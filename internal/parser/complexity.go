package parser

import (
	"fmt"
	"strings"
)

// Complexity counts branches (if, for, while, &&, ||) in source lines.
type ComplexityCount struct {
	Branches int
	Line     int
}

// CountBranches counts control-flow branches in a set of lines (e.g., a function body).
func CountBranches(lines []string) []ComplexityCount {
	var results []ComplexityCount
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isComment(trimmed) {
			continue
		}
		lower := strings.ToLower(trimmed)
		count := countBranchesInLine(lower)
		if count > 0 {
			results = append(results, ComplexityCount{Branches: count, Line: i + 1})
		}
	}
	return results
}

func countBranchesInLine(line string) int {
	keywords := []string{"if ", "else if ", "for ", "while ", "switch ", "case ", "&&", "||", "? "}
	count := 0
	for _, kw := range keywords {
		count += strings.Count(line, kw)
	}
	return count
}

// DetectComplexConditional checks for if/while conditions with many boolean operators.
func DetectComplexConditional(lines []string, warningThreshold int, alertThreshold int) []Smell {
	var smells []Smell
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isComment(trimmed) {
			continue
		}
		lower := strings.ToLower(trimmed)
		if !strings.HasPrefix(lower, "if ") && !strings.HasPrefix(lower, "while ") {
			continue
		}
		// Count && and || in the line (condition may or may not have parens)
		andCount := strings.Count(trimmed, "&&")
		orCount := strings.Count(trimmed, "||")
		totalBranches := andCount + orCount
		if totalBranches < warningThreshold {
			continue
		}
		sev := "warning"
		if totalBranches >= alertThreshold {
			sev = "alert"
		}
		smells = append(smells, Smell{
			Name:      "complex_conditional",
			Severity:  sev,
			LineStart: i + 1,
			LineEnd:   i + 1,
			Message:   fmt.Sprintf("Complex conditional with %d boolean branches at line %d", totalBranches, i+1),
			AIPrompt:  fmt.Sprintf("WARNING: Complex conditional with %d boolean branches at line %d. Decompose the conditional: extract parts to well-named boolean variables or helper methods (e.g., `if user.IsEligible()`).", totalBranches, i+1),
		})
	}
	return smells
}

// DetectLongParameterList checks function signatures for too many parameters.
func DetectLongParameterList(lines []string, warningThreshold int, alertThreshold int) []Smell {
	var smells []Smell
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isComment(trimmed) {
			continue
		}
		lower := strings.ToLower(trimmed)
		if !isFuncSig(lower) {
			continue
		}
		paramCount := countParams(trimmed)
		if paramCount < warningThreshold {
			continue
		}
		sev := "warning"
		if paramCount >= alertThreshold {
			sev = "alert"
		}
		smells = append(smells, Smell{
			Name:      "long_parameter_list",
			Severity:  sev,
			LineStart: i + 1,
			LineEnd:   i + 1,
			Message:   fmt.Sprintf("Function has %d parameters at line %d (warning at %d)", paramCount, i+1, warningThreshold),
			AIPrompt:  fmt.Sprintf("WARNING: Long parameter list (%d params) at line %d. Consider introducing a Parameter Object or breaking the function into smaller ones.", paramCount, i+1),
		})
	}
	return smells
}

func isFuncSig(line string) bool {
	return strings.Contains(line, "func ") ||
		strings.HasPrefix(line, "def ") ||
		strings.HasPrefix(line, "async def ")
}

func countParams(line string) int {
	start := strings.Index(line, "(")
	if start == -1 {
		return 0
	}
	end := strings.Index(line[start:], ")")
	if end == -1 {
		return 0
	}
	params := line[start+1 : start+end]
	params = strings.TrimSpace(params)
	if params == "" {
		return 0
	}
	// Naive: split by comma, but this won't handle nested generics.
	return strings.Count(params, ",") + 1
}

func isComment(line string) bool {
	return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") ||
		strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") ||
		strings.HasPrefix(line, "*/")
}
