package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func InitCommand() *cobra.Command {
	var skipAgents bool
	var withVSCode bool
	var agentFlag string
	var setupHook bool
	var profileFlag string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize ailinter in the current project",
		Long: `Bootstrap ailinter in the current directory.

Interactive mode (default when run in a terminal):
  ailinter init

Non-interactive mode (with flags):
  ailinter init --agent opencode --hook --vscode

Creates:
  .ailinter.toml          Configuration with thresholds
  AGENTS.md               AI agent instructions

Per-agent files (via --agent or interactive selection):
  --agent opencode        opencode.json, .opencode/agent/, .opencode/skills/
  --agent claude          .claude/settings.json, CLAUDE.md
  --agent cursor          .cursor/mcp.json, .cursor/rules/
  --agent copilot         .github/copilot-instructions.md
  --agent all             All of the above
  --vscode                .vscode/tasks.json, settings.json, extensions.json
  --hook                  .githooks/pre-commit (configure git: git config core.hooksPath .githooks)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("cannot get working directory: %w", err)
			}

			hasAnyFlag := skipAgents || withVSCode || agentFlag != "" || setupHook ||
				cmd.Flags().Changed("profile") || cmd.Flags().Changed("format")

			if !hasAnyFlag && isInteractive() {
				return runInteractiveSetup(cwd, false, false)
			}

			return runNonInteractiveSetup(cwd, skipAgents, withVSCode, agentFlag, setupHook, profileFlag)
		},
	}

	cmd.Flags().BoolVar(&skipAgents, "no-agents", false, "Skip AGENTS.md creation")
	cmd.Flags().BoolVar(&withVSCode, "vscode", false, "Create .vscode/tasks.json + settings for IDE integration")
	cmd.Flags().StringVar(&agentFlag, "agent", "", "AI agent to configure: opencode, claude, cursor, copilot, all")
	cmd.Flags().BoolVar(&setupHook, "hook", false, "Create .githooks/pre-commit for pre-commit scanning")
	cmd.Flags().StringVar(&profileFlag, "profile", "default", "Threshold profile: default, strict, relaxed")
	return cmd
}

func runNonInteractiveSetup(cwd string, skipAgents, withVSCode bool, agentFlag string, setupHook bool, profileFlag string) error {
	result := &setupResult{}

	writeConfig(cwd, profileFlag, result)
	writeAgentsMD(cwd, skipAgents, result)

	switch strings.ToLower(agentFlag) {
	case "all":
		for _, agent := range allAgentNames() {
			writeAgentFiles(cwd, agent, result)
		}
	case "opencode", "claude", "cursor", "copilot":
		writeAgentFiles(cwd, agentFlag, result)
	}

	if withVSCode {
		writeVSCodeFiles(cwd, result)
	}
	if setupHook {
		writeGitHook(cwd, result)
	}

	printResult(result)
	fmt.Println("\nailinter initialized! Run 'ailinter check .' to analyze your codebase.")
	return nil
}
