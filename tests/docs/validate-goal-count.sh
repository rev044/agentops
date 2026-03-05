#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

errors=0

echo "=== Goal Count Validation ==="

# Detect goals file format
if [[ -f "$REPO_ROOT/GOALS.md" ]]; then
    GOALS_FILE="$REPO_ROOT/GOALS.md"
    # Count gates (table rows starting with | followed by lowercase id)
    actual_count=$(grep -cE '^\| [a-z]' "$GOALS_FILE")
    echo "  GOALS.md gates: $actual_count"
elif [[ -f "$REPO_ROOT/GOALS.yaml" ]]; then
    GOALS_FILE="$REPO_ROOT/GOALS.yaml"
    actual_count=$(grep -c "^  - id:" "$GOALS_FILE")
    echo "  GOALS.yaml goals: $actual_count"
else
    echo "FAIL: No GOALS.md or GOALS.yaml found"
    exit 1
fi

# Extract README claim (line like "N measurable goals" or "N gates") — optional
readme_claim=$(grep -oE '[0-9]+ (measurable )?(goals|gates)' "$REPO_ROOT/README.md" | head -1 | grep -oE '^[0-9]+' || echo "")

if [[ -n "$readme_claim" ]]; then
    echo "  README.md claim:   $readme_claim"
    if [[ "$actual_count" -ne "$readme_claim" ]]; then
        echo "FAIL: Goals file has $actual_count entries but README.md claims $readme_claim"
        errors=$((errors + 1))
    fi
else
    echo "  README.md claim:   (none — no hardcoded count, OK)"
fi
echo ""

# Also check goals file count comment if present (YAML format)
if [[ "$GOALS_FILE" == *".yaml" ]]; then
    yaml_claim=$(grep -oE '^# [0-9]+ goals:' "$GOALS_FILE" | head -1 | grep -oE '[0-9]+' || echo "")
    if [[ -n "$yaml_claim" ]]; then
        echo "  GOALS.yaml comment: $yaml_claim"
        if [[ "$actual_count" -ne "$yaml_claim" ]]; then
            echo "FAIL: GOALS.yaml comment says $yaml_claim but actual count is $actual_count"
            errors=$((errors + 1))
        fi
    fi
fi

if [[ "$errors" -gt 0 ]]; then
    echo ""
    echo "FAIL: $errors mismatch(es) found"
    exit 1
else
    echo ""
    echo "PASS: Goal counts consistent (actual=$actual_count)"
    exit 0
fi
