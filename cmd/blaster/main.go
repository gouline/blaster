package main

import (
	"log"
	"os"

	gintemplate "github.com/foolin/gin-template"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/gouline/blaster/internal/pkg/handlers"
)

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
	router.NoRoute(handlers.NotFound)

	// Page
	router.GET("/", handlers.Index)

	// Auth
	authGroup := router.Group("/auth")
	authGroup.GET("/initiate", handlers.AuthInitiate)
	authGroup.GET("/complete", handlers.AuthComplete)
	authGroup.GET("/logout", handlers.AuthLogout)

	// API
	apiGroup := router.Group("/api")
	apiGroup.GET("/suggest", handlers.APISuggest)
	apiGroup.POST("/send", handlers.APISend)

	router.Run(":" + port)
}
