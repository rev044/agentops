#!/bin/bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if eval "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "SKILL.md has name: evolve" "grep -q '^name: evolve' '$SKILL_DIR/SKILL.md'"
check "references/ directory exists" "[ -d '$SKILL_DIR/references' ]"
check "references/ has at least 1 file" "[ \$(ls '$SKILL_DIR/references/' | wc -l) -ge 1 ]"
check "SKILL.md mentions kill switch" "grep -qi 'kill switch' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions fitness" "grep -qi 'fitness' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions GOALS.yaml" "grep -q 'GOALS.yaml' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions cycle" "grep -qi 'cycle' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions /rpi" "grep -q '/rpi' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
