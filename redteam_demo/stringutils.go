// Package stringutils provides simple string helper functions.
package stringutils

import "unicode/utf8"

// Reverse returns s with its bytes in reverse order.
func Reverse(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

// ToUpper returns s with ASCII letters uppercased; non-ASCII bytes are left unchanged.
func ToUpper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if 'a' <= c && c <= 'z' {
			b[i] = c - 'a' + 'A'
		}
	}
	return string(b)
}

// ContainsAny reports whether s contains any character present in chars.
//
// It builds a set of the runes in chars (O(len(chars))) and scans s
// rune-by-rune (O(len(s))), so distinct runes that share bytes (é U+00E9 vs
// è U+00E8) never match. Invalid UTF-8 decodes to RuneError on both sides,
// mirroring strings.ContainsAny without panicking.
func ContainsAny(s, chars string) bool {
	if len(s) == 0 || chars == "" {
		return false
	}
	set := make(map[rune]bool, utf8.RuneCountInString(chars))
	for _, r := range chars {
		set[r] = true
	}
	for _, r := range s {
		if set[r] {
			return true
		}
	}
	return false
}

// MaxLen returns s truncated to at most max bytes, or s unchanged if len(s) <= max.
// It panics if max is negative.
func MaxLen(s string, max int) string {
	if max < 0 {
		panic("stringutils.MaxLen: max must not be negative")
	}
	if len(s) <= max {
		return s
	}
	return s[:max]
}
