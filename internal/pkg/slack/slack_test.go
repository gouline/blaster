package slack

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func Test_redirectURI(t *testing.T) {
	e := echo.New()

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
		req := httptest.NewRequest(http.MethodGet, test.target, nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMETextPlain)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		actual := redirectURI(c, test.relative)
		if actual != test.expected {
			t.Errorf("for %s, %s: got %s, expected %s", test.target, test.relative, actual, test.expected)
		}
	}
}
