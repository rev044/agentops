#!/usr/bin/env bash
# merge-worktrees.sh — merge completed worktree changes into the main repo.
# Usage: merge-worktrees.sh <worktree-dir> [<worktree-dir> ...]
#
# For each worktree directory:
#   1. Verify it is a valid git worktree
#   2. Get list of changed files (vs the worktree's merge-base with HEAD)
#   3. Copy files using /bin/cp -f (bypasses interactive aliases)
#   4. Report what was copied
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

if [[ $# -eq 0 ]]; then
    echo "Usage: $(basename "$0") <worktree-dir> [<worktree-dir> ...]"
    exit 1
fi

total_copied=0
total_errors=0

for wt_dir in "$@"; do
    echo ""
    echo "==> Processing worktree: $wt_dir"

    # --- Validate worktree ---
    if [[ ! -d "$wt_dir" ]]; then
        echo "  SKIP: directory does not exist"
        total_errors=$((total_errors + 1))
        continue
    fi

    if [[ ! -f "$wt_dir/.git" ]] && [[ ! -d "$wt_dir/.git" ]]; then
        echo "  SKIP: not a git worktree (no .git)"
        total_errors=$((total_errors + 1))
        continue
    fi

    # --- Find changed files ---
    merge_base="$(cd "$wt_dir" && git merge-base HEAD main 2>/dev/null || echo "")"
    if [[ -z "$merge_base" ]]; then
        echo "  SKIP: cannot determine merge-base with main"
        total_errors=$((total_errors + 1))
        continue
    fi

    changed_files="$(cd "$wt_dir" && git diff --name-only "$merge_base" HEAD)"
    if [[ -z "$changed_files" ]]; then
        echo "  SKIP: no changed files vs main"
        continue
    fi

    # --- Copy files ---
    copied=0
    while IFS= read -r file; do
        src="$wt_dir/$file"
        dst="$REPO_ROOT/$file"

        if [[ ! -f "$src" ]]; then
            echo "  SKIP (deleted): $file"
            continue
        fi

        # Ensure destination directory exists
        dst_dir="$(dirname "$dst")"
        [[ -d "$dst_dir" ]] || mkdir -p "$dst_dir"

        /bin/cp -f "$src" "$dst"
        echo "  COPY: $file"
        copied=$((copied + 1))
    done <<< "$changed_files"

    echo "  -- $copied file(s) copied from $wt_dir"
    total_copied=$((total_copied + copied))
done

echo ""
echo "==> Summary: $total_copied file(s) copied, $total_errors worktree error(s)"
[[ "$total_errors" -eq 0 ]] || exit 1
