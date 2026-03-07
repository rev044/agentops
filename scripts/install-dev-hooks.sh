#!/usr/bin/env bash
# install-dev-hooks.sh — activate repo-managed git hooks for this worktree/repo.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
EXPECTED_HOOKS_PATH=".githooks"
CHECK_ONLY="false"

usage() {
  cat <<'EOF'
Usage: bash scripts/install-dev-hooks.sh [--check]

Options:
  --check     Verify that git uses the repo-managed .githooks path.
  -h, --help  Show this help message.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check)
      CHECK_ONLY="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

cd "$REPO_ROOT"

for hook in pre-commit pre-push; do
  if [[ ! -x "$REPO_ROOT/.githooks/$hook" ]]; then
    echo "Missing executable repo hook: .githooks/$hook" >&2
    exit 1
  fi
done

current_hooks_path="$(git config --local --get core.hooksPath || true)"

if [[ "$CHECK_ONLY" == "true" ]]; then
  if [[ "$current_hooks_path" == "$EXPECTED_HOOKS_PATH" ]]; then
    echo "Git hooks path OK: $current_hooks_path"
    exit 0
  fi
  echo "Git hooks path mismatch: expected $EXPECTED_HOOKS_PATH, got ${current_hooks_path:-<unset>}" >&2
  exit 1
fi

git config --local core.hooksPath "$EXPECTED_HOOKS_PATH"

updated_hooks_path="$(git config --local --get core.hooksPath || true)"
if [[ "$updated_hooks_path" != "$EXPECTED_HOOKS_PATH" ]]; then
  echo "Failed to set core.hooksPath to $EXPECTED_HOOKS_PATH" >&2
  exit 1
fi

echo "Configured git hooks path: $updated_hooks_path"
echo "pre-commit and pre-push will now run from .githooks/"
