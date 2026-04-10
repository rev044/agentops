---
name: dream
description: 'Private overnight operator mode. Routes interactive Dream requests to the shared `ao overnight` engine for setup, bedtime runs, and morning reports. Triggers: "$dream", "overnight", "bedtime run", "morning report", "dream setup", "dream report", "dream council", "dreamscape".'
---

# Dream Skill

`$dream` is the interactive surface for the same Dream engine used by
`ao overnight`.

## Execution Steps

### Step 1: Route the request

- `setup` -> `ao overnight setup`
- `start` or `run` -> `ao overnight start`
- `report` -> `ao overnight report`

### Step 2: Setup lane

Use `ao overnight setup` to inspect host constraints, runner availability,
scheduler mode, and keep-awake behavior.

```bash
ao overnight setup
ao overnight setup --apply --runner codex --runner claude --at 01:30
```

Default to preview. Use `--apply` only when the user explicitly wants Dream
config or scheduler artifacts persisted.

### Step 3: Bedtime run lane

Use `ao overnight start` for the actual private local run.

```bash
ao overnight start --goal "close the loop on today's auth work"
ao overnight start --goal "stabilize release follow-ups" --runner codex --runner claude --creative-lane
```

### Step 4: Morning report lane

Use `ao overnight report` to render the latest Dream result.

```bash
ao overnight report
ao overnight report --from .agents/overnight/latest
```

When rendering a report, answer four questions fast:

1. What state did I wake up to?
2. What ran overnight?
3. What degraded or failed?
4. What should I do first?

## Key Rules

- Keep Dream settings under the shared `dream.*` control plane.
- Do not promise scheduled execution on a sleeping laptop.
- Do not imply tracked source-code edits overnight.
- GitHub nightly is the public proof harness, not the private Dream engine.

## See Also

- [Dream Run Contract](../../docs/contracts/dream-run-contract.md)
- [Dream Report Contract](../../docs/contracts/dream-report.md)
