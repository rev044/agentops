package overnight

import (
	"path/filepath"
	"testing"
)

// TestWithMeasureFailureCap_SetsBothFieldsAtomically is the Micro-epic 5
// (C4) guard that the builder sets both MaxConsecutiveMeasureFailures
// and the private explicitMeasureFailureCap sentinel together. If a
// future refactor drops the builder and exposes only the int field,
// this test surfaces the regression.
func TestWithMeasureFailureCap_SetsBothFieldsAtomically(t *testing.T) {
	cases := []struct {
		name    string
		input   int
		wantCap int
	}{
		{"explicit zero (halt on first failure)", 0, 0},
		{"positive (halt after N)", 5, 5},
		{"unbounded sentinel", -1, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := RunLoopOptions{}.WithMeasureFailureCap(tc.input)
			if opts.MaxConsecutiveMeasureFailures != tc.wantCap {
				t.Fatalf("MaxConsecutiveMeasureFailures=%d want %d",
					opts.MaxConsecutiveMeasureFailures, tc.wantCap)
			}
			if !opts.explicitMeasureFailureCap {
				t.Fatal("explicitMeasureFailureCap should be true after builder")
			}
		})
	}
}

// TestNormalize_AppliesMeasureFailureCapDefault verifies that an
// uninitialized RunLoopOptions normalizes to
// defaultMaxConsecutiveMeasureFailures, while a WithMeasureFailureCap
// caller is preserved. This is the sentinel disambiguation the field
// doc spec demands.
func TestNormalize_AppliesMeasureFailureCapDefault(t *testing.T) {
	// Case 1: untouched → default.
	untouched, _ := RunLoopOptions{}.normalize()
	if untouched.MaxConsecutiveMeasureFailures != defaultMaxConsecutiveMeasureFailures {
		t.Fatalf("untouched cap=%d want %d",
			untouched.MaxConsecutiveMeasureFailures, defaultMaxConsecutiveMeasureFailures)
	}

	// Case 2: caller explicitly set 0 → preserved (not defaulted).
	explicitZero, _ := RunLoopOptions{}.WithMeasureFailureCap(0).normalize()
	if explicitZero.MaxConsecutiveMeasureFailures != 0 {
		t.Fatalf("explicit-zero cap=%d want 0 (preserved)",
			explicitZero.MaxConsecutiveMeasureFailures)
	}

	// Case 3: caller explicitly set -1 → preserved (unbounded).
	unbounded, _ := RunLoopOptions{}.WithMeasureFailureCap(-1).normalize()
	if unbounded.MaxConsecutiveMeasureFailures != -1 {
		t.Fatalf("unbounded cap=%d want -1", unbounded.MaxConsecutiveMeasureFailures)
	}

	// Case 4: caller explicitly set positive → preserved.
	explicitFive, _ := RunLoopOptions{}.WithMeasureFailureCap(5).normalize()
	if explicitFive.MaxConsecutiveMeasureFailures != 5 {
		t.Fatalf("explicit-five cap=%d want 5", explicitFive.MaxConsecutiveMeasureFailures)
	}
}

// TestLoadLatestUnflaggedIteration_WalksDegradedToStatusDone covers the
// rehydration helper for Micro-epic 5. A mixed history of done/degraded/
// halted iterations should produce the most recent StatusDone entry,
// skipping degraded and halted iterations that do not carry a valid
// FitnessAfter snapshot for delta computation.
func TestLoadLatestUnflaggedIteration_WalksDegradedToStatusDone(t *testing.T) {
	dir := t.TempDir()
	runID := "test-run-c4"

	// iter-1: Done, the valid baseline
	writeTestIteration(t, dir, runID, IterationSummary{
		ID: IterationID(runID + "-iter-1"), Index: 1, Status: StatusDone,
		FitnessAfter: map[string]any{"composite": 0.42},
	})
	// iter-2: Degraded (MEASURE failed)
	writeTestIteration(t, dir, runID, IterationSummary{
		ID: IterationID(runID + "-iter-2"), Index: 2, Status: StatusDegraded,
	})
	// iter-3: HaltedOnRegressionPostCommit
	writeTestIteration(t, dir, runID, IterationSummary{
		ID: IterationID(runID + "-iter-3"), Index: 3, Status: StatusHaltedOnRegressionPostCommit,
	})

	got, err := LoadLatestUnflaggedIteration(dir, runID)
	if err != nil {
		t.Fatalf("LoadLatestUnflaggedIteration: %v", err)
	}
	if got == nil {
		t.Fatal("got nil, expected iter-1 as the latest unflagged")
	}
	if got.Index != 1 {
		t.Fatalf("got index=%d want 1", got.Index)
	}
	if got.Status != StatusDone {
		t.Fatalf("got status=%q want %q", got.Status, StatusDone)
	}
	if _, ok := got.FitnessAfter["composite"]; !ok {
		t.Fatal("unflagged iteration should carry FitnessAfter snapshot")
	}
}

// TestLoadLatestUnflaggedIteration_NoUnflaggedReturnsNil verifies the
// nil-and-no-error return contract when every iteration is degraded or
// halted.
func TestLoadLatestUnflaggedIteration_NoUnflaggedReturnsNil(t *testing.T) {
	dir := t.TempDir()
	runID := "test-run-c4-alldegraded"
	writeTestIteration(t, dir, runID, IterationSummary{
		ID: IterationID(runID + "-iter-1"), Index: 1, Status: StatusDegraded,
	})
	writeTestIteration(t, dir, runID, IterationSummary{
		ID: IterationID(runID + "-iter-2"), Index: 2, Status: StatusDegraded,
	})

	got, err := LoadLatestUnflaggedIteration(dir, runID)
	if err != nil {
		t.Fatalf("err=%v want nil", err)
	}
	if got != nil {
		t.Fatalf("got=%+v want nil when no unflagged exists", got)
	}
}

// TestLoadLatestUnflaggedIteration_MissingDir verifies the
// "fresh-run" path returns (nil, nil).
func TestLoadLatestUnflaggedIteration_MissingDir(t *testing.T) {
	got, err := LoadLatestUnflaggedIteration(filepath.Join(t.TempDir(), "nope"), "x")
	if err != nil {
		t.Fatalf("err=%v want nil (missing dir is not an error)", err)
	}
	if got != nil {
		t.Fatal("got non-nil from missing dir")
	}
}

// writeTestIteration is a minimal test helper that persists an
// IterationSummary via the existing atomic writer. Shared by the
// LoadLatestUnflaggedIteration tests above.
func writeTestIteration(t *testing.T, dir, runID string, iter IterationSummary) {
	t.Helper()
	if err := writeIterationAtomic(dir, iter); err != nil {
		t.Fatalf("writeIterationAtomic iter-%d: %v", iter.Index, err)
	}
}
