# rpi

Run the full RPI lifecycle in a Codex-native way: direct in-session orchestration, concise progress updates, and file-backed handoff between phases.


<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->
<!-- Generated from skills-codex-overrides/catalog.json for rpi. -->

## Codex Execution Profile

1. In Codex hookless mode, inspect `.agents/ao/codex/state.json` and ensure `ao codex start` once per thread before phase orchestration.
2. Resolve bead IDs before routing; do not infer epic scope from the `ag-*` prefix alone.
3. Keep a single `epic_id` spine across discovery, crank, and validation. Never replace it with a child issue ID from `bd ready`, `bd show`, or `.agents/rpi/next-work.jsonl`.
4. If `$crank` returns `<promise>PARTIAL</promise>`, rerun `$crank` on the same `epic_id` until the epic is done, blocked, or the retry budget is exhausted.
5. Orchestrate phases directly in the current session; do not hand RPI orchestration to wrapper commands.
6. claim, release, and consume semantics exactly
7. claim before work, consume on success, release on failure or interruption

## Guardrails

1. If the invocation resolves to standalone single-issue work with no parent epic, use `$implement` instead of pretending `$rpi` is epic orchestration.

<!-- END AGENTOPS OPERATOR CONTRACT -->
