# Fitness Scoring

## Measurement

Run the fitness measurement script to produce a rolling snapshot:

```bash
bash scripts/evolve-measure-fitness.sh \
  --output .agents/evolve/fitness-latest.json \
  --timeout 60 \
  --total-timeout 75
```

**Do NOT write per-cycle `fitness-{N}-pre.json` files.** The rolling file is sufficient for work selection and regression detection.

This writes a fitness snapshot to `.agents/evolve/` atomically via a temp file plus JSON validation. The AgentOps CLI is required for fitness measurement because the wrapper shells out to `ao goals measure`. If measurement exceeds the whole-command bound or returns invalid JSON, the wrapper fails without clobbering the previous rolling snapshot.

## Baseline Capture (First Run Only)

Skip if `--skip-baseline` or `--beads-only` or baseline already exists.

```bash
if [ ! -f .agents/evolve/fitness-0-baseline.json ]; then
  bash scripts/evolve-capture-baseline.sh \
    --label "era-$(date -u +%Y%m%dT%H%M%SZ)" \
    --timeout 60
fi
```

## Post-Cycle Re-Measurement (Regression Gate)

After execution, re-measure to detect regressions:

```bash
bash scripts/evolve-measure-fitness.sh \
  --output .agents/evolve/fitness-latest-post.json \
  --timeout 60 \
  --total-timeout 75 \
  --goal "$GOAL_ID"

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

## Oscillation Detection

Before working a failing goal, check if it has oscillated (improved-to-fail transitions >= 3 times in `cycle-history.jsonl`). If so, quarantine it and try the next failing goal.

```bash
OSC_COUNT=$(jq -r "select(.target==\"$FAILING\") | .result" .agents/evolve/cycle-history.jsonl \
  | awk 'prev=="improved" && $0=="fail" {count++} {prev=$0} END {print count+0}')
if [ "$OSC_COUNT" -ge 3 ]; then
  QUARANTINED_GOALS[$FAILING]=true
  echo "{\"cycle\":${CYCLE},\"target\":\"${FAILING}\",\"result\":\"quarantined\",\"oscillations\":${OSC_COUNT},\"timestamp\":\"$(date -Iseconds)\"}" >> .agents/evolve/cycle-history.jsonl
fi
```

See also: `references/oscillation.md` for full quarantine protocol.
