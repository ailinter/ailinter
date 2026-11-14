package analyzer_test

import (
	"embed"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

//go:embed testdata/complexity/*.txt
var complexityFixtures embed.FS

func TestAnalyze_PythonComplexConditional(t *testing.T) {
	data, err := complexityFixtures.ReadFile("testdata/complexity/python_complex.txt")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	result := analyzer.Analyze(
		analyzer.SourceInput{FilePath: "complex.py", Source: string(data), Lang: "python"},
		analyzer.DefaultThresholds("python"),
	)
	if !hasSmell(result, "complex_conditional") {
		t.Fatalf("Python and/or condition should be detected: %+v", result.Smells)
	}
}

func TestAnalyze_PythonCyclomaticComplexity(t *testing.T) {
	data, err := complexityFixtures.ReadFile("testdata/complexity/python_branchy.txt")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	result := analyzer.Analyze(
		analyzer.SourceInput{FilePath: "branchy.py", Source: string(data), Lang: "python"},
		analyzer.DefaultThresholds("python"),
	)
	if !hasSmell(result, "complex_method") {
		t.Fatalf("Python and/or branches should count toward CC: %+v", result.Smells)
	}
}
