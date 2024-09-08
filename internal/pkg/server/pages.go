package server

import (
	"net/http"
	"time"

	"github.com/gouline/blaster/internal/pkg/format"
	"github.com/gouline/blaster/internal/pkg/scache"
	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

var teamCache = scache.New(12*time.Hour, 12*time.Hour)

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
	authorized := false
	teamName := ""

	if token := s.authorizedToken(c); token != "" {
		authorized = true

		cacheResponse := <-teamCache.ResponseChan(format.HashToken(token), func(key string) (interface{}, error) {
			client := slack.New(token)

			teamInfo, err := client.GetTeamInfo()
			if err != nil {
				return nil, err
			}

			return teamInfo.Name, err
		})
		if cacheResponse.Error == nil {
			teamName = cacheResponse.Value.(string)
		}

		// Build other caches
		go func() {
			<-buildSuggestCache(token)
		}()
	}

	data["debugging"] = s.config.Debug
	data["authorized"] = authorized
	data["teamName"] = teamName

	return data
}
