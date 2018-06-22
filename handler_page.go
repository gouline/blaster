package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleIndex(c *gin.Context) {
	authorized := false
	teamName := ""

	api, err := slackAPI(c)
	if err == nil {
		authorized = true
		teamInfo, err := api.GetTeamInfo()
		if err == nil {
			teamName = teamInfo.Name
		}
	}

	c.HTML(http.StatusOK, "index", gin.H{
		"title":      appName,
		"authorized": authorized,
		"teamName":   teamName,
	})
}
