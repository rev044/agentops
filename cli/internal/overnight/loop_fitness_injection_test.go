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
