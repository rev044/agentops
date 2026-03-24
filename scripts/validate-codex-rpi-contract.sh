#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

failures=0

require_contains() {
  local file="$1"
  local needle="$2"
  local message="$3"
  if ! grep -Fq -- "$needle" "$file"; then
    echo "FAIL: $message" >&2
    echo "  missing: $needle" >&2
    echo "  file: $file" >&2
    failures=$((failures + 1))
  fi
}

require_not_contains() {
  local file="$1"
  local needle="$2"
  local message="$3"
  if grep -Fq -- "$needle" "$file"; then
    echo "FAIL: $message" >&2
    echo "  unexpected: $needle" >&2
    echo "  file: $file" >&2
    failures=$((failures + 1))
  fi
}

echo "=== Codex RPI contract validation ==="

require_contains "skills-codex/rpi/SKILL.md" '$crank .agents/rpi/execution-packet.json' \
  'rpi must define the no-beads implementation handoff through execution-packet.json'
require_contains "skills-codex/rpi/SKILL.md" '$validation --complexity=<level>' \
  'rpi must define standalone validation when no epic_id exists'
require_not_contains "skills-codex/rpi/SKILL.md" '$crank <objective-id>' \
  'rpi must not use an undefined objective-id handoff'

require_contains "skills-codex/crank/SKILL.md" 'Given `$crank [epic-id | .agents/rpi/execution-packet.json | plan-file.md | "description"]`:' \
  'crank must accept execution-packet.json as a first-class input'
require_contains "skills-codex/crank/SKILL.md" '**Execution-packet/file mode:**' \
  'crank must define execution-packet/file-backed behavior explicitly'

require_contains "skills-codex/rpi/references/phase-data-contracts.md" '$crank .agents/rpi/execution-packet.json' \
  'phase-data contracts must document the no-beads discovery-to-implementation handoff'
require_contains "skills-codex/rpi/references/phase-data-contracts.md" 'standalone `$validation`' \
  'phase-data contracts must document standalone validation when no epic exists'

require_contains "skills-codex/discovery/references/output-templates.md" 'this execution packet becomes the' \
  'discovery output template must explain file-backed handoff when no epic is created'

if [[ $failures -ne 0 ]]; then
  echo "Codex RPI contract validation failed with $failures issue(s)." >&2
  exit 1
fi

echo "Codex RPI contract validation passed."
