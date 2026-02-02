#!/usr/bin/env bash
# Enhanced test helpers for Claude Code skill tests
# Adapted from superpowers (0x-chad/superpowers)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Test configuration
MAX_TURNS="${MAX_TURNS:-3}"
DEFAULT_TIMEOUT="${DEFAULT_TIMEOUT:-120}"
LOG_DIR="${LOG_DIR:-$SCRIPT_DIR/logs}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Create log directory
mkdir -p "$LOG_DIR"

# Run Claude Code with a prompt and capture output as plain text
# Usage: run_claude "prompt text" [timeout_seconds]
run_claude() {
    local prompt="$1"
    local timeout="${2:-$DEFAULT_TIMEOUT}"
    local output_file
    output_file="$(mktemp)"

    if timeout "$timeout" claude -p "$prompt" \
        --plugin-dir "$REPO_ROOT" \
        --dangerously-skip-permissions \
        --max-turns "$MAX_TURNS" \
        > "$output_file" 2>&1; then
        cat "$output_file"
        rm -f "$output_file"
        return 0
    else
        local exit_code=$?
        cat "$output_file" >&2
        rm -f "$output_file"
        return $exit_code
    fi
}

# Run Claude Code with stream-json output for skill detection
# Usage: run_claude_json "prompt text" [timeout_seconds]
# Returns: path to JSON log file
run_claude_json() {
    local prompt="$1"
    local timeout="${2:-$DEFAULT_TIMEOUT}"
    local ts
    ts="$(date +%s)"
    local log_file="$LOG_DIR/claude-${ts}-$$.jsonl"

    if timeout "$timeout" claude -p "$prompt" \
        --plugin-dir "$REPO_ROOT" \
        --dangerously-skip-permissions \
        --max-turns "$MAX_TURNS" \
        --output-format stream-json \
        > "$log_file" 2>&1; then
        echo "$log_file"
        return 0
    else
        local exit_code=$?
        echo "$log_file"
        return $exit_code
    fi
}

# Check if a skill was triggered (looks in stream-json output)
# Usage: assert_skill_triggered "log_file" "skill-name" "test name"
assert_skill_triggered() {
    local log_file="$1"
    local skill_name="$2"
    local test_name="${3:-Skill triggered}"

    if [[ ! -f "$log_file" ]]; then
        echo -e "  ${RED}[FAIL]${NC} $test_name: Log file not found"
        return 1
    fi

    # Look for Skill tool invocation with the skill name
    # Pattern: "name":"Skill" and "skill":"skillname" (with optional namespace prefix)
    local skill_pattern='"skill":"([^"]*:)?'"${skill_name}"'"'

    if grep -q '"name":"Skill"' "$log_file" && grep -qE "$skill_pattern" "$log_file"; then
        echo -e "  ${GREEN}[PASS]${NC} $test_name: $skill_name was triggered"
        return 0
    else
        echo -e "  ${RED}[FAIL]${NC} $test_name: $skill_name was NOT triggered"
        echo "  Searched for: $skill_pattern"
        if [[ -f "$log_file" ]]; then
            echo "  Tool calls found:"
            grep '"name":' "$log_file" | head -5 | sed 's/^/    /' || true
        fi
        return 1
    fi
}

# Check that no tools were called before skill invocation
# Usage: assert_no_premature_tools "log_file" "test name"
assert_no_premature_tools() {
    local log_file="$1"
    local test_name="${2:-No premature tools}"

    if [[ ! -f "$log_file" ]]; then
        echo -e "  ${RED}[FAIL]${NC} $test_name: Log file not found"
        return 1
    fi

    # Find line number of Skill invocation
    local skill_line
    skill_line="$(grep -n '"name":"Skill"' "$log_file" | head -1 | cut -d: -f1)"

    if [[ -z "$skill_line" ]]; then
        echo -e "  ${YELLOW}[SKIP]${NC} $test_name: No Skill invocation found"
        return 0
    fi

    # Check for tool calls before Skill
    local premature_tools
    premature_tools="$(head -n "$skill_line" "$log_file" | grep -E '"name":"(Bash|Read|Write|Edit|Glob|Grep)"' | head -3)"

    if [[ -n "$premature_tools" ]]; then
        echo -e "  ${RED}[FAIL]${NC} $test_name: Tools called before Skill"
        echo "$premature_tools" | head -3 | sed 's/^/    /'
        return 1
    else
        echo -e "  ${GREEN}[PASS]${NC} $test_name"
        return 0
    fi
}

# Check if output contains a pattern
# Usage: assert_contains "output" "pattern" "test name"
assert_contains() {
    local output="$1"
    local pattern="$2"
    local test_name="${3:-test}"

    if echo "$output" | grep -qi "$pattern"; then
        echo -e "  ${GREEN}[PASS]${NC} $test_name"
        return 0
    else
        echo -e "  ${RED}[FAIL]${NC} $test_name"
        echo "  Expected to find: $pattern"
        echo "  In output (first 500 chars):"
        echo "$output" | head -c 500 | sed 's/^/    /'
        return 1
    fi
}

# Check if output does NOT contain a pattern
# Usage: assert_not_contains "output" "pattern" "test name"
assert_not_contains() {
    local output="$1"
    local pattern="$2"
    local test_name="${3:-test}"

    if echo "$output" | grep -qi "$pattern"; then
        echo -e "  ${RED}[FAIL]${NC} $test_name"
        echo "  Did not expect to find: $pattern"
        return 1
    else
        echo -e "  ${GREEN}[PASS]${NC} $test_name"
        return 0
    fi
}

# Check if pattern A appears before pattern B
# Usage: assert_order "output" "pattern_a" "pattern_b" "test name"
assert_order() {
    local output="$1"
    local pattern_a="$2"
    local pattern_b="$3"
    local test_name="${4:-test}"

    local line_a
    line_a="$(echo "$output" | grep -n -i "$pattern_a" | head -1 | cut -d: -f1)"
    local line_b
    line_b="$(echo "$output" | grep -n -i "$pattern_b" | head -1 | cut -d: -f1)"

    if [[ -z "$line_a" ]]; then
        echo -e "  ${RED}[FAIL]${NC} $test_name: pattern A not found: $pattern_a"
        return 1
    fi

    if [[ -z "$line_b" ]]; then
        echo -e "  ${RED}[FAIL]${NC} $test_name: pattern B not found: $pattern_b"
        return 1
    fi

    if [[ "$line_a" -lt "$line_b" ]]; then
        echo -e "  ${GREEN}[PASS]${NC} $test_name (A at line $line_a, B at line $line_b)"
        return 0
    else
        echo -e "  ${RED}[FAIL]${NC} $test_name"
        echo "  Expected '$pattern_a' before '$pattern_b'"
        return 1
    fi
}

# Check if a specific tool was called
# Usage: assert_tool_called "log_file" "ToolName" "test name"
assert_tool_called() {
    local log_file="$1"
    local tool_name="$2"
    local test_name="${3:-Tool called}"

    if [[ ! -f "$log_file" ]]; then
        echo -e "  ${RED}[FAIL]${NC} $test_name: Log file not found"
        return 1
    fi

    if grep -q "\"name\":\"$tool_name\"" "$log_file"; then
        echo -e "  ${GREEN}[PASS]${NC} $test_name: $tool_name was called"
        return 0
    else
        echo -e "  ${RED}[FAIL]${NC} $test_name: $tool_name was NOT called"
        return 1
    fi
}

# Check if a specific tool was NOT called
# Usage: assert_tool_not_called "log_file" "ToolName" "test name"
assert_tool_not_called() {
    local log_file="$1"
    local tool_name="$2"
    local test_name="${3:-Tool not called}"

    if [[ ! -f "$log_file" ]]; then
        echo -e "  ${RED}[FAIL]${NC} $test_name: Log file not found"
        return 1
    fi

    if grep -q "\"name\":\"$tool_name\"" "$log_file"; then
        echo -e "  ${RED}[FAIL]${NC} $test_name: $tool_name WAS called (should not be)"
        return 1
    else
        echo -e "  ${GREEN}[PASS]${NC} $test_name"
        return 0
    fi
}

# Create a temporary test project directory
# Usage: test_project=$(create_test_project)
create_test_project() {
    local test_dir
    test_dir="$(mktemp -d)"
    mkdir -p "$test_dir/.agents/learnings"
    mkdir -p "$test_dir/.agents/research"
    mkdir -p "$test_dir/.beads"
    echo "$test_dir"
}

# Cleanup test project
# Usage: cleanup_test_project "$test_dir"
cleanup_test_project() {
    local test_dir="$1"
    if [[ -d "$test_dir" ]]; then
        rm -rf "$test_dir"
    fi
}

# Cleanup old log files (keep last 50)
cleanup_logs() {
    if [[ -d "$LOG_DIR" ]]; then
        local count
        count="$(find "$LOG_DIR" -name "*.jsonl" -type f | wc -l | tr -d ' ')"
        if [[ "$count" -gt 50 ]]; then
            find "$LOG_DIR" -name "*.jsonl" -type f -printf '%T@ %p\n' | \
                sort -n | head -n $((count - 50)) | cut -d' ' -f2- | xargs rm -f
        fi
    fi
}

# Print test summary
# Usage: print_summary passed failed skipped
print_summary() {
    local passed="${1:-0}"
    local failed="${2:-0}"
    local skipped="${3:-0}"
    local total=$((passed + failed + skipped))

    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════${NC}"
    echo -e "Tests: $total total"
    echo -e "  ${GREEN}Passed:${NC}  $passed"
    echo -e "  ${RED}Failed:${NC}  $failed"
    echo -e "  ${YELLOW}Skipped:${NC} $skipped"
    echo -e "${BLUE}═══════════════════════════════════════════${NC}"

    if [[ "$failed" -gt 0 ]]; then
        return 1
    fi
    return 0
}

# Export functions for use in tests
export -f run_claude
export -f run_claude_json
export -f assert_skill_triggered
export -f assert_no_premature_tools
export -f assert_contains
export -f assert_not_contains
export -f assert_order
export -f assert_tool_called
export -f assert_tool_not_called
export -f create_test_project
export -f cleanup_test_project
export -f cleanup_logs
export -f print_summary
export REPO_ROOT
export LOG_DIR
export MAX_TURNS
export DEFAULT_TIMEOUT
