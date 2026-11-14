package analyzer_test

import (
	"embed"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

//go:embed testdata/function_bodies/*.txt
var functionBodyFixtures embed.FS

func TestAnalyze_CSharpExpressionBodiedCounted(t *testing.T) {
	data, err := functionBodyFixtures.ReadFile("testdata/function_bodies/csharp_expression.txt")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	result := analyzer.Analyze(
		analyzer.SourceInput{FilePath: "expr.cs", Source: string(data), Lang: "csharp"},
		analyzer.DefaultThresholds("csharp"),
	)
	if !hasSmell(result, "function_count") {
		t.Fatalf("expression-bodied methods should count toward function_count: %+v", result.Smells)
	}
	if hasSmell(result, "lazy_element") {
		t.Fatalf("expression-bodied methods must not be lazy elements: %+v", result.Smells)
	}
}
