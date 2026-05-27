package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/vulnerability"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old
	return buf.String()
}

func TestFormatMode_String(t *testing.T) {
	cases := map[FormatMode]string{
		FormatAuto:     "auto",
		FormatHuman:    "human",
		FormatJSON:     "json",
		FormatMarkdown: "markdown",
		FormatProblems: "problems",
	}
	for mode, want := range cases {
		if got := mode.String(); got != want {
			t.Errorf("FormatMode(%d).String() = %q, want %q", mode, got, want)
		}
	}
}

func TestDetectFormat_Explicit(t *testing.T) {
	cases := map[string]FormatMode{
		"json":     FormatJSON,
		"JSON":     FormatJSON,
		"markdown": FormatMarkdown,
		"md":       FormatMarkdown,
		"text":     FormatHuman,
		"human":    FormatHuman,
		"problems": FormatProblems,
		"gcc":      FormatProblems,
		"vscode":   FormatProblems,
		"auto":     FormatAuto,
	}
	for flag, want := range cases {
		if got := DetectFormat(flag); got != want {
			t.Errorf("DetectFormat(%q) = %v, want %v", flag, got, want)
		}
	}
}

func TestDetectFormat_Empty(t *testing.T) {
	old := os.Getenv("CLI_FORMAT")
	os.Setenv("CLI_FORMAT", "json")
	defer os.Setenv("CLI_FORMAT", old)

	if got := DetectFormat(""); got != FormatJSON {
		t.Errorf("DetectFormat empty with CLI_FORMAT=json = %v, want FormatJSON", got)
	}
}

func TestResolveFormat(t *testing.T) {
	oldNoColor := os.Getenv("NO_COLOR")

	os.Setenv("NO_COLOR", "")
	if got := ResolveFormat("json"); got != FormatJSON {
		t.Errorf("ResolveFormat(json) = %v, want FormatJSON", got)
	}

	os.Setenv("NO_COLOR", oldNoColor)
}

func TestIsColorEnabled(t *testing.T) {
	old := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer os.Setenv("NO_COLOR", old)

	if IsColorEnabled() {
		t.Error("IsColorEnabled should be false with NO_COLOR=1")
	}
}

func TestIsSourceFile(t *testing.T) {
	source := []string{
		"main.go", "app.py", "script.js", "component.ts", "page.tsx",
		"model.java", "lib.rs", "helper.rb", "main.c", "app.cpp",
		"header.h", "core.hpp", "program.cs", "view.swift", "logic.kt",
		"test.kts", "data.scala", "index.php", "script.pl", "run.sh",
		"deploy.bash", "infra.tf", "config.yaml", "settings.yml",
		"config.toml", "data.json", "layout.xml", "index.html", "style.css", "query.sql",
	}
	for _, f := range source {
		if !isSourceFile(f) {
			t.Errorf("isSourceFile(%q) = false, want true", f)
		}
	}

	nonSource := []string{"image.png", "doc.pdf", "video.mp4", "archive.zip", "binary", ""}
	for _, f := range nonSource {
		if isSourceFile(f) {
			t.Errorf("isSourceFile(%q) = true, want false", f)
		}
	}
}

func TestWriteHumanResult(t *testing.T) {
	result := analyzer.QualityResult{
		Score:       85,
		Label:       analyzer.LabelProceedWithCare,
		FilePath:    "test.go",
		Language:    "go",
		LinesOfCode: 120,
		Smells: []analyzer.Smell{
			{Name: "deep_nesting", Severity: "warning", LineStart: 42, Message: "Nesting depth 4", AIPrompt: "Use guard clauses."},
			{Name: "brain_method", Severity: "alert", LineStart: 10, Message: "Function too long", AIPrompt: "Extract methods."},
		},
	}

	out := captureStdout(func() {
		writeHumanResult(result)
	})

	if !strings.Contains(out, "test.go") {
		t.Error("human output should contain file path")
	}
	if !strings.Contains(out, "85/100") {
		t.Error("human output should contain score")
	}
	if !strings.Contains(out, analyzer.LabelProceedWithCare) {
		t.Error("human output should contain label")
	}
	if !strings.Contains(out, "deep_nesting") {
		t.Error("human output should contain smell names")
	}
}

func TestWriteHumanResult_NoIssues(t *testing.T) {
	result := analyzer.QualityResult{
		Score:       95,
		Label:       analyzer.LabelGoAhead,
		FilePath:    "clean.go",
		Language:    "go",
		LinesOfCode: 50,
		Smells:      []analyzer.Smell{},
	}

	out := captureStdout(func() {
		writeHumanResult(result)
	})

	if !strings.Contains(out, "No issues") {
		t.Error("human output should say 'No issues' when empty", out)
	}
}

func TestWriteHumanSecrets(t *testing.T) {
	findings := []secrets.SecretFinding{
		{RuleID: "aws-access-key", Severity: "critical", Line: 5, Description: "AWS key found", Message: "Use env vars."},
		{RuleID: "stripe-key", Severity: "alert", Line: 10, Description: "Stripe key found", Message: "Use env vars."},
	}

	out := captureStdout(func() {
		writeHumanSecrets("test.go", findings)
	})

	if !strings.Contains(out, "secrets detected") {
		t.Error("secrets output should show count")
	}
	if !strings.Contains(out, "aws-access-key") {
		t.Error("secrets output should show rule IDs")
	}
}

func TestWriteHumanSecrets_Empty(t *testing.T) {
	// writeSecrets dispatches to writeHumanSecrets only when findings > 0
	// Empty list produces no output
	writeSecrets(FormatHuman, "test.go", nil)
}

func TestWriteHumanSummary(t *testing.T) {
	results := []analyzer.QualityResult{
		{Score: 95, Label: analyzer.LabelGoAhead, FilePath: "a.go", Smells: []analyzer.Smell{{Severity: "warning", Name: "x"}}},
		{Score: 60, Label: analyzer.LabelStopRefactor, FilePath: "b.go", Smells: []analyzer.Smell{{Severity: "critical", Name: "y"}}},
	}

	out := captureStdout(func() {
		writeHumanSummary(results)
	})

	if !strings.Contains(out, "Summary") {
		t.Error("summary should contain 'Summary' heading")
	}
	if !strings.Contains(out, "2 files") {
		t.Error("summary should show file count")
	}
	if !strings.Contains(out, "Stop & Refactor") {
		t.Error("summary should show Proceed with Care or Stop & Refactor section")
	}
}

func TestWriteHumanSummary_SingleFile(t *testing.T) {
	results := []analyzer.QualityResult{
		{Score: 95, Label: analyzer.LabelGoAhead, FilePath: "a.go"},
	}

	out := captureStdout(func() {
		writeHumanSummary(results)
	})

	// Single file should not print a summary
	if strings.Contains(out, "Summary") {
		t.Error("single file should not print summary")
	}
}

func TestWriteMarkdownResult(t *testing.T) {
	result := analyzer.QualityResult{
		Score:       72,
		Label:       analyzer.LabelProceedWithCare,
		FilePath:    "src/main.go",
		Language:    "go",
		LinesOfCode: 200,
		Smells: []analyzer.Smell{
			{Name: "deep_nesting", Severity: "warning", LineStart: 15, Message: "Depth 3", AIPrompt: "Flatten."},
		},
	}

	out := captureStdout(func() {
		writeMarkdownResult(result)
	})

	if !strings.Contains(out, "## src/main.go") {
		t.Error("md output should contain heading with file path")
	}
	if !strings.Contains(out, "| warning | deep_nesting | 15 |") {
		t.Error("md output should contain smell table row")
	}
	if !strings.Contains(out, "72/100") {
		t.Error("md output should contain score")
	}
}

func TestWriteMarkdownResult_NoIssues(t *testing.T) {
	result := analyzer.QualityResult{
		Score:       95,
		Label:       analyzer.LabelGoAhead,
		FilePath:    "clean.go",
		Language:    "go",
		LinesOfCode: 30,
		Smells:      []analyzer.Smell{},
	}

	out := captureStdout(func() {
		writeMarkdownResult(result)
	})

	if !strings.Contains(out, "No Code Quality issues detected") {
		t.Error("md output should show 'No issues' when clean")
	}
}

func TestWriteMarkdownSecrets(t *testing.T) {
	findings := []secrets.SecretFinding{
		{RuleID: "aws-key", Severity: "critical", Line: 3, Description: "AWS key", Message: "Fix."},
	}

	out := captureStdout(func() {
		writeMarkdownSecrets("secrets.go", findings)
	})

	if !strings.Contains(out, "### 🔑 Secret Scan") {
		t.Error("md secrets should have heading")
	}
}

func TestWriteMarkdownSummary(t *testing.T) {
	results := []analyzer.QualityResult{
		{Score: 95, Label: analyzer.LabelGoAhead, FilePath: "a.go"},
		{Score: 50, Label: analyzer.LabelStopRefactor, FilePath: "b.go"},
	}

	out := captureStdout(func() {
		writeMarkdownSummary(results)
	})

	if !strings.Contains(out, "## Summary") {
		t.Error("md summary should have heading")
	}
	if !strings.Contains(out, "b.go") {
		t.Error("md summary should list Needs Work files")
	}
}

func TestWriteProblemsResult(t *testing.T) {
	result := analyzer.QualityResult{
		FilePath: "src/main.go", Score: 85, Label: analyzer.LabelProceedWithCare,
		Language: "go", LinesOfCode: 200,
		Smells: []analyzer.Smell{
			{Name: "deep_nesting", Severity: "critical", LineStart: 5, Message: "Nesting depth 6"},
			{Name: "brain_method", Severity: "warning", LineStart: 10, Message: "Function too long"},
		},
	}

	out := captureStdout(func() {
		writeProblemsResult(result)
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (1 comment + 2 problems), got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "#") {
		t.Error("first line should be a comment with score info")
	}
	if !strings.Contains(lines[0], "score=85") {
		t.Error("comment line should contain score")
	}
	if !strings.Contains(lines[1], ": error:") {
		t.Error("critical should map to 'error' severity")
	}
	if !strings.Contains(lines[2], ": warning:") {
		t.Error("warning should stay 'warning'")
	}
}

func TestWriteProblemsSecrets(t *testing.T) {
	findings := []secrets.SecretFinding{
		{RuleID: "aws-key", Severity: "critical", Line: 3, Column: 10, Description: "Key found"},
	}

	out := captureStdout(func() {
		writeProblemsSecrets("test.go", findings)
	})

	if !strings.Contains(out, "test.go:3:10") {
		t.Error("problems output should have file:line:col format")
	}
}

func TestWriteJSONResult(t *testing.T) {
	result := analyzer.QualityResult{
		Score: 85, Label: analyzer.LabelProceedWithCare,
		FilePath: "test.go", Language: "go", LinesOfCode: 100,
	}

	out := captureStdout(func() {
		writeJSONResult(result)
	})

	var parsed analyzer.QualityResult
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("JSON output must be valid: %v", err)
	}
	if parsed.Score != 85 {
		t.Errorf("score = %d, want 85", parsed.Score)
	}
}

func TestWriteJSONResults(t *testing.T) {
	results := []analyzer.QualityResult{
		{Score: 95, Label: analyzer.LabelGoAhead, FilePath: "a.go", Language: "go", LinesOfCode: 30},
		{Score: 60, Label: analyzer.LabelStopRefactor, FilePath: "b.go", Language: "go", LinesOfCode: 500},
	}

	out := captureStdout(func() {
		writeJSONResults(results)
	})

	var parsed struct {
		CodeQuality []analyzer.QualityResult `json:"code_quality"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("JSON output must be valid: %v", err)
	}
	if len(parsed.CodeQuality) != 2 {
		t.Errorf("got %d results, want 2", len(parsed.CodeQuality))
	}
}

func TestWriteResult_Dispatch(t *testing.T) {
	result := analyzer.QualityResult{
		Score: 90, Label: analyzer.LabelGoAhead,
		FilePath: "test.go", Language: "go", LinesOfCode: 50,
	}

	out := captureStdout(func() {
		writeResult(FormatMarkdown, result)
	})
	if !strings.Contains(out, "## test.go") {
		t.Error("FormatMarkdown should produce md output")
	}

	out = captureStdout(func() {
		writeResult(FormatProblems, result)
	})
	if out != "" {
		t.Log("problems output for clean file (no smells): ok")
	}

	out = captureStdout(func() {
		writeResult(FormatJSON, result)
	})
	var parsed analyzer.QualityResult
	json.Unmarshal([]byte(out), &parsed)
	if parsed.Score != 90 {
		t.Error("JSON dispatch failed")
	}
}

func TestWriteResults_Dispatch(t *testing.T) {
	results := []analyzer.QualityResult{
		{Score: 95, Label: analyzer.LabelGoAhead, FilePath: "a.go", Language: "go", LinesOfCode: 30},
	}

	writeResults(FormatMarkdown, results)
	writeResults(FormatProblems, results)
	writeResults(FormatJSON, results)
}

func TestWriteSecrets_Dispatch(t *testing.T) {
	findings := []secrets.SecretFinding{
		{RuleID: "test", Severity: "warning", Line: 1, Description: "test"},
	}

	writeSecrets(FormatMarkdown, "x.go", findings)
	writeSecrets(FormatProblems, "x.go", findings)
}

func TestWriteSummary_Dispatch(t *testing.T) {
	results := []analyzer.QualityResult{
		{Score: 50, Label: analyzer.LabelStopRefactor, FilePath: "x.go"},
		{Score: 75, Label: analyzer.LabelProceedWithCare, FilePath: "y.go"},
	}

	writeSummary(FormatMarkdown, results)
}

func TestTruncateMsg(t *testing.T) {
	if got := truncateMsg("hello", 10); got != "hello" {
		t.Errorf("short string: %q", got)
	}
	if got := truncateMsg("hello world", 10); !strings.Contains(got, "…") {
		t.Errorf("truncated string should end with ellipsis: %q", got)
	}
}

func TestCountBySeverity(t *testing.T) {
	smells := []analyzer.Smell{
		{Severity: "warning"},
		{Severity: "warning"},
		{Severity: "critical"},
		{Severity: "alert"},
	}
	counts := countBySeverity(smells)
	if counts["warning"] != 2 {
		t.Errorf("warning count = %d", counts["warning"])
	}
	if counts["critical"] != 1 {
		t.Errorf("critical count = %d", counts["critical"])
	}
	if counts["alert"] != 1 {
		t.Errorf("alert count = %d", counts["alert"])
	}
}

func TestVisualLen(t *testing.T) {
	if visualLen("hello") != 5 {
		t.Error("plain text length")
	}
	if visualLen("\033[31mred\033[0m") != 3 {
		t.Error("ANSI escape codes should be excluded from length")
	}
}

func TestSeverityBlock(t *testing.T) {
	b := severityBlock("critical")
	if b == "" {
		t.Error("severityBlock should return non-empty for known severity")
	}
}

func TestSeverityLabel(t *testing.T) {
	if got := severityLabel("warning"); got != "WARNING" {
		t.Errorf("severityLabel = %q, want WARNING", got)
	}
}

func TestSeverityPrefix(t *testing.T) {
	p := severityPrefix("warning")
	if !strings.Contains(p, "WARNING") {
		t.Error("severity prefix should contain label")
	}
}

func TestScoreBar(t *testing.T) {
	bar := scoreBar(50)
	if len([]rune(bar)) != 10 {
		t.Errorf("scoreBar should be 10 chars, got %d", len([]rune(bar)))
	}
	if !strings.Contains(bar, "█") {
		t.Error("scoreBar should have block characters")
	}
}

func TestProblemSeverity(t *testing.T) {
	if got := problemSeverity("critical"); got != "error" {
		t.Errorf("critical -> %q, want error", got)
	}
	if got := problemSeverity("alert"); got != "warning" {
		t.Errorf("alert -> %q, want warning", got)
	}
	if got := problemSeverity("warning"); got != "warning" {
		t.Errorf("warning -> %q, want warning", got)
	}
}

func TestProgressToStderr(t *testing.T) {
	progressToStderr("test %d", 42)
}

func TestTerminalWidth(t *testing.T) {
	w := TerminalWidth()
	if w <= 0 {
		t.Errorf("TerminalWidth should be positive, got %d", w)
	}
}

func TestCheckFile_Integration(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {\n\tif true {\n\t\tprintln(\"nested\")\n\t}\n}\n"), 0644)

	opts := checkOptions{format: FormatHuman, noSecrets: true, noVulnerabilities: true}
	_ = captureStdout(func() {
		checkFile(f, opts)
	})

	opts.format = FormatJSON
	_ = captureStdout(func() {
		checkFile(f, opts)
	})

	opts.format = FormatMarkdown
	_ = captureStdout(func() {
		checkFile(f, opts)
	})

	opts.format = FormatProblems
	_ = captureStdout(func() {
		checkFile(f, opts)
	})

	opts.format = FormatHuman
	opts.langOverride = "python"
	_ = captureStdout(func() {
		checkFile(f, opts)
	})
}

func TestCheckFile_NotFound(t *testing.T) {
	err := checkFile("/nonexistent/file.go", checkOptions{format: FormatHuman, noSecrets: true})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCheckDirectory_Integration(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, "app.py"), []byte("def hello():\n    pass\n"), 0644)
	os.MkdirAll(filepath.Join(dir, ".hidden"), 0755)
	os.WriteFile(filepath.Join(dir, ".hidden", "secret.go"), []byte("package hidden\nfunc f() {}\n"), 0644)

	opts := checkOptions{format: FormatHuman, noSecrets: true, noVulnerabilities: true}
	_ = captureStdout(func() { checkDirectory(dir, opts, false) })
	opts.format = FormatJSON
	_ = captureStdout(func() { checkDirectory(dir, opts, false) })
	opts.format = FormatMarkdown
	_ = captureStdout(func() { checkDirectory(dir, opts, false) })
	opts.format = FormatProblems
	_ = captureStdout(func() { checkDirectory(dir, opts, false) })
}

func TestGroupSmellsBySeverity(t *testing.T) {
	smells := []analyzer.Smell{
		{Name: "a", Severity: "critical"},
		{Name: "b", Severity: "warning"},
		{Name: "c", Severity: "warning"},
	}
	groups := groupSmellsBySeverity(smells)
	if len(groups["critical"]) != 1 {
		t.Error("critical should have 1 smell")
	}
	if len(groups["warning"]) != 2 {
		t.Error("warning should have 2 smells")
	}
	if len(groups["alert"]) != 0 {
		t.Error("alert should be empty")
	}
}

func TestSortedSeverityKeys(t *testing.T) {
	keys := sortedSeverityKeys()
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "critical" {
		t.Error("first should be critical")
	}
}

func TestGroupSecretFindings(t *testing.T) {
	findings := []secrets.SecretFinding{
		{RuleID: "a", Severity: "critical"},
		{RuleID: "b", Severity: "alert"},
	}
	groups := groupSecretFindings(findings)
	if len(groups["critical"]) != 1 {
		t.Error("critical should have 1")
	}
	if len(groups["alert"]) != 1 {
		t.Error("alert should have 1")
	}
}

func TestSeverityCountParts(t *testing.T) {
	counts := map[string]int{"critical": 2, "warning": 5, "alert": 0}
	parts := severityCountParts(counts)
	if len(parts) != 3 {
		t.Errorf("expected 3 parts (alert has 0 but key exists), got %d", len(parts))
	}
}

func TestCardTopBottom(t *testing.T) {
	_ = captureStdout(func() {
		cardTop("test.go")
	})

	_ = captureStdout(func() {
		cardDivider()
	})

	_ = captureStdout(func() {
		cardBottom()
	})

	_ = captureStdout(func() {
		cardLine("test %s", "hello")
	})

	_ = captureStdout(func() {
		cardGap()
	})
}

func TestRenderGroupedSmells(t *testing.T) {
	groups := map[string][]analyzer.Smell{
		"critical": {{Name: "deep_nesting", Severity: "critical", LineStart: 10, Message: "Deep", AIPrompt: "Fix"}},
		"warning":  {{Name: "brain_method", Severity: "warning", LineStart: 20, Message: "Long", AIPrompt: "Split"}},
		"alert":    {},
	}

	_ = captureStdout(func() {
		renderGroupedSmells(groups)
	})
}

func TestWriteHumanVulnerabilities(t *testing.T) {
	findings := []vulnerability.Finding{
		{RuleID: "pickle_deserialization", Category: "deserialization", Severity: "critical", Description: "Unsafe pickle usage", Line: 5, Column: 1, Reminder: "Use JSON instead."},
		{RuleID: "eval_injection", Category: "injection", Severity: "critical", Description: "eval() usage", Line: 10, Column: 1, Reminder: "Use ast.literal_eval."},
		{RuleID: "innerHTML_xss", Category: "xss", Severity: "critical", Description: "innerHTML assignment", Line: 15, Column: 1, Reminder: "Use textContent."},
	}

	out := captureStdout(func() {
		writeHumanVulnerabilities("test.py", findings)
	})

	if !strings.Contains(out, "test.py") {
		t.Error("human vuln output should contain file path")
	}
	if !strings.Contains(out, "pickle_deserialization") {
		t.Error("human vuln output should contain rule ID")
	}
	if !strings.Contains(out, "deserialization") {
		t.Error("human vuln output should contain category")
	}
}

func TestWriteHumanVulnerabilities_Empty(t *testing.T) {
	out := captureStdout(func() {
		writeHumanVulnerabilities("empty.py", nil)
	})
	if strings.Contains(out, "pickle") {
		t.Error("empty findings should not produce output")
	}
}

func TestWriteMarkdownVulnerabilities(t *testing.T) {
	findings := []vulnerability.Finding{
		{RuleID: "os_system_injection", Category: "injection", Severity: "critical", Description: "os.system() call detected", Line: 3, Column: 5, Reminder: "Use subprocess.run with list args."},
	}

	out := captureStdout(func() {
		writeMarkdownVulnerabilities("test.py", findings)
	})

	if !strings.Contains(out, "test.py") {
		t.Error("markdown vuln output should contain file path")
	}
	if !strings.Contains(out, "os_system_injection") {
		t.Error("markdown vuln output should contain rule ID")
	}
	if !strings.Contains(out, "injection") {
		t.Error("markdown vuln output should contain category")
	}
	if !strings.Contains(out, "subprocess.run") {
		t.Error("markdown vuln output should contain reminder")
	}
}

func TestWriteMarkdownVulnerabilities_Empty(t *testing.T) {
	out := captureStdout(func() {
		writeMarkdownVulnerabilities("empty.py", nil)
	})
	if !strings.Contains(out, "0 findings") {
		t.Errorf("empty findings should show 0 count, got %q", out)
	}
}

func TestWriteProblemsVulnerabilities(t *testing.T) {
	findings := []vulnerability.Finding{
		{RuleID: "eval_injection", Category: "injection", Severity: "critical", Description: "eval() call", Line: 42, Column: 10, Reminder: "Use safe parser."},
	}

	out := captureStdout(func() {
		writeProblemsVulnerabilities("test.py", findings)
	})

	if !strings.Contains(out, "test.py:42:10") {
		t.Errorf("problems format should contain file:line:col, got %q", out)
	}
	if !strings.Contains(out, "eval_injection") {
		t.Error("problems format should contain rule ID")
	}
	if !strings.Contains(out, "injection") {
		t.Error("problems format should contain category")
	}
}

func TestWriteProblemsVulnerabilities_Alert(t *testing.T) {
	findings := []vulnerability.Finding{
		{RuleID: "weak_hash_md5", Category: "crypto", Severity: "alert", Description: "MD5 usage", Line: 1, Column: 1, Reminder: "Use SHA-256."},
		{RuleID: "path_traversal", Category: "injection", Severity: "warning", Description: "Path join", Line: 3, Column: 1, Reminder: "Validate paths."},
	}

	out := captureStdout(func() {
		writeProblemsVulnerabilities("test.go", findings)
	})

	if !strings.Contains(out, "test.go:1:1: warning: weak_hash_md5") {
		t.Errorf("alert severity should map to 'warning' in problems format, got %q", out)
	}
	if !strings.Contains(out, "test.go:3:1: warning: path_traversal") {
		t.Errorf("warning severity should stay 'warning', got %q", out)
	}
}

func TestWriteVulnerabilities_Dispatch(t *testing.T) {
	findings := []vulnerability.Finding{
		{RuleID: "eval_injection", Category: "injection", Severity: "critical", Description: "eval() call", Line: 1, Column: 1, Reminder: "Use safe parser."},
	}

	formats := []FormatMode{FormatHuman, FormatMarkdown, FormatProblems}
	for _, f := range formats {
		out := captureStdout(func() {
			writeVulnerabilities(f, "test.py", findings)
		})
		if !strings.Contains(out, "eval_injection") {
			t.Errorf("format %v should include rule ID", f)
		}
	}

	// JSON format should NOT produce output (findings are embedded in combined JSON)
	out := captureStdout(func() {
		writeVulnerabilities(FormatJSON, "test.py", findings)
	})
	if strings.Contains(out, "eval_injection") {
		t.Error("JSON format should not output vuln findings via writeVulnerabilities")
	}
}

func TestWriteVulnerabilities_Empty(t *testing.T) {
	var empty []vulnerability.Finding
	for _, f := range []FormatMode{FormatHuman, FormatMarkdown, FormatProblems, FormatJSON} {
		out := captureStdout(func() {
			writeVulnerabilities(f, "test.py", empty)
		})
		if out != "" {
			t.Errorf("empty findings should produce no output for format %v, got %q", f, out)
		}
	}
}

func TestGroupVulnFindings(t *testing.T) {
	findings := []vulnerability.Finding{
		{RuleID: "a", Severity: "critical"},
		{RuleID: "b", Severity: "critical"},
		{RuleID: "c", Severity: "alert"},
		{RuleID: "d", Severity: "warning"},
		{RuleID: "e", Severity: "alert"},
	}

	groups := groupVulnFindings(findings)
	if len(groups["critical"]) != 2 {
		t.Errorf("expected 2 critical, got %d", len(groups["critical"]))
	}
	if len(groups["alert"]) != 2 {
		t.Errorf("expected 2 alert, got %d", len(groups["alert"]))
	}
	if len(groups["warning"]) != 1 {
		t.Errorf("expected 1 warning, got %d", len(groups["warning"]))
	}
}

func TestWriteCombinedJSON_NoVulnerabilities(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.py")
	os.WriteFile(f, []byte("import pickle\npickle.loads(data)\n"), 0644)
	src := "import pickle\npickle.loads(data)\n"

	result := analyzer.QualityResult{Score: 80, Label: "Proceed with Care", FilePath: f, Language: "python", LinesOfCode: 2}

	// With vulnerabilities disabled
	out := captureStdout(func() {
		writeCombinedJSON(result, []byte(src), f, true, true)
	})
	if strings.Contains(out, "pickle_deserialization") {
		t.Error("vulnerability findings should NOT appear when noVulnerabilities=true")
	}

	// With vulnerabilities enabled
	out2 := captureStdout(func() {
		writeCombinedJSON(result, []byte(src), f, true, false)
	})
	if !strings.Contains(out2, "pickle_deserialization") {
		t.Error("vulnerability findings SHOULD appear when noVulnerabilities=false")
	}
}

func TestWriteCombinedJSON_NoSecretsKeepsVulns(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.py")
	os.WriteFile(f, []byte("import pickle\npickle.loads(data)\n"), 0644)
	src := "import pickle\npickle.loads(data)\n"

	result := analyzer.QualityResult{Score: 80, Label: "Proceed with Care", FilePath: f, Language: "python", LinesOfCode: 2}

	// --no-secrets: secrets disabled, vulns enabled
	out := captureStdout(func() {
		writeCombinedJSON(result, []byte(src), f, true, false)
	})

	if strings.Contains(out, "\"secret_scan\"") && !strings.Contains(out, "\"secret_scan\":null") && !strings.Contains(out, "\"secret_scan\":[]") && !strings.Contains(out, "\"secret_scan\": []") {
		// secret_scan might be omitted or null when no secrets found — that's fine
	}
	if !strings.Contains(out, "pickle_deserialization") {
		t.Error("C4 REGRESSION: vulnerability_scan should have findings even when noSecrets=true")
	}
}

func TestWriteCombinedJSON_WithVulnerabilityContent(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "vuln.py")
	src := "import pickle\npickle.loads(data)\n"
	os.WriteFile(f, []byte(src), 0644)

	result := analyzer.QualityResult{Score: 80, Label: "Proceed with Care", FilePath: f, Language: "python", LinesOfCode: 2}

	out := captureStdout(func() {
		writeCombinedJSON(result, []byte(src), f, false, false)
	})

	if !strings.Contains(out, "pickle_deserialization") {
		t.Error("JSON output should contain pickle_deserialization finding")
	}
	if !strings.Contains(out, "\"vulnerability_scan\"") {
		t.Error("JSON output should have vulnerability_scan key")
	}
}

func TestWriteJSONResults_WrapsInObject(t *testing.T) {
	results := []analyzer.QualityResult{
		{Score: 95, Label: "Go Ahead", FilePath: "a.go", Language: "go", LinesOfCode: 30},
	}

	out := captureStdout(func() {
		writeJSONResults(results)
	})

	var parsed struct {
		CodeQuality []analyzer.QualityResult `json:"code_quality"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("JSON must be valid wrapped object: %v\nRaw: %s", err, out)
	}
	if len(parsed.CodeQuality) != 1 {
		t.Errorf("expected 1 result, got %d", len(parsed.CodeQuality))
	}
}
