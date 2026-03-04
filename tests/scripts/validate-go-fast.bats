#!/usr/bin/env bats
# validate-go-fast.bats — Tests for scripts/validate-go-fast.sh
#
# Strategy: Stub git and go via PATH to control which files appear changed
# and whether tests pass or fail.

setup() {
    REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
    SCRIPT="$REPO_ROOT/scripts/validate-go-fast.sh"

    TMP_DIR="$(mktemp -d)"
    MOCK_BIN="$TMP_DIR/bin"
    mkdir -p "$MOCK_BIN"
}

teardown() {
    rm -rf "$TMP_DIR"
}

@test "validate-go-fast.sh exists and is executable" {
    [ -f "$SCRIPT" ]
    [ -x "$SCRIPT" ]
}

@test "validate-go-fast.sh has set -euo pipefail" {
    run grep -q 'set -euo pipefail' "$SCRIPT"
    [ "$status" -eq 0 ]
}

@test "validate-go-fast.sh skips when go not installed" {
    cat > "$MOCK_BIN/git" <<'GIT'
#!/usr/bin/env bash
exit 0
GIT
    chmod +x "$MOCK_BIN/git"

    # PATH with mock bin + essential system tools, but no go
    export PATH="$MOCK_BIN:/usr/bin:/bin"

    run bash "$SCRIPT"
    [ "$status" -eq 0 ]
    [[ "$output" == *"SKIP"*"go not installed"* ]]
}

@test "validate-go-fast.sh skips when no changed files" {
    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    # git reports no files changed
    cat > "$MOCK_BIN/git" <<'GIT'
#!/usr/bin/env bash
case "$*" in
    *"rev-parse --git-dir"*) echo ".git"; exit 0 ;;
    *"rev-parse --abbrev-ref"*) echo "origin/main"; exit 0 ;;
    *"diff --name-only"*) echo ""; exit 0 ;;
    *"show --name-only"*) echo ""; exit 0 ;;
esac
exit 0
GIT
    chmod +x "$MOCK_BIN/git"

    export PATH="$MOCK_BIN:$PATH"

    run bash "$SCRIPT"
    [ "$status" -eq 0 ]
    [[ "$output" == *"SKIP"* ]]
}

@test "validate-go-fast.sh skips when only non-Go files changed" {
    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    cat > "$MOCK_BIN/git" <<'GIT'
#!/usr/bin/env bash
case "$*" in
    *"rev-parse --git-dir"*) echo ".git"; exit 0 ;;
    *"rev-parse --abbrev-ref"*) echo "origin/main"; exit 0 ;;
    *"diff --name-only"*) echo "README.md"; exit 0 ;;
    *"show --name-only"*) echo ""; exit 0 ;;
esac
exit 0
GIT
    chmod +x "$MOCK_BIN/git"

    export PATH="$MOCK_BIN:$PATH"

    run bash "$SCRIPT"
    [ "$status" -eq 0 ]
    [[ "$output" == *"SKIP"*"no Go changes"* ]]
}

@test "validate-go-fast.sh references collect_target_files function" {
    run grep -q 'collect_target_files' "$SCRIPT"
    [ "$status" -eq 0 ]
}

@test "validate-go-fast.sh supports SLOW_THRESHOLD_SECS override" {
    run grep -q 'SLOW_THRESHOLD_SECS' "$SCRIPT"
    [ "$status" -eq 0 ]
}

@test "validate-go-fast.sh handles fork/resource failures with serial fallback" {
    run grep -q 'serial mode' "$SCRIPT"
    [ "$status" -eq 0 ]
}
