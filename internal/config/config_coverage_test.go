package config_test

import (
	"os"
	"testing"

	"github.com/ailinter/ailinter/internal/config"
)

func TestLoad_FileNotFound(t *testing.T) {
	old := os.Getenv("HOME")
	os.Setenv("HOME", "/nonexistent-home-dir")
	defer os.Setenv("HOME", old)

	c, err := config.Load()
	if err != nil {
		t.Fatalf("Load should not error when file not found: %v", err)
	}
	if c == nil {
		t.Fatal("Load should return empty config, not nil")
	}
}

func TestSet_AllKeys(t *testing.T) {
	keys := []struct {
		key   string
		value string
	}{
		{"access_token", "test-token"},
		{"onprem_url", "https://example.com"},
		{"default_path", "/tmp"},
		{"language", "python"},
		{"repo_path", "/repo"},
		{"enabled_tools", "analyze_code,scan_for_secrets"},
		{"read_only", "true"},
		{"disable_git", "false"},
	}

	for _, tc := range keys {
		t.Run(tc.key, func(t *testing.T) {
			_, err := config.Set(tc.key, tc.value)
			if err != nil {
				t.Errorf("Set(%q, %q) failed: %v", tc.key, tc.value, err)
			}
		})
	}
}

func TestSet_EmptyValues(t *testing.T) {
	_, err := config.Set("enabled_tools", "")
	if err != nil {
		t.Errorf("Set with empty value failed: %v", err)
	}
	_, err = config.Set("enabled_tools", "*")
	if err != nil {
		t.Errorf("Set with * failed: %v", err)
	}
}

func TestSet_BooleanVariants(t *testing.T) {
	_, err := config.Set("read_only", "false")
	if err != nil {
		t.Errorf("Set read_only=false failed: %v", err)
	}
	_, err = config.Set("read_only", "0")
	if err != nil {
		t.Errorf("Set read_only=0 failed: %v", err)
	}
	_, err = config.Set("disable_git", "1")
	if err != nil {
		t.Errorf("Set disable_git=1 failed: %v", err)
	}
}

func TestGet_Works(t *testing.T) {
	s, err := config.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if s == "" {
		t.Error("Get should return non-empty string")
	}
}

func TestPath(t *testing.T) {
	p := config.Path()
	if p == "" {
		t.Error("Path should not be empty")
	}
}

func TestParseToolList(t *testing.T) {
	// Covered via Set tests, but verifies the function accepts various formats
	_, err := config.Set("enabled_tools", "  tool1 , tool2 , tool3  ")
	if err != nil {
		t.Errorf("Set with spaced tools failed: %v", err)
	}
}
