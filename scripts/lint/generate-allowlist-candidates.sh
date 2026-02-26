#!/usr/bin/env bash
# Generate candidate allowlist entries from converted codex skills.
# Run AFTER conversion, BEFORE validation.
# Usage: generate-allowlist-candidates.sh <converted-skills-dir>
set -euo pipefail

CONVERTED_DIR="${1:?Usage: generate-allowlist-candidates.sh <converted-skills-dir>}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ALLOWLIST="${SCRIPT_DIR}/codex-residual-allowlist.txt"

# Load existing allowlist patterns
mapfile -t patterns < <(grep -v '^#' "$ALLOWLIST" | grep -v '^$')

# Find unallowlisted markers
candidates=()
while IFS= read -r file; do
  while IFS= read -r line; do
    allowed=false
    for pat in "${patterns[@]}"; do
      if echo "$line" | grep -qE "$pat" 2>/dev/null; then
        allowed=true; break
      fi
    done
    if ! $allowed; then
      skill_name="$(echo "$file" | sed 's|.*/\([^/]*\)/SKILL.md|\1|')"
      candidates+=("# $skill_name: $line")
    fi
  done < <(grep -inE '\bclaude\b' "$file" 2>/dev/null || true)
done < <(find "$CONVERTED_DIR" -name "SKILL.md" -type f)

if [[ ${#candidates[@]} -eq 0 ]]; then
  echo "No unallowlisted residual markers found."
  exit 0
fi

echo "Found ${#candidates[@]} unallowlisted residual markers:"
printf '%s\n' "${candidates[@]}"
echo ""
echo "Add patterns to $ALLOWLIST to allow these, or fix the converter rules."
exit 1
