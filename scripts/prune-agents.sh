#!/usr/bin/env bash
# prune-agents.sh — Enforce .agents/ retention policies
#
# Usage:
#   ./scripts/prune-agents.sh              # Dry run (default) — show what would be deleted
#   ./scripts/prune-agents.sh --execute    # Actually delete files
#
# Policies defined in .agents/README.md ## Pruning section.
# Never touches: learnings/, patterns/, plans/, research/, retros/ (knowledge assets)

set -euo pipefail

AGENTS_DIR=".agents"
DRY_RUN=true
TOTAL_FILES=0
TOTAL_BYTES=0

if [[ "${1:-}" == "--execute" ]]; then
    DRY_RUN=false
    echo "=== EXECUTE MODE — files will be deleted ==="
else
    echo "=== DRY RUN — no files will be deleted (pass --execute to delete) ==="
fi
echo ""

# Helper: list files to prune, sorted oldest first
prune_keep_newest() {
    local dir="$1"
    local keep="$2"
    local label="$3"

    if [[ ! -d "$dir" ]]; then
        return
    fi

    local count
    count=$(find "$dir" -maxdepth 1 -type f 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$count" -le "$keep" ]]; then
        echo "[$label] $count files — within limit ($keep). Nothing to prune."
        return
    fi

    local to_delete=$((count - keep))
    echo "[$label] $count files — keeping newest $keep, pruning $to_delete"

    # List oldest files first (by modification time)
    find "$dir" -maxdepth 1 -type f -print0 2>/dev/null \
        | xargs -0 ls -t 2>/dev/null \
        | tail -n "$to_delete" \
        | while read -r f; do
            local size
            size=$(stat -f%z "$f" 2>/dev/null || stat --format=%s "$f" 2>/dev/null || echo 0)
            TOTAL_BYTES=$((TOTAL_BYTES + size))
            TOTAL_FILES=$((TOTAL_FILES + 1))
            if [[ "$DRY_RUN" == true ]]; then
                echo "  would delete: $f ($(numfmt_size "$size"))"
            else
                rm -f "$f"
                echo "  deleted: $f ($(numfmt_size "$size"))"
            fi
        done
}

prune_older_than() {
    local dir="$1"
    local days="$2"
    local pattern="$3"
    local label="$4"

    if [[ ! -d "$dir" ]]; then
        return
    fi

    local found
    found=$(find "$dir" -maxdepth 1 -name "$pattern" -type f -mtime +"$days" 2>/dev/null | wc -l | tr -d ' ')

    if [[ "$found" -eq 0 ]]; then
        echo "[$label] No files older than ${days}d matching '$pattern'. Nothing to prune."
        return
    fi

    echo "[$label] $found files older than ${days}d"

    find "$dir" -maxdepth 1 -name "$pattern" -type f -mtime +"$days" -print0 2>/dev/null \
        | while IFS= read -r -d '' f; do
            local size
            size=$(stat -f%z "$f" 2>/dev/null || stat --format=%s "$f" 2>/dev/null || echo 0)
            TOTAL_BYTES=$((TOTAL_BYTES + size))
            TOTAL_FILES=$((TOTAL_FILES + 1))
            if [[ "$DRY_RUN" == true ]]; then
                echo "  would delete: $f ($(numfmt_size "$size"))"
            else
                rm -f "$f"
                echo "  deleted: $f ($(numfmt_size "$size"))"
            fi
        done
}

numfmt_size() {
    local bytes="$1"
    if [[ "$bytes" -ge 1073741824 ]]; then
        echo "$(( bytes / 1073741824 ))GB"
    elif [[ "$bytes" -ge 1048576 ]]; then
        echo "$(( bytes / 1048576 ))MB"
    elif [[ "$bytes" -ge 1024 ]]; then
        echo "$(( bytes / 1024 ))KB"
    else
        echo "${bytes}B"
    fi
}

# --- Policy: council/ — keep last 30 ---
prune_keep_newest "$AGENTS_DIR/council" 30 "council"
echo ""

# --- Policy: tooling/ — keep last run only (newest date prefix) ---
if [[ -d "$AGENTS_DIR/tooling" ]]; then
    tooling_count=$(find "$AGENTS_DIR/tooling" -maxdepth 1 -type f 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$tooling_count" -gt 0 ]]; then
        # All tooling files from the same run — keep newest by mtime, prune rest
        # Since tooling has no date-prefix convention, keep files from last 1 day
        old_tooling=$(find "$AGENTS_DIR/tooling" -maxdepth 1 -type f -mtime +1 2>/dev/null | wc -l | tr -d ' ')
        if [[ "$old_tooling" -gt 0 ]]; then
            echo "[tooling] $tooling_count total files — $old_tooling older than 1 day"
            find "$AGENTS_DIR/tooling" -maxdepth 1 -type f -mtime +1 -print0 2>/dev/null \
                | while IFS= read -r -d '' f; do
                    size=$(stat -f%z "$f" 2>/dev/null || stat --format=%s "$f" 2>/dev/null || echo 0)
                    if [[ "$DRY_RUN" == true ]]; then
                        echo "  would delete: $f"
                    else
                        rm -f "$f"
                        echo "  deleted: $f"
                    fi
                done
        else
            echo "[tooling] $tooling_count files — all from recent run. Nothing to prune."
        fi
    fi
fi
echo ""

# --- Policy: knowledge/pending/ — older than 14 days ---
prune_older_than "$AGENTS_DIR/knowledge/pending" 14 "*.md" "knowledge/pending"
echo ""

# --- Policy: rpi/ phase summaries — older than 30 days ---
prune_older_than "$AGENTS_DIR/rpi" 30 "phase-*-summary-*" "rpi/phase-summaries"
echo ""

# --- Policy: ao/sessions/ — keep last 50 ---
prune_keep_newest "$AGENTS_DIR/ao/sessions" 50 "ao/sessions"
echo ""

# --- Policy: handoff/ — keep last 10 ---
prune_keep_newest "$AGENTS_DIR/handoff" 10 "handoff"
echo ""

# --- Summary ---
echo "========================================"
if [[ "$DRY_RUN" == true ]]; then
    echo "DRY RUN COMPLETE"
    echo "Files that would be deleted: $TOTAL_FILES"
else
    echo "PRUNE COMPLETE"
    echo "Files deleted: $TOTAL_FILES"
fi
echo ""
echo "Protected directories (never pruned):"
echo "  learnings/ patterns/ plans/ research/ retros/"
echo ""
echo "Recommendation: Add .agents/tooling/ to .gitignore (1.1GB of regenerable scanner output)"
