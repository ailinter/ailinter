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
func ContainsAny(s, chars string) bool {
	if len(s) == 0 || chars == "" {
		return false
	}
	// Build a set of the runes in chars: O(len(chars)).
	// Scanning s rune-by-rune makes matching O(len(s)) and Unicode-correct:
	// distinct runes sharing a byte (é U+00E9 vs è U+00E8) no longer match.
	// Invalid UTF-8 decodes to RuneError on both sides, mirroring
	// strings.ContainsAny without panicking.
	set := make(map[rune]struct{}, utf8.RuneCountInString(chars))
	for len(chars) > 0 {
		r, size := utf8.DecodeRuneInString(chars)
		set[r] = struct{}{}
		chars = chars[size:]
	}
	for _, r := range s {
		if _, ok := set[r]; ok {
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
