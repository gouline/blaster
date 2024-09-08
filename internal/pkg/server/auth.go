package server

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gouline/blaster/internal/pkg/format"
	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

const slackBaseURL = "https://slack.com"

var (
	slackClientID     = os.Getenv("SLACK_CLIENT_ID")
	slackClientSecret = os.Getenv("SLACK_CLIENT_SECRET")

	slackAPIScopes = []string{
		"team:read",
		"users:read",
		"usergroups:read",
		"im:write",
		"chat:write:bot",
		"chat:write:user",
	}
)

// handleAuthInitiate handles /auth/initiate.
func (s *Server) handleAuthInitiate(c echo.Context) error {
	redirectURI, err := authorizeURI(format.RelativeURI(c, "/auth/complete"))
	if err != nil {
		log.Fatal(err)
	}
	return c.Redirect(http.StatusFound, redirectURI)
}

// handleAuthComplete handles /auth/complete.
func (s *Server) handleAuthComplete(c echo.Context) error {
	code := c.QueryParam("code")

	if code != "" {
		response, err := slack.GetOAuthResponse(http.DefaultClient, slackClientID, slackClientSecret, code, format.RelativeURI(c, "/auth/complete"))
		if err != nil {
			return c.String(http.StatusUnauthorized, err.Error())
		}

		s.setAuthorizedToken(c, response.AccessToken)
	}

	return c.Redirect(http.StatusFound, format.RelativeURI(c, "/"))
}

// handleAuthLogout handles /auth/logout.
func (s *Server) handleAuthLogout(c echo.Context) error {
	s.setAuthorizedToken(c, "")

	return c.Redirect(http.StatusFound, format.RelativeURI(c, "/"))
}

func authorizeURI(redirectURI string) (string, error) {
	redirectURL, err := url.Parse(slackBaseURL + "/oauth/authorize")
	if err != nil {
		return "", err
	}
	q := redirectURL.Query()
	q.Set("client_id", slackClientID)
	q.Set("scope", strings.Join(slackAPIScopes, ","))
	q.Set("redirect_uri", redirectURI)
	redirectURL.RawQuery = q.Encode()

	return redirectURL.String(), nil
}

func (s *Server) setAuthorizedToken(c echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     cookiePrefix + "slacktoken",
		Value:    url.QueryEscape(token),
		MaxAge:   86400,
		Path:     "/",
		Secure:   !s.config.Debug,
		HttpOnly: true,
	})
}

func (s *Server) authorizedToken(c echo.Context) string {
	tokenCookie, err := c.Cookie(cookiePrefix + "slacktoken")
	if err != nil {
		return ""
	}
	return tokenCookie.Value
}
