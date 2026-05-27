package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ailinter/ailinter/internal/analyzer"
	"github.com/ailinter/ailinter/internal/secrets"
	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

type FormatMode int

const (
	FormatAuto FormatMode = iota
	FormatHuman
	FormatJSON
	FormatMarkdown
	FormatProblems
)

func (f FormatMode) String() string {
	switch f {
	case FormatHuman:
		return "human"
	case FormatJSON:
		return "json"
	case FormatMarkdown:
		return "markdown"
	case FormatProblems:
		return "problems"
	default:
		return "auto"
	}
}

func DetectFormat(flagValue string) FormatMode {
	if flagValue != "" {
		switch strings.ToLower(flagValue) {
		case "json":
			return FormatJSON
		case "markdown", "md":
			return FormatMarkdown
		case "text", "human":
			return FormatHuman
		case "problems", "gcc", "vscode":
			return FormatProblems
		case "auto":
			return FormatAuto
		}
	}

	if env := os.Getenv("CLI_FORMAT"); env != "" {
		switch strings.ToLower(env) {
		case "json":
			return FormatJSON
		case "markdown", "md":
			return FormatMarkdown
		case "text", "human":
			return FormatHuman
		case "problems", "gcc", "vscode":
			return FormatProblems
		}
	}

	if env := os.Getenv("NO_COLOR"); env != "" {
		return FormatJSON
	}

	if isatty.IsTerminal(os.Stdout.Fd()) {
		return FormatHuman
	}
	return FormatJSON
}

func ResolveFormat(flagValue string) FormatMode {
	mode := DetectFormat(flagValue)
	if mode == FormatAuto {
		if isatty.IsTerminal(os.Stdout.Fd()) {
			return FormatHuman
		}
		return FormatJSON
	}
	return mode
}

func IsColorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
			return true
		}
		return false
	}
	return true
}

func TerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func writeResult(format FormatMode, result analyzer.QualityResult) {
	switch format {
	case FormatJSON:
		writeJSONResult(result)
	case FormatMarkdown:
		writeMarkdownResult(result)
	case FormatProblems:
		writeProblemsResult(result)
	default:
		writeHumanResult(result)
	}
}

func writeResults(format FormatMode, results []analyzer.QualityResult) {
	switch format {
	case FormatJSON:
		writeJSONResults(results)
	case FormatMarkdown:
		for i, r := range results {
			if i > 0 {
				fmt.Println()
			}
			writeMarkdownResult(r)
		}
	case FormatProblems:
		for _, r := range results {
			writeProblemsResult(r)
		}
	default:
		for _, r := range results {
			writeHumanResult(r)
		}
	}
}

func writeSecrets(format FormatMode, path string, findings []secrets.SecretFinding) {
	if len(findings) == 0 {
		return
	}
	switch format {
	case FormatJSON:
	case FormatMarkdown:
		writeMarkdownSecrets(path, findings)
	case FormatProblems:
		writeProblemsSecrets(path, findings)
	default:
		writeHumanSecrets(path, findings)
	}
}

func writeSummary(format FormatMode, results []analyzer.QualityResult) {
	switch format {
	case FormatJSON, FormatProblems:
	case FormatMarkdown:
		writeMarkdownSummary(results)
	default:
		writeHumanSummary(results)
	}
}

func truncateMsg(msg string, max int) string {
	if len(msg) <= max {
		return msg
	}
	return msg[:max-1] + "…"
}

func countBySeverity(smells []analyzer.Smell) map[string]int {
	counts := map[string]int{}
	for _, s := range smells {
		counts[s.Severity]++
	}
	return counts
}

func visualLen(s string) int {
	n := 0
	inEsc := false
	for _, r := range s {
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		if r == '\033' {
			inEsc = true
			continue
		}
		n++
	}
	return n
}

func progressToStderr(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func writeJSONResult(result analyzer.QualityResult) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}

func writeJSONResults(results []analyzer.QualityResult) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(results)
}

type combinedResult struct {
	CodeQuality  analyzer.QualityResult  `json:"code_quality"`
	SecretScan   []secrets.SecretFinding `json:"secret_scan,omitempty"`
}

func writeCombinedJSON(result analyzer.QualityResult, data []byte, path string) {
	output := combinedResult{CodeQuality: result}
	scanner, err := secrets.NewScanner()
	if err == nil {
		output.SecretScan = scanner.ScanBytes(data, path)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(output)
}
