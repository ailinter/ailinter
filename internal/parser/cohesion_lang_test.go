package parser_test

import (
	"embed"
	"reflect"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

//go:embed testdata/cohesion/*.txt
var cohesionFixtures embed.FS

func readCohesionFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := cohesionFixtures.ReadFile("testdata/cohesion/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}

type cohesionCase struct {
	name    string
	lang    string
	fixture string
	low     bool
	min     float64
}

var cohesionDetectors = map[string]func([]string) []parser.FunctionBloat{
	"python": func(lines []string) []parser.FunctionBloat {
		return parser.DetectFunctionBloatsIndent(lines, 4)
	},
	"ruby": func(lines []string) []parser.FunctionBloat {
		return parser.DetectFunctionBloatsRuby(lines)
	},
}

func TestAnalyzeCohesionLang(t *testing.T) {
	cases := []cohesionCase{
		{"JavascriptClass", "javascript", "js_cohesive_class.txt", false, 1.0},
		{"JavascriptUnrelated", "javascript", "js_unrelated_functions.txt", true, 0},
		{"JavascriptSharedHelpers", "javascript", "js_shared_helpers.txt", false, 0.8},
		{"JavascriptMethodChain", "javascript", "js_method_chain.txt", false, 0.8},
		{"TypeScriptClass", "typescript", "js_cohesive_class.txt", false, 1.0},
		{"PythonClass", "python", "python_cohesive_class.txt", false, 1.0},
		{"PythonUnrelated", "python", "python_unrelated.txt", true, 0},
		{"RubyClass", "ruby", "ruby_cohesive_class.txt", false, 1.0},
		{"RubyUnrelated", "ruby", "ruby_unrelated.txt", true, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertCohesion(t, tc)
		})
	}
}

func assertCohesion(t *testing.T, tc cohesionCase) {
	t.Helper()
	detector := cohesionDetectors[tc.lang]
	if detector == nil {
		detector = parser.DetectFunctionBloatsTS
	}
	lines := strings.Split(readCohesionFixture(t, tc.fixture), "\n")
	result := parser.AnalyzeCohesionLang(tc.lang, detector(lines), lines)
	if result.IsLowCohesion != tc.low {
		t.Errorf("low cohesion = %v, want %v (%+v)", result.IsLowCohesion, tc.low, result)
	}
	if result.CohesionScore < tc.min {
		t.Errorf("cohesion %.2f is below %.2f", result.CohesionScore, tc.min)
	}
}

func TestAnalyzeCohesionLang_TypedFallback(t *testing.T) {
	goLines := strings.Split(readCohesionFixture(t, "go_shared_types.txt"), "\n")
	goBloats := parser.DetectFunctionBloats(goLines)
	wantGo := parser.AnalyzeCohesion(goBloats, goLines)
	if got := parser.AnalyzeCohesionLang("go", goBloats, goLines); !reflect.DeepEqual(got, wantGo) {
		t.Errorf("go must keep the type signal: got %+v, want %+v", got, wantGo)
	}
	if wantGo.IsLowCohesion {
		t.Errorf("shared *Data should stay cohesive, got %+v", wantGo)
	}

	jsLines := strings.Split(readCohesionFixture(t, "js_cohesive_class.txt"), "\n")
	jsBloats := parser.DetectFunctionBloatsTS(jsLines)
	wantUnknown := parser.AnalyzeCohesion(jsBloats, jsLines)
	if got := parser.AnalyzeCohesionLang("", jsBloats, jsLines); !reflect.DeepEqual(got, wantUnknown) {
		t.Errorf("unknown language must fall back: got %+v, want %+v", got, wantUnknown)
	}
}
