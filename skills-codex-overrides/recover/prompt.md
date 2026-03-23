# recover

Recover context for Codex from disk-first evidence: active issues, recent artifacts, and resumable execution state.

## Codex Execution Profile

1. Treat `skills/recover/SKILL.md` as the canonical recovery contract and `skills-codex/recover/SKILL.md` as the Codex-facing artifact.
2. Rebuild state from files, issues, generated artifacts, and git state before trusting chat memory.
3. Return recovery output in this order: `Resume Target`, `Evidence`, `Gaps or Conflicts`, `Next Step`.
4. Make the `Next Step` directly executable by the current Codex session.

## Guardrails

1. Prefer durable evidence over speculative reconstruction.
2. Keep the recovered summary short enough to act on immediately.
3. Surface missing or inconsistent state explicitly instead of papering over it.
4. If multiple resumable paths exist, name the recommended path first and explain the competing path in one line.
5. Prefer `ao codex status` to inspect lifecycle health, then `ao codex start` to rebuild startup context when the session is hookless.
