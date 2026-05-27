package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Output  string    `json:"Output,omitempty"`
	Elapsed float64   `json:"Elapsed,omitempty"`
}

type PackageResult struct {
	Package  string
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
	Tests    []TestResult
	Status   string
}

type TestResult struct {
	Name     string
	Status   string
	Duration float64
	Output   []string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: test-report <test-report.json>")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}
	defer f.Close()

	packages := make(map[string]*PackageResult)
	var events []TestEvent

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var ev TestEvent
		if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			continue
		}
		events = append(events, ev)
	}

	for _, ev := range events {
		pkg, ok := packages[ev.Package]
		if !ok {
			pkg = &PackageResult{Package: ev.Package, Status: "pass"}
			packages[ev.Package] = pkg
		}

		switch ev.Action {
		case "pass", "fail", "skip":
			if ev.Test != "" {
				tr := TestResult{Name: ev.Test, Status: ev.Action, Duration: ev.Elapsed}
				pkg.Tests = append(pkg.Tests, tr)
				switch ev.Action {
				case "pass":
					pkg.Passed++
				case "fail":
					pkg.Failed++
					pkg.Status = "fail"
				case "skip":
					pkg.Skipped++
				}
			}
		case "output":
			if ev.Test != "" && len(pkg.Tests) > 0 {
				last := &pkg.Tests[len(pkg.Tests)-1]
				last.Output = append(last.Output, ev.Output)
			}
		}
	}

	// Sort packages
	var pkgNames []string
	for name := range packages {
		pkgNames = append(pkgNames, name)
	}
	sort.Strings(pkgNames)

	totalPassed := 0
	totalFailed := 0
	totalSkipped := 0

	for _, name := range pkgNames {
		pkg := packages[name]
		totalPassed += pkg.Passed
		totalFailed += pkg.Failed
		totalSkipped += pkg.Skipped
	}

	fmt.Print(generateHTML(pkgNames, packages, totalPassed, totalFailed, totalSkipped))
}

func generateHTML(pkgNames []string, packages map[string]*PackageResult, passed, failed, skipped int) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>ailinter Test Report</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; background: #1a1a2e; color: #e0e0e0; }
h1 { color: #7c3aed; }
.summary { display: flex; gap: 20px; margin: 20px 0; }
.stat-card { background: #16213e; border-radius: 12px; padding: 20px; flex: 1; text-align: center; border: 2px solid #2a2a4a; }
.stat-card h2 { margin: 0; font-size: 36px; }
.stat-card p { margin: 5px 0 0; color: #888; }
.passed { color: #22c55e; }
.failed { color: #ef4444; }
.skipped { color: #f59e0b; }
.package { background: #16213e; border-radius: 12px; margin: 15px 0; overflow: hidden; border: 1px solid #2a2a4a; }
.package-header { display: flex; justify-content: space-between; align-items: center; padding: 15px 20px; cursor: pointer; }
.package-header:hover { background: #1e2a4a; }
.package-name { font-weight: 600; font-size: 16px; }
.package-status { font-weight: 700; text-transform: uppercase; font-size: 14px; }
.package-stats { font-size: 14px; color: #888; }
.test-list { display: none; padding: 0 20px 15px; }
.test-list.open { display: block; }
.test-item { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; border-radius: 6px; margin: 4px 0; }
.test-item:nth-child(odd) { background: rgba(255,255,255,0.03); }
.test-name { font-family: 'SF Mono', 'Monaco', monospace; font-size: 13px; }
.test-duration { font-size: 12px; color: #888; }
.test-output { font-family: 'SF Mono', 'Monaco', monospace; font-size: 11px; background: #0d1117; padding: 10px; border-radius: 6px; margin: 5px 0; white-space: pre-wrap; color: #8b949e; display: none; }
.test-output.open { display: block; }
.test-fail { border-left: 3px solid #ef4444; }
.test-pass { border-left: 3px solid #22c55e; }
.test-skip { border-left: 3px solid #f59e0b; }
.fail-text { color: #ef4444; }
.footer { text-align: center; margin: 40px 0 20px; color: #555; font-size: 12px; }
</style>
<script>
function toggle(id) { var el = document.getElementById(id); el.classList.toggle('open'); }
</script>
</head>
<body>
<h1>ailinter Test Report</h1>
`)

	fmt.Fprintf(&b, `<div class="summary">
<div class="stat-card"><h2 class="passed">%d</h2><p>Passed</p></div>
<div class="stat-card"><h2 class="failed">%d</h2><p>Failed</p></div>
<div class="stat-card"><h2 class="skipped">%d</h2><p>Skipped</p></div>
<div class="stat-card"><h2>%d</h2><p>Total</p></div>
</div>
`, passed, failed, skipped, passed+failed+skipped)

	for i, name := range pkgNames {
		pkg := packages[name]
		statusClass := "passed"
		if pkg.Status == "fail" {
			statusClass = "failed"
		}
		id := fmt.Sprintf("pkg-%d", i)

		fmt.Fprintf(&b, `<div class="package">
<div class="package-header" onclick="toggle('%s')">
<span class="package-name">%s</span>
<span class="package-stats">%d pass / %d fail / %d skip</span>
<span class="package-status %s">%s</span>
</div>
<div class="test-list" id="%s">
`, id, shortPkg(name), pkg.Passed, pkg.Failed, pkg.Skipped, statusClass, pkg.Status, id)

		for j, test := range pkg.Tests {
			itemClass := "test-pass"
			if test.Status == "fail" {
				itemClass = "test-fail"
			} else if test.Status == "skip" {
				itemClass = "test-skip"
			}

			tid := fmt.Sprintf("test-%d-%d", i, j)
			fmt.Fprintf(&b, `<div class="test-item %s">
<span class="test-name %s-%s">%s</span>
<span class="test-duration">%.2fs</span>
</div>`, itemClass, test.Status, "text", test.Name, test.Duration)

			if len(test.Output) > 0 {
				fmt.Fprintf(&b, `<div class="test-output open" id="%s">`, tid+"-out")
				for _, line := range test.Output {
					fmt.Fprintf(&b, "%s", escapeHTML(line))
				}
				fmt.Fprint(&b, `</div>`)
			}
		}
		fmt.Fprint(&b, `</div></div>`)
	}

	b.WriteString(`<div class="footer">Generated by ailinter</div>
</body></html>`)

	return b.String()
}

func shortPkg(pkg string) string {
	return strings.TrimPrefix(pkg, "github.com/ailinter/ailinter/")
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
