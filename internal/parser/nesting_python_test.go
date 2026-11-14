package parser_test

import (
	"embed"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

//go:embed testdata/nesting/*.txt
var nestingFixtures embed.FS

func nestingFixture(t *testing.T, name string) []string {
	t.Helper()
	data, err := nestingFixtures.ReadFile("testdata/nesting/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return strings.Split(string(data), "\n")
}

func TestAnalyzeNestingIndent_Python(t *testing.T) {
	result := parser.AnalyzeNestingIndent(nestingFixture(t, "python_depth_5.txt"))
	if result.MaxDepth < 5 {
		t.Errorf("expected Python depth at least 5, got %+v", result)
	}
}

func TestDetectBumpyRoadForFunctionIndent_Python(t *testing.T) {
	result := parser.DetectBumpyRoadForFunctionIndent(nestingFixture(t, "python_bumpy.txt"), 2, 2)
	if !result.IsBumpyRoad {
		t.Fatalf("expected a bumpy road, got %+v", result)
	}
}
