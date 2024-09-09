package server

import (
	"fmt"
	"os"

	"github.com/gouline/blaster/internal/pkg/slack"
	"github.com/gouline/blaster/internal/pkg/templates"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

const (
	appName = "Blaster"
)

type Config struct {
	Logger *zap.Logger

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
	var err error
	s := &Server{
		config: config,
		echo:   echo.New(),
	}

	s.echo.Debug = config.Debug

	s.slack, err = slack.New(slack.Config{
		Logger:       config.Logger,
		ClientID:     config.SlackClientID,
		ClientSecret: config.SlackClientSecret,
	})
	if err != nil {
		return s, fmt.Errorf("failed to init Slack: %w", err)
	}

	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.Gzip())
	if s.echo.Debug {
		s.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogURI:    true,
			LogStatus: true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				config.Logger.Info("request",
					zap.String("URI", v.URI),
					zap.Int("status", v.Status),
				)
				return nil
			},
		}))
	}

	// Static
	if f, err := os.Stat(config.StaticRoot); os.IsNotExist(err) {
		return s, fmt.Errorf("static not found: %w", err)
	} else if err == nil && !f.IsDir() {
		return s, fmt.Errorf("static not directory")
	}
	s.echo.Static("/static", config.StaticRoot)

	// Templates
	s.echo.Renderer, err = templates.New(templates.Config{
		Logger:     config.Logger,
		RootPath:   config.TemplatesRoot,
		LayoutFile: "layout.html",
	})
	if err != nil {
		return nil, fmt.Errorf("templates parsing failed: %w", err)
	}

	// Slack auth
	s.echo.Use(s.slack.Middleware)
	s.echo.GET("/login", s.slack.HandleLogin)
	s.echo.GET("/logout", s.slack.HandleLogout)

	// Pages
	s.echo.GET("/", s.handleIndex)
	s.echo.RouteNotFound("/*", s.handleNotFound)

	// API
	apiGroup := s.echo.Group("/api")
	apiGroup.GET("/suggest", s.handleAPISuggest)
	apiGroup.POST("/send", s.handleAPISend)

	return s, nil
}

// Start starts HTTP or HTTPS server, depending on the presence of cert/key.
func (s *Server) Start() error {
	s.config.Logger.Info("Starting server",
		zap.String("host", s.config.Host),
		zap.String("port", s.config.Port),
		zap.String("certFile", s.config.CertFile),
		zap.String("keyFile", s.config.KeyFile))

	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	if s.config.CertFile != "" && s.config.KeyFile != "" {
		return s.echo.StartTLS(addr, s.config.CertFile, s.config.KeyFile)
	}
	return s.echo.Start(addr)
}
