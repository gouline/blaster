package utils

import (
	"regexp"

	"github.com/gin-gonic/gin"
)

// NewAllSymbolsRegexp returns a compiled regular expression with
// all symbols on the keyboard available for filtering.
func NewAllSymbolsRegexp() *regexp.Regexp {
	reg, _ := regexp.Compile("[!-/:-@[-`{-~]+")
	return reg
}

// RelativeURI returns a URI relative to request host.
func RelativeURI(c *gin.Context, path string) string {
	return "http://" + c.Request.Host + path
}
