package stringutils

import (
	"strings"
	"testing"
	"time"
)

// TestContainsAnyLinear guards against accidental quadratic behavior in
// ContainsAny: a 1MB x 1MB no-match call must complete in under 1 second.
// On the reference sandbox this runs in ~60ms, so 1s leaves ~16x headroom
// while still failing loudly on any O(n*m) or per-byte-scan regression.
func TestContainsAnyLinear(t *testing.T) {
	s := strings.Repeat("a", 1<<20)     // 1MiB of 'a'
	chars := strings.Repeat("b", 1<<20) // 1MiB of 'b'
	start := time.Now()
	if ContainsAny(s, chars) {
		t.Fatal("ContainsAny reported a match for disjoint inputs")
	}
	if elapsed := time.Since(start); elapsed >= time.Second {
		t.Fatalf("ContainsAny(1MB, 1MB) took %v, want < 1s", elapsed)
	}
}
