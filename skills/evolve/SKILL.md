---
name: evolve
description: Goal-driven fitness-scored improvement loop. Measures goals, picks worst gap, runs /rpi, compounds via knowledge flywheel. Also pulls from open beads when goals all pass.
skill_api_version: 1
user-invocable: true
context:
  window: fork
  intent:
    mode: task
  sections:
    exclude: [HISTORY]
  intel_scope: full
metadata:
  tier: execution
  dependencies:
    - rpi         # required - executes each improvement cycle
    - post-mortem # required - auto-runs at teardown to harvest learnings
    - athena      # optional - knowledge warmup when --athena is passed
  triggers:
    - evolve
    - improve everything
    - autonomous improvement
    - run until done
---

# /evolve — Goal-Driven Compounding Loop

> Measure what's wrong. Fix the worst thing. Measure again. Compound.

Always-on autonomous loop over `/rpi`. Work selection order:
1. **Harvested `.agents/rpi/next-work.jsonl` work** (freshest concrete follow-up)
2. **Open ready beads work** (`bd ready`)
3. **Failing goals and directive gaps** (`ao goals measure`)
4. **Testing improvements** (missing/thin coverage, missing regression tests)
5. **Validation tightening and bug-hunt passes** (gates, audits, bug sweeps)
6. **Complexity / TODO / FIXME / drift / dead code / stale docs / stale research mining**
7. **Concrete feature suggestions** derived from repo purpose when no sharper work exists

**Dormancy is last resort.** Empty current queues mean "run the generator layers", not "stop". Only go dormant after the queue layers and generator layers come up empty across multiple consecutive passes.

```bash
/evolve                      # Run until kill switch, max-cycles, or real dormancy
/evolve --max-cycles=5       # Cap at 5 cycles
/evolve --dry-run            # Show what would be worked on, don't execute
/evolve --beads-only         # Skip goals measurement, work beads backlog only
/evolve --quality            # Quality-first mode: prioritize post-mortem findings
/evolve --quality --max-cycles=10  # Quality mode with cycle cap
/evolve --athena             # Mine → Defrag warmup before first cycle
/evolve --athena --max-cycles=5  # Warm knowledge base then run 5 cycles
/evolve --test-first         # Default strict-quality /rpi execution path
/evolve --no-test-first      # Explicit opt-out from test-first mode
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--max-cycles=N` | unlimited | Stop after `N` completed cycles |
| `--dry-run` | off | Show planned cycle actions without executing |
| `--beads-only` | off | Skip goal measurement and run backlog-only selection |
| `--skip-baseline` | off | Skip first-run baseline snapshot |
| `--quality` | off | Prioritize harvested post-mortem findings |
| `--athena` | off | Run `ao mine` + `ao defrag` warmup before cycle 1 |
| `--test-first` | on | Pass strict-quality defaults through to `/rpi` |
| `--no-test-first` | off | Explicitly disable test-first passthrough to `/rpi` |

## Execution Steps

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

### Step 0: Setup

```bash
mkdir -p .agents/evolve
ao lookup --query "autonomous improvement cycle" --limit 5 2>/dev/null || true
```

Recover cycle number, queue/generator streaks, and the last claimed work item from disk (survives context compaction):
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

# Recover generator state and queue claim state
if [ -f .agents/evolve/session-state.json ]; then
  GENERATOR_EMPTY_STREAK=$(jq -r '.generator_empty_streak // 0' .agents/evolve/session-state.json 2>/dev/null || echo 0)
  LAST_SELECTED_SOURCE=$(jq -r '.last_selected_source // empty' .agents/evolve/session-state.json 2>/dev/null || true)
  CLAIMED_WORK_REF=$(jq -r '.claimed_work.ref // empty' .agents/evolve/session-state.json 2>/dev/null || true)
else
  GENERATOR_EMPTY_STREAK=0
  LAST_SELECTED_SOURCE=""
  CLAIMED_WORK_REF=""
fi

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

# Pre-populate quarantine list from cycle history (lightweight local scan)
if [ -f .agents/evolve/cycle-history.jsonl ]; then
  while IFS= read -r goal; do
    QUARANTINED_GOALS[$goal]=true
    echo "Quarantined oscillating goal: $goal"
  done < <(
    jq -r '.target' .agents/evolve/cycle-history.jsonl 2>/dev/null \
    | awk '{
        if (prev != "" && prev != $0) transitions[$0]++
        prev = $0
      }
      END {
        for (g in transitions) if (transitions[g] >= 3) print g
      }'
  )
fi
```

Parse flags: `--max-cycles=N` (default unlimited), `--dry-run`, `--beads-only`, `--skip-baseline`, `--quality`, `--athena`.

Track cycle-level execution state:

```text
evolve_state = {
  cycle: <current cycle number>,
  mode: <standard|quality|beads-only>,
  test_first: <true by default; false only when --no-test-first>,
  generator_empty_streak: <consecutive passes where all generator layers returned nothing>,
  last_selected_source: <harvested|beads|goal|directive|testing|validation|bug-hunt|drift|feature>,
  claimed_work: <null or queue reference being worked>,
  queue_refresh_count: <incremented after every /rpi cycle>
}
```

Persist `evolve_state` to `.agents/evolve/session-state.json` at each cycle boundary, after queue claims, after queue release/finalize, and during teardown. `cycle-history.jsonl` remains the canonical cycle ledger; `session-state.json` carries resume-only state that has not yet earned a committed cycle entry.

### Step 0.2: Athena Warmup (--athena only)

Skip if `--athena` was not passed or if `--dry-run`.

Run the mechanical half of the Athena cycle to surface fresh signal before the first evolve cycle:

```bash
mkdir -p .agents/mine .agents/defrag
echo "Athena warmup: mining signal..."
ao mine --since 26h --quiet 2>/dev/null || echo "(ao mine unavailable — skipping)"

echo "Athena warmup: defrag sweep..."
ao defrag --prune --dedup --quiet 2>/dev/null || echo "(ao defrag unavailable — skipping)"
```

Then read `.agents/mine/latest.json` and `.agents/defrag/latest.json` and note (in 1-2 sentences each):
- Any **orphaned research** files that look relevant to current goals
- Any **code hotspots** (high-CC functions with recent edits) that may be the root cause of failing goals
- Any **duplicate learnings** merged by defrag — context on what's been cleaned up

These notes inform work selection throughout the evolve session. Store them in a session variable (in-memory), not a file.

### Step 0.5: Baseline (first run only)

Skip if `--skip-baseline` or `--beads-only` or baseline already exists.

```bash
if [ ! -f .agents/evolve/fitness-0-baseline.json ]; then
  bash scripts/evolve-capture-baseline.sh \
    --label "era-$(date -u +%Y%m%dT%H%M%SZ)" \
    --timeout 60
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

This writes a fitness snapshot to `.agents/evolve/`. If `ao goals measure` is unavailable or fails, stop with error. The CLI is required for fitness measurement.

### Step 3: Select Work

Selection is a ladder, not a one-shot check. After every productive cycle, return to the TOP of this step and re-read the queue before considering dormancy.

**Step 3.1: Harvested work first**

Read `.agents/rpi/next-work.jsonl` and pick the highest-value unconsumed item for this repo. Prefer:
- exact repo match before `*`, then legacy unscoped entries
- already-harvested concrete implementation work before process work
- higher severity before lower severity

When evolve picks a queue item, **claim it first**:
- set `claim_status: "in_progress"`
- set `claimed_by: "evolve:cycle-N"`
- set `claimed_at: "<timestamp>"`
- keep `consumed: false` until the `/rpi` cycle and regression gate both succeed

If the cycle fails, regresses, or is interrupted before success, release the claim and leave the item available for the next cycle.

**Step 3.2: Open ready beads**

If no harvested item is ready, check `bd ready`. Pick the highest-priority unblocked issue.

**Step 3.3: Failing goals and directive gaps** (skip if `--beads-only`)

First assess directives, then goals:
- top-priority directive gap from `ao goals measure --directives`
- highest-weight failing goals (skip quarantined oscillators)
- lower-weight failing goals

This step exists even when all queued work is empty. Goals are the third source, not the stop condition.

```bash
DIRECTIVES=$(ao goals measure --directives 2>/dev/null)
FAILING=$(jq -r '.goals[] | select(.result=="fail") | .id' .agents/evolve/fitness-latest.json | head -1)
```

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

**Step 3.4: Testing improvements**

When queues and goals are empty, generate concrete testing work instead of idling:
- find packages/files with thin or missing tests
- look for missing regression tests around recent bug-fix paths
- identify flaky or absent headless/runtime smokes

Convert any real finding into durable work:
- add a bead when the work needs tracked backlog ownership, or
- append a queue item under the shared next-work contract when it should flow directly back into `/rpi`

**Step 3.5: Validation tightening and bug-hunt passes**

If testing improvement generation returns nothing, run bug-hunt and validation sweeps:
- missing validation gates
- weak lint/contract coverage
- bug-hunt style audits for risky areas
- stale assumptions between docs, contracts, and runtime truth

Again: convert findings into beads or queue items, then immediately select the highest-priority result and continue.

**Step 3.6: Drift / hotspot / dead-code mining**

If the prior generators are empty, mine for:
- complexity hotspots
- stale TODO/FIXME markers
- dead code
- stale docs
- stale research
- drift between generated artifacts and source-of-truth files

Do not stop here. Normalize findings into tracked work and continue.

**Step 3.7: Feature suggestions**

If all concrete remediation layers are empty, propose one or more specific feature ideas grounded in the repo purpose, write them as durable work, and continue:
- create a bead when the feature needs review/backlog treatment
- or append a queue item with `source: "feature-suggestion"` when it is ready for the next `/rpi` cycle

**Quality mode (`--quality`)** — inverted cascade (findings before directives):

Step 3.0q: Unconsumed high-severity post-mortem findings:
```bash
HIGH=$(jq -r 'select(.consumed==false) | .items[] | select(.severity=="high") | .title' \
  .agents/rpi/next-work.jsonl 2>/dev/null | head -1)
```

Step 3.1q: Unconsumed medium-severity findings.

Step 3.2q: Open ready beads.

Step 3.3q: Emergency gates (weight >= 5) and top directive gaps.

Step 3.4q: Testing improvements.

Step 3.5q: Validation tightening / bug-hunt / drift mining.

Step 3.6q: Feature suggestions.

This inverts the standard cascade only at the top of the ladder: findings BEFORE goals and directives. It does NOT skip the generator layers.

When evolve picks a finding, claim it first in next-work.jsonl:
- Set `claim_status: "in_progress"`, `claimed_by: "evolve-quality:cycle-N"`, `claimed_at: "<timestamp>"`
- Set `consumed: true` only after the /rpi cycle and regression gate succeed
- If the /rpi cycle fails (regression), clear the claim and leave `consumed: false`

See `references/quality-mode.md` for scoring and full details.

**Nothing found?** HARD GATE — only consider dormancy after the generator layers also came up empty:

```bash
# Count trailing idle/unchanged entries in cycle-history.jsonl (portable, no tac)
IDLE_STREAK=$(awk '/"result"\s*:\s*"(idle|unchanged)"/{streak++; next} {streak=0} END{print streak+0}' \
  .agents/evolve/cycle-history.jsonl 2>/dev/null)

if [ "$GENERATOR_EMPTY_STREAK" -ge 2 ] && [ "$IDLE_STREAK" -ge 2 ]; then
  # Queue layers are empty AND producer layers were empty for the 3rd consecutive pass — STOP
  echo "Stagnation reached after repeated empty queue + generator passes. Dormancy is the last-resort outcome."
  # go to Teardown — do NOT log another idle entry
fi
```

If the queue layers were empty but a generator pass has not been exhausted 3 times yet, persist the new generator streak in `session-state.json` and loop back to Step 1. Empty pre-cycle queues are not a stop reason by themselves.

A cycle is idle only if NO work source returned actionable work and every generator layer also came up empty. A cycle that targeted an oscillating goal and skipped it counts as idle only after the remaining ladder was exhausted.

If `--dry-run`: report what would be worked on and go to Teardown.

### Step 4: Execute

Primary engine: use `/rpi` for any implementation-quality work. `/implement` and `/crank` are allowed only when a bead already contains execution-ready scope and skipping discovery is clearly the better path.

For a **harvested item, failing goal, directive gap, testing improvement, validation tightening task, bug-hunt result, drift finding, or feature suggestion**:
```
Invoke /rpi "{normalized work title}" --auto --max-cycles=1
```

For a **beads issue**:
```
Prefer: /rpi "Land {issue_id}: {title}" --auto --max-cycles=1
Fallback: /implement {issue_id}
```
Or for an epic with children: `Invoke /crank {epic_id}`.

If Step 3 created durable work instead of executing it immediately, re-enter Step 3 and let the newly-created queue/bead item win through the normal selection order.

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
if [ -f scripts/check-wiring-closure.sh ]; then
  bash scripts/check-wiring-closure.sh
else
  echo "WARNING: scripts/check-wiring-closure.sh not found — skipping wiring check"
fi
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

Queue finalization after the regression gate:
- **success:** finalize any claimed queue item with `consumed: true`, `consumed_by`, and `consumed_at`; clear transient claim fields
- **failure/regression:** clear `claim_status`, `claimed_by`, and `claimed_at`; keep `consumed: false`; record the release in `session-state.json`

After the cycle's `/post-mortem` finishes, immediately re-read `.agents/rpi/next-work.jsonl` before selecting the next item. Never assume the queue state from before the cycle.

### Step 6: Log Cycle + Commit

Two paths: productive cycles get committed, idle cycles are local-only.

**PRODUCTIVE cycles** (result is improved, regressed, or harvested):

```bash
# Quality mode: compute quality_score BEFORE writing the JSONL entry
QUALITY_SCORE_ARGS=()
if [ "$QUALITY_MODE" = "true" ]; then
  REMAINING_HIGH=$(jq -r 'select(.consumed==false) | .items[] | select(.severity=="high")' \
    .agents/rpi/next-work.jsonl 2>/dev/null | wc -l | tr -d ' ')
  REMAINING_MEDIUM=$(jq -r 'select(.consumed==false) | .items[] | select(.severity=="medium")' \
    .agents/rpi/next-work.jsonl 2>/dev/null | wc -l | tr -d ' ')
  QUALITY_SCORE=$((100 - (REMAINING_HIGH * 10) - (REMAINING_MEDIUM * 3)))
  [ "$QUALITY_SCORE" -lt 0 ] && QUALITY_SCORE=0
  QUALITY_SCORE_ARGS=(--quality-score "$QUALITY_SCORE")
fi

ENTRY_JSON="$(
  bash scripts/evolve-log-cycle.sh \
    --cycle "$CYCLE" \
    --target "$TARGET" \
    --result "$OUTCOME" \
    --canonical-sha "$(git rev-parse --short HEAD)" \
    --cycle-start-sha "$CYCLE_START_SHA" \
    --goals-passing "$PASSING" \
    --goals-total "$TOTAL" \
    "${QUALITY_SCORE_ARGS[@]}"
)"
OUTCOME="$(printf '%s\n' "$ENTRY_JSON" | jq -r '.result')"
REAL_CHANGES=$(git diff --name-only "${CYCLE_START_SHA}..HEAD" -- ':!.agents/**' ':!GOALS.yaml' ':!GOALS.md' \
  2>/dev/null | wc -l | tr -d ' ')

# Telemetry
bash scripts/log-telemetry.sh evolve cycle-complete cycle=${CYCLE} goal=${TARGET} outcome=${OUTCOME} 2>/dev/null || true

if [ "$OUTCOME" = "unchanged" ]; then
  # No-delta cycle: leave local-only so history stays honest and stagnation logic can see it.
  :
elif [ "$REAL_CHANGES" -gt 0 ]; then
  # Full commit: real code was changed
  git add .agents/evolve/cycle-history.jsonl
  git commit -m "evolve: cycle ${CYCLE} -- ${TARGET} ${OUTCOME}"
else
  # Productive cycle with non-agent repo delta already committed by a sub-skill:
  # stage the ledger but do not create a standalone follow-up commit.
  git add .agents/evolve/cycle-history.jsonl
fi

PRODUCTIVE_THIS_SESSION=$((PRODUCTIVE_THIS_SESSION + 1))
```

**IDLE cycles** (nothing found even after generator layers):

```bash
bash scripts/evolve-log-cycle.sh \
  --cycle "$CYCLE" \
  --target "idle" \
  --result "unchanged" >/dev/null
# No git add, no git commit, no fitness snapshot write
```

### Step 7: Loop or Stop

```bash
while true; do
  # Step 1 .. Step 6
  # Stop if kill switch, max-cycles, or a real safety breaker triggers
  # Otherwise increment cycle and re-enter selection
  CYCLE=$((CYCLE + 1))
done
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
2. Run `/post-mortem "evolve session: ${CYCLE} cycles"` to harvest learnings.
3. Push only if unpushed commits exist:
```bash
UNPUSHED=$(git log origin/main..HEAD --oneline 2>/dev/null | wc -l)
[ "$UNPUSHED" -gt 0 ] && git push
```
4. Report summary:

```
## /evolve Complete
Cycles: N | Productive: X | Regressed: Y (reverted) | Idle: Z
Stop reason: stagnation | circuit-breaker | max-cycles | kill-switch
```

In quality mode, the report includes additional fields:
```
## /evolve Complete (quality mode)
Cycles: N | Findings resolved: X | Goals fixed: Y | Idle: Z
Quality score: start → end (delta)
Remaining unconsumed: H high, M medium
Stop reason: stagnation | circuit-breaker | max-cycles | kill-switch
```

## Examples

**User says:** `/evolve --max-cycles=5`
**What happens:** Evolve re-enters the full selection ladder after every `/rpi` cycle and runs producer layers instead of idling on empty queues.

**User says:** `/evolve --beads-only`
**What happens:** Evolve skips goals measurement and works through `bd ready` backlog.

**User says:** `/evolve --dry-run`
**What happens:** Evolve shows what would be worked on without executing.

**User says:** `/evolve --athena`
**What happens:** Evolve runs `ao mine` + `ao defrag` at session start to surface fresh signal (orphaned research, code hotspots, oscillating goals) before the first evolve cycle. Use before a long autonomous run or after a burst of development activity.

**User says:** `/evolve`
**What happens:** See `references/examples.md` for a worked overnight flow that moves through beads -> harvested work -> goals -> testing -> bug hunt -> feature suggestion before dormancy is considered.

See `references/examples.md` for detailed walkthroughs.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Loop exits immediately | Remove `~/.config/evolve/KILL` or `.agents/evolve/STOP` |
| Stagnation after repeated empty passes | Queue layers and producer layers were empty across multiple passes — dormancy is the fallback outcome |
| `ao goals measure` hangs | Use `--timeout 30` flag or `--beads-only` to skip |
| Regression gate reverts | Review reverted changes, narrow scope, re-run; claimed queue items must be released back to available state |

See `references/cycle-history.md` for advanced troubleshooting.

## References

- `references/cycle-history.md` — JSONL format, recovery protocol, kill switch
- `references/compounding.md` — Knowledge flywheel and work harvesting
- `references/goals-schema.md` — GOALS.yaml format and continuous metrics
- `references/parallel-execution.md` — Parallel /swarm architecture
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
