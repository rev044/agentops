# rpi

Run the full RPI lifecycle in a Codex-native way: direct in-session orchestration, concise progress updates, and file-backed handoff between phases.

## Codex Execution Profile

1. Treat `skills/rpi/SKILL.md` as the canonical lifecycle contract and `skills-codex/rpi/SKILL.md` as the Codex-facing artifact.
2. Orchestrate phases directly in the current session; do not hand RPI orchestration to wrapper commands.
3. Prefer Codex sub-agents only for bounded sidecar work inside a phase, not for the lead orchestration path.
4. Re-read `.agents/rpi/next-work.jsonl` after each cycle and honor claim, release, and consume semantics exactly.

## Guardrails

1. Keep commentary updates short and operational; report phase transitions, blockers, and validation outcomes.
2. Preserve queue correctness: claim before work, consume on success, release on failure or interruption.
3. Treat harvested work as durable state on disk, not ephemeral chat context.
4. If a Codex-native override and the source skill diverge, keep behavior aligned with the source contract and then update the override.
