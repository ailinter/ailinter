package secrets

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/detect"
)

//go:embed betterleaks.toml
var betterleaksConfig string

// SecretFinding represents a detected secret with AI guidance.
type SecretFinding struct {
	RuleID      string  `json:"rule_id"`
	Description string  `json:"description"`
	Line        int     `json:"line"`
	Column      int     `json:"column"`
	Secret      string  `json:"secret"`
	Entropy     float64 `json:"entropy"`
	Severity    string  `json:"severity"`
	Message     string  `json:"message"`
}

// Scanner wraps the gitleaks detector.
type Scanner struct {
	detector *detect.Detector
}

// NewScanner creates a scanner with the best available rule set.
// Tries betterleaks 269-rule config first, falls back to gitleaks 150-rule default.
func NewScanner() (*Scanner, error) {
	d, err := newDetectorFromConfig(betterleaksConfig)
	if err != nil {
		d2, err2 := detect.NewDetectorDefaultConfig()
		if err2 != nil {
			return nil, fmt.Errorf("failed to initialize gitleaks detector (betterleaks: %v, default: %v)", err, err2)
		}
		return &Scanner{detector: d2}, nil
	}
	return &Scanner{detector: d}, nil
}

// NewScannerConfig creates a scanner from a custom TOML config string.
func NewScannerConfig(configTOML string) (*Scanner, error) {
	d, err := newDetectorFromConfig(configTOML)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize detector from config: %w", err)
	}
	return &Scanner{detector: d}, nil
}

func newDetectorFromConfig(tomlContent string) (*detect.Detector, error) {
	v := viper.New()
	v.SetConfigType("toml")
	if err := v.ReadConfig(strings.NewReader(tomlContent)); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var vc config.ViperConfig
	if err := v.Unmarshal(&vc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	cfg, err := vc.Translate()
	if err != nil {
		return nil, fmt.Errorf("failed to translate config: %w", err)
	}
	return detect.NewDetector(cfg), nil
}

// ScanBytes scans raw content and returns findings.
func (s *Scanner) ScanBytes(content []byte, filePath string) []SecretFinding {
	findings := s.detector.DetectBytes(content)
	filtered := filterByGitleaksAllow(findings, content)
	return convertFindings(filtered, filePath)
}

// ScanString scans a string and returns findings.
func (s *Scanner) ScanString(content string, filePath string) []SecretFinding {
	findings := s.detector.DetectString(content)
	filtered := filterByGitleaksAllow(findings, []byte(content))
	return convertFindings(filtered, filePath)
}

// HasFindings returns true if there are any secrets detected.
func (s *Scanner) HasFindings(content []byte) bool {
	findings := s.detector.DetectBytes(content)
	filtered := filterByGitleaksAllow(findings, content)
	return len(filtered) > 0
}
