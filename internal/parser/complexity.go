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
	return DetectLongParameterListFor("", lines, warningThreshold, alertThreshold)
}

// DetectLongParameterListFor recognises signatures in the language of the file and counts
// parameters across a possibly multi-line parameter list.
func DetectLongParameterListFor(lang string, lines []string, warningThreshold int, alertThreshold int) []Smell {
	var smells []Smell
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || isComment(trimmed) || !isSignatureLine(lang, trimmed) {
			continue
		}
		startLine := i + 1
		params, end := signatureParams(lang, lines, i)
		i = end
		paramCount := countParams(params)
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
			LineStart: startLine,
			LineEnd:   startLine,
			Message:   fmt.Sprintf("Function has %d parameters at line %d (warning at %d)", paramCount, startLine, warningThreshold),
			AIPrompt:  fmt.Sprintf("WARNING: Long parameter list (%d params) at line %d. Consider introducing a Parameter Object or breaking the function into smaller ones.", paramCount, startLine),
		})
	}
	return smells
}

var signatureStopWords = map[string]bool{
	"if": true, "for": true, "while": true, "switch": true, "catch": true,
	"return": true, "new": true, "throw": true, "await": true, "else": true,
	"do": true, "case": true, "when": true, "guard": true, "defer": true,
	"go": true, "range": true, "match": true,
}

// isSignatureLine uses the language first, then a conservative generic heuristic: a line with a
// parameter list whose prefix is not a control keyword or a call.
func isSignatureLine(lang string, line string) bool {
	lower := strings.ToLower(line)
	if signatureStopWords[strings.Fields(lower)[0]] {
		return false
	}
	switch lang {
	case "go":
		return strings.Contains(lower, "func ")
	case "python":
		return strings.HasPrefix(lower, "def ") || strings.HasPrefix(lower, "async def ")
	case "ruby":
		return strings.HasPrefix(lower, "def ")
	case "rust":
		return strings.HasPrefix(lower, "fn ") || strings.Contains(lower, " fn ")
	case "kotlin":
		return strings.Contains(lower, "fun ")
	case "swift":
		return strings.Contains(lower, "func ")
	case "php":
		return strings.Contains(lower, "function ")
	}
	return looksLikeSignature(lower, line)
}

func looksLikeSignature(lower string, line string) bool {
	open := strings.Index(line, "(")
	if open < 0 {
		return false
	}
	prefix := strings.TrimSpace(line[:open])
	if prefix == "" {
		return false
	}
	fields := strings.Fields(prefix)
	if len(fields) >= 2 && !strings.Contains(prefix, "=") {
		return true
	}
	if !strings.Contains(line[open:], ")") {
		return false
	}
	if strings.HasPrefix(lower, "constructor(") {
		return true
	}
	tail := strings.TrimSpace(line[strings.LastIndex(line, ")")+1:])
	if tail == "{" || strings.HasPrefix(tail, "{") || strings.HasPrefix(tail, "=>") {
		return true
	}
	return false
}

func stripGoReceiver(lang string, line string) string {
	if lang != "go" || !strings.HasPrefix(strings.TrimSpace(line), "func (") {
		return line
	}
	open := strings.Index(line, "(")
	close := strings.Index(line[open:], ")")
	if close < 0 {
		return line
	}
	return line[open+close+1:]
}

// signatureParams returns the text inside the first parameter list and the line where it ends.
func signatureParams(lang string, lines []string, start int) (string, int) {
	end := start
	text := stripGoReceiver(lang, lines[start])
	for strings.Count(text, "(") > strings.Count(text, ")") && end+1 < len(lines) && end < start+20 {
		end++
		text += " " + lines[end]
	}
	open := strings.Index(text, "(")
	if open < 0 {
		return "", start
	}
	close := strings.Index(text[open+1:], ")")
	if close < 0 {
		return "", end
	}
	return text[open+1 : open+1+close], end
}

func countParams(params string) int {
	params = strings.TrimSpace(params)
	if params == "" || strings.EqualFold(params, "void") {
		return 0
	}
	depth := 0
	count := 1
	var quote rune
	for _, ch := range params {
		quote = updateQuote(quote, ch)
		depth = updateDepth(depth, ch)
		count += commaCount(quote, depth, ch)
	}
	return count
}

func commaCount(quote rune, depth int, ch rune) int {
	if quote != 0 {
		return 0
	}
	if ch != ',' {
		return 0
	}
	if depth != 0 {
		return 0
	}
	return 1
}

func updateQuote(current rune, ch rune) rune {
	if current == 0 && (ch == '\'' || ch == '"' || ch == '`') {
		return ch
	}
	if ch == current {
		return 0
	}
	return current
}

var openParens = map[rune]bool{'(': true, '[': true, '{': true, '<': true}
var closeParens = map[rune]bool{')': true, ']': true, '}': true, '>': true}

func updateDepth(depth int, ch rune) int {
	if openParens[ch] {
		return depth + 1
	}
	if closeParens[ch] && depth > 0 {
		return depth - 1
	}
	return depth
}

func isComment(line string) bool {
	return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") ||
		strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") ||
		strings.HasPrefix(line, "*/")
}

func isFuncSig(line string) bool {
	return strings.Contains(line, "func ") || strings.HasPrefix(line, "def ") || strings.HasPrefix(line, "async def ")
}
