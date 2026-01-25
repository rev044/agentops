#!/usr/bin/env bash
# Discover documentable features in a codebase
# Usage: discover-features.sh <type> [repo-root]
#   type: CODING | INFORMATIONAL | OPS
# Output: JSON list of discovered features

set -euo pipefail

TYPE="${1:-CODING}"
REPO_ROOT="${2:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"

discover_coding() {
    # Find Python/TypeScript modules with documentable content
    echo "["
    first=true

    # Python services
    while IFS= read -r dir; do
        name=$(basename "$dir")
        score=0
        sources=()
        endpoints=()

        # Check for main.py or __init__.py
        [[ -f "$dir/main.py" ]] && sources+=("$dir/main.py") && score=$((score + 1))
        [[ -f "$dir/__init__.py" ]] && sources+=("$dir/__init__.py")

        # Check for API endpoints
        if grep -rq "@app\.\(get\|post\|put\|delete\)\|@router\." "$dir" 2>/dev/null; then
            score=$((score + 2))
            mapfile -t endpoints < <(grep -rh "@app\.\(get\|post\|put\|delete\)\|@router\." "$dir" 2>/dev/null | head -5 | sed 's/.*"\([^"]*\)".*/\1/' || true)
        fi

        # Check for Prometheus metrics
        grep -rq "Counter(\|Gauge(\|Histogram(" "$dir" 2>/dev/null && score=$((score + 1))

        # Check for config vars
        grep -rq "os.getenv\|os.environ" "$dir" 2>/dev/null && score=$((score + 1))

        # Check for existing docs
        doc_path="docs/code-map/${name}.md"
        documented="false"
        [[ -f "$REPO_ROOT/$doc_path" ]] && documented="true"

        # Only include if score >= 3
        if [[ $score -ge 3 ]]; then
            $first || echo ","
            first=false
            cat <<EOF
  {
    "name": "$name",
    "type": "service",
    "score": $score,
    "documented": $documented,
    "doc_path": "$doc_path",
    "sources": $(printf '%s\n' "${sources[@]:-}" | jq -R . | jq -s .),
    "endpoints": $(printf '%s\n' "${endpoints[@]:-}" | jq -R . | jq -s .)
  }
EOF
        fi
    done < <(find "$REPO_ROOT/services" -maxdepth 1 -type d 2>/dev/null | tail -n +2)

    echo "]"
}

discover_informational() {
    # Find corpus sections and validate structure
    echo "["
    first=true

    while IFS= read -r dir; do
        name=$(basename "$dir")
        [[ "$name" == "docs" ]] && continue

        md_count=$(find "$dir" -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
        has_readme="false"
        [[ -f "$dir/README.md" ]] && has_readme="true"

        if [[ $md_count -gt 0 ]]; then
            $first || echo ","
            first=false
            cat <<EOF
  {
    "name": "$name",
    "type": "section",
    "files": $md_count,
    "has_readme": $has_readme,
    "path": "${dir#"$REPO_ROOT"/}"
  }
EOF
        fi
    done < <(find "$REPO_ROOT/docs" -maxdepth 2 -type d 2>/dev/null)

    echo "]"
}

discover_ops() {
    # Find Helm charts and config files
    echo "["
    first=true

    # Helm charts
    while IFS= read -r chart; do
        [[ -z "$chart" ]] && continue
        dir=$(dirname "$chart")
        name=$(basename "$dir")

        values_count=$(find "$dir" -name "values*.yaml" -type f 2>/dev/null | wc -l | tr -d ' ')
        has_runbook="false"
        [[ -f "$dir/RUNBOOK.md" || -f "$dir/docs/runbook.md" ]] && has_runbook="true"

        $first || echo ","
        first=false
        cat <<EOF
  {
    "name": "$name",
    "type": "helm-chart",
    "values_files": $values_count,
    "has_runbook": $has_runbook,
    "path": "${dir#"$REPO_ROOT"/}"
  }
EOF
    done < <(find "$REPO_ROOT" -name "Chart.yaml" -type f 2>/dev/null)

    echo "]"
}

case "$TYPE" in
    CODING) discover_coding ;;
    INFORMATIONAL) discover_informational ;;
    OPS) discover_ops ;;
    *) echo "Unknown type: $TYPE" >&2; exit 1 ;;
esac
