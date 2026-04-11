> Extracted from council/SKILL.md on 2026-04-11.

# Task Types & First-Pass Rigor Gate

## Task Types

| Type | Trigger Words | Perspective Focus |
|------|---------------|-------------------|
| **validate** | validate, check, review, assess, critique, feedback, improve | Is this correct? What's wrong? What could be better? |
| **brainstorm** | brainstorm, explore, options, approaches | What are the alternatives? Pros/cons? |
| **research** | research, investigate, deep dive, explore deeply, analyze, examine, evaluate, compare | What can we discover? What are the properties, trade-offs, and structure? |

Natural language works — the skill infers task type from your prompt.

## First-pass rigor gate for plan/spec validation (MANDATORY)

When mode is `validate` and the target is a plan/spec/contract (or contains boundary rules, state transitions, or conformance tables), judges must apply this gate before returning `PASS`:

1. Canonical mutation + ack sequence is explicit, single-path, and non-contradictory.
2. Consume-at-most-once path is crash-safe with explicit atomic boundary and restart recovery semantics.
3. Status/precedence behavior is defined with a field-level truth table and anomaly reason codes for conflicting evidence.
4. Conformance includes explicit boundary failpoint tests and deterministic assertions for replay/no-duplicate-effect outcomes.

Verdict policy for this gate:
- Missing or contradictory gate item: minimum `WARN`.
- Missing deterministic conformance coverage for any gate item: minimum `WARN`.
- Critical lifecycle invariant not mechanically verifiable: `FAIL`.

## Quick Mode (`--quick`)

Single-agent inline validation. No subprocess spawning, no Task tool, no Codex. The current agent performs a structured self-review using the same output schema as a full council.

**When to use:** Routine checks, mid-implementation sanity checks, pre-commit quick scan.

**Execution:** Gather context (files, diffs) -> perform structured self-review inline using the council output_schema (verdict, confidence, findings, recommendation) -> write report to `.agents/council/YYYY-MM-DD-quick-<target>.md` labeled as `Mode: quick (single-agent)`.

**Limitations:** No cross-perspective disagreement, no cross-vendor insights, lower confidence ceiling. Not suitable for security audits or architecture decisions.
