# docs/council-log/

Committed provenance for high-stakes `/council` validations.

## Purpose

When a council verdict is load-bearing for a merged commit (e.g., it caught a stale scope estimate, blocked a flawed plan, or validated an architectural decision), the consolidated report belongs in git history — not in the transient `.agents/council/` directory. This directory exists so future readers of `git log` can trace **why** a decision was made back to the multi-judge evidence that supported it.

## How artifacts land here

A `/council` invocation with `--commit-ready` writes its consolidated report to `docs/council-log/YYYY-MM-DD-<type>-<target>.md` in addition to the usual `.agents/council/` path. Per-judge reports stay in `.agents/council/` (transient, gitignored); only the **consolidated verdict with evidence-mode test assertions** lands here.

Example:
```
/council --evidence --commit-ready validate "M8 C1 Option A scope estimate"
```

Produces:
- `.agents/council/2026-04-11-m8-option-a-judge1.md` (local, transient)
- `.agents/council/2026-04-11-m8-option-a-judge2.md` (local, transient)
- `docs/council-log/2026-04-11-validate-m8-option-a.md` (committed, shareable)

## What belongs here (and what doesn't)

**DO commit:**
- Councils that changed a load-bearing decision (scope estimate, architectural choice, migration go/no-go)
- Councils that caught a stale or incorrect prior-session claim
- Councils whose verdict is referenced by a commit message or PR body
- `--evidence`-mode councils with test assertions future readers can re-verify

**DO NOT commit:**
- Routine `/vibe` checks mid-implementation (those belong in `.agents/council/`)
- Brainstorming exploration (ephemeral)
- Research councils that produced no actionable verdict
- Debug councils run to help the agent think (not decisions)

## File naming convention

```
docs/council-log/YYYY-MM-DD-<mode>-<target-slug>.md
```

Where:
- `YYYY-MM-DD` is the UTC date of the council invocation
- `<mode>` is `validate`, `brainstorm`, or `research`
- `<target-slug>` is a kebab-case summary of what was validated (e.g., `m8-option-a`, `auth-migration-plan`)

Multiple councils on the same day with the same target get a `-<n>` suffix: `2026-04-11-validate-m8-option-a-2.md`.

## Relationship to `.agents/council/`

| Aspect | `.agents/council/` | `docs/council-log/` |
|---|---|---|
| Gitignored? | Yes | No |
| Content | All council output (per-judge + consolidated + packet) | Consolidated report only |
| Purpose | Transient working memory | Durable provenance |
| Retention | Until flywheel prune / local cleanup | Permanent (survives rebase) |
| Who reads it | Current session's agents | Future agents + humans reviewing git history |

## Entry format

Each committed entry should lead with:

```markdown
---
title: <Human-readable title>
date: YYYY-MM-DD
mode: validate | brainstorm | research
flags: --evidence --mixed ...
verdict: PASS | WARN | FAIL
confidence: HIGH | MEDIUM | LOW
judge_count: N
---
```

Followed by:
1. TL;DR — what was decided and why
2. Consolidated findings (each with `test_assertions` when `--evidence` was set)
3. Disagreements (if any) with attribution
4. Related commits/beads/PRs

See `2026-04-11-validate-m8-assumption-validation.md` for the canonical example (the council that caught the na-h61 scope inflation).
