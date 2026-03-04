#!/usr/bin/env bash
# audit-assertion-density.sh
# Analyze Go test files for assertion density.
# Usage:
#   bash scripts/audit-assertion-density.sh           # Report mode
#   bash scripts/audit-assertion-density.sh --check   # Gate mode (exit 1 if hollow tests found)
set -euo pipefail

CHECK_MODE=false
THRESHOLD=1.5
TARGET_DIR="${1:-cli/cmd/ao}"

if [[ "${1:-}" == "--check" ]]; then
    CHECK_MODE=true
    TARGET_DIR="${2:-cli/cmd/ao}"
fi

# For each *_test.go file, compute assertion density
HOLLOW_COUNT=0
TOTAL_FILES=0

for f in $(find "$TARGET_DIR" -name "*_test.go" -type f | sort); do
    TOTAL_FILES=$((TOTAL_FILES + 1))
    FUNCS=$(grep -c "^func Test" "$f" 2>/dev/null || echo 0)
    ASSERTS=$(grep -cE 't\.(Error|Fatal|Errorf|Fatalf)|assert\.|require\.' "$f" 2>/dev/null || echo 0)

    if [[ "$FUNCS" -gt 0 ]]; then
        # Use awk for float division
        RATIO=$(awk "BEGIN {printf \"%.1f\", $ASSERTS / $FUNCS}")
        HOLLOW=$(awk "BEGIN {print ($RATIO < $THRESHOLD) ? 1 : 0}")

        if [[ "$HOLLOW" -eq 1 ]]; then
            HOLLOW_COUNT=$((HOLLOW_COUNT + 1))
            echo "HOLLOW: $f (ratio=$RATIO, funcs=$FUNCS, asserts=$ASSERTS)"
        else
            echo "OK:     $f (ratio=$RATIO)"
        fi
    fi
done

echo ""
echo "Summary: $HOLLOW_COUNT hollow / $TOTAL_FILES total test files"

if [[ "$CHECK_MODE" == "true" && "$HOLLOW_COUNT" -gt 0 ]]; then
    echo "FAIL: $HOLLOW_COUNT files below threshold ($THRESHOLD)"
    exit 1
fi
