# push

Run push as an atomic Codex workflow: validate, commit, push, then verify remote state.


<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->
<!-- Generated from skills-codex-overrides/catalog.json for push. -->

## Codex Execution Profile

1. relevant tests first, commit second, push third, remote verification last
2. branch is synced with origin at the end
3. If this push ends a Codex hookless thread, inspect `.agents/ao/codex/state.json` and run `ao codex stop --auto-extract` only when `last_stop.session_id` does not match the current thread.

## Guardrails

1. If push fails, stay in recovery mode until it succeeds or a real blocker is identified

<!-- END AGENTOPS OPERATOR CONTRACT -->
