package metalinter

import (
	"fmt"
	"sync"
)

// LintGo runs all Go meta-linters on the given package paths.
// Returns findings from all embedded tools.
// The pkgPaths are file paths or directory paths containing Go files.
func LintGo(pkgPaths []string) ([]Finding, error) {
	var (
		findings []Finding
		mu       sync.Mutex
		wg       sync.WaitGroup
		errCh    = make(chan error, 5)
	)

	addFindings := func(fs []Finding) {
		mu.Lock()
		findings = append(findings, fs...)
		mu.Unlock()
	}

	// Run go vet + staticcheck via shared analysis driver
	wg.Add(1)
	go func() {
		defer wg.Done()
		fs, err := runGoVetStaticcheck(pkgPaths)
		if err != nil {
			errCh <- fmt.Errorf("govet/staticcheck: %w", err)
			return
		}
		addFindings(fs)
	}()

	// Run gofmt on all Go files
	wg.Add(1)
	go func() {
		defer wg.Done()
		fs, err := runGofmt(pkgPaths)
		if err != nil {
			errCh <- fmt.Errorf("gofmt: %w", err)
			return
		}
		addFindings(fs)
	}()

	// Run misspell on all Go files
	wg.Add(1)
	go func() {
		defer wg.Done()
		fs, err := runMisspell(pkgPaths)
		if err != nil {
			errCh <- fmt.Errorf("misspell: %w", err)
			return
		}
		addFindings(fs)
	}()

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return findings, fmt.Errorf("meta-linter errors: %v", errs)
	}

	return findings, nil
}

// collectGoFiles returns the paths as-is for now.
// In the future, this may expand directories into .go file paths.
func collectGoFiles(paths []string) []string {
	return append([]string(nil), paths...)
}
