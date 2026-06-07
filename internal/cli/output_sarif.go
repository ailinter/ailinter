package cli

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/git"
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
	Category         string  `json:"category,omitempty"`
	SecuritySeverity float64 `json:"security-severity,omitempty"`
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
		// Use per-file path when available (directory mode), fall back to scanPath.
		fp := f.FilePath
		if fp == "" {
			fp = scanPath
		}
		entries = append(entries, sarifEntry{
			ruleID:   f.RuleID,
			severity: f.Severity,
			filePath: fp,
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
		// Use per-file path when available (directory mode), fall back to scanPath.
		fp := f.FilePath
		if fp == "" {
			fp = scanPath
		}
		entries = append(entries, sarifEntry{
			ruleID:   f.RuleID,
			severity: f.Severity,
			filePath: fp,
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
//
// Paths are normalized to be repo-relative for GitHub Code Scanning
// compatibility. The SARIF spec requires URIs to be relative to the
// repository root; absolute filesystem paths (e.g.
// /Users/user/project/src/main.go) cause "Preview unavailable" in the
// GitHub Security tab.
func writeSARIFLog(w io.Writer, entries []sarifEntry) error {
	normalizeEntryPaths(entries)

	rules, ruleIndex := buildSARIFRules(entries)
	results := buildSARIFResults(entries, ruleIndex)
	log := buildSARIFLog(rules, results)
	return encodeSARIF(w, log)
}

// normalizeEntryPaths converts absolute file paths in all entries to paths
// relative to the git repository root. This ensures GitHub Code Scanning
// can map SARIF URIs back to repository files.
//
// If the git repo root cannot be determined (not a git repository, git not
// installed), paths are left unchanged (absolute). This is a graceful
// degradation — the SARIF is still valid, just not compatible with GitHub
// Code Scanning.
func normalizeEntryPaths(entries []sarifEntry) {
	if len(entries) == 0 {
		return
	}

	repoRoot := findRepoRoot(entries)
	if repoRoot == "" {
		return
	}

	for i := range entries {
		if filepath.IsAbs(entries[i].filePath) {
			rel, err := filepath.Rel(repoRoot, entries[i].filePath)
			if err == nil && !strings.HasPrefix(rel, "..") {
				entries[i].filePath = rel
			}
		}
	}
}

// findRepoRoot discovers the git repository root from the first absolute
// file path in the entries. It tries the file's parent directory first
// (since git -C requires a directory), then falls back to common prefix
// detection if git fails.
func findRepoRoot(entries []sarifEntry) string {
	for _, e := range entries {
		if !filepath.IsAbs(e.filePath) {
			continue
		}
		// Try the parent directory (file itself won't work with git -C).
		root, err := git.FindRepoRoot(filepath.Dir(e.filePath))
		if err == nil {
			return root
		}
		// Fall back to trying the file's grandparent chain — some files
		// may be in symlinked directories where git -C fails on the
		// symlinked path but the .git is discoverable via other means.
	}
	return ""
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
// Each rule includes a security-severity score (for GitHub severity taxonomy)
// and a stable name/description from the curated ruleDescriptions map.
func buildSARIFRules(entries []sarifEntry) ([]SARIFRule, map[string]int) {
	// First pass: compute maximum severity score per rule across all entries.
	ruleMaxSev := make(map[string]float64)
	for _, e := range entries {
		s := severityToScore(e.severity)
		if s > ruleMaxSev[e.ruleID] {
			ruleMaxSev[e.ruleID] = s
		}
	}

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

		name := ruleName(e.ruleID)
		rules = append(rules, SARIFRule{
			ID:   e.ruleID,
			Name: name,
			ShortDescription: SARIFMessage{
				Text: name,
			},
			HelpURI: helpURI,
			Properties: SARIFProperties{
				Category:         e.category,
				SecuritySeverity: ruleMaxSev[e.ruleID],
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

// severityToScore maps an ailinter severity string to a numeric score (0.0-10.0)
// for GitHub's security-severity taxonomy. The score determines the severity
// label shown in the GitHub Security tab:
//   - 9.0+  → Critical
//   - 7.0-8.9 → High
//   - 4.0-6.9 → Medium
//   - 0.1-3.9 → Low
//   - 0.0 or unset → falls back to SARIF level
func severityToScore(severity string) float64 {
	switch strings.ToLower(severity) {
	case "critical":
		return 9.5
	case "error":
		return 8.0
	case "alert":
		return 5.0
	case "warning":
		return 4.0
	default:
		return 1.0
	}
}

// ruleDescriptions maps known rule IDs to stable, human-readable descriptions.
// This ensures rule names are consistent across all SARIF output instances
// and are not dependent on the first finding's message text.
var ruleDescriptions = map[string]string{
	// Code Quality
	"deep_nesting":        "Deeply nested control flow",
	"brain_method":        "Overly large method / function",
	"bumpy_road":          "Code with bumpy readability — many short blocks",
	"complex_conditional": "Complex boolean condition with many branches",
	"complex_method":      "Method with high cyclomatic complexity",
	"long_parameter_list": "Function with too many parameters",
	"primitive_obsession": "Excessive use of primitive types",
	"duplicated_code":     "Duplicated code block",
	"god_class":           "Class with too many responsibilities",
	"file_bloat":          "File exceeds size thresholds",
	"paragraph_of_code":   "Long block of consecutive non-blank lines",
	"lazy_element":        "Unnecessary intermediate variable or function",
	"message_chains":      "Long method call chain",
	"long_scope_variable": "Variable used far from declaration",
	"long_switch":         "Long switch statement",
	"global_data":         "Excessive global state",
	"excessive_comments":  "Too many comments relative to code",
	"low_cohesion":        "Low cohesion in class/module",
	"brain_class":         "Overly large class",
	"code_duplication":    "Duplicate code pattern",

	// Secrets
	"generic-api-key":     "Generic API key detected",
	"gcp-api-key":         "Google Cloud Platform API key",
	"aws-access-key":      "AWS access key",
	"stripe-access-token": "Stripe access token",

	// Vulnerabilities
	"go_path_traversal":          "Path traversal via filepath.Join with user input",
	"go_sql_injection":           "SQL query built with string formatting",
	"go_exec_shell_injection":    "Command injection via exec.Command with shell",
	"go_template_html_xss":       "XSS via template.HTML() bypassing auto-escaping",
	"go_ssrf_http":               "SSRF via http.Get() with user-controlled URL",
	"eval_injection":             "eval() with untrusted input",
	"innerHTML_xss":              "innerHTML assignment with untrusted content",
	"aes_ecb_mode":               "AES in ECB mode (leaks plaintext structure)",
	"weak_hash_md5":              "MD5 hash (collision vulnerability)",
	"weak_hash_sha1":             "SHA-1 hash (collision vulnerability)",
	"weak_crypto_des":            "DES/3DES encryption (deprecated)",
	"tls_verification_disabled":  "TLS certificate verification disabled",
	"xml_unsafe_parse":           "XML parsing vulnerable to XXE",
	"unsafe_yaml_load":           "Unsafe YAML deserialization",
	"pickle_wrapper_load":        "Unsafe pickle deserialization",
	"python_subprocess_shell":    "subprocess with shell=True",
	"react_dangerously_set_html": "dangerouslySetInnerHTML with untrusted content",
	"document_write_xss":         "document.write() with untrusted input",
	"new_function_injection":     "new Function() with dynamic string",

	// Metalint
	"spelling":    "Spelling error detected",
	"gofmt":       "Code formatting issue",
	"go_vet":      "Go vet issue",
	"staticcheck": "Static analysis issue",
}

// ruleName returns a stable, human-readable name for a rule ID.
// It first checks the curated ruleDescriptions map, then falls back
// to converting snake_case to Title Case (e.g., "go_path_traversal" → "Go Path Traversal").
func ruleName(ruleID string) string {
	if desc, ok := ruleDescriptions[ruleID]; ok {
		return desc
	}
	// Fallback: convert snake_case to Title Case
	words := strings.Split(ruleID, "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
