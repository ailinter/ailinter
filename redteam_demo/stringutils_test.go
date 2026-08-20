package stringutils

import (
	"strings"
	"testing"
)

func TestReverse(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"a", "a"},
		{"abc", "cba"},
		{"hello world", "dlrow olleh"},
		{"12345", "54321"},
	}
	for _, c := range cases {
		if got := Reverse(c.in); got != c.want {
			t.Errorf("Reverse(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestToUpper(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"hello", "HELLO"},
		{"Hello World", "HELLO WORLD"},
		{"abc123!@#", "ABC123!@#"},
		{"café", "CAFé"}, // é is non-ASCII, must stay unchanged
		{"abc DEF", "ABC DEF"},
	}
	for _, c := range cases {
		if got := ToUpper(c.in); got != c.want {
			t.Errorf("ToUpper(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestContainsAny(t *testing.T) {
	cases := []struct {
		name     string
		s, chars string
		want     bool
	}{
		{"ascii-present", "hello", "aeiou", true},
		{"ascii-absent", "hello", "xyz", false},
		{"single-rune", "hello", "h", true},
		{"empty-s", "", "a", false},
		{"empty-chars", "hello", "", false},
		{"both-empty", "", "", false},
		// Discriminating cases: a byte-wise revert would fail these.
		{"accent-distinct", "é", "è", false},
		{"accent-same", "é", "é", true},
		{"accent-word-absent", "café", "è", false},
		{"accent-word-present", "café", "é", true},
		{"cjk-distinct", "日", "本", false},
		{"cjk-present", "日本語", "本", true},
		{"emoji-distinct", "🙂", "🙃", false},
		{"emoji-same", "🙂", "🙂", true},
		{"invalid-utf8-no-match", "\xff", "a", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ContainsAny(c.s, c.chars); got != c.want {
				t.Errorf("ContainsAny(%q, %q) = %v, want %v", c.s, c.chars, got, c.want)
			}
		})
	}
}

// TestContainsAnyMatchesStrings is a differential check: ContainsAny must
// agree with strings.ContainsAny on a fixed corpus. Any divergence (e.g. a
// revert to naive byte-wise matching) fails CI.
func TestContainsAnyMatchesStrings(t *testing.T) {
	corpus := []struct{ s, chars string }{
		{"hello", "aeiou"},
		{"hello", "xyz"},
		{"", "abc"},
		{"abc", ""},
		{"", ""},
		{"café", "è"},
		{"café", "é"},
		{"日本語", "本"},
		{"🙂🙃", "🙂"},
		{"a\xffb", "\xff"},
	}
	for _, c := range corpus {
		want := strings.ContainsAny(c.s, c.chars)
		if got := ContainsAny(c.s, c.chars); got != want {
			t.Errorf("ContainsAny(%q, %q) = %v, strings.ContainsAny = %v", c.s, c.chars, got, want)
		}
	}
}

func TestMaxLen(t *testing.T) {
	cases := []struct {
		s    string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 3, "hel"},
		{"", 0, ""},
		{"abcdef", 0, ""},
	}
	for _, c := range cases {
		if got := MaxLen(c.s, c.max); got != c.want {
			t.Errorf("MaxLen(%q, %d) = %q, want %q", c.s, c.max, got, c.want)
		}
	}
}

func TestMaxLenPanicsOnNegative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MaxLen(\"x\", -1) did not panic")
		}
	}()
	MaxLen("x", -1)
}
