# Artifact Consistency Validation

Cross-reference validation: scan knowledge artifacts for broken internal references.

```bash
# Preferred: run the helper script (handles fenced code blocks + placeholders).
skills/flywheel/scripts/artifact-consistency.sh

# Optional: include each broken reference for cleanup work.
skills/flywheel/scripts/artifact-consistency.sh --verbose
```

The helper script:
- Scans `.agents/**/*.md` excluding `.agents/ao/*`
- Ignores fenced code blocks
- Extracts references to `.agents/...(.md|.json|.jsonl)`
- Skips template placeholders (`YYYY`, `<...>`, `{...}`, wildcards)
- Reports `TOTAL_REFS`, `BROKEN_REFS`, `CONSISTENCY`, `STATUS`
- With `--verbose`, emits `BROKEN_REF=<source> -> <target>` lines

## Health Indicator

| Consistency | Status |
|-------------|--------|
| >90% | Healthy |
| 70-90% | Warning |
| <70% | Critical |
