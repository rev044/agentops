#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CLI_DIR="$REPO_ROOT/cli"

FLOOR="${CMD_AO_COVERAGE_FLOOR:-78.0}"
MAX_ZERO="${CMD_AO_ZERO_COVERAGE_MAX:-20}"

if ! command -v go >/dev/null 2>&1; then
  echo "SKIP: go is not installed"
  exit 0
fi

tmp_cov="$(mktemp)"
tmp_out="$(mktemp)"
trap 'rm -f "$tmp_cov" "$tmp_out"' EXIT

echo "Running cmd/ao coverage gate (floor=${FLOOR}%, zero-max=${MAX_ZERO})..."

if ! (cd "$CLI_DIR" && go test -coverprofile="$tmp_cov" -covermode=atomic ./cmd/ao >"$tmp_out" 2>&1); then
  echo "FAIL: go test failed for ./cmd/ao"
  cat "$tmp_out"
  exit 1
fi

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

echo "PASS: cmd/ao coverage ${total_pct}% (zero-coverage functions: ${zero_count})"
