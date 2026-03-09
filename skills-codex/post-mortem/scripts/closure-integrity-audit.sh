#!/usr/bin/env bash
set -euo pipefail

SCOPE="auto"
EPIC_ID=""

usage() {
  cat <<'EOF'
Usage: bash skills/post-mortem/scripts/closure-integrity-audit.sh [--scope auto|commit|staged|worktree] <epic-id>
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --scope)
      SCOPE="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      if [[ -z "$EPIC_ID" ]]; then
        EPIC_ID="$1"
        shift
      else
        echo "Unknown arg: $1" >&2
        usage >&2
        exit 2
      fi
      ;;
  esac
done

case "$SCOPE" in
  auto|commit|staged|worktree) ;;
  *)
    echo "Invalid --scope: $SCOPE" >&2
    usage >&2
    exit 2
    ;;
esac

[[ -n "$EPIC_ID" ]] || {
  echo "epic id is required" >&2
  usage >&2
  exit 2
}

command -v jq >/dev/null 2>&1 || {
  echo "jq is required" >&2
  exit 1
}

command -v bd >/dev/null 2>&1 || {
  echo "bd is required" >&2
  exit 1
}

collect_children() {
  bd children "$EPIC_ID" 2>/dev/null | grep -oE '[a-z]{2}-[a-z0-9]+\.[0-9]+' | sort -u
}

extract_scoped_files() {
  local child="$1"
  bd show "$child" 2>/dev/null \
    | grep -oE '`[^`]+\.(go|py|ts|sh|md|yaml|yml|json)`' \
    | tr -d '`' \
    | sort -u
}

json_array_from_stream() {
  if ! cat | sed '/^[[:space:]]*$/d' | sort -u | jq -R . | jq -s .; then
    printf '[]\n'
  fi
}

commit_ref_exists() {
  local child="$1"
  git log --oneline --all --grep="$child" 2>/dev/null | grep -q .
}

commit_file_exists() {
  local file
  for file in "$@"; do
    if git log --oneline --diff-filter=ACMR -- "$file" 2>/dev/null | head -1 | grep -q .; then
      return 0
    fi
  done
  return 1
}

staged_matches_json() {
  if [[ "$#" -eq 0 ]]; then
    printf '[]\n'
    return 0
  fi
  git diff --cached --name-only --diff-filter=ACMR -- "$@" 2>/dev/null | json_array_from_stream
}

worktree_matches_json() {
  if [[ "$#" -eq 0 ]]; then
    printf '[]\n'
    return 0
  fi

  {
    git diff --name-only --diff-filter=ACMR -- "$@" 2>/dev/null || true
    git ls-files --others --exclude-standard -- "$@" 2>/dev/null || true
  } | json_array_from_stream
}

build_child_result() {
  local child="$1"
  local scoped_json="$2"
  local mode="$3"
  local detail="$4"
  local matches_json="$5"
  local status="$6"

  jq -n \
    --arg child_id "$child" \
    --arg status "$status" \
    --arg evidence_mode "$mode" \
    --arg detail "$detail" \
    --argjson scoped_files "$scoped_json" \
    --argjson matched_files "$matches_json" \
    '{
      child_id: $child_id,
      status: $status,
      evidence_mode: $evidence_mode,
      detail: $detail,
      scoped_files: $scoped_files,
      matched_files: $matched_files
    }'
}

classify_child() {
  local child="$1"
  local scoped_json staged_json worktree_json
  local -a scoped_files=()

  mapfile -t scoped_files < <(extract_scoped_files "$child")
  scoped_json="$(printf '%s\n' "${scoped_files[@]}" | json_array_from_stream)"

  case "$SCOPE" in
    auto|commit)
      if commit_ref_exists "$child"; then
        build_child_result "$child" "$scoped_json" "commit" "matched child id in git history" '[]' "pass"
        return 0
      fi
      if [[ "${#scoped_files[@]}" -gt 0 ]] && commit_file_exists "${scoped_files[@]}"; then
        build_child_result "$child" "$scoped_json" "commit" "matched scoped files in git history" "$scoped_json" "pass"
        return 0
      fi
      if [[ "$SCOPE" == "commit" ]]; then
        build_child_result "$child" "$scoped_json" "none" "no qualifying commit evidence" '[]' "fail"
        return 0
      fi
      ;;
  esac

  if [[ "${#scoped_files[@]}" -eq 0 ]]; then
    build_child_result "$child" "$scoped_json" "none" "no scoped files and no qualifying commit evidence" '[]' "fail"
    return 0
  fi

  case "$SCOPE" in
    auto|staged)
      staged_json="$(staged_matches_json "${scoped_files[@]}")"
      if echo "$staged_json" | jq -e 'length > 0' >/dev/null 2>&1; then
        build_child_result "$child" "$scoped_json" "staged" "matched scoped files in git index" "$staged_json" "pass"
        return 0
      fi
      if [[ "$SCOPE" == "staged" ]]; then
        build_child_result "$child" "$scoped_json" "none" "no qualifying staged evidence" '[]' "fail"
        return 0
      fi
      ;;
  esac

  case "$SCOPE" in
    auto|worktree)
      worktree_json="$(worktree_matches_json "${scoped_files[@]}")"
      if echo "$worktree_json" | jq -e 'length > 0' >/dev/null 2>&1; then
        build_child_result "$child" "$scoped_json" "worktree" "matched scoped files in working tree" "$worktree_json" "pass"
        return 0
      fi
      ;;
  esac

  build_child_result "$child" "$scoped_json" "none" "no qualifying commit, staged, or worktree evidence" '[]' "fail"
}

tmp_results="$(mktemp)"
trap 'rm -f "$tmp_results"' EXIT

while IFS= read -r child; do
  [[ -n "$child" ]] || continue
  classify_child "$child" >> "$tmp_results"
done < <(collect_children)

if [[ ! -s "$tmp_results" ]]; then
  jq -n \
    --arg epic_id "$EPIC_ID" \
    --arg scope "$SCOPE" \
    '{
      epic_id: $epic_id,
      scope: $scope,
      summary: {
        checked_children: 0,
        passed: 0,
        failed: 0
      },
      children: [],
      failures: []
    }'
  exit 0
fi

jq -s \
  --arg epic_id "$EPIC_ID" \
  --arg scope "$SCOPE" \
  '{
    epic_id: $epic_id,
    scope: $scope,
    summary: {
      checked_children: length,
      passed: ([.[] | select(.status == "pass")] | length),
      failed: ([.[] | select(.status == "fail")] | length),
      evidence_modes: {
        commit: ([.[] | select(.status == "pass" and .evidence_mode == "commit") | .child_id] | sort),
        staged: ([.[] | select(.status == "pass" and .evidence_mode == "staged") | .child_id] | sort),
        worktree: ([.[] | select(.status == "pass" and .evidence_mode == "worktree") | .child_id] | sort)
      }
    },
    children: .,
    failures: [.[] | select(.status == "fail") | {child_id, detail}]
  }' "$tmp_results"
