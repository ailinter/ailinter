package secrets

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/zricethezav/gitleaks/v8/report"
)

func convertFindings(findings []report.Finding, filePath string) []SecretFinding {
	result := make([]SecretFinding, 0, len(findings))
	for _, f := range findings {
		if isFalsePositive(f) {
			continue
		}
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

// -- Severity classification --

func classifySeverity(f report.Finding) string {
	if f.Entropy >= 4.5 {
		return "critical"
	}
	if f.Entropy >= 3.5 {
		return "alert"
	}
	return "warning"
}

// -- Secret redaction --

func redactSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}

// -- AI prompt generation --

func generateAIPrompt(f report.Finding) string {
	return fmt.Sprintf(
		"CRITICAL: %s detected on line %d. %s. "+
			"You MUST rewrite this to use environment variables (e.g., os.Getenv(\"VAR_NAME\")). "+
			"Never hardcode secrets in source code.",
		f.RuleID, f.StartLine, f.Description,
	)
}

// -- Finding filtering --

func filterByGitleaksAllow(findings []report.Finding, content []byte) []report.Finding {
	if len(findings) == 0 {
		return findings
	}
	var filtered []report.Finding
	for _, f := range findings {
		if hasGitleaksAllow(content, f.StartLine) {
			continue
		}
		filtered = append(filtered, f)
	}
	return filtered
}

func hasGitleaksAllow(content []byte, findingLine int) bool {
	lines := bytes.Split(content, []byte("\n"))
	for _, lineIdx := range []int{findingLine - 1, findingLine - 2, findingLine} {
		if lineIdx < 0 || lineIdx >= len(lines) {
			continue
		}
		trimmed := bytes.TrimSpace(lines[lineIdx])
		if bytes.Contains(trimmed, []byte("gitleaks:allow")) {
			return true
		}
	}
	return false
}

// -- False positive detection --
//
// isFalsePositive catches strings that gitleaks may flag via entropy heuristics
// but are clearly not secrets. This supplements gitleaks' built-in allowlist/filter:
//   - Known placeholder/example values (hardcoded switch cases)
//   - Prefix-based allowlist for common non-secret patterns (e.g., test data)
//   - JWT header fragments without payload/signature

func isFalsePositive(f report.Finding) bool {
	secret := strings.ToLower(f.Secret)
	switch secret {
	case "your-api-key-here", "your-token-here", "your-secret-here",
		"default_value", "placeholder", "example",
		"super_secret_jwt_value", "hardcoded_password_123":
		return true
	}
	if strings.HasPrefix(secret, "this_is_just_a_") {
		return true
	}
	if isJWTHeaderOnly(f.Secret) {
		return true
	}
	return false
}

func isJWTHeaderOnly(secret string) bool {
	if !strings.HasPrefix(secret, "eyJ") || strings.Contains(secret, ".") {
		return false
	}
	return len(secret) <= 50
}
