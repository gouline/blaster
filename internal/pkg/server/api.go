package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gouline/blaster/internal/pkg/format"
	"github.com/gouline/blaster/internal/pkg/scache"
	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

var suggestCache = scache.New(5*time.Minute, 10*time.Minute)

// APISuggest handles /api/suggest.
func (s *Server) handleAPISuggest(c echo.Context) error {
	token := s.authorizedToken(c)
	if token == "" {
		return c.String(http.StatusUnauthorized, "no token")
	}

	cacheResponse := <-buildSuggestCache(token)
	if cacheResponse.Error != nil {
		return c.String(http.StatusInternalServerError, cacheResponse.Error.Error())
	}

	allSuggestions := cacheResponse.Value.([]suggestion)

	// Filter out all symbols from term
	symbolReg := format.NewAllSymbolsRegexp()
	term := strings.ToLower(" " + c.QueryParam("term"))
	term = symbolReg.ReplaceAllString(term, "")

	// Filter users by term
	suggestions := []suggestion{}
	for _, suggestion := range allSuggestions {
		if strings.Contains(suggestion.Search, term) {
			suggestions = append(suggestions, suggestion)
			if len(suggestions) == 10 {
				break
			}
		}
	}

	return c.JSON(http.StatusOK, suggestions)
}

// handleAPISend handles /api/send.
func (s *Server) handleAPISend(c echo.Context) error {
	token := s.authorizedToken(c)
	if token == "" {
		return c.String(http.StatusUnauthorized, "no token")
	}

	client := slack.New(token)

	// Bind JSON request
	var request sendRequest
	err := c.Bind(&request)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Open/get channel by user ID
	channel, _, _, err := client.OpenConversation(&slack.OpenConversationParameters{
		Users: []string{
			request.User,
		},
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Post message to opened channel
	_, _, err = client.PostMessage(
		channel.ID,
		slack.MsgOptionText(request.Message, false),
		slack.MsgOptionAsUser(request.AsUser),
	)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, struct{}{})
}

func buildSuggestCache(token string) <-chan scache.Response {
	return suggestCache.ResponseChan(format.HashToken(token), func(key string) (interface{}, error) {
		client := slack.New(token)

		symbolReg := format.NewAllSymbolsRegexp()

		var suggestions []suggestion

		userLookup := map[string]suggestion{}

		// Get all users
		users, err := client.GetUsers()
		if err != nil {
			return nil, err
		}

		suggestions = []suggestion{}

		for _, user := range users {
			if user.Deleted || user.IsBot {
				continue
			}

			realName := user.Profile.RealName
			displayName := user.Profile.DisplayName

			// Format label based on availability
			label := realName
			if displayName != "" {
				label += " (" + displayName + ")"
			}

			// Filter out all symbols from search string
			search := fmt.Sprintf(" %s %s", strings.ToLower(realName), strings.ToLower(displayName))
			search = symbolReg.ReplaceAllString(search, "")

			// Sanitize labels and values
			sanitize := func(s string) string {
				return strings.Replace(s, ",", "", -1)
			}

			s := suggestion{
				Type:   "user",
				Label:  sanitize(label),
				Value:  sanitize(user.ID),
				Search: search,
			}

			suggestions = append(suggestions, s)

			userLookup[user.ID] = s
		}

		usergroups, err := client.GetUserGroups(slack.GetUserGroupsOptionIncludeUsers(true))
		if err != nil {
			return nil, err
		}

		for _, usergroup := range usergroups {
			if !usergroup.IsUserGroup {
				continue
			}

			children := []suggestion{}

			for _, userID := range usergroup.Users {
				user, found := userLookup[userID]
				if !found {
					continue
				}

				children = append(children, user)
			}

			name := usergroup.Name
			handle := usergroup.Handle
			label := name + " (" + handle + ")"

			// Filter out all symbols from search string
			search := fmt.Sprintf(" %s %s", strings.ToLower(name), strings.ToLower(handle))
			search = symbolReg.ReplaceAllString(search, "")

			suggestions = append(suggestions, suggestion{
				Type:     "usergroup",
				Label:    label,
				Value:    "null",
				Search:   search,
				Children: children,
			})
		}

		return suggestions, nil
	})
}

type suggestion struct {
	Type     string       `json:"type"`
	Label    string       `json:"label"`
	Value    string       `json:"value"`
	Children []suggestion `json:"children,omitempty"`
	Search   string       `json:"-"`
}

type sendRequest struct {
	User    string `json:"user"`
	Message string `json:"message"`
	AsUser  bool   `json:"as_user"`
}
