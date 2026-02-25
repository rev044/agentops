---
name: evolve
description: 'Goal-driven fitness-scored improvement loop. Measures goals, picks worst gap, runs $rpi, compounds via knowledge flywheel. Also pulls from open beads when goals all pass.'
---


# $evolve — Goal-Driven Compounding Loop

> Measure what's wrong. Fix the worst thing. Measure again. Compound.

Thin fitness-scored loop over `$rpi`. Three work sources in priority order:
1. **Failing GOALS.yaml goals** (highest-weight first)
2. **Open beads issues** (`bd ready` — when all goals pass)
3. **Harvested next-work.jsonl** (from prior $rpi post-mortems)

**Dormancy is success.** When all sources are empty, stop. Don't manufacture work.

```bash
$evolve                      # Run until kill switch or stagnation
$evolve --max-cycles=5       # Cap at 5 cycles
$evolve --dry-run            # Show what would be worked on, don't execute
$evolve --beads-only         # Skip goals measurement, work beads backlog only
$evolve --quality            # Quality-first mode: prioritize post-mortem findings
$evolve --quality --max-cycles=10  # Quality mode with cycle cap
```

## Execution Steps

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

### Step 0: Setup

```bash
mkdir -p .agents/evolve
ao inject 2>/dev/null || true
```

Recover cycle number and idle streak from disk (survives context compaction):
```bash
if [ -f .agents/evolve/cycle-history.jsonl ]; then
  CYCLE=$(( $(tail -1 .agents/evolve/cycle-history.jsonl | jq -r '.cycle // 0') + 1 ))
else
  CYCLE=1
fi
SESSION_START_SHA=$(git rev-parse HEAD)

# Recover idle streak from disk (not in-memory — survives compaction)
# Portable: forward-scanning awk counts trailing idle run without tac (unavailable on stock macOS)
IDLE_STREAK=$(awk '/"result"\s*:\s*"(idle|unchanged)"/{streak++; next} {streak=0} END{print streak+0}' \
  .agents/evolve/cycle-history.jsonl 2>/dev/null)

PRODUCTIVE_THIS_SESSION=0

# Circuit breaker: stop if last productive cycle was >60 minutes ago
LAST_PRODUCTIVE_TS=$(grep -v '"idle"\|"unchanged"' .agents/evolve/cycle-history.jsonl 2>/dev/null \
  | tail -1 | jq -r '.timestamp // empty')
if [ -n "$LAST_PRODUCTIVE_TS" ]; then
  NOW_EPOCH=$(date +%s)
  LAST_EPOCH=$(date -j -f "%Y-%m-%dT%H:%M:%S%z" "$LAST_PRODUCTIVE_TS" +%s 2>/dev/null \
    || date -d "$LAST_PRODUCTIVE_TS" +%s 2>/dev/null || echo 0)
  if [ "$LAST_EPOCH" -gt 1000000000 ] && [ $((NOW_EPOCH - LAST_EPOCH)) -ge 3600 ]; then
    echo "CIRCUIT BREAKER: No productive work in 60+ minutes. Stopping."
    # go to Teardown
  fi
fi

# Track oscillating goals (improved→fail→improved→fail) to avoid burning cycles
declare -A QUARANTINED_GOALS  # goal_id → true if oscillation count >= 3
```

Parse flags: `--max-cycles=N` (default unlimited), `--dry-run`, `--beads-only`, `--skip-baseline`, `--quality`.

### Step 0.5: Baseline (first run only)

Skip if `--skip-baseline` or `--beads-only` or baseline already exists.

```bash
if [ ! -f .agents/evolve/fitness-0-baseline.json ]; then
  ao goals measure --json --timeout 60 > .agents/evolve/fitness-0-baseline.json
fi
```

### Step 1: Kill Switch Check

Run at the TOP of every cycle:

```bash
CYCLE_START_SHA=$(git rev-parse HEAD)
[ -f ~/.config/evolve/KILL ] && echo "KILL: $(cat ~/.config/evolve/KILL)" && exit 0
[ -f .agents/evolve/STOP ] && echo "STOP: $(cat .agents/evolve/STOP 2>/dev/null)" && exit 0
```

### Step 2: Measure Fitness

Skip if `--beads-only`.

```bash
ao goals measure --json --timeout 60 > .agents/evolve/fitness-latest.json
```

**Do NOT write per-cycle `fitness-{N}-pre.json` files.** The rolling file is sufficient for work selection and regression detection.

This writes a fitness snapshot to `.agents/evolve/`. If `ao goals measure` is unavailable, read `GOALS.yaml` and run each goal's `check` command manually. Mark timeouts as `"result": "skip"`.

### Step 3: Select Work

**Step 3.0: Emergency gates** (skip if `--beads-only`):
```bash
# Check for failing gates with weight >= 5
FAILING=$(jq -r '.goals[] | select(.result=="fail" and .weight>=5) | .id' .agents/evolve/fitness-latest.json | head -1)
```
Pick highest-weight failing goal. Skip quarantined goals (oscillation ≥ 3).

**Oscillation check:** Before working a failing goal, check if it has oscillated (improved→fail transitions ≥ 3 times in cycle-history.jsonl). If so, quarantine it and try the next failing goal. See `references/oscillation.md`.
```bash
# Count improved→fail transitions for this goal
OSC_COUNT=$(jq -r "select(.target==\"$FAILING\") | .result" .agents/evolve/cycle-history.jsonl \
  | awk 'prev=="improved" && $0=="fail" {count++} {prev=$0} END {print count+0}')
if [ "$OSC_COUNT" -ge 3 ]; then
  QUARANTINED_GOALS[$FAILING]=true
  echo "{\"cycle\":${CYCLE},\"target\":\"${FAILING}\",\"result\":\"quarantined\",\"oscillations\":${OSC_COUNT},\"timestamp\":\"$(date -Iseconds)\"}" >> .agents/evolve/cycle-history.jsonl
fi
```

**Step 3.1: Directive gap** (skip if `--beads-only`):
```bash
# Get directives from GOALS.md
DIRECTIVES=$(ao goals measure --directives 2>/dev/null)
```
If directives exist, assess the top-priority directive (lowest number, non-quarantined):
- Check git log for recent commits addressing it
- If gap detected → generate `suggested_work` description from the directive
- Use this as the work item for Step 4

**Step 3.2: Harvested work** from `.agents/rpi/next-work.jsonl` (unconsumed entries).

**Step 3.3: Open beads**:
```bash
READY_ISSUE=$(bd ready -n 1 2>/dev/null | head -1 | awk '{print $2}')
```
Pick highest-priority unblocked issue. If `bd ready` fails or returns empty, fall through. Do not treat bd failure as idle.

**Step 3.4: Next directive** — assess priority 2, 3, etc. directives for gaps.

**Step 3.5: Lower-weight failing gates** — failing goals with weight < 5.

**Quality mode (`--quality`)** — inverted cascade (findings before directives):

Step 3.0q: Unconsumed high-severity post-mortem findings:
```bash
HIGH=$(jq -r 'select(.consumed==false) | .items[] | select(.severity=="high") | .title' \
  .agents/rpi/next-work.jsonl 2>/dev/null | head -1)
```

Step 3.1q: Unconsumed medium-severity findings.

Step 3.2q: Emergency gates (weight >= 5).

Step 3.3q: Top directive gap.

Step 3.4q: Open beads (`bd ready`).

Step 3.5q: Next directive.

This inverts the standard cascade: findings BEFORE goals and directives.
Rationale: harvested work has 100% first-attempt success rate.

When evolve picks a finding, mark it consumed in next-work.jsonl:
- Set `consumed: true`, `consumed_by: "evolve-quality:cycle-N"`, `consumed_at: "<timestamp>"`
- If the $rpi cycle fails (regression), un-mark the finding (set consumed back to false).

See `references/quality-mode.md` for scoring and full details.

**Nothing found?** HARD GATE — re-derive idle streak from disk:

```bash
# Count trailing idle/unchanged entries in cycle-history.jsonl (portable, no tac)
IDLE_STREAK=$(awk '/"result"\s*:\s*"(idle|unchanged)"/{streak++; next} {streak=0} END{print streak+0}' \
  .agents/evolve/cycle-history.jsonl 2>/dev/null)

if [ "$IDLE_STREAK" -ge 2 ]; then
  # This would be the 3rd consecutive idle cycle — STOP
  echo "Stagnation reached (3 idle cycles). Dormancy is success."
  # go to Teardown — do NOT log another idle entry
fi
```

If IDLE_STREAK < 2: this is idle cycle 1 or 2. Go to Step 6 (idle path).

A cycle is idle if NO work source returned actionable work. A cycle that targeted an oscillating goal and skipped it counts as idle.

If `--dry-run`: report what would be worked on and go to Teardown.

### Step 4: Execute

For a **failing goal**:
```
Invoke $rpi "Improve {goal_id}: {description}" --auto --max-cycles=1
```

For a **beads issue**:
```
Invoke $implement {issue_id}
```
Or for an epic with children: `Invoke $crank {epic_id}`.

For **harvested work**:
```
Invoke $rpi "{item_title}" --auto --max-cycles=1
```
Then mark the item consumed in next-work.jsonl.

### Step 5: Regression Gate

After execution, verify nothing broke:

```bash
# Detect and run project build+test
if [ -f Makefile ]; then make test
elif [ -f package.json ]; then npm test
elif [ -f go.mod ]; then go build ./... && go vet ./... && go test ./... -count=1 -timeout 120s
elif [ -f Cargo.toml ]; then cargo build && cargo test
elif [ -f pyproject.toml ] || [ -f setup.py ]; then python -m pytest
else echo "No recognized build system found"; fi

# Cross-cutting constraint check (catches wiring regressions)
bash scripts/check-wiring-closure.sh
```

If not `--beads-only`, also re-measure to produce a post-cycle snapshot:
```bash
ao goals measure --json --timeout 60 --goal $GOAL_ID > .agents/evolve/fitness-latest-post.json

# Extract goal counts for cycle history entry
PASSING=$(jq '[.goals[] | select(.result=="pass")] | length' .agents/evolve/fitness-latest-post.json 2>/dev/null || echo 0)
TOTAL=$(jq '.goals | length' .agents/evolve/fitness-latest-post.json 2>/dev/null || echo 0)
```

**If regression detected** (previously-passing goal now fails):
```bash
git revert HEAD --no-edit  # single commit
# or for multiple commits:
git revert --no-commit ${CYCLE_START_SHA}..HEAD && git commit -m "revert: evolve cycle ${CYCLE} regression"
```
Set outcome to "regressed".

### Step 6: Log Cycle + Commit

Two paths: productive cycles get committed, idle cycles are local-only.

**PRODUCTIVE cycles** (result is improved, regressed, or harvested):

```bash
# Quality mode: compute quality_score BEFORE writing the JSONL entry
QUALITY_SCORE_FIELD=""
if [ "$QUALITY_MODE" = "true" ]; then
  REMAINING_HIGH=$(jq -r 'select(.consumed==false) | .items[] | select(.severity=="high")' \
    .agents/rpi/next-work.jsonl 2>/dev/null | wc -l | tr -d ' ')
  REMAINING_MEDIUM=$(jq -r 'select(.consumed==false) | .items[] | select(.severity=="medium")' \
    .agents/rpi/next-work.jsonl 2>/dev/null | wc -l | tr -d ' ')
  QUALITY_SCORE=$((100 - (REMAINING_HIGH * 10) - (REMAINING_MEDIUM * 3)))
  [ "$QUALITY_SCORE" -lt 0 ] && QUALITY_SCORE=0
  QUALITY_SCORE_FIELD=",\"quality_score\":${QUALITY_SCORE}"
fi

# Append to cycle history (atomic write)
# Note: flock is Linux-native. On macOS use plain >> append if single-process.
ENTRY="{\"cycle\":${CYCLE},\"target\":\"${TARGET}\",\"result\":\"${OUTCOME}\",\"sha\":\"$(git rev-parse --short HEAD)\",\"timestamp\":\"$(date -Iseconds)\",\"goals_passing\":${PASSING},\"goals_total\":${TOTAL}${QUALITY_SCORE_FIELD}}"
flock .agents/evolve/cycle-history.jsonl -c "echo '${ENTRY}' >> .agents/evolve/cycle-history.jsonl" 2>/dev/null \
  || echo "${ENTRY}" >> .agents/evolve/cycle-history.jsonl

# Verify write
LAST=$(tail -1 .agents/evolve/cycle-history.jsonl | jq -r '.cycle')
[ "$LAST" != "$CYCLE" ] && echo "FATAL: cycle log write failed" && exit 1

# Telemetry
bash scripts/log-telemetry.sh evolve cycle-complete cycle=${CYCLE} goal=${TARGET} outcome=${OUTCOME} 2>/dev/null || true

# Check if this cycle changed real code (not just artifacts)
# Note: Using ${CYCLE_START_SHA} instead of HEAD~1 safely handles sub-skill multi-commit cases
REAL_CHANGES=$(git diff --name-only ${CYCLE_START_SHA}..HEAD -- ':!.agents/*' ':!GOALS.yaml' 2>/dev/null | wc -l | tr -d ' ')

if [ "$REAL_CHANGES" -gt 0 ]; then
  # Full commit: real code was changed
  git add .agents/evolve/cycle-history.jsonl
  git commit -m "evolve: cycle ${CYCLE} -- ${TARGET} ${OUTCOME}"
else
  # Artifact-only cycle: stage JSONL but don't create a standalone commit
  # The $rpi or $implement sub-skill already committed its own artifact changes
  git add .agents/evolve/cycle-history.jsonl
  # Do NOT create a standalone commit for artifact-only work
fi

PRODUCTIVE_THIS_SESSION=$((PRODUCTIVE_THIS_SESSION + 1))
```

**IDLE cycles** (nothing found):

```bash
# Append locally — NOT committed (disposable if compaction occurs)
flock .agents/evolve/cycle-history.jsonl -c \
  "echo '{\"cycle\":${CYCLE},\"target\":\"idle\",\"result\":\"unchanged\",\"timestamp\":\"$(date -Iseconds)\"}' >> .agents/evolve/cycle-history.jsonl" \
  2>/dev/null \
  || echo "{\"cycle\":${CYCLE},\"target\":\"idle\",\"result\":\"unchanged\",\"timestamp\":\"$(date -Iseconds)\"}" >> .agents/evolve/cycle-history.jsonl
# No git add, no git commit, no fitness snapshot write
```

### Step 7: Loop or Stop

```bash
CYCLE=$((CYCLE + 1))
# Stop if max-cycles reached
# Otherwise: go to Step 1
```

Push only when productive work has accumulated:
```bash
if [ $((PRODUCTIVE_THIS_SESSION % 5)) -eq 0 ] && [ "$PRODUCTIVE_THIS_SESSION" -gt 0 ]; then
  git push
fi
```

### Teardown

1. Commit any staged but uncommitted cycle-history.jsonl (from artifact-only cycles):
```bash
if git diff --cached --name-only | grep -q cycle-history.jsonl; then
  git commit -m "evolve: session teardown -- artifact-only cycles logged"
fi
```
2. Run `$post-mortem "evolve session: ${CYCLE} cycles"` to harvest learnings.
3. Push only if unpushed commits exist:
```bash
UNPUSHED=$(git log origin/main..HEAD --oneline 2>/dev/null | wc -l)
[ "$UNPUSHED" -gt 0 ] && git push
```
4. Report summary:

```
## $evolve Complete
Cycles: N | Productive: X | Regressed: Y (reverted) | Idle: Z
Stop reason: stagnation | circuit-breaker | max-cycles | kill-switch
```

In quality mode, the report includes additional fields:
```
## $evolve Complete (quality mode)
Cycles: N | Findings resolved: X | Goals fixed: Y | Idle: Z
Quality score: start → end (delta)
Remaining unconsumed: H high, M medium
Stop reason: stagnation | circuit-breaker | max-cycles | kill-switch
```

## Examples

**Basic:** `$evolve --max-cycles=5` — measures goals, fixes highest-weight failure, gates, repeats for 5 cycles.

**Beads only:** `$evolve --beads-only` — skips goals measurement, works through `bd ready` backlog.

**Dry run:** `$evolve --dry-run` — shows what would be worked on without executing.

See `references/examples.md` for detailed walkthroughs.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Loop exits immediately | Remove `~/.config/evolve/KILL` or `.agents/evolve/STOP` |
| Stagnation after 3 idle cycles | All work sources empty — this is success |
| `ao goals measure` hangs | Use `--timeout 30` flag or `--beads-only` to skip |
| Regression gate reverts | Review reverted changes, narrow scope, re-run |

See `references/cycle-history.md` for advanced troubleshooting.

## References

- `references/cycle-history.md` — JSONL format, recovery protocol, kill switch
- `references/compounding.md` — Knowledge flywheel and work harvesting
- `references/goals-schema.md` — GOALS.yaml format and continuous metrics
- `references/parallel-execution.md` — Parallel $swarm architecture
- `references/teardown.md` — Trajectory computation and session summary
- `references/examples.md` — Detailed usage examples
- `references/artifacts.md` — Generated files registry
- `references/oscillation.md` — Oscillation detection and quarantine
- `references/quality-mode.md` — Quality-first mode: scoring, priority cascade, artifacts

## See Also

- `skills/rpi/SKILL.md` — Full lifecycle orchestrator (called per cycle)
- `skills/crank/SKILL.md` — Epic execution (called for beads epics)
- `GOALS.yaml` — Fitness goals for this repo

## Reference Documents

- [references/artifacts.md](references/artifacts.md)
- [references/compounding.md](references/compounding.md)
- [references/cycle-history.md](references/cycle-history.md)
- [references/examples.md](references/examples.md)
- [references/goals-schema.md](references/goals-schema.md)
- [references/oscillation.md](references/oscillation.md)
- [references/parallel-execution.md](references/parallel-execution.md)
- [references/quality-mode.md](references/quality-mode.md)
- [references/teardown.md](references/teardown.md)

---

## References

### artifacts.md

# $evolve Artifacts

## Committed to Git

| File | Purpose |
|------|---------|
| `GOALS.yaml` | Fitness goals (repo root) |
| `.agents/evolve/fitness-0-baseline.json` | Cycle-0 baseline snapshot (comparison anchor) |
| `.agents/evolve/cycle-history.jsonl` | Cycle outcomes log (includes commit SHAs) |

## Local Only (gitignored)

| File | Purpose |
|------|---------|
| `.agents/evolve/fitness-latest.json` | Pre-cycle fitness snapshot (rolling, overwritten each cycle) |
| `.agents/evolve/fitness-latest-post.json` | Post-cycle fitness snapshot (for regression comparison) |
| `.agents/evolve/session-summary.md` | Session wrap-up |
| `.agents/evolve/session-fitness-delta.md` | Session fitness trajectory (baseline to final delta) |
| `.agents/evolve/STOP` | Local kill switch |
| `~/.config/evolve/KILL` | External kill switch |

## Removed (legacy)

These files are no longer generated. Older repos may have them in git history:

| File | Replacement |
|------|-------------|
| `.agents/evolve/fitness-{N}-pre.json` | `fitness-latest.json` (rolling) |
| `.agents/evolve/fitness-{N}-post.json` | `fitness-latest-post.json` (rolling) |
| `.agents/evolve/cycle-0-report.md` | Inlined into session-summary.md |
| `.agents/evolve/last-sweep-date` | No longer needed (baseline gate uses fitness-0-baseline.json) |
| `.agents/evolve/KILLED.json` | Kill switch acknowledgment removed (STOP file is sufficient) |

### compounding.md

# How Compounding Works

Two mechanisms feed the loop:

**1. Knowledge flywheel (each cycle is smarter):**
```
Session 1:
  ao inject (nothing yet)         → cycle runs blind
  $rpi fixes test-pass-rate       → post-mortem runs ao forge
  Learnings extracted: "tests/skills/run-all.sh validates frontmatter"

Session 2:
  ao inject (loads Session 1 learnings)  → cycle knows about frontmatter validation
  $rpi fixes doc-coverage                → approach informed by prior learning
  Learnings extracted: "references/ dirs need at least one .md file"
```

**2. Work harvesting (each cycle discovers the next):**
```
Cycle 1: $rpi fixes test-pass-rate
  → post-mortem harvests: "add missing smoke test for $evolve" → next-work.jsonl

Cycle 2: all GOALS.yaml goals pass
  → $evolve reads next-work.jsonl (filtered to current repo + cross-repo '*')
  → picks "add missing smoke test"
  → $rpi fixes it → post-mortem harvests: "update SKILL-TIERS count"

Cycle 3: reads next-work.jsonl → picks "update SKILL-TIERS count" → ...
```

The loop keeps running as long as post-mortem keeps finding follow-up work. Each $rpi cycle generates next-work items from its own post-mortem. The system feeds itself.

**Priority cascade:**
```
GOALS.yaml goals (explicit, human-authored)  → fix these first
Open beads (bd ready)                        → work when goals pass
next-work.jsonl (harvested from post-mortem) → work on these when beads empty
nothing left                                 → re-measure (external changes may create new work)
3 consecutive idle cycles                    → stagnation stop (nothing left to improve)
60-minute circuit breaker                    → stop if no productive cycle in 60 min
kill switch                                  → immediate stop
```

The loop does NOT stop just because goals are met. It re-measures, checks for harvested work, and only stops after 3 consecutive cycles with truly nothing to do. Idle cycles are NOT committed to git — only appended locally. The idle streak is re-derived from disk at each session start, so compaction cannot corrupt it. Use the kill switch for intentional stops.

### cycle-history.md

# Cycle History Format and Recovery Protocol

## Compaction Resilience

The evolve loop MUST survive context compaction. Every cycle commits its
artifacts to git before proceeding. The `cycle-history.jsonl` file is the
recovery point -- on session restart, read it to determine cycle number
and resume from Step 1.

## Cycle History JSONL Format

Append one line per cycle to `.agents/evolve/cycle-history.jsonl`.

### Canonical Schema

All new entries MUST use this schema:

```json
{
  "cycle": 123,
  "target": "goal-id-or-idle",
  "result": "improved|regressed|unchanged|harvested|quarantined",
  "sha": "abc1234",
  "timestamp": "2026-02-23T12:00:00-05:00",
  "goals_passing": 59,
  "goals_total": 59
}
```

**Field standardization:**
- Use `target` (not `goal_id`) — this is what recent cycles already use
- Use `sha` (not `commit_sha`) — shorter, matches recent convention
- Always include `goals_passing` and `goals_total` — enables trajectory plotting
- Optional fields: `quality_score` (quality mode), `idle_streak` (idle cycles), `parallel` + `goal_ids` (parallel mode)

**Legacy field names:** Older entries may use `goal_id` instead of `target` and `commit_sha` instead of `sha`. Tools reading cycle-history.jsonl should handle both conventions.

**Sequential cycle entry:**
```jsonl
{"cycle": 1, "target": "test-pass-rate", "result": "improved", "sha": "abc1234", "goals_passing": 18, "goals_total": 23, "timestamp": "2026-02-11T21:00:00Z"}
{"cycle": 2, "target": "doc-coverage", "result": "regressed", "sha": "def5678", "goals_passing": 17, "goals_total": 23, "timestamp": "2026-02-11T21:30:00Z"}
```

**Idle cycle entry** (not committed to git):
```jsonl
{"cycle": 3, "target": "idle", "result": "unchanged", "timestamp": "2026-02-11T22:00:00Z"}
```

**Parallel cycle entry** (use `goal_ids` array and `parallel: true`):
```jsonl
{"cycle": 4, "goal_ids": ["test-pass-rate", "doc-coverage", "lint-clean"], "result": "improved", "sha": "ghi9012", "goals_passing": 22, "goals_total": 23, "parallel": true, "timestamp": "2026-02-11T22:30:00Z"}
```

### Mandatory Fields

Every productive cycle log entry MUST include:

| Field | Description |
|-------|-------------|
| `cycle` | Cycle number (1-indexed) |
| `target` | Target goal ID, or `"idle"` for idle cycles |
| `result` | One of: `improved`, `regressed`, `unchanged`, `harvested`, `quarantined` |
| `sha` | Git SHA after cycle commit (omitted for idle cycles) |
| `goals_passing` | Count of goals with result "pass" (omitted for idle cycles) |
| `goals_total` | Total goals measured (omitted for idle cycles) |
| `timestamp` | ISO 8601 timestamp |

These enable fitness trajectory plotting across cycles.

### Telemetry

Log telemetry at the end of each cycle:
```bash
bash scripts/log-telemetry.sh evolve cycle-complete cycle=${CYCLE} score=${SCORE} goals_passing=${PASSING} goals_total=${TOTAL}
```

### Compaction-Proofing: Commit After Productive Cycles

Only **productive cycles** (improved, regressed, harvested) are committed. Idle
cycles are appended to cycle-history.jsonl locally but NOT committed — they are
disposable if compaction occurs, and the idle streak is re-derived from disk at
session start.

```bash
# Productive cycle: commit cycle-history.jsonl only
git add .agents/evolve/cycle-history.jsonl
git commit -m "evolve: cycle ${CYCLE} -- ${TARGET} ${OUTCOME}"

# Parallel productive cycle:
git add .agents/evolve/cycle-history.jsonl
git commit -m "evolve: cycle ${CYCLE} -- parallel wave [${goal_ids}] ${outcome}"

# Idle cycle: append locally, do NOT commit
echo '{"cycle":N,"target":"idle","result":"unchanged",...}' >> .agents/evolve/cycle-history.jsonl
# No git add, no git commit
```

### 60-Minute Circuit Breaker

At session start (Step 0), after recovering the idle streak, check the timestamp
of the last productive cycle. If it was more than 60 minutes ago, go directly to
Teardown. This prevents runaway sessions that accumulate idle cycles without
producing value.

```bash
LAST_PRODUCTIVE_TS=$(grep -v '"idle"\|"unchanged"' .agents/evolve/cycle-history.jsonl 2>/dev/null \
  | tail -1 | jq -r '.timestamp // empty')
# If >3600s since last productive cycle AND timestamp parsed correctly: CIRCUIT BREAKER → Teardown
# Guard: LAST_EPOCH > 1e9 prevents false trigger on date parse failure
```

## Recovery Protocol

On session restart or after compaction:

1. Read `.agents/evolve/cycle-history.jsonl` to find last completed cycle number
2. Set `evolve_state.cycle` to last cycle + 1
3. Resume from Step 1 (kill switch check)
4. The baseline snapshot (`fitness-0-baseline.json`) is preserved -- do not regenerate

## Kill Switch

Two paths, checked at every cycle boundary:

| File | Purpose | Who Creates It |
|------|---------|---------------|
| `~/.config/evolve/KILL` | Permanent stop (outside repo) | Human |
| `.agents/evolve/STOP` | One-time local stop | Human or automation |

To stop $evolve:
```bash
echo "Taking a break" > ~/.config/evolve/KILL    # Permanent
echo "done for today" > .agents/evolve/STOP       # Local, one-time
```

To re-enable:
```bash
rm ~/.config/evolve/KILL
rm .agents/evolve/STOP
```

## Flags Reference

| Flag | Default | Description |
|------|---------|-------------|
| `--max-cycles=N` | unlimited | Optional hard cap. Without this, loop runs forever. |
| `--test-first` | off | Pass `--test-first` through to `$rpi` -> `$crank` |
| `--dry-run` | off | Measure fitness and show plan, don't execute |
| `--skip-baseline` | off | Skip cycle-0 baseline sweep |
| `--parallel` | off | Enable parallel goal execution via $swarm per cycle |
| `--max-parallel=N` | 3 | Max goals to fix in parallel (cap: 5). Only with `--parallel`. |

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `$evolve` exits immediately with "KILL SWITCH ACTIVE" | Kill switch file exists | Remove `~/.config/evolve/KILL` or `.agents/evolve/STOP` to re-enable |
| "No goals to measure" error | GOALS.yaml missing or empty | Create GOALS.yaml in repo root with fitness goals (see goals-schema.md) |
| Cycle completes but fitness unchanged | Goal check command is always passing or always failing | Verify check command logic in GOALS.yaml produces exit code 0 (pass) or non-zero (fail) |
| Regression revert fails | Multiple commits in cycle or uncommitted changes | Check cycle-start SHA in fitness snapshot, commit or stash changes before retrying |
| Harvested work never consumed | All goals passing but `next-work.jsonl` not read | Check file exists and has `consumed: false` entries. Agent picks harvested work after goals met. |
| Loop stops after N cycles | `--max-cycles` was set (or old default of 10) | Omit `--max-cycles` flag -- default is now unlimited. Loop runs until kill switch or 3 idle cycles. |

### examples.md

# $evolve Examples

## Infinite Autonomous Improvement

**User says:** `$evolve`

**What happens:**
1. Agent checks kill switch files (none found, continues)
2. Agent measures fitness against GOALS.yaml (3 of 5 goals passing)
3. Agent selects worst-failing goal by weight (test-pass-rate)
4. Agent invokes `$rpi "Improve test-pass-rate"` with full lifecycle
5. Agent re-measures fitness post-cycle (test-pass-rate now passing, all others unchanged)
6. Agent logs cycle to history, increments cycle counter
7. Agent loops to next cycle, selects next failing goal
8. After all goals pass, agent checks harvested work from post-mortem — finds 3 items
9. Agent works through harvested items, each generating more via post-mortem
10. After 3 consecutive idle cycles (no failing goals, no harvested work), agent runs `$post-mortem` and writes session summary
11. To stop earlier: create `~/.config/evolve/KILL` or `.agents/evolve/STOP`

**Result:** Runs forever — fixing goals, consuming harvested work, re-measuring. Only stops on kill switch or stagnation (3 idle cycles).

## Dry-Run Mode

**User says:** `$evolve --dry-run`

**What happens:**
1. Agent measures fitness (3 of 5 goals passing)
2. Agent identifies worst-failing goal (doc-coverage, weight 5)
3. Agent reports what would be worked on: "Dry run: would work on 'doc-coverage' (weight: 5)"
4. Agent shows harvested work queue (2 items from prior RPI cycles)
5. Agent stops without executing

**Result:** Fitness report and next-action preview without code changes.

## Regression with Revert

**User says:** `$evolve --max-cycles=3`

**What happens:**
1. Agent improves goal A in cycle 1 (commit abc123)
2. Agent measures fitness post-cycle: goal A passes, but goal B now fails (regression)
3. Agent reverts commit abc123 with annotated message
4. Agent logs regression to history, moves to next goal
5. Agent completes 3 cycles (cap reached), runs post-mortem

**Result:** Fitness regressions are auto-reverted, preventing compounding failures.

## Parallel Goal Improvement

**User says:** `$evolve --parallel`

**What happens:**
1. Agent checks kill switch (none found)
2. Agent measures fitness against GOALS.yaml (4 of 7 goals failing)
3. Agent selects top 3 independent failing goals by weight, filtered for independence via `select_parallel_goals`
4. Agent creates TaskList tasks for each goal, sets up artifact isolation
5. Agent invokes `$swarm --worktrees` — spawns 3 fresh-context workers in isolated worktrees
6. Each worker runs a full `$rpi` cycle independently (research → plan → crank → vibe → post-mortem)
7. `$swarm` completes — all 3 workers done, lead merges worktrees
8. Agent re-measures ALL goals (single regression gate for entire wave)
9. No regression detected — logs cycle with `goal_ids` array and `parallel: true`
10. Next cycle: 1 remaining failing goal → sequential (only 1 goal, no parallelism needed)
11. After all goals pass, checks harvested work, eventually stagnation → teardown

**Result:** 3 goals fixed in ~1 cycle instead of 3 sequential cycles. ~3x speedup for independent goals. Each worker's $post-mortem feeds the knowledge flywheel independently.

## Parallel with Regression Revert

**User says:** `$evolve --parallel --max-cycles=2`

**What happens:**
1. Cycle 1: 3 parallel goals attempted via `$swarm --worktrees`
2. Post-wave regression gate detects goal C started failing after goals A and B were improved
3. Agent reverts entire parallel wave (all merged worktree commits) using `cycle_start_sha`
4. Logs cycle with `result: "regressed"` and all 3 `goal_ids`
5. Cycle 2: Agent retries — `select_parallel_goals` still selects same 3 (different check scripts)
6. This time no regression — all 3 improvements are clean
7. Max cycles reached (2), runs teardown with `$post-mortem`

**Result:** Parallel regression detected and reverted cleanly. Entire wave rolled back as a unit. Retry in next cycle succeeds.

### goals-schema.md

# GOALS.yaml Schema

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

## Fitness Snapshot Format

Each cycle writes a fitness snapshot with **continuous values** (not just pass/fail):

```json
{
  "cycle": 1,
  "timestamp": "2026-02-12T15:45:00-05:00",
  "cycle_start_sha": "abc1234",
  "goals": [
    {
      "id": "go-coverage-floor",
      "result": "pass",
      "weight": 2,
      "value": 86.1,
      "threshold": 80
    },
    {
      "id": "doc-coverage",
      "result": "pass",
      "weight": 2,
      "value": 20,
      "threshold": 16
    },
    {
      "id": "go-cli-builds",
      "result": "pass",
      "weight": 5,
      "value": null,
      "threshold": null
    }
  ]
}
```

- **value**: The continuous metric extracted from the check command (null for binary-only goals)
- **threshold**: The pass/fail threshold (null for binary-only goals)
- **cycle_start_sha**: Git SHA at cycle start, used for multi-commit revert on regression

Pre-cycle snapshot: `fitness-latest.json` (rolling, overwritten each cycle)
Post-cycle snapshot: `fitness-latest-post.json` (rolling, for regression comparison)

## Cycle-0 Baseline

Before the first improvement cycle runs, the system captures a baseline fitness snapshot (`fitness-0-baseline.json`). This serves as the comparison anchor for measuring session-wide progress.

The baseline includes:
- **All goals** from GOALS.yaml, measured in their initial state
- **Cycle-0 report** (`cycle-0-report.md`) — summary of which goals are failing and their weights
- **No regression comparisons** — this is the starting point

When the session ends (at Teardown), the system computes the **session fitness trajectory** by comparing the baseline against the final cycle snapshot. This produces `session-fitness-delta.md`, which shows which goals improved, regressed, or stayed unchanged over the entire $evolve session.

## Meta-Goals

Meta-goals validate the validation system itself. Use them to prevent exception lists (allowlists, skip lists) from accumulating stale entries unnoticed.

```yaml
# Meta-goals validate the validation system itself
goals:
  - id: allowlist-hygiene
    description: "Every dead-code allowlist entry should have 0 non-test callers"
    check: "bash scripts/check-allowlist-hygiene.sh"
    weight: 7

  - id: skip-list-hygiene
    description: "Every skip-list entry should still reference an existing test"
    check: "bash scripts/check-skip-list-hygiene.sh"
    weight: 5
```

**When to add a meta-goal:** After pruning any allowlist or exception list, always add a corresponding meta-goal that fails if entries have callers/references. Allowlists without meta-goals are technical debt magnets — they grow silently across epics.

## Maintaining GOALS.yaml

Use `$goals` to maintain the fitness specification:
- `$goals` — run all checks, report pass/fail by pillar
- `$goals generate` — scan repo for uncovered areas, propose new goals
- `$goals prune` — find stale/broken goals, propose removals or updates

## GOALS.md Format (Version 4)

GOALS.md extends the YAML format with strategic intent:

```markdown
# Goals

<Mission statement>

## North Stars
- <Aspiration>

## Anti Stars
- <What to avoid>

## Directives

### 1. <Title>
<Description>
**Steer:** increase | decrease | hold | explore

## Gates
| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| id | `command` | N | Description |
```

### Evolve Integration

When GOALS.md is detected, evolve uses the directive-based cascade (Step 3.1):
1. `ao goals measure --directives` returns the directive list as JSON
2. Top-priority directive (lowest number) is assessed for gaps
3. If gap found → generates work item from directive description + steer
4. Directive becomes the work source for the cycle

When `--beads-only` is passed, directive assessment is skipped entirely.

### Format Detection

`ao goals measure` auto-detects format. When both GOALS.yaml and GOALS.md exist, GOALS.md takes precedence.

### oscillation.md

# Oscillation Detection

## What Is Oscillation?

A goal **oscillates** when it alternates between passing and failing across
evolve cycles. Typically this means the fix for the goal causes a side-effect
that breaks something else, which gets reverted, which re-exposes the original
failure — creating a loop.

Example from cycle-history.jsonl:
```
cycle 5:  goal-X → improved
cycle 8:  goal-X → fail
cycle 12: goal-X → improved
cycle 15: goal-X → fail
```

## Detection

Count **improved→fail transitions** for the same `target` in
`.agents/evolve/cycle-history.jsonl`:

```bash
# Count oscillations for a given goal
jq -r "select(.target==\"$GOAL_ID\") | .result" .agents/evolve/cycle-history.jsonl \
  | awk 'prev=="improved" && $0=="fail" {count++} {prev=$0} END {print count+0}'
```

## Threshold

**3 oscillations** (improved→fail transitions) within a single session
triggers quarantine. The goal is skipped in Step 3 selection.

## Effect

- Quarantined goals are skipped during work selection (Step 3)
- Skipping a quarantined goal counts as idle (no actionable work found)
- The quarantine is session-scoped — a new session resets the count
- Quarantine events are logged in cycle-history.jsonl with `"result": "quarantined"`

## Recovery

1. Human identifies the root cause (usually a conflict between two goals)
2. Fix the underlying issue manually
3. Start a new evolve session (quarantine resets)
4. Or: remove the quarantine by deleting the goal from the skip list

## Why Not Just Increase the Skip Threshold?

The 3-consecutive-regression skip in Step 3 only catches monotonic failure.
Oscillation is worse — it burns cycles alternating between "fixed" and "broken"
without the stagnation detector ever triggering (because the goal intermittently
passes). The oscillation detector catches this pattern explicitly.

### parallel-execution.md

# Parallel Goal Execution

## Architecture

When `--parallel` is enabled, `$evolve` uses `$swarm` to execute multiple independent
goal improvements concurrently instead of fixing one goal per cycle.

```
$evolve --parallel (Fitness Loop)
  │
  ├─ Step 2: Measure ALL goals
  │
  ├─ Step 3: Select top N independent failing goals (max_parallel, default 3)
  │  └─ select_parallel_goals: heuristic independence via check-script overlap
  │
  ├─ Step 4: Parallel execution via $swarm
  │  ├─ TaskCreate for each selected goal
  │  ├─ Artifact isolation: .agents/evolve/parallel-rpi/{goal.id}/
  │  ├─ Git isolation: $swarm --worktrees (each worker in /tmp/evolve-{goal.id})
  │  └─ $swarm spawns N fresh-context workers, each runs full $rpi cycle:
  │     └─ research → plan → pre-mortem → crank → vibe → post-mortem
  │
  ├─ Step 5: Single regression gate (re-measure ALL goals after wave)
  │  ├─ If ANY goal regressed → revert ENTIRE parallel wave
  │  └─ If clean → log cycle with goal_ids array
  │
  └─ Step 6-7: Log, loop (same as sequential)
```

## The Fractal Pattern

Swarm is the universal coordination primitive at every level:

```
LEVEL 0: $evolve --parallel
  └─ $swarm (parallel goal improvements)     ← NEW: swarm at evolve level
     └─ LEVEL 1: $rpi (per-goal lifecycle)
        └─ research → plan → crank → vibe → post-mortem
           └─ LEVEL 2: $crank (epic execution)
              └─ $swarm (parallel issue implementation)  ← existing: swarm at crank level
                 └─ LEVEL 3: workers (atomic tasks)
```

Each level creates fresh context for the next (Ralph Wiggum pattern).
The pattern is always: **one leader + N fresh-context workers + validation + cleanup**.

## Goal Independence Detection

`select_parallel_goals` uses a heuristic check:

1. Start with highest-weight failing goal
2. For each remaining eligible goal (weight-sorted):
   - Compare check commands for shared scripts/paths
   - If independent: add to selection (up to max_parallel)
   - If overlapping: skip (handled in next cycle)

**This is a heuristic, not a guarantee.** Goals don't declare which files their
improvements will modify — only which scripts verify them. Two goals with different
check scripts may still modify overlapping files.

**The regression gate (Step 5) is the real safety net.** If parallel goals conflict,
the regression check detects it and reverts the entire wave. This makes false
negatives in independence detection safe (they just cost one wasted cycle).

## Artifact Isolation

Each parallel $rpi worker needs isolated artifact directories to prevent collision:

| Directory | Purpose | Isolation |
|-----------|---------|-----------|
| `.agents/evolve/parallel-rpi/{goal.id}/` | $rpi phase summaries, next-work | Per-goal subdirectory |
| `.agents/evolve/parallel-results/{goal.id}.md` | Worker result summary | Per-goal file |
| `/tmp/evolve-{goal.id}` | Git worktree | Per-goal worktree via $swarm |

Without isolation, N concurrent $rpi cycles would collide on `.agents/rpi/`
(phase summaries, next-work.jsonl) and git index locks.

## Git Isolation

Parallel workers MUST use worktree isolation (via `$swarm --worktrees`):

- Each worker operates in `/tmp/evolve-{goal.id}` worktree
- No git lock conflicts (each worktree has its own index)
- Lead merges worktrees after all complete, before regression gate
- On regression: revert all merged commits using `cycle_start_sha`

## Regression Handling

**Sequential mode:** Revert commits from one goal's $rpi cycle.

**Parallel mode:** Revert ALL commits from the entire parallel wave.
The `cycle_start_sha` (captured before the wave) anchors the revert point.
All N goal improvements are rolled back together — even goals that individually
succeeded. This is by design: if goals interfere, we can't know which one
caused the regression without testing each in isolation.

## Cycle History Schema

Sequential cycles use `target` (string). Parallel cycles use `goal_ids` (array) with `parallel: true`:

```jsonl
{"cycle": 1, "target": "test-pass-rate", "result": "improved", "sha": "abc1234", ...}
{"cycle": 2, "goal_ids": ["doc-coverage", "lint-clean"], "result": "improved", "sha": "def5678", "parallel": true, ...}
```

Legacy entries may use `goal_id` instead of `target` and `commit_sha` instead of `sha`. Tools should handle both.

## Compounding

Each parallel $rpi worker runs its own $post-mortem, which feeds the knowledge
flywheel independently. Learnings from all N parallel cycles compound into the
flywheel, feeding the next $evolve cycle.

## When to Use

| Scenario | Mode |
|----------|------|
| 1-2 failing goals | Sequential (default) — parallelism overhead not worth it |
| 3+ independent failing goals | `--parallel` — significant speedup |
| Goals with overlapping files | Sequential — parallel would cause conflicts |
| First run on new repo | Sequential — learn the codebase before parallelizing |

## Constraints

- Max 5 parallel goals per wave (`--max-parallel` cap)
- Default 3 parallel goals (balance between speedup and resource usage)
- Each $rpi worker needs a full context window — budget accordingly
- Worktree isolation required (no shared-worktree parallel $rpi)

### quality-mode.md

# Quality Mode

## When to Use

Use `--quality` when:
- Post-mortem findings are accumulating faster than they're consumed
- All GOALS pass but `next-work.jsonl` has unconsumed high-severity items
- You want to resolve context-hot findings from a just-completed epic
- Running immediately after `$post-mortem` to action its findings

Do NOT use `--quality` when:
- GOALS have critical failures (build broken, tests failing)
- No `next-work.jsonl` exists or is empty
- You want standard fitness-driven improvement

## Quality Score

Simple severity-weighted score:

```
score = 100 - (high_count * 10) - (medium_count * 3)
```

Where counts are unconsumed findings remaining in next-work.jsonl.

| Score | Meaning |
|-------|---------|
| 90-100 | Excellent — few or no findings remaining |
| 70-89 | Good — medium-severity items remain |
| 50-69 | Attention needed — high-severity items remain |
| <50 | Quality debt — many high-severity findings |

## Priority Cascade (Quality Mode)

1. High-severity unconsumed findings → $rpi
2. Medium-severity unconsumed findings → $rpi
3. Failing GOALS.yaml goals → $rpi (standard behavior)
4. Open beads → $implement
5. Nothing → stagnation (3 idle cycles)

## Marking Findings Consumed

When evolve picks a finding from next-work.jsonl, mark it consumed:
- Set `consumed: true`
- Set `consumed_by: "evolve-quality:cycle-N"`
- Set `consumed_at: "<timestamp>"`

If the $rpi cycle fails (regression), un-mark the finding (set consumed back to false).

## Artifacts

| File | Purpose |
|------|---------|
| `cycle-history.jsonl` | Same as standard mode + `quality_score` field |
| `fitness-latest.json` | Same as standard mode (goals measurement) |
| `quality-trajectory.md` | Quality score over time (written at teardown) |

## Interaction with Standard Mode

Quality mode and standard mode share:
- The same cycle-history.jsonl
- The same fitness measurement (goals are still checked)
- The same stagnation detection (3 idle cycles)
- The same circuit breaker (60 minutes)

They differ in work selection priority: quality mode picks findings first, standard picks goals first.

### teardown.md

# Teardown Procedure

**Auto-run $post-mortem on the full evolution session:**

```
$post-mortem "evolve session: $CYCLE cycles, goals improved: X, harvested: Y"
```

This captures learnings from the ENTIRE evolution run (all cycles, all $rpi invocations) in one council review. The post-mortem harvests follow-up items into `next-work.jsonl`, feeding the next `$evolve` session.

**Compute session fitness trajectory:**

```bash
# Check if both baseline and final snapshot exist
if [ -f .agents/evolve/fitness-0-baseline.json ] && [ -f .agents/evolve/fitness-latest.json ]; then
  baseline = load(".agents/evolve/fitness-0-baseline.json")
  final = load(".agents/evolve/fitness-latest.json")

  # Compute delta — goals that flipped between baseline and final
  improved_count = 0
  regressed_count = 0
  unchanged_count = 0
  delta_rows = []

  for final_goal in final.goals:
    baseline_goal = baseline.goals.find(g => g.id == final_goal.id)
    baseline_result = baseline_goal ? baseline_goal.result : "unknown"
    final_result = final_goal.result

    if baseline_result == "fail" and final_result == "pass":
      delta = "improved"
      improved_count += 1
    elif baseline_result == "pass" and final_result == "fail":
      delta = "regressed"
      regressed_count += 1
    else:
      delta = "unchanged"
      unchanged_count += 1

    delta_rows.append({goal_id: final_goal.id, baseline_result, final_result, delta})

  # Write session-fitness-delta.md with trajectory table
  cat > .agents/evolve/session-fitness-delta.md << EOF
  # Session Fitness Trajectory

  | goal_id | baseline_result | final_result | delta |
  |---------|-----------------|--------------|-------|
  $(for row in delta_rows: "| ${row.goal_id} | ${row.baseline_result} | ${row.final_result} | ${row.delta} |")

  **Summary:** ${improved_count} improved, ${regressed_count} regressed, ${unchanged_count} unchanged
  EOF

  # Include delta summary in user-facing teardown report
  log "Fitness trajectory: ${improved_count} improved, ${regressed_count} regressed, ${unchanged_count} unchanged"
fi
```

**Then write session summary:**

```bash
cat > .agents/evolve/session-summary.md << EOF
# $evolve Session Summary

**Date:** $(date -Iseconds)
**Cycles:** $CYCLE of $MAX_CYCLES
**Goals measured:** $(wc -l < GOALS.yaml goals)

## Cycle History
$(cat .agents/evolve/cycle-history.jsonl)

## Final Fitness
$(cat .agents/evolve/fitness-latest.json)

## Post-Mortem
<path to post-mortem report from above>

## Next Steps
- Run \`$evolve\` again to continue improving
- Run \`$evolve --dry-run\` to check current fitness without executing
- Create \`~/.config/evolve/KILL\` to prevent future runs
- Create \`.agents/evolve/STOP\` for a one-time local stop
EOF
```

Report to user:
```
## $evolve Complete

Cycles: N of M
Goals improved: X
Goals regressed: Y (reverted)
Goals unchanged: Z
Post-mortem: <verdict> (see <report-path>)

Run `$evolve` again to continue improving.
```


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
check "SKILL.md has name: evolve" "grep -q '^name: evolve' '$SKILL_DIR/SKILL.md'"
check "references/ directory exists" "[ -d '$SKILL_DIR/references' ]"
check "references/ has at least 1 file" "[ \$(ls '$SKILL_DIR/references/' | wc -l) -ge 1 ]"
check "SKILL.md mentions kill switch" "grep -qi 'kill switch' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions fitness" "grep -qi 'fitness' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions GOALS.yaml" "grep -q 'GOALS.yaml' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions cycle" "grep -qi 'cycle' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions $rpi" "grep -q '$rpi' '$SKILL_DIR/SKILL.md'"
# Behavioral contracts from retro learnings (2026-02-12)
check "SKILL.md has KILL file path" "grep -q 'KILL' '$SKILL_DIR/SKILL.md'"
check "SKILL.md documents regression detection" "grep -qi 'regression' '$SKILL_DIR/SKILL.md'"
check "SKILL.md documents snapshot enforcement" "grep -qi 'snapshot' '$SKILL_DIR/SKILL.md'"
check "SKILL.md documents session_start_sha" "grep -qi 'session.start.sha\|cycle_start_sha' '$SKILL_DIR/SKILL.md'"
check "SKILL.md documents continuous values" "grep -qi 'continuous\|value.*threshold' '$SKILL_DIR/SKILL.md'"
check "SKILL.md documents full regression gate" "grep -qi 'full.*regression\|all goals' '$SKILL_DIR/SKILL.md'"
check "SKILL.md documents post-cycle snapshot" "grep -q 'fitness-.*-post' '$SKILL_DIR/SKILL.md'"
check "SKILL.md documents oscillation detection" "grep -qi 'oscillat' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


