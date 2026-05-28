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
#   3. ailinter check — full quality + secrets + vulnerabilities + meta-linters
#
# Exit code: 0 if all checks pass, 1 if any fail.

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color
PASS=true

echo ""
echo "=== ailinter Pre-Commit Quality Gate ==="
echo ""

# --- Step 1: go vet ---
echo "  [1/3] Running go vet..."
if ! go vet ./...; then
    echo -e "  ${RED}✗ go vet failed${NC}"
    PASS=false
else
    echo -e "  ${GREEN}✓ go vet passed${NC}"
fi

# --- Step 2: go fmt ---
echo ""
echo "  [2/3] Checking go fmt..."
UNFORMATTED=$(go fmt ./...)
if [ -n "$UNFORMATTED" ]; then
    echo -e "  ${YELLOW}⚠  Unformatted files:${NC}"
    echo "$UNFORMATTED" | sed 's/^/    /'
    echo -e "  ${YELLOW}  Run 'go fmt ./...' to fix${NC}"
    PASS=false
else
    echo -e "  ${GREEN}✓ go fmt passed${NC}"
fi

# --- Step 3: ailinter self-check ---
echo ""
echo "  [3/3] Running ailinter self-check..."
BIN="$(git rev-parse --show-toplevel)/bin/ailinter"
if [ -x "$BIN" ]; then
    if ! "$BIN" check ./... --format problems 2>&1 | head -50; then
        echo -e "  ${RED}✗ ailinter check found issues${NC}"
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
