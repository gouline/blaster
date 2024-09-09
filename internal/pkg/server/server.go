package server

import (
	"fmt"

	"github.com/gouline/blaster/internal/pkg/slack"
	"github.com/gouline/blaster/internal/pkg/templates"
	"github.com/labstack/echo/v4"
)

const (
	appName = "Blaster"
)

type Config struct {
	Debug bool

	Host     string
	Port     string
	CertFile string
	KeyFile  string

	StaticRoot    string
	TemplatesRoot string

	SlackClientID     string
	SlackClientSecret string
}

type Server struct {
	config Config
	echo   *echo.Echo
	slack  *slack.Slack
}

func New(config Config) (*Server, error) {
	s := &Server{
		config: config,
		echo:   echo.New(),
		slack:  slack.New(config.SlackClientID, config.SlackClientSecret),
	}

	s.echo.Debug = config.Debug

	// Slack auth
	s.echo.Use(s.slack.Middleware)
	s.echo.GET("/login", s.slack.HandleLogin)
	s.echo.GET("/logout", s.slack.HandleLogout)

	s.echo.Static("/static", config.StaticRoot)

	var err error
	s.echo.Renderer, err = templates.New(config.TemplatesRoot, "layout.html")
	if err != nil {
		return nil, err
	}

	// Pages
	s.echo.GET("/", s.handleIndex)
	s.echo.RouteNotFound("/*", s.handleNotFound)

	// API
	apiGroup := s.echo.Group("/api")
	apiGroup.GET("/suggest", s.handleAPISuggest)
	apiGroup.POST("/send", s.handleAPISend)

	return s, nil
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	if s.config.CertFile != "" && s.config.KeyFile != "" {
		return s.echo.StartTLS(addr, s.config.CertFile, s.config.KeyFile)
	}
	return s.echo.Start(addr)
}
