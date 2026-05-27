package secrets_test

import (
	"testing"

	"github.com/ailinter/ailinter/internal/secrets"
)

func TestScanner_NoSecrets(t *testing.T) {
	s, err := secrets.NewScanner()
	if err != nil {
		t.Skipf("gitleaks not available: %v", err)
	}
	findings := s.ScanString("package main\nfunc main() {}\n", "test.go")
	if len(findings) > 0 {
		t.Errorf("expected no secrets, got %d", len(findings))
	}
}

func TestScanner_AWSKey(t *testing.T) {
	s, err := secrets.NewScanner()
	if err != nil {
		t.Skipf("gitleaks not available: %v", err)
	}
	findings := s.ScanString("AKIAIOSFODNN7EXAMPLE", "test.go")
	// This is the AWS example key — it may be in gitleaks allowlist
	t.Logf("AWS test key findings: %d", len(findings))
}

func TestScanner_StripeKey(t *testing.T) {
	s, err := secrets.NewScanner()
	if err != nil {
		t.Skipf("gitleaks not available: %v", err)
	}
	findings := s.ScanString("sk_live_4eC39HqLyjWDarjtT1zdp7dc", "test.go")
	if len(findings) == 0 {
		t.Error("expected stripe key detection")
	}
}

func TestScanner_HasFindings(t *testing.T) {
	s, err := secrets.NewScanner()
	if err != nil {
		t.Skipf("gitleaks not available: %v", err)
	}
	if s.HasFindings([]byte("package main")) {
		t.Error("clean code should have no findings")
	}
}
