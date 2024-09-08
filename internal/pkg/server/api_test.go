package server

import (
	"slices"
	"testing"

	"github.com/gouline/blaster/internal/pkg/slack"
)

func Test_sanitizeSearchTerm(t *testing.T) {
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
		actual := sanitizeSearchTerm(test.s)
		if actual != test.expected {
			t.Errorf("for %s: got %s, expected %s", test.s, actual, test.expected)
		}
	}
}

func Test_sanitizeCSV(t *testing.T) {
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
		actual := sanitizeCSV(test.s)
		if actual != test.expected {
			t.Errorf("for %s: got %s, expected %s", test.s, actual, test.expected)
		}
	}
}

func Test_suggestionLabel(t *testing.T) {
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
		actual := suggestionLabel(test.name, test.displayName)
		if actual != test.expected {
			t.Errorf("for %s, %s: got %s, expected %s", test.name, test.displayName, actual, test.expected)
		}
	}
}

func Test_suggestDestinations(t *testing.T) {
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
	}

	for _, test := range []struct {
		term        string
		expectedIDs []string
	}{
		{
			term:        "mi",
			expectedIDs: []string{"mg"},
		},
		{
			term:        "mark",
			expectedIDs: []string{"mark", "mk"},
		},
		{
			term:        "kno",
			expectedIDs: []string{"mk"},
		},
	} {
		actuals := suggestDestinations(test.term, destinations)
		actualValues := []string{}
		for _, actual := range actuals {
			actualValues = append(actualValues, actual.Value)
		}

		if !slices.Equal(actualValues, test.expectedIDs) {
			t.Errorf("for %s: got %s, expected %s", test.term, actualValues, test.expectedIDs)
		}
	}
}
