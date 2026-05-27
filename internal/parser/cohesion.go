package parser

import (
	"fmt"
	"strings"
)

// CohesionResult describes module-level cohesion.
type CohesionResult struct {
	TotalFuncs      int
	SharedTypeCount int
	IsolatedFuncs   int
	CohesionScore   float64 // 0-1, higher = more cohesive
	IsLowCohesion   bool
}

// AnalyzeCohesion estimates module cohesion by checking how many functions
// share common types in their signatures (parameters, returns).
// Low cohesion: many isolated functions that don't share types with others.
func AnalyzeCohesion(bloats []FunctionBloat, lines []string) CohesionResult {
	validFuncs := make([]FunctionBloat, 0)
	for _, b := range bloats {
		if b.Name != "unknown" && b.Name != "anonymous" && b.LineCount >= 2 {
			validFuncs = append(validFuncs, b)
		}
	}

	if len(validFuncs) <= 3 {
		return CohesionResult{TotalFuncs: len(validFuncs), CohesionScore: 1.0}
	}

	// Extract types referenced in each function's first few lines (signature+body start)
	funcTypes := make(map[string]map[string]bool) // funcName -> set of types
	for _, fn := range validFuncs {
		end := fn.LineStart + fn.LineCount - 1
		if end > len(lines) {
			end = len(lines)
		}
		start := fn.LineStart - 1
		if start < 0 {
			start = 0
		}
		// Only look at first ~10 lines of function for type references
		scopeEnd := start + 10
		if scopeEnd > end {
			scopeEnd = end
		}
		body := strings.Join(lines[start:scopeEnd], " ")
		types := extractTypeReferences(body)
		funcTypes[fn.Name] = types
	}

	// Count shared types across functions
	totalTypeCount := 0
	isolatedCount := 0
	for fnName, types := range funcTypes {
		if len(types) == 0 || !hasSharedType(fnName, types, funcTypes) {
			isolatedCount++
		}
		totalTypeCount += len(types)
	}

	// Cohesion score: ratio of functions with shared types
	n := len(validFuncs)
	sharedCount := n - isolatedCount
	score := float64(sharedCount) / float64(n)

	isLow := score < 0.5 && n >= 5

	return CohesionResult{
		TotalFuncs:    n,
		IsolatedFuncs: isolatedCount,
		CohesionScore: score,
		IsLowCohesion: isLow,
	}
}

// DetectLowCohesionSmell converts cohesion analysis to a Smell.
func DetectLowCohesionSmell(result CohesionResult, warningThreshold int, badThreshold int) *Smell {
	if !result.IsLowCohesion {
		return nil
	}
	sev := "warning"
	if result.CohesionScore < 0.3 {
		sev = "alert"
	}

	return &Smell{
		Name:     "low_cohesion",
		Severity: sev,
		Message:  fmt.Sprintf("Low cohesion: %d/%d functions are isolated (score: %.1f)", result.IsolatedFuncs, result.TotalFuncs, result.CohesionScore*100),
		AIPrompt: fmt.Sprintf("WARNING: Low cohesion detected — %d of %d functions don't share types with others. Consider extracting unrelated functions into separate modules (SRP).",
			result.IsolatedFuncs, result.TotalFuncs),
	}
}

// extractTypeReferences finds type-like identifiers in function signatures.
func extractTypeReferences(text string) map[string]bool {
	types := make(map[string]bool)

	// Common type patterns in Go, C++, Python, etc.
	// Look for capitalized words (likely types in most languages)
	words := strings.Fields(text)
	for _, w := range words {
		// Strip punctuation
		w = strings.Trim(w, "()[]{}*&,.;:\"'")
		if len(w) < 3 {
			continue
		}
		// Skip common keywords
		lower := strings.ToLower(w)
		switch lower {
		case "func", "function", "def", "var", "const", "return", "if", "for",
			"while", "switch", "case", "break", "continue", "else", "nil",
			"null", "true", "false", "error", "err", "int", "string",
			"float64", "float32", "bool", "byte", "void", "auto", "self",
			"this", "new", "make", "len", "append", "range", "defer", "go":
			continue
		}
		// Capitalized words are likely type names
		if w[0] >= 'A' && w[0] <= 'Z' {
			types[w] = true
		}
		// Pointer types: *Type
		if strings.HasPrefix(w, "*") && len(w) > 1 {
			typeName := strings.TrimPrefix(w, "*")
			if typeName[0] >= 'A' && typeName[0] <= 'Z' {
				types[typeName] = true
			}
		}
	}
	return types
}

func hasSharedType(fnName string, types map[string]bool, all map[string]map[string]bool) bool {
	for otherName, otherTypes := range all {
		if fnName == otherName {
			continue
		}
		for t := range types {
			if otherTypes[t] {
				return true
			}
		}
	}
	return false
}
