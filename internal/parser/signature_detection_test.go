package parser_test

import (
	"embed"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

//go:embed testdata/signatures/*.txt
var signatureFixtures embed.FS

func signatureLines(t *testing.T, name string) []string {
	t.Helper()
	data, err := signatureFixtures.ReadFile("testdata/signatures/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return strings.Split(string(data), "\n")
}

type longParamCase struct {
	name    string
	lang    string
	fixture string
}

var longParamCases = []longParamCase{
	{"Java", "java", "java_long.txt"},
	{"Kotlin", "kotlin", "kotlin_long.txt"},
	{"Swift", "swift", "swift_long.txt"},
	{"Rust", "rust", "rust_long.txt"},
	{"JavaScript", "javascript", "js_long.txt"},
	{"TypeScriptMultiline", "typescript", "ts_multiline.txt"},
	{"Python", "python", "python_long.txt"},
	{"GoMethod", "go", "go_method.txt"},
}

func TestDetectLongParameterListFor(t *testing.T) {
	for _, tc := range longParamCases {
		t.Run(tc.name, func(t *testing.T) {
			smells := parser.DetectLongParameterListFor(tc.lang, signatureLines(t, tc.fixture), 4, 7)
			if len(smells) == 0 {
				t.Fatalf("expected long_parameter_list for %s", tc.name)
			}
		})
	}
}

type primitiveCase struct {
	name    string
	lang    string
	fixture string
	want    bool
}

var primitiveCases = []primitiveCase{
	{"Java", "java", "java_primitives.txt", true},
	{"TypeScript", "typescript", "ts_primitives.txt", true},
	{"JavaScriptUntyped", "javascript", "js_primitives.txt", false},
}

func TestDetectPrimitiveObsessionFor(t *testing.T) {
	for _, tc := range primitiveCases {
		t.Run(tc.name, func(t *testing.T) {
			smells := parser.DetectPrimitiveObsessionFor(tc.lang, signatureLines(t, tc.fixture))
			if (len(smells) > 0) != tc.want {
				t.Errorf("primitive_obsession present = %v, want %v", len(smells) > 0, tc.want)
			}
		})
	}
}
