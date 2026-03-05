# Quarantined Tests

Tests in this directory require external runtimes (Claude Code, Codex CLI, OpenCode, OL) that are not available in CI. They are **not** run by `.github/workflows/validate.yml` or `tests/run-all.sh`.

## Why quarantined

These tests were never executed in CI and rotted silently. Moving them here makes the active test surface explicit while preserving the scripts for manual use.

## Running manually

```bash
# Individual test suites (require their respective runtimes)
bash tests/_quarantine/claude-code/run-all.sh
bash tests/_quarantine/codex/run-all.sh
bash tests/_quarantine/rpi-e2e/run-full-rpi.sh
bash tests/_quarantine/team-runner/run-all.sh
bash tests/_quarantine/skill-triggering/run-all.sh
```

## Promoting back to active

To move a test back to `tests/`:

1. Ensure it runs in CI (add to `validate.yml`)
2. Verify it passes without external runtimes OR add the runtime to CI
3. Move the file back: `git mv tests/_quarantine/<dir>/<file> tests/<dir>/`
