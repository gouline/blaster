package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/gouline/blaster/internal/pkg/slack"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareAuth(t *testing.T) {
	r := newRequestTester(http.MethodGet, "/?code=123", nil)

	err := r.Server.middlewareAuth(func(c echo.Context) error {
		return nil
	})(r.Context)

	if assert.NoError(t, err) {
		if assert.Equal(t, http.StatusSeeOther, r.Response.Code, r.Response.Body) {
			setCookies := r.Response.Header()["Set-Cookie"]
			if assert.NotEmpty(t, setCookies) {
				assert.Contains(t, setCookies[0], url.QueryEscape("\"token\":\"authenticated\""))
			}
		}
	}
}

func TestMiddlewareAuthError(t *testing.T) {
	r := newRequestTester(http.MethodGet, "/?code=123", nil)
	r.Session.AuthenticateError = errors.New("simulated")

	err := r.Server.middlewareAuth(func(c echo.Context) error {
		return nil
	})(r.Context)

	if assert.NoError(t, err) {
		if assert.Equal(t, http.StatusUnauthorized, r.Response.Code, r.Response.Body) {
			assert.Contains(t, r.Response.Body.String(), "simulated")
		}
	}
}

func TestMiddlewareAuthPassthrough(t *testing.T) {
	r := newRequestTester(http.MethodGet, "/", nil)

	err := r.Server.middlewareAuth(func(c echo.Context) error {
		return nil
	})(r.Context)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, r.Response.Code)
	}
}

func TestHandleAuthLogin(t *testing.T) {
	for _, test := range []struct {
		session      slack.Session
		expectedCode int
	}{
		{
			session: &mockSlackSession{
				ClientSession: &slack.ClientSession{},
			},
			expectedCode: http.StatusFound,
		},
		{
			session: &mockSlackSession{
				ClientSession:     &slack.ClientSession{},
				AuthorizeURLError: errors.New("dummy"),
			},
			expectedCode: http.StatusInternalServerError,
		},
	} {
		r := newRequestTester(http.MethodGet, "/", nil)
		r.Context.Set(cookieSession, test.session)

		if assert.NoError(t, r.Server.handleAuthLogin(r.Context)) {
			assert.Equal(t, test.expectedCode, r.Response.Code)
			if r.Response.Code == http.StatusFound {
				redirectLocation := r.Response.Header()["Location"]
				redirectURL, err := url.Parse(redirectLocation[0])
				if assert.NoError(t, err) {
					assert.Equal(t, "/oauth/authorize", redirectURL.Path)
					assert.Equal(t, mockClientID, redirectURL.Query()["client_id"][0])
				}
			}
		}
	}
}

func TestHandleAuthLogout(t *testing.T) {
	r := newRequestTester(http.MethodGet, "/", nil)
	r.Authenticate("existing", "")

	if assert.NoError(t, r.Server.handleAuthLogout(r.Context)) {
		assert.Equal(t, http.StatusFound, r.Response.Code)

		setCookie := r.Response.Header()["Set-Cookie"][0]
		assert.Contains(t, setCookie, "Max-Age=0;")
		fmt.Println(setCookie)
	}
}

func TestSessionFallbackCreation(t *testing.T) {
	r := newRequestTester(http.MethodGet, "/", nil)
	r.Context.Set(cookieSession, nil)
	r.Request.AddCookie(&http.Cookie{
		Name:     cookieSession,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		Value:    url.QueryEscape(slack.NewSession().Marshal()),
		MaxAge:   86400,
	})
	assert.NotNil(t, r.Server.session(r.Context))
}

func TestSessionMalformedCookie(t *testing.T) {
	for _, test := range []struct {
		cookie *http.Cookie
	}{
		{nil},
		{&http.Cookie{
			Name:  cookieSession,
			Path:  "/",
			Value: "%7z",
		}},
		{&http.Cookie{
			Name:  cookieSession,
			Path:  "/",
			Value: url.QueryEscape(slack.NewSession().Marshal()),
		}},
	} {

		r := newRequestTester(http.MethodGet, "/", nil)
		if test.cookie != nil {
			r.Request.AddCookie(test.cookie)
		}
		assert.NotNil(t, r.Server.session(r.Context))
	}
}
