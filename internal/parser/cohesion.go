package parser

import (
	"fmt"
	"regexp"
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

var jsPatterns = struct {
	functionDecl *regexp.Regexp
	classDecl    *regexp.Regexp
	varDecl      *regexp.Regexp
	thisProp     *regexp.Regexp
}{
	functionDecl: regexp.MustCompile(`^\s*(?:export\s+)?(?:async\s+)?function\s+([A-Za-z_$][\w$]*)`),
	classDecl:    regexp.MustCompile(`^\s*(?:export\s+)?(?:default\s+)?class\s+([A-Za-z_$][\w$]*)`),
	varDecl:      regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$][\w$]*)`),
	thisProp:     regexp.MustCompile(`\bthis\.([A-Za-z_$][\w$]*)`),
}

var pythonPatterns = struct {
	defDecl   *regexp.Regexp
	classDecl *regexp.Regexp
	assign    *regexp.Regexp
	selfProp  *regexp.Regexp
}{
	defDecl:   regexp.MustCompile(`^\s*(?:async\s+)?def\s+([A-Za-z_]\w*)`),
	classDecl: regexp.MustCompile(`^\s*class\s+([A-Za-z_]\w*)`),
	assign:    regexp.MustCompile(`^([A-Za-z_]\w*)\s*=`),
	selfProp:  regexp.MustCompile(`\bself\.([A-Za-z_]\w*)`),
}

var rubyPatterns = struct {
	defDecl   *regexp.Regexp
	classDecl *regexp.Regexp
	constDecl *regexp.Regexp
	ivar      *regexp.Regexp
}{
	defDecl:   regexp.MustCompile(`^\s*def\s+(?:self\.)?([A-Za-z_]\w*[!?=]?)`),
	classDecl: regexp.MustCompile(`^\s*(?:class|module)\s+([A-Z]\w*)`),
	constDecl: regexp.MustCompile(`^\s*([A-Z]\w*)\s*=`),
	ivar:      regexp.MustCompile(`@([A-Za-z_]\w*)`),
}

type symbolRules struct {
	moduleDecls []*regexp.Regexp
	fieldRefs   *regexp.Regexp
	fieldPrefix string
}

var symbolRulesByLanguage = map[string]symbolRules{
	"javascript": {
		moduleDecls: []*regexp.Regexp{jsPatterns.functionDecl, jsPatterns.classDecl, jsPatterns.varDecl},
		fieldRefs:   jsPatterns.thisProp,
		fieldPrefix: "this.",
	},
	"typescript": {
		moduleDecls: []*regexp.Regexp{jsPatterns.functionDecl, jsPatterns.classDecl, jsPatterns.varDecl},
		fieldRefs:   jsPatterns.thisProp,
		fieldPrefix: "this.",
	},
	"python": {
		moduleDecls: []*regexp.Regexp{pythonPatterns.defDecl, pythonPatterns.classDecl, pythonPatterns.assign},
		fieldRefs:   pythonPatterns.selfProp,
		fieldPrefix: "self.",
	},
	"ruby": {
		moduleDecls: []*regexp.Regexp{rubyPatterns.defDecl, rubyPatterns.classDecl, rubyPatterns.constDecl},
		fieldRefs:   rubyPatterns.ivar,
		fieldPrefix: "@",
	},
}

var identifierPattern = regexp.MustCompile(`[A-Za-z_$][\w$]*`)

var typeKeywords = map[string]bool{
	"func": true, "function": true, "def": true, "var": true, "const": true,
	"return": true, "if": true, "for": true, "while": true, "switch": true,
	"case": true, "break": true, "continue": true, "else": true, "nil": true,
	"null": true, "true": true, "false": true, "error": true, "err": true,
	"int": true, "string": true, "float64": true, "float32": true, "bool": true,
	"byte": true, "void": true, "auto": true, "self": true, "this": true,
	"new": true, "make": true, "len": true, "append": true, "range": true,
	"defer": true, "go": true,
}

// cohesionSets holds one symbol set per function. A function is isolated when its set does not
// intersect any other set.
type cohesionSets []map[string]bool

// AnalyzeCohesion estimates cohesion from shared type names in function signatures.
// It is the right signal for typed languages; use AnalyzeCohesionLang for untyped ones.
func AnalyzeCohesion(bloats []FunctionBloat, lines []string) CohesionResult {
	validFuncs := validFunctions(bloats)
	if len(validFuncs) <= 3 {
		return CohesionResult{TotalFuncs: len(validFuncs), CohesionScore: 1.0}
	}

	sets := make(cohesionSets, len(validFuncs))
	for i, fn := range validFuncs {
		sets[i] = extractTypeReferences(functionBody(fn, lines, 10))
	}

	isolated, total := countIsolated(sets)
	return cohesionFromCounts(len(validFuncs), isolated, total)
}

// AnalyzeCohesionLang picks the cohesion signal that fits the language.
//
// Typed languages are measured by shared types in signatures. JavaScript, TypeScript, Python and
// Ruby have no reliable signature types, so that metric marks every function isolated and caps the
// score at 85 for any module with five or more functions. For them, cohesion is measured by shared
// module symbols instead: top-level declarations a function references, plus the fields its methods
// share (`this.*`, `self.*`, or `@ivar`). A module of unrelated helpers still scores low; a class
// or a module built on its own helpers scores high.
func AnalyzeCohesionLang(lang string, bloats []FunctionBloat, lines []string) CohesionResult {
	if rules, ok := symbolRulesByLanguage[lang]; ok {
		return analyzeSymbolCohesion(bloats, lines, rules)
	}
	return AnalyzeCohesion(bloats, lines)
}

// analyzeSymbolCohesion measures how many functions share a symbol with another function.
func analyzeSymbolCohesion(bloats []FunctionBloat, lines []string, rules symbolRules) CohesionResult {
	validFuncs := validFunctions(bloats)
	if len(validFuncs) <= 3 {
		return CohesionResult{TotalFuncs: len(validFuncs), CohesionScore: 1.0}
	}

	moduleNames := moduleLevelNames(validFuncs, lines, rules)
	sets := make(cohesionSets, len(validFuncs))
	for i, fn := range validFuncs {
		sets[i] = functionSymbols(fn, lines, moduleNames, rules)
	}

	return cohesionFromCounts(len(validFuncs), countIsolatedSets(sets), 0)
}

func countIsolated(sets cohesionSets) (int, int) {
	isolated, total := 0, 0
	for i, types := range sets {
		if len(types) == 0 || !sets.shares(types, i) {
			isolated++
		}
		total += len(types)
	}
	return isolated, total
}

func countIsolatedSets(sets cohesionSets) int {
	isolated := 0
	for i := range sets {
		if !sets.shares(sets[i], i) {
			isolated++
		}
	}
	return isolated
}

func (sets cohesionSets) shares(set map[string]bool, at int) bool {
	for other, otherSet := range sets {
		if other == at {
			continue
		}
		if intersects(set, otherSet) {
			return true
		}
	}
	return false
}

// cohesionFromCounts is the single place that turns counted functions into a result.
func cohesionFromCounts(functions, isolated, sharedTypes int) CohesionResult {
	score := 1.0
	if functions > 0 {
		score = float64(functions-isolated) / float64(functions)
	}
	return CohesionResult{
		TotalFuncs:      functions,
		SharedTypeCount: sharedTypes,
		IsolatedFuncs:   isolated,
		CohesionScore:   score,
		IsLowCohesion:   lowCohesion(score, functions),
	}
}

func lowCohesion(score float64, functions int) bool {
	if functions < 5 {
		return false
	}
	return score < 0.5
}

// validFunctions keeps the functions that carry enough signal to judge.
func validFunctions(bloats []FunctionBloat) []FunctionBloat {
	valid := make([]FunctionBloat, 0, len(bloats))
	for _, b := range bloats {
		if isValidFunction(b) {
			valid = append(valid, b)
		}
	}
	return valid
}

func isValidFunction(b FunctionBloat) bool {
	if b.LineCount < 2 {
		return false
	}
	if b.Name == "unknown" {
		return false
	}
	return b.Name != "anonymous"
}

// functionBody returns the first maxLines of a function, joined for scanning.
func functionBody(fn FunctionBloat, lines []string, maxLines int) string {
	start := max(0, fn.LineStart-1)
	end := min(len(lines), start+fn.LineCount)
	if maxLines > 0 && end > start+maxLines {
		end = start + maxLines
	}
	return strings.Join(lines[start:end], " ")
}

// moduleLevelNames finds declarations outside every function: top-level functions, classes and
// constants. These are the symbols a function can share with a sibling.
func moduleLevelNames(bloats []FunctionBloat, lines []string, rules symbolRules) map[string]bool {
	inside := insideFunctionLines(bloats, len(lines))
	names := make(map[string]bool)
	addFunctionNames(names, bloats)
	for i, line := range lines {
		if inside[i] {
			continue
		}
		addDeclaredName(names, line, rules.moduleDecls)
	}
	return names
}

func addFunctionNames(names map[string]bool, bloats []FunctionBloat) {
	for _, fn := range bloats {
		if isNamedFunction(fn.Name) {
			names[fn.Name] = true
		}
	}
}

func insideFunctionLines(bloats []FunctionBloat, total int) []bool {
	inside := make([]bool, total)
	for _, fn := range bloats {
		markFunctionRange(inside, fn)
	}
	return inside
}

func markFunctionRange(inside []bool, fn FunctionBloat) {
	start := max(0, fn.LineStart-1)
	end := min(len(inside), start+fn.LineCount)
	for i := start; i < end; i++ {
		inside[i] = true
	}
}

func addDeclaredName(names map[string]bool, line string, patterns []*regexp.Regexp) {
	for _, pattern := range patterns {
		if match := pattern.FindStringSubmatch(line); match != nil {
			names[match[1]] = true
		}
	}
}

func isNamedFunction(name string) bool {
	return name != "" && name != "unknown" && name != "anonymous" && name != "arrow"
}

// functionSymbols collects the module declarations a function references and the fields it reads
// or writes through the language's instance-state syntax.
func functionSymbols(fn FunctionBloat, lines []string, moduleNames map[string]bool, rules symbolRules) map[string]bool {
	symbols := make(map[string]bool)
	if isNamedFunction(fn.Name) {
		symbols[fn.Name] = true
		symbols[rules.fieldPrefix+fn.Name] = true
	}
	body := functionBody(fn, lines, 0)
	for _, match := range rules.fieldRefs.FindAllStringSubmatch(body, -1) {
		symbols[rules.fieldPrefix+match[1]] = true
	}
	addModuleReferences(symbols, body, moduleNames)
	return symbols
}

func addModuleReferences(symbols map[string]bool, body string, moduleNames map[string]bool) {
	for _, word := range identifierPattern.FindAllString(body, -1) {
		if moduleNames[word] {
			symbols[word] = true
		}
	}
}

func intersects(first, second map[string]bool) bool {
	for symbol := range first {
		if second[symbol] {
			return true
		}
	}
	return false
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
		AIPrompt: fmt.Sprintf("WARNING: Low cohesion detected — %d of %d functions don't share types or module symbols with others. Consider extracting unrelated functions into separate modules (SRP).",
			result.IsolatedFuncs, result.TotalFuncs),
	}
}

// extractTypeReferences finds type-like identifiers in function signatures.
func extractTypeReferences(text string) map[string]bool {
	types := make(map[string]bool)
	for _, word := range strings.Fields(text) {
		if name, ok := typeName(word); ok {
			types[name] = true
		}
	}
	return types
}

// typeName returns a type-like word after stripping punctuation and keywords.
func typeName(word string) (string, bool) {
	word = strings.Trim(word, "()[]{}*&,.;:\"'")
	if len(word) < 3 {
		return "", false
	}
	if typeKeywords[strings.ToLower(word)] {
		return "", false
	}
	if word[0] < 'A' || word[0] > 'Z' {
		return "", false
	}
	return word, true
}
