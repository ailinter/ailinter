package metalinter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFinding_ProblemsFormat(t *testing.T) {
	f := Finding{
		Tool:    "gofmt",
		Code:    "formatting",
		File:    "test.go",
		Line:    10,
		Column:  5,
		Message: "file is not gofmt-ed",
	}
	want := "test.go:10:5: [gofmt] file is not gofmt-ed (formatting)"
	if got := f.ProblemsFormat(); got != want {
		t.Errorf("ProblemsFormat() = %q, want %q", got, want)
	}
}

func TestFinding_ProblemsFormat_ZeroColumn(t *testing.T) {
	f := Finding{
		Tool:    "govet",
		Code:    "SA1000",
		File:    "pkg.go",
		Line:    3,
		Column:  0,
		Message: "unused variable",
	}
	if got := f.ProblemsFormat(); !strings.Contains(got, "pkg.go:3:1:") {
		t.Errorf("ProblemsFormat() with zero column = %q, expected column 1", got)
	}
}

func TestCollectGoFiles(t *testing.T) {
	paths := []string{"a.go", "b.go", "cmd/"}
	got := collectGoFiles(paths)
	if len(got) != 3 {
		t.Errorf("collectGoFiles() len = %d, want 3", len(got))
	}
	if got[0] != "a.go" {
		t.Errorf("collectGoFiles()[0] = %s, want a.go", got[0])
	}
}

func TestCollectGoFiles_Empty(t *testing.T) {
	got := collectGoFiles(nil)
	if len(got) != 0 {
		t.Error("collectGoFiles(nil) should return empty slice")
	}
}

func TestCheckFileGofmt_WellFormatted(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "well.go")
	content := []byte("package test\n\nfunc main() {\n\tprintln(\"hello\")\n}\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	finding, err := checkFileGofmt(path)
	if err != nil {
		t.Fatalf("checkFileGofmt() err = %v", err)
	}
	if finding != nil {
		t.Errorf("checkFileGofmt() for well-formatted file should return nil, got %+v", finding)
	}
}

func TestCheckFileGofmt_BadlyFormatted(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.go")
	// Missing newline, bad indentation
	content := []byte("package test\nfunc main(){println(\"hello\")}")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	finding, err := checkFileGofmt(path)
	if err != nil {
		t.Fatalf("checkFileGofmt() err = %v", err)
	}
	if finding == nil {
		t.Fatal("checkFileGofmt() for badly formatted file should return a finding")
	}
	if finding.Tool != "gofmt" {
		t.Errorf("Tool = %s, want gofmt", finding.Tool)
	}
	if finding.Severity != SeverityWarning {
		t.Errorf("Severity = %s, want warning", finding.Severity)
	}
}

func TestCheckFileGofmt_Nonexistent(t *testing.T) {
	finding, err := checkFileGofmt("/nonexistent/file.go")
	if err != nil {
		t.Errorf("checkFileGofmt() on nonexistent file should not error, got %v", err)
	}
	if finding != nil {
		t.Errorf("checkFileGofmt() on nonexistent file should return nil finding")
	}
}

func TestCheckFileGofmt_NonGoFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "readme.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	finding, err := checkFileGofmt(path)
	if err != nil {
		t.Fatalf("checkFileGofmt() err = %v", err)
	}
	// Non-Go content may parse without error (format.Source handles it) 
	// or return a finding. Either is acceptable behavior.
	_ = finding // behavior is implementation-defined for non-.go files
}

func TestRunGofmt_EmptyPaths(t *testing.T) {
	findings, err := runGofmt(nil)
	if err != nil {
		t.Errorf("runGofmt(nil) err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("runGofmt(nil) should return empty findings, got %d", len(findings))
	}
}

func TestCheckDirGofmt_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	findings, err := checkDirGofmt(tmp)
	if err != nil {
		t.Errorf("checkDirGofmt(empty) err = %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("checkDirGofmt(empty) should return empty findings, got %d", len(findings))
	}
}

func TestCheckDirGofmt_WithFiles(t *testing.T) {
	tmp := t.TempDir()
	goodPath := filepath.Join(tmp, "good.go")
	os.WriteFile(goodPath, []byte("package test\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"), 0644)
	badPath := filepath.Join(tmp, "bad.go")
	os.WriteFile(badPath, []byte("package test\nfunc main(){println(\"x\")}"), 0644)
	// Non-Go file should be skipped
	os.WriteFile(filepath.Join(tmp, "readme.txt"), []byte("hello"), 0644)

	findings, err := checkDirGofmt(tmp)
	if err != nil {
		t.Fatalf("checkDirGofmt() err = %v", err)
	}
	if len(findings) != 1 {
		t.Errorf("checkDirGofmt() should return 1 finding (badly formatted), got %d", len(findings))
	}
}

func TestSeverityConstants(t *testing.T) {
	if SeverityError != "error" {
		t.Errorf("SeverityError = %q", SeverityError)
	}
	if SeverityWarning != "warning" {
		t.Errorf("SeverityWarning = %q", SeverityWarning)
	}
}

func TestCategoryConstants(t *testing.T) {
	if CategoryBug != "bug" {
		t.Errorf("CategoryBug = %q", CategoryBug)
	}
	if CategoryStyle != "style" {
		t.Errorf("CategoryStyle = %q", CategoryStyle)
	}
}
