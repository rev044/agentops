#!/usr/bin/env bash
# bootstrap-maturity.sh — One-time migration: add maturity: provisional to
# existing .md learnings that lack a maturity field.
#
# Usage: bash scripts/bootstrap-maturity.sh [learnings-dir]
# Default learnings dir: .agents/learnings
#
# Safe to run multiple times — skips files that already have maturity field.

set -euo pipefail

LEARNINGS_DIR="${1:-.agents/learnings}"

if [[ ! -d "$LEARNINGS_DIR" ]]; then
    echo "Directory not found: $LEARNINGS_DIR"
    exit 1
fi

total=0
updated=0
skipped=0

for file in "$LEARNINGS_DIR"/*.md; do
    [[ -f "$file" ]] || continue
    total=$((total + 1))

    # Check if file has YAML frontmatter
    head_line=$(head -1 "$file")
    if [[ "$head_line" != "---" ]]; then
        skipped=$((skipped + 1))
        continue
    fi

    # Check if maturity field already exists in frontmatter
    if grep -q "^maturity:" "$file"; then
        skipped=$((skipped + 1))
        continue
    fi

    # Add maturity: provisional after the opening ---
    # Use a temp file to avoid in-place edit portability issues
    tmpfile=$(mktemp)
    {
        echo "---"
        echo "maturity: provisional"
        tail -n +2 "$file"
    } > "$tmpfile"
    mv "$tmpfile" "$file"
    updated=$((updated + 1))
done

echo "Bootstrap maturity complete:"
echo "  Total .md files: $total"
echo "  Updated (added maturity: provisional): $updated"
echo "  Skipped (already has maturity or no frontmatter): $skipped"
