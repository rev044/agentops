#!/usr/bin/env bash
set -euo pipefail

# Warn when a Go test file is >5 commits behind its source file.
# Always exits 0 (warn-only).

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

STALE=0
CURRENT=0

for source in $(find cli/cmd/ao/ cli/internal/*/ -maxdepth 1 -name '*.go' ! -name '*_test.go' 2>/dev/null | sort); do
  dir=$(dirname "$source")
  base=$(basename "$source" .go)
  test_file="${dir}/${base}_test.go"

  # Skip if no matching test file
  [[ -f "$test_file" ]] || continue

  src_last=$(git log -1 --format=%H -- "$source" 2>/dev/null || true)
  test_last=$(git log -1 --format=%H -- "$test_file" 2>/dev/null || true)

  # Skip if either file has no commit history (new/untracked)
  if [[ -z "$src_last" || -z "$test_last" ]]; then
    CURRENT=$((CURRENT + 1))
    echo "  ok: $source"
    continue
  fi

  # Count commits touching source since the test's last commit
  gap=$(git log --oneline "${test_last}..HEAD" -- "$source" | wc -l | tr -d ' ')

  if [[ "$gap" -gt 5 ]]; then
    STALE=$((STALE + 1))
    echo "WARN: $source is $gap commits ahead of test"
  else
    CURRENT=$((CURRENT + 1))
    echo "  ok: $source"
  fi
done

echo "SUMMARY: $STALE stale, $CURRENT current"
exit 0
