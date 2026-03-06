#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
current_branch="$(git branch --show-current)"
common_dir="$(git rev-parse --git-common-dir)"
canonical_root="${common_dir%/.git}"

if [[ -z "$current_branch" ]]; then
    echo "FAIL: current worktree is detached; run this gate from a branch-attached task worktree" >&2
    exit 1
fi

if [[ ! -d "$canonical_root" ]]; then
    echo "FAIL: canonical root not found: $canonical_root" >&2
    exit 1
fi

canonical_branch="$(git -C "$canonical_root" branch --show-current)"
if [[ -z "$canonical_branch" ]]; then
    echo "FAIL: canonical root $canonical_root is detached; it must stay on main" >&2
    exit 1
fi

if [[ "$canonical_branch" != "main" ]]; then
    echo "FAIL: canonical root $canonical_root is on $canonical_branch; expected main" >&2
    exit 1
fi

if [[ -n "$(git -C "$canonical_root" status --porcelain)" ]]; then
    echo "FAIL: canonical root $canonical_root has uncommitted changes" >&2
    exit 1
fi

declare -A allowed_worktrees=()
allowed_worktrees["main"]="$canonical_root"
allowed_worktrees["$current_branch"]="$repo_root"

if [[ -n "${WORKTREE_DISPOSITION_EXTRA_ALLOWED_BRANCHES:-}" ]]; then
    IFS=',' read -r -a extra_branches <<<"$WORKTREE_DISPOSITION_EXTRA_ALLOWED_BRANCHES"
    for branch in "${extra_branches[@]}"; do
        branch="${branch//[[:space:]]/}"
        [[ -n "$branch" ]] || continue

        worktree_path="$(git for-each-ref --format='%(worktreepath)' "refs/heads/$branch")"
        if [[ -n "$worktree_path" ]]; then
            allowed_worktrees["$branch"]="$worktree_path"
        fi
    done
fi

failures=()
while IFS='|' read -r branch worktree_path; do
    [[ -n "$worktree_path" ]] || continue

    expected_path="${allowed_worktrees[$branch]:-}"
    if [[ -z "$expected_path" ]]; then
        failures+=("$branch attached at $worktree_path")
        continue
    fi

    if [[ "$worktree_path" != "$expected_path" ]]; then
        failures+=("$branch attached at $worktree_path (expected $expected_path)")
    fi
done < <(git for-each-ref --format='%(refname:short)|%(worktreepath)' refs/heads)

if (( ${#failures[@]} > 0 )); then
    printf 'FAIL: unexpected branch-attached worktrees detected:\n' >&2
    printf '  - %s\n' "${failures[@]}" >&2
    exit 1
fi

echo "PASS: canonical root $canonical_root is clean on main; current branch $current_branch is attached at $repo_root"
