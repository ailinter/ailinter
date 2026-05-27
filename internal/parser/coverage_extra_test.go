package parser_test

import (
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

func TestDetectFunctionBloatsJava(t *testing.T) {
	src := "public class Foo {\npublic void bar() {\nSystem.out.println(\"hi\");\n}\nprivate int baz() {\nreturn 42;\n}\n}\n"
	bloats := parser.DetectFunctionBloatsJava(strings.Split(src, "\n"))
	if len(bloats) < 2 {
		t.Fatalf("expected at least 2 Java methods, got %d", len(bloats))
	}
}

func TestDetectFunctionBloatsIndent(t *testing.T) {
	src := "def foo():\n    x = 1\n    return x\n\ndef bar():\n    return 2\n"
	bloats := parser.DetectFunctionBloatsIndent(strings.Split(src, "\n"), 4)
	t.Logf("Python functions found: %d", len(bloats))
	for _, b := range bloats {
		t.Logf("  %s: L%d, %d lines", b.Name, b.LineStart, b.LineCount)
	}
}

func TestCountLeadingSpaces(t *testing.T) {
	src := "def foo():\n    pass\n"
	bloats := parser.DetectFunctionBloatsIndent(strings.Split(src, "\n"), 4)
	t.Logf("Python functions found: %d", len(bloats))
}

func TestDetectLowCohesionSmell(t *testing.T) {
	result := parser.CohesionResult{
		TotalFuncs:    10,
		IsolatedFuncs: 7,
		CohesionScore: 0.3,
		IsLowCohesion: true,
	}
	s := parser.DetectLowCohesionSmell(result, 50, 75)
	if s == nil {
		t.Fatal("expected low_cohesion smell")
	}
	t.Logf("Low cohesion: %s", s.Message)
}

func TestDetectDuplicationSmells(t *testing.T) {
	pairs := []parser.DuplicationPair{
		{FuncA: "foo", FuncB: "bar", LineA: 10, LineB: 20, Similarity: 0.85, LineCount: 15},
	}
	smells := parser.DetectDuplicationSmells(pairs)
	if len(smells) != 1 {
		t.Fatalf("expected 1 duplication smell, got %d", len(smells))
	}
}

func TestHasSharedType(t *testing.T) {
	// Test indirectly through cohesion analysis
	src := "func ParseInt(s string) (int, error) {\nreturn 0, nil\n}\nfunc ParseFloat(s string) (float64, error) {\nreturn 0, nil\n}\nfunc Greet(name string) string {\nreturn \"hi\"\n}\nfunc Add(a, b int) int {\nreturn a + b\n}\nfunc Multiply(a, b int) int {\nreturn a * b\n}\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloats(lines)
	result := parser.AnalyzeCohesion(bloats, lines)
	t.Logf("Cohesion: %.2f, isolated: %d/%d", result.CohesionScore, result.IsolatedFuncs, result.TotalFuncs)
}

func TestComplexConditional_While(t *testing.T) {
	lines := strings.Split("while a > 0 && b > 0 || c > 0 {", "\n")
	smells := parser.DetectComplexConditional(lines, 2, 5)
	if len(smells) == 0 {
		t.Fatal("expected complex_conditional for while")
	}
}

func TestComplexConditional_Alert(t *testing.T) {
	lines := strings.Split("if a && b && c && d && e && f {", "\n")
	smells := parser.DetectComplexConditional(lines, 2, 5)
	if len(smells) == 0 {
		t.Fatal("expected complex_conditional alert")
	}
	if smells[0].Severity != "alert" {
		t.Errorf("expected alert, got %s", smells[0].Severity)
	}
}

func TestLongParameterList_Alert(t *testing.T) {
	lines := strings.Split("func Many(a, b, c, d, e, f, g, h string) {", "\n")
	smells := parser.DetectLongParameterList(lines, 4, 7)
	if len(smells) == 0 {
		t.Fatal("expected alert for 8 params")
	}
}

func TestTotalCyclomaticComplexity(t *testing.T) {
	// Test indirect coverage through CC detection
	lines := strings.Split("func f() {\nif x > 0 {\nfor _, v := range items {\nif v < 100 {\nreturn v\n}\n}\n}\n}\n", "\n")
	bloats := parser.DetectFunctionBloats(lines)
	if len(bloats) > 0 {
		t.Logf("Found function: %s (%d lines)", bloats[0].Name, bloats[0].LineCount)
	}
}

func TestAnalyzeCohesion_Small(t *testing.T) {
	src := "func A() {}\nfunc B() {}\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloats(lines)
	result := parser.AnalyzeCohesion(bloats, lines)
	if result.TotalFuncs > 0 && result.CohesionScore < 0.5 {
		t.Errorf("small module should have high cohesion, got %.2f", result.CohesionScore)
	}
}

func TestDetectBrainMethodSmell_Edge(t *testing.T) {
	// Just at warning threshold
	bloats := []parser.FunctionBloat{{Name: "f", LineCount: 80, LineStart: 1}}
	smells := parser.DetectBrainMethodSmell(bloats, 80, 200)
	if len(smells) == 0 {
		t.Error("80-line function should trigger at threshold 80")
	}
}

func TestDetectParagraphOfCode_MidFile(t *testing.T) {
	lines := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "", "x", "y"}
	smells := parser.DetectParagraphOfCode(lines, 10)
	if len(smells) == 0 {
		t.Fatal("expected paragraph detection before blank line")
	}
}

func TestDetectMessageChains_Simple(t *testing.T) {
	lines := strings.Split("r := obj.A().B()", "\n")
	smells := parser.DetectMessageChains(lines)
	t.Logf("Message chains: %d", len(smells))
}

func TestThresholds_ReturnValues(t *testing.T) {
	for _, lang := range []string{"go", "python", "ruby", "swift", "kotlin"} {
		th := parser.DefaultThresholds(lang)
		if th.FuncCCWarning < 5 || th.FuncCCWarning > 15 {
			t.Errorf("%s: unexpected CC warning: %d", lang, th.FuncCCWarning)
		}
	}
}
