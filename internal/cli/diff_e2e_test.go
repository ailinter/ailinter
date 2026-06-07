package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ailinter/ailinter/internal/git"
)

// TestDiffE2E tests the end-to-end diff-aware scanning against a temporary
// git repository. It creates an initial commit, then makes changes and
// verifies that only the changed lines produce results.
func TestDiffE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	dir := t.TempDir()

	// Initialize a git repo.
	for _, cmd := range []*exec.Cmd{
		exec.Command("git", "init"),
		exec.Command("git", "config", "user.name", "test"),
		exec.Command("git", "config", "user.email", "test@test.com"),
	} {
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git setup failed: %v\n%s", err, out)
		}
	}

	// Create an initial file with clean code.
	initialSrc := `package main
func hello() {
	println("hello")
}
`
	mainFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(mainFile, []byte(initialSrc), 0644); err != nil {
		t.Fatal(err)
	}

	// Commit the clean file.
	for _, cmd := range []*exec.Cmd{
		exec.Command("git", "add", "main.go"),
		exec.Command("git", "commit", "-m", "initial"),
	} {
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git commit failed: %v\n%s", err, out)
		}
	}

	// Modify to add deeply nested code (which will generate smells).
	modifiedSrc := `package main
func hello() {
	println("hello")
}
func deeplyNested() {
	for i := 0; i < 10; i++ {
		if true {
			for j := 0; j < 10; j++ {
				if true {
					for k := 0; k < 10; k++ {
						println("too deep")
					}
				}
			}
		}
	}
}
`
	if err := os.WriteFile(mainFile, []byte(modifiedSrc), 0644); err != nil {
		t.Fatal(err)
	}

	// Get the repo root.
	repoRoot, err := git.FindRepoRoot(dir)
	if err != nil {
		t.Fatalf("FindRepoRoot failed: %v", err)
	}

	// Verify ChangedLines returns non-empty ranges.
	ranges, err := git.ChangedLines(repoRoot, "HEAD", "main.go")
	if err != nil {
		t.Fatalf("ChangedLines failed: %v", err)
	}
	t.Logf("Changed line ranges: %+v", ranges)
	if len(ranges) == 0 {
		t.Fatal("expected at least one line range")
	}

	// Run check with diff mode.
	opts := checkOptions{
		format:            FormatProblems,
		noSecrets:         true,
		noVulnerabilities: true,
		diffRef:           "HEAD",
	}

	out := captureStdout(func() {
		err := executeCheck(mainFile, opts, false)
		if err != nil {
			t.Fatalf("executeCheck with diff should not error: %v", err)
		}
	})

	t.Logf("Diff scan output:\n%s", out)

	// The diff scan should show deep_nesting (which is in the changed lines).
	if !strings.Contains(out, "deep_nesting") {
		t.Error("expected deep_nesting in diff scan output")
	}

	// Run without diff to compare (full scan).
	optsFull := checkOptions{
		format:            FormatProblems,
		noSecrets:         true,
		noVulnerabilities: true,
	}

	outFull := captureStdout(func() {
		err := executeCheck(mainFile, optsFull, false)
		if err != nil {
			t.Fatalf("executeCheck full should not error: %v", err)
		}
	})

	t.Logf("Full scan output:\n%s", outFull)

	if !strings.Contains(outFull, "deep_nesting") {
		t.Error("expected deep_nesting in full scan output")
	}

	// Verify that --diff on a committed (unchanged) file produces no issues.
	// Create another clean file and commit it.
	otherFile := filepath.Join(dir, "clean.go")
	os.WriteFile(otherFile, []byte("package main\nfunc clean() {}\n"), 0644)
	exec.Command("git", "-C", dir, "add", "clean.go").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "add clean").Run()

	optsNoChanges := checkOptions{
		format:            FormatProblems,
		noSecrets:         true,
		noVulnerabilities: true,
		diffRef:           "HEAD",
	}

	outNoChanges := captureStdout(func() {
		err := executeCheck(otherFile, optsNoChanges, false)
		if err != nil {
			t.Fatalf("executeCheck with diff on unchanged file: %v", err)
		}
	})
	if outNoChanges != "" {
		t.Errorf("expected no output for unchanged file, got: %q", outNoChanges)
	}
}

// TestDiffDirectoryE2E tests diff mode with a directory target.
func TestDiffDirectoryE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	dir := t.TempDir()

	for _, cmd := range []*exec.Cmd{
		exec.Command("git", "init"),
		exec.Command("git", "config", "user.name", "test"),
		exec.Command("git", "config", "user.email", "test@test.com"),
	} {
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git setup failed: %v\n%s", err, out)
		}
	}

	// Create two files and commit them.
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\nfunc a() {}\n"), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("package main\nfunc b() {}\n"), 0644)

	exec.Command("git", "-C", dir, "add", ".").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "initial").Run()

	// Now modify only a.go.
	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\nfunc a() {\n\tif true {\n\t\tif true {\n\t\t\tif true {\n\t\t\t\tif true {\n\t\t\t\t\tprintln(\"deep\")\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t}\n}\n"), 0644)

	// Run directory scan with diff mode.
	opts := checkOptions{
		format:            FormatProblems,
		noSecrets:         true,
		noVulnerabilities: true,
		diffRef:           "HEAD",
	}

	out := captureStdout(func() {
		err := executeCheck(dir, opts, false)
		if err != nil {
			t.Fatalf("executeCheck dir with diff should not error: %v", err)
		}
	})

	t.Logf("Directory diff scan output:\n%s", out)

	// Should only find issues in a.go (the changed file), not b.go.
	if !strings.Contains(out, "a.go") {
		t.Error("expected a.go in diff directory output")
	}
	if strings.Contains(out, "b.go") {
		t.Error("did NOT expect b.go in diff directory output (it wasn't changed)")
	}
	if !strings.Contains(out, "deep_nesting") {
		t.Error("expected deep_nesting in diff directory output")
	}
}
