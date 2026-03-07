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
  "canonical_sha": "abc1234",
  "timestamp": "2026-02-23T12:00:00-05:00",
  "goals_passing": 59,
  "goals_total": 59
}
```

**Field standardization:**
- Use `target` (not `goal_id`) — this is what recent cycles already use
- Use `sha` as the compatibility alias for `canonical_sha`
- Use `canonical_sha` for the implementation commit the cycle actually delivered
- Use `log_sha` only when the bookkeeping/log commit is distinct from `canonical_sha`
- Always include `goals_passing` and `goals_total` — enables trajectory plotting
- Optional fields: `quality_score` (quality mode), `idle_streak` (idle cycles), `parallel` + `goal_ids` (parallel mode)

**Legacy field names:** Older entries may use `goal_id` instead of `target` and `commit_sha` instead of `sha`. Tools reading cycle-history.jsonl should handle both conventions.

**Sequential cycle entry:**
```jsonl
{"cycle": 1, "target": "test-pass-rate", "result": "improved", "sha": "abc1234", "canonical_sha": "abc1234", "goals_passing": 18, "goals_total": 23, "timestamp": "2026-02-11T21:00:00Z"}
{"cycle": 2, "target": "doc-coverage", "result": "regressed", "sha": "def5678", "canonical_sha": "def5678", "log_sha": "fedcba9", "goals_passing": 17, "goals_total": 23, "timestamp": "2026-02-11T21:30:00Z"}
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
| `sha` | Compatibility alias for the implementation SHA (omitted for idle cycles) |
| `canonical_sha` | Implementation commit the cycle actually delivered |
| `goals_passing` | Count of goals with result "pass" (omitted for idle cycles) |
| `goals_total` | Total goals measured (omitted for idle cycles) |
| `timestamp` | ISO 8601 timestamp |

`log_sha` is optional and should only be written when the log/bookkeeping commit
differs from `canonical_sha`. These fields enable fitness trajectory plotting
without losing retrospective provenance.

### Substantive-Delta Rule

Do not record `result: "improved"` when a cycle produces no non-agent repo delta.
If the cycle touched only `.agents/` artifacts or otherwise made no substantive
repo change, rewrite the outcome to `unchanged` and keep it local-only. This
prevents ledger churn from being misread as product progress.

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
# Productive cycle: log via the canonical writer, then commit
bash scripts/evolve-log-cycle.sh \
  --cycle "$CYCLE" \
  --target "$TARGET" \
  --result "$OUTCOME" \
  --canonical-sha "$(git rev-parse --short HEAD)" \
  --cycle-start-sha "$CYCLE_START_SHA" \
  --goals-passing "$PASSING" \
  --goals-total "$TOTAL"

# Parallel productive cycle:
bash scripts/evolve-log-cycle.sh \
  --cycle "$CYCLE" \
  --target "parallel-wave" \
  --goal-ids "${goal_ids_csv}" \
  --parallel \
  --result "$OUTCOME" \
  --canonical-sha "$(git rev-parse --short HEAD)" \
  --goals-passing "$PASSING" \
  --goals-total "$TOTAL"

# Idle or no-delta cycle: append locally, do NOT commit
bash scripts/evolve-log-cycle.sh --cycle "$CYCLE" --target "idle" --result "unchanged" >/dev/null
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
4. The active baseline pointer (`active-baseline.txt`) is preserved -- do not regenerate the current era baseline

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

## Flags Reference

| Flag | Default | Description |
|------|---------|-------------|
| `--max-cycles=N` | unlimited | Optional hard cap. Without this, loop runs forever. |
| `--test-first` | off | Pass `--test-first` through to `/rpi` -> `/crank` |
| `--dry-run` | off | Measure fitness and show plan, don't execute |
| `--skip-baseline` | off | Skip cycle-0 baseline sweep |
| `--parallel` | off | Enable parallel goal execution via /swarm per cycle |
| `--max-parallel=N` | 3 | Max goals to fix in parallel (cap: 5). Only with `--parallel`. |

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `/evolve` exits immediately with "KILL SWITCH ACTIVE" | Kill switch file exists | Remove `~/.config/evolve/KILL` or `.agents/evolve/STOP` to re-enable |
| "No goals to measure" error | GOALS.yaml missing or empty | Create GOALS.yaml in repo root with fitness goals (see goals-schema.md) |
| Cycle completes but fitness unchanged | Goal check command is always passing or always failing | Verify check command logic in GOALS.yaml produces exit code 0 (pass) or non-zero (fail) |
| Regression revert fails | Multiple commits in cycle or uncommitted changes | Check cycle-start SHA in fitness snapshot, commit or stash changes before retrying |
| Harvested work never consumed | All goals passing but `next-work.jsonl` not read | Check file exists and has `consumed: false` entries. Agent picks harvested work after goals met. |
| Loop stops after N cycles | `--max-cycles` was set (or old default of 10) | Omit `--max-cycles` flag -- default is now unlimited. Loop runs until kill switch or 3 idle cycles. |
