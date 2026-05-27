package parser_test

import (
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

func TestDuplication_Identical(t *testing.T) {
	src := "func ProcessA(data []int) int {\ns := 0\nfor _, v := range data {\nif v > 0 {\ns += v\ns *= 2\ns -= 1\n}\n}\nreturn s\n}\nfunc ProcessB(data []int) int {\ns := 0\nfor _, v := range data {\nif v > 0 {\ns += v\ns *= 2\ns -= 1\n}\n}\nreturn s\n}\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloats(lines)
	pairs := parser.DetectDuplications(bloats, lines, 10, 0.75)
	if len(pairs) == 0 {
		t.Fatal("expected duplication pair for identical functions")
	}
	t.Logf("Duplication: %s and %s (%.0f%%) ", pairs[0].FuncA, pairs[0].FuncB, pairs[0].Similarity*100)
}

func TestDuplication_Different(t *testing.T) {
	src := "func Add(a, b int) int {\nreturn a + b\n}\nfunc Multiply(a, b int) int {\nresult := 0\nfor i := 0; i < b; i++ {\nresult += a\n}\nreturn result\n}\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloats(lines)
	pairs := parser.DetectDuplications(bloats, lines, 10, 0.75)
	if len(pairs) > 0 {
		t.Errorf("unexpected duplication: %s ~ %s", pairs[0].FuncA, pairs[0].FuncB)
	}
}

func TestLowCohesion(t *testing.T) {
	src := "func ParseInt(s string) (int, error) {\n}\nfunc FormatInt(n int) string {\n}\nfunc ParseFloat(s string) (float64, error) {\n}\nfunc ParseBool(s string) (bool, error) {\n}\nfunc Contains(s, substr string) bool {\n}\n"
	lines := strings.Split(src, "\n")
	bloats := parser.DetectFunctionBloats(lines)
	result := parser.AnalyzeCohesion(bloats, lines)
	t.Logf("Cohesion: %.0f, isolated: %d/%d, low: %v", result.CohesionScore, result.IsolatedFuncs, result.TotalFuncs, result.IsLowCohesion)
}

func TestMessageChains(t *testing.T) {
	lines := strings.Split("r := obj.Method().Chain().Call()", "\n")
	smells := parser.DetectMessageChains(lines)
	// Should detect at least 1 chain
	t.Logf("Message chains found: %d", len(smells))
}

func TestPrimitiveObsession(t *testing.T) {
	lines := strings.Split("func CreateUser(name string, email string, age int, active bool, score float64) error {", "\n")
	smells := parser.DetectPrimitiveObsession(lines)
	if len(smells) == 0 {
		t.Fatal("expected primitive_obsession for function with 5 primitive params")
	}
	t.Logf("Primitive obsession: %d smells", len(smells))
}

func TestDetectLazyElement(t *testing.T) {
	bloats := []parser.FunctionBloat{
		{Name: "foo", LineCount: 1, LineStart: 1},
		{Name: "bar", LineCount: 5, LineStart: 3},
	}
	smells := parser.DetectLazyElements(bloats, 3)
	if len(smells) == 0 {
		t.Fatal("expected lazy_element for 1-line function")
	}
}

func TestDetectParagraphOfCode(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "x := 1"
	}
	smells := parser.DetectParagraphOfCode(lines, 15)
	if len(smells) == 0 {
		t.Fatal("expected paragraph_of_code for 20 consecutive lines")
	}
}

func TestThresholds_AllLanguages(t *testing.T) {
	langs := []string{"go", "python", "javascript", "typescript", "cpp", "java", "rust", "ruby", "swift", "kotlin", "csharp"}
	for _, lang := range langs {
		th := parser.DefaultThresholds(lang)
		if th.FuncLOCWarning == 0 {
			t.Errorf("%s: FuncLOCWarning is 0", lang)
		}
		if th.NestingWarning == 0 {
			t.Errorf("%s: NestingWarning is 0", lang)
		}
	}
}

func TestLanguageDetection(t *testing.T) {
	tests := map[string]string{
		".go":    "go",
		".py":    "python",
		".js":    "javascript",
		".ts":    "typescript",
		".cpp":   "cpp",
		".java":  "java",
		".rs":    "rust",
		".rb":    "ruby",
		".swift": "swift",
		".kt":    "kotlin",
		".cs":    "csharp",
		".xyz":   "",
	}
	for ext, want := range tests {
		got := parser.DetectedLanguage(ext)
		if got != want {
			t.Errorf("DetectedLanguage(%s) = %q, want %q", ext, got, want)
		}
	}
}
