package cli

import (
	"fmt"
	"strings"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/parser"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/vulnerability"
)

func writeMarkdownResult(result analyzer.QualityResult) {
	fmt.Printf("## %s\n\n", result.FilePath)
	fmt.Printf("- **Language:** %s | **Lines:** %d | **Score:** %d/100 — %s\n\n",
		result.Language, result.LinesOfCode, result.Score, result.Label)

	if len(result.Smells) == 0 {
		fmt.Print("*No Code Quality issues detected.*\n\n")
		return
	}

	fmt.Printf("### Detected Issues (%d)\n\n", len(result.Smells))
	fmt.Println("| Severity | Smell | Line | Description |")
	fmt.Println("|----------|-------|------|-------------|")
	for _, s := range result.Smells {
		msg := strings.ReplaceAll(s.Message, "|", "\\|")
		fmt.Printf("| %s | %s | %d | %s |\n",
			s.Severity, s.Name, s.LineStart, msg)
		if s.AIPrompt != "" {
			fmt.Printf("  \n_%s_\n", s.AIPrompt)
		}
	}
	fmt.Println()
}

func writeMarkdownSecrets(path string, findings []secrets.SecretFinding) {
	fmt.Printf("\n### 🔑 Secret Scan: %s (%d findings)\n\n", path, len(findings))
	fmt.Println("| Severity | Rule | Line | Description |")
	fmt.Println("|----------|------|------|-------------|")
	for _, f := range findings {
		fmt.Printf("| %s | %s | %d | %s |\n",
			f.Severity, f.RuleID, f.Line, f.Description)
		fmt.Printf("  \n_%s_\n", f.Message)
	}
	fmt.Println()
}

func writeMarkdownSummary(results []analyzer.QualityResult) {
	if len(results) <= 1 {
		return
	}
	fmt.Printf("---\n\n")
	fmt.Printf("## Summary\n\n")
	fmt.Printf("- **Files analyzed:** %d\n", len(results))
	var counts = map[string]int{}
	for _, r := range results {
		counts[parser.Classify(r.Score)]++
	}
	fmt.Printf("- **Go Ahead:** %d | **Care:** %d | **Needs Work:** %d | **Stop:** %d\n",
		counts[parser.LabelGoAhead],
		counts[parser.LabelProceedWithCare],
		counts[parser.LabelNeedsWork],
		counts[parser.LabelStopRefactor])
	var highRisk []string
	for _, r := range results {
		if parser.Classify(r.Score) == parser.LabelStopRefactor {
			highRisk = append(highRisk, r.FilePath)
		}
	}
	if len(highRisk) > 0 {
		fmt.Printf("- **%s files (%d):**\n", parser.LabelStopRefactor, len(highRisk))
		for _, f := range highRisk {
			fmt.Printf("  - `%s`\n", f)
		}
	}
	fmt.Println()
}

func writeMarkdownVulnerabilities(path string, findings []vulnerability.Finding) {
	fmt.Printf("\n### Shield Vulnerability Scan: %s (%d findings)\n\n", path, len(findings))
	fmt.Println("| Severity | Category | Pattern | Description |")
	fmt.Println("|----------|----------|---------|-------------|")
	for _, f := range findings {
		fmt.Printf("| %s | %s | %s | %s |\n",
			f.Severity, f.Category, f.RuleID, f.Description)
		fmt.Printf("  \n_%s_\n", strings.ReplaceAll(f.Reminder, "\n", " "))
	}
	fmt.Println()
}
