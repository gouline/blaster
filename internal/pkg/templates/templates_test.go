package templates

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	templates, err := New(Config{
		Logger:     zap.Must(zap.NewDevelopment()),
		RootPath:   "examples",
		LayoutFile: "layout.html",
	})
	if !assert.NoError(t, err) {
		return
	}

	for _, test := range []struct {
		name     string
		expected bool
	}{
		{
			name:     "README.md",
			expected: false,
		},
		{
			name:     "about.html",
			expected: true,
		},
		{
			name:     "home.html",
			expected: true,
		},
	} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)

		if err := templates.Render(rec, test.name, map[string]interface{}{}, c); test.expected {
			assert.NoError(t, err, "template: %s", test.name)
		} else {
			assert.Error(t, err, "template: %s", test.name)
		}
	}
}

func TestNewChecks(t *testing.T) {
	logger := zap.Must(zap.NewDevelopment())
	var err error

	_, err = New(Config{
		Logger:     logger,
		RootPath:   "examples1",
		LayoutFile: "layout.html",
	})
	assert.ErrorContains(t, err, "root not found")

	_, err = New(Config{
		Logger:     logger,
		RootPath:   "examples/about.html",
		LayoutFile: "layout.html",
	})
	assert.ErrorContains(t, err, "root not directory")

	_, err = New(Config{
		Logger:     logger,
		RootPath:   "examples",
		LayoutFile: "layout.html1",
	})
	assert.ErrorContains(t, err, "layout not found")

	_, err = New(Config{
		Logger:     logger,
		RootPath:   "examples",
		LayoutFile: ".",
	})
	assert.ErrorContains(t, err, "layout is directory")
}
