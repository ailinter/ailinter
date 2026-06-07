package parser

import (
	"fmt"
	"strings"
)

// Smell represents a detected code smell.
type Smell struct {
	Name      string `json:"name"`     // e.g. "deep_nesting", "brain_method", "bumpy_road"
	Severity  string `json:"severity"` // "warning", "alert", "critical"
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
	Message   string `json:"message"`   // human-readable description
	AIPrompt  string `json:"ai_prompt"` // injection text for the LLM
}

// QualityResult is the full output of an analysis.
type QualityResult struct {
	Score       int     `json:"score"` // 0-100
	Label       string  `json:"label"` // one of LabelGoAhead, LabelProceedWithCare, LabelNeedsWork, LabelStopRefactor
	Smells      []Smell `json:"smells"`
	FilePath    string  `json:"file_path"`
	Language    string  `json:"language"`
	LinesOfCode int     `json:"lines_of_code"`
}

// DetectedLanguage maps a file extension to a language name.
func DetectedLanguage(ext string) string {
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".cpp", ".cc", ".cxx", ".c++":
		return "cpp"
	case ".c":
		return "c"
	case ".h", ".hpp":
		return "cpp"
	case ".java":
		return "java"
	case ".rs":
		return "rust"
	case ".rb":
		return "ruby"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".cs":
		return "csharp"
	case ".yaml", ".yml":
		return "yaml"
	case ".html", ".htm":
		return "html"
	default:
		return ""
	}
}

// Tiers for code quality score classification.
const (
	LabelGoAhead         = "Go Ahead"
	LabelProceedWithCare = "Proceed with Care"
	LabelNeedsWork       = "Needs Work"
	LabelStopRefactor    = "Stop & Refactor"
)

// Score tier thresholds — single source of truth for the classify() function.
// Templates reference these so they auto-update when tiers change.
const (
	GoAheadThreshold         = 80
	ProceedWithCareThreshold = 60
	NeedsWorkThreshold       = 40
)

// ScoreTier describes one score tier for documentation and templates.
type ScoreTier struct {
	MinScore int
	MaxScore int
	Label    string
	Guidance string
}

// ScoreTiers returns all quality score tiers in descending order.
// Templates call this to generate reference tables that always
// match the actual classify() function.
func ScoreTiers() []ScoreTier {
	return []ScoreTier{
		{
			MinScore: GoAheadThreshold,
			MaxScore: 100,
			Label:    LabelGoAhead,
			Guidance: "Safe for AI modification",
		},
		{
			MinScore: ProceedWithCareThreshold,
			MaxScore: GoAheadThreshold - 1,
			Label:    LabelProceedWithCare,
			Guidance: "Use guard clauses, prefer small changes, re-check after each edit",
		},
		{
			MinScore: NeedsWorkThreshold,
			MaxScore: ProceedWithCareThreshold - 1,
			Label:    LabelNeedsWork,
			Guidance: "Significant issues — refactor incrementally in small steps",
		},
		{
			MinScore: 0,
			MaxScore: NeedsWorkThreshold - 1,
			Label:    LabelStopRefactor,
			Guidance: "Refactor BEFORE AI modification. Run get_refactoring_strategy() for detected issues.",
		},
	}
}

// TierReferenceTable returns a markdown table of score tiers for
// inclusion in AGENTS.md and agent configuration templates.
func TierReferenceTable() string {
	var b strings.Builder
	b.WriteString("| Score | Label | AI Guidance |\n")
	b.WriteString("|-------|-------|-------------|\n")
	for _, t := range ScoreTiers() {
		rangeStr := fmt.Sprintf("%d-%d", t.MinScore, t.MaxScore)
		if t.MinScore == 0 {
			rangeStr = fmt.Sprintf("<%d", t.MaxScore+1)
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", rangeStr, t.Label, t.Guidance))
	}
	return b.String()
}

// Tiers for vulnerability severity classification.
const (
	VulnLabelClean     = "Clean"
	VulnLabelMonitor   = "Monitor"
	VulnLabelRemediate = "Remediate"
)

func classify(score int) string {
	switch {
	case score >= GoAheadThreshold:
		return LabelGoAhead
	case score >= ProceedWithCareThreshold:
		return LabelProceedWithCare
	case score >= NeedsWorkThreshold:
		return LabelNeedsWork
	default:
		return LabelStopRefactor
	}
}

// VulnClassify returns a vulnerability tier based on findings.
func VulnClassify(findings []struct {
	Severity string
}) string {
	hasAlert := false
	hasWarning := false
	for _, f := range findings {
		switch f.Severity {
		case "critical", "alert":
			hasAlert = true
		case "warning":
			hasWarning = true
		}
	}
	if hasAlert {
		return VulnLabelRemediate
	}
	if hasWarning {
		return VulnLabelMonitor
	}
	return VulnLabelClean

}

// TestClassifyHelper is exported for testing the classify function.
func TestClassifyHelper(score int) string {
	return classify(score)
}

// Classify returns the tier label for a given score.
func Classify(score int) string {
	return classify(score)
}
