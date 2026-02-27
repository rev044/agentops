#!/usr/bin/env bash
# test-gate.sh — Tests for scripts/spec-consistency-gate.sh
#
# Exercises each failure mode using temporary fixture directories.
# Uses the same pass/fail accumulator pattern as other test scripts.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
GATE_SCRIPT="$REPO_ROOT/scripts/spec-consistency-gate.sh"
FIXTURES_DIR="$SCRIPT_DIR/fixtures"

PASS_COUNT=0
FAIL_COUNT=0
TMP_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

pass() {
  echo "PASS: $1"
  PASS_COUNT=$((PASS_COUNT + 1))
}

fail() {
  echo "FAIL: $1"
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

# --- Helpers ---

run_gate() {
  local contracts_dir="$1"
  local output_file="$2"
  set +e
  "$GATE_SCRIPT" "$contracts_dir" >"$output_file" 2>&1
  local status=$?
  set -e
  return "$status"
}

assert_passes() {
  local description="$1"
  local contracts_dir="$2"
  local output_file
  output_file="$(mktemp "$TMP_DIR/out-XXXXXX")"

  if run_gate "$contracts_dir" "$output_file"; then
    pass "$description"
  else
    echo "--- gate output ($description) ---"
    cat "$output_file"
    echo "---"
    fail "$description (expected PASS, got FAIL)"
  fi
}

assert_fails_with() {
  local description="$1"
  local contracts_dir="$2"
  local expected_text="$3"
  local output_file
  output_file="$(mktemp "$TMP_DIR/out-XXXXXX")"

  if run_gate "$contracts_dir" "$output_file"; then
    echo "--- gate output ($description) ---"
    cat "$output_file"
    echo "---"
    fail "$description (expected FAIL, got PASS)"
    return
  fi

  if grep -Fq "$expected_text" "$output_file"; then
    pass "$description"
  else
    echo "--- gate output ($description) ---"
    cat "$output_file"
    echo "---"
    fail "$description (missing expected text: '$expected_text')"
  fi
}

# --- Fixture builder ---
# Creates a minimal valid contract in a temp dir
make_valid_contract() {
  local dir="$1"
  local filename="${2:-contract-valid.md}"
  local issue="${3:-ag-test.1}"

  mkdir -p "$dir"
  cat >"$dir/$filename" <<EOF
# Contract: Valid

\`\`\`yaml
issue:      $issue
framework:  go
category:   feature
\`\`\`

## Problem

Test fixture for a valid contract.

## Invariants

1. First invariant always holds.
2. Second invariant always holds.
3. Third invariant always holds.

## Test Cases

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | happy path input | Returns success result | #1 |
| 2 | error path input | Returns error response | #2 |
| 3 | boundary input | Returns expected boundary | #3 |
EOF
}

# ============================================================================
echo "================================"
echo "Testing spec-consistency-gate.sh"
echo "================================"
echo ""

# --- Test: script is executable ---
if [[ -x "$GATE_SCRIPT" ]]; then
  pass "spec-consistency-gate.sh is executable"
else
  fail "spec-consistency-gate.sh is not executable (run: chmod +x $GATE_SCRIPT)"
fi

# --- Test: missing directory exits 0 (graceful skip) ---
missing_dir="$TMP_DIR/nonexistent"
output_file="$(mktemp "$TMP_DIR/out-XXXXXX")"
set +e
"$GATE_SCRIPT" "$missing_dir" >"$output_file" 2>&1
missing_exit=$?
set -e
if [[ "$missing_exit" -eq 0 ]]; then
  pass "missing directory: exits 0 (graceful skip)"
else
  echo "--- output ---"
  cat "$output_file"
  echo "---"
  fail "missing directory: expected exit 0, got $missing_exit"
fi

# --- Test: empty directory exits 0 (graceful skip) ---
empty_dir="$TMP_DIR/empty"
mkdir -p "$empty_dir"
output_file="$(mktemp "$TMP_DIR/out-XXXXXX")"
set +e
"$GATE_SCRIPT" "$empty_dir" >"$output_file" 2>&1
empty_exit=$?
set -e
if [[ "$empty_exit" -eq 0 ]]; then
  pass "empty directory: exits 0 (graceful skip)"
else
  echo "--- output ---"
  cat "$output_file"
  echo "---"
  fail "empty directory: expected exit 0, got $empty_exit"
fi

# --- Test: valid contract passes ---
valid_dir="$TMP_DIR/valid"
make_valid_contract "$valid_dir" "contract-valid.md" "ag-valid.1"
assert_passes "valid contract: gate passes" "$valid_dir"

# --- Test: check 1 — missing frontmatter block fails ---
no_fm_dir="$TMP_DIR/no-frontmatter"
mkdir -p "$no_fm_dir"
cat >"$no_fm_dir/contract-no-fm.md" <<'EOF'
# Contract: No Frontmatter

## Problem

No yaml block present.

## Invariants

1. First.
2. Second.
3. Third.

## Test Cases

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | input | success | #1 |
| 2 | input | error | #2 |
| 3 | input | success | #3 |
EOF
assert_fails_with \
  "check 1: missing frontmatter block fails" \
  "$no_fm_dir" \
  "no \`\`\`yaml frontmatter block found"

# --- Test: check 1 — missing 'issue' field fails ---
no_issue_dir="$TMP_DIR/no-issue"
mkdir -p "$no_issue_dir"
cat >"$no_issue_dir/contract-no-issue.md" <<'EOF'
# Contract: Missing Issue

```yaml
framework:  go
category:   feature
```

## Invariants

1. First.
2. Second.
3. Third.

## Test Cases

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | input | success | #1 |
| 2 | input | error | #2 |
| 3 | input | success | #3 |
EOF
assert_fails_with \
  "check 1: missing 'issue' field fails" \
  "$no_issue_dir" \
  "frontmatter field 'issue' is missing or empty"

# --- Test: check 2 — fewer than 3 invariants fails ---
few_inv_dir="$TMP_DIR/few-invariants"
mkdir -p "$few_inv_dir"
cat >"$few_inv_dir/contract-few-inv.md" <<'EOF'
```yaml
issue:      ag-test.2
framework:  shell
category:   ci
```

## Invariants

1. Only invariant.
2. Second invariant.

## Test Cases

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | input | success | #1 |
| 2 | input | error | #2 |
| 3 | input | success | #1 |
EOF
assert_fails_with \
  "check 2: fewer than 3 invariants fails" \
  "$few_inv_dir" \
  "## Invariants has 2 item(s) (need >=3)"

# --- Test: check 2 — fewer than 3 test rows fails ---
few_rows_dir="$TMP_DIR/few-rows"
mkdir -p "$few_rows_dir"
cat >"$few_rows_dir/contract-few-rows.md" <<'EOF'
```yaml
issue:      ag-test.3
framework:  python
category:   feature
```

## Invariants

1. First.
2. Second.
3. Third.

## Test Cases

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | input | success | #1 |
| 2 | input | error | #2 |
EOF
assert_fails_with \
  "check 2: fewer than 3 test rows fails" \
  "$few_rows_dir" \
  "## Test Cases has 2 row(s) (need >=3)"

# --- Test: check 3 — duplicate issue across contracts fails ---
dup_dir="$TMP_DIR/dup-issue"
make_valid_contract "$dup_dir" "contract-alpha.md" "ag-dup.1"
make_valid_contract "$dup_dir" "contract-beta.md" "ag-dup.1"
assert_fails_with \
  "check 3: duplicate issue ID across contracts fails" \
  "$dup_dir" \
  "is referenced by 2 contracts"

# --- Test: check 4 — no error-path test row warns (not fails) ---
no_error_path_dir="$TMP_DIR/no-error-path"
mkdir -p "$no_error_path_dir"
cat >"$no_error_path_dir/contract-no-error.md" <<'EOF'
```yaml
issue:      ag-test.4
framework:  go
category:   feature
```

## Invariants

1. First.
2. Second.
3. Third.

## Test Cases

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | valid input A | Returns success | #1 |
| 2 | valid input B | Returns success | #2 |
| 3 | boundary input | Returns success | #3 |
EOF
output_file="$(mktemp "$TMP_DIR/out-XXXXXX")"
set +e
"$GATE_SCRIPT" "$no_error_path_dir" >"$output_file" 2>&1
no_error_exit=$?
set -e
if [[ "$no_error_exit" -eq 0 ]]; then
  if grep -q "no error-path" "$output_file"; then
    pass "check 4: no error-path rows produces WARN (not FAIL), exits 0"
  else
    fail "check 4: expected WARN about no error-path rows"
  fi
else
  echo "--- output ---"
  cat "$output_file"
  echo "---"
  fail "check 4: no error-path rows should WARN (exit 0), not FAIL (exit 1)"
fi

# --- Test: check 6 — placeholder text produces WARN (not FAIL) ---
placeholder_dir="$TMP_DIR/placeholder"
mkdir -p "$placeholder_dir"
cat >"$placeholder_dir/contract-tbd.md" <<'EOF'
```yaml
issue:      ag-test.5
framework:  go
category:   chore
```

## Problem

TODO: fill this in.

## Invariants

1. First.
2. Second.
3. TBD third.

## Test Cases

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | input-a | success | #1 |
| 2 | input-b | error response | #2 |
| 3 | input-c | success | #3 |
EOF
output_file="$(mktemp "$TMP_DIR/out-XXXXXX")"
set +e
"$GATE_SCRIPT" "$placeholder_dir" >"$output_file" 2>&1
placeholder_exit=$?
set -e
if [[ "$placeholder_exit" -eq 0 ]]; then
  if grep -q "placeholder text found" "$output_file"; then
    pass "check 6: placeholder text produces WARN (not FAIL), exits 0"
  else
    fail "check 6: expected WARN about placeholder text"
  fi
else
  echo "--- output ---"
  cat "$output_file"
  echo "---"
  fail "check 6: placeholder text should WARN (exit 0), not FAIL (exit 1)"
fi

# --- Test: fixture files from fixtures/ dir ---
for fixture in "$FIXTURES_DIR"/valid-*.md; do
  [[ -f "$fixture" ]] || continue
  fixture_name="$(basename "$fixture")"
  single_dir="$TMP_DIR/fixture-valid-$fixture_name"
  mkdir -p "$single_dir"
  /bin/cp "$fixture" "$single_dir/contract-${fixture_name}"
  assert_passes "fixture $fixture_name: valid contract passes" "$single_dir"
done

echo ""
echo "================================"
echo "Results: $PASS_COUNT PASS, $FAIL_COUNT FAIL"
echo "================================"

if [[ $FAIL_COUNT -gt 0 ]]; then
  exit 1
fi
exit 0
