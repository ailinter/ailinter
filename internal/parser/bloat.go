package parser

import (
	"fmt"
	"strings"
)

// FileBloatSmell returns a smell if the file exceeds line thresholds.
func DetectFileBloat(totalLoc int, warningThreshold int, alertThreshold int, criticalThreshold int) *Smell {
	if totalLoc < warningThreshold {
		return nil
	}
	sev := "warning"
	if totalLoc >= criticalThreshold {
		sev = "critical"
	} else if totalLoc >= alertThreshold {
		sev = "alert"
	}
	return &Smell{
		Name:     "file_bloat",
		Severity: sev,
		Message:  fmt.Sprintf("File is %d lines (warn:%d, alert:%d, critical:%d)", totalLoc, warningThreshold, alertThreshold, criticalThreshold),
		AIPrompt: fmt.Sprintf("WARNING: File is %d lines — God Class risk. Consider splitting into smaller, cohesive modules.", totalLoc),
	}
}

// FunctionBloat holds per-function length information.
type FunctionBloat struct {
	Name      string
	LineCount int
	LineStart int
}

// DetectBrainMethodSmell checks function lengths against thresholds.
func DetectBrainMethodSmell(bloats []FunctionBloat, warningThreshold int, alertThreshold int) []Smell {
	var smells []Smell
	for _, b := range bloats {
		if b.LineCount < warningThreshold {
			continue
		}
		sev := "warning"
		if b.LineCount >= alertThreshold {
			sev = "alert"
		}
		smells = append(smells, Smell{
			Name:      "brain_method",
			Severity:  sev,
			LineStart: b.LineStart,
			LineEnd:   b.LineStart + b.LineCount,
			Message:   fmt.Sprintf("Function '%s' is %d lines (warning at %d)", b.Name, b.LineCount, warningThreshold),
			AIPrompt:  fmt.Sprintf("WARNING: Brain Method detected. Function '%s' is %d lines. Break into smaller, well-named sub-functions using Extract Method.", b.Name, b.LineCount),
		})
	}
	return smells
}

// Shared helpers used by all language detectors.

func trimLine(line string) string {
	s := strings.TrimSpace(line)
	if strings.HasPrefix(s, "//") || strings.HasPrefix(s, "#") {
		return ""
	}
	return s
}

func countLeadingSpaces(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}

func countOpens(line string) int  { return strings.Count(line, "{") }
func countCloses(line string) int { return strings.Count(line, "}") }
