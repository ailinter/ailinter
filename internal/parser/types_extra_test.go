package parser

import (
	"testing"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{95, LabelGoAhead},
		{85, LabelGoAhead},
		{80, LabelGoAhead},
		{79, LabelProceedWithCare},
		{65, LabelProceedWithCare},
		{60, LabelProceedWithCare},
		{59, LabelNeedsWork},
		{40, LabelNeedsWork},
		{30, LabelStopRefactor},
		{29, LabelStopRefactor},
		{15, LabelStopRefactor},
		{0, LabelStopRefactor},
	}
	for _, tc := range tests {
		if got := Classify(tc.score); got != tc.want {
			t.Errorf("Classify(%d) = %q, want %q", tc.score, got, tc.want)
		}
	}
}

func TestClassifyBoundaries(t *testing.T) {
	if Classify(80) != LabelGoAhead {
		t.Error("80 should be Go Ahead")
	}
	if Classify(60) != LabelProceedWithCare {
		t.Error("60 should be Proceed with Care")
	}
	if Classify(30) != LabelStopRefactor {
		t.Error("30 should be Stop & Refactor")
	}
	if Classify(29) != LabelStopRefactor {
		t.Error("29 should be Stop & Refactor")
	}
}

type testFinding struct {
	Severity string
}

func TestVulnClassify(t *testing.T) {
	tests := []struct {
		name     string
		findings []testFinding
		want     string
	}{
		{"empty", []testFinding{}, VulnLabelClean},
		{"only warning", []testFinding{{Severity: "warning"}}, VulnLabelMonitor},
		{"has alert", []testFinding{{Severity: "alert"}}, VulnLabelRemediate},
		{"has critical", []testFinding{{Severity: "critical"}}, VulnLabelRemediate},
		{"mixed", []testFinding{{Severity: "warning"}, {Severity: "critical"}}, VulnLabelRemediate},
		{"unknown severity", []testFinding{{Severity: "info"}}, VulnLabelClean},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Convert to anonymous struct
			findings := make([]struct{ Severity string }, len(tc.findings))
			for i, f := range tc.findings {
				findings[i] = struct{ Severity string }{Severity: f.Severity}
			}
			if got := VulnClassify(findings); got != tc.want {
				t.Errorf("VulnClassify(%v) = %q, want %q", tc.findings, got, tc.want)
			}
		})
	}
}
