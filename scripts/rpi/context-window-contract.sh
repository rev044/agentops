#!/usr/bin/env bash
set -euo pipefail

# Contract: deterministic context-window tooling is available for large-repo
# /rpi workflows and can initialize + traverse bounded shards.

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

has_rg() {
  command -v rg >/dev/null 2>&1
}

search_n() {
  local pattern="$1"
  local file="$2"
  if has_rg; then
    rg -n "$pattern" "$file"
  else
    grep -nE "$pattern" "$file"
  fi
}

search_q() {
  local pattern="$1"
  local file="$2"
  if has_rg; then
    rg -q "$pattern" "$file"
  else
    grep -qE "$pattern" "$file"
  fi
}

search_c() {
  local pattern="$1"
  local file="$2"
  if has_rg; then
    rg -c "$pattern" "$file"
  else
    grep -cE "$pattern" "$file"
  fi
}

# Support both GOALS.md (current) and GOALS.yaml (legacy)
if [ -f GOALS.md ]; then
  GOALS_FILE="GOALS.md"
elif [ -f GOALS.yaml ]; then
  GOALS_FILE="GOALS.yaml"
else
  echo "FAIL: neither GOALS.md nor GOALS.yaml found" >&2
  exit 1
fi

if [ "$GOALS_FILE" = "GOALS.md" ]; then
  # GOALS.md: first non-heading, non-empty line after "# Goals" is the mission
  mission_line="$(sed -n '/^# Goals/,/^##/{/^# Goals/d; /^##/d; /^$/d; p;}' "$GOALS_FILE" | head -n 1 || true)"
  if [ -z "$mission_line" ]; then
    echo "FAIL: GOALS.md mission statement is required (first line after # Goals)" >&2
    exit 1
  fi
  mission_value="$(printf '%s' "$mission_line" | sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//')"
  if [ "${#mission_value}" -lt 20 ]; then
    echo "FAIL: GOALS.md mission must be at least 20 chars" >&2
    exit 1
  fi
  # Check for at least one directive (### heading under ## Directives)
  directive_count="$(search_c '^### [0-9]+\.' "$GOALS_FILE" || true)"
  if [ "${directive_count:-0}" -lt 1 ]; then
    echo "FAIL: GOALS.md must include at least one numbered directive" >&2
    exit 1
  fi
else
  # Legacy GOALS.yaml path
  mission_line="$(search_n '^[[:space:]]*mission:[[:space:]]*' "$GOALS_FILE" | head -n 1 || true)"
  if [ -z "$mission_line" ]; then
    echo "FAIL: GOALS.yaml mission field is required" >&2
    exit 1
  fi
  mission_value="${mission_line#*:}"
  mission_value="$(printf '%s' "$mission_value" | sed -E "s/^[[:space:]]+//; s/[[:space:]]+$//; s/^['\"]//; s/['\"]$//")"
  if [ "${#mission_value}" -lt 20 ]; then
    echo "FAIL: GOALS.yaml mission must be at least 20 chars" >&2
    exit 1
  fi
  if ! search_q '^[[:space:]]*goals:[[:space:]]*$' "$GOALS_FILE"; then
    echo "FAIL: GOALS.yaml goals list is required" >&2
    exit 1
  fi
  goal_count="$(search_c '^[[:space:]]*-[[:space:]]*id:[[:space:]]*[^[:space:]]+' "$GOALS_FILE" || true)"
  if [ "${goal_count:-0}" -lt 1 ]; then
    echo "FAIL: GOALS.yaml goals list must include at least one id entry" >&2
    exit 1
  fi
fi

echo "PASS: goals schema contract ($GOALS_FILE)"

scripts/rpi/generate-context-shards.py \
  --max-units 80 \
  --max-bytes 300000 \
  --out .agents/rpi/context-shards/latest.json \
  --check \
  --quiet

scripts/rpi/init-shard-progress.py \
  --manifest .agents/rpi/context-shards/latest.json \
  --progress .agents/rpi/context-shards/progress.json \
  --quiet

scripts/rpi/run-shard.py \
  --manifest .agents/rpi/context-shards/latest.json \
  --progress .agents/rpi/context-shards/progress.json \
  --shard-id 1 \
  --limit 1 >/dev/null

echo "PASS: context-window contract"
