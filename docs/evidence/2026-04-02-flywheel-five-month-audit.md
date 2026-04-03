# Five-Month Flywheel Audit: Where Knowledge Compounding Breaks Down

> Empirical audit date: 2026-04-02
> Data range: November 2025 — April 2026
> Rigs audited: 14 active crew workspaces across Gas Town

## Executive Summary

After 5 months operating the AgentOps knowledge flywheel across 14 agent
workspaces, the system has accumulated 2,275 learning artifacts. Despite
this volume, the flywheel does not compound. Knowledge is written but not
read, harvested but not curated, and scored but not trusted. This document
presents the empirical evidence, diagnoses five structural failure modes,
and proposes the minimum changes needed to make the flywheel spin.

## Raw Inventory

| Metric | Value |
|--------|-------|
| Total learning files (all rigs) | 2,275 |
| Global store files (harvested) | 1,612 |
| Active rigs producing learnings | 14 |
| Date range | Nov 2025 — Apr 2026 |
| Top producers | neo (189), ichigo (163), nami (139), kenpachi (104) |
| Global curated (root-level, hand-written) | 24 files |
| Global patterns | 6 categories |
| Council reports (nami) | 1 |
| Learnings cited by any other artifact | 0 (0%) |

## Per-Rig Metadata Quality

| Rig | Learnings | Has Maturity | Has Utility | Auto-Generated |
|-----|-----------|-------------|-------------|----------------|
| nami | 28 | 100% | 0% | 100% |
| ichigo | 162 | 2% | 54% | 1% |
| neo | 187 | 61% | 33% | 56% |
| kenpachi | 104 | 97% | 3% | 96% |
| Global store | 1,612 | varies | varies | ~70% est. |

**Key finding:** No two rigs tag learnings the same way. The scoring
pipeline needs both maturity and utility to rank effectively. Across the
fleet, the overlap of "has both" is consistently low.

## Five Structural Failure Modes

### Failure 1: Write-Only Knowledge (0% Citation Rate)

Of 28 learnings in nami's local store, **zero** are referenced by any
other artifact — no plan, no council report, no implementation notes, no
handoff document. Knowledge flows one direction: from session into file.
It never flows back into a decision.

**Root cause:** No mechanical surface forces retrieval at decision points.
The `ao lookup` command exists but nothing calls it during `/plan`,
`/pre-mortem`, or `/implement`. Injection was disabled in ag-8km (manual
startup mode) and never replaced with decision-point retrieval.

**Evidence:** `grep -rl <filename> .agents/ --exclude-dir=learnings`
returns 0 matches for all 28 nami learnings.

### Failure 2: Auto-Extraction Produces Garbage

100% of nami's 28 learnings are auto-extracted. Sample content:

> "didn't work because the goals agent is still running and re-wrote it.
> Let me force the revert:"

This is a conversation fragment, not a learning. It has no title (filename
is a truncated sentence), no utility score, no source bead, and maturity
permanently stuck at "provisional."

**Root cause:** The forge/auto-extract pipeline has no minimum quality
threshold. Any text chunk with a vaguely learning-shaped structure gets
persisted. No human-in-the-loop curation step, no automated quality
filter checking semantic coherence.

**Contrast:** The 24 manually curated global learnings (e.g., "Mine Before
You Manufacture") have structured sections, clear takeaways, source
attribution, and high confidence ratings. These are genuinely useful.

### Failure 3: Inconsistent Metadata Across Rigs

- **nami:** 100% maturity, 0% utility (maturity-only)
- **ichigo:** 2% maturity, 54% utility (utility-only)
- **kenpachi:** 97% maturity, 3% utility (maturity-only)
- **neo:** 61% maturity, 33% utility (mixed)

The scoring pipeline (`inject_scoring.go`) uses both maturity and utility
for composite ranking: `(z_norm(freshness) + lambda * z_norm(utility)) * maturityWeight`.
When one is missing, scoring degrades. Learnings without utility default
to 0.5, which is generous for auto-extracted garbage.

**Root cause:** Each rig's extraction pipeline evolved independently.
No shared schema enforcement at write time. Harvest copies but does not
validate or normalize.

### Failure 4: Harvesting Without Curating (1,612 Files, No Filter)

The global store at `~/.agents/learnings/` contains 1,612 files across
9 subdirectories:

| Subdirectory | Files |
|-------------|-------|
| research/ | 773 |
| learning/ | 671 |
| jren-platform/ | 54 |
| platform/ | 40 |
| pattern/ | 32 |
| infra/ | 8 |
| devops/ | 5 |
| operations/ | 4 |
| tooling/ | 12 |
| Root (curated) | 24 |

Low-quality auto-extracted fragments get promoted alongside curated work.
The `harvest_confidence: 0.6` threshold was too permissive (raised to 0.5
in 3945832f, but old files remain). No deduplication — the same garbage
fragment from nami appears in both local and global stores.

### Failure 5: Retrieval Returns Wrong Results on Live Data

The `ao badge` shows retrieval sigma = 0.13 (on 0-1 scale). The synthetic
retrieval bench scores P@3 = 0.93, but on live data, `ao lookup --query
"hook authoring"` returns ACM cluster factory research and Medium post
preparation. Completely unrelated.

**Root cause:** 1,612 global files with inconsistent metadata dilute the
signal. The scoring pipeline works (proven by synthetic bench), but the
corpus quality destroys it.

## What Actually Works

Three things produce value despite the failures:

1. **Manual curation.** The 24 root-level global learnings are high
   quality. "Mine Before You Manufacture," "Enforcement Pyramid Pattern,"
   "Version Pinning is a Timebomb" — genuinely reusable knowledge.

2. **Single-session flywheel loops.** The ag-470 case study proved that
   within a single RPI session, knowledge compounds: research feeds plan,
   pre-mortem catches 80% of bugs, post-mortem extracts learnings. The
   intra-session flywheel works. The cross-session flywheel does not.

3. **The retrieval engine itself.** After the token-AND filter fix
   (ag-9t0), synthetic P@3 = 0.93. The engine works; the corpus is the
   bottleneck.

## Root Cause: The Chain Breaks in Three Places

```
Generate --> Extract --> Store --> Retrieve --> Apply --> Generate
                ^           ^                    ^
                |           |                    |
           Failure 2    Failures 3,4         Failure 1
           (garbage)    (noise drowns       (nothing reads
                         signal)             knowledge back)
```

The flywheel generates knowledge every session. But extraction produces
noise (Failure 2), storage mixes noise with signal (Failures 3, 4), and
nothing forces retrieval at decision points (Failure 1). The result:
2,275 files that nobody reads.

## Minimum Viable Fixes (Ordered by Impact)

| Priority | Fix | Effort | Addresses |
|----------|-----|--------|-----------|
| 1 | Wire `ao lookup` into `/plan` and `/pre-mortem` | M | Failure 1 (0% citation) |
| 2 | Add quality gate to auto-extraction (min 50 chars, heading, coherent) | S | Failure 2 (garbage) |
| 3 | Enforce maturity+utility at write time | S | Failure 3 (metadata) |
| 4 | Purge global store of low-quality fragments | M | Failure 4 (noise) |
| 5 | Add live-data retrieval bench to CI | S | Failure 5 (retrieval) |
| 6 | Archive learnings with 0 citations after 60 days | M | Failures 1, 4 |

## Success Metrics: "The Flywheel Spins"

| # | Metric | Target | Current |
|---|--------|--------|---------|
| 1 | Citation rate | > 10% | 0% |
| 2 | Retrieval sigma (live) | > 0.5 | 0.13 |
| 3 | Auto-extract rejection rate | > 50% | 0% |
| 4 | Time-to-first-retrieval | < 2 sessions | never |
| 5 | Global growth < local creation | yes | no (1,612 uncurated) |

Current state: 0/5 met.

## The Thesis

The flywheel infrastructure works. The retrieval engine scores 0.93 on
clean data. The pre-mortem catches 80% of bugs. The intra-session loop
compounds knowledge reliably. What's missing is the connective tissue
between sessions: quality gates at extraction, schema enforcement at
storage, and mechanical retrieval at decision points. These are small
changes — but without them, 5 months of accumulated knowledge is a
write-only log that no agent ever reads.
