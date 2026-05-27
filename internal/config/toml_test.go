package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ailinter/ailinter/internal/config"
)

func TestLoadProjectThresholds_FindsConfigWalkingUp(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "deep", "nested", "src")
	os.MkdirAll(subdir, 0755)

	cfg := `
extends = "default"

[rules]
deep_nesting = { weight = 1.0, warning = 10, alert = 15 }
brain_method = { warning_lines = 120, alert_lines = 400 }
`
	os.WriteFile(filepath.Join(dir, ".ailinter.toml"), []byte(cfg), 0644)

	thresholds := config.LoadProjectThresholds(subdir, "go")
	if thresholds.NestingWarning != 10 {
		t.Errorf("NestingWarning should be 10 from config, got %d", thresholds.NestingWarning)
	}
	if thresholds.NestingAlert != 15 {
		t.Errorf("NestingAlert should be 15 from config, got %d", thresholds.NestingAlert)
	}
	if thresholds.FuncLOCWarning != 120 {
		t.Errorf("FuncLOCWarning should be 120 from config, got %d", thresholds.FuncLOCWarning)
	}
}

func TestLoadProjectThresholds_NoConfigFallsBackToDefaults(t *testing.T) {
	dir := t.TempDir()
	thresholds := config.LoadProjectThresholds(dir, "go")
	if thresholds.NestingWarning != 4 {
		t.Errorf("NestingWarning default for Go should be 4, got %d", thresholds.NestingWarning)
	}
	if thresholds.FuncLOCWarning != 80 {
		t.Errorf("FuncLOCWarning default for Go should be 80, got %d", thresholds.FuncLOCWarning)
	}
}

func TestLoadProjectThresholds_InvalidTomlFallsBack(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".ailinter.toml"), []byte("this is not valid toml }}}}"), 0644)
	thresholds := config.LoadProjectThresholds(dir, "go")
	if thresholds.NestingWarning != 4 {
		t.Errorf("should fall back to defaults on invalid TOML, got %d", thresholds.NestingWarning)
	}
}

func TestLoadProjectThresholds_PartialOverride(t *testing.T) {
	dir := t.TempDir()
	cfg := `
extends = "default"

[rules]
deep_nesting = { warning = 6 }
`
	os.WriteFile(filepath.Join(dir, ".ailinter.toml"), []byte(cfg), 0644)
	thresholds := config.LoadProjectThresholds(dir, "go")
	if thresholds.NestingWarning != 6 {
		t.Errorf("NestingWarning should be overridden to 6, got %d", thresholds.NestingWarning)
	}
	// Alert should stay at default since not overridden
	if thresholds.NestingAlert != 5 {
		t.Errorf("NestingAlert should stay at default 5, got %d", thresholds.NestingAlert)
	}
}

func TestLoadProjectThresholds_AllDetectors(t *testing.T) {
	dir := t.TempDir()
	cfg := `
extends = "default"

[rules]
deep_nesting = { warning = 5, alert = 10 }
brain_method = { warning_lines = 100, alert_lines = 500 }
file_bloat = { warning_lines = 1500, alert_lines = 3000 }
complex_conditional = { branches_warning = 4, branches_alert = 15 }
cyclomatic_complexity = { warning = 12, alert = 25 }
bumpy_road = { bumps_warning = 3 }
long_parameter_list = { warning = 6, alert = 10 }
lazy_element = { min_lines = 5 }
paragraph_of_code = { max_consecutive = 30 }
excessive_comments = { ratio = 0.5 }
global_data = { warning = 10 }
long_scope_variable = { min_lines = 75 }
duplicated_code = { min_lines = 15, min_similarity = 0.85 }
long_switch = { warning = 15, alert = 30 }
`
	os.WriteFile(filepath.Join(dir, ".ailinter.toml"), []byte(cfg), 0644)
	thresholds := config.LoadProjectThresholds(dir, "go")

	checks := []struct {
		name     string
		got, want int
	}{
		{"NestingWarning", thresholds.NestingWarning, 5},
		{"NestingAlert", thresholds.NestingAlert, 10},
		{"FuncLOCWarning", thresholds.FuncLOCWarning, 100},
		{"FuncLOCAlert", thresholds.FuncLOCAlert, 500},
		{"FileLOCWarning", thresholds.FileLOCWarning, 1500},
		{"FileLOCAlert", thresholds.FileLOCAlert, 3000},
		{"BumpyRoadBumpsWarning", thresholds.BumpyRoadBumpsWarning, 3},
		{"MaxArgumentsWarn", thresholds.MaxArgumentsWarn, 6},
		{"MaxArgumentsAlert", thresholds.MaxArgumentsAlert, 10},
		{"LazyMinLines", thresholds.LazyMinLines, 5},
		{"ParagraphMaxConsecutive", thresholds.ParagraphMaxConsecutive, 30},
		{"GlobalDataWarning", thresholds.GlobalDataWarning, 10},
		{"LongScopeVarLines", thresholds.LongScopeVarLines, 75},
		{"ComplexCondBranchesWarn", thresholds.ComplexCondBranchesWarn, 4},
		{"ComplexCondBranchesAlert", thresholds.ComplexCondBranchesAlert, 15},
		{"FuncCCWarning", thresholds.FuncCCWarning, 12},
		{"FuncCCAlert", thresholds.FuncCCAlert, 25},
		{"DupMinLines", thresholds.DupMinLines, 15},
		{"LongSwitchWarn", thresholds.LongSwitchWarn, 15},
		{"LongSwitchAlert", thresholds.LongSwitchAlert, 30},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
	// Float checks
	if thresholds.CommentRatioWarning != 0.5 {
		t.Errorf("CommentRatioWarning = %f, want 0.5", thresholds.CommentRatioWarning)
	}
	if thresholds.DupMinSimilarity != 0.85 {
		t.Errorf("DupMinSimilarity = %f, want 0.85", thresholds.DupMinSimilarity)
	}
}

func TestLoadProjectThresholds_EmptyFileDefaults(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".ailinter.toml"), []byte(""), 0644)
	thresholds := config.LoadProjectThresholds(dir, "go")
	if thresholds.NestingWarning != 4 {
		t.Errorf("empty config should fall back to defaults, got %d", thresholds.NestingWarning)
	}
}

func TestLoadProjectThresholds_PythonDefaults(t *testing.T) {
	dir := t.TempDir()
	thresholds := config.LoadProjectThresholds(dir, "python")
	if thresholds.FuncLOCWarning != 70 {
		t.Errorf("Python FuncLOCWarning default should be 70, got %d", thresholds.FuncLOCWarning)
	}
	if thresholds.NestingWarning != 4 {
		t.Errorf("Python NestingWarning default should be 4, got %d", thresholds.NestingWarning)
	}
}

func TestLoadProjectThresholds_JavaScriptDefaults(t *testing.T) {
	dir := t.TempDir()
	thresholds := config.LoadProjectThresholds(dir, "javascript")
	if thresholds.NestingWarning != 3 {
		t.Errorf("JavaScript NestingWarning default should be 3, got %d", thresholds.NestingWarning)
	}
	if thresholds.FuncLOCWarning != 60 {
		t.Errorf("JavaScript FuncLOCWarning default should be 60, got %d", thresholds.FuncLOCWarning)
	}
}
