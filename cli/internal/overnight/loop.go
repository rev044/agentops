package overnight

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/boshu2/agentops/cli/internal/lifecycle"
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
// Each iteration: INGEST → NewCheckpoint → REDUCE → MEASURE → delta.
// Halt conditions (first to fire wins):
//
//   - Budget exhausted: ctx deadline expired or cumulative elapsed
//     exceeds opts.RunTimeout.
//   - Max iterations: opts.MaxIterations > 0 and reached.
//   - Plateau: opts.PlateauWindowK consecutive |delta| < opts.PlateauEpsilon.
//   - Regression: any metric dropped by more than opts.RegressionFloor.
//
// WarnOnly mode (the default for the first 2-3 production runs) turns
// plateau and regression halts into degraded notes instead of stopping,
// so operators can calibrate thresholds empirically before ratcheting
// to strict mode.
//
// REDUCE runs with nil close-loop callbacks at this layer — Dream's
// command-layer integration (cli/cmd/ao/overnight.go) is responsible
// for wiring cmd-package helpers into the loop if/when it wants
// in-process close-loop promotion. Until then, close-loop is a
// degraded-note stage inside REDUCE.
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

	// Derive a ctx that also honors the run-timeout wall-clock budget.
	loopCtx, cancel := context.WithTimeout(ctx, opts.RunTimeout)
	defer cancel()

	startedAt := time.Now()
	plateau := NewPlateauState(opts.PlateauWindowK, opts.PlateauEpsilon)
	var prevSnapshot *FitnessSnapshot
	iterIndex := 0

	for {
		// Budget / cancellation check BEFORE starting a new iteration.
		select {
		case <-loopCtx.Done():
			if errors.Is(loopCtx.Err(), context.DeadlineExceeded) {
				result.BudgetExhausted = true
				fmt.Fprintln(log, "overnight: RunLoop halted — wall-clock budget exhausted")
			} else {
				fmt.Fprintf(log, "overnight: RunLoop halted — context cancelled: %v\n", loopCtx.Err())
			}
			return result, nil
		default:
		}

		if opts.MaxIterations > 0 && iterIndex >= opts.MaxIterations {
			fmt.Fprintf(log, "overnight: RunLoop halted — max iterations (%d) reached\n", opts.MaxIterations)
			return result, nil
		}

		iterIndex++
		iterID := IterationID("iter-" + strconv.Itoa(iterIndex))
		iterStart := time.Now()
		fmt.Fprintf(log, "overnight: iteration %d starting (id=%s)\n", iterIndex, iterID)

		iter := IterationSummary{
			ID:        iterID,
			Index:     iterIndex,
			StartedAt: iterStart,
			Status:    "done",
		}

		// --- INGEST ---
		ingest, ingestErr := RunIngest(loopCtx, opts, log)
		if ingestErr != nil {
			iter.Status = "failed"
			iter.Error = fmt.Sprintf("ingest: %v", ingestErr)
			iter.FinishedAt = time.Now()
			iter.Duration = iter.FinishedAt.Sub(iterStart).String()
			result.Iterations = append(result.Iterations, iter)
			fmt.Fprintf(log, "overnight: iteration %d INGEST failed: %v\n", iterIndex, ingestErr)
			return result, fmt.Errorf("overnight: iteration %d ingest: %w", iterIndex, ingestErr)
		}
		if ingest != nil {
			iter.Ingest = ingestSummary(ingest)
			iter.Degraded = append(iter.Degraded, ingest.Degraded...)
		}

		// --- CHECKPOINT ---
		cp, cpErr := NewCheckpoint(opts.Cwd, string(iterID), opts.CheckpointMaxBytes)
		if cpErr != nil {
			iter.Status = "failed"
			iter.Error = fmt.Sprintf("checkpoint: %v", cpErr)
			iter.FinishedAt = time.Now()
			iter.Duration = iter.FinishedAt.Sub(iterStart).String()
			result.Iterations = append(result.Iterations, iter)
			fmt.Fprintf(log, "overnight: iteration %d checkpoint failed: %v\n", iterIndex, cpErr)
			return result, fmt.Errorf("overnight: iteration %d checkpoint: %w", iterIndex, cpErr)
		}

		// --- REDUCE ---
		// Close-loop callbacks are intentionally nil at this layer.
		// Downstream wiring (cli/cmd/ao/overnight.go) can provide its
		// own RunLoop wrapper if it wants in-process close-loop.
		var emptyCloseLoop lifecycle.CloseLoopOpts
		reduce, reduceErr := RunReduce(loopCtx, opts, ingest, cp, emptyCloseLoop, log)
		if reduceErr != nil {
			iter.Status = "rolled-back"
			iter.Error = fmt.Sprintf("reduce: %v", reduceErr)
			if reduce != nil {
				iter.Reduce = reduceSummary(reduce)
			}
			iter.FinishedAt = time.Now()
			iter.Duration = iter.FinishedAt.Sub(iterStart).String()
			result.Iterations = append(result.Iterations, iter)
			// Checkpoint is already rolled back by RunReduce itself.
			fmt.Fprintf(log, "overnight: iteration %d REDUCE failed (rolled back): %v\n", iterIndex, reduceErr)
			return result, fmt.Errorf("overnight: iteration %d reduce: %w", iterIndex, reduceErr)
		}
		if reduce != nil {
			iter.Reduce = reduceSummary(reduce)
			iter.Degraded = append(iter.Degraded, reduce.Degraded...)
		}

		// --- COMMIT ---
		// REDUCE succeeded. Promote the staging tree into live so the
		// next iteration (and MEASURE below) sees the compounded corpus.
		if commitErr := cp.Commit(); commitErr != nil {
			iter.Status = "failed"
			iter.Error = fmt.Sprintf("commit: %v", commitErr)
			iter.FinishedAt = time.Now()
			iter.Duration = iter.FinishedAt.Sub(iterStart).String()
			result.Iterations = append(result.Iterations, iter)
			fmt.Fprintf(log, "overnight: iteration %d commit failed: %v\n", iterIndex, commitErr)
			return result, fmt.Errorf("overnight: iteration %d commit: %w", iterIndex, commitErr)
		}

		// Post-commit metadata integrity check (ratchet-forward per pm-V7).
		// Cannot unwind a successful commit; record a findings entry and
		// surface via degraded so the morning report shows the strip.
		if postReport := VerifyMetadataRoundTripPostCommit(cp); !postReport.Pass {
			msg := fmt.Sprintf("post-commit metadata integrity: %d stripped field(s)", len(postReport.StrippedFields))
			iter.Degraded = append(iter.Degraded, msg)
			fmt.Fprintf(log, "overnight: iteration %d %s\n", iterIndex, msg)
			// Log a structured finding the router will intentionally skip
			// (filename bypasses findingFilenameRe); surfaces in morning report.
			_ = logPostCommitFinding(opts.Cwd, iterIndex, postReport)
		}

		// --- MEASURE ---
		measure, measureErr := RunMeasure(loopCtx, opts, log)
		if measureErr != nil {
			iter.Status = "degraded"
			iter.Error = fmt.Sprintf("measure: %v", measureErr)
			iter.FinishedAt = time.Now()
			iter.Duration = iter.FinishedAt.Sub(iterStart).String()
			result.Iterations = append(result.Iterations, iter)
			// Measure failure is not a rollback trigger (post-commit);
			// the compounded corpus is already committed. Surface the
			// error but continue the loop.
			fmt.Fprintf(log, "overnight: iteration %d MEASURE failed (continuing): %v\n", iterIndex, measureErr)
			continue
		}
		if measure != nil {
			iter.Measure = measureSummary(measure)
			iter.Degraded = append(iter.Degraded, measure.Degraded...)
		}

		// --- DELTA + HALT CHECK ---
		currSnapshot := measure.FitnessSnapshot
		iter.FitnessAfter = snapshotToMap(currSnapshot)
		if prevSnapshot != nil {
			iter.FitnessBefore = snapshotToMap(*prevSnapshot)
			composite, regressions, regressed := currSnapshot.Delta(prevSnapshot, nil, opts.RegressionFloor)
			iter.FitnessDelta = composite
			if regressed && !opts.WarnOnly {
				iter.Status = "rolled-back"
				iter.FinishedAt = time.Now()
				iter.Duration = iter.FinishedAt.Sub(iterStart).String()
				result.Iterations = append(result.Iterations, iter)
				result.RegressionReason = fmt.Sprintf("iteration %d: %d metric(s) breached regression floor %g: %v",
					iterIndex, len(regressions), opts.RegressionFloor, regressionNames(regressions))
				fmt.Fprintf(log, "overnight: RunLoop halted — %s\n", result.RegressionReason)
				return result, nil
			}
			if regressed {
				iter.Degraded = append(iter.Degraded,
					fmt.Sprintf("regression beyond floor (warn-only): %d metric(s)", len(regressions)))
			}
			if plateau.Observe(composite) && !opts.WarnOnly {
				iter.FinishedAt = time.Now()
				iter.Duration = iter.FinishedAt.Sub(iterStart).String()
				result.Iterations = append(result.Iterations, iter)
				result.PlateauReason = plateau.Reason()
				fmt.Fprintf(log, "overnight: RunLoop halted — %s\n", result.PlateauReason)
				return result, nil
			}
		}

		iter.FinishedAt = time.Now()
		iter.Duration = iter.FinishedAt.Sub(iterStart).String()
		result.Iterations = append(result.Iterations, iter)
		snap := currSnapshot
		prevSnapshot = &snap

		fmt.Fprintf(log, "overnight: iteration %d done (elapsed=%s, total=%s)\n",
			iterIndex, iter.Duration, time.Since(startedAt))
	}
}

// ingestSummary marshals an IngestResult into the opaque map the
// IterationSummary uses so downstream JSON consumers see it without
// pulling in Go types.
func ingestSummary(r *IngestResult) map[string]any {
	return map[string]any{
		"harvest_preview_count": r.HarvestPreviewCount,
		"forge_artifacts_mined": r.ForgeArtifactsMined,
		"provenance_audited":    r.ProvenanceAudited,
		"mine_findings_new":     r.MineFindingsNew,
	}
}

// reduceSummary marshals a ReduceResult into the opaque map the
// IterationSummary uses.
func reduceSummary(r *ReduceResult) map[string]any {
	return map[string]any{
		"harvest_promoted":    r.HarvestPromoted,
		"dedup_merged":        r.DedupMerged,
		"maturity_tempered":   r.MaturityTempered,
		"defrag_pruned":       r.DefragPruned,
		"close_loop_promoted": r.CloseLoopPromoted,
		"findings_routed":     r.FindingsRouted,
		"inject_refreshed":    r.InjectRefreshed,
		"rolled_back":         r.RolledBack,
	}
}

// measureSummary marshals a MeasureResult into the opaque map the
// IterationSummary uses.
func measureSummary(r *MeasureResult) map[string]any {
	out := map[string]any{
		"findings_resolved": r.FindingsResolved,
		"inject_visibility": r.InjectVisibility,
	}
	if r.Fitness != nil {
		out["fitness"] = r.Fitness
	}
	return out
}

// snapshotToMap serializes a FitnessSnapshot's metric map for the
// IterationSummary's fitness_before / fitness_after fields.
func snapshotToMap(s FitnessSnapshot) map[string]any {
	out := make(map[string]any, len(s.Metrics))
	for k, v := range s.Metrics {
		out[k] = v
	}
	return out
}

// regressionNames extracts the metric names from a MetricRegression slice
// for inclusion in the RegressionReason string.
func regressionNames(rs []MetricRegression) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.Name)
	}
	return out
}

// logPostCommitFinding writes a structured finding to .agents/findings/
// describing a post-commit metadata integrity failure. The filename
// intentionally does NOT match findingFilenameRe, so the findings router
// will skip it — post-commit strips belong in the morning report as
// warnings, not as auto-routed work.
func logPostCommitFinding(cwd string, iterIndex int, report MetadataIntegrityReport) error {
	findingsDir := filepath.Join(cwd, ".agents", "findings")
	if err := os.MkdirAll(findingsDir, 0o755); err != nil {
		return err
	}
	date := time.Now().Format("2006-01-02")
	name := fmt.Sprintf("f-%s-postcommit-iter-%d.md", date, iterIndex)
	path := filepath.Join(findingsDir, name)

	var b strings.Builder
	fmt.Fprintf(&b, "---\ntitle: Post-commit metadata integrity drift (iter %d)\n", iterIndex)
	fmt.Fprintf(&b, "type: finding\nseverity: high\ndate: %s\n---\n\n", date)
	fmt.Fprintf(&b, "The post-commit VerifyMetadataRoundTripPostCommit check detected %d stripped field(s) "+
		"after Dream iteration %d landed. These fields were present in the pre-REDUCE baseline but are "+
		"missing from the post-commit live tree. This indicates late-stage corruption in the commit swap "+
		"itself, since the pre-commit check would have caught a reducer-layer strip.\n\n",
		len(report.StrippedFields), iterIndex)
	fmt.Fprintln(&b, "## Stripped fields")
	fmt.Fprintln(&b, "")
	for _, sf := range report.StrippedFields {
		fmt.Fprintf(&b, "- `%s`: key `%s`\n", sf.File, sf.Key)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
