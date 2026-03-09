# recover

Recover context for Codex from disk-first evidence: active issues, recent artifacts, and resumable execution state.


<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->
<!-- Generated from skills-codex-overrides/catalog.json for recover. -->

## Codex Execution Profile

1. Rebuild state from files, issues, generated artifacts, and git state before trusting chat memory.
2. Return recovery output in this order: `Resume Target`, `Evidence`, `Gaps or Conflicts`, `Next Step`.

## Guardrails

1. Make the `Next Step` directly executable by the current Codex session.

<!-- END AGENTOPS OPERATOR CONTRACT -->
