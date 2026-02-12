#!/bin/bash
# task-validation-gate.sh - TaskCompleted hook: validate task metadata before completion
# Reads task JSON from stdin, checks metadata.validation rules.
# Exit 0 = pass (or no validation). Exit 2 = block completion.

# Kill switch
[ "${AGENTOPS_HOOKS_DISABLED:-}" = "1" ] && exit 0
[ "${AGENTOPS_TASK_VALIDATION_DISABLED:-}" = "1" ] && exit 0

# Read all stdin
INPUT=$(cat)

# Require jq — fail open without it
if ! command -v jq >/dev/null 2>&1; then
    exit 0
fi

# Error log directory (repo-local)
ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
ROOT="$(cd "$ROOT" 2>/dev/null && pwd -P 2>/dev/null || printf '%s' "$ROOT")"
ERROR_LOG_DIR="$ROOT/.agents/ao"
ERROR_LOG="$ERROR_LOG_DIR/hook-errors.log"

# Execute validations from repo root so relative paths are predictable.
cd "$ROOT" 2>/dev/null || true

# Restricted command execution: allowlist-based, no shell interpretation
run_restricted() {
    local cmd="$1"

    # Block shell metacharacters — prevents injection via crafted metadata
    if [[ "$cmd" =~ [\;\|\&\`\$\(\)\<\>] ]]; then
        log_error "BLOCKED: shell metacharacters in command: $cmd"
        echo "VALIDATION BLOCKED: shell metacharacters not allowed in command" >&2
        exit 2
    fi

    # Split command string into array (word-split on whitespace)
    read -ra cmd_parts <<< "$cmd"
    local binary="${cmd_parts[0]}"

    # Binary must be a bare name (no path separators)
    if [[ "$binary" == */* ]]; then
        log_error "BLOCKED: path in binary name: $binary (full: $cmd)"
        echo "VALIDATION BLOCKED: binary must be a bare name, not a path" >&2
        exit 2
    fi

    # Strict allowlist of permitted binaries
    local allowed="go pytest npm npx make bash"
    local found=0
    for a in $allowed; do
        if [ "$binary" = "$a" ]; then
            found=1
            break
        fi
    done
    if [ "$found" -ne 1 ]; then
        log_error "BLOCKED: command not in allowlist: $binary (full: $cmd)"
        echo "VALIDATION BLOCKED: command '$binary' not in allowlist ($allowed)" >&2
        exit 2
    fi

    # Block bash -c (shell interpretation escape)
    if [ "$binary" = "bash" ]; then
        for arg in "${cmd_parts[@]:1}"; do
            if [ "$arg" = "-c" ]; then
                log_error "BLOCKED: bash -c not allowed: $cmd"
                echo "VALIDATION BLOCKED: 'bash -c' not allowed (use direct commands)" >&2
                exit 2
            fi
        done
    fi

    # Execute as array — no shell interpretation
    "${cmd_parts[@]}" >/dev/null 2>&1
}

log_error() {
    mkdir -p "$ERROR_LOG_DIR" 2>/dev/null
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) task-validation-gate: $1" >> "$ERROR_LOG" 2>/dev/null
}

# Write structured failure details to last-failure.json
write_failure() {
    local type="$1"
    local command="$2"
    local exit_code="$3"
    local details="$4"

    mkdir -p "$ERROR_LOG_DIR" 2>/dev/null

    # Extract task subject from INPUT
    local task_subject
    task_subject=$(echo "$INPUT" | jq -r '.subject // "unknown"' 2>/dev/null)
    [ -z "$task_subject" ] || [ "$task_subject" = "null" ] && task_subject="unknown"

    local ts
    ts=$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")

    # Try jq first (preferred)
    if command -v jq >/dev/null 2>&1; then
        jq -n \
            --arg ts "$ts" \
            --arg type "$type" \
            --arg command "$command" \
            --argjson exit_code "$exit_code" \
            --arg task_subject "$task_subject" \
            --arg details "$details" \
            '{ts:$ts,type:$type,command:$command,exit_code:$exit_code,task_subject:$task_subject,details:$details}' \
            > "$ERROR_LOG_DIR/last-failure.json" 2>/dev/null
    else
        # Fallback: use printf and basic escaping (escape quotes and backslashes)
        local escaped_command escaped_subject escaped_details
        escaped_command=$(printf '%s' "$command" | sed 's/["\\]/\\&/g')
        escaped_subject=$(printf '%s' "$task_subject" | sed 's/["\\]/\\&/g')
        escaped_details=$(printf '%s' "$details" | sed 's/["\\]/\\&/g')

        printf '{"ts":"%s","type":"%s","command":"%s","exit_code":%d,"task_subject":"%s","details":"%s"}\n' \
            "$ts" "$type" "$escaped_command" "$exit_code" "$escaped_subject" "$escaped_details" \
            > "$ERROR_LOG_DIR/last-failure.json" 2>/dev/null
    fi
}

# Resolve user-provided file paths to repo-rooted absolute paths.
# Returns non-zero if path escapes ROOT or cannot be normalized.
resolve_repo_path() {
    local raw_path="$1"
    local candidate dir base normalized_dir normalized_path

    [ -n "$raw_path" ] || return 1
    case "$raw_path" in
        *$'\n'*|*$'\r'*) return 1 ;;
    esac

    if [[ "$raw_path" = /* ]]; then
        candidate="$raw_path"
    else
        candidate="$ROOT/$raw_path"
    fi

    dir=$(dirname -- "$candidate")
    base=$(basename -- "$candidate")
    normalized_dir=$(cd "$dir" 2>/dev/null && pwd -P) || return 1
    normalized_path="$normalized_dir/$base"

    case "$normalized_path" in
        "$ROOT"|"$ROOT"/*)
            printf '%s\n' "$normalized_path"
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# Extract metadata.validation — fail open on parse errors
if ! VALIDATION=$(echo "$INPUT" | jq -r '.metadata.validation // empty' 2>/dev/null); then
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
            if [ -n "$FILE_PATH" ] && [ "$FILE_PATH" != "null" ]; then
                RESOLVED_FILE=$(resolve_repo_path "$FILE_PATH") || {
                    log_error "blocked files_exist path outside repo root: $FILE_PATH"
                    write_failure "files_exist" "resolve_repo_path" 1 "path escapes repo root: $FILE_PATH"
                    echo "VALIDATION FAILED: files_exist — path escapes repo root: $FILE_PATH" >&2
                    exit 2
                }
                if [ ! -f "$RESOLVED_FILE" ]; then
                    # Collect all missing files from this check
                    MISSING_FILES="$FILE_PATH"
                    for j in $(seq $((i + 1)) $((FILE_COUNT - 1))); do
                        NEXT_FILE=$(echo "$FILES_EXIST" | jq -r ".[$j]" 2>/dev/null)
                        if [ -n "$NEXT_FILE" ] && [ "$NEXT_FILE" != "null" ]; then
                            NEXT_RESOLVED=$(resolve_repo_path "$NEXT_FILE" 2>/dev/null) || continue
                            if [ ! -f "$NEXT_RESOLVED" ]; then
                                MISSING_FILES="$MISSING_FILES, $NEXT_FILE"
                            fi
                        fi
                    done
                    write_failure "files_exist" "test -f" 1 "missing files: $MISSING_FILES"
                    echo "VALIDATION FAILED: files_exist — missing files: $MISSING_FILES" >&2
                    exit 2
                fi
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
                RESOLVED_CHECK_FILE=$(resolve_repo_path "$CHECK_FILE") || {
                    log_error "blocked content_check path outside repo root: $CHECK_FILE"
                    write_failure "content_check" "resolve_repo_path" 1 "path escapes repo root: $CHECK_FILE"
                    echo "VALIDATION FAILED: content_check — path escapes repo root: $CHECK_FILE" >&2
                    exit 2
                }
                if ! grep -qF "$CHECK_PATTERN" "$RESOLVED_CHECK_FILE" 2>/dev/null; then
                    write_failure "content_check" "grep" 1 "pattern '$CHECK_PATTERN' not found in file $CHECK_FILE"
                    echo "VALIDATION FAILED: content_check — pattern '$CHECK_PATTERN' not found in file $CHECK_FILE" >&2
                    echo "  Expected pattern: $CHECK_PATTERN" >&2
                    echo "  File: $CHECK_FILE" >&2
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
        write_failure "test" "$TESTS_CMD" "$?" "test command failed"
        echo "VALIDATION FAILED: tests — command failed: $TESTS_CMD" >&2
        echo "  Suggested: /bug-hunt --test-failure .agents/ao/last-failure.json" >&2
        exit 2
    fi
fi

# 4. lint: command string
LINT_CMD=$(echo "$VALIDATION" | jq -r '.lint // empty' 2>/dev/null)
if [ -n "$LINT_CMD" ] && [ "$LINT_CMD" != "null" ]; then
    if ! run_restricted "$LINT_CMD"; then
        write_failure "lint" "$LINT_CMD" "$?" "lint command failed"
        echo "VALIDATION FAILED: lint — command failed: $LINT_CMD" >&2
        echo "  Suggested: /bug-hunt --test-failure .agents/ao/last-failure.json" >&2
        exit 2
    fi
fi

# 5. command: command string
GENERIC_CMD=$(echo "$VALIDATION" | jq -r '.command // empty' 2>/dev/null)
if [ -n "$GENERIC_CMD" ] && [ "$GENERIC_CMD" != "null" ]; then
    if ! run_restricted "$GENERIC_CMD"; then
        write_failure "command" "$GENERIC_CMD" "$?" "command failed"
        echo "VALIDATION FAILED: command — command failed: $GENERIC_CMD" >&2
        echo "  Suggested: /bug-hunt --test-failure .agents/ao/last-failure.json" >&2
        exit 2
    fi
fi

# All checks passed
exit 0
