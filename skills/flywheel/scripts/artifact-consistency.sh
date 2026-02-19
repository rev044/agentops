#!/usr/bin/env bash
set -euo pipefail

VERBOSE=0
if [[ "${1:-}" == "--verbose" ]]; then
  VERBOSE=1
  shift
fi

ROOT="${1:-.agents}"

if [[ ! -d "$ROOT" ]]; then
  echo "TOTAL_REFS=0"
  echo "BROKEN_REFS=0"
  echo "CONSISTENCY=100"
  echo "STATUS=Healthy"
  exit 0
fi

total_refs=0
broken_refs=0
broken_lines=()

# Scan markdown while excluding ao telemetry/session data.
while IFS= read -r -d '' file; do
  # Strip fenced code blocks to avoid counting template snippets as broken links.
  refs=$(awk '
    BEGIN { in_code=0 }
    /^```/ { in_code=!in_code; next }
    in_code { next }
    {
      line=$0
      while (match(line, /\.agents\/[A-Za-z0-9._\/-]+\.(md|json|jsonl)/)) {
        print substr(line, RSTART, RLENGTH)
        line = substr(line, RSTART + RLENGTH)
      }
    }
  ' "$file" | sort -u)

  while IFS= read -r ref; do
    [[ -z "$ref" ]] && continue

    # Skip template placeholders and non-literal paths.
    if [[ "$ref" =~ YYYY|\<|\>|\{|\}|\* ]]; then
      continue
    fi

    total_refs=$((total_refs + 1))

    # Normalize leading ./, then check relative to repo root.
    normalized="${ref#./}"
    if [[ ! -f "$normalized" ]]; then
      broken_refs=$((broken_refs + 1))
      if (( VERBOSE )); then
        broken_lines+=("$file -> $normalized")
      fi
    fi
  done <<< "$refs"
done < <(find "$ROOT" -type f -name "*.md" -not -path "$ROOT/ao/*" -print0)

if (( total_refs > 0 )); then
  consistency=$(( (total_refs - broken_refs) * 100 / total_refs ))
else
  consistency=100
fi

status="Critical"
if (( consistency > 90 )); then
  status="Healthy"
elif (( consistency >= 70 )); then
  status="Warning"
fi

echo "TOTAL_REFS=$total_refs"
echo "BROKEN_REFS=$broken_refs"
echo "CONSISTENCY=$consistency"
echo "STATUS=$status"

if (( VERBOSE )) && (( broken_refs > 0 )); then
  for line in "${broken_lines[@]}"; do
    echo "BROKEN_REF=$line"
  done
fi
