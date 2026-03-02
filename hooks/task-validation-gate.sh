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

    # Block shell metacharacters and control chars — prevents injection via crafted metadata
    # Note: newline checked separately — \n inside [...] matches literal 'n' in ERE
    if [[ "$cmd" == *$'\n'* ]] || [[ "$cmd" =~ [\;\|\&\`\$\(\)\<\>\'\"\\\] ]]; then
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
    # NOTE: npx removed (downloads+executes arbitrary npm packages = RCE)
    # NOTE: bash removed (bash <script> bypasses -c block = arbitrary execution)
    local allowed="go pytest npm make"
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

    # Execute as array — no shell interpretation
    "${cmd_parts[@]}" >/dev/null 2>&1
}

# Like run_restricted but captures stdout+stderr instead of discarding.
# Used by behavioral assertions to check output patterns.
run_restricted_capture() {
    local cmd="$1"

    # Block shell metacharacters and control chars
    if [[ "$cmd" == *$'\n'* ]] || [[ "$cmd" =~ [\;\|\&\`\$\(\)\<\>\'\"\\\] ]]; then
        log_error "BLOCKED: shell metacharacters in assertion command: $cmd"
        echo "VALIDATION BLOCKED: shell metacharacters not allowed in assertion command" >&2
        return 1
    fi

    read -ra cmd_parts <<< "$cmd"
    local binary="${cmd_parts[0]}"

    # Binary must be a bare name
    if [[ "$binary" == */* ]]; then
        log_error "BLOCKED: path in assertion binary: $binary"
        echo "VALIDATION BLOCKED: binary must be a bare name" >&2
        return 1
    fi

    # Strict allowlist
    local allowed="go pytest npm make"
    local found=0
    for a in $allowed; do
        if [ "$binary" = "$a" ]; then
            found=1
            break
        fi
    done
    if [ "$found" -ne 1 ]; then
        log_error "BLOCKED: assertion command not in allowlist: $binary"
        echo "VALIDATION BLOCKED: command '$binary' not in allowlist ($allowed)" >&2
        return 1
    fi

    # Execute and capture output
    "${cmd_parts[@]}" 2>&1
}

log_error() {
    mkdir -p "$ERROR_LOG_DIR" 2>/dev/null
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) task-validation-gate: $1" >> "$ERROR_LOG" 2>/dev/null
}

# Source hook-helpers from plugin install dir, not repo root (security: prevents malicious repo sourcing)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/hook-helpers.sh
. "$SCRIPT_DIR/../lib/hook-helpers.sh"

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

collect_changed_files() {
    {
        git diff --cached --name-only --diff-filter=ACMR 2>/dev/null || true
        git diff --name-only --diff-filter=ACMR 2>/dev/null || true
        git ls-files --others --exclude-standard 2>/dev/null || true
    } | awk 'NF' | sort -u
}

derive_companion_path() {
    local source_file="$1"
    local companion_template="$2"
    local source_dir source_name source_basename source_ext companion

    source_dir=$(dirname -- "$source_file")
    source_name=$(basename -- "$source_file")
    source_basename="$source_name"
    source_ext=""
    if [[ "$source_name" == *.* ]]; then
        source_basename="${source_name%.*}"
        source_ext=".${source_name##*.}"
    fi

    companion="$companion_template"
    companion="${companion//\{dir\}/$source_dir}"
    companion="${companion//\{basename\}/$source_basename}"
    companion="${companion//\{ext\}/$source_ext}"
    companion="${companion#./}"
    while [[ "$companion" == *"//"* ]]; do
        companion="${companion//\/\//\/}"
    done
    printf '%s\n' "$companion"
}

is_impl_issue_type() {
    case "$1" in
        feature|bug|task) return 0 ;;
        *) return 1 ;;
    esac
}

is_validation_exempt_issue_type() {
    case "$1" in
        docs|chore|ci) return 0 ;;
        *) return 1 ;;
    esac
}

# Extract metadata.validation — fail open on parse errors
if ! VALIDATION=$(echo "$INPUT" | jq -r '.metadata.validation // empty' 2>/dev/null); then
    log_error "JSON parse error on stdin"
    exit 0
fi

# Resolve issue type from common payload locations.
ISSUE_TYPE=$(echo "$INPUT" | jq -r '
    (
        .issue_type // .type // .kind //
        .metadata.issue_type // .metadata.type // .metadata.kind //
        .issue.issue_type // .issue.type // .issue.kind // empty
    )
    | if type == "string" then ascii_downcase | gsub("^\\s+|\\s+$"; "") else "" end
' 2>/dev/null)

# No validation metadata.
if [ -z "$VALIDATION" ] || [ "$VALIDATION" = "null" ]; then
    if is_impl_issue_type "$ISSUE_TYPE"; then
        write_failure "validation_metadata" "metadata.validation" 1 "missing metadata.validation for issue_type '$ISSUE_TYPE'"
        echo "VALIDATION FAILED: metadata.validation is required for issue_type '$ISSUE_TYPE' (feature|bug|task)" >&2
        exit 2
    fi
    # Explicit exemption path for non-implementation tasks.
    if is_validation_exempt_issue_type "$ISSUE_TYPE"; then
        exit 0
    fi

    # Unknown/legacy tasks remain fail-open for backward compatibility.
    exit 0
fi

# Normalize common validation fields once for policy checks and execution.
FILES_EXIST=$(echo "$VALIDATION" | jq -c '
    if (.files_exist // null) == null then
        []
    elif (.files_exist | type) == "array" then
        .files_exist
    else
        []
    end
' 2>/dev/null)
FILES_EXIST="${FILES_EXIST:-[]}"

CONTENT_CHECKS=$(echo "$VALIDATION" | jq -c '
    if (.content_check // null) == null then
        []
    elif (.content_check | type) == "array" then
        .content_check
    elif (.content_check | type) == "object" then
        [ .content_check ]
    else
        []
    end
' 2>/dev/null)
CONTENT_CHECKS="${CONTENT_CHECKS:-[]}"

TESTS_CMD=$(echo "$VALIDATION" | jq -r '.tests // empty' 2>/dev/null)

# Strict policy for implementation tasks: tests + structural evidence required.
if is_impl_issue_type "$ISSUE_TYPE"; then
    if [ -z "$TESTS_CMD" ] || [ "$TESTS_CMD" = "null" ]; then
        write_failure "validation_metadata" "metadata.validation.tests" 1 "missing tests for issue_type '$ISSUE_TYPE'"
        echo "VALIDATION FAILED: metadata.validation.tests is required for issue_type '$ISSUE_TYPE'" >&2
        exit 2
    fi

    FILE_COUNT=$(echo "$FILES_EXIST" | jq -r 'length' 2>/dev/null)
    CHECK_COUNT=$(echo "$CONTENT_CHECKS" | jq -r 'length' 2>/dev/null)
    FILE_COUNT="${FILE_COUNT:-0}"
    CHECK_COUNT="${CHECK_COUNT:-0}"
    if [ "$FILE_COUNT" -eq 0 ] && [ "$CHECK_COUNT" -eq 0 ]; then
        write_failure "validation_metadata" "metadata.validation" 1 "missing structural checks for issue_type '$ISSUE_TYPE'"
        echo "VALIDATION FAILED: metadata.validation for issue_type '$ISSUE_TYPE' must include files_exist or content_check" >&2
        exit 2
    fi
fi

# --- Validation checks ---

# 1. files_exist: array of paths
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

# 3. paired_files: array of {pattern, exclude, companion, message}
PAIRED_RULES=$(echo "$VALIDATION" | jq -r '.paired_files // empty' 2>/dev/null)
if [ -n "$PAIRED_RULES" ] && [ "$PAIRED_RULES" != "null" ]; then
    RULE_COUNT=$(echo "$PAIRED_RULES" | jq -r 'length' 2>/dev/null)
    if [ -n "$RULE_COUNT" ] && [ "$RULE_COUNT" -gt 0 ] 2>/dev/null; then
        CHANGED_FILES=$(collect_changed_files)
        if [ -n "$CHANGED_FILES" ]; then
            for i in $(seq 0 $((RULE_COUNT - 1))); do
                RULE_PATTERN=$(echo "$PAIRED_RULES" | jq -r ".[$i].pattern // empty" 2>/dev/null)
                RULE_EXCLUDE=$(echo "$PAIRED_RULES" | jq -r ".[$i].exclude // empty" 2>/dev/null)
                RULE_COMPANION=$(echo "$PAIRED_RULES" | jq -r ".[$i].companion // empty" 2>/dev/null)
                RULE_MESSAGE=$(echo "$PAIRED_RULES" | jq -r ".[$i].message // empty" 2>/dev/null)

                if [ -z "$RULE_PATTERN" ] || [ -z "$RULE_COMPANION" ]; then
                    continue
                fi

                while IFS= read -r CHANGED_FILE; do
                    [ -n "$CHANGED_FILE" ] || continue

                    if [[ "$CHANGED_FILE" != $RULE_PATTERN ]]; then
                        continue
                    fi
                    if [ -n "$RULE_EXCLUDE" ] && [[ "$CHANGED_FILE" == $RULE_EXCLUDE ]]; then
                        continue
                    fi

                    DERIVED_COMPANION=$(derive_companion_path "$CHANGED_FILE" "$RULE_COMPANION")
                    RESOLVED_COMPANION=$(resolve_repo_path "$DERIVED_COMPANION") || {
                        log_error "blocked paired_files companion outside repo root: $DERIVED_COMPANION"
                        write_failure "paired_files" "resolve_repo_path" 1 "path escapes repo root: $DERIVED_COMPANION"
                        echo "VALIDATION FAILED: paired_files — path escapes repo root: $DERIVED_COMPANION" >&2
                        exit 2
                    }
                    REL_COMPANION=$(to_repo_relative_path "$RESOLVED_COMPANION")
                    REL_COMPANION="${REL_COMPANION#./}"

                    if ! printf '%s\n' "$CHANGED_FILES" | grep -Fx -- "$REL_COMPANION" >/dev/null 2>&1; then
                        FAIL_DETAIL="missing companion '$REL_COMPANION' for changed file '$CHANGED_FILE'"
                        write_failure "paired_files" "$RULE_PATTERN" 1 "$FAIL_DETAIL"
                        if [ -n "$RULE_MESSAGE" ]; then
                            echo "VALIDATION FAILED: paired_files — $RULE_MESSAGE" >&2
                        else
                            echo "VALIDATION FAILED: paired_files — $FAIL_DETAIL" >&2
                        fi
                        echo "  Changed file: $CHANGED_FILE" >&2
                        echo "  Expected companion: $REL_COMPANION" >&2
                        exit 2
                    fi
                done <<< "$CHANGED_FILES"
            done
        fi
    fi
fi

# 4. tests: command string
if [ -n "$TESTS_CMD" ] && [ "$TESTS_CMD" != "null" ]; then
    if ! run_restricted "$TESTS_CMD"; then
        write_failure "test" "$TESTS_CMD" "$?" "test command failed"
        echo "VALIDATION FAILED: tests — command failed: $TESTS_CMD" >&2
        echo "  Suggested: /bug-hunt --test-failure .agents/ao/last-failure.json" >&2
        exit 2
    fi
fi

# 5. lint: command string
LINT_CMD=$(echo "$VALIDATION" | jq -r '.lint // empty' 2>/dev/null)
if [ -n "$LINT_CMD" ] && [ "$LINT_CMD" != "null" ]; then
    if ! run_restricted "$LINT_CMD"; then
        write_failure "lint" "$LINT_CMD" "$?" "lint command failed"
        echo "VALIDATION FAILED: lint — command failed: $LINT_CMD" >&2
        echo "  Suggested: /bug-hunt --test-failure .agents/ao/last-failure.json" >&2
        exit 2
    fi
fi

# 6. command: command string
GENERIC_CMD=$(echo "$VALIDATION" | jq -r '.command // empty' 2>/dev/null)
if [ -n "$GENERIC_CMD" ] && [ "$GENERIC_CMD" != "null" ]; then
    if ! run_restricted "$GENERIC_CMD"; then
        write_failure "command" "$GENERIC_CMD" "$?" "command failed"
        echo "VALIDATION FAILED: command — command failed: $GENERIC_CMD" >&2
        echo "  Suggested: /bug-hunt --test-failure .agents/ao/last-failure.json" >&2
        exit 2
    fi
fi

# 7. assertions: array of {command, expect_pattern, description}
# Behavioral assertions — run command and check output for expected pattern.
# Closes the gap between "tests run" and "tests verify correct behavior."
ASSERTIONS=$(echo "$VALIDATION" | jq -c '.assertions // []' 2>/dev/null)
ASSERTION_COUNT=$(echo "$ASSERTIONS" | jq -r 'length' 2>/dev/null || echo 0)

if [ -n "$ASSERTION_COUNT" ] && [ "$ASSERTION_COUNT" -gt 0 ] 2>/dev/null; then
    for i in $(seq 0 $((ASSERTION_COUNT - 1))); do
        A_CMD=$(echo "$ASSERTIONS" | jq -r ".[$i].command" 2>/dev/null)
        A_PATTERN=$(echo "$ASSERTIONS" | jq -r ".[$i].expect_pattern" 2>/dev/null)
        A_DESC=$(echo "$ASSERTIONS" | jq -r ".[$i].description // \"assertion $i\"" 2>/dev/null)

        # Skip if command or pattern empty
        [ -z "$A_CMD" ] || [ "$A_CMD" = "null" ] && continue
        [ -z "$A_PATTERN" ] || [ "$A_PATTERN" = "null" ] && continue

        # Run through restricted executor with output capture
        A_OUTPUT=$(run_restricted_capture "$A_CMD") || {
            write_failure "behavioral_assertion" "$A_CMD" 2 "Command failed: $A_DESC"
            echo "VALIDATION FAILED: behavioral assertion '$A_DESC' — command failed: $A_CMD" >&2
            exit 2
        }

        # Check output for expected pattern (regex via grep -E)
        if ! echo "$A_OUTPUT" | grep -qE "$A_PATTERN"; then
            write_failure "behavioral_assertion" "$A_CMD" 2 "Expected pattern '$A_PATTERN' not found in output: $A_DESC"
            echo "VALIDATION FAILED: behavioral assertion '$A_DESC' — expected pattern '$A_PATTERN' not found in output" >&2
            echo "  Command: $A_CMD" >&2
            echo "  Pattern: $A_PATTERN" >&2
            exit 2
        fi
    done
fi

# 8. embedded_parity: auto-check when hook files changed
# Enforces hooks/ <-> cli/embedded/hooks/ parity after any hook modification.
CHANGED_FILES_LIST=$(collect_changed_files 2>/dev/null || true)
HOOK_CHANGED=0
if [ -n "$CHANGED_FILES_LIST" ]; then
    if echo "$CHANGED_FILES_LIST" | grep -qE '^(hooks/|lib/hook-helpers\.sh)'; then
        HOOK_CHANGED=1
    fi
fi

if [ "$HOOK_CHANGED" -eq 1 ]; then
    PARITY_SCRIPT="$ROOT/scripts/validate-embedded-sync.sh"
    if [ -f "$PARITY_SCRIPT" ] && [ -x "$PARITY_SCRIPT" ]; then
        if ! bash "$PARITY_SCRIPT" >/dev/null 2>&1; then
            write_failure "embedded_parity" "validate-embedded-sync.sh" 1 "hooks/ and cli/embedded/hooks/ are out of sync"
            echo "VALIDATION FAILED: embedded_parity — hooks and embedded copies are out of sync" >&2
            echo "  Run: cd cli && make sync-hooks" >&2
            exit 2
        fi
    fi
fi

# All checks passed
exit 0
