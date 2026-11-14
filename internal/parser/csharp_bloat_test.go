package parser_test

import (
	"embed"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

//go:embed testdata/function_bodies/*.txt
var functionBodyFixtures embed.FS

func functionBodyLines(t *testing.T, name string) []string {
	t.Helper()
	data, err := functionBodyFixtures.ReadFile("testdata/function_bodies/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return strings.Split(string(data), "\n")
}

func TestDetectFunctionBloatsCSharp_ExpressionBodied(t *testing.T) {
	bloats := parser.DetectFunctionBloatsCSharp(functionBodyLines(t, "csharp_expression.txt"))
	count := 0
	for _, b := range bloats {
		if b.ExpressionBody {
			count++
		}
	}
	if count != 26 {
		t.Fatalf("expected 26 expression-bodied methods, got %d (%+v)", count, bloats)
	}
}
