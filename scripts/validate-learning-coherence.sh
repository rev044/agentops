#!/usr/bin/env bash
set -euo pipefail

# validate-learning-coherence.sh — Quality gate for learning files
# Checks learning files for hallucination indicators and structural issues.
# Exit 0 = all pass, Exit 1 = failures found (for CI gating).

LEARNINGS_DIR="${1:-.agents/learnings}"
VERBOSE="${VERBOSE:-false}"
FAILURES=0
CHECKED=0

log() { [[ "$VERBOSE" == "true" ]] && echo "$@" || true; }

check_file() {
    local file="$1"
    local basename
    basename=$(basename "$file")

    # Skip non-markdown
    [[ "$file" == *.md ]] || return 0

    CHECKED=$((CHECKED + 1))
    local content
    content=$(cat "$file")

    # 1. Check for frontmatter
    if ! echo "$content" | head -1 | grep -q '^---$'; then
        echo "FAIL: $basename — missing YAML frontmatter"
        FAILURES=$((FAILURES + 1))
        return
    fi

    # 2. Extract body (after second ---)
    local body
    body=$(echo "$content" | sed '1,/^---$/d' | sed '1,/^---$/d')

    # 3. Check minimum content (< 50 words = likely stub)
    local word_count
    word_count=$(echo "$body" | wc -w | tr -d ' ')
    if [[ "$word_count" -lt 50 ]]; then
        echo "FAIL: $basename — too short ($word_count words, minimum 50)"
        FAILURES=$((FAILURES + 1))
        return
    fi

    # 4. Check for required frontmatter fields
    local frontmatter
    frontmatter=$(echo "$content" | sed -n '2,/^---$/p' | sed '$d')

    for field in "id:" "date:" "confidence:"; do
        if ! echo "$frontmatter" | grep -q "$field"; then
            echo "FAIL: $basename — missing frontmatter field: $field"
            FAILURES=$((FAILURES + 1))
            return
        fi
    done

    # 5. Check confidence range (0.0-1.0)
    local confidence
    confidence=$(echo "$frontmatter" | grep 'confidence:' | head -1 | awk '{print $2}')
    if [[ -n "$confidence" ]]; then
        if ! awk -v c="$confidence" 'BEGIN { exit !(c >= 0.0 && c <= 1.0) }' 2>/dev/null; then
            echo "FAIL: $basename — confidence out of range: $confidence (must be 0.0-1.0)"
            FAILURES=$((FAILURES + 1))
            return
        fi
    fi

    # 6. Check for boilerplate/template indicators
    if echo "$body" | grep -qi "no significant learnings\|placeholder\|TODO: fill in\|template content"; then
        echo "FAIL: $basename — contains boilerplate/placeholder content"
        FAILURES=$((FAILURES + 1))
        return
    fi

    log "PASS: $basename ($word_count words)"
}

# Main
if [[ ! -d "$LEARNINGS_DIR" ]]; then
    echo "No learnings directory found at $LEARNINGS_DIR"
    exit 0
fi

for file in "$LEARNINGS_DIR"/*.md; do
    [[ -f "$file" ]] || continue
    check_file "$file"
done

echo ""
echo "Coherence check: $CHECKED files checked, $FAILURES failures"

if [[ "$FAILURES" -gt 0 ]]; then
    echo "Coherence gate FAILED"
    exit 1
fi

echo "Coherence gate passed"
exit 0
