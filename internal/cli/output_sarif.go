package cli

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/metalinter"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/version"
	"github.com/ailinter/ailinter/internal/vulnerability"
)

// ---------------------------------------------------------------------------
// SARIF v2.1.0 types — see https://docs.oasis-open.org/sarif/sarif/v2.1.0/
// ---------------------------------------------------------------------------

// SARIFLog is the top-level SARIF log file.
type SARIFLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SARIFRun `json:"runs"`
}

// SARIFRun represents a single analysis run.
type SARIFRun struct {
	Tool       SARIFTool     `json:"tool"`
	Results    []SARIFResult `json:"results"`
	ColumnKind string        `json:"columnKind,omitempty"`
}

// SARIFTool describes the analysis tool.
type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

// SARIFDriver describes the tool driver (AILINTER itself).
type SARIFDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationURI string      `json:"informationUri"`
	Rules          []SARIFRule `json:"rules"`
}

// SARIFRule describes a single detection rule / check.
type SARIFRule struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	ShortDescription SARIFMessage    `json:"shortDescription"`
	HelpURI          string          `json:"helpUri,omitempty"`
	Properties       SARIFProperties `json:"properties,omitempty"`
}

// SARIFProperties holds optional metadata on a rule.
type SARIFProperties struct {
	Category string `json:"category,omitempty"`
}

// SARIFResult is a single finding in a SARIF run.
type SARIFResult struct {
	RuleID    string          `json:"ruleId"`
	RuleIndex int             `json:"ruleIndex"`
	Level     string          `json:"level"` // "error", "warning", "note"
	Message   SARIFMessage    `json:"message"`
	Locations []SARIFLocation `json:"locations,omitempty"`
}

// SARIFMessage is a human-readable message.
type SARIFMessage struct {
	Text string `json:"text"`
}

// SARIFLocation associates a finding with a source location.
type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

// SARIFPhysicalLocation points to a specific file + region.
type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region,omitempty"`
}

// SARIFArtifactLocation identifies a file by URI.
type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

// SARIFRegion identifies a specific line/column range.
type SARIFRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn,omitempty"`
}

// ---------------------------------------------------------------------------
// Internal intermediate representation used to build SARIF output.
// ---------------------------------------------------------------------------

type sarifEntry struct {
	ruleID   string
	severity string
	filePath string
	line     int
	column   int
	message  string
	category string
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// WriteSARIFCombined collects quality, secret, vulnerability, and meta-lint
// findings and writes them as a single SARIF v2.1.0 log to w.
//
// qualityResults carry their own FilePath. Secrets and vulns use the provided
// filePath (customarily the resolved file path for single-file mode, or a
// directory path for directory mode). // gitleaks:allow
func WriteSARIFCombined(w io.Writer, qualityResults []analyzer.QualityResult, secretFindings []secrets.SecretFinding, vulnFindings []vulnerability.Finding, mlFindings []metalinter.Finding, scanPath string) error {
	entries := appendQualityEntries(nil, qualityResults)
	entries = appendSecretEntries(entries, secretFindings, scanPath)
	entries = appendVulnEntries(entries, vulnFindings, scanPath)
	entries = appendMetaLintEntries(entries, mlFindings)
	return writeSARIFLog(w, entries)
}

func appendQualityEntries(entries []sarifEntry, results []analyzer.QualityResult) []sarifEntry {
	for _, r := range results {
		for _, s := range r.Smells {
			if s.LineStart <= 0 {
				continue
			}
			entries = append(entries, sarifEntry{
				ruleID:   s.Name,
				severity: s.Severity,
				filePath: r.FilePath,
				line:     s.LineStart,
				column:   1,
				message:  s.Message,
				category: "quality",
			})
		}
	}
	return entries
}

func appendSecretEntries(entries []sarifEntry, findings []secrets.SecretFinding, scanPath string) []sarifEntry {
	for _, f := range findings {
		col := f.Column
		if col < 1 {
			col = 1
		}
		line := f.Line
		if line < 1 {
			line = 1
		}
		msg := f.Description
		if msg == "" {
			msg = f.Message
		}
		entries = append(entries, sarifEntry{
			ruleID:   f.RuleID,
			severity: f.Severity,
			filePath: scanPath,
			line:     line,
			column:   col,
			message:  msg,
			category: "secret",
		})
	}
	return entries
}

func appendVulnEntries(entries []sarifEntry, findings []vulnerability.Finding, scanPath string) []sarifEntry {
	for _, f := range findings {
		col := f.Column
		if col < 1 {
			col = 1
		}
		line := f.Line
		if line < 1 {
			line = 1
		}
		msg := f.Description
		if msg == "" {
			msg = f.Category
		}
		entries = append(entries, sarifEntry{
			ruleID:   f.RuleID,
			severity: f.Severity,
			filePath: scanPath,
			line:     line,
			column:   col,
			message:  msg,
			category: "vulnerability",
		})
	}
	return entries
}

func appendMetaLintEntries(entries []sarifEntry, findings []metalinter.Finding) []sarifEntry {
	for _, f := range findings {
		col := f.Column
		if col < 1 {
			col = 1
		}
		line := f.Line
		if line < 1 {
			line = 1
		}
		ruleID := f.Code
		if ruleID == "" {
			ruleID = f.Tool
		}
		entries = append(entries, sarifEntry{
			ruleID:   ruleID,
			severity: f.Severity,
			filePath: f.File,
			line:     line,
			column:   col,
			message:  f.Message,
			category: "meta-lint",
		})
	}
	return entries
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// writeSARIFLog builds a complete SARIF v2.1.0 log from the intermediate
// entry list and encodes it to w.
func writeSARIFLog(w io.Writer, entries []sarifEntry) error {
	rules, ruleIndex := buildSARIFRules(entries)
	results := buildSARIFResults(entries, ruleIndex)
	log := buildSARIFLog(rules, results)
	return encodeSARIF(w, log)
}

func buildSARIFResults(entries []sarifEntry, ruleIndex map[string]int) []SARIFResult {
	results := make([]SARIFResult, 0, len(entries))
	for _, e := range entries {
		idx, ok := ruleIndex[e.ruleID]
		if !ok {
			idx = 0
		}
		results = append(results, SARIFResult{
			RuleID:    e.ruleID,
			RuleIndex: idx,
			Level:     mapSeverityToSARIF(e.severity),
			Message:   SARIFMessage{Text: e.message},
			Locations: []SARIFLocation{{
				PhysicalLocation: SARIFPhysicalLocation{
					ArtifactLocation: SARIFArtifactLocation{URI: e.filePath},
					Region: SARIFRegion{
						StartLine:   e.line,
						StartColumn: e.column,
					},
				},
			}},
		})
	}
	return results
}

func buildSARIFLog(rules []SARIFRule, results []SARIFResult) SARIFLog {
	if results == nil {
		results = []SARIFResult{}
	}
	return SARIFLog{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []SARIFRun{{
			Tool: SARIFTool{
				Driver: SARIFDriver{
					Name:           "AILINTER",
					Version:        version.Semver(),
					InformationURI: "https://ailinter.dev",
					Rules:          rules,
				},
			},
			Results:    results,
			ColumnKind: "utf16CodeUnits",
		}},
	}
}

func encodeSARIF(w io.Writer, log SARIFLog) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}

// buildSARIFRules extracts a de-duplicated list of rules from entries.
// Returns the rules slice and a map of ruleID → ruleIndex.
func buildSARIFRules(entries []sarifEntry) ([]SARIFRule, map[string]int) {
	seen := make(map[string]bool)
	var rules []SARIFRule
	ruleIndex := make(map[string]int)

	for _, e := range entries {
		if seen[e.ruleID] {
			continue
		}
		seen[e.ruleID] = true
		ruleIndex[e.ruleID] = len(rules)

		helpURI := "https://ailinter.dev/docs"
		switch e.category {
		case "secret":
			helpURI = "https://ailinter.dev/docs/secrets"
		case "vulnerability":
			helpURI = "https://ailinter.dev/docs/vulnerabilities"
		case "quality":
			helpURI = "https://ailinter.dev/docs/quality"
		case "meta-lint":
			helpURI = "https://ailinter.dev/docs/meta-lint"
		}

		rules = append(rules, SARIFRule{
			ID:   e.ruleID,
			Name: e.ruleID,
			ShortDescription: SARIFMessage{
				Text: e.message,
			},
			HelpURI: helpURI,
			Properties: SARIFProperties{
				Category: e.category,
			},
		})
	}

	return rules, ruleIndex
}

// sarifOutput returns an io.Writer for SARIF output.
// If outputPath is non-empty, it creates/opens that file.
// Otherwise it returns os.Stdout.
func sarifOutput(outputPath string) io.Writer {
	if outputPath != "" {
		f, err := os.Create(outputPath)
		if err != nil {
			// Fall back to stdout on file error
			return os.Stdout
		}
		return f
	}
	return os.Stdout
}

// mapSeverityToSARIF maps ailinter severity levels to SARIF severity levels.
func mapSeverityToSARIF(severity string) string {
	switch strings.ToLower(severity) {
	case "critical", "error":
		return "error"
	case "alert", "warning":
		return "warning"
	default:
		return "note"
	}
}
