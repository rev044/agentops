#!/usr/bin/env bats
# pre-push-gate.bats — Tests for scripts/pre-push-gate.sh
#
# Strategy: We stub out external commands (go, git, scripts/*) via PATH
# manipulation so each gate check can be tested in isolation.

setup() {
    REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
    SCRIPT="$REPO_ROOT/scripts/pre-push-gate.sh"

    TMP_DIR="$(mktemp -d)"
    MOCK_BIN="$TMP_DIR/bin"
    mkdir -p "$MOCK_BIN"

    # Build a fake repo with the real script copied in so SCRIPT_DIR resolves here
    FAKE_REPO="$TMP_DIR/repo"
    mkdir -p \
        "$FAKE_REPO/scripts" \
        "$FAKE_REPO/cli" \
        "$FAKE_REPO/hooks" \
        "$FAKE_REPO/cli/embedded/hooks" \
        "$FAKE_REPO/skills/heal-skill/scripts" \
        "$FAKE_REPO/tests/skills"
    /bin/cp "$SCRIPT" "$FAKE_REPO/scripts/pre-push-gate.sh"
    chmod +x "$FAKE_REPO/scripts/pre-push-gate.sh"
    touch "$FAKE_REPO/cli/go.mod"
    # Dummy hooks for sync check (matching content = in sync)
    echo "content" > "$FAKE_REPO/hooks/session-start.sh"
    echo "content" > "$FAKE_REPO/cli/embedded/hooks/session-start.sh"
    echo "content" > "$FAKE_REPO/hooks/hooks.json"
    echo "content" > "$FAKE_REPO/cli/embedded/hooks/hooks.json"

    GATE="$FAKE_REPO/scripts/pre-push-gate.sh"
    make_stub "$FAKE_REPO/scripts/check-worktree-disposition.sh"
    make_stub "$FAKE_REPO/scripts/validate-skill-runtime-parity.sh"
    make_stub "$FAKE_REPO/scripts/validate-codex-skill-parity.sh"
    make_stub "$FAKE_REPO/scripts/validate-codex-install-bundle.sh"
    make_stub "$FAKE_REPO/scripts/validate-codex-runtime-sections.sh"
    make_stub "$FAKE_REPO/scripts/validate-codex-generated-artifacts.sh"
    make_stub "$FAKE_REPO/scripts/validate-skill-runtime-formats.sh"
    make_stub "$FAKE_REPO/scripts/validate-skill-cli-snippets.sh"
    make_stub "$FAKE_REPO/scripts/validate-headless-runtime-skills.sh"
    make_stub "$FAKE_REPO/skills/heal-skill/scripts/heal.sh"
    make_stub "$FAKE_REPO/tests/skills/run-all.sh"
    make_stub "$FAKE_REPO/scripts/validate-skill-schema.sh"
    make_stub "$FAKE_REPO/scripts/validate-manifests.sh"
}

teardown() {
    rm -rf "$TMP_DIR"
}

# Helper: create a stub script that exits with given code
make_stub() {
    local path="$1"
    local exit_code="${2:-0}"
    cat > "$path" <<STUB
#!/usr/bin/env bash
exit $exit_code
STUB
    chmod +x "$path"
}

@test "pre-push-gate.sh exists and is executable" {
    [ -f "$SCRIPT" ]
    [ -x "$SCRIPT" ]
}

@test "pre-push-gate.sh has set -euo pipefail" {
    run grep -q 'set -euo pipefail' "$SCRIPT"
    [ "$status" -eq 0 ]
}

@test "pre-push-gate.sh checks all 19 gates" {
    # Verify the script references all gate sections
    run grep -c '# --- [0-9]' "$SCRIPT"
    [ "$status" -eq 0 ]
    [ "$output" -ge 19 ]
}

@test "pre-push-gate.sh exits 1 on go build failure" {
    # Create a mock go that fails on build
    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
if [[ "$1" == "build" ]]; then exit 1; fi
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    # Create a mock git that reports Go changes
    cat > "$MOCK_BIN/git" <<'GIT'
#!/usr/bin/env bash
if [[ "$*" == *"diff --name-only"* ]]; then echo "cli/cmd/ao/main.go"; fi
if [[ "$*" == *"rev-parse"* ]]; then echo "/tmp"; fi
exit 0
GIT
    chmod +x "$MOCK_BIN/git"

    # Provide passing stubs for all other checks
    make_stub "$FAKE_REPO/scripts/validate-go-fast.sh"
    make_stub "$FAKE_REPO/scripts/check-go-command-test-pair.sh"
    make_stub "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    make_stub "$FAKE_REPO/scripts/sync-skill-counts.sh"

    cd "$FAKE_REPO"
    export PATH="$MOCK_BIN:$PATH"

    run bash "$GATE"
    [ "$status" -eq 1 ]
    [[ "$output" == *"FAIL"*"go build"* ]]
}

@test "pre-push-gate.sh passes when no Go changes" {
    # Mock git to report no Go changes
    cat > "$MOCK_BIN/git" <<'GIT'
#!/usr/bin/env bash
if [[ "$*" == *"diff --name-only"* ]]; then echo ""; fi
exit 0
GIT
    chmod +x "$MOCK_BIN/git"

    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    make_stub "$FAKE_REPO/scripts/validate-go-fast.sh"
    make_stub "$FAKE_REPO/scripts/check-go-command-test-pair.sh"
    make_stub "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    make_stub "$FAKE_REPO/scripts/sync-skill-counts.sh"

    cd "$FAKE_REPO"
    export PATH="$MOCK_BIN:$PATH"

    run bash "$GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == *"passed"* ]]
}

@test "pre-push-gate.sh detects stale embedded hooks" {
    # Make hooks differ
    echo "new-content" > "$FAKE_REPO/hooks/session-start.sh"
    echo "old-content" > "$FAKE_REPO/cli/embedded/hooks/session-start.sh"

    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    cat > "$MOCK_BIN/git" <<'GIT'
#!/usr/bin/env bash
if [[ "$*" == *"diff --name-only"* ]]; then echo ""; fi
exit 0
GIT
    chmod +x "$MOCK_BIN/git"

    make_stub "$FAKE_REPO/scripts/validate-go-fast.sh"
    make_stub "$FAKE_REPO/scripts/check-go-command-test-pair.sh"
    make_stub "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    make_stub "$FAKE_REPO/scripts/sync-skill-counts.sh"

    cd "$FAKE_REPO"
    export PATH="$MOCK_BIN:$PATH"

    run bash "$GATE"
    [ "$status" -eq 1 ]
    [[ "$output" == *"embedded hooks stale"* ]]
}

@test "pre-push-gate.sh fails when validate-go-fast.sh fails" {
    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    cat > "$MOCK_BIN/git" <<'GIT'
#!/usr/bin/env bash
if [[ "$*" == *"diff --name-only"* ]]; then echo ""; fi
exit 0
GIT
    chmod +x "$MOCK_BIN/git"

    make_stub "$FAKE_REPO/scripts/validate-go-fast.sh" 1
    make_stub "$FAKE_REPO/scripts/check-go-command-test-pair.sh"
    make_stub "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    make_stub "$FAKE_REPO/scripts/sync-skill-counts.sh"

    cd "$FAKE_REPO"
    export PATH="$MOCK_BIN:$PATH"

    run bash "$GATE"
    [ "$status" -eq 1 ]
    [[ "$output" == *"FAIL"*"go test -race"* ]]
}

@test "pre-push-gate.sh counts multiple failures" {
    # Make everything fail
    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
if [[ "$1" == "build" ]]; then exit 1; fi
if [[ "$1" == "vet" ]]; then exit 1; fi
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    cat > "$MOCK_BIN/git" <<'GIT'
#!/usr/bin/env bash
if [[ "$*" == *"diff --name-only"* ]]; then echo "cli/cmd/ao/main.go"; fi
exit 0
GIT
    chmod +x "$MOCK_BIN/git"

    make_stub "$FAKE_REPO/scripts/validate-go-fast.sh" 1
    make_stub "$FAKE_REPO/scripts/check-go-command-test-pair.sh" 1
    make_stub "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh" 1
    make_stub "$FAKE_REPO/scripts/sync-skill-counts.sh" 1

    # Make hooks differ too
    echo "new" > "$FAKE_REPO/hooks/session-start.sh"
    echo "old" > "$FAKE_REPO/cli/embedded/hooks/session-start.sh"

    cd "$FAKE_REPO"
    export PATH="$MOCK_BIN:$PATH"

    run bash "$GATE"
    [ "$status" -eq 1 ]
    [[ "$output" == *"BLOCKED"* ]]
}

@test "pre-push-gate.sh fails when worktree disposition check fails" {
    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    cat > "$MOCK_BIN/git" <<'GIT'
#!/usr/bin/env bash
if [[ "$*" == *"diff --name-only"* ]]; then echo ""; fi
exit 0
GIT
    chmod +x "$MOCK_BIN/git"

    make_stub "$FAKE_REPO/scripts/validate-go-fast.sh"
    make_stub "$FAKE_REPO/scripts/check-go-command-test-pair.sh"
    make_stub "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    make_stub "$FAKE_REPO/scripts/sync-skill-counts.sh"
    make_stub "$FAKE_REPO/scripts/check-worktree-disposition.sh" 1

    cd "$FAKE_REPO"
    export PATH="$MOCK_BIN:$PATH"

    run bash "$GATE"
    [ "$status" -eq 1 ]
    [[ "$output" == *"FAIL"*"worktree disposition"* ]]
}
