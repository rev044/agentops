# swarm

Orchestrate parallel work with Codex sub-agents and deterministic wave control.

## Codex Execution Profile

1. Treat `skills/swarm/SKILL.md` as canonical swarm contract.
2. Prefer Codex sub-agents for wave execution and result collection.
3. Use file-backed result handoff under `.agents/swarm/` for consolidation.

## Guardrails

1. Keep runtime fallback notes short and avoid Claude-team-first language.
2. Preserve wave dependency integrity before spawning workers.
3. Require each worker output to include status, changed files, and blockers.
