package utils

import "regexp"

// NewAllSymbolsRegexp returns a compiled regular expression with
// all symbols on the keyboard available for filtering.
func NewAllSymbolsRegexp() *regexp.Regexp {
	reg, _ := regexp.Compile("[!-/:-@[-`{-~]+")
	return reg
}
