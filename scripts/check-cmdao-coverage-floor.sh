#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CLI_DIR="$REPO_ROOT/cli"

FLOOR="${CMD_AO_COVERAGE_FLOOR:-85.0}"
MAX_ZERO="${CMD_AO_ZERO_COVERAGE_MAX:-95}"
MAX_HANDLER_ZERO="${CMD_AO_HANDLER_ZERO_MAX:-5}"

if ! command -v go >/dev/null 2>&1; then
  echo "SKIP: go is not installed"
  exit 0
fi

tmp_cov="$(mktemp)"
tmp_out="$(mktemp)"
tmp_first_fail="$(mktemp)"
trap 'rm -f "$tmp_cov" "$tmp_out" "$tmp_first_fail"' EXIT

echo "Running cmd/ao coverage gate (floor=${FLOOR}%, zero-max=${MAX_ZERO}, handler-zero-max=${MAX_HANDLER_ZERO})..."

attempt=1
while true; do
  if (cd "$CLI_DIR" && go test -coverprofile="$tmp_cov" -covermode=atomic ./cmd/ao >"$tmp_out" 2>&1); then
    break
  fi

  if [[ "$attempt" -ge 2 ]]; then
    echo "FAIL: go test failed for ./cmd/ao"
    if [[ -s "$tmp_first_fail" ]]; then
      echo "First attempt output:"
      cat "$tmp_first_fail"
      echo ""
      echo "Second attempt output:"
    fi
    cat "$tmp_out"
    exit 1
  fi

  cp "$tmp_out" "$tmp_first_fail"
  echo "WARN: initial covered go test failed for ./cmd/ao; retrying once to filter transient flake..."
  attempt=$((attempt + 1))
done

coverage_report="$(cd "$CLI_DIR" && go tool cover -func="$tmp_cov")"
total_pct="$(printf '%s\n' "$coverage_report" | awk '/^total:/ {gsub("%","",$3); print $3}')"
if [[ -z "$total_pct" ]]; then
  echo "FAIL: unable to determine total coverage for ./cmd/ao"
  exit 1
fi

zero_count="$(printf '%s\n' "$coverage_report" | awk '$1 != "total:" {gsub("%","",$3); if ($3 == "0.0") c++} END {print c+0}')"

if ! awk -v value="$total_pct" -v floor="$FLOOR" 'BEGIN { exit !(value+0 >= floor+0) }'; then
  echo "FAIL: cmd/ao coverage ${total_pct}% is below floor ${FLOOR}%"
  exit 1
fi

if (( zero_count > MAX_ZERO )); then
  echo "FAIL: cmd/ao zero-coverage functions ${zero_count} exceeds max ${MAX_ZERO}"
  exit 1
fi

handler_zero_count="$(printf '%s\n' "$coverage_report" | awk '$1 != "total:" && $1 ~ /Handler/ {gsub("%","",$3); if ($3 == "0.0") c++} END {print c+0}')"

if (( handler_zero_count > MAX_HANDLER_ZERO )); then
  echo "FAIL: cmd/ao handler-family zero-coverage functions ${handler_zero_count} exceeds max ${MAX_HANDLER_ZERO}"
  exit 1
fi

echo "PASS: cmd/ao coverage ${total_pct}% (zero-coverage functions: ${zero_count}, handler-zero: ${handler_zero_count})"
