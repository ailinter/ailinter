package metalinter

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/client9/misspell"
)

// runMisspell checks Go files for common misspellings.
func runMisspell(paths []string) ([]Finding, error) {
	goFiles := collectGoFiles(paths)
	if len(goFiles) == 0 {
		return nil, nil
	}

	replacer := misspell.New()

	var findings []Finding

	for _, path := range goFiles {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			dirFindings, err := checkDirMisspell(path, replacer)
			if err != nil {
				return findings, err
			}
			findings = append(findings, dirFindings...)
			continue
		}
		if !strings.HasSuffix(path, ".go") {
			continue
		}
		fileFindings := checkFileMisspell(path, replacer)
		findings = append(findings, fileFindings...)
	}

	return findings, nil
}

// checkDirMisspell walks a directory and checks all .go files for misspellings.
func checkDirMisspell(dir string, r *misspell.Replacer) ([]Finding, error) {
	var findings []Finding
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
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
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		fileFindings := checkFileMisspell(path, r)
		findings = append(findings, fileFindings...)
		return nil
	})
	return findings, err
}

// checkFileMisspell checks a single file for misspellings and returns findings.
func checkFileMisspell(path string, r *misspell.Replacer) []Finding {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	// Replace method returns corrected string AND a list of diffs.
	_, diffs := r.Replace(string(data))

	var findings []Finding
	for _, d := range diffs {
		findings = append(findings, Finding{
			Tool:     "misspell",
			Code:     "spelling",
			Severity: SeverityInfo,
			File:     path,
			Line:     d.Line,
			Column:   d.Column,
			Message:  d.Original + " is a misspelling of " + d.Corrected,
			Category: CategoryStyle,
		})
	}
	return findings
}
