package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ailinter/ailinter/internal/config"
	"github.com/ailinter/ailinter/internal/parser"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/ailinter/ailinter/internal/vulnerability"
)

// ReportData holds all the data needed to render a markdown quality report.
type ReportData struct {
	Target       string
	Timestamp    string
	Results      []QualityResult
	Secrets      []secrets.SecretFinding
	Vulns        []vulnerability.Finding
	OverallScore int
	OverallLabel string
}

// GenerateReport runs a full scan and returns rendered markdown.
func GenerateReport(target string, respectGitignore bool) (*ReportData, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("cannot access %s: %w", target, err)
	}

	var results []QualityResult
	var secretFindings []secrets.SecretFinding
	var vulnFindings []vulnerability.Finding

	if info.IsDir() {
		results, secretFindings, vulnFindings = scanDir(target, respectGitignore)
	} else {
		r, sec, vul := scanFile(target)
		results = append(results, r)
		secretFindings = append(secretFindings, sec...)
		vulnFindings = append(vulnFindings, vul...)
	}

	overallScore := computeOverallScore(results)
	overallLabel := parser.Classify(overallScore)

	return &ReportData{
		Target:       target,
		Timestamp:    time.Now().Format(time.RFC3339),
		Results:      results,
		Secrets:      secretFindings,
		Vulns:        vulnFindings,
		OverallScore: overallScore,
		OverallLabel: overallLabel,
	}, nil
}

func scanFile(path string) (QualityResult, []secrets.SecretFinding, []vulnerability.Finding) {
	resolved, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return QualityResult{Score: 0, Label: LabelStopRefactor, FilePath: path}, nil, nil
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return QualityResult{Score: 0, Label: LabelStopRefactor, FilePath: resolved}, nil, nil
	}
	if isBinaryFile(data) {
		return QualityResult{Score: 0, Label: LabelStopRefactor, FilePath: resolved}, nil, nil
	}

	ext := filepath.Ext(resolved)
	lang := DetectedLanguage(ext)
	if lang == "" {
		lang = "go"
	}
	thresholds := config.LoadProjectThresholds(resolved, lang)
	result := Analyze(SourceInput{FilePath: resolved, Source: string(data), Lang: lang}, thresholds)

	scanner, _ := secrets.NewScanner()
	var sec []secrets.SecretFinding
	if scanner != nil {
		sec = scanner.ScanBytes(data, resolved)
	}

	vulnScanner := vulnerability.NewScanner()
	vuln := vulnScanner.Scan(string(data), resolved)

	return result, sec, vuln
}

func scanDir(dir string, respectGitignore bool) ([]QualityResult, []secrets.SecretFinding, []vulnerability.Finding) {
	resolvedDir, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return nil, nil, nil
	}

	var results []QualityResult
	var secFindings []secrets.SecretFinding
	var vulnFindings []vulnerability.Finding

	var gitignorePats []string
	if respectGitignore {
		gitignorePats = loadGitignoreSimple(resolvedDir)
	}

	scanner, _ := secrets.NewScanner()
	vulnScanner := vulnerability.NewScanner()

	filepath.WalkDir(resolvedDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			if d != nil && d.IsDir() {
				base := filepath.Base(path)
				if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if !isSourceFileReport(path) {
			return nil
		}
		if len(gitignorePats) > 0 && isGitignoredReport(path, resolvedDir, gitignorePats) {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil || isBinaryFile(data) {
			return nil
		}

		ext := filepath.Ext(path)
		lang := DetectedLanguage(ext)
		if lang == "" {
			lang = "go"
		}
		thresholds := config.LoadProjectThresholds(path, lang)
		result := Analyze(SourceInput{FilePath: path, Source: string(data), Lang: lang}, thresholds)
		results = append(results, result)

		if scanner != nil {
			secFindings = append(secFindings, scanner.ScanBytes(data, path)...)
		}
		if vulnScanner != nil {
			vulnFindings = append(vulnFindings, vulnScanner.Scan(string(data), path)...)
		}

		return nil
	})

	return results, secFindings, vulnFindings
}

func computeOverallScore(results []QualityResult) int {
	if len(results) == 0 {
		return 0
	}
	sum := 0
	for _, r := range results {
		sum += r.Score
	}
	return sum / len(results)
}

// RenderMarkdown generates the full markdown report string.
func (rd *ReportData) RenderMarkdown() string {
	var b strings.Builder

	b.WriteString("# Code Quality Report\n\n")

	// Header with emoji
	emoji := scoreEmoji(rd.OverallScore)
	b.WriteString(fmt.Sprintf("**Target:** `%s`  \n", rd.Target))
	b.WriteString(fmt.Sprintf("**Timestamp:** %s  \n", rd.Timestamp))
	b.WriteString(fmt.Sprintf("**Overall Score:** %s %d/100 — %s  \n\n", emoji, rd.OverallScore, rd.OverallLabel))

	// Score breakdown
	b.WriteString("## Score Breakdown\n\n")
	if len(rd.Results) == 1 {
		r := rd.Results[0]
		b.WriteString(fmt.Sprintf("| Metric | Value |\n"))
		b.WriteString(fmt.Sprintf("|--------|-------|\n"))
		b.WriteString(fmt.Sprintf("| Score | %d/100 |\n", r.Score))
		b.WriteString(fmt.Sprintf("| Label | %s |\n", r.Label))
		b.WriteString(fmt.Sprintf("| Language | %s |\n", r.Language))
		b.WriteString(fmt.Sprintf("| Lines of Code | %d |\n", r.LinesOfCode))
		b.WriteString(fmt.Sprintf("| Issues Found | %d |\n", len(r.Smells)))
	} else {
		b.WriteString(fmt.Sprintf("| Category | Count |\n"))
		b.WriteString(fmt.Sprintf("|----------|-------|\n"))
		b.WriteString(fmt.Sprintf("| Files Analyzed | %d |\n", len(rd.Results)))
		b.WriteString(fmt.Sprintf("| Average Score | %d/100 |\n", rd.OverallScore))

		var goAhead, care, needsWork, stop int
		for _, r := range rd.Results {
			switch r.Label {
			case LabelGoAhead:
				goAhead++
			case LabelProceedWithCare:
				care++
			case LabelNeedsWork:
				needsWork++
			default:
				stop++
			}
		}
		b.WriteString(fmt.Sprintf("| Go Ahead (80-100) | %d |\n", goAhead))
		b.WriteString(fmt.Sprintf("| Proceed with Care (60-79) | %d |\n", care))
		b.WriteString(fmt.Sprintf("| Needs Work (40-59) | %d |\n", needsWork))
		b.WriteString(fmt.Sprintf("| Stop & Refactor (0-39) | %d |\n", stop))

		totalIssues := 0
		sevCounts := map[string]int{}
		for _, r := range rd.Results {
			totalIssues += len(r.Smells)
			for _, s := range r.Smells {
				sevCounts[s.Severity]++
			}
		}
		b.WriteString(fmt.Sprintf("| Total Issues | %d |\n", totalIssues))
		if len(sevCounts) > 0 {
			for _, sev := range []string{"critical", "alert", "warning"} {
				if c := sevCounts[sev]; c > 0 {
					b.WriteString(fmt.Sprintf("| %s Issues | %d |\n", strings.Title(sev), c))
				}
			}
		}
	}
	b.WriteString("\n")

	// Detector results summary
	b.WriteString("## Detector Results\n\n")
	b.WriteString(fmt.Sprintf("| Detector | Status |\n"))
	b.WriteString(fmt.Sprintf("|----------|--------|\n"))
	b.WriteString(fmt.Sprintf("| Code Quality | %s |\n", detectorStatus("quality", rd.Results)))
	b.WriteString(fmt.Sprintf("| Secret Scan | %s |\n", detectorStatus("secrets", rd.Results)))
	b.WriteString(fmt.Sprintf("| Vulnerability Scan | %s |\n", detectorStatus("vulns", rd.Results)))
	b.WriteString("\n")

	// Secret scan summary
	b.WriteString("## Secret Scan\n\n")
	b.WriteString(fmt.Sprintf("**Target:** `%s`  \n\n", rd.Target))
	if len(rd.Secrets) == 0 {
		b.WriteString("✅ **Clean** — No secrets detected.\n\n")
	} else {
		b.WriteString(fmt.Sprintf("⚠️ **%d secret(s) detected.**\n\n", len(rd.Secrets)))
		b.WriteString("| Line | Rule | Severity | Description |\n")
		b.WriteString("|------|------|----------|-------------|\n")
		for _, f := range rd.Secrets {
			b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
				f.Line, f.RuleID, f.Severity, f.Description))
		}
		b.WriteString("\n")
	}

	// Vulnerability summary
	b.WriteString("## Vulnerability Scan\n\n")
	b.WriteString(fmt.Sprintf("**Target:** `%s`  \n\n", rd.Target))
	if len(rd.Vulns) == 0 {
		b.WriteString("✅ **Clean** — No vulnerability patterns detected.\n\n")
	} else {
		b.WriteString(fmt.Sprintf("⚠️ **%d vulnerability pattern(s) detected.**\n\n", len(rd.Vulns)))
		b.WriteString("| Line | Category | Rule | Severity |\n")
		b.WriteString("|------|----------|------|----------|\n")
		for _, f := range rd.Vulns {
			b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
				f.Line, f.Category, f.RuleID, f.Severity))
		}
		b.WriteString("\n")
	}

	// All detected issues table
	b.WriteString("## All Issues\n\n")
	totalIssues := 0
	for _, r := range rd.Results {
		totalIssues += len(r.Smells)
	}
	if totalIssues == 0 {
		b.WriteString("✅ No issues detected across all targets.\n\n")
	} else {
		b.WriteString("| File | Line | Severity | Issue | Description |\n")
		b.WriteString("|------|------|----------|-------|-------------|\n")
		for _, r := range rd.Results {
			for _, s := range r.Smells {
				b.WriteString(fmt.Sprintf("| `%s` | %d | %s | %s | %s |\n",
					filepath.Base(r.FilePath), s.LineStart, s.Severity, s.Name, s.Message))
			}
		}
	}
	b.WriteString("\n")

	// Footer
	b.WriteString("---\n\n")
	b.WriteString(fmt.Sprintf("_Report generated by [ailinter](https://ailinter.dev) — Code Quality for AI-Assisted Development._\n"))

	return b.String()
}

func scoreEmoji(score int) string {
	switch {
	case score >= 80:
		return "🟢"
	case score >= 60:
		return "🟡"
	default:
		return "🔴"
	}
}

func detectorStatus(detector string, results []QualityResult) string {
	switch detector {
	case "quality":
		for _, r := range results {
			if len(r.Smells) > 0 {
				return fmt.Sprintf("⚠️ %d issue(s) found", countAllSmells(results))
			}
		}
		return "✅ Passed"
	default:
		return "✅ Passed"
	}
}

func countAllSmells(results []QualityResult) int {
	n := 0
	for _, r := range results {
		n += len(r.Smells)
	}
	return n
}

func loadGitignoreSimple(root string) []string {
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

func isGitignoredReport(path, root string, patterns []string) bool {
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

func isSourceFileReport(path string) bool {
	sourceExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
		".java": true, ".rs": true, ".rb": true, ".c": true, ".cpp": true,
		".h": true, ".hpp": true, ".cs": true, ".swift": true, ".kt": true,
		".kts": true, ".scala": true, ".php": true, ".pl": true, ".sh": true,
		".bash": true, ".tf": true, ".yaml": true, ".yml": true, ".toml": true,
		".json": true, ".xml": true, ".html": true, ".css": true, ".sql": true,
		".properties": true, ".ini": true, ".cfg": true, ".conf": true,
	}
	if sourceExts[filepath.Ext(path)] {
		return true
	}
	sourceBases := map[string]bool{
		".env": true, "Dockerfile": true, "Makefile": true,
		".gitignore": true, ".gitattributes": true, ".npmrc": true,
		".editorconfig": true, ".dockerignore": true,
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

func isBinaryFile(data []byte) bool {
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
