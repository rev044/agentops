#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if eval "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "SKILL.md has name: crank" "grep -q '^name: crank' '$SKILL_DIR/SKILL.md'"
check "FIRE loop documented" "grep -q 'FIRE' '$SKILL_DIR/SKILL.md'"
check "No phantom bd cook refs" "! grep -q 'bd cook' '$SKILL_DIR/SKILL.md'"
check "No phantom gt convoy refs" "! grep -q 'gt convoy' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
