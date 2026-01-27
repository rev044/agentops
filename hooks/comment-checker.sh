#!/usr/bin/env bash
# comment-checker.sh - Check comment density in modified files
# PostToolUse hook for Write/Edit operations
# Competitor adoption: Reminds agents to prefer self-documenting code

set -euo pipefail

# Read hook input from stdin
input=$(cat)

# Extract file path from tool result
# PostToolUse provides tool_input with file_path
FILE=$(echo "$input" | jq -r '.tool_input.file_path // .tool_input.path // ""' 2>/dev/null)

# Skip if no file path found
if [[ -z "$FILE" || "$FILE" == "null" ]]; then
    exit 0
fi

# Skip if file doesn't exist
if [[ ! -f "$FILE" ]]; then
    exit 0
fi

# Security: Verify file is within current project directory (prevent path traversal)
PROJ_DIR=$(pwd)
REAL_FILE=$(realpath "$FILE" 2>/dev/null || echo "")
if [[ -z "$REAL_FILE" || ! "$REAL_FILE" =~ ^"$PROJ_DIR" ]]; then
    # File is outside project directory - skip silently
    exit 0
fi

# Skip binary files
file_type=$(file --mime-type -b "$FILE" 2>/dev/null || echo "unknown")
if [[ ! "$file_type" =~ ^text/ ]]; then
    exit 0
fi

# Skip very small files (less than 10 lines)
total_lines=$(wc -l < "$FILE" 2>/dev/null | tr -d ' ')
if [[ "$total_lines" -lt 10 ]]; then
    exit 0
fi

# Detect file type for language-specific comment patterns
ext="${FILE##*.}"
case "$ext" in
    py)
        # Python: # comments only (skip shebang, docstrings counted separately)
        comment_pattern='^\s*#[^!]'
        ;;
    sh|bash|zsh)
        # Shell: # comments (skip shebang)
        comment_pattern='^\s*#[^!]'
        ;;
    go|js|ts|tsx|jsx|java|c|cpp|h|hpp|rs)
        # C-family: // single-line comments
        comment_pattern='^\s*//'
        ;;
    md|markdown)
        # Markdown: skip entirely (all content is "documentation")
        exit 0
        ;;
    *)
        # Fallback: conservative pattern (# or // at line start only)
        comment_pattern='^\s*(#[^!]|//)'
        ;;
esac

# Count comment lines using language-specific pattern
# Note: This is a heuristic - may have false positives in strings
comment_lines=$(grep -cE "$comment_pattern" "$FILE" 2>/dev/null || echo "0")

# Handle edge case: all comments (docstring file, license header)
if [[ "$comment_lines" -eq "$total_lines" ]]; then
    exit 0
fi

# Handle edge case: no comments at all (that's fine)
if [[ "$comment_lines" -eq 0 ]]; then
    exit 0
fi

# Calculate density percentage
density=$((comment_lines * 100 / total_lines))

# Threshold: >15% triggers gentle reminder
if [[ "$density" -gt 15 ]]; then
    # Output reminder to stderr (informational, doesn't block)
    echo "# Comment density: ${density}% (${comment_lines}/${total_lines} lines) in ${FILE##*/}" >&2
    echo "# Prefer self-documenting code over comments. Remove obvious comments." >&2
fi

exit 0
