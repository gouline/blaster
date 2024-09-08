package templates

import (
	"net/http/httptest"
	"testing"
)

func Test_New(t *testing.T) {
	templates, err := New("examples", "layout.html")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
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
		tmpl, ok := templates.templates[test.name]
		if ok != test.expected {
			if test.expected {
				t.Errorf("template %s expected but not found", test.name)
			} else {
				t.Errorf("template %s not expected but found", test.name)
			}
		}
		if !ok {
			continue
		}

		rec := httptest.NewRecorder()
		err := tmpl.Execute(rec, map[string]interface{}{})
		if err != nil {
			t.Errorf("template %s error: %s", test.name, err)
		}
	}
}
