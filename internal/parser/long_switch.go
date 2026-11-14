package parser

import (
	"fmt"
	"strings"
)

func DetectLongSwitch(lines []string, warnThreshold int, alertThreshold int) []Smell {
	return DetectLongSwitchFor("", lines, warnThreshold, alertThreshold)
}

func DetectLongSwitchFor(lang string, lines []string, warnThreshold int, alertThreshold int) []Smell {
	var smells []Smell

	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if !isSwitchStartFor(lang, trimmed) {
			i++
			continue
		}

		branches, endLine := countSwitchBranchesFor(lang, lines, i)
		if branches >= warnThreshold {
			sev := "warning"
			if branches >= alertThreshold {
				sev = "alert"
			}
			smells = append(smells, Smell{
				Name:      "long_switch",
				Severity:  sev,
				LineStart: i + 1,
				LineEnd:   endLine,
				Message:   fmt.Sprintf("Long switch with %d branches at lines %d-%d", branches, i+1, endLine),
				AIPrompt:  fmt.Sprintf("Long switch with %d branches. LLMs are prone to errors when modifying long switch chains. Replace with map lookup, strategy pattern, or polymorphism.", branches),
			})
		}
		i = endLine
	}

	return smells
}

func isSwitchStartFor(lang string, line string) bool {
	if isSwitchStart(line) {
		return true
	}
	return lang == "ruby" && strings.HasPrefix(strings.TrimSpace(line), "case ") && !strings.Contains(line, ":")
}

func isSwitchStart(line string) bool {
	t := strings.TrimSpace(line)
	if isComment(t) || t == "" {
		return false
	}

	patterns := []string{
		"switch ", "switch(",
		"match ", "match(",
	}

	for _, p := range patterns {
		if strings.HasPrefix(t, p) || strings.Contains(t, " "+p) || strings.Contains(t, "\t"+p) {
			return true
		}
	}

	if strings.HasPrefix(t, "case ") {
		after := strings.TrimPrefix(t, "case ")
		if strings.Contains(after, "when ") || strings.Contains(after, "of ") {
			return true
		}
	}

	if strings.HasPrefix(t, "when ") || strings.Contains(t, " when ") {
		return true
	}

	return false
}

func countSwitchBranchesFor(lang string, lines []string, start int) (int, int) {
	branchCount := 0
	depth := 0
	inBlock := false
	endLine := start + 1

	for i := start; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if isComment(trimmed) || trimmed == "" {
			continue
		}

		opens := strings.Count(trimmed, "{")
		closes := strings.Count(trimmed, "}")

		depth += opens

		if depth >= 1 && opens > 0 {
			inBlock = true
		}

		if inBlock {
			if isSwitchBranchFor(lang, trimmed) {
				branchCount++
			}
		}

		depth -= closes
		if depth < 0 {
			depth = 0
		}

		if inBlock && depth == 0 {
			endLine = i + 1
			break
		}

		endLine = i + 1

		if !strings.Contains(trimmed, "{") && !strings.Contains(trimmed, "}") &&
			inBlock && depth == 0 {
			break
		}
	}

	if branchCount == 0 && start < len(lines) {
		return countIndentedBranchesFor(lang, lines, start)
	}

	return branchCount, endLine
}

func isSwitchBranchFor(lang string, line string) bool {
	if isSwitchBranch(line) {
		return true
	}
	if lang == "kotlin" && strings.Contains(line, "->") {
		return true
	}
	if (lang == "rust" || lang == "scala" || lang == "csharp") && strings.Contains(line, "=>") {
		return true
	}
	return false
}

func isSwitchBranch(line string) bool {
	t := strings.TrimSpace(line)

	markers := []string{
		"case ", "case\t",
		"default:", "default ",
	}

	for _, m := range markers {
		if strings.HasPrefix(t, m) || strings.Contains(t, " "+m) || strings.Contains(t, "\t"+m) {
			return true
		}
	}

	if strings.HasPrefix(t, "when ") && strings.Contains(t, "then") {
		return true
	}

	if strings.HasPrefix(t, "when ") && !strings.Contains(t, "=") && !strings.Contains(t, "(") {
		return true
	}

	if strings.Contains(t, "=>") {
		before := strings.TrimSpace(strings.SplitN(t, "=>", 2)[0])
		if !strings.Contains(before, "=") && (!strings.Contains(before, " ") || strings.Count(before, " ") <= 2) {
			return true
		}
	}

	return false
}

func countIndentedBranchesFor(lang string, lines []string, start int) (int, int) {
	if start >= len(lines) {
		return 0, start + 1
	}

	baseIndent := detectIndent(lines[start])
	if baseIndent < 0 {
		return 0, start + 1
	}

	branchCount := 0
	endLine := start + 1

	for i := start + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || isComment(trimmed) {
			continue
		}

		currentIndent := detectIndent(lines[i])
		if currentIndent < baseIndent || (lang != "ruby" && currentIndent <= baseIndent) {
			endLine = i
			break
		}

		if isIndentedBranchFor(lang, trimmed) {
			branchCount++
		}
		endLine = i + 1
	}

	return branchCount, endLine
}

func isIndentedBranchFor(lang string, line string) bool {
	if isIndentedBranch(line) {
		return true
	}
	return lang == "kotlin" && strings.Contains(line, "->")
}

func isIndentedBranch(line string) bool {
	t := strings.TrimSpace(line)

	if strings.HasPrefix(t, "case ") || strings.HasPrefix(t, "case\t") {
		return true
	}

	if strings.HasPrefix(t, "when ") || strings.Contains(t, " when ") {
		return true
	}

	return false
}

func detectIndent(line string) int {
	count := 0
	for _, c := range line {
		if c == ' ' {
			count++
		} else if c == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}
