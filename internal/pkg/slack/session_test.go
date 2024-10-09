package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashToken(t *testing.T) {
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
		actual := (&ClientSession{Token: test.a}).tokenHash() == (&ClientSession{Token: test.b}).tokenHash()
		assert.Equal(t, test.expected, actual, "% ?= %s", test.a, test.b)
	}
}

func TestMarshalUnmarshalNormal(t *testing.T) {
	original := &ClientSession{Token: "123", Team: "abc"}
	data := original.Marshal()
	recreated := &ClientSession{}
	recreated.Unmarshal(data)
	assert.Equal(t, original, recreated)
	assert.Equal(t, original.TeamName(), recreated.TeamName())
	assert.Equal(t, original.IsAuthenticated(), recreated.IsAuthenticated())
}

func TestMarshalEmpty(t *testing.T) {
	original := &ClientSession{}
	data := original.Marshal()
	recreated := &ClientSession{}
	recreated.Unmarshal(data)
	assert.Equal(t, original, recreated)
	assert.Equal(t, original.TeamName(), recreated.TeamName())
	assert.Equal(t, original.IsAuthenticated(), recreated.IsAuthenticated())
}
