package metalinter

import (
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/client9/misspell"
	"golang.org/x/tools/go/analysis"
)

func TestResolveAnalyzers_Empty(t *testing.T) {
	got := resolveAnalyzers(nil)
	if len(got) != 0 {
		t.Errorf("resolveAnalyzers(nil) = %d items, want 0", len(got))
	}
}

func TestResolveAnalyzers_NoDeps(t *testing.T) {
	a1 := &analysis.Analyzer{Name: "analyzer1"}
	a2 := &analysis.Analyzer{Name: "analyzer2"}
	got := resolveAnalyzers([]*analysis.Analyzer{a1, a2})
	if len(got) != 2 {
		t.Fatalf("resolveAnalyzers() = %d items, want 2", len(got))
	}
}

func TestResolveAnalyzers_WithDeps(t *testing.T) {
	a1 := &analysis.Analyzer{Name: "base"}
	a2 := &analysis.Analyzer{Name: "dependent", Requires: []*analysis.Analyzer{a1}}
	got := resolveAnalyzers([]*analysis.Analyzer{a2, a1})
	if len(got) != 2 {
		t.Fatalf("resolveAnalyzers() = %d items, want 2", len(got))
	}
	if got[0] != a1 {
		t.Errorf("resolveAnalyzers: expected base first, got %s", got[0].Name)
	}
	if got[1] != a2 {
		t.Errorf("resolveAnalyzers: expected dependent second, got %s", got[1].Name)
	}
}

func TestResolveAnalyzers_CircularDependency(t *testing.T) {
	a1 := &analysis.Analyzer{Name: "a1"}
	a2 := &analysis.Analyzer{Name: "a2"}
	a1.Requires = []*analysis.Analyzer{a2}
	a2.Requires = []*analysis.Analyzer{a1}
	got := resolveAnalyzers([]*analysis.Analyzer{a1, a2})
	if len(got) != 2 {
		t.Fatalf("resolveAnalyzers(circular) = %d items, want 2", len(got))
	}
}

func TestResolveAnalyzers_DeepChain(t *testing.T) {
	a1 := &analysis.Analyzer{Name: "a1"}
	a2 := &analysis.Analyzer{Name: "a2", Requires: []*analysis.Analyzer{a1}}
	a3 := &analysis.Analyzer{Name: "a3", Requires: []*analysis.Analyzer{a2}}
	got := resolveAnalyzers([]*analysis.Analyzer{a3, a1, a2})
	if len(got) != 3 {
		t.Fatalf("resolveAnalyzers() = %d items, want 3", len(got))
	}
	if got[0] != a1 || got[1] != a2 || got[2] != a3 {
		t.Errorf("resolveAnalyzers(deep) order wrong: got %s, %s, %s", got[0].Name, got[1].Name, got[2].Name)
	}
}

func TestToolNameForAnalyzer(t *testing.T) {
	tests := []struct {
		name     string
		analyzer string
		want     string
	}{
		{"ineffassign", "ineffassign", "ineffassign"},
		{"SA prefix", "SA1000", "staticcheck"},
		{"SA prefix #2", "SA4017", "staticcheck"},
		{"S prefix", "S1017", "staticcheck"},
		{"ST prefix", "ST1000", "staticcheck"},
		{"U prefix", "U1000", "staticcheck"},
		{"govet default", "printf", "govet"},
		{"govet default #2", "assign", "govet"},
		{"govet default #3", "bools", "govet"},
		{"unknown analyzer", "somecustom", "govet"},
		{"empty name", "", "govet"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &analysis.Analyzer{Name: tt.analyzer}
			if got := toolNameForAnalyzer(a); got != tt.want {
				t.Errorf("toolNameForAnalyzer(%q) = %q, want %q", tt.analyzer, got, tt.want)
			}
		})
	}
}

func TestCategorizeAnalyzer(t *testing.T) {
	tests := []struct {
		name     string
		analyzer string
		want     string
	}{
		{"ineffassign", "ineffassign", CategoryUnused},
		{"assign", "assign", CategoryUnused},
		{"unusedresult", "unusedresult", CategoryUnused},
		{"printf", "printf", CategoryCorrectness},
		{"slog", "slog", CategoryCorrectness},
		{"stdmethods", "stdmethods", CategoryCorrectness},
		{"copylock", "copylock", CategoryBug},
		{"bools", "bools", CategoryBug},
		{"nilfunc", "nilfunc", CategoryBug},
		{"nilness", "nilness", CategoryBug},
		{"atomic", "atomic", CategoryBug},
		{"loopclosure", "loopclosure", CategoryBug},
		{"lostcancel", "lostcancel", CategoryBug},
		{"composite", "composite", CategoryStyle},
		{"structtag", "structtag", CategoryStyle},
		{"SA prefix", "SA1000", CategoryCorrectness},
		{"S prefix", "S1017", CategoryStyle},
		{"ST prefix", "ST1000", CategoryStyle},
		{"U prefix", "U1000", CategoryUnused},
		{"unknown analyzer", "somecustom", CategoryCorrectness},
		{"empty name", "", CategoryCorrectness},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &analysis.Analyzer{Name: tt.analyzer}
			if got := categorizeAnalyzer(a); got != tt.want {
				t.Errorf("categorizeAnalyzer(%q) = %q, want %q", tt.analyzer, got, tt.want)
			}
		})
	}
}

func TestBuildResultOf(t *testing.T) {
	a1 := &analysis.Analyzer{Name: "dep1"}
	a2 := &analysis.Analyzer{Name: "dep2"}
	parent := &analysis.Analyzer{Name: "parent", Requires: []*analysis.Analyzer{a1, a2}}

	results := map[*analysis.Analyzer]any{
		a1: "result1",
	}

	got := buildResultOf(parent, results)
	if len(got) != 1 {
		t.Fatalf("buildResultOf() = %d items, want 1", len(got))
	}
	if got[a1] != "result1" {
		t.Errorf("buildResultOf()[a1] = %v, want 'result1'", got[a1])
	}
	if _, ok := got[a2]; ok {
		t.Errorf("buildResultOf() should not include a2 (not in results)")
	}
}

func TestBuildResultOf_Empty(t *testing.T) {
	a := &analysis.Analyzer{Name: "solo"}
	got := buildResultOf(a, nil)
	if len(got) != 0 {
		t.Errorf("buildResultOf() with nil results = %d items, want 0", len(got))
	}
}

func TestBuildResultOf_NoDeps(t *testing.T) {
	a := &analysis.Analyzer{Name: "solo"}
	got := buildResultOf(a, map[*analysis.Analyzer]any{})
	if len(got) != 0 {
		t.Errorf("buildResultOf() with no deps = %d items, want 0", len(got))
	}
}

func TestResolvePos_NilPkg(t *testing.T) {
	file, line, col := resolvePos(token.NoPos, nil)
	if file != "" || line != 0 || col != 0 {
		t.Errorf("resolvePos(nil) = (%q, %d, %d), want empty", file, line, col)
	}
}

func TestResolvePos_InvalidPos(t *testing.T) {
	file, line, col := resolvePos(token.NoPos, nil)
	if file != "" || line != 0 || col != 0 {
		t.Errorf("resolvePos(token.NoPos) = (%q, %d, %d), want empty", file, line, col)
	}
}

func TestResolvePos_NilFset(t *testing.T) {
	pos := token.Pos(10)
	file, line, col := resolvePos(pos, nil)
	if file != "" || line != 0 || col != 0 {
		t.Errorf("resolvePos(nil pkg) = (%q, %d, %d), want empty", file, line, col)
	}
}

func TestLineCol_InvalidPos(t *testing.T) {
	line, col := lineCol(nil, token.NoPos)
	if line != 0 || col != 0 {
		t.Errorf("lineCol(nil, NoPos) = (%d, %d), want (0,0)", line, col)
	}
}

func TestLineCol_ValidPos(t *testing.T) {
	fset := token.NewFileSet()
	f := fset.AddFile("test.go", -1, 100)
	f.SetLines([]int{0, 20, 40, 60})

	pos := f.Pos(25)
	line, col := lineCol(f, pos)
	if line == 0 && col == 0 {
		t.Errorf("lineCol() for valid pos returned (0,0), expected non-zero")
	}
}

func TestToRelative(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"empty string", ""},
		{"relative path", "some/relative/file.go"},
		{"dot-prefixed", "./local/file.go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toRelative(tt.path)
			if tt.path == "" {
				if got != "" {
					t.Errorf("toRelative(\"\") = %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tt.path) && got != tt.path {
				t.Errorf("toRelative(%q) = %q, expected something containing %q", tt.path, got, tt.path)
			}
		})
	}
}

func TestDiagnosticToFinding(t *testing.T) {
	fset := token.NewFileSet()
	f := fset.AddFile("test.go", -1, 100)
	f.SetLinesForContent([]byte("package p\n"))

	a := &analysis.Analyzer{Name: "printf"}
	d := analysis.Diagnostic{
		Pos:      f.Pos(0),
		Message:  "unexpected printf argument",
		Category: "printf",
	}

	finding := diagnosticToFinding(a, d, nil)
	if finding.Tool != "govet" {
		t.Errorf("Tool = %q, want govet", finding.Tool)
	}
	if finding.Code != "printf" {
		t.Errorf("Code = %q, want printf", finding.Code)
	}
	if finding.Severity != SeverityWarning {
		t.Errorf("Severity = %q, want warning", finding.Severity)
	}
	if finding.Message != "unexpected printf argument" {
		t.Errorf("Message = %q, want 'unexpected printf argument'", finding.Message)
	}
	if finding.Category != "printf" {
		t.Errorf("Category = %q, want printf (from diagnostic Category)", finding.Category)
	}
}

func TestDiagnosticToFinding_EmptyCategory(t *testing.T) {
	fset := token.NewFileSet()
	f := fset.AddFile("test.go", -1, 100)
	f.SetLinesForContent([]byte("package p\n"))

	a := &analysis.Analyzer{Name: "bools"}
	d := analysis.Diagnostic{
		Pos:     f.Pos(0),
		Message: "redundant boolean",
	}

	finding := diagnosticToFinding(a, d, nil)
	if finding.Code != "bools" {
		t.Errorf("Code = %q, want bools (analyzer name fallback)", finding.Code)
	}
	if finding.Category != CategoryBug {
		t.Errorf("Category = %q, want bug", finding.Category)
	}
}

func TestDiagnosticToFinding_StaticcheckCategory(t *testing.T) {
	fset := token.NewFileSet()
	f := fset.AddFile("test.go", -1, 100)
	f.SetLinesForContent([]byte("package p\n"))

	a := &analysis.Analyzer{Name: "SA1000"}
	d := analysis.Diagnostic{
		Pos:     f.Pos(0),
		Message: "invalid call",
	}

	finding := diagnosticToFinding(a, d, nil)
	if finding.Tool != "staticcheck" {
		t.Errorf("Tool = %q, want staticcheck", finding.Tool)
	}
	if finding.Category != CategoryCorrectness {
		t.Errorf("Category = %q, want correctness", finding.Category)
	}
}

func TestCheckFileMisspell(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "misspelled.go")
	content := []byte("package main\n\nfunc main() {\n\tprintln(\"this was occured\")\n}\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	r := misspell.New()
	findings := checkFileMisspell(path, r)

	if len(findings) == 0 {
		t.Fatal("checkFileMisspell() should detect misspelling of 'occured'")
	}
	found := false
	for _, f := range findings {
		if strings.Contains(f.Message, "occured") {
			found = true
			if f.Tool != "misspell" {
				t.Errorf("Tool = %q, want misspell", f.Tool)
			}
			if f.Severity != SeverityInfo {
				t.Errorf("Severity = %q, want info", f.Severity)
			}
			if f.Category != CategoryStyle {
				t.Errorf("Category = %q, want style", f.Category)
			}
			if f.Line <= 0 {
				t.Errorf("Line = %d, want > 0", f.Line)
			}
			break
		}
	}
	if !found {
		t.Errorf("no finding for 'occured', findings: %v", findings)
	}
}

func TestCheckFileMisspell_Multiple(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "multi.go")
	content := []byte("package main\n\n// This accomodates recieve requests\nvar x = \"occured\"\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	r := misspell.New()
	findings := checkFileMisspell(path, r)

	if len(findings) < 2 {
		t.Fatalf("checkFileMisspell() should detect multiple misspellings, got %d", len(findings))
	}
}

func TestCheckFileMisspell_Nonexistent(t *testing.T) {
	r := misspell.New()
	findings := checkFileMisspell("/nonexistent/path.go", r)
	if len(findings) != 0 {
		t.Errorf("checkFileMisspell() on nonexistent file should return empty, got %d", len(findings))
	}
}

func TestCheckDirMisspell(t *testing.T) {
	tmp := t.TempDir()
	subdir := filepath.Join(tmp, "pkg")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(subdir, "test.go")
	content := []byte("package pkg\n\n// This was occured\nvar X = 1\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	r := misspell.New()
	findings, err := checkDirMisspell(tmp, r)
	if err != nil {
		t.Fatalf("checkDirMisspell() err = %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("checkDirMisspell() should detect misspellings")
	}
}

func TestCheckDirMisspell_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	r := misspell.New()
	findings, err := checkDirMisspell(tmp, r)
	if err != nil {
		t.Fatalf("checkDirMisspell(empty) err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("checkDirMisspell(empty) should return empty, got %d", len(findings))
	}
}

func TestCheckDirMisspell_SkipsHiddenDirs(t *testing.T) {
	tmp := t.TempDir()
	hidden := filepath.Join(tmp, ".hidden")
	if err := os.Mkdir(hidden, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(hidden, "test.go")
	os.WriteFile(path, []byte("package test\n// occured\n"), 0644)

	r := misspell.New()
	findings, err := checkDirMisspell(tmp, r)
	if err != nil {
		t.Fatalf("checkDirMisspell() err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("checkDirMisspell() should skip hidden dirs, got %d findings", len(findings))
	}
}

func TestCheckDirMisspell_SkipsNonGoFiles(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "readme.txt")
	os.WriteFile(path, []byte("occured"), 0644)

	r := misspell.New()
	findings, err := checkDirMisspell(tmp, r)
	if err != nil {
		t.Fatalf("checkDirMisspell() err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("checkDirMisspell() should skip non-.go files, got %d findings", len(findings))
	}
}

func TestRunMisspell_Empty(t *testing.T) {
	findings, err := runMisspell(nil)
	if err != nil {
		t.Errorf("runMisspell(nil) err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("runMisspell(nil) should return empty, got %d", len(findings))
	}
}

func TestRunMisspell_FilePaths(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bug.go")
	os.WriteFile(path, []byte("package main\n// occured\n"), 0644)

	findings, err := runMisspell([]string{path})
	if err != nil {
		t.Fatalf("runMisspell() err = %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("runMisspell() should detect misspellings")
	}
}

func TestRunMisspell_NonGoFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "notes.txt")
	os.WriteFile(path, []byte("occured"), 0644)

	findings, err := runMisspell([]string{path})
	if err != nil {
		t.Fatalf("runMisspell() err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("runMisspell() should skip non-.go files, got %d findings", len(findings))
	}
}

func TestRunMisspell_DirPaths(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.go")
	os.WriteFile(path, []byte("package main\n// occured\n"), 0644)

	findings, err := runMisspell([]string{tmp})
	if err != nil {
		t.Fatalf("runMisspell(dir) err = %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("runMisspell(dir) should detect misspellings in directory")
	}
}

func TestRunMisspell_NonexistentPath(t *testing.T) {
	findings, err := runMisspell([]string{"/nonexistent/file.go"})
	if err != nil {
		t.Fatalf("runMisspell(nonexistent) err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("runMisspell(nonexistent) should return empty, got %d", len(findings))
	}
}

func TestRunGofmt_FilePaths(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "good.go")
	os.WriteFile(path, []byte("package test\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"), 0644)

	findings, err := runGofmt([]string{path})
	if err != nil {
		t.Fatalf("runGofmt() err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("runGofmt() on well-formatted file should return empty, got %d", len(findings))
	}
}

func TestRunGofmt_BadlyFormatted(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.go")
	os.WriteFile(path, []byte("package test\nfunc main(){println(\"x\")}"), 0644)

	findings, err := runGofmt([]string{path})
	if err != nil {
		t.Fatalf("runGofmt() err = %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("runGofmt() should detect badly formatted file")
	}
}

func TestRunGofmt_SkipsNonGo(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "notes.txt")
	os.WriteFile(path, []byte("hello"), 0644)

	findings, err := runGofmt([]string{path})
	if err != nil {
		t.Fatalf("runGofmt() err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("runGofmt() should skip non-.go files, got %d findings", len(findings))
	}
}

func TestCheckFileGofmt_ParseError(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "parse_err.go")
	content := []byte("package test\n\nfunc main() {\n\tprintln(\"hello\"\n}\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	finding, err := checkFileGofmt(path)
	if err != nil {
		t.Fatalf("checkFileGofmt(parse_err) err = %v", err)
	}
	if finding == nil {
		t.Fatal("checkFileGofmt(parse_err) should return a finding for parse errors")
	}
	if finding.Tool != "gofmt" {
		t.Errorf("Tool = %q, want gofmt", finding.Tool)
	}
	if finding.Severity != SeverityWarning {
		t.Errorf("Severity = %q, want warning", finding.Severity)
	}
	if finding.Category != CategoryFormatting {
		t.Errorf("Category = %q, want formatting", finding.Category)
	}
	if !strings.Contains(finding.Message, "parse errors") {
		t.Errorf("Message = %q, should contain 'parse errors'", finding.Message)
	}
}

func TestRunGofmt_NonexistentPath(t *testing.T) {
	findings, err := runGofmt([]string{"/nonexistent/file.go"})
	if err != nil {
		t.Fatalf("runGofmt(nonexistent) err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("runGofmt(nonexistent) should return empty, got %d", len(findings))
	}
}
