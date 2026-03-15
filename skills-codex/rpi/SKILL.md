---
name: rpi
description: 'Full RPI lifecycle orchestrator. Delegates to $discovery, $crank, $validation phase skills. One command, full lifecycle with complexity classification, --from routing, and optional loop. Triggers: "rpi", "full lifecycle", "research plan implement", "end to end".'
metadata:
  tier: meta
---


# $rpi — Full RPI Lifecycle Orchestrator

> **Quick Ref:** One command, full lifecycle. `$discovery` → `$crank` → `$validation`. Thin wrapper that delegates to phase orchestrators.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

**THE THREE-PHASE RULE:** RPI has THREE mandatory phases (unless complexity == `fast`). You MUST run all three — discovery, implementation, AND validation — in a single session. Do NOT stop after implementation. Do NOT ask the user if they want to commit after Phase 2. Phase 2 completing is NOT the end — it is the midpoint. Validation (Phase 3) is where learnings are captured and the knowledge flywheel turns. Skipping it breaks the flywheel.

**FULLY AUTONOMOUS BY DEFAULT.** Unless `--interactive` is explicitly set, RPI runs hands-free from start to finish. Do NOT:
- Ask the user for confirmation between phases
- Ask "want me to commit?" or "should I continue?"
- Pause to summarize and wait for input
- Request clarification mid-execution
- Stop to ask about approach or strategy

The human's only touchpoint is AFTER Phase 3 completes. If something is genuinely blocked (3 retries exhausted), then and only then do you stop and report. Everything else runs autonomously. The user invoked `$rpi` because they want you to GO — not to narrate.

**Phase completion tracking:** After each phase, log progress:
```
PHASE 1 COMPLETE ✓ (discovery) — proceeding to Phase 2
PHASE 2 COMPLETE ✓ (implementation) — proceeding to Phase 3
PHASE 3 COMPLETE ✓ (validation) — RPI DONE
```

## Quick Start

```bash
$rpi "add user authentication"                        # full lifecycle
$rpi --interactive "add user authentication"          # human gates in discovery only
$rpi --from=discovery "add auth"                      # resume discovery
$rpi --from=implementation ag-23k                      # skip to crank with existing epic
$rpi --from=validation                                 # run validation only
$rpi --loop --max-cycles=3 "add auth"                 # iterate-on-fail loop
$rpi --deep "refactor payment module"                  # force full council ceremony
$rpi --fast-path "fix typo in readme"                  # force lightweight ceremony
$rpi --no-test-first "add auth"                       # opt out of strict-quality
```

## Architecture

```
$rpi <goal | epic-id> [--from=<phase>] [--interactive] [--deep|--fast-path]
  │
  ├── Step 0: Setup + complexity classification
  │
  ├── Phase 1: $discovery <goal>
  │   └── brainstorm → ao search → research → plan → pre-mortem (gate)
  │   └── Outputs: epic-id, execution-packet.json
  │
  ├── Phase 2: $crank <epic-id>
  │   └── wave-based implementation + validation + rework
  │   └── Gate: DONE → proceed, BLOCKED/PARTIAL → retry (max 3)
  │
  └── Phase 3: $validation <epic-id>
      └── vibe → post-mortem → retro → forge
      └── Gate: PASS/WARN → finish, FAIL → re-crank (max 3)
```

**Phase orchestrators own all sub-skill sequencing, retry gates, and phase budgets.**
`$rpi` owns only: setup, complexity classification, phase routing, implementation gate, validation-fail-to-crank loop, and final report.

### Anti-Patterns (DO NOT)

| Anti-Pattern | Why It's Wrong | Correct Behavior |
|--------------|----------------|------------------|
| Stop after Phase 2 and ask to commit | Skips validation — no quality check, no learnings, flywheel doesn't turn | Proceed directly to Phase 3 |
| Call `$vibe` directly instead of `$validation` | `$vibe` is one sub-step; `$validation` wraps vibe + post-mortem + retro + forge | Always call `$validation` from `$rpi` |
| Ask "want me to commit?" between phases | Interrupts autonomous flow — user invoked `$rpi` for hands-free execution | Commit only after ALL phases complete |
| Ask the user ANY question during execution | RPI is autonomous unless `--interactive` — questions break the flow | Make best judgment and proceed; report at end |
| Run Phase 1 inline instead of delegating to `$discovery` | Loses brainstorm → search → research → plan → pre-mortem sequencing | Delegate via `$discovery ` |
| Summarize findings and wait after Phase 1 | Discovery output is an input to Phase 2, not a deliverable | Proceed immediately to Phase 2 |
| Pause to explain what you're about to do | Narration wastes time — the user wants results, not commentary | Execute, then report at the end |

## Execution Steps

### Step 0: Setup + Classify

```bash
mkdir -p .agents/rpi
```

**Determine starting phase:**
- default: `discovery`
- `--from=implementation` (aliases: `crank`) → skip to Phase 2
- `--from=validation` (aliases: `vibe`, `post-mortem`) → skip to Phase 3
- aliases `research`, `plan`, `pre-mortem`, `brainstorm` map to `discovery`
- If input looks like an epic ID (`ag-*`) and `--from` is not set, start at implementation.

**Classify complexity:**

| Level | Criteria | Behavior |
|-------|----------|----------|
| `fast` | Goal <=30 chars, no complex/scope keywords | Discovery → crank only. Skip validation. |
| `standard` | Goal 31-120 chars, or 1 scope keyword | Full 3-phase. Gates use `--quick`. |
| `full` | Complex-operation keyword, 2+ scope keywords, or >120 chars | Full 3-phase. Gates use full council. |

**Complex-operation keywords:** `refactor`, `migrate`, `migration`, `rewrite`, `redesign`, `rearchitect`, `overhaul`, `restructure`, `reorganize`, `decouple`, `deprecate`, `split`, `extract module`, `port`

**Scope keywords:** `all`, `entire`, `across`, `everywhere`, `every file`, `every module`, `system-wide`, `global`, `throughout`, `codebase`

**Overrides:** `--deep` forces `full`. `--fast-path` forces `fast`.

Log:
```
RPI mode: rpi-phased (complexity: <level>)
```

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

### Phase 1: Discovery

Delegate to `$discovery`:

```
$discovery <goal> [--interactive] --complexity=<level>
```

After `$discovery` completes:
1. Check completion marker: `<promise>DONE</promise>` or `<promise>BLOCKED</promise>`
2. If BLOCKED: stop. Discovery handles its own retries (max 3 pre-mortem attempts). Manual intervention needed.
3. If DONE: extract epic-id from `.agents/rpi/execution-packet.json`
4. Store `rpi_state.epic_id` and `rpi_state.verdicts.pre_mortem`
5. Log: `PHASE 1 COMPLETE ✓ (discovery) — proceeding to Phase 2`

### Phase 2: Implementation

Requires `rpi_state.epic_id`.

```
$crank <epic-id> [--test-first] [--no-test-first]
```

**Implementation gate (max 3 attempts):**
- `<promise>DONE</promise>`: proceed to validation
- `<promise>BLOCKED</promise>`: retry with block context (max 2 retries)
  - Re-invoke `$crank` with epic-id + block reason
  - If still BLOCKED after 3 total: stop, manual intervention needed
- `<promise>PARTIAL</promise>`: retry remaining (max 2 retries)
  - Re-invoke `$crank` with epic-id (picks up unclosed issues)
  - If still PARTIAL after 3 total: stop, manual intervention needed

Record:
```bash
ao ratchet record implement 2>/dev/null || true
```

Log: `PHASE 2 COMPLETE ✓ (implementation) — proceeding to Phase 3`

**DO NOT STOP HERE.** Do not ask the user to commit. Do not summarize and wait. Proceed IMMEDIATELY to Phase 3. Implementation without validation is incomplete work — the flywheel does not turn, learnings are not captured, and quality is unverified.

### Phase 3: Validation

**Skip if:** complexity == `fast` (fast-path runs discovery + crank only).

**MANDATORY for `standard` and `full` complexity.** This is not optional. `$validation` is the Phase 3 orchestrator — it wraps `$vibe` + `$post-mortem` + `$retro` + `$forge`. Do NOT call `$vibe` directly from `$rpi` — call `$validation` which handles the full sequence.

```
$validation <epic-id> --complexity=<level>
```

**Validation-to-crank loop (max 3 total attempts):**
- `<promise>DONE</promise>`: finish RPI
- `<promise>FAIL</promise>`: vibe found defects
  1. Extract findings from validation output
  2. Re-invoke `$crank` with epic-id + findings context (preserve `--test-first` / `--no-test-first` from original invocation)
  3. Re-invoke `$validation`
  4. If still FAIL after 3 total: stop, manual intervention needed

Record:
```bash
ao ratchet record vibe 2>/dev/null || true
```

Log: `PHASE 3 COMPLETE ✓ (validation) — RPI DONE`

### Step Final: Report + Loop

**Report:** Summarize all phase verdicts and epic status.

**Optional loop (`--loop`):** If validation verdict is FAIL and `cycle < max_cycles`:
1. Extract 3 concrete fixes from the post-mortem report
2. Increment `rpi_state.cycle`
3. Re-invoke `$rpi` from discovery with a tightened goal
4. PASS/WARN stops the loop

**Optional spawn-next (`--spawn-next`):** After PASS/WARN finish:
1. Read `.agents/rpi/next-work.jsonl` for harvested follow-up items
2. Report with suggested next `$rpi` command
3. Do NOT auto-invoke

Read `references/report-template.md` for full output format.
Read `references/error-handling.md` for failure semantics.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--from=<phase>` | `discovery` | Start from `discovery`, `implementation`, or `validation` |
| `--interactive` | off | Human gates in discovery |
| `--loop` | off | Post-mortem FAIL triggers new cycle |
| `--max-cycles=<n>` | `3` | Max cycles when `--loop` enabled (default 3) |
| `--spawn-next` | off | Surface follow-up work after completion |
| `--test-first` | on | Strict-quality (passed to `$crank`) |
| `--no-test-first` | off | Opt out of strict-quality |
| `--fast-path` | auto | Force fast complexity |
| `--deep` | auto | Force full complexity |
| `--dry-run` | off | Report without mutating queue |
| `--no-budget` | off | Disable phase time budgets (passed to phase skills) |

## Phase Data Contracts

All transitions use filesystem artifacts (no in-memory coupling). The execution packet (`.agents/rpi/execution-packet.json`) carries `contract_surfaces` (repo execution profile), `done_criteria`, and queue claim/finalize metadata between phases. Sub-skills include $plan, $vibe, $post-mortem, and $pre-mortem. For detailed contract schemas, read `references/phase-data-contracts.md`.

## Complexity-Scaled Council Gates

### Phase 3: Pre-mortem
- complexity == "low": inline review, no spawning (--quick)
- complexity == "medium": inline fast default (--quick)
- complexity == "high": full council, 2-judge minimum
- Retry gate: max 3 total attempts

### Phase 5: Final Vibe
- complexity == "low": inline review, no spawning (--quick)
- complexity == "medium": inline fast default (--quick)
- complexity == "high": full council, 2-judge minimum
- Retry gate: max 3 total attempts

### Phase 6: Post-mortem
- complexity == "low": inline review, no spawning (--quick)
- complexity == "medium": inline fast default (--quick)
- complexity == "high": full council, 2-judge minimum
- Retry gate: max 3 total attempts

## Examples

Read `references/examples.md` for full lifecycle, resume, and interactive examples.

## Troubleshooting

Read `references/troubleshooting.md` for common problems and solutions.

**See also:** [discovery](../discovery/SKILL.md), [crank](../crank/SKILL.md), [validation](../validation/SKILL.md)

## Reference Documents

- [references/complexity-scaling.md](references/complexity-scaling.md)
- [references/context-windowing.md](references/context-windowing.md)
- [references/gate-retry-logic.md](references/gate-retry-logic.md)
- [references/gate4-loop-and-spawn.md](references/gate4-loop-and-spawn.md)
- [references/phase-budgets.md](references/phase-budgets.md)
- [references/phase-data-contracts.md](references/phase-data-contracts.md)
- [references/report-template.md](references/report-template.md)
- [references/error-handling.md](references/error-handling.md)
- [references/examples.md](references/examples.md)
- [references/troubleshooting.md](references/troubleshooting.md)

## Local Resources

### references/

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

### scripts/

- `scripts/validate.sh`


