#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

echo "=== Skill runtime format validation ==="

echo "--- Claude/cloud skill format ---"
bash ./tests/skills/lint-skills.sh

echo "--- Codex skill format ---"
bash ./scripts/lint-codex-native.sh --strict

echo "--- Codex runtime sections ---"
bash ./scripts/validate-codex-runtime-sections.sh

echo "Skill runtime format validation passed."
