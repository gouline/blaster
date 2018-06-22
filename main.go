package main

import (
	"log"
	"os"

	"github.com/foolin/gin-template"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

const (
	appName      = "Blaster"
	cookiePrefix = "blaster_"
)

var isDebugging = gin.IsDebugging()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
		log.Printf("$PORT must be set, defaulting to %s", port)
	}

	router := gin.Default()

	router.HTMLRender = gintemplate.New(gintemplate.TemplateConfig{
		Root:      "views",
		Extension: ".tmpl.html",
		Master:    "layouts/master",
	})

	router.Static("/static", "static")

	// Page
	router.GET("/", handleIndex)

	// Auth
	authPrefix := "/auth"
	router.GET(authPrefix+"/initiate", handleAuthInitiate)
	router.GET(authPrefix+"/complete", handleAuthComplete)
	router.GET(authPrefix+"/logout", handleAuthLogout)

	// API
	apiPrefix := "/api"
	router.GET(apiPrefix+"/suggest", handleAPISuggest)
	router.POST(apiPrefix+"/send", handleAPISend)

	router.Run(":" + port)
}

func relativeURI(c *gin.Context, path string) string {
	return "http://" + c.Request.Host + path
}

type checkResponse struct {
	Token string `json:"token"`
}
