# swarm

Orchestrate parallel work with Codex sub-agents and deterministic wave control.


<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->
<!-- Generated from skills-codex-overrides/catalog.json for swarm. -->

## Codex Execution Profile

1. Assign explicit ownership per worker before spawning: issue id, file set, and expected output.
2. Use file-backed result handoff under `.agents/swarm/` for consolidation and deterministic merge order.

## Guardrails

1. Do not give two workers overlapping write ownership in the same wave unless the merge plan is explicit.

<!-- END AGENTOPS OPERATOR CONTRACT -->
