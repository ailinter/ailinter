package parser_test

import (
	"embed"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

//go:embed testdata/global_data/*.txt
var globalFixtures embed.FS

func readGlobalFixture(t *testing.T, name string) []string {
	t.Helper()
	data, err := globalFixtures.ReadFile("testdata/global_data/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return strings.Split(string(data), "\n")
}

func detectPython(lines []string) []parser.FunctionBloat {
	return parser.DetectFunctionBloatsIndent(lines, 4)
}

type globalDataCase struct {
	name    string
	fixture string
	detect  func([]string) []parser.FunctionBloat
	want    bool
}

func TestDetectGlobalData_LanguageScopes(t *testing.T) {
	cases := []globalDataCase{
		{"JSClassFieldsAreNotGlobals", "js_class.txt", parser.DetectFunctionBloatsTS, false},
		{"JSModuleStateIsGlobal", "js_globals.txt", parser.DetectFunctionBloatsTS, true},
		{"PythonGlobalsAfterFunction", "python_after.txt", detectPython, true},
		{"PythonGlobalsBeforeFunction", "python_before.txt", detectPython, true},
		{"PythonClassFieldsAreNotGlobals", "python_class.txt", detectPython, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertGlobalData(t, tc)
		})
	}
}

func assertGlobalData(t *testing.T, tc globalDataCase) {
	t.Helper()
	lines := readGlobalFixture(t, tc.fixture)
	smell := parser.DetectGlobalData(lines, tc.detect(lines), 5)
	if (smell != nil) != tc.want {
		t.Fatalf("global_data present = %v, want %v (%+v)", smell != nil, tc.want, smell)
	}
}
