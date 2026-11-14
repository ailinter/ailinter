package analyzer_test

import (
	"embed"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

//go:embed testdata/signatures/*.txt
var signatureFixtures embed.FS

func signatureSource(t *testing.T, name string) string {
	t.Helper()
	data, err := signatureFixtures.ReadFile("testdata/signatures/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}

type signatureCase struct {
	name    string
	lang    string
	fixture string
	smell   string
	want    bool
}

var signatureCases = []signatureCase{
	{"JavaLong", "java", "java_long.txt", "long_parameter_list", true},
	{"JavaScriptLong", "javascript", "js_long.txt", "long_parameter_list", true},
	{"TypeScriptMultiline", "typescript", "ts_multiline.txt", "long_parameter_list", true},
	{"JavaPrimitives", "java", "java_primitives.txt", "primitive_obsession", true},
	{"JavaScriptPrimitivesSkipped", "javascript", "js_primitives.txt", "primitive_obsession", false},
}

func TestAnalyze_SignatureSmellsByLanguage(t *testing.T) {
	for _, tc := range signatureCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.Analyze(
				analyzer.SourceInput{FilePath: tc.fixture, Source: signatureSource(t, tc.fixture), Lang: tc.lang},
				analyzer.DefaultThresholds(tc.lang),
			)
			if hasSmell(result, tc.smell) != tc.want {
				t.Errorf("%s present = %v, want %v", tc.smell, hasSmell(result, tc.smell), tc.want)
			}
		})
	}
}
