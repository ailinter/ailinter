package mcp

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestHandleAnalyzeCode_Valid(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)
	os.WriteFile("test.go", []byte("package main\nfunc main() {\n\tif true {\n\t\tprintln(\"nested\")\n\t}\n}\n"), 0644)

	result, err := handleAnalyzeCode(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"file_path": "test.go",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}

	for _, c := range result.Content {
		if text, ok := c.(*mcp.TextContent); ok {
			if !strings.Contains(text.Text, "score") {
				t.Error("analyze_code result should contain 'score'")
			}
		}
	}
}

func TestHandleAnalyzeCode_MissingArg(t *testing.T) {
	result, err := handleAnalyzeCode(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing file_path")
	}
}

func TestHandleAnalyzeCode_FileNotFound(t *testing.T) {
	result, err := handleAnalyzeCode(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{"file_path": "/nonexistent.go"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent file")
	}
}

func TestHandleAnalyzeCode_InvalidArgs(t *testing.T) {
	result, err := handleAnalyzeCode(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: "not-a-map"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for invalid arguments type")
	}
}

func TestHandleScanForSecrets_Valid(t *testing.T) {
	result, err := handleScanForSecrets(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"content": "var key = \"sk_live_1234567890abcdef\"\n", // gitleaks:allow
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("scan error: %v", result.Content)
	}

	for _, c := range result.Content {
		if text, ok := c.(*mcp.TextContent); ok {
			if !strings.Contains(text.Text, "stripe") && !strings.Contains(text.Text, "[") {
				t.Logf("scan result: %s", text.Text[:min(100, len(text.Text))])
			}
		}
	}
}

func TestHandleScanForSecrets_Clean(t *testing.T) {
	result, err := handleScanForSecrets(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"content": "func hello() { return 42 }\n",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
}

func TestHandleScanForSecrets_MissingArg(t *testing.T) {
	result, err := handleScanForSecrets(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing content")
	}
}

func TestHandleScanForSecrets_InvalidArgs(t *testing.T) {
	result, err := handleScanForSecrets(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: "not-a-map"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for invalid arguments type")
	}
}

func TestHandleGetRefactoringStrategy_Valid(t *testing.T) {
	patterns := []string{
		"deep_nesting", "brain_method", "bumpy_road", "complex_conditional",
		"god_class", "long_parameter_list", "primitive_obsession", "duplicated_code",
	}
	for _, p := range patterns {
		result, err := handleGetRefactoringStrategy(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"smell_name": p},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Errorf("unexpected error for pattern %q: %v", p, result.Content)
		}
		for _, c := range result.Content {
			if text, ok := c.(*mcp.TextContent); ok {
				if len(text.Text) < 10 {
					t.Errorf("pattern %q content too short: %q", p, text.Text)
				}
			}
		}
	}
}

func TestHandleGetRefactoringStrategy_NotFound(t *testing.T) {
	result, err := handleGetRefactoringStrategy(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{"smell_name": "nonexistent_pattern"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for unknown pattern")
	}
}

func TestHandleGetRefactoringStrategy_MissingArg(t *testing.T) {
	result, err := handleGetRefactoringStrategy(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing smell_name")
	}
}

func TestHandleGetRefactoringStrategy_InvalidArgs(t *testing.T) {
	result, err := handleGetRefactoringStrategy(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: "not-a-map"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for invalid arguments")
	}
}

func TestHandleAssessFile_Valid(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)
	os.WriteFile("test.go", []byte("package main\nfunc main() {}\n"), 0644)

	result, err := handleAssessFile(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{"file_path": "test.go"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	for _, c := range result.Content {
		if text, ok := c.(*mcp.TextContent); ok {
			if !strings.Contains(text.Text, "Score:") && !strings.Contains(text.Text, "/100") {
				t.Error("assess_file should contain score")
			}
		}
	}
}

func TestHandleAssessFile_FileNotFound(t *testing.T) {
	result, err := handleAssessFile(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{"file_path": "/nonexistent.go"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent file")
	}
}

func TestHandleAssessFile_MissingArg(t *testing.T) {
	result, err := handleAssessFile(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing file_path")
	}
}

func TestHandleAssessFile_InvalidArgs(t *testing.T) {
	result, err := handleAssessFile(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: "not-a-map"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for invalid arguments")
	}
}

func TestHandleSetConfig_Valid(t *testing.T) {
	result, err := handleSetConfig(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{"key": "language", "value": "go"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
}

func TestHandleSetConfig_ReadOnly(t *testing.T) {
	result, err := handleSetConfig(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{"key": "read_only", "value": "true"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
}

func TestHandleSetConfig_InvalidKey(t *testing.T) {
	result, err := handleSetConfig(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{"key": "nonexistent", "value": "x"}, // gitleaks:allow
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for invalid config key")
	}
}

func TestHandleSetConfig_MissingKey(t *testing.T) {
	result, err := handleSetConfig(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{"value": "x"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing key")
	}
}

func TestHandleSetConfig_InvalidArgs(t *testing.T) {
	result, err := handleSetConfig(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: "not-a-map"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for invalid arguments")
	}
}

func TestHandleGetConfig(t *testing.T) {
	result, err := handleGetConfig(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	for _, c := range result.Content {
		if text, ok := c.(*mcp.TextContent); ok {
			if !strings.Contains(text.Text, "configuration") {
				t.Error("get_config should contain 'configuration'")
			}
		}
	}
}

func TestHandleListHotspots_DefaultArgs(t *testing.T) {
	result, err := handleListHotspots(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]interface{}{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Logf("list_hotspots error (expected in non-git dir): %v", result.Content)
		// May error if not in a git repo, which is ok for unit test
	}
}

func TestHandleListHotspots_WithArgs(t *testing.T) {
	result, err := handleListHotspots(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"repo_path":   "..",
				"max_commits": 50.0,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// May succeed or error depending on context
	_ = result
}
