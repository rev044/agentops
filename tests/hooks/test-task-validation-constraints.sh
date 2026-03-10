#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
HOOKS_DIR="$REPO_ROOT/hooks"

PASS=0
FAIL=0
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

pass() {
    echo "PASS: $1"
    PASS=$((PASS + 1))
}

fail() {
    echo "FAIL: $1"
    FAIL=$((FAIL + 1))
}

setup_repo() {
    local dir="$1"
    mkdir -p "$dir/.agents/constraints"
    git -C "$dir" init -q >/dev/null 2>&1
}

write_index() {
    local dir="$1"
    local body="$2"
    printf '%s\n' "$body" > "$dir/.agents/constraints/index.json"
}

run_hook() {
    local dir="$1"
    local payload="$2"
    local output
    local ec=0
    output=$(cd "$dir" && printf '%s' "$payload" | bash "$HOOKS_DIR/task-validation-gate.sh" 2>&1) || ec=$?
    printf '%s\n' "$output"
    return "$ec"
}

# 1. draft constraints remain inert
REPO_DRAFT="$TMPDIR/draft"
setup_repo "$REPO_DRAFT"
mkdir -p "$REPO_DRAFT/docs"
printf 'hello\n' > "$REPO_DRAFT/docs/guide.md"
write_index "$REPO_DRAFT" '{"schema_version":1,"constraints":[{"id":"c-draft","title":"draft","status":"draft","compiled_at":"2026-03-10T00:00:00Z","applies_to":{"scope":"files","issue_types":["feature"],"path_globs":["docs/*.md"]},"detector":{"kind":"content_pattern","mode":"must_contain","pattern":"SAFE_MARKER","message":"SAFE_MARKER required"}}]}'
if run_hook "$REPO_DRAFT" '{"metadata":{"files":["docs/guide.md"]}}' >/dev/null; then
    pass "draft constraints remain inert"
else
    fail "draft constraints remain inert"
fi

# 2. active constraints require issue_type when scoped by issue_types
REPO_ISSUE_TYPE="$TMPDIR/issue-type"
setup_repo "$REPO_ISSUE_TYPE"
mkdir -p "$REPO_ISSUE_TYPE/docs"
printf 'SAFE_MARKER\n' > "$REPO_ISSUE_TYPE/docs/guide.md"
write_index "$REPO_ISSUE_TYPE" '{"schema_version":1,"constraints":[{"id":"c-issue-type","title":"issue type required","status":"active","compiled_at":"2026-03-10T00:00:00Z","applies_to":{"scope":"files","issue_types":["feature"],"path_globs":["docs/*.md"]},"detector":{"kind":"content_pattern","mode":"must_contain","pattern":"SAFE_MARKER","message":"SAFE_MARKER required"}}]}'
EC=0
OUTPUT=$(run_hook "$REPO_ISSUE_TYPE" '{"metadata":{"files":["docs/guide.md"]}}') || EC=$?
if [ "$EC" -eq 2 ] && echo "$OUTPUT" | grep -q 'metadata.issue_type'; then
    pass "active constraints require metadata.issue_type"
else
    fail "active constraints require metadata.issue_type"
fi

# 3. active content_pattern passes with matching literal
REPO_PATTERN_PASS="$TMPDIR/pattern-pass"
setup_repo "$REPO_PATTERN_PASS"
mkdir -p "$REPO_PATTERN_PASS/docs"
printf 'SAFE_MARKER\n' > "$REPO_PATTERN_PASS/docs/guide.md"
write_index "$REPO_PATTERN_PASS" '{"schema_version":1,"constraints":[{"id":"c-pattern-pass","title":"must contain","status":"active","compiled_at":"2026-03-10T00:00:00Z","applies_to":{"scope":"files","issue_types":["feature"],"path_globs":["docs/*.md"]},"detector":{"kind":"content_pattern","mode":"must_contain","pattern":"SAFE_MARKER","message":"SAFE_MARKER required"}}]}'
if run_hook "$REPO_PATTERN_PASS" '{"metadata":{"issue_type":"feature","files":["docs/guide.md"]}}' >/dev/null; then
    pass "active content_pattern passes when literal is present"
else
    fail "active content_pattern passes when literal is present"
fi

# 4. active content_pattern blocks missing literal
REPO_PATTERN_FAIL="$TMPDIR/pattern-fail"
setup_repo "$REPO_PATTERN_FAIL"
mkdir -p "$REPO_PATTERN_FAIL/docs"
printf 'hello\n' > "$REPO_PATTERN_FAIL/docs/guide.md"
write_index "$REPO_PATTERN_FAIL" '{"schema_version":1,"constraints":[{"id":"c-pattern-fail","title":"must contain","status":"active","compiled_at":"2026-03-10T00:00:00Z","applies_to":{"scope":"files","issue_types":["feature"],"path_globs":["docs/*.md"]},"detector":{"kind":"content_pattern","mode":"must_contain","pattern":"SAFE_MARKER","message":"SAFE_MARKER required"}}]}'
EC=0
OUTPUT=$(run_hook "$REPO_PATTERN_FAIL" '{"metadata":{"issue_type":"feature","files":["docs/guide.md"]}}') || EC=$?
if [ "$EC" -eq 2 ] && echo "$OUTPUT" | grep -q 'SAFE_MARKER required'; then
    pass "active content_pattern blocks missing literal"
else
    fail "active content_pattern blocks missing literal"
fi

# 5. active paired_files blocks missing companion
REPO_PAIRED="$TMPDIR/paired"
setup_repo "$REPO_PAIRED"
mkdir -p "$REPO_PAIRED/cli/cmd/ao"
printf 'package main\n' > "$REPO_PAIRED/cli/cmd/ao/foo.go"
write_index "$REPO_PAIRED" '{"schema_version":1,"constraints":[{"id":"c-paired","title":"paired files","status":"active","compiled_at":"2026-03-10T00:00:00Z","applies_to":{"scope":"files","issue_types":["feature"],"path_globs":["cli/cmd/ao/*.go"]},"detector":{"kind":"paired_files","pattern":"cli/cmd/ao/*.go","exclude":"*_test.go","companion":"{dir}/{basename}_test{ext}","message":"missing paired test change"}}]}'
EC=0
OUTPUT=$(run_hook "$REPO_PAIRED" '{"metadata":{"issue_type":"feature","files":["cli/cmd/ao/foo.go"]}}') || EC=$?
if [ "$EC" -eq 2 ] && echo "$OUTPUT" | grep -q 'missing paired test change'; then
    pass "active paired_files blocks missing companion"
else
    fail "active paired_files blocks missing companion"
fi

# 6. active restricted_command runs allowlisted command
REPO_COMMAND="$TMPDIR/command"
setup_repo "$REPO_COMMAND"
write_index "$REPO_COMMAND" '{"schema_version":1,"constraints":[{"id":"c-command","title":"go version","status":"active","compiled_at":"2026-03-10T00:00:00Z","applies_to":{"scope":"files","issue_types":["feature"]},"detector":{"kind":"restricted_command","command":"go version","message":"go version must succeed"}}]}'
if run_hook "$REPO_COMMAND" '{"metadata":{"issue_type":"feature"}}' >/dev/null; then
    pass "active restricted_command runs allowlisted command"
else
    fail "active restricted_command runs allowlisted command"
fi

echo
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
