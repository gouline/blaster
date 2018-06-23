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

	c.HTML(http.StatusOK, "index", gin.H{
		"title":      appName,
		"authorized": authorized,
		"teamName":   teamName,
	})
}
