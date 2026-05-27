package parser_test

import (
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/parser"
)

func sm(s string) []string { return strings.Split(s, "\n") }

func TestNesting_Shallow(t *testing.T) {
	lines := sm("func f() {\n\tx := 1\n\treturn x\n}")
	s := parser.DetectNestingSmell(lines, 4, 5)
	if s != nil {
		t.Errorf("shallow function should not trigger nesting, got %s", s.Message)
	}
}

func TestNesting_Deep(t *testing.T) {
	lines := sm("func f() {\nif a {\nif b {\nif c {\nif d {\nif e {\nx=1\n}\n}\n}\n}\n}\n}")
	s := parser.DetectNestingSmell(lines, 4, 5)
	if s == nil {
		t.Fatal("expected nesting smell for 5-deep")
	}
	if s.Severity != "alert" {
		t.Errorf("expected alert severity, got %s", s.Severity)
	}
}

func TestNesting_BelowThreshold(t *testing.T) {
	lines := sm("func f() {\nif a {\nx=1\n}\n}")
	s := parser.DetectNestingSmell(lines, 4, 5)
	if s != nil {
		t.Errorf("1-deep nesting should not trigger: %s", s.Message)
	}
}

func TestFileBloat_Small(t *testing.T) {
	s := parser.DetectFileBloat(50, 100, 200, 400)
	if s != nil {
		t.Error("small file should not trigger bloat")
	}
}

func TestFileBloat_Warning(t *testing.T) {
	s := parser.DetectFileBloat(150, 100, 200, 400)
	if s == nil {
		t.Fatal("expected file_bloat warning")
	}
	if s.Severity != "warning" {
		t.Errorf("expected warning, got %s", s.Severity)
	}
}

func TestFileBloat_Alert(t *testing.T) {
	s := parser.DetectFileBloat(250, 100, 200, 400)
	if s == nil {
		t.Fatal("expected file_bloat alert")
	}
	if s.Severity != "alert" {
		t.Errorf("expected alert, got %s", s.Severity)
	}
}

func TestFileBloat_Critical(t *testing.T) {
	s := parser.DetectFileBloat(500, 100, 200, 400)
	if s == nil {
		t.Fatal("expected file_bloat critical")
	}
	if s.Severity != "critical" {
		t.Errorf("expected critical, got %s", s.Severity)
	}
}

func TestBrainMethodSmell_Normal(t *testing.T) {
	bloats := []parser.FunctionBloat{{Name: "foo", LineCount: 40, LineStart: 1}}
	smells := parser.DetectBrainMethodSmell(bloats, 80, 200)
	if len(smells) > 0 {
		t.Error("40-line function should not trigger brain_method")
	}
}

func TestBrainMethodSmell_Warning(t *testing.T) {
	bloats := []parser.FunctionBloat{{Name: "foo", LineCount: 90, LineStart: 1}}
	smells := parser.DetectBrainMethodSmell(bloats, 80, 200)
	if len(smells) == 0 {
		t.Fatal("expected brain_method warning")
	}
}

func TestBrainMethodSmell_Alert(t *testing.T) {
	bloats := []parser.FunctionBloat{{Name: "foo", LineCount: 250, LineStart: 1}}
	smells := parser.DetectBrainMethodSmell(bloats, 80, 200)
	if len(smells) == 0 {
		t.Fatal("expected brain_method alert")
	}
}
