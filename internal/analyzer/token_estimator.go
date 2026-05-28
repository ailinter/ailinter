package analyzer

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// TokenEstimator estimates token counts and cost savings for AI interactions.
type TokenEstimator struct {
	FilePath     string
	CurrentScore int
	FileBytes    int64
}

// ModelPricing holds pricing info for an AI model.
type ModelPricing struct {
	Name  string
	CostPerMTokens float64 // USD per million input tokens
}

var supportedModels = []ModelPricing{
	{Name: "Claude Opus 4", CostPerMTokens: 15.00},
	{Name: "GPT-4.5", CostPerMTokens: 10.00},
	{Name: "Claude 3.5 Sonnet", CostPerMTokens: 3.00},
}

// NewTokenEstimator creates an estimator for a given file.
func NewTokenEstimator(filePath string, currentScore int) *TokenEstimator {
	t := &TokenEstimator{
		FilePath:     filePath,
		CurrentScore: currentScore,
	}
	if info, err := os.Stat(filePath); err == nil {
		t.FileBytes = info.Size()
	}
	return t
}

// CurrentTokens estimates the number of tokens in the current file.
// Approx 1 token per 4 characters (standard LLM approximation).
func (t *TokenEstimator) CurrentTokens() int {
	return int(math.Ceil(float64(t.FileBytes) / 4.0))
}

// ReductionFactor returns the estimated token reduction factor after refactoring to 80+.
// Based on current quality score:
//   - 0-36: 50% reduction
//   - 37-49: 50% reduction
//   - 50-64: 45% reduction
//   - 65-79: 40% reduction
//   - 80+: 10% reduction (already clean)
func ReductionFactor(score int) float64 {
	switch {
	case score < 50:
		return 0.50
	case score < 65:
		return 0.45
	case score < 80:
		return 0.40
	default:
		return 0.10
	}
}

// EstimatedTokensAfterRefactor estimates tokens after refactoring to score 80+.
func (t *TokenEstimator) EstimatedTokensAfterRefactor() int {
	current := float64(t.CurrentTokens())
	return int(math.Round(current * (1.0 - ReductionFactor(t.CurrentScore))))
}

// TokenSavingsPerRead returns the number of tokens saved per single read of the file.
func (t *TokenEstimator) TokenSavingsPerRead() int {
	return t.CurrentTokens() - t.EstimatedTokensAfterRefactor()
}

// InteractionCost returns the cost of reading the file once with a given model.
func (t *TokenEstimator) InteractionCost(model ModelPricing) float64 {
	tokens := float64(t.CurrentTokens())
	return tokens / 1_000_000.0 * model.CostPerMTokens
}

// SavingsPerRead returns per-interaction savings for a given model.
func (t *TokenEstimator) SavingsPerRead(model ModelPricing) float64 {
	savedTokens := float64(t.TokenSavingsPerRead())
	return math.Round(savedTokens/1_000_000.0*model.CostPerMTokens*1000) / 1000
}

// MonthlyEnterpriseSavings calculates monthly savings for 20 devs, 50 AI calls/day, 6 files read per call.
func (t *TokenEstimator) MonthlyEnterpriseSavings(model ModelPricing) float64 {
	perInteraction := t.SavingsPerRead(model)
	// 20 devs * 50 AI calls/day * 22 working days * 6 files read per call
	savings := perInteraction * 20.0 * 50.0 * 22.0 * 6.0
	return math.Round(savings*100) / 100
}

// IterationReductionSavings calculates total savings including ~40% fewer iterations from clean code.
func (t *TokenEstimator) IterationReductionSavings(model ModelPricing) float64 {
	monthly := t.MonthlyEnterpriseSavings(model)
	// Hidden savings: clean code needs ~40% fewer iterations
	total := monthly * 1.4
	return math.Round(total*100) / 100
}

// FormatEstimateOutput returns the formatted token estimate string.
func (t *TokenEstimator) FormatEstimateOutput() string {
	fileName := filepath.Base(t.FilePath)
	currentTokens := t.CurrentTokens()
	afterTokens := t.EstimatedTokensAfterRefactor()
	savings := t.TokenSavingsPerRead()
	pct := int(ReductionFactor(t.CurrentScore) * 100)

	var out string
	out += "Token Savings Estimate\n"
	out += "──────────────────────\n"
	out += fmt.Sprintf("File: %s (%d tokens, score: %d)\n", fileName, currentTokens, t.CurrentScore)
	out += fmt.Sprintf("After refactoring: ~%d tokens (%d%% reduction at score 80+)\n\n", afterTokens, pct)

	out += "Per AI interaction (reads file once):\n"
	for _, model := range supportedModels {
		perRead := t.SavingsPerRead(model)
		out += fmt.Sprintf("  %-20s $%.3f saved per read\n", model.Name+":", perRead)
	}

	out += "\nMonthly (20 devs, 50 AI calls/day, 6 files read per call):\n"
	for _, model := range supportedModels {
		monthly := t.MonthlyEnterpriseSavings(model)
		out += fmt.Sprintf("  %-20s $%.0f/month saved\n", model.Name+":", monthly)
	}

	out += "\nHidden savings: Clean code needs ~40% fewer iterations.\n"
	for _, model := range supportedModels {
		total := t.IterationReductionSavings(model)
		out += fmt.Sprintf("Total estimated savings: $%.0f/month (%s, including iteration reduction)\n", total, model.Name)
	}

	return out
}

// CostForModel returns the per-interaction cost for a specific model.
func (t *TokenEstimator) CostForModel(modelName string) float64 {
	for _, m := range supportedModels {
		if m.Name == modelName {
			return t.InteractionCost(m)
		}
	}
	return 0
}
