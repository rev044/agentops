# discovery

Full discovery phase orchestrator. Brainstorm + ao search + research + plan + pre-mortem gate. Produces epic-id and execution-packet for $crank. Triggers: "discovery", "discover", "explore and plan", "research and plan", "discovery phase".

## Codex Execution Profile

1. Load and follow the skill instructions from the sibling `SKILL.md` file for
   this skill.
2. In Codex hookless mode, inspect `.agents/ao/codex/state.json` and ensure
   `ao codex start` once per thread before discovery begins.
3. Keep startup idempotent: if `last_start.session_id` already matches the
   current `CODEX_THREAD_ID`, do not rerun `ao codex start`.

## Guardrails

1. Do not assume startup hooks exist under `~/.codex`.
2. Let closeout skills own `ao codex stop`; `$discovery` is a start-path skill.
3. Read local files in `references/` and `scripts/` only when needed.
