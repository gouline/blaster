package slack

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	baseURL      = "https://slack.com"
	cookiePrefix = "slack_"
)

var (
	scopes = []string{
		"team:read",
		"users:read",
		"usergroups:read",
		"im:write",
		"chat:write:bot",
		"chat:write:user",
	}
)

type Config struct {
	Logger *zap.Logger

	ClientID     string
	ClientSecret string
}

type Slack struct {
	config Config
}

func New(config Config) (*Slack, error) {
	s := &Slack{
		config: config,
	}

	if config.ClientID == "" || config.ClientSecret == "" {
		return s, fmt.Errorf("missing Slack credentials")
	}

	return s, nil
}

func (s *Slack) authorizeURI(redirectURI string) (string, error) {
	redirectURL, err := url.Parse(baseURL + "/oauth/authorize")
	if err != nil {
		return "", err
	}
	q := redirectURL.Query()
	q.Set("client_id", s.config.ClientID)
	q.Set("scope", strings.Join(scopes, ","))
	q.Set("redirect_uri", redirectURI)
	redirectURL.RawQuery = q.Encode()

	return redirectURL.String(), nil
}

// token fetches authorized token from HTTP cookie.
func (s *Slack) token(c echo.Context) string {
	cookie, err := c.Cookie(cookiePrefix + "token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// setToken sets authorized token to HTTP cookie.
func (s *Slack) setToken(c echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     cookiePrefix + "token",
		Value:    url.QueryEscape(token),
		MaxAge:   86400,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	})
}

// redirectURI creates a stable URI for redirects.
// Removes query parameters and trailing slashes.
func redirectURI(c echo.Context, uri string) string {
	url, _ := url.Parse(uri)
	url.RawQuery = ""
	if url.Scheme == "" {
		url.Scheme = c.Scheme()
	}
	if url.Host == "" {
		url.Host = c.Request().Host
	}
	url.Path, _ = strings.CutSuffix(url.Path, "/")
	return url.String()
}
