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
