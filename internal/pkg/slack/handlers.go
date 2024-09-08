package slack

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

// Middleware detects 'code' query parameter and completes authentication.
func (s *Slack) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if code := c.QueryParam("code"); code != "" {
			redirectURI := redirectURI(c, c.Request().RequestURI)
			response, err := slack.GetOAuthResponse(http.DefaultClient, s.clientID, s.clientSecret, code, redirectURI)
			if err != nil {
				return c.String(http.StatusUnauthorized, err.Error())
			}

			s.setToken(c, response.AccessToken)
			return c.Redirect(http.StatusSeeOther, redirectURI)
		}

		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}

// HandleLogin initiates Slack authorization.
func (s *Slack) HandleLogin(c echo.Context) error {
	redirectURI := redirectURI(c, c.Request().Referer())
	redirectURI, err := s.authorizeURI(redirectURI)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.Redirect(http.StatusFound, redirectURI)
}

// HandleLogout clears Slack credentials.
func (s *Slack) HandleLogout(c echo.Context) error {
	s.setToken(c, "")

	redirectURI := redirectURI(c, c.Request().Referer())
	return c.Redirect(http.StatusFound, redirectURI)
}
