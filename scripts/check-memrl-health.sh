#!/usr/bin/env bash
set -euo pipefail

# check-memrl-health.sh
# Validates that the MemRL (Memory Reinforcement Learning) feedback loop is
# properly wired end-to-end. Catches disconnects between maturity scanning,
# utility tracking, citation feedback, session hooks, and close-loop transitions.
#
# Each check verifies one critical link in the feedback chain. A broken link
# means learnings silently stop maturing or utility scores stop updating.
#
# Exit 0: all links healthy
# Exit 1: one or more links broken

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"

failures=0

pass() { echo "  PASS: $1"; }
fail() { echo "  FAIL: $1"; failures=$((failures + 1)); }

echo "=== MemRL Feedback Loop Health Check ==="
echo ""

# ── Check A: Maturity scanner globs both *.md and *.jsonl ──
# Without both globs, the scanner misses half the learnings corpus.
if grep -q 'globLearningFiles' "$ROOT/cli/internal/ratchet/maturity.go" && \
   grep -q '"[*].md"' "$ROOT/cli/internal/ratchet/maturity.go" 2>/dev/null || \
   grep -q '\.md' "$ROOT/cli/internal/ratchet/maturity.go"; then
    # More precise: verify glob function handles both extensions
    if grep -q '\.jsonl' "$ROOT/cli/internal/ratchet/maturity.go" && \
       grep -q '\.md' "$ROOT/cli/internal/ratchet/maturity.go"; then
        pass "Maturity scanner globs both *.md and *.jsonl"
    else
        fail "Maturity scanner missing glob for *.md or *.jsonl (cli/internal/ratchet/maturity.go)"
    fi
else
    fail "Maturity scanner missing globLearningFiles function (cli/internal/ratchet/maturity.go)"
fi

# ── Check B: Markdown utility updater tracks helpful_count ──
# Without helpful_count, the maturity ratchet has no signal to promote learnings.
if grep -q 'helpful_count' "$ROOT/cli/cmd/ao/feedback.go"; then
    pass "Feedback updater tracks helpful_count"
else
    fail "Feedback updater missing helpful_count tracking (cli/cmd/ao/feedback.go)"
fi

# ── Check C: Citation feedback does NOT hardcode reward=1.0 ──
# A hardcoded reward defeats the purpose of RL — every citation gets the same score.
if grep -q 'updateLearningUtility.*1\.0' "$ROOT/cli/cmd/ao/flywheel_citation_feedback.go"; then
    fail "Citation feedback hardcodes reward=1.0 — defeats RL signal (cli/cmd/ao/flywheel_citation_feedback.go)"
else
    pass "Citation feedback does not hardcode reward=1.0"
fi

# ── Check D: SessionEnd hook has --scan --apply ──
# Without --scan --apply, maturity transitions never fire at session boundaries.
if grep -q '\-\-scan.*\-\-apply' "$ROOT/hooks/session-end-maintenance.sh"; then
    pass "SessionEnd hook invokes maturity --scan --apply"
else
    fail "SessionEnd hook missing --scan --apply (hooks/session-end-maintenance.sh)"
fi

# ── Check E: Close-loop calls applyAllMaturityTransitions ──
# The close-loop command must call the transition function or maturity is dead code.
if grep -q 'applyAllMaturityTransitions' "$ROOT/cli/cmd/ao/flywheel_close_loop.go"; then
    pass "Close-loop calls applyAllMaturityTransitions"
else
    fail "Close-loop missing applyAllMaturityTransitions call (cli/cmd/ao/flywheel_close_loop.go)"
fi

# ── Check F: Dedup command exists ──
# Without dedup, duplicate learnings accumulate and pollute utility scores.
if [[ -f "$ROOT/cli/cmd/ao/dedup.go" ]]; then
    pass "Dedup command exists"
else
    fail "Dedup command missing (cli/cmd/ao/dedup.go)"
fi

# ── Summary ──
echo ""
if [[ "$failures" -gt 0 ]]; then
    echo "FAILED: $failures MemRL feedback loop disconnect(s) detected"
    exit 1
fi

echo "OK: All MemRL feedback loop links verified"
exit 0
