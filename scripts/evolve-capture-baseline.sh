#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: scripts/evolve-capture-baseline.sh --label <slug> [options]

Capture an immutable evolve baseline snapshot for the current goal era.

Required:
  --label <slug>              Baseline label (letters, numbers, ., _, -)

Options:
  --repo-root <path>          Repo root (default: current directory)
  --output-dir <path>         Baseline directory (default: .agents/evolve/baselines)
  --legacy-path <path>        Compatibility mirror path (default: .agents/evolve/fitness-0-baseline.json)
  --active-path <path>        Active baseline pointer (default: .agents/evolve/active-baseline.txt)
  --index-path <path>         Baseline index path (default: .agents/evolve/baselines/index.jsonl)
  --timeout <seconds>         ao goals measure timeout (default: 60)
  --force                     Overwrite an existing label
  -h, --help                  Show help
EOF
}

die() {
  echo "ERROR: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

iso_timestamp() {
  date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -Iseconds
}

resolve_path() {
  local base="$1"
  local path="$2"
  if [[ "$path" = /* ]]; then
    printf '%s\n' "$path"
  else
    printf '%s/%s\n' "$base" "$path"
  fi
}

relative_to_repo() {
  local repo_root="$1"
  local path="$2"
  case "$path" in
    "$repo_root"/*)
      printf '%s\n' "${path#"$repo_root"/}"
      ;;
    *)
      printf '%s\n' "$path"
      ;;
  esac
}

LABEL=""
REPO_ROOT="$(pwd)"
OUTPUT_DIR=".agents/evolve/baselines"
LEGACY_PATH=".agents/evolve/fitness-0-baseline.json"
ACTIVE_PATH=".agents/evolve/active-baseline.txt"
INDEX_PATH=".agents/evolve/baselines/index.jsonl"
TIMEOUT="60"
FORCE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --label)
      LABEL="${2:-}"
      shift 2
      ;;
    --repo-root)
      REPO_ROOT="${2:-}"
      shift 2
      ;;
    --output-dir)
      OUTPUT_DIR="${2:-}"
      shift 2
      ;;
    --legacy-path)
      LEGACY_PATH="${2:-}"
      shift 2
      ;;
    --active-path)
      ACTIVE_PATH="${2:-}"
      shift 2
      ;;
    --index-path)
      INDEX_PATH="${2:-}"
      shift 2
      ;;
    --timeout)
      TIMEOUT="${2:-}"
      shift 2
      ;;
    --force)
      FORCE=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown option: $1"
      ;;
  esac
done

require_cmd ao
require_cmd git
require_cmd jq

[[ -n "$LABEL" ]] || die "--label is required"
[[ "$LABEL" =~ ^[A-Za-z0-9._-]+$ ]] || die "--label must contain only letters, numbers, ., _, or -"
[[ "$TIMEOUT" =~ ^[0-9]+$ ]] || die "--timeout must be numeric"

OUTPUT_DIR_PATH="$(resolve_path "$REPO_ROOT" "$OUTPUT_DIR")"
LEGACY_PATH_ABS="$(resolve_path "$REPO_ROOT" "$LEGACY_PATH")"
ACTIVE_PATH_ABS="$(resolve_path "$REPO_ROOT" "$ACTIVE_PATH")"
INDEX_PATH_ABS="$(resolve_path "$REPO_ROOT" "$INDEX_PATH")"
OUTPUT_PATH="${OUTPUT_DIR_PATH}/${LABEL}.json"

mkdir -p "$OUTPUT_DIR_PATH" "$(dirname "$LEGACY_PATH_ABS")" "$(dirname "$ACTIVE_PATH_ABS")" "$(dirname "$INDEX_PATH_ABS")"

if [[ -f "$OUTPUT_PATH" && "$FORCE" != true ]]; then
  die "baseline already exists for label: $LABEL"
fi

TMP_FILE="$(mktemp "${TMPDIR:-/tmp}/evolve-baseline.XXXXXX")"
trap 'rm -f "$TMP_FILE"' EXIT

(
  cd "$REPO_ROOT"
  ao goals measure --json --timeout "$TIMEOUT" > "$TMP_FILE"
)

jq -e '.goals | type == "array"' "$TMP_FILE" >/dev/null 2>&1 || die "baseline snapshot is not valid goals JSON"

GOALS_TOTAL="$(jq -r '.goals | length' "$TMP_FILE")"
TIMESTAMP="$(iso_timestamp)"
HEAD_SHA="$(git -C "$REPO_ROOT" rev-parse --short HEAD)"
RELATIVE_OUTPUT="$(relative_to_repo "$REPO_ROOT" "$OUTPUT_PATH")"

mv "$TMP_FILE" "$OUTPUT_PATH"
trap - EXIT
cp "$OUTPUT_PATH" "$LEGACY_PATH_ABS"
printf '%s\n' "$RELATIVE_OUTPUT" > "$ACTIVE_PATH_ABS"

INDEX_ENTRY="$(
  jq -cn \
    --arg label "$LABEL" \
    --arg path "$RELATIVE_OUTPUT" \
    --arg captured_at "$TIMESTAMP" \
    --arg sha "$HEAD_SHA" \
    --argjson goals_total "$GOALS_TOTAL" \
    '{
      label: $label,
      path: $path,
      captured_at: $captured_at,
      sha: $sha,
      goals_total: $goals_total
    }'
)"
printf '%s\n' "$INDEX_ENTRY" >> "$INDEX_PATH_ABS"
printf '%s\n' "$RELATIVE_OUTPUT"
