package refactoring_test

import (
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/refactoring"
)

func TestLookup_DeepNesting(t *testing.T) {
	p := refactoring.Lookup("deep_nesting")
	if p == nil {
		t.Fatal("deep_nesting pattern not found")
	}
	if !strings.Contains(p.Content, "Guard Clauses") {
		t.Error("expected Guard Clauses in pattern")
	}
}

func TestLookup_BrainMethod(t *testing.T) {
	p := refactoring.Lookup("brain_method")
	if p == nil {
		t.Fatal("brain_method pattern not found")
	}
	if !strings.Contains(p.Content, "Extract Method") {
		t.Error("expected Extract Method reference")
	}
}

func TestLookup_ComplexMethod(t *testing.T) {
	p := refactoring.Lookup("complex_method")
	if p == nil {
		t.Fatal("complex_method pattern not found")
	}
	if !strings.Contains(p.Content, "Cyclomatic Complexity") {
		t.Error("expected Cyclomatic Complexity reference")
	}
	if !strings.Contains(p.Content, "State Machine") {
		t.Error("expected State Machine pattern reference")
	}
	if !strings.Contains(p.Content, "Table Lookup") {
		t.Error("expected Table Lookup pattern reference")
	}
}

func TestLookup_BumpyRoad(t *testing.T) {
	p := refactoring.Lookup("bumpy_road")
	if p == nil {
		t.Fatal("bumpy_road pattern not found")
	}
}

func TestLookup_ComplexConditional(t *testing.T) {
	p := refactoring.Lookup("complex_conditional")
	if p == nil {
		t.Fatal("complex_conditional pattern not found")
	}
}

func TestLookup_GodClass(t *testing.T) {
	p := refactoring.Lookup("god_class")
	if p == nil {
		t.Fatal("god_class pattern not found")
	}
}

func TestLookup_LongParameterList(t *testing.T) {
	p := refactoring.Lookup("long_parameter_list")
	if p == nil {
		t.Fatal("long_parameter_list pattern not found")
	}
}

func TestLookup_PrimitiveObsession(t *testing.T) {
	p := refactoring.Lookup("primitive_obsession")
	if p == nil {
		t.Fatal("primitive_obsession pattern not found")
	}
}

func TestLookup_DuplicatedCode(t *testing.T) {
	p := refactoring.Lookup("duplicated_code")
	if p == nil {
		t.Fatal("duplicated_code pattern not found")
	}
}

func TestLookup_NotFound(t *testing.T) {
	p := refactoring.Lookup("nonexistent_pattern")
	if p != nil {
		t.Error("expected nil for unknown pattern")
	}
}

func TestListPatterns(t *testing.T) {
	patterns := refactoring.ListPatterns()
	if len(patterns) < 23 {
		t.Errorf("expected at least 23 patterns, got %d", len(patterns))
	}
}

func TestLookup_MagicNumber(t *testing.T) {
	p := refactoring.Lookup("magic_number")
	if p == nil {
		t.Fatal("magic_number pattern not found")
	}
	if !strings.Contains(p.Content, "Named Constants") {
		t.Error("expected Named Constants reference")
	}
}

func TestLookup_LowCohesion(t *testing.T) {
	p := refactoring.Lookup("low_cohesion")
	if p == nil {
		t.Fatal("low_cohesion pattern not found")
	}
	if !strings.Contains(p.Content, "Extract Unrelated") {
		t.Error("expected Extract Unrelated reference")
	}
}

func TestLookup_DataClass(t *testing.T) {
	p := refactoring.Lookup("data_class")
	if p == nil {
		t.Fatal("data_class pattern not found")
	}
	if !strings.Contains(p.Content, "Tell Don't Ask") {
		t.Error("expected Tell Don't Ask reference")
	}
}

func TestLookup_RefusedBequest(t *testing.T) {
	p := refactoring.Lookup("refused_bequest")
	if p == nil {
		t.Fatal("refused_bequest pattern not found")
	}
	if !strings.Contains(p.Content, "Delegation") {
		t.Error("expected Delegation reference")
	}
}

func TestLookup_ShotgunSurgery(t *testing.T) {
	p := refactoring.Lookup("shotgun_surgery")
	if p == nil {
		t.Fatal("shotgun_surgery pattern not found")
	}
	if !strings.Contains(p.Content, "Consolidate") {
		t.Error("expected Consolidate reference")
	}
}

func TestLookup_ParallelInheritance(t *testing.T) {
	p := refactoring.Lookup("parallel_inheritance")
	if p == nil {
		t.Fatal("parallel_inheritance pattern not found")
	}
	if !strings.Contains(p.Content, "Parallel") {
		t.Error("expected Parallel reference")
	}
}

func TestLookup_LongMethod(t *testing.T) {
	p := refactoring.Lookup("long_method")
	if p == nil {
		t.Fatal("long_method pattern not found")
	}
	if !strings.Contains(p.Content, "Brain Method") {
		t.Error("expected Brain Method reference in alias pattern")
	}
}
