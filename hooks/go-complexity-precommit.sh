#!/usr/bin/env bash
# PostToolUse hook: check cyclomatic complexity of modified Go files.
# Non-blocking (exit 0) — logs warnings to stderr for visibility.
set -euo pipefail

# Only trigger on Edit/Write tool use against .go files
TOOL_NAME="${CLAUDE_TOOL_NAME:-}"
FILE_PATH="${CLAUDE_TOOL_INPUT_FILE_PATH:-}"

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
