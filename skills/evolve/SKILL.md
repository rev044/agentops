---
name: evolve
description: Autonomous fitness-scored improvement loop. Measures goals, picks worst gap, runs /rpi, compounds via knowledge flywheel.
tier: orchestration
dependencies:
  - rpi         # required - executes each improvement cycle
  - post-mortem # required - auto-runs at teardown to harvest learnings
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

**Capture session-start SHA** (for multi-commit revert):
```bash
SESSION_START_SHA=$(git rev-parse HEAD)
```

Initialize state:
```
evolve_state = {
  cycle: 0,
  max_cycles: <from flag, default 10>,
  dry_run: <from flag, default false>,
  test_first: <from flag, default false>,
  session_start_sha: $SESSION_START_SHA,
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

Record results with **continuous values** (not just pass/fail):
```bash
# Write fitness snapshot
cat > .agents/evolve/fitness-${CYCLE}.json << EOF
{
  "cycle": $CYCLE,
  "timestamp": "$(date -Iseconds)",
  "cycle_start_sha": "$(git rev-parse HEAD)",
  "goals": [
    {"id": "$goal_id", "result": "$result", "weight": $weight, "value": $metric_value, "threshold": $threshold},
    ...
  ]
}
EOF
```

For goals with measurable metrics, extract the continuous value:
- `go-coverage-floor`: parse `go test -cover` output → `"value": 85.7, "threshold": 80`
- `doc-coverage`: count skills with references/ → `"value": 20, "threshold": 16`
- `shellcheck-clean`: count of warnings → `"value": 0, "threshold": 0`
- Other goals: `"value": null` (binary pass/fail only)

**Snapshot enforcement (HARD GATE):** After writing the snapshot, validate it:
```bash
if ! jq empty ".agents/evolve/fitness-${CYCLE}.json" 2>/dev/null; then
  echo "ERROR: Fitness snapshot write failed or invalid JSON. Refusing to proceed."
  exit 1
fi
```
Do NOT proceed to Step 3 without a valid fitness snapshot.

**Bootstrap mode:** If a check command fails to execute (command not found, permission denied), mark that goal as `"result": "skip"` with a warning. Do NOT block the entire loop because one check is broken.

### Step 3: Select Work

```
failing_goals = [g for g in goals if g.result == "fail"]

if not failing_goals:
  # Before stopping, check harvested work from prior /rpi cycles
  if [ -f .agents/rpi/next-work.jsonl ]; then
    items = read_unconsumed(next-work.jsonl)  # entries with consumed: false
    if items:
      selected_item = max(items, by=severity)  # highest severity first
      log "All goals met. Picking harvested work: {selected_item.title}"
      # Execute as an /rpi cycle (Step 4), then mark consumed
      /rpi "{selected_item.title}" --auto --max-cycles=1 --test-first   # if --test-first set
      /rpi "{selected_item.title}" --auto --max-cycles=1                 # otherwise
      mark_consumed(selected_item)  # set consumed: true, consumed_by, consumed_at
      # Skip Steps 4-5 (already executed above), go to Step 6 (log cycle)
      log_cycle(cycle, goal_id="next-work:{selected_item.title}", result="harvested")
      continue loop  # → Step 1 (kill switch check)

  log "All goals met, no harvested work. Done."
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

**If `--dry-run`:** Report the selected goal (or harvested item) and stop.

```
log "Dry run: would work on '{selected.id}' (weight: {selected.weight})"
log "Description: {selected.description}"
log "Check command: {selected.check}"

# Also show queued harvested work
if [ -f .agents/rpi/next-work.jsonl ]; then
  items = read_unconsumed(next-work.jsonl)
  if items:
    log "Harvested work queue ({len(items)} items):"
    for item in items:
      log "  - [{item.severity}] {item.title} ({item.type})"

STOP → go to Teardown
```

**Otherwise:** Run a full /rpi cycle on the selected goal.

```
/rpi "Improve {selected.id}: {selected.description}" --auto --max-cycles=1 --test-first   # if --test-first set
/rpi "Improve {selected.id}: {selected.description}" --auto --max-cycles=1                 # otherwise
```

This internally runs the full lifecycle:
- `/research` — understand the problem
- `/plan` — decompose into issues
- `/pre-mortem` — validate the plan
- `/crank` — implement (spawns workers)
- `/vibe` — validate the code
- `/post-mortem` — extract learnings + `ao forge` (COMPOUNDING)

**Wait for /rpi to complete before proceeding.**

### Step 5: Full-Fitness Regression Gate

**CRITICAL: Re-run ALL goals, not just the target.**

After /rpi completes, re-run MEASURE_FITNESS on **every goal** (same as Step 2). Write result to `fitness-{CYCLE}-post.json`.

Compare the pre-cycle snapshot (`fitness-{CYCLE}.json`) against the post-cycle snapshot (`fitness-{CYCLE}-post.json`) for **ALL goals**:

```
# Load pre-cycle results
pre_results = load("fitness-{CYCLE}.json")

# Re-measure ALL goals (writes fitness-{CYCLE}-post.json)
post_results = MEASURE_FITNESS()

# Check the target goal
if selected_goal.post_result == "pass":
  outcome = "improved"
else:
  outcome = "unchanged"

# FULL REGRESSION CHECK: compare ALL goals, not just the target
newly_failing = []
for goal in post_results.goals:
  pre = pre_results.find(goal.id)
  if pre.result == "pass" and goal.result == "fail":
    newly_failing.append(goal.id)

if newly_failing:
  outcome = "regressed"
  log "REGRESSION: {newly_failing} started failing after fixing {selected.id}"

  # Multi-commit revert using cycle start SHA
  cycle_start_sha = pre_results.cycle_start_sha
  commit_count = $(git rev-list --count ${cycle_start_sha}..HEAD)
  if commit_count == 0:
    log "No commits to revert"
  elif commit_count == 1:
    git revert HEAD --no-edit
  else:
    git revert --no-commit ${cycle_start_sha}..HEAD
    git commit -m "revert: evolve cycle ${CYCLE} regression in {newly_failing}"
  log "Reverted ${commit_count} commits. Moving to next goal."
```

**Snapshot enforcement:** Validate `fitness-{CYCLE}-post.json` was written and is valid JSON before proceeding.

### Step 6: Log Cycle

Append to `.agents/evolve/cycle-history.jsonl`:

```jsonl
{"cycle": 1, "goal_id": "test-pass-rate", "result": "improved", "commit_sha": "abc1234", "timestamp": "2026-02-11T21:00:00Z"}
{"cycle": 2, "goal_id": "doc-coverage", "result": "regressed", "commit_sha": "def5678", "reverted_to": "abc1234", "timestamp": "2026-02-11T21:30:00Z"}
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

**Auto-run /post-mortem on the full evolution session:**

```
/post-mortem "evolve session: $CYCLE cycles, goals improved: X, harvested: Y"
```

This captures learnings from the ENTIRE evolution run (all cycles, all /rpi invocations) in one council review. The post-mortem harvests follow-up items into `next-work.jsonl`, feeding the next `/evolve` session.

**Then write session summary:**

```bash
cat > .agents/evolve/session-summary.md << EOF
# /evolve Session Summary

**Date:** $(date -Iseconds)
**Cycles:** $CYCLE of $MAX_CYCLES
**Goals measured:** $(wc -l < GOALS.yaml goals)

## Cycle History
$(cat .agents/evolve/cycle-history.jsonl)

## Final Fitness
$(cat .agents/evolve/fitness-${CYCLE}.json)

## Post-Mortem
<path to post-mortem report from above>

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
Post-mortem: <verdict> (see <report-path>)

Run `/evolve` again to continue improving.
```

---

Read `references/compounding.md` for details on how the knowledge flywheel and work harvesting compound across cycles.

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
| `--test-first` | off | Pass `--test-first` through to `/rpi` → `/crank` |
| `--dry-run` | off | Measure fitness and show plan, don't execute |

---

Read `references/goals-schema.md` for the GOALS.yaml format.

---

## Artifacts

| File | Purpose |
|------|---------|
| `GOALS.yaml` | Fitness goals (repo root) |
| `.agents/evolve/fitness-{N}.json` | Pre-cycle fitness snapshot (continuous values) |
| `.agents/evolve/fitness-{N}-post.json` | Post-cycle fitness snapshot (for regression comparison) |
| `.agents/evolve/cycle-history.jsonl` | Cycle outcomes log (includes commit SHAs) |
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
