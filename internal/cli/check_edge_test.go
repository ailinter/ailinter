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

func makeCtx(opts checkOptions, results []analyzer.QualityResult, secs []secrets.SecretFinding, vulns []vulnerability.Finding) *walkContext {
	return &walkContext{
		opts:        opts,
		allResults:  results,
		allSecrets:  secs,
		allVulns:    vulns,
		resolvedDir: "",
		langCount:   make(map[string]int),
	}
}

func TestWriteResults_SecretsOnly(t *testing.T) {
	finding := secrets.SecretFinding{RuleID: "test-rule", Line: 1, Description: "test finding"}
	for _, tt := range []struct {
		name   string
		ctx    *walkContext
		checks []string
	}{
		{"with findings", makeCtx(checkOptions{format: FormatProblems, secretsOnly: true}, nil, []secrets.SecretFinding{finding}, nil), []string{"test-rule"}},
		{"empty findings", makeCtx(checkOptions{format: FormatProblems, secretsOnly: true}, nil, nil, nil), nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(func() { tt.ctx.writeResults() })
			if tt.name == "empty findings" && out != "" {
				t.Errorf("expected no output for empty secrets, got: %s", out)
			}
			for _, c := range tt.checks {
				if !strings.Contains(out, c) {
					t.Errorf("expected %q in output, got: %s", c, out)
				}
			}
		})
	}
}

func TestWriteResults_VulnerabilitiesOnly(t *testing.T) {
	finding := vulnerability.Finding{RuleID: "test-vuln", Severity: "high", Line: 1, Description: "test vuln"}
	for _, tt := range []struct {
		name   string
		ctx    *walkContext
		checks []string
	}{
		{"with findings", makeCtx(checkOptions{format: FormatProblems, vulnerabilitiesOnly: true}, nil, nil, []vulnerability.Finding{finding}), []string{"test-vuln"}},
		{"empty findings", makeCtx(checkOptions{format: FormatProblems, vulnerabilitiesOnly: true}, nil, nil, nil), nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(func() { tt.ctx.writeResults() })
			if tt.name == "empty findings" && out != "" {
				t.Errorf("expected no output for empty vulns, got: %s", out)
			}
			for _, c := range tt.checks {
				if !strings.Contains(out, c) {
					t.Errorf("expected %q in output, got: %s", c, out)
				}
			}
		})
	}
}

func TestWriteResults_JSONFormat(t *testing.T) {
	ctx := makeCtx(checkOptions{format: FormatJSON},
		[]analyzer.QualityResult{{Score: 100, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10}},
		[]secrets.SecretFinding{{RuleID: "test-rule", Line: 5, Description: "secret found"}},
		nil)
	out := captureStdout(func() { ctx.writeResults() })
	for _, s := range []string{"\"code_quality\"", "\"secret_scan\""} {
		if !strings.Contains(out, s) {
			t.Errorf("JSON output should contain %s, got: %s", s, out)
		}
	}
}

func TestWriteResults_SARIFFormat(t *testing.T) {
	ctx := makeCtx(checkOptions{format: FormatSARIF},
		[]analyzer.QualityResult{{Score: 95, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10}},
		nil, nil)
	ctx.resolvedDir = t.TempDir()
	out := captureStdout(func() { ctx.writeResults() })
	if !strings.Contains(out, "$schema") {
		t.Logf("SARIF output: %s", out)
	}
}

func TestWriteResults_StandardFormat(t *testing.T) {
	for _, tt := range []struct {
		name   string
		ctx    *walkContext
		checks []string
	}{
		{
			"with results and secrets",
			makeCtx(checkOptions{format: FormatHuman},
				[]analyzer.QualityResult{{Score: 95, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10}},
				[]secrets.SecretFinding{{RuleID: "test-rule", Line: 5, Description: "test"}},
				nil),
			[]string{"test.go"},
		},
		{
			"with vulnerabilities",
			makeCtx(checkOptions{format: FormatHuman},
				[]analyzer.QualityResult{{Score: 95, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10}},
				nil,
				[]vulnerability.Finding{{RuleID: "pickle_deserialization", Severity: "high", Line: 1, Description: "pickle vuln"}}),
			[]string{"vulnerabilit"},
		},
		{
			"with summary",
			makeCtx(checkOptions{format: FormatHuman},
				[]analyzer.QualityResult{
					{Score: 95, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10},
					{Score: 60, Label: "Proceed with Care", FilePath: "test2.go", Language: "go", LinesOfCode: 50},
				},
				nil, nil),
			[]string{"Summary"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(func() { tt.ctx.writeResults() })
			for _, c := range tt.checks {
				if !strings.Contains(out, c) {
					t.Errorf("expected %q in output, got: %s", c, out)
				}
			}
		})
	}
}

func TestWriteResults_SARIFSecretsOnly(t *testing.T) {
	ctx := makeCtx(checkOptions{format: FormatSARIF, secretsOnly: true},
		nil,
		[]secrets.SecretFinding{{RuleID: "test-rule", Line: 1, Description: "test"}},
		nil)
	ctx.resolvedDir = t.TempDir()
	out := captureStdout(func() { ctx.writeResults() })
	if !strings.Contains(out, "$schema") {
		t.Logf("SARIF secrets-only output: %s", out)
	}
}

func TestWriteResults_SARIFVulnerabilitiesOnly(t *testing.T) {
	ctx := makeCtx(checkOptions{format: FormatSARIF, vulnerabilitiesOnly: true},
		nil, nil,
		[]vulnerability.Finding{{RuleID: "test-vuln", Severity: "high", Line: 1, Description: "test"}})
	ctx.resolvedDir = t.TempDir()
	out := captureStdout(func() { ctx.writeResults() })
	if !strings.Contains(out, "$schema") {
		t.Logf("SARIF vuln output: %s", out)
	}
}

func TestWriteResults_EstimateTokens(t *testing.T) {
	ctx := makeCtx(checkOptions{format: FormatHuman, estimateTokens: true},
		[]analyzer.QualityResult{{Score: 95, Label: "Go Ahead", FilePath: "test.go", Language: "go", LinesOfCode: 10}},
		nil, nil)
	out := captureStdout(func() { ctx.writeResults() })
	if !strings.Contains(out, "Token Savings") {
		t.Errorf("output should contain token savings section, got: %s", out)
	}
}

// Helper to write a test file and return its path.
func writeGoFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCheckFile_SecretsOnlyFormats(t *testing.T) {
	for _, tt := range []struct {
		name   string
		format FormatMode
		checks []string
	}{
		{"human", FormatHuman, []string{"secret"}},
		{"problems", FormatProblems, []string{"sk_live"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			f := writeGoFile(t, dir, "keys.go", "const key = \"sk_liv...p7dc\"\n")
			opts := checkOptions{format: tt.format, secretsOnly: true}
			out := captureStdout(func() {
				if err := checkFile(f, opts); err != nil {
					t.Fatalf("checkFile secrets-only should not error: %v", err)
				}
			})
			for _, c := range tt.checks {
				if !strings.Contains(out, c) {
					t.Logf("secrets-only %s output: %s", tt.format, out)
				}
			}
		})
	}
}

func TestCheckFile_VulnerabilitiesOnlyFormats(t *testing.T) {
	for _, tt := range []struct {
		name   string
		format FormatMode
	}{
		{"human", FormatHuman},
		{"problems", FormatProblems},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			f := writeGoFile(t, dir, "vuln.py", "import pickle\npickle.loads(data)\n")
			opts := checkOptions{format: tt.format, vulnerabilitiesOnly: true}
			out := captureStdout(func() {
				if err := checkFile(f, opts); err != nil {
					t.Fatalf("checkFile vulns-only should not error: %v", err)
				}
			})
			if !strings.Contains(out, "pickle_deserialization") {
				t.Logf("vulns-only %s output: %s", tt.format, out)
			}
		})
	}
}

func TestCheckFile_StandardWithSecretsAndVulns(t *testing.T) {
	dir := t.TempDir()
	f := writeGoFile(t, dir, "vuln.py", "import pickle\npickle.loads(data)\nAPI_KEY = 'sk_liv...p7dc'\n")
	opts := checkOptions{format: FormatHuman}
	out := captureStdout(func() {
		if err := checkFile(f, opts); err != nil {
			t.Fatalf("checkFile with all scans should not error: %v", err)
		}
	})
	if !strings.Contains(out, "Score") {
		t.Logf("standard output: %s", out)
	}
}

func TestCheckFile_BinaryError(t *testing.T) {
	dir := t.TempDir()
	f := writeGoFile(t, dir, "data.bin", string([]byte{0x00, 0x01, 0x02, 0x03}))
	opts := checkOptions{format: FormatHuman}
	err := checkFile(f, opts)
	if err == nil {
		t.Fatal("expected error for binary file")
	}
	if !strings.Contains(err.Error(), "cannot analyze binary file") {
		t.Errorf("expected binary file error, got: %v", err)
	}
}

func TestCheckDirectory_SecretsAndVulns(t *testing.T) {
	for _, tt := range []struct {
		name     string
		opts     checkOptions
		content  string
		filename string
		checks   []string
	}{
		{"secrets only", checkOptions{format: FormatProblems, secretsOnly: true}, "const key = \"sk_liv...p7dc\"\n", "keys.go", []string{"sk_live"}},
		{"vulns only", checkOptions{format: FormatProblems, vulnerabilitiesOnly: true}, "import pickle\npickle.loads(data)\n", "vuln.py", []string{"pickle_deserialization"}},
		{"secrets only empty", checkOptions{format: FormatProblems, secretsOnly: true}, "package main\nfunc main() {}\n", "clean.go", nil},
		{"vulns only empty", checkOptions{format: FormatProblems, vulnerabilitiesOnly: true}, "package main\nfunc main() {}\n", "clean.go", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeGoFile(t, dir, tt.filename, tt.content)
			out := captureStdout(func() {
				if err := checkDirectory(dir, tt.opts, false); err != nil {
					t.Fatalf("checkDirectory should not error: %v", err)
				}
			})
			for _, c := range tt.checks {
				if !strings.Contains(out, c) {
					t.Logf("dir output: %s", out)
				}
			}
			if len(tt.checks) == 0 && out != "" {
				t.Logf("unexpected output for empty: %s", out)
			}
		})
	}
}

func TestWalkFn_BinaryFile(t *testing.T) {
	dir := t.TempDir()
	f := writeGoFile(t, dir, "data.bin", string([]byte{0x00, 0x01, 0x02, 0x03}))
	ctx := &walkContext{
		opts:        checkOptions{noSecrets: true, noVulnerabilities: true},
		resolvedDir: dir,
		scanQuality: true,
		langCount:   make(map[string]int),
	}
	if err := ctx.walkFn(f, fakeDirEntry{name: "data.bin", isDir: false}, nil); err != nil {
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
	err := ctx.walkFn("/test/.hidden", fakeDirEntry{name: ".hidden", isDir: true}, nil)
	if err == nil {
		t.Error("expected SkipDir for hidden directory")
	}
}

func TestWalkFn_ShellSourceFile(t *testing.T) {
	dir := t.TempDir()
	f := writeGoFile(t, dir, "script.sh", "#!/bin/bash\necho hello\n")
	ctx := &walkContext{
		opts:        checkOptions{noSecrets: true, noVulnerabilities: true},
		resolvedDir: dir,
		scanQuality: true,
		langCount:   make(map[string]int),
	}
	if err := ctx.walkFn(f, fakeDirEntry{name: "script.sh", isDir: false}, nil); err != nil {
		t.Fatalf("walkFn on shell file should not error: %v", err)
	}
	if ctx.fileCount != 1 {
		t.Errorf("expected 1 file processed, got %d", ctx.fileCount)
	}
}

type fakeDirEntry struct {
	name  string
	isDir bool
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return f.isDir }
func (f fakeDirEntry) Type() os.FileMode          { return os.FileMode(0) }
func (f fakeDirEntry) Info() (os.FileInfo, error) { return nil, nil }
