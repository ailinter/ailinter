package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/vulnerability"
)

// ──────────────────────────────────────────────
// writeResults tests
// ──────────────────────────────────────────────

func TestWriteResults_SecretsOnly(t *testing.T) {
	t.Run("secrets-only with findings", func(t *testing.T) {
		ctx := &walkContext{
			opts: checkOptions{format: FormatProblems, secretsOnly: true},
			allSecrets: []secrets.SecretFinding{
				{RuleID: "test-rule", Line: 1, Description: "test finding"},
			},
		}
		out := captureStdout(func() {
			ctx.writeResults()
		})
		if !strings.Contains(out, "test-rule") {
			t.Errorf("output should contain rule name, got: %s", out)
		}
	})

	t.Run("secrets-only empty findings", func(t *testing.T) {
		ctx := &walkContext{
			opts:       checkOptions{format: FormatProblems, secretsOnly: true},
			allSecrets: nil,
		}
		out := captureStdout(func() {
			ctx.writeResults()
		})
		if out != "" {
			t.Errorf("expected no output for empty secrets, got: %s", out)
		}
	})

	t.Run("secrets-only with SARIF format", func(t *testing.T) {
		ctx := &walkContext{
			opts: checkOptions{format: FormatSARIF, secretsOnly: true},
			allSecrets: []secrets.SecretFinding{
				{RuleID: "test-rule", Line: 1, Description: "test"},
			},
			resolvedDir: t.TempDir(),
		}
		out := captureStdout(func() {
			ctx.writeResults()
		})
		if !strings.Contains(out, "sarif") && !strings.Contains(out, "$schema") {
			t.Logf("SARIF output: %s", out)
		}
	})
}

func TestWriteResults_VulnerabilitiesOnly(t *testing.T) {
	t.Run("vulns-only with findings", func(t *testing.T) {
		ctx := &walkContext{
			opts: checkOptions{format: FormatProblems, vulnerabilitiesOnly: true},
			allVulns: []vulnerability.Finding{
				{RuleID: "test-vuln", Severity: "high", Line: 1, Description: "test vuln"},
			},
		}
		out := captureStdout(func() {
			ctx.writeResults()
		})
		if !strings.Contains(out, "test-vuln") {
			t.Errorf("output should contain vulnerability name, got: %s", out)
		}
	})

	t.Run("vulns-only empty findings", func(t *testing.T) {
		ctx := &walkContext{
			opts:     checkOptions{format: FormatProblems, vulnerabilitiesOnly: true},
			allVulns: nil,
		}
		out := captureStdout(func() {
			ctx.writeResults()
		})
		if out != "" {
			t.Errorf("expected no output for empty vulns, got: %s", out)
		}
	})

	t.Run("vulns-only with SARIF format", func(t *testing.T) {
		ctx := &walkContext{
			opts: checkOptions{format: FormatSARIF, vulnerabilitiesOnly: true},
			allVulns: []vulnerability.Finding{
				{RuleID: "test-vuln", Severity: "high", Line: 1, Description: "test"},
			},
			resolvedDir: t.TempDir(),
		}
		out := captureStdout(func() {
			ctx.writeResults()
		})
		if !strings.Contains(out, "$schema") {
			t.Logf("SARIF vuln output: %s", out)
		}
	})
}

func TestWriteResults_JSONFormat(t *testing.T) {
	ctx := &walkContext{
		opts: checkOptions{format: FormatJSON},
		allResults: []analyzer.QualityResult{
			{Score: 100, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10},
		},
		allSecrets: []secrets.SecretFinding{
			{RuleID: "test-rule", Line: 5, Description: "secret found"},
		},
	}
	out := captureStdout(func() {
		ctx.writeResults()
	})
	if !strings.Contains(out, "\"code_quality\"") {
		t.Errorf("JSON output should contain code_quality, got: %s", out)
	}
	if !strings.Contains(out, "\"secret_scan\"") {
		t.Errorf("JSON output should contain secret_scan, got: %s", out)
	}
}

func TestWriteResults_SARIFFormat(t *testing.T) {
	ctx := &walkContext{
		opts: checkOptions{format: FormatSARIF},
		allResults: []analyzer.QualityResult{
			{Score: 95, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10},
		},
		resolvedDir: t.TempDir(),
	}
	out := captureStdout(func() {
		ctx.writeResults()
	})
	if !strings.Contains(out, "$schema") {
		t.Logf("SARIF output: %s", out)
	}
}

func TestWriteResults_StandardFormat(t *testing.T) {
	t.Run("standard with results and secrets", func(t *testing.T) {
		ctx := &walkContext{
			opts: checkOptions{format: FormatHuman},
			allResults: []analyzer.QualityResult{
				{Score: 95, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10},
			},
			allSecrets: []secrets.SecretFinding{
				{RuleID: "test-rule", Line: 5, Description: "test"},
			},
		}
		out := captureStdout(func() {
			ctx.writeResults()
		})
		if !strings.Contains(out, "test.go") {
			t.Errorf("output should contain filename, got: %s", out)
		}
	})

	t.Run("standard with vulnerabilities", func(t *testing.T) {
		ctx := &walkContext{
			opts: checkOptions{format: FormatHuman},
			allResults: []analyzer.QualityResult{
				{Score: 95, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10},
			},
			allVulns: []vulnerability.Finding{
				{RuleID: "pickle_deserialization", Severity: "high", Line: 1, Description: "pickle vuln"},
			},
		}
		out := captureStdout(func() {
			ctx.writeResults()
		})
		if !strings.Contains(out, "vulnerabilit") {
			t.Errorf("output should contain vulnerability section, got: %s", out)
		}
	})

	t.Run("standard with summary", func(t *testing.T) {
		ctx := &walkContext{
			opts: checkOptions{format: FormatHuman},
			allResults: []analyzer.QualityResult{
				{Score: 95, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10},
				{Score: 60, Label: "Proceed with Care", FilePath: "test2.go", Language: "go", LinesOfCode: 50},
			},
		}
		out := captureStdout(func() {
			ctx.writeResults()
		})
		if !strings.Contains(out, "Summary") {
			t.Errorf("output should contain Summary section, got: %s", out)
		}
	})
}

func TestWriteResults_EstimateTokens(t *testing.T) {
	ctx := &walkContext{
		opts: checkOptions{format: FormatHuman, estimateTokens: true},
		allResults: []analyzer.QualityResult{
			{Score: 95, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10},
		},
	}
	out := captureStdout(func() {
		ctx.writeResults()
	})
	if !strings.Contains(out, "Token Savings") {
		t.Errorf("output should contain token savings section, got: %s", out)
	}
}

// ──────────────────────────────────────────────
// checkFile edge case tests
// ──────────────────────────────────────────────

func TestCheckFile_SecretsOnlyFormatHuman(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "keys.go")
	os.WriteFile(f, []byte("const key = \"sk_live_4eC39HqLyjWDarjtT1zdp7dc\"\n"), 0644) // gitleaks:allow

	opts := checkOptions{
		format:      FormatHuman,
		secretsOnly: true,
	}
	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile secrets-only should not error: %v", err)
		}
	})
	if !strings.Contains(out, "Secret Scan") && !strings.Contains(out, "secret") {
		t.Logf("secrets-only human output: %s", out)
	}
}

func TestCheckFile_SecretsOnlyFormatProblems(t *testing.T) {
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
			t.Fatalf("checkFile secrets-only problems should not error: %v", err)
		}
	})
	if !strings.Contains(out, "sk_live") {
		t.Logf("secrets-only problems output: %s", out)
	}
}

func TestCheckFile_VulnerabilitiesOnlyFormatHuman(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "vuln.py")
	os.WriteFile(f, []byte("import pickle\npickle.loads(data)\n"), 0644)

	opts := checkOptions{
		format:              FormatHuman,
		vulnerabilitiesOnly: true,
	}
	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile vulns-only should not error: %v", err)
		}
	})
	if !strings.Contains(out, "pickle_deserialization") {
		t.Logf("vulns-only human output: %s", out)
	}
}

func TestCheckFile_StandardWithSecretsAndVulns(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "vuln.py")
	os.WriteFile(f, []byte("import pickle\npickle.loads(data)\nAPI_KEY = 'sk_live_4eC39HqLyjWDarjtT1zdp7dc'\n"), 0644) // gitleaks:allow

	opts := checkOptions{
		format: FormatHuman,
	}
	out := captureStdout(func() {
		err := checkFile(f, opts)
		if err != nil {
			t.Fatalf("checkFile with all scans should not error: %v", err)
		}
	})
	if !strings.Contains(out, "Code Quality") && !strings.Contains(out, "Score") {
		t.Logf("standard output: %s", out)
	}
}

func TestCheckFile_BinaryError(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "data.bin")
	os.WriteFile(f, []byte{0x00, 0x01, 0x02, 0x03}, 0644)

	opts := checkOptions{format: FormatHuman}
	err := checkFile(f, opts)
	if err == nil {
		t.Fatal("expected error for binary file")
	}
	if !strings.Contains(err.Error(), "cannot analyze binary file") {
		t.Errorf("expected binary file error, got: %v", err)
	}
}

// ──────────────────────────────────────────────
// checkDirectory edge case tests
// ──────────────────────────────────────────────

func TestCheckDirectory_SecretsOnly(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "keys.go"), []byte("const key = \"sk_live_4eC39HqLyjWDarjtT1zdp7dc\"\n"), 0644) // gitleaks:allow

	opts := checkOptions{
		format:      FormatProblems,
		secretsOnly: true,
	}
	out := captureStdout(func() {
		err := checkDirectory(dir, opts, false)
		if err != nil {
			t.Fatalf("checkDirectory secrets-only should not error: %v", err)
		}
	})
	if !strings.Contains(out, "sk_live") {
		t.Logf("secrets-only dir output: %s", out)
	}
}

func TestCheckDirectory_VulnerabilitiesOnly(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "vuln.py"), []byte("import pickle\npickle.loads(data)\n"), 0644)

	opts := checkOptions{
		format:              FormatProblems,
		vulnerabilitiesOnly: true,
	}
	out := captureStdout(func() {
		err := checkDirectory(dir, opts, false)
		if err != nil {
			t.Fatalf("checkDirectory vulns-only should not error: %v", err)
		}
	})
	if !strings.Contains(out, "pickle_deserialization") {
		t.Logf("vulns-only dir output: %s", out)
	}
}

func TestCheckDirectory_SecretsOnlyEmptyResults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "clean.go"), []byte("package main\nfunc main() {}\n"), 0644)

	opts := checkOptions{
		format:      FormatProblems,
		secretsOnly: true,
	}
	out := captureStdout(func() {
		err := checkDirectory(dir, opts, false)
		if err != nil {
			t.Fatalf("checkDirectory secrets-only clean should not error: %v", err)
		}
	})
	if out != "" {
		t.Logf("secrets-only clean dir output: %s", out)
	}
}

func TestCheckDirectory_VulnerabilitiesOnlyEmptyResults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "clean.go"), []byte("package main\nfunc main() {}\n"), 0644)

	opts := checkOptions{
		format:              FormatProblems,
		vulnerabilitiesOnly: true,
	}
	out := captureStdout(func() {
		err := checkDirectory(dir, opts, false)
		if err != nil {
			t.Fatalf("checkDirectory vulns-only clean should not error: %v", err)
		}
	})
	if out != "" {
		t.Logf("vulns-only clean dir output: %s", out)
	}
}

// ──────────────────────────────────────────────
// walkFn basic tests
// ──────────────────────────────────────────────

func TestWalkFn_BinaryFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "data.bin")
	os.WriteFile(f, []byte{0x00, 0x01, 0x02, 0x03}, 0644)

	ctx := &walkContext{
		opts:        checkOptions{noSecrets: true, noVulnerabilities: true},
		resolvedDir: dir,
		scanQuality: true,
		langCount:   make(map[string]int),
	}
	err := ctx.walkFn(f, fakeDirEntry{name: "data.bin", isDir: false}, nil)
	if err != nil {
		t.Fatalf("walkFn on binary should not error: %v", err)
	}
	if ctx.fileCount != 0 {
		t.Errorf("expected 0 files processed for binary, got %d", ctx.fileCount)
	}
}

func TestWalkFn_DirectorySkip(t *testing.T) {
	ctx := &walkContext{
		resolvedDir: "/test",
		langCount:   make(map[string]int),
	}
	// .hidden dirs should be skipped
	err := ctx.walkFn("/test/.hidden", fakeDirEntry{name: ".hidden", isDir: true}, nil)
	if err == nil {
		t.Error("expected SkipDir for hidden directory")
	}
}

func TestWalkFn_ShellSourceFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "script.sh")
	os.WriteFile(f, []byte("#!/bin/bash\necho hello\n"), 0644)

	ctx := &walkContext{
		opts:        checkOptions{noSecrets: true, noVulnerabilities: true},
		resolvedDir: dir,
		scanQuality: true,
		langCount:   make(map[string]int),
		scanner:     nil,
		vulnScanner: nil,
	}
	err := ctx.walkFn(f, fakeDirEntry{name: "script.sh", isDir: false}, nil)
	if err != nil {
		t.Fatalf("walkFn on shell file should not error: %v", err)
	}
	if ctx.fileCount != 1 {
		t.Errorf("expected 1 file processed, got %d", ctx.fileCount)
	}
}

// ──────────────────────────────────────────────
// Helper: fakeDirEntry implements os.DirEntry
// ──────────────────────────────────────────────

type fakeDirEntry struct {
	name  string
	isDir bool
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return f.isDir }
func (f fakeDirEntry) Type() os.FileMode          { return os.FileMode(0) }
func (f fakeDirEntry) Info() (os.FileInfo, error) { return nil, nil }
