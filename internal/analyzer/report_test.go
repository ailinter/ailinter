package analyzer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/vulnerability"
)

// reportTestBuilder helps build ReportData instances without long parameter lists.
type reportTestBuilder struct {
	target       string
	score        int
	label        string
	results      []analyzer.QualityResult
	secrets      []secrets.SecretFinding
	vulns        []vulnerability.Finding
}

func (b reportTestBuilder) build() *analyzer.ReportData {
	return &analyzer.ReportData{
		Target:       b.target,
		Timestamp:    "2026-05-28T12:00:00Z",
		OverallScore: b.score,
		OverallLabel: b.label,
		Results:      b.results,
		Secrets:      b.secrets,
		Vulns:        b.vulns,
	}
}

func TestComputeOverallScore(t *testing.T) {
	for _, tt := range []struct {
		name     string
		results  []analyzer.QualityResult
		expected int
	}{
		{name: "empty", results: nil, expected: 0},
		{name: "single", results: []analyzer.QualityResult{{Score: 42}}, expected: 42},
		{name: "multiple", results: []analyzer.QualityResult{{Score: 80}, {Score: 90}, {Score: 100}}, expected: 90},
		{name: "edge", results: []analyzer.QualityResult{{Score: 0}, {Score: 100}}, expected: 50},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.TestComputeOverallScoreHelper(tt.results)
			if got != tt.expected {
				t.Errorf("got %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestScoreEmoji(t *testing.T) {
	for _, tt := range []struct {
		score int
		want  string
	}{
		{100, "\U0001f7e2"}, {80, "\U0001f7e2"},
		{79, "\U0001f7e1"}, {60, "\U0001f7e1"},
		{59, "\U0001f534"}, {0, "\U0001f534"}, {-1, "\U0001f534"},
	} {
		t.Run("", func(t *testing.T) {
			got := analyzer.TestScoreEmojiHelper(tt.score)
			if got != tt.want {
				t.Errorf("scoreEmoji(%d) = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

func TestDetectorStatus(t *testing.T) {
	for _, tt := range []struct {
		name     string
		detector string
		results  []analyzer.QualityResult
		contains string
	}{
		{name: "no results", detector: "quality", results: nil, contains: "Passed"},
		{name: "with smells", detector: "quality", results: []analyzer.QualityResult{{Smells: []analyzer.Smell{{Name: "test"}}}}, contains: "1 issue(s) found"},
		{name: "no smells", detector: "quality", results: []analyzer.QualityResult{{Smells: nil}}, contains: "Passed"},
		{name: "non quality", detector: "secrets", results: nil, contains: "Passed"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.TestDetectorStatusHelper(tt.detector, tt.results)
			if !containsStr(got, tt.contains) {
				t.Errorf("got %q, want it to contain %q", got, tt.contains)
			}
		})
	}
}

func TestCountAllSmells(t *testing.T) {
	for _, tt := range []struct {
		name     string
		results  []analyzer.QualityResult
		expected int
	}{
		{name: "nil", results: nil, expected: 0},
		{name: "empty", results: []analyzer.QualityResult{}, expected: 0},
		{name: "no smells", results: []analyzer.QualityResult{{Smells: nil}}, expected: 0},
		{name: "with smells", results: []analyzer.QualityResult{{Smells: []analyzer.Smell{{Name: "a"}, {Name: "b"}}}}, expected: 2},
		{name: "mixed", results: []analyzer.QualityResult{
			{Smells: []analyzer.Smell{{Name: "a"}}},
			{Smells: []analyzer.Smell{{Name: "b"}, {Name: "c"}}},
			{Smells: nil},
		}, expected: 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.TestCountAllSmellsHelper(tt.results)
			if got != tt.expected {
				t.Errorf("got %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestIsSourceFileReport(t *testing.T) {
	for _, tt := range []struct {
		path string
		want bool
	}{
		{"file.go", true}, {"file.py", true}, {"file.js", true},
		{"file.ts", true}, {"file.java", true}, {"file.rs", true},
		{"file.rb", true}, {"file.c", true}, {"file.cpp", true},
		{"file.h", true}, {"file.cs", true}, {"file.swift", true},
		{"file.kt", true}, {"file.php", true}, {"file.sh", true},
		{"file.yaml", true}, {"file.toml", true}, {"file.json", true},
		{"file.xml", true}, {"file.html", true}, {"file.sql", true},
		{".env", true}, {"Dockerfile", true}, {"Makefile", true},
		{".gitignore", true}, {".npmrc", true}, {".editorconfig", true},
		{".env.local", true}, {"Dockerfile.dev", true},
		{"image.png", false}, {"image.jpg", false}, {"app.exe", false},
		{"app.bin", false}, {"archive.zip", false},
	} {
		t.Run(tt.path, func(t *testing.T) {
			got := analyzer.TestIsSourceFileReportHelper(tt.path)
			if got != tt.want {
				t.Errorf("isSourceFileReport(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsBinaryFile(t *testing.T) {
	for _, tt := range []struct {
		name string
		data []byte
		want bool
	}{
		{"nil", nil, false},
		{"empty", []byte{}, false},
		{"text", []byte("hello world\npackage main\n"), false},
		{"null byte", []byte("hello\x00world"), true},
		{"null end", append([]byte("hello"), 0), true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.TestIsBinaryFileHelper(tt.data)
			if got != tt.want {
				t.Errorf("isBinaryFile(%v) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestLoadGitignoreSimple(t *testing.T) {
	setup := func(t *testing.T, content string) string {
		t.Helper()
		dir := t.TempDir()
		if content != "" {
			if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}

	t.Run("normal", func(t *testing.T) {
		dir := setup(t, "# comments\n*.log\n/build/\n/vendor\n\nsecret.txt\n")
		pats := analyzer.TestLoadGitignoreSimpleHelper(dir)
		want := []string{"*.log", "build/", "vendor", "secret.txt"}
		if len(pats) != len(want) {
			t.Fatalf("got %d patterns, want %d: %v", len(pats), len(want), pats)
		}
		for i := range pats {
			if pats[i] != want[i] {
				t.Errorf("pattern[%d] = %q, want %q", i, pats[i], want[i])
			}
		}
	})

	t.Run("comments filtered", func(t *testing.T) {
		dir := setup(t, "# comment\n\n\n# another\n*.tmp\n")
		pats := analyzer.TestLoadGitignoreSimpleHelper(dir)
		if len(pats) != 1 || pats[0] != "*.tmp" {
			t.Fatalf("got %v, want [*.tmp]", pats)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		dir := setup(t, "")
		pats := analyzer.TestLoadGitignoreSimpleHelper(dir)
		if pats != nil {
			t.Fatalf("expected nil, got %v", pats)
		}
	})
}

func TestIsGitignoredReport(t *testing.T) {
	for _, tt := range []struct {
		name     string
		path     string
		root     string
		patterns []string
		want     bool
	}{
		{"basename", "/repo/secret.txt", "/repo", []string{"secret.txt"}, true},
		{"rel path", "/repo/subdir/file.go", "/repo", []string{"subdir/file.go"}, true},
		{"dir prefix", "/repo/build/output.o", "/repo", []string{"build/"}, true},
		{"no match", "/repo/main.go", "/repo", []string{"*.log", "vendor/"}, false},
		{"glob", "/repo/debug.log", "/repo", []string{"*.log"}, true},
		{"outer", "/other/main.go", "/repo", []string{"subdir/*"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.TestIsGitignoredReportHelper(tt.path, tt.root, tt.patterns)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderMarkdown_SingleFile(t *testing.T) {
	rd := reportTestBuilder{
		target: "/test/main.go", score: 85, label: "Go Ahead",
		results: []analyzer.QualityResult{{Score: 85, Label: "Go Ahead", FilePath: "/test/main.go", Language: "go", LinesOfCode: 50}},
	}.build()
	md := rd.RenderMarkdown()
	assertContains(t, md, "Code Quality Report")
	assertContains(t, md, "/test/main.go")
	assertContains(t, md, "85/100")
	assertContains(t, md, "No issues detected")
	assertContains(t, md, "No secrets detected")
	assertContains(t, md, "No vulnerability patterns detected")
}

func TestRenderMarkdown_MultiFile(t *testing.T) {
	rd := reportTestBuilder{
		target: "/test", score: 70, label: "Proceed with Care",
		results: []analyzer.QualityResult{
			{Score: 90, Label: "Go Ahead", FilePath: "/test/a.go", Language: "go", LinesOfCode: 30, Smells: []analyzer.Smell{{Name: "test", Severity: "warning", LineStart: 5, Message: "test message"}}},
			{Score: 50, Label: "Needs Work", FilePath: "/test/b.go", Language: "go", LinesOfCode: 200},
		},
	}.build()
	md := rd.RenderMarkdown()
	assertContains(t, md, "Files Analyzed")
	assertContains(t, md, "Average Score")
	assertContains(t, md, "Go Ahead (80-100)")
	assertContains(t, md, "1 issue(s) found")
}

func TestRenderMarkdown_WithSecrets(t *testing.T) {
	rd := reportTestBuilder{
		target: "/test/main.go", score: 80, label: "Go Ahead",
		results: []analyzer.QualityResult{{Score: 80, Label: "Go Ahead", FilePath: "/test/main.go", Language: "go", LinesOfCode: 30}},
		secrets: []secrets.SecretFinding{{RuleID: "generic-api-key", Severity: "high", Line: 10, Description: "API key detected"}},
	}.build()
	md := rd.RenderMarkdown()
	assertContains(t, md, "1 secret(s) detected")
	assertContains(t, md, "generic-api-key")
}

func TestRenderMarkdown_WithVulnerabilities(t *testing.T) {
	rd := reportTestBuilder{
		target: "/test/main.go", score: 80, label: "Go Ahead",
		results: []analyzer.QualityResult{{Score: 80, Label: "Go Ahead", FilePath: "/test/main.go", Language: "go", LinesOfCode: 30}},
		vulns: []vulnerability.Finding{{Category: "injection", RuleID: "eval_injection", Severity: "critical", Line: 15}},
	}.build()
	md := rd.RenderMarkdown()
	assertContains(t, md, "1 vulnerability pattern(s) detected")
	assertContains(t, md, "eval_injection")
}

func TestRenderMarkdown_CleanNoIssues(t *testing.T) {
	rd := reportTestBuilder{
		target: "/test/clean.go", score: 100, label: "Go Ahead",
		results: []analyzer.QualityResult{{Score: 100, Label: "Go Ahead", FilePath: "/test/clean.go", Language: "go", LinesOfCode: 10}},
	}.build()
	md := rd.RenderMarkdown()
	assertContains(t, md, "No issues detected")
	assertContains(t, md, "No secrets detected")
	assertContains(t, md, "No vulnerability patterns detected")
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateReport_SingleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	writeFile(t, path, "package main\nfunc main() {}\n")

	rd, err := analyzer.GenerateReport(path, false)
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}
	if rd.Target != path {
		t.Errorf("Target = %q, want %q", rd.Target, path)
	}
	if rd.OverallScore <= 0 {
		t.Errorf("OverallScore = %d, want > 0", rd.OverallScore)
	}
	if len(rd.Results) != 1 {
		t.Errorf("got %d results, want 1", len(rd.Results))
	}
}

func TestGenerateReport_NonexistentPath(t *testing.T) {
	_, err := analyzer.GenerateReport("/nonexistent/path.go", false)
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestGenerateReport_Directory(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.go"), "package main\nfunc main() {}\n")
	writeFile(t, filepath.Join(dir, "util.go"), "package util\nfunc Help() string { return \"ok\" }\n")

	rd, err := analyzer.GenerateReport(dir, false)
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}
	if rd == nil {
		t.Fatal("GenerateReport() returned nil")
	}
	if len(rd.Results) < 1 {
		t.Errorf("got %d results, want >= 1", len(rd.Results))
	}
	if rd.OverallScore <= 0 {
		t.Errorf("OverallScore = %d, want > 0", rd.OverallScore)
	}
}

func assertContains(t testing.TB, s, substr string) {
	t.Helper()
	if !containsStr(s, substr) {
		t.Errorf("expected string to contain %q", substr)
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
