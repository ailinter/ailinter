package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// ChainPattern matches method chains like a.b().c()
var ChainPattern = regexp.MustCompile(`\.\w+\([^)]*\)\.\w+\(`)

func DetectMessageChains(lines []string) []Smell {
	var smells []Smell
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isComment(trimmed) {
			continue
		}
		if len(ChainPattern.FindAllStringIndex(trimmed, -1)) >= 2 {
			smells = append(smells, Smell{
				Name: "message_chains", Severity: "warning",
				LineStart: i + 1, LineEnd: i + 1,
				Message:  fmt.Sprintf("Message chain detected at line %d", i+1),
				AIPrompt: fmt.Sprintf("NOTE: Long message chain at line %d. Consider using Hide Delegate pattern.", i+1),
			})
		}
	}
	return smells
}

func DetectPrimitiveObsession(lines []string) []Smell {
	var smells []Smell
	primitiveTypes := []string{"string", "int", "float64", "float32", "bool", "int64", "int32", "byte", "rune", "uint"}
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !isFuncSig(strings.ToLower(trimmed)) {
			continue
		}
		primCount := 0
		lower := strings.ToLower(trimmed)
		for _, pt := range primitiveTypes {
			// Count type occurrences followed by comma, paren, space, or end-of-line
			primCount += strings.Count(lower, pt+",")
			primCount += strings.Count(lower, pt+")")
			primCount += strings.Count(lower, pt+" ")
			// Check if line ends with the type
			if strings.HasSuffix(lower, pt) {
				primCount++
			}
		}
		if primCount >= 4 {
			smells = append(smells, Smell{
				Name: "primitive_obsession", Severity: "warning",
				LineStart: i + 1, LineEnd: i + 1,
				Message:  fmt.Sprintf("Primitive obsession: %d primitive-type parameters at line %d", primCount, i+1),
				AIPrompt: fmt.Sprintf("NOTE: %d primitive-type parameters at line %d. Consider introducing Value Objects.", primCount, i+1),
			})
		}
	}
	return smells
}

func DetectLazyElements(bloats []FunctionBloat, minLines int) []Smell {
	var smells []Smell
	for _, b := range bloats {
		if b.Name == "unknown" || b.Name == "anonymous" {
			continue
		}
		if b.LineCount < minLines && b.LineCount > 0 {
			smells = append(smells, Smell{
				Name: "lazy_element", Severity: "warning",
				LineStart: b.LineStart,
				Message:   fmt.Sprintf("Lazy element: function '%s' is only %d lines — consider inlining", b.Name, b.LineCount),
				AIPrompt:  fmt.Sprintf("NOTE: Function '%s' is very small (%d lines). Consider inlining if called once.", b.Name, b.LineCount),
			})
		}
	}
	return smells
}

func DetectParagraphOfCode(lines []string, maxConsecutive int) []Smell {
	var smells []Smell
	consecutive := 0
	startLine := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isComment(trimmed) {
			if consecutive > maxConsecutive {
				smells = append(smells, Smell{
					Name: "paragraph_of_code", Severity: "warning",
					LineStart: startLine, LineEnd: i,
					Message:  fmt.Sprintf("Paragraph of %d consecutive non-blank lines at lines %d-%d", consecutive, startLine, i),
					AIPrompt: fmt.Sprintf("NOTE: Paragraph of %d consecutive lines.", consecutive),
				})
			}
			consecutive = 0
		} else {
			if consecutive == 0 {
				startLine = i + 1
			}
			consecutive++
		}
	}
	// Trailing paragraph at end of file
	if consecutive > maxConsecutive {
		smells = append(smells, Smell{
			Name: "paragraph_of_code", Severity: "warning",
			LineStart: startLine, LineEnd: len(lines),
			Message:  fmt.Sprintf("Paragraph of %d consecutive non-blank lines at lines %d-%d", consecutive, startLine, len(lines)),
			AIPrompt: fmt.Sprintf("NOTE: Paragraph of %d consecutive lines.", consecutive),
		})
	}
	return smells
}

func deduplicateSmells(smells []Smell) []Smell {
	seen := make(map[string]bool)
	var result []Smell
	for _, s := range smells {
		key := fmt.Sprintf("%s:%d:%s", s.Name, s.LineStart, s.Severity)
		if !seen[key] {
			seen[key] = true
			result = append(result, s)
		}
	}
	return result
}

// DetectExcessiveComments flag files where comment lines exceed a ratio of total lines.
func DetectExcessiveComments(lines []string, ratioThreshold float64) *Smell {
	if len(lines) < 10 {
		return nil
	}
	commentLines := 0
	nonBlank := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		nonBlank++
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, " *") ||
			strings.HasPrefix(trimmed, "*/") {
			commentLines++
		}
	}
	if nonBlank == 0 {
		return nil
	}
	ratio := float64(commentLines) / float64(nonBlank)
	if ratio < ratioThreshold {
		return nil
	}
	sev := "warning"
	if ratio >= ratioThreshold*2 {
		sev = "alert"
	}
	return &Smell{
		Name:     "excessive_comments",
		Severity: sev,
		Message:  fmt.Sprintf("Excessive comments: %.0f%% of non-blank lines are comments (%d/%d)", ratio*100, commentLines, nonBlank),
		AIPrompt: fmt.Sprintf("NOTE: %.0f%% of this file is comments (%d/%d lines). Consider renaming functions and variables to be self-documenting instead of relying on comments.",
			ratio*100, commentLines, nonBlank),
	}
}

// DetectGlobalData flags file-level mutable variable declarations outside functions.
func DetectGlobalData(lines []string, minGlobals int) *Smell {
	globalCount := 0
	inFunction := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isComment(trimmed) {
			continue
		}
		if strings.Contains(trimmed, "func ") || strings.HasPrefix(trimmed, "def ") ||
			strings.HasPrefix(trimmed, "class ") {
			inFunction = true
			continue
		}
		if trimmed == "}" || (inFunction && trimmed == "") {
			inFunction = false
			continue
		}
		if inFunction {
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "var ") || strings.HasPrefix(lower, "const ") ||
			(strings.Contains(lower, " = ") && !strings.Contains(lower, "func") && !strings.Contains(lower, "import") && !strings.Contains(lower, "package")) {
			globalCount++
		}
	}
	if globalCount < minGlobals {
		return nil
	}
	return &Smell{
		Name:     "global_data",
		Severity: "warning",
		Message:  fmt.Sprintf("Global data: %d mutable top-level declarations", globalCount),
		AIPrompt: fmt.Sprintf("NOTE: %d top-level mutable variable declarations detected. Consider encapsulating global state in structs, using dependency injection, or making globals immutable.",
			globalCount),
	}
}

// DetectLongScopeVariables flags variables declared far from their last usage.
func DetectLongScopeVariables(lines []string, bloats []FunctionBloat, minLineGap int) []Smell {
	var smells []Smell
	for _, fn := range bloats {
		if fn.LineCount < minLineGap || fn.Name == "unknown" || fn.Name == "anonymous" {
			continue
		}
		end := fn.LineStart + fn.LineCount - 1
		if end > len(lines) {
			end = len(lines)
		}
		start := fn.LineStart - 1
		if start < 0 {
			start = 0
		}
		body := lines[start:end]

		declared := make(map[string]int) // varName → declaration line
		for i, line := range body {
			trimmed := strings.TrimSpace(line)
			if isComment(trimmed) || trimmed == "" {
				continue
			}
			absLine := start + i + 1
			// Detect variable declarations: var x =, x :=, let x =, const x =
			if varName := extractVarDecl(trimmed); varName != "" {
				if _, exists := declared[varName]; !exists {
					declared[varName] = absLine
				}
			}
			// Check all declared names for usage on this line
			for vName, declLine := range declared {
				if strings.Contains(trimmed, vName) && absLine > declLine {
					lastUse := absLine
					gap := lastUse - declLine
					if gap >= minLineGap {
						// Only flag if the variable is used across a long gap
						// and the declaration is near the start of the function
						if declLine <= fn.LineStart+10 {
							smells = append(smells, Smell{
								Name:      "long_scope_variable",
								Severity:  "warning",
								LineStart: declLine,
								LineEnd:   lastUse,
								Message:   fmt.Sprintf("Variable '%s' declared at line %d, last used at line %d (%d line gap)", vName, declLine, lastUse, gap),
								AIPrompt: fmt.Sprintf("NOTE: Variable '%s' spans %d lines (declared at line %d, used at line %d). Consider reducing scope by introducing a new function or moving the declaration closer to its use.",
									vName, gap, declLine, lastUse),
							})
						}
						delete(declared, vName) // Only flag once per variable
					} else {
						// Update last use position
						declared[vName] = absLine
					}
					break
				}
			}
		}
	}
	return smells
}

func extractVarDecl(line string) string {
	lower := strings.ToLower(line)
	if strings.HasPrefix(lower, "var ") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	if strings.Contains(lower, ":=") {
		parts := strings.SplitN(line, ":=", 2)
		name := strings.TrimSpace(parts[0])
		if name != "" {
			return name
		}
	}
	if strings.HasPrefix(lower, "let ") || strings.HasPrefix(lower, "const ") {
		parts := strings.Fields(line)
		for i, p := range parts {
			if i > 0 && p != "=" && p != "" && !strings.HasPrefix(p, "//") {
				return strings.TrimRight(p, "=;")
			}
		}
	}
	return ""
}
