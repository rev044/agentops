# codex-team

Coordinate multi-agent Codex implementation work with clear ownership boundaries.

## Codex Execution Profile

1. Treat `skills/codex-team/SKILL.md` as canonical team execution contract.
2. Assign explicit file ownership per worker.
3. Merge wave results in deterministic order after each completion batch.

## Guardrails

1. Keep conflict resolution explicit when workers touch adjacent code paths.
2. Require each worker to ignore unrelated edits made by others.
3. Keep status reporting compact: done, blocked, next action.
