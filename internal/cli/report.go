package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/spf13/cobra"
)

func ReportCommand() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "report <file|dir>",
		Short: "Generate a CODE_QUALITY.md report",
		Long: `Run a full analysis on a file or directory and generate a detailed
CODE_QUALITY.md report with score breakdown, detector results, secret
scan summary, vulnerability summary, and all issues in a table.

The report is a clean, commit-worthy markdown file with timestamp
and a link to ailinter.dev.

Flags:
  --output <path>   Output path for the report (default: CODE_QUALITY.md)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeReport(args[0], outputPath)
		},
	}

	cmd.Flags().StringVar(&outputPath, "output", "CODE_QUALITY.md", "Output path for the report")
	return cmd
}

func executeReport(target, outputPath string) error {
	respectGitignore := true

	reportData, err := analyzer.GenerateReport(target, respectGitignore)
	if err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}

	markdown := reportData.RenderMarkdown()

	absOutput, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("cannot resolve output path: %w", err)
	}

	if err := os.WriteFile(absOutput, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("cannot write report: %w", err)
	}

	fmt.Printf("Report written to %s\n", absOutput)
	fmt.Printf("Target: %s  |  Score: %d/100 (%s)\n", target, reportData.OverallScore, reportData.OverallLabel)
	fmt.Printf("Files: %d  |  Issues: %d\n", len(reportData.Results), countAllSmells(reportData.Results))
	if len(reportData.Secrets) > 0 {
		fmt.Printf("Secrets: %d found ⚠️\n", len(reportData.Secrets))
	} else {
		fmt.Println("Secrets: clean ✅")
	}
	if len(reportData.Vulns) > 0 {
		fmt.Printf("Vulnerabilities: %d found ⚠️\n", len(reportData.Vulns))
	} else {
		fmt.Println("Vulnerabilities: clean ✅")
	}

	return nil
}

func countAllSmells(results []analyzer.QualityResult) int {
	n := 0
	for _, r := range results {
		n += len(r.Smells)
	}
	return n
}
