#!/usr/bin/env bash
# pending-cleaner.sh — Archive stale pending queue entries (>2 days old)
# Called during session start to prevent queue buildup.
# Exit 0 always — never block session start.

ROOT=$(git rev-parse --show-toplevel 2>/dev/null || exit 0)

PENDING_DIR="$ROOT/.agents/ao/pending"
ARCHIVE_DIR="$ROOT/.agents/ao/archive"
LOG_FILE="$ROOT/.agents/ao/hook-errors.log"

# Nothing to do if pending dir doesn't exist
if [ ! -d "$PENDING_DIR" ]; then
    exit 0
fi

# Find stale files (older than 2 days)
stale_files=$(find "$PENDING_DIR" -name "*.jsonl" -mtime +2 2>/dev/null)

if [ -z "$stale_files" ]; then
    exit 0
fi

# Ensure archive dir exists
mkdir -p "$ARCHIVE_DIR" 2>/dev/null || exit 0

timestamp=$(date +%Y%m%d-%H%M%S)

echo "$stale_files" | while IFS= read -r file; do
    [ -f "$file" ] || continue
    basename=$(basename "$file")
    archive_name="${timestamp}-${basename}"

    # Atomic: write archive first, then remove original
    if cp "$file" "$ARCHIVE_DIR/$archive_name" 2>/dev/null; then
        rm -f "$file" 2>/dev/null
        echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) archived stale: $basename -> $archive_name" >> "$LOG_FILE" 2>/dev/null
    else
        echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) ERROR: failed to archive $basename" >> "$LOG_FILE" 2>/dev/null
    fi
done

exit 0
