#!/usr/bin/env bash
# PreToolUse hook: run short Go tests before git commit.
# Blocking (exit 2) if tests fail — prevents committing broken code.
set -euo pipefail

TOOL_NAME="${CLAUDE_TOOL_NAME:-}"
COMMAND="${CLAUDE_TOOL_INPUT_COMMAND:-}"

# Only trigger on Bash tool
[[ "$TOOL_NAME" == "Bash" ]] || exit 0

# Only trigger on git commit commands
case "$COMMAND" in
  *"git commit"*) ;;
  *) exit 0 ;;
esac

# Need go compiler
command -v go &>/dev/null || exit 0

# Must be in a Go project (has go.mod)
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null)" || exit 0
GO_MOD=""
for candidate in "$REPO_ROOT/go.mod" "$REPO_ROOT/cli/go.mod"; do
  if [[ -f "$candidate" ]]; then
    GO_MOD="$candidate"
    break
  fi
done
[[ -n "$GO_MOD" ]] || exit 0

GO_DIR="$(dirname "$GO_MOD")"

echo "Running pre-commit Go tests..." >&2
if ! (cd "$GO_DIR" && go test ./... -count=1 -short 2>&1 | tail -5) >&2; then
  echo "BLOCKED: Go tests failed. Fix before committing." >&2
  exit 2
fi

exit 0
