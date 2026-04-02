# Knowledge Flywheel: Five-Month Empirical Audit

> **Date:** 2026-04-02 | **Evidence:** [.agents/evidence/2026-04-02-flywheel-five-month-audit.md](../.agents/evidence/2026-04-02-flywheel-five-month-audit.md)

## TL;DR

After 5 months and 2,275 learning artifacts across 14 workspaces, the
AgentOps knowledge flywheel does not compound across sessions. Knowledge
is written but never retrieved (0% citation rate). The intra-session
flywheel works (proven by 80% pre-mortem accuracy, 6x ROI). The
cross-session flywheel does not.

## The Promise vs Reality

**The promise:** "Session 1, your agent spends 2 hours debugging. Session
15, a new agent finds the answer in 10 seconds — because the flywheel
promoted the lesson."

**The reality:** 2,275 learning files exist. Zero are cited by any
subsequent plan, pre-mortem, or implementation. The flywheel generates
knowledge every session but nothing mechanically retrieves it at decision
points. Learnings are a write-only log.

## What Works (Intra-Session)

Within a single RPI session, the flywheel compounds reliably:

| Loop | Evidence |
|------|----------|
| Research feeds plan | Symbol-level specs from file:line citations |
| Pre-mortem catches bugs | 80% prediction accuracy, 6x ROI (measured) |
| Failures become learnings | Novel discoveries captured with source beads |
| Post-mortem seeds next work | Follow-up items harvested to next-work.jsonl |

**Source:** [ag-470 case study](../.agents/evidence/2026-04-02-flywheel-case-study.md)

## What Breaks (Cross-Session)

The knowledge chain has five stages. Three are broken:

```
Generate --> Extract --> Store --> Retrieve --> Apply --> Generate
                ^           ^                    ^
            BROKEN       BROKEN              BROKEN
```

### 1. Extraction Produces Garbage

100% of auto-extracted learnings in the primary development rig are
conversation fragments — truncated sentences with no titles, no utility
scores, no source beads. Example: *"didn't work because the goals agent
is still running and re-wrote it."*

### 2. Storage Mixes Signal With Noise

1,612 files in the global store. No quality gate at harvest time. The
same garbage fragment appears in both local and global stores. Metadata
is inconsistent: some rigs tag maturity, others tag utility, none tag both.

### 3. Nothing Retrieves at Decision Points

`ao lookup` exists. No skill calls it. After injection was disabled
(ag-8km), no replacement was added. The mechanical surface that would
close the loop — retrieving prior knowledge during planning — does not
exist.

## Measured State

| Metric | Target | Current |
|--------|--------|---------|
| Citation rate | > 10% | **0%** |
| Retrieval precision (live) | > 0.5 | **0.13** |
| Auto-extract rejection rate | > 50% | **0%** |
| Time-to-first-retrieval | < 2 sessions | **never** |
| Global growth < local creation | yes | **no** |

## Path Forward

Six changes, ordered by impact:

1. **Wire retrieval into decision points.** `/plan` and `/pre-mortem`
   must call `ao lookup` with the current goal and surface relevant
   learnings. This is the single highest-impact change.

2. **Gate auto-extraction quality.** Minimum 50 characters, must have
   a heading, must form a coherent sentence. Reject fragments.

3. **Enforce metadata at write time.** Both maturity and utility required.
   No exceptions.

4. **Purge the global store.** Remove auto-extracted fragments below a
   coherence threshold. Reduce 1,612 to the ~100 that actually contain
   signal.

5. **Add live retrieval bench to CI.** The synthetic bench scores 0.93.
   The live bench scores 0.13. Only measuring live data catches real
   problems.

6. **Archive uncited learnings after 60 days.** If nothing referenced it
   in 60 days, it's not useful. Move to cold storage.

## What Success Looks Like

The flywheel spins when a learning created in Session N is retrieved and
applied in Session N+2 without human intervention. That requires all
three broken links to be fixed: quality extraction, clean storage, and
mechanical retrieval. The infrastructure exists — the retrieval engine
scores 0.93 on clean data. The bottleneck is corpus quality and the
missing retrieval surface at decision points.
