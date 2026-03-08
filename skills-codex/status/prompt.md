# status

Render repo status for Codex as a terse operator dashboard: active work, latest gates, and the next concrete move.

## Codex Execution Profile

1. Treat `skills/status/SKILL.md` as the canonical dashboard contract and `skills-codex/status/SKILL.md` as the Codex-facing artifact.
2. Default to a one-screen layout with three blocks in this order: `Current Work`, `Latest Gates`, `Next Action`.
3. Use exact issue ids, branch/worktree state, and file-backed artifacts instead of conversational summaries.
4. Make the last line a concrete next action when one exists.

## Guardrails

1. Do not expand into a long narrative when a compact dashboard answers the question.
2. Prefer current repo evidence over stale conversational context.
3. If state is missing or contradictory, say which file or command is missing instead of guessing.
4. Keep the output resumable after compaction.
