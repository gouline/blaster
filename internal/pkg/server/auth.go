package server

import (
	"net/http"
	"net/url"

	"github.com/gouline/blaster/internal/pkg/slack"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Middleware detects 'code' query parameter and completes authentication.
func (s *Server) middlewareAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := s.session(c)
		redirectURI := redirectURI(c, c.Request().RequestURI)
		authenticated, err := session.Authenticate(s.config.SlackClientID, s.config.SlackClientSecret, redirectURI, c.QueryParams())
		if authenticated {
			if err != nil {
				return c.String(http.StatusUnauthorized, err.Error())
			}
			s.setSession(c, session)
			return c.Redirect(http.StatusSeeOther, redirectURI)
		}

		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}

// HandleLogin initiates Slack authorization.
func (s *Server) handleAuthLogin(c echo.Context) error {
	redirectURI := redirectURI(c, c.Request().Referer())
	authorizeURL, err := s.session(c).AuthorizeURL(s.config.SlackClientID, redirectURI)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusFound, authorizeURL)
}

// handleAuthLogout clears session.
func (s *Server) handleAuthLogout(c echo.Context) error {
	session := s.session(c)
	session.Reset()
	s.setSession(c, session)

	return c.Redirect(http.StatusFound, redirectURI(c, c.Request().Referer()))
}

// session retrieves current session from context or cookie.
func (s *Server) session(c echo.Context) slack.Session {
	session, ok := c.Get(cookieSession).(slack.Session)
	if !ok {
		session = &slack.ClientSession{}
		if cookie, err := c.Cookie(cookieSession); err == nil {
			cookieValue, err := url.QueryUnescape(cookie.Value)
			if err != nil {
				s.config.Logger.Warn("found unescapable cookie", zap.Error(err))
			} else {
				session.Unmarshal(cookieValue)
			}
		}
		c.Set(cookieSession, session)
	}
	return session
}

// setSession sets new session to context and cookie.
func (s *Server) setSession(c echo.Context, session slack.Session) {
	cookie := &http.Cookie{
		Name:     cookieSession,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	}
	if session.IsAuthenticated() {
		cookie.Value = url.QueryEscape(session.Marshal())
		cookie.MaxAge = 86400
	} else {
		cookie.Value = ""
		cookie.MaxAge = -1
	}
	c.SetCookie(cookie)
	c.Set(cookieSession, session)
}
