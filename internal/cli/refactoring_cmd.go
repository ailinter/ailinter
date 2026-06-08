package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ailinter/ailinter/internal/refactoring"
	"github.com/spf13/cobra"
)

// refactoringAliases maps alternative smell names to their canonical pattern files.
var refactoringAliases = map[string]string{
	"long_method":                "brain_method",
	"high_cyclomatic_complexity": "complex_method",
}

func GetRefactoringStrategyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-refactoring-strategy <smell_name>",
		Short: "Get refactoring strategy for a code smell",
		Long: `Returns the refactoring strategy for a given code smell, including
step-by-step instructions, before/after Go code examples, and verification steps.

Use 'get-refactoring-strategy list' to list all available patterns.

Aliases: long_method → brain_method, high_cyclomatic_complexity → complex_method`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			smellName := strings.TrimSpace(args[0])

			// Resolve aliases before lookup
			if alias, ok := refactoringAliases[smellName]; ok {
				smellName = alias
			}

			// Handle list subcommand
			if smellName == "list" {
				patterns := refactoring.ListPatterns()
				sort.Strings(patterns)
				fmt.Println("Available refactoring strategy patterns:")
				fmt.Println()
				for _, name := range patterns {
					fmt.Printf("  - %s\n", name)
				}
				fmt.Println()
				fmt.Println("Aliases:")
				aliasNames := make([]string, 0, len(refactoringAliases))
				for alias := range refactoringAliases {
					aliasNames = append(aliasNames, alias)
				}
				sort.Strings(aliasNames)
				for _, alias := range aliasNames {
					fmt.Printf("  - %s → %s\n", alias, refactoringAliases[alias])
				}
				return nil
			}

			pattern := refactoring.Lookup(smellName)
			if pattern == nil {
				available := refactoring.ListPatterns()
				return fmt.Errorf("no refactoring strategy found for '%s'\nAvailable: %s", smellName, strings.Join(available, ", "))
			}

			fmt.Print(pattern.Content)

			// Ensure output ends with newline
			if !strings.HasSuffix(pattern.Content, "\n") {
				fmt.Println()
			}

			return nil
		},
	}

	return cmd
}

// RunRefactoringStrategy is a convenience wrapper for programmatic use.
func RunRefactoringStrategy(smellName string) {
	pattern := refactoring.Lookup(smellName)
	if pattern == nil {
		fmt.Fprintf(os.Stderr, "no refactoring strategy found for '%s'\n", smellName)
		os.Exit(1)
	}
	fmt.Print(pattern.Content)
	if !strings.HasSuffix(pattern.Content, "\n") {
		fmt.Println()
	}
}
