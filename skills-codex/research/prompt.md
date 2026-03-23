# research

Run repository research with Codex-native agents and concise artifact output.


<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->
<!-- Generated from skills-codex-overrides/catalog.json for research. -->

## Codex Execution Profile

1. In Codex hookless mode, inspect `.agents/ao/codex/state.json` and ensure `ao codex start` once per thread before research.
2. Prefer `spawn_agent` / `send_input` / `wait` for parallel exploration.
3. Write findings to `.agents/research/` with file-level references and concrete evidence.

## Guardrails

1. Keep backend fallback logic explicit: codex sub-agents, then inline.

<!-- END AGENTOPS OPERATOR CONTRACT -->
