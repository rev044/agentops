# crank

Execute epics hands-free with Codex-native wave progression.

## Codex Execution Profile

1. Treat `skills/crank/SKILL.md` as canonical execution contract.
2. Run waves from beads dependencies and process only currently unblocked issues.
3. Keep retries bounded and report blockers with exact issue ids.

## Guardrails

1. Prefer Codex sub-agents through `$swarm` for parallel issue execution.
2. Do not blur done/partial/blocked status boundaries.
3. Include validation metadata checks in worker instructions when available.
