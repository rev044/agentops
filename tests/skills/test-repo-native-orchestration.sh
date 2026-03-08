#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL + 1)); }

check_contains() {
    local file="$1"
    local pattern="$2"
    local message="$3"
    if grep -qE "$pattern" "$file"; then
        pass "$message"
    else
        fail "$message"
    fi
}

EVOLVE="$REPO_ROOT/skills/evolve/SKILL.md"
EVOLVE_EXAMPLES="$REPO_ROOT/skills/evolve/references/examples.md"
RPI="$REPO_ROOT/skills/rpi/SKILL.md"
RPI_CONTRACT="$REPO_ROOT/skills/rpi/references/phase-data-contracts.md"
INDEX="$REPO_ROOT/docs/INDEX.md"
PROFILE_DOC="$REPO_ROOT/docs/contracts/repo-execution-profile.md"
PROFILE_SCHEMA="$REPO_ROOT/docs/contracts/repo-execution-profile.schema.json"

check_contains "$EVOLVE" 'repo execution profile' "/evolve documents repo execution profile bootstrap"
check_contains "$EVOLVE" 'startup_reads' "/evolve documents ordered startup reads"
check_contains "$EVOLVE" 'validation_commands' "/evolve records repo validation commands"
check_contains "$EVOLVE" 'definition_of_done' "/evolve records repo done criteria"
check_contains "$EVOLVE_EXAMPLES" 'repo execution profile' "/evolve examples mention repo bootstrap"

check_contains "$RPI" 'execution packet' "/rpi documents execution packet handoff"
check_contains "$RPI" 'contract_surfaces' "/rpi documents execution packet contract surfaces"
check_contains "$RPI" 'done_criteria' "/rpi documents execution packet done criteria"
check_contains "$RPI_CONTRACT" 'repo execution profile' "/rpi phase data contract names the repo profile"
check_contains "$RPI_CONTRACT" 'execution_packet' "/rpi phase data contract names the execution packet artifact"

check_contains "$INDEX" 'contracts/repo-execution-profile.md' "docs index catalogs repo execution profile doc"
check_contains "$INDEX" 'contracts/repo-execution-profile.schema.json' "docs index catalogs repo execution profile schema"
check_contains "$PROFILE_DOC" 'repo-execution-profile.schema.json' "repo execution profile doc references its schema"
check_contains "$PROFILE_SCHEMA" '"definition_of_done"' "repo execution profile schema includes definition_of_done"

echo
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
