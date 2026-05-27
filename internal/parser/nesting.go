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

// DetectNestingSmell checks whether nesting exceeds a threshold and returns a smell if so.
func DetectNestingSmell(lines []string, warningThreshold int, alertThreshold int) *Smell {
	r := AnalyzeNesting(lines)
	if r.MaxDepth < warningThreshold {
		return nil
	}

	sev := "warning"
	if r.MaxDepth >= alertThreshold {
		sev = "alert"
	}

	return &Smell{
		Name:      "deep_nesting",
		Severity:  sev,
		LineStart: r.DeepestLine,
		LineEnd:   r.DeepestLine,
		Message:   fmt.Sprintf("Nesting depth %d at line %d", r.MaxDepth, r.DeepestLine),
		AIPrompt:  fmt.Sprintf("WARNING: Deep nested complexity (depth %d) detected at line %d. Use Guard Clauses to flatten nested logic and Extract Method to separate concerns.", r.MaxDepth, r.DeepestLine),
	}
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
