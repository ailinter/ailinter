package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/metalinter"
)

func TestMCPCommand_Details(t *testing.T) {
	cmd := MCPCommand("1.2.3")
	if cmd == nil {
		t.Fatal("MCPCommand returned nil")
	}
	if cmd.Use != "mcp" {
		t.Errorf("Use = %q, want mcp", cmd.Use)
	}
	if cmd.Short != "Start ailinter as an MCP server (stdio)" {
		t.Errorf("Short = %q, want expected", cmd.Short)
	}
	if !strings.Contains(cmd.Long, "Model Context Protocol") {
		t.Error("Long should describe MCP")
	}
	if cmd.Args != nil {
		t.Error("MCPCommand should not have Args validator")
	}
	if cmd.RunE == nil {
		t.Error("MCPCommand should have RunE")
	}
}

func TestMCPCommand_StdioFlag(t *testing.T) {
	cmd := MCPCommand("1.2.3")
	flag := cmd.Flags().Lookup("stdio")
	if flag != nil {
		t.Error("MCPCommand should not have --stdio flag (it's a subcommand, not arg)")
	}
}

func TestReportCommand_Details(t *testing.T) {
	cmd := ReportCommand()
	if cmd == nil {
		t.Fatal("ReportCommand returned nil")
	}
	if cmd.Use != "report <file|dir>" {
		t.Errorf("Use = %q, want report <file|dir>", cmd.Use)
	}
	if cmd.Short != "Generate a CODE_QUALITY.md report" {
		t.Errorf("Short = %q, want expected", cmd.Short)
	}
}

func TestReportCommand_ExactArgs(t *testing.T) {
	cmd := ReportCommand()
	if cmd.Args == nil {
		t.Fatal("ReportCommand should have Args validator")
	}
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error for 0 args")
	}
	if err := cmd.Args(cmd, []string{"a", "b"}); err == nil {
		t.Error("expected error for 2 args")
	}
	if err := cmd.Args(cmd, []string{"a"}); err != nil {
		t.Errorf("expected no error for 1 arg, got: %v", err)
	}
}

func TestReportCommand_OutputFlag(t *testing.T) {
	cmd := ReportCommand()
	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("--output flag not found")
	}
	if flag.DefValue != "CODE_QUALITY.md" {
		t.Errorf("default = %q, want CODE_QUALITY.md", flag.DefValue)
	}
}

func TestExecuteReport_InvalidPath(t *testing.T) {
	err := executeReport("/nonexistent/path/test.go", filepath.Join(t.TempDir(), "report.md"))
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
	if !strings.Contains(err.Error(), "report generation failed") {
		t.Errorf("expected error to mention 'report generation failed', got: %v", err)
	}
}

func TestExecuteReport_ValidPath(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)
	outputPath := filepath.Join(dir, "report.md")
	err := executeReport(f, outputPath)
	if err != nil {
		t.Fatalf("executeReport should not error: %v", err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("report file should exist: %v", err)
	}
}

func TestWriteJSONMetaLint(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "test.go", Line: 10, Column: 5, Message: "unused variable", Category: "bug"},
		{Tool: "gofmt", Code: "formatting", Severity: "info", File: "test.go", Line: 1, Column: 0, Message: "file is not gofmt-ed", Category: "formatting"},
	}
	out := captureStdout(func() { writeJSONMetaLint(findings) })
	var parsed struct {
		MetaLint []metalinter.Finding `json:"meta_lint"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("expected valid JSON: %v\noutput: %s", err, out)
	}
	if len(parsed.MetaLint) != 2 {
		t.Errorf("got %d findings, want 2", len(parsed.MetaLint))
	}
	if parsed.MetaLint[0].Tool != "govet" {
		t.Errorf("first tool = %q, want govet", parsed.MetaLint[0].Tool)
	}
}

func TestWriteJSONMetaLint_Empty(t *testing.T) {
	out := captureStdout(func() { writeJSONMetaLint(nil) })
	var parsed struct {
		MetaLint []metalinter.Finding `json:"meta_lint"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("expected valid JSON: %v\noutput: %s", err, out)
	}
	if len(parsed.MetaLint) != 0 {
		t.Errorf("got %d findings, want 0", len(parsed.MetaLint))
	}
}

func metaLintFinding(tool, code, sev, file string, line, col int, msg, cat string) metalinter.Finding {
	return metalinter.Finding{Tool: tool, Code: code, Severity: sev, File: file, Line: line, Column: col, Message: msg, Category: cat}
}

func TestWriteMetaLintFindings(t *testing.T) {
	for _, tt := range []struct {
		name     string
		format   FormatMode
		findings []metalinter.Finding
		checks   []string // substrings to look for
	}{
		{"json", FormatJSON, []metalinter.Finding{metaLintFinding("staticcheck", "S1017", "warning", "test.go", 5, 1, "should use strings.Replace", "style")}, []string{"meta_lint"}},
		{"problems", FormatProblems, []metalinter.Finding{metaLintFinding("govet", "SA1000", "warning", "test.go", 10, 5, "unused variable", "bug")}, []string{"test.go:10:5", "govet"}},
		{"markdown", FormatMarkdown, []metalinter.Finding{metaLintFinding("gofmt", "formatting", "info", "test.go", 1, 0, "file is not gofmt-ed", "formatting")}, []string{"Meta-Lint Findings", "gofmt", "| 1 |"}},
		{"human", FormatHuman, []metalinter.Finding{metaLintFinding("govet", "SA1000", "warning", "test.go", 10, 5, "unused variable", "bug")}, []string{"meta-lint", "govet"}},
		{"auto (default human)", FormatAuto, []metalinter.Finding{metaLintFinding("misspell", "spelling", "warning", "test.go", 3, 1, "typo", "style")}, []string{"meta-lint"}},
		{"empty json", FormatJSON, nil, []string{}},
		{"empty human", FormatHuman, nil, []string{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(func() { writeMetaLintFindings(tt.format, tt.findings) })
			if tt.format == FormatJSON && tt.findings == nil && out != "" {
				t.Errorf("expected no output for empty json, got: %s", out)
			}
			if tt.format == FormatHuman && tt.findings == nil && out != "" {
				t.Errorf("expected no output for empty human, got: %s", out)
			}
			for _, c := range tt.checks {
				if !strings.Contains(out, c) {
					t.Errorf("expected %q in output, got: %s", c, out)
				}
			}
		})
	}
}

func TestWriteMarkdownMetaLint(t *testing.T) {
	for _, tt := range []struct {
		name     string
		findings []metalinter.Finding
		checks   []string
		antiChecks []string
	}{
		{
			"empty",
			nil,
			[]string{"Meta-Lint Findings (0)"},
			nil,
		},
		{
			"with findings",
			[]metalinter.Finding{
				metaLintFinding("govet", "SA1000", "warning", "src/main.go", 42, 5, "unused variable", "bug"),
				metaLintFinding("gofmt", "formatting", "info", "src/main.go", 1, 0, "file is not gofmt-ed", "formatting"),
			},
			[]string{"Meta-Lint Findings (2)", "govet", "gofmt", "| govet | SA1000 | warning | src/main.go | 42:5 |"},
			nil,
		},
		{
			"column format",
			[]metalinter.Finding{metaLintFinding("govet", "SA1000", "warning", "test.go", 10, 3, "message", "bug")},
			[]string{"| 10:3 |"},
			nil,
		},
		{
			"line only",
			[]metalinter.Finding{metaLintFinding("govet", "SA1000", "warning", "test.go", 10, 0, "message", "bug")},
			[]string{"| 10 |"},
			[]string{"| 10:0 |"},
		},
		{
			"line zero",
			[]metalinter.Finding{metaLintFinding("govet", "SA1000", "warning", "test.go", 0, 0, "message", "bug")},
			[]string{"| test.go |  |"},
			nil,
		},
		{
			"pipe in message",
			[]metalinter.Finding{metaLintFinding("govet", "SA1000", "warning", "test.go", 1, 0, "msg with | pipe", "bug")},
			[]string{"msg with"},
			nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(func() { writeMarkdownMetaLint(tt.findings) })
			for _, c := range tt.checks {
				if !strings.Contains(out, c) {
					t.Errorf("expected %q in output, got: %s", c, out)
				}
			}
			for _, a := range tt.antiChecks {
				if strings.Contains(out, a) {
					t.Errorf("expected NOT to find %q in output, got: %s", a, out)
				}
			}
		})
	}
}

func TestWriteHumanMetaLint(t *testing.T) {
	for _, tt := range []struct {
		name     string
		findings []metalinter.Finding
		checks   []string
		antiChecks []string
	}{
		{
			"empty",
			nil,
			[]string{"meta-lint"},
			nil,
		},
		{
			"with findings",
			[]metalinter.Finding{
				metaLintFinding("govet", "SA1000", "warning", "src/main.go", 42, 5, "unused variable x", "bug"),
				metaLintFinding("gofmt", "formatting", "info", "src/main.go", 1, 0, "file is not gofmt-ed", "formatting"),
			},
			[]string{"meta-lint", "govet: 1 finding", "gofmt: 1 finding", "unused variable", "L42"},
			nil,
		},
		{
			"code column",
			[]metalinter.Finding{metaLintFinding("govet", "SA1000", "warning", "test.go", 10, 0, "unused", "bug")},
			[]string{"[SA1000]"},
			nil,
		},
		{
			"file and location",
			[]metalinter.Finding{metaLintFinding("govet", "SA1000", "warning", "src/main.go", 10, 3, "unused", "bug")},
			[]string{"src/main.go", "L10:3"},
			nil,
		},
		{
			"no file",
			[]metalinter.Finding{metaLintFinding("govet", "SA1000", "warning", "", 10, 0, "unused", "bug")},
			nil,
			[]string{"└"},
		},
		{
			"no code",
			[]metalinter.Finding{metaLintFinding("govet", "", "warning", "test.go", 10, 0, "unused", "bug")},
			nil,
			[]string{"[]"},
		},
		{
			"column line format",
			[]metalinter.Finding{metaLintFinding("govet", "SA1000", "warning", "test.go", 10, 3, "unused", "bug")},
			[]string{"L10:3"},
			nil,
		},
		{
			"line no column",
			[]metalinter.Finding{metaLintFinding("gofmt", "formatting", "info", "test.go", 1, 0, "formatting", "formatting")},
			[]string{" L1"},
			[]string{"L1:0"},
		},
		{
			"multiple tools",
			[]metalinter.Finding{
				metaLintFinding("govet", "SA1000", "warning", "a.go", 1, 0, "unused", "bug"),
				metaLintFinding("staticcheck", "SA1000", "warning", "b.go", 2, 0, "ineffective assign", "bug"),
				metaLintFinding("gofmt", "formatting", "info", "c.go", 0, 0, "formatting", "formatting"),
				metaLintFinding("misspell", "spelling", "warning", "d.go", 3, 0, "typo", "style"),
				metaLintFinding("ineffassign", "ineffassign", "warning", "e.go", 4, 0, "ineffective assign", "unused"),
			},
			[]string{"govet", "staticcheck", "gofmt", "misspell", "ineffassign"},
			nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStdout(func() { writeHumanMetaLint(tt.findings) })
			for _, c := range tt.checks {
				if !strings.Contains(out, c) {
					t.Errorf("expected %q in output, got: %s", c, out)
				}
			}
			for _, a := range tt.antiChecks {
				if strings.Contains(out, a) {
					t.Errorf("expected NOT to find %q in output, got: %s", a, out)
				}
			}
		})
	}
}

func TestGroupMetaLintFindings(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Message: "a", Category: "bug"},
		{Tool: "govet", Message: "b", Category: "bug"},
		{Tool: "gofmt", Message: "c", Category: "formatting"},
		{Tool: "misspell", Message: "d", Category: "style"},
	}
	groups := groupMetaLintFindings(findings)
	if len(groups["govet"]) != 2 {
		t.Errorf("govet should have 2, got %d", len(groups["govet"]))
	}
	if len(groups["gofmt"]) != 1 {
		t.Errorf("gofmt should have 1, got %d", len(groups["gofmt"]))
	}
	if len(groups["misspell"]) != 1 {
		t.Errorf("misspell should have 1, got %d", len(groups["misspell"]))
	}
	if len(groups["staticcheck"]) != 0 {
		t.Errorf("staticcheck should have 0, got %d", len(groups["staticcheck"]))
	}
}

func TestCardWidthNow(t *testing.T) {
	w := cardWidthNow()
	if w != 100 {
		t.Errorf("cardWidthNow = %d, want 100", w)
	}
}

func TestCardLineRaw(t *testing.T) {
	out := captureStdout(func() { cardLineRaw("some raw text") })
	if !strings.Contains(out, "some raw text") {
		t.Errorf("expected raw text in output, got: %s", out)
	}
}

func TestExecuteCheck_NonExistentPath(t *testing.T) {
	opts := checkOptions{format: FormatHuman, noSecrets: true, noVulnerabilities: true}
	err := executeCheck("/nonexistent/path/file.go", opts, false)
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
	if !strings.Contains(err.Error(), "cannot access") {
		t.Errorf("expected 'cannot access' in error, got: %v", err)
	}
}

func TestExecuteCheck_WithTempFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)
	opts := checkOptions{format: FormatHuman, noSecrets: true, noVulnerabilities: true}
	err := executeCheck(f, opts, false)
	if err != nil {
		t.Fatalf("executeCheck with temp file should not error: %v", err)
	}
}

func TestExecuteCheck_WithTempDir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\nfunc main() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("package main\nfunc f() {}\n"), 0644)
	opts := checkOptions{format: FormatHuman, noSecrets: true, noVulnerabilities: true}
	err := executeCheck(dir, opts, false)
	if err != nil {
		t.Fatalf("executeCheck with temp dir should not error: %v", err)
	}
}

func TestCheckDirectory_WithGoFiles(t *testing.T) {
	for _, tt := range []struct {
		name   string
		format FormatMode
		checks []string
	}{
		{"problems", FormatProblems, []string{"deep_nesting"}},
		{"markdown", FormatMarkdown, []string{"##"}},
		{"human", FormatHuman, []string{"main.go"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {\n\tif true {\n\t\tif true {\n\t\t\tif true {\n\t\t\t\tif true {\n\t\t\t\t\tprintln(\"deep\")\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t}\n}\n"), 0644)
			opts := checkOptions{format: tt.format, noSecrets: true, noVulnerabilities: true}
			out := captureStdout(func() {
				if err := checkDirectory(dir, opts, false); err != nil {
					t.Fatalf("checkDirectory should not error: %v", err)
				}
			})
			for _, c := range tt.checks {
				if !strings.Contains(out, c) {
					t.Errorf("expected %q in output, got: %s", c, out)
				}
			}
		})
	}
}

func TestCheckDirectory_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	opts := checkOptions{format: FormatHuman, noSecrets: true, noVulnerabilities: true}
	err := checkDirectory(dir, opts, false)
	if err != nil {
		t.Fatalf("checkDirectory on empty dir should not error: %v", err)
	}
}

func TestCheckDirectory_HiddenDirsSkipped(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)
	os.MkdirAll(filepath.Join(dir, ".hidden"), 0755)
	os.WriteFile(filepath.Join(dir, ".hidden", "secret.go"), []byte("package main\nfunc f() {}\n"), 0644)
	opts := checkOptions{format: FormatProblems, noSecrets: true, noVulnerabilities: true}
	out := captureStdout(func() {
		if err := checkDirectory(dir, opts, false); err != nil {
			t.Fatalf("checkDirectory should not error: %v", err)
		}
	})
	if strings.Contains(out, "secret.go") {
		t.Error(".hidden dir should be skipped")
	}
}

func TestCheckDirectory_WithGitignore(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.tmp\n"), 0644)
	os.WriteFile(filepath.Join(dir, "temp.tmp"), []byte("temp data\n"), 0644)
	opts := checkOptions{format: FormatProblems, noSecrets: true, noVulnerabilities: true}
	out := captureStdout(func() {
		if err := checkDirectory(dir, opts, true); err != nil {
			t.Fatalf("checkDirectory should not error: %v", err)
		}
	})
	if strings.Contains(out, "temp.tmp") {
		t.Error("gitignored files should not appear in output")
	}
}

func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCheckFile_QuietMode(t *testing.T) {
	for _, tt := range []struct {
		name     string
		content  string
		filename string
		opts     checkOptions
		wantErr  bool
		errCheck string
	}{
		{
			"suppresses normal output",
			"package main\nfunc main() {}\n",
			"main.go",
			checkOptions{format: FormatHuman, quiet: true, noSecrets: true, noVulnerabilities: true},
			false, "",
		},
		{
			"suppresses secrets output",
			"const key = \"sk_liv...p7dc\"\n",
			"keys.go",
			checkOptions{format: FormatProblems, quiet: true},
			false, "",
		},
		{
			"suppresses vulnerabilities output",
			"import pickle\npickle.loads(data)\n",
			"vuln.py",
			checkOptions{format: FormatProblems, quiet: true},
			false, "",
		},
		{
			"suppresses secrets-only output",
			"const key = \"sk_liv...p7dc\"\n",
			"keys.go",
			checkOptions{format: FormatProblems, quiet: true, secretsOnly: true},
			false, "",
		},
		{
			"suppresses vulnerabilities-only output",
			"import pickle\npickle.loads(data)\n",
			"vuln.py",
			checkOptions{format: FormatProblems, quiet: true, vulnerabilitiesOnly: true},
			false, "",
		},
		{
			"suppresses JSON output",
			"package main\nfunc main() {}\n",
			"main.go",
			checkOptions{format: FormatJSON, quiet: true},
			false, "",
		},
		{
			"errors still returned",
			"",
			"",
			checkOptions{format: FormatHuman, quiet: true},
			true, "cannot",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var f string
			if tt.filename != "" {
				dir := t.TempDir()
				f = writeTestFile(t, dir, tt.filename, tt.content)
			} else {
				f = "/nonexistent/path/file.go"
			}
			out := captureStdout(func() {
				err := checkFile(f, tt.opts)
				if tt.wantErr {
					if err == nil {
						t.Fatal("expected error")
					}
					if tt.errCheck != "" && !strings.Contains(err.Error(), tt.errCheck) {
						t.Errorf("expected %q in error, got: %v", tt.errCheck, err)
					}
					return
				}
				if err != nil {
					t.Fatalf("checkFile should not error: %v", err)
				}
			})
			if out != "" {
				t.Errorf("quiet mode should produce no stdout output, got: %q", out)
			}
		})
	}
}

func TestCheckDirectory_QuietMode(t *testing.T) {
	for _, tt := range []struct {
		name     string
		files    map[string]string
		opts     checkOptions
	}{
		{
			"suppresses all output",
			map[string]string{"a.go": "package main\nfunc main() {}\n", "b.go": "package main\nfunc f() {}\n"},
			checkOptions{format: FormatHuman, quiet: true, noSecrets: true, noVulnerabilities: true},
		},
		{
			"suppresses with secrets",
			map[string]string{"keys.go": "const key = \"sk_liv...p7dc\"\n"},
			checkOptions{format: FormatProblems, quiet: true},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tt.files {
				writeTestFile(t, dir, name, content)
			}
			out := captureStdout(func() {
				if err := checkDirectory(dir, tt.opts, false); err != nil {
					t.Fatalf("checkDirectory should not error: %v", err)
				}
			})
			if out != "" {
				t.Errorf("quiet mode should produce no stdout, got: %q", out)
			}
		})
	}
}

func TestCheckCommand_QuietFlag(t *testing.T) {
	cmd := CheckCommand()
	flag := cmd.Flags().Lookup("quiet")
	if flag == nil {
		t.Fatal("--quiet flag not found")
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default false, got %s", flag.DefValue)
	}
	if flag.Shorthand != "q" {
		t.Errorf("expected shorthand 'q', got %q", flag.Shorthand)
	}
}
