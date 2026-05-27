package parser

import (
	"fmt"
	"math"
	"strings"
)

// DuplicationPair represents a pair of duplicated code regions.
type DuplicationPair struct {
	FuncA      string
	FuncB      string
	LineA      int
	LineB      int
	Similarity float64
	LineCount  int
}

// DetectDuplications checks for duplicated code across function bodies.
// Uses normalized fingerprint comparison with configurable minimum similarity.
func DetectDuplications(bloats []FunctionBloat, lines []string, minLines int, minSimilarity float64) []DuplicationPair {
	validFuncs := filterValidFunctions(bloats, minLines)
	fingerprints := computeFingerprints(validFuncs, lines)
	return findDuplicationPairs(validFuncs, fingerprints, minSimilarity)
}

func filterValidFunctions(bloats []FunctionBloat, minLines int) []FunctionBloat {
	validFuncs := make([]FunctionBloat, 0)
	for _, b := range bloats {
		if b.LineCount >= minLines && b.Name != "unknown" && b.Name != "anonymous" {
			validFuncs = append(validFuncs, b)
		}
	}
	return validFuncs
}

func computeFingerprints(funcs []FunctionBloat, lines []string) map[string]string {
	fingerprints := make(map[string]string)
	for _, fn := range funcs {
		end := fn.LineStart + fn.LineCount - 1
		if end > len(lines) {
			end = len(lines)
		}
		start := fn.LineStart - 1
		if start < 0 {
			start = 0
		}
		fingerprints[fn.Name] = normalizeForFingerprint(strings.Join(lines[start:end], "\n"))
	}
	return fingerprints
}

func findDuplicationPairs(funcs []FunctionBloat, fingerprints map[string]string, minSimilarity float64) []DuplicationPair {
	var pairs []DuplicationPair
	for i := 0; i < len(funcs); i++ {
		for j := i + 1; j < len(funcs); j++ {
			fnA, fnB := funcs[i], funcs[j]
			sim := jaccardSimilarity(fingerprints[fnA.Name], fingerprints[fnB.Name])
			if sim >= minSimilarity {
				pairs = append(pairs, DuplicationPair{
					FuncA: fnA.Name, FuncB: fnB.Name,
					LineA: fnA.LineStart, LineB: fnB.LineStart,
					Similarity: math.Round(sim*100) / 100,
					LineCount:  min(fnA.LineCount, fnB.LineCount),
				})
			}
		}
	}
	return pairs
}

// DetectDuplicationSmells converts duplication pairs into Smell entries.
func DetectDuplicationSmells(pairs []DuplicationPair) []Smell {
	var smells []Smell
	for _, p := range pairs {
		smells = append(smells, Smell{
			Name:      "code_duplication",
			Severity:  "warning",
			LineStart: p.LineA,
			LineEnd:   p.LineB,
			Message: fmt.Sprintf("Code duplication: '%s' (line %d) and '%s' (line %d) are %.0f%% similar (%d lines each)",
				p.FuncA, p.LineA, p.FuncB, p.LineB, p.Similarity*100, p.LineCount),
			AIPrompt: fmt.Sprintf("DRY VIOLATION: Functions '%s' and '%s' are %.0f%% similar. Extract the common logic into a shared helper function.",
				p.FuncA, p.FuncB, p.Similarity*100),
		})
	}
	return smells
}

// normalizeForFingerprint strips whitespace, comments, and normalizes identifiers.
func normalizeForFingerprint(code string) string {
	lines := strings.Split(code, "\n")
	var normalized []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isComment(trimmed) {
			continue
		}
		// Strip string literals, normalize identifiers
		cleaned := normalizeLine(trimmed)
		normalized = append(normalized, cleaned)
	}
	return strings.Join(normalized, "\n")
}

func normalizeLine(line string) string {
	// Replace string literals with placeholder
	inString := false
	var result strings.Builder
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '"' || ch == '`' {
			inString = !inString
			result.WriteByte('"')
			continue
		}
		if inString {
			result.WriteByte('x')
			continue
		}
		// Replace numbers with placeholder
		if ch >= '0' && ch <= '9' {
			result.WriteByte('0')
			continue
		}
		result.WriteByte(ch)
	}
	return result.String()
}

// jaccardSimilarity computes similarity between two texts using word-level Jaccard index.
func jaccardSimilarity(a, b string) float64 {
	wordsA := tokenize(a)
	wordsB := tokenize(b)

	if len(wordsA) == 0 && len(wordsB) == 0 {
		return 1.0
	}

	setA := make(map[string]bool)
	for _, w := range wordsA {
		setA[w] = true
	}
	setB := make(map[string]bool)
	for _, w := range wordsB {
		setB[w] = true
	}

	intersection := 0
	for w := range setA {
		if setB[w] {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func tokenize(text string) []string {
	// Split on non-alphanumeric, filter very short tokens
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_')
	})
	var tokens []string
	for _, f := range fields {
		if len(f) > 2 {
			tokens = append(tokens, strings.ToLower(f))
		}
	}
	return tokens
}

// sha256Fingerprint computes a compact fingerprint for a text segment.
