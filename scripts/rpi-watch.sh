#!/usr/bin/env bash
set -euo pipefail

# rpi-watch.sh
# Live monitor for ao rpi phased runs:
# - state snapshot
# - per-run orchestration log tail
# - tmux session pane tail(s)

usage() {
  cat <<'EOF'
Usage:
  scripts/rpi-watch.sh [run-id] [--once] [--interval SECONDS]

Examples:
  scripts/rpi-watch.sh
  scripts/rpi-watch.sh e9dd5b00348b
  scripts/rpi-watch.sh --once
  scripts/rpi-watch.sh 60d881b65b19 --interval 1
EOF
}

RUN_ID=""
ONCE=0
INTERVAL=2

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    --once)
      ONCE=1
      shift
      ;;
    --interval)
      if [[ $# -lt 2 ]]; then
        echo "error: --interval requires a value" >&2
        exit 1
      fi
      INTERVAL="$2"
      shift 2
      ;;
    *)
      if [[ -z "$RUN_ID" ]]; then
        RUN_ID="$1"
        shift
      else
        echo "error: unexpected argument: $1" >&2
        usage
        exit 1
      fi
      ;;
  esac
done

if ! command -v jq >/dev/null 2>&1; then
  echo "error: jq is required" >&2
  exit 1
fi

if ! command -v tmux >/dev/null 2>&1; then
  echo "error: tmux is required" >&2
  exit 1
fi

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
RUNS_DIR="$REPO_ROOT/.agents/rpi/runs"
LOG_PATH="$REPO_ROOT/.agents/rpi/phased-orchestration.log"

if [[ ! -d "$RUNS_DIR" ]]; then
  echo "error: runs directory not found: $RUNS_DIR" >&2
  exit 1
fi

if [[ -z "$RUN_ID" ]]; then
  RUN_ID="$(find "$RUNS_DIR" -mindepth 1 -maxdepth 1 -type d -print0 \
    | xargs -0 stat -f '%m %N' \
    | sort -nr \
    | awk 'NR==1 {print $2}')"
  RUN_ID="$(basename "${RUN_ID:-}")"
fi

if [[ -z "$RUN_ID" ]]; then
  echo "error: could not determine run id (no run directories found)" >&2
  exit 1
fi

RUN_DIR="$RUNS_DIR/$RUN_ID"
STATE_PATH="$RUN_DIR/phased-state.json"
HEARTBEAT_PATH="$RUN_DIR/heartbeat.txt"
RUN_SHORT="${RUN_ID:0:8}"

if [[ ! -d "$RUN_DIR" ]]; then
  echo "error: run directory not found: $RUN_DIR" >&2
  exit 1
fi

render_once() {
  if [[ -t 1 ]]; then
    clear
  fi

  echo "=== RPI Watch ==="
  echo "time: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  echo "run:  $RUN_ID"
  echo "root: $REPO_ROOT"
  echo

  local phase status backend goal worktree reason
  phase=""
  status=""
  backend=""
  goal=""
  worktree=""
  reason=""
  if [[ -f "$STATE_PATH" ]]; then
    phase="$(jq -r '.phase // "?"' "$STATE_PATH")"
    status="$(jq -r '.terminal_status // "running"' "$STATE_PATH")"
    backend="$(jq -r '.backend // "unknown"' "$STATE_PATH")"
    goal="$(jq -r '.goal // ""' "$STATE_PATH")"
    worktree="$(jq -r '.worktree_path // ""' "$STATE_PATH")"
    reason="$(jq -r '.terminal_reason // ""' "$STATE_PATH")"
    echo "state: phase=$phase backend=$backend status=$status"
    [[ -n "$reason" ]] && echo "reason: $reason"
    [[ -n "$worktree" ]] && echo "worktree: $worktree"
    [[ -n "$goal" ]] && echo "goal: $goal"
  else
    echo "state: (missing $STATE_PATH)"
  fi

  if [[ -f "$HEARTBEAT_PATH" ]]; then
    echo "heartbeat: $(tr -d '\n' < "$HEARTBEAT_PATH")"
  else
    echo "heartbeat: (missing)"
  fi

  echo
  echo "--- Orchestration Log Tail ---"
  local found_log=0
  local candidate_logs=("$LOG_PATH")
  if [[ -n "$worktree" ]]; then
    candidate_logs+=("$worktree/.agents/rpi/phased-orchestration.log")
  fi
  for candidate in "${candidate_logs[@]}"; do
    if [[ -f "$candidate" ]]; then
      local matches
      matches="$(rg "\[$RUN_ID\]" "$candidate" || true)"
      if [[ -n "$matches" ]]; then
        echo "log: $candidate"
        echo "$matches" | tail -n 12
        found_log=1
        break
      fi
    fi
  done
  if [[ "$found_log" -eq 0 ]]; then
    echo "(no matching log lines for run $RUN_ID)"
  fi

  echo
  echo "--- Tmux Sessions ---"
  local sessions
  sessions="$(tmux list-sessions -F '#{session_name}' 2>/dev/null | rg "^ao-rpi-${RUN_SHORT}-p" || true)"
  if [[ -z "$sessions" ]]; then
    echo "(no matching tmux sessions)"
    return
  fi

  while IFS= read -r session; do
    [[ -z "$session" ]] && continue
    echo
    echo "[$session]"
    tmux capture-pane -p -t "$session" | tail -n 20 || true
  done <<< "$sessions"
}

while true; do
  render_once
  if [[ "$ONCE" -eq 1 ]]; then
    break
  fi
  sleep "$INTERVAL"
done
