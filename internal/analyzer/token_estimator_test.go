package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTokenEstimator_NewTokenEstimator(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.go")
	if err := os.WriteFile(path, []byte("package test\n\n// Hello world\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	te := NewTokenEstimator(path, 42)
	if te.FilePath != path {
		t.Errorf("FilePath = %s, want %s", te.FilePath, path)
	}
	if te.CurrentScore != 42 {
		t.Errorf("CurrentScore = %d, want 42", te.CurrentScore)
	}
	if te.FileBytes == 0 {
		t.Error("FileBytes should be > 0 for existing file")
	}
}

func TestTokenEstimator_NewTokenEstimator_Nonexistent(t *testing.T) {
	te := NewTokenEstimator("/nonexistent/path/file.go", 80)
	if te.FilePath != "/nonexistent/path/file.go" {
		t.Errorf("FilePath = %s", te.FilePath)
	}
	if te.FileBytes != 0 {
		t.Errorf("FileBytes = %d, want 0 for nonexistent file", te.FileBytes)
	}
}

func TestTokenEstimator_CurrentTokens(t *testing.T) {
	te := &TokenEstimator{FileBytes: 400}
	if got := te.CurrentTokens(); got != 100 {
		t.Errorf("CurrentTokens() = %d, want 100 (400 bytes / 4)", got)
	}
}

func TestReductionFactor(t *testing.T) {
	tests := []struct {
		score int
		want  float64
	}{
		{0, 0.50},
		{36, 0.50},
		{49, 0.50},
		{50, 0.45},
		{64, 0.45},
		{65, 0.40},
		{79, 0.40},
		{80, 0.10},
		{100, 0.10},
	}
	for _, tt := range tests {
		if got := ReductionFactor(tt.score); got != tt.want {
			t.Errorf("ReductionFactor(%d) = %f, want %f", tt.score, got, tt.want)
		}
	}
}

func TestTokenEstimator_EstimatedTokensAfterRefactor(t *testing.T) {
	// File with 400 bytes = 100 tokens, score 30 => 50% reduction => 50 tokens
	te := &TokenEstimator{FileBytes: 400, CurrentScore: 30}
	got := te.EstimatedTokensAfterRefactor()
	if got != 50 {
		t.Errorf("EstimatedTokensAfterRefactor() = %d, want 50", got)
	}
}

func TestTokenEstimator_TokenSavingsPerRead(t *testing.T) {
	te := &TokenEstimator{FileBytes: 400, CurrentScore: 30}
	// CurrentTokens = 100, AfterRefactor = 50, savings = 50
	if got := te.TokenSavingsPerRead(); got != 50 {
		t.Errorf("TokenSavingsPerRead() = %d, want 50", got)
	}
}

func TestTokenEstimator_InteractionCost(t *testing.T) {
	te := &TokenEstimator{FileBytes: 400, CurrentScore: 30}
	// 100 tokens at $15/MToken = 100/1e6 * 15 = $0.0015
	model := ModelPricing{Name: "test", CostPerMTokens: 15.0}
	got := te.InteractionCost(model)
	if got < 0.001 || got > 0.002 {
		t.Errorf("InteractionCost() = %f, want ~0.0015", got)
	}
}

func TestTokenEstimator_SavingsPerRead(t *testing.T) {
	te := &TokenEstimator{FileBytes: 400, CurrentScore: 30}
	model := ModelPricing{Name: "test", CostPerMTokens: 15.0}
	got := te.SavingsPerRead(model)
	// 50 saved tokens at $15/MToken = 50/1e6 * 15 = $0.00075, rounded to $0.001
	if got < 0.0 {
		t.Errorf("SavingsPerRead() = %f, expected non-negative", got)
	}
}

func TestTokenEstimator_MonthlyEnterpriseSavings(t *testing.T) {
	te := &TokenEstimator{FileBytes: 400, CurrentScore: 30}
	model := ModelPricing{Name: "test", CostPerMTokens: 15.0}
	got := te.MonthlyEnterpriseSavings(model)
	// Should be positive and reasonable
	if got < 0 {
		t.Errorf("MonthlyEnterpriseSavings() = %f, expected positive", got)
	}
}

func TestTokenEstimator_IterationReductionSavings(t *testing.T) {
	te := &TokenEstimator{FileBytes: 400, CurrentScore: 30}
	model := ModelPricing{Name: "test", CostPerMTokens: 15.0}
	got := te.IterationReductionSavings(model)
	monthly := te.MonthlyEnterpriseSavings(model)
	if got < monthly {
		t.Errorf("IterationReductionSavings() = %f, should be > monthly savings %f", got, monthly)
	}
}

func TestTokenEstimator_FormatEstimateOutput(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.go")
	if err := os.WriteFile(path, []byte("package test\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	te := NewTokenEstimator(path, 42)
	out := te.FormatEstimateOutput()
	if !strings.Contains(out, "Token Savings Estimate") {
		t.Error("FormatEstimateOutput() should contain header")
	}
	if !strings.Contains(out, "test.go") {
		t.Error("FormatEstimateOutput() should contain filename")
	}
}

func TestTokenEstimator_CostForModel(t *testing.T) {
	te := &TokenEstimator{FileBytes: 400, CurrentScore: 30}
	if got := te.CostForModel("Claude Opus 4"); got <= 0 {
		t.Errorf("CostForModel(Claude Opus 4) = %f, expected positive", got)
	}
	if got := te.CostForModel("NonExistent"); got != 0 {
		t.Errorf("CostForModel(NonExistent) = %f, want 0", got)
	}
}
