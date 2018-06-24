package main

import (
	"log"
	"os"

	rice "github.com/GeertJohan/go.rice"
	gintemplate "github.com/foolin/gin-template"
	"github.com/foolin/gin-template/supports/gorice"
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

	projectRoot := "../.."

	router := gin.Default()

	// Templates
	router.HTMLRender = gorice.NewWithConfig(rice.MustFindBox(projectRoot+"/views"), gintemplate.TemplateConfig{
		Root:      "views",
		Extension: ".tmpl.html",
		Master:    "layouts/master",
	})

	// Static
	router.StaticFS("/static", rice.MustFindBox(projectRoot+"/static").HTTPBox())

	// Page
	router.GET("/", handleIndex)
	router.NoRoute(handleNotFound)

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
