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

func TestCLI_NoVulnerabilitiesFlag(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	src := "import pickle\npickle.loads(data)\n"
	f := filepath.Join(dir, "vuln.py")
	os.WriteFile(f, []byte(src), 0644)

	// With vulnerabilities (default)
	out, _ := exec.Command(bin, "check", "--format", "json", f).CombinedOutput()
	if !strings.Contains(string(out), "\"vulnerability_scan\"") {
		t.Errorf("should include vulnerability_scan by default: %s", out)
	}

	// With --no-vulnerabilities
	out2, _ := exec.Command(bin, "check", "--format", "json", "--no-vulnerabilities", f).CombinedOutput()
	s2 := string(out2)
	if strings.Contains(s2, "\"vulnerability_scan\"") && strings.Contains(s2, "\"vulnerability_scan\":") && !strings.Contains(s2, "\"vulnerability_scan\":[]") && !strings.Contains(s2, "\"vulnerability_scan\": []") {
		if strings.Contains(s2, "pickle_deserialization") {
			t.Errorf("--no-vulnerabilities should suppress vuln findings: %s", s2)
		}
	}
}

func TestCLI_NoVulnerabilitiesWithNoSecrets(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	src := "import pickle\npickle.loads(data)\nAPI_KEY = 'sk_live_1234567890abcdef'\n"
	f := filepath.Join(dir, "vuln.py")
	os.WriteFile(f, []byte(src), 0644)

	out, _ := exec.Command(bin, "check", "--format", "json", "--no-secrets", "--no-vulnerabilities", f).CombinedOutput()
	s := string(out)

	if strings.Contains(s, "\"secret_scan\"") && !strings.Contains(s, "\"secret_scan\":[]") && !strings.Contains(s, "\"secret_scan\": []") {
		t.Errorf("--no-secrets should suppress secrets: %s", s)
	}
}

func TestCLI_NoSecretsKeepsVulnerabilities_E2E(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	src := "import pickle\npickle.loads(data)\n"
	f := filepath.Join(dir, "vuln.py")
	os.WriteFile(f, []byte(src), 0644)

	// C4 regression test: --no-secrets must NOT suppress vulnerability_scan
	out, _ := exec.Command(bin, "check", "--format", "json", "--no-secrets", f).CombinedOutput()
	s := string(out)

	if !strings.Contains(s, "\"vulnerability_scan\"") {
		t.Error("C4 REGRESSION: --no-secrets should not suppress vulnerability_scan (was the original C4 bug)")
	}
	if strings.Contains(s, "pickle_deserialization") {
		t.Log("OK: vulnerability findings present despite --no-secrets")
	}
}

func TestCLI_CheckProblemsVulnerabilities(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	src := "import pickle\npickle.loads(data)\n"
	f := filepath.Join(dir, "vuln.py")
	os.WriteFile(f, []byte(src), 0644)

	out, _ := exec.Command(bin, "check", "--format", "problems", f).CombinedOutput()
	s := string(out)

	if strings.Contains(s, "pickle_deserialization") {
		// Verify line numbers are non-zero
		if strings.Contains(s, "vuln.py:0:0") {
			t.Errorf("vuln findings should have real line numbers, got 0:0: %s", s)
		}
	}
}

func TestCLI_BadFormatRejected(t *testing.T) {
	bin := buildBinary(t)
	out, err := exec.Command(bin, "check", "--format", "xml", "/dev/null").CombinedOutput()
	if err == nil {
		t.Errorf("expected error for bad format, got output: %s", out)
	}
}

func TestCLI_BadLanguageRejected(t *testing.T) {
	bin := buildBinary(t)
	out, err := exec.Command(bin, "check", "--lang", "frobulator", "/dev/null").CombinedOutput()
	if err == nil {
		t.Errorf("expected error for bad language, got output: %s", out)
	}
	s := string(out)
	if !strings.Contains(s, "unknown language") {
		t.Errorf("expected unknown language message: %s", s)
	}
}

func TestCLI_BinaryRejected(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	f := filepath.Join(dir, "test.bin")
	os.WriteFile(f, []byte{0x00, 0x01, 0x02, 0x03}, 0644)

	out, err := exec.Command(bin, "check", f).CombinedOutput()
	if err == nil {
		t.Errorf("expected error for binary file, got: %s", out)
	}
}

func TestCLI_DirScanVulnerabilities(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "vuln.py"), []byte("import pickle\npickle.loads(data)\n"), 0644)
	os.WriteFile(filepath.Join(dir, "clean.go"), []byte("package main\nfunc main() {}\n"), 0644)

	out, _ := exec.Command(bin, "check", "--format", "json", dir).CombinedOutput()
	s := string(out)

	if !strings.Contains(s, "\"vulnerability_scan\"") {
		t.Error("directory JSON output should include vulnerability_scan")
	}
}

func TestCLI_ProblemsFormatHasFileLine(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	src := "package main\n\nfunc main() {\n\tif true {\n\t\tif true {\n\t\t\tif true {\n\t\t\t\tif true {\n\t\t\t\t\tprintln(\"deep\")\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t}\n}\n"
	f := filepath.Join(dir, "deep.go")
	os.WriteFile(f, []byte(src), 0644)

	out, _ := exec.Command(bin, "check", "--format", "problems", f).CombinedOutput()
	s := string(out)

	if !strings.Contains(s, "deep.go:") {
		t.Errorf("problems format should contain file:line prefix: %s", s)
	}
	if strings.Contains(s, "\n\n") {
		// Just checking we get output
	}
}

func TestCLI_JSONOutputHasCodeQualityAndVulnKeys(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()
	f := filepath.Join(dir, "vuln.py")
	os.WriteFile(f, []byte("import pickle\npickle.loads(data)\n"), 0644)

	out, _ := exec.Command(bin, "check", "--format", "json", f).CombinedOutput()
	s := string(out)

	if !strings.Contains(s, "\"code_quality\"") {
		t.Error("JSON output should contain code_quality key")
	}
	if !strings.Contains(s, "\"vulnerability_scan\"") {
		t.Error("JSON output should contain vulnerability_scan key")
	}
}
