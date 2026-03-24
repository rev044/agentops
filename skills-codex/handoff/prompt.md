# handoff

Create Codex-native handoffs that survive compaction: exact state, exact files, exact next step.


<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->
<!-- Generated from skills-codex-overrides/catalog.json for handoff. -->

## Codex Execution Profile

1. Capture the current objective, completed work, unresolved blockers, and the next command or file to inspect.
2. Prefer durable paths, issue ids, and validation evidence over conversational summaries.
3. Finish handoff-driven session closeout by running `ao codex ensure-stop --auto-extract`; the CLI already skips duplicate closeout for the same Codex thread.

## Guardrails

1. Do not leave the next session guessing what to do first.

<!-- END AGENTOPS OPERATOR CONTRACT -->
