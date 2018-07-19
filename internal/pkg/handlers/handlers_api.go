package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mgouline/slack"
	"github.com/traversals/blaster/internal/pkg/scache"
	"github.com/traversals/blaster/internal/pkg/utils"
)

var suggestCache = scache.New(5*time.Minute, 10*time.Minute)

// APISuggest handles /api/suggest.
func APISuggest(c *gin.Context) {
	token := authorizedToken(c)
	if token == "" {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("no token"))
		return
	}

	cacheResponse := <-buildSuggestCache(token)
	if cacheResponse.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, cacheResponse.Error)
		return
	}

	allSuggestions := cacheResponse.Value.([]suggestion)

	// Filter out all symbols from term
	symbolReg := utils.NewAllSymbolsRegexp()
	term := strings.ToLower(" " + c.Query("term"))
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

	c.JSON(http.StatusOK, suggestions)
}

func buildSuggestCache(token string) <-chan scache.Response {
	return suggestCache.ResponseChan(hashedToken(token), func(key string) (interface{}, error) {
		client := slack.New(token)

		symbolReg := utils.NewAllSymbolsRegexp()

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

		usergroups, err := client.GetUserGroups(true)

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

// APISend handles /api/send.
func APISend(c *gin.Context) {
	token := authorizedToken(c)
	if token == "" {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("no token"))
		return
	}

	client := slack.New(token)

	// Bind JSON request
	var request sendRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Open/get channel by user ID
	_, _, channelID, err := client.OpenIMChannel(request.User)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Post message to opened channel
	_, _, err = client.PostMessage(channelID, request.Message, slack.PostMessageParameters{
		AsUser: request.AsUser,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, struct{}{})
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
