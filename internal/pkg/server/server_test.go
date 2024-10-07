package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gouline/blaster/internal/pkg/slack"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const (
	mockClientID     = "dummy_id"
	mockClientSecret = "dummy_secret"
)

type requestTester struct {
	Server   *Server
	Request  *http.Request
	Response *httptest.ResponseRecorder
	Echo     *echo.Echo
	Context  echo.Context
	Session  *mockSlackSession
}

func (r *requestTester) Authenticate(token, team string) {
	r.Session = &mockSlackSession{ClientSession: &slack.ClientSession{
		Token: token,
		Team:  team,
	}}
	r.Context.Set(cookieSession, r.Session)
}

func newRequestTester(method, target string, body io.Reader) *requestTester {
	r := &requestTester{}

	var err error
	r.Server, err = New(Config{
		Logger:            zap.Must(zap.NewDevelopment()),
		Debug:             true,
		SlackClientID:     mockClientID,
		SlackClientSecret: mockClientSecret,
		StaticRoot:        "../../../static",
		TemplatesRoot:     "../../../templates",
	})
	if err != nil {
		fmt.Println(err)
		panic("must server")
	}

	r.Request = httptest.NewRequest(method, target, body)
	r.Response = httptest.NewRecorder()
	r.Context = r.Server.echo.NewContext(r.Request, r.Response)
	r.Authenticate("", "")

	return r
}

type mockSlackSession struct {
	*slack.ClientSession
	AuthenticateError    error
	AuthorizeURLError    error
	GetDestinationsError error
	PostMessageError     error
}

func (s *mockSlackSession) Authenticate(clientID, clientSecret, redirectURI string, query url.Values) (bool, error) {
	if _, ok := query["code"]; ok {
		s.ClientSession.Token = "authenticated"
		return true, s.AuthenticateError
	}
	return false, s.AuthenticateError
}

func (s *mockSlackSession) AuthorizeURL(clientID, redirectURI string) (string, error) {
	if s.AuthorizeURLError != nil {
		return "", s.AuthorizeURLError
	}
	return s.ClientSession.AuthorizeURL(clientID, redirectURI)
}

func (s *mockSlackSession) GetDestinations() ([]*slack.Destination, error) {
	return []*slack.Destination{}, s.GetDestinationsError
}

func (s *mockSlackSession) PostMessage(user, message string, asUser bool) error {
	return s.PostMessageError
}

func TestServerChecks(t *testing.T) {
	for _, test := range []struct {
		config        Config
		errorContains string
	}{
		{
			config: Config{
				Logger:            zap.Must(zap.NewDevelopment()),
				SlackClientID:     mockClientID,
				SlackClientSecret: mockClientSecret,
				StaticRoot:        "../../static",
				TemplatesRoot:     "../../../templates",
			},
			errorContains: "static not found",
		},
		{
			config: Config{
				Logger:            zap.Must(zap.NewDevelopment()),
				SlackClientID:     mockClientID,
				SlackClientSecret: mockClientSecret,
				StaticRoot:        "../../../static/css/main.css",
				TemplatesRoot:     "../../../templates",
			},
			errorContains: "static not directory",
		},
		{
			config: Config{
				Logger:            zap.Must(zap.NewDevelopment()),
				SlackClientID:     mockClientID,
				SlackClientSecret: mockClientSecret,
				StaticRoot:        "../../../static",
				TemplatesRoot:     "../../templates",
			},
			errorContains: "templates parsing failed",
		},
		{
			config: Config{
				Logger:        zap.Must(zap.NewDevelopment()),
				StaticRoot:    "../../../static",
				TemplatesRoot: "../../templates",
			},
			errorContains: "Slack client credentials",
		},
	} {
		_, err := New(test.config)
		if assert.Error(t, err) {
			assert.ErrorContains(t, err, test.errorContains)
		}
	}

}

func TestRedirectURI(t *testing.T) {
	for _, test := range []struct {
		target   string
		relative string
		expected string
	}{
		{
			target:   "https://example.com/",
			relative: "/context/path?test=1",
			expected: "https://example.com/context/path",
		},
		{
			target:   "https://example.com/",
			relative: "/",
			expected: "https://example.com",
		},
		{
			target:   "https://example.com/",
			relative: "",
			expected: "https://example.com",
		},
	} {
		r := newRequestTester(http.MethodGet, test.target, nil)
		r.Request.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)

		assert.Equal(t, test.expected, redirectURI(r.Context, test.relative), "target: %s", test.target)
	}
}
