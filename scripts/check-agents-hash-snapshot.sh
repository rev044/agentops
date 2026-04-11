#!/usr/bin/env bash
set -euo pipefail

# check-agents-hash-snapshot.sh — content-hash gate for ~/.agents/ mutation
# during test runs.
#
# Complements scripts/check-home-isolation.sh which only checks for the
# t.Setenv("HOME", ...) pattern statically via grep. This gate catches any
# mutation of the protected ~/.agents/ subtrees, including:
#   - Tests that write to ~/.agents/ without HOME isolation
#   - Tests that write then use os.Chtimes to hide the mtime (the
#     mtime-bypass attack; demonstrated in cli/internal/overnight/recovery_test.go:285
#     for legitimate recovery fixture setup, but the same API can hide
#     test-side poisoning)
#
# Usage:
#   SNAP=$(./scripts/check-agents-hash-snapshot.sh capture)
#   go test ./...
#   ./scripts/check-agents-hash-snapshot.sh diff "$SNAP"
#
# Scoped to the subtrees Dream and flywheel tests are most likely to
# corrupt: learnings, patterns, findings, rpi/next-work.jsonl.

# shasum presence check — skip gracefully on minimal containers
if ! command -v shasum >/dev/null 2>&1; then
    echo "WARN: shasum not available; ~/.agents hash gate skipped" >&2
    exit 0
fi

AGENTS_HUB="${AGENTS_HUB_OVERRIDE:-$HOME/.agents}"
SCOPES=(learnings patterns findings rpi/next-work.jsonl)

snapshot() {
    for scope in "${SCOPES[@]}"; do
        local target="$AGENTS_HUB/$scope"
        if [[ -f "$target" ]]; then
            shasum -a 256 "$target" 2>/dev/null || true
        elif [[ -d "$target" ]]; then
            find "$target" -type f -print0 2>/dev/null \
                | xargs -0 shasum -a 256 2>/dev/null \
                | sort
        fi
    done
}

cmd="${1:-}"
case "$cmd" in
    capture)
        out="${AGENTOPS_HASH_SNAPSHOT:-$(mktemp -t agents-snap.XXXXXX)}"
        snapshot > "$out"
        echo "$out"
        ;;
    diff)
        snap_file="${2:-}"
        if [[ -z "$snap_file" || ! -f "$snap_file" ]]; then
            echo "ERROR: diff requires a valid snapshot file path" >&2
            exit 2
        fi
        current=$(mktemp -t agents-snap.XXXXXX)
        snapshot > "$current"
        if ! diff -q "$snap_file" "$current" >/dev/null 2>&1; then
            echo "FAIL: ~/.agents mutation detected during tests" >&2
            echo "--- snapshot before (path: $snap_file)" >&2
            echo "+++ snapshot after" >&2
            diff "$snap_file" "$current" | head -40 >&2
            rm -f "$current"
            exit 1
        fi
        rm -f "$current"
        ;;
    *)
        echo "usage: $0 {capture|diff <snapshot-file>}" >&2
        exit 2
        ;;
esac
