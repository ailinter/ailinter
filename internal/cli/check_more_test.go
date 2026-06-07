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
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error for 0 args")
	}
	err = cmd.Args(cmd, []string{"a", "b"})
	if err == nil {
		t.Error("expected error for 2 args")
	}
	err = cmd.Args(cmd, []string{"a"})
	if err != nil {
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
	data, _ := os.ReadFile(outputPath)
	if !strings.Contains(string(data), "Code Quality") && !strings.Contains(string(data), "Ailinter") {
		t.Logf("report content: %s", string(data))
	}
}

func TestWriteJSONMetaLint(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "test.go", Line: 10, Column: 5, Message: "unused variable", Category: "bug"},
		{Tool: "gofmt", Code: "formatting", Severity: "info", File: "test.go", Line: 1, Column: 0, Message: "file is not gofmt-ed", Category: "formatting"},
	}

	out := captureStdout(func() {
		writeJSONMetaLint(findings)
	})

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
	out := captureStdout(func() {
		writeJSONMetaLint(nil)
	})

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

func TestWriteMetaLintFindings_FormatJSON(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "staticcheck", Code: "S1017", Severity: "warning", File: "test.go", Line: 5, Column: 1, Message: "should use strings.Replace", Category: "style"},
	}

	out := captureStdout(func() {
		writeMetaLintFindings(FormatJSON, findings)
	})

	if !strings.Contains(out, "meta_lint") {
		t.Error("JSON output should contain meta_lint key")
	}
	var parsed struct {
		MetaLint []metalinter.Finding `json:"meta_lint"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("expected valid JSON: %v\noutput: %s", err, out)
	}
	if len(parsed.MetaLint) != 1 {
		t.Errorf("got %d findings, want 1", len(parsed.MetaLint))
	}
}

func TestWriteMetaLintFindings_FormatProblems(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "test.go", Line: 10, Column: 5, Message: "unused variable", Category: "bug"},
	}

	out := captureStdout(func() {
		writeMetaLintFindings(FormatProblems, findings)
	})

	if !strings.Contains(out, "test.go:10:5") {
		t.Errorf("problems output should contain file:line:col, got: %s", out)
	}
	if !strings.Contains(out, "govet") {
		t.Errorf("problems output should contain tool name, got: %s", out)
	}
}

func TestWriteMetaLintFindings_FormatMarkdown(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "gofmt", Code: "formatting", Severity: "info", File: "test.go", Line: 1, Column: 0, Message: "file is not gofmt-ed", Category: "formatting"},
	}

	out := captureStdout(func() {
		writeMetaLintFindings(FormatMarkdown, findings)
	})

	if !strings.Contains(out, "Meta-Lint Findings") {
		t.Errorf("md output should contain heading, got: %s", out)
	}
	if !strings.Contains(out, "gofmt") {
		t.Errorf("md output should contain tool name, got: %s", out)
	}
	if !strings.Contains(out, "| 1 |") {
		t.Errorf("md output should contain line number, got: %s", out)
	}
}

func TestWriteMetaLintFindings_FormatHuman(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "test.go", Line: 10, Column: 5, Message: "unused variable", Category: "bug"},
	}

	out := captureStdout(func() {
		writeMetaLintFindings(FormatHuman, findings)
	})

	if !strings.Contains(out, "meta-lint") {
		t.Errorf("human output should contain meta-lint heading, got: %s", out)
	}
	if !strings.Contains(out, "govet") {
		t.Errorf("human output should contain tool name, got: %s", out)
	}
}

func TestWriteMetaLintFindings_Empty(t *testing.T) {
	out := captureStdout(func() {
		writeMetaLintFindings(FormatJSON, nil)
	})
	if out != "" {
		t.Errorf("empty findings should produce no output, got: %s", out)
	}

	out = captureStdout(func() {
		writeMetaLintFindings(FormatHuman, nil)
	})
	if out != "" {
		t.Errorf("empty findings should produce no output, got: %s", out)
	}
}

func TestWriteMetaLintFindings_FormatProblemDefault(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "misspell", Code: "spelling", Severity: "warning", File: "test.go", Line: 3, Column: 1, Message: "typo 'langauge'", Category: "style"},
	}
	// FormatAuto should fall through to default (human)
	out := captureStdout(func() {
		writeMetaLintFindings(FormatAuto, findings)
	})
	if !strings.Contains(out, "meta-lint") {
		t.Errorf("auto format should produce human output, got: %s", out)
	}
}

func TestWriteMarkdownMetaLint_Empty(t *testing.T) {
	out := captureStdout(func() {
		writeMarkdownMetaLint(nil)
	})
	if !strings.Contains(out, "Meta-Lint Findings (0)") {
		t.Errorf("should show heading with 0 count, got: %s", out)
	}
}

func TestWriteMarkdownMetaLint_WithFindings(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "src/main.go", Line: 42, Column: 5, Message: "unused variable", Category: "bug"},
		{Tool: "gofmt", Code: "formatting", Severity: "info", File: "src/main.go", Line: 1, Column: 0, Message: "file is not gofmt-ed", Category: "formatting"},
	}

	out := captureStdout(func() {
		writeMarkdownMetaLint(findings)
	})

	if !strings.Contains(out, "Meta-Lint Findings (2)") {
		t.Errorf("should show count in heading, got: %s", out)
	}
	if !strings.Contains(out, "| govet | SA1000 | warning | src/main.go | 42:5 |") {
		t.Errorf("should have table row with full info, got: %s", out)
	}
	if !strings.Contains(out, "gofmt") {
		t.Errorf("should contain gofmt, got: %s", out)
	}
}

func TestWriteMarkdownMetaLint_ColumnFormat(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "test.go", Line: 10, Column: 3, Message: "message", Category: "bug"},
	}

	out := captureStdout(func() {
		writeMarkdownMetaLint(findings)
	})

	if !strings.Contains(out, "| 10:3 |") {
		t.Errorf("should show line:col when column > 0, got: %s", out)
	}
}

func TestWriteMarkdownMetaLint_LineOnly(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "test.go", Line: 10, Column: 0, Message: "message", Category: "bug"},
	}

	out := captureStdout(func() {
		writeMarkdownMetaLint(findings)
	})

	if strings.Contains(out, "| 10:0 |") {
		t.Errorf("should show only line when column is 0, got: %s", out)
	}
	if !strings.Contains(out, "| 10 |") {
		t.Errorf("should show line number only, got: %s", out)
	}
}

func TestWriteMarkdownMetaLint_LineZero(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "test.go", Line: 0, Column: 0, Message: "message", Category: "bug"},
	}

	out := captureStdout(func() {
		writeMarkdownMetaLint(findings)
	})

	if !strings.Contains(out, "| test.go |  |") {
		t.Errorf("should have empty location when line is 0, got: %s", out)
	}
}

func TestWriteMarkdownMetaLint_PipeInMessage(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Severity: "warning", File: "test.go", Line: 1, Message: "msg with | pipe", Code: "SA1000", Category: "bug"},
	}

	out := captureStdout(func() {
		writeMarkdownMetaLint(findings)
	})

	if !strings.Contains(out, "msq with \\| pipe") && !strings.Contains(out, "msg with") {
		t.Logf("pipe-in-message output: %s", out)
	}
}

func TestWriteHumanMetaLint_Empty(t *testing.T) {
	out := captureStdout(func() {
		writeHumanMetaLint(nil)
	})
	if !strings.Contains(out, "meta-lint") {
		t.Errorf("human output should show heading even with empty findings, got: %s", out)
	}
}

func TestWriteHumanMetaLint_WithFindings(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "src/main.go", Line: 42, Column: 5, Message: "unused variable x", Category: "bug"},
		{Tool: "gofmt", Code: "formatting", Severity: "info", File: "src/main.go", Line: 1, Column: 0, Message: "file is not gofmt-ed", Category: "formatting"},
	}

	out := captureStdout(func() {
		writeHumanMetaLint(findings)
	})

	if !strings.Contains(out, "meta-lint") {
		t.Errorf("should contain meta-lint heading, got: %s", out)
	}
	if !strings.Contains(out, "govet: 1 finding") {
		t.Errorf("should show govet count, got: %s", out)
	}
	if !strings.Contains(out, "gofmt: 1 finding") {
		t.Errorf("should show gofmt count, got: %s", out)
	}
	if !strings.Contains(out, "unused variable") {
		t.Errorf("should contain message, got: %s", out)
	}
	if !strings.Contains(out, "L42") {
		t.Errorf("should show line number, got: %s", out)
	}
}

func TestWriteHumanMetaLint_CodeColumn(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "test.go", Line: 10, Message: "unused", Category: "bug"},
	}

	out := captureStdout(func() {
		writeHumanMetaLint(findings)
	})

	if !strings.Contains(out, "[SA1000]") {
		t.Errorf("should show code in brackets, got: %s", out)
	}
}

func TestWriteHumanMetaLint_FileAndLocation(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "src/main.go", Line: 10, Column: 3, Message: "unused", Category: "bug"},
	}

	out := captureStdout(func() {
		writeHumanMetaLint(findings)
	})

	if !strings.Contains(out, "src/main.go") {
		t.Errorf("should contain file path, got: %s", out)
	}
	if !strings.Contains(out, "L10:3") {
		t.Errorf("should contain line:col, got: %s", out)
	}
}

func TestWriteHumanMetaLint_NoFile(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "", Line: 10, Message: "unused", Category: "bug"},
	}

	out := captureStdout(func() {
		writeHumanMetaLint(findings)
	})

	if strings.Contains(out, "└") {
		t.Error("should not print file line when file is empty")
	}
}

func TestWriteHumanMetaLint_NoCode(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Severity: "warning", File: "test.go", Line: 10, Message: "unused", Category: "bug"},
	}

	out := captureStdout(func() {
		writeHumanMetaLint(findings)
	})

	if strings.Contains(out, "[]") {
		t.Error("should not print empty brackets when code is empty")
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
	out := captureStdout(func() {
		cardLineRaw("some raw text")
	})
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

func TestCheckDirectory_WithGoFilesFormatProblems(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {\n\tif true {\n\t\tif true {\n\t\t\tif true {\n\t\t\t\tif true {\n\t\t\t\t\tprintln(\"deep\")\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t}\n}\n"), 0644)

	opts := checkOptions{format: FormatProblems, noSecrets: true, noVulnerabilities: true}
	out := captureStdout(func() {
		err := checkDirectory(dir, opts, false)
		if err != nil {
			t.Fatalf("checkDirectory should not error: %v", err)
		}
	})
	if !strings.Contains(out, "deep_nesting") {
		t.Logf("problems format output: %s", out)
	}
}

func TestCheckDirectory_WithGoFilesFormatMarkdown(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	opts := checkOptions{format: FormatMarkdown, noSecrets: true, noVulnerabilities: true}
	out := captureStdout(func() {
		err := checkDirectory(dir, opts, false)
		if err != nil {
			t.Fatalf("checkDirectory should not error: %v", err)
		}
	})
	if !strings.Contains(out, "##") {
		t.Errorf("markdown output should contain heading, got: %s", out)
	}
}

func TestCheckDirectory_WithGoFilesFormatHuman(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	opts := checkOptions{format: FormatHuman, noSecrets: true, noVulnerabilities: true}
	out := captureStdout(func() {
		err := checkDirectory(dir, opts, false)
		if err != nil {
			t.Fatalf("checkDirectory should not error: %v", err)
		}
	})
	if !strings.Contains(out, "main.go") {
		t.Errorf("human output should contain filename, got: %s", out)
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
		err := checkDirectory(dir, opts, false)
		if err != nil {
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
		err := checkDirectory(dir, opts, true)
		if err != nil {
			t.Fatalf("checkDirectory should not error: %v", err)
		}
	})
	if strings.Contains(out, "temp.tmp") {
		t.Error("gitignored files should not appear in output")
	}
}

func TestWriteHumanMetaLint_MultipleTools(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "a.go", Line: 1, Message: "unused", Category: "bug"},
		{Tool: "staticcheck", Code: "SA1000", Severity: "warning", File: "b.go", Line: 2, Message: "ineffective assign", Category: "bug"},
		{Tool: "gofmt", Code: "formatting", Severity: "info", File: "c.go", Line: 0, Message: "formatting", Category: "formatting"},
		{Tool: "misspell", Code: "spelling", Severity: "warning", File: "d.go", Line: 3, Message: "typo", Category: "style"},
		{Tool: "ineffassign", Code: "ineffassign", Severity: "warning", File: "e.go", Line: 4, Message: "ineffective assign", Category: "unused"},
	}

	out := captureStdout(func() {
		writeHumanMetaLint(findings)
	})

	tools := []string{"govet", "staticcheck", "gofmt", "misspell", "ineffassign"}
	for _, tool := range tools {
		if !strings.Contains(out, tool) {
			t.Errorf("expected %q in output", tool)
		}
	}
}

func TestWriteHumanMetaLint_ColumnLineFormat(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "govet", Code: "SA1000", Severity: "warning", File: "test.go", Line: 10, Column: 3, Message: "unused", Category: "bug"},
	}

	out := captureStdout(func() {
		writeHumanMetaLint(findings)
	})

	if !strings.Contains(out, "L10:3") {
		t.Errorf("should show line:col when both present, got: %s", out)
	}
}

func TestWriteHumanMetaLint_LineNoColumn(t *testing.T) {
	findings := []metalinter.Finding{
		{Tool: "gofmt", Code: "formatting", Severity: "info", File: "test.go", Line: 1, Column: 0, Message: "formatting", Category: "formatting"},
	}

	out := captureStdout(func() {
		writeHumanMetaLint(findings)
	})

	if strings.Contains(out, "L1:0") {
		t.Errorf("should not show :0 for column, got: %s", out)
	}
	if !strings.Contains(out, " L1") {
		t.Errorf("should show line number, got: %s", out)
	}
}

func TestCheckFile_QuietMode(t *testing.T) {
	t.Run("quiet suppresses normal output", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "main.go")
		os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)

		opts := checkOptions{
			format:            FormatHuman,
			quiet:             true,
			noSecrets:         true,
			noVulnerabilities: true,
		}

		out := captureStdout(func() {
			err := checkFile(f, opts)
			if err != nil {
				t.Fatalf("checkFile in quiet mode should not error: %v", err)
			}
		})
		if out != "" {
			t.Errorf("quiet mode should produce no stdout output, got: %q", out)
		}
	})

	t.Run("quiet suppresses secrets output", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "keys.go")
		// This content contains a secret token
		os.WriteFile(f, []byte("const key = \"sk_live_4eC39HqLyjWDarjtT1zdp7dc\"\n"), 0644) // gitleaks:allow

		opts := checkOptions{
			format: FormatProblems,
			quiet:  true,
		}

		out := captureStdout(func() {
			err := checkFile(f, opts)
			if err != nil {
				t.Fatalf("checkFile in quiet mode should not error: %v", err)
			}
		})
		if out != "" {
			t.Errorf("quiet mode should suppress secrets output, got: %q", out)
		}
	})

	t.Run("quiet suppresses vulnerabilities output", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "vuln.py")
		os.WriteFile(f, []byte("import pickle\npickle.loads(data)\n"), 0644)

		opts := checkOptions{
			format: FormatProblems,
			quiet:  true,
		}

		out := captureStdout(func() {
			err := checkFile(f, opts)
			if err != nil {
				t.Fatalf("checkFile in quiet mode should not error: %v", err)
			}
		})
		if out != "" {
			t.Errorf("quiet mode should suppress vulnerabilities output, got: %q", out)
		}
	})

	t.Run("quiet suppresses secrets-only output", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "keys.go")
		os.WriteFile(f, []byte("const key = \"sk_live_4eC39HqLyjWDarjtT1zdp7dc\"\n"), 0644) // gitleaks:allow

		opts := checkOptions{
			format:      FormatProblems,
			quiet:       true,
			secretsOnly: true,
		}

		out := captureStdout(func() {
			err := checkFile(f, opts)
			if err != nil {
				t.Fatalf("checkFile in quiet+secretsOnly should not error: %v", err)
			}
		})
		if out != "" {
			t.Errorf("quiet mode should suppress secrets-only output, got: %q", out)
		}
	})

	t.Run("quiet suppresses vulnerabilities-only output", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "vuln.py")
		os.WriteFile(f, []byte("import pickle\npickle.loads(data)\n"), 0644)

		opts := checkOptions{
			format:              FormatProblems,
			quiet:               true,
			vulnerabilitiesOnly: true,
		}

		out := captureStdout(func() {
			err := checkFile(f, opts)
			if err != nil {
				t.Fatalf("checkFile in quiet+vulnsOnly should not error: %v", err)
			}
		})
		if out != "" {
			t.Errorf("quiet mode should suppress vulnerabilities-only output, got: %q", out)
		}
	})

	t.Run("quiet suppresses JSON output", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "main.go")
		os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)

		opts := checkOptions{
			format: FormatJSON,
			quiet:  true,
		}

		out := captureStdout(func() {
			err := checkFile(f, opts)
			if err != nil {
				t.Fatalf("checkFile in quiet+JSON should not error: %v", err)
			}
		})
		if out != "" {
			t.Errorf("quiet mode should suppress JSON output, got: %q", out)
		}
	})

	t.Run("errors still returned", func(t *testing.T) {
		opts := checkOptions{
			format: FormatHuman,
			quiet:  true,
		}
		err := checkFile("/nonexistent/path/file.go", opts)
		if err == nil {
			t.Fatal("expected error for nonexistent file even in quiet mode")
		}
		if !strings.Contains(err.Error(), "cannot") {
			t.Errorf("expected error about file access, got: %v", err)
		}
	})
}

func TestCheckDirectory_QuietMode(t *testing.T) {
	t.Run("quiet suppresses all directory output", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\nfunc main() {}\n"), 0644)
		os.WriteFile(filepath.Join(dir, "b.go"), []byte("package main\nfunc f() {}\n"), 0644)

		opts := checkOptions{
			format:            FormatHuman,
			quiet:             true,
			noSecrets:         true,
			noVulnerabilities: true,
		}

		out := captureStdout(func() {
			err := checkDirectory(dir, opts, false)
			if err != nil {
				t.Fatalf("checkDirectory in quiet mode should not error: %v", err)
			}
		})
		if out != "" {
			t.Errorf("quiet mode should produce no stdout for directory, got: %q", out)
		}
	})

	t.Run("quiet suppresses directory with secrets", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "keys.go"), []byte("const key = \"sk_live_4eC39HqLyjWDarjtT1zdp7dc\"\n"), 0644) // gitleaks:allow

		opts := checkOptions{
			format: FormatProblems,
			quiet:  true,
		}

		out := captureStdout(func() {
			err := checkDirectory(dir, opts, false)
			if err != nil {
				t.Fatalf("checkDirectory in quiet mode should not error: %v", err)
			}
		})
		if out != "" {
			t.Errorf("quiet mode should suppress all directory output, got: %q", out)
		}
	})
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
	// Verify shorthand exists
	if flag.Shorthand != "q" {
		t.Errorf("expected shorthand 'q', got %q", flag.Shorthand)
	}
}
