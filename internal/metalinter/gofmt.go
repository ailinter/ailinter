package metalinter

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

// runGofmt checks Go formatting by parsing and re-formatting each file.
// Returns findings for files that are not gofmt-compliant.
func runGofmt(paths []string) ([]Finding, error) {
	goFiles := collectGoFiles(paths)
	if len(goFiles) == 0 {
		return nil, nil
	}

	var findings []Finding

	for _, path := range goFiles {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			dirFindings, err := checkDirGofmt(path)
			if err != nil {
				return findings, err
			}
			findings = append(findings, dirFindings...)
			continue
		}
		if !strings.HasSuffix(path, ".go") {
			continue
		}
		fFinding, err := checkFileGofmt(path)
		if err != nil {
			return findings, err
		}
		if fFinding != nil {
			findings = append(findings, *fFinding)
		}
	}

	return findings, nil
}

// checkDirGofmt walks a directory and checks all .go files for formatting.
func checkDirGofmt(dir string) ([]Finding, error) {
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
		fFinding, err := checkFileGofmt(path)
		if err != nil {
			return err
		}
		if fFinding != nil {
			findings = append(findings, *fFinding)
		}
		return nil
	})
	return findings, err
}

// checkFileGofmt checks a single Go file for gofmt compliance.
// Returns a Finding if the file is not properly formatted, or nil if it is.
func checkFileGofmt(path string) (*Finding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}

	formatted, err := format.Source(data)
	if err != nil {
		// Parse error — not a formatting issue per se, but report it
		return &Finding{
			Tool:     "gofmt",
			Severity: SeverityWarning,
			File:     path,
			Message:  fmt.Sprintf("file has parse errors (cannot check formatting): %v", err),
			Category: CategoryFormatting,
		}, nil
	}

	if !bytes.Equal(data, formatted) {
		// Find the first line that differs
		origLines := bytes.Split(data, []byte("\n"))
		fmtLines := bytes.Split(formatted, []byte("\n"))
		diffLine := 1
		for i := 0; i < len(origLines) && i < len(fmtLines); i++ {
			if !bytes.Equal(origLines[i], fmtLines[i]) {
				diffLine = i + 1
				break
			}
		}

		return &Finding{
			Tool:     "gofmt",
			Code:     "formatting",
			Severity: SeverityWarning,
			File:     path,
			Line:     diffLine,
			Column:   1,
			Message:  "file is not gofmt-ed with gofmt (gofmt style)",
			Category: CategoryFormatting,
		}, nil
	}

	return nil, nil
}
