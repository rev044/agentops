#!/usr/bin/env bats

setup() {
    REPO_ROOT="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"
    SCRIPT="$REPO_ROOT/scripts/check-evolve-cycle-logging.sh"

    TMP_DIR="$(mktemp -d)"
    FAKE_REPO="$TMP_DIR/repo"
    mkdir -p "$FAKE_REPO/scripts" "$FAKE_REPO/.agents/evolve"
    /bin/cp "$SCRIPT" "$FAKE_REPO/scripts/check-evolve-cycle-logging.sh"
    chmod +x "$FAKE_REPO/scripts/check-evolve-cycle-logging.sh"
}

teardown() {
    rm -rf "$TMP_DIR"
}

@test "check-evolve-cycle-logging.sh exists and is executable" {
    [ -f "$SCRIPT" ]
    [ -x "$SCRIPT" ]
}

@test "check-evolve-cycle-logging.sh accepts canonical rows and warns on legacy numeric drift" {
    cat > "$FAKE_REPO/.agents/evolve/cycle-history.jsonl" <<'EOF'
{"cycle":1,"target":"legacy-goal","result":"improved","sha":"abc1234","goals_passing":"17","goals_total":"18","timestamp":"2026-03-07T16:00:00Z"}
{"cycle":2,"target":"canonical-goal","result":"improved","sha":"def5678","canonical_sha":"def5678","log_sha":"feedcafe","goals_passing":17,"goals_total":18,"timestamp":"2026-03-07T16:05:00Z"}
EOF

    cd "$FAKE_REPO"
    run bash scripts/check-evolve-cycle-logging.sh

    [ "$status" -eq 0 ]
    [[ "$output" == *"OK with warnings"* ]]
    [[ "$output" == *"non-numeric goals_passing"* ]]
}

@test "check-evolve-cycle-logging.sh fails on invalid JSON rows" {
    cat > "$FAKE_REPO/.agents/evolve/cycle-history.jsonl" <<'EOF'
{"cycle":1,"target":"ok","result":"unchanged","timestamp":"2026-03-07T16:00:00Z"}
not json
EOF

    cd "$FAKE_REPO"
    run bash scripts/check-evolve-cycle-logging.sh

    [ "$status" -eq 1 ]
    [[ "$output" == *"not valid JSON"* ]]
}
