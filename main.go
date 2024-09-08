package main

import (
	"fmt"
	"os"

	"github.com/gouline/blaster/internal/pkg/server"
)

func main() {
	s, err := server.NewServer(server.Config{
		Debug:         os.Getenv("DEBUG") == "1",
		Host:          os.Getenv("HOST"),
		Port:          os.Getenv("PORT"),
		CertFile:      os.Getenv("CERT_FILE"),
		KeyFile:       os.Getenv("KEY_FILE"),
		StaticRoot:    "static",
		TemplatesRoot: "templates",
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create server: %s", err))
	}

	s.Run()
}
