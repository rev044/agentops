#!/usr/bin/env bash
# preflight-swarm.sh — pre-flight checklist for parallel worktree sprints.
# Verifies branch state, pulls latest, builds the CLI, and runs go vet.
# Exit 0 if all checks pass; non-zero on first failure.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

errors=0
step() { printf '\n==> %s\n' "$1"; }
pass() { printf '  PASS: %s\n' "$1"; }
warn() { printf '  WARN: %s\n' "$1"; }
fail() { printf '  FAIL: %s\n' "$1"; errors=$((errors + 1)); }

step "Checking git repository"
if ! git rev-parse --git-dir >/dev/null 2>&1; then
    fail "not inside a git repository"; exit 1
fi
pass "git repository detected"

step "Checking branch"
branch="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$branch" == "main" ]]; then pass "on main branch"
else warn "on branch '$branch' (expected main)"; fi

step "Checking working tree"
dirty="$(git status --porcelain 2>/dev/null)"
if [[ -z "$dirty" ]]; then pass "working tree is clean"
else fail "uncommitted changes detected"
    printf '%s\n' "$dirty" | head -10 | sed 's/^/    /'; fi

step "Pulling latest from origin/main"
if git pull --ff-only origin main 2>&1 | sed 's/^/  /'; then
    pass "up to date with origin/main"
else fail "pull failed — resolve manually"; fi

step "Building CLI (make build)"
if (cd cli && make build) 2>&1 | tail -3 | sed 's/^/  /'; then
    pass "cli binary built"
else fail "make build failed"; fi

step "Running go vet"
if (cd cli && go vet ./cmd/ao/...) 2>&1; then
    pass "go vet clean"
else fail "go vet reported issues"; fi

# --- Shell Hygiene ---
step "Checking shell hygiene (interactive aliases)"
alias_warnings=0
for cmd in cp mv rm; do
    # Check if the command is aliased to its interactive variant
    alias_def="$(alias "$cmd" 2>/dev/null || true)"
    if [[ "$alias_def" == *"-i"* ]]; then
        warn "'$cmd' is aliased to interactive mode: $alias_def"
        warn "  Use /bin/$cmd -f in scripts to bypass"
        alias_warnings=$((alias_warnings + 1))
    fi
done
if [[ "$alias_warnings" -eq 0 ]]; then
    pass "no interactive aliases on cp/mv/rm"
else
    warn "$alias_warnings interactive alias(es) found — agents should use /bin/cp -f, /bin/mv -f, /bin/rm -f"
fi

# --- Summary ---
step "Preflight summary"
commit="$(git rev-parse --short HEAD)"
bin="cli/bin/ao"
if [[ -f "$bin" ]]; then
    bin_age="$(( $(date +%s) - $(stat -f %m "$bin" 2>/dev/null || stat -c %Y "$bin" 2>/dev/null) ))s ago"
else bin_age="missing"; fi

printf '  %-14s %s\n' "Branch:" "$branch"
printf '  %-14s %s\n' "Commit:" "$commit"
printf '  %-14s %s\n' "Binary age:" "$bin_age"
printf '  %-14s %s\n' "Clean:" "$( [[ -z "$dirty" ]] && echo "yes" || echo "no" )"
printf '  %-14s %s\n' "Errors:" "$errors"

if [[ "$errors" -gt 0 ]]; then
    echo ""; echo "FAIL: $errors preflight check(s) failed"; exit 1
fi
echo ""; echo "PASS: all preflight checks passed — ready for worktree sprint"
