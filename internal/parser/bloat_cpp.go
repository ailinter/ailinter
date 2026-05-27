package parser

import "strings"

// DetectFunctionBloatsCPP detects C/C++ function boundaries.
func DetectFunctionBloatsCPP(lines []string) []FunctionBloat {
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
			if isFuncDefCPP(trimmed) {
				funcName = extractFuncNameCPP(trimmed)
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

		if isFuncDefCPP(trimmed) {
			funcName = extractFuncNameCPP(trimmed)
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

var cppNotFuncPrefixes = []string{"if ", "else ", "for ", "while ", "switch ", "return ", "break", "continue", "case ", "default:", "goto ", "try ", "catch ", "throw ", "enum ", "struct ", "class ", "namespace ", "template", "using ", "typedef ", "public:", "private:", "protected:", "#", "//"}

func isFuncDefCPP(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
		return false
	}
	if !strings.Contains(trimmed, "(") {
		return false
	}
	lower := strings.ToLower(trimmed)
	for _, p := range cppNotFuncPrefixes {
		if strings.HasPrefix(lower, p) {
			return false
		}
	}

	// Destructors: ~ClassName() { or ClassName::~ClassName() {
	if strings.Contains(trimmed, "~") && strings.Contains(trimmed, "(") {
		tildeIdx := strings.Index(trimmed, "~")
		parenIdx := strings.Index(trimmed, "(")
		if tildeIdx < parenIdx {
			return true
		}
	}

	// Operator overloads: operator+ / operator== / operator<< / etc.
	if strings.Contains(lower, "operator") && strings.Contains(trimmed, "(") {
		return true
	}

	// Constructor initializer lists: Foo() : bar(1), baz(2)
	// Already handled by the general ( detection below

	before := strings.TrimSpace(trimmed[:strings.Index(trimmed, "(")])
	if before == "" {
		return false
	}
	words := strings.Fields(before)
	return len(words) >= 1 && (strings.HasSuffix(trimmed, "{") || strings.HasSuffix(trimmed, ")") || strings.Contains(trimmed, ") {") || strings.Contains(trimmed, ") :"))
}

func extractFuncNameCPP(line string) string {
	idx := strings.Index(line, "(")
	if idx == -1 {
		return "unknown"
	}
	words := strings.Fields(strings.TrimSpace(line[:idx]))
	if len(words) == 0 {
		return "unknown"
	}
	name := words[len(words)-1]
	if ci := strings.LastIndex(name, "::"); ci != -1 {
		name = name[ci+2:]
	}
	return name
}
