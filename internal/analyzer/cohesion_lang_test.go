package analyzer_test

import (
	"embed"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
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

func hasSmell(result analyzer.QualityResult, name string) bool {
	for _, smell := range result.Smells {
		if smell.Name == name {
			return true
		}
	}
	return false
}

func TestAnalyze_JavascriptClassDoesNotGetLowCohesion(t *testing.T) {
	result := analyzer.Analyze(
		analyzer.SourceInput{
			FilePath: "counter.js",
			Source:   readCohesionFixture(t, "js_cohesive_class.txt"),
			Lang:     "javascript",
		},
		analyzer.DefaultThresholds("javascript"),
	)

	if hasSmell(result, "low_cohesion") {
		t.Fatalf("a class sharing this.value must not be low cohesion: %+v", result.Smells)
	}
	if result.Score < 95 {
		t.Errorf("cohesive JavaScript should score at least 95, got %d", result.Score)
	}
}

func TestAnalyze_JavascriptUnrelatedFunctionsStillGetLowCohesion(t *testing.T) {
	result := analyzer.Analyze(
		analyzer.SourceInput{
			FilePath: "helpers.js",
			Source:   readCohesionFixture(t, "js_unrelated_functions.txt"),
			Lang:     "javascript",
		},
		analyzer.DefaultThresholds("javascript"),
	)

	if !hasSmell(result, "low_cohesion") {
		t.Fatalf("unrelated helpers should still be flagged: %+v", result.Smells)
	}
}

func TestAnalyze_TypeScriptClassDoesNotGetLowCohesion(t *testing.T) {
	result := analyzer.Analyze(
		analyzer.SourceInput{
			FilePath: "counter.ts",
			Source:   readCohesionFixture(t, "js_cohesive_class.txt"),
			Lang:     "typescript",
		},
		analyzer.DefaultThresholds("typescript"),
	)

	if hasSmell(result, "low_cohesion") {
		t.Fatalf("TypeScript uses the same symbol signal: %+v", result.Smells)
	}
}

func TestAnalyze_GoCohesionUnchanged(t *testing.T) {
	result := analyzer.Analyze(
		analyzer.SourceInput{
			FilePath: "data.go",
			Source:   readCohesionFixture(t, "go_shared_types.txt"),
			Lang:     "go",
		},
		analyzer.DefaultThresholds("go"),
	)

	if hasSmell(result, "low_cohesion") {
		t.Fatalf("shared *Data should keep Go cohesion high: %+v", result.Smells)
	}
}
