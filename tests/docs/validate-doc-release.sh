#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
FREEZE_OVERRIDE="${DOC_RELEASE_FREEZE_OVERRIDE:-false}"
FREEZE_REASON="${DOC_RELEASE_FREEZE_REASON:-}"

errors=0

is_truthy() {
  local value
  value="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')"
  case "$value" in
    1|true|yes|y|on)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

run_check() {
  local name="$1"
  shift

  echo "=== $name ==="
  if "$@"; then
    echo "PASS: $name"
  else
    echo "FAIL: $name"
    errors=$((errors + 1))
  fi
  echo ""
}

validate_message_freeze() {
  local goreleaser_file="$REPO_ROOT/.goreleaser.yml"
  local expected_header_prefix='`brew upgrade agentops`'
  local expected_header_checksums='[checksums](https://github.com/boshu2/agentops/releases/download/{{ .Tag }}/checksums.txt)'
  local expected_header_provenance='[verify provenance](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds)'
  local expected_footer='This release was planned, built, and validated by AgentOps agents. The flywheel turns.'
  local mismatches=0

  if [[ ! -f "$goreleaser_file" ]]; then
    echo "MISMATCH: .goreleaser.yml not found"
    return 1
  fi

  if is_truthy "$FREEZE_OVERRIDE"; then
    if [[ -z "$FREEZE_REASON" ]]; then
      echo "ERROR: DOC_RELEASE_FREEZE_OVERRIDE is set, but DOC_RELEASE_FREEZE_REASON is empty"
      return 1
    fi
    echo "OVERRIDE: message freeze check bypassed"
    echo "Reason: $FREEZE_REASON"
    return 0
  fi

  if ! grep -Fq "$expected_header_prefix" "$goreleaser_file"; then
    echo "MISMATCH: release header install command changed in .goreleaser.yml"
    mismatches=$((mismatches + 1))
  fi

  if ! grep -Fq "$expected_header_checksums" "$goreleaser_file"; then
    echo "MISMATCH: release header checksums link changed in .goreleaser.yml"
    mismatches=$((mismatches + 1))
  fi

  if ! grep -Fq "$expected_header_provenance" "$goreleaser_file"; then
    echo "MISMATCH: release header provenance link changed in .goreleaser.yml"
    mismatches=$((mismatches + 1))
  fi

  if ! grep -Fq "$expected_footer" "$goreleaser_file"; then
    echo "MISMATCH: release footer message changed in .goreleaser.yml"
    mismatches=$((mismatches + 1))
  fi

  if [[ "$mismatches" -gt 0 ]]; then
    echo "To intentionally bypass in CI, set:"
    echo "  DOC_RELEASE_FREEZE_OVERRIDE=true"
    echo "  DOC_RELEASE_FREEZE_REASON='<why this change is required>'"
    return 1
  fi

  echo "PASS: release message freeze intact"
  return 0
}

run_check "Link validation" bash "$REPO_ROOT/tests/docs/validate-links.sh"
run_check "Skill count validation" bash "$REPO_ROOT/tests/docs/validate-skill-count.sh"
run_check "Skill count sync check" bash "$REPO_ROOT/scripts/sync-skill-counts.sh" --check
run_check "Release message freeze validation" validate_message_freeze

if [[ "$errors" -gt 0 ]]; then
  echo "FAIL: doc-release gate failed ($errors check(s) failed)"
  exit 1
fi

echo "PASS: doc-release gate succeeded"
