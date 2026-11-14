package parser

import (
	"fmt"
	"strings"
)

// NestingResult holds the outcome of nesting depth analysis.
type NestingResult struct {
	MaxDepth    int
	DeepestLine int
}

// AnalyzeNesting scans lines and returns the maximum nesting depth found.
func AnalyzeNesting(lines []string) NestingResult {
	var depth, maxDepth int
	var deepestLine int

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		opens := countOpeners(trimmed)
		closes := countClosers(trimmed)

		depth += opens
		if depth > maxDepth {
			maxDepth = depth
			deepestLine = i + 1
		}
		depth -= closes
		if depth < 0 {
			depth = 0
		}
	}

	return NestingResult{MaxDepth: maxDepth, DeepestLine: deepestLine}
}

// AnalyzeNestingIndent measures nesting by indentation, for languages such as Python that have no
// braces. Lines that continue a bracket expression or a docstring do not add a level.
func AnalyzeNestingIndent(lines []string) NestingResult {
	result := NestingResult{}
	for i, depth := range indentDepths(lines) {
		if depth > result.MaxDepth {
			result.MaxDepth = depth
			result.DeepestLine = i + 1
		}
	}
	return result
}

// DetectNestingSmell checks whether nesting exceeds a threshold and returns a smell if so.
func DetectNestingSmell(lines []string, warningThreshold int, alertThreshold int) *Smell {
	return nestingSmell(AnalyzeNesting(lines), warningThreshold, alertThreshold)
}

// DetectNestingSmellFor chooses brace or indentation nesting by language.
func DetectNestingSmellFor(lang string, lines []string, warningThreshold int, alertThreshold int) *Smell {
	if lang == "python" {
		return nestingSmell(AnalyzeNestingIndent(lines), warningThreshold, alertThreshold)
	}
	if lang == "javascript" || lang == "typescript" {
		return nestingSmell(AnalyzeNestingJS(lines), warningThreshold, alertThreshold)
	}
	return DetectNestingSmell(lines, warningThreshold, alertThreshold)
}

func nestingSmell(result NestingResult, warningThreshold int, alertThreshold int) *Smell {
	if result.MaxDepth < warningThreshold {
		return nil
	}
	sev := "warning"
	if result.MaxDepth >= alertThreshold {
		sev = "alert"
	}
	return &Smell{
		Name:      "deep_nesting",
		Severity:  sev,
		LineStart: result.DeepestLine,
		LineEnd:   result.DeepestLine,
		Message:   fmt.Sprintf("Nesting depth %d at line %d", result.MaxDepth, result.DeepestLine),
		AIPrompt:  fmt.Sprintf("WARNING: Deep nested complexity (depth %d) detected at line %d. Use Guard Clauses to flatten nested logic and Extract Method to separate concerns.", result.MaxDepth, result.DeepestLine),
	}
}

// indentDepths returns the indentation level of each code line, or -1 for blank, comment,
// docstring and bracket-continuation lines.
func indentDepths(lines []string) []int {
	depths := make([]int, len(lines))
	for i := range depths {
		depths[i] = -1
	}
	stack := make([]int, 0)
	bracketDepth := 0
	inTriple := false
	triple := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if inTriple {
			if strings.Contains(trimmed, triple) {
				inTriple = false
			}
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
			triple = trimmed[:3]
			inTriple = strings.Count(trimmed, triple) == 1
			continue
		}
		if bracketDepth > 0 {
			bracketDepth = max(0, bracketDepth+netBrackets(trimmed))
			continue
		}

		indent := countLeadingSpaces(line)
		for len(stack) > 0 && indent < stack[len(stack)-1] {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 || indent > stack[len(stack)-1] {
			stack = append(stack, indent)
		}
		depths[i] = len(stack) - 1
		bracketDepth = max(0, bracketDepth+netBrackets(trimmed))
	}
	return depths
}

func netBrackets(line string) int {
	opens := strings.Count(line, "(") + strings.Count(line, "[") + strings.Count(line, "{")
	closes := strings.Count(line, ")") + strings.Count(line, "]") + strings.Count(line, "}")
	return opens - closes
}

func countOpeners(line string) int {
	// Count opening braces only — each { represents one nesting level in brace languages.
	// For Python we'd use a different approach, but for MVP brace-counting is sufficient.
	// Avoid counting { inside strings/comments (rough heuristic).
	count := strings.Count(line, "{")
	return count
}

func countClosers(line string) int {
	return strings.Count(line, "}")
}

// AnalyzeNestingJS counts block braces after removing strings, template literals and comments.
func AnalyzeNestingJS(lines []string) NestingResult {
	cleaned := make([]string, len(lines))
	state := jsStripState{}
	for i, line := range lines {
		cleaned[i] = stripJSLine(line, &state)
	}
	return AnalyzeNesting(cleaned)
}

type jsStripState struct {
	block    bool
	template bool
	quote    rune
}

func stripJSLine(line string, state *jsStripState) string {
	var b strings.Builder
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if state.block {
			if ch == '*' && i+1 < len(runes) && runes[i+1] == '/' {
				state.block = false
				i++
			}
			continue
		}
		if state.template {
			if ch == '`' {
				state.template = false
			}
			continue
		}
		if state.quote != 0 {
			if ch == state.quote {
				state.quote = 0
			}
			continue
		}
		if ch == '/' && i+1 < len(runes) && runes[i+1] == '/' {
			break
		}
		if ch == '/' && i+1 < len(runes) && runes[i+1] == '*' {
			state.block = true
			i++
			continue
		}
		if ch == '\'' || ch == '"' {
			state.quote = ch
			continue
		}
		if ch == '`' {
			state.template = true
			continue
		}
		b.WriteRune(ch)
	}
	return b.String()
}
