// Package git provides utilities for working with git repositories,
// including diff-aware analysis support.
package git

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

// LineRange represents a range of changed lines from a git diff (1-based).
type LineRange struct {
	Start int
	End   int
}

// ChangedFiles returns files changed relative to the given ref.
// It runs: git diff <ref> --name-only
func ChangedFiles(repoRoot, ref string) ([]string, error) {
	args := []string{"-C", repoRoot, "diff", ref, "--name-only"}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}

// ChangedFilesStaged returns files changed in the staging area (git diff --cached).
func ChangedFilesStaged(repoRoot, ref string) ([]string, error) {
	args := []string{"-C", repoRoot, "diff", "--cached", ref, "--name-only"}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}

// ChangedLines returns the line ranges changed in a file relative to the given ref.
// Uses git diff with -U0 (no context) to get only the changed line numbers.
func ChangedLines(repoRoot, ref, file string) ([]LineRange, error) {
	cmd := exec.Command("git", "-C", repoRoot, "diff", ref, "-U0", "--", file)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseHunkHeaders(out), nil
}

// parseHunkHeaders extracts line ranges from git diff -U0 output.
// Format: @@ -oldStart,oldCount +newStart,newCount @@
func parseHunkHeaders(diff []byte) []LineRange {
	var ranges []LineRange
	lines := bytes.Split(diff, []byte("\n"))
	for _, line := range lines {
		r := parseHunkLine(line)
		if r != nil {
			ranges = append(ranges, *r)
		}
	}
	return ranges
}

// parseHunkLine parses a single hunk header line.
// Returns nil if the line is not a valid hunk header or has no additions.
func parseHunkLine(line []byte) *LineRange {
	if !bytes.HasPrefix(line, []byte("@@")) {
		return nil
	}
	// Parse @@ -a,b +c,d @@
	parts := bytes.SplitN(line, []byte(" "), 4)
	if len(parts) < 3 {
		return nil
	}
	newPart := bytes.TrimPrefix(parts[2], []byte("+"))
	commaIdx := bytes.IndexByte(newPart, ',')
	if commaIdx == -1 {
		return nil
	}
	start, err := strconv.Atoi(string(newPart[:commaIdx]))
	if err != nil {
		return nil
	}
	count, err := strconv.Atoi(string(newPart[commaIdx+1:]))
	if err != nil || count <= 0 {
		return nil
	}
	return &LineRange{Start: start, End: start + count - 1}
}

// FindRepoRoot finds the git repository root from a given path.
func FindRepoRoot(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
