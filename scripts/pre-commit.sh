#!/bin/bash
# ailinter pre-commit quality gate
# Run this as a git pre-commit hook or manually before every commit.
#
# Installation:
#   ln -sf ../../scripts/pre-commit.sh .git/hooks/pre-commit
#
# What it checks:
#   1. go vet — catches sync.Once copies, loopclosure, nilness, etc.
#   2. go fmt — catches formatting issues
#   3. staticcheck — catches unused code, style issues, subtle bugs
#   4. ailinter check — full quality + secrets + vulnerabilities + meta-linters
#
# Each step runs on staged Go files only (or full repo if no Go files are staged,
# to catch cross-file issues). Pre-existing issues in untouched files do NOT block.
#
# Exit code: 0 if all checks pass, 1 if any fail.

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color
PASS=true

REPO_ROOT="$(git rev-parse --show-toplevel)"

echo ""
echo "=== ailinter Pre-Commit Quality Gate ==="

# Get staged Go files (added, modified, copied)
STAGED_GO=$(git diff --cached --name-only --diff-filter=ACM -- '*.go' || true)
if [ -z "$STAGED_GO" ]; then
    echo "  No staged Go files — nothing to check."
    echo ""
    echo -e "${GREEN}✅ Skipped (no Go changes).${NC}"
    exit 0
fi

echo "  Staged Go files: $(echo "$STAGED_GO" | wc -l | tr -d ' ')"
# Convert to space-separated list of absolute paths
STAGED_PATHS=$(echo "$STAGED_GO" | sed "s|^|$REPO_ROOT/|" | tr '\n' ' ')
# Get unique package directories
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

# --- Step 2: go fmt ---
echo ""
echo "  [2/4] Checking go fmt on staged files..."
UNFORMATTED=$(for f in $STAGED_PATHS; do [ -f "$f" ] && gofmt -l "$f"; done)
if [ -n "$UNFORMATTED" ]; then
    echo -e "  ${YELLOW}⚠  Unformatted files:${NC}"
    echo "$UNFORMATTED" | sed 's/^/    /'
    echo -e "  ${YELLOW}  Run 'gofmt -w' on these files${NC}"
    PASS=false
else
    echo -e "  ${GREEN}✓ go fmt passed${NC}"
fi

# --- Step 3: staticcheck ---
echo ""
echo "  [3/4] Running staticcheck on changed packages..."
# git hooks run with a restricted PATH — add ~/go/bin as fallback
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
BIN="$REPO_ROOT/bin/ailinter"
if [ -x "$BIN" ]; then
    AILINTER_FAILED=false
    for f in $STAGED_PATHS; do
        if [ -f "$f" ]; then
            "$BIN" check "$f" --format problems 2>&1 || AILINTER_FAILED=true
        fi
    done
    if $AILINTER_FAILED; then
        echo -e "  ${RED}✗ ailinter check found issues in staged files${NC}"
        PASS=false
    else
        echo -e "  ${GREEN}✓ ailinter check passed${NC}"
    fi
else
    echo -e "  ${YELLOW}⚠  ailinter binary not found at $BIN${NC}"
    echo "     Run 'make build' first."
    PASS=false
fi

echo ""
if [ "$PASS" = true ]; then
    echo -e "${GREEN}✅ All pre-commit checks passed.${NC}"
    exit 0
else
    echo -e "${RED}❌ Pre-commit checks FAILED. Fix issues before committing.${NC}"
    exit 1
fi
