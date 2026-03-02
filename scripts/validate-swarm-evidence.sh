#!/usr/bin/env bash
# validate-swarm-evidence.sh — Validate swarm worker result evidence against schema contract.
# Usage: bash scripts/validate-swarm-evidence.sh <result-file.json>
# Exit 0 = PASS, Exit 1 = FAIL (with structured error output)
set -euo pipefail

RESULT_FILE="${1:-}"
if [[ -z "$RESULT_FILE" || ! -f "$RESULT_FILE" ]]; then
    echo "Usage: bash scripts/validate-swarm-evidence.sh <result-file.json>"
    exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
    echo "ERROR: jq required for evidence validation"
    exit 1
fi

ERRORS=0
WARNINGS=0

# Parse result
TYPE=$(jq -r '.type // "unknown"' "$RESULT_FILE")
STATUS=$(jq -r '.status // "unknown"' "$RESULT_FILE")

# Blocked results skip evidence validation
if [[ "$TYPE" == "blocked" ]] || [[ "$STATUS" == "blocked" ]]; then
    echo "SKIP: blocked result — evidence not required"
    exit 0
fi

# Completion results require evidence
if [[ "$TYPE" == "completion" ]]; then
    # Check artifacts exist
    ARTIFACTS_COUNT=$(jq -r '.artifacts | length // 0' "$RESULT_FILE" 2>/dev/null || echo 0)
    if [[ "$ARTIFACTS_COUNT" -eq 0 ]]; then
        echo "FAIL: completion result missing artifacts array"
        ERRORS=$((ERRORS + 1))
    fi

    # Check evidence block exists
    HAS_EVIDENCE=$(jq -e '.evidence' "$RESULT_FILE" >/dev/null 2>&1 && echo "yes" || echo "no")
    if [[ "$HAS_EVIDENCE" == "no" ]]; then
        echo "FAIL: completion result missing evidence block"
        ERRORS=$((ERRORS + 1))
    else
        # Check required_checks array
        REQ_CHECKS=$(jq -r '.evidence.required_checks // [] | length' "$RESULT_FILE")
        if [[ "$REQ_CHECKS" -eq 0 ]]; then
            echo "FAIL: evidence.required_checks is empty — at least one check required"
            ERRORS=$((ERRORS + 1))
        fi

        # For each required check, verify it exists in checks and has PASS verdict
        for check_name in $(jq -r '.evidence.required_checks[]' "$RESULT_FILE" 2>/dev/null); do
            CHECK_EXISTS=$(jq -e ".evidence.checks[\"$check_name\"]" "$RESULT_FILE" >/dev/null 2>&1 && echo "yes" || echo "no")
            if [[ "$CHECK_EXISTS" == "no" ]]; then
                echo "FAIL: required check '$check_name' missing from evidence.checks"
                ERRORS=$((ERRORS + 1))
                continue
            fi

            VERDICT=$(jq -r ".evidence.checks[\"$check_name\"].verdict // \"missing\"" "$RESULT_FILE")
            if [[ "$VERDICT" == "FAIL" ]]; then
                echo "FAIL: required check '$check_name' has FAIL verdict"
                ERRORS=$((ERRORS + 1))
            elif [[ "$VERDICT" == "SKIP" ]]; then
                echo "WARN: required check '$check_name' has SKIP verdict"
                WARNINGS=$((WARNINGS + 1))
            elif [[ "$VERDICT" != "PASS" ]]; then
                echo "FAIL: required check '$check_name' has invalid verdict: $VERDICT"
                ERRORS=$((ERRORS + 1))
            fi
        done
    fi
fi

# Summary
if [[ $ERRORS -gt 0 ]]; then
    echo ""
    echo "EVIDENCE VALIDATION FAILED: $ERRORS error(s), $WARNINGS warning(s)"
    exit 1
fi

if [[ $WARNINGS -gt 0 ]]; then
    echo ""
    echo "EVIDENCE VALIDATION WARN: $WARNINGS warning(s)"
    exit 0
fi

echo "EVIDENCE VALIDATION PASS"
exit 0
