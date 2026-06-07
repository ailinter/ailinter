package config_test

import (
	"os"
	"testing"

	"github.com/ailinter/ailinter/internal/config"
)

func TestLoad_Empty(t *testing.T) {
	// Override config path for test
	os.Setenv("HOME", t.TempDir())
	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("Load returned nil config")
	}
}

func TestLoad_NotFound(t *testing.T) {
	os.Setenv("HOME", t.TempDir())
	c, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	// Default config should have zero values
	if c == nil {
		t.Fatal("Load returned nil")
	}
}

func TestSetGet_Roundtrip(t *testing.T) {
	os.Setenv("HOME", t.TempDir())
	_, err := config.Set("language", "go")
	if err != nil {
		t.Fatal(err)
	}
	got, err := config.Get()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("Get returned empty string")
	}
}

func TestSet_ReadOnly(t *testing.T) {
	os.Setenv("HOME", t.TempDir())
	_, err := config.Set("read_only", "true")
	if err != nil {
		t.Fatal(err)
	}
	c, _ := config.Load()
	if !c.ReadOnly {
		t.Error("read_only should be true after setting")
	}
}

func TestSet_UnknownKey(t *testing.T) {
	os.Setenv("HOME", t.TempDir())
	_, err := config.Set("nonexistent", "value")
	if err == nil {
		t.Error("expected error for unknown key")
	}
}

func TestSet_EnabledTools(t *testing.T) {
	os.Setenv("HOME", t.TempDir())
	_, err := config.Set("enabled_tools", "analyze_code_health,scan_for_secrets")
	if err != nil {
		t.Fatal(err)
	}
	c, _ := config.Load()
	if len(c.EnabledTools) != 2 {
		t.Errorf("expected 2 enabled tools, got %d", len(c.EnabledTools))
	}
}

func TestSet_DisableGit(t *testing.T) {
	os.Setenv("HOME", t.TempDir())
	config.Set("disable_git", "true")
	c, _ := config.Load()
	if !c.DisableGit {
		t.Error("disable_git should be true")
	}
}

func TestSet_RepoPath(t *testing.T) {
	os.Setenv("HOME", t.TempDir())
	config.Set("repo_path", "/tmp/test")
	c, _ := config.Load()
	if c.RepoPath != "/tmp/test" {
		t.Errorf("expected /tmp/test, got %s", c.RepoPath)
	}
}

func TestSet_OnPremURL(t *testing.T) {
	os.Setenv("HOME", t.TempDir())
	config.Set("onprem_url", "https://example.com")
	c, _ := config.Load()
	if c.OnPremURL != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", c.OnPremURL)
	}
}

func TestSet_AccessToken(t *testing.T) {
	os.Setenv("HOME", t.TempDir())
	config.Set("access_token", "test-token-123") // gitleaks:allow
	c, _ := config.Load()
	if c.AccessToken != "test-token-123" {
		t.Error("access_token not set")
	}
}

func TestConfig_Path(t *testing.T) {
	p := config.Path()
	if p == "" {
		t.Error("Config.Path() returned empty")
	}
	t.Logf("Config path: %s", p)
}
