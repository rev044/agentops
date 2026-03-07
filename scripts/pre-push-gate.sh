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
#  17. Skill runtime formats
#  18. Skill CLI snippets
#  19. Headless runtime skill smoke
#
# Usage:
#   scripts/pre-push-gate.sh [--scope auto|upstream|staged|worktree|head]
#   (also called from .githooks/pre-push)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

errors=0
SCOPE="${PRE_PUSH_GO_SCOPE:-upstream}"
pass() { echo -e "${GREEN}  ok${NC}  $1"; }
fail() { echo -e "${RED}FAIL${NC}  $1"; errors=$((errors + 1)); }
indent_output() {
    while IFS= read -r line; do
        printf '    %s\n' "$line"
    done <<<"$1"
}

usage() {
    cat <<'EOF'
Usage: scripts/pre-push-gate.sh [--scope auto|upstream|staged|worktree|head]
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
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

echo "pre-push gate: validating before push..."

# --- 1. Go build + vet ---
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

# --- 2. Go race tests on changed scope ---
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

# --- 3. Command/test pairing for command-surface changes ---
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

# --- 4. Embedded hooks sync ---
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
if [[ -x scripts/sync-skill-counts.sh ]]; then
    if scripts/sync-skill-counts.sh --check >/dev/null 2>&1; then
        pass "skill counts in sync"
    else
        fail "skill counts out of sync (run: scripts/sync-skill-counts.sh)"
    fi
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

# --- 9. Codex skill parity ---
if [[ -x scripts/validate-codex-skill-parity.sh ]]; then
    if codex_parity_output="$(scripts/validate-codex-skill-parity.sh 2>&1)"; then
        pass "codex skill parity"
    else
        fail "codex skill parity"
        indent_output "$codex_parity_output"
    fi
else
    fail "missing executable: scripts/validate-codex-skill-parity.sh"
fi

# --- 10. Codex install bundle parity ---
if [[ -x scripts/validate-codex-install-bundle.sh ]]; then
    if codex_bundle_output="$(scripts/validate-codex-install-bundle.sh 2>&1)"; then
        pass "codex install bundle parity"
    else
        fail "codex install bundle parity"
        indent_output "$codex_bundle_output"
    fi
else
    fail "missing executable: scripts/validate-codex-install-bundle.sh"
fi

# --- 11. Codex runtime section format ---
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

# --- 12. Skill integrity ---
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

# --- 13. Skill lint suite ---
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

# --- 14. Skill schema validation ---
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

# --- 15. Manifest schema validation ---
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

# --- 16. Codex generated artifacts ---
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

# --- 17. Skill runtime formats ---
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

# --- 18. Skill CLI snippets ---
if [[ -x scripts/validate-skill-cli-snippets.sh ]]; then
    if skill_cli_output="$(scripts/validate-skill-cli-snippets.sh 2>&1)"; then
        pass "skill CLI snippets"
    else
        fail "skill CLI snippets"
        indent_output "$skill_cli_output"
    fi
else
    fail "missing executable: scripts/validate-skill-cli-snippets.sh"
fi

# --- 19. Headless runtime skill smoke ---
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

# --- Summary ---
echo ""
if [[ $errors -gt 0 ]]; then
    echo -e "${RED}pre-push gate: BLOCKED ($errors failures)${NC}"
    exit 1
else
    echo -e "${GREEN}pre-push gate: passed${NC}"
    exit 0
fi
