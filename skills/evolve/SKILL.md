---
name: evolve
description: Autonomous fitness-scored improvement loop. Measures goals, picks worst gap, runs /rpi, compounds via knowledge flywheel.
tier: orchestration
triggers:
  - evolve
  - improve everything
  - autonomous improvement
  - run until done
---

# /evolve — Autonomous Compounding Loop

> **Purpose:** Measure what's wrong. Fix the worst thing. Measure again. Compound.

Thin fitness-scored loop over `/rpi`. The knowledge flywheel provides compounding — each cycle loads learnings from all prior cycles.

## Quick Start

```bash
/evolve                      # Run until all goals met or --max-cycles hit
/evolve --max-cycles=5       # Cap at 5 improvement cycles
/evolve --dry-run            # Measure fitness, show what would be worked on, don't execute
```

## Execution Steps

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

### Step 0: Setup

```bash
mkdir -p .agents/evolve
```

Load accumulated learnings (COMPOUNDING):
```bash
ao inject 2>/dev/null || true
```

Parse flags:
- `--max-cycles=N` (default: 10) — hard cap on improvement cycles
- `--dry-run` — measure and report only, no execution

Initialize state:
```
evolve_state = {
  cycle: 0,
  max_cycles: <from flag, default 10>,
  dry_run: <from flag, default false>,
  history: []
}
```

### Step 1: Kill Switch Check

Check at the TOP of every cycle iteration:

```bash
# External kill (outside repo — can't be accidentally deleted by agents)
if [ -f ~/.config/evolve/KILL ]; then
  echo "KILL SWITCH ACTIVE: $(cat ~/.config/evolve/KILL)"
  # Write acknowledgment
  echo "{\"killed_at\": \"$(date -Iseconds)\", \"cycle\": $CYCLE}" > .agents/evolve/KILLED.json
  exit 0
fi

# Local convenience stop
if [ -f .agents/evolve/STOP ]; then
  echo "STOP file detected: $(cat .agents/evolve/STOP 2>/dev/null)"
  exit 0
fi
```

If either file exists, log reason and **stop immediately**. Do not proceed to measurement.

### Step 2: Measure Fitness (MEASURE_FITNESS)

Read `GOALS.yaml` from repo root. For each goal:

```bash
# Run the check command
if eval "$goal_check" > /dev/null 2>&1; then
  # Exit code 0 = PASS
  result = "pass"
else
  # Non-zero = FAIL
  result = "fail"
fi
```

Record results:
```bash
# Write fitness snapshot
cat > .agents/evolve/fitness-${CYCLE}.json << EOF
{
  "cycle": $CYCLE,
  "timestamp": "$(date -Iseconds)",
  "goals": [
    {"id": "$goal_id", "result": "$result", "weight": $weight},
    ...
  ]
}
EOF
```

**Bootstrap mode:** If a check command fails to execute (command not found, permission denied), mark that goal as `"result": "skip"` with a warning. Do NOT block the entire loop because one check is broken.

### Step 3: Select Work

```
failing_goals = [g for g in goals if g.result == "fail"]

if not failing_goals:
  log "All goals met. Nothing to improve."
  STOP → go to Teardown

# Sort by weight (highest priority first)
failing_goals.sort(by=weight, descending)

# Simple strike check: skip goals that failed the last 3 consecutive cycles
for goal in failing_goals:
  recent = last_3_cycles_for(goal.id)
  if all(r.result == "regressed" for r in recent):
    log "Skipping {goal.id}: regressed 3 consecutive cycles. Needs human attention."
    continue
  selected = goal
  break

if no goal selected:
  log "All failing goals have regressed 3+ times. Human intervention needed."
  STOP → go to Teardown
```

### Step 4: Execute

**If `--dry-run`:** Report the selected goal and stop.

```
log "Dry run: would work on '{selected.id}' (weight: {selected.weight})"
log "Description: {selected.description}"
log "Check command: {selected.check}"
STOP → go to Teardown
```

**Otherwise:** Run a full /rpi cycle on the selected goal.

```
/rpi "Improve {selected.id}: {selected.description}" --auto --max-cycles=1
```

This internally runs the full lifecycle:
- `/research` — understand the problem
- `/plan` — decompose into issues
- `/pre-mortem` — validate the plan
- `/crank` — implement (spawns workers)
- `/vibe` — validate the code
- `/post-mortem` — extract learnings + `ao forge` (COMPOUNDING)

**Wait for /rpi to complete before proceeding.**

### Step 5: Re-Measure and Check Regression

After /rpi completes, re-run MEASURE_FITNESS (same as Step 2).

Compare before/after:

```
# Check the target goal
if selected_goal.result == "pass":
  outcome = "improved"
elif selected_goal.result == "fail":
  # Check if OTHER goals regressed
  newly_failing = [g for g in goals if g.was_passing_before and g.result == "fail"]
  if newly_failing:
    outcome = "regressed"
    log "REGRESSION: {newly_failing} started failing after fixing {selected.id}"
    # Revert: find the most recent merge commit and revert it
    git log --oneline -5  # Find the merge
    git revert HEAD --no-edit  # Revert the last commit
    log "Reverted regression. Moving to next goal."
  else:
    outcome = "unchanged"
```

### Step 6: Log Cycle

Append to `.agents/evolve/cycle-history.jsonl`:

```jsonl
{"cycle": 1, "goal_id": "test-pass-rate", "result": "improved", "timestamp": "2026-02-11T21:00:00Z"}
{"cycle": 2, "goal_id": "doc-coverage", "result": "regressed", "timestamp": "2026-02-11T21:30:00Z"}
```

### Step 7: Loop or Stop

```
evolve_state.cycle += 1

if evolve_state.cycle >= evolve_state.max_cycles:
  log "Max cycles ({max_cycles}) reached."
  STOP → go to Teardown

# Otherwise: loop back to Step 1 (kill switch check)
```

### Teardown

```bash
# Write session summary
cat > .agents/evolve/session-summary.md << EOF
# /evolve Session Summary

**Date:** $(date -Iseconds)
**Cycles:** $CYCLE of $MAX_CYCLES
**Goals measured:** $(wc -l < GOALS.yaml goals)

## Cycle History
$(cat .agents/evolve/cycle-history.jsonl)

## Final Fitness
$(cat .agents/evolve/fitness-${CYCLE}.json)

## Next Steps
- Run \`/evolve\` again to continue improving
- Run \`/evolve --dry-run\` to check current fitness without executing
- Create \`~/.config/evolve/KILL\` to prevent future runs
- Create \`.agents/evolve/STOP\` for a one-time local stop
EOF
```

Report to user:
```
## /evolve Complete

Cycles: N of M
Goals improved: X
Goals regressed: Y (reverted)
Goals unchanged: Z

Run `/evolve` again to continue, or `/post-mortem` to wrap up.
```

---

## How Compounding Works

```
Session 1:
  ao inject (nothing yet)         → cycle runs blind
  /rpi fixes test-pass-rate       → post-mortem runs ao forge
  Learnings extracted: "tests/skills/run-all.sh validates frontmatter"

Session 2:
  ao inject (loads Session 1 learnings)  → cycle knows about frontmatter validation
  /rpi fixes doc-coverage                → approach informed by prior learning
  Learnings extracted: "references/ dirs need at least one .md file"

Session N:
  ao inject (loads ALL prior learnings)  → system knows patterns from N-1 sessions
  Each cycle is smarter than the last
```

The flywheel is the compounding mechanism. /evolve just feeds it consistently.

---

## Kill Switch

Two paths, checked at every cycle boundary:

| File | Purpose | Who Creates It |
|------|---------|---------------|
| `~/.config/evolve/KILL` | Permanent stop (outside repo) | Human |
| `.agents/evolve/STOP` | One-time local stop | Human or automation |

To stop /evolve:
```bash
echo "Taking a break" > ~/.config/evolve/KILL    # Permanent
echo "done for today" > .agents/evolve/STOP       # Local, one-time
```

To re-enable:
```bash
rm ~/.config/evolve/KILL
rm .agents/evolve/STOP
```

---

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--max-cycles=N` | 10 | Hard cap on improvement cycles per session |
| `--dry-run` | off | Measure fitness and show plan, don't execute |

---

## GOALS.yaml Schema

```yaml
version: 1
mission: "What this repo does"

goals:
  - id: unique-identifier
    description: "Human-readable description"
    check: "shell command — exit 0 = pass, non-zero = fail"
    weight: 1-10  # Higher = fix first
```

Goals are checked in weight order (highest first). The first failing goal with the highest weight is selected for improvement.

---

## Artifacts

| File | Purpose |
|------|---------|
| `GOALS.yaml` | Fitness goals (repo root) |
| `.agents/evolve/fitness-{N}.json` | Fitness snapshot per cycle |
| `.agents/evolve/cycle-history.jsonl` | Cycle outcomes log |
| `.agents/evolve/session-summary.md` | Session wrap-up |
| `.agents/evolve/STOP` | Local kill switch |
| `.agents/evolve/KILLED.json` | Kill acknowledgment |
| `~/.config/evolve/KILL` | External kill switch |

---

## See Also

- `skills/rpi/SKILL.md` — Full lifecycle orchestrator (called per cycle)
- `skills/vibe/SKILL.md` — Code validation (called by /rpi)
- `skills/council/SKILL.md` — Multi-model judgment (called by /rpi)
- `GOALS.yaml` — Fitness goals for this repo
