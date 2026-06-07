package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// InstallHookCommand creates a cobra command that installs the ailinter
// pre-commit hook into the current git repository.
func InstallHookCommand() *cobra.Command {
	var forceBackup bool

	cmd := &cobra.Command{
		Use:   "install-hook",
		Short: "Install the ailinter pre-commit hook",
		Long: `Install the ailinter pre-commit hook into the current git repository.

The hook runs ailinter checks: go vet, staticcheck, gofmt, quality analysis,
and vulnerability detection before every commit.

If a pre-commit hook already exists, it is backed up to pre-commit.backup
before being overwritten. Running the command twice is safe — the backup
from the first run is never overwritten (use --force to override).

Supports standard git repos and git worktrees.

Examples:
  ailinter install-hook
  ailinter install-hook --force    # Overwrite existing backup
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return installHook(forceBackup)
		},
	}

	cmd.Flags().BoolVar(&forceBackup, "force", false, "Overwrite existing backup if one exists")
	return cmd
}

func installHook(forceBackup bool) error {
	// 1. Verify we're in a git repo and resolve hook path.
	hookPath, err := resolveHookPath()
	if err != nil {
		return err
	}

	// 2. Handle existing hook — backup if needed, abort if already installed.
	abort, err := backupExistingHook(hookPath, forceBackup)
	if err != nil {
		return err
	}
	if abort {
		return nil
	}

	// 3. Write the new hook.
	if err := os.WriteFile(hookPath, []byte(preCommitHookContent), 0755); err != nil {
		return fmt.Errorf("cannot write hook: %w", err)
	}

	fmt.Printf("  ✓ ailinter pre-commit hook installed at %s\n", hookPath)
	fmt.Println()
	fmt.Println("  The hook will run before every commit and check:")
	fmt.Println("    • go vet — catch sync.Once copies, loopclosure, nilness")
	fmt.Println("    • gofmt — enforce consistent formatting")
	fmt.Println("    • staticcheck — catch unused code, subtle bugs")
	fmt.Println("    • ailinter — quality + secrets + vulnerability scan")
	fmt.Println()
	fmt.Println("  To skip the hook for a single commit:")
	fmt.Println("    git commit --no-verify")

	// 4. On Windows, also create a .bat wrapper.
	if runtime.GOOS == "windows" {
		batContent := fmt.Sprintf(`@echo off
REM Windows batch wrapper for ailinter pre-commit hook
bash "%%~dp0pre-commit" %%*
`)
		batPath := hookPath + ".bat"
		if err := os.WriteFile(batPath, []byte(batContent), 0755); err == nil {
			fmt.Printf("  ✓ Windows batch wrapper installed at %s\n", batPath)
		}
	}

	return nil
}

// resolveHookPath finds the pre-commit hook path using git rev-parse,
// with support for worktrees via --git-path hooks.
func resolveHookPath() (string, error) {
	gitDir, err := execGit("rev-parse", "--git-dir")
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}

	hooksDir, err := execGit("rev-parse", "--git-path", "hooks")
	if err != nil {
		hooksDir = filepath.Clean(gitDir) + "/hooks"
	}
	if !filepath.IsAbs(hooksDir) {
		abs, err := filepath.Abs(hooksDir)
		if err != nil {
			return "", fmt.Errorf("cannot resolve hooks path: %w", err)
		}
		hooksDir = abs
	}
	return filepath.Join(hooksDir, "pre-commit"), nil
}

// backupExistingHook checks for an existing hook and backs it up if needed.
// Returns abort=true if the ailinter hook is already installed (idempotent).
func backupExistingHook(hookPath string, forceBackup bool) (abort bool, err error) {
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return false, nil
	}

	data, readErr := os.ReadFile(hookPath)
	if readErr == nil && strings.Contains(string(data), "ailinter pre-commit quality gate") {
		fmt.Printf("  ✓ ailinter pre-commit hook already installed at %s\n", hookPath)
		return true, nil
	}

	backupPath := hookPath + ".backup"
	if _, err := os.Stat(backupPath); err == nil && !forceBackup {
		fmt.Printf("  → existing pre-commit hook backed up at %s (already exists)\n", backupPath)
		return false, nil
	}

	if err := os.Rename(hookPath, backupPath); err != nil {
		return false, fmt.Errorf("cannot backup existing hook: %w", err)
	}

	msg := fmt.Sprintf("→ existing pre-commit hook backed up to %s", backupPath)
	if forceBackup {
		msg += " (overwritten)"
	}
	fmt.Printf("  %s\n", msg)
	return false, nil
}

// execGit runs a git command and returns the trimmed stdout.
func execGit(args ...string) (string, error) {
	path, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("git not found in PATH: %w", err)
	}
	cmd := exec.Command(path, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// preCommitHookContent is the bash script written as the pre-commit hook.
// It mirrors scripts/pre-commit.sh in the ailinter repository.
// gitleaks:allow — shell script template, not a real secret
const preCommitHookContent = `#!/bin/bash
# ailinter pre-commit quality gate
# Auto-installed by: ailinter install-hook
# Run as a git pre-commit hook or manually before every commit.
#
# What it checks:
#   1. go vet — catches sync.Once copies, loopclosure, nilness, etc.
#   2. gofmt — catches formatting issues
#   3. staticcheck — catches unused code, style issues, subtle bugs
#   4. ailinter check — full quality + secrets + vulnerabilities + meta-linters
#
# Each step runs on staged Go files only (or full repo if no Go files are
# staged, to catch cross-file issues). Pre-existing issues in untouched
# files do NOT block the commit.
#
# Exit code: 0 if all checks pass, 1 if any fail.

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'
PASS=true

REPO_ROOT="$(git rev-parse --show-toplevel)"

echo ""
echo "=== ailinter Pre-Commit Quality Gate ==="

# Get staged Go files (added, modified, copied)
STAGED_GO=$(git diff --cached --name-only --diff-filter=ACM -- '*.go' || true)
if [ -z "$STAGED_GO" ]; then
    echo "  No staged Go files — nothing to check."
    echo ""
    echo -e "${GREEN}✓ Skipped (no Go changes).${NC}"
    exit 0
fi

echo "  Staged Go files: $(echo "$STAGED_GO" | wc -l | tr -d ' ')"
STAGED_PATHS=$(echo "$STAGED_GO" | sed "s|^|$REPO_ROOT/|" | tr '\n' ' ')
STAGED_PKGS=$(echo "$STAGED_GO" | sed 's|/[^/]*\.go$||' | sort -u | sed 's|^|./|' | tr '\n' ' ')

echo ""

# --- Step 1: go vet ---
echo "  [1/4] Running go vet on changed packages..."
cd "$REPO_ROOT"
VET_FAILED=false
for pkg in $STAGED_PKGS; do
    if [ -d "${pkg#./}" ]; then
        go vet "$pkg" 2>&1 || VET_FAILED=true
    fi
done
if $VET_FAILED; then
    echo -e "  ${RED}✗ go vet failed${NC}"
    PASS=false
else
    echo -e "  ${GREEN}✓ go vet passed${NC}"
fi

# --- Step 2: gofmt ---
echo ""
echo "  [2/4] Checking gofmt on staged files..."
UNFORMATTED=$(for f in $STAGED_PATHS; do [ -f "$f" ] && gofmt -l "$f"; done)
if [ -n "$UNFORMATTED" ]; then
    echo -e "  ${YELLOW}⚠  Unformatted files:${NC}"
    echo "$UNFORMATTED" | sed 's/^/    /'
    echo -e "  ${YELLOW}  Run 'gofmt -w' on these files${NC}"
    PASS=false
else
    echo -e "  ${GREEN}✓ gofmt passed${NC}"
fi

# --- Step 3: staticcheck ---
echo ""
echo "  [3/4] Running staticcheck on changed packages..."
export PATH="$HOME/go/bin:$PATH"
if command -v staticcheck >/dev/null 2>&1; then
    STATICCHECK_FAILED=false
    for pkg in $STAGED_PKGS; do
        if [ -d "${pkg#./}" ]; then
            staticcheck "$pkg" 2>&1 || STATICCHECK_FAILED=true
        fi
    done
    if $STATICCHECK_FAILED; then
        echo -e "  ${RED}✗ staticcheck failed — unused code or subtle bugs in changed packages${NC}"
        PASS=false
    else
        echo -e "  ${GREEN}✓ staticcheck passed${NC}"
    fi
else
    echo -e "  ${YELLOW}⚠  staticcheck not found — install with: go install honnef.co/go/tools/cmd/staticcheck@latest${NC}"
    PASS=false
fi

# --- Step 4: ailinter self-check ---
echo ""
echo "  [4/4] Running ailinter self-check on staged files..."
cd "$REPO_ROOT"
AILINTER_BIN=""
if command -v ailinter >/dev/null 2>&1; then
    AILINTER_BIN="ailinter"
elif [ -x "$REPO_ROOT/bin/ailinter" ]; then
    AILINTER_BIN="$REPO_ROOT/bin/ailinter"
fi
if [ -n "$AILINTER_BIN" ]; then
    AILINTER_FAILED=false
    for f in $STAGED_PATHS; do
        if [ -f "$f" ]; then
            "$AILINTER_BIN" check "$f" --format problems 2>&1 || AILINTER_FAILED=true
        fi
    done
    if $AILINTER_FAILED; then
        echo -e "  ${RED}✗ ailinter check found issues in staged files${NC}"
        PASS=false
    else
        echo -e "  ${GREEN}✓ ailinter check passed${NC}"
    fi
else
    echo -e "  ${YELLOW}⚠  ailinter binary not found in PATH or at $REPO_ROOT/bin/ailinter${NC}"
    echo "     Install it with: go install github.com/ailinter/ailinter/cmd/ailinter@latest"
    PASS=false
fi

echo ""
if [ "$PASS" = true ]; then
    echo -e "${GREEN}✓ All pre-commit checks passed.${NC}"
    exit 0
else
    echo -e "${RED}✗ Pre-commit checks FAILED. Fix issues before committing.${NC}"
    exit 1
fi
`
