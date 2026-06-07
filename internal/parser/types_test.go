package parser

import (
	"strings"
	"testing"
)

func TestScoreTiers_MatchClassify(t *testing.T) {
	t.Parallel()
	for score := 0; score <= 100; score++ {
		expected := TestClassifyHelper(score)
		var found string
		for _, tier := range ScoreTiers() {
			if score >= tier.MinScore && score <= tier.MaxScore {
				found = tier.Label
				break
			}
		}
		if found != expected {
			t.Errorf("score %d: classify()=%q but ScoreTiers() covers as %q", score, expected, found)
		}
	}
}

func TestScoreTiers_Boundaries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		score int
		label string
	}{
		{100, LabelGoAhead},
		{80, LabelGoAhead},
		{79, LabelProceedWithCare},
		{60, LabelProceedWithCare},
		{59, LabelNeedsWork},
		{40, LabelNeedsWork},
		{39, LabelStopRefactor},
		{0, LabelStopRefactor},
		{-1, LabelStopRefactor},
	}
	for _, tc := range tests {
		t.Run(tc.label, func(t *testing.T) {
			label := TestClassifyHelper(tc.score)
			if label != tc.label {
				t.Errorf("score %d: got %q, want %q", tc.score, label, tc.label)
			}
		})
	}
}

func TestScoreTiers_AllTiersPresent(t *testing.T) {
	t.Parallel()
	tiers := ScoreTiers()
	if len(tiers) != 4 {
		t.Fatalf("expected 4 tiers, got %d", len(tiers))
	}

	labels := map[string]bool{}
	for _, tier := range tiers {
		labels[tier.Label] = true
	}
	expected := []string{LabelGoAhead, LabelProceedWithCare, LabelNeedsWork, LabelStopRefactor}
	for _, l := range expected {
		if !labels[l] {
			t.Errorf("missing tier label: %s", l)
		}
	}
}

func TestScoreTiers_OrderedDescending(t *testing.T) {
	t.Parallel()
	tiers := ScoreTiers()
	for i := 1; i < len(tiers); i++ {
		if tiers[i].MinScore > tiers[i-1].MinScore {
			t.Errorf("tiers not in descending order: %d (idx %d) > %d (idx %d)",
				tiers[i].MinScore, i, tiers[i-1].MinScore, i-1)
		}
	}
}

func TestTierReferenceTable_ContainsAllTiers(t *testing.T) {
	t.Parallel()
	table := TierReferenceTable()
	for _, tier := range ScoreTiers() {
		if !strings.Contains(table, tier.Label) {
			t.Errorf("TierReferenceTable missing label: %s", tier.Label)
		}
	}
}

func TestTierReferenceTable_Format(t *testing.T) {
	t.Parallel()
	table := TierReferenceTable()
	if !strings.HasPrefix(table, "| Score |") {
		t.Error("table should start with header row")
	}
	if !strings.Contains(table, "|---") {
		t.Error("table should contain separator row")
	}
	// Verify all 4 tiers produce data rows
	lines := strings.Split(strings.TrimSpace(table), "\n")
	if len(lines) != 6 { // header + separator + 4 data rows
		t.Errorf("expected 6 lines (header+sep+4 tiers), got %d", len(lines))
	}
}

func TestTierReferenceTable_RangesAreCorrect(t *testing.T) {
	t.Parallel()
	table := TierReferenceTable()
	// Should contain the correct range representations
	checks := []string{
		"| 80-100",
		"| 60-79",
		"| 40-59",
		"| <40",
	}
	for _, c := range checks {
		if !strings.Contains(table, c) {
			t.Errorf("table should contain %q", c)
		}
	}
}

func TestScoreTierConstants_MatchClassify(t *testing.T) {
	t.Parallel()
	if GoAheadThreshold != 80 {
		t.Errorf("GoAheadThreshold = %d, want 80", GoAheadThreshold)
	}
	if ProceedWithCareThreshold != 60 {
		t.Errorf("ProceedWithCareThreshold = %d, want 60", ProceedWithCareThreshold)
	}
	if NeedsWorkThreshold != 40 {
		t.Errorf("NeedsWorkThreshold = %d, want 40", NeedsWorkThreshold)
	}
}
