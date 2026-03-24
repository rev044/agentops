#!/usr/bin/env bash
set -euo pipefail

# test-codex-hookless-lifecycle.sh
# Verifies the non-release Codex hookless lifecycle path against a locally built ao binary.
#
# What it checks:
# 1) builds ao from the current worktree unless --ao-bin is supplied
# 2) seeds a temp Codex home with session_index + history fallback data
# 3) seeds a temp repo with repo-local learnings
# 4) verifies ao codex start writes startup context/state and retrieved citations
# 5) verifies ao lookup and ao search --local --cite record citations
# 6) verifies ao codex stop uses history fallback and writes lifecycle state
# 7) verifies ao codex status reports coherent hookless health and citation counts
# 8) verifies the Codex RPI no-beads contract stays executable via the repo validator
#
# Usage:
#   bash scripts/test-codex-hookless-lifecycle.sh
#   bash scripts/test-codex-hookless-lifecycle.sh --keep-temp
#   bash scripts/test-codex-hookless-lifecycle.sh --ao-bin /path/to/ao

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

AO_BIN=""
KEEP_TEMP="false"
TMP_ROOT=""

usage() {
  cat <<'EOF'
test-codex-hookless-lifecycle.sh

Options:
  --ao-bin <path>   Use an existing ao binary instead of building ./cli/cmd/ao
  --keep-temp       Preserve the temp Codex home + repo for inspection
  --help            Show this help

Examples:
  bash scripts/test-codex-hookless-lifecycle.sh
  bash scripts/test-codex-hookless-lifecycle.sh --keep-temp
  bash scripts/test-codex-hookless-lifecycle.sh --ao-bin ./cli/bin/ao
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --ao-bin)
      AO_BIN="${2:-}"
      shift 2
      ;;
    --keep-temp)
      KEEP_TEMP="true"
      shift 1
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

info() {
  echo "INFO: $*"
}

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "Required command not found: $1"
}

cleanup() {
  if [[ "$KEEP_TEMP" != "true" && -n "${TMP_ROOT:-}" && -d "$TMP_ROOT" ]]; then
    rm -rf "$TMP_ROOT"
  fi
}
trap cleanup EXIT

require_cmd bash
require_cmd git
require_cmd jq
require_cmd rg

TMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/codex-hookless-lifecycle-test.XXXXXX")"
HOME_ROOT="$TMP_ROOT/home"
REPO_DIR="$TMP_ROOT/repo"
CODEX_HOME="$HOME_ROOT/.codex"
SESSION_ID="019d1bf7-58ea-79e1-9f5d-02109d930081"
QUERY="explicit codex lifecycle"

mkdir -p "$CODEX_HOME" "$REPO_DIR/.agents/learnings"

if [[ -z "$AO_BIN" ]]; then
  require_cmd go
  AO_BIN="$TMP_ROOT/ao"
  info "Building ao from current worktree"
  (
    cd "$REPO_ROOT/cli"
    go build -o "$AO_BIN" ./cmd/ao
  )
fi

[[ -x "$AO_BIN" ]] || fail "ao binary is not executable: $AO_BIN"

cat > "$CODEX_HOME/session_index.jsonl" <<EOF
{"id":"$SESSION_ID","thread_name":"Hookless fallback smoke","updated_at":"2026-03-23T12:00:00Z"}
EOF

cat > "$CODEX_HOME/history.jsonl" <<EOF
{"session_id":"$SESSION_ID","ts":1766945655,"text":"Review prior learnings for codex fallback"}
{"session_id":"$SESSION_ID","ts":1766945658,"text":"Implement explicit ao codex stop closeout"}
EOF

cat > "$REPO_DIR/.agents/learnings/codex-lifecycle.md" <<'EOF'
---
id: codex-lifecycle
type: learning
date: 2026-03-23
source: codex-test
maturity: provisional
utility: 0.9
---

# Explicit Codex lifecycle

Use ao codex start and ao codex stop when runtime hooks are unavailable.
EOF

cat > "$REPO_DIR/.agents/learnings/codex-lifecycle.jsonl" <<'EOF'
{"id":"codex-lifecycle-jsonl","summary":"Explicit Codex lifecycle uses ao codex start and ao codex stop when runtime hooks are unavailable.","maturity":"provisional","utility":0.9}
EOF

git init -q "$REPO_DIR"

run_ao() {
  (
    cd "$REPO_DIR"
    HOME="$HOME_ROOT" \
      CODEX_THREAD_ID="$SESSION_ID" \
      CODEX_INTERNAL_ORIGINATOR_OVERRIDE="Codex Desktop" \
      "$AO_BIN" "$@"
  )
}

assert_json() {
  local json="$1"
  local expr="$2"
  local message="$3"
  printf '%s' "$json" | jq -e "$expr" >/dev/null || fail "$message"
}

info "Running ao codex start"
start_json="$(run_ao codex start --json --query "$QUERY")"
assert_json "$start_json" '.runtime.mode == "codex-hookless-fallback"' "codex start did not detect hookless Codex mode"
assert_json "$start_json" '.learnings | length >= 1' "codex start did not surface repo-local learnings"

START_CONTEXT_PATH="$(printf '%s' "$start_json" | jq -r '.startup_context_path')"
STATE_PATH="$(printf '%s' "$start_json" | jq -r '.state_path')"
MEMORY_PATH="$(printf '%s' "$start_json" | jq -r '.memory_path')"
[[ -f "$START_CONTEXT_PATH" ]] || fail "startup context not written: $START_CONTEXT_PATH"
[[ -f "$STATE_PATH" ]] || fail "lifecycle state not written on start: $STATE_PATH"
[[ -n "$MEMORY_PATH" && "$MEMORY_PATH" != "null" ]] || fail "codex start did not report a memory path"

CITATIONS_PATH="$REPO_DIR/.agents/ao/citations.jsonl"
[[ -f "$CITATIONS_PATH" ]] || fail "citations ledger missing after codex start"
rg -q '"citation_type":"retrieved"' "$CITATIONS_PATH" || fail "codex start did not record retrieved citations"

info "Running ao lookup for curated retrieval"
lookup_json="$(run_ao lookup --query "$QUERY" --limit 1 --json)"
assert_json "$lookup_json" '.learnings | length >= 1' "lookup did not return the seeded learning"

info "Running ao search --local with assisted citation"
search_json="$(run_ao search "$QUERY" --local --json --cite reference)"
assert_json "$search_json" 'length >= 1' "local search did not return the seeded learning"
rg -q '"citation_type":"reference"' "$CITATIONS_PATH" || fail "search --cite reference did not record a reference citation"

info "Running ao codex stop"
stop_json="$(run_ao codex stop --json)"
assert_json "$stop_json" '.runtime.mode == "codex-hookless-fallback"' "codex stop did not report hookless Codex mode"
assert_json "$stop_json" '.transcript_source == "history-fallback"' "codex stop did not use history fallback in the temp Codex home"
assert_json "$stop_json" '.synthetic_transcript == true' "codex stop did not mark the history transcript as synthetic"

TRANSCRIPT_PATH="$(printf '%s' "$stop_json" | jq -r '.transcript_path')"
HANDOFF_PATH="$(printf '%s' "$stop_json" | jq -r '.session.handoff_written')"
[[ -f "$TRANSCRIPT_PATH" ]] || fail "history fallback transcript missing: $TRANSCRIPT_PATH"
[[ -f "$HANDOFF_PATH" ]] || fail "closeout handoff artifact missing: $HANDOFF_PATH"

info "Running ao codex status"
status_json="$(run_ao codex status --json)"
assert_json "$status_json" '.runtime.mode == "codex-hookless-fallback"' "codex status did not report hookless Codex mode"
printf '%s' "$status_json" | jq -e --arg path "$START_CONTEXT_PATH" '.state.last_start.startup_context_path == $path' >/dev/null \
  || fail "codex status did not preserve the startup context path"
assert_json "$status_json" '.state.last_stop.transcript_source == "history-fallback"' "codex status did not preserve the stop transcript source"
assert_json "$status_json" '.retrieval.learnings >= 1' "codex status did not report retrievable learnings"
assert_json "$status_json" '.capture.sessions_indexed >= 1' "codex status did not report indexed sessions"
assert_json "$status_json" '.citations.retrieved >= 2' "codex status retrieved citation count is too low"
assert_json "$status_json" '.citations.reference >= 1' "codex status reference citation count is too low"
assert_json "$status_json" '.citations.total >= 3' "codex status total citation count is too low"

echo ""
info "Running Codex RPI contract validation"
bash "$REPO_ROOT/scripts/validate-codex-rpi-contract.sh"

echo "PASS: Codex hookless lifecycle smoke verified"
echo "  ao binary: $AO_BIN"
echo "  temp root: $TMP_ROOT"
echo "  repo dir: $REPO_DIR"

if [[ "$KEEP_TEMP" != "true" ]]; then
  echo "  temp files: cleaned on exit"
fi
