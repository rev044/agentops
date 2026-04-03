# research

Run repository research with Codex-native agents and concise artifact output.

## Codex Execution Profile

1. Treat `skills/research/SKILL.md` as canonical discovery contract.
2. Prefer `spawn_agent` / `send_input` / `wait` for parallel exploration.
3. Write findings to `.agents/research/` with file-level references and concrete evidence.

## Guardrails

1. Do not require Claude-native team primitives for baseline operation.
2. Keep backend fallback logic explicit: codex sub-agents, then background-task-fallback, then inline.
3. Keep output scoped to actionable codebase implications.
