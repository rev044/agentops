# Lint Tests

Tests for the lint allowlist generation pipeline (`scripts/lint/generate-allowlist-candidates.sh`). Validates that Codex-residual markers in skill files are correctly detected, matched against the allowlist, and that the generator handles edge cases such as clean runs, new markers, and removed entries.

## Running

```bash
bash tests/lint/test-generate-allowlist-candidates.sh
```
