#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
WORKSPACE_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
TARGET_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"

TARGET_ID="fixture-evidence-only-closure"
TARGET_TYPE="task"
PRODUCER="post-mortem"
EVIDENCE_SUMMARY="Evidence-only closure artifact emitted for follow-up validation."
declare -a VALIDATION_COMMANDS=()
declare -a ARTIFACTS=()
declare -a NOTES=()

usage() {
  cat <<'EOF'
Usage: bash skills/post-mortem/scripts/write-evidence-only-closure.sh [options]

Options:
  --repo-root <path>            Target repo root. Defaults to $REPO_ROOT or current git root.
  --target-id <id>              Target issue/epic/policy identifier. Default: fixture-evidence-only-closure.
  --target-type <type>          Target type (for example: issue, epic, policy). Default: task.
  --producer <name>             Producer name recorded in the artifact. Default: post-mortem.
  --validation-command <cmd>    Validation command to record. Repeatable. Defaults to manifest validation for the target root.
  --evidence-summary <text>     Human summary for the closure evidence.
  --artifact <path>             Repo-relative or absolute evidence artifact path. Repeatable.
  --note <text>                 Additional note to include in the evidence block. Repeatable.
  --help                        Show this help.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo-root)
      TARGET_ROOT="${2:-}"
      shift 2
      ;;
    --target-id)
      TARGET_ID="${2:-}"
      shift 2
      ;;
    --target-type)
      TARGET_TYPE="${2:-}"
      shift 2
      ;;
    --producer)
      PRODUCER="${2:-}"
      shift 2
      ;;
    --validation-command)
      VALIDATION_COMMANDS+=("${2:-}")
      shift 2
      ;;
    --evidence-summary)
      EVIDENCE_SUMMARY="${2:-}"
      shift 2
      ;;
    --artifact)
      ARTIFACTS+=("${2:-}")
      shift 2
      ;;
    --note)
      NOTES+=("${2:-}")
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "Unknown arg: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ "$TARGET_ROOT" != /* ]]; then
  TARGET_ROOT="$(cd "$TARGET_ROOT" && pwd)"
fi

[[ -n "$TARGET_ID" ]] || { echo "--target-id is required" >&2; exit 1; }
[[ -n "$TARGET_TYPE" ]] || { echo "--target-type is required" >&2; exit 1; }
[[ -n "$EVIDENCE_SUMMARY" ]] || { echo "--evidence-summary is required" >&2; exit 1; }

mkdir -p "$TARGET_ROOT/schemas"
schema_source="$WORKSPACE_ROOT/schemas/evidence-only-closure.v1.schema.json"
schema_target="$TARGET_ROOT/schemas/evidence-only-closure.v1.schema.json"
if [[ "$(cd "$(dirname "$schema_source")" && pwd)/$(basename "$schema_source")" != "$(cd "$(dirname "$schema_target")" && pwd)/$(basename "$schema_target")" ]]; then
  cp "$schema_source" "$schema_target"
fi

ROOT="$TARGET_ROOT"
source "$WORKSPACE_ROOT/lib/hook-helpers.sh"

if [[ "${#VALIDATION_COMMANDS[@]}" -eq 0 ]]; then
  VALIDATION_COMMANDS=("bash scripts/validate-manifests.sh --repo-root $TARGET_ROOT")
fi

validation_commands_json="$(
  printf '%s\n' "${VALIDATION_COMMANDS[@]}" \
    | jq -R . \
    | jq -s .
)"

artifacts_json="$(
  if [[ "${#ARTIFACTS[@]}" -eq 0 ]]; then
    printf '[]\n'
  else
    printf '%s\n' "${ARTIFACTS[@]}" | jq -R . | jq -s .
  fi
)"

notes_json="$(
  if [[ "${#NOTES[@]}" -eq 0 ]]; then
    printf '[]\n'
  else
    printf '%s\n' "${NOTES[@]}" | jq -R . | jq -s .
  fi
)"

git_branch="$(git -C "$TARGET_ROOT" branch --show-current 2>/dev/null || true)"
head_sha="$(git -C "$TARGET_ROOT" rev-parse HEAD 2>/dev/null || true)"
mapfile -t modified_files < <(git -C "$TARGET_ROOT" status --short 2>/dev/null | awk '{print $2}')
if [[ "${#modified_files[@]}" -eq 0 ]]; then
  modified_files_json='[]'
  git_dirty='false'
else
  modified_files_json="$(printf '%s\n' "${modified_files[@]}" | jq -R . | jq -s .)"
  git_dirty='true'
fi

repo_state_json="$(
  jq -n \
    --arg repo_root "$TARGET_ROOT" \
    --arg git_branch "$git_branch" \
    --arg head_sha "$head_sha" \
    --argjson git_dirty "$git_dirty" \
    --argjson modified_files "$modified_files_json" \
    '{
      repo_root: $repo_root,
      git_branch: $git_branch,
      git_dirty: $git_dirty,
      head_sha: $head_sha,
      modified_files: $modified_files
    }'
)"

evidence_json="$(
  jq -n \
    --arg summary "$EVIDENCE_SUMMARY" \
    --argjson artifacts "$artifacts_json" \
    --argjson notes "$notes_json" \
    '{
      summary: $summary,
      artifacts: $artifacts,
      notes: $notes
    }'
)"

write_evidence_only_closure_packet \
  "$TARGET_ID" \
  "$TARGET_TYPE" \
  "$PRODUCER" \
  "$validation_commands_json" \
  "$repo_state_json" \
  "$evidence_json"
