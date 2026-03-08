# recover

Recover context for Codex from disk-first evidence: active issues, recent artifacts, and resumable execution state.

## Codex Execution Profile

1. Treat `skills/recover/SKILL.md` as the canonical recovery contract and `skills-codex/recover/SKILL.md` as the Codex-facing artifact.
2. Rebuild state from files, issues, and generated artifacts before trusting chat memory.
3. Return the minimum context needed to resume work immediately.

## Guardrails

1. Prefer durable evidence over speculative reconstruction.
2. Keep the recovered summary short enough to act on.
3. Surface missing or inconsistent state explicitly instead of papering over it.
