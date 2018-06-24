package main

import (
	"net/http"
	"time"

	"github.com/nlopes/slack"
	"github.com/traversals/blaster/pkg/scache"

	"github.com/gin-gonic/gin"
)

var teamCache = scache.New(12*time.Hour, 12*time.Hour)

func handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index", baseH(c, gin.H{
		"title": appName,
	}))
}

func handleNotFound(c *gin.Context) {
	c.HTML(http.StatusNotFound, "error", baseH(c, gin.H{
		"title":   "404 Not Found",
		"message": "Unfortunately, this page doesn’t seem to exist. Are you sure about that URL?",
	}))
}

func baseH(c *gin.Context, h gin.H) gin.H {
	authorized := false
	teamName := ""

	token := authorizedToken(c)
	if token != "" {
		authorized = true

		cacheResponse := <-teamCache.ResponseChan(hashedToken(token), func(key string) (interface{}, error) {
			client := slack.New(token)

			teamInfo, err := client.GetTeamInfo()
			if err != nil {
				return nil, err
			}

			return teamInfo.Name, err
		})
		if cacheResponse.Error == nil {
			teamName = cacheResponse.Value.(string)
		}
	}

	h["authorized"] = authorized
	h["teamName"] = teamName

	return h
}
