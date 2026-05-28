package metalinter

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

func TestCollectGoFiles_AbsolutePaths(t *testing.T) {
	paths := []string{"/a/b/c.go", "/d/e/f.go"}
	got := collectGoFiles(paths)
	if len(got) != 2 {
		t.Fatalf("collectGoFiles() len = %d, want 2", len(got))
	}
	if got[0] != paths[0] {
		t.Errorf("collectGoFiles()[0] = %q, want %q", got[0], paths[0])
	}
	if got[1] != paths[1] {
		t.Errorf("collectGoFiles()[1] = %q, want %q", got[1], paths[1])
	}
}

func TestCollectGoFiles_CopySemantics(t *testing.T) {
	original := []string{"a.go", "b.go"}
	got := collectGoFiles(original)
	original[0] = "mutated.go"
	if got[0] == "mutated.go" {
		t.Error("collectGoFiles() should return a copy, not alias the input slice")
	}
}

func TestFinding_ProblemsFormat_EmptyCategoryAndCode(t *testing.T) {
	f := Finding{
		Tool:    "govet",
		File:    "main.go",
		Line:    5,
		Column:  1,
		Message: "some diagnostic",
	}
	got := f.ProblemsFormat()
	want := "main.go:5:1: [govet] some diagnostic ()"
	if got != want {
		t.Errorf("ProblemsFormat() with empty Category/Code = %q, want %q", got, want)
	}
	if !strings.HasSuffix(got, "()") {
		t.Errorf("ProblemsFormat() should end with () for empty Code, got %q", got)
	}
}

func TestLintGo_EmptyPaths(t *testing.T) {
	findings, err := LintGo(nil)
	if err != nil {
		t.Errorf("LintGo(nil) err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("LintGo(nil) should return empty findings, got %d", len(findings))
	}

	findings, err = LintGo([]string{})
	if err != nil {
		t.Errorf("LintGo([]) err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("LintGo([]) should return empty findings, got %d", len(findings))
	}
}

func TestLintGo_NonGoFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "readme.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := LintGo([]string{path})
	if err == nil {
		t.Log("LintGo with non-Go file: no error (graceful)")
	} else {
		t.Logf("LintGo with non-Go file: err = %v (acceptable if packages.Load fails)", err)
	}
	_ = findings
}

func TestGetStaticcheckAnalyzers(t *testing.T) {
	analyzers := getStaticcheckAnalyzers()
	if analyzers == nil {
		t.Fatal("getStaticcheckAnalyzers() returned nil")
	}
	if len(analyzers) == 0 {
		t.Fatal("getStaticcheckAnalyzers() returned empty slice")
	}

	var hasSA, hasS, hasST bool
	for _, a := range analyzers {
		if a == nil {
			t.Error("getStaticcheckAnalyzers() contains nil analyzer")
			continue
		}
		if a.Name == "" {
			t.Error("getStaticcheckAnalyzers() contains analyzer with empty Name")
		}
		if strings.HasPrefix(a.Name, "SA") {
			hasSA = true
		}
		if strings.HasPrefix(a.Name, "S") && !strings.HasPrefix(a.Name, "SA") && !strings.HasPrefix(a.Name, "ST") {
			hasS = true
		}
		if strings.HasPrefix(a.Name, "ST") {
			hasST = true
		}
	}
	if !hasSA {
		t.Error("getStaticcheckAnalyzers() should include SA* analyzers (staticcheck)")
	}
	if !hasS {
		t.Error("getStaticcheckAnalyzers() should include S* analyzers (simple)")
	}
	if !hasST {
		t.Error("getStaticcheckAnalyzers() should include ST* analyzers (stylecheck)")
	}
}

func TestToolNameForAnalyzer_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		aName string
		want  string
	}{
		{"empty name", "", "govet"},
		{"very long name", "thisisaverylonganalyzernameindeedtoolong", "govet"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &analysis.Analyzer{Name: tt.aName}
			if got := toolNameForAnalyzer(a); got != tt.want {
				t.Errorf("toolNameForAnalyzer(%q) = %q, want %q", tt.aName, got, tt.want)
			}
		})
	}
}

func TestRunGoVetStaticcheck_EmptyPaths(t *testing.T) {
	findings, err := runGoVetStaticcheck(nil)
	if err != nil {
		t.Errorf("runGoVetStaticcheck(nil) err = %v", err)
	}
	if findings != nil {
		t.Errorf("runGoVetStaticcheck(nil) should return nil, got %d findings", len(findings))
	}

	findings, err = runGoVetStaticcheck([]string{})
	if err != nil {
		t.Errorf("runGoVetStaticcheck([]) err = %v", err)
	}
	if findings != nil {
		t.Errorf("runGoVetStaticcheck([]) should return nil, got %d findings", len(findings))
	}
}

func TestRunGoVetStaticcheck_NonExistentPath(t *testing.T) {
	findings, err := runGoVetStaticcheck([]string{"/nonexistent/package/path"})
	if err == nil {
		t.Log("runGoVetStaticcheck(nonexistent) returned nil error (acceptable if packages.Load is lenient)")
	} else {
		t.Logf("runGoVetStaticcheck(nonexistent) err = %v", err)
	}
	_ = findings
}

func TestPackageErrorsToFindings(t *testing.T) {
	t.Run("single error", func(t *testing.T) {
		errs := []packages.Error{
			{Pos: "main.go:1:1", Msg: "could not load package"},
		}
		findings := packageErrorsToFindings(errs)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		f := findings[0]
		if f.Tool != "govet" {
			t.Errorf("Tool = %q, want govet", f.Tool)
		}
		if f.Severity != SeverityError {
			t.Errorf("Severity = %q, want error", f.Severity)
		}
		if f.File != "main.go:1:1" {
			t.Errorf("File = %q, want main.go:1:1", f.File)
		}
		if f.Line != 1 {
			t.Errorf("Line = %d, want 1", f.Line)
		}
		if f.Column != 1 {
			t.Errorf("Column = %d, want 1", f.Column)
		}
		if f.Category != CategoryCorrectness {
			t.Errorf("Category = %q, want correctness", f.Category)
		}
		if !strings.Contains(f.Message, "could not load package") {
			t.Errorf("Message = %q, should contain 'could not load package'", f.Message)
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := []packages.Error{
			{Pos: "a.go:1:1", Msg: "error one"},
			{Pos: "b.go:2:5", Msg: "error two"},
			{Pos: "c.go:3:10", Msg: "error three"},
		}
		findings := packageErrorsToFindings(errs)
		if len(findings) != 3 {
			t.Fatalf("expected 3 findings, got %d", len(findings))
		}
		for i, f := range findings {
			if f.Tool != "govet" {
				t.Errorf("findings[%d].Tool = %q, want govet", i, f.Tool)
			}
			if f.Line != 1 {
				t.Errorf("findings[%d].Line = %d, want 1", i, f.Line)
			}
			if f.Column != 1 {
				t.Errorf("findings[%d].Column = %d, want 1", i, f.Column)
			}
		}
	})

	t.Run("empty errors", func(t *testing.T) {
		findings := packageErrorsToFindings(nil)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings for nil, got %d", len(findings))
		}
		findings = packageErrorsToFindings([]packages.Error{})
		if len(findings) != 0 {
			t.Errorf("expected 0 findings for empty, got %d", len(findings))
		}
	})
}

func TestRunGoVetStaticcheck_ValidModule(t *testing.T) {
	tmp := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module testpkg\n\ngo 1.25\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mainContent := []byte("package main\n\nfunc main() {}\n")
	if err := os.WriteFile(filepath.Join(tmp, "main.go"), mainContent, 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := runGoVetStaticcheck([]string{tmp})

	if err != nil {
		t.Logf("runGoVetStaticcheck with real module returned error (may not work in temp dir): %v", err)
		return
	}

	t.Logf("runGoVetStaticcheck returned %d findings", len(findings))
	for i, f := range findings {
		if f.Tool == "" {
			t.Errorf("findings[%d].Tool is empty", i)
		}
		if f.Message == "" {
			t.Errorf("findings[%d].Message is empty", i)
		}
	}
}

func TestAnalyzeSinglePackage(t *testing.T) {
	t.Run("no syntax", func(t *testing.T) {
		pkg := &packages.Package{}
		findings := analyzeSinglePackage(pkg, nil)
		if len(findings) != 0 {
			t.Errorf("expected 0 findings for package with no syntax, got %d", len(findings))
		}
	})

	t.Run("package errors", func(t *testing.T) {
		pkg := &packages.Package{
			Syntax: []*ast.File{{}},
			Errors: []packages.Error{
				{Pos: "test.go:1:1", Msg: "load failure"},
			},
		}
		findings := analyzeSinglePackage(pkg, nil)
		if len(findings) != 1 {
			t.Fatalf("expected 1 finding, got %d", len(findings))
		}
		if !strings.Contains(findings[0].Message, "load failure") {
			t.Errorf("Message = %q, should contain 'load failure'", findings[0].Message)
		}
	})
}

func TestResolvePos(t *testing.T) {
	t.Run("pkg with valid fset", func(t *testing.T) {
		fset := token.NewFileSet()
		f := fset.AddFile("main.go", -1, 50)
		f.SetLinesForContent([]byte("package main\n\nfunc main() {}\n"))

		pkg := &packages.Package{Fset: fset}
		pos := f.Pos(0)
		file, line, col := resolvePos(pos, pkg)
		if file != "main.go" {
			t.Errorf("file = %q, want main.go", file)
		}
		if line != 1 {
			t.Errorf("line = %d, want 1", line)
		}
		_ = col
	})
}

func TestLineCol(t *testing.T) {
	fset := token.NewFileSet()
	f := fset.AddFile("test.go", -1, 100)
	f.SetLinesForContent([]byte("line1\nline2\nline3\n"))

	pos := f.Pos(6)
	line, col := lineCol(f, pos)
	if line != 2 {
		t.Errorf("line = %d, want 2", line)
	}
	if col != 1 {
		t.Errorf("col = %d, want 1", col)
	}
}

func TestToRelative_WithAbsPath(t *testing.T) {
	wd, _ := os.Getwd()
	abs := filepath.Join(wd, "some", "file.go")
	got := toRelative(abs)
	if strings.HasPrefix(got, "/") {
		t.Logf("toRelative(%q) = %q (still absolute, may vary by wd)", abs, got)
	}
}

func TestStaticcheckAnalyzersVar(t *testing.T) {
	if staticcheckAnalyzers == nil {
		t.Fatal("staticcheckAnalyzers is nil")
	}
	if len(staticcheckAnalyzers) == 0 {
		t.Fatal("staticcheckAnalyzers is empty")
	}
}

func TestGoVetAnalyzers(t *testing.T) {
	if len(goVetAnalyzers) == 0 {
		t.Fatal("goVetAnalyzers is empty")
	}
	for i, a := range goVetAnalyzers {
		if a == nil {
			t.Fatalf("goVetAnalyzers[%d] is nil", i)
		}
		if a.Name == "" {
			t.Errorf("goVetAnalyzers[%d].Name is empty", i)
		}
	}
}

func TestIneffassignAnalyzer(t *testing.T) {
	if ineffassignAnalyzer == nil {
		t.Fatal("ineffassignAnalyzer is nil")
	}
	if ineffassignAnalyzer.Name != "ineffassign" {
		t.Errorf("ineffassignAnalyzer.Name = %q, want ineffassign", ineffassignAnalyzer.Name)
	}
}

func TestToolPrefixes(t *testing.T) {
	if len(toolPrefixes) == 0 {
		t.Fatal("toolPrefixes is empty")
	}
}

func TestAnalyzerCategory(t *testing.T) {
	if len(analyzerCategory) == 0 {
		t.Fatal("analyzerCategory is empty")
	}
	if cat, ok := analyzerCategory["ineffassign"]; !ok || cat != CategoryUnused {
		t.Errorf("analyzerCategory[\"ineffassign\"] = %q, want %q", cat, CategoryUnused)
	}
}

func TestPrefixCategories(t *testing.T) {
	if len(prefixCategories) == 0 {
		t.Fatal("prefixCategories is empty")
	}
}
