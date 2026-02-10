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
trap 'rm -rf "$TMPDIR"' EXIT

# ============================================================
echo "=== prompt-nudge.sh ==="
# ============================================================

# Test 1: Empty prompt → silent exit
OUTPUT=$(echo '{"prompt":""}' | AGENTOPS_HOOKS_DISABLED=0 bash "$HOOKS_DIR/prompt-nudge.sh" 2>&1 || true)
if [ -z "$OUTPUT" ]; then pass "empty prompt produces no output"; else fail "empty prompt produces no output"; fi

# Test 2: Kill switch disables hook
OUTPUT=$(echo '{"prompt":"implement a feature"}' | AGENTOPS_HOOKS_DISABLED=1 bash "$HOOKS_DIR/prompt-nudge.sh" 2>&1 || true)
if [ -z "$OUTPUT" ]; then pass "kill switch disables hook"; else fail "kill switch disables hook"; fi

# Test 3: No chain.jsonl → silent exit
OUTPUT=$(echo '{"prompt":"implement something"}' | bash "$HOOKS_DIR/prompt-nudge.sh" 2>&1 || true)
if [ -z "$OUTPUT" ]; then pass "no chain.jsonl produces no output"; else fail "no chain.jsonl produces no output"; fi

# Test 4: jq -n produces valid JSON with special characters
SAFE_JSON=$(jq -n --arg nudge 'Test "nudge" with <special> chars & more' '{"hookSpecificOutput":{"additionalContext":$nudge}}')
if echo "$SAFE_JSON" | jq . >/dev/null 2>&1; then pass "jq -n produces valid JSON with special chars"; else fail "jq -n produces valid JSON with special chars"; fi

# Test 5: JSON injection resistance — special characters in nudge
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

# Test 6: Non-git command → pass through
echo '{"tool_input":{"command":"ls -la"}}' | AGENTOPS_HOOKS_DISABLED=0 bash "$HOOKS_DIR/push-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "non-git command passes through"; else fail "non-git command passes through"; fi

# Test 7: Kill switch allows git push
echo '{"tool_input":{"command":"git push origin main"}}' | AGENTOPS_HOOKS_DISABLED=1 bash "$HOOKS_DIR/push-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "kill switch allows git push"; else fail "kill switch allows git push"; fi

# Test 8: Worker exemption allows git push
echo '{"tool_input":{"command":"git push origin main"}}' | AGENTOPS_WORKER=1 bash "$HOOKS_DIR/push-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "worker exemption allows git push"; else fail "worker exemption allows git push"; fi

# Test 9: Empty command → pass through
echo '{"tool_input":{"command":""}}' | bash "$HOOKS_DIR/push-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "empty command passes through"; else fail "empty command passes through"; fi

# Test 10: Mock chain.jsonl (vibe+post-mortem done) → allow push
MOCK_PUSH="$TMPDIR/mock-push"
mkdir -p "$MOCK_PUSH/.agents/ao" "$MOCK_PUSH/.git/refs" "$MOCK_PUSH/.git/objects"
echo 'ref: refs/heads/main' > "$MOCK_PUSH/.git/HEAD"
echo '{"gate":"vibe","status":"locked"}' > "$MOCK_PUSH/.agents/ao/chain.jsonl"
echo '{"gate":"post-mortem","status":"locked"}' >> "$MOCK_PUSH/.agents/ao/chain.jsonl"
EC=0
cd "$MOCK_PUSH" && echo '{"tool_input":{"command":"git push origin main"}}' | bash "$HOOKS_DIR/push-gate.sh" >/dev/null 2>&1 || EC=$?
cd "$REPO_ROOT"
if [ "$EC" -eq 0 ]; then pass "git push allowed when vibe+post-mortem done"; else fail "git push allowed when vibe+post-mortem done (exit=$EC)"; fi

# Test 11: Mock chain.jsonl (vibe pending) → block push
MOCK_BLOCK="$TMPDIR/mock-block"
mkdir -p "$MOCK_BLOCK/.agents/ao" "$MOCK_BLOCK/.git/refs" "$MOCK_BLOCK/.git/objects"
echo 'ref: refs/heads/main' > "$MOCK_BLOCK/.git/HEAD"
echo '{"gate":"vibe","status":"pending"}' > "$MOCK_BLOCK/.agents/ao/chain.jsonl"
EC=0
OUTPUT=$(cd "$MOCK_BLOCK" && echo '{"tool_input":{"command":"git push origin main"}}' | bash "$HOOKS_DIR/push-gate.sh" 2>&1 || true)
cd "$REPO_ROOT"
if echo "$OUTPUT" | grep -q "BLOCKED"; then pass "git push blocked when vibe pending"; else fail "git push blocked when vibe pending"; fi

# ============================================================
echo ""
echo "=== task-validation-gate.sh ==="
# ============================================================

# Test 12: No validation metadata → pass
echo '{"metadata":{}}' | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "no validation metadata passes"; else fail "no validation metadata passes"; fi

# Test 13: Kill switch
echo '{"metadata":{"validation":{"files_exist":["/nonexistent"]}}}' | AGENTOPS_HOOKS_DISABLED=1 bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "kill switch passes validation"; else fail "kill switch passes validation"; fi

# Test 14: files_exist — existing file
INPUT=$(jq -n --arg f "$HOOKS_DIR/push-gate.sh" '{"metadata":{"validation":{"files_exist":[$f]}}}')
echo "$INPUT" | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "files_exist with existing file passes"; else fail "files_exist with existing file passes"; fi

# Test 15: files_exist — missing file
EC=0
echo '{"metadata":{"validation":{"files_exist":["/tmp/nonexistent-file-12345"]}}}' | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1 || EC=$?
if [ "$EC" -eq 2 ]; then pass "files_exist with missing file blocks (exit 2)"; else fail "files_exist with missing file blocks (exit=$EC, expected 2)"; fi

# Test 16: content_check — pattern found (uses grep -qF, literal match)
echo "function authenticate() {}" > "$TMPDIR/test-content.js"
INPUT=$(jq -n --arg f "$TMPDIR/test-content.js" '{"metadata":{"validation":{"content_check":[{"file":$f,"pattern":"function authenticate"}]}}}')
echo "$INPUT" | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "content_check with matching pattern passes"; else fail "content_check with matching pattern passes"; fi

# Test 17: content_check — pattern not found
INPUT=$(jq -n --arg f "$TMPDIR/test-content.js" '{"metadata":{"validation":{"content_check":[{"file":$f,"pattern":"class UserService"}]}}}')
EC=0
echo "$INPUT" | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1 || EC=$?
if [ "$EC" -eq 2 ]; then pass "content_check with missing pattern blocks (exit 2)"; else fail "content_check with missing pattern blocks (exit=$EC, expected 2)"; fi

# Test 18: content_check — regex injection safe (grep -qF uses literal strings)
echo 'hello.*world' > "$TMPDIR/test-regex.txt"
INPUT=$(jq -n --arg f "$TMPDIR/test-regex.txt" '{"metadata":{"validation":{"content_check":[{"file":$f,"pattern":"hello.*world"}]}}}')
echo "$INPUT" | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "content_check treats regex chars as literals (grep -qF)"; else fail "content_check treats regex chars as literals (grep -qF)"; fi

# Test 19: Allowlist blocks disallowed commands
EC=0
echo '{"metadata":{"validation":{"tests":"curl http://evil.com"}}}' | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1 || EC=$?
if [ "$EC" -eq 2 ]; then pass "allowlist blocks curl command"; else fail "allowlist blocks curl command (exit=$EC, expected 2)"; fi

# Test 20: Allowlist allows go command
echo '{"metadata":{"validation":{"tests":"go version"}}}' | bash "$HOOKS_DIR/task-validation-gate.sh" >/dev/null 2>&1
if [ $? -eq 0 ]; then pass "allowlist allows go command"; else fail "allowlist allows go command"; fi

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
