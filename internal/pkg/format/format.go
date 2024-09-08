package format

import (
	"crypto/sha1"
	"fmt"
	"regexp"

	"github.com/labstack/echo/v4"
)

// NewAllSymbolsRegexp returns a compiled regular expression with
// all symbols on the keyboard available for filtering.
func NewAllSymbolsRegexp() *regexp.Regexp {
	return regexp.MustCompile("[!-/:-@[-`{-~]+")
}

// RelativeURI returns a URI relative to request host.
func RelativeURI(c echo.Context, path string) string {
	return fmt.Sprintf("%s://%s%s", c.Scheme(), c.Request().Host, path)
}

// HashToken hashes raw auth token with SHA-1.
func HashToken(token string) string {
	h := sha1.New()
	h.Write([]byte(token))
	return fmt.Sprintf("%x", h.Sum(nil))
}
