package server

import (
	"fmt"

	"github.com/gouline/blaster/internal/pkg/templates"
	"github.com/labstack/echo/v4"
)

const (
	appName      = "Blaster"
	cookiePrefix = "blaster_"
)

type Config struct {
	Debug bool

	Host string
	Port string

	CertFile string
	KeyFile  string

	StaticRoot    string
	TemplatesRoot string
}

type Server struct {
	config Config
	echo   *echo.Echo
}

func NewServer(config Config) (*Server, error) {
	s := &Server{
		config: config,
		echo:   echo.New(),
	}

	s.echo.Debug = config.Debug

	s.echo.Static("/static", config.StaticRoot)

	var err error
	s.echo.Renderer, err = templates.NewTemplates(config.TemplatesRoot, "layout.html")
	if err != nil {
		return nil, err
	}

	s.echo.RouteNotFound("/*", s.handleNotFound)

	s.echo.GET("/", s.handleIndex)

	authGroup := s.echo.Group("/auth")
	authGroup.GET("/initiate", s.handleAuthInitiate)
	authGroup.GET("/complete", s.handleAuthComplete)
	authGroup.GET("/logout", s.handleAuthLogout)

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
