#!/usr/bin/env bash
set -euo pipefail

# coverage-ratchet.sh — Per-package coverage ratchet gate.
#
# Compares current per-package coverage against a checked-in baseline.
# Fails if any package drops below its baseline.
#
# Usage:
#   scripts/coverage-ratchet.sh --check     # Compare against baseline (CI gate)
#   scripts/coverage-ratchet.sh --update    # Refresh baseline from current state
#   scripts/coverage-ratchet.sh --show      # Display current vs baseline

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CLI_DIR="$REPO_ROOT/cli"
BASELINE="$REPO_ROOT/.coverage-baseline.json"

if ! command -v go >/dev/null 2>&1; then
    echo "SKIP: go is not installed"
    exit 0
fi

if ! command -v jq >/dev/null 2>&1; then
    echo "SKIP: jq is not installed"
    exit 0
fi

# Collect per-package coverage via go test -cover
collect_coverage() {
    local tmp_json
    tmp_json="$(mktemp)"
    trap 'rm -f "$tmp_json"' RETURN

    echo "{" > "$tmp_json"
    local first=true

    # Test each internal package and extract coverage percentage
    for pkg_dir in "$CLI_DIR"/internal/*/; do
        [[ -d "$pkg_dir" ]] || continue
        pkg_name="$(basename "$pkg_dir")"

        # Skip packages with no test files
        if ! ls "$pkg_dir"/*_test.go >/dev/null 2>&1; then
            continue
        fi

        coverage_line="$(cd "$CLI_DIR" && go test -cover "./internal/$pkg_name" 2>&1 | grep -oE 'coverage: [0-9.]+%' | head -1 || true)"
        if [[ -z "$coverage_line" ]]; then
            continue
        fi

        pct="$(echo "$coverage_line" | grep -oE '[0-9.]+')"
        if [[ -n "$pct" ]]; then
            if [[ "$first" == true ]]; then
                first=false
            else
                echo "," >> "$tmp_json"
            fi
            printf '  "internal/%s": %.1f' "$pkg_name" "$pct" >> "$tmp_json"
        fi
    done

    # cmd/ao coverage (the monolith)
    cmd_coverage="$(cd "$CLI_DIR" && go test -cover ./cmd/ao 2>&1 | grep -oE 'coverage: [0-9.]+%' | head -1 || true)"
    if [[ -n "$cmd_coverage" ]]; then
        cmd_pct="$(echo "$cmd_coverage" | grep -oE '[0-9.]+')"
        if [[ -n "$cmd_pct" ]]; then
            if [[ "$first" == true ]]; then
                first=false
            else
                echo "," >> "$tmp_json"
            fi
            printf '  "cmd/ao": %.1f' "$cmd_pct" >> "$tmp_json"
        fi
    fi

    echo "" >> "$tmp_json"
    echo "}" >> "$tmp_json"

    cat "$tmp_json"
}

case "${1:-}" in
    --update)
        echo "Collecting current coverage for baseline update..."
        collect_coverage > "$BASELINE"
        echo "Baseline written to $BASELINE"
        jq '.' "$BASELINE"
        ;;

    --check)
        if [[ ! -f "$BASELINE" ]]; then
            echo "WARN: No baseline found at $BASELINE"
            echo "Run 'scripts/coverage-ratchet.sh --update' to create one."
            echo "Skipping ratchet check."
            exit 0
        fi

        echo "Collecting current coverage..."
        current="$(mktemp)"
        trap 'rm -f "$current"' EXIT
        collect_coverage > "$current"

        echo "Comparing against baseline..."
        echo ""

        drops=0
        while IFS= read -r pkg; do
            baseline_pct="$(jq -r --arg p "$pkg" '.[$p] // empty' "$BASELINE")"
            current_pct="$(jq -r --arg p "$pkg" '.[$p] // empty' "$current")"

            if [[ -z "$baseline_pct" ]] || [[ -z "$current_pct" ]]; then
                continue
            fi

            if awk -v c="$current_pct" -v b="$baseline_pct" 'BEGIN { exit !(c+0 < b+0) }'; then
                printf "  DROP  %-40s  %.1f%% → %.1f%%\n" "$pkg" "$baseline_pct" "$current_pct"
                drops=$((drops + 1))
            else
                printf "  ok    %-40s  %.1f%% (baseline: %.1f%%)\n" "$pkg" "$current_pct" "$baseline_pct"
            fi
        done < <(jq -r 'keys[]' "$BASELINE" | sort)

        # Check for new packages not in baseline
        while IFS= read -r pkg; do
            if ! jq -e --arg p "$pkg" '.[$p]' "$BASELINE" >/dev/null 2>&1; then
                current_pct="$(jq -r --arg p "$pkg" '.[$p]' "$current")"
                printf "  NEW   %-40s  %.1f%%\n" "$pkg" "$current_pct"
            fi
        done < <(jq -r 'keys[]' "$current" | sort)

        echo ""
        if [[ "$drops" -gt 0 ]]; then
            echo "FAIL: $drops package(s) dropped below baseline coverage"
            exit 1
        else
            echo "PASS: all packages at or above baseline"
        fi
        ;;

    --show)
        if [[ ! -f "$BASELINE" ]]; then
            echo "No baseline found. Run --update first."
            exit 0
        fi
        echo "Current baseline ($BASELINE):"
        jq '.' "$BASELINE"
        ;;

    *)
        echo "Usage: scripts/coverage-ratchet.sh [--check|--update|--show]"
        echo ""
        echo "  --check   Compare current coverage against baseline (CI gate)"
        echo "  --update  Refresh baseline from current state"
        echo "  --show    Display current baseline"
        exit 1
        ;;
esac
