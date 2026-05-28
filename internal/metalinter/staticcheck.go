package metalinter

import (
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"honnef.co/go/tools/unused"

	"golang.org/x/tools/go/analysis"
)

// staticcheckAnalyzers collects the most useful staticcheck analyzers for Phase 1.
// We select SA* (staticcheck), S* (simple), ST* (stylecheck), and U1000 (unused).
//
// Staticcheck v0.7.0 provides analyzers as honnef.co/go/tools/analysis/lint.Analyzer wrappers.
// We extract the underlying *analysis.Analyzer from each.
func getStaticcheckAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer

	// Staticcheck SA* analyzers (correctness)
	for _, l := range staticcheck.Analyzers {
		analyzers = append(analyzers, l.Analyzer)
	}

	// Simple S* analyzers (style simplifications)
	for _, l := range simple.Analyzers {
		analyzers = append(analyzers, l.Analyzer)
	}

	// Stylecheck ST* analyzers (style)
	for _, l := range stylecheck.Analyzers {
		analyzers = append(analyzers, l.Analyzer)
	}

	// Unused U1000 (unused code)
	if unused.Analyzer != nil {
		analyzers = append(analyzers, unused.Analyzer.Analyzer)
	}

	return analyzers
}

// staticcheckAnalyzers is the initialized list.
var staticcheckAnalyzers = getStaticcheckAnalyzers()
