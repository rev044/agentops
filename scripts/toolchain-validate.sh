#!/usr/bin/env bash
set -euo pipefail

# Toolchain Validate - Run all available linters/scanners
# Outputs structured findings to .agents/tooling/
#
# Usage: ./scripts/toolchain-validate.sh [OPTIONS]
#
# Options:
#   --quick   Skip slow tools (tests, comprehensive scans)
#   --json    Output summary as JSON to stdout
#   --gate    Exit non-zero on CRITICAL or HIGH findings
#
# Exit Codes:
#   0 - Pass (no critical/high findings, or --gate not specified)
#   1 - Script error
#   2 - CRITICAL findings found (with --gate)
#   3 - HIGH findings only (with --gate)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"
OUTPUT_DIR="$REPO_ROOT/.agents/tooling"

# Parse arguments
QUICK=false
JSON_OUTPUT=false
GATE=false

for arg in "$@"; do
    case $arg in
        --quick) QUICK=true ;;
        --json) JSON_OUTPUT=true ;;
        --gate) GATE=true ;;
        --help|-h)
            head -20 "$0" | grep "^#" | sed 's/^# *//'
            exit 0
            ;;
        *)
            echo "Unknown option: $arg" >&2
            exit 1
            ;;
    esac
done

# Initialize output directory
mkdir -p "$OUTPUT_DIR"

# Counters
CRITICAL_COUNT=0
HIGH_COUNT=0
MEDIUM_COUNT=0
LOW_COUNT=0
TOOLS_RUN=0
TOOLS_SKIPPED=0

# Tool output files and status
declare -A TOOL_FILES
declare -A TOOL_STATUS

log() {
    if [[ "$JSON_OUTPUT" != "true" ]]; then
        echo "$1"
    fi
}

run_tool() {
    local name="$1"
    local output_file="$OUTPUT_DIR/${name}.txt"
    shift

    TOOL_FILES["$name"]="$output_file"

    if ! command -v "$1" &>/dev/null; then
        log "  [SKIP] $name - not installed"
        echo "NOT_INSTALLED" > "$output_file"
        TOOL_STATUS["$name"]="not_installed"
        TOOLS_SKIPPED=$((TOOLS_SKIPPED + 1))
        return 1
    fi

    log "  [RUN]  $name"
    TOOLS_RUN=$((TOOLS_RUN + 1))
    return 0
}

# ============================================================================
# TOOL: ruff (Python linting)
# ============================================================================
run_ruff() {
    local output_file="$OUTPUT_DIR/ruff.txt"

    if ! run_tool "ruff" ruff; then return 0; fi

    # Check if there are Python files
    if ! find "$REPO_ROOT" -name "*.py" -type f | head -1 | grep -q .; then
        echo "NO_PYTHON_FILES" > "$output_file"
        TOOL_STATUS["ruff"]="skipped"
        return 0
    fi

    # Run ruff and capture output
    if ruff check "$REPO_ROOT" --output-format=text > "$output_file" 2>&1; then
        echo "CLEAN" > "$output_file"
        TOOL_STATUS["ruff"]="pass"
    else
        # Count issues by severity (ruff uses E=error, W=warning, F=fatal)
        local errors warnings
        errors=$(grep -cE "^[^:]+:[0-9]+:[0-9]+: [EF]" "$output_file" 2>/dev/null || true)
        errors=${errors:-0}
        errors=$(echo "$errors" | tr -d '[:space:]')
        warnings=$(grep -cE "^[^:]+:[0-9]+:[0-9]+: [WC]" "$output_file" 2>/dev/null || true)
        warnings=${warnings:-0}
        warnings=$(echo "$warnings" | tr -d '[:space:]')
        HIGH_COUNT=$((HIGH_COUNT + errors))
        MEDIUM_COUNT=$((MEDIUM_COUNT + warnings))
        TOOL_STATUS["ruff"]="findings"
    fi
}

# ============================================================================
# TOOL: golangci-lint (Go linting)
# ============================================================================
run_golangci() {
    local output_file="$OUTPUT_DIR/golangci-lint.txt"

    if ! run_tool "golangci-lint" golangci-lint; then return 0; fi

    # Check if there are Go files
    if ! find "$REPO_ROOT" -name "*.go" -type f | head -1 | grep -q .; then
        echo "NO_GO_FILES" > "$output_file"
        TOOL_STATUS["golangci-lint"]="skipped"
        return 0
    fi

    # Run golangci-lint (redirect all output to file)
    if golangci-lint run "$REPO_ROOT/..." > "$output_file" 2>&1; then
        echo "CLEAN" > "$output_file"
        TOOL_STATUS["golangci-lint"]="pass"
    else
        # Count issues by matching file:line:col: pattern (skip headers/footers)
        local issues
        issues=$(grep -cE "^[^:]+:[0-9]+:[0-9]+:" "$output_file" 2>/dev/null || true)
        issues=${issues:-0}
        issues=$(echo "$issues" | tr -d '[:space:]')
        HIGH_COUNT=$((HIGH_COUNT + issues))
        TOOL_STATUS["golangci-lint"]="findings"
    fi
}

# ============================================================================
# TOOL: gitleaks (secret scanning)
# ============================================================================
run_gitleaks() {
    local output_file="$OUTPUT_DIR/gitleaks.txt"

    if ! run_tool "gitleaks" gitleaks; then return 0; fi

    # Run gitleaks (use --no-color to avoid ANSI codes, redirect stderr to file too)
    if gitleaks detect --source="$REPO_ROOT" --no-git --no-color > "$output_file" 2>&1; then
        echo "CLEAN" > "$output_file"
        TOOL_STATUS["gitleaks"]="pass"
    else
        # Count leaks - gitleaks outputs one block per finding
        local leaks
        leaks=$(grep -c "Secret:" "$output_file" 2>/dev/null || true)
        leaks=${leaks:-0}
        leaks=$(echo "$leaks" | tr -d '[:space:]')
        CRITICAL_COUNT=$((CRITICAL_COUNT + leaks))
        TOOL_STATUS["gitleaks"]="findings"
    fi
}

# ============================================================================
# TOOL: shellcheck (shell script linting)
# ============================================================================
run_shellcheck() {
    local output_file="$OUTPUT_DIR/shellcheck.txt"

    if ! run_tool "shellcheck" shellcheck; then return 0; fi

    # Find all shell scripts
    local scripts
    scripts=$(find "$REPO_ROOT" -name "*.sh" -type f ! -path "*/.git/*" 2>/dev/null || true)

    if [[ -z "$scripts" ]]; then
        echo "NO_SHELL_FILES" > "$output_file"
        TOOL_STATUS["shellcheck"]="skipped"
        return 0
    fi

    # Run shellcheck
    echo "$scripts" | xargs shellcheck -f gcc 2>&1 > "$output_file" || true

    if [[ ! -s "$output_file" ]]; then
        echo "CLEAN" > "$output_file"
        TOOL_STATUS["shellcheck"]="pass"
    else
        # Count by severity (shellcheck gcc format: "file:line:col: error: message")
        local errors warnings
        errors=$(grep -cE ": error:" "$output_file" 2>/dev/null || true)
        errors=${errors:-0}
        errors=$(echo "$errors" | tr -d '[:space:]')
        warnings=$(grep -cE ": warning:" "$output_file" 2>/dev/null || true)
        warnings=${warnings:-0}
        warnings=$(echo "$warnings" | tr -d '[:space:]')
        HIGH_COUNT=$((HIGH_COUNT + errors))
        MEDIUM_COUNT=$((MEDIUM_COUNT + warnings))
        if [[ $errors -gt 0 || $warnings -gt 0 ]]; then
            TOOL_STATUS["shellcheck"]="findings"
        else
            TOOL_STATUS["shellcheck"]="pass"
        fi
    fi
}

# ============================================================================
# TOOL: radon (Python complexity)
# ============================================================================
run_radon() {
    local output_file="$OUTPUT_DIR/radon.txt"

    if ! run_tool "radon" radon; then return 0; fi

    # Check if there are Python files
    if ! find "$REPO_ROOT" -name "*.py" -type f | head -1 | grep -q .; then
        echo "NO_PYTHON_FILES" > "$output_file"
        TOOL_STATUS["radon"]="skipped"
        return 0
    fi

    # Run radon for cyclomatic complexity (min C = 11+)
    radon cc "$REPO_ROOT" -a -s --min C 2>&1 > "$output_file" || true

    if [[ ! -s "$output_file" ]]; then
        echo "CLEAN" > "$output_file"
        TOOL_STATUS["radon"]="pass"
    else
        # Count high complexity functions
        local complex
        complex=$(grep -cE "^\s+[A-Z] " "$output_file" 2>/dev/null || true)
        complex=${complex:-0}
        complex=$(echo "$complex" | tr -d '[:space:]')
        HIGH_COUNT=$((HIGH_COUNT + complex))
        if [[ $complex -gt 0 ]]; then
            TOOL_STATUS["radon"]="findings"
        else
            TOOL_STATUS["radon"]="pass"
        fi
    fi
}

# ============================================================================
# TOOL: pytest (Python tests) - skipped in quick mode
# ============================================================================
run_pytest() {
    if [[ "$QUICK" == "true" ]]; then
        log "  [SKIP] pytest - quick mode"
        TOOLS_SKIPPED=$((TOOLS_SKIPPED + 1))
        echo "SKIPPED_QUICK_MODE" > "$OUTPUT_DIR/pytest.txt"
        TOOL_STATUS["pytest"]="skipped"
        return 0
    fi

    local output_file="$OUTPUT_DIR/pytest.txt"

    if ! run_tool "pytest" pytest; then return 0; fi

    # Check if there are test files
    if ! find "$REPO_ROOT" -name "test_*.py" -o -name "*_test.py" | head -1 | grep -q .; then
        echo "NO_TEST_FILES" > "$output_file"
        TOOL_STATUS["pytest"]="skipped"
        return 0
    fi

    # Run pytest with minimal output
    if pytest "$REPO_ROOT" --tb=short -q > "$output_file" 2>&1; then
        echo "PASS" >> "$output_file"
        TOOL_STATUS["pytest"]="pass"
    else
        local failures
        failures=$(grep -cE "^FAILED" "$output_file" 2>/dev/null || true)
        failures=${failures:-0}
        failures=$(echo "$failures" | tr -d '[:space:]')
        CRITICAL_COUNT=$((CRITICAL_COUNT + failures))
        TOOL_STATUS["pytest"]="findings"
    fi
}

# ============================================================================
# TOOL: go test - skipped in quick mode
# ============================================================================
run_gotest() {
    if [[ "$QUICK" == "true" ]]; then
        log "  [SKIP] go test - quick mode"
        TOOLS_SKIPPED=$((TOOLS_SKIPPED + 1))
        echo "SKIPPED_QUICK_MODE" > "$OUTPUT_DIR/gotest.txt"
        TOOL_STATUS["go-test"]="skipped"
        return 0
    fi

    local output_file="$OUTPUT_DIR/gotest.txt"

    if ! run_tool "go-test" go; then return 0; fi

    # Check if there are Go test files
    if ! find "$REPO_ROOT" -name "*_test.go" -type f | head -1 | grep -q .; then
        echo "NO_TEST_FILES" > "$output_file"
        TOOL_STATUS["go-test"]="skipped"
        return 0
    fi

    # Run go test
    if (cd "$REPO_ROOT" && go test ./... -short) > "$output_file" 2>&1; then
        echo "PASS" >> "$output_file"
        TOOL_STATUS["go-test"]="pass"
    else
        local failures
        failures=$(grep -c "^--- FAIL" "$output_file" 2>/dev/null || true)
        failures=${failures:-0}
        failures=$(echo "$failures" | tr -d '[:space:]')
        CRITICAL_COUNT=$((CRITICAL_COUNT + failures))
        TOOL_STATUS["go-test"]="findings"
    fi
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

log ""
log "Toolchain Validation"
log "===================="
log "Target: $REPO_ROOT"
log "Output: $OUTPUT_DIR"
log ""

# Run all tools
log "Running tools..."
run_ruff
run_golangci
run_gitleaks
run_shellcheck
run_radon
run_pytest
run_gotest

log ""

# Compute gate status once
if [[ $CRITICAL_COUNT -gt 0 ]]; then
    GATE_STATUS="BLOCKED_CRITICAL"
elif [[ $HIGH_COUNT -gt 0 ]]; then
    GATE_STATUS="BLOCKED_HIGH"
else
    GATE_STATUS="PASS"
fi

# Build tools JSON object
TOOLS_JSON="{"
first=true
for tool in ruff golangci-lint gitleaks shellcheck radon pytest go-test; do
    status="${TOOL_STATUS[$tool]:-not_run}"
    if [[ "$first" == "true" ]]; then
        first=false
    else
        TOOLS_JSON="$TOOLS_JSON,"
    fi
    TOOLS_JSON="$TOOLS_JSON \"$tool\": \"$status\""
done
TOOLS_JSON="$TOOLS_JSON }"

# Generate summary
SUMMARY=$(cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "target": "$REPO_ROOT",
  "tools_run": $TOOLS_RUN,
  "tools_skipped": $TOOLS_SKIPPED,
  "tools": $TOOLS_JSON,
  "findings": {
    "critical": $CRITICAL_COUNT,
    "high": $HIGH_COUNT,
    "medium": $MEDIUM_COUNT,
    "low": $LOW_COUNT
  },
  "gate_status": "$GATE_STATUS",
  "output_dir": "$OUTPUT_DIR"
}
EOF
)

# Write summary file
echo "$SUMMARY" > "$OUTPUT_DIR/summary.json"

# Output based on mode
if [[ "$JSON_OUTPUT" == "true" ]]; then
    echo "$SUMMARY"
else
    log "Summary"
    log "-------"
    log "  Tools run: $TOOLS_RUN"
    log "  Tools skipped: $TOOLS_SKIPPED"
    log ""
    log "  Findings:"
    log "    CRITICAL: $CRITICAL_COUNT"
    log "    HIGH:     $HIGH_COUNT"
    log "    MEDIUM:   $MEDIUM_COUNT"
    log "    LOW:      $LOW_COUNT"
    log ""

    if [[ "$GATE_STATUS" == "BLOCKED_CRITICAL" ]]; then
        log "  Gate: BLOCKED (${CRITICAL_COUNT} critical findings)"
    elif [[ "$GATE_STATUS" == "BLOCKED_HIGH" ]]; then
        log "  Gate: BLOCKED (${HIGH_COUNT} high findings)"
    else
        log "  Gate: PASS"
    fi

    log ""
    log "Full output: $OUTPUT_DIR"
fi

# Exit code logic
if [[ "$GATE" == "true" ]]; then
    if [[ "$GATE_STATUS" == "BLOCKED_CRITICAL" ]]; then
        exit 2
    elif [[ "$GATE_STATUS" == "BLOCKED_HIGH" ]]; then
        exit 3
    fi
fi

exit 0
