package parser

import "strings"

// DetectFunctionBloatsJava detects Java/C# method boundaries.
func DetectFunctionBloatsJava(lines []string) []FunctionBloat {
	return runBraceDetector(lines, isFuncDefJava, extractFuncNameJava)
}

// DetectFunctionBloatsRust detects Rust fn boundaries.
func DetectFunctionBloatsRust(lines []string) []FunctionBloat {
	return runBraceDetector(lines, isFuncDefRust, extractFuncNameRust)
}

// DetectFunctionBloatsRuby detects Ruby def/end boundaries.
func DetectFunctionBloatsRuby(lines []string) []FunctionBloat {
	return detectByKeywords(lines, "def ", "end")
}

// DetectFunctionBloatsSwift detects Swift func boundaries.
func DetectFunctionBloatsSwift(lines []string) []FunctionBloat {
	return runBraceDetector(lines, isFuncDefSwift, extractFuncNameSwift)
}

// DetectFunctionBloatsKotlin detects Kotlin fun boundaries.
func DetectFunctionBloatsKotlin(lines []string) []FunctionBloat {
	return runBraceDetector(lines, isFuncDefKotlin, extractFuncNameKotlin)
}

// detectByKeywords tracks function bodies delimited by keyword pairs (e.g., def/end for Ruby).
func detectByKeywords(lines []string, openKW string, closeKW string) []FunctionBloat {
	var results []FunctionBloat
	depth := 0
	type funcState struct {
		name       string
		startLine  int
		startDepth int
	}
	var active []funcState

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		lower := strings.ToLower(trimmed)

		opens := countKeywordOccurrences(lower, openKW) + countBraceLikeOpens(lower)
		closes := countKeywordOccurrences(lower, closeKW) + countCloses(lower)

		// Start a function whenever we see the open keyword (after trimming)
		if strings.HasPrefix(lower, openKW) || strings.Contains(lower, " "+openKW) || strings.HasPrefix(lower, "self."+openKW) {
			name := extractRubyFuncName(trimmed, openKW)
			active = append(active, funcState{name: name, startLine: i + 1, startDepth: depth})
		}

		depth += opens - closes
		if depth < 0 {
			depth = 0
		}

		// End functions whose start depth >= current depth
		for j := len(active) - 1; j >= 0; j-- {
			if depth <= active[j].startDepth {
				results = append(results, FunctionBloat{
					Name: active[j].name, LineCount: (i + 1) - active[j].startLine + 1, LineStart: active[j].startLine,
				})
				active = append(active[:j], active[j+1:]...)
			}
		}
	}
	// Remaining active functions
	for _, f := range active {
		results = append(results, FunctionBloat{
			Name: f.name, LineCount: len(lines) - f.startLine + 1, LineStart: f.startLine,
		})
	}
	return results
}

func countKeywordOccurrences(line, kw string) int {
	n := 0
	rest := line
	for {
		idx := strings.Index(rest, kw)
		if idx == -1 {
			break
		}
		before := idx == 0 || !isIdentChar(rest[idx-1])
		// If keyword ends with space, the space itself is the boundary
		after := strings.HasSuffix(kw, " ") || idx+len(kw) >= len(rest) || !isIdentChar(rest[idx+len(kw)])
		if before && after {
			n++
		}
		rest = rest[idx+1:]
	}
	return n
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func countBraceLikeOpens(line string) int {
	keywords := []string{"do", "if", "unless", "while", "until", "for", "case", "begin", "module", "class"}
	n := 0
	for _, kw := range keywords {
		n += countKeywordOccurrences(line, kw)
	}
	n += strings.Count(line, "{")
	return n
}

func extractRubyFuncName(line string, kw string) string {
	rest := strings.TrimPrefix(line, kw)
	rest = strings.TrimSpace(rest)
	// Handle self.method_name
	if strings.HasPrefix(rest, "self.") {
		rest = rest[5:]
	}
	idx := strings.IndexAny(rest, " (\n")
	if idx == -1 {
		idx = len(rest)
	}
	return strings.TrimSpace(rest[:idx])
}

func isFuncDefSwift(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
		return false
	}
	lower := strings.ToLower(trimmed)
	notFunc := []string{"if ", "else ", "for ", "while ", "switch ", "guard ", "return ", "break", "continue", "case ", "default:", "import ", "var ", "let "}
	for _, p := range notFunc {
		if strings.HasPrefix(lower, p) {
			return false
		}
	}
	return (strings.HasPrefix(lower, "func ") || strings.HasPrefix(lower, "private func ") ||
		strings.HasPrefix(lower, "public func ") || strings.HasPrefix(lower, "internal func ") ||
		strings.HasPrefix(lower, "override func ") || strings.HasPrefix(lower, "mutating func ") ||
		strings.HasPrefix(lower, "static func ") || strings.HasPrefix(lower, "class func ")) &&
		strings.Contains(trimmed, "(")
}

func extractFuncNameSwift(line string) string {
	trimmed := strings.TrimSpace(line)
	// Strip modifiers
	for _, p := range []string{"private ", "public ", "internal ", "override ", "mutating ", "static ", "class "} {
		if strings.HasPrefix(strings.ToLower(trimmed), p) {
			trimmed = strings.TrimSpace(trimmed[len(p):])
		}
	}
	trimmed = strings.TrimPrefix(trimmed, "func ")
	idx := strings.Index(trimmed, "(")
	if idx == -1 {
		idx = len(trimmed)
	}
	return strings.TrimSpace(trimmed[:idx])
}

func isFuncDefKotlin(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
		return false
	}
	lower := strings.ToLower(trimmed)
	notFunc := []string{"if ", "else ", "for ", "while ", "when ", "return ", "break", "continue", "import ", "package ", "val ", "var ", "class ", "interface ", "object ", "enum ", "data class", "sealed "}
	for _, p := range notFunc {
		if strings.HasPrefix(lower, p) {
			return false
		}
	}

	// Strip Kotlin modifiers to find the 'fun' keyword
	keywordFound := false
	rest := lower
	for {
		rest = strings.TrimSpace(rest)
		if strings.HasPrefix(rest, "fun ") {
			keywordFound = true
			break
		}
		// Known modifiers that can precede 'fun'
		modifiers := []string{"private ", "public ", "protected ", "internal ", "override ", "open ", "final ",
			"suspend ", "inline ", "operator ", "infix ", "tailrec ", "external ",
			"abstract ", "expect ", "actual "}
		matched := false
		for _, m := range modifiers {
			if strings.HasPrefix(rest, m) {
				rest = rest[len(m):]
				matched = true
				break
			}
		}
		// Also handle annotations: @Something fun ...
		if strings.HasPrefix(rest, "@") {
			spaceIdx := strings.Index(rest, " ")
			if spaceIdx != -1 {
				rest = rest[spaceIdx+1:]
				matched = true
			}
		}
		if !matched {
			break
		}
	}
	if !keywordFound {
		return false
	}

	c := strings.TrimSpace(trimmed)
	return strings.Contains(c, "(") && (strings.HasSuffix(c, "{") || strings.HasSuffix(c, ")") || strings.Contains(c, ") =") || strings.Contains(c, "):"))
}

func extractFuncNameKotlin(line string) string {
	// Find the 'fun' keyword and extract name after it
	lower := strings.ToLower(line)
	idx := strings.Index(lower, "fun ")
	if idx == -1 {
		return "unknown"
	}
	rest := strings.TrimSpace(line[idx+4:])
	// Handle <T> generics after fun
	if strings.HasPrefix(rest, "<") {
		gt := strings.Index(rest, ">")
		if gt != -1 {
			rest = strings.TrimSpace(rest[gt+1:])
		}
	}
	paren := strings.Index(rest, "(")
	space := strings.Index(rest, " ")
	colon := strings.Index(rest, ":")
	// Function name is the token before ( or : or space
	end := paren
	if end == -1 {
		end = len(rest)
	}
	// Check for return type colon before (
	if colon != -1 && colon < end {
		end = colon
	}
	// Check for space before return type
	if space != -1 && space < end {
		end = space
	}
	return strings.TrimSpace(rest[:end])
}

// DetectIndentSize auto-detects whether a file uses spaces or tabs for indentation.
func DetectIndentSize(lines []string) int {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Check tabs first — countLeadingSpaces counts tabs as 4 spaces
		tabs := countLeadingTabs(line)
		if tabs > 0 {
			return 1
		}
		spaces := countLeadingSpaces(line)
		if spaces > 0 {
			return 4
		}
	}
	return 4
}

func countLeadingTabs(line string) int {
	n := 0
	for n < len(line) && line[n] == '\t' {
		n++
	}
	return n
}

// DetectFunctionBloatsIndent detects Python functions using indentation.
// Handles: def, async def, @decorators, class methods, nested functions, tabs/spaces.
func DetectFunctionBloatsIndent(lines []string, indentSize int) []FunctionBloat {
	var results []FunctionBloat
	type funcStack struct {
		name   string
		start  int
		indent int
	}
	var stack []funcStack

	measureIndent := func(line string) int {
		if indentSize == 1 {
			return countLeadingTabs(line)
		}
		return countLeadingSpaces(line) / indentSize
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Skip decorators — they precede the function definition
		if strings.HasPrefix(trimmed, "@") && (i+1 < len(lines)) {
			continue
		}

		indent := measureIndent(line)

		// Check whether we detect a new function
		if strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "async def ") {
			// Close any inner functions that are deeper than this new function
			for len(stack) > 0 && indent <= stack[len(stack)-1].indent {
				f := stack[len(stack)-1]
				results = append(results, FunctionBloat{
					Name: f.name, LineCount: i - f.start + 1, LineStart: f.start,
				})
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, funcStack{
				name:   extractPythonFuncName(trimmed),
				start:  i + 1,
				indent: indent,
			})
			continue
		}

		// Close functions when indentation returns to their level or above
		for len(stack) > 0 {
			top := stack[len(stack)-1]
			if indent <= top.indent {
				results = append(results, FunctionBloat{
					Name: top.name, LineCount: i - top.start + 1, LineStart: top.start,
				})
				stack = stack[:len(stack)-1]
			} else {
				break
			}
		}

		// If no active functions and this line could be a def — detect it
		// (handles cases where 'def' appears mid-line, unlikely but defensive)
		if len(stack) == 0 && (strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "async def ")) {
			stack = append(stack, funcStack{
				name:   extractPythonFuncName(trimmed),
				start:  i + 1,
				indent: indent,
			})
		}
	}

	// Flush remaining stack at EOF
	for _, f := range stack {
		results = append(results, FunctionBloat{
			Name: f.name, LineCount: len(lines) - f.start + 1, LineStart: f.start,
		})
	}
	return results
}

func runBraceDetector(lines []string, isDef func(string) bool, extractName func(string) string) []FunctionBloat {
	var results []FunctionBloat
	var inFunc bool
	var awaitingBody bool
	var funcName string
	var funcStart, braceDepth int

	reset := func() {
		inFunc = false
		awaitingBody = false
		braceDepth = 0
	}

	for i, line := range lines {
		trimmed := trimLine(line)

		if awaitingBody {
			depthChange := countOpens(trimmed) - countCloses(trimmed)
			braceDepth += depthChange
			if braceDepth > 0 {
				awaitingBody = false
				inFunc = true
			}
			// If another function definition arrives while awaiting body,
			// replace the name (handles annotations on preceding lines)
			if isDef(trimmed) {
				funcName = extractName(trimmed)
				funcStart = i + 1
				braceDepth = 0
				depthChange2 := countOpens(trimmed) - countCloses(trimmed)
				braceDepth = depthChange2
				if braceDepth > 0 {
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
				reset()
			}
			continue
		}

		if isDef(trimmed) {
			funcName = extractName(trimmed)
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

var javaMethodModifiers = []string{"public ", "private ", "protected ", "static ", "synchronized ", "final ", "abstract ", "native ", "default "}
var javaReturnTypePrefixes = []string{"void ", "int ", "long ", "double ", "float ", "boolean ", "byte ", "short ", "char ", "string "}

func isFuncDefJava(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
		return false
	}
	if !strings.Contains(trimmed, "(") {
		return false
	}
	lower := strings.ToLower(trimmed)

	// Annotations: any @ prefix is a strong signal of a method definition
	if strings.HasPrefix(trimmed, "@") {
		return true
	}

	// Modifiers as prefixes
	for _, m := range javaMethodModifiers {
		if strings.HasPrefix(lower, m) {
			return true
		}
	}

	// Return types as prefixes (after optional modifier matching above)
	for _, rt := range javaReturnTypePrefixes {
		if strings.HasPrefix(lower, rt) {
			return true
		}
	}

	// Generic return types: identifier<...> name(
	// Pattern: Word followed by < generics > and (
	if braceIdx := strings.Index(lower, "("); braceIdx != -1 {
		before := strings.TrimSpace(lower[:braceIdx])
		words := strings.Fields(before)
		if len(words) >= 2 {
			first := words[0]
			// Check if first word is NOT a control flow keyword
			controlKWs := []string{"if", "else", "for", "while", "switch", "return", "throw", "new", "try", "catch"}
			for _, kw := range controlKWs {
				if first == kw {
					return false
				}
			}
			// If the first word is a Java type (starts with uppercase or is a known type),
			// and there's at least one more word before (, it's likely a method
			if first[0] >= 'A' && first[0] <= 'Z' {
				return true
			}
		}
	}

	return false
}

func extractFuncNameJava(line string) string {
	idx := strings.Index(line, "(")
	if idx == -1 {
		return "unknown"
	}
	words := strings.Fields(strings.TrimSpace(line[:idx]))
	if len(words) == 0 {
		return "unknown"
	}
	return words[len(words)-1]
}

func isFuncDefRust(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "///") {
		return false
	}
	lower := strings.ToLower(trimmed)
	return strings.HasPrefix(lower, "fn ") || strings.HasPrefix(lower, "pub fn ") || strings.HasPrefix(lower, "pub(crate) fn ") || strings.HasPrefix(lower, "async fn ") || strings.HasPrefix(lower, "pub async fn ")
}

func extractFuncNameRust(line string) string {
	trimmed := strings.TrimSpace(line)
	lower := strings.ToLower(trimmed)
	for _, p := range []string{"pub ", "pub(crate) ", "async ", "pub async "} {
		if strings.HasPrefix(lower, p) {
			trimmed = strings.TrimSpace(trimmed[len(p):])
			lower = strings.ToLower(trimmed)
		}
	}
	trimmed = strings.TrimPrefix(trimmed, "fn ")
	idx := strings.Index(trimmed, "(")
	if idx == -1 {
		return "unknown"
	}
	return strings.TrimSpace(trimmed[:idx])
}

func extractPythonFuncName(line string) string {
	l := line
	if strings.HasPrefix(l, "async ") {
		l = l[6:]
	}
	l = strings.TrimPrefix(l, "def ")
	idx := strings.Index(l, "(")
	if idx == -1 {
		return "unknown"
	}
	return strings.TrimSpace(l[:idx])
}

// DetectFunctionBloatsTS detects functions in JavaScript and TypeScript.
// Handles: function, async function, export function, class methods, arrow => {.
// Tracks paren-depth to avoid miscounting destructuring {} inside parameter lists.
func DetectFunctionBloatsTS(lines []string) []FunctionBloat {
	var results []FunctionBloat
	var inFunc bool
	var funcName string
	var funcStart int
	var braceDepth, parenDepth int
	var bodyOpened bool

	reset := func() {
		inFunc = false
		braceDepth = 0
		parenDepth = 0
		bodyOpened = false
	}

	isTSFuncStart := func(line string) bool {
		l := strings.ToLower(strings.TrimSpace(line))
		if isComment(l) || l == "" {
			return false
		}
		// Arrow functions: const x = () => { or (a) => {
		if strings.Contains(l, "=>") && strings.Contains(l, "{") {
			before := l[:strings.LastIndex(l, "=>")]
			if !strings.Contains(before, "//") && !strings.Contains(before, "\"") {
				return true
			}
		}
		// Named functions: function foo() / async function foo()
		if strings.Contains(l, "function ") {
			return true
		}
		// Class methods/constructors: name() {, async name() {, constructor() {
		// Must NOT look like a function call: foo(); bar(x);
		hasOpeningParen := strings.Contains(l, "(")
		hasClosingParen := strings.Contains(l, ")")
		hasBrace := strings.Contains(l, "{") || strings.Contains(l, "=>")
		if hasOpeningParen && hasBrace {
			// Exclude function CALLS (name(); or foo().bar())
			if hasClosingParen {
				closeIdx := strings.LastIndex(l, ")")
				braceIdx := strings.LastIndex(l, "{")
				arrowIdx := strings.Index(l, "=>")
				if arrowIdx == -1 {
					arrowIdx = 999999
				}
				// If ) is before {, it's likely a method declaration
				if closeIdx < braceIdx && braceIdx < arrowIdx {
					beforeParen := strings.TrimSpace(l[:strings.Index(l, "(")])
					parts := strings.Fields(beforeParen)
					// Strip modifiers
					filtered := []string{}
					for _, p := range parts {
						pl := strings.ToLower(p)
						if pl == "async" || pl == "static" || pl == "public" ||
							pl == "private" || pl == "protected" || pl == "export" ||
							pl == "default" || pl == "abstract" || pl == "get" || pl == "set" {
							continue
						}
						filtered = append(filtered, p)
					}
					if len(filtered) >= 1 && len(filtered) <= 2 {
						return true
					}
				}
			}
		}
		return false
	}

	extractTSFuncName := func(line string) string {
		l := strings.TrimSpace(line)
		l = strings.TrimPrefix(l, "export ")
		l = strings.TrimPrefix(l, "default ")
		l = strings.TrimPrefix(l, "async ")
		l = strings.TrimPrefix(l, "static ")
		l = strings.TrimPrefix(l, "public ")
		l = strings.TrimPrefix(l, "private ")
		l = strings.TrimPrefix(l, "protected ")
		l = strings.TrimPrefix(l, "abstract ")
		l = strings.TrimSpace(l)

		if strings.HasPrefix(l, "function ") {
			rest := strings.TrimSpace(l[9:])
			if end := strings.IndexAny(rest, "(<"); end != -1 {
				return strings.TrimSpace(rest[:end])
			}
			return "anonymous"
		}

		if strings.HasPrefix(l, "constructor") {
			return "constructor"
		}

		// Class method or getter/setter: name(...) { or get name() {
		parenIdx := strings.Index(l, "(")
		if parenIdx > 0 {
			before := strings.TrimSpace(l[:parenIdx])
			// Handle get/set prefix
			if strings.HasPrefix(before, "get ") || strings.HasPrefix(before, "set ") {
				return strings.TrimSpace(before)
			}
			return before
		}

		if strings.Contains(l, "=>") {
			eq := strings.Index(l, "=")
			arrow := strings.Index(l, "=>")
			if eq != -1 && eq < arrow {
				before := strings.TrimSpace(l[:eq])
				words := strings.Fields(before)
				if len(words) >= 2 {
					return words[len(words)-1]
				}
			}
			return "arrow"
		}

		return "unknown"
	}

	for i, line := range lines {
		trimmed := trimLine(line)
		if inFunc {
			for _, ch := range trimmed {
				switch ch {
				case '{':
					if parenDepth == 0 {
						braceDepth++
						bodyOpened = true
					}
				case '}':
					if parenDepth == 0 {
						braceDepth--
					}
				case '(':
					parenDepth++
				case ')':
					parenDepth--
				}
			}
			if parenDepth < 0 {
				parenDepth = 0
			}
			if bodyOpened && braceDepth <= 0 && parenDepth == 0 {
				results = append(results, FunctionBloat{
					Name: funcName, LineCount: (i + 1) - funcStart + 1, LineStart: funcStart,
				})
				reset()
			}
			continue
		}

		if isTSFuncStart(trimmed) {
			inFunc = true
			funcName = extractTSFuncName(trimmed)
			funcStart = i + 1
			braceDepth = 0
			parenDepth = 0
			bodyOpened = false

			for _, ch := range trimmed {
				switch ch {
				case '{':
					if parenDepth == 0 {
						braceDepth = 1
						bodyOpened = true
					}
				case '}':
					if parenDepth == 0 {
						braceDepth--
					}
				case '(':
					parenDepth++
				case ')':
					parenDepth--
				}
			}
			if parenDepth < 0 {
				parenDepth = 0
			}
		}
	}

	if inFunc {
		results = append(results, FunctionBloat{
			Name: funcName, LineCount: len(lines) - funcStart + 1, LineStart: funcStart,
		})
	}

	return results
}
