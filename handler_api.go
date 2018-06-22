package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/gin-gonic/gin"
	"github.com/nlopes/slack"
)

var suggestCache = cache.New(5*time.Minute, 10*time.Minute)

func handleAPISuggest(c *gin.Context) {
	api, err := slackAPI(c)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	term := strings.ToLower(c.Query("term"))

	var allSuggestions []suggestion

	cacheKey := authorizedTokenHashed(c)
	cached, found := suggestCache.Get(cacheKey)
	if found {
		// Retrieve from cache
		allSuggestions = cached.([]suggestion)
	} else {
		// Get all users
		users, err := api.GetUsers()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		allSuggestions = []suggestion{}

		for _, user := range users {
			realName := user.Profile.RealName
			displayName := user.Profile.DisplayName

			label := realName
			if label == "" {
				label = displayName
			}

			allSuggestions = append(allSuggestions, suggestion{
				Type:   "user",
				Label:  label,
				Value:  user.ID,
				Search: fmt.Sprintf("%s %s", strings.ToLower(realName), strings.ToLower(displayName)),
			})
		}

		suggestCache.Set(cacheKey, allSuggestions, cache.DefaultExpiration)
	}

	// Filter users by term
	suggestions := []suggestion{}
	for _, suggestion := range allSuggestions {
		if strings.Contains(suggestion.Search, term) {
			suggestions = append(suggestions, suggestion)
		}
	}

	c.JSON(http.StatusOK, suggestions)
}

func handleAPISend(c *gin.Context) {
	api, err := slackAPI(c)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	// Bind JSON request
	var request sendRequest
	err = c.BindJSON(&request)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Open/get channel by user ID
	_, _, channelID, err := api.OpenIMChannel(request.User)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Post message to opened channel
	_, _, err = api.PostMessage(channelID, request.Message, slack.PostMessageParameters{
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
