package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseHunkHeaders tests the hunk header parsing logic. This is pure logic
// that doesn't require git, so we test it directly.
func TestParseHunkHeaders(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		expected []LineRange
	}{
		{
			name:     "empty diff",
			diff:     "",
			expected: nil,
		},
		{
			name: "single hunk",
			diff: `diff --git a/main.go b/main.go
index abc..def 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,5 @@
 package main
-func main() {}
+func main() {
+	println("hello")
+}
`,
			expected: []LineRange{{Start: 1, End: 5}},
		},
		{
			name: "multiple hunks",
			diff: `diff --git a/main.go b/main.go
index abc..def 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,5 @@
 package main
-func main() {}
+func main() {
+	println("hello")
+}
@@ -10,2 +12,5 @@
 func foo() {
-	println("old")
+	println("new")
+	println("new2")
+	println("new3")
 }
`,
			expected: []LineRange{
				{Start: 1, End: 5},
				{Start: 12, End: 16},
			},
		},
		{
			name: "new file (no old)",
			diff: `diff --git a/new.go b/new.go
new file mode 100644
index 000..abc
--- /dev/null
+++ b/new.go
@@ -0,0 +1,3 @@
+package main
+func main() {}
+
`,
			expected: []LineRange{{Start: 1, End: 3}},
		},
		{
			name: "deleted file should not produce ranges",
			diff: `diff --git a/old.go b/old.go
deleted file mode 100644
index abc..000
--- a/old.go
+++ /dev/null
@@ -1,5 +0,0 @@
-package main
-func main() {
-	println("removed")
-}
`,
			expected: nil,
		},
		{
			name: "single line change",
			diff: `@@ -10,1 +10,1 @@
-func old() {}
+func new() {}
`,
			expected: []LineRange{{Start: 10, End: 10}},
		},
		{
			name:     "malformed hunk header",
			diff:     `@@ not a valid hunk @@`,
			expected: nil,
		},
		{
			name: "zero count additions (just context removal)",
			diff: `@@ -1,3 +1,0 @@
-foo
-bar
-baz
`,
			expected: nil, // count=0, so no range
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHunkHeaders([]byte(tt.diff))
			if len(result) != len(tt.expected) {
				t.Errorf("got %d ranges, want %d: %+v", len(result), len(tt.expected), result)
				return
			}
			for i, r := range result {
				if r.Start != tt.expected[i].Start || r.End != tt.expected[i].End {
					t.Errorf("range %d: got {Start=%d, End=%d}, want {Start=%d, End=%d}",
						i, r.Start, r.End, tt.expected[i].Start, tt.expected[i].End)
				}
			}
		})
	}
}

// TestChangedFilesInRepo tests ChangedFiles and ChangedLines against a real
// temporary git repository.
func TestChangedFilesInRepo(t *testing.T) {
	dir := t.TempDir()

	// Initialize git repo
	initCmd := exec.Command("git", "init")
	initCmd.Dir = dir
	out, err := initCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	// Configure git user for commits
	for _, cfg := range []string{"user.name", "user.email"} {
		val := "test"
		if cfg == "user.email" {
			val = "test@test.com"
		}
		cmd := exec.Command("git", "config", cfg, val)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("git config %s failed: %v", cfg, err)
		}
	}

	// Create initial commit
	initial := filepath.Join(dir, "main.go")
	if err := os.WriteFile(initial, []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exec.Command("git", "-C", dir, "add", "main.go").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "initial").Run()

	// Add new content to simulate changes
	if err := os.WriteFile(initial, []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test ChangedFiles with HEAD (shows uncommitted changes)
	repoRoot, err := FindRepoRoot(dir)
	if err != nil {
		t.Fatalf("FindRepoRoot failed: %v", err)
	}

	files, err := ChangedFiles(repoRoot, "HEAD")
	if err != nil {
		t.Fatalf("ChangedFiles failed: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one changed file")
	}
	found := false
	for _, f := range files {
		if strings.Contains(f, "main.go") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected main.go in changed files, got: %v", files)
	}

	// Test ChangedLines
	ranges, err := ChangedLines(repoRoot, "HEAD", "main.go")
	if err != nil {
		t.Fatalf("ChangedLines failed: %v", err)
	}
	if len(ranges) == 0 {
		t.Fatal("expected at least one line range")
	}
	t.Logf("Changed lines for main.go: %+v", ranges)
}

// TestNoChanges tests when there are no changes.
func TestNoChanges(t *testing.T) {
	dir := t.TempDir()

	initCmd := exec.Command("git", "init")
	initCmd.Dir = dir
	out, err := initCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	for _, cfg := range []string{"user.name", "user.email"} {
		val := "test"
		if cfg == "user.email" {
			val = "test@test.com"
		}
		cmd := exec.Command("git", "config", cfg, val)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("git config %s failed: %v", cfg, err)
		}
	}

	initial := filepath.Join(dir, "main.go")
	os.WriteFile(initial, []byte("package main\nfunc main() {}\n"), 0644)
	exec.Command("git", "-C", dir, "add", "main.go").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "initial").Run()

	repoRoot, err := FindRepoRoot(dir)
	if err != nil {
		t.Fatalf("FindRepoRoot failed: %v", err)
	}

	// No changes should return empty
	files, err := ChangedFiles(repoRoot, "HEAD")
	if err != nil {
		t.Fatalf("ChangedFiles failed: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 changed files, got %d: %v", len(files), files)
	}
}

// TestFindRepoRootOutsideRepo tests that FindRepoRoot fails outside a git repo.
func TestFindRepoRootOutsideRepo(t *testing.T) {
	dir := t.TempDir()
	_, err := FindRepoRoot(dir)
	if err == nil {
		t.Error("expected error when not in a git repository")
	}
}
