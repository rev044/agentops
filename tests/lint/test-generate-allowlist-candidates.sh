#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
GENERATOR="$REPO_ROOT/scripts/lint/generate-allowlist-candidates.sh"
PASS=0; FAIL=0

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT

# The generator reads ALLOWLIST relative to its SCRIPT_DIR (scripts/lint/codex-residual-allowlist.txt)
# We need to create a mock allowlist at the expected location relative to where we cd

# Test 1: Clean — all markers match allowlist
echo "Test 1: Clean markers (all allowlisted)..."
mkdir -p "$tmpdir/test1/skill1"
echo 'Use claude code for this' > "$tmpdir/test1/skill1/SKILL.md"
mkdir -p "$REPO_ROOT/scripts/lint"
ORIG_ALLOWLIST=""
if [[ -f "$REPO_ROOT/scripts/lint/codex-residual-allowlist.txt" ]]; then
  ORIG_ALLOWLIST=$(cat "$REPO_ROOT/scripts/lint/codex-residual-allowlist.txt")
fi
echo 'claude' > "$REPO_ROOT/scripts/lint/codex-residual-allowlist.txt"
if bash "$GENERATOR" "$tmpdir/test1"; then
  echo "  PASS"
  PASS=$((PASS+1))
else
  echo "  FAIL (expected exit 0)"
  FAIL=$((FAIL+1))
fi

# Test 2: Dirty — unallowlisted marker
echo "Test 2: Dirty markers (unallowlisted)..."
mkdir -p "$tmpdir/test2/skill2"
echo 'Invoke claude directly' > "$tmpdir/test2/skill2/SKILL.md"
echo 'NOMATCH' > "$REPO_ROOT/scripts/lint/codex-residual-allowlist.txt"
if bash "$GENERATOR" "$tmpdir/test2" >/dev/null 2>&1; then
  echo "  FAIL (expected exit 1)"
  FAIL=$((FAIL+1))
else
  echo "  PASS"
  PASS=$((PASS+1))
fi

# Test 3: No markers at all
echo "Test 3: No markers..."
mkdir -p "$tmpdir/test3/skill3"
echo 'No special markers here' > "$tmpdir/test3/skill3/SKILL.md"
echo 'claude' > "$REPO_ROOT/scripts/lint/codex-residual-allowlist.txt"
if bash "$GENERATOR" "$tmpdir/test3"; then
  echo "  PASS"
  PASS=$((PASS+1))
else
  echo "  FAIL (expected exit 0)"
  FAIL=$((FAIL+1))
fi

# Restore original allowlist
if [[ -n "$ORIG_ALLOWLIST" ]]; then
  echo "$ORIG_ALLOWLIST" > "$REPO_ROOT/scripts/lint/codex-residual-allowlist.txt"
fi

echo ""
echo "Results: $PASS passed, $FAIL failed"
[[ $FAIL -eq 0 ]] && exit 0 || exit 1
