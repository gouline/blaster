package main

import (
	"log"
	"os"

	gintemplate "github.com/foolin/gin-template"
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

	// Templates
	router.HTMLRender = gintemplate.New(gintemplate.TemplateConfig{
		Root:      "views",
		Extension: ".tmpl.html",
		Master:    "layouts/master",
	})

	// Static
	router.Static("/static", "static")
	router.NoRoute(handleNotFound)

	// Page
	router.GET("/", handleIndex)

	// Auth
	authGroup := router.Group("/auth")
	authGroup.GET("/initiate", handleAuthInitiate)
	authGroup.GET("/complete", handleAuthComplete)
	authGroup.GET("/logout", handleAuthLogout)

	// API
	apiGroup := router.Group("/api")
	apiGroup.GET("/suggest", handleAPISuggest)
	apiGroup.POST("/send", handleAPISend)

	router.Run(":" + port)
}

func relativeURI(c *gin.Context, path string) string {
	return "http://" + c.Request.Host + path
}
