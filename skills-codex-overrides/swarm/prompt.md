# swarm

Orchestrate parallel work with Codex sub-agents and deterministic wave control.

## Codex Execution Profile

1. Treat `skills/swarm/SKILL.md` as the canonical swarm contract and this override as the Codex operator layer.
2. Prefer Codex sub-agents for wave execution and result collection.
3. Assign explicit ownership per worker before spawning: issue id, file set, and expected output.
4. Use file-backed result handoff under `.agents/swarm/` for consolidation and deterministic merge order.
5. Require every worker report to end with `status`, `changed files`, `blockers`, and `next action`.

## Guardrails

1. Keep runtime fallback notes short and avoid Claude-team-first language.
2. Preserve wave dependency integrity before spawning workers.
3. Do not give two workers overlapping write ownership in the same wave unless the merge plan is explicit.
4. If a worker returns partial results, convert that into a concrete follow-up instead of silently absorbing it.
