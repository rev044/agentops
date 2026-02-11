#!/bin/bash
# test-hooks.sh - Integration tests for hook scripts
# Validates that hooks produce valid output and handle edge cases.
# Usage: ./tests/hooks/test-hooks.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
HOOKS_DIR="$REPO_ROOT/hooks"
PASS=0
FAIL=0

red() { printf '\033[0;31m%s\033[0m\n' "$1"; }
green() { printf '\033[0;32m%s\033[0m\n' "$1"; }

pass() {
    green "  PASS: $1"
    PASS=$((PASS + 1))
}

fail() {
    red "  FAIL: $1"
    FAIL=$((FAIL + 1))
}

# Pre-flight: jq required
if ! command -v jq >/dev/null 2>&1; then
    red "ERROR: jq is required for hook tests"
    exit 1
fi

TMPDIR=$(mktemp -d)
REPO_FIXTURE_DIR="$REPO_ROOT/.agents/ao/test-hooks-$$"
mkdir -p "$REPO_FIXTURE_DIR"
trap 'rm -rf "$TMPDIR" "$REPO_FIXTURE_DIR"' EXIT

# ============================================================
echo "=== prompt-nudge.sh ==="
# ============================================================

# Test 1: Empty prompt => silent exit
OUTPUT=$(echo '{"prompt":""}' | AGENTOPS_HOOKS_DISABLED=0 bash "$HOOKS_DIR/prompt-nudge.sh" 2>&1 || true)
if [ -z "$OUTPUT" ]; then pass "empty prompt produces no output"; else fail "empty prompt produces no output"; fi

# Test 2: Kill switch disables hook
OUTPUT=$(echo '{"prompt":"implement a feature"}' | AGENTOPS_HOOKS_DISABLED=1 bash "$HOOKS_DIR/prompt-nudge.sh" 2>&1 || true)
if [ -z "$OUTPUT" ]; then pass "kill switch disables hook"; else fail "kill switch disables hook"; fi

# Test 3: No chain.jsonl => silent exit
OUTPUT=$(echo '{"prompt":"implement something"}' | bash "$HOOKS_DIR/prompt-nudge.sh" 2>&1 || true)
if [ -z "$OUTPUT" ]; then pass "no chain.jsonl produces no output"; else fail "no chain.jsonl produces no output"; fi

# Test 4: jq -n produces valid JSON with special characters
SAFE_JSON=$(jq -n --arg nudge 'Test "nudge" with <special> chars & more' '{"hookSpecificOutput":{"additionalContext":$nudge}}')
if echo "$SAFE_JSON" | jq . >/dev/null 2>&1; then pass "jq -n produces valid JSON with special chars"; else fail "jq -n produces valid JSON with special chars"; fi

# Test 5: JSON injection resistance - special characters in nudge
for PAYLOAD in '"' '\\' '$(whoami)' '`id`' '<script>' "'; DROP TABLE" '{"nested":"json"}'; do
    RESULT=$(jq -n --arg nudge "$PAYLOAD" '{"hookSpecificOutput":{"additionalContext":$nudge}}')
    if echo "$RESULT" | jq -e '.hookSpecificOutput.additionalContext' >/dev/null 2>&1; then
        pass "jq escapes payload: $PAYLOAD"
    else
        fail "jq escapes payload: $PAYLOAD"
    fi
done

# ============================================================
echo ""
echo "=== push-gate.sh ==="
# ============================================================

# Test 6: Non-git command => pass through
echo '{"tool_input":{"command":"ls -la"}}' | AGENTOPS_HOOKS_DISABLED=0 bash "$HOOKS_DIR/push-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "non-git command passes through"; else fail "non-git command passes through"; fi

# Test 7: Kill switch allows git push
echo '{"tool_input":{"command":"git push origin main"}}' | AGENTOPS_HOOKS_DISABLED=1 bash "$HOOKS_DIR/push-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "kill switch allows git push"; else fail "kill switch allows git push"; fi

# Test 8: Worker exemption allows git push
echo '{"tool_input":{"command":"git push origin main"}}' | AGENTOPS_WORKER=1 bash "$HOOKS_DIR/push-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "worker exemption allows git push"; else fail "worker exemption allows git push"; fi

# Test 9: Empty command => pass through
echo '{"tool_input":{"command":""}}' | bash "$HOOKS_DIR/push-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "empty command passes through"; else fail "empty command passes through"; fi

# Test 10: Mock chain.jsonl (vibe+post-mortem done) => allow push
MOCK_PUSH="$TMPDIR/mock-push"
mkdir -p "$MOCK_PUSH/.agents/ao" "$MOCK_PUSH/.git/refs" "$MOCK_PUSH/.git/objects"
echo 'ref: refs/heads/main' > "$MOCK_PUSH/.git/HEAD"
echo '{"gate":"vibe","status":"locked"}' > "$MOCK_PUSH/.agents/ao/chain.jsonl"
echo '{"gate":"post-mortem","status":"locked"}' >> "$MOCK_PUSH/.agents/ao/chain.jsonl"
EC=0
(cd "$MOCK_PUSH" && echo '{"tool_input":{"command":"git push origin main"}}' | bash "$HOOKS_DIR/push-gate.sh" >/dev/null 2>&1) || EC=$?
if [ "$EC" -eq 0 ]; then pass "git push allowed when vibe+post-mortem done"; else fail "git push allowed when vibe+post-mortem done (exit=$EC)"; fi

# Test 11: Mock chain.jsonl (vibe pending) => block push
MOCK_BLOCK="$TMPDIR/mock-block"
mkdir -p "$MOCK_BLOCK/.agents/ao" "$MOCK_BLOCK/.git/refs" "$MOCK_BLOCK/.git/objects"
echo 'ref: refs/heads/main' > "$MOCK_BLOCK/.git/HEAD"
echo '{"gate":"vibe","status":"pending"}' > "$MOCK_BLOCK/.agents/ao/chain.jsonl"
OUTPUT=$(cd "$MOCK_BLOCK" && echo '{"tool_input":{"command":"git push origin main"}}' | bash "$HOOKS_DIR/push-gate.sh" 2>&1 || true)
if echo "$OUTPUT" | grep -q "BLOCKED"; then pass "git push blocked when vibe pending"; else fail "git push blocked when vibe pending"; fi

# ============================================================
echo ""
echo "=== session-start.sh / precompact-snapshot.sh ==="
# ============================================================

# Test 12: session-start emits valid JSON
SESSION_JSON=$(bash "$HOOKS_DIR/session-start.sh" 2>/dev/null || true)
if echo "$SESSION_JSON" | jq -e '.hookSpecificOutput.hookEventName == "SessionStart"' >/dev/null 2>&1; then
    pass "session-start emits SessionStart JSON"
else
    fail "session-start emits SessionStart JSON"
fi

# Test 13: session-start kill switch suppresses output
OUTPUT=$(AGENTOPS_HOOKS_DISABLED=1 bash "$HOOKS_DIR/session-start.sh" 2>&1 || true)
if [ -z "$OUTPUT" ]; then pass "session-start kill switch suppresses output"; else fail "session-start kill switch suppresses output"; fi

# Test 14: session-start roots .agents paths to git root from subdir
MOCK_SESSION="$TMPDIR/mock-session"
mkdir -p "$MOCK_SESSION/subdir"
git -C "$MOCK_SESSION" init -q >/dev/null 2>&1
(cd "$MOCK_SESSION/subdir" && bash "$HOOKS_DIR/session-start.sh" >/dev/null 2>&1 || true)
if [ -d "$MOCK_SESSION/.agents/research" ] && [ ! -d "$MOCK_SESSION/subdir/.agents/research" ]; then
    pass "session-start writes .agents to repo root"
else
    fail "session-start writes .agents to repo root"
fi

# Test 15: precompact emits JSON when data exists
MOCK_PRECOMPACT="$TMPDIR/mock-precompact"
mkdir -p "$MOCK_PRECOMPACT/.agents"
git -C "$MOCK_PRECOMPACT" init -q >/dev/null 2>&1
PRECOMPACT_JSON=$(cd "$MOCK_PRECOMPACT" && bash "$HOOKS_DIR/precompact-snapshot.sh" 2>/dev/null || true)
if echo "$PRECOMPACT_JSON" | jq -e '.hookSpecificOutput.additionalContext' >/dev/null 2>&1; then
    pass "precompact emits additionalContext JSON"
else
    fail "precompact emits additionalContext JSON"
fi

# Test 16: precompact kill switch suppresses output
OUTPUT=$(cd "$MOCK_PRECOMPACT" && AGENTOPS_HOOKS_DISABLED=1 bash "$HOOKS_DIR/precompact-snapshot.sh" 2>&1 || true)
if [ -z "$OUTPUT" ]; then pass "precompact kill switch suppresses output"; else fail "precompact kill switch suppresses output"; fi

# ============================================================
echo ""
echo "=== pending-cleaner.sh ==="
# ============================================================

# Test 17: fail-open outside git repo
NON_GIT="$TMPDIR/non-git"
mkdir -p "$NON_GIT"
EC=0
(cd "$NON_GIT" && bash "$HOOKS_DIR/pending-cleaner.sh" >/dev/null 2>&1) || EC=$?
if [ "$EC" -eq 0 ]; then pass "pending-cleaner fail-open outside git repo"; else fail "pending-cleaner fail-open outside git repo"; fi

# Test 18: stale pending.jsonl auto-clears and logs alerts
MOCK_PENDING="$TMPDIR/mock-pending"
mkdir -p "$MOCK_PENDING"
git -C "$MOCK_PENDING" init -q >/dev/null 2>&1
mkdir -p "$MOCK_PENDING/.agents/ao"
PENDING_FILE="$MOCK_PENDING/.agents/ao/pending.jsonl"
printf '{"session":"1"}\n{"session":"2"}\n' > "$PENDING_FILE"
touch -t 202401010101 "$PENDING_FILE"
(cd "$MOCK_PENDING" && AGENTOPS_PENDING_STALE_SECONDS=1 AGENTOPS_PENDING_ALERT_LINES=1 bash "$HOOKS_DIR/pending-cleaner.sh" >/dev/null 2>&1 || true)
if [ ! -s "$PENDING_FILE" ]; then pass "stale pending.jsonl auto-cleared"; else fail "stale pending.jsonl auto-cleared"; fi
if ls "$MOCK_PENDING/.agents/ao/archive"/pending-*.jsonl >/dev/null 2>&1; then pass "stale pending.jsonl archived before clear"; else fail "stale pending.jsonl archived before clear"; fi
if grep -q 'ALERT stale pending.jsonl detected' "$MOCK_PENDING/.agents/ao/hook-errors.log" 2>/dev/null && grep -q 'AUTOCLEAR stale pending.jsonl' "$MOCK_PENDING/.agents/ao/hook-errors.log" 2>/dev/null; then
    pass "pending-cleaner logs stale alert and autoclear telemetry"
else
    fail "pending-cleaner logs stale alert and autoclear telemetry"
fi

# Test 19: pending-cleaner kill switch prevents auto-clear
printf '{"session":"keep"}\n' > "$PENDING_FILE"
touch -t 202401010101 "$PENDING_FILE"
(cd "$MOCK_PENDING" && AGENTOPS_HOOKS_DISABLED=1 AGENTOPS_PENDING_STALE_SECONDS=1 bash "$HOOKS_DIR/pending-cleaner.sh" >/dev/null 2>&1 || true)
if [ -s "$PENDING_FILE" ]; then pass "pending-cleaner kill switch preserves queue"; else fail "pending-cleaner kill switch preserves queue"; fi

# ============================================================
echo ""
echo "=== task-validation-gate.sh ==="
# ============================================================

REPO_CONTENT_FILE="$REPO_FIXTURE_DIR/test-content.js"
REPO_REGEX_FILE="$REPO_FIXTURE_DIR/test-regex.txt"
echo "function authenticate() {}" > "$REPO_CONTENT_FILE"
echo 'hello.*world' > "$REPO_REGEX_FILE"

# Test 20: No validation metadata => pass
echo '{"metadata":{}}' | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "no validation metadata passes"; else fail "no validation metadata passes"; fi

# Test 21: Global kill switch
echo '{"metadata":{"validation":{"files_exist":["/nonexistent"]}}}' | AGENTOPS_HOOKS_DISABLED=1 bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "global kill switch passes validation"; else fail "global kill switch passes validation"; fi

# Test 22: Hook-specific kill switch
echo '{"metadata":{"validation":{"files_exist":["/nonexistent"]}}}' | AGENTOPS_TASK_VALIDATION_DISABLED=1 bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "task-validation kill switch passes validation"; else fail "task-validation kill switch passes validation"; fi

# Test 23: files_exist - existing repo file (relative path)
INPUT=$(jq -n '{"metadata":{"validation":{"files_exist":["hooks/push-gate.sh"]}}}')
echo "$INPUT" | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "files_exist with existing repo file passes"; else fail "files_exist with existing repo file passes"; fi

# Test 24: files_exist - missing file blocks
EC=0
echo '{"metadata":{"validation":{"files_exist":["hooks/nonexistent-file-12345.sh"]}}}' | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1 || EC=$?
if [ "$EC" -eq 2 ]; then pass "files_exist with missing file blocks (exit 2)"; else fail "files_exist with missing file blocks (exit=$EC, expected 2)"; fi

# Test 25: files_exist blocks path traversal outside repo root
EC=0
echo '{"metadata":{"validation":{"files_exist":["../README.md"]}}}' | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1 || EC=$?
if [ "$EC" -eq 2 ]; then pass "files_exist blocks path traversal outside repo"; else fail "files_exist blocks path traversal outside repo (exit=$EC, expected 2)"; fi

# Test 26: content_check - pattern found within repo root
INPUT=$(jq -n --arg f "$REPO_CONTENT_FILE" '{"metadata":{"validation":{"content_check":[{"file":$f,"pattern":"function authenticate"}]}}}')
echo "$INPUT" | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "content_check with matching pattern passes"; else fail "content_check with matching pattern passes"; fi

# Test 27: content_check - pattern not found
INPUT=$(jq -n --arg f "$REPO_CONTENT_FILE" '{"metadata":{"validation":{"content_check":[{"file":$f,"pattern":"class UserService"}]}}}')
EC=0
echo "$INPUT" | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1 || EC=$?
if [ "$EC" -eq 2 ]; then pass "content_check with missing pattern blocks (exit 2)"; else fail "content_check with missing pattern blocks (exit=$EC, expected 2)"; fi

# Test 28: content_check - regex injection safe (grep -qF literal match)
INPUT=$(jq -n --arg f "$REPO_REGEX_FILE" '{"metadata":{"validation":{"content_check":[{"file":$f,"pattern":"hello.*world"}]}}}')
echo "$INPUT" | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "content_check treats regex chars as literals"; else fail "content_check treats regex chars as literals"; fi

# Test 29: content_check blocks files outside repo root
echo "outside" > "$TMPDIR/outside.txt"
INPUT=$(jq -n --arg f "$TMPDIR/outside.txt" '{"metadata":{"validation":{"content_check":[{"file":$f,"pattern":"outside"}]}}}')
EC=0
echo "$INPUT" | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1 || EC=$?
if [ "$EC" -eq 2 ]; then pass "content_check blocks paths outside repo"; else fail "content_check blocks paths outside repo (exit=$EC, expected 2)"; fi

# Test 30: Allowlist blocks disallowed commands
EC=0
echo '{"metadata":{"validation":{"tests":"curl http://evil.com"}}}' | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1 || EC=$?
if [ "$EC" -eq 2 ]; then pass "allowlist blocks curl command"; else fail "allowlist blocks curl command (exit=$EC, expected 2)"; fi

# Test 31: Allowlist allows go command
echo '{"metadata":{"validation":{"tests":"go version"}}}' | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "allowlist allows go command"; else fail "allowlist allows go command"; fi

# ============================================================
echo ""
echo "=== validate-hook-preflight.sh ==="
# ============================================================

# Test 32: Hook preflight validator passes
if "$REPO_ROOT/scripts/validate-hook-preflight.sh" >/dev/null 2>&1; then
    pass "validate-hook-preflight.sh passes"
else
    fail "validate-hook-preflight.sh passes"
fi

# ============================================================
echo ""
echo "=== Results ==="
# ============================================================

TOTAL=$((PASS + FAIL))
echo "Total: $TOTAL | Pass: $PASS | Fail: $FAIL"

if [ "$FAIL" -gt 0 ]; then
    red "FAILED"
    exit 1
else
    green "ALL PASSED"
    exit 0
fi
