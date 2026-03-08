#!/usr/bin/env bash
set -euo pipefail

# validate-learning-coherence.sh — Quality gate for learning files
# Catches hallucinated/garbage learnings before flywheel ingest.
# Exit 0 = all pass, Exit 1 = failures found (for CI gating).
#
# Two learning formats are valid:
#   1. MemRL-managed: YAML frontmatter with utility/confidence/maturity fields
#      plus a non-empty body section.
#   2. Manual: YAML frontmatter with id/date fields + substantive body (>=50 words)
#
# Failure signals (garbage indicators):
#   - No YAML frontmatter AND < 30 words (truly empty junk)
#   - Has frontmatter but zero recognized fields (malformed)
#   - Frontmatter-only artifacts (no body content)
#   - Boilerplate/placeholder content detected
#   - Confidence value outside 0.0-1.0 range

LEARNINGS_DIR="${1:-.agents/learnings}"
VERBOSE="${VERBOSE:-false}"
FAILURES=0
WARNINGS=0
CHECKED=0

log() { [[ "$VERBOSE" == "true" ]] && echo "$@" || true; }

is_learning_artifact() {
    local file="$1"
    local basename
    basename=$(basename "$file")
    [[ "$basename" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}-.*\.md$ ]]
}

check_file() {
    local file="$1"
    local basename
    basename=$(basename "$file")

    # Skip non-markdown
    [[ "$file" == *.md ]] || return 0

    CHECKED=$((CHECKED + 1))
    local content
    content=$(cat "$file")
    local total_words
    total_words=$(echo "$content" | wc -w | tr -d ' ')

    # 1. Check for frontmatter presence
    local has_frontmatter=false
    if echo "$content" | head -1 | grep -q '^---$'; then
        has_frontmatter=true
    fi

    # 2. No frontmatter: require at least 30 words of content
    if [[ "$has_frontmatter" == "false" ]]; then
        if [[ "$total_words" -lt 30 ]]; then
            echo "FAIL: $basename — no frontmatter and too short ($total_words words)"
            FAILURES=$((FAILURES + 1))
            return
        fi
        log "PASS: $basename (no frontmatter, $total_words words)"
        return
    fi

    # 3. Extract frontmatter (between first and second ---)
    local frontmatter
    frontmatter=$(echo "$content" | sed -n '2,/^---$/p' | sed '$d')

    # 4. Check for at least 1 recognized field in frontmatter
    # MemRL fields: utility, confidence, maturity, reward_count
    # Manual fields: id, date, source_epic, tags
    local recognized=0
    for field in "utility:" "confidence:" "maturity:" "id:" "date:" "source_epic:"; do
        if echo "$frontmatter" | grep -q "$field"; then
            recognized=$((recognized + 1))
        fi
    done

    if [[ "$recognized" -eq 0 ]]; then
        echo "FAIL: $basename — frontmatter has no recognized fields"
        FAILURES=$((FAILURES + 1))
        return
    fi

    # 5. Check confidence value if present
    #    Valid: numeric 0.0-1.0  OR  string high/medium/low
    local confidence
    confidence=$(echo "$frontmatter" | grep 'confidence:' | head -1 | awk '{print $2}' || true)
    if [[ -n "$confidence" ]]; then
        case "$confidence" in
            high|medium|low) ;; # string confidence is valid
            *)
                if ! awk -v c="$confidence" 'BEGIN { exit !(c >= 0.0 && c <= 1.0) }' 2>/dev/null; then
                    echo "FAIL: $basename — confidence out of range: $confidence (must be 0.0-1.0 or high/medium/low)"
                    FAILURES=$((FAILURES + 1))
                    return
                fi
                ;;
        esac
    fi

    # 6. Extract body (after second ---) and require non-empty content.
    local body
    body=$(echo "$content" | awk 'BEGIN{n=0} /^---$/{n++; if(n==2){found=1; next}} found{print}')
    local body_words
    body_words=$(echo "$body" | wc -w | tr -d ' ')

    if [[ "$body_words" -eq 0 ]]; then
        echo "FAIL: $basename — frontmatter-only learning (missing body content)"
        FAILURES=$((FAILURES + 1))
        return
    fi

    if [[ -n "$body" ]]; then
        if echo "$body" | grep -qi "no significant learnings\|placeholder\|TODO: fill in\|template content"; then
            echo "FAIL: $basename — contains boilerplate/placeholder content"
            FAILURES=$((FAILURES + 1))
            return
        fi
    fi

    # 7. For manual learnings (has id: or date: but not utility:), require body >= 50 words
    local is_manual=false
    if echo "$frontmatter" | grep -q "id:\|date:\|source_epic:" && \
       ! echo "$frontmatter" | grep -q "utility:"; then
        is_manual=true
    fi

    if [[ "$is_manual" == "true" ]]; then
        if [[ "$body_words" -lt 50 ]]; then
            echo "WARN: $basename — manual learning with thin body ($body_words words)"
            WARNINGS=$((WARNINGS + 1))
            # Warn, don't fail — manual learnings may be concise
        fi
    fi

    log "PASS: $basename (fields=$recognized)"
}

# Main
if [[ ! -d "$LEARNINGS_DIR" ]]; then
    echo "No learnings directory found at $LEARNINGS_DIR"
    exit 0
fi

for file in "$LEARNINGS_DIR"/*.md; do
    [[ -f "$file" ]] || continue
    is_learning_artifact "$file" || continue
    check_file "$file"
done

echo ""
echo "Coherence check: $CHECKED files checked, $FAILURES failures, $WARNINGS warnings"

if [[ "$FAILURES" -gt 0 ]]; then
    echo "Coherence gate FAILED"
    exit 1
fi

echo "Coherence gate passed"
exit 0
