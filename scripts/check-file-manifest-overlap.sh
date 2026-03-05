#!/usr/bin/env bash
set -euo pipefail

# check-file-manifest-overlap.sh — Detect file ownership conflicts in swarm task manifests
# Input: JSON file path (array of {id, subject, files[]})
# Exit 0 if no conflicts, exit 1 if overlapping files detected

# Requires bash 4+ for associative arrays
if [[ "${BASH_VERSINFO[0]}" -lt 4 ]]; then
  echo "WARN: bash 4+ required (have ${BASH_VERSION}) — skipping manifest overlap check"
  exit 0
fi

INPUT="${1:--}"
if [[ "$INPUT" == "-" ]]; then INPUT="/dev/stdin"; fi

if ! command -v jq &>/dev/null; then
  echo "WARN: jq not found — skipping manifest overlap check"
  exit 0
fi

CONFLICTS=0
declare -A FILE_OWNERS

while IFS=$'\t' read -r task_id file; do
  if [[ -n "${FILE_OWNERS[$file]:-}" ]]; then
    echo "CONFLICT: $file claimed by task ${FILE_OWNERS[$file]} and task $task_id"
    CONFLICTS=$((CONFLICTS + 1))
  else
    FILE_OWNERS["$file"]="$task_id"
  fi
done < <(jq -r '.[] | .id as $id | .files[]? | [$id, .] | @tsv' "$INPUT")

if [[ $CONFLICTS -gt 0 ]]; then
  echo "Found $CONFLICTS file overlap conflict(s)"
  exit 1
fi

echo "No file manifest overlaps detected"
exit 0
