package main

import (
	"fmt"
	"os"

	"github.com/ailinter/ailinter/internal/cli"
	"github.com/spf13/cobra"
)

var version = "0.5.0-dev"

func main() {
	root := &cobra.Command{
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

	root.AddCommand(cli.CheckCommand())
	root.AddCommand(cli.MCPCommand())
	root.AddCommand(cli.InitCommand())
	root.AddCommand(rulesCommand())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rulesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage ailinter rules and thresholds",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List active rules for a language",
		RunE: func(c *cobra.Command, args []string) error {
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
			return nil
		},
	})

	return cmd
}
