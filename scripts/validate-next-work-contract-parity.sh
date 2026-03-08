#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

usage() {
  cat <<'EOF'
Usage: bash scripts/validate-next-work-contract-parity.sh [repo-root]
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -gt 1 ]]; then
  usage >&2
  exit 2
fi

if [[ $# -eq 1 ]]; then
  ROOT="$1"
fi

if [[ "$ROOT" != /* ]]; then
  ROOT="$(cd "$ROOT" && pwd)"
fi

SCHEMA="$ROOT/.agents/rpi/next-work.schema.md"
HARVEST_REF="$ROOT/skills/post-mortem/references/harvest-next-work.md"
POST_MORTEM_SKILL="$ROOT/skills/post-mortem/SKILL.md"
POST_MORTEM_CODEX_SKILL="$ROOT/skills-codex/post-mortem/SKILL.md"
PHASE_CONTRACT="$ROOT/skills/rpi/references/phase-data-contracts.md"
GATE4="$ROOT/skills/rpi/references/gate4-loop-and-spawn.md"
RUNTIME="$ROOT/cli/cmd/ao/rpi_loop.go"
SMOKE="$ROOT/tests/smoke-test.sh"

failures=0

fail() {
  echo "FAIL: $1" >&2
  failures=$((failures + 1))
}

have_rg() {
  command -v rg >/dev/null 2>&1
}

require_file() {
  local path="$1"
  [[ -f "$path" ]] || fail "missing file: ${path#$ROOT/}"
}

contains_fixed_file() {
  local needle="$1"
  local path="$2"

  if have_rg; then
    rg -Fq "$needle" "$path"
    return
  fi

  grep -Fq -- "$needle" "$path"
}

contains_fixed_stdin() {
  local needle="$1"

  if have_rg; then
    rg -Fq "$needle"
    return
  fi

  grep -Fq -- "$needle"
}

require_contains() {
  local path="$1"
  local needle="$2"
  local label="$3"
  if ! contains_fixed_file "$needle" "$path"; then
    fail "$label"
  fi
}

require_section_contains() {
  local path="$1"
  local start="$2"
  local end="$3"
  local needle="$4"
  local label="$5"
  local section

  section="$(awk -v start="$start" -v end="$end" '
    index($0, start) {in_section=1}
    in_section {print}
    in_section && index($0, end) {exit}
  ' "$path")"

  if [[ -z "$section" ]]; then
    fail "missing section $start in ${path#$ROOT/}"
    return
  fi

  if ! printf '%s\n' "$section" | contains_fixed_stdin "$needle"; then
    fail "$label"
  fi
}

require_section_not_contains() {
  local path="$1"
  local start="$2"
  local end="$3"
  local needle="$4"
  local label="$5"
  local section

  section="$(awk -v start="$start" -v end="$end" '
    index($0, start) {in_section=1}
    in_section {print}
    in_section && index($0, end) {exit}
  ' "$path")"

  if [[ -z "$section" ]]; then
    fail "missing section $start in ${path#$ROOT/}"
    return
  fi

  if printf '%s\n' "$section" | contains_fixed_stdin "$needle"; then
    fail "$label"
  fi
}

for path in \
  "$SCHEMA" \
  "$HARVEST_REF" \
  "$POST_MORTEM_SKILL" \
  "$POST_MORTEM_CODEX_SKILL" \
  "$PHASE_CONTRACT" \
  "$GATE4" \
  "$RUNTIME" \
  "$SMOKE"; do
  require_file "$path"
done

require_contains "$SCHEMA" "schema_version: 1.3" \
  "next-work schema is not at v1.3"
require_contains "$SCHEMA" 'Item lifecycle inside `items[]` is authoritative.' \
  "next-work schema must declare item lifecycle authority"
require_contains "$SCHEMA" "may be empty when a post-mortem finds nothing actionable" \
  "next-work schema must permit empty items arrays"
require_contains "$SCHEMA" "consumers may rewrite existing lines to claim, release, fail, or consume individual items" \
  "next-work schema must describe rewrite semantics"
require_contains "$SCHEMA" "Legacy Compatibility" \
  "next-work schema must document legacy flat rows"

entry_fields=(
  source_epic timestamp items consumed claim_status claimed_by claimed_at
  consumed_by consumed_at failed_at
)
item_fields=(
  title type severity source description evidence target_repo consumed
  claim_status claimed_by claimed_at consumed_by consumed_at failed_at
)
item_types=(
  tech-debt improvement pattern-fix process-improvement feature bug task
)
item_sources=(
  council-finding retro-learning retro-pattern evolve-generator
  feature-suggestion backlog-processing
)

for field in "${entry_fields[@]}"; do
  require_contains "$SCHEMA" "\`$field\`" \
    "next-work schema missing entry field \`$field\`"
  require_contains "$RUNTIME" "json:\"$field" \
    "runtime next-work structs missing json field $field"
done

for field in "${item_fields[@]}"; do
  require_contains "$SCHEMA" "\`$field\`" \
    "next-work schema missing item field \`$field\`"
done

for value in "${item_types[@]}"; do
  require_contains "$SCHEMA" "\`$value\`" \
    "next-work schema missing type enum value $value"
  require_contains "$HARVEST_REF" "$value" \
    "harvest-next-work reference missing type enum value $value"
done

for value in high medium low; do
  require_contains "$SCHEMA" "\`$value\`" \
    "next-work schema missing severity enum value $value"
done

for value in "${item_sources[@]}"; do
  require_contains "$SCHEMA" "\`$value\`" \
    "next-work schema missing source enum value $value"
  require_contains "$HARVEST_REF" "$value" \
    "harvest-next-work reference missing source enum value $value"
done

for value in available in_progress consumed; do
  require_contains "$SCHEMA" "\`$value\`" \
    "next-work schema missing claim_status value $value"
  require_contains "$RUNTIME" "\"$value\"" \
    "runtime next-work logic missing claim_status value $value"
done

require_contains "$HARVEST_REF" ".agents/rpi/next-work.schema.md" \
  "harvest-next-work must reference the tracked next-work schema"
require_contains "$POST_MORTEM_SKILL" ".agents/rpi/next-work.schema.md" \
  "post-mortem skill must reference the tracked next-work schema"
require_contains "$POST_MORTEM_CODEX_SKILL" ".agents/rpi/next-work.schema.md" \
  "generated Codex post-mortem skill must reference the tracked next-work schema"
require_contains "$GATE4" ".agents/rpi/next-work.schema.md" \
  "rpi gate4 reference must point at the tracked next-work schema"
require_contains "$PHASE_CONTRACT" "item lifecycle as authoritative" \
  "phase-data-contracts must describe item lifecycle authority"
require_contains "$PHASE_CONTRACT" 'entry aggregate flips to `consumed=true` only after every child item is consumed' \
  "phase-data-contracts must describe aggregate consumption rule"
require_contains "$GATE4" "Never mark an item consumed at pick-time" \
  "rpi gate4 must retain claim-before-consume rule"
require_contains "$SMOKE" "validate-next-work-contract-parity.sh" \
  "smoke-test must execute the next-work contract parity validator"

require_contains "$RUNTIME" "case \"feature\", \"improvement\", \"tech-debt\", \"pattern-fix\", \"bug\", \"task\":" \
  "RPI runtime is missing workTypeRank coverage for pattern-fix"
require_contains "$RUNTIME" 'omitted item `claim_status` semantically' \
  "runtime comments or docs should preserve omitted claim_status semantics"

for skill in "$POST_MORTEM_SKILL" "$POST_MORTEM_CODEX_SKILL"; do
  require_section_contains "$skill" '#### Step ACT.3: Feed Next-Work' '#### Step ACT.4: Update Marker' \
    ".agents/rpi/next-work.schema.md" \
    "${skill#$ROOT/} ACT.3 must reference the tracked next-work schema"
  require_section_contains "$skill" '#### Step ACT.3: Feed Next-Work' '#### Step ACT.4: Update Marker' \
    "source_epic:" \
    "${skill#$ROOT/} ACT.3 must show batched next-work entries"
  require_section_contains "$skill" '#### Step ACT.3: Feed Next-Work' '#### Step ACT.4: Update Marker' \
    "items:" \
    "${skill#$ROOT/} ACT.3 must show batched item arrays"
  require_section_contains "$skill" '#### Step ACT.3: Feed Next-Work' '#### Step ACT.4: Update Marker' \
    'claim_status: "available"' \
    "${skill#$ROOT/} ACT.3 must initialize entry claim_status"
  require_section_not_contains "$skill" '#### Step ACT.3: Feed Next-Work' '#### Step ACT.4: Update Marker' \
    'echo "{\"title\":' \
    "${skill#$ROOT/} ACT.3 still contains the legacy flat-row append example"
done

if [[ "$failures" -gt 0 ]]; then
  echo "next-work contract parity validation FAILED ($failures finding(s))." >&2
  exit 1
fi

echo "next-work contract parity validation passed."
