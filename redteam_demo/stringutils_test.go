package stringutils

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestReverse(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"a", "a"},
		{"abc", "cba"},
		{"hello world", "dlrow olleh"},
		{"12345", "54321"},
		// Multi-byte: runes must reverse whole, result stays valid UTF-8.
		{"café", "éfac"},
		{"日本語", "語本日"},
		{"🙂ok", "ko🙂"},
		{"aé日", "日éa"},
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

// containsAnyCase is one row of the ContainsAny corpus: a substring s, a
// chars set, and the membership result ContainsAny must report.
type containsAnyCase struct {
	name     string
	s, chars string
	want     bool
}

// containsAnyCases returns the rune-matching corpus exercised by
// TestContainsAny and assertContainsAny.
func containsAnyCases() []containsAnyCase {
	// Corpus for TestContainsAny: ContainsAny must agree with
	// strings.ContainsAny on every row, including byte-sharing runes.
	return []containsAnyCase{
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
}

func TestContainsAny(t *testing.T) {
	// Run every ContainsAny corpus row as its own subtest.
	for _, c := range containsAnyCases() {
		t.Run(c.name, func(t *testing.T) {
			assertContainsAny(t, c)
		})
	}
}

// assertContainsAny reports a subtest failure when the row mismatches.
func assertContainsAny(t *testing.T, c containsAnyCase) {
	// Spot-check the hardened path: ContainsAny must equal c.want for this row.
	if got := ContainsAny(c.s, c.chars); got != c.want {
		t.Errorf("ContainsAny(%q, %q) = %v, want %v", c.s, c.chars, got, c.want)
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

// maxLenCase is one row of the MaxLen corpus: an input string, a byte budget,
// and the truncated result MaxLen must return.
type maxLenCase struct {
	s    string
	max  int
	want string
}

// maxLenCases returns the truncation corpus for TestMaxLen.
func maxLenCases() []maxLenCase {
	// Rows exercised by TestMaxLen and TestMaxLenPanicsOnNegative. MaxLen
	// must truncate each row to its byte budget.
	return []maxLenCase{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 3, "hel"},
		{"", 0, ""},
		{"abcdef", 0, ""},
		// Byte budgets that cut through multi-byte runes must back off to a
		// rune boundary — the result is always valid UTF-8.
		{"日本語", 9, "日本語"}, // len == 9, unchanged
		{"日本語", 8, "日本"},  // 語 is 3 bytes; 8 cuts it → back to 6
		{"日本語", 3, "日"},   // boundary at 3
		{"日本語", 2, ""},    // 2 < 日(3B) → nothing fits
		{"aé", 2, "a"},    // é is 2 bytes (0xc3 0xa9); 2 cuts it → back to 1
	}
}

func TestMaxLen(t *testing.T) {
	// MaxLen must truncate every corpus row to its byte budget.
	for _, c := range maxLenCases() {
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

// TestMaxLenReturnsValidUTF8 pins the rune-boundary guarantee: for any input
// and byte budget, the result must be valid UTF-8 when the input is.
func TestMaxLenReturnsValidUTF8(t *testing.T) {
	inputs := []string{"", "hello", "café", "日本語", "🙂ok", "a\xc3\xa9b日"}
	for _, s := range inputs {
		for max := 0; max <= len(s)+1; max++ {
			got := MaxLen(s, max)
			if !utf8.ValidString(got) {
				t.Errorf("MaxLen(%q, %d) = %q — invalid UTF-8", s, max, got)
			}
			if len(got) > max {
				t.Errorf("MaxLen(%q, %d) = %q — %d bytes > budget", s, max, got, len(got))
			}
		}
	}
}
