package main

import (
	"strings"
	"testing"
)

func TestIsValidLanguage(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want bool
	}{
		{"go", "go", true},
		{"python", "python", true},
		{"javascript", "javascript", true},
		{"typescript", "typescript", true},
		{"java", "java", true},
		{"csharp", "csharp", true},
		{"ruby", "ruby", true},
		{"swift", "swift", true},
		{"kotlin", "kotlin", true},
		{"rust", "rust", true},
		{"cpp", "cpp", true},
		{"c", "c", true},
		{"empty string", "", false},
		{"frobulator", "frobulator", false},
		{"uppercase Go", "Go", false},
		{"uppercase PYTHON", "PYTHON", false},
		{"c++", "c++", false},
		{"yaml", "yaml", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidLanguage(tt.lang); got != tt.want {
				t.Errorf("isValidLanguage(%q) = %v, want %v", tt.lang, got, tt.want)
			}
		})
	}
}

func TestRunListRules(t *testing.T) {
	tests := []struct {
		name    string
		lang    string
		wantErr bool
		errMsg  string
	}{
		{"valid go", "go", false, ""},
		{"valid python", "python", false, ""},
		{"invalid language", "frobulator", true, "unknown language"},
		{"empty language", "", true, "unknown language"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runListRules(tt.lang)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPrintDefaultRules(t *testing.T) {
	out := captureMainOutput(func() {
		printDefaultRules()
	})

	checks := []string{
		"Go", "Python", "JS/TS", "Java",
		"Nesting depth", "Cyclomatic complexity", "Function LOC", "File LOC", "Max function arguments", "Bumpy Road bumps",
		"4", "9", "80", "1000",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("printDefaultRules() output should contain %q", c)
		}
	}
}

func TestPrintLanguageRules(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		out := captureMainOutput(func() {
			printLanguageRules("go")
		})
		if !strings.Contains(out, "Rules for go") {
			t.Error("output should contain language name")
		}
		if !strings.Contains(out, "4") || !strings.Contains(out, "80") || !strings.Contains(out, "1000") {
			t.Error("output should contain go thresholds")
		}
	})

	t.Run("unknown language uses defaults", func(t *testing.T) {
		out := captureMainOutput(func() {
			printLanguageRules("rust")
		})
		if !strings.Contains(out, "rust") {
			t.Error("output should contain language name")
		}
		if !strings.Contains(out, "4") || !strings.Contains(out, "80") || !strings.Contains(out, "1000") {
			t.Error("unknown language should use default thresholds")
		}
	})

	knownLangs := []string{"go", "python", "javascript", "typescript", "java"}
	for _, lang := range knownLangs {
		t.Run(lang, func(t *testing.T) {
			out := captureMainOutput(func() {
				printLanguageRules(lang)
			})
			if !strings.Contains(out, lang) {
				t.Errorf("output should contain language name %q", lang)
			}
			if !strings.Contains(out, "Nesting depth") {
				t.Error("output should contain Nesting depth metric")
			}
		})
	}
}

func TestPrintTelemetryInfo(t *testing.T) {
	out := captureMainOutput(func() {
		printTelemetryInfo()
	})

	checks := []string{
		"Metrics collected",
		"cli.invocations",
		"mcp.tool_calls",
		"files.analyzed",
		"installations",
		"quality.score",
		"smells.detected", // gitleaks:allow
		"secrets.detected",
		"duration.seconds",
		"errors",
		"AILINTER_NO_TELEMETRY",
		"NOT collected",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("printTelemetryInfo() output should contain %q", c)
		}
	}
}

func TestTelemetryCommand(t *testing.T) {
	cmd := telemetryCommand()
	if cmd.Use != "telemetry" {
		t.Errorf("Use = %q, want %q", cmd.Use, "telemetry")
	}
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
	if cmd.Long == "" {
		t.Error("Long should not be empty")
	}
	if cmd.Run == nil {
		t.Error("Run should not be nil")
	}
}

func TestRulesCommand(t *testing.T) {
	cmd := rulesCommand()
	if cmd.Use != "rules" {
		t.Errorf("Use = %q, want %q", cmd.Use, "rules")
	}
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}

	listCmd, _, err := cmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("rules list subcommand not found: %v", err)
	}
	if listCmd == nil {
		t.Fatal("list command is nil")
	}
	if listCmd.Use != "list" {
		t.Errorf("list command Use = %q, want %q", listCmd.Use, "list")
	}

	langFlag := listCmd.Flags().Lookup("lang")
	if langFlag == nil {
		t.Fatal("list command should have --lang flag")
	}
	if langFlag.DefValue != "" {
		t.Errorf("--lang default = %q, want %q", langFlag.DefValue, "")
	}
}

func TestRunListRulesE(t *testing.T) {
	t.Run("with lang go", func(t *testing.T) {
		cmd := listRulesCommand()
		cmd.SetArgs([]string{"--lang", "go"})
		out := captureMainOutput(func() {
			if err := cmd.Execute(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Go") && !strings.Contains(out, "Nesting depth") {
			t.Error("output should contain language rules")
		}
	})

	t.Run("with invalid lang", func(t *testing.T) {
		cmd := listRulesCommand()
		cmd.SetArgs([]string{"--lang", "frobulator"})
		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "unknown language") {
			t.Errorf("error = %q, want containing %q", err.Error(), "unknown language")
		}
	})

	t.Run("with no lang flag", func(t *testing.T) {
		cmd := listRulesCommand()
		out := captureMainOutput(func() {
			if err := cmd.Execute(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Go") {
			t.Error("output should contain default rules")
		}
	})
}

func TestNewRootCommand(t *testing.T) {
	cmd := newRootCommand()
	if cmd.Use != "ailinter" {
		t.Errorf("Use = %q, want %q", cmd.Use, "ailinter")
	}
	if cmd.Version == "" {
		t.Error("Version should not be empty")
	}
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestMain_Runs(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("main() panicked: %v", r)
		}
	}()
}
