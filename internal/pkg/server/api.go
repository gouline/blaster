package server

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gouline/blaster/internal/pkg/slack"
	"github.com/labstack/echo/v4"
)

// APISuggest handles /api/suggest.
func (s *Server) handleAPISuggest(c echo.Context) error {
	slackCtx := s.slack.Context(c)
	if !slackCtx.Authorized {
		return c.String(http.StatusUnauthorized, "no token")
	}

	destinations, err := slackCtx.Destinations()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	suggestions := suggestDestinations(c.QueryParam("term"), destinations)

	return c.JSON(http.StatusOK, suggestions)
}

// handleAPISend handles /api/send.
func (s *Server) handleAPISend(c echo.Context) error {
	slackCtx := s.slack.Context(c)
	if !slackCtx.Authorized {
		return c.NoContent(http.StatusUnauthorized)
	}

	var request sendRequest
	if err := c.Bind(&request); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if err := slackCtx.SendMessage(request.User, request.Message, request.AsUser); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, struct{}{})
}

// suggestDestinations filters destionations into suggestions by search term.
func suggestDestinations(term string, destinations []*slack.Destination) []*suggestion {
	term = " " + sanitizeSearchTerm(term)

	suggestions := []*suggestion{}
	for _, dest := range destinations {
		searchable := " " + sanitizeSearchTerm(strings.ToLower(dest.Name)+" "+strings.ToLower(dest.DisplayName))

		if strings.Contains(searchable, term) {
			children := []*suggestion{}
			for _, child := range dest.Children {
				children = append(children, &suggestion{
					Type:  sanitizeCSV(child.Type),
					Label: sanitizeCSV(suggestionLabel(child.Name, child.DisplayName)),
					Value: sanitizeCSV(child.ID),
				})
			}

			suggestions = append(suggestions, &suggestion{
				Type:     sanitizeCSV(dest.Type),
				Label:    sanitizeCSV(suggestionLabel(dest.Name, dest.DisplayName)),
				Value:    sanitizeCSV(dest.ID),
				Children: children,
			})

			if len(suggestions) == 10 {
				break
			}
		}
	}

	return suggestions
}

// sanitizeSearchTerm removes lowercases search term and removes skippable characters.
func sanitizeSearchTerm(s string) string {
	symbolReg := regexp.MustCompile("[!-/:-@[-`{-~]+")
	return symbolReg.ReplaceAllString(strings.ToLower(s), "")
}

// sanitizeCSV removes commas for comma-separated values.
func sanitizeCSV(s string) string {
	return strings.Replace(s, ",", "", -1)
}

// suggestionLabel formats label with a mandatory name and an optional display name.
func suggestionLabel(name, displayName string) string {
	label := name
	if displayName != "" {
		label += " (" + displayName + ")"
	}
	return label
}

type sendRequest struct {
	User    string `json:"user"`
	Message string `json:"message"`
	AsUser  bool   `json:"as_user"`
}

type suggestion struct {
	Type     string        `json:"type"`
	Label    string        `json:"label"`
	Value    string        `json:"value"`
	Children []*suggestion `json:"children,omitempty"`
}
