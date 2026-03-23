# rpi

Run the full RPI lifecycle in a Codex-native way: direct in-session orchestration, concise progress updates, and file-backed handoff between phases.


<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->
<!-- Generated from skills-codex-overrides/catalog.json for rpi. -->

## Codex Execution Profile

1. In Codex hookless mode, inspect `.agents/ao/codex/state.json` and ensure `ao codex start` once per thread before phase orchestration.
2. Orchestrate phases directly in the current session; do not hand RPI orchestration to wrapper commands.
3. claim, release, and consume semantics exactly

## Guardrails

1. claim before work, consume on success, release on failure or interruption

<!-- END AGENTOPS OPERATOR CONTRACT -->
