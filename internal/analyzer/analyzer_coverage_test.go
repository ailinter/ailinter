package analyzer_test

import (
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

func TestAnalyze_AllLanguages(t *testing.T) {
	langs := []string{"go", "python", "cpp", "java", "rust", "ruby", "swift", "kotlin", "csharp"}
	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			result := analyzer.Analyze("test."+lang, "// empty", lang, analyzer.DefaultThresholds(lang))
			if result.Score < 10 || result.Score > 100 {
				t.Errorf("%s: score out of range: %d", lang, result.Score)
			}
		})
	}
}

func TestAnalyze_EmptyFile(t *testing.T) {
	result := analyzer.Analyze("empty.go", "", "go", analyzer.DefaultThresholds("go"))
	if result.Score != 100 {
		t.Errorf("empty file should score 100, got %d", result.Score)
	}
}

func TestAnalyze_VeryLargeFile(t *testing.T) {
	src := "package main\n"
	for i := 0; i < 2000; i++ {
		src += "var x" + string(rune('a'+i%26)) + " = 1\n"
	}
	result := analyzer.Analyze("big.go", src, "go", analyzer.DefaultThresholds("go"))
	// Large but flat file — should detect file_bloat, score moderately
	t.Logf("Score: %d, Smells: %d", result.Score, len(result.Smells))
}

func TestAnalyze_ManyFunctions(t *testing.T) {
	src := "package main\n"
	for i := 0; i < 30; i++ {
		src += "func f" + string(rune('a'+i%26)) + "() {\nprintln(\"hi\")\n}\n"
	}
	result := analyzer.Analyze("many.go", src, "go", analyzer.DefaultThresholds("go"))
	t.Logf("Score: %d, Smells: %d", result.Score, len(result.Smells))
	for _, s := range result.Smells {
		t.Logf("  [%s] %s", s.Severity, s.Name)
	}
}

func TestAnalyze_BumpyRoadGo(t *testing.T) {
	src := "package main\nfunc TwoBumps(items []int) int {\nc := 0\nfor _, it := range items {\nif it > 0 {\nif it < 100 {\nc += it\n}\n}\n}\nfor _, it := range items {\nif it < 0 {\nif it > -100 {\nc -= it\n}\n}\n}\nreturn c\n}\n"
	result := analyzer.Analyze("bumps.go", src, "go", analyzer.DefaultThresholds("go"))
	t.Logf("Score: %d, Smells: %d", result.Score, len(result.Smells))
}

func TestAnalyze_PythonFile(t *testing.T) {
	src := "def greet(name):\n    return f'Hello {name}'\n"
	result := analyzer.Analyze("test.py", src, "python", analyzer.DefaultThresholds("python"))
	if result.Language != "python" {
		t.Errorf("expected python, got %s", result.Language)
	}
}

func TestAnalyze_JavaScriptFile(t *testing.T) {
	src := "function greet(name) {\n    return 'Hello ' + name\n}\n"
	result := analyzer.Analyze("test.js", src, "javascript", analyzer.DefaultThresholds("javascript"))
	if result.Language != "javascript" {
		t.Errorf("expected javascript, got %s", result.Language)
	}
}

func TestSeverityWeight(t *testing.T) {
	// Test that different severities produce different scores
	src := "package main\nfunc main() {}\n"
	clean := analyzer.Analyze("a.go", src, "go", analyzer.DefaultThresholds("go"))

	srcDeep := "package main\nfunc main() {\nif true{\nif true{\nif true{\nif true{\nif true{\n}\n}\n}\n}\n}\n}\n"
	deep := analyzer.Analyze("b.go", srcDeep, "go", analyzer.DefaultThresholds("go"))

	if deep.Score >= clean.Score {
		t.Errorf("deeply nested file (%d) should score lower than clean file (%d)", deep.Score, clean.Score)
	}
}
