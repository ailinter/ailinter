package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/config"
	"github.com/ailinter/ailinter/internal/refactoring"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/telemetry"
	"github.com/ailinter/ailinter/internal/vulnerability"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// clientNameMap maps known MCP client names to short identifiers for telemetry.
var clientNameMap = map[string]string{
	"cursor":         "cursor",
	"claude":         "claude",
	"claude desktop": "claude",
	"claude code":    "claude",
	"cline":          "cline",
	"github.copilot": "copilot",
	"github copilot": "copilot",
	"copilot":        "copilot",
	"windsurf":       "windsurf",
	"continue":       "continue",
	"cody":           "cody",
	"goose":          "goose",
	"sourcegraph":    "cody",
}

// normalizeClientName converts a raw MCP client name to a short identifier.
// Falls back to the lowercased name if not in the known mapping.
func normalizeClientName(name string) string {
	if name == "" {
		return "unknown"
	}
	lower := strings.ToLower(name)
	if mapped, ok := clientNameMap[lower]; ok {
		return mapped
	}
	return lower
}

// clientDetectionHook returns an OnBeforeInitialize hook that auto-detects
// the MCP client name from the Initialize handshake and sets it on the
// telemetry package. The AILINTER_MCP_CLIENT env var takes precedence as
// an explicit override.
func clientDetectionHook() server.OnBeforeInitializeFunc {
	return func(ctx context.Context, id any, message *mcp.InitializeRequest) {
		// Env var takes precedence as explicit override
		if c := os.Getenv("AILINTER_MCP_CLIENT"); c != "" {
			telemetry.SetMCPClient(c)
			return
		}
		// Auto-detect from initialize handshake
		clientName := normalizeClientName(message.Params.ClientInfo.Name)
		telemetry.SetMCPClient(clientName)
	}
}

// Serve starts the MCP server on stdio.
func Serve(version string) error {
	s := server.NewMCPServer(
		"ailinter",
		version,
		server.WithToolCapabilities(true),
		server.WithHooks(&server.Hooks{
			OnBeforeInitialize: []server.OnBeforeInitializeFunc{
				clientDetectionHook(),
			},
		}),
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
		mcp.WithDescription("🔧 NEXT STEP after analyze_code or assess_file reports issues. Returns exact, actionable refactoring instructions with before/after examples and verification steps. Supports: deep_nesting, brain_method, bumpy_road, complex_conditional, god_class, long_parameter_list, primitive_obsession, duplicated_code."),
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
	start := time.Now()
	telemetry.RecordMCPToolCall("analyze_code")
	defer func() {
		telemetry.RecordDuration("analyze_code", "", time.Since(start).Seconds())
	}()

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		telemetry.RecordError("mcp_invalid_args")
		return mcp.NewToolResultError("invalid arguments"), nil
	}
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	resolvedPath, err := resolveAndValidatePath(filePath)
	if err != nil {
		telemetry.RecordError("mcp_path_validation")
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		telemetry.RecordError("mcp_file_read")
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	if isBinaryContent(data) {
		telemetry.RecordError("mcp_binary_file")
		return mcp.NewToolResultError("cannot analyze binary file"), nil
	}

	ext := filepath.Ext(resolvedPath)
	lang := analyzer.DetectedLanguage(ext)
	if lang == "" {
		lang = "go"
	}

	thresholds := config.LoadProjectThresholds(resolvedPath, lang)
	result := analyzer.Analyze(analyzer.SourceInput{FilePath: resolvedPath, Source: string(data), Lang: lang}, thresholds)

	telemetry.RecordFileAnalyzed(lang, ext)
	telemetry.RecordQualityScore(lang, result.Score)
	for _, s := range result.Smells {
		telemetry.RecordSmellsDetected(s.Name, lang, 1)
	}

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
	start := time.Now()
	telemetry.RecordMCPToolCall("scan_for_secrets")
	defer func() {
		telemetry.RecordDuration("scan_for_secrets", "", time.Since(start).Seconds())
	}()

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		telemetry.RecordError("mcp_invalid_args")
		return mcp.NewToolResultError("invalid arguments"), nil
	}
	content, ok := args["content"].(string)
	if !ok || content == "" {
		return mcp.NewToolResultError("content is required"), nil
	}

	scanner, err := secrets.NewScanner()
	if err != nil {
		telemetry.RecordError("mcp_secret_scanner_init")
		return mcp.NewToolResultError(fmt.Sprintf("secret scanner init failed: %v", err)), nil
	}

	findings := scanner.ScanString(content, "<inline>")
	telemetry.RecordSecretsDetected("", "", "", len(findings))
	return mcp.NewToolResultStructuredOnly(findings), nil
}

func handleGetRefactoringStrategy(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	telemetry.RecordMCPToolCall("get_refactoring_strategy")
	defer func() {
		telemetry.RecordDuration("get_refactoring_strategy", "", time.Since(start).Seconds())
	}()

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		telemetry.RecordError("mcp_invalid_args")
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
	start := time.Now()
	telemetry.RecordMCPToolCall("assess_file")
	defer func() {
		telemetry.RecordDuration("assess_file", "", time.Since(start).Seconds())
	}()

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		telemetry.RecordError("mcp_invalid_args")
		return mcp.NewToolResultError("invalid arguments"), nil
	}
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return mcp.NewToolResultError("file_path is required"), nil
	}

	resolvedPath, err := resolveAndValidatePath(filePath)
	if err != nil {
		telemetry.RecordError("mcp_path_validation")
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		telemetry.RecordError("mcp_file_read")
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	if isBinaryContent(data) {
		telemetry.RecordError("mcp_binary_file")
		return mcp.NewToolResultError("cannot analyze binary file"), nil
	}

	ext := filepath.Ext(resolvedPath)
	lang := analyzer.DetectedLanguage(ext)
	if lang == "" {
		lang = "go"
	}

	thresholds := config.LoadProjectThresholds(resolvedPath, lang)
	result := analyzer.Analyze(analyzer.SourceInput{FilePath: resolvedPath, Source: string(data), Lang: lang}, thresholds)

	telemetry.RecordFileAnalyzed(lang, ext)
	telemetry.RecordQualityScore(lang, result.Score)

	return mcp.NewToolResultText(buildAssessmentSummary(result)), nil
}

// smellRefactoringMap maps detected smell names to their refactoring strategy names.
func smellRefactoringMap() map[string]string {
	return map[string]string{
		"deep_nesting":               "deep_nesting",
		"brain_method":               "brain_method",
		"brain_class":                "god_class",
		"bumpy_road":                 "bumpy_road",
		"complex_conditional":        "complex_conditional",
		"complex_method":             "complex_method",
		"god_class":                  "god_class",
		"long_parameter_list":        "long_parameter_list",
		"primitive_obsession":        "primitive_obsession",
		"duplicated_code":            "duplicated_code",
		"message_chains":             "message_chains",
		"file_bloat":                 "file_bloat",
		"global_data":                "global_data",
		"lazy_element":               "lazy_element",
		"paragraph_of_code":          "paragraph_of_code",
		"excessive_comments":         "excessive_comments",
		"long_scope_variable":        "long_scope_variable",
		"long_switch":                "long_switch",
		"magic_number":               "magic_number",
		"low_cohesion":               "low_cohesion",
		"data_class":                 "data_class",
		"refused_bequest":            "refused_bequest",
		"shotgun_surgery":            "shotgun_surgery",
		"parallel_inheritance":       "parallel_inheritance",
		"long_method":                "brain_method",
		"high_cyclomatic_complexity": "complex_method",
	}
}

func buildRefactoringSteps(smellNames []string) string {
	if len(smellNames) == 0 {
		return ""
	}
	smellCalls := make(map[string]bool)
	for _, name := range smellNames {
		if strategy, ok := smellRefactoringMap()[name]; ok {
			smellCalls[strategy] = true
		}
	}
	if len(smellCalls) == 0 {
		return ""
	}
	result := "\n\nREFACTORING STEPS: Call get_refactoring_strategy() for each detected smell:"
	strategies := make([]string, 0, len(smellCalls))
	for s := range smellCalls {
		strategies = append(strategies, s)
	}
	sort.Strings(strategies)
	for _, s := range strategies {
		result += fmt.Sprintf("\n  - get_refactoring_strategy(%q)", s)
	}
	result += "\n  Then refactor in 3-5 small steps, re-running analyze_code after each to verify improvement."
	return result
}

func buildAssessmentSummary(result analyzer.QualityResult) string {
	summary := fmt.Sprintf("%s — Score: %d/100", result.Label, result.Score)
	if len(result.Smells) > 0 {
		summary += fmt.Sprintf("\nDetected %d issues:", len(result.Smells))
		smellNames := make([]string, len(result.Smells))
		for i, s := range result.Smells {
			summary += fmt.Sprintf("\n  - %s (%s): %s", s.Name, s.Severity, s.Message)
			smellNames[i] = s.Name
		}
		summary += buildRefactoringSteps(smellNames)
	}
	rec := assessmentRecommendation(result.Label, len(result.Smells))
	if rec != "" {
		summary += "\n\n" + rec
	}
	return summary
}

func assessmentRecommendation(label string, smellCount int) string {
	switch label {
	case analyzer.LabelStopRefactor:
		return stopRefactorRec(smellCount > 0)
	case analyzer.LabelNeedsWork:
		return needsWorkRec(smellCount > 0)
	case analyzer.LabelProceedWithCare:
		return "RECOMMENDATION: Proceed with Care — use guard clauses and small isolated changes. Re-check after each edit with analyze_code."
	default:
		return ""
	}
}

func stopRefactorRec(hasSmells bool) string {
	if hasSmells {
		return "RECOMMENDATION: STOP — refactor before any AI modification. Call get_refactoring_strategy() for the specific smell names above. Do NOT modify this file until the score is above 80."
	}
	return "RECOMMENDATION: STOP — refactor before any AI modification. The file is in a risky state and cannot be safely edited as-is."
}

func needsWorkRec(hasSmells bool) string {
	if hasSmells {
		return "RECOMMENDATION: Needs Work — significant complexity issues. Use get_refactoring_strategy() for the smells above. Refactor incrementally in 3-5 small steps, re-running analyze_code after each."
	}
	return "RECOMMENDATION: Needs Work — significant complexity issues. Refactor incrementally in 3-5 small steps before adding features."
}

func handleSetConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	telemetry.RecordMCPToolCall("set_config")
	defer func() {
		telemetry.RecordDuration("set_config", "", time.Since(start).Seconds())
	}()

	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		telemetry.RecordError("mcp_invalid_args")
		return mcp.NewToolResultError("invalid arguments"), nil
	}
	key, _ := args["key"].(string)
	value, _ := args["value"].(string)
	if key == "" {
		return mcp.NewToolResultError("key is required"), nil
	}
	cfg, err := config.SetAndGet(key, value)
	if err != nil {
		telemetry.RecordError("mcp_set_config")
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultStructured(cfg, fmt.Sprintf("Configuration updated:\n%s", mustMarshalIndent(cfg))), nil
}

func handleGetConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	telemetry.RecordMCPToolCall("get_config")
	defer func() {
		telemetry.RecordDuration("get_config", "", time.Since(start).Seconds())
	}()

	cfg, err := config.GetConfig()
	if err != nil {
		telemetry.RecordError("mcp_get_config")
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultStructured(cfg, fmt.Sprintf("Current configuration:\n%s", mustMarshalIndent(cfg))), nil
}

func handleListHotspots(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()
	telemetry.RecordMCPToolCall("list_hotspots")
	defer func() {
		telemetry.RecordDuration("list_hotspots", "", time.Since(start).Seconds())
	}()

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
		telemetry.RecordError("mcp_list_hotspots")
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
