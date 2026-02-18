---
name: evolve
description: Goal-driven fitness-scored improvement loop. Measures goals, picks worst gap, runs /rpi, compounds via knowledge flywheel.
metadata:
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

# /evolve — Goal-Driven Compounding Loop

> **Purpose:** Measure what's wrong. Fix the worst thing. Measure again. Compound.

Thin fitness-scored loop over `/rpi`. The knowledge flywheel provides compounding — each cycle loads learnings from all prior cycles.

## Compaction Resilience

The evolve loop MUST survive context compaction. Every cycle commits its
artifacts to git before proceeding. The `cycle-history.jsonl` file is the
recovery point — on session restart, read it to determine cycle number
and resume from Step 1.

## Known Good Properties

- Severity-based selection naturally orders: code health → architecture →
  testing → documentation → cleanup. This is the correct ordering.
  Do not add special-case logic to front-load doc fixes.

**Dormancy is success.** When all goals pass and no harvested work remains, the system enters dormancy — a valid, healthy state. The system does not manufacture work to justify its existence. Nothing to do means everything is working.

## Quick Start

```bash
/evolve                      # Run forever until kill switch or stagnation
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
- `--max-cycles=N` (default: **unlimited**) — optional hard cap. Without this flag, the loop runs **forever** until kill switch or stagnation.
- `--dry-run` — measure and report only, no execution
- `--skip-baseline` — skip the Step 0.5 baseline sweep

**Capture session-start SHA** (for multi-commit revert):
```bash
SESSION_START_SHA=$(git rev-parse HEAD)
```

Initialize state:
```
evolve_state = {
  cycle: 0,
  max_cycles: <from flag, or Infinity if not set>,
  dry_run: <from flag, default false>,
  test_first: <from flag, default false>,
  session_start_sha: $SESSION_START_SHA,
  idle_streak: 0,         # consecutive cycles with nothing to do
  max_idle_streak: 3,     # stop after this many consecutive idle cycles
  history: []
}
```

### Step 0.5: Cycle-0 Baseline Sweep

Capture a baseline fitness snapshot before the first cycle so every later cycle
has a comparison anchor.  Skipped on resume (idempotent).

```
if [ "$SKIP_BASELINE" = "true" ]; then
  log "Skipping baseline sweep (--skip-baseline flag set)"
  exit 0
fi

if ! [ -f .agents/evolve/fitness-0-baseline.json ]; then
  # **Preferred (when ao CLI available):**
  if command -v ao &>/dev/null; then
    ao goals measure --json > .agents/evolve/fitness-0-baseline.json
  fi

  # **Fallback (no ao CLI):**
  baseline = MEASURE_FITNESS()            # run every GOALS.yaml goal
  baseline.cycle = 0
  write ".agents/evolve/fitness-0-baseline.json" baseline

  # Baseline report
  failing = [g for g in baseline.goals if g.result == "fail"]
  failing.sort(by=weight, descending)
  cat > .agents/evolve/cycle-0-report.md << EOF
  # Cycle-0 Baseline
  **Total goals:** ${len(baseline.goals)}
  **Passing:** ${len(baseline.goals) - len(failing)}
  **Failing:** ${len(failing)}
  $(for g in failing: "- [weight ${g.weight}] ${g.id}: ${g.result}")
  EOF

  log "Baseline captured: ${len(failing)}/${len(baseline.goals)} goals failing"
fi

# Wiring closure check: every check-*.sh must appear in GOALS.yaml
unwired=$(comm -23 \
  <(ls scripts/check-*.sh 2>/dev/null | xargs -I{} basename {} | sort) \
  <(grep -oP 'scripts/check-\S+\.sh' GOALS.yaml | xargs -I{} basename {} | sort))
if [ -n "$unwired" ]; then
  for script in $unwired; do
    add_to_next_work("Unwired script: $script — wire to GOALS.yaml or delete",
                     severity="high", type="tech-debt")
  done
  log "Found $(echo "$unwired" | wc -l | tr -d ' ') unwired scripts — added to next-work"
fi
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

Read `GOALS.yaml` from repo root.

**Preferred (when ao CLI available):**
```bash
if command -v ao &>/dev/null; then
  ao goals measure --json > .agents/evolve/fitness-${CYCLE}-snapshot.json
fi
```

**Fallback (no ao CLI):**

For each goal:

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
  # Cycle-0 Comprehensive Sweep (optional full-repo scan)
  # Before consuming harvested work, optionally discover items the harvest missed.
  # This is because manual sweeps have found issues automated harvests didn't catch.

  if ! [ -f .agents/evolve/last-sweep-date ] || \
     [ $(date +%s) -gt $(( $(stat -f %m .agents/evolve/last-sweep-date) + 604800 )) ]; then
    # Sweep is stale (> 7 days) or missing — run lightweight scan
    log "Running cycle-0 comprehensive sweep (stale/missing: .agents/evolve/last-sweep-date)"

    # Lightweight sweep: shellcheck, go vet, known anti-patterns
    shellcheck hooks/*.sh 2>&1 | grep -v "^$" | while read line; do
      add_to_next_work("shellcheck finding: $line", severity="medium", type="bug")
    done

    go vet ./cli/... 2>&1 | grep -v "^$" | while read line; do
      add_to_next_work("go vet finding: $line", severity="medium", type="bug")
    done

    # grep for known anti-patterns (e.g., hardcoded secrets, TODO markers)
    grep -r "TODO|FIXME|XXX" --include="*.go" --include="*.sh" . 2>/dev/null | while read line; do
      add_to_next_work("code marker: $line", severity="low", type="tech-debt")
    done

    # Coverage floor single-pass: scan ALL packages at once
    # When discovering coverage floor gaps, process everything in one cycle.
    if [ -f GOALS.yaml ] && grep -q "coverage-floor" GOALS.yaml; then
      # Run coverage for ALL packages in a single pass
      go test -cover ./... 2>/dev/null | grep -E "^ok|^FAIL" | while read line; do
        pkg=$(echo "$line" | awk '{print $2}')
        cov=$(echo "$line" | grep -oP '\d+\.\d+%' | tr -d '%')
        if [ -n "$cov" ]; then
          # Check if package has a floor in GOALS.yaml
          # If not tracked, add it; if tracked with >3% headroom, tighten it
          add_to_next_work("Coverage floor check: $pkg at ${cov}%",
                           severity="medium", type="coverage-floor")
        fi
      done
      log "Coverage floor single-pass complete for all packages"
    fi

    # Mark sweep complete
    touch .agents/evolve/last-sweep-date
    log "Cycle-0 sweep complete. New findings added to next-work.jsonl"
  fi

  # Coverage floor processing guidance:
  # When the explore agent or sweep finds coverage floor headroom:
  # - Run coverage for ALL packages in a single pass
  # - Compare ALL floors to actual in one table
  # - Tighten ALL floors with >3% headroom in a single cycle
  # - Add ALL untracked packages in the same cycle
  # Do NOT split floor-tightening across multiple cycles.

  # All goals pass — check harvested work from prior /rpi cycles
  if [ -f .agents/rpi/next-work.jsonl ]; then
    # Detect current repo for filtering
    CURRENT_REPO=$(bd config --get prefix 2>/dev/null \
      || basename "$(git remote get-url origin 2>/dev/null)" .git 2>/dev/null \
      || basename "$(pwd)")

    all_items = read_unconsumed(next-work.jsonl)  # entries with consumed: false
    # Filter by target_repo: include items where target_repo matches
    # CURRENT_REPO, target_repo is "*" (cross-repo), or field is absent (backward compat).
    # Skip items whose target_repo names a different repo.
    items = [i for i in all_items
             if i.target_repo in (CURRENT_REPO, "*", None)]
    if items:
      evolve_state.idle_streak = 0  # reset — we found work
      selected_item = max(items, by=severity)  # highest severity first
      log "All goals met. Picking harvested work: {selected_item.title}"
      # Execute as an /rpi cycle (Step 4), then mark consumed
      /rpi "{selected_item.title}" --auto --max-cycles=1 --test-first   # if --test-first set
      /rpi "{selected_item.title}" --auto --max-cycles=1                 # otherwise
      mark_consumed(selected_item)  # set consumed: true, consumed_by, consumed_at
      # Skip Steps 4-5 (already executed above), go to Step 6 (log cycle)
      log_cycle(cycle, goal_id="next-work:{selected_item.title}", result="harvested")
      continue loop  # → Step 1 (kill switch check)

  # Nothing to do THIS cycle — but don't quit yet
  evolve_state.idle_streak += 1
  log "All goals met, no harvested work. Idle streak: {idle_streak}/{max_idle_streak}"

  if evolve_state.idle_streak >= evolve_state.max_idle_streak:
    log "Stagnation: {max_idle_streak} consecutive idle cycles. Nothing left to improve."
    STOP → go to Teardown

  # NOT stagnant yet — re-measure next cycle (external changes, new harvested work)
  log "Re-measuring next cycle in case conditions changed..."
  continue loop  # → Step 1 (kill switch check)

# Meta-goal guidance: after pruning any allowlist, add a meta-goal that
# prevents re-accumulation. The meta-goal should fail if allowlist entries
# have callers. Allowlists without meta-goals are technical debt magnets.
# See references/goals-schema.md for the meta-goal pattern.

# We have failing goals — reset idle streak
evolve_state.idle_streak = 0

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

# Also show queued harvested work (filtered to current repo)
if [ -f .agents/rpi/next-work.jsonl ]; then
  all_items = read_unconsumed(next-work.jsonl)
  items = [i for i in all_items
           if i.target_repo in (CURRENT_REPO, "*", None)]
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
{"cycle": 1, "goal_id": "test-pass-rate", "result": "improved", "commit_sha": "abc1234", "goals_passing": 18, "goals_total": 23, "timestamp": "2026-02-11T21:00:00Z"}
{"cycle": 2, "goal_id": "doc-coverage", "result": "regressed", "commit_sha": "def5678", "reverted_to": "abc1234", "goals_passing": 17, "goals_total": 23, "timestamp": "2026-02-11T21:30:00Z"}
```

**MANDATORY fields in every cycle log entry:**
- `goals_passing`: count of goals with result "pass"
- `goals_total`: total goals measured
- `goals_added`: count of new goals added this cycle (0 if none)

These enable fitness trajectory plotting across cycles.

**Telemetry logging (end of each cycle):**
```bash
bash scripts/log-telemetry.sh evolve cycle-complete cycle=${CYCLE} score=${SCORE} goals_passing=${PASSING} goals_total=${TOTAL}
```

**Compaction-proofing: commit after every cycle.**
Uncommitted state does not survive context compaction. ALWAYS commit cycle
artifacts before starting the next cycle:

```bash
git add .agents/evolve/cycle-history.jsonl .agents/evolve/fitness-*.json
git commit -m "evolve: cycle ${CYCLE} — ${selected.id} ${outcome}" --allow-empty
```

### Step 7: Loop or Stop

```
evolve_state.cycle += 1

# Only stop for max-cycles if the user explicitly set one
if evolve_state.max_cycles != Infinity and evolve_state.cycle >= evolve_state.max_cycles:
  log "Max cycles ({max_cycles}) reached."
  STOP → go to Teardown

# Otherwise: loop back to Step 1 (kill switch check) — run forever
```

### Teardown

Read `references/teardown.md` for the full teardown procedure: post-mortem on the full evolution session, fitness trajectory computation, and session summary generation.

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
| `--max-cycles=N` | unlimited | Optional hard cap. Without this, loop runs forever. |
| `--test-first` | off | Pass `--test-first` through to `/rpi` → `/crank` |
| `--dry-run` | off | Measure fitness and show plan, don't execute |
| `--skip-baseline` | off | Skip cycle-0 baseline sweep |

---

Read `references/goals-schema.md` for the GOALS.yaml format.

---

## Artifacts

See `references/artifacts.md` for the full list of generated files and their purposes.

---

## Examples

See `references/examples.md` for detailed examples including infinite improvement, dry-run mode, and regression with revert.

---

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `/evolve` exits immediately with "KILL SWITCH ACTIVE" | Kill switch file exists | Remove `~/.config/evolve/KILL` or `.agents/evolve/STOP` to re-enable |
| "No goals to measure" error | GOALS.yaml missing or empty | Create GOALS.yaml in repo root with fitness goals (see references/goals-schema.md) |
| Cycle completes but fitness unchanged | Goal check command is always passing or always failing | Verify check command logic in GOALS.yaml produces exit code 0 (pass) or non-zero (fail) |
| Regression revert fails | Multiple commits in cycle or uncommitted changes | Check cycle-start SHA in fitness snapshot, commit or stash changes before retrying |
| Harvested work never consumed | All goals passing but `next-work.jsonl` not read | Check file exists and has `consumed: false` entries. Agent picks harvested work after goals met. |
| Loop stops after N cycles | `--max-cycles` was set (or old default of 10) | Omit `--max-cycles` flag — default is now unlimited. Loop runs until kill switch or 3 idle cycles. |

---

## See Also

- `skills/rpi/SKILL.md` — Full lifecycle orchestrator (called per cycle)
- `skills/vibe/SKILL.md` — Code validation (called by /rpi)
- `skills/council/SKILL.md` — Multi-model judgment (called by /rpi)
- `GOALS.yaml` — Fitness goals for this repo
