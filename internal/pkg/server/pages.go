package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// handleIndex handles /.
func (s *Server) handleIndex(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", s.baseData(c, map[string]interface{}{
		"title": appName,
	}))
}

// handleNotFound handles 404 Not Found errors.
func (s *Server) handleNotFound(c echo.Context) error {
	return c.Render(http.StatusNotFound, "error.html", s.baseData(c, map[string]interface{}{
		"title":   "404 Not Found",
		"message": "Unfortunately, this page doesn't seem to exist. Are you sure about that URL?",
	}))
}

func (s *Server) baseData(c echo.Context, data map[string]interface{}) map[string]interface{} {
	data["slack"] = s.slack.Context(c)
	return data
}
