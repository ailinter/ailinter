package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/ailinter/ailinter/internal/cli"
	"github.com/ailinter/ailinter/internal/telemetry"
	"github.com/spf13/cobra"
)

var version = "v0.8.6"

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			version = v
		}
	}
}

func main() {
	telemetry.Version = version
	telemetry.Init(context.Background())
	defer telemetry.Shutdown(context.Background())

	telemetry.RecordInstallation()

	root := newRootCommand()
	root.AddCommand(cli.CheckCommand())
	root.AddCommand(cli.ReportCommand())
	root.AddCommand(cli.MCPCommand(version))
	root.AddCommand(cli.InitCommand())
	root.AddCommand(cli.KnowledgeCommand())
	root.AddCommand(rulesCommand())
	root.AddCommand(telemetryCommand())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ailinter",
		Short: "ailinter — AI Linter & Code Quality for AI-Assisted Development",
		Long: `ailinter is a Code Quality and safety tool for AI-assisted development.

Provides four pillars of protection:
  1. Code Quality Radar — structural analysis (nesting, bumpy roads, complexity)
  2. Secret Scanning — 150+ rules to catch hardcoded credentials
  3. Refactoring Guide — exact patterns to fix detected issues
  4. IaC + Dependency Guard — catch infrastructure misconfigurations and hallucinated packages

Run as a CLI tool or an MCP server for AI assistants.`,
		Version: version,
	}
}

func rulesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage ailinter rules and thresholds",
	}
	cmd.AddCommand(listRulesCommand())
	return cmd
}

func listRulesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active rules for a language",
		RunE:  runListRulesE,
	}
	cmd.Flags().String("lang", "", "Filter rules for a specific language (go, python, javascript, typescript, java, csharp, ruby, swift, kotlin, rust, cpp, c)")
	return cmd
}

func runListRulesE(c *cobra.Command, args []string) error {
	lang, _ := c.Flags().GetString("lang")
	if lang == "" {
		printDefaultRules()
		return nil
	}
	return runListRules(lang)
}

func runListRules(lang string) error {
	if !isValidLanguage(lang) {
		return fmt.Errorf("unknown language: %s (valid: go, python, javascript, typescript, java, csharp, ruby, swift, kotlin, rust, cpp, c)", lang)
	}
	printLanguageRules(lang)
	return nil
}

func printDefaultRules() {
	fmt.Println("Default rules for common languages:")
	fmt.Println()
	fmt.Println("  Metric                  |  Go  | Python | JS/TS | Java")
	fmt.Println("  ----------------------- | ---- | ------ | ----- | ----")
	fmt.Println("  Nesting depth (warn)    |   4  |    4   |   3   |   4")
	fmt.Println("  Cyclomatic complexity   |   9  |    9   |   9   |   9")
	fmt.Println("  Function LOC (warn)     |  80  |   70   |  60   |  70")
	fmt.Println("  File LOC (warn)         | 1000 |  600   | 700   | 600")
	fmt.Println("  Max function arguments  |   4  |    4   |   4   |   5")
	fmt.Println("  Bumpy Road bumps        |   2  |    2   |   2   |   2")
	fmt.Println()
	fmt.Println("Customize thresholds via .ailinter.toml in your project root.")
}

func isValidLanguage(lang string) bool {
	switch lang {
	case "go", "python", "javascript", "typescript", "java", "csharp", "ruby", "swift", "kotlin", "rust", "cpp", "c":
		return true
	}
	return false
}

func telemetryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "telemetry",
		Short: "Show telemetry collection details",
		Long:  "ailinter collects anonymous usage statistics by default. No source code, file paths, or PII is ever collected.",
		Run:   runTelemetry,
	}
}

func runTelemetry(cmd *cobra.Command, args []string) {
	printTelemetryInfo()
}

func printTelemetryInfo() {
	fmt.Println("Metrics collected (all anonymous):")
	fmt.Println()
	fmt.Println("  Metric              | Tags")
	fmt.Println("  ------------------- | ----")
	fmt.Println("  cli.invocations     | command, os, arch, version")
	fmt.Println("  mcp.tool_calls      | tool, version")
	fmt.Println("  files.analyzed      | language, extension, version")
	fmt.Println("  installations       | version")
	fmt.Println("  quality.score       | language (histogram)")
	fmt.Println("  smells.detected     | smell_type, language")
	fmt.Println("  secrets.detected    | category, severity, file_ext")
	fmt.Println("  duration.seconds    | operation, language (histogram)")
	fmt.Println("  errors              | error_type")
	fmt.Println()
	fmt.Println("Opt out: AILINTER_NO_TELEMETRY=1")
	fmt.Println()
	fmt.Println("NOT collected: source code, file paths, IPs, hostnames, env vars, raw secrets, git metadata.")
	fmt.Println()
}

type langRules struct {
	NestingWarn string
	FuncLOCWarn string
	FileLOCWarn string
	MaxArgs     string
}

var languageRules = map[string]langRules{
	"go":         {"4", "80", "1000", "4"},
	"python":     {"4", "70", "600", "4"},
	"javascript": {"3", "60", "700", "4"},
	"typescript": {"3", "60", "700", "4"},
	"java":       {"4", "70", "600", "5"},
}

func printLanguageRules(lang string) {
	fmt.Printf("Rules for %s:\n\n", lang)
	fmt.Println("  Metric                  | Value")
	fmt.Println("  ----------------------- | -----")
	rules, ok := languageRules[lang]
	if !ok {
		rules = langRules{"4", "80", "1000", "4"}
	}
	fmt.Printf("  Nesting depth (warn)    |   %s\n", rules.NestingWarn)
	fmt.Println("  Cyclomatic complexity   |   9")
	fmt.Printf("  Function LOC (warn)     |   %s\n", rules.FuncLOCWarn)
	fmt.Printf("  File LOC (warn)         |   %s\n", rules.FileLOCWarn)
	fmt.Printf("  Max function arguments  |   %s\n", rules.MaxArgs)
	fmt.Println("  Bumpy Road bumps        |   2")
	fmt.Println()
	fmt.Println("Customize thresholds via .ailinter.toml in your project root.")
}
