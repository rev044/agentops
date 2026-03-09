#!/usr/bin/env bash
set -euo pipefail

SCOPE="auto"
EPIC_ID=""
COLLECTION_DETAIL=""

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

json_array_from_stream() {
  if ! sed '/^[[:space:]]*$/d' | sort -u | jq -R . | jq -s .; then
    printf '[]\n'
  fi
}

regex_escape_extended() {
  printf '%s' "$1" | sed -e 's/[][(){}.^$*+?|\\-]/\\&/g'
}

bd_show_json() {
  local issue_id="$1"
  bd show "$issue_id" --json 2>/dev/null | jq -ec 'if type == "array" then .[0] // empty else . end'
}

extract_description_from_show_text() {
  awk '
    /^DESCRIPTION$/ { in_desc = 1; next }
    in_desc {
      if ($0 ~ /^(LABELS|DEPENDENCIES|DEPENDENTS|CHILDREN|COMMENTS|REFERENCES|NOTES):?[[:space:]]*$/) {
        exit
      }
      print
    }
  '
}

collect_children_from_bd_children_json() {
  local json_output
  json_output="$(bd children "$EPIC_ID" --json 2>/dev/null)" || return 1
  printf '%s\n' "$json_output" \
    | jq -er '
        .[]? |
        if type == "object" then
          (.id // .child_id // .issue_id // empty)
        elif type == "string" then
          .
        else
          empty
        end
      ' 2>/dev/null \
    | sed '/^[[:space:]]*$/d' \
    | sort -u
}

collect_children_from_bd_show_json() {
  local json_output
  json_output="$(bd show "$EPIC_ID" --json 2>/dev/null)" || return 1
  printf '%s\n' "$json_output" \
    | jq -er '
        .[]? |
        ((.dependents // .children // [])[]? |
          select((.dependency_type // .type // "parent-child") == "parent-child") |
          (.id // .child_id // .issue_id // empty))
      ' 2>/dev/null \
    | sed '/^[[:space:]]*$/d' \
    | sort -u
}

collect_children_from_human_output() {
  local human_output
  human_output="$(bd show "$EPIC_ID" 2>/dev/null)" || return 1
  printf '%s\n' "$human_output" \
    | awk '
        /^CHILDREN$/ { in_children = 1; next }
        in_children && /^[[:space:]]*$/ { exit }
        in_children { print }
      ' \
    | grep -oE '[[:alnum:]]+-[[:alnum:]]+(\.[0-9]+)+' \
    | sort -u
}

collect_children() {
  local children_output=""

  if children_output="$(collect_children_from_bd_children_json)" && [[ -n "$children_output" ]]; then
    printf '%s\n' "$children_output"
    return 0
  fi

  if children_output="$(collect_children_from_bd_show_json)" && [[ -n "$children_output" ]]; then
    printf '%s\n' "$children_output"
    return 0
  fi

  if children_output="$(collect_children_from_human_output)" && [[ -n "$children_output" ]]; then
    printf '%s\n' "$children_output"
    return 0
  fi

  COLLECTION_DETAIL="no child issues discovered from bd children/show output"
  return 1
}

extract_validation_block_from_text() {
  awk '
    /^```validation[[:space:]]*$/ { in_block = 1; next }
    in_block && /^```[[:space:]]*$/ { exit }
    in_block { print }
  '
}

extract_validation_files_from_block() {
  local validation_block="$1"

  [[ -n "$validation_block" ]] || return 0
  printf '%s\n' "$validation_block" \
    | jq -r '
        def as_items:
          if type == "array" then .[]
          else .
          end;
        (
          (.files // [])[]?,
          (.files_exist // [])[]?,
          ((.content_check // empty) | as_items | .file?),
          ((.content_checks // empty) | as_items | .file?),
          ((.paired_files // empty) | as_items | .file?)
        )
        | select(type == "string" and length > 0)
      ' 2>/dev/null || true
}

extract_files_section_from_text() {
  awk '
    /^Files:[[:space:]]*$/ { in_files = 1; next }
    in_files {
      if ($0 ~ /^[[:space:]]*$/ || $0 ~ /^```/) {
        exit
      }
      if ($0 !~ /^[[:space:]]*-/) {
        exit
      }
      sub(/^[[:space:]]*-[[:space:]]*/, "", $0)
      gsub(/`/, "", $0)
      sub(/[[:space:]]+\(.*\)$/, "", $0)
      if ($0 ~ /^([.[:alnum:]_-]+\/)*[.[:alnum:]_-]+\.(go|py|ts|sh|md|yaml|yml|json)$/) {
        print
      }
    }
  '
}

extract_backticked_files_from_text() {
  grep -oE '`[^`]+\.(go|py|ts|sh|md|yaml|yml|json)`' | tr -d '`' || true
}

extract_scoped_files() {
  local child="$1"
  local description=""
  local child_json=""
  local human_output=""
  local validation_block=""

  if child_json="$(bd_show_json "$child" 2>/dev/null)"; then
    description="$(printf '%s\n' "$child_json" | jq -r '.description // ""')"
  else
    human_output="$(bd show "$child" 2>/dev/null || true)"
    description="$(printf '%s\n' "$human_output" | extract_description_from_show_text)"
  fi

  validation_block="$(printf '%s\n' "$description" | extract_validation_block_from_text)"

  {
    extract_validation_files_from_block "$validation_block"
    printf '%s\n' "$description" | extract_files_section_from_text
    printf '%s\n' "$description" | extract_backticked_files_from_text
  } | sed '/^[[:space:]]*$/d' | sort -u
}

issue_timestamp() {
  local child_json="$1"
  local field="$2"
  printf '%s\n' "$child_json" | jq -r --arg field "$field" '.[$field] // empty'
}

commit_ref_exists() {
  local child="$1"
  local escaped_child
  local pattern

  escaped_child="$(regex_escape_extended "$child")"
  pattern="(^|[^[:alnum:]_.-])${escaped_child}([^[:alnum:]_.-]|$)"
  git log --format='%H' --all --extended-regexp --grep="$pattern" 2>/dev/null | grep -q .
}

commit_matches_json() {
  local since="$1"
  local until="$2"
  shift 2
  local file
  local -a matched_files=()
  local -a git_args=(log --format=%H --all --diff-filter=ACMR)

  [[ -n "$since" ]] && git_args+=("--since=$since")
  [[ -n "$until" ]] && git_args+=("--until=$until")

  for file in "$@"; do
    if git "${git_args[@]}" -- "$file" 2>/dev/null | grep -q .; then
      matched_files+=("$file")
    fi
  done

  if [[ "${#matched_files[@]}" -eq 0 ]]; then
    printf '[]\n'
    return 0
  fi

  printf '%s\n' "${matched_files[@]}" | json_array_from_stream
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
  local child_json=""
  local created_at=""
  local closed_at=""
  local scoped_json commit_json staged_json worktree_json
  local -a scoped_files=()

  if child_json="$(bd_show_json "$child" 2>/dev/null)"; then
    created_at="$(issue_timestamp "$child_json" "created_at")"
    closed_at="$(issue_timestamp "$child_json" "closed_at")"
    if [[ -z "$closed_at" ]]; then
      closed_at="$(issue_timestamp "$child_json" "updated_at")"
    fi
  fi

  mapfile -t scoped_files < <(extract_scoped_files "$child")
  scoped_json="$(printf '%s\n' "${scoped_files[@]}" | json_array_from_stream)"

  case "$SCOPE" in
    auto|commit)
      if commit_ref_exists "$child"; then
        build_child_result "$child" "$scoped_json" "commit" "matched child id in git history" '[]' "pass"
        return 0
      fi
      commit_json="$(commit_matches_json "$created_at" "$closed_at" "${scoped_files[@]}")"
      if echo "$commit_json" | jq -e 'length > 0' >/dev/null 2>&1; then
        build_child_result "$child" "$scoped_json" "commit" "matched scoped files in git history during issue lifetime" "$commit_json" "pass"
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
children_file="$(mktemp)"
trap 'rm -f "$tmp_results" "$children_file"' EXIT

if ! collect_children >"$children_file"; then
  jq -n \
    --arg epic_id "$EPIC_ID" \
    --arg scope "$SCOPE" \
    --arg detail "${COLLECTION_DETAIL:-failed to collect child issues}" \
    '{
      epic_id: $epic_id,
      scope: $scope,
      summary: {
        checked_children: 0,
        passed: 0,
        failed: 1,
        collection_failed: true
      },
      children: [],
      failures: [
        {
          child_id: null,
          detail: $detail
        }
      ]
    }'
  exit 1
fi

children_output="$(cat "$children_file")"
while IFS= read -r child; do
  [[ -n "$child" ]] || continue
  classify_child "$child" >> "$tmp_results"
done <<< "$children_output"

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
