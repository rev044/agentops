#!/usr/bin/env bash
# pre-push-gate.sh — lightweight validation before push
#
# Runs the minimum checks to prevent broken code from landing on main.
# Designed to be fast (~10-20s cached) while catching the failures that
# ci-local-release.sh would catch later.
#
# Checks:
#   1. Go build + vet (if cli/ changed)
#   2. Go race tests on changed packages (via validate-go-fast.sh)
#   3. Command/test pairing for cli/cmd/ao Go changes
#   4. cmd/ao coverage floor gate
#   5. Embedded hooks sync (cli/embedded/ matches hooks/)
#   6. Skill count sync
#   7. Worktree disposition
#   8. Skill runtime/CLI parity
#   9. Codex skill parity
#  10. Codex install bundle parity
#  11. Codex runtime section format
#  12. Skill integrity (references/xrefs)
#  13. Skill lint suite
#  14. Skill schema validation
#  15. Manifest schema validation
#  16. Codex generated artifacts
#  17. Codex backbone prompts
#  18. Codex override coverage
#  19. Next-work contract parity
#  20. Skill runtime formats
#  21. Skill CLI snippets
#  22. Headless runtime skill smoke
#
# Usage:
#   scripts/pre-push-gate.sh [--scope auto|upstream|staged|worktree|head]
#   scripts/pre-push-gate.sh --fast [--scope ...]   # only checks relevant to changed files
#   (also called from .githooks/pre-push)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

if [[ -n "${GIT_DIR:-}" && -z "${GIT_WORK_TREE:-}" ]]; then
    export GIT_WORK_TREE="$REPO_ROOT"
fi

run_without_git_env() {
    local var_name
    local -a env_args=(env)
    while IFS='=' read -r var_name _; do
        [[ "$var_name" == GIT_* ]] || continue
        env_args+=("-u" "$var_name")
    done < <(env)
    "${env_args[@]}" "$@"
}

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

errors=0
skipped=0
SCOPE="${PRE_PUSH_GO_SCOPE:-upstream}"
FAST_MODE=false
pass() { echo -e "${GREEN}  ok${NC}  $1"; }
fail() { echo -e "${RED}FAIL${NC}  $1"; errors=$((errors + 1)); }
skip() { echo -e "  --  $1 (skipped)"; skipped=$((skipped + 1)); }
indent_output() {
    while IFS= read -r line; do
        printf '    %s\n' "$line"
    done <<<"$1"
}

usage() {
    cat <<'EOF'
Usage: scripts/pre-push-gate.sh [--fast] [--scope auto|upstream|staged|worktree|head]

Options:
  --fast    Only run checks relevant to changed files (~15-30s vs ~3min)
  --scope   How to determine changed files (default: upstream)
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --fast)
            FAST_MODE=true
            shift
            ;;
        --scope)
            SCOPE="${2:-}"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown arg: $1" >&2
            usage >&2
            exit 2
            ;;
    esac
done

case "$SCOPE" in
    auto|upstream|staged|worktree|head) ;;
    *)
        echo "Invalid --scope: $SCOPE" >&2
        usage >&2
        exit 2
        ;;
esac

collect_all_changed() {
    case "$SCOPE" in
        upstream)
            git diff --name-only '@{upstream}...HEAD' 2>/dev/null || true
            ;;
        staged)
            git diff --name-only --cached 2>/dev/null || true
            ;;
        worktree)
            {
                git diff --name-only --cached 2>/dev/null || true
                git diff --name-only 2>/dev/null || true
            } | sed '/^[[:space:]]*$/d' | sort -u
            ;;
        head)
            git show --name-only --pretty=format: HEAD 2>/dev/null || true
            ;;
        auto)
            {
                git diff --name-only '@{upstream}...HEAD' 2>/dev/null || true
                git diff --name-only --cached 2>/dev/null || true
                git diff --name-only 2>/dev/null || true
            } | sed '/^[[:space:]]*$/d' | sort -u
            ;;
    esac
}

# --- Fast mode: detect changed file categories ---
HAS_GO=1
HAS_SKILL=1
HAS_HOOK=1

if [[ "$FAST_MODE" == "true" ]]; then
    all_changed="$(collect_all_changed)"
    if echo "$all_changed" | grep -qE '^cli/'; then
        HAS_GO=1
    else
        HAS_GO=0
    fi
    if echo "$all_changed" | grep -qE '^skills/|^skills-codex|^tests/skills/'; then
        HAS_SKILL=1
    else
        HAS_SKILL=0
    fi
    if echo "$all_changed" | grep -qE '^hooks/|^lib/'; then
        HAS_HOOK=1
    else
        HAS_HOOK=0
    fi
fi

needs_check() {
    local category="$1"
    if [[ "$FAST_MODE" != "true" ]]; then
        return 0
    fi
    case "$category" in
        go)    [[ "$HAS_GO" -eq 1 ]] ;;
        skill) [[ "$HAS_SKILL" -eq 1 ]] ;;
        hook)  [[ "$HAS_HOOK" -eq 1 ]] ;;
        always) return 0 ;;
        *)     return 0 ;;
    esac
}

collect_go_changed() {
    case "$SCOPE" in
        upstream)
            git diff --name-only '@{upstream}...HEAD' -- 'cli/*.go' 'cli/**/*.go' 'cli/go.mod' 'cli/go.sum' 2>/dev/null || true
            ;;
        staged)
            git diff --name-only --cached -- 'cli/*.go' 'cli/**/*.go' 'cli/go.mod' 'cli/go.sum' 2>/dev/null || true
            ;;
        worktree)
            {
                git diff --name-only --cached -- 'cli/*.go' 'cli/**/*.go' 'cli/go.mod' 'cli/go.sum' 2>/dev/null || true
                git diff --name-only -- 'cli/*.go' 'cli/**/*.go' 'cli/go.mod' 'cli/go.sum' 2>/dev/null || true
            } | sed '/^[[:space:]]*$/d' | sort -u
            ;;
        head)
            git show --name-only --pretty=format: HEAD -- 'cli/*.go' 'cli/**/*.go' 'cli/go.mod' 'cli/go.sum' 2>/dev/null || true
            ;;
        auto)
            {
                git diff --name-only '@{upstream}...HEAD' -- 'cli/*.go' 'cli/**/*.go' 'cli/go.mod' 'cli/go.sum' 2>/dev/null || true
                git diff --name-only --cached -- 'cli/*.go' 'cli/**/*.go' 'cli/go.mod' 'cli/go.sum' 2>/dev/null || true
                git diff --name-only -- 'cli/*.go' 'cli/**/*.go' 'cli/go.mod' 'cli/go.sum' 2>/dev/null || true
            } | sed '/^[[:space:]]*$/d' | sort -u
            ;;
    esac
}

if [[ "$FAST_MODE" == "true" ]]; then
    echo "pre-push gate (fast): validating changed files before push..."
    echo "  go=$HAS_GO skill=$HAS_SKILL hook=$HAS_HOOK"
else
    echo "pre-push gate: validating before push..."
fi

# --- 1. Go build + vet ---
if needs_check go; then
    if command -v go >/dev/null 2>&1 && [[ -f cli/go.mod ]]; then
        go_changed="$(collect_go_changed)"
        if [[ -n "$go_changed" ]]; then
            if (cd cli && go build -o /dev/null ./cmd/ao 2>&1); then
                pass "go build"
            else
                fail "go build"
            fi
            if (cd cli && go vet ./... 2>&1); then
                pass "go vet"
            else
                fail "go vet"
            fi
        else
            pass "go build (no Go changes)"
        fi
    fi
else
    skip "go build + vet"
fi

# --- 2. Go race tests on changed scope ---
if needs_check go; then
    if [[ -x scripts/validate-go-fast.sh ]]; then
        if go_fast_output="$(scripts/validate-go-fast.sh --scope "$SCOPE" 2>&1)"; then
            pass "go test -race (changed scope)"
        else
            fail "go test -race (changed scope)"
            indent_output "$go_fast_output"
        fi
    else
        fail "missing executable: scripts/validate-go-fast.sh"
    fi
else
    skip "go test -race"
fi

# --- 3. Command/test pairing for command-surface changes ---
if needs_check go; then
    if [[ -x scripts/check-go-command-test-pair.sh ]]; then
        if pair_output="$(scripts/check-go-command-test-pair.sh 2>&1)"; then
            pass "command/test pairing"
        else
            fail "command/test pairing"
            indent_output "$pair_output"
        fi
    else
        fail "missing executable: scripts/check-go-command-test-pair.sh"
    fi
else
    skip "command/test pairing"
fi

# --- 4. cmd/ao coverage floor ---
if needs_check go; then
    if [[ -x scripts/check-cmdao-coverage-floor.sh ]]; then
        if coverage_output="$(scripts/check-cmdao-coverage-floor.sh 2>&1)"; then
            pass "cmd/ao coverage floor"
        else
            fail "cmd/ao coverage floor"
            indent_output "$coverage_output"
        fi
    else
        fail "missing executable: scripts/check-cmdao-coverage-floor.sh"
    fi
else
    skip "cmd/ao coverage floor"
fi

# --- 4b. Per-package coverage ratchet (default mode only, not --fast) ---
if [[ "$FAST_MODE" != "true" ]] && needs_check go; then
    if [[ -x scripts/coverage-ratchet.sh ]] && [[ -f .coverage-baseline.json ]]; then
        if ratchet_output="$(scripts/coverage-ratchet.sh --check 2>&1)"; then
            pass "coverage ratchet (per-package)"
        else
            fail "coverage ratchet (per-package)"
            indent_output "$ratchet_output"
        fi
    else
        skip "coverage ratchet (missing script or baseline)"
    fi
else
    skip "coverage ratchet"
fi

# --- 5. Embedded hooks sync ---
stale=0
for src in hooks/session-start.sh hooks/hooks.json; do
    embedded="cli/embedded/$src"
    if [[ -f "$src" ]] && [[ -f "$embedded" ]]; then
        if ! diff -q "$src" "$embedded" >/dev/null 2>&1; then
            stale=1
            break
        fi
    fi
done
if [[ "$stale" -eq 1 ]]; then
    fail "embedded hooks stale (run: cd cli && make sync-hooks)"
else
    pass "embedded hooks in sync"
fi

# --- 6. Skill count sync ---
if needs_check skill; then
    if [[ -x scripts/sync-skill-counts.sh ]]; then
        if scripts/sync-skill-counts.sh --check >/dev/null 2>&1; then
            pass "skill counts in sync"
        else
            fail "skill counts out of sync (run: scripts/sync-skill-counts.sh)"
        fi
    fi
else
    skip "skill counts"
fi

# --- 7. Worktree disposition ---
if [[ -x scripts/check-worktree-disposition.sh ]]; then
    if disposition_output="$(scripts/check-worktree-disposition.sh 2>&1)"; then
        pass "worktree disposition"
    else
        fail "worktree disposition"
        indent_output "$disposition_output"
    fi
else
    fail "missing executable: scripts/check-worktree-disposition.sh"
fi

# --- 8. Skill runtime/CLI parity ---
if needs_check skill; then
    if [[ -x scripts/validate-skill-runtime-parity.sh ]]; then
        if skill_runtime_output="$(scripts/validate-skill-runtime-parity.sh 2>&1)"; then
            pass "skill runtime parity"
        else
            fail "skill runtime parity"
            indent_output "$skill_runtime_output"
        fi
    else
        fail "missing executable: scripts/validate-skill-runtime-parity.sh"
    fi
else
    skip "skill runtime parity"
fi

# --- 9. Codex skill parity --- (removed: skills-codex/ is manually maintained)
skip "codex skill parity (manually maintained)"

# --- 10. Codex install bundle parity --- (removed: skills-codex/ is manually maintained)
skip "codex install bundle parity (manually maintained)"

# --- 11. Codex runtime section format ---
if needs_check skill; then
    if [[ -x scripts/validate-codex-runtime-sections.sh ]]; then
        if codex_runtime_output="$(scripts/validate-codex-runtime-sections.sh 2>&1)"; then
            pass "codex runtime sections"
        else
            fail "codex runtime sections"
            indent_output "$codex_runtime_output"
        fi
    else
        fail "missing executable: scripts/validate-codex-runtime-sections.sh"
    fi
else
    skip "codex runtime sections"
fi

# --- 12. Skill integrity ---
if needs_check skill; then
    if [[ -x skills/heal-skill/scripts/heal.sh ]]; then
        if skill_integrity_output="$(bash skills/heal-skill/scripts/heal.sh --strict 2>&1)"; then
            pass "skill integrity"
        else
            fail "skill integrity"
            indent_output "$skill_integrity_output"
        fi
    else
        fail "missing executable: skills/heal-skill/scripts/heal.sh"
    fi
else
    skip "skill integrity"
fi

# --- 13. Skill lint suite ---
if needs_check skill; then
    if [[ -x tests/skills/run-all.sh ]]; then
        if skill_lint_output="$(bash tests/skills/run-all.sh 2>&1)"; then
            pass "skill lint suite"
        else
            fail "skill lint suite"
            indent_output "$skill_lint_output"
        fi
    else
        fail "missing executable: tests/skills/run-all.sh"
    fi
else
    skip "skill lint suite"
fi

# --- 14. Skill schema validation ---
if needs_check skill; then
    if [[ -x scripts/validate-skill-schema.sh ]]; then
        if skill_schema_output="$(scripts/validate-skill-schema.sh 2>&1)"; then
            pass "skill schema validation"
        else
            fail "skill schema validation"
            indent_output "$skill_schema_output"
        fi
    else
        fail "missing executable: scripts/validate-skill-schema.sh"
    fi
else
    skip "skill schema validation"
fi

# --- 15. Manifest schema validation ---
if needs_check skill; then
    if [[ -x scripts/validate-manifests.sh ]]; then
        if manifest_output="$(scripts/validate-manifests.sh --repo-root . 2>&1)"; then
            pass "manifest schema validation"
        else
            fail "manifest schema validation"
            indent_output "$manifest_output"
        fi
    else
        fail "missing executable: scripts/validate-manifests.sh"
    fi
else
    skip "manifest schema validation"
fi

# --- 16. Codex generated artifacts ---
if needs_check skill; then
    if [[ -x scripts/validate-codex-generated-artifacts.sh ]]; then
        if codex_generated_output="$(scripts/validate-codex-generated-artifacts.sh --scope "$SCOPE" 2>&1)"; then
            pass "codex generated artifacts"
        else
            fail "codex generated artifacts"
            indent_output "$codex_generated_output"
        fi
    else
        fail "missing executable: scripts/validate-codex-generated-artifacts.sh"
    fi
else
    skip "codex generated artifacts"
fi

# --- 17. Codex backbone prompts ---
if needs_check skill; then
    if [[ -x scripts/validate-codex-backbone-prompts.sh ]]; then
        if codex_backbone_output="$(scripts/validate-codex-backbone-prompts.sh 2>&1)"; then
            pass "codex backbone prompts"
        else
            fail "codex backbone prompts"
            indent_output "$codex_backbone_output"
        fi
    else
        fail "missing executable: scripts/validate-codex-backbone-prompts.sh"
    fi
else
    skip "codex backbone prompts"
fi

# --- 18. Codex override coverage ---
if needs_check skill; then
    if [[ -x scripts/validate-codex-override-coverage.sh ]]; then
        if codex_override_output="$(scripts/validate-codex-override-coverage.sh 2>&1)"; then
            pass "codex override coverage"
        else
            fail "codex override coverage"
            indent_output "$codex_override_output"
        fi
    else
        fail "missing executable: scripts/validate-codex-override-coverage.sh"
    fi
else
    skip "codex override coverage"
fi

# --- 19. Next-work contract parity ---
if [[ -x scripts/validate-next-work-contract-parity.sh ]]; then
    if next_work_contract_output="$(scripts/validate-next-work-contract-parity.sh 2>&1)"; then
        pass "next-work contract parity"
    else
        fail "next-work contract parity"
        indent_output "$next_work_contract_output"
    fi
else
    fail "missing executable: scripts/validate-next-work-contract-parity.sh"
fi

# --- 20. Skill runtime formats ---
if needs_check skill; then
    if [[ -x scripts/validate-skill-runtime-formats.sh ]]; then
        if codex_lint_output="$(scripts/validate-skill-runtime-formats.sh 2>&1)"; then
            pass "skill runtime formats"
        else
            fail "skill runtime formats"
            indent_output "$codex_lint_output"
        fi
    else
        fail "missing executable: scripts/validate-skill-runtime-formats.sh"
    fi
else
    skip "skill runtime formats"
fi

# --- 21. Skill CLI snippets ---
if needs_check skill; then
    if [[ -x scripts/validate-skill-cli-snippets.sh ]]; then
        if skill_cli_output="$(run_without_git_env scripts/validate-skill-cli-snippets.sh 2>&1)"; then
            pass "skill CLI snippets"
        else
            fail "skill CLI snippets"
            indent_output "$skill_cli_output"
        fi
    else
        fail "missing executable: scripts/validate-skill-cli-snippets.sh"
    fi
else
    skip "skill CLI snippets"
fi

# --- 22. Headless runtime skill smoke ---
# Skip in fast mode — requires nested Claude/Codex which fails inside Claude sessions
if needs_check always && [[ "$FAST_MODE" != "true" ]]; then
    if [[ -x scripts/validate-headless-runtime-skills.sh ]]; then
        if runtime_smoke_output="$(scripts/validate-headless-runtime-skills.sh 2>&1)"; then
            pass "headless runtime skills"
            indent_output "$runtime_smoke_output"
        else
            fail "headless runtime skills"
            indent_output "$runtime_smoke_output"
        fi
    else
        fail "missing executable: scripts/validate-headless-runtime-skills.sh"
    fi
else
    skip "headless runtime skills"
fi

# --- 23. CLI docs parity (generate-cli-reference.sh --check) ---
if needs_check go; then
    if [[ -x scripts/generate-cli-reference.sh ]]; then
        if cli_docs_output="$(run_without_git_env scripts/generate-cli-reference.sh --check 2>&1)"; then
            pass "CLI docs parity"
        else
            fail "CLI docs parity (run: scripts/generate-cli-reference.sh)"
            indent_output "$cli_docs_output"
        fi
    else
        fail "missing executable: scripts/generate-cli-reference.sh"
    fi
else
    skip "CLI docs parity"
fi

# --- Summary ---
echo ""
if [[ $errors -gt 0 ]]; then
    if [[ "$FAST_MODE" == "true" ]]; then
        echo -e "${RED}pre-push gate (fast): BLOCKED ($errors failures, $skipped skipped)${NC}"
    else
        echo -e "${RED}pre-push gate: BLOCKED ($errors failures)${NC}"
    fi
    exit 1
else
    if [[ "$FAST_MODE" == "true" ]]; then
        echo -e "${GREEN}pre-push gate (fast): passed ($skipped skipped)${NC}"
    else
        echo -e "${GREEN}pre-push gate: passed${NC}"
    fi
    exit 0
fi
