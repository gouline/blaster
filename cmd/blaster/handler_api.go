package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nlopes/slack"
	"github.com/traversals/blaster/pkg/scache"
)

var suggestCache = scache.New(5*time.Minute, 10*time.Minute)

func handleAPISuggest(c *gin.Context) {
	token := authorizedToken(c)
	if token == "" {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("no token"))
		return
	}

	cacheResponse := <-suggestCache.ResponseChan(hashedToken(token), func(key string) (interface{}, error) {
		client := slack.New(token)

		var suggestions []suggestion

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

			suggestions = append(suggestions, suggestion{
				Type:   "user",
				Label:  label,
				Value:  user.ID,
				Search: fmt.Sprintf("%s %s", strings.ToLower(realName), strings.ToLower(displayName)),
			})
		}

		return suggestions, nil
	})
	if cacheResponse.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, cacheResponse.Error)
		return
	}

	allSuggestions := cacheResponse.Value.([]suggestion)

	// Filter users by term
	term := strings.ToLower(c.Query("term"))
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

func handleAPISend(c *gin.Context) {
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
		AsUser: false,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, struct{}{})
}

type suggestion struct {
	Type   string `json:"type"`
	Label  string `json:"label"`
	Value  string `json:"value"`
	Search string `json:"-"`
}

type sendRequest struct {
	User    string `json:"user"`
	Message string `json:"message"`
}
