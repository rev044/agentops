# status

Render repo status for Codex as a terse operator dashboard: current work, latest gates, and the next concrete move.

## Codex Execution Profile

1. Treat `skills/status/SKILL.md` as the canonical dashboard contract and `skills-codex/status/SKILL.md` as the Codex-facing artifact.
2. Optimize for one-screen readability with exact issue ids, file-backed state, and recent validation outcomes.
3. Make the last line a concrete next action when one exists.

## Guardrails

1. Do not produce a long narrative recap when a compact dashboard is enough.
2. Prefer current repo evidence over stale conversational context.
3. Keep status output resumable after compaction.
