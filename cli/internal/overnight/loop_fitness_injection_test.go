package overnight

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/overnight/fixture"
)

// Micro-epic 6 (C5) — deterministic RunLoop L2 tests.
//
// These tests exercise the full INGEST → REDUCE → COMMIT → MEASURE pipeline
// on a real fixture (via fixture.GenerateFixture) BUT replace the MEASURE
// fitness computation with a deterministic injector. That lets us drive
// specific plateau/regression/warn-only rescue paths without depending on
// the exact numeric behaviour of corpus.Compute against the fixture corpus.
//
// Determinism: SetTestFitnessInjector is a package-private global with a
// mutex. Every test in this file calls t.Cleanup(SetTestFitnessInjector(nil))
// to restore the nil baseline. None of these tests call t.Parallel().
//
// The handoff specifies three cases:
//  1. Plateau halt under strict mode.
//  2. Regression halt under strict mode.
//  3. Regression tolerated under warn-only mode (with the Micro-epic 4
//     ratchet budget high enough that the loop does not exhaust rescues).

// injectConstantFitness returns an injector that reports a constant
// composite across every iteration. Used to drive the plateau path.
func injectConstantFitness(value float64) func(int) (FitnessSnapshot, error) {
	return func(iterIndex int) (FitnessSnapshot, error) {
		return FitnessSnapshot{
			Metrics: map[string]float64{
				"composite": value,
			},
			CapturedAt: time.Unix(int64(iterIndex), 0).UTC(),
		}, nil
	}
}

// injectRegressionOnSecondIteration reports value=good on iter 1 and
// value=bad on iter 2+, producing a delta that exceeds the default
// RegressionFloor (0.05). Used to drive the strict regression halt path.
func injectRegressionOnSecondIteration(good, bad float64) func(int) (FitnessSnapshot, error) {
	return func(iterIndex int) (FitnessSnapshot, error) {
		value := good
		if iterIndex >= 2 {
			value = bad
		}
		return FitnessSnapshot{
			Metrics: map[string]float64{
				"composite": value,
			},
			CapturedAt: time.Unix(int64(iterIndex), 0).UTC(),
		}, nil
	}
}

// TestRunLoop_PlateauHaltStrict verifies that in strict mode
// (WarnOnly=false) a constant fitness across PlateauWindowK iterations
// triggers a plateau halt with a non-empty PlateauReason and no
// RegressionReason.
func TestRunLoop_PlateauHaltStrict(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	restore := stubInjectRefresh(t)
	defer restore()

	SetTestFitnessInjector(injectConstantFitness(0.5))
	t.Cleanup(func() { SetTestFitnessInjector(nil) })

	dir := t.TempDir()
	if err := fixture.GenerateFixture(dir, fixture.DefaultOpts()); err != nil {
		t.Fatalf("GenerateFixture: %v", err)
	}

	opts := RunLoopOptions{
		Cwd:            dir,
		OutputDir:      filepath.Join(dir, ".agents", "overnight", "plateau-strict"),
		RunID:          "c5-plateau-strict",
		RunTimeout:     30 * time.Second,
		MaxIterations:  10,
		PlateauEpsilon: 0.01,
		PlateauWindowK: 2,
		WarnOnly:       false, // strict mode: plateau halts immediately
		LogWriter:      io.Discard,
	}
	result, err := RunLoop(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunLoop: %v", err)
	}
	if result == nil {
		t.Fatal("result == nil")
	}
	if result.PlateauReason == "" {
		t.Fatalf("expected PlateauReason to be set; got empty. Iterations=%d", len(result.Iterations))
	}
	if result.RegressionReason != "" {
		t.Fatalf("expected empty RegressionReason under plateau path; got %q", result.RegressionReason)
	}
	// Constant composite means iter 1 has no previous snapshot (no delta
	// check), iter 2 sees delta=0 (first plateau observation), iter 3
	// should fire the halt (K=2 consecutive sub-epsilon).
	if len(result.Iterations) < 2 || len(result.Iterations) > 5 {
		t.Fatalf("iteration count=%d not in plausible plateau range [2,5]", len(result.Iterations))
	}
}

// TestRunLoop_RegressionHaltStrict verifies that a drop exceeding the
// regression floor in strict mode halts the loop with
// StatusHaltedOnRegressionPostCommit on the halted iter and a non-empty
// RegressionReason.
func TestRunLoop_RegressionHaltStrict(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	restore := stubInjectRefresh(t)
	defer restore()

	// Drop from 0.9 to 0.1 = 0.8 drop, far exceeding the default 0.05
	// regression floor.
	SetTestFitnessInjector(injectRegressionOnSecondIteration(0.9, 0.1))
	t.Cleanup(func() { SetTestFitnessInjector(nil) })

	dir := t.TempDir()
	if err := fixture.GenerateFixture(dir, fixture.DefaultOpts()); err != nil {
		t.Fatalf("GenerateFixture: %v", err)
	}

	opts := RunLoopOptions{
		Cwd:             dir,
		OutputDir:       filepath.Join(dir, ".agents", "overnight", "regression-strict"),
		RunID:           "c5-regression-strict",
		RunTimeout:      30 * time.Second,
		MaxIterations:   5,
		PlateauEpsilon:  0.01,
		PlateauWindowK:  2,
		RegressionFloor: 0.05,
		WarnOnly:        false, // strict mode: regression halts immediately
		LogWriter:       io.Discard,
	}
	result, err := RunLoop(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunLoop: %v", err)
	}
	if result == nil {
		t.Fatal("result == nil")
	}
	if result.RegressionReason == "" {
		t.Fatalf("expected RegressionReason to be set; got empty")
	}
	if len(result.Iterations) < 2 {
		t.Fatalf("expected at least 2 iterations (iter 1 + iter 2 regression halt), got %d",
			len(result.Iterations))
	}
	// The halted iter should be the regression-halted one.
	lastIter := result.Iterations[len(result.Iterations)-1]
	if lastIter.Status != StatusHaltedOnRegressionPostCommit {
		t.Fatalf("last iter status=%q want %q", lastIter.Status, StatusHaltedOnRegressionPostCommit)
	}
}

// TestRunLoop_RegressionIgnoredWarnOnly verifies that the same
// regression-on-iter-2 injector does NOT halt under warn-only mode
// (with sufficient budget). Instead the loop continues to iter 3+ and
// the halted-on-regression status does not appear in any persisted
// iteration.
func TestRunLoop_RegressionIgnoredWarnOnly(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	restore := stubInjectRefresh(t)
	defer restore()

	SetTestFitnessInjector(injectRegressionOnSecondIteration(0.9, 0.1))
	t.Cleanup(func() { SetTestFitnessInjector(nil) })

	dir := t.TempDir()
	if err := fixture.GenerateFixture(dir, fixture.DefaultOpts()); err != nil {
		t.Fatalf("GenerateFixture: %v", err)
	}

	opts := RunLoopOptions{
		Cwd:             dir,
		OutputDir:       filepath.Join(dir, ".agents", "overnight", "regression-warn"),
		RunID:           "c5-regression-warn",
		RunTimeout:      30 * time.Second,
		MaxIterations:   3, // bounded low to avoid running forever
		PlateauEpsilon:  0.01,
		PlateauWindowK:  2,
		RegressionFloor: 0.05,
		WarnOnly:        true, // warn-only: regression becomes degraded note
		LogWriter:       io.Discard,
		// No WarnOnlyBudget set → nil → legacy infinite-rescue path from
		// Micro-epic 4. This keeps the test focused on the "warn-only
		// ignores regression" semantic.
	}
	result, err := RunLoop(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunLoop: %v", err)
	}
	if result == nil {
		t.Fatal("result == nil")
	}
	// Regression should NOT have fired a halt.
	if result.RegressionReason != "" {
		t.Fatalf("expected empty RegressionReason in warn-only mode; got %q",
			result.RegressionReason)
	}
	// No iteration should carry StatusHaltedOnRegressionPostCommit.
	for _, iter := range result.Iterations {
		if iter.Status == StatusHaltedOnRegressionPostCommit {
			t.Fatalf("iter %d incorrectly halted under warn-only: status=%q",
				iter.Index, iter.Status)
		}
	}
	// Some iterations should have the warn-only degraded annotation.
	// This is tight coupling to the message format but is worth it to
	// prove the code path executed.
	foundWarnOnlyNote := false
	for _, iter := range result.Iterations {
		for _, note := range iter.Degraded {
			if strings.Contains(note, "warn-only") {
				foundWarnOnlyNote = true
				break
			}
		}
		if foundWarnOnlyNote {
			break
		}
	}
	if !foundWarnOnlyNote {
		t.Fatal("expected at least one iteration to carry a warn-only degraded note")
	}
}

// injectOscillatingFitness alternates between good and bad values on every
// iteration. Used to drive repeated regression events — each "bad" iter
// produces a delta = bad - good (negative, exceeding the default floor),
// firing the regression path on every other iteration. Plateau is NOT
// tripped because the deltas are large, so plateau.halted stays false and
// the regression path is the only rescue consumer.
func injectOscillatingFitness(good, bad float64) func(int) (FitnessSnapshot, error) {
	return func(iterIndex int) (FitnessSnapshot, error) {
		value := good
		if iterIndex%2 == 0 {
			value = bad
		}
		return FitnessSnapshot{
			Metrics: map[string]float64{
				"composite": value,
			},
			CapturedAt: time.Unix(int64(iterIndex), 0).UTC(),
		}, nil
	}
}

// TestRunLoop_WarnOnlyBudgetExhausted (M6 follow-up, per na-jn9) verifies
// that the warn-only rescue budget drains correctly across repeated
// regression events and that rescue N+1 falls through to a strict halt with
// WarnOnlyBudgetRemaining == 0.
//
// Setup: oscillating composite fitness (good↔bad per iter) + WarnOnly=true
// + WarnOnlyBudget initial=2. Each "bad" iter exceeds RegressionFloor and
// consumes exactly one rescue. After N rescues are consumed, effectiveWarnOnly
// flips to false inside loop.go (Remaining<=0), and the next regression event
// halts strictly with a non-empty RegressionReason.
//
// Plateau path is explicitly NOT exercised here because plateau.halted is
// sticky (one-shot per run), so it can only consume a single rescue. The
// regression path is stateless per-iteration and can consume one rescue per
// event — the shape na-jn9 asks for.
func TestRunLoop_WarnOnlyBudgetExhausted(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	restore := stubInjectRefresh(t)
	defer restore()

	// good=0.9, bad=0.1 → |delta|=0.8, well above the default 0.05 floor.
	SetTestFitnessInjector(injectOscillatingFitness(0.9, 0.1))
	t.Cleanup(func() { SetTestFitnessInjector(nil) })

	dir := t.TempDir()
	if err := fixture.GenerateFixture(dir, fixture.DefaultOpts()); err != nil {
		t.Fatalf("GenerateFixture: %v", err)
	}

	const initialBudget = 2
	opts := RunLoopOptions{
		Cwd:             dir,
		OutputDir:       filepath.Join(dir, ".agents", "overnight", "warn-only-exhausted"),
		RunID:           "m6-warn-only-exhausted",
		RunTimeout:      30 * time.Second,
		MaxIterations:   20, // generous ceiling; strict halt should fire well before this
		PlateauEpsilon:  0.01,
		PlateauWindowK:  2,
		RegressionFloor: 0.05,
		WarnOnly:        true,
		WarnOnlyBudget: &WarnOnlyRatchet{
			Initial:   initialBudget,
			Remaining: initialBudget,
		},
		LogWriter: io.Discard,
	}
	result, err := RunLoop(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunLoop: %v", err)
	}
	if result == nil {
		t.Fatal("result == nil")
	}

	// Budget should be fully drained — effectiveWarnOnly flips only once
	// Remaining hits 0.
	if result.WarnOnlyBudgetRemaining != 0 {
		t.Fatalf("WarnOnlyBudgetRemaining = %d, want 0 (budget should be exhausted)",
			result.WarnOnlyBudgetRemaining)
	}
	if result.WarnOnlyBudgetInitial != initialBudget {
		t.Fatalf("WarnOnlyBudgetInitial = %d, want %d",
			result.WarnOnlyBudgetInitial, initialBudget)
	}

	// A strict regression halt should have fired once rescues ran out.
	// RegressionReason on the top-level result is only set when the loop
	// itself halts on regression (rescues do not set it).
	if result.RegressionReason == "" {
		t.Fatalf("expected non-empty RegressionReason after rescue budget exhausted; got empty")
	}
	// Exhaustion suffix is added by loop.go when the strict halt fires
	// under a drained budget — see loop.go:466-468.
	if !strings.Contains(result.RegressionReason, "warn-only budget exhausted") {
		t.Fatalf("RegressionReason = %q; expected to contain %q",
			result.RegressionReason, "warn-only budget exhausted")
	}

	// Sanity: loop must not run to MaxIterations — strict halt should
	// fire on rescue N+1 (conservatively well within 2*initialBudget+2).
	if len(result.Iterations) >= opts.MaxIterations {
		t.Fatalf("loop ran %d iterations, expected strict halt well before MaxIterations=%d",
			len(result.Iterations), opts.MaxIterations)
	}

	// Last iteration must be the strict regression halt, not a rescue.
	lastIter := result.Iterations[len(result.Iterations)-1]
	if lastIter.Status != StatusHaltedOnRegressionPostCommit {
		t.Fatalf("last iter status = %q, want %q",
			lastIter.Status, StatusHaltedOnRegressionPostCommit)
	}

	// Count warn-only degraded notes across iterations. We expect
	// exactly `initialBudget` rescue notes — the regression path consumes
	// one rescue per bad iter, and the strict halt iter does not emit a
	// warn-only note (it takes the halt branch).
	warnOnlyNoteCount := 0
	for _, iter := range result.Iterations {
		for _, note := range iter.Degraded {
			if strings.Contains(note, "warn-only") {
				warnOnlyNoteCount++
				break
			}
		}
	}
	if warnOnlyNoteCount < initialBudget {
		t.Fatalf("warn-only rescue notes = %d, want at least %d (the initial budget) before exhaustion",
			warnOnlyNoteCount, initialBudget)
	}
}
