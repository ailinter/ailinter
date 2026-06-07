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
	Project      *ProjectConfig `json:"project,omitempty"`
}

// ProjectConfig holds the project-level .ailinter.toml configuration.
type ProjectConfig struct {
	Path    string    `json:"path,omitempty"`
	Extends string    `json:"extends,omitempty"`
	Rules   RulesInfo `json:"rules,omitempty"`
}

// RulesInfo holds readable rule overrides from .ailinter.toml.
type RulesInfo struct {
	DeepNesting          *int     `json:"deep_nesting_warning,omitempty"`
	BrainMethod          *int     `json:"brain_method_warning_lines,omitempty"`
	FileBloat            *int     `json:"file_bloat_warning_lines,omitempty"`
	CyclomaticComplexity *int     `json:"cyclomatic_complexity_warning,omitempty"`
	BumpyRoad            *int     `json:"bumpy_road_bumps_warning,omitempty"`
	LongParameterList    *int     `json:"long_parameter_list_warning,omitempty"`
	ComplexConditional   *int     `json:"complex_conditional_branches_warning,omitempty"`
	LongSwitch           *int     `json:"long_switch_warning,omitempty"`
	ExcessiveComments    *float64 `json:"excessive_comments_ratio,omitempty"`
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
	if c.AccessToken != "" {
		c.AccessToken = "***"
	}
	cwd, err := os.Getwd()
	if err == nil {
		c.Project = LoadProjectConfigFile(cwd)
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data), nil
}

// Set updates a config value by key and persists.
func Set(key, value string) (string, error) {
	c, err := applyAndSave(key, value)
	if err != nil {
		return "", err
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	return string(data), nil
}

// GetConfig returns the current config struct with sensitive fields redacted.
func GetConfig() (*Config, error) {
	c, err := Load()
	if err != nil {
		return nil, err
	}
	if c.AccessToken != "" {
		c.AccessToken = "***"
	}
	cwd, err := os.Getwd()
	if err == nil {
		c.Project = LoadProjectConfigFile(cwd)
	}
	return c, nil
}

// SetAndGet sets a config value and returns the updated config struct with sensitive fields redacted.
func SetAndGet(key, value string) (*Config, error) {
	return applyAndSave(key, value)
}

func applyAndSave(key, value string) (*Config, error) {
	c, err := Load()
	if err != nil {
		return nil, err
	}

	switch key {
	case "access_token": // gitleaks:allow
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
		return nil, fmt.Errorf("unknown config key: %s (valid: access_token, onprem_url, default_path, language, repo_path, enabled_tools, read_only, disable_git)", key)
	}

	if err := Save(c); err != nil {
		return nil, err
	}

	if c.AccessToken != "" {
		c.AccessToken = "***"
	}
	cwd, err := os.Getwd()
	if err == nil {
		c.Project = LoadProjectConfigFile(cwd)
	}
	return c, nil
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
