#!/usr/bin/env bash
set -euo pipefail

# Flywheel Proof Run: Verifies knowledge compounds across sessions.
# Exit 0 = PASS, exit 1 = FAIL

# Setup: Create isolated test directory
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT

mkdir -p "$TEST_DIR/.agents/learnings"

# Session 1: Seed a learning manually
cat > "$TEST_DIR/.agents/learnings/proof-learning-1.md" << 'LEARNING'
---
utility: 0.8
source_bead: proof-test
source_phase: validate
---
# Proof Run Learning
When testing flywheel compounding, always verify inject retrieves prior session learnings.
LEARNING

# Session 2: Verify inject retrieves it
RESULT=$(cd "$TEST_DIR" && ao inject "flywheel compounding" --format json --no-cite 2>/dev/null) || {
    echo "FAIL: ao inject command failed"
    exit 1
}

COUNT=$(echo "$RESULT" | jq '.learnings | length' 2>/dev/null) || {
    echo "FAIL: Could not parse inject output as JSON"
    echo "Raw output: $RESULT"
    exit 1
}

if [ "$COUNT" -ge 1 ]; then
    echo "PASS: Session 2 retrieved $COUNT learnings from session 1"
else
    echo "FAIL: Session 2 retrieved 0 learnings"
    exit 1
fi

# Session 3: Verify scoring is operational (with decay)
RESULT2=$(cd "$TEST_DIR" && ao inject "flywheel compounding" --apply-decay --format json --no-cite 2>/dev/null) || {
    echo "FAIL: ao inject --apply-decay command failed"
    exit 1
}

SCORE=$(echo "$RESULT2" | jq '.learnings[0].composite_score // 0' 2>/dev/null) || SCORE=0
if awk "BEGIN{exit !($SCORE > 0)}"; then
    echo "PASS: Scoring and decay operational (score=$SCORE)"
else
    echo "FAIL: Scoring not working (score=$SCORE)"
    exit 1
fi

echo ""
echo "FLYWHEEL PROOF: PASS"
echo "Knowledge compounds across sessions."
