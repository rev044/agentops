#!/usr/bin/env bash
# Archive stale research files to reduce search noise.
# Moves .agents/research/ files older than N days to .agents/archive/research/.
# Uses date from filename (YYYY-MM-DD prefix) instead of mtime, since ao mine
# and other tools touch files and reset mtime.
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

# Compute cutoff date (macOS and GNU date compatible)
CUTOFF=$(date -v-"${DAYS}"d +%Y-%m-%d 2>/dev/null || date -d "-${DAYS} days" +%Y-%m-%d 2>/dev/null)
if [[ -z "$CUTOFF" ]]; then
    echo "Error: could not compute cutoff date"
    exit 1
fi

# Find files with YYYY-MM-DD prefix older than cutoff
STALE_FILES=()
for f in "$RESEARCH_DIR"/*.md; do
    [[ -f "$f" ]] || continue
    # Extract date from filename (e.g., 2026-02-22-cmd-ao-complexity-scout.md)
    FILE_DATE=$(basename "$f" | grep -oE '^[0-9]{4}-[0-9]{2}-[0-9]{2}' || echo "")
    if [[ -z "$FILE_DATE" ]]; then
        # No date prefix — skip (don't archive undated files)
        continue
    fi
    if [[ "$FILE_DATE" < "$CUTOFF" ]]; then
        STALE_FILES+=("$f")
    fi
done

if [[ ${#STALE_FILES[@]} -eq 0 ]]; then
    echo "No research files older than $DAYS days (cutoff: $CUTOFF)."
    exit 0
fi

echo "Found ${#STALE_FILES[@]} research files older than $DAYS days (cutoff: $CUTOFF)."

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
