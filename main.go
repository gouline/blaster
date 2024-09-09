package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gouline/blaster/internal/pkg/server"
)

func main() {
	s, err := server.New(server.Config{
		Debug:             os.Getenv("DEBUG") == "1",
		Host:              os.Getenv("HOST"),
		Port:              os.Getenv("PORT"),
		CertFile:          os.Getenv("CERT_FILE"),
		KeyFile:           os.Getenv("KEY_FILE"),
		StaticRoot:        "static",
		TemplatesRoot:     "templates",
		SlackClientID:     os.Getenv("SLACK_CLIENT_ID"),
		SlackClientSecret: os.Getenv("SLACK_CLIENT_SECRET"),
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create server: %s", err))
	}

	log.Fatal(s.Run())
}
