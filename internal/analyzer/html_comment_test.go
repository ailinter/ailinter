package analyzer_test

import (
	"embed"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

//go:embed testdata/comments/html_comments.txt
var htmlCommentFixture embed.FS

func TestAnalyze_HTMLCommentsStillAnalyzed(t *testing.T) {
	data, err := htmlCommentFixture.ReadFile("testdata/comments/html_comments.txt")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	result := analyzer.Analyze(
		analyzer.SourceInput{FilePath: "comments.html", Source: string(data), Lang: "html"},
		analyzer.DefaultThresholds("html"),
	)
	if !hasSmell(result, "excessive_comments") {
		t.Fatalf("HTML comments should be analyzed: %+v", result.Smells)
	}
}
