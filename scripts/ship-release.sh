#!/bin/bash
# ship-release.sh <version> — tag-and-release ritual for ailinter.
#
# How releases work here:
#   * Every code change lands via PR (CI gates + code-reviewer) on main.
#   * CHANGELOG.md updates are part of the PR that precedes a release.
#   * Pushing a tag vX.Y.Z triggers .github/workflows/release.yml, which
#     builds the 5-binary matrix, creates the GitHub Release with
#     generated notes, and updates the Homebrew formula.
#
# Usage (from a clean, synced main):
#   scripts/ship-release.sh v1.1.0
#
# Requirements:
#   * local main == origin/main (no unpushed commits, no dirty files)
#   * CHANGELOG.md already reflects the release (updated via PR)

set -euo pipefail

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  echo "usage: $0 <version>  (e.g. v1.1.0)" >&2
  exit 2
fi
if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "X version must look like v1.1.0 (got: $VERSION)" >&2
  exit 2
fi

cd "$(git rev-parse --show-toplevel)"

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "X working tree dirty - commit or stash first" >&2
  exit 1
fi

git fetch origin --prune --quiet
LOCAL="$(git rev-parse HEAD)"
REMOTE="$(git rev-parse origin/main)"
if [ "$LOCAL" != "$REMOTE" ]; then
  echo "X local main != origin/main" >&2
  echo "  local : $LOCAL" >&2
  echo "  remote: $REMOTE" >&2
  exit 1
fi

if git rev-parse -q --verify "refs/tags/$VERSION" >/dev/null; then
  echo "X tag $VERSION already exists" >&2
  exit 1
fi

echo "== creating annotated tag $VERSION at $(git log -1 --format=%h) =="
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"
echo "OK tag $VERSION pushed - release.yml is building the matrix + GH Release."
