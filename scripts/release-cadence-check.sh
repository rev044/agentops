#!/usr/bin/env bash
set -euo pipefail

# release-cadence-check.sh
# Enforces weekly release cadence policy.
# - WARN: last release <7 days ago (non-security)
# - FAIL: last release <1 day ago (non-security)
# - PASS: no recent release, or security hotfix
#
# Usage:
#   ./scripts/release-cadence-check.sh              # default (warn mode)
#   ./scripts/release-cadence-check.sh --strict      # fail on <7 days
#   ./scripts/release-cadence-check.sh --security     # bypass (security hotfix)
#
# Exit codes:
#   0 = pass/warn
#   1 = blocked (release too recent)

STRICT=false
SECURITY=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --strict) STRICT=true; shift ;;
        --security) SECURITY=true; shift ;;
        *) echo "Unknown flag: $1" >&2; exit 1 ;;
    esac
done

# Find the latest semver tag
LATEST_TAG=$(git tag --sort=-version:refname -l 'v*' 2>/dev/null | head -1)

if [[ -z "$LATEST_TAG" ]]; then
    echo "PASS: No previous releases found (first release)"
    exit 0
fi

# Get the tag date as Unix timestamp
TAG_DATE=$(git log -1 --format=%ct "$LATEST_TAG" 2>/dev/null)
if [[ -z "$TAG_DATE" ]]; then
    echo "WARN: Could not determine date for $LATEST_TAG"
    exit 0
fi

NOW=$(date +%s)
DAYS_AGO=$(( (NOW - TAG_DATE) / 86400 ))

if [[ "$SECURITY" == "true" ]]; then
    echo "PASS: Security hotfix — cadence bypass (last release: $LATEST_TAG, ${DAYS_AGO}d ago)"
    exit 0
fi

if [[ "$DAYS_AGO" -lt 1 ]]; then
    echo "FAIL: $LATEST_TAG was released today. No multiple releases per day (unless security hotfix)."
    echo "  Use --security flag if this is a security hotfix."
    exit 1
fi

if [[ "$DAYS_AGO" -lt 7 ]]; then
    if [[ "$STRICT" == "true" ]]; then
        echo "FAIL: $LATEST_TAG was released ${DAYS_AGO}d ago. Weekly release train policy requires 7-day spacing."
        echo "  Use --security flag if this is a security hotfix."
        exit 1
    fi
    echo "WARN: $LATEST_TAG was released ${DAYS_AGO}d ago (weekly cadence target: 7 days)."
    echo "  Batch non-security changes into the next weekly release."
    exit 0
fi

echo "PASS: Last release $LATEST_TAG was ${DAYS_AGO}d ago (cadence OK)"
exit 0
