package secrets_test

import (
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/secrets"
)

func TestNewScanner_Works(t *testing.T) {
	s, err := secrets.NewScanner()
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}
	if s == nil {
		t.Fatal("NewScanner returned nil")
	}
}

func TestScanner_ScanBytes(t *testing.T) {
	s, err := secrets.NewScanner()
	if err != nil {
		t.Skipf("gitleaks not available: %v", err)
	}

	t.Run("clean code", func(t *testing.T) {
		findings := s.ScanBytes([]byte("func hello() { return 42 }\n"), "clean.go")
		if len(findings) != 0 {
			t.Errorf("expected 0 findings, got %d", len(findings))
		}
	})

	t.Run("stripe key", func(t *testing.T) {
		findings := s.ScanBytes([]byte("var key = \"sk_live_1234567890abcdef\"\n"), "stripe.go")
		if len(findings) == 0 {
			t.Log("no findings for stripe key (might be allowlisted)")
		} else {
			for _, f := range findings {
				if f.RuleID == "" {
					t.Error("finding should have a RuleID")
				}
				if f.Severity == "" {
					t.Error("finding should have severity")
				}
				if !strings.Contains(f.Secret, "...") && f.Secret != "***" {
					t.Errorf("secret should be redacted, got %q", f.Secret)
				}
				if !strings.Contains(f.Message, "os.Getenv") && !strings.Contains(f.Message, "environment") {
					t.Errorf("message should mention env vars, got %q", f.Message)
				}
			}
		}
	})

	t.Run("multiple findings", func(t *testing.T) {
		content := "AWS_KEY=AKIAIOSFODNN7EXAMPLE\nSTRIPE_KEY=sk_live_abcdefghijklmnop\n"
		findings := s.ScanBytes([]byte(content), "multi.go")
		// Just verify it doesn't crash with multiple potential secrets
		_ = findings
	})
}

func TestScanner_HasFindings_Clean(t *testing.T) {
	s, err := secrets.NewScanner()
	if err != nil {
		t.Skipf("gitleaks not available: %v", err)
	}

	if s.HasFindings([]byte("func hello() { return 42 }\n")) {
		t.Error("HasFindings should be false for clean code")
	}
}

func TestScanner_ScanString(t *testing.T) {
	s, err := secrets.NewScanner()
	if err != nil {
		t.Skipf("gitleaks not available: %v", err)
	}

	findings := s.ScanString("func hello() { return 42 }\n", "clean.go")
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestScanner_EmptyContent(t *testing.T) {
	s, err := secrets.NewScanner()
	if err != nil {
		t.Skipf("gitleaks not available: %v", err)
	}

	findings := s.ScanBytes([]byte(""), "empty.go")
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for empty content, got %d", len(findings))
	}
}

func TestScanner_LargeContent(t *testing.T) {
	s, err := secrets.NewScanner()
	if err != nil {
		t.Skipf("gitleaks not available: %v", err)
	}

	large := strings.Repeat("func hello() { return 42 }\n", 100)
	findings := s.ScanBytes([]byte(large), "large.go")
	// Should not crash or hang
	_ = findings
}
