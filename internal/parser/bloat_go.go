package parser

import "strings"

// DetectFunctionBloats identifies function boundaries for brace-based languages.
func DetectFunctionBloats(lines []string) []FunctionBloat {
	var results []FunctionBloat
	var inFunc bool
	var awaitingBody bool
	var funcName string
	var funcStart, braceDepth int

	for i, line := range lines {
		trimmed := trimLine(line)

		if awaitingBody {
			depthChange := countOpens(trimmed) - countCloses(trimmed)
			braceDepth += depthChange
			if braceDepth > 0 {
				awaitingBody = false
				inFunc = true
			}
			if isFuncDefGo(trimmed) {
				funcName = extractFuncNameGo(trimmed)
				funcStart = i + 1
				diff := countOpens(trimmed) - countCloses(trimmed)
				braceDepth = diff
				if diff > 0 {
					awaitingBody = false
					inFunc = true
				}
			}
			continue
		}

		if inFunc {
			braceDepth += countOpens(trimmed) - countCloses(trimmed)
			if braceDepth <= 0 {
				results = append(results, FunctionBloat{
					Name: funcName, LineCount: (i + 1) - funcStart + 1, LineStart: funcStart,
				})
				inFunc = false
			}
			continue
		}

		if isFuncDefGo(trimmed) {
			funcName = extractFuncNameGo(trimmed)
			funcStart = i + 1
			opens := countOpens(trimmed)
			closes := countCloses(trimmed)
			braceDepth = opens - closes
			if braceDepth > 0 {
				inFunc = true
			} else {
				awaitingBody = true
			}
		}
	}
	return results
}

func isFuncDefGo(line string) bool {
	l := strings.ToLower(line)
	return strings.Contains(l, "func ") || strings.HasPrefix(l, "func(") ||
		strings.Contains(l, "function ") || strings.Contains(l, "= function(") ||
		(strings.Contains(l, "=>") && strings.Contains(l, "{"))
}

func extractFuncNameGo(line string) string {
	// JavaScript: function foo() { }
	if idx := strings.Index(line, "function "); idx != -1 {
		rest := strings.TrimSpace(line[idx+9:])
		if end := strings.IndexAny(rest, " (\n"); end != -1 {
			return strings.TrimSpace(rest[:end])
		}
		return "anonymous"
	}
	// JavaScript/TS: const foo = (...) => {
	if strings.Contains(line, "=>") && strings.Contains(line, "=") {
		before := strings.TrimSpace(line[:strings.Index(line, "=")])
		words := strings.Fields(before)
		if len(words) >= 2 {
			return words[len(words)-1]
		}
		return "arrow"
	}
	// Go: func foo() {
	idx := strings.Index(line, "func ")
	if idx == -1 {
		return "unknown"
	}
	rest := strings.TrimSpace(line[idx+5:])
	if strings.HasPrefix(rest, "(") {
		depth := 1
		i := 1
		for i < len(rest) && depth > 0 {
			if rest[i] == '(' {
				depth++
			}
			if rest[i] == ')' {
				depth--
			}
			i++
		}
		rest = strings.TrimSpace(rest[i:])
	}
	end := strings.Index(rest, "(")
	if end == -1 {
		end = len(rest)
	}
	return strings.TrimSpace(rest[:end])
}
