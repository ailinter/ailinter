package analyzer

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/ailinter/ailinter/internal/config"
	"github.com/ailinter/ailinter/internal/parser"
)

// Re-export parser types for backward compatibility.
type (
	Smell            = parser.Smell
	QualityResult    = parser.QualityResult
	Thresholds       = parser.Thresholds
	FunctionBloat    = parser.FunctionBloat
	DuplicationPair  = parser.DuplicationPair
	CohesionResult   = parser.CohesionResult
	GitHotspotResult = parser.GitHotspotResult
	HotspotEntry     = parser.HotspotEntry
)

var (
	DefaultThresholds = parser.DefaultThresholds
	DetectedLanguage  = parser.DetectedLanguage
)

func AnalyzeGitHotspots(repoPath string, maxDepth int) GitHotspotResult {
	result := parser.AnalyzeGitHotspots(repoPath, maxDepth)
	if result.Error != "" {
		return result
	}
	if len(result.Entries) == 0 {
		return result
	}

	sort.Slice(result.Entries, func(i, j int) bool {
		return result.Entries[i].CommitCount > result.Entries[j].CommitCount
	})

	analyzeLimit := 50
	if len(result.Entries) < analyzeLimit {
		analyzeLimit = len(result.Entries)
	}

	for i := 0; i < analyzeLimit; i++ {
		e := &result.Entries[i]
		absPath := filepath.Join(repoPath, e.FilePath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		ext := filepath.Ext(e.FilePath)
		lang := parser.DetectedLanguage(ext)
		if lang == "" {
			continue
		}

		thresholds := config.LoadProjectThresholds(absPath, lang)
		qr := Analyze(e.FilePath, string(data), lang, thresholds)
		e.QualityScore = float64(qr.Score)
		e.Priority = float64(e.CommitCount) * (100.0 - e.QualityScore + 1)
	}

	sort.Slice(result.Entries, func(i, j int) bool {
		return result.Entries[i].Priority > result.Entries[j].Priority
	})

	return result
}

const (
	LabelGoAhead         = parser.LabelGoAhead
	LabelProceedWithCare = parser.LabelProceedWithCare
	LabelNeedsWork       = parser.LabelNeedsWork
	LabelStopRefactor    = parser.LabelStopRefactor
)

// Analyze runs all detectors on source and returns a QualityResult.
// Score is 0-100.
func Analyze(filePath, source, lang string, t Thresholds) QualityResult {
	lines := splitLines(source)
	loc := len(lines)
	smells := []Smell{}

	addSmell(&smells, parser.DetectNestingSmell(lines, t.NestingWarning, t.NestingAlert))
	addSmell(&smells, parser.DetectFileBloat(loc, t.FileLOCWarning, t.FileLOCAlert, t.FileLOCCritical))

	bloats := detectFunctions(lang, lines)
	for _, s := range parser.DetectBrainMethodSmell(bloats, t.FuncLOCWarning, t.FuncLOCAlert) {
		smells = append(smells, s)
	}

	for _, fn := range bloats {
		funcLines := sliceFuncBody(lines, fn)
		if fn.LineCount >= 10 {
			addSmell(&smells, parser.DetectBumpyRoadSmellAt(funcLines, t.BumpyRoadNestingDepth, t.BumpyRoadBumpsWarning, fn.LineStart))
		}
		cc := cyclomaticComplexity(funcLines)
		if cc >= t.FuncCCWarning {
			smells = append(smells, complexMethodSmell(fn.Name, cc, fn.LineStart, t))
		}
	}

	smells = append(smells, parser.DetectComplexConditional(lines, t.ComplexCondBranchesWarn, t.ComplexCondBranchesAlert)...)
	smells = append(smells, parser.DetectLongParameterList(lines, t.MaxArgumentsWarn, t.MaxArgumentsAlert)...)

	fnCount := countReal(bloats)
	addSmell(&smells, functionCountSmell(fnCount, t))
	addSmell(&smells, brainClassSmell(loc, fnCount, t))

	smells = append(smells, parser.DetectMessageChains(lines)...)
	smells = append(smells, parser.DetectParagraphOfCode(lines, t.ParagraphMaxConsecutive)...)
	smells = append(smells, parser.DetectLazyElements(bloats, t.LazyMinLines)...)
	smells = append(smells, parser.DetectPrimitiveObsession(lines)...)

	dupPairs := parser.DetectDuplications(bloats, lines, t.DupMinLines, t.DupMinSimilarity)
	smells = append(smells, parser.DetectDuplicationSmells(dupPairs)...)

	cohesion := parser.AnalyzeCohesion(bloats, lines)
	addSmell(&smells, parser.DetectLowCohesionSmell(cohesion, 50, 75))

	addSmell(&smells, parser.DetectExcessiveComments(lines, t.CommentRatioWarning))
	addSmell(&smells, parser.DetectGlobalData(lines, t.GlobalDataWarning))
	smells = append(smells, parser.DetectLongScopeVariables(lines, bloats, t.LongScopeVarLines)...)
	smells = append(smells, parser.DetectLongSwitch(lines, t.LongSwitchWarn, t.LongSwitchAlert)...)

	raw := computeScore(smells, loc, fnCount, cohesion, dupPairs, t)
	score := int(math.Round(raw * 10.0))
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	label := classifyLabel(score)

	return QualityResult{
		Score: score, Label: label, Smells: smells,
		FilePath: filePath, Language: lang, LinesOfCode: loc,
	}
}

func detectFunctions(lang string, lines []string) []FunctionBloat {
	switch lang {
	case "python":
		return parser.DetectFunctionBloatsIndent(lines, parser.DetectIndentSize(lines))
	case "javascript", "typescript":
		return parser.DetectFunctionBloatsTS(lines)
	case "cpp", "c", "c++", "cc":
		return parser.DetectFunctionBloatsCPP(lines)
	case "java", "csharp", "cs":
		return parser.DetectFunctionBloatsJava(lines)
	case "rust":
		return parser.DetectFunctionBloatsRust(lines)
	case "ruby":
		return parser.DetectFunctionBloatsRuby(lines)
	case "swift":
		return parser.DetectFunctionBloatsSwift(lines)
	case "kotlin":
		return parser.DetectFunctionBloatsKotlin(lines)
	default:
		return parser.DetectFunctionBloats(lines)
	}
}

func computeScore(smells []Smell, loc, fnCount int, coh CohesionResult, dups []DuplicationPair, t Thresholds) float64 {
	base := 10.0

	// LOC penalty — logarithmic, capped at 3.0
	if loc > t.FileLOCWarning {
		ratio := float64(loc) / float64(t.FileLOCWarning)
		locPenalty := math.Log2(ratio) * 1.5
		if locPenalty > 3.0 {
			locPenalty = 3.0
		}
		base -= locPenalty
	}

	// Function count penalty — capped at 1.5
	if fnCount > t.FuncCountWarning {
		ratio := float64(fnCount) / float64(t.FuncCountWarning)
		fnPenalty := math.Log2(ratio) * 0.8
		if fnPenalty > 1.5 {
			fnPenalty = 1.5
		}
		base -= fnPenalty
	}

	// Per-smell penalties — capped at 3.0 total
	smellPenalty := 0.0
	for _, s := range smells {
		switch s.Severity {
		case "critical":
			smellPenalty += 0.5
		case "alert":
			smellPenalty += 0.25
		case "warning":
			smellPenalty += 0.1
		}
	}
	if smellPenalty > 3.0 {
		smellPenalty = 3.0
	}
	base -= smellPenalty

	// Duplication — 0.15 per pair, capped at 2.0
	if len(dups) > 0 {
		dupPenalty := float64(len(dups)) * 0.15
		if dupPenalty > 2.0 {
			dupPenalty = 2.0
		}
		base -= dupPenalty
	}

	// Cohesion — capped at 1.5
	if coh.IsLowCohesion && coh.TotalFuncs >= 5 {
		cohPenalty := (1.0 - coh.CohesionScore) * 2.0
		if cohPenalty > 1.5 {
			cohPenalty = 1.5
		}
		base -= cohPenalty
	}

	if base < 1.0 {
		base = 1.0
	}
	return base
}

func addSmell(smells *[]Smell, s *Smell) {
	if s != nil {
		*smells = append(*smells, *s)
	}
}

func splitLines(source string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			lines = append(lines, source[start:i])
			start = i + 1
		}
	}
	if start < len(source) {
		lines = append(lines, source[start:])
	}
	return lines
}

func sliceFuncBody(lines []string, fn FunctionBloat) []string {
	end := fn.LineStart + fn.LineCount - 1
	if end > len(lines) {
		end = len(lines)
	}
	start := fn.LineStart - 1
	if start < 0 {
		start = 0
	}
	return lines[start:end]
}

func cyclomaticComplexity(lines []string) int {
	counts := parser.CountBranches(lines)
	cc := 1
	for _, c := range counts {
		cc += c.Branches
	}
	return cc
}

func countReal(bloats []FunctionBloat) int {
	n := 0
	for _, b := range bloats {
		if b.Name != "unknown" && b.Name != "anonymous" && b.LineCount >= 3 {
			n++
		}
	}
	return n
}

func complexMethodSmell(name string, cc, line int, t Thresholds) Smell {
	s := "warning"
	if cc >= t.FuncCCAlert {
		s = "alert"
	}
	return Smell{
		Name: "complex_method", Severity: s, LineStart: line,
		Message:  fmt.Sprintf("%s CC=%d", name, cc),
		AIPrompt: fmt.Sprintf("Complex method '%s' (CC=%d). Simplify.", name, cc),
	}
}

func functionCountSmell(n int, t Thresholds) *Smell {
	if n < t.FuncCountWarning {
		return nil
	}
	return &Smell{
		Name: "function_count", Severity: "warning",
		Message:  fmt.Sprintf("%d functions", n),
		AIPrompt: "Too many functions. Consider splitting (SRP).",
	}
}

func brainClassSmell(loc, fnCount int, t Thresholds) *Smell {
	if loc < t.FileLOCWarning || fnCount < t.BrainClassMinFunc {
		return nil
	}
	return &Smell{
		Name: "brain_class", Severity: "warning",
		Message:  fmt.Sprintf("Brain Class: %d lines, %d funcs", loc, fnCount),
		AIPrompt: "Brain Class. Split into smaller, cohesive modules.",
	}
}

func classifyLabel(score int) string {
	return parser.Classify(score)
}
