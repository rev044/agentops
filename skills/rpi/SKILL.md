---
name: rpi
description: 'Full RPI lifecycle orchestrator. Research → Plan → Pre-mortem → Crank → Vibe → Post-mortem. One command, sequential skill invocations with human gates and hands-free validation. Triggers: "rpi", "full lifecycle", "end to end", "research to production".'
metadata:
  tier: orchestration
  dependencies:
    - research    # required - Phase 1
    - plan        # required - Phase 2
    - pre-mortem  # required - Phase 3 (gate)
    - crank       # required - Phase 4 (implementation)
    - vibe        # required - Phase 5 (gate)
    - post-mortem # required - Phase 6
    - ratchet     # required - checkpoint tracking
  internal: false
---

# /rpi — Full RPI Lifecycle Orchestrator

> **Quick Ref:** One command, full lifecycle. Research → Plan → Pre-mortem → Crank → Vibe → Post-mortem. The session IS the lead. Sub-skills manage their own teams.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

## Quick Start

```bash
/rpi "add user authentication"              # Full lifecycle, fully autonomous (default)
/rpi --interactive "add user authentication" # Human gates at research + plan
/rpi --from=plan "add auth"                 # Skip research, start from plan
/rpi --from=crank ag-23k                    # Skip to crank with existing epic
/rpi --from=vibe                            # Just run final validation + post-mortem
/rpi --loop --max-cycles=3 "add auth"       # Gate 4 loop: post-mortem FAIL -> spawn another /rpi cycle
/rpi --test-first "add auth"               # Spec-first TDD (contracts → tests → impl)
```

## Architecture

```
/rpi <goal | epic-id> [--from=<phase>] [--interactive]
  │  (session = the lead, no TeamCreate)
  │
  ├── Phase 1: /research ── auto (default) or human gate (--interactive)
  ├── Phase 2: /plan ────── auto (default) or human gate (--interactive)
  ├── Phase 3: /pre-mortem ── auto (FAIL → retry loop)
  ├── Phase 4: /crank ────── autonomous (manages own teams)
  ├── Phase 5: /vibe ──────── auto (FAIL → retry loop)
  └── Phase 6: /post-mortem ── auto (council + retro + flywheel)
```

**Human gates (default):** 0 — fully autonomous, all gates are retry loops
**Human gates (--interactive mode):** 2 (research + plan approval, owned by those skills)
**Retry gates:** pre-mortem FAIL → re-plan, vibe FAIL → re-crank, crank BLOCKED/PARTIAL → re-crank (max 3 attempts each)
**Gate 4 (optional):** post-mortem FAIL → spawn another /rpi cycle (enabled via `--loop`)

Read `references/phase-data-contracts.md` for the full phase-to-phase data contract table.

## Execution Steps

Given `/rpi <goal | epic-id> [--from=<phase>] [--interactive]`:

### Step 0: Setup

```bash
mkdir -p .agents/rpi
```

**Large-repo mode (recommended when context is tight):**

If the repo is large (for example, >1500 tracked files), initialize deterministic
context-window shards before Phase 1:

```bash
scripts/rpi/context-window-contract.sh
```

Read `references/context-windowing.md` for shard generation, progress tracking,
and bounded one-shard-at-a-time execution.

Determine the starting phase:
- Default: Phase 1 (research)
- `--from=plan`: Start at Phase 2
- `--from=pre-mortem`: Start at Phase 3
- `--from=crank`: Start at Phase 4 (requires epic-id)
- `--from=vibe`: Start at Phase 5
- `--from=post-mortem`: Start at Phase 6

If input looks like an epic-id (matches `ag-*` or similar bead prefix pattern), treat it as an existing epic and skip to the appropriate phase (default: crank if no --from specified).

**Check for harvested next-work from prior RPI cycles:**

```bash
if [ -f .agents/rpi/next-work.jsonl ]; then
  # Read unconsumed entries (consumed: false)
  # Filter by target_repo: include entries where target_repo matches
  # the current repo name OR target_repo is "*" (wildcard).
  # Skip entries whose target_repo names a different repo.
  # Entries with no target_repo field are treated as matching (backward compat).
  # Current repo is derived from: basename of git remote origin URL, or
  # failing that, basename of the working directory.
  # Schema: .agents/rpi/next-work.schema.md
fi
```

If unconsumed, repo-matched entries exist in `.agents/rpi/next-work.jsonl`:
- In `--auto` mode: use the highest-severity item's title as the goal (no user prompt)
- In `--interactive` mode: present items via AskUserQuestion and let user choose or provide custom goal
- If goal was already provided by the user, ignore next-work.jsonl (explicit goal takes precedence)

Initialize state:
```
rpi_state = {
  goal: "<goal string>",
  epic_id: null,     # populated after Phase 2
  phase: "<starting phase>",
  auto: <true unless --interactive flag present>,
  test_first: <true if --test-first flag present>,
  complexity: null,  # auto-detected after Phase 2: "low", "medium", or "high"
  cycle: 1,          # RPI iteration number (incremented on --spawn-next)
  parent_epic: null,  # epic ID from prior cycle (if spawned from next-work)
  verdicts: {}       # populated as phases complete
}
```

### Phase 1: Research

**Skip if:** `--from` is set to a later phase.

```
Skill(skill="research", args="<goal> --auto")   # always --auto unless --interactive
```

By default, /research runs with `--auto` (skips human gate, proceeds automatically).
With `--interactive`, the research skill shows its human gate (AskUserQuestion). /rpi trusts the outcome:
- User approves → research complete, proceed
- User abandons → /rpi stops with message: "Research abandoned by user."

**After research completes:**
1. Record: which research file was produced
2. Write phase summary (keep context lean):
   ```
   Read the research output file.
   Write a 500-token summary to .agents/rpi/phase-1-summary-YYYY-MM-DD-<goal-slug>.md
   ```
3. Ratchet checkpoint:
   ```bash
   ao ratchet record research 2>/dev/null || true
   bash scripts/checkpoint-commit.sh rpi "phase-1" "research complete" 2>/dev/null || true
   bash scripts/log-telemetry.sh rpi phase-complete phase=1 phase_name=research 2>/dev/null || true
   ```

### Phase 2: Plan

**Skip if:** `--from` is set to a later phase.

```
Skill(skill="plan", args="<goal> --auto")   # always --auto unless --interactive
```

By default, /plan runs with `--auto` (skips human gate, proceeds automatically).
With `--interactive`, the plan skill shows its human gate. /rpi trusts the outcome.

**After plan completes:**
1. Extract epic-id:
   ```bash
   # Find most recent epic
   EPIC_ID=$(bd list --type epic --status open 2>/dev/null | head -1 | grep -o 'ag-[a-z0-9]*')
   ```
   Store in `rpi_state.epic_id`.

2. **Detect complexity level:**

   Read `references/complexity-scaling.md` for the classification table.

   ```bash
   ISSUE_COUNT=$(bd children <epic-id> 2>/dev/null | wc -l | tr -d ' ')
   WAVE_COUNT=<derived from dependency depth>
   # Guard: if bd unavailable, ISSUE_COUNT=0 is misleading — default to medium
   if ! command -v bd &>/dev/null; then
     log "Warning: bd unavailable — defaulting complexity to medium"
     rpi_state.complexity = "medium"  # skip auto-detection
   fi
   ```

   Compute complexity (skip if already set by bd-unavailable guard):
   - If `--fast-path` flag set → `rpi_state.complexity = "low"`
   - Else if `--deep` flag set → `rpi_state.complexity = "high"`
   - **Error if both `--fast-path` and `--deep` are set** — mutually exclusive, exit with error
   - Else if ISSUE_COUNT <= 2 AND WAVE_COUNT == 1 → `rpi_state.complexity = "low"`
   - Else if ISSUE_COUNT >= 7 OR WAVE_COUNT >= 3 → `rpi_state.complexity = "high"`
   - Else → `rpi_state.complexity = "medium"`

   Log: "Complexity: {level} ({issue_count} issues, {wave_count} waves)"

3. Write phase summary to `.agents/rpi/phase-2-summary-YYYY-MM-DD-<goal-slug>.md`

4. Ratchet checkpoint:
   ```bash
   ao ratchet record plan 2>/dev/null || true
   bash scripts/checkpoint-commit.sh rpi "phase-2" "plan complete" 2>/dev/null || true
   bash scripts/log-telemetry.sh rpi phase-complete phase=2 phase_name=plan 2>/dev/null || true
   ```

### Phase 3: Pre-mortem

**Skip if:** `--from` is set to a later phase.

```
Skill(skill="pre-mortem", args="--quick")   # if rpi_state.complexity == "low" (inline, no spawning)
Skill(skill="pre-mortem", args="--quick")   # if rpi_state.complexity == "medium" (inline, fast default)
Skill(skill="pre-mortem")                    # if rpi_state.complexity == "high" (full 2-judge council)
```

**Rationale:** Pre-mortem's `--quick` mode (inline single-agent structured review) catches the same class of plan bugs as full council at ~10% of the cost. Full council is reserved for high-complexity epics (7+ issues or 3+ waves) where the stakes justify multi-judge overhead. Use `--deep` explicitly for critical migrations or security plans.

Pre-mortem auto-discovers the most recent plan. No args needed.

**After pre-mortem completes:** Extract verdict (PASS/WARN/FAIL) from council report. Read `references/gate-retry-logic.md` for detailed Pre-mortem gate retry behavior.

- PASS/WARN: auto-proceed
- FAIL: re-plan with feedback, re-run pre-mortem (max 3 total attempts)

Store verdict in `rpi_state.verdicts.pre_mortem`. Write phase summary to `.agents/rpi/phase-3-summary-YYYY-MM-DD-<goal-slug>.md`.

Ratchet checkpoint:
```bash
ao ratchet record pre-mortem 2>/dev/null || true
bash scripts/checkpoint-commit.sh rpi "phase-3" "pre-mortem complete" 2>/dev/null || true
bash scripts/log-telemetry.sh rpi phase-complete phase=3 phase_name=pre-mortem 2>/dev/null || true
```

### Phase 4: Crank (Implementation)

**Requires:** `rpi_state.epic_id` (from Phase 2 or --from=crank with epic-id argument)

```
Skill(skill="crank", args="<epic-id> --test-first")   # if --test-first set
Skill(skill="crank", args="<epic-id>")                 # otherwise
```

Crank manages its own waves, teams, and internal validation. /rpi waits for completion.

**After crank completes:** Check `<promise>` tags for completion status. Read `references/gate-retry-logic.md` for detailed Crank gate retry behavior.

- DONE: proceed to Phase 5
- BLOCKED/PARTIAL: re-crank with context (max 3 total attempts)

Write phase summary to `.agents/rpi/phase-4-summary-YYYY-MM-DD-<goal-slug>.md`.

Ratchet checkpoint:
```bash
ao ratchet record implement 2>/dev/null || true
bash scripts/checkpoint-commit.sh rpi "phase-4" "crank complete" 2>/dev/null || true
bash scripts/log-telemetry.sh rpi phase-complete phase=4 phase_name=crank 2>/dev/null || true
```

### Phase 5: Final Vibe

```
Skill(skill="vibe", args="--quick recent")   # if rpi_state.complexity == "low" (inline, no spawning)
Skill(skill="vibe", args="--quick recent")   # if rpi_state.complexity == "medium" (inline, fast default)
Skill(skill="vibe", args="recent")            # if rpi_state.complexity == "high" (full 2-judge council)
```

**Rationale:** Vibe's `--quick` mode (inline single-agent review + complexity analysis) catches the same bugs at ~10% cost. Full council reserved for high-complexity epics. The complexity analysis and mechanical checks still run in --quick mode — only multi-agent spawning is skipped.

Vibe runs complexity analysis + council on recent changes (cross-wave consistency check).

**After vibe completes:** Extract verdict (PASS/WARN/FAIL) and apply gate logic. Read `references/gate-retry-logic.md` for detailed Vibe gate retry behavior.

- PASS/WARN: auto-proceed
- FAIL: re-crank with feedback, re-vibe (max 3 total attempts)

Store verdict in `rpi_state.verdicts.vibe`. Write phase summary to `.agents/rpi/phase-5-summary-YYYY-MM-DD-<goal-slug>.md`.

Ratchet checkpoint:
```bash
ao ratchet record vibe 2>/dev/null || true
bash scripts/checkpoint-commit.sh rpi "phase-5" "vibe complete" 2>/dev/null || true
bash scripts/log-telemetry.sh rpi phase-complete phase=5 phase_name=vibe 2>/dev/null || true
```

### Phase 6: Post-mortem

```
Skill(skill="post-mortem", args="--quick <epic-id>")   # if rpi_state.complexity == "low" (inline, no spawning)
Skill(skill="post-mortem", args="--quick <epic-id>")   # if rpi_state.complexity == "medium" (inline, fast default)
Skill(skill="post-mortem", args="<epic-id>")            # if rpi_state.complexity == "high" (full 2-judge council)
```

**Rationale:** Post-mortem's primary value is the retro + flywheel feed, not the council verdict. `--quick` still extracts learnings and feeds the flywheel — it just skips multi-agent spawning. Full council reserved for high-complexity epics where cross-model consensus on lessons learned is worth the cost.

Post-mortem runs council + retro + flywheel feed. By default, /rpi ends after post-mortem (enable Gate 4 loop via `--loop`).

**After post-mortem completes:**
1. Ratchet checkpoint (with cycle lineage):
   ```bash
   ao ratchet record post-mortem --cycle=<rpi_state.cycle> --parent-epic=<rpi_state.parent_epic> 2>/dev/null || true
   bash scripts/checkpoint-commit.sh rpi "phase-6" "post-mortem complete" 2>/dev/null || true
   bash scripts/log-telemetry.sh rpi phase-complete phase=6 phase_name=post-mortem 2>/dev/null || true
   ```

Read `references/gate4-loop-and-spawn.md` for Gate 4 loop (`--loop`) and spawn-next work details.

### Step Final: Report

Read `references/report-template.md` for the full report and flywheel output templates.

Read `references/error-handling.md` for error handling details.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--from=<phase>` | `research` | Start from this phase (research, plan, pre-mortem, crank, vibe, post-mortem) |
| `--interactive` | off | Enable human gates in /research and /plan. Without this flag, /rpi runs fully autonomous. |
| `--auto` | on | (Legacy, now default) Fully autonomous — zero human gates. Kept for backwards compatibility. |
| `--loop` | off | Enable Gate 4 loop: after /post-mortem, iterate only when post-mortem verdict is FAIL (spawns another /rpi cycle). |
| `--max-cycles=<n>` | `1` | Hard cap on total /rpi cycles when `--loop` is set (recommended: 3). |
| `--spawn-next` | off | After post-mortem, read harvested next-work items and report suggested next `/rpi` command. Marks consumed entries. |
| `--test-first` | off | Pass `--test-first` to `/crank` for spec-first TDD |
| `--fast-path` | auto | Force low complexity (--quick for gates). Auto-detected when ≤2 issues and 1 wave. Incompatible with `--deep`. |
| `--deep` | auto | Force high complexity (--deep on pre-mortem and vibe). Auto-detected when 7+ issues or 3+ waves. Incompatible with `--fast-path`. |
| `--dry-run` | off | With `--spawn-next`: report items without marking consumed. Useful for testing the consumption flow. |

## Examples

### Full Lifecycle from Scratch

**User says:** `/rpi "add user authentication"`

**What happens:**
1. Phase 1: Research agent explores auth patterns in codebase
2. Phase 2: Plan creates epic `ag-5k2` with 5 issues in 2 waves
3. Phase 3: Pre-mortem validates plan — PASS
4. Phase 4: Crank spawns 5 workers across 2 waves, all complete
5. Phase 5: Vibe validates recent changes — PASS
6. Phase 6: Post-mortem extracts learnings, harvests 2 tech-debt items

**Result:** Auth system implemented end-to-end. Suggested next `/rpi` command for highest-priority tech-debt.

### Resume from Crank

**User says:** `/rpi --from=crank ag-5k2`

**What happens:**
1. Skips research, plan, pre-mortem (already done)
2. Phase 4: Crank resumes or restarts epic `ag-5k2`
3. Phase 5: Vibe validates
4. Phase 6: Post-mortem wraps up

**Result:** Fast resumption for already-planned work.

### Interactive Mode

**User says:** `/rpi --interactive "refactor payment module"`

**What happens:**
1. Phase 1: Research completes, asks for human approval
2. User approves research
3. Phase 2: Plan completes, asks for human approval
4. User approves plan with revisions
5. Phases 3-6: Fully autonomous (pre-mortem auto-retries on FAIL)

**Result:** Human-guided research and planning, autonomous execution.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Pre-mortem retry loop hits max attempts | Plan has fundamental issues that retry loop cannot fix | Review council findings, manually revise plan, re-run `/rpi --from=plan` |
| Vibe retry loop hits max attempts | Implementation has critical flaws that re-crank cannot fix | Review vibe findings, manually fix code, re-run `/rpi --from=vibe` |
| Crank blocks on missing dependency | Epic issue references unavailable blocker | Check `bd show <epic-id>` for dep graph, fix or remove blocker |
| Post-mortem harvests no next-work | Council found no tech debt or improvements | Flywheel stable — no follow-up needed |
| `--loop` causes infinite cycles | Gate 4 loop enabled but post-mortem always returns FAIL | Set `--max-cycles=3` to cap iterations, review why FAIL persists |
| Large repo context overflow | Repo has >1500 files and agents run out of context | Enable context-windowing via `scripts/rpi/context-window-contract.sh` before Phase 1 |

## See Also

- `skills/research/SKILL.md` — Phase 1
- `skills/plan/SKILL.md` — Phase 2
- `skills/pre-mortem/SKILL.md` — Phase 3
- `skills/crank/SKILL.md` — Phase 4
- `skills/vibe/SKILL.md` — Phase 5
- `skills/post-mortem/SKILL.md` — Phase 6
