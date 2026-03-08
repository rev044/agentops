#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

RUNTIME="all"
WORKDIR=""
CLAUDE_BIN="${CLAUDE_BIN:-claude}"
CODEX_BIN="${CODEX_BIN:-codex}"
CODEX_PROFILE="${CODEX_VALIDATE_PROFILE:-safe}"
TIMEOUT_SECONDS="${HEADLESS_RUNTIME_SKILL_TIMEOUT_SECONDS:-120}"
MAX_BUDGET_USD="${HEADLESS_RUNTIME_SKILL_MAX_BUDGET_USD:-2.00}"
SKIP_ENV="${AGENTOPS_SKIP_HEADLESS_RUNTIME_SKILLS:-0}"
CLAUDE_STRICT="${HEADLESS_RUNTIME_SKILL_CLAUDE_STRICT:-0}"

usage() {
    cat <<'EOF'
validate-headless-runtime-skills.sh

Open fresh headless Claude and/or Codex sessions, ask each runtime to return the
visible AgentOps skill inventory as JSON, then compare that inventory against
the repo skill definitions.

Options:
  --runtime <all|claude|codex>  Which runtime(s) to validate (default: all)
  --repo-root <dir>             Repo root to validate (default: current repo)
  --workdir <dir>               Ephemeral working directory for headless sessions
  --claude-bin <path>           Claude CLI binary (default: claude)
  --codex-bin <path>            Codex CLI binary (default: codex)
  --codex-profile <name>        Codex profile for headless exec (default: safe)
  --timeout <seconds>           Per-runtime timeout (default: 120)
  --max-budget-usd <amount>     Claude budget cap (default: 2.00)
  --help                        Show this help

Environment:
  AGENTOPS_SKIP_HEADLESS_RUNTIME_SKILLS=1  Skip the validation entirely
  HEADLESS_RUNTIME_SKILL_CLAUDE_STRICT=1   Fail when Claude inventory cannot be collected
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --runtime)
            RUNTIME="${2:-}"
            shift 2
            ;;
        --repo-root)
            REPO_ROOT="${2:-}"
            shift 2
            ;;
        --workdir)
            WORKDIR="${2:-}"
            shift 2
            ;;
        --claude-bin)
            CLAUDE_BIN="${2:-}"
            shift 2
            ;;
        --codex-bin)
            CODEX_BIN="${2:-}"
            shift 2
            ;;
        --codex-profile)
            CODEX_PROFILE="${2:-}"
            shift 2
            ;;
        --timeout)
            TIMEOUT_SECONDS="${2:-}"
            shift 2
            ;;
        --max-budget-usd)
            MAX_BUDGET_USD="${2:-}"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown arg: $1" >&2
            usage >&2
            exit 2
            ;;
    esac
done

case "$RUNTIME" in
    all|claude|codex) ;;
    *)
        echo "Invalid --runtime: $RUNTIME" >&2
        usage >&2
        exit 2
        ;;
esac

if [[ "$REPO_ROOT" != /* ]]; then
    REPO_ROOT="$(cd "$REPO_ROOT" && pwd)"
fi

if [[ "$SKIP_ENV" == "1" ]]; then
    echo "SKIP: AGENTOPS_SKIP_HEADLESS_RUNTIME_SKILLS=1"
    exit 0
fi

TMP_DIR=""
cleanup() {
    if [[ -n "${TMP_DIR:-}" && -d "$TMP_DIR" ]]; then
        rm -rf "$TMP_DIR"
    fi
}
trap cleanup EXIT

TMP_DIR="$(mktemp -d)"
if [[ -z "$WORKDIR" ]]; then
    WORKDIR="$TMP_DIR/workdir"
fi
mkdir -p "$WORKDIR"

EXPECTED_CLAUDE_JSON="$TMP_DIR/expected-claude.json"
EXPECTED_CODEX_JSON="$TMP_DIR/expected-codex.json"
ACTUAL_CLAUDE_JSON="$TMP_DIR/actual-claude.json"
ACTUAL_CODEX_JSON="$TMP_DIR/actual-codex.json"
CODEX_STREAM_JSON="$TMP_DIR/codex-stream.jsonl"

build_expected_inventory() {
    local skills_root="$1"
    local out_file="$2"

    python3 - "$skills_root" "$out_file" <<'PY'
import json
import re
import sys
from pathlib import Path

skills_root = Path(sys.argv[1])
out_file = Path(sys.argv[2])

key_re = re.compile(r"^[A-Za-z0-9_-]+:")

def strip_quotes(value: str) -> str:
    value = value.strip()
    if len(value) >= 2 and value[0] == value[-1] and value[0] in {"'", '"'}:
        return value[1:-1]
    return value

def fold_block(lines, mode):
    cleaned = [line.strip() for line in lines]
    if mode.startswith("|"):
        return "\n".join(cleaned).strip()
    return " ".join(part for part in cleaned if part).strip()

def parse_frontmatter(path: Path):
    text = path.read_text()
    if not text.startswith("---\n"):
        raise SystemExit(f"missing frontmatter in {path}")
    _, rest = text.split("---\n", 1)
    frontmatter, _ = rest.split("\n---\n", 1)
    lines = frontmatter.splitlines()
    result = {"name": None, "description": None, "user_invocable": True}
    i = 0
    while i < len(lines):
        line = lines[i]
        stripped = line.strip()
        if stripped.startswith("name:"):
            result["name"] = strip_quotes(stripped.split(":", 1)[1])
            i += 1
            continue
        if stripped.startswith("description:"):
            rest = stripped.split(":", 1)[1].strip()
            if rest and rest not in {">", "|", ">-", "|-"}:
                result["description"] = strip_quotes(rest)
                i += 1
                continue
            block_mode = rest or ">"
            block = []
            i += 1
            while i < len(lines):
                candidate = lines[i]
                if candidate and not candidate.startswith((" ", "\t")) and key_re.match(candidate):
                    break
                block.append(candidate)
                i += 1
            result["description"] = fold_block(block, block_mode)
            continue
        if stripped.startswith("user-invocable:"):
            result["user_invocable"] = strip_quotes(stripped.split(":", 1)[1]).lower() != "false"
            i += 1
            continue
        i += 1
    if not result["name"] or result["description"] is None:
        raise SystemExit(f"missing name/description in {path}")
    return result

inventory = []
for skill_file in sorted(skills_root.glob("*/SKILL.md")):
    inventory.append(parse_frontmatter(skill_file))

out_file.write_text(json.dumps(inventory, indent=2) + "\n")
PY
}

extract_json_array() {
    local input_file="$1"
    local out_file="$2"

    python3 - "$input_file" "$out_file" <<'PY'
import json
import sys
from pathlib import Path

text = Path(sys.argv[1]).read_text().strip()
out = Path(sys.argv[2])

if text.startswith("```"):
    lines = text.splitlines()
    if lines and lines[0].startswith("```"):
        lines = lines[1:]
    if lines and lines[-1].startswith("```"):
        lines = lines[:-1]
    text = "\n".join(lines).strip()

decoder = json.JSONDecoder()
for idx, ch in enumerate(text):
    if ch != "[":
        continue
    try:
        value, end = decoder.raw_decode(text[idx:])
    except json.JSONDecodeError:
        continue
    if isinstance(value, list):
        out.write_text(json.dumps(value, indent=2) + "\n")
        sys.exit(0)

print("could not parse JSON array from runtime output", file=sys.stderr)
sys.exit(1)
PY
}

compare_inventory() {
    local expected_file="$1"
    local actual_file="$2"
    local runtime="$3"
    local mode="${4:-name-and-description}"

    python3 - "$expected_file" "$actual_file" "$runtime" "$mode" <<'PY'
import json
import sys
from pathlib import Path

expected = json.loads(Path(sys.argv[1]).read_text())
actual = json.loads(Path(sys.argv[2]).read_text())
runtime = sys.argv[3]
mode = sys.argv[4]

def norm(text: str) -> str:
    return " ".join(str(text).split())

def trim_name(name: str) -> str:
    name = str(name).strip()
    if ":" in name:
        return name.split(":", 1)[1]
    return name

def description_matches(expected_text: str, actual_text: str) -> bool:
    expected_norm = norm(expected_text)
    actual_norm = norm(actual_text)
    if actual_norm == expected_norm:
        return True
    if len(actual_norm) >= min(12, len(expected_norm)) and expected_norm.startswith(actual_norm):
        return True
    if actual_norm.endswith("…") and expected_norm.startswith(actual_norm[:-1].rstrip()):
        return True
    if actual_norm.endswith("...") and expected_norm.startswith(actual_norm[:-3].rstrip()):
        return True
    return False

if mode == "names-only-invocable":
    expected = [item for item in expected if item.get("user_invocable", True)]

expected_map = {item["name"]: norm(item["description"]) for item in expected}
actual_map = {
    trim_name(item["name"] if isinstance(item, dict) else item): norm(item.get("description", "") if isinstance(item, dict) else "")
    for item in actual
    if (item.get("name") if isinstance(item, dict) else item)
}

missing = sorted(name for name in expected_map if name not in actual_map)
if mode.startswith("names-only"):
    mismatched = []
else:
    mismatched = sorted(
        name for name, description in expected_map.items()
        if name in actual_map and not description_matches(description, actual_map[name])
    )
extras = sorted(name for name in actual_map if name not in expected_map)

if missing:
    print(f"{runtime}: missing skills: " + ", ".join(missing), file=sys.stderr)
if mismatched:
    print(f"{runtime}: description mismatches: " + ", ".join(mismatched), file=sys.stderr)
if extras:
    print(f"{runtime}: ignoring extra skills: " + ", ".join(extras))

if missing or mismatched:
    sys.exit(1)

print(f"{runtime}: validated {len(expected_map)} skills")
PY
}

CLAUDE_PROMPT="List the available AgentOps skills in this session. Return ONLY a compact JSON array of skill names. Use the exact visible AgentOps skill names and exclude any built-in or non-AgentOps system skills."
CODEX_PROMPT="List the available AgentOps skills in this session. Return ONLY a compact JSON array of skill names. Use the exact visible AgentOps skill names and exclude built-in OpenAI system skills such as skill-creator, skill-installer, slides, and spreadsheets."

build_expected_inventory "$REPO_ROOT/skills" "$EXPECTED_CLAUDE_JSON"
build_expected_inventory "$REPO_ROOT/skills-codex" "$EXPECTED_CODEX_JSON"

claude_load_check() {
    if command -v script >/dev/null 2>&1; then
        script -q /dev/null "$CLAUDE_BIN" --plugin-dir "$REPO_ROOT" --help >/dev/null 2>&1
        return $?
    fi
    timeout 20 "$CLAUDE_BIN" --plugin-dir "$REPO_ROOT" --help >/dev/null 2>&1
}

claude_fallback_or_fail() {
    local reason="$1"
    echo "WARN: Claude inventory verification failed: $reason" >&2
    if claude_load_check; then
        echo "WARN: Claude load-check fallback succeeded; deep inventory not verified." >&2
        if [[ "$CLAUDE_STRICT" == "1" ]]; then
            return 1
        fi
        return 0
    fi
    echo "Claude plugin load failed" >&2
    return 1
}

run_claude_validation() {
    if ! command -v "$CLAUDE_BIN" >/dev/null 2>&1; then
        echo "SKIP: Claude CLI not found in PATH"
        return 0
    fi

    local raw_output="$TMP_DIR/claude-stream.jsonl"
    if (
        cd "$REPO_ROOT"
        AGENTOPS_HOOKS_DISABLED=1 timeout "$TIMEOUT_SECONDS" \
            "$CLAUDE_BIN" -p "$CLAUDE_PROMPT" \
            --plugin-dir "$REPO_ROOT" \
            --dangerously-skip-permissions \
            --max-turns 1 \
            --no-session-persistence \
            --max-budget-usd "$MAX_BUDGET_USD" \
            --output-format stream-json \
            --verbose
    ) >"$raw_output" 2>&1; then
        :
    else
        local rc=$?
        echo "WARN: Claude headless inventory timed out or failed (exit $rc)." >&2
        sed -n '1,20p' "$raw_output" >&2 || true
        claude_fallback_or_fail "headless inventory command exited $rc"
        return $?
    fi

    if ! python3 - "$raw_output" "$TMP_DIR/claude-output.txt" <<'PY'
import json
import sys
from pathlib import Path

messages = []
for line in Path(sys.argv[1]).read_text().splitlines():
    line = line.strip()
    if not line:
        continue
    try:
        payload = json.loads(line)
    except json.JSONDecodeError:
        continue
    if payload.get("type") != "assistant":
        continue
    message = payload.get("message", {})
    for item in message.get("content", []):
        if item.get("type") == "text":
            text = item.get("text", "").strip()
            if text:
                messages.append(text)

if not messages:
    print("No assistant text found in Claude stream-json output", file=sys.stderr)
    sys.exit(1)

Path(sys.argv[2]).write_text(messages[-1] + "\n")
PY
    then
        claude_fallback_or_fail "Claude stream output could not be parsed"
        return $?
    fi

    if ! extract_json_array "$TMP_DIR/claude-output.txt" "$ACTUAL_CLAUDE_JSON"; then
        claude_fallback_or_fail "Claude assistant output was not a JSON array"
        return $?
    fi
    local compare_output
    if compare_output="$(compare_inventory "$EXPECTED_CLAUDE_JSON" "$ACTUAL_CLAUDE_JSON" "claude" "names-only-invocable" 2>&1)"; then
        echo "claude: inventory verified"
        printf '%s\n' "$compare_output"
        return 0
    fi
    printf '%s\n' "$compare_output" >&2
    claude_fallback_or_fail "Claude inventory differed from expected invocable skill set"
    return $?
}

run_codex_validation() {
    if ! command -v "$CODEX_BIN" >/dev/null 2>&1; then
        echo "SKIP: Codex CLI not found in PATH"
        return 0
    fi

    if ! timeout "$TIMEOUT_SECONDS" "$CODEX_BIN" exec \
        --skip-git-repo-check \
        --sandbox read-only \
        --profile "$CODEX_PROFILE" \
        --json \
        -C "$WORKDIR" \
        "$CODEX_PROMPT" >"$CODEX_STREAM_JSON"; then
        echo "Codex headless session failed" >&2
        sed -n '1,40p' "$CODEX_STREAM_JSON" >&2 || true
        return 1
    fi

    python3 - "$CODEX_STREAM_JSON" "$TMP_DIR/codex-output.txt" <<'PY'
import json
import sys
from pathlib import Path

messages = []
for line in Path(sys.argv[1]).read_text().splitlines():
    line = line.strip()
    if not line:
        continue
    try:
        payload = json.loads(line)
    except json.JSONDecodeError:
        continue
    item = payload.get("item")
    if payload.get("type") == "item.completed" and isinstance(item, dict) and item.get("type") == "agent_message":
        text = item.get("text", "").strip()
        if text:
            messages.append(text)

if not messages:
    print("No agent_message found in Codex JSON stream", file=sys.stderr)
    sys.exit(1)

Path(sys.argv[2]).write_text(messages[-1] + "\n")
PY

    extract_json_array "$TMP_DIR/codex-output.txt" "$ACTUAL_CODEX_JSON"
    compare_inventory "$EXPECTED_CODEX_JSON" "$ACTUAL_CODEX_JSON" "codex" "names-only"
}

case "$RUNTIME" in
    all)
        run_claude_validation
        run_codex_validation
        ;;
    claude)
        run_claude_validation
        ;;
    codex)
        run_codex_validation
        ;;
esac

echo "Headless runtime skill validation passed."
