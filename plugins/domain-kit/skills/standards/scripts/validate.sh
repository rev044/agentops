#!/bin/bash
# Validate standards skill
set -euo pipefail

# Determine SKILL_DIR relative to this script (works in plugins or ~/.claude)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC2034  # Reserved for future use
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

ERRORS=0
CHECKS=0

check_exists() {
    local desc="$1"
    local path="$2"

    CHECKS=$((CHECKS + 1))
    if [ -e "$path" ]; then
        echo "✓ $desc"
    else
        echo "✗ $desc ($path not found)"
        ERRORS=$((ERRORS + 1))
    fi
}

echo "=== Standards Skill Validation ==="
echo ""


# Verify all documented language references exist
check_exists "Python standard" "$SKILL_DIR/references/python.md"
check_exists "Go standard" "$SKILL_DIR/references/go.md"
check_exists "TypeScript standard" "$SKILL_DIR/references/typescript.md"
check_exists "Shell standard" "$SKILL_DIR/references/shell.md"
check_exists "YAML standard" "$SKILL_DIR/references/yaml.md"
check_exists "Markdown standard" "$SKILL_DIR/references/markdown.md"
check_exists "JSON standard" "$SKILL_DIR/references/json.md"
check_exists "Tags standard" "$SKILL_DIR/references/tags.md"

# Verify OpenAI standards exist
check_exists "OpenAI overview" "$SKILL_DIR/references/openai.md"
check_exists "OpenAI prompts" "$SKILL_DIR/references/openai-prompts.md"
check_exists "OpenAI functions" "$SKILL_DIR/references/openai-functions.md"
check_exists "OpenAI responses" "$SKILL_DIR/references/openai-responses.md"
check_exists "OpenAI reasoning" "$SKILL_DIR/references/openai-reasoning.md"
check_exists "OpenAI GPT-OSS" "$SKILL_DIR/references/openai-gptoss.md"

# Verify each reference has required sections
for ref in python go typescript shell yaml markdown json; do
    ref_path="$SKILL_DIR/references/$ref.md"
    if [ -f "$ref_path" ]; then
        CHECKS=$((CHECKS + 1))
        if grep -q "Common Errors\|AI Agent Guidelines\|Anti-Pattern" "$ref_path" 2>/dev/null; then
            echo "✓ $ref.md has standard sections"
        else
            echo "✗ $ref.md missing standard sections (Common Errors, Anti-Patterns, or AI Agent Guidelines)"
            ERRORS=$((ERRORS + 1))
        fi
    fi
done

# Verify OpenAI references have required sections
for ref in openai openai-prompts openai-functions openai-responses openai-reasoning openai-gptoss; do
    ref_path="$SKILL_DIR/references/$ref.md"
    if [ -f "$ref_path" ]; then
        CHECKS=$((CHECKS + 1))
        if grep -q "Common Errors\|AI Agent Guidelines\|Anti-Pattern" "$ref_path" 2>/dev/null; then
            echo "✓ $ref.md has standard sections"
        else
            echo "✗ $ref.md missing standard sections (Common Errors, Anti-Patterns, or AI Agent Guidelines)"
            ERRORS=$((ERRORS + 1))
        fi
    fi
done

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Standards skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Standards skill validation passed"
    exit 0
fi
