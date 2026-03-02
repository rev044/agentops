#!/usr/bin/env bash
# PostToolUse hook: run go vet on modified Go files after edits.
# Non-blocking (exit 0) — logs warnings to stderr for visibility.
set -euo pipefail

TOOL_NAME="${CLAUDE_TOOL_NAME:-}"
FILE_PATH="${CLAUDE_TOOL_INPUT_FILE_PATH:-}"

# Only trigger on Edit/Write
case "$TOOL_NAME" in
  Edit|Write) ;;
  *) exit 0 ;;
esac

# Only .go files
[[ "$FILE_PATH" == *.go ]] || exit 0

# Need go compiler
command -v go &>/dev/null || exit 0

# Find the package directory for the modified file
PKG_DIR="$(dirname "$FILE_PATH")"
[[ -d "$PKG_DIR" ]] || exit 0

# Run go vet on the package
VETOUT=$(cd "$PKG_DIR" && go vet ./... 2>&1) || true
if [[ -n "$VETOUT" ]]; then
  echo "go vet warning in $(basename "$PKG_DIR"):" >&2
  echo "$VETOUT" | head -20 >&2
fi

exit 0
