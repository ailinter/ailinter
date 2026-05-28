package secrets

import (
	"strings"
	"testing"

	"github.com/zricethezav/gitleaks/v8/report"
)

func TestConvertFindings(t *testing.T) {
	findings := []report.Finding{
		{
			RuleID:      "aws-access-key",
			Description: "AWS Access Key",
			StartLine:   10,
			StartColumn: 5,
			Secret:      "AKIAIOSFODNN7EXAMPLE",
			Entropy:     4.5,
		},
		{
			RuleID:      "stripe-key",
			Description: "Stripe API Key",
			StartLine:   20,
			StartColumn: 15,
			Secret:      "sk_live_4eC39HqLyjWDarjtT1zdp7dc",
			Entropy:     3.0,
		},
		{
			RuleID:      "short-secret",
			Description: "Short secret",
			StartLine:   30,
			StartColumn: 1,
			Secret:      "abc",
			Entropy:     2.0,
		},
	}

	result := convertFindings(findings, "test.env")

	if len(result) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(result))
	}

	t.Run("aws key", func(t *testing.T) {
		f := result[0]
		if f.RuleID != "aws-access-key" {
			t.Errorf("RuleID = %q, want %q", f.RuleID, "aws-access-key")
		}
		if f.Description != "AWS Access Key" {
			t.Errorf("Description = %q, want %q", f.Description, "AWS Access Key")
		}
		if f.Line != 10 {
			t.Errorf("Line = %d, want 10", f.Line)
		}
		if f.Column != 5 {
			t.Errorf("Column = %d, want 5", f.Column)
		}
		if f.Severity != "critical" {
			t.Errorf("Severity = %q, want %q", f.Severity, "critical")
		}
		if !strings.Contains(f.Message, "aws-access-key") {
			t.Errorf("Message should contain rule ID, got %q", f.Message)
		}
		if !strings.Contains(f.Secret, "...") {
			t.Errorf("Secret should be redacted, got %q", f.Secret)
		}
	})

	t.Run("stripe key", func(t *testing.T) {
		f := result[1]
		if f.RuleID != "stripe-key" {
			t.Errorf("RuleID = %q, want %q", f.RuleID, "stripe-key")
		}
		if f.Severity != "warning" {
			t.Errorf("Severity = %q, want %q", f.Severity, "warning")
		}
		if f.Secret != "sk_l...p7dc" {
			t.Errorf("Secret = %q, want %q", f.Secret, "sk_l...p7dc")
		}
	})

	t.Run("short secret", func(t *testing.T) {
		f := result[2]
		if f.Secret != "***" {
			t.Errorf("Secret = %q, want %q", f.Secret, "***")
		}
	})
}

func TestConvertFindings_FalsePositiveFilter(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{"api placeholder", "your-api-key-here"},
		{"token placeholder", "your-token-here"},
		{"secret placeholder", "your-secret-here"},
		{"default value", "default_value"},
		{"generic placeholder", "placeholder"},
		{"example value", "example"},
		{"jwt test value", "super_secret_jwt_value"},
		{"hardcoded password", "hardcoded_password_123"},
		{"prefix filter", "this_is_just_a_test_value"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			findings := []report.Finding{
				{RuleID: "test-rule", Description: "test", Secret: tc.secret, Entropy: 4.0},
			}
			result := convertFindings(findings, "test.txt")
			if len(result) != 0 {
				t.Errorf("expected 0 findings for %q, got %d", tc.secret, len(result))
			}
		})
	}
}

func TestFilterByGitleaksAllow(t *testing.T) {
	t.Run("no allow comment", func(t *testing.T) {
		content := []byte("var x = \"hello\"\nvar y = \"world\"\n")
		findings := []report.Finding{
			{RuleID: "test", StartLine: 1},
			{RuleID: "test2", StartLine: 2},
		}
		result := filterByGitleaksAllow(findings, content)
		if len(result) != 2 {
			t.Errorf("expected 2 findings, got %d", len(result))
		}
	})

	t.Run("allow on same line", func(t *testing.T) {
		content := []byte("var key = \"secret\" // gitleaks:allow\n")
		findings := []report.Finding{
			{RuleID: "test", StartLine: 1},
		}
		result := filterByGitleaksAllow(findings, content)
		if len(result) != 0 {
			t.Errorf("expected 0 findings, got %d", len(result))
		}
	})

	t.Run("allow on line above", func(t *testing.T) {
		content := []byte("// gitleaks:allow\nvar key = \"secret\"\n")
		findings := []report.Finding{
			{RuleID: "test", StartLine: 2},
		}
		result := filterByGitleaksAllow(findings, content)
		if len(result) != 0 {
			t.Errorf("expected 0 findings, got %d", len(result))
		}
	})

	t.Run("allow on line below", func(t *testing.T) {
		content := []byte("var key = \"secret\"\n// gitleaks:allow\n")
		findings := []report.Finding{
			{RuleID: "test", StartLine: 1},
		}
		result := filterByGitleaksAllow(findings, content)
		if len(result) != 0 {
			t.Errorf("expected 0 findings, got %d", len(result))
		}
	})

	t.Run("empty findings", func(t *testing.T) {
		content := []byte("anything here\n")
		result := filterByGitleaksAllow(nil, content)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("mixed allow and no allow", func(t *testing.T) {
		content := []byte("safe line\n// gitleaks:allow\nvar key = \"secret\"\nvar other = \"data\"\n")
		findings := []report.Finding{
			{RuleID: "a", StartLine: 3},
			{RuleID: "b", StartLine: 4},
		}
		result := filterByGitleaksAllow(findings, content)
		if len(result) != 1 {
			t.Errorf("expected 1 finding, got %d", len(result))
		}
		if result[0].StartLine != 4 {
			t.Errorf("expected remaining finding on line 4, got line %d", result[0].StartLine)
		}
	})
}

func TestHasGitleaksAllow(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		findingLine int
		want        bool
	}{
		{
			name:        "allow on same line",
			content:     "var key = \"x\" // gitleaks:allow\n",
			findingLine: 1,
			want:        true,
		},
		{
			name:        "allow one line above",
			content:     "// gitleaks:allow\nvar key = \"x\"\n",
			findingLine: 2,
			want:        true,
		},
		{
			name:        "allow one line below",
			content:     "var key = \"x\"\n// gitleaks:allow\n",
			findingLine: 1,
			want:        true,
		},
		{
			name:        "no allow comment",
			content:     "var key = \"x\"\nvar clean = \"y\"\n",
			findingLine: 1,
			want:        false,
		},
		{
			name:        "empty content",
			content:     "",
			findingLine: 1,
			want:        false,
		},
		{
			name:        "finding on negative line",
			content:     "some content\n",
			findingLine: 0,
			want:        false,
		},
		{
			name:        "finding past end of content",
			content:     "line one\n",
			findingLine: 100,
			want:        false,
		},
		{
			name:        "allow with leading whitespace",
			content:     "   // gitleaks:allow\nvar key = \"x\"\n",
			findingLine: 2,
			want:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasGitleaksAllow([]byte(tc.content), tc.findingLine); got != tc.want {
				t.Errorf("hasGitleaksAllow(%q, %d) = %v, want %v", tc.content, tc.findingLine, got, tc.want)
			}
		})
	}
}

func TestGenerateAIPrompt(t *testing.T) {
	f := report.Finding{
		RuleID:      "aws-access-key",
		Description: "AWS Access Key ID detected",
		StartLine:   42,
	}
	msg := generateAIPrompt(f)

	if !strings.Contains(msg, "aws-access-key") {
		t.Errorf("prompt should contain rule ID, got: %s", msg)
	}
	if !strings.Contains(msg, "os.Getenv") {
		t.Errorf("prompt should mention os.Getenv, got: %s", msg)
	}
	if !strings.Contains(msg, "AWS Access Key") {
		t.Errorf("prompt should contain description, got: %s", msg)
	}
	if !strings.Contains(msg, "line 42") {
		t.Errorf("prompt should mention line number, got: %s", msg)
	}
}

func TestNewScannerConfig(t *testing.T) {
	t.Run("valid minimal config", func(t *testing.T) {
		configTOML := `
[[rules]]
id = "test-rule"
description = "Test rule"
regex = '''test_secret_[A-Za-z0-9]+'''
`
		s, err := NewScannerConfig(configTOML)
		if err != nil {
			t.Fatalf("NewScannerConfig failed: %v", err)
		}
		if s == nil {
			t.Fatal("NewScannerConfig returned nil")
		}
	})

	t.Run("invalid toml", func(t *testing.T) {
		_, err := NewScannerConfig("not valid toml {{{")
		if err == nil {
			t.Error("expected error for invalid TOML")
		}
	})
}

func TestConvertFindings_EmptyInput(t *testing.T) {
	result := convertFindings(nil, "test.txt")
	if len(result) != 0 {
		t.Errorf("expected 0 findings for nil input, got %d", len(result))
	}

	result = convertFindings([]report.Finding{}, "test.txt")
	if len(result) != 0 {
		t.Errorf("expected 0 findings for empty input, got %d", len(result))
	}
}
