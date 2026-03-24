# rpi

Run the full RPI lifecycle in a Codex-native way: direct in-session orchestration, concise progress updates, and file-backed handoff between phases.

## Codex Execution Profile

1. Treat `skills/rpi/SKILL.md` as the canonical lifecycle contract and `skills-codex/rpi/SKILL.md` as the Codex-facing artifact.
2. When beads are present, resolve bead IDs before routing; when beads are absent, preserve the current goal or execution-packet objective across phases.
3. Keep a single lifecycle objective spine across discovery, crank, and validation. Never replace it with a child issue ID or one ready slice from `bd ready`, `bd show`, or `.agents/rpi/next-work.jsonl`.
4. If discovery does not yield an epic id, invoke `$crank .agents/rpi/execution-packet.json` and standalone `$validation` instead of inventing one.
5. If `$crank` returns `<promise>PARTIAL</promise>`, rerun `$crank` on the same lifecycle objective until the work is done, blocked, or the retry budget is exhausted.
6. Orchestrate phases directly in the current session; do not hand RPI orchestration to wrapper commands.
7. Prefer Codex sub-agents only for bounded sidecar work inside a phase, not for the lead orchestration path.
8. Re-read `.agents/rpi/next-work.jsonl` after each cycle and honor claim, release, and consume semantics exactly.

## Guardrails

1. Keep commentary updates short and operational; report phase transitions, blockers, and validation outcomes.
2. Preserve queue correctness: claim before work, consume on success, release on failure or interruption.
3. Treat harvested work as durable state on disk, not ephemeral chat context.
4. Do not stop after a partial phase result; only stop on `<promise>BLOCKED</promise>`, retry-budget exhaustion, or final completion.
5. If a Codex-native override and the source skill diverge, keep behavior aligned with the source contract and then update the override.
