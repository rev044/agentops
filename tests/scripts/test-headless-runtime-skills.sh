#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/scripts/validate-headless-runtime-skills.sh"

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL + 1)); }

if [[ ! -f "$SCRIPT" ]]; then
    echo "FAIL: missing script: $SCRIPT" >&2
    exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

make_fixture() {
    local root="$1"

    mkdir -p \
        "$root/skills/athena" \
        "$root/skills/research" \
        "$root/skills-codex/athena" \
        "$root/skills-codex/research"

    cat > "$root/skills/athena/SKILL.md" <<'EOF'
---
name: athena
description: >
  Active knowledge intelligence. Runs Mine → Grow → Defrag cycle.
skill_api_version: 1
---
EOF

    cat > "$root/skills/research/SKILL.md" <<'EOF'
---
name: research
description: 'Deep codebase exploration.'
skill_api_version: 1
---
EOF

    cat > "$root/skills-codex/athena/SKILL.md" <<'EOF'
---
name: athena
description: 'Active knowledge intelligence. Runs Mine → Grow → Defrag cycle.'
skill_api_version: 1
---
EOF

    cat > "$root/skills-codex/research/SKILL.md" <<'EOF'
---
name: research
description: 'Deep codebase exploration.'
skill_api_version: 1
---
EOF
}

make_mock_claude() {
    local bin_dir="$1"
    local mode="${2:-pass}"
    cat > "$bin_dir/claude" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

prompt=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    -p)
      prompt="${2:-}"
      shift 2
      ;;
    *)
      if [[ "$1" == "--help" ]]; then
        echo "Claude help"
        exit 0
      fi
      shift
      ;;
  esac
done

mode="__MODE__"
if [[ "$mode" == "fallback" ]]; then
  exit 124
fi

if [[ "$prompt" == *"compact JSON array of skill names"* ]]; then
  if [[ "$mode" == "malformed" ]]; then
    python3 - <<'PY'
import json

for payload in (
    {"type": "assistant", "message": {"content": [{"type": "text", "text": 'not-json'}]}},
    {"type": "result"},
):
    print(json.dumps(payload))
PY
    exit 0
  fi
  python3 - <<'PY'
import json

for payload in (
    {"type": "assistant", "message": {"content": [{"type": "text", "text": '["athena","research"]'}]}},
    {"type": "result"},
):
    print(json.dumps(payload))
PY
  exit 0
fi

echo "unexpected Claude prompt: $prompt" >&2
exit 1
EOF
    python3 - <<'PY' "$bin_dir/claude" "$mode"
from pathlib import Path
import sys

path = Path(sys.argv[1])
mode = sys.argv[2]
path.write_text(path.read_text().replace("__MODE__", mode))
PY
    chmod +x "$bin_dir/claude"
}

make_mock_codex() {
    local bin_dir="$1"
    local mode="${2:-pass}"
    cat > "$bin_dir/codex" <<EOF
#!/usr/bin/env bash
set -euo pipefail

if [[ "\$1" != "exec" ]]; then
  echo "unexpected codex command: \$*" >&2
  exit 1
fi

python3 - <<'PY'
import json

mode = ${mode@Q}
if mode == "missing":
    text = '[{"name":"research","description":"Deep codebase exploration."}]'
else:
    text = '[{"name":"athena","description":"Active knowledge intelligence. Runs Mine → Grow → Defrag cycle."},{"name":"research","description":"Deep codebase exploration."}]'

for payload in (
    {"type": "thread.started", "thread_id": "fixture"},
    {"type": "turn.started"},
    {"type": "item.completed", "item": {"id": "item_0", "type": "agent_message", "text": text}},
    {"type": "turn.completed"},
):
    print(json.dumps(payload))
PY
EOF
    chmod +x "$bin_dir/codex"
}

test_passes_with_mocked_runtimes() {
    local repo="$TMP_DIR/pass-repo"
    local bin_dir="$TMP_DIR/pass-bin"
    mkdir -p "$bin_dir"
    make_fixture "$repo"
    make_mock_claude "$bin_dir"
    make_mock_codex "$bin_dir" pass

    if PATH="$bin_dir:$PATH" bash "$SCRIPT" --repo-root "$repo" --workdir "$TMP_DIR/workdir-pass" >"$TMP_DIR/pass.log" 2>&1; then
        pass "passes with mocked Claude and Codex inventories"
    else
        fail "passes with mocked Claude and Codex inventories"
        sed -n '1,80p' "$TMP_DIR/pass.log" >&2
    fi
}

test_fails_when_codex_inventory_is_missing_skill() {
    local repo="$TMP_DIR/fail-repo"
    local bin_dir="$TMP_DIR/fail-bin"
    mkdir -p "$bin_dir"
    make_fixture "$repo"
    make_mock_claude "$bin_dir"
    make_mock_codex "$bin_dir" missing

    if PATH="$bin_dir:$PATH" bash "$SCRIPT" --repo-root "$repo" --runtime codex --workdir "$TMP_DIR/workdir-fail" >"$TMP_DIR/fail.log" 2>&1; then
        fail "fails when Codex inventory is missing a skill"
    elif rg -q 'missing skills: athena' "$TMP_DIR/fail.log"; then
        pass "fails when Codex inventory is missing a skill"
    else
        fail "fails when Codex inventory is missing a skill"
        sed -n '1,80p' "$TMP_DIR/fail.log" >&2
    fi
}

test_warns_and_passes_when_claude_inventory_falls_back_to_help() {
    local repo="$TMP_DIR/fallback-repo"
    local bin_dir="$TMP_DIR/fallback-bin"
    mkdir -p "$bin_dir"
    make_fixture "$repo"
    make_mock_claude "$bin_dir" fallback

    if PATH="$bin_dir:$PATH" bash "$SCRIPT" --repo-root "$repo" --runtime claude --workdir "$TMP_DIR/workdir-fallback" >"$TMP_DIR/fallback.log" 2>&1; then
        if rg -q 'Claude load-check fallback succeeded; deep inventory not verified' "$TMP_DIR/fallback.log"; then
            pass "warns and passes when Claude inventory falls back to help"
        else
            fail "warns and passes when Claude inventory falls back to help"
            sed -n '1,80p' "$TMP_DIR/fallback.log" >&2
        fi
    else
        fail "warns and passes when Claude inventory falls back to help"
        sed -n '1,80p' "$TMP_DIR/fallback.log" >&2
    fi
}

test_warns_and_passes_when_claude_output_is_malformed() {
    local repo="$TMP_DIR/malformed-repo"
    local bin_dir="$TMP_DIR/malformed-bin"
    mkdir -p "$bin_dir"
    make_fixture "$repo"
    make_mock_claude "$bin_dir" malformed

    if PATH="$bin_dir:$PATH" bash "$SCRIPT" --repo-root "$repo" --runtime claude --workdir "$TMP_DIR/workdir-malformed" >"$TMP_DIR/malformed.log" 2>&1; then
        if rg -q 'Claude assistant output was not a JSON array' "$TMP_DIR/malformed.log" && \
            rg -q 'Claude load-check fallback succeeded; deep inventory not verified' "$TMP_DIR/malformed.log"; then
            pass "warns and passes when Claude output is malformed"
        else
            fail "warns and passes when Claude output is malformed"
            sed -n '1,80p' "$TMP_DIR/malformed.log" >&2
        fi
    else
        fail "warns and passes when Claude output is malformed"
        sed -n '1,80p' "$TMP_DIR/malformed.log" >&2
    fi
}

test_strict_mode_fails_when_claude_falls_back() {
    local repo="$TMP_DIR/strict-repo"
    local bin_dir="$TMP_DIR/strict-bin"
    mkdir -p "$bin_dir"
    make_fixture "$repo"
    make_mock_claude "$bin_dir" fallback

    if PATH="$bin_dir:$PATH" HEADLESS_RUNTIME_SKILL_CLAUDE_STRICT=1 bash "$SCRIPT" --repo-root "$repo" --runtime claude --workdir "$TMP_DIR/workdir-strict" >"$TMP_DIR/strict.log" 2>&1; then
        fail "strict mode fails when Claude falls back"
    elif rg -q 'Claude load-check fallback succeeded; deep inventory not verified' "$TMP_DIR/strict.log"; then
        pass "strict mode fails when Claude falls back"
    else
        fail "strict mode fails when Claude falls back"
        sed -n '1,80p' "$TMP_DIR/strict.log" >&2
    fi
}

echo "== test-headless-runtime-skills =="
test_passes_with_mocked_runtimes
test_fails_when_codex_inventory_is_missing_skill
test_warns_and_passes_when_claude_inventory_falls_back_to_help
test_warns_and_passes_when_claude_output_is_malformed
test_strict_mode_fails_when_claude_falls_back

echo ""
echo "Results: $PASS PASS, $FAIL FAIL"
if [[ "$FAIL" -gt 0 ]]; then
    exit 1
fi
exit 0
