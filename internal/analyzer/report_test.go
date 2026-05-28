package analyzer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/vulnerability"
)

func TestComputeOverallScore(t *testing.T) {
	tests := []struct {
		name     string
		results  []analyzer.QualityResult
		expected int
	}{
		{
			name:     "empty results",
			results:  nil,
			expected: 0,
		},
		{
			name: "single result",
			results: []analyzer.QualityResult{
				{Score: 42},
			},
			expected: 42,
		},
		{
			name: "multiple results",
			results: []analyzer.QualityResult{
				{Score: 80},
				{Score: 90},
				{Score: 100},
			},
			expected: 90,
		},
		{
			name: "min and max",
			results: []analyzer.QualityResult{
				{Score: 0},
				{Score: 100},
			},
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.TestComputeOverallScoreHelper(tt.results)
			if got != tt.expected {
				t.Errorf("computeOverallScore(%v) = %d, want %d", tt.results, got, tt.expected)
			}
		})
	}
}

func TestScoreEmoji(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{100, "\U0001f7e2"},
		{80, "\U0001f7e2"},
		{79, "\U0001f7e1"},
		{60, "\U0001f7e1"},
		{59, "\U0001f534"},
		{0, "\U0001f534"},
		{-1, "\U0001f534"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := analyzer.TestScoreEmojiHelper(tt.score)
			if got != tt.want {
				t.Errorf("scoreEmoji(%d) = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

func TestDetectorStatus(t *testing.T) {
	tests := []struct {
		name     string
		detector string
		results  []analyzer.QualityResult
		contains string
	}{
		{
			name:     "quality no results",
			detector: "quality",
			results:  nil,
			contains: "Passed",
		},
		{
			name:     "quality with smells",
			detector: "quality",
			results: []analyzer.QualityResult{
				{Smells: []analyzer.Smell{{Name: "test"}}},
			},
			contains: "1 issue(s) found",
		},
		{
			name:     "quality without smells",
			detector: "quality",
			results: []analyzer.QualityResult{
				{Smells: nil},
			},
			contains: "Passed",
		},
		{
			name:     "non-quality detector",
			detector: "secrets",
			results:  nil,
			contains: "Passed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.TestDetectorStatusHelper(tt.detector, tt.results)
			if !containsStr(got, tt.contains) {
				t.Errorf("detectorStatus(%q, _) = %q, want it to contain %q", tt.detector, got, tt.contains)
			}
		})
	}
}

func TestCountAllSmells(t *testing.T) {
	tests := []struct {
		name     string
		results  []analyzer.QualityResult
		expected int
	}{
		{
			name:     "nil results",
			results:  nil,
			expected: 0,
		},
		{
			name:     "empty results",
			results:  []analyzer.QualityResult{},
			expected: 0,
		},
		{
			name: "single result no smells",
			results: []analyzer.QualityResult{
				{Smells: nil},
			},
			expected: 0,
		},
		{
			name: "single result with smells",
			results: []analyzer.QualityResult{
				{Smells: []analyzer.Smell{{Name: "a"}, {Name: "b"}}},
			},
			expected: 2,
		},
		{
			name: "multiple results with mixed smells",
			results: []analyzer.QualityResult{
				{Smells: []analyzer.Smell{{Name: "a"}}},
				{Smells: []analyzer.Smell{{Name: "b"}, {Name: "c"}}},
				{Smells: nil},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.TestCountAllSmellsHelper(tt.results)
			if got != tt.expected {
				t.Errorf("countAllSmells(%v) = %d, want %d", tt.results, got, tt.expected)
			}
		})
	}
}

func TestIsSourceFileReport(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "file.go", want: true},
		{path: "file.py", want: true},
		{path: "file.js", want: true},
		{path: "file.ts", want: true},
		{path: "file.tsx", want: true},
		{path: "file.java", want: true},
		{path: "file.rs", want: true},
		{path: "file.rb", want: true},
		{path: "file.c", want: true},
		{path: "file.cpp", want: true},
		{path: "file.h", want: true},
		{path: "file.hpp", want: true},
		{path: "file.cs", want: true},
		{path: "file.swift", want: true},
		{path: "file.kt", want: true},
		{path: "file.kts", want: true},
		{path: "file.scala", want: true},
		{path: "file.php", want: true},
		{path: "file.pl", want: true},
		{path: "file.sh", want: true},
		{path: "file.bash", want: true},
		{path: "file.tf", want: true},
		{path: "file.yaml", want: true},
		{path: "file.yml", want: true},
		{path: "file.toml", want: true},
		{path: "file.json", want: true},
		{path: "file.xml", want: true},
		{path: "file.html", want: true},
		{path: "file.css", want: true},
		{path: "file.sql", want: true},
		{path: "file.properties", want: true},
		{path: "file.ini", want: true},
		{path: "file.cfg", want: true},
		{path: "file.conf", want: true},
		{path: ".env", want: true},
		{path: "Dockerfile", want: true},
		{path: "Makefile", want: true},
		{path: ".gitignore", want: true},
		{path: ".gitattributes", want: true},
		{path: ".npmrc", want: true},
		{path: ".editorconfig", want: true},
		{path: ".dockerignore", want: true},
		{path: ".env.local", want: true},
		{path: ".env.production", want: true},
		{path: "Dockerfile.dev", want: true},
		{path: "Dockerfile.prod", want: true},
		{path: "image.png", want: false},
		{path: "image.jpg", want: false},
		{path: "app.exe", want: false},
		{path: "app.bin", want: false},
		{path: "archive.zip", want: false},
		{path: "archive.tar.gz", want: false},
		{path: "video.mp4", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := analyzer.TestIsSourceFileReportHelper(tt.path)
			if got != tt.want {
				t.Errorf("isSourceFileReport(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsBinaryFile(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{name: "nil", data: nil, want: false},
		{name: "empty", data: []byte{}, want: false},
		{name: "text without null bytes", data: []byte("hello world\npackage main\n"), want: false},
		{name: "text with null byte", data: []byte("hello\x00world"), want: true},
		{name: "null at end", data: append([]byte("hello"), 0), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.TestIsBinaryFileHelper(tt.data)
			if got != tt.want {
				t.Errorf("isBinaryFile(%v) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestLoadGitignoreSimple(t *testing.T) {
	t.Run("normal patterns", func(t *testing.T) {
		dir := t.TempDir()
		content := "# comments\n*.log\n/build/\n/vendor\n\nsecret.txt\n"
		if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		pats := analyzer.TestLoadGitignoreSimpleHelper(dir)
		expected := []string{"*.log", "build/", "vendor", "secret.txt"}
		if len(pats) != len(expected) {
			t.Fatalf("got %d patterns, want %d: %v", len(pats), len(expected), pats)
		}
		for i, p := range pats {
			if p != expected[i] {
				t.Errorf("pattern[%d] = %q, want %q", i, p, expected[i])
			}
		}
	})

	t.Run("comments and empty lines filtered", func(t *testing.T) {
		dir := t.TempDir()
		content := "# this is a comment\n\n\n# another comment\n*.tmp\n"
		if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		pats := analyzer.TestLoadGitignoreSimpleHelper(dir)
		if len(pats) != 1 || pats[0] != "*.tmp" {
			t.Fatalf("got %v, want [*.tmp]", pats)
		}
	})

	t.Run("file not existing", func(t *testing.T) {
		dir := t.TempDir()
		pats := analyzer.TestLoadGitignoreSimpleHelper(dir)
		if pats != nil {
			t.Fatalf("expected nil, got %v", pats)
		}
	})
}

func TestIsGitignoredReport(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		root     string
		patterns []string
		want     bool
	}{
		{
			name:     "match basename",
			path:     "/repo/secret.txt",
			root:     "/repo",
			patterns: []string{"secret.txt"},
			want:     true,
		},
		{
			name:     "match rel path",
			path:     "/repo/subdir/file.go",
			root:     "/repo",
			patterns: []string{"subdir/file.go"},
			want:     true,
		},
		{
			name:     "directory prefix match",
			path:     "/repo/build/output.o",
			root:     "/repo",
			patterns: []string{"build/"},
			want:     true,
		},
		{
			name:     "no match",
			path:     "/repo/main.go",
			root:     "/repo",
			patterns: []string{"*.log", "vendor/"},
			want:     false,
		},
		{
			name:     "glob basename",
			path:     "/repo/debug.log",
			root:     "/repo",
			patterns: []string{"*.log"},
			want:     true,
		},
		{
			name:     "outer path no match rel",
			path:     "/other/main.go",
			root:     "/repo",
			patterns: []string{"subdir/*"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.TestIsGitignoredReportHelper(tt.path, tt.root, tt.patterns)
			if got != tt.want {
				t.Errorf("isGitignoredReport(%q, %q, %v) = %v, want %v", tt.path, tt.root, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestRenderMarkdown(t *testing.T) {
	t.Run("single file result", func(t *testing.T) {
		rd := &analyzer.ReportData{
			Target:       "/test/main.go",
			Timestamp:    "2026-05-28T12:00:00Z",
			OverallScore: 85,
			OverallLabel: "Go Ahead",
			Results: []analyzer.QualityResult{
				{Score: 85, Label: "Go Ahead", FilePath: "/test/main.go", Language: "go", LinesOfCode: 50},
			},
		}
		md := rd.RenderMarkdown()
		assertContains(t, md, "Code Quality Report")
		assertContains(t, md, "/test/main.go")
		assertContains(t, md, "85/100")
		assertContains(t, md, "Score Breakdown")
		assertContains(t, md, "Detector Results")
		assertContains(t, md, "Secret Scan")
		assertContains(t, md, "Vulnerability Scan")
		assertContains(t, md, "All Issues")
		assertContains(t, md, "No issues detected")
		assertContains(t, md, "No secrets detected")
		assertContains(t, md, "No vulnerability patterns detected")
	})

	t.Run("multi file results", func(t *testing.T) {
		rd := &analyzer.ReportData{
			Target:       "/test",
			Timestamp:    "2026-05-28T12:00:00Z",
			OverallScore: 70,
			OverallLabel: "Proceed with Care",
			Results: []analyzer.QualityResult{
				{Score: 90, Label: "Go Ahead", FilePath: "/test/a.go", Language: "go", LinesOfCode: 30, Smells: []analyzer.Smell{{Name: "test", Severity: "warning", LineStart: 5, Message: "test message"}}},
				{Score: 50, Label: "Needs Work", FilePath: "/test/b.go", Language: "go", LinesOfCode: 200},
			},
		}
		md := rd.RenderMarkdown()
		assertContains(t, md, "Files Analyzed")
		assertContains(t, md, "Average Score")
		assertContains(t, md, "Go Ahead (80-100)")
		assertContains(t, md, "Needs Work (40-59)")
		assertContains(t, md, "Detector Results")
		assertContains(t, md, "1 issue(s) found")
	})

	t.Run("with secrets found", func(t *testing.T) {
		rd := &analyzer.ReportData{
			Target:       "/test/main.go",
			Timestamp:    "2026-05-28T12:00:00Z",
			OverallScore: 80,
			OverallLabel: "Go Ahead",
			Results: []analyzer.QualityResult{
				{Score: 80, Label: "Go Ahead", FilePath: "/test/main.go", Language: "go", LinesOfCode: 30},
			},
			Secrets: []secrets.SecretFinding{
				{RuleID: "generic-api-key", Severity: "high", Line: 10, Description: "API key detected"},
			},
		}
		md := rd.RenderMarkdown()
		assertContains(t, md, "1 secret(s) detected")
		assertContains(t, md, "generic-api-key")
		assertContains(t, md, "API key detected")
	})

	t.Run("with vulnerabilities found", func(t *testing.T) {
		rd := &analyzer.ReportData{
			Target:       "/test/main.go",
			Timestamp:    "2026-05-28T12:00:00Z",
			OverallScore: 80,
			OverallLabel: "Go Ahead",
			Results: []analyzer.QualityResult{
				{Score: 80, Label: "Go Ahead", FilePath: "/test/main.go", Language: "go", LinesOfCode: 30},
			},
			Vulns: []vulnerability.Finding{
				{Category: "injection", RuleID: "eval_injection", Severity: "critical", Line: 15},
			},
		}
		md := rd.RenderMarkdown()
		assertContains(t, md, "1 vulnerability pattern(s) detected")
		assertContains(t, md, "eval_injection")
		assertContains(t, md, "injection")
	})

	t.Run("clean no issues", func(t *testing.T) {
		rd := &analyzer.ReportData{
			Target:       "/test/clean.go",
			Timestamp:    "2026-05-28T12:00:00Z",
			OverallScore: 100,
			OverallLabel: "Go Ahead",
			Results: []analyzer.QualityResult{
				{Score: 100, Label: "Go Ahead", FilePath: "/test/clean.go", Language: "go", LinesOfCode: 10},
			},
		}
		md := rd.RenderMarkdown()
		assertContains(t, md, "No issues detected")
		assertContains(t, md, "No secrets detected")
		assertContains(t, md, "No vulnerability patterns detected")
	})
}

func TestGenerateReport(t *testing.T) {
	t.Run("single file", func(t *testing.T) {
		dir := t.TempDir()
		src := "package main\nfunc main() {}\n"
		path := filepath.Join(dir, "main.go")
		if err := os.WriteFile(path, []byte(src), 0644); err != nil {
			t.Fatal(err)
		}

		rd, err := analyzer.GenerateReport(path, false)
		if err != nil {
			t.Fatalf("GenerateReport() error = %v", err)
		}
		if rd == nil {
			t.Fatal("GenerateReport() returned nil")
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
	})

	t.Run("nonexistent path", func(t *testing.T) {
		_, err := analyzer.GenerateReport("/nonexistent/path.go", false)
		if err == nil {
			t.Fatal("expected error for nonexistent path")
		}
	})

	t.Run("directory", func(t *testing.T) {
		dir := t.TempDir()
		src1 := "package main\nfunc main() {}\n"
		src2 := "package util\nfunc Help() string { return \"ok\" }\n"
		if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src1), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "util.go"), []byte(src2), 0644); err != nil {
			t.Fatal(err)
		}

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
	})
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
