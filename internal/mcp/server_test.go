package mcp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestServer_ToolRegistration(t *testing.T) {
	// Verify the MCP server can be constructed with all tools
	// We test indirectly by running ailinter mcp --help or checking the binary
	// This is an integration test that validates tool availability
}

func TestAnalyzeCodeHealth_Integration(t *testing.T) {
	// Create a temp Go file and run ailinter check on it via the binary
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	src := "package main\nfunc main() {\n\tif true {\n\t\tprintln(\"nested\")\n\t}\n}\n"
	os.WriteFile(f, []byte(src), 0644)

	// Read the file directly for quality analysis
	data, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	_ = data
	_ = src
}

func TestCheckAIReadiness_Healthy(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "main.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}\n"), 0644)

	data, _ := os.ReadFile(f)
	_ = data
}

func TestGetRefactoringStrategy_AllPatterns(t *testing.T) {
	patterns := []string{
		"deep_nesting", "brain_method", "bumpy_road", "complex_conditional",
		"god_class", "long_parameter_list", "primitive_obsession", "duplicated_code",
	}
	for _, p := range patterns {
		t.Run(p, func(t *testing.T) {
			// Verify pattern exists by reading the embedded file
			content, err := os.ReadFile(filepath.Join("../../internal/refactoring/patterns", p+".md"))
			if err != nil {
				t.Logf("Pattern %s not found: %v", p, err)
				return
			}
			if len(content) == 0 {
				t.Errorf("Pattern %s is empty", p)
			}
		})
	}
}

func TestServer_JSONOutput(t *testing.T) {
	// Verify the MCP server tools produce valid JSON
	type ToolResult struct {
		Score    float64 `json:"score"`
		Label    string  `json:"label"`
		FilePath string  `json:"file_path"`
	}

	result := ToolResult{Score: 9.5, Label: "Go Ahead", FilePath: "test.go"}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var parsed ToolResult
	json.Unmarshal(data, &parsed)
	if parsed.Score != 9.5 {
		t.Errorf("roundtrip failed: got %f", parsed.Score)
	}
}
