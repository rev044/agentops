#!/usr/bin/env bash
set -euo pipefail

# check-doctor-health.sh
# Validates that ao doctor runs without required failures.
# Used by ci-local-release.sh to catch path/namespace drift.
#
# Exit codes:
#   0 = doctor passes (HEALTHY or DEGRADED with no required failures)
#   1 = doctor fails or binary unavailable

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
AO_BIN="$REPO_ROOT/cli/bin/ao"

if [[ ! -x "$AO_BIN" ]]; then
    echo "ao binary not found at $AO_BIN — run 'cd cli && make build' first"
    exit 1
fi

# Run doctor in JSON mode for machine parsing
output=$("$AO_BIN" doctor --json 2>&1) || {
    echo "ao doctor exited with error"
    echo "$output"
    exit 1
}

result=$(echo "$output" | jq -r '.result')
summary=$(echo "$output" | jq -r '.summary')

echo "Doctor: $summary ($result)"

# Fail only on UNHEALTHY (required check failures)
if [[ "$result" == "UNHEALTHY" ]]; then
    echo ""
    echo "Required check(s) failed:"
    echo "$output" | jq -r '.checks[] | select(.status == "fail") | "  \(.name): \(.detail)"'
    exit 1
fi

# Warn on DEGRADED but don't fail
if [[ "$result" == "DEGRADED" ]]; then
    echo ""
    echo "Warnings (non-blocking):"
    echo "$output" | jq -r '.checks[] | select(.status == "warn") | "  \(.name): \(.detail)"'
fi

exit 0
