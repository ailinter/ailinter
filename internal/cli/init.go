package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func InitCommand() *cobra.Command {
	var skipAgents bool
	var withVSCode bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize ailinter in the current project",
		Long: `Bootstrap ailinter in the current directory by creating:
  - .ailinter.toml (configuration with language-appropriate defaults)
  - AGENTS.md (AI agent instructions for this project)
  - --vscode: also creates .vscode/tasks.json for IDE integration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("cannot get working directory: %w", err)
			}

			configPath := filepath.Join(cwd, ".ailinter.toml")
			if _, err := os.Stat(configPath); err == nil {
				fmt.Printf(".ailinter.toml already exists at %s\n", configPath)
			} else {
				os.WriteFile(configPath, []byte(defaultConfig), 0644)
				fmt.Println("Created .ailinter.toml")
			}

			if !skipAgents {
				agentsPath := filepath.Join(cwd, "AGENTS.md")
				if _, err := os.Stat(agentsPath); err == nil {
					fmt.Printf("AGENTS.md already exists at %s (skipping)\n", agentsPath)
				} else {
					os.WriteFile(agentsPath, []byte(defaultAgentsMD), 0644)
					fmt.Println("Created AGENTS.md")
				}
			}

			if withVSCode {
				vscodeDir := filepath.Join(cwd, ".vscode")
				os.MkdirAll(vscodeDir, 0755)
				tasksPath := filepath.Join(vscodeDir, "tasks.json")
				if _, err := os.Stat(tasksPath); err == nil {
					fmt.Printf(".vscode/tasks.json already exists (skipping)\n")
				} else {
					os.WriteFile(tasksPath, []byte(defaultVSCodeTasks), 0644)
					fmt.Println("Created .vscode/tasks.json")
				}
			}

			fmt.Println("\nailinter initialized! Run 'ailinter check .' to analyze your codebase.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&skipAgents, "no-agents", false, "Skip AGENTS.md creation")
	cmd.Flags().BoolVar(&withVSCode, "vscode", false, "Also create .vscode/tasks.json for IDE problem integration")
	return cmd
}
