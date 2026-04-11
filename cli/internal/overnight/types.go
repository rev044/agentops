package overnight

import (
	"io"
	"time"
)

// IterationID is the stable identifier for a single iteration inside a
// Dream run. Format: "<run-id>-iter-<N>" where N is 1-based.
type IterationID string

// RunLoopOptions bundles every knob RunLoop consumes from the caller.
//
// All callers flow through the overnight command (cli/cmd/ao/overnight.go);
// the options struct is intentionally flat so flag-to-field mapping stays
// easy to audit. Defaults are documented inline; zero values are never
// interpreted as "do something special" — validators in RunLoop substitute
// the documented default instead.
type RunLoopOptions struct {
	// Cwd is the repository root. Dream's .agents/ lives under this path.
	Cwd string

	// OutputDir is where per-run artifacts are written. Typically
	// ".agents/overnight/<run-id>/".
	OutputDir string

	// RunID is the caller-assigned identifier for this Dream run. Used to
	// namespace per-iteration persistence under <OutputDir>/<RunID>/iterations/
	// so two runs sharing an OutputDir (the default `.agents/overnight/latest`
	// is shared by every run) cannot cross-contaminate each other's history.
	// Also used as the "<run-id>" prefix in IterationID values per the contract
	// at line 10. Required; RunLoop returns an error if empty.
	RunID string

	// RunTimeout caps the outer loop's wall-clock. Default: 2h.
	// Capped at 6h regardless of flag value.
	RunTimeout time.Duration

	// MaxIterations caps the outer loop count. 0 = budget-bounded only.
	MaxIterations int

	// PlateauEpsilon is the |delta| threshold that counts as "no progress"
	// for a single iteration. Default: 0.01.
	PlateauEpsilon float64

	// PlateauWindowK is the number of consecutive sub-epsilon deltas
	// required to declare a plateau. Default: 2. Never less than 2 — a
	// single noisy sample is absorbed by design.
	PlateauWindowK int

	// RegressionFloor is the maximum allowed drop for any single metric in
	// the fitness vector between iterations. Default: 0.05.
	RegressionFloor float64

	// WarnOnly, when true, turns plateau and regression detection into
	// warnings instead of halts. Used during the first 2-3 production
	// Dream runs to calibrate thresholds empirically. Default: true (opt
	// in to strict mode once thresholds are calibrated).
	WarnOnly bool

	// QueuePath optionally points at an operator-pinned roadmap
	// (markdown) whose items Dream works in order before falling through
	// to fitness-driven work selection. Reuses the evolve pinned-queue
	// format; see skills/evolve/references/pinned-queue.md.
	QueuePath string

	// CheckpointMaxBytes caps disk usage across all concurrent checkpoint
	// snapshots in a single run. Default: 512 MB. On exceed, NewCheckpoint
	// returns an error and the iteration degrades (never half-mutates).
	CheckpointMaxBytes int64

	// LockStaleAfter is the threshold past which a .agents/overnight/run.lock
	// file with a dead PID is reclaimed on startup. Default: 12h.
	LockStaleAfter time.Duration

	// LogWriter receives structured progress output. Typically Dream's
	// existing overnight.log file. Nil is allowed — RunLoop substitutes
	// io.Discard.
	LogWriter io.Writer
}

// defaultRunTimeout is the documented default wall-clock cap.
const defaultRunTimeout = 2 * time.Hour

// maxRunTimeout is the hard upper bound regardless of flag value.
const maxRunTimeout = 6 * time.Hour

// defaultPlateauEpsilon is the documented default plateau threshold.
const defaultPlateauEpsilon = 0.01

// defaultPlateauWindowK is the documented default plateau window.
const defaultPlateauWindowK = 2

// defaultRegressionFloor is the documented default per-metric regression floor.
const defaultRegressionFloor = 0.05

// defaultCheckpointMaxBytes is the documented default checkpoint-storage cap.
const defaultCheckpointMaxBytes = int64(512 * 1024 * 1024) // 512 MB

// defaultLockStaleAfter is the documented default stale-lock reclaim threshold.
const defaultLockStaleAfter = 12 * time.Hour

// normalize returns a copy of opts with zero/out-of-range values replaced
// by documented defaults. It never returns an error; every correction is
// a silent substitution recorded in the RunLoopResult's Degraded list.
func (opts RunLoopOptions) normalize() (RunLoopOptions, []string) {
	var degraded []string
	if opts.RunTimeout <= 0 {
		opts.RunTimeout = defaultRunTimeout
	}
	if opts.RunTimeout > maxRunTimeout {
		degraded = append(degraded, "RunTimeout clamped to 6h hard max")
		opts.RunTimeout = maxRunTimeout
	}
	if opts.PlateauEpsilon <= 0 {
		opts.PlateauEpsilon = defaultPlateauEpsilon
	}
	if opts.PlateauWindowK < defaultPlateauWindowK {
		// K=1 is ill-defined; a single noisy sample would halt the loop.
		if opts.PlateauWindowK > 0 {
			degraded = append(degraded, "PlateauWindowK raised to 2 (minimum)")
		}
		opts.PlateauWindowK = defaultPlateauWindowK
	}
	if opts.RegressionFloor <= 0 {
		opts.RegressionFloor = defaultRegressionFloor
	}
	if opts.CheckpointMaxBytes <= 0 {
		opts.CheckpointMaxBytes = defaultCheckpointMaxBytes
	}
	if opts.LockStaleAfter <= 0 {
		opts.LockStaleAfter = defaultLockStaleAfter
	}
	return opts, degraded
}
