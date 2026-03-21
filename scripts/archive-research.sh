#!/usr/bin/env bash
# Archive stale research files to reduce search noise.
# Moves .agents/research/ files older than N days to .agents/archive/research/.
# Usage: scripts/archive-research.sh [--days N] [--dry-run]
set -euo pipefail

DAYS=30
DRY_RUN=false
RESEARCH_DIR=".agents/research"
ARCHIVE_DIR=".agents/archive/research"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --days) DAYS="$2"; shift 2 ;;
        --dry-run) DRY_RUN=true; shift ;;
        -h|--help)
            echo "Usage: $0 [--days N] [--dry-run]"
            echo "  --days N    Archive files older than N days (default: 30)"
            echo "  --dry-run   Show what would be moved without moving"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

if [[ ! -d "$RESEARCH_DIR" ]]; then
    echo "No research directory found at $RESEARCH_DIR"
    exit 0
fi

# Find files older than N days
STALE_FILES=()
while IFS= read -r -d '' file; do
    STALE_FILES+=("$file")
done < <(find "$RESEARCH_DIR" -maxdepth 1 -name "*.md" -mtime +"$DAYS" -print0 2>/dev/null)

if [[ ${#STALE_FILES[@]} -eq 0 ]]; then
    echo "No research files older than $DAYS days."
    exit 0
fi

echo "Found ${#STALE_FILES[@]} research files older than $DAYS days."

if [[ "$DRY_RUN" == "true" ]]; then
    echo "Dry run — would move:"
    for f in "${STALE_FILES[@]}"; do
        echo "  $(basename "$f")"
    done
    exit 0
fi

mkdir -p "$ARCHIVE_DIR"

MOVED=0
for f in "${STALE_FILES[@]}"; do
    mv "$f" "$ARCHIVE_DIR/"
    MOVED=$((MOVED + 1))
done

echo "Archived $MOVED files to $ARCHIVE_DIR/"
