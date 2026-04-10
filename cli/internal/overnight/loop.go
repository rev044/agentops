package overnight

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

// IterationSummary is one pass of the INGEST → REDUCE → MEASURE wave.
//
// Persisted into overnightSummary.Iterations (dream-report schema v2) for
// downstream inspection by the morning report renderer and council judges.
type IterationSummary struct {
	ID           IterationID `json:"id" yaml:"id"`
	Index        int         `json:"index" yaml:"index"` // 1-based
	StartedAt    time.Time   `json:"started_at" yaml:"started_at"`
	FinishedAt   time.Time   `json:"finished_at" yaml:"finished_at"`
	Duration     string      `json:"duration" yaml:"duration"`
	Status       string      `json:"status" yaml:"status"` // done, degraded, rolled-back, failed

	// Stage sub-summaries. Each is an opaque map so v1 readers can ignore
	// them without schema awareness.
	Ingest  map[string]any `json:"ingest,omitempty" yaml:"ingest,omitempty"`
	Reduce  map[string]any `json:"reduce,omitempty" yaml:"reduce,omitempty"`
	Measure map[string]any `json:"measure,omitempty" yaml:"measure,omitempty"`

	// FitnessBefore and FitnessAfter are opaque maps for v1 compatibility.
	// The real FitnessVector lives in cli/internal/corpus/fitness.go
	// (introduced by Issue 15) and is marshaled into these maps by MEASURE.
	FitnessBefore map[string]any `json:"fitness_before,omitempty" yaml:"fitness_before,omitempty"`
	FitnessAfter  map[string]any `json:"fitness_after,omitempty" yaml:"fitness_after,omitempty"`
	FitnessDelta  float64        `json:"fitness_delta" yaml:"fitness_delta"`

	Degraded []string `json:"degraded,omitempty" yaml:"degraded,omitempty"`
	Error    string   `json:"error,omitempty" yaml:"error,omitempty"`
}

// RunLoopResult is the aggregate output of a single Dream run.
//
// The command-layer caller copies these fields into overnightSummary's new
// v2 fields (Iterations, FitnessDelta, PlateauReason, RegressionReason)
// before finalizing the morning report.
type RunLoopResult struct {
	Iterations       []IterationSummary `json:"iterations" yaml:"iterations"`
	FitnessDelta     map[string]any     `json:"fitness_delta,omitempty" yaml:"fitness_delta,omitempty"`
	PlateauReason    string             `json:"plateau_reason,omitempty" yaml:"plateau_reason,omitempty"`
	RegressionReason string             `json:"regression_reason,omitempty" yaml:"regression_reason,omitempty"`
	BudgetExhausted  bool               `json:"budget_exhausted" yaml:"budget_exhausted"`
	Degraded         []string           `json:"degraded,omitempty" yaml:"degraded,omitempty"`
}

// ErrNotImplemented is returned by stage stubs that will be filled in by
// later waves (Wave 3 for stage drivers, Wave 4 for integration). Callers
// in tests distinguish this from a real error so skeleton compiles green
// without faking behavior.
var ErrNotImplemented = errors.New("overnight: stage not implemented yet (skeleton wave)")

// RunLoop executes the bounded Dream compounding loop.
//
// Wave 1 (this file) ships the public surface only: normalized options,
// a log-writer shim, and a deterministic "no iterations ran, budget 0"
// degraded-result that lets downstream wiring compile. Real iteration
// logic lands in Wave 3 via RunIngest / RunReduce / RunMeasure.
//
// Returning (non-nil, non-nil) is explicitly allowed: a partial result
// with a non-nil error describes "we ran N iterations, then hit error E."
// The caller should persist the partial result AND surface the error.
func RunLoop(ctx context.Context, opts RunLoopOptions) (*RunLoopResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	opts, degraded := opts.normalize()
	log := opts.LogWriter
	if log == nil {
		log = io.Discard
	}

	result := &RunLoopResult{
		Iterations: nil,
		Degraded:   degraded,
	}

	if opts.Cwd == "" {
		return result, fmt.Errorf("overnight: RunLoopOptions.Cwd is required")
	}
	if opts.OutputDir == "" {
		return result, fmt.Errorf("overnight: RunLoopOptions.OutputDir is required")
	}

	fmt.Fprintf(log, "overnight: RunLoop starting (budget=%s, max_iter=%d, epsilon=%g, K=%d, warn_only=%v)\n",
		opts.RunTimeout, opts.MaxIterations, opts.PlateauEpsilon, opts.PlateauWindowK, opts.WarnOnly)

	// Wave 3 replaces this early-return with the real
	// INGEST → REDUCE → MEASURE driver loop. For Wave 1 the package
	// compiles and callers observe a degraded no-op result so the overall
	// shape flows through the command layer without surprises.
	result.Degraded = append(result.Degraded, "RunLoop skeleton: no iterations executed (Wave 1 ships public surface only)")

	// Respect context cancellation even in the stub path — callers rely
	// on this contract in tests.
	select {
	case <-ctx.Done():
		return result, ctx.Err()
	default:
	}

	fmt.Fprintln(log, "overnight: RunLoop skeleton complete (no iterations executed)")
	return result, nil
}
