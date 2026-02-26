# beads

Operate issue workflows with Codex-friendly execution and clear dependency control.

## Codex Execution Profile

1. Treat `skills/beads/SKILL.md` as canonical issue-tracking contract.
2. Use beads ids and dependency edges as the single source of wave truth.
3. Keep issue updates tightly coupled to implementation state changes.

## Guardrails

1. Avoid runtime-specific assistant assumptions in operator instructions.
2. Ensure dependency direction is explicit whenever adding `bd dep` edges.
3. Require acceptance and validation details in issue descriptions for executable handoff.
