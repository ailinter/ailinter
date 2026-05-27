package parser

import (
	"fmt"
	"strings"
)

// BumpRange represents one "bump" in a nesting sprawl.
type BumpRange struct {
	StartLine int `json:"start_line"`
	EndLine   int `json:"end_line"`
	MaxDepth  int `json:"max_depth"`
	LineCount int `json:"line_count"`
}

// BumpyRoadResult holds the output of a nesting sprawl analysis on one function.
type BumpyRoadResult struct {
	Bumps       []BumpRange `json:"bumps"`
	IsBumpyRoad bool        `json:"is_bumpy_road"`
	Severity    string      `json:"severity"` // "ok", "warning", "alert", "critical"
	MaxDepth    int         `json:"max_depth"`
}

// DetectBumpyRoadForFunction analyzes indentation within a function body.
func DetectBumpyRoadForFunction(lines []string, bumpDepthThreshold int, bumpsWarning int) BumpyRoadResult {
	result := BumpyRoadResult{}
	depth := 0
	inBump := false
	var currentBump *BumpRange

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isComment(trimmed) {
			continue
		}
		opens := countOpenersInLine(trimmed)
		closes := countClosersInLine(trimmed)
		netDelta := opens - closes

		// Inline braces (struct literals, interface{}, map literals):
		// if depth changes and reverts on the same line, skip bump tracking.
		if opens > 0 && closes > 0 && netDelta == 0 && opens == closes {
			continue
		}

		depth += netDelta
		if depth < 0 {
			depth = 0
		}

		if depth >= bumpDepthThreshold && !inBump {
			inBump = true
			currentBump = &BumpRange{StartLine: i + 1, MaxDepth: depth, LineCount: 0}
		}
		if inBump && currentBump != nil {
			currentBump.LineCount++
			if depth > currentBump.MaxDepth {
				currentBump.MaxDepth = depth
			}
			if depth > result.MaxDepth {
				result.MaxDepth = depth
			}
		}
		if depth < bumpDepthThreshold && inBump && currentBump != nil {
			currentBump.EndLine = i + 1
			result.Bumps = append(result.Bumps, *currentBump)
			inBump = false
			currentBump = nil
		}
	}
	if inBump && currentBump != nil {
		currentBump.EndLine = len(lines)
		result.Bumps = append(result.Bumps, *currentBump)
	}
	return classifyBumpyRoad(result, bumpsWarning)
}

func countOpenersInLine(line string) int {
	return strings.Count(line, "{")
}

func countClosersInLine(line string) int {
	return strings.Count(line, "}")
}

func classifyBumpyRoad(result BumpyRoadResult, bumpsWarning int) BumpyRoadResult {
	filtered := make([]BumpRange, 0, len(result.Bumps))
	maxDepth := 0
	for _, b := range result.Bumps {
		if b.LineCount > 3 {
			filtered = append(filtered, b)
			if b.MaxDepth > maxDepth {
				maxDepth = b.MaxDepth
			}
		}
	}
	result.Bumps = filtered
	result.MaxDepth = maxDepth

	n := len(result.Bumps)
	if n < bumpsWarning {
		result.Severity = "ok"
		return result
	}
	switch {
	case n >= 4 || result.MaxDepth >= 4:
		result.Severity = "critical"
		result.IsBumpyRoad = true
	case n >= 3 || result.MaxDepth >= 3:
		result.Severity = "alert"
		result.IsBumpyRoad = true
	case n >= bumpsWarning && result.MaxDepth >= 2:
		result.Severity = "warning"
		result.IsBumpyRoad = true
	default:
		result.Severity = "ok"
	}
	return result
}

// DetectBumpyRoadSmell wraps the nesting sprawl result into a Smell.
func DetectBumpyRoadSmell(lines []string, bumpDepth int, bumpsWarning int) *Smell {
	return DetectBumpyRoadSmellAt(lines, bumpDepth, bumpsWarning, 0)
}

func DetectBumpyRoadSmellAt(lines []string, bumpDepth int, bumpsWarning int, funcStartLine int) *Smell {
	result := DetectBumpyRoadForFunction(lines, bumpDepth, bumpsWarning)
	if !result.IsBumpyRoad {
		return nil
	}

	ranges := ""
	for i, b := range result.Bumps {
		if i > 0 {
			ranges += ", "
		}
		ranges += fmt.Sprintf("%d-%d", funcStartLine+b.StartLine-1, funcStartLine+b.EndLine-1)
	}

	lineStart := funcStartLine + result.Bumps[0].StartLine - 1
	lineEnd := funcStartLine + result.Bumps[len(result.Bumps)-1].EndLine - 1

	return &Smell{
		Name:      "bumpy_road",
		Severity:  result.Severity,
		LineStart: lineStart,
		LineEnd:   lineEnd,
		Message:   fmt.Sprintf("Bumpy Road: %d bumps (max depth %d) at ranges: %s", len(result.Bumps), result.MaxDepth, ranges),
		AIPrompt: fmt.Sprintf("WARNING: Bumpy Road detected — %d separate blocks of deeply nested logic. "+
			"COGNITIVE NOTE: Each nested block taxes working memory (limited to ~3-4 items). "+
			"REFACTORING: Use Extract Method for each bump. Create well-named private functions. "+
			"Target: 0 bumps (flat, readable 'code highway').", len(result.Bumps)),
	}
}
