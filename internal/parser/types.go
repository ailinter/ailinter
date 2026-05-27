package parser

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

// Tiers for vulnerability severity classification.
const (
	VulnLabelClean     = "Clean"
	VulnLabelMonitor   = "Monitor"
	VulnLabelRemediate = "Remediate"
)

func classify(score int) string {
	switch {
	case score >= 80:
		return LabelGoAhead
	case score >= 60:
		return LabelProceedWithCare
	case score >= 40:
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

// severityWeight maps severity strings to penalty weights.
func severityWeight(severity string) float64 {
	switch severity {
	case "warning":
		return 0.5
	case "alert":
		return 1.0
	case "critical":
		return 2.0
	default:
		return 0.5
	}
}

// TestClassifyHelper is exported for testing the classify function.
func TestClassifyHelper(score int) string {
	return classify(score)
}

// Classify returns the tier label for a given score.
func Classify(score int) string {
	return classify(score)
}
