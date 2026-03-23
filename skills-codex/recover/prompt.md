# recover

Recover context for Codex from disk-first evidence: active issues, recent artifacts, and resumable execution state.


<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->
<!-- Generated from skills-codex-overrides/catalog.json for recover. -->

## Codex Execution Profile

1. Treat `skills/recover/SKILL.md` as the canonical recovery contract and `skills-codex/recover/SKILL.md` as the Codex-facing artifact.
2. Rebuild state from files, issues, generated artifacts, and git state before trusting chat memory.
3. Return recovery output in this order: `Resume Target`, `Evidence`, `Gaps or Conflicts`, `Next Step`.
4. In Codex hookless mode, inspect `.agents/ao/codex/state.json` and run `ao codex start` only when `last_start.session_id` does not match the current thread.

## Guardrails

1. Make the `Next Step` directly executable by the current Codex session.

<!-- END AGENTOPS OPERATOR CONTRACT -->
