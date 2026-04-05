#!/usr/bin/env bash
# Regression test: close-before-commit grace window
# Verifies that a bead closed BEFORE its qualifying commit still passes
# when the commit lands within the 24h grace window.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
AUDIT_SCRIPT="$REPO_ROOT/skills/post-mortem/scripts/closure-integrity-audit.sh"
WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/closure-grace-XXXXXX")"
PASS=0
FAIL=0

cleanup() { rm -rf "$WORK_DIR"; }
trap cleanup EXIT

pass() { PASS=$((PASS + 1)); echo "PASS: $1"; }
fail() { FAIL=$((FAIL + 1)); echo "FAIL: $1"; }

# Setup: isolated git repo with bd mock
BD_DIR="$WORK_DIR/bd-data"
BIN_DIR="$WORK_DIR/bin"
REPO_DIR="$WORK_DIR/repo"
mkdir -p "$BD_DIR" "$BIN_DIR" "$REPO_DIR"

# Create a mock bd that reads from JSON files
cat > "$BIN_DIR/bd" <<'MOCK'
#!/usr/bin/env bash
case "$1" in
  children)
    cat "$BD_DIR/children.json" 2>/dev/null || echo '[]'
    ;;
  show)
    if [[ "${3:-}" == "--json" ]]; then
      cat "$BD_DIR/show-${2}.json" 2>/dev/null || echo '[]'
    else
      cat "$BD_DIR/show-${2}.txt" 2>/dev/null || echo "NOT FOUND"
    fi
    ;;
esac
MOCK
chmod +x "$BIN_DIR/bd"
sed -i.bak "s|\$BD_DIR|$BD_DIR|g" "$BIN_DIR/bd" 2>/dev/null || \
  sed -i '' "s|\$BD_DIR|$BD_DIR|g" "$BIN_DIR/bd"

export PATH="$BIN_DIR:$PATH"
export BD_DIR

# Initialize isolated repo
(
  cd "$REPO_DIR"
  git init -q
  git config user.email "test@test.com"
  git config user.name "Test"
  echo "init" > README.md
  git add README.md
  git commit -q -m "init"
)

# Create a file that the issue scopes to
SCOPED_FILE="cli/cmd/ao/fix.go"
mkdir -p "$REPO_DIR/cli/cmd/ao"

# Scenario: bead closed at T, qualifying commit lands at T+2h (within grace)
CLOSE_TIME="2026-03-20T10:00:00+00:00"
COMMIT_TIME="2026-03-20T12:00:00+00:00"

# Setup bd mock data for test-epic
cat > "$BD_DIR/children.json" <<JSON
[{"id": "test-epic.1"}]
JSON

cat > "$BD_DIR/show-test-epic.1.json" <<JSON
[{
  "id": "test-epic.1",
  "status": "closed",
  "created_at": "2026-03-19T10:00:00+00:00",
  "closed_at": "$CLOSE_TIME",
  "description": "Fix the handler logic.\n\nFiles:\n- \`$SCOPED_FILE\`"
}]
JSON

# Create qualifying commit AFTER close time
(
  cd "$REPO_DIR"
  echo "package ao" > "$SCOPED_FILE"
  git add "$SCOPED_FILE"
  GIT_AUTHOR_DATE="$COMMIT_TIME" GIT_COMMITTER_DATE="$COMMIT_TIME" \
    git commit -q -m "fix: handler logic"
)

# Test 1: Without grace, this would fail (commit is after closed_at)
# With grace, it should pass
result="$(cd "$REPO_DIR" && bash "$AUDIT_SCRIPT" --scope commit test-epic 2>&1)"
verdict="$(echo "$result" | jq -r '.children[0].status')"
detail="$(echo "$result" | jq -r '.children[0].detail')"

if [[ "$verdict" == "pass" ]] && [[ "$detail" == *"grace window"* ]]; then
  pass "close-before-commit detected via grace window"
else
  fail "close-before-commit should pass via grace window (got status=$verdict detail=$detail)"
fi

# Test 2: Commit way outside grace window (T+48h) should fail
(
  cd "$REPO_DIR"
  git reset --hard HEAD~1 -q
  mkdir -p "$(dirname "$SCOPED_FILE")"
  LATE_TIME="2026-03-22T10:00:00+00:00"
  echo "package ao" > "$SCOPED_FILE"
  git add "$SCOPED_FILE"
  GIT_AUTHOR_DATE="$LATE_TIME" GIT_COMMITTER_DATE="$LATE_TIME" \
    git commit -q -m "fix: late handler logic"
)

result="$(cd "$REPO_DIR" && bash "$AUDIT_SCRIPT" --scope commit test-epic 2>&1)"
verdict="$(echo "$result" | jq -r '.children[0].status')"
ftype="$(echo "$result" | jq -r '.failures[0].failure_type')"

if [[ "$verdict" == "fail" ]] && [[ "$ftype" == "timing_miss" ]]; then
  pass "commit outside grace window correctly classified as timing_miss"
else
  fail "commit outside grace should be timing_miss (got status=$verdict failure_type=$ftype)"
fi

# Test 3: Issue with no scoped files should be parser_miss
cat > "$BD_DIR/show-test-epic.1.json" <<JSON
[{
  "id": "test-epic.1",
  "status": "closed",
  "created_at": "2026-03-19T10:00:00+00:00",
  "closed_at": "$CLOSE_TIME",
  "description": "Fix the handler logic without specifying files."
}]
JSON

result="$(cd "$REPO_DIR" && bash "$AUDIT_SCRIPT" --scope commit test-epic 2>&1)"
ftype="$(echo "$result" | jq -r '.failures[0].failure_type')"

if [[ "$ftype" == "parser_miss" ]]; then
  pass "missing scoped files correctly classified as parser_miss"
else
  fail "missing scoped files should be parser_miss (got $ftype)"
fi

# Test 4: Bead with no scoped files AND no evidence-only packet should FAIL
cat > "$BD_DIR/children.json" <<JSON
[{"id": "test-epic.2"}]
JSON

cat > "$BD_DIR/show-test-epic.2.json" <<JSON
[{
  "id": "test-epic.2",
  "status": "closed",
  "created_at": "2026-03-19T10:00:00+00:00",
  "closed_at": "$CLOSE_TIME",
  "description": "Refactored internal logic with no specific files mentioned."
}]
JSON

# Ensure no evidence-only packet exists
rm -rf "$REPO_DIR/.agents/releases/evidence-only-closures" "$REPO_DIR/.agents/council/evidence-only-closures"

result="$(cd "$REPO_DIR" && bash "$AUDIT_SCRIPT" --scope auto test-epic 2>&1)"
verdict="$(echo "$result" | jq -r '.children[0].status')"
ftype="$(echo "$result" | jq -r '.failures[0].failure_type')"

if [[ "$verdict" == "fail" ]] && [[ "$ftype" == "parser_miss" ]]; then
  pass "no scoped files + no evidence-only packet correctly fails as parser_miss"
else
  fail "no scoped files + no evidence-only packet should be parser_miss (got status=$verdict failure_type=$ftype)"
fi

# Test 5: Bead with evidence-only packet but invalid schema should WARN (pass with packet mode but packet_is_valid rejects it)
cat > "$BD_DIR/children.json" <<JSON
[{"id": "test-epic.3"}]
JSON

cat > "$BD_DIR/show-test-epic.3.json" <<JSON
[{
  "id": "test-epic.3",
  "status": "closed",
  "created_at": "2026-03-19T10:00:00+00:00",
  "closed_at": "$CLOSE_TIME",
  "description": "Policy-only closure with no code delta."
}]
JSON

# Create an invalid evidence-only packet (missing required fields)
mkdir -p "$REPO_DIR/.agents/releases/evidence-only-closures"
cat > "$REPO_DIR/.agents/releases/evidence-only-closures/test-epic.3.json" <<JSON
{
  "target_id": "test-epic.3",
  "evidence_mode": "invalid_mode",
  "evidence": {"artifacts": []}
}
JSON

result="$(cd "$REPO_DIR" && bash "$AUDIT_SCRIPT" --scope auto test-epic 2>&1)"
verdict="$(echo "$result" | jq -r '.children[0].status')"
ftype="$(echo "$result" | jq -r '.failures[0].failure_type')"

if [[ "$verdict" == "fail" ]] && [[ "$ftype" == "parser_miss" ]]; then
  pass "invalid evidence-only packet correctly falls through to parser_miss"
else
  fail "invalid evidence-only packet should fall through to parser_miss (got status=$verdict failure_type=$ftype)"
fi

# Test 6: Bead with expired grace window should FAIL
cat > "$BD_DIR/children.json" <<JSON
[{"id": "test-epic.1"}]
JSON

EXPIRED_CLOSE="2026-03-15T10:00:00+00:00"
cat > "$BD_DIR/show-test-epic.1.json" <<JSON
[{
  "id": "test-epic.1",
  "status": "closed",
  "created_at": "2026-03-10T10:00:00+00:00",
  "closed_at": "$EXPIRED_CLOSE",
  "description": "Fix the handler logic.\n\nFiles:\n- \`$SCOPED_FILE\`"
}]
JSON

# Reset repo - commit is at 2026-03-20T12:00:00, close was 2026-03-15 (5 days before commit, well outside 24h grace)
(
  cd "$REPO_DIR"
  # Remove any evidence-only packets
  rm -rf .agents
  git reset --hard HEAD~1 -q 2>/dev/null || true
  mkdir -p "$(dirname "$SCOPED_FILE")"
  LATE_TIME="2026-03-20T12:00:00+00:00"
  echo "package ao" > "$SCOPED_FILE"
  git add "$SCOPED_FILE"
  GIT_AUTHOR_DATE="$LATE_TIME" GIT_COMMITTER_DATE="$LATE_TIME" \
    git commit -q -m "fix: handler logic"
)

result="$(cd "$REPO_DIR" && bash "$AUDIT_SCRIPT" --scope commit test-epic 2>&1)"
verdict="$(echo "$result" | jq -r '.children[0].status')"
ftype="$(echo "$result" | jq -r '.failures[0].failure_type')"

if [[ "$verdict" == "fail" ]] && [[ "$ftype" == "timing_miss" ]]; then
  pass "expired grace window correctly classified as timing_miss"
else
  fail "expired grace window should be timing_miss (got status=$verdict failure_type=$ftype)"
fi

echo ""
echo "Results: $PASS passed, $FAIL failed"
[[ "$FAIL" -eq 0 ]]
