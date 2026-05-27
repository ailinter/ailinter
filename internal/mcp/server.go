package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/config"
	"github.com/ailinter/ailinter/internal/refactoring"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/vulnerability"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Serve starts the MCP server on stdio.
func Serve(version string) error {
	s := server.NewMCPServer(
		"ailinter",
		version,
		server.WithToolCapabilities(true),
	)

	// Tool 1: analyze_code
	s.AddTool(mcp.NewTool(
		"analyze_code",
		mcp.WithDescription("Analyze a source file for Code Quality issues (complexity, nesting, size, bumpy roads) and security vulnerabilities (injection, XSS, deserialization, weak crypto, XXE). Returns a quality score (0-100) and detailed findings."),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Absolute or relative path to the source file to analyze"),
		),
	), handleAnalyzeCode)

	// Tool 2: scan_for_secrets
	s.AddTool(mcp.NewTool(
		"scan_for_secrets",
		mcp.WithDescription("Scan source code content for hardcoded secrets: API keys, tokens, passwords, private keys. Uses 150+ detection rules."),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The source code text to scan for secrets"),
		),
	), handleScanForSecrets)

	// Tool 3: get_refactoring_strategy
	s.AddTool(mcp.NewTool(
		"get_refactoring_strategy",
		mcp.WithDescription("Get exact step-by-step refactoring instructions for a specific code smell. Includes before/after examples and verification steps."),
		mcp.WithString("smell_name",
			mcp.Required(),
			mcp.Description("The code smell to get a refactoring strategy for (e.g., deep_nesting, brain_method, bumpy_road, complex_conditional, god_class, long_parameter_list, primitive_obsession, duplicated_code)"),
		),
	), handleGetRefactoringStrategy)

	// Tool 4: assess_file
	s.AddTool(mcp.NewTool(
		"assess_file",
		mcp.WithDescription("Quick assessment of whether a file is safe for AI modification. Returns 'Go Ahead', 'Proceed with Care', or 'Stop & Refactor' with a summary."),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file to assess"),
		),
	), handleAssessFile)

	// Tool 5: set_config (Phase 4)
	s.AddTool(mcp.NewTool(
		"set_config",
		mcp.WithDescription("Set an ailinter configuration value. Valid keys: access_token, onprem_url, default_path, language, repo_path, enabled_tools, read_only, disable_git."),
		mcp.WithString("key", mcp.Required(), mcp.Description("Configuration key to set")),
		mcp.WithString("value", mcp.Required(), mcp.Description("Value to set")),
	), handleSetConfig)

	// Tool 6: get_config (Phase 4)
	s.AddTool(mcp.NewTool(
		"get_config",
		mcp.WithDescription("View current ailinter configuration."),
	), handleGetConfig)

	// Tool 7: list_hotspots
	s.AddTool(mcp.NewTool(
		"list_hotspots",
		mcp.WithDescription("List frequently-changed files with low quality scores. Requires the repo to be a git repository."),
		mcp.WithString("repo_path", mcp.Description("Path to the git repository (defaults to current directory)")),
		mcp.WithNumber("max_commits", mcp.Description("Maximum commits to scan (default: 500)")),
	), handleListHotspots)

	return server.ServeStdio(s)
}

func handleAnalyzeCode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments"), nil
	}
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	resolvedPath, err := resolveAndValidatePath(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	if isBinaryContent(data) {
		return mcp.NewToolResultError("cannot analyze binary file"), nil
	}

	ext := filepath.Ext(resolvedPath)
	lang := analyzer.DetectedLanguage(ext)
	if lang == "" {
		lang = "go"
	}

	thresholds := config.LoadProjectThresholds(resolvedPath, lang)
	result := analyzer.Analyze(resolvedPath, string(data), lang, thresholds)

	vulnScanner := vulnerability.NewScanner()
	vulnFindings := vulnScanner.Scan(string(data), resolvedPath)

	combined := struct {
		analyzer.QualityResult
		VulnerabilityScan []vulnerability.Finding `json:"vulnerability_scan,omitempty"`
	}{
		QualityResult:     result,
		VulnerabilityScan: vulnFindings,
	}

	return mcp.NewToolResultStructuredOnly(combined), nil
}

func handleScanForSecrets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments"), nil
	}
	content, ok := args["content"].(string)
	if !ok || content == "" {
		return mcp.NewToolResultError("content is required"), nil
	}

	scanner, err := secrets.NewScanner()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("secret scanner init failed: %v", err)), nil
	}

	findings := scanner.ScanString(content, "<inline>")
	return mcp.NewToolResultStructuredOnly(findings), nil
}

func handleGetRefactoringStrategy(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments"), nil
	}
	smellName, ok := args["smell_name"].(string)
	if !ok || smellName == "" {
		return mcp.NewToolResultError("smell_name is required"), nil
	}

	pattern := refactoring.Lookup(smellName)
	if pattern == nil {
		available := refactoring.ListPatterns()
		return mcp.NewToolResultError(fmt.Sprintf("no pattern found for '%s'. Available: %v", smellName, available)), nil
	}

	return mcp.NewToolResultText(pattern.Content), nil
}

func handleAssessFile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments"), nil
	}
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	resolvedPath, err := resolveAndValidatePath(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	if isBinaryContent(data) {
		return mcp.NewToolResultError("cannot analyze binary file"), nil
	}

	ext := filepath.Ext(resolvedPath)
	lang := analyzer.DetectedLanguage(ext)
	if lang == "" {
		lang = "go"
	}

	thresholds := config.LoadProjectThresholds(resolvedPath, lang)
	result := analyzer.Analyze(resolvedPath, string(data), lang, thresholds)

	return mcp.NewToolResultText(buildAssessmentSummary(result)), nil
}

func buildAssessmentSummary(result analyzer.QualityResult) string {
	summary := fmt.Sprintf("%s — Score: %d/100", result.Label, result.Score)
	if len(result.Smells) > 0 {
		summary += fmt.Sprintf("\nDetected %d issues:", len(result.Smells))
		for _, s := range result.Smells {
			summary += fmt.Sprintf("\n  - %s (%s): %s", s.Name, s.Severity, s.Message)
		}
	}
	rec := assessmentRecommendation(result.Label)
	if rec != "" {
		summary += "\n\n" + rec
	}
	return summary
}

func assessmentRecommendation(label string) string {
	switch label {
	case analyzer.LabelStopRefactor:
		return "RECOMMENDATION: Stop & Refactor before AI modification. Run get_refactoring_strategy() for detected issues."
	case analyzer.LabelProceedWithCare:
		return "RECOMMENDATION: Proceed with Care — use guard clauses and small isolated changes. Re-check after each edit."
	default:
		return ""
	}
}

func handleSetConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments"), nil
	}
	key, _ := args["key"].(string)
	value, _ := args["value"].(string)
	if key == "" {
		return mcp.NewToolResultError("key is required"), nil
	}
	cfg, err := config.SetAndGet(key, value)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultStructured(cfg, fmt.Sprintf("Configuration updated:\n%s", mustMarshalIndent(cfg))), nil
}

func handleGetConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultStructured(cfg, fmt.Sprintf("Current configuration:\n%s", mustMarshalIndent(cfg))), nil
}

func handleListHotspots(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]interface{})
	repoPath, _ := args["repo_path"].(string)
	if repoPath == "" {
		repoPath = "."
	}
	maxCommits := 500
	if v, ok := args["max_commits"].(float64); ok && v > 0 {
		maxCommits = int(v)
	}

	result := analyzer.AnalyzeGitHotspots(repoPath, maxCommits)
	if result.Error != "" {
		return mcp.NewToolResultError(result.Error), nil
	}

	entries := result.Entries[:min(20, len(result.Entries))]
	fallback := fmt.Sprintf("Frequently-changed files in %s (%d files analyzed, showing top 20):\n", repoPath, len(result.Entries))
	output, _ := json.MarshalIndent(entries, "", "  ")
	return mcp.NewToolResultStructured(entries, fallback+string(output)), nil
}

func resolveAndValidatePath(filePath string) (string, error) {
	cleaned := filepath.Clean(filePath)
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine working directory: %w", err)
	}
	resolvedCwd, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		return "", fmt.Errorf("cannot resolve working directory: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}
	prefix := resolvedCwd + string(os.PathSeparator)
	if resolved != resolvedCwd && !strings.HasPrefix(resolved, prefix) {
		return "", fmt.Errorf("access denied: path '%s' is outside the working directory", filePath)
	}
	return resolved, nil
}

func isBinaryContent(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	checkLen := 8000
	if len(data) < checkLen {
		checkLen = len(data)
	}
	for _, b := range data[:checkLen] {
		if b == 0 {
			return true
		}
	}
	return false
}

func mustMarshalIndent(v any) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}
