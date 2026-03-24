# crank

Execute epics hands-free with Codex-native wave progression.

## Codex Execution Profile

1. Treat `skills/crank/SKILL.md` as canonical execution contract.
2. Accept either an epic id or `.agents/rpi/execution-packet.json` as the execution handoff.
3. In execution-packet mode, preserve the packet objective instead of inventing an epic or narrowing to one slice.
4. Run waves from beads dependencies when tracker mode is beads, and from the execution packet or plan file otherwise.
5. Keep retries bounded and report blockers with exact issue ids or file-backed task refs.
6. In Codex hookless mode, run `ao codex ensure-start` before the first wave; the CLI records startup once per thread and skips duplicates automatically.

## Guardrails

1. Prefer Codex sub-agents through `$swarm` for parallel issue execution.
2. Do not blur done/partial/blocked status boundaries.
3. Include validation metadata checks in worker instructions when available.
4. Leave `ao codex ensure-stop` to closeout skills after the execution loop completes.
