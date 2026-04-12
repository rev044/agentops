#!/usr/bin/env bash
# AgentOps Hook: constraint-compiler
# Compiles high-scoring learnings tagged "constraint" or "anti-pattern" into
# template shell hooks under .agents/constraints/.
#
# Usage: bash hooks/constraint-compiler.sh <learning-path>
set -euo pipefail

##############################################################################
# Argument validation
##############################################################################

if [ $# -lt 1 ]; then
    echo "Usage: constraint-compiler.sh <learning-path>" >&2
    exit 1
fi

LEARNING_PATH="$1"

if [ ! -f "$LEARNING_PATH" ]; then
    echo "ERROR: Learning file not found: $LEARNING_PATH" >&2
    exit 1
fi

##############################################################################
# Repository root
##############################################################################

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || echo ".")"
CONSTRAINT_DIR="$ROOT/.agents/constraints"
INDEX_FILE="$CONSTRAINT_DIR/index.json"

mkdir -p "$CONSTRAINT_DIR"

##############################################################################
# Parse YAML frontmatter from learning file
##############################################################################

# Extract the frontmatter block (between --- delimiters)
IN_FRONTMATTER=0
FRONTMATTER=""
BODY=""
PAST_FRONTMATTER=0

while IFS= read -r line; do
    if [ "$PAST_FRONTMATTER" -eq 1 ]; then
        BODY="${BODY}${line}
"
        continue
    fi
    if [ "$line" = "---" ]; then
        if [ "$IN_FRONTMATTER" -eq 0 ]; then
            IN_FRONTMATTER=1
            continue
        else
            PAST_FRONTMATTER=1
            continue
        fi
    fi
    if [ "$IN_FRONTMATTER" -eq 1 ]; then
        FRONTMATTER="${FRONTMATTER}${line}
"
    fi
done < "$LEARNING_PATH"

##############################################################################
# Extract fields from frontmatter (portable grep/sed, no jq for YAML)
##############################################################################

extract_field() {
    local field="$1"
    local source="$2"
    printf '%s' "$source" | grep "^${field}:" | head -1 | sed "s/^${field}:[[:space:]]*//" | sed 's/^["'"'"']//;s/["'"'"']$//' | sed 's/[[:space:]]*$//'
}

TITLE="$(extract_field "title" "$FRONTMATTER")"
LEARNING_ID="$(extract_field "id" "$FRONTMATTER")"
DATE_FIELD="$(extract_field "date" "$FRONTMATTER")"
TAGS_LINE="$(printf '%s' "$FRONTMATTER" | grep "^tags:" | head -1 | sed 's/^tags:[[:space:]]*//')"

# Fallback: derive title from first heading if not in frontmatter
if [ -z "$TITLE" ]; then
    TITLE="$(printf '%s' "$BODY" | grep "^# " | head -1 | sed 's/^# //')"
fi

# Fallback: derive ID from filename if not in frontmatter
if [ -z "$LEARNING_ID" ]; then
    LEARNING_ID="$(basename "$LEARNING_PATH" .md | sed 's/[^a-zA-Z0-9_-]/-/g')"
fi

# Fallback: use today for date
if [ -z "$DATE_FIELD" ]; then
    DATE_FIELD="$(date -u +%Y-%m-%d)"
fi

# Validate that tags include "constraint" or "anti-pattern"
HAS_CONSTRAINT_TAG=0
case "$TAGS_LINE" in
    *constraint*) HAS_CONSTRAINT_TAG=1 ;;
esac
case "$TAGS_LINE" in
    *anti-pattern*) HAS_CONSTRAINT_TAG=1 ;;
esac

if [ "$HAS_CONSTRAINT_TAG" -eq 0 ]; then
    echo "SKIP: Learning '$LEARNING_ID' not tagged 'constraint' or 'anti-pattern'" >&2
    exit 0
fi

##############################################################################
# Build content summary (first non-empty paragraph from body, max 200 chars)
##############################################################################

SUMMARY=""
IN_PARAGRAPH=0
while IFS= read -r line; do
    # Skip headings and blank lines to find first paragraph
    case "$line" in
        "#"*) continue ;;
        "")
            if [ "$IN_PARAGRAPH" -eq 1 ]; then
                break
            fi
            continue
            ;;
        *)
            IN_PARAGRAPH=1
            if [ -n "$SUMMARY" ]; then
                SUMMARY="${SUMMARY} ${line}"
            else
                SUMMARY="$line"
            fi
            ;;
    esac
done <<EOF
$BODY
EOF

# Truncate summary to 200 chars
if [ "${#SUMMARY}" -gt 200 ]; then
    SUMMARY="$(printf '%.200s' "$SUMMARY")..."
fi

# Content excerpt for the TODO (first 120 chars of summary)
EXCERPT="$SUMMARY"
if [ "${#EXCERPT}" -gt 120 ]; then
    EXCERPT="$(printf '%.120s' "$EXCERPT")..."
fi

##############################################################################
# Generate constraint template
##############################################################################

CONSTRAINT_FILE="$CONSTRAINT_DIR/${LEARNING_ID}.sh"
COMPILED_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

cat > "$CONSTRAINT_FILE" <<TEMPLATE
#!/bin/bash
# Constraint: ${TITLE}
# Source: ${LEARNING_ID}
# Compiled: ${COMPILED_DATE}
# Status: draft
# Description: ${SUMMARY}
# TODO: Replace this placeholder with a real detection pattern.
# The learning says: "${EXCERPT}"
if false; then  # <-- replace with detection pattern
    echo "CONSTRAINT VIOLATION: ${TITLE}"
    exit 1
fi
TEMPLATE

chmod +x "$CONSTRAINT_FILE"
echo "Generated constraint template: $CONSTRAINT_FILE"

##############################################################################
# Update index.json (create or merge)
##############################################################################

json_escape() {
    local s="$1"
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    s="${s//$'\n'/\\n}"
    s="${s//$'\r'/\\r}"
    s="${s//$'\t'/\\t}"
    printf '%s' "$s"
}

NEW_ENTRY="$(printf '{"id":"%s","title":"%s","source":"%s","status":"draft","compiled_at":"%s","file":"%s"}' \
    "$(json_escape "$LEARNING_ID")" \
    "$(json_escape "$TITLE")" \
    "$(json_escape "$LEARNING_PATH")" \
    "$(json_escape "$COMPILED_DATE")" \
    "$(json_escape ".agents/constraints/${LEARNING_ID}.sh")")"
PENDING_INDEX_FILE="$CONSTRAINT_DIR/index.pending.jsonl"

if [ -f "$INDEX_FILE" ]; then
    # Check if jq is available for safe JSON manipulation
    if command -v jq >/dev/null 2>&1; then
        # Remove existing entry with same ID, then append new one
        UPDATED="$(jq --arg id "$LEARNING_ID" \
                      --arg title "$TITLE" \
                      --arg source "$LEARNING_PATH" \
                      --arg compiled "$COMPILED_DATE" \
                      --arg file ".agents/constraints/${LEARNING_ID}.sh" '
            .schema_version = (.schema_version // 1)
            | .constraints = ((.constraints // []) | map(select(.id != $id)) + [{
                id: $id,
                title: $title,
                source: $source,
                status: "draft",
                compiled_at: $compiled,
                file: $file
            }])
        ' "$INDEX_FILE")"
        printf '%s\n' "$UPDATED" > "$INDEX_FILE"
    else
        # Non-destructive fallback: keep existing index and queue a pending update.
        printf '%s\n' "$NEW_ENTRY" >> "$PENDING_INDEX_FILE"
        echo "WARNING: jq not found, preserving existing index and queueing update in $PENDING_INDEX_FILE" >&2
    fi
else
    # Create fresh index
    if command -v jq >/dev/null 2>&1; then
        jq -n --arg id "$LEARNING_ID" \
              --arg title "$TITLE" \
              --arg source "$LEARNING_PATH" \
              --arg compiled "$COMPILED_DATE" \
              --arg file ".agents/constraints/${LEARNING_ID}.sh" '
            {
              schema_version: 1,
              constraints: [{
                id: $id,
                title: $title,
                source: $source,
                status: "draft",
                compiled_at: $compiled,
                file: $file
              }]
            }
        ' > "$INDEX_FILE"
    else
        printf '{"schema_version":1,"constraints":[%s]}\n' "$NEW_ENTRY" > "$INDEX_FILE"
    fi
fi

echo "Updated index: $INDEX_FILE"
