package cli

import (
	"fmt"
	"strings"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/parser"
	"github.com/ailinter/ailinter/internal/secrets"
)

var severityBlockMap = map[string]string{
	"critical": "█",
	"alert":    "▓",
	"warning":  "▒",
}

var severityBlockMapColor = map[string]string{
	"critical": "\033[1;31m█\033[0m",
	"alert":    "\033[1;35m▓\033[0m",
	"warning":  "\033[33m▒\033[0m",
}

func severityBlock(severity string) string {
	if IsColorEnabled() {
		if b, ok := severityBlockMapColor[severity]; ok {
			return b
		}
	}
	if b, ok := severityBlockMap[severity]; ok {
		return b
	}
	return " "
}

func severityLabel(severity string) string {
	return strings.ToUpper(severity)
}

func severityPrefix(severity string) string {
	return severityBlock(severity) + strings.ToUpper(severity)
}

func severityColumn(severity string) string {
	s := severityPrefix(severity)
	vw := visualLen(s)
	if vw >= 9 {
		return s
	}
	return s + strings.Repeat(" ", 9-vw)
}

func scoreBar(score int) string {
	blocks := 10
	filled := score * blocks / 100
	var b strings.Builder

	color := ""
	reset := ""
	if IsColorEnabled() {
		switch parser.Classify(score) {
		case parser.LabelGoAhead:
			color = "\033[32m"
		case parser.LabelProceedWithCare:
			color = "\033[33m"
		default:
			color = "\033[31m"
		}
		reset = "\033[0m"
	}

	b.WriteString(color)
	for i := 0; i < blocks; i++ {
		if i < filled {
			b.WriteString("█")
		} else {
			b.WriteString("░")
		}
	}
	b.WriteString(reset)
	return b.String()
}

func labelBadge(label string) string {
	color := ""
	reset := ""
	if IsColorEnabled() {
		switch label {
		case parser.LabelGoAhead:
			color = "\033[32m"
		case parser.LabelProceedWithCare:
			color = "\033[33m"
		default:
			color = "\033[31m"
		}
		reset = "\033[0m"
	}
	switch label {
	case parser.LabelGoAhead:
		return color + "✓ " + label + reset
	case parser.LabelProceedWithCare:
		return color + "~ " + label + reset
	default:
		return color + "✗ " + label + reset
	}
}

func groupSmellsBySeverity(smells []analyzer.Smell) map[string][]analyzer.Smell {
	groups := map[string][]analyzer.Smell{
		"critical": {},
		"alert":    {},
		"warning":  {},
	}
	for _, s := range smells {
		groups[s.Severity] = append(groups[s.Severity], s)
	}
	return groups
}

func sortedSeverityKeys() []string {
	return []string{"critical", "alert", "warning"}
}

func groupSecretFindings(findings []secrets.SecretFinding) map[string][]secrets.SecretFinding {
	groups := map[string][]secrets.SecretFinding{
		"critical": {},
		"alert":    {},
		"warning":  {},
	}
	for _, f := range findings {
		groups[f.Severity] = append(groups[f.Severity], f)
	}
	return groups
}

func cleanPrompt(p string) string {
	p = strings.TrimPrefix(p, "WARNING: ")
	p = strings.TrimPrefix(p, "NOTE: ")
	p = strings.TrimPrefix(p, "CRITICAL: ")
	return p
}

func lastSentence(s string) string {
	last := strings.LastIndex(s, ". ")
	if last == -1 {
		return s
	}
	return strings.TrimSpace(s[last+2:])
}

func padCol(s string, width int) string {
	if len(s) > width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func severityCountParts(counts map[string]int) []string {
	var parts []string
	for _, sev := range sortedSeverityKeys() {
		if c, ok := counts[sev]; ok {
			parts = append(parts, fmt.Sprintf("%s%s: %d", severityBlock(sev), severityLabel(sev), c))
		}
	}
	return parts
}

func maybePlural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// --- retained for test compatibility ---

var severityOrder = map[string]int{"critical": 0, "alert": 1, "warning": 2}
const boxH = "─"

func cardWidthNow() int { return 100 }
func cardTop(path string)     { fmt.Printf("╭ %s ╮\n", path) }
func cardDivider()            { fmt.Println("├──┤") }
func cardBottom()             { fmt.Println("╰──╯") }
func cardGap()                { fmt.Println() }
func cardLine(format string, args ...interface{}) { fmt.Printf(format+"\n", args...) }
func cardLineRaw(text string) { fmt.Println(text) }
func renderGroupedSmells(groups map[string][]analyzer.Smell) {
	for _, sev := range []string{"critical", "alert", "warning"} {
		for _, s := range groups[sev] {
			fmt.Printf("  %s  %-18s  L%-3d  %s\n", severityPrefix(s.Severity), s.Name, s.LineStart, s.Message)
		}
	}
}

var headerPrinted bool

func writeHumanResult(result analyzer.QualityResult) {
	fmt.Printf("\n%s  %s  ·  %d lines\n", result.FilePath, result.Language, result.LinesOfCode)
	fmt.Printf("  Code Quality: %d/100  %s\n", result.Score, scoreBar(result.Score))
	fmt.Printf("  Recommendation: %s\n", labelBadge(result.Label))

	if len(result.Smells) == 0 {
		fmt.Println("  ✓ No issues detected")
		return
	}

	fmt.Printf("  %d issue%s  (%d critical, %d alert, %d warning)\n",
		len(result.Smells), maybePlural(len(result.Smells)),
		countBySeverity(result.Smells)["critical"],
		countBySeverity(result.Smells)["alert"],
		countBySeverity(result.Smells)["warning"])

	printSmellHeader()
	renderSmells(groupSmellsBySeverity(result.Smells))
	headerPrinted = false
}

func printSmellHeader() {
	if headerPrinted {
		return
	}
	header := fmt.Sprintf("  %s │ %s │ %s │ %s",
		padCol("severity", 9), padCol("issue", 20), padCol("line", 5), "description")
	if IsColorEnabled() {
		header = "\033[2m" + header + "\033[0m"
	}
	fmt.Println(header)
	headerPrinted = true
}

func renderSmells(groups map[string][]analyzer.Smell) {
	for _, sev := range sortedSeverityKeys() {
		items := groups[sev]
		if len(items) == 0 {
			continue
		}
		for _, s := range items {
			text := s.Message
			if s.AIPrompt != "" {
				p := cleanPrompt(s.AIPrompt)
				if last := lastSentence(p); last != "" && !strings.Contains(strings.ToLower(text), strings.ToLower(last)) {
					text += " — " + last
				}
			}
			sevCol := severityColumn(s.Severity)
			nameCol := fmt.Sprintf("%-20s", s.Name)
			lineCol := fmt.Sprintf("L%-4d", s.LineStart)
			fmt.Printf("  %s │ %s │ %s │ %s\n", sevCol, nameCol, lineCol, text)
		}
	}
}

func writeHumanSecrets(path string, findings []secrets.SecretFinding) {
	fmt.Printf("\n%s  ·  secrets\n", path)
	fmt.Printf("  %d secret%s detected\n", len(findings), maybePlural(len(findings)))

	groups := groupSecretFindings(findings)
	for _, sev := range sortedSeverityKeys() {
		items := groups[sev]
		if len(items) == 0 {
			continue
		}
		for _, f := range items {
			sevCol := severityColumn(f.Severity)
			nameCol := fmt.Sprintf("%-20s", f.RuleID)
			lineCol := fmt.Sprintf("L%-4d", f.Line)
			fmt.Printf("  %s │ %s │ %s │ %s — %s\n",
				sevCol, nameCol, lineCol, f.Description, cleanPrompt(f.Message))
		}
	}
}

func writeHumanSummary(results []analyzer.QualityResult) {
	if len(results) <= 1 {
		return
	}

	var totalIssues int
	var bySev = map[string]int{}
	for _, r := range results {
		totalIssues += len(r.Smells)
		for _, s := range r.Smells {
			bySev[s.Severity]++
		}
	}

	var highRisk []string
	for _, r := range results {
		if parser.Classify(r.Score) == parser.LabelStopRefactor {
			highRisk = append(highRisk, r.FilePath)
		}
	}

	fmt.Printf("\n  Summary\n")
	fmt.Printf("  %d files analyzed\n", len(results))
	if totalIssues > 0 {
		fmt.Printf("  %d total issues  (%d critical, %d alert, %d warning)\n",
			totalIssues, bySev["critical"], bySev["alert"], bySev["warning"])
	}
	if len(highRisk) > 0 {
		fmt.Printf("  %s (%d):\n", parser.LabelStopRefactor, len(highRisk))
		for _, f := range highRisk {
			fmt.Printf("    %s\n", f)
		}
	}
}
