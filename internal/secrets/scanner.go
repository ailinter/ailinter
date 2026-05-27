package secrets

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/report"
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
	return convertFindings(findings, filePath)
}

// ScanString scans a string and returns findings.
func (s *Scanner) ScanString(content string, filePath string) []SecretFinding {
	findings := s.detector.DetectString(content)
	return convertFindings(findings, filePath)
}

// HasFindings returns true if there are any secrets detected.
func (s *Scanner) HasFindings(content []byte) bool {
	return len(s.detector.DetectBytes(content)) > 0
}

func convertFindings(findings []report.Finding, filePath string) []SecretFinding {
	result := make([]SecretFinding, 0, len(findings))
	for _, f := range findings {
		sev := classifySeverity(f)
		result = append(result, SecretFinding{
			RuleID:      f.RuleID,
			Description: f.Description,
			Line:        f.StartLine,
			Column:      f.StartColumn,
			Secret:      redactSecret(f.Secret),
			Entropy:     float64(f.Entropy),
			Severity:    sev,
			Message:     generateAIPrompt(f),
		})
	}
	return result
}

func classifySeverity(f report.Finding) string {
	if f.Entropy >= 4.5 {
		return "critical"
	}
	if f.Entropy >= 3.5 {
		return "alert"
	}
	return "warning"
}

func redactSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}

func generateAIPrompt(f report.Finding) string {
	return fmt.Sprintf(
		"CRITICAL: %s detected on line %d. %s. "+
			"You MUST rewrite this to use environment variables (e.g., os.Getenv(\"VAR_NAME\")). "+
			"Never hardcode secrets in source code.",
		f.RuleID, f.StartLine, f.Description,
	)
}
