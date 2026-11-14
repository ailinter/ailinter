package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/config"
	"github.com/ailinter/ailinter/internal/git"
	"github.com/ailinter/ailinter/internal/metalinter"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/telemetry"
	"github.com/ailinter/ailinter/internal/vulnerability"

	"github.com/spf13/cobra"
)

type checkOptions struct {
	format              FormatMode
	quiet               bool
	noSecrets           bool
	noVulnerabilities   bool
	secretsOnly         bool
	vulnerabilitiesOnly bool
	langOverride        string
	estimateTokens      bool
	metaLint            bool
	outputPath          string
	diffRef             string // if non-empty, scan only lines changed relative to this git ref
}

func CheckCommand() *cobra.Command {
	var (
		formatFlag          string
		jsonFlag            bool
		quiet               bool
		noSecrets           bool
		noVulnerabilities   bool
		secretsOnly         bool
		vulnerabilitiesOnly bool
		langOverride        string
		noGitignore         bool
		estimateTokens      bool
		metaLint            bool
		noMetaLint          bool
		outputPath          string
		diffRef             string
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
  sarif     SARIF v2.1.0 JSON (GitHub Code Scanning, enterprise CI)

Quiet mode:
  --quiet, -q  Suppress all analysis output (no results, secrets, vulnerabilities,
               meta-lint, token estimates, or summaries). Errors are still printed
               to stderr. Useful for CI when only the exit code matters.

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
				quiet:               quiet,
				noSecrets:           noSecrets,
				noVulnerabilities:   noVulnerabilities,
				secretsOnly:         secretsOnly,
				vulnerabilitiesOnly: vulnerabilitiesOnly,
				langOverride:        langOverride,
				estimateTokens:      estimateTokens,
				metaLint:            metaLint && !noMetaLint,
				outputPath:          outputPath,
				diffRef:             diffRef,
			}
			return executeCheck(args[0], opts, !noGitignore)
		},
	}

	cmd.Flags().StringVar(&formatFlag, "format", "", "Output format: auto, human, json, markdown, problems, sarif")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output results as JSON")
	cmd.Flags().Lookup("json").Hidden = true                                                                                                  // gitleaks:allow
	cmd.Flags().BoolVar(&noSecrets, "no-secrets", false, "Skip secret scanning")                                                              // gitleaks:allow
	cmd.Flags().BoolVar(&noVulnerabilities, "no-vulnerabilities", false, "Skip vulnerability scanning")                                       // gitleaks:allow
	cmd.Flags().BoolVar(&secretsOnly, "secrets-only", false, "Scan ONLY for secrets (skip code quality and vulnerabilities)")                 // gitleaks:allow
	cmd.Flags().BoolVar(&vulnerabilitiesOnly, "vulnerabilities-only", false, "Scan ONLY for vulnerabilities (skip code quality and secrets)") // gitleaks:allow
	cmd.Flags().StringVar(&langOverride, "lang", "", "Force language (auto-detected by default)")
	cmd.Flags().BoolVar(&noGitignore, "no-gitignore", false, "Do not respect .gitignore patterns when scanning directories") // gitleaks:allow
	cmd.Flags().BoolVar(&estimateTokens, "estimate-tokens", false, "Show AI token cost estimation after analysis")
	cmd.Flags().BoolVar(&metaLint, "meta-lint", true, "Run embedded meta-linters (go vet, staticcheck, gofmt, misspell, ineffassign) [default: on]")
	cmd.Flags().BoolVar(&noMetaLint, "no-meta-lint", false, "Skip embedded meta-linters")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress all output except errors (errors still go to stderr)")
	cmd.Flags().StringVar(&outputPath, "output", "", "Write output to file instead of stdout (useful with --format sarif)")
	cmd.Flags().StringVar(&diffRef, "diff", "", "Only scan lines changed relative to a git ref (e.g., 'main', 'HEAD~1', 'HEAD'). Use 'HEAD' for uncommitted changes.")

	return cmd
}

func executeCheck(target string, opts checkOptions, respectGitignore bool) error {
	flags := map[string]string{}
	if opts.quiet {
		flags["quiet"] = "true"
	}
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
	if opts.diffRef != "" {
		flags["diff"] = opts.diffRef
	}
	telemetry.RecordCLIInvocationWithFlags("check", flags)

	info, err := os.Stat(target)
	if err != nil {
		telemetry.RecordError("file_access")
		return fmt.Errorf("cannot access %s: %w", target, err)
	}

	if opts.diffRef != "" {
		return executeCheckDiff(target, opts, respectGitignore, info)
	}

	if info.IsDir() {
		return checkDirectory(target, opts, respectGitignore)
	}
	return checkFile(target, opts)
}

// executeCheckDiff handles diff-aware scanning: resolves the git ref, finds
// changed files, and only reports issues on changed lines.
func executeCheckDiff(target string, opts checkOptions, respectGitignore bool, info os.FileInfo) error {
	resolved, err := resolveSafePath(target)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	// git -C expects a directory; if resolved is a file, use its parent.
	gitDir := resolved
	if !info.IsDir() {
		gitDir = filepath.Dir(resolved)
	}

	repoRoot, err := git.FindRepoRoot(gitDir)
	if err != nil {
		return fmt.Errorf("not a git repository (or unable to find root): %w", err)
	}

	// Get changed files relative to the ref.
	changedFiles, err := git.ChangedFiles(repoRoot, opts.diffRef)
	if err != nil {
		return fmt.Errorf("git diff failed: %w", err)
	}

	// Build a set of changed files for fast lookup.
	changedSet := make(map[string]bool, len(changedFiles))
	for _, f := range changedFiles {
		changedSet[f] = true
	}

	if info.IsDir() {
		return checkDirectoryDiff(resolved, repoRoot, changedSet, opts, respectGitignore)
	}

	// For a specific file, check if it was changed.
	rel, err := filepath.Rel(repoRoot, resolved)
	if err != nil {
		return fmt.Errorf("cannot compute relative path: %w", err)
	}
	if !changedSet[rel] {
		// File was not changed — nothing to report.
		if !opts.quiet {
			fmt.Fprintf(os.Stderr, "no changes detected in %s (relative to %s)\n", target, opts.diffRef)
		}
		return nil
	}

	ranges, err := git.ChangedLines(repoRoot, opts.diffRef, rel)
	if err != nil {
		return fmt.Errorf("git diff failed for %s: %w", target, err)
	}
	if len(ranges) == 0 {
		if !opts.quiet {
			fmt.Fprintf(os.Stderr, "no line-level changes detected in %s (relative to %s)\n", target, opts.diffRef)
		}
		return nil
	}

	optsWithRanges := opts
	return checkFile(target, optsWithRanges, ranges...)
}

// makeWalkContext creates a walkContext from the provided options.
func makeWalkContext(resolvedDir string, opts checkOptions, respectGitignore bool, isDiffMode bool, diffRepoRoot string, diffChangedSet map[string]bool) *walkContext {
	scanQuality := !opts.secretsOnly && !opts.vulnerabilitiesOnly
	scanSecrets := !opts.noSecrets || opts.secretsOnly
	scanVulns := !opts.noVulnerabilities || opts.vulnerabilitiesOnly

	configDir := config.FindConfigDir(resolvedDir)
	ctx := &walkContext{
		opts:           opts,
		resolvedDir:    resolvedDir,
		gitignorePats:  nil,
		excludePats:    config.LoadExcludedFiles(resolvedDir),
		configDir:      configDir,
		scanQuality:    scanQuality,
		scanner:        nil,
		vulnScanner:    nil,
		langCount:      make(map[string]int),
		isDiffMode:     isDiffMode,
		diffRepoRoot:   diffRepoRoot,
		diffChangedSet: diffChangedSet,
		diffRef:        opts.diffRef,
		outputPath:     opts.outputPath,
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
	return ctx
}

// checkDirectoryDiff scans only changed files within a directory.
func checkDirectoryDiff(resolvedDir, repoRoot string, changedSet map[string]bool, opts checkOptions, respectGitignore bool) error {
	ctx := makeWalkContext(resolvedDir, opts, respectGitignore, true, repoRoot, changedSet)

	err := filepath.WalkDir(resolvedDir, ctx.walkFn)
	if err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	telemetry.RecordDirScan(ctx.fileCount, ctx.langCount)
	if !opts.quiet {
		ctx.writeResults()
	}
	if opts.metaLint {
		mlFindings, err := metalinter.LintGo([]string{resolvedDir})
		if err == nil && len(mlFindings) > 0 {
			if !opts.quiet {
				writeMetaLintFindings(opts.format, mlFindings)
			}
		}
	}
	return nil
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

func checkFile(path string, opts checkOptions, diffRanges ...git.LineRange) error {
	start := time.Now()
	resolved, err := resolveSafePath(path)
	if err != nil {
		return err
	}

	if err := checkFileExcluded(resolved, opts); err != nil {
		return nil // excluded, not an error
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
		return checkFileSecretsOnly(resolved, data, opts, start)
	}
	if opts.vulnerabilitiesOnly {
		return checkFileVulnerabilitiesOnly(resolved, data, opts, start)
	}

	lang := opts.detectLang(resolved)
	ext := filepath.Ext(resolved)
	thresholds := config.LoadProjectThresholds(resolved, lang)
	result := analyzer.Analyze(analyzer.SourceInput{FilePath: resolved, Source: string(data), Lang: lang}, thresholds)

	if len(diffRanges) > 0 {
		diffModeAnnotate(&result, diffRanges)
	}

	telemetry.RecordFileAnalyzed(lang, ext)
	telemetry.RecordQualityScore(lang, result.Score)
	for _, s := range result.Smells {
		telemetry.RecordSmellsDetected(s.Name, lang, 1)
	}

	if opts.quiet {
		telemetry.RecordDuration("check_file", lang, time.Since(start).Seconds())
		return nil
	}

	switch opts.format {
	case FormatJSON:
		writeCombinedJSON(result, data, resolved, opts.noSecrets, opts.noVulnerabilities)
	case FormatSARIF:
		if err := checkFileSARIF(result, data, resolved, ext, opts); err != nil {
			return err
		}
	default:
		writeResult(opts.format, result)
		checkFileExtraScans(resolved, data, ext, opts)
		if opts.estimateTokens {
			estimator := analyzer.NewTokenEstimator(resolved, result.Score)
			fmt.Print("\n" + estimator.FormatEstimateOutput())
		}
	}

	telemetry.RecordDuration("check_file", lang, time.Since(start).Seconds())
	return nil
}

// checkFileExcluded returns an error if the file is excluded, nil otherwise.
func checkFileExcluded(resolved string, opts checkOptions) error {
	resolvedDir := filepath.Dir(resolved)
	excludePats := config.LoadExcludedFiles(resolvedDir)
	configDir := config.FindConfigDir(resolvedDir)
	if len(excludePats) > 0 && configDir != "" && config.IsExcluded(resolved, excludePats, configDir) {
		if !opts.quiet {
			fmt.Fprintf(os.Stderr, "skipped (excluded by .ailinter.toml): %s\n", resolved)
		}
		return errExcluded
	}
	return nil
}

var errExcluded = fmt.Errorf("excluded")

func checkFileSecretsOnly(resolved string, data []byte, opts checkOptions, start time.Time) error {
	if !opts.quiet {
		if opts.format == FormatSARIF {
			scanner, secErr := secrets.NewScanner()
			var secFindings []secrets.SecretFinding
			if secErr == nil {
				secFindings = scanner.ScanBytes(data, resolved)
			}
			out := sarifOutput(opts.outputPath)
			WriteSARIFCombined(out, nil, secFindings, nil, nil, resolved)
			if closer, ok := out.(io.Closer); ok && out != os.Stdout {
				closer.Close()
			}
		} else {
			scanAndWriteSecrets(resolved, data, opts.format)
		}
	}
	telemetry.RecordDuration("check_file", "", time.Since(start).Seconds())
	return nil
}

func checkFileVulnerabilitiesOnly(resolved string, data []byte, opts checkOptions, start time.Time) error {
	if !opts.quiet {
		if opts.format == FormatSARIF {
			vulnScanner := vulnerability.NewScanner()
			vulnFindings := vulnScanner.Scan(string(data), resolved)
			out := sarifOutput(opts.outputPath)
			WriteSARIFCombined(out, nil, nil, vulnFindings, nil, resolved)
			if closer, ok := out.(io.Closer); ok && out != os.Stdout {
				closer.Close()
			}
		} else {
			vulnScanner := vulnerability.NewScanner()
			vulnFindings := vulnScanner.Scan(string(data), resolved)
			writeVulnerabilities(opts.format, resolved, vulnFindings)
		}
	}
	telemetry.RecordDuration("check_file", "", time.Since(start).Seconds())
	return nil
}

func checkFileSARIF(result analyzer.QualityResult, data []byte, resolved, ext string, opts checkOptions) error {
	var mlFindings []metalinter.Finding
	if opts.metaLint && ext == ".go" {
		mlFindings, _ = metalinter.LintGo([]string{resolved})
	}
	var secFindings []secrets.SecretFinding
	if !opts.noSecrets {
		scanner, secErr := secrets.NewScanner()
		if secErr == nil {
			secFindings = scanner.ScanBytes(data, resolved)
		}
	}
	var vulnFindings []vulnerability.Finding
	if !opts.noVulnerabilities {
		vulnScanner := vulnerability.NewScanner()
		vulnFindings = vulnScanner.Scan(string(data), resolved)
	}
	out := sarifOutput(opts.outputPath)
	if err := WriteSARIFCombined(out, []analyzer.QualityResult{result}, secFindings, vulnFindings, mlFindings, resolved); err != nil {
		return fmt.Errorf("failed to write SARIF output: %w", err)
	}
	if closer, ok := out.(io.Closer); ok && out != os.Stdout {
		closer.Close()
	}
	return nil
}

func checkFileExtraScans(resolved string, data []byte, ext string, opts checkOptions) {
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
}

// diffModeAnnotate filters a quality result to only show smells in the
// given line ranges, and adds a note to the message indicating diff mode.
func diffModeAnnotate(result *analyzer.QualityResult, ranges []git.LineRange) {
	var filtered []analyzer.Smell
	for _, s := range result.Smells {
		if smellInRanges(s, ranges) {
			filtered = append(filtered, s)
		}
	}
	result.Smells = filtered
}

func checkDirectory(dir string, opts checkOptions, respectGitignore bool) error {
	resolvedDir, err := resolveSafePath(dir)
	if err != nil {
		return err
	}

	ctx := makeWalkContext(resolvedDir, opts, respectGitignore, false, "", nil)

	err = filepath.WalkDir(resolvedDir, ctx.walkFn)
	if err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	telemetry.RecordDirScan(ctx.fileCount, ctx.langCount)
	if !opts.quiet {
		ctx.writeResults()
	}
	if opts.metaLint {
		mlFindings, err := metalinter.LintGo([]string{resolvedDir})
		if err == nil && len(mlFindings) > 0 {
			if !opts.quiet {
				writeMetaLintFindings(opts.format, mlFindings)
			}
		}
	}
	return nil
}

// smellInRanges checks if a smell overlaps with any of the given line ranges.
func smellInRanges(s analyzer.Smell, ranges []git.LineRange) bool {
	for _, r := range ranges {
		// A smell overlaps if its start line is within the range,
		// or if its end line is within the range (for multi-line smells).
		if s.LineStart >= r.Start && s.LineStart <= r.End {
			return true
		}
		if s.LineEnd >= r.Start && s.LineStart <= r.End {
			return true
		}
	}
	return false
}

type walkContext struct {
	opts          checkOptions
	resolvedDir   string
	gitignorePats []string
	excludePats   []string
	configDir     string
	scanQuality   bool
	scanner       *secrets.Scanner
	vulnScanner   *vulnerability.Scanner
	allResults    []analyzer.QualityResult
	allSecrets    []secrets.SecretFinding
	allVulns      []vulnerability.Finding
	fileCount     int
	langCount     map[string]int
	outputPath    string

	// Diff-aware scan fields.
	isDiffMode     bool
	diffRepoRoot   string
	diffChangedSet map[string]bool
	diffRef        string
}

func (ctx *walkContext) walkFn(path string, d os.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() {
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" || base == "testdata" {
			return filepath.SkipDir
		}
		return nil
	}
	if ctx.shouldSkipFile(path) {
		return nil
	}

	// In diff mode, skip files that weren't changed.
	if ctx.isDiffMode {
		rel, relErr := filepath.Rel(ctx.diffRepoRoot, path)
		if relErr != nil || !ctx.diffChangedSet[rel] {
			return nil
		}
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

		// In diff mode, filter smells to only those in changed lines.
		if ctx.isDiffMode {
			rel, _ := filepath.Rel(ctx.diffRepoRoot, path)
			ranges, rangeErr := git.ChangedLines(ctx.diffRepoRoot, ctx.diffRef, rel)
			if rangeErr == nil && len(ranges) > 0 {
				diffModeAnnotate(&result, ranges)
			}
		}

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
	if !isSourceFile(path) {
		return true
	}
	if len(ctx.gitignorePats) > 0 && isGitignored(path, ctx.resolvedDir, ctx.gitignorePats) {
		return true
	}
	if len(ctx.excludePats) > 0 && config.IsExcluded(path, ctx.excludePats, ctx.configDir) {
		return true
	}
	return false
}

func (ctx *walkContext) writeResults() {
	if ctx.opts.secretsOnly {
		if len(ctx.allSecrets) > 0 {
			if ctx.opts.format == FormatSARIF {
				out := sarifOutput(ctx.outputPath)
				WriteSARIFCombined(out, nil, ctx.allSecrets, nil, nil, ctx.resolvedDir)
			} else {
				writeSecrets(ctx.opts.format, "<directory>", ctx.allSecrets)
			}
		}
		return
	}
	if ctx.opts.vulnerabilitiesOnly {
		if len(ctx.allVulns) > 0 {
			if ctx.opts.format == FormatSARIF {
				out := sarifOutput(ctx.outputPath)
				WriteSARIFCombined(out, nil, nil, ctx.allVulns, nil, ctx.resolvedDir)
			} else {
				writeVulnerabilities(ctx.opts.format, "<directory>", ctx.allVulns)
			}
		}
		return
	}
	if ctx.opts.format == FormatJSON {
		writeCombinedDirJSON(ctx.allResults, ctx.allSecrets, ctx.allVulns)
		return
	}
	if ctx.opts.format == FormatSARIF {
		mlFindings, _ := metalinter.LintGo([]string{ctx.resolvedDir})
		out := sarifOutput(ctx.outputPath)
		WriteSARIFCombined(out, ctx.allResults, ctx.allSecrets, ctx.allVulns, mlFindings, ctx.resolvedDir)
		if closer, ok := out.(io.Closer); ok && out != os.Stdout {
			closer.Close()
		}
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
