# codex-team

Coordinate multi-agent Codex implementation work with clear ownership boundaries.

## Codex Execution Profile

1. Treat `skills/codex-team/SKILL.md` as canonical team execution contract.
2. Assign explicit file ownership per worker before coding starts.
3. Require each worker to report `done`, `blocked`, `changed files`, and `next action`.
4. Merge wave results in deterministic order after each completion batch.
5. When a worker is blocked by another worker's output, pause that lane instead of creating speculative edits.

## Guardrails

1. Keep conflict resolution explicit when workers touch adjacent code paths.
2. Require each worker to ignore unrelated edits made by others.
3. Keep status reporting compact and operator-facing.
4. Summaries should identify which worker owns the next move, not just what happened.
