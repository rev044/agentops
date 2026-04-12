#!/usr/bin/env bash
# Plan Metadata Parser - Extract YAML frontmatter from plans
#
# Usage:
#   plan-metadata-parser.sh <plan_file> <field>
#
# Examples:
#   plan-metadata-parser.sh PLAN_FOO.md status
#   plan-metadata-parser.sh PLAN_FOO.md completion_criteria
#
# Returns:
#   - Field value (single line)
#   - Multi-line values for arrays (one per line)
#   - Exit 0 on success, 1 on error

set -euo pipefail

PLAN_FILE="${1:-}"
FIELD="${2:-}"

if [ -z "$PLAN_FILE" ] || [ -z "$FIELD" ]; then
  echo "Usage: $0 <plan_file> <field>" >&2
  exit 1
fi

if [ ! -f "$PLAN_FILE" ]; then
  echo "Error: Plan file not found: $PLAN_FILE" >&2
  exit 1
fi

# Extract YAML frontmatter (between --- delimiters)
# Then extract the requested field

# Check if file has YAML frontmatter
if ! head -1 "$PLAN_FILE" | grep -q "^---$"; then
  echo "Error: Plan file has no YAML frontmatter: $PLAN_FILE" >&2
  exit 1
fi

# Extract frontmatter (lines between first --- and second ---)
FRONTMATTER=$(awk '/^---$/{flag=!flag; next} flag' "$PLAN_FILE" | head -n 100)

# Parse field (handle both single-line and multi-line values)
case "$FIELD" in
  completion_criteria)
    # Extract completion criteria list (lines starting with "  - [")
    echo "$FRONTMATTER" | sed -n '/^completion_criteria:/,/^[a-z_]/ {
      /^  - \[/p
    }'
    ;;
  related_plans)
    # Extract related plans list
    echo "$FRONTMATTER" | sed -n '/^related_plans:/,/^[a-z_]/ {
      /^  - /p
    }' | sed 's/^  - //'
    ;;
  *)
    # Extract single-line field
    echo "$FRONTMATTER" | grep "^${FIELD}:" | sed "s/^${FIELD}: *//" || echo ""
    ;;
esac
