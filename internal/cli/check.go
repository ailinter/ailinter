package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/config"
	"github.com/ailinter/ailinter/internal/secrets"

	"github.com/spf13/cobra"
)

func CheckCommand() *cobra.Command {
	var (
		formatFlag    string
		jsonFlag      bool
		noSecrets     bool
		langOverride  string
		noGitignore   bool
	)

	cmd := &cobra.Command{
		Use:   "check <file|dir>",
		Short: "Analyze files for Code Quality issues and secrets",
		Long: `Analyze source files for structural Code Quality issues (deep nesting, brain methods,
bumpy roads, complex conditionals, etc.) and scan for hardcoded secrets.

Returns a quality score from 0-100 and detailed findings with AI guidance.

By default, respects .gitignore patterns when scanning directories.
Use --no-gitignore to disable this behavior.

Output formats:
  auto      Auto-detect based on terminal (default)
  human     Colored text for terminal display
  json      Structured JSON output
  markdown  Markdown formatted (ideal for LLMs)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			mode := ResolveFormat(formatFlag)
			if jsonFlag {
				mode = FormatJSON
			}

			info, err := os.Stat(target)
			if err != nil {
				return fmt.Errorf("cannot access %s: %w", target, err)
			}

			if info.IsDir() {
				return checkDirectory(target, mode, noSecrets, langOverride, !noGitignore)
			}
			return checkFile(target, mode, noSecrets, langOverride)
		},
	}

	cmd.Flags().StringVar(&formatFlag, "format", "", "Output format: auto, human, json, markdown")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output results as JSON")
	cmd.Flags().Lookup("json").Hidden = true
	cmd.Flags().BoolVar(&noSecrets, "no-secrets", false, "Skip secret scanning")
	cmd.Flags().StringVar(&langOverride, "lang", "", "Force language (auto-detected by default)")
	cmd.Flags().BoolVar(&noGitignore, "no-gitignore", false, "Do not respect .gitignore patterns when scanning directories")

	return cmd
}

func checkFile(path string, format FormatMode, noSecrets bool, langOverride string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	lang := langOverride
	if lang == "" {
		ext := filepath.Ext(path)
		lang = analyzer.DetectedLanguage(ext)
		if lang == "" {
			lang = "go"
		}
	}

	thresholds := config.LoadProjectThresholds(path, lang)
	result := analyzer.Analyze(path, string(data), lang, thresholds)

	if format == FormatJSON {
		writeCombinedJSON(result, data, path)
	} else {
		writeResult(format, result)
		if !noSecrets {
			scanAndWriteSecrets(path, data, format)
		}
	}

	return nil
}

func checkDirectory(dir string, format FormatMode, noSecrets bool, langOverride string, respectGitignore bool) error {
	var allResults []analyzer.QualityResult
	var gitignorePatterns []string

	if respectGitignore {
		gitignorePatterns = loadGitignore(dir)
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isSourceFile(path) {
			return nil
		}

		if len(gitignorePatterns) > 0 && isGitignored(path, dir, gitignorePatterns) {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		lang := langOverride
		if lang == "" {
			ext := filepath.Ext(path)
			lang = analyzer.DetectedLanguage(ext)
			if lang == "" {
				lang = "go" // default language for config files, dotfiles, etc.
			}
		}

		thresholds := config.LoadProjectThresholds(path, lang)
		result := analyzer.Analyze(path, string(data), lang, thresholds)
		allResults = append(allResults, result)

		if !noSecrets {
			scanAndWriteSecrets(path, data, format)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	if format == FormatJSON {
		writeJSONResults(allResults)
	} else {
		writeResults(format, allResults)
		writeSummary(format, allResults)
	}

	return nil
}

func scanAndWriteSecrets(path string, data []byte, format FormatMode) {
	scanner, err := secrets.NewScanner()
	if err != nil {
		return
	}
	findings := scanner.ScanBytes(data, path)
	if format == FormatJSON {
		if len(findings) > 0 {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(findings)
		}
	} else {
		writeSecrets(format, path, findings)
	}
}

func isSourceFile(path string) bool {
	ext := filepath.Ext(path)
	base := filepath.Base(path)
	switch ext {
	case ".go", ".py", ".js", ".ts", ".tsx", ".java", ".rs", ".rb",
		".c", ".cpp", ".h", ".hpp", ".cs", ".swift", ".kt", ".kts", ".scala",
		".php", ".pl", ".sh", ".bash", ".tf", ".yaml", ".yml", ".toml",
		".json", ".xml", ".html", ".css", ".sql",
		".properties", ".ini", ".cfg", ".conf", ".env":
		return true
	}
	// Handle dotfiles without extensions: .env, .env.prod, .gitignore, Dockerfile, etc.
	if base == ".env" || strings.HasPrefix(base, ".env.") || base == "Dockerfile" ||
		strings.HasPrefix(base, "Dockerfile.") || base == "Makefile" ||
		base == ".gitignore" || base == ".gitattributes" || base == ".npmrc" ||
		base == ".editorconfig" || base == ".dockerignore" {
		return true
	}
	return false
}

// loadGitignore reads .gitignore from the scan root and returns patterns.
func loadGitignore(root string) []string {
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return nil
	}
	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "/")
		patterns = append(patterns, line)
	}
	return patterns
}

// isGitignored checks if a path matches any gitignore pattern.
func isGitignored(path, root string, patterns []string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	for _, p := range patterns {
		// Handle simple glob and prefix patterns
		matched, _ := filepath.Match(p, filepath.Base(path))
		if matched {
			return true
		}
		matched, _ = filepath.Match(p, rel)
		if matched {
			return true
		}
		// Check if the relative path starts with a directory pattern (e.g., "bin/")
		if strings.HasSuffix(p, "/") && strings.HasPrefix(rel, p) {
			return true
		}
	}
	return false
}
