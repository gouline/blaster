package server

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/gouline/blaster/internal/pkg/slack"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHandleAPIUnauthenticated(t *testing.T) {
	for _, test := range []struct {
		f func(r *requestTester) error
	}{
		{
			func(r *requestTester) error {
				return r.Server.handleAPISuggest(r.Context)
			},
		},
		{
			func(r *requestTester) error {
				return r.Server.handleAPISend(r.Context)
			},
		},
	} {
		r := newRequestTester(http.MethodGet, "/", nil)
		if assert.NoError(t, test.f(r)) {
			assert.Equal(t, http.StatusUnauthorized, r.Response.Code)
		}
	}
}

func TestHandleAPISuggest(t *testing.T) {
	r := newRequestTester(http.MethodGet, "/", nil)
	r.Authenticate("1", "")

	if assert.NoError(t, r.Server.handleAPISuggest(r.Context)) {
		assert.Equal(t, http.StatusOK, r.Response.Code)
	}
}

func TestHandleAPISuggestError(t *testing.T) {
	r := newRequestTester(http.MethodGet, "/", nil)
	r.Authenticate("1", "")
	r.Session.GetDestinationsError = errors.New("simulated")

	if assert.NoError(t, r.Server.handleAPISuggest(r.Context)) {
		assert.Equal(t, http.StatusInternalServerError, r.Response.Code)
		assert.Contains(t, r.Response.Body.String(), "simulated")
	}
}

func TestHandleAPISend(t *testing.T) {
	r := newRequestTester(http.MethodPost, "/", strings.NewReader("{\"user\":\"1\",\"message\":\"test\",\"as_user\":true}"))
	r.Request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	r.Authenticate("1", "")

	if assert.NoError(t, r.Server.handleAPISend(r.Context)) {
		assert.Equal(t, http.StatusOK, r.Response.Code)
	}
}

func TestHandleAPISendBindError(t *testing.T) {
	r := newRequestTester(http.MethodPost, "/", strings.NewReader("{\"user:\"1\",\"message\":\"test\",\"as_user\":true}"))
	r.Request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	r.Authenticate("1", "")

	if assert.NoError(t, r.Server.handleAPISend(r.Context)) {
		assert.Equal(t, http.StatusBadRequest, r.Response.Code)
	}
}

func TestHandleAPISendError(t *testing.T) {
	r := newRequestTester(http.MethodPost, "/", nil)
	r.Authenticate("1", "")
	r.Session.PostMessageError = errors.New("simulated")

	if assert.NoError(t, r.Server.handleAPISend(r.Context)) {
		assert.Equal(t, http.StatusInternalServerError, r.Response.Code)
		assert.Contains(t, r.Response.Body.String(), "simulated")
	}
}

func TestSanitizeSearchTerm(t *testing.T) {
	for _, test := range []struct {
		s        string
		expected string
	}{
		{
			s:        "test s!t@r#i$n^g [1,2)",
			expected: "test string 12",
		},
		{
			s:        "тестовая с!т@р#о$к^а (9.0]",
			expected: "тестовая строка 90",
		},
		{
			s:        "测!试@字#符$串%5^6",
			expected: "测试字符串56",
		},
	} {
		assert.Equal(t, test.expected, sanitizeSearchTerm(test.s))
	}
}

func TestSanitizeCSV(t *testing.T) {
	for _, test := range []struct {
		s        string
		expected string
	}{
		{
			s:        "",
			expected: "",
		},
		{
			s:        ",",
			expected: "",
		},
		{
			s:        ",,",
			expected: "",
		},
		{
			s:        "something, else",
			expected: "something else",
		},
	} {
		assert.Equal(t, test.expected, sanitizeCSV(test.s))
	}
}

func TestSuggestionLabel(t *testing.T) {
	for _, test := range []struct {
		name        string
		displayName string
		expected    string
	}{
		{
			name:        "",
			displayName: "",
			expected:    "",
		},
		{
			name:        "",
			displayName: "Mike",
			expected:    " (Mike)",
		},
		{
			name:        "mg",
			displayName: "Mike",
			expected:    "mg (Mike)",
		},
		{
			name:        "mg",
			displayName: "",
			expected:    "mg",
		},
	} {
		assert.Equal(t, test.expected, suggestionLabel(test.name, test.displayName), "name: %s", test.name)
	}
}

func TestSuggestDestinations(t *testing.T) {
	destinations := []*slack.Destination{
		{
			Type:        "user",
			Name:        "mg",
			DisplayName: "Mike",
			ID:          "mg",
		},
		{
			Type:        "user",
			Name:        "mark",
			DisplayName: "Mark",
			ID:          "mark",
		},
		{
			Type:        "user",
			Name:        "mk",
			DisplayName: "Mark Knopfler",
			ID:          "mk",
		},
		{
			Type:        "usergroup",
			Name:        "developers",
			DisplayName: "Developers",
			ID:          "ug",
			Children: []*slack.Destination{
				{
					Type:        "user",
					Name:        "jane",
					DisplayName: "Jane",
					ID:          "jd",
				},
			},
		},
	}

	for _, test := range []struct {
		term                string
		expectedIDs         []string
		expectedChildrenIDs []string
	}{
		{
			term:                "mi",
			expectedIDs:         []string{"mg"},
			expectedChildrenIDs: []string{},
		},
		{
			term:                "mark",
			expectedIDs:         []string{"mark", "mk"},
			expectedChildrenIDs: []string{},
		},
		{
			term:                "kno",
			expectedIDs:         []string{"mk"},
			expectedChildrenIDs: []string{},
		},
		{
			term:                "dev",
			expectedIDs:         []string{"ug"},
			expectedChildrenIDs: []string{"jd"},
		},
	} {
		actualValues := []string{}
		actualChildren := []string{}
		for _, actual := range suggestDestinations(test.term, destinations) {
			actualValues = append(actualValues, actual.Value)
			for _, child := range actual.Children {
				actualChildren = append(actualChildren, child.Value)
			}
		}
		assert.Equal(t, test.expectedIDs, actualValues, "term: %s:", test.term)
		assert.Equal(t, test.expectedChildrenIDs, actualChildren, "term: %s:", test.term)
	}
}
