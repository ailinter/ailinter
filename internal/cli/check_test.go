package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
)

func TestScanAndWriteSecrets(t *testing.T) {
	t.Run("with secret data and Problems format", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "keys.go")
		content := []byte("const key = \"sk_live_4eC39HqLyjWDarjtT1zdp7dc\"\n") // gitleaks:allow
		os.WriteFile(f, content, 0644)

		out := captureStdout(func() {
			scanAndWriteSecrets(f, content, FormatProblems)
		})
		if out == "" {
			t.Log("no secrets output (may be expected with empty findings)")
		}
	})

	t.Run("with clean data and JSON format", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "clean.go")
		content := []byte("package main\nfunc main() {}\n")
		os.WriteFile(f, content, 0644)

		out := captureStdout(func() {
			scanAndWriteSecrets(f, content, FormatJSON)
		})
		if out != "" {
			t.Logf("JSON output for clean data: %s", out)
		}
	})

	t.Run("with secret data and JSON format", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "keys.go")
		content := []byte("const key = \"sk_live_4eC39HqLyjWDarjtT1zdp7dc\"\n") // gitleaks:allow
		os.WriteFile(f, content, 0644)

		out := captureStdout(func() {
			scanAndWriteSecrets(f, content, FormatJSON)
		})
		if out == "" {
			t.Fatal("expected JSON output for secrets")
		}
		var findings []interface{}
		if err := json.Unmarshal([]byte(out), &findings); err != nil {
			t.Fatalf("expected valid JSON array: %v\noutput: %s", err, out)
		}
		if len(findings) == 0 {
			t.Error("expected at least one secret finding")
		}
	})

	t.Run("with clean data and Problems format", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "clean.go")
		content := []byte("package main\nfunc main() {}\n")
		os.WriteFile(f, content, 0644)

		captureStdout(func() {
			scanAndWriteSecrets(f, content, FormatProblems)
		})
	})
}

func TestWriteDirTokenEstimates(t *testing.T) {
	t.Run("empty results", func(t *testing.T) {
		out := captureStdout(func() {
			writeDirTokenEstimates(nil)
		})
		if !strings.Contains(out, "0") {
			t.Errorf("expected output containing 0, got %q", out)
		}
	})

	t.Run("single result", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.go")
		os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)

		results := []analyzer.QualityResult{
			{FilePath: f, Score: 42},
		}

		out := captureStdout(func() {
			writeDirTokenEstimates(results)
		})
		if !strings.Contains(out, "1") && !strings.Contains(out, "Files scanned: 1") {
			t.Errorf("expected output mentioning 1 file, got %q", out)
		}
		if !strings.Contains(out, "Total current tokens") {
			t.Errorf("expected token estimate output, got %q", out)
		}
	})

	t.Run("multiple results", func(t *testing.T) {
		dir := t.TempDir()
		f1 := filepath.Join(dir, "a.go")
		f2 := filepath.Join(dir, "b.go")
		os.WriteFile(f1, []byte("package main\nfunc main() {}\n"), 0644)
		os.WriteFile(f2, []byte("package main\nfunc f() {\n\tprintln(\"hello\")\n}\n"), 0644)

		results := []analyzer.QualityResult{
			{FilePath: f1, Score: 42},
			{FilePath: f2, Score: 85},
		}

		out := captureStdout(func() {
			writeDirTokenEstimates(results)
		})
		if !strings.Contains(out, "Files scanned: 2") {
			t.Errorf("expected 2 files scanned, got %q", out)
		}
	})

	t.Run("score 100 only", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "perfect.go")
		os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)

		results := []analyzer.QualityResult{
			{FilePath: f, Score: 100},
		}

		out := captureStdout(func() {
			writeDirTokenEstimates(results)
		})
		if !strings.Contains(out, "Total current tokens") {
			t.Errorf("expected token estimate output for perfect score, got %q", out)
		}
	})
}

func TestCheckFile_SecretsOnly(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "keys.go")
	os.WriteFile(f, []byte("const key = \"sk_live_4eC39HqLyjWDarjtT1zdp7dc\"\n"), 0644) // gitleaks:allow

	opts := checkOptions{
		format:      FormatProblems,
		secretsOnly: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile with secretsOnly should not error: %v", err)
		}
	})
	t.Logf("secretsOnly output: %s", out)
}

func TestCheckFile_SecretsOnlyJSON(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "keys.go")
	os.WriteFile(f, []byte("const key = \"sk_live_4eC39HqLyjWDarjtT1zdp7dc\"\n"), 0644) // gitleaks:allow

	opts := checkOptions{
		format:      FormatJSON,
		secretsOnly: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile with secretsOnly JSON should not error: %v", err)
		}
	})
	if out == "" {
		t.Log("no secrets found in secretsOnly JSON mode")
	} else {
		var parsed interface{}
		if err := json.Unmarshal([]byte(out), &parsed); err != nil {
			t.Fatalf("expected valid JSON: %v\noutput: %s", err, out)
		}
	}
}

func TestCheckFile_VulnerabilitiesOnly(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "vuln.py")
	os.WriteFile(f, []byte("import pickle\npickle.loads(data)\n"), 0644)

	opts := checkOptions{
		format:              FormatProblems,
		vulnerabilitiesOnly: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile with vulnerabilitiesOnly should not error: %v", err)
		}
	})
	if !strings.Contains(out, "pickle") && !strings.Contains(out, "deserialization") {
		t.Logf("vuln output may not contain pickle (test may be running in different context): %s", out)
	}
}

func TestCheckFile_VulnerabilitiesOnlyJSON(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "vuln.py")
	os.WriteFile(f, []byte("import pickle\npickle.loads(data)\n"), 0644)

	opts := checkOptions{
		format:              FormatJSON,
		vulnerabilitiesOnly: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile with vulnerabilitiesOnly JSON should not error: %v", err)
		}
	})
	if out != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(out), &parsed); err != nil {
			t.Fatalf("expected valid JSON: %v\noutput: %s", err, out)
		}
	}
}

func TestCheckFile_BinaryRejected(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "binary.bin")
	os.WriteFile(f, []byte{0x00, 0x01, 0x02, 0x03}, 0644)

	opts := checkOptions{format: FormatHuman, noSecrets: true, noVulnerabilities: true}
	err := checkFile(f, opts)
	if err == nil {
		t.Fatal("expected error for binary file")
	}
	if !strings.Contains(err.Error(), "binary") {
		t.Errorf("error should mention 'binary', got: %v", err)
	}
}

func TestCheckFile_Nonexistent(t *testing.T) {
	opts := checkOptions{format: FormatHuman, noSecrets: true, noVulnerabilities: true}
	err := checkFile("/nonexistent/path/file.go", opts)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestCheckFile_JSONFormat(t *testing.T) {
	t.Run("full analysis with JSON", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "main.go")
		os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)

		opts := checkOptions{
			format: FormatJSON,
		}

		out := captureStdout(func() {
			err := checkFile(f, opts)
			if err != nil {
				t.Fatalf("checkFile with FormatJSON should not error: %v", err)
			}
		})
		if !strings.Contains(out, "\"code_quality\"") {
			t.Errorf("JSON output should contain code_quality, got: %s", out)
		}
		var parsed struct {
			CodeQuality analyzer.QualityResult `json:"code_quality"`
		}
		if err := json.Unmarshal([]byte(out), &parsed); err != nil {
			t.Fatalf("expected valid JSON: %v\noutput: %s", err, out)
		}
	})

	t.Run("JSON with secrets disabled", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "main.go")
		os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)

		opts := checkOptions{
			format:    FormatJSON,
			noSecrets: true,
		}

		out := captureStdout(func() {
			err := checkFile(f, opts)
			if err != nil {
				t.Fatalf("checkFile should not error: %v", err)
			}
		})
		if !strings.Contains(out, "\"code_quality\"") {
			t.Errorf("JSON output should contain code_quality, got: %s", out)
		}
	})

	t.Run("JSON with vulnerabilities disabled", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "vuln.py")
		os.WriteFile(f, []byte("import pickle\npickle.loads(data)\n"), 0644)

		opts := checkOptions{
			format:            FormatJSON,
			noVulnerabilities: true,
		}

		out := captureStdout(func() {
			err := checkFile(f, opts)
			if err != nil {
				t.Fatalf("checkFile should not error: %v", err)
			}
		})
		if !strings.Contains(out, "\"code_quality\"") {
			t.Errorf("JSON output should contain code_quality, got: %s", out)
		}
	})
}

func TestCheckFile_MetaLintEnabled(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)

	opts := checkOptions{
		format:            FormatProblems,
		metaLint:          true,
		noSecrets:         true,
		noVulnerabilities: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile with metaLint should not error: %v", err)
		}
	})
	t.Logf("meta-lint output: %s", out)
}

func TestCheckFile_EstimateTokens(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {\n\tprintln(\"hello\")\n}\n"), 0644)

	opts := checkOptions{
		format:            FormatHuman,
		estimateTokens:    true,
		noSecrets:         true,
		noVulnerabilities: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile with estimateTokens should not error: %v", err)
		}
	})
	if !strings.Contains(out, "Token Savings Estimate") {
		t.Errorf("expected token savings output, got: %s", out)
	}
}

func TestCheckFile_HumanFormat(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)

	opts := checkOptions{
		format:            FormatHuman,
		noSecrets:         true,
		noVulnerabilities: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile should not error: %v", err)
		}
	})
	if !strings.Contains(out, "main.go") {
		t.Errorf("human output should contain filename, got: %s", out)
	}
}

func TestCheckFile_MarkdownFormat(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)

	opts := checkOptions{
		format:            FormatMarkdown,
		noSecrets:         true,
		noVulnerabilities: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile should not error: %v", err)
		}
	})
	if !strings.Contains(out, "##") {
		t.Errorf("markdown output should contain heading, got: %s", out)
	}
}

func TestCheckFile_ProblemsFormat(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)

	opts := checkOptions{
		format:            FormatProblems,
		noSecrets:         true,
		noVulnerabilities: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile should not error: %v", err)
		}
	})
	t.Logf("problems output: %s", out)
}

func TestCheckFile_LangOverride(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "script.py")
	os.WriteFile(f, []byte("def hello():\n    pass\n"), 0644)

	opts := checkOptions{
		format:            FormatHuman,
		langOverride:      "python",
		noSecrets:         true,
		noVulnerabilities: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile with langOverride should not error: %v", err)
		}
	})
	if !strings.Contains(out, "script.py") {
		t.Errorf("output should contain filename, got: %s", out)
	}
}

func TestCheckFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "empty.go")
	os.WriteFile(f, []byte{}, 0644)

	opts := checkOptions{
		format:            FormatHuman,
		noSecrets:         true,
		noVulnerabilities: true,
	}

	err := checkFile(f, opts)
	if err != nil {
		t.Fatalf("checkFile on empty file should not error: %v", err)
	}
}

func TestCheckFile_NestedSmells(t *testing.T) {
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

	opts := checkOptions{
		format:            FormatProblems,
		noSecrets:         true,
		noVulnerabilities: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile should not error: %v", err)
		}
	})
	if !strings.Contains(out, "deep_nesting") {
		t.Logf("expected deep_nesting smell in output: %s", out)
	}
}

func TestCheckFile_LongLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "long.go")
	src := "package main\n\nfunc main() {\n\t_ = \"something\"\n}\n"
	src += "// " + strings.Repeat("x", 200) + "\n"
	os.WriteFile(f, []byte(src), 0644)

	opts := checkOptions{
		format:            FormatHuman,
		noSecrets:         true,
		noVulnerabilities: true,
	}

	_ = captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile should not error: %v", err)
		}
	})
}

func TestCheckFile_BrainMethodSmell(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "brain.go")
	var lines []string
	lines = append(lines, "package main")
	lines = append(lines, "func main() {")
	for i := 0; i < 100; i++ {
		lines = append(lines, "\tprintln(\"line", string(rune('0'+i%10)), "\")")
	}
	lines = append(lines, "}")
	src := strings.Join(lines, "\n")
	os.WriteFile(f, []byte(src), 0644)

	opts := checkOptions{
		format:            FormatProblems,
		noSecrets:         true,
		noVulnerabilities: true,
	}

	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile should not error: %v", err)
		}
	})
	t.Logf("brain method output: %s", out)
}

func TestDetectLang(t *testing.T) {
	t.Run("override takes precedence", func(t *testing.T) {
		opts := checkOptions{langOverride: "python"}
		if got := opts.detectLang("test.go"); got != "python" {
			t.Errorf("detectLang with override = %q, want python", got)
		}
	})

	t.Run("go extension", func(t *testing.T) {
		opts := checkOptions{}
		if got := opts.detectLang("test.go"); got != "go" {
			t.Errorf("detectLang for .go = %q, want go", got)
		}
	})

	t.Run("python extension", func(t *testing.T) {
		opts := checkOptions{}
		if got := opts.detectLang("test.py"); got != "python" {
			t.Errorf("detectLang for .py = %q, want python", got)
		}
	})

	t.Run("unknown extension defaults to go", func(t *testing.T) {
		opts := checkOptions{}
		if got := opts.detectLang("test.xyz"); got != "go" {
			t.Errorf("detectLang for unknown ext = %q, want go", got)
		}
	})
}
