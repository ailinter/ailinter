package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func projectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..")
}

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "ailinter")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/ailinter")
	cmd.Dir = projectRoot()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func TestCLI_Version(t *testing.T) {
	bin := buildBinary(t)
	out, _ := exec.Command(bin, "--version").CombinedOutput()
	t.Logf("version: %s", strings.TrimSpace(string(out)))
}

func TestCLI_CheckHealthyFile(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)
	out, _ := exec.Command(bin, "check", f).CombinedOutput()
	if !strings.Contains(string(out), "Go Ahead") {
		t.Errorf("expected Clean: %s", out)
	}
}

func TestCLI_CheckJSON_FormatFlag(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)
	out, _ := exec.Command(bin, "check", "--format", "json", f).CombinedOutput()
	if !strings.Contains(string(out), "\"score\"") {
		t.Errorf("expected JSON score: %s", out)
	}
}

func TestCLI_CheckJSON_LegacyFlag(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)
	out, _ := exec.Command(bin, "check", "--json", f).CombinedOutput()
	if !strings.Contains(string(out), "\"score\"") {
		t.Errorf("expected JSON score: %s", out)
	}
}

func TestCLI_CheckMarkdown(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)
	out, _ := exec.Command(bin, "check", "--format", "markdown", f).CombinedOutput()
	s := string(out)
	if !strings.Contains(s, "## ") {
		t.Errorf("expected markdown heading: %s", s)
	}
	if !strings.Contains(s, "**Score:**") {
		t.Errorf("expected markdown score: %s", s)
	}
}

func TestCLI_CheckHuman(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)
	out, _ := exec.Command(bin, "check", "--format", "human", f).CombinedOutput()
	if !strings.Contains(string(out), "Go Ahead") {
		t.Errorf("expected Clean: %s", out)
	}
}

func TestCLI_CheckDir(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\nfunc main() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("package main\nfunc f() {}\n"), 0644)
	out, _ := exec.Command(bin, "check", dir).CombinedOutput()
	s := string(out)
	if !strings.Contains(s, "a.go") || !strings.Contains(s, "b.go") {
		t.Errorf("expected both files: %s", s)
	}
}

func TestCLI_CheckDirJSON(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\nfunc main() {}\n"), 0644)
	out, _ := exec.Command(bin, "check", "--format", "json", dir).CombinedOutput()
	if !strings.Contains(string(out), "\"score\"") {
		t.Errorf("expected JSON array: %s", out)
	}
}

func TestCLI_CheckDirMarkdown(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\nfunc main() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("package main\nfunc f() {}\n"), 0644)
	out, _ := exec.Command(bin, "check", "--format", "markdown", dir).CombinedOutput()
	s := string(out)
	if !strings.Contains(s, "## ") {
		t.Errorf("expected markdown: %s", s)
	}
	if !strings.Contains(s, "Summary") {
		t.Errorf("expected summary: %s", s)
	}
}

func TestCLI_CheckProblems(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	f := filepath.Join(dir, "deep.go")
	src := `package main

func main() {
	if true {
		if true {
			if true {
				if true {
					println("deep")
				}
			}
		}
	}
}
`

	os.WriteFile(f, []byte(src), 0644)
	out, _ := exec.Command(bin, "check", "--format", "problems", f).CombinedOutput()
	s := string(out)
	if !strings.Contains(s, ": warning: deep_nesting") {
		t.Errorf("expected gcc diagnostic: %s", s)
	}
	if !strings.Contains(s, "deep.go:") {
		t.Errorf("expected file:line prefix: %s", s)
	}
}

func TestCLI_Init(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	cmd := exec.Command(bin, "init")
	cmd.Dir = dir
	cmd.Run()
	if _, err := os.Stat(filepath.Join(dir, ".ailinter.toml")); err != nil {
		t.Errorf(".ailinter.toml not created: %v", err)
	}
}

func TestCLI_InitNoAgents(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	cmd := exec.Command(bin, "init", "--no-agents")
	cmd.Dir = dir
	cmd.Run()
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
		t.Error("AGENTS.md should not exist with --no-agents")
	}
}

func TestCLI_RulesList(t *testing.T) {
	bin := buildBinary(t)
	out, _ := exec.Command(bin, "rules", "list").CombinedOutput()
	s := string(out)
	if !strings.Contains(s, "Go") {
		t.Errorf("expected Go table: %s", s)
	}
}
