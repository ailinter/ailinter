package metalinter

import (
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/packages"
)

// goVetAnalyzers is the set of go vet analyzers we run in-process.
// These correspond to the most useful checks from `go vet`.
var goVetAnalyzers = []*analysis.Analyzer{
	asmdecl.Analyzer,
	assign.Analyzer,
	atomic.Analyzer,
	bools.Analyzer,
	composite.Analyzer,
	copylock.Analyzer,
	deepequalerrors.Analyzer,
	errorsas.Analyzer,
	httpresponse.Analyzer,
	ifaceassert.Analyzer,
	loopclosure.Analyzer,
	lostcancel.Analyzer,
	nilfunc.Analyzer,
	nilness.Analyzer,
	printf.Analyzer,
	shift.Analyzer,
	sigchanyzer.Analyzer,
	slog.Analyzer,
	stdmethods.Analyzer,
	stringintconv.Analyzer,
	structtag.Analyzer,
	testinggoroutine.Analyzer,
	tests.Analyzer,
	timeformat.Analyzer,
	unmarshal.Analyzer,
	unreachable.Analyzer,
	unsafeptr.Analyzer,
	unusedresult.Analyzer,
}

// runGoVetStaticcheck loads Go packages and runs all embedded analyzers
// (go vet passes + staticcheck + ineffassign) on them.
func runGoVetStaticcheck(pkgPaths []string) ([]Finding, error) {
	if len(pkgPaths) == 0 {
		return nil, nil
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedTypesSizes |
			packages.NeedModule,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, pkgPaths...)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	// Build combined analyzer list from go vet + staticcheck + ineffassign
	analyzers := make([]*analysis.Analyzer, 0, len(goVetAnalyzers)+len(staticcheckAnalyzers)+1)
	analyzers = append(analyzers, goVetAnalyzers...)
	analyzers = append(analyzers, staticcheckAnalyzers...)
	analyzers = append(analyzers, ineffassignAnalyzer)

	// Resolve analyzer dependencies topologically
	ordered := resolveAnalyzers(analyzers)

	allFindings := analyzePackages(pkgs, ordered)
	return allFindings, nil
}

// analyzePackages runs all analyzers on all loaded packages.
func analyzePackages(pkgs []*packages.Package, ordered []*analysis.Analyzer) []Finding {
	var allFindings []Finding
	for _, pkg := range pkgs {
		findings := analyzeSinglePackage(pkg, ordered)
		allFindings = append(allFindings, findings...)
	}
	return allFindings
}

// analyzeSinglePackage runs analyzers on a single package, handling errors.
func analyzeSinglePackage(pkg *packages.Package, ordered []*analysis.Analyzer) []Finding {
	if len(pkg.Syntax) == 0 {
		return nil
	}
	if len(pkg.Errors) > 0 {
		return packageErrorsToFindings(pkg.Errors)
	}
	return runAnalyzersOnPackage(pkg, ordered)
}

// packageErrorsToFindings converts package load errors to findings.
func packageErrorsToFindings(errs []packages.Error) []Finding {
	findings := make([]Finding, 0, len(errs))
	for _, e := range errs {
		findings = append(findings, Finding{
			Tool:     "govet",
			Severity: SeverityError,
			File:     e.Pos,
			Line:     1,
			Column:   1,
			Message:  fmt.Sprintf("package error: %s", e.Msg),
			Category: CategoryCorrectness,
		})
	}
	return findings
}

// runAnalyzersOnPackage runs all resolved analyzers on a single loaded package.
func runAnalyzersOnPackage(pkg *packages.Package, analyzers []*analysis.Analyzer) []Finding {
	var findings []Finding
	results := make(map[*analysis.Analyzer]any)

	for _, a := range analyzers {
		findings = append(findings, runAnalyzer(a, pkg, results)...)
	}
	return findings
}

// runAnalyzer runs a single analyzer on a package and returns findings.
func runAnalyzer(a *analysis.Analyzer, pkg *packages.Package, results map[*analysis.Analyzer]any) []Finding {
	var findings []Finding

	// Collect dependency results
	resultOf := buildResultOf(a, results)

	var diags []analysis.Diagnostic
	pass := newPass(a, pkg, resultOf, &diags)

	result, runErr := a.Run(pass)
	if runErr != nil {
		return []Finding{{
			Tool:     toolNameForAnalyzer(a),
			Severity: SeverityError,
			Message:  fmt.Sprintf("analyzer %s failed: %v", a.Name, runErr),
			Category: CategoryCorrectness,
		}}
	}
	if result != nil {
		results[a] = result
	}

	for _, d := range diags {
		findings = append(findings, diagnosticToFinding(a, d, pkg))
	}
	return findings
}

// buildResultOf collects dependency results for an analyzer.
func buildResultOf(a *analysis.Analyzer, results map[*analysis.Analyzer]any) map[*analysis.Analyzer]any {
	resultOf := make(map[*analysis.Analyzer]any)
	for _, dep := range a.Requires {
		if res, ok := results[dep]; ok {
			resultOf[dep] = res
		}
	}
	return resultOf
}

// newPass creates an analysis.Pass for running an analyzer.
func newPass(a *analysis.Analyzer, pkg *packages.Package, resultOf map[*analysis.Analyzer]any, diags *[]analysis.Diagnostic) *analysis.Pass {
	return &analysis.Pass{
		Analyzer:          a,
		Fset:              pkg.Fset,
		Files:             pkg.Syntax,
		Pkg:               pkg.Types,
		TypesInfo:         pkg.TypesInfo,
		TypesSizes:        pkg.TypesSizes,
		ResultOf:          resultOf,
		Report:            func(d analysis.Diagnostic) { *diags = append(*diags, d) },
		AllObjectFacts:    func() []analysis.ObjectFact { return nil },
		AllPackageFacts:   func() []analysis.PackageFact { return nil },
		ImportObjectFact:  func(obj types.Object, fact analysis.Fact) bool { return false },
		ImportPackageFact: func(pkg2 *types.Package, fact analysis.Fact) bool { return false },
		ExportObjectFact:  func(obj types.Object, fact analysis.Fact) {},
		ExportPackageFact: func(fact analysis.Fact) {},
	}
}

// diagnosticToFinding converts an analysis.Diagnostic to a Finding.
func diagnosticToFinding(a *analysis.Analyzer, d analysis.Diagnostic, pkg *packages.Package) Finding {
	filePath, line, col := resolvePos(d.Pos, pkg)
	filePath = toRelative(filePath)
	code := d.Category
	if code == "" {
		code = a.Name
	}
	category := d.Category
	if category == "" {
		category = categorizeAnalyzer(a)
	}
	return Finding{
		Tool:     toolNameForAnalyzer(a),
		Code:     code,
		Severity: SeverityWarning,
		File:     filePath,
		Line:     line,
		Column:   col,
		Message:  d.Message,
		Category: category,
	}
}

// resolvePos extracts filename, line, and column from a token position.
func resolvePos(pos token.Pos, pkg *packages.Package) (string, int, int) {
	if pkg == nil || pkg.Fset == nil || !pos.IsValid() {
		return "", 0, 0
	}
	tokFile := pkg.Fset.File(pos)
	if tokFile == nil {
		return "", 0, 0
	}
	p := tokFile.Position(pos)
	return tokFile.Name(), p.Line, p.Column
}

// toRelative converts an absolute path to relative if possible.
func toRelative(path string) string {
	if path == "" {
		return ""
	}
	wd, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(wd, path)
	if err != nil {
		return path
	}
	return rel
}

// lineCol extracts 1-based line and column from a token.Pos.
func lineCol(tokFile *token.File, pos token.Pos) (int, int) {
	if !pos.IsValid() {
		return 0, 0
	}
	p := tokFile.Position(pos)
	return p.Line, p.Column
}

// toolPrefixes maps analyzer name prefixes to their tool name.
var toolPrefixes = []struct {
	prefix string
	tool   string
}{
	{"SA", "staticcheck"},
	{"S", "staticcheck"},
	{"ST", "staticcheck"},
	{"U", "staticcheck"},
}

// toolNameForAnalyzer maps an analyzer to its tool name.
func toolNameForAnalyzer(a *analysis.Analyzer) string {
	if a.Name == "ineffassign" {
		return "ineffassign"
	}
	for _, tp := range toolPrefixes {
		if strings.HasPrefix(a.Name, tp.prefix) {
			return tp.tool
		}
	}
	return "govet"
}

// analyzerCategory maps analyzer names to their category.
// Use prefix matching for groups (SA*, S*, ST*) and exact match for individuals.
var analyzerCategory = func() map[string]string {
	m := map[string]string{
		"ineffassign":  CategoryUnused,
		"assign":       CategoryUnused,
		"unusedresult": CategoryUnused,
		"printf":       CategoryCorrectness,
		"slog":         CategoryCorrectness,
		"stdmethods":   CategoryCorrectness,
		"copylock":     CategoryBug,
		"bools":        CategoryBug,
		"nilfunc":      CategoryBug,
		"nilness":      CategoryBug,
		"atomic":       CategoryBug,
		"loopclosure":  CategoryBug,
		"lostcancel":   CategoryBug,
		"composite":    CategoryStyle,
		"structtag":    CategoryStyle,
	}
	return m
}()

// prefixCategories maps prefix patterns for staticcheck/dynamic analyzer groups.
var prefixCategories = []struct {
	prefix   string
	category string
}{
	{"SA", CategoryCorrectness},
	{"S", CategoryStyle},
	{"ST", CategoryStyle},
	{"U", CategoryUnused},
}

// categorizeAnalyzer returns the category for an analyzer based on its name.
func categorizeAnalyzer(a *analysis.Analyzer) string {
	if cat, ok := analyzerCategory[a.Name]; ok {
		return cat
	}
	for _, pc := range prefixCategories {
		if strings.HasPrefix(a.Name, pc.prefix) {
			return pc.category
		}
	}
	return CategoryCorrectness
}

// resolveAnalyzers topologically sorts analyzers by their dependencies.
func resolveAnalyzers(analyzers []*analysis.Analyzer) []*analysis.Analyzer {
	visited := make(map[*analysis.Analyzer]bool)
	var order []*analysis.Analyzer

	var visit func(a *analysis.Analyzer)
	visit = func(a *analysis.Analyzer) {
		if visited[a] {
			return
		}
		visited[a] = true
		for _, dep := range a.Requires {
			visit(dep)
		}
		order = append(order, a)
	}

	for _, a := range analyzers {
		visit(a)
	}

	return order
}
