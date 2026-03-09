#!/usr/bin/env bash
# preflight-refs.sh — Check required reference docs exist for post-mortem skill.
#
# Exit codes:
#   0 = all refs present (silent)
#   1 = missing refs + --strict flag (BLOCK)
#   2 = missing refs without --strict (WARN, non-blocking)
#
# Flags:
#   --strict                  Promote WARN to BLOCK (exit 1) on missing refs
#   --skip-checkpoint-policy  Skip checkpoint-policy.md check
#   --json                    Output structured JSON instead of plain text
set -euo pipefail

SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"

REQUIRED_REFS=(
  "skills/post-mortem/references/checkpoint-policy.md"
  "skills/post-mortem/references/metadata-verification.md"
  "skills/post-mortem/references/closure-integrity-audit.md"
)

# --- Parse flags ---
strict=false
skip_checkpoint_policy=false
json_output=false

for arg in "$@"; do
  case "$arg" in
    --strict)                 strict=true ;;
    --skip-checkpoint-policy) skip_checkpoint_policy=true ;;
    --json)                   json_output=true ;;
    *) echo "WARN: unknown flag: $arg" >&2 ;;
  esac
done

# --- Check refs ---
missing=0
skipped=0
missing_list=()
skipped_list=()

for ref in "${REQUIRED_REFS[@]}"; do
  # Apply skip guard clause
  if [[ "$skip_checkpoint_policy" == true ]] && [[ "$ref" == *"checkpoint-policy.md" ]]; then
    skipped=$((skipped + 1))
    skipped_list+=("$ref")
    continue
  fi

  if [ ! -f "$ref" ]; then
    missing=$((missing + 1))
    missing_list+=("$ref")
    if [ "$json_output" = false ]; then
      echo "WARN: missing required reference: $ref"
    fi
  fi
done

# --- Output ---
if [ "$json_output" = true ]; then
  # Build JSON arrays
  missing_json="["
  first=true
  for r in "${missing_list[@]+"${missing_list[@]}"}"; do
    [ "$first" = true ] && first=false || missing_json+=","
    missing_json+="\"$r\""
  done
  missing_json+="]"

  skipped_json="["
  first=true
  for r in "${skipped_list[@]+"${skipped_list[@]}"}"; do
    [ "$first" = true ] && first=false || skipped_json+=","
    skipped_json+="\"$r\""
  done
  skipped_json+="]"

  echo "{\"missing\":${missing},\"skipped\":${skipped},\"missing_refs\":${missing_json},\"skipped_refs\":${skipped_json}}"
fi

# --- Exit ---
if [ "$missing" -gt 0 ]; then
  if [ "$json_output" = false ]; then
    echo "WARN: post-mortem reference preflight incomplete (${missing} missing)."
  fi
  if [ "$strict" = true ]; then
    exit 1
  else
    exit 2
  fi
fi

exit 0
