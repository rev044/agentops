# validation

Full validation phase orchestrator. Vibe + post-mortem + retro + forge. Reviews implementation quality, extracts learnings, feeds the knowledge flywheel. Triggers: "validation", "validate", "validate work", "review and learn", "validation phase", "post-implementation review".

## Codex Execution Profile

1. Load and follow the skill instructions from the sibling `SKILL.md` file for
   this skill.
2. In Codex hookless mode, standalone validation should inspect
   `.agents/ao/codex/state.json` and ensure `ao codex stop --auto-extract` once
   per thread when closeout has not already run.
3. Keep closeout idempotent: if `last_stop.session_id` already matches the
   current `CODEX_THREAD_ID`, do not rerun `ao codex stop`.

## Guardrails

1. Do not assume session-end hooks exist under `~/.codex`.
2. Let `$validation` own Codex closeout only for standalone validation; when it invokes `$post-mortem` for an epic, `$post-mortem` owns the closeout.
3. Read local files in `references/` and `scripts/` only when needed.
