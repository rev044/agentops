#!/usr/bin/env bash
# test_helper.bash — Shared setup/teardown for hook bats tests
#
# Provides:
#   - REPO_ROOT, HOOKS_DIR globals
#   - TMP_TEST_DIR: per-test temp directory (auto-cleaned)
#   - MOCK_REPO: a git-initialized mock repo with lib/ helpers copied in
#   - setup_mock_repo DIR: creates a git-initialized repo at DIR with lib/ copied
#   - run_hook HOOK_SCRIPT JSON_STRING [env assignments ...]: pipes JSON to hook

# ── Common setup ─────────────────────────────────────────────────────
_helper_setup() {
    REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
    HOOKS_DIR="$REPO_ROOT/hooks"
    export REPO_ROOT HOOKS_DIR

    # Per-test temp directory
    TMP_TEST_DIR="$(mktemp -d)"

    # Default mock repo
    MOCK_REPO="$TMP_TEST_DIR/mock-repo"
    setup_mock_repo "$MOCK_REPO"

    # Ensure hooks are NOT globally disabled by default
    export AGENTOPS_HOOKS_DISABLED=0
}

# ── Common teardown ──────────────────────────────────────────────────
_helper_teardown() {
    rm -rf "$TMP_TEST_DIR"
    # Clean up any dedup flags that may have leaked into the real repo
    rm -f "$REPO_ROOT/.agents/ao/.intent-echo-fired" 2>/dev/null
    rm -f "$REPO_ROOT/.agents/ao/.new-user-welcome-needed" 2>/dev/null
    rm -f "$REPO_ROOT/.agents/ao/.read-streak" 2>/dev/null
    rm -f "$REPO_ROOT/.agents/ao/.ratchet-advance-fired" 2>/dev/null
}

# ── Helpers ──────────────────────────────────────────────────────────

# setup_mock_repo DIR
#   Creates a git-initialized repo at DIR with lib/ helpers copied in.
setup_mock_repo() {
    local dir="$1"
    mkdir -p "$dir/.agents/ao" "$dir/lib"
    git -C "$dir" init -q >/dev/null 2>&1
    [ -f "$REPO_ROOT/lib/hook-helpers.sh" ] && /bin/cp "$REPO_ROOT/lib/hook-helpers.sh" "$dir/lib/hook-helpers.sh"
    [ -f "$REPO_ROOT/lib/chain-parser.sh" ] && /bin/cp "$REPO_ROOT/lib/chain-parser.sh" "$dir/lib/chain-parser.sh"
}

# run_hook HOOK_SCRIPT JSON_STRING
#   Pipes JSON_STRING to HOOK_SCRIPT via stdin, captures stdout+stderr.
#   After call, $status and $output are set (like bats `run`).
run_hook() {
    local hook="$1"
    local json="$2"
    run bash -c 'printf "%s" "$1" | bash "$2" 2>&1' -- "$json" "$hook"
}
