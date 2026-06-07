package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/metalinter"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/vulnerability"
)

func TestWriteSARIFCombined_QualityOnly(t *testing.T) {
	t.Parallel()

	results := []analyzer.QualityResult{
		{
			Score:    65,
			FilePath: "src/main.go",
			Smells: []analyzer.Smell{
				{Name: "deep_nesting", Severity: "warning", LineStart: 10, Message: "Nesting depth 5"},
				{Name: "brain_method", Severity: "alert", LineStart: 20, Message: "Function too long (120 lines)"},
			},
		},
	}

	var buf bytes.Buffer
	err := WriteSARIFCombined(&buf, results, nil, nil, nil, "")
	if err != nil {
		t.Fatalf("WriteSARIFCombined returned error: %v", err)
	}

	var log SARIFLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("SARIF output must be valid JSON: %v\nRaw:\n%s", err, buf.String())
	}

	if log.Version != "2.1.0" {
		t.Errorf("version = %q, want 2.1.0", log.Version)
	}
	if len(log.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(log.Runs))
	}

	run := log.Runs[0]
	if run.Tool.Driver.Name != "AILINTER" {
		t.Errorf("tool name = %q, want AILINTER", run.Tool.Driver.Name)
	}
	if len(run.Tool.Driver.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(run.Tool.Driver.Rules))
	}
	if len(run.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(run.Results))
	}

	// Verify result content
	foundDeep := false
	foundBrain := false
	for _, r := range run.Results {
		if r.RuleID == "deep_nesting" {
			foundDeep = true
			if r.Level != "warning" {
				t.Errorf("deep_nesting level = %q, want warning", r.Level)
			}
			if len(r.Locations) != 1 {
				t.Fatal("expected 1 location")
			}
			if r.Locations[0].PhysicalLocation.Region.StartLine != 10 {
				t.Errorf("startLine = %d, want 10", r.Locations[0].PhysicalLocation.Region.StartLine)
			}
		}
		if r.RuleID == "brain_method" {
			foundBrain = true
			if r.Level != "warning" {
				t.Errorf("brain_method level = %q, want warning (alert -> warning)", r.Level)
			}
		}
	}
	if !foundDeep {
		t.Error("deep_nesting result not found")
	}
	if !foundBrain {
		t.Error("brain_method result not found")
	}
}

func TestWriteSARIFCombined_SecretsVulnsMetaLint(t *testing.T) {
	t.Parallel()

	secrets := []secrets.SecretFinding{
		{RuleID: "aws-access-key", Severity: "critical", Line: 5, Column: 10, Description: "AWS key found"},
	}

	vulns := []vulnerability.Finding{
		{RuleID: "pickle_deserialization", Category: "deserialization", Severity: "critical", Line: 15, Column: 1, Description: "Unsafe pickle"},
	}

	ml := []metalinter.Finding{
		{Tool: "gofmt", Code: "formatting", Severity: "warning", File: "test.go", Line: 3, Column: 1, Message: "File not gofmt-ed"},
	}

	var buf bytes.Buffer
	err := WriteSARIFCombined(&buf, nil, secrets, vulns, ml, "/path/to/scan")
	if err != nil {
		t.Fatalf("WriteSARIFCombined returned error: %v", err)
	}

	var log SARIFLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("SARIF output must be valid JSON: %v\nRaw:\n%s", err, buf.String())
	}

	run := log.Runs[0]
	if len(run.Results) != 3 {
		t.Errorf("expected 3 results, got %d", len(run.Results))
	}
	if len(run.Tool.Driver.Rules) != 3 {
		t.Errorf("expected 3 rules, got %d", len(run.Tool.Driver.Rules))
	}

	// Verify severity mapping
	for _, r := range run.Results {
		switch r.RuleID {
		case "aws-access-key":
			if r.Level != "error" {
				t.Errorf("aws-access-key level = %q, want error (critical -> error)", r.Level)
			}
		case "pickle_deserialization":
			if r.Level != "error" {
				t.Errorf("pickle_deserialization level = %q, want error", r.Level)
			}
		case "formatting":
			if r.Level != "warning" {
				t.Errorf("formatting level = %q, want warning", r.Level)
			}
		}
	}
}

func TestWriteSARIFCombined_EmptyResults(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := WriteSARIFCombined(&buf, nil, nil, nil, nil, "")
	if err != nil {
		t.Fatalf("WriteSARIFCombined returned error: %v", err)
	}

	var log SARIFLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("SARIF output must be valid JSON: %v", err)
	}

	if len(log.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(log.Runs))
	}

	run := log.Runs[0]
	if len(run.Results) != 0 {
		t.Errorf("expected 0 results for empty input, got %d", len(run.Results))
	}
	// Rules should also be empty when no findings
	if len(run.Tool.Driver.Rules) != 0 {
		t.Errorf("expected 0 rules for empty input, got %d", len(run.Tool.Driver.Rules))
	}
}

func TestMapSeverityToSARIF(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"critical": "error",
		"error":    "error",
		"alert":    "warning",
		"warning":  "warning",
		"info":     "note",
		"note":     "note",
		"unknown":  "note",
		"":         "note",
	}

	for input, want := range cases {
		if got := mapSeverityToSARIF(input); got != want {
			t.Errorf("mapSeverityToSARIF(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestWriteSARIFCombined_DeduplicatesRules(t *testing.T) {
	t.Parallel()

	// Multiple findings with the same RuleID should produce only one rule
	results := []analyzer.QualityResult{
		{
			FilePath: "a.go",
			Smells: []analyzer.Smell{
				{Name: "deep_nesting", Severity: "warning", LineStart: 5, Message: "Nesting depth 5"},
			},
		},
		{
			FilePath: "b.go",
			Smells: []analyzer.Smell{
				{Name: "deep_nesting", Severity: "warning", LineStart: 10, Message: "Nesting depth 5"},
			},
		},
	}

	var buf bytes.Buffer
	err := WriteSARIFCombined(&buf, results, nil, nil, nil, "")
	if err != nil {
		t.Fatalf("WriteSARIFCombined returned error: %v", err)
	}

	var log SARIFLog
	json.Unmarshal(buf.Bytes(), &log)
	run := log.Runs[0]

	if len(run.Tool.Driver.Rules) != 1 {
		t.Errorf("expected 1 rule (deduplicated), got %d", len(run.Tool.Driver.Rules))
	}
	if len(run.Results) != 2 {
		t.Errorf("expected 2 results (one per file), got %d", len(run.Results))
	}
}

func TestSARIFLog_HasRequiredFields(t *testing.T) {
	t.Parallel()

	results := []analyzer.QualityResult{
		{
			Score:    75,
			FilePath: "test.go",
			Smells: []analyzer.Smell{
				{Name: "test_rule", Severity: "warning", LineStart: 1, Message: "Test message"},
			},
		},
	}

	var buf bytes.Buffer
	WriteSARIFCombined(&buf, results, nil, nil, nil, "")

	var log SARIFLog
	json.Unmarshal(buf.Bytes(), &log)

	// Verify all required SARIF v2.1.0 fields are present
	if log.Schema == "" {
		t.Error("missing $schema")
	}
	if log.Version != "2.1.0" {
		t.Errorf("version = %q", log.Version)
	}
	if len(log.Runs) == 0 {
		t.Fatal("no runs")
	}

	run := log.Runs[0]
	if run.Tool.Driver.Name == "" {
		t.Error("missing tool driver name")
	}
	if run.Tool.Driver.Version == "" {
		t.Error("missing tool driver version")
	}
	if len(run.Results) == 0 {
		t.Error("no results")
	}

	r := run.Results[0]
	if r.RuleID == "" {
		t.Error("missing ruleId")
	}
	if r.Level == "" {
		t.Error("missing level")
	}
	if r.Message.Text == "" {
		t.Error("missing message text")
	}
	if len(r.Locations) == 0 {
		t.Fatal("missing locations")
	}
	loc := r.Locations[0]
	if loc.PhysicalLocation.ArtifactLocation.URI == "" {
		t.Error("missing artifact location URI")
	}
	if loc.PhysicalLocation.Region.StartLine == 0 {
		t.Error("missing startLine")
	}
}
