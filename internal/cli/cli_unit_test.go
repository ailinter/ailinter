package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckCommand_Exists(t *testing.T) {
	cmd := CheckCommand()
	if cmd == nil {
		t.Fatal("CheckCommand returned nil")
	}
}

func TestMCPCommand_Exists(t *testing.T) {
	cmd := MCPCommand("0.1.0")
	if cmd == nil {
		t.Fatal("MCPCommand returned nil")
	}
}

func TestInitCommand_Exists(t *testing.T) {
	cmd := InitCommand()
	if cmd == nil {
		t.Fatal("InitCommand returned nil")
	}
}

func TestInitCommand_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)

	cmd := InitCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".ailinter.toml")); err != nil {
		t.Error(".ailinter.toml not created")
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Error("AGENTS.md not created")
	}
}

func TestInitCommand_NoAgents(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)

	cmd := InitCommand()
	cmd.SetArgs([]string{"--no-agents"})
	cmd.Execute()

	if _, err := os.Stat(filepath.Join(dir, ".ailinter.toml")); err != nil {
		t.Error(".ailinter.toml not created")
	}
}

func TestCheckCommand_FileNotFound(t *testing.T) {
	cmd := CheckCommand()
	cmd.SetArgs([]string{"/nonexistent/path.go"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestInitCommand_Idempotent(t *testing.T) {
	dir := t.TempDir()
	os.Chdir(dir)
	cmd := InitCommand()
	cmd.SetArgs([]string{})
	cmd.Execute() // first init
	cmd = InitCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute() // second init — should not error
	if err != nil {
		t.Errorf("second init should not error: %v", err)
	}
}

func TestIsValidLanguageName(t *testing.T) {
	valid := []string{"go", "python", "javascript", "typescript", "java", "csharp", "ruby", "swift", "kotlin", "rust", "cpp", "c"}
	for _, lang := range valid {
		if !isValidLanguageName(lang) {
			t.Errorf("expected '%s' to be valid", lang)
		}
	}

	invalid := []string{"", "frobulator", "yaml", "html", "Go", "PYTHON", "c++"}
	for _, lang := range invalid {
		if isValidLanguageName(lang) {
			t.Errorf("expected '%s' to be invalid", lang)
		}
	}
}

func TestCheckCommand_HasNoVulnerabilitiesFlag(t *testing.T) {
	cmd := CheckCommand()
	flag := cmd.Flags().Lookup("no-vulnerabilities")
	if flag == nil {
		t.Fatal("--no-vulnerabilities flag not found")
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default false, got %s", flag.DefValue)
	}
}

func TestCheckCommand_HasAllFlags(t *testing.T) {
	cmd := CheckCommand()
	expected := []string{"format", "no-secrets", "no-vulnerabilities", "lang", "no-gitignore"}
	for _, name := range expected {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("flag --%s not found", name)
		}
	}
}

func TestLoadGitignore(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\n/node_modules/\n.env\n"), 0644)

	patterns := loadGitignore(dir)
	if len(patterns) != 3 {
		t.Fatalf("expected 3 patterns, got %d: %v", len(patterns), patterns)
	}
	if patterns[0] != "*.log" {
		t.Errorf("expected '*.log', got %q", patterns[0])
	}
	if patterns[1] != "node_modules/" {
		t.Errorf("expected 'node_modules/', got %q", patterns[1])
	}
}

func TestLoadGitignore_CommentsAndEmpty(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("# This is a comment\n\n*.log\n\n# Another comment\n.env\n"), 0644)

	patterns := loadGitignore(dir)
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns (comments/empty skipped), got %d: %v", len(patterns), patterns)
	}
}

func TestLoadGitignore_NotExist(t *testing.T) {
	dir := t.TempDir()
	patterns := loadGitignore(dir)
	if len(patterns) != 0 {
		t.Errorf("expected 0 patterns for missing .gitignore, got %d", len(patterns))
	}
}

func TestIsGitignored(t *testing.T) {
	patterns := []string{"*.log", "node_modules/", "secrets/", ".env"}
	root := "/project"

	tests := []struct {
		path string
		want bool
	}{
		{"/project/app.log", true},
		{"/project/src/debug.log", true},
		{"/project/node_modules/pkg/index.js", true},
		{"/project/node_modules", false},
		{"/project/secrets/key.txt", true},
		{"/project/.env", true},
		{"/project/main.go", false},
		{"/project/readme.md", false},
		{"/project/src/app.py", false},
	}
	for _, tc := range tests {
		if got := isGitignored(tc.path, root, patterns); got != tc.want {
			t.Errorf("isGitignored(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestIsGitignored_LeadingSlash(t *testing.T) {
	patterns := parseGitignoreContent("/build/\noutput/\n*.tmp\n")
	root := "/project"

	if !isGitignored("/project/build/artifact.o", root, patterns) {
		t.Error("build/ should be gitignored")
	}
	if !isGitignored("/project/output/release.exe", root, patterns) {
		t.Error("output/ should be gitignored")
	}
}

func parseGitignoreContent(content string) []string {
	var patterns []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "/")
		patterns = append(patterns, line)
	}
	return patterns
}

func TestIsSourceFile_Internal(t *testing.T) {
	sources := []string{"main.go", "app.py", "script.js", "component.tsx", "App.java", "lib.rs", "gem.rb", "program.c", "Program.cs", "main.swift", "App.kt", "config.yaml", "index.html", "style.css", "query.sql", ".env", "Dockerfile", "Makefile", ".gitignore", "docker-compose.yml"}
	for _, p := range sources {
		if !isSourceFile(p) {
			t.Errorf("expected %q to be a source file", p)
		}
	}

	nonSources := []string{"image.png", "archive.zip", "binary.exe", "data.bin", "video.mp4"}
	for _, p := range nonSources {
		if isSourceFile(p) {
			t.Errorf("expected %q NOT to be a source file", p)
		}
	}
}

func TestIsBinary(t *testing.T) {
	if isBinary(nil) {
		t.Error("nil should not be binary")
	}
	if isBinary([]byte{}) {
		t.Error("empty should not be binary")
	}
	if !isBinary([]byte{0x00, 0x01, 0x02}) {
		t.Error("null byte should be binary")
	}
	if isBinary([]byte("package main\nfunc main() {}\n")) {
		t.Error("Go source should not be binary")
	}
}

func TestCheckCommand_BadFormatRejected(t *testing.T) {
	cmd := CheckCommand()
	cmd.SetArgs([]string{"--format", "xml", "/dev/null"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for bad format")
	}
}

func TestCheckCommand_BadLangRejected(t *testing.T) {
	cmd := CheckCommand()
	cmd.SetArgs([]string{"--lang", "frobulator", "/dev/null"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for bad language")
	}
}
