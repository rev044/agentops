# Artifact Consistency Validation

Cross-reference validation: scan knowledge artifacts for broken internal references.

```bash
# Scan .agents/ markdown files for references to other .agents/ files
TOTAL_REFS=0
BROKEN_REFS=0
BROKEN_LIST=""

for file in $(find .agents/ -name "*.md" -not -path ".agents/ao/*" 2>/dev/null); do
  # Find references to .agents/ paths
  refs=$(grep -oE '\.agents/[a-zA-Z0-9/_-]+\.(md|json|jsonl)' "$file" 2>/dev/null || true)
  for ref in $refs; do
    TOTAL_REFS=$((TOTAL_REFS + 1))
    if [ ! -f "$ref" ]; then
      BROKEN_REFS=$((BROKEN_REFS + 1))
      BROKEN_LIST="${BROKEN_LIST}\n  - $file -> $ref"
    fi
  done
done

# Compute consistency score
if [ "$TOTAL_REFS" -gt 0 ]; then
  CONSISTENCY=$(( (TOTAL_REFS - BROKEN_REFS) * 100 / TOTAL_REFS ))
else
  CONSISTENCY=100
fi

echo "Artifact Consistency: ${CONSISTENCY}% ($BROKEN_REFS broken of $TOTAL_REFS references)"
```

## Health Indicator

| Consistency | Status |
|-------------|--------|
| >90% | Healthy |
| 70-90% | Warning |
| <70% | Critical |
