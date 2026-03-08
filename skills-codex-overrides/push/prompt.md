# push

Run push as an atomic Codex workflow: validate, commit, push, then verify remote state.

## Codex Execution Profile

1. Treat `skills/push/SKILL.md` as the canonical push contract and `skills-codex/push/SKILL.md` as the Codex-facing artifact.
2. Keep the gate order explicit: relevant tests first, commit second, push third, remote verification last.
3. Report the exact validations run and whether the branch is synced with origin at the end.

## Guardrails

1. Do not skip failing gates just to get a push out.
2. Keep commit scope aligned with the actual change set.
3. If push fails, stay in recovery mode until it succeeds or a real blocker is identified.
