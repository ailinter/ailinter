package config

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGetConfig_ReturnsStruct(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", oldHome)

	Set("language", "python")
	Set("repo_path", "/tmp/repo")

	cfg, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}
	if cfg.Language != "python" {
		t.Errorf("expected python, got %s", cfg.Language)
	}
	if cfg.RepoPath != "/tmp/repo" {
		t.Errorf("expected /tmp/repo, got %s", cfg.RepoPath)
	}
}

func TestGetConfig_RedactsAccessToken(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", oldHome)

	Set("access_token", "sk-live-secret-token-value-123")

	cfg, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AccessToken != "***" {
		t.Errorf("expected redacted token '***', got %q", cfg.AccessToken)
	}
}

func TestSetAndGet_Works(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", oldHome)

	cfg, err := SetAndGet("language", "go")
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("SetAndGet returned nil")
	}
	if cfg.Language != "go" {
		t.Errorf("expected go, got %s", cfg.Language)
	}
}

func TestSetAndGet_RedactsAccessToken(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", oldHome)

	cfg, err := SetAndGet("access_token", "my-secret-token-abc123")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AccessToken != "***" {
		t.Errorf("expected redacted token '***' in SetAndGet, got %q", cfg.AccessToken)
	}

	loaded, _ := Load()
	if loaded.AccessToken != "my-secret-token-abc123" {
		t.Error("token should be stored unredacted on disk")
	}
}

func TestGetConfig_Empty(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", oldHome)

	cfg, err := GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("GetConfig should return empty config, not nil")
	}
	if cfg.AccessToken != "" && cfg.AccessToken != "***" {
		t.Errorf("empty config access_token should be '' or '***', got %q", cfg.AccessToken)
	}
}

func TestSet_RedactsAccessTokenInReturn(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer os.Setenv("HOME", oldHome)

	result, err := Set("access_token", "secret-value-12345")
	if err != nil {
		t.Fatal(err)
	}

	var parsed Config
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.AccessToken != "***" {
		t.Errorf("Set return value should redact access_token, got %q", parsed.AccessToken)
	}
}
