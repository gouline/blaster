package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nlopes/slack"
)

func handleAPISuggest(c *gin.Context) {
	api, err := slackAPI(c)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	term := strings.ToLower(c.Query("term"))

	// Get all users
	users, err := api.GetUsers()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Filter users by term
	suggestions := []suggestion{}
	for _, user := range users {
		id := user.ID
		realName := user.Profile.RealName
		displayName := user.Profile.DisplayName

		if strings.Contains(strings.ToLower(realName), term) ||
			strings.Contains(strings.ToLower(displayName), term) {

			label := realName
			if label == "" {
				label = displayName
			}

			suggestions = append(suggestions, suggestion{
				Type:  "user",
				Label: label,
				Value: id,
			})
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
	Type  string `json:"type"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type sendRequest struct {
	User    string `json:"user"`
	Message string `json:"message"`
}
