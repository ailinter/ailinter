package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func InitCommand() *cobra.Command {
	var (
		skipAgents  bool
		withVSCode  bool
		agentFlag   string
		setupHook   bool
		profileFlag string
	)
	opts := &setupOpts{}

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
				cmd.Flags().Changed("profile")

			if !hasAnyFlag && isInteractive() {
				return runInteractiveSetup(cwd, false, false)
			}

			opts.skipAgents = skipAgents
			opts.withVSCode = withVSCode
			opts.agentFlag = agentFlag
			opts.setupHook = setupHook
			opts.profileFlag = profileFlag
			return runNonInteractiveSetup(cwd, opts)
		},
	}

	cmd.Flags().BoolVar(&skipAgents, "no-agents", false, "Skip AGENTS.md creation")
	cmd.Flags().BoolVar(&withVSCode, "vscode", false, "Create .vscode/tasks.json + settings for IDE integration")
	cmd.Flags().StringVar(&agentFlag, "agent", "", "AI agent to configure: opencode, claude, cursor, copilot, all")
	cmd.Flags().BoolVar(&setupHook, "hook", false, "Create .githooks/pre-commit for pre-commit scanning")
	cmd.Flags().StringVar(&profileFlag, "profile", "default", "Threshold profile: default, strict, relaxed")
	return cmd
}

type setupOpts struct {
	skipAgents  bool
	withVSCode  bool
	agentFlag   string
	setupHook   bool
	profileFlag string
}

func runNonInteractiveSetup(cwd string, opts *setupOpts) error {
	result := &setupResult{}

	writeConfig(cwd, opts.profileFlag, result)
	writeAgentsMD(cwd, opts.skipAgents, result)

	writeAgentSetups(cwd, opts.agentFlag, result)

	if opts.withVSCode {
		writeVSCodeFiles(cwd, result)
	}
	if opts.setupHook {
		writeGitHook(cwd, result)
	}

	printResult(result)
	fmt.Println("\nailinter initialized! Run 'ailinter check .' to analyze your codebase.")
	return nil
}

func writeAgentSetups(cwd, agentFlag string, result *setupResult) {
	agent := strings.ToLower(agentFlag)
	if agent == "all" {
		for _, a := range allAgents {
			writeAgentFiles(cwd, a, result)
		}
		return
	}
	if kind, ok := parseAgentName(agent); ok {
		writeAgentFiles(cwd, kind, result)
	}
}

func parseAgentName(name string) (agentKind, bool) {
	if canonical, ok := agentAliases[name]; ok {
		return canonical, true
	}
	return "", false
}
