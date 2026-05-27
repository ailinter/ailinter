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
	if len(patterns) < 8 {
		t.Errorf("expected at least 8 patterns, got %d", len(patterns))
	}
}
