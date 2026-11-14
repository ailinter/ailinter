package analyzer_test

import (
	"embed"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

//go:embed testdata/js_nesting/*.txt
var jsNestingFixtures embed.FS

type jsNestingCase struct {
	fixture string
	want    bool
}

var jsNestingCases = []jsNestingCase{
	{"template.txt", false},
	{"nested.txt", true},
}

func TestAnalyze_JSDeepNesting(t *testing.T) {
	for _, tc := range jsNestingCases {
		t.Run(tc.fixture, func(t *testing.T) {
			data, err := jsNestingFixtures.ReadFile("testdata/js_nesting/" + tc.fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			result := analyzer.Analyze(
				analyzer.SourceInput{FilePath: tc.fixture, Source: string(data), Lang: "javascript"},
				analyzer.DefaultThresholds("javascript"),
			)
			if hasSmell(result, "deep_nesting") != tc.want {
				t.Errorf("deep_nesting present = %v, want %v", hasSmell(result, "deep_nesting"), tc.want)
			}
		})
	}
}
