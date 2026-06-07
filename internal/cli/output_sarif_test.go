package cli

import (
	"bytes"
	"encoding/json"
	"strconv"
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

func TestSeverityToScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity string
		label    string // expected GitHub label bucket
		min      float64
		max      float64
	}{
		{"critical", "Critical", 9.0, 10.0},
		{"error", "High", 7.0, 8.9},
		{"alert", "Medium", 4.0, 6.9},
		{"warning", "Medium", 4.0, 6.9},
		{"info", "Low", 0.1, 3.9},
		{"note", "Low", 0.1, 3.9},
		{"unknown", "Low", 0.1, 3.9},
		{"", "Low", 0.1, 3.9},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			got := severityToScore(tt.severity)
			if got < tt.min || got > tt.max {
				t.Errorf("severityToScore(%q) = %.1f, want in [%.1f, %.1f] (GitHub: %s)",
					tt.severity, got, tt.min, tt.max, tt.label)
			}
		})
	}
}

func TestSARIFRuleStableDescriptions(t *testing.T) {
	t.Parallel()

	// Known rule IDs that should have stable descriptions from the curated map.
	knownRuleIDs := []string{
		"deep_nesting",
		"brain_method",
		"bumpy_road",
		"complex_conditional",
		"generic-api-key",
		"aws-access-key",
		"go_sql_injection",
		"go_path_traversal",
		"spelling",
		"gofmt",
		"staticcheck",
	}

	for _, ruleID := range knownRuleIDs {
		t.Run(ruleID, func(t *testing.T) {
			name := ruleName(ruleID)
			if name == "" {
				t.Errorf("ruleName(%q) returned empty string", ruleID)
			}
			if name == ruleID {
				t.Errorf("ruleName(%q) = %q, expected a human-readable description, not the raw ID", ruleID, name)
			}
			// Verify the description is stable: calling twice gives the same result
			name2 := ruleName(ruleID)
			if name != name2 {
				t.Errorf("ruleName(%q) is not stable: first call = %q, second call = %q", ruleID, name, name2)
			}
		})
	}
}

func TestSARIFRuleFallbackName(t *testing.T) {
	t.Parallel()

	// Unknown rule IDs should fall back to Title Case conversion.
	tests := []struct {
		ruleID string
		want   string
	}{
		{"my_custom_rule", "My Custom Rule"},
		{"single", "Single"},
		{"already-title", "Already-title"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ruleID, func(t *testing.T) {
			got := ruleName(tt.ruleID)
			if got != tt.want {
				t.Errorf("ruleName(%q) = %q, want %q", tt.ruleID, got, tt.want)
			}
		})
	}
}

func TestSARIFRulesHaveSecuritySeverity(t *testing.T) {
	t.Parallel()

	results := []analyzer.QualityResult{
		{
			Score:    65,
			FilePath: "src/main.go",
			Smells: []analyzer.Smell{
				{Name: "deep_nesting", Severity: "warning", LineStart: 10, Message: "Nesting depth 5"},
				{Name: "brain_method", Severity: "alert", LineStart: 20, Message: "Function too long (120 lines)"},
				{Name: "bumpy_road", Severity: "critical", LineStart: 30, Message: "Bumpy road detected"},
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
		t.Fatalf("SARIF output must be valid JSON: %v", err)
	}

	run := log.Runs[0]
	if len(run.Tool.Driver.Rules) == 0 {
		t.Fatal("expected at least one rule")
	}

	for _, rule := range run.Tool.Driver.Rules {
		t.Run(rule.ID, func(t *testing.T) {
			// Parse security-severity string to float for validation
			var sevFloat float64
			if rule.Properties.SecuritySeverity != "" {
				var err error
				sevFloat, err = strconv.ParseFloat(rule.Properties.SecuritySeverity, 64)
				if err != nil {
					t.Fatalf("rule %q has invalid security-severity string %q: %v",
						rule.ID, rule.Properties.SecuritySeverity, err)
				}
			}
			// Every rule must have a security-severity or the GitHub taxonomy won't work
			if sevFloat == 0 {
				t.Errorf("rule %q has security-severity = 0, expected non-zero", rule.ID)
			}
			// Must be in valid range
			if sevFloat < 0.0 || sevFloat > 10.0 {
				t.Errorf("rule %q has security-severity = %.1f, outside valid range [0.0, 10.0]",
					rule.ID, sevFloat)
			}
			// Must have a stable name (not the raw ruleID)
			if rule.Name == rule.ID {
				t.Errorf("rule %q has Name = %q, expected a human-readable name", rule.ID, rule.Name)
			}
			// Must have stable short description
			if rule.ShortDescription.Text == "" {
				t.Errorf("rule %q has empty shortDescription", rule.ID)
			}
			// Category should be set
			if rule.Properties.Category == "" {
				t.Errorf("rule %q has empty category", rule.ID)
			}
		})
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
