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

func DefaultThresholds(lang string) Thresholds {
	return parser.DefaultThresholds(lang)
}

func DetectedLanguage(ext string) string {
	return parser.DetectedLanguage(ext)
}

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
		data, err := os.ReadFile(repoPath + "/" + filepath.Clean(e.FilePath))
		if err != nil {
			continue
		}

		ext := filepath.Ext(e.FilePath)
		lang := parser.DetectedLanguage(ext)
		if lang == "" {
			continue
		}

		thresholds := config.LoadProjectThresholds(repoPath, lang)
		qr := Analyze(SourceInput{FilePath: e.FilePath, Source: string(data), Lang: lang}, thresholds)
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

// SourceInput bundles the source file identity for analysis.
type SourceInput struct {
	FilePath string
	Source   string
	Lang     string
}

// Analyze runs all detectors on source and returns a QualityResult.
// Score is 0-100.
func Analyze(src SourceInput, t Thresholds) QualityResult {
	lines := splitLines(src.Source)
	loc := len(lines)
	var smells []Smell
	addSmell(&smells, parser.DetectNestingSmell(lines, t.NestingWarning, t.NestingAlert))
	addSmell(&smells, parser.DetectFileBloat(loc, t.FileLOCWarning, t.FileLOCAlert, t.FileLOCCritical))
	bloats := detectFunctions(src.Lang, lines)
	smells = append(smells, detectFunctionSmells(lines, bloats, t)...)
	smells = append(smells, detectCodeLevelSmells(lines, bloats, t)...)

	dupPairs := parser.DetectDuplications(bloats, lines, t.DupMinLines, t.DupMinSimilarity)
	smells = append(smells, parser.DetectDuplicationSmells(dupPairs)...)
	cohesion := parser.AnalyzeCohesion(bloats, lines)
	addSmell(&smells, parser.DetectLowCohesionSmell(cohesion, 50, 75))

	raw := computeScore(scoreParams{smells: smells, loc: loc, fnCount: countReal(bloats), coh: cohesion, dups: dupPairs, t: t})
	s := int(math.Round(raw * 10.0))
	if s < 0 {
		s = 0
	} else if s > 100 {
		s = 100
	}
	return QualityResult{Score: s, Label: classifyLabel(s), Smells: smells, FilePath: src.FilePath, Language: src.Lang, LinesOfCode: loc}
}
func detectFunctionSmells(lines []string, bloats []FunctionBloat, t Thresholds) []Smell {
	var smells []Smell
	for _, s := range parser.DetectBrainMethodSmell(bloats, t.FuncLOCWarning, t.FuncLOCAlert) {
		smells = append(smells, s)
	}

	for _, fn := range bloats {
		funcLines := sliceFuncBody(lines, fn)
		if fn.LineCount >= 10 {
			addSmell(&smells, parser.DetectBumpyRoadSmellAt(funcLines, t.BumpyRoadNestingDepth, t.BumpyRoadBumpsWarning, fn.LineStart))
		}

		cc := cyclomaticComplexity(funcLines)
		if cc < t.FuncCCWarning {
			continue
		}
		sev := "warning"
		if cc >= t.FuncCCAlert {
			sev = "alert"
		}
		smells = append(smells, Smell{
			Name: "complex_method", LineStart: fn.LineStart,
			Severity: sev,
			Message:  fmt.Sprintf("%s CC=%d", fn.Name, cc),
			AIPrompt: fmt.Sprintf("Complex method '%s' (CC=%d). Simplify.", fn.Name, cc),
		})
	}
	return smells
}

func detectCodeLevelSmells(lines []string, bloats []FunctionBloat, t Thresholds) []Smell {
	smells := []Smell{}
	smells = append(smells, parser.DetectComplexConditional(lines, t.ComplexCondBranchesWarn, t.ComplexCondBranchesAlert)...)
	smells = append(smells, parser.DetectLongParameterList(lines, t.MaxArgumentsWarn, t.MaxArgumentsAlert)...)
	fnCount := countReal(bloats)
	addSmell(&smells, functionCountSmell(fnCount, t))
	addSmell(&smells, brainClassSmell(len(lines), fnCount, t))
	smells = append(smells, parser.DetectMessageChains(lines)...)
	smells = append(smells, parser.DetectParagraphOfCode(lines, t.ParagraphMaxConsecutive)...)
	smells = append(smells, parser.DetectLazyElements(bloats, t.LazyMinLines)...)
	smells = append(smells, parser.DetectPrimitiveObsession(lines)...)
	smells = append(smells, parser.DetectLongScopeVariables(lines, bloats, t.LongScopeVarLines)...)
	smells = append(smells, parser.DetectLongSwitch(lines, t.LongSwitchWarn, t.LongSwitchAlert)...)
	addSmell(&smells, parser.DetectExcessiveComments(lines, t.CommentRatioWarning))
	addSmell(&smells, parser.DetectGlobalData(lines, t.GlobalDataWarning))
	return smells
}

type langDetector func(lines []string) []FunctionBloat

var functionDetectors = map[string]langDetector{
	"python":     funcDetectIndent,
	"javascript": parser.DetectFunctionBloatsTS,
	"typescript": parser.DetectFunctionBloatsTS,
	"cpp":        parser.DetectFunctionBloatsCPP,
	"c":          parser.DetectFunctionBloatsCPP,
	"c++":        parser.DetectFunctionBloatsCPP,
	"cc":         parser.DetectFunctionBloatsCPP,
	"java":       parser.DetectFunctionBloatsJava,
	"csharp":     parser.DetectFunctionBloatsJava,
	"cs":         parser.DetectFunctionBloatsJava,
	"rust":       parser.DetectFunctionBloatsRust,
	"ruby":       parser.DetectFunctionBloatsRuby,
	"swift":      parser.DetectFunctionBloatsSwift,
	"kotlin":     parser.DetectFunctionBloatsKotlin,
}

func funcDetectIndent(lines []string) []FunctionBloat {
	return parser.DetectFunctionBloatsIndent(lines, parser.DetectIndentSize(lines))
}

func detectFunctions(lang string, lines []string) []FunctionBloat {
	if fn, ok := functionDetectors[lang]; ok {
		return fn(lines)
	}
	return parser.DetectFunctionBloats(lines)
}

type scoreParams struct {
	smells  []Smell
	loc     int
	fnCount int
	coh     CohesionResult
	dups    []DuplicationPair
	t       Thresholds
}

func computeScore(p scoreParams) float64 {
	base := 10.0
	base -= locPenalty(p.loc, p.t.FileLOCWarning)
	base -= fnCountPenalty(p.fnCount, p.t.FuncCountWarning)
	base -= smellPenalty(p.smells)
	base -= dupPenalty(p.dups)
	base -= cohesionPenalty(p.coh)
	return clampBase(base, 1.0)
}

func locPenalty(loc, fileLOCWarning int) float64 {
	if loc <= fileLOCWarning {
		return 0
	}
	ratio := float64(loc) / float64(fileLOCWarning)
	p := math.Log2(ratio) * 1.5
	if p > 3.0 {
		p = 3.0
	}
	return p
}

func fnCountPenalty(fnCount, funcCountWarning int) float64 {
	if fnCount <= funcCountWarning {
		return 0
	}
	ratio := float64(fnCount) / float64(funcCountWarning)
	p := math.Log2(ratio) * 0.8
	if p > 1.5 {
		p = 1.5
	}
	return p
}

func smellPenalty(smells []Smell) float64 {
	p := 0.0
	for _, s := range smells {
		switch s.Severity {
		case "critical":
			p += 0.5
		case "alert":
			p += 0.25
		case "warning":
			p += 0.1
		}
	}
	if p > 3.0 {
		p = 3.0
	}
	return p
}

func dupPenalty(dups []DuplicationPair) float64 {
	if len(dups) == 0 {
		return 0
	}
	p := float64(len(dups)) * 0.15
	if p > 2.0 {
		p = 2.0
	}
	return p
}

func cohesionPenalty(coh CohesionResult) float64 {
	if !coh.IsLowCohesion || coh.TotalFuncs < 5 {
		return 0
	}
	p := (1.0 - coh.CohesionScore) * 2.0
	if p > 1.5 {
		p = 1.5
	}
	return p
}

func clampBase(v, min float64) float64 {
	if v < min {
		return min
	}
	return v
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

func isRealFunction(b FunctionBloat) bool {
	return b.Name != "unknown" && b.Name != "anonymous" && b.LineCount >= 3
}

func countReal(bloats []FunctionBloat) int {
	n := 0
	for _, b := range bloats {
		if isRealFunction(b) {
			n++
		}
	}
	return n
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
