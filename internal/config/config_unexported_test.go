package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_UnmarshalError(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad-config.json")
	os.WriteFile(f, []byte("{invalid json}"), 0644)

	oldPath := configPath
	configPath = f
	defer func() { configPath = oldPath }()

	_, err := Load()
	if err == nil {
		t.Error("expected unmarshal error")
	}
}

func TestSave_MkdirError(t *testing.T) {
	dir := t.TempDir()
	// Create a file where the config directory should be
	blocker := filepath.Join(dir, "blocker")
	os.WriteFile(blocker, []byte("x"), 0644)
	f := filepath.Join(blocker, "config.json")

	oldPath := configPath
	configPath = f
	defer func() { configPath = oldPath }()

	err := Save(&Config{Language: "go"})
	if err == nil {
		t.Error("expected MkdirAll error")
	}
}

func TestSet_UnknownKey(t *testing.T) {
	_, err := Set("unknown_key", "value")
	if err == nil {
		t.Error("expected error for unknown config key")
	}
}

func TestAPPDATAPath(t *testing.T) {
	oldAPPDATA := os.Getenv("APPDATA")
	oldHOME := os.Getenv("HOME")

	os.Setenv("APPDATA", "/tmp/appdata-test")
	defer os.Setenv("APPDATA", oldAPPDATA)
	defer os.Setenv("HOME", oldHOME)

	// Reset init to re-compute path
	oldPath := configPath
	defer func() { configPath = oldPath }()
	configPath = filepath.Join(os.Getenv("APPDATA"), "ailinter", "config.json")

	p := Path()
	if p == "" {
		t.Error("APPDATA path should not be empty")
	}
}

func TestSave_JSONRoundtrip(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "config.json")

	oldPath := configPath
	configPath = f
	defer func() { configPath = oldPath }()

	cfg := &Config{
		Language:     "python",
		ReadOnly:     true,
		EnabledTools: []string{"analyze_code", "scan_for_secrets"},
		DisableGit:   false,
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Language != "python" {
		t.Errorf("Language = %q, want python", loaded.Language)
	}
	if !loaded.ReadOnly {
		t.Error("ReadOnly should be true")
	}
	if len(loaded.EnabledTools) != 2 {
		t.Errorf("EnabledTools = %v, want 2 items", loaded.EnabledTools)
	}
}

func TestGet_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "config.json")

	oldPath := configPath
	configPath = f
	defer func() { configPath = oldPath }()

	cfg := &Config{Language: "go", RepoPath: "/test"}
	data, _ := json.Marshal(cfg)
	os.WriteFile(f, data, 0644)

	result, err := Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result == "" {
		t.Error("Get should return non-empty JSON")
	}
}

func TestSet_ParseToolListEmpty(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "config.json")

	oldPath := configPath
	configPath = f
	defer func() { configPath = oldPath }()

	_, err := Set("enabled_tools", "")
	if err != nil {
		t.Errorf("Set with empty tools failed: %v", err)
	}
	c, _ := Load()
	if len(c.EnabledTools) != 0 {
		t.Errorf("EnabledTools should be nil for empty value, got %v", c.EnabledTools)
	}
}
