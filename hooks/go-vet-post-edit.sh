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

# Assertion density check for test files
if [[ "$FILE_PATH" == *_test.go ]] && [[ -f "$FILE_PATH" ]]; then
  EMPTY_TESTS=""
  while IFS= read -r func_name; do
    [[ -z "$func_name" ]] && continue
    # Extract ~50 lines after function declaration, check for assertion patterns
    if ! grep -A 50 "func ${func_name}" "$FILE_PATH" 2>/dev/null | \
         grep -qE '(t\.(Fatal|Fatalf|Error|Errorf|Fail|FailNow)|assert\.|require\.|want.*got|got.*want|expected.*actual|actual.*expected)'; then
      EMPTY_TESTS="${EMPTY_TESTS}  ${func_name}\n"
    fi
  done < <(grep -oE 'func (Test[A-Za-z0-9_]+)' "$FILE_PATH" 2>/dev/null | awk '{print $2}')

  if [[ -n "$EMPTY_TESTS" ]]; then
    echo "Assertion density warning: test functions with no assertions detected:" >&2
    printf '%b' "$EMPTY_TESTS" >&2
    echo "  These tests will pass regardless of code behavior." >&2
  fi
fi

exit 0
