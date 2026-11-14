package analyzer_test

import (
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/parser"
)

func TestAnalyze_DataLanguagesSkipCodeSmells(t *testing.T) {
	source := `{"a":{"b":{"c":1}}}`
	langs := []string{"json", "toml", "xml", "yaml", "html", "css", "properties", "env"}
	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			result := analyzer.Analyze(
				analyzer.SourceInput{FilePath: "data." + lang, Source: source, Lang: lang},
				analyzer.DefaultThresholds(lang),
			)
			if len(result.Smells) != 0 {
				t.Errorf("data language %s should have no code smells: %+v", lang, result.Smells)
			}
			if result.Score != 100 {
				t.Errorf("data language %s score = %d, want 100", lang, result.Score)
			}
			if result.Language != lang {
				t.Errorf("language = %q, want %q", result.Language, lang)
			}
		})
	}
}

func TestAnalyze_CodeLanguagesStillAnalyzed(t *testing.T) {
	result := analyzer.Analyze(
		analyzer.SourceInput{
			FilePath: "app.php",
			Source:   "<?php\nfunction one() {\n    if (true) {\n        return 1;\n    }\n    return 0;\n}\n",
			Lang:     "php",
		},
		analyzer.DefaultThresholds("php"),
	)
	if result.Language != "php" {
		t.Errorf("language = %q, want php", result.Language)
	}
	if result.Score == 0 {
		t.Error("code language should produce a score")
	}
}

func TestIsCodeLanguage(t *testing.T) {
	assertCodeLanguages(t, []string{"go", "javascript", "python", "php", "scala", "shell", "sql", "html", "css"}, true)
	assertCodeLanguages(t, []string{"json", "yaml", "toml", "xml", "env"}, false)
}

func assertCodeLanguages(t *testing.T, langs []string, want bool) {
	t.Helper()
	for _, lang := range langs {
		if got := parser.IsCodeLanguage(lang); got != want {
			t.Errorf("IsCodeLanguage(%s) = %v, want %v", lang, got, want)
		}
	}
}
