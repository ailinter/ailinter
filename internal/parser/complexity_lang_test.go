package parser_test

import (
	"embed"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

//go:embed testdata/complexity/*.txt
var complexityFixtures embed.FS

func complexityLines(t *testing.T, name string) []string {
	t.Helper()
	data, err := complexityFixtures.ReadFile("testdata/complexity/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return strings.Split(string(data), "\n")
}

type conditionalCase struct {
	name    string
	lang    string
	fixture string
	want    bool
}

var conditionalCases = []conditionalCase{
	{"PythonAnd", "python", "python_complex.txt", true},
	{"GoAnd", "go", "go_complex.txt", true},
	{"PythonClean", "python", "python_clean.txt", false},
}

func TestDetectComplexConditionalFor(t *testing.T) {
	for _, tc := range conditionalCases {
		t.Run(tc.name, func(t *testing.T) {
			smells := parser.DetectComplexConditionalFor(tc.lang, complexityLines(t, tc.fixture), 2, 10)
			if (len(smells) > 0) != tc.want {
				t.Errorf("complex_conditional present = %v, want %v", len(smells) > 0, tc.want)
			}
		})
	}
}
