---
name: goals
description: 'Maintain GOALS.yaml and GOALS.md fitness specifications. Measure fitness, manage directives, track drift, add/prune goals. Triggers: "goals", "goal status", "show goals", "add goals", "prune goals", "clean goals", "goal drift", "goal history", "export goals", "meta goals", "migrate goals".'
---


# $goals — Fitness Goal Maintenance

> Maintain GOALS.yaml and GOALS.md fitness specifications. Use `ao work goals` CLI for all operations.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

## Quick Start

```bash
$goals                    # Measure fitness (default)
$goals init               # Bootstrap GOALS.md interactively
$goals steer              # Manage directives
$goals add                # Add a new goal
$goals drift              # Compare snapshots for regressions
$goals history            # Show measurement history
$goals export             # Export snapshot as JSON for CI
$goals meta               # Run meta-goals only
$goals validate           # Validate structure
$goals prune              # Remove stale gates
$goals migrate            # Migrate YAML to Markdown
```

## Format Support

| Format | File | Version | Features |
|--------|------|---------|----------|
| YAML | GOALS.yaml | 1-3 | Goals with checks, weights, pillars |
| Markdown | GOALS.md | 4 | Goals + mission + north/anti stars + directives |

When both files exist, GOALS.md takes precedence.

## Mode Selection

Parse the user's input:

| Input | Mode | CLI Command |
|-------|------|-------------|
| `$goals`, `$goals measure`, "goal status" | **measure** | `ao work goals measure` |
| `$goals init`, "bootstrap goals" | **init** | `ao work goals init` |
| `$goals steer`, "manage directives" | **steer** | `ao work goals steer` |
| `$goals add`, "add goal" | **add** | `ao work goals add` |
| `$goals drift`, "goal drift" | **drift** | `ao work goals drift` |
| `$goals history`, "goal history" | **history** | `ao work goals history` |
| `$goals export`, "export goals" | **export** | `ao work goals export` |
| `$goals meta`, "meta goals" | **meta** | `ao work goals meta` |
| `$goals validate`, "validate goals" | **validate** | `ao work goals validate` |
| `$goals prune`, "prune goals", "clean goals" | **prune** | `ao work goals prune` |
| `$goals migrate`, "migrate goals" | **migrate** | `ao work goals migrate` |

## Measure Mode (default) — Observe

### Step 1: Run Measurement

```bash
ao work goals measure --json
```

Parse the JSON output. Extract per-goal pass/fail, overall fitness score.

### Step 2: Directive Gap Assessment (GOALS.md only)

If the goals file is GOALS.md format:

```bash
ao work goals measure --directives
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
ao work goals init
```

Or with defaults:
```bash
ao work goals init --non-interactive
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
ao work goals steer add "Title" --description="..." --steer=increase
ao work goals steer remove 3
ao work goals steer prioritize 2 1
```

## Add Mode

Add a single goal to the goals file. Format-aware — writes to GOALS.yaml or GOALS.md depending on which format is detected.

```bash
ao work goals add <id> <check-command> --weight=5 --description="..." --type=health
```

| Flag | Default | Description |
|------|---------|-------------|
| `--weight` | 5 | Goal weight (1-10) |
| `--description` | — | Human-readable description |
| `--type` | — | Goal type (health, architecture, quality, meta) |

Example:
```bash
ao work goals add go-coverage-floor "bash scripts/check-coverage.sh" --weight=3 --description="Go test coverage above 60%"
```

## Drift Mode

Compare the latest measurement snapshot against a previous one to detect regressions.

```bash
ao work goals drift                    # Compare latest vs previous snapshot
ao work goals drift --since 2026-02-20  # Compare against a specific date
```

Reports which goals improved, regressed, or stayed unchanged.

## History Mode

Show measurement history over time for all goals or a specific goal.

```bash
ao work goals history                        # All goals, all time
ao work goals history --goal go-coverage     # Single goal
ao work goals history --since 2026-02-01     # Since a specific date
ao work goals history --goal go-coverage --since 2026-02-01  # Combined
```

Useful for spotting trends and identifying oscillating goals.

## Export Mode

Export the latest fitness snapshot as JSON for CI consumption or external tooling.

```bash
ao work goals export
```

Outputs the snapshot to stdout in the fitness snapshot schema (see `references/goals-schema.md`).

## Meta Mode

Run only meta-goals (goals that validate the validation system itself). Useful for checking allowlist hygiene, skip-list freshness, and other self-referential checks.

```bash
ao work goals meta --json
```

See `references/goals-schema.md` for the meta-goal pattern.

## Validate Mode

```bash
ao work goals validate --json
```

Reports: goal count, version, format, directive count, any structural errors or warnings.

## Prune Mode

```bash
ao work goals prune --dry-run    # List stale gates
ao work goals prune              # Remove stale gates
```

Identifies gates whose check commands reference nonexistent paths. Removes them and re-renders the file.

## Migrate Mode

Convert between goal file formats.

```bash
ao work goals migrate --to-md      # Convert GOALS.yaml → GOALS.md
ao work goals migrate               # Migrate GOALS.yaml to latest YAML version
```

The `--to-md` flag creates a GOALS.md with mission, north/anti stars sections, and converts existing goals into the Gates table format. The original YAML file is backed up.

## Examples

### Checking fitness and directive gaps

**User says:** `$goals`

**What happens:**
1. Runs `ao work goals measure --json` to get gate results
2. If GOALS.md format, runs `ao work goals measure --directives` to get directive list
3. Assesses each directive against recent work
4. Reports combined fitness + directive gap dashboard

**Result:** Dashboard showing gate pass rates and directive progress.

### Bootstrapping goals for a new project

**User says:** `$goals init`

**What happens:**
1. Runs `ao work goals init` which prompts for mission, stars, directives, and auto-detects gates
2. Creates GOALS.md in the project root

**Result:** New GOALS.md ready for `$evolve` consumption.

### Tracking regressions across releases

**User says:** `$goals drift --since 2026-02-20`

**What happens:**
1. Runs `ao work goals drift --since 2026-02-20`
2. Compares current snapshot against the one from that date
3. Reports improved/regressed/unchanged goals

**Result:** Regression report highlighting which goals changed direction.

### Adding a new goal after a post-mortem

**User says:** `$goals add go-parser-fuzz "cd cli && go test -fuzz=. ./internal/goals/ -fuzztime=10s" --weight=3 --description="Markdown parser survives fuzz testing"`

**What happens:**
1. Runs `ao work goals add` with the provided arguments
2. Writes the new goal in the correct format (YAML or Markdown)

**Result:** New goal added, measurable on next `$goals` run.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| "goals file already exists" | Init called on existing project | Use `$goals` to measure, or delete file to re-init |
| "directives require GOALS.md format" | Tried steer on YAML file | Run `ao work goals migrate --to-md` first |
| No directives in measure output | GOALS.yaml doesn't support directives | Migrate to GOALS.md with `ao work goals migrate --to-md` |
| Gates referencing deleted scripts | Scripts were renamed or removed | Run `$goals prune` to clean up |
| Drift shows no history | No prior snapshots saved | Run `ao work goals measure` at least twice first |
| Export returns empty | No snapshot file exists | Run `ao work goals measure` to create initial snapshot |

## See Also

- `$evolve` — consumes goals for fitness-scored improvement loops
- `references/goals-schema.md` — schema definition for both formats
- `references/generation-heuristics.md` — goal quality criteria

## Reference Documents

- [references/generation-heuristics.md](references/generation-heuristics.md)
- [references/goals-schema.md](references/goals-schema.md)

## Local Resources

### references/

- [references/generation-heuristics.md](references/generation-heuristics.md)
- [references/goals-schema.md](references/goals-schema.md)

### scripts/

- `scripts/validate.sh`


