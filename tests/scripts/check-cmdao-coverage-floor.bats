#!/usr/bin/env bats
# check-cmdao-coverage-floor.bats — Tests for scripts/check-cmdao-coverage-floor.sh
#
# Strategy: Stub go via PATH to control coverage output without running
# actual Go tests.

setup() {
    REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
    SCRIPT="$REPO_ROOT/scripts/check-cmdao-coverage-floor.sh"

    TMP_DIR="$(mktemp -d)"
    MOCK_BIN="$TMP_DIR/bin"
    mkdir -p "$MOCK_BIN"

    # Script expects cli/go.mod parent
    FAKE_REPO="$TMP_DIR/repo"
    mkdir -p "$FAKE_REPO/scripts" "$FAKE_REPO/cli"
    /bin/cp "$SCRIPT" "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    chmod +x "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
}

teardown() {
    rm -rf "$TMP_DIR"
}

@test "check-cmdao-coverage-floor.sh exists and is executable" {
    [ -f "$SCRIPT" ]
    [ -x "$SCRIPT" ]
}

@test "check-cmdao-coverage-floor.sh has set -euo pipefail" {
    run grep -q 'set -euo pipefail' "$SCRIPT"
    [ "$status" -eq 0 ]
}

@test "check-cmdao-coverage-floor.sh skips when go not installed" {
    # Symlink essentials so the script can initialize, but exclude go
    ln -sf "$(command -v bash)" "$MOCK_BIN/bash"
    ln -sf "$(command -v dirname)" "$MOCK_BIN/dirname"

    run env PATH="$MOCK_BIN" bash "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    [ "$status" -eq 0 ]
    [[ "$output" == *"SKIP"*"go is not installed"* ]]
}

@test "check-cmdao-coverage-floor.sh uses CMD_AO_COVERAGE_FLOOR env" {
    run grep -q 'CMD_AO_COVERAGE_FLOOR' "$SCRIPT"
    [ "$status" -eq 0 ]
}

@test "check-cmdao-coverage-floor.sh default floor is 81%" {
    run grep -q '81.0' "$SCRIPT"
    [ "$status" -eq 0 ]
}

@test "check-cmdao-coverage-floor.sh passes above floor" {
    # Stub go to simulate passing tests with coverage above floor
    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
case "$1" in
    test)
        # Write a fake coverprofile
        for arg in "$@"; do
            case "$arg" in
                -coverprofile=*)
                    cov_file="${arg#-coverprofile=}"
                    echo "mode: atomic" > "$cov_file"
                    echo "example.com/cli/cmd/ao/main.go:10.1,20.1 5 1" >> "$cov_file"
                    ;;
            esac
        done
        echo "ok example.com/cli/cmd/ao 1.234s"
        exit 0
        ;;
    tool)
        # go tool cover -func outputs coverage summary
        if [[ "$2" == "cover" && "$3" == "-func="* ]]; then
            echo "example.com/cli/cmd/ao/main.go:10:	main		100.0%"
            echo "total:						(statements)	90.5%"
            exit 0
        fi
        ;;
esac
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    cd "$FAKE_REPO"
    export PATH="$MOCK_BIN:$PATH"

    run bash "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    [ "$status" -eq 0 ]
    [[ "$output" == *"PASS"* ]]
    [[ "$output" == *"90.5%"* ]]
}

@test "check-cmdao-coverage-floor.sh fails below floor" {
    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
case "$1" in
    test)
        for arg in "$@"; do
            case "$arg" in
                -coverprofile=*)
                    cov_file="${arg#-coverprofile=}"
                    echo "mode: atomic" > "$cov_file"
                    echo "example.com/cli/cmd/ao/main.go:10.1,20.1 5 1" >> "$cov_file"
                    ;;
            esac
        done
        echo "ok example.com/cli/cmd/ao 1.234s"
        exit 0
        ;;
    tool)
        if [[ "$2" == "cover" && "$3" == "-func="* ]]; then
            echo "example.com/cli/cmd/ao/main.go:10:	main		50.0%"
            echo "total:						(statements)	50.0%"
            exit 0
        fi
        ;;
esac
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    cd "$FAKE_REPO"
    export PATH="$MOCK_BIN:$PATH"

    run bash "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    [ "$status" -eq 1 ]
    [[ "$output" == *"FAIL"*"below floor"* ]]
}

@test "check-cmdao-coverage-floor.sh fails on too many zero-coverage functions" {
    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
case "$1" in
    test)
        for arg in "$@"; do
            case "$arg" in
                -coverprofile=*)
                    cov_file="${arg#-coverprofile=}"
                    echo "mode: atomic" > "$cov_file"
                    echo "example.com/cli/cmd/ao/main.go:10.1,20.1 5 1" >> "$cov_file"
                    ;;
            esac
        done
        exit 0
        ;;
    tool)
        if [[ "$2" == "cover" && "$3" == "-func="* ]]; then
            # 100 zero-coverage functions (exceeds MAX_ZERO=95)
            for i in $(seq 1 100); do
                echo "example.com/cli/cmd/ao/main.go:$i:	func${i}		0.0%"
            done
            echo "example.com/cli/cmd/ao/main.go:100:	main		100.0%"
            echo "total:						(statements)	90.0%"
            exit 0
        fi
        ;;
esac
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    cd "$FAKE_REPO"
    export PATH="$MOCK_BIN:$PATH"

    run bash "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    [ "$status" -eq 1 ]
    [[ "$output" == *"FAIL"*"zero-coverage functions"* ]]
}

@test "check-cmdao-coverage-floor.sh fails on too many handler zero-coverage" {
    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
case "$1" in
    test)
        for arg in "$@"; do
            case "$arg" in
                -coverprofile=*)
                    cov_file="${arg#-coverprofile=}"
                    echo "mode: atomic" > "$cov_file"
                    echo "example.com/cli/cmd/ao/main.go:10.1,20.1 5 1" >> "$cov_file"
                    ;;
            esac
        done
        exit 0
        ;;
    tool)
        if [[ "$2" == "cover" && "$3" == "-func="* ]]; then
            # 6 handler-family zero-coverage functions (exceeds MAX_HANDLER_ZERO=5)
            # The script checks $1 (file path) for "Handler", so put it in the path
            # but only 6 total zero (within MAX_ZERO=8)
            for i in $(seq 1 6); do
                echo "example.com/cli/cmd/ao/Handler${i}.go:$i:	run		0.0%"
            done
            # Many non-zero functions to keep total coverage high
            for i in $(seq 1 50); do
                echo "example.com/cli/cmd/ao/main.go:$((100+i)):	func${i}		100.0%"
            done
            echo "total:						(statements)	90.0%"
            exit 0
        fi
        ;;
esac
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    cd "$FAKE_REPO"
    export PATH="$MOCK_BIN:$PATH"

    run bash "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    [ "$status" -eq 1 ]
    [[ "$output" == *"FAIL"*"handler-family"* ]]
}

@test "check-cmdao-coverage-floor.sh respects custom floor via env" {
    cat > "$MOCK_BIN/go" <<'GO'
#!/usr/bin/env bash
case "$1" in
    test)
        for arg in "$@"; do
            case "$arg" in
                -coverprofile=*)
                    cov_file="${arg#-coverprofile=}"
                    echo "mode: atomic" > "$cov_file"
                    ;;
            esac
        done
        exit 0
        ;;
    tool)
        if [[ "$2" == "cover" && "$3" == "-func="* ]]; then
            echo "total:						(statements)	50.0%"
            exit 0
        fi
        ;;
esac
exit 0
GO
    chmod +x "$MOCK_BIN/go"

    cd "$FAKE_REPO"
    export PATH="$MOCK_BIN:$PATH"
    export CMD_AO_COVERAGE_FLOOR=40.0

    run bash "$FAKE_REPO/scripts/check-cmdao-coverage-floor.sh"
    [ "$status" -eq 0 ]
    [[ "$output" == *"PASS"* ]]
}
