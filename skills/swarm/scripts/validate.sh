#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if eval "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "SKILL.md has name: swarm" "grep -q '^name: swarm' '$SKILL_DIR/SKILL.md'"
check "Local mode documented" "grep -q 'Local' '$SKILL_DIR/SKILL.md'"
check "Distributed mode documented" "grep -q 'Distributed' '$SKILL_DIR/SKILL.md'"
check "TeamCreate lifecycle documented" "grep -q 'TeamCreate' '$SKILL_DIR/SKILL.md'"
check "TeamDelete lifecycle documented" "grep -q 'TeamDelete' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
