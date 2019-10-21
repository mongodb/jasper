package testutil

import (
	"strings"
	"unicode"
)

// RemoveWhitespace returns the string with all whitespace removed.
func RemoveWhitespace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}
