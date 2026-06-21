package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NotFound(t *testing.T) {
	// Override configPath to a nonexistent path
	oldPath := configPath
	configPath = filepath.Join(t.TempDir(), "nonexistent.json")
	defer func() { configPath = oldPath }()

	c, err := Load()
	if err != nil {
		t.Fatalf("Load() on missing file should return defaults, got error: %v", err)
	}
	if c == nil {
		t.Fatal("Load() should return non-nil config")
	}
}

func TestLoad_Save_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	oldPath := configPath
	configPath = filepath.Join(dir, "ailinter_config.json")
	defer func() { configPath = oldPath }()

	c := &Config{
		AccessToken:  "test-token",
		DefaultPath:  "/home/test",
		Language:     "python",
		EnabledTools: []string{"govet", "staticcheck"},
		ReadOnly:     true,
	}

	if err := Save(c); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file was not created: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after save failed: %v", err)
	}

	if loaded.AccessToken != "test-token" {
		t.Errorf("AccessToken = %q, want test-token", loaded.AccessToken)
	}
	if loaded.Language != "python" {
		t.Errorf("Language = %q, want python", loaded.Language)
	}
	if len(loaded.EnabledTools) != 2 {
		t.Errorf("EnabledTools = %v, want 2 tools", loaded.EnabledTools)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	oldPath := configPath
	configPath = filepath.Join(dir, "bad_config.json")
	defer func() { configPath = oldPath }()

	os.WriteFile(configPath, []byte("{invalid json}"), 0644)

	_, err := Load()
	if err == nil {
		t.Error("Load() with invalid JSON should return error")
	}
}

func TestGet(t *testing.T) {
	dir := t.TempDir()
	oldPath := configPath
	configPath = filepath.Join(dir, "get_test.json")
	defer func() { configPath = oldPath }()

	// Create a config with a token
	c := &Config{
		AccessToken: "secret-token",
		Language:    "go",
	}
	Save(c)

	result, err := Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	// Token should be redacted
	if contains(result, "secret-token") {
		t.Errorf("Get() should redact access token, got: %s", result)
	}
	// Should contain language
	if !contains(result, "go") {
		t.Errorf("Get() should contain language setting, got: %s", result)
	}
}

func TestSet(t *testing.T) {
	dir := t.TempDir()
	oldPath := configPath
	configPath = filepath.Join(dir, "set_test.json")
	defer func() { configPath = oldPath }()

	// Set a value
	result, err := Set("language", "python")
	if err != nil {
		t.Fatalf("Set(language) failed: %v", err)
	}
	if !contains(result, "python") {
		t.Errorf("result should contain python, got: %s", result)
	}

	// Verify it persisted
	loaded, _ := Load()
	if loaded.Language != "python" {
		t.Errorf("Language = %q after Set, want python", loaded.Language)
	}
}

func TestSet_UnknownKey(t *testing.T) {
	dir := t.TempDir()
	oldPath := configPath
	configPath = filepath.Join(dir, "set_unknown.json")
	defer func() { configPath = oldPath }()

	_, err := Set("nonexistent_key", "value")
	if err == nil {
		t.Error("Set with unknown key should return error")
	}
}

func TestSet_BoolKeys(t *testing.T) {
	dir := t.TempDir()
	oldPath := configPath
	configPath = filepath.Join(dir, "set_bool.json")
	defer func() { configPath = oldPath }()

	// Test read_only
	Set("read_only", "true")
	loaded, _ := Load()
	if !loaded.ReadOnly {
		t.Error("ReadOnly should be true after Set")
	}

	// Test disable_git
	Set("disable_git", "1")
	loaded, _ = Load()
	if !loaded.DisableGit {
		t.Error("DisableGit should be true after Set")
	}

	// Test enabled_tools
	Set("enabled_tools", "govet, staticcheck, misspell")
	loaded, _ = Load()
	if len(loaded.EnabledTools) != 3 {
		t.Errorf("Expected 3 tools, got %v", loaded.EnabledTools)
	}
}

func TestSet_EmptyTools(t *testing.T) {
	dir := t.TempDir()
	oldPath := configPath
	configPath = filepath.Join(dir, "set_tools.json")
	defer func() { configPath = oldPath }()

	Set("enabled_tools", "")
	loaded, _ := Load()
	if loaded.EnabledTools != nil {
		t.Errorf("Empty tools should be nil, got %v", loaded.EnabledTools)
	}

	Set("enabled_tools", "*")
	loaded, _ = Load()
	if loaded.EnabledTools != nil {
		t.Errorf("Wildcard tools should be nil, got %v", loaded.EnabledTools)
	}
}

func TestGetConfig(t *testing.T) {
	dir := t.TempDir()
	oldPath := configPath
	configPath = filepath.Join(dir, "get_config.json")
	defer func() { configPath = oldPath }()

	c := &Config{AccessToken: "my-token", DisableGit: true}
	Save(c)

	cfg, err := GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() failed: %v", err)
	}
	if cfg.AccessToken != "***" {
		t.Errorf("AccessToken should be redacted, got %q", cfg.AccessToken)
	}
}

func TestSetAndGet(t *testing.T) {
	dir := t.TempDir()
	oldPath := configPath
	configPath = filepath.Join(dir, "set_and_get.json")
	defer func() { configPath = oldPath }()

	cfg, err := SetAndGet("onprem_url", "https://example.com")
	if err != nil {
		t.Fatalf("SetAndGet() failed: %v", err)
	}
	if cfg.OnPremURL != "https://example.com" {
		t.Errorf("OnPremURL = %q, want https://example.com", cfg.OnPremURL)
	}
}

func TestPath(t *testing.T) {
	// Path should not be empty
	if p := Path(); p == "" {
		t.Error("Path() should return non-empty path")
	}
}

func TestParseToolList(t *testing.T) {
	tests := []struct {
		input string
		want  int // expected length
	}{
		{"", 0},
		{"*", 0},
		{"govet", 1},
		{"govet, staticcheck", 2},
		{"govet, , staticcheck", 2},
	}
	for _, tc := range tests {
		got := parseToolList(tc.input)
		if len(got) != tc.want {
			t.Errorf("parseToolList(%q) = %v (len %d), want len %d", tc.input, got, len(got), tc.want)
		}
	}
}

func TestSplitAndTrim(t *testing.T) {
	got := splitAndTrim("a, b,  c  ,,d")
	if len(got) != 4 {
		t.Fatalf("expected 4 items, got %d: %v", len(got), got)
	}
	if got[0] != "a" || got[1] != "b" || got[2] != "c" || got[3] != "d" {
		t.Errorf("unexpected items: %v", got)
	}
}

// contains reports whether s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
