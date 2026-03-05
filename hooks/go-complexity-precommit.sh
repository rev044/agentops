#!/usr/bin/env bash
# PostToolUse hook: check cyclomatic complexity of modified Go files.
# Non-blocking (exit 0) — logs warnings to stderr for visibility.
set -euo pipefail

# Validate a command is in an allowlist before executing.
# Usage: _validate_restricted_cmd "command_string" allowed_array
# Returns 1 if the command binary is not in the allowlist.
_validate_restricted_cmd() {
    local cmd="$1"
    shift
    local -a allowlist=("$@")
    local binary
    binary=$(echo "$cmd" | awk '{print $1}')
    for allowed in "${allowlist[@]}"; do
        if [ "$binary" = "$allowed" ]; then
            return 0
        fi
    done
    echo "BLOCKED: '$binary' not in allowlist: ${allowlist[*]}" >&2
    return 1
}

# Only trigger on Edit/Write tool use against .go files
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // ""' 2>/dev/null) || TOOL_NAME=""
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // ""' 2>/dev/null) || FILE_PATH=""

# Normalize: strip repo root prefix if path is absolute
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
if [[ "$FILE_PATH" == /* ]]; then
  FILE_PATH="${FILE_PATH#"$REPO_ROOT"/}"
fi

case "$TOOL_NAME" in
  Edit|Write) ;;
  *) exit 0 ;;
esac

[[ "$FILE_PATH" == *.go ]] || exit 0
[[ "$FILE_PATH" == cli/* ]] || exit 0

# Skip test files
[[ "$FILE_PATH" != *_test.go ]] || exit 0

if ! command -v gocyclo &>/dev/null; then
  exit 0
fi

# Check complexity of the modified file
VIOLATIONS=$(gocyclo -over 15 "$FILE_PATH" 2>/dev/null || true)
if [[ -n "$VIOLATIONS" ]]; then
  echo "Complexity warning in $FILE_PATH:" >&2
  echo "$VIOLATIONS" >&2
  echo "Consider extracting helpers to reduce CC below 15." >&2
fi

# Always exit 0 — this is advisory, not blocking
exit 0
