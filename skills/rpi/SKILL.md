---
name: rpi
description: 'Full RPI lifecycle orchestrator. Delegates to /discovery, /crank, /validation phase skills. One command, full lifecycle with complexity classification, --from routing, and optional loop. Triggers: "rpi", "full lifecycle", "research plan implement", "end to end".'
skill_api_version: 1
user-invocable: true
metadata:
  tier: meta
  dependencies:
    - discovery   # phase 1 orchestrator
    - crank       # phase 2 orchestrator
    - validation  # phase 3 orchestrator
    - ratchet     # checkpoint tracking
  internal: false
---

# /rpi — Full RPI Lifecycle Orchestrator

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

## DAG — Execute This Sequentially

```
mkdir -p .agents/rpi
classify(goal) → complexity, start_phase
```

**From `--from` or start_phase, enter the DAG at the matching step and run every step after it:**

```
STEP 1  ──  if start_phase <= discovery:
              Skill(skill="discovery", args="<goal> [--interactive] --complexity=<level>")
              BLOCKED? → stop (manual intervention)
              DONE?    → read epic-id from .agents/rpi/execution-packet.json

STEP 2  ──  Skill(skill="crank", args="<epic-id> [--test-first] [--no-test-first]")
              BLOCKED/PARTIAL? → retry (max 3), then stop
              DONE? → ao ratchet record implement 2>/dev/null || true

STEP 3  ──  if complexity != fast:
              Skill(skill="validation", args="<epic-id> --complexity=<level>")
              FAIL? → re-crank + re-validate (max 3 total), then stop
              DONE? → ao ratchet record vibe 2>/dev/null || true

STEP 4  ──  report(verdicts)
              if --loop && FAIL && cycle < max_cycles: restart from STEP 1
              if --spawn-next: read .agents/rpi/next-work.jsonl, suggest next
```

**That's it.** Steps 1→2→3→4. No stopping between steps. No summarizing. No asking. Enter at `--from`, run to the end. The human's only touchpoint is after STEP 4.

---

## Setup + Classify (STEP 0 detail)

**Determine start_phase:**
- default: `discovery`
- `--from=implementation` (aliases: `crank`) → STEP 2
- `--from=validation` (aliases: `vibe`, `post-mortem`) → STEP 3
- aliases `research`, `plan`, `pre-mortem`, `brainstorm` → STEP 1
- Input looks like epic ID (`ag-*`) and no `--from` → STEP 2

**Classify complexity:**

| Level | Criteria | Behavior |
|-------|----------|----------|
| `fast` | Goal <=30 chars, no complex/scope keywords | STEP 3 skipped |
| `standard` | Goal 31-120 chars, or 1 scope keyword | Full DAG. Gates use `--quick` |
| `full` | Complex-operation keyword, 2+ scope keywords, or >120 chars | Full DAG. Gates use full council |

**Complex-operation keywords:** `refactor`, `migrate`, `migration`, `rewrite`, `redesign`, `rearchitect`, `overhaul`, `restructure`, `reorganize`, `decouple`, `deprecate`, `split`, `extract module`, `port`

**Scope keywords:** `all`, `entire`, `across`, `everywhere`, `every file`, `every module`, `system-wide`, `global`, `throughout`, `codebase`

**Overrides:** `--deep` forces `full`. `--fast-path` forces `fast`.

Log: `RPI mode: rpi-phased (complexity: <level>)`

Initialize state:
```
rpi_state = {
  goal: "<goal string>",
  epic_id: null,
  phase: "<discovery|implementation|validation>",
  complexity: "<fast|standard|full>",
  test_first: <true by default; false only when --no-test-first>,
  cycle: 1,
  max_cycles: <3 when --loop; overridden by --max-cycles>,
  verdicts: {}
}
```

## Gate Logic Detail

**STEP 1 gate (discovery):**
- `<promise>DONE</promise>`: extract epic-id from `.agents/rpi/execution-packet.json`, proceed to STEP 2
- `<promise>BLOCKED</promise>`: stop — discovery handles its own retries (max 3 pre-mortem attempts)

**STEP 2 gate (implementation, max 3 attempts):**
- `<promise>DONE</promise>`: proceed to STEP 3
- `<promise>BLOCKED</promise>`: retry with block context (max 2 retries)
- `<promise>PARTIAL</promise>`: retry remaining (max 2 retries)

**STEP 3 gate (validation-to-crank loop, max 3 total):**
- `<promise>DONE</promise>`: proceed to STEP 4
- `<promise>FAIL</promise>`: extract findings → re-invoke `/crank` with findings → re-invoke `/validation`

**STEP 4 (report + optional loop):**
- Summarize all phase verdicts and epic status. See `references/report-template.md`.
- `--loop` + FAIL + cycle < max_cycles: extract 3 fixes from post-mortem, increment cycle, restart from STEP 1
- `--spawn-next`: read `.agents/rpi/next-work.jsonl`, suggest next `/rpi` command (do NOT auto-invoke)

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--from=<phase>` | `discovery` | Enter DAG at `discovery`, `implementation`, or `validation` |
| `--interactive` | off | Human gates in discovery only |
| `--auto` | on | Fully autonomous. Inverse of `--interactive` |
| `--loop` | off | Post-mortem FAIL triggers new cycle |
| `--max-cycles=<n>` | `3` | Max cycles when `--loop` enabled |
| `--spawn-next` | off | Surface follow-up work after completion |
| `--test-first` | on | Strict-quality (passed to `/crank`) |
| `--no-test-first` | off | Opt out of strict-quality |
| `--fast-path` | auto | Force fast complexity (skip STEP 3) |
| `--deep` | auto | Force full complexity |
| `--dry-run` | off | Report without mutating queue |
| `--no-budget` | off | Disable phase time budgets |

## Quick Start

```bash
/rpi "add user authentication"                        # full DAG
/rpi --interactive "add user authentication"          # human gates in discovery only
/rpi --from=implementation ag-23k                      # enter at STEP 2
/rpi --from=validation                                 # enter at STEP 3
/rpi --loop --max-cycles=3 "add auth"                 # iterate-on-fail loop
/rpi --deep "refactor payment module"                  # force full council
/rpi --fast-path "fix typo in readme"                  # skip STEP 3
```

## Complexity-Scaled Council Gates

### Pre-mortem (STEP 5 in discovery)
complexity == "low": inline review, no spawning (--quick) | complexity == "medium": inline fast default (--quick) | complexity == "high": full council, 2-judge minimum. Retry gate: max 3 total attempts.

### Final Vibe (STEP 1 in validation)
complexity == "low": inline review, no spawning (--quick) | complexity == "medium": inline fast default (--quick) | complexity == "high": full council, 2-judge minimum. Retry gate: max 3 total attempts.

### Post-mortem (STEP 2 in validation)
complexity == "low": inline review, no spawning (--quick) | complexity == "medium": inline fast default (--quick) | complexity == "high": full council, 2-judge minimum. Retry gate: max 3 total attempts.

## Phase Data Contracts

All transitions use filesystem artifacts (no in-memory coupling). The execution packet (`.agents/rpi/execution-packet.json`) carries `contract_surfaces`, `done_criteria`, and queue metadata between steps. See `references/phase-data-contracts.md`.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Discovery blocks | Narrow goal scope or use `--interactive` |
| Implementation loops | Check `bd children <epic>` for blockers |
| Validation FAIL exhausted | Fix findings, re-run `--from=validation` |

Read `references/troubleshooting.md` for more.

**See also:** [discovery](../discovery/SKILL.md), [crank](../crank/SKILL.md), [validation](../validation/SKILL.md)

## Reference Documents

- [references/autonomous-execution.md](references/autonomous-execution.md)
- [references/complexity-scaling.md](references/complexity-scaling.md)
- [references/context-windowing.md](references/context-windowing.md)
- [references/error-handling.md](references/error-handling.md)
- [references/examples.md](references/examples.md)
- [references/gate-retry-logic.md](references/gate-retry-logic.md)
- [references/gate4-loop-and-spawn.md](references/gate4-loop-and-spawn.md)
- [references/phase-budgets.md](references/phase-budgets.md)
- [references/phase-data-contracts.md](references/phase-data-contracts.md)
- [references/report-template.md](references/report-template.md)
- [references/troubleshooting.md](references/troubleshooting.md)
