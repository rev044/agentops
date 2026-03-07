#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "warning: scripts/check-skill-flag-refs.sh is deprecated; use scripts/validate-skill-cli-snippets.sh" >&2
exec "$SCRIPT_DIR/validate-skill-cli-snippets.sh" "$@"
