---
title: M8 C1 Option A — Council Consolidation
date: 2026-04-11
mode: validate --tdd
judges: 3 (2 completed, 1 blocked on write permissions)
verdict: WARN (assumptions wrong, scope was mis-estimated, but fix is feasible)
---

# Consolidated Verdict: WARN — My original assumptions were wrong in two directions

## TL;DR

1. `collectLearnings` **does exist** (at `cli/cmd/ao/inject_learnings.go:50`, not fitness.go) — my grep missed it. It has 8 real callers. **But it is an INJECT-side loader, not a fitness-MEASURE function.** The bead description conflated two unrelated functions.
2. **The bigger discovery (Judge 2, high-confidence):** The staging-tree infrastructure **already exists in the codebase.** REDUCE already writes to `cp.StagingDir`. `NewCheckpoint` already deep-copies subpaths. `cp.Commit()` already does atomic rename. **The M8 fix is not "add staging" — it is "reorder MEASURE before COMMIT" so the pre-commit fitness check actually uses the already-existing staging tree.**
3. **Scope is roughly 1/3 of the original ~395 LOC estimate.** The heavy lifting was already landed. What remains: (a) move `RunMeasure` call site from `loop.go:381` to between `loop.go:330` and `335`, (b) thread `cp.Rollback()` into the regression-halt branch so live tree is never touched on fail, (c) re-plumb **12+ integration tests** that currently assume post-commit MEASURE.
4. **The real risk is the test re-plumbing, not the code change.** The 1-line reorder is trivial; the 12+ tests asserting on current post-commit semantics are where the hours go and where bugs hide.

## Claim-by-Claim Results

| Claim | Original Assumption | Judge Verdict | Corrected Truth |
|---|---|---|---|
| C1: Target is `cli/cmd/ao/fitness.go::collectLearnings` | FAIL | `fitness.go` doesn't exist, but `collectLearnings` lives in `inject_learnings.go:50` with 8 real callers (retrieval_bench.go x3, lookup.go x2, inject.go, codex.go, context_ranked_intel.go). **But it has nothing to do with M8** — it loads learnings for INJECT, not MEASURE. Bead description is wrong. |
| C2: Fitness surface lives in corpus/fitness.go + overnight/fitness.go + overnight/loop.go | PASS | Verified. `Compute(cwd string)` at `corpus/fitness.go:74`. `FitnessSnapshot.Delta` at `overnight/fitness.go:77`. MEASURE→Compute call at `overnight/measure.go:108`. |
| C3: Current loop is Option B (post-commit MEASURE, no rollback on regression) | PASS | Verified. Line 335 = `cp.Commit()`. Line 381 = `RunMeasure`. Regression halt at line 444 = `StatusHaltedOnRegressionPostCommit` with **no rollback action** — live tree stays mutated on fail. |

## Judge 2's Critical Discovery (NEW — was not in my original read)

The **staging infrastructure already exists**. The prior session's ~395 LOC estimate appears to have assumed a net-new staging system, but the code already has:

| Component | Location | Status |
|---|---|---|
| Checkpoint deep-copy of `.agents/` subpaths | `checkpoint.go:194-223` (`NewCheckpoint`) | ✅ Already landed |
| REDUCE writes into `cp.StagingDir` | `reduce.go:186` + 8 sub-stages | ✅ Already landed |
| Atomic commit via rename | `checkpoint.go:280-320` (`Commit`) | ✅ Already landed |
| Rollback primitive | `checkpoint.go` (paired with Commit) | ✅ Already landed |
| Metadata round-trip post-commit check | `checkpoint.go:409-452` | ✅ Already landed |

**What's actually missing for Option A:**
1. `RunMeasure(opts.Cwd)` is called AFTER `cp.Commit()` instead of BEFORE. This is a sequencing bug, not a missing feature.
2. On regression, the halt path returns without calling `cp.Rollback()` (because by that point, commit already swapped live tree).
3. The fitness delta logic in `loop.go:425-430` reads `measure.FitnessSnapshot` — which today is computed post-commit. It needs to read a pre-commit measure result.

## Revised Scope Estimate (Judge 3 slot — inline)

Judge 3 was blocked on write permissions; doing the scope audit myself from the other judges' evidence:

| Change | LOC estimate | Risk |
|---|---|---|
| Move `RunMeasure` call from loop.go:381 → between 330 and 335 | ~5 LOC | LOW (sequencing only) |
| Wire `cp.Rollback()` into regression-halt branch | ~10 LOC | LOW (primitive exists) |
| Update fitness delta logic to use pre-commit measure (`loop.go:425-430`) | ~10 LOC | MEDIUM (may reveal hidden reads from post-commit state) |
| Preserve `VerifyMetadataRoundTripPostCommit` as insurance gate | 0 LOC (keep as-is) | LOW |
| Update `IterationSummary.FitnessAfter` semantics to "what it would be if committed" | ~5 LOC + doc | MEDIUM (behavioral contract change — needs docs/contracts update) |
| Re-plumb 12+ integration tests that assert on post-commit MEASURE semantics | **~80-150 LOC** | **HIGH** — this is where the hours go |
| New L2 tests for Option A semantic: pre-commit regression → live tree unmutated, staging discarded, rollback fires | ~60-100 LOC | MEDIUM |
| Contract doc updates (`docs/contracts/dream-run-contract.md`, `docs/contracts/dream-report.md`) | ~20 LOC | LOW |
| **Total realistic** | **~180-300 LOC** | — |

**The ~395 LOC estimate in the bead was roughly 30-50% inflated.** The real blocker isn't LOC — it's the test re-plumbing, and the fact that we don't yet know which 12+ tests need to change until we actually do the reorder and watch the suite fail.

## TDD Test Matrix (from Judges 1+2, completed by me)

Three L2 assertions from Judge 2 (verbatim — these are the correctness proofs):

```go
// T1: pre-commit regression → live tree unmutated
func TestM8_RegressionPreCommit_LiveTreeUnmutated(t *testing.T) {
    liveHubPath := filepath.Join(opts.Cwd, ".agents", "learnings", "baseline.md")
    preIterMtime := getFileModTime(liveHubPath)
    // run iteration with injected regression
    postIterMtime := getFileModTime(liveHubPath)
    assert(preIterMtime == postIterMtime, "live tree was mutated despite regression")
}

// T2: staging cleanup after regression
func TestM8_RegressionPreCommit_StagingDiscarded(t *testing.T) {
    stagingPath := filepath.Join(opts.Cwd, ".agents", "overnight", "staging", iterID)
    _, err := os.Stat(stagingPath)
    assert(os.IsNotExist(err), "staging dir was not cleaned up after regression")
}

// T3: fitness delta computed pre-commit
func TestM8_FitnessComputed_PreCommit(t *testing.T) {
    // inject external .agents/ change between Measure and old Commit,
    // verify iteration used pre-commit snapshot
}
```

Plus the ones I'm adding for completeness:

```go
// T4: pass-through case — no regression → normal commit path unchanged
func TestM8_NoRegression_CommitProceedsNormally(t *testing.T) {
    // assert: live tree mutated, staging cleaned up, iteration status = committed
}

// T5: crash-safety — process killed between pre-commit measure and commit decision
func TestM8_CrashBetweenMeasureAndCommit_RecoversCleanly(t *testing.T) {
    // inject os.Exit(1) after measure but before commit
    // assert on restart: staging discarded, no half-committed state
}

// T6: integration with M4 warn-only ratchet
func TestM8_RegressionPreCommit_WarnOnlyRescuePreservesBudget(t *testing.T) {
    // regression detected pre-commit + warn-only budget > 0
    // assert: rescue consumed, commit proceeds (warn-only honors the rescue)
    // this is the subtle interaction that could silently break M4
}
```

## Decision Guidance for Bo

**The /council -tdd validation changes the recommendation fundamentally.**

**Original assumption:** 395 LOC new staging infrastructure → needs dedicated full-complexity mixed council session.

**Corrected reality:** ~180-300 LOC behavioral fix using existing staging infrastructure. The "big scary refactor" is actually a sequencing bug with expensive test cleanup.

**Can I do it in this session?** Technically yes, with caveats:
- The code change is small.
- The test re-plumbing is where I'll burn hours and risk silent weakening.
- Doing this without `--mixed` council means I'm taking solo responsibility for preserving ~12 integration tests' intent while rewriting their fixtures. That's the exact "large document consistency" failure mode flagged in your global CLAUDE.md.
- **Better approach given the discovery:** run the **TDD test suite FIRST** (write T1-T6 as RED tests against current Option B code, watch them fail), then do the minimal reorder, watch T1-T6 go GREEN + watch existing tests fail, then re-plumb the failing tests one at a time with the test-failure output driving each fix.

**My recommendation:** Proceed this session using strict TDD, but timebox it. If the existing-test re-plumbing turns out to be more than ~3 tests broken, STOP and defer the remainder to `--mixed` council. The test count is the real risk signal.

**Gate condition to abort mid-session and defer:**
- If >5 existing tests break in non-obvious ways
- If the `cp.Rollback()` primitive turns out to not exist or be more complex than Judge 2 claims
- If the `VerifyMetadataRoundTripPostCommit` interaction reveals a second sequencing dependency
