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

# --- 1. PR-only trunk enforcement -----------------------------------------
# Blocks any write to refs/heads/main — direct pushes AND deletions (a delete
# arrives as local_sha == zeros). Non-main ref deletes/updates are allowed.
while read local_ref local_sha remote_ref remote_sha; do
  case "$remote_ref" in
    refs/heads/main)
      if [ "$local_sha" = "0000000000000000000000000000000000000000" ]; then
        echo "X Deleting main is blocked (PR-only policy)." >&2
      else
        echo "X Direct push to main blocked (PR-only policy)." >&2
        echo "  Create a short-lived branch and open a PR - CI + review run there." >&2
      fi
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
