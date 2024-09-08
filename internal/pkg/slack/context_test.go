package slack

import "testing"

func Test_hashToken(t *testing.T) {
	for _, test := range []struct {
		a        string
		b        string
		expected bool
	}{
		{
			a:        "",
			b:        "",
			expected: true,
		},
		{
			a:        "392n784y9238",
			b:        "392n784y9238",
			expected: true,
		},
		{
			a:        "392n784y9238",
			b:        "392n784y923",
			expected: false,
		},
		{
			a:        "392n784y9238",
			b:        "",
			expected: false,
		},
	} {
		actual := hashToken(test.a) == hashToken(test.b)
		if actual != test.expected {
			t.Errorf("for %s, %s: got %t, expected %t", test.a, test.b, actual, test.expected)
		}
	}
}
