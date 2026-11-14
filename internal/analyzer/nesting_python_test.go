package analyzer_test

import (
	"embed"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

//go:embed testdata/nesting/*.txt
var nestingFixtures embed.FS

func nestingSource(t *testing.T, name string) string {
	t.Helper()
	data, err := nestingFixtures.ReadFile("testdata/nesting/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}

func TestAnalyze_PythonDeepNesting(t *testing.T) {
	result := analyzer.Analyze(
		analyzer.SourceInput{FilePath: "nested.py", Source: nestingSource(t, "python_depth_5.txt"), Lang: "python"},
		analyzer.DefaultThresholds("python"),
	)
	if !hasSmell(result, "deep_nesting") {
		t.Fatalf("Python indentation nesting should be detected: %+v", result.Smells)
	}
}

func TestAnalyze_PythonBumpyRoad(t *testing.T) {
	result := analyzer.Analyze(
		analyzer.SourceInput{FilePath: "bumpy.py", Source: nestingSource(t, "python_bumpy.txt"), Lang: "python"},
		analyzer.DefaultThresholds("python"),
	)
	if !hasSmell(result, "bumpy_road") {
		t.Fatalf("Python indentation bumps should be detected: %+v", result.Smells)
	}
}

func TestAnalyze_GoBraceNestingUnchanged(t *testing.T) {
	result := analyzer.Analyze(
		analyzer.SourceInput{FilePath: "deep.go", Source: nestingSource(t, "go_depth_5.txt"), Lang: "go"},
		analyzer.DefaultThresholds("go"),
	)
	if !hasSmell(result, "deep_nesting") {
		t.Fatalf("Go brace nesting must keep working: %+v", result.Smells)
	}
}
