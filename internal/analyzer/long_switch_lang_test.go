package analyzer_test

import (
	"embed"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

//go:embed testdata/switch/*.txt
var switchFixtures embed.FS

type switchLangCase struct {
	lang    string
	fixture string
}

var switchLangCases = []switchLangCase{
	{"kotlin", "kotlin_when.txt"},
	{"rust", "rust_match.txt"},
	{"ruby", "ruby_case.txt"},
}

func TestAnalyze_LongSwitchLanguages(t *testing.T) {
	for _, tc := range switchLangCases {
		t.Run(tc.fixture, func(t *testing.T) {
			data, err := switchFixtures.ReadFile("testdata/switch/" + tc.fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			result := analyzer.Analyze(
				analyzer.SourceInput{FilePath: tc.fixture, Source: string(data), Lang: tc.lang},
				analyzer.DefaultThresholds(tc.lang),
			)
			if !hasSmell(result, "long_switch") {
				t.Fatalf("%s should report long_switch: %+v", tc.fixture, result.Smells)
			}
		})
	}
}
