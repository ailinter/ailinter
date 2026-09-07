#!/bin/bash
# ailinter pre-push gate — PR-only trunk + fast sanity before any push.
#
# Install:
#   ln -sf ../../scripts/pre-push.sh .git/hooks/pre-push
#
# Policy:
#   1. Never push to main directly. Work happens on short-lived branches
#      (chore/*, fix/*, feat/*) and lands via PR — GitHub CI (coverage
#      >=80/70, quality >=90, secrets-fail) + the code-reviewer run there.
#      Release tags are pushed separately by scripts/ship-release.sh.
#   2. Branch pushes get a fast local sanity pass (build + go vet). The full
#      gate lives in CI; this only catches obvious breakage before round-trip.
#
# Exit 0 = push allowed · Exit 1 = push blocked.

set -uo pipefail

ZERO=0000000000000000000000000000000000000000

# --- 1. PR-only trunk enforcement -----------------------------------------
while read local_ref local_sha remote_ref remote_sha; do
  [ "$local_sha" = "$ZERO" ] && continue   # deleting a ref — always allow
  case "$remote_ref" in
    refs/heads/main)
      echo "X Direct push to main blocked (PR-only policy)." >&2
      echo "  Create a short-lived branch and open a PR - CI + review run there." >&2
      exit 1
      ;;
  esac
done

# --- 2. Fast sanity on branch pushes ---------------------------------------
echo "== pre-push: build + go vet =="
if ! make build >/dev/null 2>&1; then
  echo "X build failed - fix before pushing" >&2
  exit 1
fi
if ! out=$(go vet ./cmd/... ./internal/... 2>&1); then
  echo "X go vet failed:" >&2
  echo "$out" | sed -n '1,20p' >&2
  exit 1
fi
echo "OK pre-push gate passed"
exit 0
