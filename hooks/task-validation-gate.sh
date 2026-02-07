#!/bin/bash
# task-validation-gate.sh - TaskCompleted hook: validate task metadata before completion
# Reads task JSON from stdin, checks metadata.validation rules.
# Exit 0 = pass (or no validation). Exit 2 = block completion.

# Kill switch
[ "${AGENTOPS_HOOKS_DISABLED:-}" = "1" ] && exit 0

# Read all stdin
INPUT=$(cat)

# Require jq — fail open without it
if ! command -v jq >/dev/null 2>&1; then
    exit 0
fi

# Error log directory (repo-local)
ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo ".")
ERROR_LOG_DIR="$ROOT/.agents/ao"
ERROR_LOG="$ERROR_LOG_DIR/hook-errors.log"

# Restricted command execution: only allow simple commands, no shell metacharacters
run_restricted() {
    local cmd="$1"
    # Block shell metacharacters that enable injection
    if echo "$cmd" | grep -qE '[;&|`$(){}]'; then
        log_error "BLOCKED: shell metacharacters in validation command: $cmd"
        echo "VALIDATION BLOCKED: command contains disallowed shell characters" >&2
        exit 2
    fi
    # Execute without eval — word-split the command naturally
    /bin/sh -c "$cmd" >/dev/null 2>&1
}

log_error() {
    mkdir -p "$ERROR_LOG_DIR" 2>/dev/null
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) task-validation-gate: $1" >> "$ERROR_LOG" 2>/dev/null
}

# Extract metadata.validation — fail open on parse errors
VALIDATION=$(echo "$INPUT" | jq -r '.metadata.validation // empty' 2>/dev/null)
if [ $? -ne 0 ]; then
    log_error "JSON parse error on stdin"
    exit 0
fi

# No validation metadata → pass through
if [ -z "$VALIDATION" ] || [ "$VALIDATION" = "null" ]; then
    exit 0
fi

# --- Validation checks ---

# 1. files_exist: array of paths
FILES_EXIST=$(echo "$VALIDATION" | jq -r '.files_exist // empty' 2>/dev/null)
if [ -n "$FILES_EXIST" ] && [ "$FILES_EXIST" != "null" ]; then
    FILE_COUNT=$(echo "$FILES_EXIST" | jq -r 'length' 2>/dev/null)
    if [ -n "$FILE_COUNT" ] && [ "$FILE_COUNT" -gt 0 ] 2>/dev/null; then
        for i in $(seq 0 $((FILE_COUNT - 1))); do
            FILE_PATH=$(echo "$FILES_EXIST" | jq -r ".[$i]" 2>/dev/null)
            if [ -n "$FILE_PATH" ] && [ "$FILE_PATH" != "null" ] && [ ! -f "$FILE_PATH" ]; then
                echo "VALIDATION FAILED: files_exist — $FILE_PATH not found" >&2
                exit 2
            fi
        done
    fi
fi

# 2. content_check: array of {file, pattern}
CONTENT_CHECKS=$(echo "$VALIDATION" | jq -r '.content_check // empty' 2>/dev/null)
if [ -n "$CONTENT_CHECKS" ] && [ "$CONTENT_CHECKS" != "null" ]; then
    CHECK_COUNT=$(echo "$CONTENT_CHECKS" | jq -r 'length' 2>/dev/null)
    if [ -n "$CHECK_COUNT" ] && [ "$CHECK_COUNT" -gt 0 ] 2>/dev/null; then
        for i in $(seq 0 $((CHECK_COUNT - 1))); do
            CHECK_FILE=$(echo "$CONTENT_CHECKS" | jq -r ".[$i].file" 2>/dev/null)
            CHECK_PATTERN=$(echo "$CONTENT_CHECKS" | jq -r ".[$i].pattern" 2>/dev/null)
            if [ -n "$CHECK_FILE" ] && [ "$CHECK_FILE" != "null" ] && [ -n "$CHECK_PATTERN" ] && [ "$CHECK_PATTERN" != "null" ]; then
                if ! grep -q "$CHECK_PATTERN" "$CHECK_FILE" 2>/dev/null; then
                    echo "VALIDATION FAILED: content_check — pattern '$CHECK_PATTERN' not found in $CHECK_FILE" >&2
                    exit 2
                fi
            fi
        done
    fi
fi

# 3. tests: command string
TESTS_CMD=$(echo "$VALIDATION" | jq -r '.tests // empty' 2>/dev/null)
if [ -n "$TESTS_CMD" ] && [ "$TESTS_CMD" != "null" ]; then
    if ! run_restricted "$TESTS_CMD"; then
        echo "VALIDATION FAILED: tests — command failed: $TESTS_CMD" >&2
        exit 2
    fi
fi

# 4. lint: command string
LINT_CMD=$(echo "$VALIDATION" | jq -r '.lint // empty' 2>/dev/null)
if [ -n "$LINT_CMD" ] && [ "$LINT_CMD" != "null" ]; then
    if ! run_restricted "$LINT_CMD"; then
        echo "VALIDATION FAILED: lint — command failed: $LINT_CMD" >&2
        exit 2
    fi
fi

# 5. command: command string
GENERIC_CMD=$(echo "$VALIDATION" | jq -r '.command // empty' 2>/dev/null)
if [ -n "$GENERIC_CMD" ] && [ "$GENERIC_CMD" != "null" ]; then
    if ! run_restricted "$GENERIC_CMD"; then
        echo "VALIDATION FAILED: command — command failed: $GENERIC_CMD" >&2
        exit 2
    fi
fi

# All checks passed
exit 0
