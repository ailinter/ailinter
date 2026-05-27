package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents ailinter's persistent configuration.
type Config struct {
	AccessToken  string         `json:"access_token,omitempty"`
	OnPremURL    string         `json:"onprem_url,omitempty"`
	DefaultPath  string         `json:"default_path,omitempty"`
	EnabledTools []string       `json:"enabled_tools,omitempty"`
	ReadOnly     bool           `json:"read_only,omitempty"`
	Language     string         `json:"language,omitempty"`
	RepoPath     string         `json:"repo_path,omitempty"`
	Thresholds   map[string]int `json:"thresholds,omitempty"`
	DisableGit   bool           `json:"disable_git,omitempty"`
}

var configPath string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	switch {
	case os.Getenv("APPDATA") != "":
		configPath = filepath.Join(os.Getenv("APPDATA"), "ailinter", "config.json")
	case home != "":
		configPath = filepath.Join(home, ".config", "ailinter", "config.json")
	default:
		configPath = "ailinter-config.json"
	}
}

// Path returns the config file path.
func Path() string {
	return configPath
}

// Load reads the configuration from disk.
func Load() (*Config, error) {
	c := &Config{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil // return defaults
		}
		return nil, err
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return c, nil
}

// Save writes the configuration to disk.
func Save(c *Config) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// Get returns a user-friendly representation of the current config.
func Get() (string, error) {
	c, err := Load()
	if err != nil {
		return "", err
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data), nil
}

// Set updates a config value by key and persists.
func Set(key, value string) (string, error) {
	c, err := Load()
	if err != nil {
		return "", err
	}

	switch key {
	case "access_token":
		c.AccessToken = value
	case "onprem_url":
		c.OnPremURL = value
	case "default_path":
		c.DefaultPath = value
	case "language":
		c.Language = value
	case "repo_path":
		c.RepoPath = value
	case "enabled_tools":
		c.EnabledTools = parseToolList(value)
	case "read_only":
		c.ReadOnly = value == "true" || value == "1"
	case "disable_git":
		c.DisableGit = value == "true" || value == "1"
	default:
		return "", fmt.Errorf("unknown config key: %s (valid: access_token, onprem_url, default_path, language, repo_path, enabled_tools, read_only, disable_git)", key)
	}

	if err := Save(c); err != nil {
		return "", err
	}

	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data), nil
}

func parseToolList(value string) []string {
	if value == "" || value == "*" {
		return nil
	}
	var result []string
	for _, t := range splitAndTrim(value) {
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

func splitAndTrim(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
