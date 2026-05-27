package mcp_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func startClient(t *testing.T) (*client.Client, func()) {
	t.Helper()

	c, err := client.NewStdioMCPClient("../../bin/ailinter", nil, "mcp")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "ailinter-test", Version: "1.0.0"}

	_, err = c.Initialize(ctx, initReq)
	if err != nil {
		cancel()
		c.Close()
		t.Fatalf("initialize failed: %v", err)
	}

	cleanup := func() {
		cancel()
		c.Close()
	}

	return c, cleanup
}

func TestMCP_ToolsList(t *testing.T) {
	c, cleanup := startClient(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}

	t.Logf("Registered %d tools:", len(result.Tools))
	for _, tool := range result.Tools {
		t.Logf("  - %s: %s", tool.Name, truncate(tool.Description, 80))
	}

	if len(result.Tools) < 7 {
		t.Errorf("expected at least 7 tools, got %d", len(result.Tools))
	}
}

func TestMCP_AnalyzeCode(t *testing.T) {
	c, cleanup := startClient(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	os.WriteFile("integration_test_file.go", []byte("package main\nfunc main() {}\n"), 0644)
	defer os.Remove("integration_test_file.go")

	result, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "analyze_code",
			Arguments: map[string]interface{}{
				"file_path": "integration_test_file.go",
			},
		},
	})
	if err != nil {
		t.Fatalf("analyze_code: %v", err)
	}
	if result.IsError {
		t.Fatalf("analyze_code error: %v", result.Content)
	}

	printContent(t, "analyze_code", result.Content)
}

func TestMCP_ScanForSecrets(t *testing.T) {
	c, cleanup := startClient(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	content := "package main\nvar AWS_ACCESS_KEY_ID = \"AKIAIOSFODNN7EXAMPLE\"\n"

	result, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "scan_for_secrets",
			Arguments: map[string]interface{}{
				"content": content,
			},
		},
	})
	if err != nil {
		t.Fatalf("scan_for_secrets: %v", err)
	}

	printContent(t, "scan_for_secrets", result.Content)
}

func TestMCP_GetRefactoringStrategy(t *testing.T) {
	c, cleanup := startClient(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_refactoring_strategy",
			Arguments: map[string]interface{}{
				"smell_name": "deep_nesting",
			},
		},
	})
	if err != nil {
		t.Fatalf("get_refactoring_strategy: %v", err)
	}

	printContent(t, "get_refactoring_strategy", result.Content)
}

func TestMCP_AssessFile(t *testing.T) {
	c, cleanup := startClient(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	os.WriteFile("integration_test_assess.go", []byte("package main\nfunc main() {\n\tif true {\n\t\tif true {\n\t\t\tprintln(\"x\")\n\t\t}\n\t}\n}\n"), 0644)
	defer os.Remove("integration_test_assess.go")

	result, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "assess_file",
			Arguments: map[string]interface{}{
				"file_path": "integration_test_assess.go",
			},
		},
	})
	if err != nil {
		t.Fatalf("assess_file: %v", err)
	}

	printContent(t, "assess_file", result.Content)
}

func TestMCP_ConfigGetSet(t *testing.T) {
	c, cleanup := startClient(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set language
	_, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "set_config",
			Arguments: map[string]interface{}{
				"key":   "language",
				"value": "go",
			},
		},
	})
	if err != nil {
		t.Fatalf("set_config: %v", err)
	}

	// Get config
	result, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_config",
		},
	})
	if err != nil {
		t.Fatalf("get_config: %v", err)
	}

	printContent(t, "get_config", result.Content)
}

func TestMCP_ListHotspots(t *testing.T) {
	c, cleanup := startClient(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_hotspots",
			Arguments: map[string]interface{}{
				"repo_path":   ".",
				"max_commits": 100.0,
			},
		},
	})
	if err != nil {
		t.Fatalf("list_hotspots: %v", err)
	}

	printContent(t, "list_hotspots", result.Content)
}

func printContent(t *testing.T, tool string, content []mcp.Content) {
	t.Helper()
	for _, c := range content {
		if text, ok := c.(*mcp.TextContent); ok {
			t.Logf("[%s] %s", tool, truncate(text.Text, 300))
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
