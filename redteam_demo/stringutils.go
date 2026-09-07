// Package stringutils provides simple string helper functions.
package stringutils

import "unicode/utf8"

// Reverse returns s with its characters in reverse order. It operates on
// runes, so multi-byte characters (é, 日, 🙂) are reversed whole and the
// result is always valid UTF-8 for valid-UTF-8 input. Invalid bytes decode
// to RuneError, mirroring the package's rune-based contract.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
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

// MaxLen returns s truncated to at most max bytes. If the byte budget cuts
// through a multi-byte rune, the result backs off to the nearest rune
// boundary, so the returned string is always valid UTF-8. s is returned
// unchanged when len(s) <= max. It panics if max is negative.
func MaxLen(s string, max int) string {
	if max < 0 {
		panic("stringutils.MaxLen: max must not be negative")
	}
	if len(s) <= max {
		return s
	}
	end := max
	for end > 0 && !utf8.RuneStart(s[end]) {
		end--
	}
	return s[:end]
}
