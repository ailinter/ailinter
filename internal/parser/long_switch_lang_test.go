package parser_test

import (
	"embed"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

//go:embed testdata/switch/*.txt
var switchFixtures embed.FS

func switchLines(t *testing.T, name string) []string {
	t.Helper()
	data, err := switchFixtures.ReadFile("testdata/switch/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return strings.Split(string(data), "\n")
}

type switchCase struct {
	lang    string
	fixture string
}

var switchCases = []switchCase{
	{"kotlin", "kotlin_when.txt"},
	{"rust", "rust_match.txt"},
	{"ruby", "ruby_case.txt"},
}

func TestDetectLongSwitchFor(t *testing.T) {
	for _, tc := range switchCases {
		t.Run(tc.fixture, func(t *testing.T) {
			smells := parser.DetectLongSwitchFor(tc.lang, switchLines(t, tc.fixture), 10, 20)
			if len(smells) == 0 {
				t.Fatalf("expected long_switch for %s", tc.fixture)
			}
		})
	}
}
