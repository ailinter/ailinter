package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/config"
	"github.com/ailinter/ailinter/internal/metalinter"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/telemetry"
	"github.com/ailinter/ailinter/internal/vulnerability"

	"github.com/spf13/cobra"
)

type checkOptions struct {
	format              FormatMode
	noSecrets           bool
	noVulnerabilities   bool
	secretsOnly         bool
	vulnerabilitiesOnly bool
	langOverride        string
	estimateTokens      bool
	metaLint            bool
}

func CheckCommand() *cobra.Command {
	var (
		formatFlag          string
		jsonFlag            bool
		noSecrets           bool
		noVulnerabilities   bool
		secretsOnly         bool
		vulnerabilitiesOnly bool
		langOverride        string
		noGitignore         bool
		estimateTokens      bool
		metaLint            bool
		noMetaLint          bool
	)

	cmd := &cobra.Command{
		Use:   "check <file|dir>",
		Short: "Analyze files for Code Quality, secrets, and vulnerabilities",
		Long: `Analyze source files for structural Code Quality issues (deep nesting, brain methods,
bumpy roads, complex conditionals, etc.), scan for hardcoded secrets, and
detect security vulnerabilities (injection, XSS, deserialization, weak crypto, XXE).

Returns a quality score from 0-100 and detailed findings with AI guidance.

By default, respects .gitignore patterns when scanning directories.
Use --no-gitignore to disable this behavior.

Output formats:
  auto      Auto-detect based on terminal (default)
  human     Colored text for terminal display
  json      Structured JSON output
  markdown  Markdown formatted (ideal for LLMs)
  problems  GCC-style output for IDE problem matchers (VS Code)

Targeted scans:
  --secrets-only         Scan ONLY for secrets (skip code quality and vulnerabilities)
  --vulnerabilities-only Scan ONLY for vulnerabilities (skip code quality and secrets)

Token estimation (--estimate-tokens):
  After quality analysis, estimates AI token costs and potential savings
  from refactoring to score 80+, with per-interaction and monthly
  enterprise savings across Claude Opus 4, GPT-4.5, and Claude 3.5 Sonnet.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mode, err := ResolveFormatStrict(formatFlag)
			if err != nil {
				return err
			}
			if jsonFlag {
				mode = FormatJSON
			}
			if langOverride != "" && !isValidLanguageName(langOverride) {
				return fmt.Errorf("unknown language: %s (valid: go, python, javascript, typescript, java, csharp, ruby, swift, kotlin, rust, cpp, c)", langOverride)
			}
			opts := checkOptions{
				format:              mode,
				noSecrets:           noSecrets,
				noVulnerabilities:   noVulnerabilities,
				secretsOnly:         secretsOnly,
				vulnerabilitiesOnly: vulnerabilitiesOnly,
				langOverride:        langOverride,
				estimateTokens:      estimateTokens,
				metaLint:            metaLint && !noMetaLint,
			}
			return executeCheck(args[0], opts, !noGitignore)
		},
	}

	cmd.Flags().StringVar(&formatFlag, "format", "", "Output format: auto, human, json, markdown, problems")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output results as JSON")
	cmd.Flags().Lookup("json").Hidden = true
	cmd.Flags().BoolVar(&noSecrets, "no-secrets", false, "Skip secret scanning")
	cmd.Flags().BoolVar(&noVulnerabilities, "no-vulnerabilities", false, "Skip vulnerability scanning")
	cmd.Flags().BoolVar(&secretsOnly, "secrets-only", false, "Scan ONLY for secrets (skip code quality and vulnerabilities)")
	cmd.Flags().BoolVar(&vulnerabilitiesOnly, "vulnerabilities-only", false, "Scan ONLY for vulnerabilities (skip code quality and secrets)")
	cmd.Flags().StringVar(&langOverride, "lang", "", "Force language (auto-detected by default)")
	cmd.Flags().BoolVar(&noGitignore, "no-gitignore", false, "Do not respect .gitignore patterns when scanning directories")
	cmd.Flags().BoolVar(&estimateTokens, "estimate-tokens", false, "Show AI token cost estimation after analysis")
	cmd.Flags().BoolVar(&metaLint, "meta-lint", true, "Run embedded meta-linters (go vet, staticcheck, gofmt, misspell, ineffassign) [default: on]")
	cmd.Flags().BoolVar(&noMetaLint, "no-meta-lint", false, "Skip embedded meta-linters")

	return cmd
}

func executeCheck(target string, opts checkOptions, respectGitignore bool) error {
	flags := map[string]string{}
	if opts.noSecrets {
		flags["no-secrets"] = "true"
	}
	if opts.noVulnerabilities {
		flags["no-vulnerabilities"] = "true"
	}
	if opts.secretsOnly {
		flags["secrets-only"] = "true"
	}
	if opts.vulnerabilitiesOnly {
		flags["vulnerabilities-only"] = "true"
	}
	if opts.format != FormatAuto {
		flags["format"] = opts.format.String()
	}
	if opts.langOverride != "" {
		flags["lang"] = opts.langOverride
	}
	if opts.estimateTokens {
		flags["estimate-tokens"] = "true"
	}
	if !opts.metaLint {
		flags["no-meta-lint"] = "true"
	}
	telemetry.RecordCLIInvocationWithFlags("check", flags)

	info, err := os.Stat(target)
	if err != nil {
		telemetry.RecordError("file_access")
		return fmt.Errorf("cannot access %s: %w", target, err)
	}

	if info.IsDir() {
		return checkDirectory(target, opts, respectGitignore)
	}
	return checkFile(target, opts)
}

func (opts checkOptions) detectLang(path string) string {
	if opts.langOverride != "" {
		return opts.langOverride
	}
	ext := filepath.Ext(path)
	lang := analyzer.DetectedLanguage(ext)
	if lang == "" {
		lang = "go"
	}
	return lang
}

func checkFile(path string, opts checkOptions) error {
	start := time.Now()
	resolved, err := resolveSafePath(path)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		telemetry.RecordError("file_read")
		return fmt.Errorf("failed to read %s: %w", resolved, err)
	}

	if isBinary(data) {
		telemetry.RecordError("binary_file")
		return fmt.Errorf("cannot analyze binary file: %s", resolved)
	}

	if opts.secretsOnly {
		scanAndWriteSecrets(resolved, data, opts.format)
		telemetry.RecordDuration("check_file", "", time.Since(start).Seconds())
		return nil
	}
	if opts.vulnerabilitiesOnly {
		vulnScanner := vulnerability.NewScanner()
		vulnFindings := vulnScanner.Scan(string(data), resolved)
		writeVulnerabilities(opts.format, resolved, vulnFindings)
		telemetry.RecordDuration("check_file", "", time.Since(start).Seconds())
		return nil
	}

	lang := opts.detectLang(resolved)
	ext := filepath.Ext(resolved)
	thresholds := config.LoadProjectThresholds(resolved, lang)
	result := analyzer.Analyze(analyzer.SourceInput{FilePath: resolved, Source: string(data), Lang: lang}, thresholds)

	telemetry.RecordFileAnalyzed(lang, ext)
	telemetry.RecordQualityScore(lang, result.Score)
	for _, s := range result.Smells {
		telemetry.RecordSmellsDetected(s.Name, lang, 1)
	}

	if opts.format == FormatJSON {
		writeCombinedJSON(result, data, resolved, opts.noSecrets, opts.noVulnerabilities)
		telemetry.RecordDuration("check_file", lang, time.Since(start).Seconds())
		return nil
	}

	writeResult(opts.format, result)
	if !opts.noSecrets {
		scanAndWriteSecrets(resolved, data, opts.format)
	}
	if !opts.noVulnerabilities {
		vulnScanner := vulnerability.NewScanner()
		vulnFindings := vulnScanner.Scan(string(data), resolved)
		writeVulnerabilities(opts.format, resolved, vulnFindings)
	}
	if opts.metaLint && ext == ".go" {
		mlFindings, err := metalinter.LintGo([]string{resolved})
		if err == nil && len(mlFindings) > 0 {
			writeMetaLintFindings(opts.format, mlFindings)
		}
	}

	if opts.estimateTokens {
		estimator := analyzer.NewTokenEstimator(resolved, result.Score)
		fmt.Print("\n" + estimator.FormatEstimateOutput())
	}

	telemetry.RecordDuration("check_file", lang, time.Since(start).Seconds())
	return nil
}

func checkDirectory(dir string, opts checkOptions, respectGitignore bool) error {
	resolvedDir, err := resolveSafePath(dir)
	if err != nil {
		return err
	}

	scanQuality := !opts.secretsOnly && !opts.vulnerabilitiesOnly
	scanSecrets := !opts.noSecrets || opts.secretsOnly
	scanVulns := !opts.noVulnerabilities || opts.vulnerabilitiesOnly

	ctx := &walkContext{
		opts:          opts,
		resolvedDir:   resolvedDir,
		gitignorePats: nil,
		scanQuality:   scanQuality,
		scanner:       nil,
		vulnScanner:   nil,
		langCount:     make(map[string]int),
	}
	if respectGitignore {
		ctx.gitignorePats = loadGitignore(resolvedDir)
	}
	if scanSecrets {
		ctx.scanner, _ = secrets.NewScanner()
	}
	if scanVulns {
		ctx.vulnScanner = vulnerability.NewScanner()
	}

	err = filepath.WalkDir(resolvedDir, ctx.walkFn)
	if err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	telemetry.RecordDirScan(ctx.fileCount, ctx.langCount)
	ctx.writeResults()
	if opts.metaLint {
		// Run meta-linters on the directory (Go mode)
		mlFindings, err := metalinter.LintGo([]string{resolvedDir})
		if err == nil && len(mlFindings) > 0 {
			writeMetaLintFindings(opts.format, mlFindings)
		}
	}
	return nil
}

type walkContext struct {
	opts          checkOptions
	resolvedDir   string
	gitignorePats []string
	scanQuality   bool
	scanner       *secrets.Scanner
	vulnScanner   *vulnerability.Scanner
	allResults    []analyzer.QualityResult
	allSecrets    []secrets.SecretFinding
	allVulns      []vulnerability.Finding
	fileCount     int
	langCount     map[string]int
}

func (ctx *walkContext) walkFn(path string, d os.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() {
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" {
			return filepath.SkipDir
		}
		return nil
	}
	if ctx.shouldSkipFile(path) {
		return nil
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return nil
	}
	if isBinary(data) {
		return nil
	}

	if ctx.scanQuality {
		lang := ctx.opts.detectLang(path)
		ctx.fileCount++
		ctx.langCount[lang]++
		thresholds := config.LoadProjectThresholds(path, lang)
		result := analyzer.Analyze(analyzer.SourceInput{FilePath: path, Source: string(data), Lang: lang}, thresholds)
		ctx.allResults = append(ctx.allResults, result)
	}

	if ctx.scanner != nil {
		findings := ctx.scanner.ScanBytes(data, path)
		ctx.allSecrets = append(ctx.allSecrets, findings...)
	}
	if ctx.vulnScanner != nil {
		vulnFindings := ctx.vulnScanner.Scan(string(data), path)
		ctx.allVulns = append(ctx.allVulns, vulnFindings...)
	}
	return nil
}

func (ctx *walkContext) shouldSkipFile(path string) bool {
	return !isSourceFile(path) ||
		(len(ctx.gitignorePats) > 0 && isGitignored(path, ctx.resolvedDir, ctx.gitignorePats))
}

func (ctx *walkContext) writeResults() {
	if ctx.opts.secretsOnly {
		if len(ctx.allSecrets) > 0 {
			writeSecrets(ctx.opts.format, "<directory>", ctx.allSecrets)
		}
		return
	}
	if ctx.opts.vulnerabilitiesOnly {
		if len(ctx.allVulns) > 0 {
			writeVulnerabilities(ctx.opts.format, "<directory>", ctx.allVulns)
		}
		return
	}
	if ctx.opts.format == FormatJSON {
		writeCombinedDirJSON(ctx.allResults, ctx.allSecrets, ctx.allVulns)
		return
	}
	writeResults(ctx.opts.format, ctx.allResults)
	if len(ctx.allSecrets) > 0 {
		writeSecrets(ctx.opts.format, "<directory>", ctx.allSecrets)
	}
	if len(ctx.allVulns) > 0 {
		writeVulnerabilities(ctx.opts.format, "<directory>", ctx.allVulns)
	}
	writeSummary(ctx.opts.format, ctx.allResults)
	if ctx.opts.estimateTokens {
		writeDirTokenEstimates(ctx.allResults)
	}
}

func scanAndWriteSecrets(path string, data []byte, format FormatMode) {
	scanner, err := secrets.NewScanner()
	if err != nil {
		telemetry.RecordError("secret_scanner_init")
		return
	}
	findings := scanner.ScanBytes(data, path)
	telemetry.RecordSecretsDetected("", "", filepath.Ext(path), len(findings))
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

var sourceExts = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
	".java": true, ".rs": true, ".rb": true, ".c": true, ".cpp": true,
	".h": true, ".hpp": true, ".cs": true, ".swift": true, ".kt": true,
	".kts": true, ".scala": true, ".php": true, ".pl": true, ".sh": true,
	".bash": true, ".tf": true, ".yaml": true, ".yml": true, ".toml": true,
	".json": true, ".xml": true, ".html": true, ".css": true, ".sql": true,
	".properties": true, ".ini": true, ".cfg": true, ".conf": true, ".env": true,
}

var sourceBases = map[string]bool{
	".env": true, "Dockerfile": true, "Makefile": true,
	".gitignore": true, ".gitattributes": true, ".npmrc": true,
	".editorconfig": true, ".dockerignore": true,
}

func isSourceFile(path string) bool {
	if sourceExts[filepath.Ext(path)] {
		return true
	}
	base := filepath.Base(path)
	if sourceBases[base] {
		return true
	}
	if strings.HasPrefix(base, ".env.") || strings.HasPrefix(base, "Dockerfile.") {
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
		matched, _ := filepath.Match(p, filepath.Base(path))
		if matched {
			return true
		}
		matched, _ = filepath.Match(p, rel)
		if matched {
			return true
		}
		if strings.HasSuffix(p, "/") && strings.HasPrefix(rel, p) {
			return true
		}
	}
	return false
}

func resolveSafePath(path string) (string, error) {
	cleaned := filepath.Clean(path)
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path %s: %w", path, err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}
	return resolved, nil
}

func isBinary(data []byte) bool {
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

func isValidLanguageName(lang string) bool {
	switch lang {
	case "go", "python", "javascript", "typescript", "java", "csharp", "ruby", "swift", "kotlin", "rust", "cpp", "c":
		return true
	}
	return false
}

func writeDirTokenEstimates(results []analyzer.QualityResult) {
	var totalCurrent, totalAfter int64
	for _, r := range results {
		est := analyzer.NewTokenEstimator(r.FilePath, r.Score)
		totalCurrent += int64(est.CurrentTokens())
		totalAfter += int64(est.EstimatedTokensAfterRefactor())
	}
	savings := totalCurrent - totalAfter

	fmt.Println()
	fmt.Println("Token Savings Estimate (all files)")
	fmt.Println("────────────────────────────────────")
	fmt.Printf("Files scanned: %d\n", len(results))
	fmt.Printf("Total current tokens: %d\n", totalCurrent)
	fmt.Printf("After refactoring (est): %d\n", totalAfter)
	fmt.Printf("Total tokens saved: %d\n\n", savings)

	fmt.Println("Monthly (20 devs, 50 AI calls/day, 6 files read per call):")
	models := []struct {
		name string
		cost float64
	}{
		{"Claude Opus 4", 15.00},
		{"GPT-4.5", 10.00},
		{"Claude 3.5 Sonnet", 3.00},
	}
	for _, m := range models {
		monthly := float64(savings) / 1_000_000.0 * m.cost * 20.0 * 50.0 * 22.0 * 6.0
		fmt.Printf("  %-20s $%.0f/month saved\n", m.name+":", monthly)
	}
}
