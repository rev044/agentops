# status

Render repo status for Codex as a terse operator dashboard: active work, latest gates, and the next concrete move.


<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->
<!-- Generated from skills-codex-overrides/catalog.json for status. -->

## Codex Execution Profile

1. In Codex hookless mode, run `ao codex ensure-start` before gathering dashboard state; the CLI records startup once per thread and skips duplicates automatically.
2. Default to a one-screen layout with three blocks in this order: `Current Work`, `Latest Gates`, `Next Action`.
3. Use exact issue ids, branch/worktree state, and file-backed artifacts instead of conversational summaries.

## Guardrails

1. Keep the output resumable after compaction.

<!-- END AGENTOPS OPERATOR CONTRACT -->
