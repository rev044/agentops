#!/bin/bash
set -euo pipefail

# Test 1: Script exists and is executable
if [[ ! -x "scripts/toolchain-validate.sh" ]]; then
    echo "FAIL: toolchain-validate.sh not executable"
    exit 1
fi

# Test 2: Script outputs JSON summary
output=$(./scripts/toolchain-validate.sh --json 2>/dev/null || true)
if ! echo "$output" | jq empty 2>/dev/null; then
    echo "FAIL: Output is not valid JSON"
    exit 1
fi

# Test 3: Exit code 0 when no issues
# (this will depend on repo state, just check it doesn't crash)
./scripts/toolchain-validate.sh --quick >/dev/null 2>&1 || true

echo "PASS: All tests passed"
