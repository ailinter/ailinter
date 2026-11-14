package analyzer_test

import (
	"embed"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

//go:embed testdata/comments/*.txt
var commentFixtures embed.FS

func TestAnalyze_PythonDocstringComments(t *testing.T) {
	data, err := commentFixtures.ReadFile("testdata/comments/python_docstring.txt")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	result := analyzer.Analyze(
		analyzer.SourceInput{FilePath: "docstring.py", Source: string(data), Lang: "python"},
		analyzer.DefaultThresholds("python"),
	)
	if !hasSmell(result, "excessive_comments") {
		t.Fatalf("a long Python docstring should count as comments: %+v", result.Smells)
	}
}
