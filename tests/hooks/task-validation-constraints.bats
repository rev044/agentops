#!/usr/bin/env bats

setup() {
    export REPO_ROOT="$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)"
    export HOOKS_DIR="$REPO_ROOT/hooks"
    export TEST_REPO
    TEST_REPO="$(mktemp -d)"
    mkdir -p "$TEST_REPO/.agents/constraints"
    git -C "$TEST_REPO" init -q >/dev/null 2>&1
}

teardown() {
    rm -rf "$TEST_REPO"
}

write_index() {
    local body="$1"
    printf '%s\n' "$body" > "$TEST_REPO/.agents/constraints/index.json"
}

run_hook() {
    local payload="$1"
    (cd "$TEST_REPO" && printf '%s' "$payload" | bash "$HOOKS_DIR/task-validation-gate.sh" 2>&1)
}

@test "task-validation active constraint: malformed index blocks" {
    printf '{not-json\n' > "$TEST_REPO/.agents/constraints/index.json"
    run run_hook '{"metadata":{"issue_type":"feature"}}'
    [ "$status" -eq 2 ]
    [[ "$output" == *"constraint index is unreadable"* ]]
}

@test "task-validation active constraint: unsupported detector kind blocks" {
    mkdir -p "$TEST_REPO/docs"
    printf 'hello\n' > "$TEST_REPO/docs/guide.md"
    write_index '{"schema_version":1,"constraints":[{"id":"c-unknown","title":"unknown","status":"active","compiled_at":"2026-03-10T00:00:00Z","applies_to":{"scope":"files","issue_types":["feature"],"path_globs":["docs/*.md"]},"detector":{"kind":"mystery"}}]}'
    run run_hook '{"metadata":{"issue_type":"feature","files":["docs/guide.md"]}}'
    [ "$status" -eq 2 ]
    [[ "$output" == *"unsupported active detector kind"* ]]
}

@test "task-validation active constraint: must_not_contain blocks forbidden literal" {
    mkdir -p "$TEST_REPO/docs"
    printf 'SECRET_TOKEN\n' > "$TEST_REPO/docs/guide.md"
    write_index '{"schema_version":1,"constraints":[{"id":"c-forbidden","title":"forbidden token","status":"active","compiled_at":"2026-03-10T00:00:00Z","applies_to":{"scope":"files","issue_types":["feature"],"path_globs":["docs/*.md"]},"detector":{"kind":"content_pattern","mode":"must_not_contain","pattern":"SECRET_TOKEN","message":"forbidden token present"}}]}'
    run run_hook '{"metadata":{"issue_type":"feature","files":["docs/guide.md"]}}'
    [ "$status" -eq 2 ]
    [[ "$output" == *"forbidden token present"* ]]
}

@test "task-validation active constraint: restricted_command obeys allowlist" {
    write_index '{"schema_version":1,"constraints":[{"id":"c-curl","title":"curl blocked","status":"active","compiled_at":"2026-03-10T00:00:00Z","applies_to":{"scope":"files","issue_types":["feature"]},"detector":{"kind":"restricted_command","command":"curl http://evil.test","message":"curl must stay blocked"}}]}'
    run run_hook '{"metadata":{"issue_type":"feature"}}'
    [ "$status" -eq 2 ]
    [[ "$output" == *"not in allowlist"* ]]
}

@test "task-validation active constraint: issue_type mismatch skips constraint" {
    mkdir -p "$TEST_REPO/docs"
    printf 'hello\n' > "$TEST_REPO/docs/guide.md"
    write_index '{"schema_version":1,"constraints":[{"id":"c-skip","title":"skip docs","status":"active","compiled_at":"2026-03-10T00:00:00Z","applies_to":{"scope":"files","issue_types":["feature"],"path_globs":["docs/*.md"]},"detector":{"kind":"content_pattern","mode":"must_contain","pattern":"SAFE_MARKER","message":"SAFE_MARKER required"}}]}'
    run run_hook '{"metadata":{"issue_type":"docs","files":["docs/guide.md"]}}'
    [ "$status" -eq 0 ]
}
