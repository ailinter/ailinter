package parser_test

import (
	"embed"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

//go:embed testdata/js_nesting/*.txt
var jsNestingFixtures embed.FS

func jsNestingLines(t *testing.T, name string) []string {
	t.Helper()
	data, err := jsNestingFixtures.ReadFile("testdata/js_nesting/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return strings.Split(string(data), "\n")
}

func TestAnalyzeNestingJS_IgnoresTemplateLiterals(t *testing.T) {
	lines := jsNestingLines(t, "template.txt")
	raw := parser.AnalyzeNesting(lines)
	clean := parser.AnalyzeNestingJS(lines)
	if clean.MaxDepth >= raw.MaxDepth {
		t.Errorf("template braces should not add depth: raw=%d clean=%d", raw.MaxDepth, clean.MaxDepth)
	}
	if clean.MaxDepth >= 3 {
		t.Errorf("a template literal should stay shallow, got %d", clean.MaxDepth)
	}
}
