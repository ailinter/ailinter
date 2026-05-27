package secrets

import (
	"testing"

	"github.com/zricethezav/gitleaks/v8/report"
)

func TestIsFalsePositive(t *testing.T) {
	tests := []struct {
		secret string
		want   bool
	}{
		{"your-api-key-here", true},
		{"YOUR-API-KEY-HERE", true},
		{"your-token-here", true},
		{"your-secret-here", true},
		{"default_value", true},
		{"placeholder", true},
		{"example", true},
		{"super_secret_jwt_value", true},
		{"hardcoded_password_123", true},
		{"sk_live_4eC39HqLyjWDarjtT1zdp7dc", false},
		{"AKIAIOSFODNN7EXAMPLE", false},
		{"real-secret-value-123", false},
	}
	for _, tc := range tests {
		f := report.Finding{Secret: tc.secret}
		if got := isFalsePositive(f); got != tc.want {
			t.Errorf("isFalsePositive(%q) = %v, want %v", tc.secret, got, tc.want)
		}
	}
}

func TestIsJWTHeaderOnly(t *testing.T) {
	tests := []struct {
		secret string
		want   bool
	}{
		{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", true}, // JWT header only (~36 chars, <=50)
		{"eyJhIn0", true}, // short 3-char header with payload
		{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature", false}, // full JWT (has dots)
		{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0", false}, // no signature but has dot
		{"not-a-jwt", false},
		{"", false},
		{"sk_live_12345", false},
	}
	for _, tc := range tests {
		if got := isJWTHeaderOnly(tc.secret); got != tc.want {
			t.Errorf("isJWTHeaderOnly(%q) = %v, want %v", tc.secret, got, tc.want)
		}
	}
}

func TestRedactSecret(t *testing.T) {
	tests := []struct {
		secret string
		want   string
	}{
		{"abc", "***"},
		{"12345678", "***"},
		{"sk_live_4eC39HqLyjWDarjtT1zdp7dc", "sk_l...p7dc"},
		{"ghp_1234567890abcdef1234567890abcdef", "ghp_...cdef"},
	}
	for _, tc := range tests {
		if got := redactSecret(tc.secret); got != tc.want {
			t.Errorf("redactSecret(%q) = %q, want %q", tc.secret, got, tc.want)
		}
	}
}

func TestClassifySeverity(t *testing.T) {
	tests := []struct {
		entropy float32
		want    string
	}{
		{5.0, "critical"},
		{4.5, "critical"},
		{4.0, "alert"},
		{3.5, "alert"},
		{3.0, "warning"},
		{1.0, "warning"},
	}
	for _, tc := range tests {
		f := report.Finding{Entropy: tc.entropy}
		if got := classifySeverity(f); got != tc.want {
			t.Errorf("classifySeverity(entropy=%v) = %q, want %q", tc.entropy, got, tc.want)
		}
	}
}
