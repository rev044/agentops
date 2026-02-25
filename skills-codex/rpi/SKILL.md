---
name: rpi
description: 'Full RPI lifecycle orchestrator. Discovery (research+plan+pre-mortem) → Implementation (crank) → Validation (vibe+post-mortem). One command, sequential skill invocations with retry gates and fresh phase contexts.'
---


# $rpi — Full RPI Lifecycle Orchestrator

> **Quick Ref:** One command, full lifecycle. Discovery → Implementation → Validation. The session is the lead; sub-skills manage their own teams.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

## Quick Start

```bash
$rpi "add user authentication"                        # full lifecycle
$rpi --interactive "add user authentication"          # human gates in discovery only
$rpi --from=discovery "add auth"                      # resume discovery
$rpi --from=implementation ag-23k                      # skip to crank with existing epic
$rpi --from=validation                                 # run vibe + post-mortem only
$rpi --loop --max-cycles=3 "add auth"                 # optional iterate-on-fail loop
$rpi --test-first "add auth"                          # pass --test-first to $crank
```

## CLI Toolchain Configuration

RPI control-plane command paths are configurable through `.agentops/config.yaml` or environment variables:

```yaml
rpi:
  runtime_mode: auto        # auto|direct|stream
  runtime_command: claude   # runtime process command
  ao_command: ao            # ratchet/checkpoint command
  bd_command: bd            # epic/child query command
  tmux_command: tmux        # status liveness probe command
```

Environment variable overrides:
- `AGENTOPS_RPI_RUNTIME` / `AGENTOPS_RPI_RUNTIME_MODE`
- `AGENTOPS_RPI_RUNTIME_COMMAND`
- `AGENTOPS_RPI_AO_COMMAND`
- `AGENTOPS_RPI_BD_COMMAND`
- `AGENTOPS_RPI_TMUX_COMMAND`

Safety defaults:
- `git`, `bash`, and `ps` remain fixed system tools in the RPI control plane.
- Command precedence is `flags > env > config > defaults` where flags exist.

## Architecture

```
$rpi <goal | epic-id> [--from=<phase>] [--interactive]
  │ (session = lead, no TeamCreate)
  │
  ├── Phase 1: Discovery
  │   ├── $research
  │   ├── $plan
  │   └── $pre-mortem (gate)
  │
  ├── Phase 2: Implementation
  │   └── $crank (autonomous execution)
  │
  └── Phase 3: Validation
      ├── $vibe (gate)
      └── $post-mortem (retro + flywheel)
```

**Human gates (default):** 0 — fully autonomous.
**Human gates (`--interactive`):** discovery approvals in `$research` and `$plan`.
**Retry gates:** pre-mortem FAIL → re-plan, implementation BLOCKED/PARTIAL → re-crank, vibe FAIL → re-crank (max 3 attempts each).
**Optional loop (`--loop`):** post-mortem FAIL can spawn another RPI cycle.

### Phase Data Contracts

All phase transitions use filesystem-based artifacts (no in-memory coupling):

| Transition | Key Artifacts | How Next Phase Reads Them |
|------------|---------------|---------------------------|
| Start -> Discovery | Goal string | Passed as argument to `$research` |
| Discovery -> Implementation | Epic ID, pre-mortem verdict, phase-1 summary | `phased-state.json` + `.agents/rpi/phase-1-summary-*.md` |
| Implementation -> Validation | Crank completion status, phase-2 summary | `bd children <epic-id>` + `.agents/rpi/phase-2-summary-*.md` |
| Validation -> Next Cycle (optional) | Vibe/post-mortem verdicts, harvested follow-up work | Council reports + `.agents/rpi/next-work.jsonl` |

## Execution Steps

Given `$rpi <goal | epic-id> [--from=<phase>] [--interactive]`:

### Step 0: Setup

```bash
mkdir -p .agents/rpi
```

Determine starting phase:
- default: `discovery`
- `--from=implementation` (alias: `crank`)
- `--from=validation` (aliases: `vibe`, `post-mortem`)
- aliases `research`, `plan`, and `pre-mortem` map to `discovery`

If input looks like an epic ID (`ag-*`) and `--from` is not set, start at implementation.

Initialize state:

```
rpi_state = {
  goal: "<goal string>",
  epic_id: null,
  phase: "<discovery|implementation|validation>",
  auto: <true unless --interactive>,
  test_first: <true if --test-first>,
  complexity: null,
  cycle: 1,
  parent_epic: null,
  verdicts: {}
}
```

### Phase 1: Discovery

Discovery is one context window that runs research, planning, and pre-mortem together:

```text
$research <goal> [--auto]
$plan <goal> [--auto]
$pre-mortem
```

After discovery completes:
1. Extract epic ID from `bd list --type epic --status open` and store in `rpi_state.epic_id`.
2. Extract pre-mortem verdict (PASS/WARN/FAIL) from latest pre-mortem council report.
3. Store verdict in `rpi_state.verdicts.pre_mortem`.
4. Write summary to `.agents/rpi/phase-1-summary-YYYY-MM-DD-<goal-slug>.md`.
5. Record ratchet and telemetry:

```bash
ao ratchet record research 2>/dev/null || true
bash scripts/checkpoint-commit.sh rpi "phase-1" "discovery complete" 2>/dev/null || true
bash scripts/log-telemetry.sh rpi phase-complete phase=1 phase_name=discovery 2>/dev/null || true
```

Gate behavior:
- PASS/WARN: proceed to implementation.
- FAIL: re-run `$plan` with findings context (top 5 structured findings: description, fix, ref), then `$pre-mortem` again. Max 3 total attempts. If still FAIL after 3, stop and require manual intervention.

### Phase 2: Implementation

Requires `rpi_state.epic_id`.

```text
$crank <epic-id> [--test-first]
```

After implementation completes:
1. Check completion via crank output / epic child statuses.
2. Gate result:
   - DONE: proceed to validation
   - BLOCKED: re-run `$crank` with block context (max 3 total attempts). If still BLOCKED, stop and require manual intervention.
   - PARTIAL: re-run `$crank` with epic-id; it picks up unclosed issues (max 3 total attempts). If still PARTIAL, stop and require manual intervention.
3. Write summary to `.agents/rpi/phase-2-summary-YYYY-MM-DD-<goal-slug>.md`.
4. Record ratchet and telemetry:

```bash
ao ratchet record implement 2>/dev/null || true
bash scripts/checkpoint-commit.sh rpi "phase-2" "implementation complete" 2>/dev/null || true
bash scripts/log-telemetry.sh rpi phase-complete phase=2 phase_name=implementation 2>/dev/null || true
```

### Phase 3: Validation

Validation runs final review and lifecycle close-out:

```text
$vibe recent            # use --quick recent for low/medium complexity
$post-mortem <epic-id>  # use --quick for low/medium complexity
```

After validation completes:
1. Extract vibe verdict and store `rpi_state.verdicts.vibe`.
2. If present, extract post-mortem verdict and store `rpi_state.verdicts.post_mortem`.
3. Gate result:
   - PASS/WARN: finish RPI
   - FAIL: re-run implementation with findings, then re-run validation (max 3 total attempts)
4. Write summary to `.agents/rpi/phase-3-summary-YYYY-MM-DD-<goal-slug>.md`.
5. Record ratchet and telemetry:

```bash
ao ratchet record vibe 2>/dev/null || true
bash scripts/checkpoint-commit.sh rpi "phase-3" "validation complete" 2>/dev/null || true
bash scripts/log-telemetry.sh rpi phase-complete phase=3 phase_name=validation 2>/dev/null || true
```

**Optional loop (`--loop`):** If post-mortem verdict is FAIL and `--loop` is enabled, extract 3 concrete fixes from the report and re-invoke `$rpi` from Phase 1 with a tightened goal (up to `--max-cycles` total cycles). PASS/WARN stops the loop.

**Optional spawn-next (`--spawn-next`):** After a PASS/WARN finish, read `.agents/rpi/next-work.jsonl` for harvested follow-up items. Report them to the user with a suggested next `$rpi` command but do NOT auto-invoke.

### Step Final: Report

Read `references/report-template.md` for the final output format and next-work handoff pattern.

Read `references/error-handling.md` for failure semantics and retries.

## Complexity-Aware Ceremony

RPI automatically classifies each goal's complexity at startup and adjusts the ceremony accordingly. This prevents trivial tasks from paying the full validation overhead of a refactor.

### Classification Levels

| Level | Criteria | Behavior |
|-------|----------|----------|
| `fast` | Goal ≤30 chars, no complex/scope keywords | Skips Phase 3 (validation). Runs discovery → implementation only. |
| `standard` | Goal 31–120 chars, or 1 scope keyword | Full 3-phase lifecycle. Gates use `--quick` shortcuts. |
| `full` | Complex-operation keyword (refactor, migrate, rewrite, …), 2+ scope keywords, or >120 chars | Full 3-phase lifecycle. Gates use full council (no shortcuts). |

### Keyword Signals

**Complex-operation keywords** (trigger `full`): `refactor`, `migrate`, `migration`, `rewrite`, `redesign`, `rearchitect`, `overhaul`, `restructure`, `reorganize`, `decouple`, `deprecate`, `split`, `extract module`, `port`

**Scope keywords** (1 → `standard`; 2+ → `full`): `all`, `entire`, `across`, `everywhere`, `every file`, `every module`, `system-wide`, `global`, `throughout`, `codebase`

All keywords are matched as **whole words** to prevent false positives (e.g. "support" does not match "port").

### Logged Output

At RPI start you will see:

```
RPI mode: rpi-phased (complexity: fast)
Complexity: fast — skipping validation phase (phase 3)
```

or for standard/full:

```
RPI mode: rpi-phased (complexity: standard)
```

The complexity level is persisted in `.agents/rpi/phased-state.json` as the `complexity` field.

### Override

- `--fast-path`: force fast-path regardless of classification (useful for quick patches).
- `--deep`: force full-ceremony regardless of classification (useful for sensitive changes).

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--from=<phase>` | `discovery` | Start from `discovery`, `implementation`, or `validation` (aliases accepted) |
| `--interactive` | off | Enable human gates in discovery (`$research`, `$plan`) |
| `--auto` | on | Legacy flag; autonomous is default |
| `--loop` | off | Enable post-mortem FAIL loop into next cycle |
| `--max-cycles=<n>` | `1` | Max cycle count when `--loop` is enabled |
| `--spawn-next` | off | Surface harvested follow-up work after post-mortem |
| `--test-first` | off | Pass `--test-first` to `$crank` |
| `--fast-path` | auto | Force low-complexity gate mode (`--quick`) |
| `--deep` | auto | Force high-complexity gate mode (full council) |
| `--dry-run` | off | Report actions without mutating next-work queue |

## Examples

### Full Lifecycle

**User says:** `$rpi "add user authentication"`

**What happens:**
1. Discovery runs `$research`, `$plan`, `$pre-mortem` and produces epic `ag-5k2`.
2. Implementation runs `$crank ag-5k2` until children are complete.
3. Validation runs `$vibe` then `$post-mortem`, extracts learnings, and suggests next work.

### Resume from Implementation

**User says:** `$rpi --from=implementation ag-5k2`

**What happens:**
1. Skips discovery.
2. Runs `$crank ag-5k2`.
3. Runs validation (`$vibe` + `$post-mortem`).

### Interactive Discovery

**User says:** `$rpi --interactive "refactor payment module"`

**What happens:**
1. Discovery runs with human gates in `$research` and `$plan`.
2. Implementation and validation remain autonomous.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Supervisor spiraled branch count | Detached HEAD healing or legacy `codex/auto-rpi-*` naming created detached branches | Keep `--detached-heal` off for supervisor mode (default), prefer detached worktree execution, then run cleanup: `ao rpi cleanup --all --prune-worktrees --prune-branches --dry-run` to preview, then rerun without `--dry-run`. |
| Discovery retries hit max attempts | Plan has unresolved risks | Review pre-mortem findings, re-run `$rpi --from=discovery` |
| Implementation retries hit max attempts | Epic has blockers or unresolved dependencies | Inspect `bd show <epic-id>`, fix blockers, re-run `$rpi --from=implementation` |
| Validation retries hit max attempts | Vibe found critical defects repeatedly | Apply findings, re-run `$rpi --from=validation` |
| Missing epic ID at implementation start | Discovery did not produce a parseable epic | Verify latest open epic with `bd list --type epic --status open` |
| Large-repo context pressure | Too much context in one window | Use `references/context-windowing.md` and summarize phase outputs aggressively |

### Emergency control

- Cancel in-flight RPI work immediately: `ao rpi cancel --all` (or `--run-id <id>` for one run).
- Remove stale worktrees and legacy branches: `ao rpi cleanup --all --prune-worktrees --prune-branches`.

## See Also

- `skills/research/SKILL.md` — discovery exploration
- `skills/plan/SKILL.md` — discovery decomposition
- `skills/pre-mortem/SKILL.md` — discovery risk gate
- `skills/crank/SKILL.md` — implementation execution
- `skills/vibe/SKILL.md` — validation gate
- `skills/post-mortem/SKILL.md` — validation close-out

## Reference Documents

- [references/complexity-scaling.md](references/complexity-scaling.md)
- [references/context-windowing.md](references/context-windowing.md)
- [references/gate-retry-logic.md](references/gate-retry-logic.md)
- [references/gate4-loop-and-spawn.md](references/gate4-loop-and-spawn.md)
- [references/phase-data-contracts.md](references/phase-data-contracts.md)

---

## References

### complexity-scaling.md

# Complexity Scaling

Automatic complexity detection determines the level of validation ceremony applied to each RPI cycle.

## Classification Table

| Level | Issue Count | Wave Count | Ceremony |
|-------|------------|------------|----------|
| **low** | ≤2 | 1 | fast-path: `--quick` on pre-mortem, vibe, post-mortem |
| **medium** | 3-6 | 1-2 | lean: `--quick` on pre-mortem, vibe, post-mortem (same as low — inline review is sufficient for most work) |
| **high** | 7+ OR 3+ waves | any | thorough: full council on pre-mortem and vibe, standard post-mortem |

> **Design rationale (2026-02-18):** `--quick` (inline single-agent structured review) catches the same class of bugs as full multi-judge council at ~10% of the token cost. The value of multi-judge consensus scales with stakes, not linearly with issue count. Medium-complexity epics (3-6 issues) don't benefit enough from multi-agent spawning to justify the 5-10x cost multiplier. Full council is reserved for high-stakes work where cross-model disagreement has real ROI.

## Detection

Complexity is auto-detected after plan completes (Phase 2) by examining:
- Issue count: `bd children <epic-id> | wc -l`
- Wave count: derived from dependency depth

## Flag Precedence (explicit always wins)

| Flag | Effect |
|------|--------|
| `--fast-path` | Forces `low` regardless of auto-detection |
| `--deep` (passed to $rpi) | Forces `high` regardless of auto-detection |
| No flag | Auto-detect from epic structure |

Existing mandatory pre-mortem gate (3+ issues) still applies regardless of complexity level.

### context-windowing.md

# Large-Repo Context Windowing

Use this mode when the repo is too large for stable single-window analysis.

## Why

Trying to read everything in one pass causes context collapse and unstable decisions.
Deterministic shards let `$rpi` process all files incrementally with bounded load.

## Setup Contract

```bash
scripts/rpi/context-window-contract.sh
```

This verifies:
- `GOALS.yaml` is valid
- shard manifest generation works
- shard progress state initializes and validates
- shard runner can traverse shard 1

## Generate Shards

```bash
scripts/rpi/generate-context-shards.py \
  --max-units 80 \
  --max-bytes 300000 \
  --out .agents/rpi/context-shards/latest.json \
  --check
```

## Initialize Progress State

```bash
scripts/rpi/init-shard-progress.py \
  --manifest .agents/rpi/context-shards/latest.json \
  --progress .agents/rpi/context-shards/progress.json \
  --check
```

## Run One Shard (Bounded)

```bash
scripts/rpi/run-shard.py \
  --manifest .agents/rpi/context-shards/latest.json \
  --progress .agents/rpi/context-shards/progress.json \
  --shard-id 1 \
  --limit 20 \
  --mark in_progress \
  --notes "phase-1 analysis start"
```

## Operating Pattern

1. Generate shard manifest once per material repo change.
2. Process one shard at a time.
3. Write concise shard summaries to `.agents/rpi/phase-*.md`.
4. Mark shard status (`todo`, `in_progress`, `done`).
5. Continue until all shards are `done`.

This keeps CPU and context budgets bounded while preserving full-file coverage.

### error-handling.md

# Error Handling

| Failure | Behavior |
|---------|----------|
| Skill invocation fails | Log error, retry once. If still fails, stop with checkpoint. |
| User abandons at sub-skill gate | $rpi stops with checkpoint (only in --interactive mode) |
| $crank returns BLOCKED | Re-crank with context (max 2 retries). If still blocked, stop. |
| $crank returns PARTIAL | Re-crank remaining items with context (max 2 retries). If still partial, stop. |
| Pre-mortem FAIL | Re-plan with fail feedback, re-run pre-mortem (max 3 total attempts) |
| Vibe FAIL | Re-crank with fail feedback, re-run vibe (max 3 total attempts) |
| Max retries exhausted | Stop with message + path to last report. Manual intervention needed. |
| Context feels degraded | Log warning, suggest starting new session with --from |

### gate-retry-logic.md

# Gate and Retry Logic

Detailed retry behavior for each gated phase. All gates use a max-3-attempts pattern (1 initial + 2 retries).

## Pre-mortem Gate (Phase 3)

Extract verdict from council report:

```bash
REPORT=$(ls -t .agents/council/*pre-mortem*.md 2>/dev/null | head -1)
```

Read the report file and find the verdict line (`## Council Verdict: PASS / WARN / FAIL`).

Gate logic:
- **PASS:** Auto-proceed. Log: "Pre-mortem: PASS"
- **WARN:** Auto-proceed. Log: "Pre-mortem: WARN -- see report for concerns"
- **FAIL:** Retry loop (max 2 retries):
  1. Read the full pre-mortem report to extract specific failure reasons
  1a. Extract top 5 findings with structured fields:
      ```
      For each finding (max 5), extract:
        FINDING: <description> | FIX: <fix or recommendation> | REF: <ref or location>

      Fallback for v1 findings: fix = finding.fix || finding.recommendation || "No fix specified"
                                 ref = finding.ref || finding.location || "No reference"
      ```
  2. Log: "Pre-mortem: FAIL (attempt N/3) -- retrying plan with feedback"
  3. Re-invoke `$plan` with the goal AND the failure context including structured findings:
     ```
     Skill(skill="plan", args="<goal> --auto --context 'Pre-mortem FAIL: <key concerns>\nStructured findings:\nFINDING: X | FIX: Y | REF: Z\nFINDING: A | FIX: B | REF: C'")
     ```
  4. Re-invoke `$pre-mortem` on the new plan
  5. If still FAIL after 3 total attempts, stop with message:
     "Pre-mortem failed 3 times. Last report: <path>. Manual intervention needed."

Store verdict in `rpi_state.verdicts.pre_mortem`.

## Implementation Gate (Phase 2)

Check completion status from crank's output. Look for `<promise>` tags:

- **`<promise>DONE</promise>`:** Proceed to Validation (Phase 3)
- **`<promise>BLOCKED</promise>`:** Retry (max 2 retries):
  1. Read crank output to extract block reason
  2. Log: "Crank: BLOCKED (attempt N/3) -- retrying with context"
  3. Re-invoke `$crank` with epic-id and block context (include `--test-first` if set)
  4. If still BLOCKED after 3 total attempts, stop with message:
     "Crank blocked 3 times. Reason: <reason>. Manual intervention needed."
- **`<promise>PARTIAL</promise>`:** Retry remaining (max 2 retries):
  1. Read crank output to identify remaining items
  2. Log: "Crank: PARTIAL (attempt N/3) -- retrying remaining items"
  3. Re-invoke `$crank` with epic-id (it picks up unclosed issues; include `--test-first` if set)
  4. If still PARTIAL after 3 total attempts, stop with message:
     "Crank partial after 3 attempts. Remaining: <items>. Manual intervention needed."

## Validation Gate (Phase 3)

Extract verdict from council report:

```bash
REPORT=$(ls -t .agents/council/*vibe*.md 2>/dev/null | head -1)
```

Read and extract verdict.

Gate logic:
- **PASS:** Auto-proceed. Log: "Vibe: PASS"
- **WARN:** Auto-proceed. Log: "Vibe: WARN -- see report for concerns"
- **FAIL:** Retry loop (max 2 retries):
  1. Read the full vibe report to extract specific failure reasons
  1a. Extract top 5 findings with structured fields:
      ```
      For each finding (max 5), extract:
        FINDING: <description> | FIX: <fix or recommendation> | REF: <ref or location>

      Fallback for v1 findings: fix = finding.fix || finding.recommendation || "No fix specified"
                                 ref = finding.ref || finding.location || "No reference"
      ```
  2. Log: "Vibe: FAIL (attempt N/3) -- retrying crank with feedback"
  3. Re-invoke `$crank` with the epic-id AND the failure context including structured findings:
     ```
     Skill(skill="crank", args="<epic-id> --context 'Vibe FAIL: <key issues>\nStructured findings:\nFINDING: X | FIX: Y | REF: Z' --test-first")   # if --test-first set
     Skill(skill="crank", args="<epic-id> --context 'Vibe FAIL: <key issues>\nStructured findings:\nFINDING: X | FIX: Y | REF: Z'")                 # otherwise
     ```
  4. Re-invoke `$vibe` on the new changes
  5. If still FAIL after 3 total attempts, stop with message:
     "Vibe failed 3 times. Last report: <path>. Manual intervention needed."

Store verdict in `rpi_state.verdicts.vibe`.

### gate4-loop-and-spawn.md

# Gate 4 Loop and Spawn Next Work

## Post-Validation Loop (Optional) -- Post-mortem to Spawn Another $rpi

**Default behavior:** $rpi ends after Validation (Phase 3).

**Enable loop:** pass `--loop` (and optionally `--max-cycles=<n>`).

**Gate 4 goal:** make the "ITERATE vs TEMPER" decision explicit, and if iteration is required, run another full $rpi cycle with tighter context.

**Loop decision input:** the most recent post-mortem council verdict.

1. Find the most recent post-mortem report:
   ```bash
   REPORT=$(ls -t .agents/council/*post-mortem*.md 2>/dev/null | head -1)
   ```
2. Read `REPORT` and extract the verdict line (`## Council Verdict: PASS / WARN / FAIL`).
3. Apply gate logic (only when `--loop` is set). If verdict is PASS or WARN, stop (TEMPER path). If verdict is FAIL, iterate (spawn another $rpi cycle), up to `--max-cycles`.
4. Iterate behavior (spawn). Read the post-mortem report and extract 3 concrete fixes, then re-invoke $rpi from Phase 1 with a tightened goal that includes the fixes:
   ```
   $rpi "<original goal> (Iteration <n>): Fix <item1>; <item2>; <item3>" --test-first   # if --test-first set
   $rpi "<original goal> (Iteration <n>): Fix <item1>; <item2>; <item3>"                 # otherwise
   ```
   If still FAIL after `--max-cycles` total cycles, stop and require manual intervention (file follow-up bd issues).

## Spawn Next Work (Optional) -- Post-mortem to Queue Next RPI

**Enable:** pass `--spawn-next` flag.

**Complementary to Gate 4:** Gate 4 (`--loop`) handles FAIL->iterate (same goal, tighter). `--spawn-next` handles PASS/WARN->new-goal (different work harvested from post-mortem).

1. Read `.agents/rpi/next-work.jsonl` for unconsumed entries (schema: `.agents/rpi/next-work.schema.md`).
   Filter entries by `target_repo`:
   - **Include** if `target_repo` matches the current repo name, OR `target_repo` is `"*"` (wildcard), OR the field is absent (backward compatibility).
   - **Skip** if `target_repo` names a different repo.
   - Current repo is derived from: `basename` of `git remote get-url origin`, or failing that, `basename "$PWD"`.
2. If unconsumed, repo-matched entries exist:
   - If `--dry-run` is set: report items but do NOT mutate next-work.jsonl (skip consumption). Log: "Dry run -- items not marked consumed."
   - Otherwise: mark the current cycle's entry as consumed (set `consumed: true`, `consumed_by: <epic-id>`, `consumed_at: <now>`)
   - Report harvested items to user with suggested next command:
     ```
     ## Next Work Available

     Post-mortem harvested N follow-up items from <source_epic>:
     1. <title> (severity: <severity>, type: <type>)
     ...

     To start the next RPI cycle:
       $rpi "<highest-severity item title>"
     ```
   - Do NOT auto-invoke `$rpi` -- the user decides when to start the next cycle
3. If no unconsumed entries: report "No follow-up work harvested. Flywheel stable."

**Note:** Only `--spawn-next` mutates next-work.jsonl (marks consumed). Phase 0 read is read-only.

## Repo-Scoped Filtering (target_repo)

Both Phase 0 and `--spawn-next` filter next-work entries by `target_repo`:

| `target_repo` value | Behavior |
|---------------------|----------|
| Matches current repo | Included |
| `"*"` (wildcard) | Included — applies to any repo |
| Absent / null | Included — backward compatible with pre-v1.2 entries |
| Different repo name | Skipped — intended for a different rig |

The current repo name is resolved as: `basename $(git remote get-url origin 2>/dev/null)` with `.git` suffix stripped, falling back to `basename "$PWD"` when no remote is configured.

This prevents cross-repo pollution when `.agents/rpi/next-work.jsonl` is shared or synced across rigs.

### phase-data-contracts.md

# Phase Data Contracts

How each consolidated phase passes data to the next. Artifacts are filesystem-based; no in-memory coupling between phases.

| Transition | Output | Extraction | Input to Next |
|------------|--------|------------|---------------|
| → Discovery | Research doc, plan doc, pre-mortem report, epic ID | Latest files in `.agents/research/`, `.agents/plans/`, `.agents/council/`; epic from `bd list --type epic --status open` | `epic_id`, `pre_mortem` verdict, and discovery summary are persisted in phased state |
| Discovery → Implementation | Epic execution context + discovery summary | `phased-state.json` + `.agents/rpi/phase-1-summary.md` | `$crank <epic-id>` with prior-phase context |
| Implementation → Validation | Completed/partial crank status + implementation summary | `bd children <epic-id>` + `.agents/rpi/phase-2-summary.md` | `$vibe` + `$post-mortem` with implementation context |
| Validation → Next Cycle (optional) | Vibe/post-mortem verdicts + harvested follow-up work | Latest council reports + `.agents/rpi/next-work.jsonl` | Stop, loop (`--loop`), or suggest next `$rpi` (`--spawn-next`) |

### report-template.md

# Final Report Template

After all phases complete, summarize the entire lifecycle to the user.

## Summary Report

```markdown
## $rpi Complete

**Goal:** <goal>
**Epic:** <epic-id>
**Cycle:** <rpi_state.cycle> (parent: <rpi_state.parent_epic or "none">)

| Phase | Verdict/Status |
|-------|---------------|
| Research | Complete |
| Plan | Complete (<N> issues, <M> waves) |
| Pre-mortem | <PASS/WARN/FAIL> |
| Crank | <DONE/BLOCKED/PARTIAL> |
| Vibe | <PASS/WARN/FAIL> |
| Post-mortem | Complete |

**Artifacts:**
- Research: .agents/research/...
- Plan: .agents/plans/...
- Pre-mortem: .agents/council/...
- Vibe: .agents/council/...
- Post-mortem: .agents/council/...
- Learnings: .agents/learnings/...
- Next Work: .agents/rpi/next-work.jsonl
```

## Flywheel Section

**ALWAYS include the flywheel section** (regardless of `--spawn-next` flag):

```markdown
## Flywheel: Next Cycle

Post-mortem harvested N follow-up items (M process-improvements, K tech-debt):

| # | Title | Type | Severity |
|---|-------|------|----------|
| 1 | ... | process-improvement | high |

Ready to run:
    $rpi "<highest-severity item title>"
```

The `--spawn-next` flag controls whether items are **marked consumed** in `next-work.jsonl`. The suggestion is ALWAYS shown. This ensures every `$rpi` cycle ends by pointing at the next one -- the flywheel never stops spinning unless there's nothing to improve.


---

## Scripts

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "SKILL.md has name: rpi" "grep -q '^name: rpi' '$SKILL_DIR/SKILL.md'"
check "references/ directory exists" "[ -d '$SKILL_DIR/references' ]"
check "references/ has at least 3 files" "[ \$(ls '$SKILL_DIR/references/' | wc -l) -ge 3 ]"
check "SKILL.md mentions research phase" "grep -qi 'research' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions plan phase" "grep -qi '$plan' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions pre-mortem phase" "grep -qi 'pre-mortem' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions crank phase" "grep -qi '$crank' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions vibe phase" "grep -qi '$vibe' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions post-mortem phase" "grep -qi 'post-mortem' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


