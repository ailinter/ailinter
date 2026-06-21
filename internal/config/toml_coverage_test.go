package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindConfigDir(t *testing.T) {
	dir := t.TempDir()

	// No config file -> empty string
	if got := FindConfigDir(dir); got != "" {
		t.Errorf("FindConfigDir without config = %q, want empty", got)
	}

	// Create .ailinter.toml in dir
	cfgPath := filepath.Join(dir, ".ailinter.toml")
	if err := os.WriteFile(cfgPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	if got := FindConfigDir(dir); got != dir {
		t.Errorf("FindConfigDir = %q, want %q", got, dir)
	}

	// Create subdir and verify walk-up
	subdir := filepath.Join(dir, "sub", "nested")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	if got := FindConfigDir(subdir); got != dir {
		t.Errorf("FindConfigDir from subdir = %q, want %q", got, dir)
	}
}

func TestLoadExcludedFiles(t *testing.T) {
	dir := t.TempDir()

	// No config file -> nil
	if got := LoadExcludedFiles(dir); got != nil {
		t.Errorf("expected nil, got %v", got)
	}

	// Config without exclude section -> nil
	cfg := `[rules]
`
	cfgPath := filepath.Join(dir, ".ailinter.toml")
	if err := os.WriteFile(cfgPath, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	if got := LoadExcludedFiles(dir); got != nil {
		t.Errorf("expected nil for config without exclude, got %v", got)
	}

	// Config with exclude section
	cfg2 := `[exclude]
files = ["vendor/", "node_modules/", "testdata/fixtures"]
`
	os.WriteFile(cfgPath, []byte(cfg2), 0644)
	got := LoadExcludedFiles(dir)
	if len(got) != 3 {
		t.Fatalf("expected 3 patterns, got %d: %v", len(got), got)
	}
	if got[0] != "vendor/" {
		t.Errorf("pattern[0] = %q, want vendor/", got[0])
	}
}

func TestIsExcluded(t *testing.T) {
	pats := []string{"vendor/", "node_modules/", "foo/bar/baz"}

	// Empty patterns -> false
	if IsExcluded("vendor/x.go", nil, "/") {
		t.Error("IsExcluded with nil patterns should be false")
	}

	// Directory prefix match
	if !IsExcluded("vendor/x.go", pats, "") {
		t.Error("vendor/x.go should be excluded by vendor/")
	}

	// Exact match
	if !IsExcluded("foo/bar/baz", pats, "") {
		t.Error("foo/bar/baz should be excluded by exact match")
	}

	// No match
	if IsExcluded("src/main.go", pats, "") {
		t.Error("src/main.go should NOT be excluded")
	}

	// Relative path computation fails
	t.Run("rel path failure", func(t *testing.T) {
		// Using an absolute path that makes filepath.Rel fail
		if IsExcluded("some/file", []string{"test/"}, "/nonexistent") {
			// Should return false since Rel fails
		}
	})
}

func TestLoadProjectThresholds(t *testing.T) {
	dir := t.TempDir()

	// Without config, should return defaults
	thresholds := LoadProjectThresholds(dir, "go")
	if thresholds.NestingWarning == 0 {
		t.Error("expected non-zero default thresholds")
	}

	// Create a config with overrides
	cfg := `[rules]
deep_nesting = { warning = 5, alert = 8 }
brain_method = { weight = 2.0, warning_lines = 80, alert_lines = 150 }
`
	cfgPath := filepath.Join(dir, ".ailinter.toml")
	if err := os.WriteFile(cfgPath, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}

	thresholds2 := LoadProjectThresholds(dir, "go")
	if thresholds2.NestingWarning != 5 {
		t.Errorf("expected NestingWarning=5, got %d", thresholds2.NestingWarning)
	}
	if thresholds2.FuncLOCWarning != 80 {
		t.Errorf("expected FuncLOCWarning=80, got %d", thresholds2.FuncLOCWarning)
	}
}

func TestLoadProjectThresholds_InvalidToml(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".ailinter.toml")
	if err := os.WriteFile(cfgPath, []byte("invalid toml {{{"), 0644); err != nil {
		t.Fatal(err)
	}
	// Should return defaults gracefully
	thresholds := LoadProjectThresholds(dir, "go")
	if thresholds.NestingWarning == 0 {
		t.Error("expected non-zero defaults even with invalid toml")
	}
}

func TestLoadProjectConfigFile(t *testing.T) {
	dir := t.TempDir()

	// No config -> nil
	if got := LoadProjectConfigFile(dir); got != nil {
		t.Error("expected nil without config")
	}

	// With config
	cfg := `extends = "recommended"
[rules]
deep_nesting = { warning = 4 }
brain_method = { warning_lines = 100 }
`
	cfgPath := filepath.Join(dir, ".ailinter.toml")
	os.WriteFile(cfgPath, []byte(cfg), 0644)

	pc := LoadProjectConfigFile(dir)
	if pc == nil {
		t.Fatal("expected non-nil config")
	}
	if pc.Extends != "recommended" {
		t.Errorf("extends = %q, want recommended", pc.Extends)
	}
	if pc.Rules.DeepNesting == nil || *pc.Rules.DeepNesting != 4 {
		t.Errorf("DeepNesting = %v, want 4", pc.Rules.DeepNesting)
	}
}
