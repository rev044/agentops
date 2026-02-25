---
name: goals
description: 'Maintain GOALS.yaml and GOALS.md fitness specifications. Measure fitness, manage directives, track drift, add/prune goals. Triggers: "goals", "goal status", "show goals", "add goals", "prune goals", "clean goals", "goal drift", "goal history", "export goals", "meta goals", "migrate goals".'
---


# $goals — Fitness Goal Maintenance

> Maintain GOALS.yaml and GOALS.md fitness specifications. Use `ao goals` CLI for all operations.

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
| `$goals`, `$goals measure`, "goal status" | **measure** | `ao goals measure` |
| `$goals init`, "bootstrap goals" | **init** | `ao goals init` |
| `$goals steer`, "manage directives" | **steer** | `ao goals steer` |
| `$goals add`, "add goal" | **add** | `ao goals add` |
| `$goals drift`, "goal drift" | **drift** | `ao goals drift` |
| `$goals history`, "goal history" | **history** | `ao goals history` |
| `$goals export`, "export goals" | **export** | `ao goals export` |
| `$goals meta`, "meta goals" | **meta** | `ao goals meta` |
| `$goals validate`, "validate goals" | **validate** | `ao goals validate` |
| `$goals prune`, "prune goals", "clean goals" | **prune** | `ao goals prune` |
| `$goals migrate`, "migrate goals" | **migrate** | `ao goals migrate` |

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

## Add Mode

Add a single goal to the goals file. Format-aware — writes to GOALS.yaml or GOALS.md depending on which format is detected.

```bash
ao goals add <id> <check-command> --weight=5 --description="..." --type=health
```

| Flag | Default | Description |
|------|---------|-------------|
| `--weight` | 5 | Goal weight (1-10) |
| `--description` | — | Human-readable description |
| `--type` | — | Goal type (health, architecture, quality, meta) |

Example:
```bash
ao goals add go-coverage-floor "bash scripts/check-coverage.sh" --weight=3 --description="Go test coverage above 60%"
```

## Drift Mode

Compare the latest measurement snapshot against a previous one to detect regressions.

```bash
ao goals drift                    # Compare latest vs previous snapshot
ao goals drift --since 2026-02-20  # Compare against a specific date
```

Reports which goals improved, regressed, or stayed unchanged.

## History Mode

Show measurement history over time for all goals or a specific goal.

```bash
ao goals history                        # All goals, all time
ao goals history --goal go-coverage     # Single goal
ao goals history --since 2026-02-01     # Since a specific date
ao goals history --goal go-coverage --since 2026-02-01  # Combined
```

Useful for spotting trends and identifying oscillating goals.

## Export Mode

Export the latest fitness snapshot as JSON for CI consumption or external tooling.

```bash
ao goals export
```

Outputs the snapshot to stdout in the fitness snapshot schema (see `references/goals-schema.md`).

## Meta Mode

Run only meta-goals (goals that validate the validation system itself). Useful for checking allowlist hygiene, skip-list freshness, and other self-referential checks.

```bash
ao goals meta --json
```

See `references/goals-schema.md` for the meta-goal pattern.

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

## Migrate Mode

Convert between goal file formats.

```bash
ao goals migrate --to-md      # Convert GOALS.yaml → GOALS.md
ao goals migrate               # Migrate GOALS.yaml to latest YAML version
```

The `--to-md` flag creates a GOALS.md with mission, north/anti stars sections, and converts existing goals into the Gates table format. The original YAML file is backed up.

## Examples

### Checking fitness and directive gaps

**User says:** `$goals`

**What happens:**
1. Runs `ao goals measure --json` to get gate results
2. If GOALS.md format, runs `ao goals measure --directives` to get directive list
3. Assesses each directive against recent work
4. Reports combined fitness + directive gap dashboard

**Result:** Dashboard showing gate pass rates and directive progress.

### Bootstrapping goals for a new project

**User says:** `$goals init`

**What happens:**
1. Runs `ao goals init` which prompts for mission, stars, directives, and auto-detects gates
2. Creates GOALS.md in the project root

**Result:** New GOALS.md ready for `$evolve` consumption.

### Tracking regressions across releases

**User says:** `$goals drift --since 2026-02-20`

**What happens:**
1. Runs `ao goals drift --since 2026-02-20`
2. Compares current snapshot against the one from that date
3. Reports improved/regressed/unchanged goals

**Result:** Regression report highlighting which goals changed direction.

### Adding a new goal after a post-mortem

**User says:** `$goals add go-parser-fuzz "cd cli && go test -fuzz=. ./internal/goals/ -fuzztime=10s" --weight=3 --description="Markdown parser survives fuzz testing"`

**What happens:**
1. Runs `ao goals add` with the provided arguments
2. Writes the new goal in the correct format (YAML or Markdown)

**Result:** New goal added, measurable on next `$goals` run.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| "goals file already exists" | Init called on existing project | Use `$goals` to measure, or delete file to re-init |
| "directives require GOALS.md format" | Tried steer on YAML file | Run `ao goals migrate --to-md` first |
| No directives in measure output | GOALS.yaml doesn't support directives | Migrate to GOALS.md with `ao goals migrate --to-md` |
| Gates referencing deleted scripts | Scripts were renamed or removed | Run `$goals prune` to clean up |
| Drift shows no history | No prior snapshots saved | Run `ao goals measure` at least twice first |
| Export returns empty | No snapshot file exists | Run `ao goals measure` to create initial snapshot |

## See Also

- `$evolve` — consumes goals for fitness-scored improvement loops
- `references/goals-schema.md` — schema definition for both formats
- `references/generation-heuristics.md` — goal quality criteria

## Reference Documents

- [references/generation-heuristics.md](references/generation-heuristics.md)
- [references/goals-schema.md](references/goals-schema.md)

---

## References

### generation-heuristics.md

# Goal Generation Heuristics

## Goal Quality Criteria

A good goal:

1. **Mechanically verifiable** — `check` is a shell command that exits 0 (pass) or non-zero (fail). No human judgment required.
2. **Descriptive** — `description` says what it measures, not how. "Go CLI compiles without errors" not "run go build".
3. **Weighted by impact** — 5 = build/test integrity, 3-4 = feature fitness, 1-2 = hygiene.
4. **Pillar-mapped** — Maps to one of: knowledge-compounding, validated-acceleration, goal-driven-automation, zero-friction-workflow. Infrastructure goals omit `pillar`.
5. **Not trivially true** — Check can actually fail in a realistic scenario. `test -f README.md` is trivially true.
6. **Not duplicative** — No two goals test the same thing. Check existing IDs before proposing.

## Scan Sources

| Source | What to look for | Goal type |
|--------|-----------------|-----------|
| `PRODUCT.md` | Value props, design principles, theoretical pillars without goals | Pillar |
| `README.md` | Claims, badges, features without verification | Pillar |
| `skills/*/SKILL.md` | Skills with no goal referencing them | Pillar or Infra |
| `tests/`, `hooks/` | Scripts not covered by goals | Infrastructure |
| `docs/` | Doc files referenced but not covered | Infrastructure |
| Existing goals | Checks referencing deleted paths | Prune candidates |

## Theoretical Pillar Coverage

Generate mode should check that all 4 theoretical pillars have goals:

### 1. Systems Theory (Meadows)

Targets leverage points #3-#6 (information flows, rules, self-organization, goals). Goals should verify that the system operates at these leverage points rather than lower ones (parameters, buffers).

### 2. DevOps (Three Ways)

- **Flow** maps to `zero-friction-workflow` and `goal-driven-automation`
- **Feedback** maps to `validated-acceleration`
- **Continual Learning** maps to `knowledge-compounding`

Goals should cover all three ways.

### 3. Brownian Ratchet

The pattern: chaos + filter + ratchet = directional progress from undirected energy. Goals should verify:
- Chaos source exists (agent sessions generate varied outputs)
- Filter exists (council validates, vibe checks)
- Ratchet exists (knowledge flywheel captures and persists gains)

### 4. Knowledge Flywheel

Escape velocity condition: `signal_rate x retrieval_rate > decay_rate` (informally: you learn faster than you forget). Goals should verify:
- Signal generation (extract, forge, retro produce learnings)
- Retrieval (inject loads learnings into sessions)
- Decay resistance (learnings are persisted, not just in-memory)

## Weight Guidelines

| Weight | Category | Examples |
|--------|----------|----------|
| 5 | **Critical** | Build passes, tests pass, manifests valid |
| 4 | **Important** | Full test suite, hook safety, mission alignment |
| 3 | **Feature fitness** | Skill behaviors, positioning, documentation |
| 2 | **Hygiene** | Lint, coverage floors, doc counts |
| 1 | **Nice to have** | Stubs, aspirational checks |

## ID Conventions

- Use kebab-case: `go-cli-builds`, `readme-compounding-hero`
- Prefix with domain: `readme-`, `go-`, `skill-`, `hook-`
- Keep under 40 characters
- Must be unique across all goals

## Directive Quality Criteria

When generating or evaluating directives for GOALS.md:

1. **Actionable** — Describes work that can be decomposed into issues. "Expand test coverage" not "Be better at testing."
2. **Steerable** — Has a clear direction (increase/decrease/hold/explore). If you can't assign a steer, it's too vague.
3. **Measurable progress** — You can tell whether work addressed it (even if not fully completed).
4. **Not a gate** — Directives describe intent, not pass/fail thresholds. "Reduce complexity" is a directive; "complexity < 15" is a gate.
5. **Prioritized** — Lower number = higher priority. Directive 1 is worked before directive 2.

### Steer Values

| Steer | Meaning | Example |
|-------|---------|---------|
| `increase` | Do more of this | "Expand test coverage" |
| `decrease` | Reduce this | "Reduce complexity budget" |
| `hold` | Maintain current level | "Keep API compatibility" |
| `explore` | Investigate options | "Evaluate new CI provider" |

### Directive-Gate Relationship

Directives generate gates over time:
- Directive "Expand test coverage" → Gate `test-coverage-floor` (check: coverage > 80%)
- Directive "Reduce complexity" → Gate `complexity-budget` (check: gocyclo -over 15 = 0 findings)

When a directive is fully addressed (gate exists and passes), consider removing the directive and keeping the gate.

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

GOALS.md extends the YAML format with strategic intent sections:

```markdown
# Goals

<Mission statement — one sentence>

## North Stars

- <Aspiration 1>
- <Aspiration 2>

## Anti Stars

- <What we explicitly avoid>

## Directives

### 1. <Title>

<Description of the strategic intent>

**Steer:** increase | decrease | hold | explore

### 2. <Title>

<Description>

**Steer:** <direction>

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| build-passing | `cd cli && make build` | 8 | CLI builds without errors |
| test-passing | `cd cli && make test` | 7 | All unit tests pass |
```

### Key Differences from YAML

| Feature | YAML (v1-3) | Markdown (v4) |
|---------|-------------|---------------|
| Goals/Gates | `goals:` array | `## Gates` table |
| Mission | `mission:` field | First paragraph after `# Goals` |
| Directives | Not supported | `## Directives` section |
| North/Anti Stars | Not supported | `## North Stars` / `## Anti Stars` |
| Version | `version: N` | Implicit (always 4) |

### CLI Commands

```bash
ao goals measure                  # Measure gates (both formats)
ao goals measure --directives     # Output directives as JSON
ao goals validate                 # Validate structure
ao goals init                     # Bootstrap GOALS.md interactively
ao goals steer add <title>        # Add directive
ao goals steer remove <number>    # Remove directive
ao goals steer prioritize <n> <p> # Reorder directive
ao goals migrate --to-md          # Convert YAML → Markdown
ao goals prune                    # Remove stale gates
```

### Format Auto-Detection

`LoadGoals()` auto-detects format:
1. `.md` extension → markdown parser
2. `.yaml`/`.yml` extension → check if `GOALS.md` exists alongside → prefer markdown
3. Default `GOALS.yaml` path → check if `GOALS.md` exists → prefer markdown


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
check "SKILL.md has name: goals" "grep -q '^name: goals' '$SKILL_DIR/SKILL.md'"
check "SKILL.md has tier: product" "grep -q '^[[:space:]]*tier:[[:space:]]*product' '$SKILL_DIR/SKILL.md'"
check "SKILL.md documents status mode" "grep -q '## Status Mode' '$SKILL_DIR/SKILL.md'"
check "SKILL.md documents generate mode" "grep -q '## Generate Mode' '$SKILL_DIR/SKILL.md'"
check "SKILL.md documents prune mode" "grep -q '## Prune Mode' '$SKILL_DIR/SKILL.md'"
check "SKILL.md references GOALS.yaml" "grep -q 'GOALS.yaml' '$SKILL_DIR/SKILL.md'"
check "SKILL.md references $evolve" "grep -q '$evolve' '$SKILL_DIR/SKILL.md'"
check "references/generation-heuristics.md exists" "[ -f '$SKILL_DIR/references/generation-heuristics.md' ]"
check "generation-heuristics has quality criteria" "grep -q 'Quality Criteria' '$SKILL_DIR/references/generation-heuristics.md'"
check "generation-heuristics has scan sources" "grep -q 'Scan Sources' '$SKILL_DIR/references/generation-heuristics.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


