package cli

import (
	"fmt"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/secrets"
)

func problemSeverity(severity string) string {
	switch severity {
	case "critical":
		return "error"
	case "alert":
		return "warning"
	default:
		return "warning"
	}
}

func writeProblemsResult(result analyzer.QualityResult) {
	for _, s := range result.Smells {
		fmt.Printf("%s:%d:%d: %s: %s: %s\n",
			result.FilePath, s.LineStart, 1,
			problemSeverity(s.Severity), s.Name, s.Message)
	}
}

func writeProblemsSecrets(path string, findings []secrets.SecretFinding) {
	for _, f := range findings {
		fmt.Printf("%s:%d:%d: %s: %s: %s\n",
			path, f.Line, f.Column,
			problemSeverity(f.Severity), f.RuleID, f.Description)
	}
}
