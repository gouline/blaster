package main

import (
	"fmt"
	"log"
	"os"

	gintemplate "github.com/foolin/gin-template"
	"github.com/gin-gonic/gin"
	"github.com/gouline/blaster/internal/pkg/handlers"
)

func main() {
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
		log.Printf("$HOST can be set, defaulting to %s", host)
	}

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

	router.Run(fmt.Sprintf("%s:%s", host, port))
}
