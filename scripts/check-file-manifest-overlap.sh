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
WARNINGS=0
declare -A FILE_OWNERS

# Validate JSON input is a non-empty array
TASK_COUNT=$(jq -r 'if type == "array" then length else 0 end' "$INPUT" 2>/dev/null || echo 0)
if [[ "$TASK_COUNT" -eq 0 ]]; then
  echo "SKIP: empty or invalid task array — nothing to check"
  exit 0
fi

# Warn about tasks missing file manifests
MISSING_MANIFEST=$(jq -r '.[] | select(.files == null or (.files | length) == 0) | .id // "unknown"' "$INPUT" 2>/dev/null)
if [[ -n "$MISSING_MANIFEST" ]]; then
  while IFS= read -r task_id; do
    echo "WARN: task $task_id has no file manifest — cannot detect overlaps for this task"
    WARNINGS=$((WARNINGS + 1))
  done <<< "$MISSING_MANIFEST"
fi

while IFS=$'\t' read -r task_id file; do
  # Skip empty file entries
  [[ -z "$file" ]] && continue
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

if [[ $WARNINGS -gt 0 ]]; then
  echo "No file manifest overlaps detected ($WARNINGS task(s) missing manifests)"
else
  echo "No file manifest overlaps detected"
fi
exit 0
