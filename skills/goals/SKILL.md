---
name: goals
description: 'Maintain GOALS.yaml fitness specification. Generate new goals from repo state, prune stale goals, update drifted checks. Triggers: "goals", "goal status", "show goals", "generate goals", "add goals", "prune goals", "update goals", "clean goals".'
skill_api_version: 1
metadata:
  tier: product
  dependencies: []
---

# /goals — Fitness Goal Maintenance

> Maintain GOALS.yaml and GOALS.md fitness specifications. Use `ao goals` CLI for all operations.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

## Quick Start

```bash
/goals                    # Measure fitness (default)
/goals init               # Bootstrap GOALS.md interactively
/goals steer              # Manage directives
/goals validate           # Validate structure
/goals prune              # Remove stale gates
```

## Format Support

| Format | File | Version | Features |
|--------|------|---------|----------|
| YAML | GOALS.yaml | 1-3 | Goals with checks, weights, pillars |
| Markdown | GOALS.md | 4 | Goals + mission + north/anti stars + directives |

When both files exist, GOALS.md takes precedence.

## Mode Selection (5 OODA Verbs)

Parse the user's input:

| Input | Mode | OODA Phase |
|-------|------|------------|
| `/goals`, `/goals measure`, "goal status" | **measure** | Observe |
| `/goals init`, "bootstrap goals" | **init** | — |
| `/goals steer`, "manage directives" | **steer** | Orient/Decide |
| `/goals validate`, "validate goals" | **validate** | — |
| `/goals prune`, "prune goals", "clean goals" | **prune** | — |

## Measure Mode (default) — Observe

### Step 1: Run Measurement

```bash
ao goals measure --json
```

Parse the JSON output. Extract per-goal pass/fail, overall fitness score.

### Step 2: Directive Gap Assessment (GOALS.md only)

If the goals file is GOALS.md format:

```bash
ao goals measure --directives
```

For each directive, assess whether recent work has addressed it:
- Check git log for commits mentioning the directive title
- Check beads/issues related to the directive topic
- Rate each directive: addressed / partially-addressed / gap

### Step 3: Report

Present fitness dashboard:
```
Fitness: 5/7 passing (71%)

Gates:
  [PASS] build-passing (weight 8)
  [FAIL] test-passing (weight 7)
    └─ 3 test failures in pool_test.go

Directives:
  1. Expand Test Coverage — gap (no recent test additions)
  2. Reduce Complexity — partially-addressed (2 refactors this week)
```

## Init Mode

```bash
ao goals init
```

Or with defaults:
```bash
ao goals init --non-interactive
```

Creates a new GOALS.md with mission, north/anti stars, first directive, and auto-detected gates. Error if file already exists.

## Steer Mode — Orient/Decide

### Step 1: Show Current State

Run measure mode first to show current fitness and directive status.

### Step 2: Propose Adjustments

Based on measurement:
- If a directive is fully addressed → suggest removing or replacing
- If fitness is declining → suggest new gates
- If idle rate is high → suggest new directives

### Step 3: Execute Changes

Use CLI commands:
```bash
ao goals steer add "Title" --description="..." --steer=increase
ao goals steer remove 3
ao goals steer prioritize 2 1
```

## Validate Mode

```bash
ao goals validate --json
```

Reports: goal count, version, format, directive count, any structural errors or warnings.

## Prune Mode

```bash
ao goals prune --dry-run    # List stale gates
ao goals prune              # Remove stale gates
```

Identifies gates whose check commands reference nonexistent paths. Removes them and re-renders the file.

## Examples

### Checking fitness and directive gaps

**User says:** `/goals`

**What happens:**
1. Runs `ao goals measure --json` to get gate results
2. If GOALS.md format, runs `ao goals measure --directives` to get directive list
3. Assesses each directive against recent work
4. Reports combined fitness + directive gap dashboard

**Result:** Dashboard showing gate pass rates and directive progress.

### Bootstrapping goals for a new project

**User says:** `/goals init`

**What happens:**
1. Runs `ao goals init` which prompts for mission, stars, directives, and auto-detects gates
2. Creates GOALS.md in the project root

**Result:** New GOALS.md ready for `/evolve` consumption.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| "goals file already exists" | Init called on existing project | Use `/goals` to measure, or delete file to re-init |
| "directives require GOALS.md format" | Tried steer on YAML file | Run `ao goals migrate --to-md` first |
| No directives in measure output | GOALS.yaml doesn't support directives | Migrate to GOALS.md with `ao goals migrate --to-md` |
| Gates referencing deleted scripts | Scripts were renamed or removed | Run `/goals prune` to clean up |

## See Also

- `/evolve` — consumes goals for fitness-scored improvement loops
- `references/goals-schema.md` — schema definition for both formats
- `references/generation-heuristics.md` — goal quality criteria

## Reference Documents

- [references/generation-heuristics.md](references/generation-heuristics.md)
- [references/goals-schema.md](references/goals-schema.md)
