#!/usr/bin/env bash
# Detect project type: CODING, INFORMATIONAL, or OPS
# Usage: detect-project.sh [repo-root]
# Output: JSON with type, signals, and doc directories

set -euo pipefail

REPO_ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
REPO_NAME=$(basename "$REPO_ROOT")

# Initialize scores
CODING_SCORE=0
INFORMATIONAL_SCORE=0
OPS_SCORE=0

# Detect documentation directories
DOC_DIRS=()
for pattern in "docs/code-map" "docs/api" "docs/architecture" "docs" "doc" "documentation"; do
    if [[ -d "$REPO_ROOT/$pattern" ]]; then
        count=$(find "$REPO_ROOT/$pattern" -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
        DOC_DIRS+=("$pattern:$count")
    fi
done

# Count source files (excluding venv, node_modules, etc.)
py_count=$(find "$REPO_ROOT" -name "*.py" -type f 2>/dev/null | grep -v __pycache__ | grep -v .venv | grep -v venv/ | wc -l | tr -d ' ')
ts_count=$(find "$REPO_ROOT" \( -name "*.ts" -o -name "*.tsx" \) -type f 2>/dev/null | grep -v node_modules | wc -l | tr -d ' ')
md_count=$(find "$REPO_ROOT" -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
helm_count=$(find "$REPO_ROOT" -name "values.yaml" -type f 2>/dev/null | wc -l | tr -d ' ')

# CODING signals
[[ -d "$REPO_ROOT/services" ]] && CODING_SCORE=$((CODING_SCORE + 3))
[[ -d "$REPO_ROOT/src" ]] && CODING_SCORE=$((CODING_SCORE + 2))
[[ -f "$REPO_ROOT/pyproject.toml" || -f "$REPO_ROOT/package.json" ]] && CODING_SCORE=$((CODING_SCORE + 2))
[[ -d "$REPO_ROOT/docs/code-map" ]] && CODING_SCORE=$((CODING_SCORE + 3))
[[ $py_count -gt 50 || $ts_count -gt 50 ]] && CODING_SCORE=$((CODING_SCORE + 2))

# INFORMATIONAL signals
[[ -d "$REPO_ROOT/docs/corpus" ]] && INFORMATIONAL_SCORE=$((INFORMATIONAL_SCORE + 3))
[[ -d "$REPO_ROOT/docs/standards" ]] && INFORMATIONAL_SCORE=$((INFORMATIONAL_SCORE + 2))
[[ $md_count -gt 100 ]] && INFORMATIONAL_SCORE=$((INFORMATIONAL_SCORE + 3))
[[ ! -d "$REPO_ROOT/services" && ! -d "$REPO_ROOT/src" ]] && INFORMATIONAL_SCORE=$((INFORMATIONAL_SCORE + 2))
[[ -d "$REPO_ROOT/docs/tutorials" || -d "$REPO_ROOT/docs/how-to" ]] && INFORMATIONAL_SCORE=$((INFORMATIONAL_SCORE + 2))

# OPS signals
[[ -d "$REPO_ROOT/charts" ]] && OPS_SCORE=$((OPS_SCORE + 3))
[[ -d "$REPO_ROOT/apps" || -d "$REPO_ROOT/applications" ]] && OPS_SCORE=$((OPS_SCORE + 2))
[[ $helm_count -gt 5 ]] && OPS_SCORE=$((OPS_SCORE + 3))
find "$REPO_ROOT" -name "config.env" -type f 2>/dev/null | grep -q . && OPS_SCORE=$((OPS_SCORE + 2))

# Determine type (highest score wins, tie-breaker: CODING > OPS > INFORMATIONAL)
if [[ $CODING_SCORE -ge $OPS_SCORE && $CODING_SCORE -ge $INFORMATIONAL_SCORE ]]; then
    TYPE="CODING"
    SCORE=$CODING_SCORE
elif [[ $OPS_SCORE -ge $INFORMATIONAL_SCORE ]]; then
    TYPE="OPS"
    SCORE=$OPS_SCORE
else
    TYPE="INFORMATIONAL"
    SCORE=$INFORMATIONAL_SCORE
fi

# Output JSON
cat <<EOF
{
  "repo": "$REPO_NAME",
  "type": "$TYPE",
  "confidence": $SCORE,
  "scores": {
    "coding": $CODING_SCORE,
    "informational": $INFORMATIONAL_SCORE,
    "ops": $OPS_SCORE
  },
  "sources": {
    "python": $py_count,
    "typescript": $ts_count,
    "markdown": $md_count,
    "helm": $helm_count
  },
  "doc_dirs": [$(printf '"%s",' "${DOC_DIRS[@]}" | sed 's/,$//')]
}
EOF
