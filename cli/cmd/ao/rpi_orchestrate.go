package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// orchPhase represents the current lifecycle phase of an RPI orchestration run.
type orchPhase string

const (
	orchPhaseInit       orchPhase = "init"
	orchPhaseDiscovery  orchPhase = "discovery"
	orchPhaseImpl       orchPhase = "implementation"
	orchPhaseValidation orchPhase = "validation"
	orchPhaseDone       orchPhase = "done"
	orchPhaseFailed     orchPhase = "failed"
)

// beadWorkerStatus represents the execution status of a single bead worker.
type beadWorkerStatus string

const (
	beadPending beadWorkerStatus = "pending"
	beadRunning beadWorkerStatus = "running"
	beadDone    beadWorkerStatus = "done"
	beadFailed  beadWorkerStatus = "failed"
)

// beadWorker tracks state for one bead's isolated worker session.
type beadWorker struct {
	BeadID    string           `json:"bead_id"`
	WorkerID  string           `json:"worker_id"`
	Status    beadWorkerStatus `json:"status"`
	Attempts  int              `json:"attempts"`
	StartedAt string           `json:"started_at,omitempty"`
	DoneAt    string           `json:"done_at,omitempty"`
	Error     string           `json:"error,omitempty"`
}

// orchState persists the full orchestration run state to disk between phase transitions.
type orchState struct {
	SchemaVersion  int            `json:"schema_version"`
	RunID          string         `json:"run_id"`
	Goal           string         `json:"goal"`
	Phase          orchPhase      `json:"phase"`
	EpicID         string         `json:"epic_id,omitempty"`
	Beads          []beadWorker   `json:"beads,omitempty"`
	Attempts       map[string]int `json:"attempts"`
	StartedAt      string         `json:"started_at"`
	UpdatedAt      string         `json:"updated_at"`
	TerminalStatus string         `json:"terminal_status,omitempty"`
	TerminalReason string         `json:"terminal_reason,omitempty"`
}

// orchOpts configures the orchestration engine.
type orchOpts struct {
	MaxAttempts    int
	BDCommand      string
	RuntimeCommand string
	AOCommand      string
	PhaseTimeout   time.Duration
	NoWorktree     bool
}

// defaultOrchOpts returns production-safe defaults.
func defaultOrchOpts() orchOpts {
	return orchOpts{
		MaxAttempts:    3,
		BDCommand:      "bd",
		RuntimeCommand: "claude",
		AOCommand:      "ao",
		PhaseTimeout:   90 * time.Minute,
	}
}

// generateWorkerID returns a short random hex identifier for a worker session.
func generateWorkerID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("w-%x", b)
}

// saveOrchState persists state to orchestration-state.json in the run registry dir.
func saveOrchState(root, runID string, state *orchState) error {
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	dir := rpiRunRegistryDir(root, runID)
	if dir == "" {
		return fmt.Errorf("could not resolve run registry dir for run %s", runID)
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create run dir: %w", err)
	}
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal orchState: %w", err)
	}
	path := filepath.Join(dir, "orchestration-state.json")
	if err := os.WriteFile(path, data, 0o640); err != nil {
		return fmt.Errorf("write orchestration-state.json: %w", err)
	}
	return nil
}

// beadIDPattern matches bead IDs like ag-123 or ag-000.1
var beadIDPattern = regexp.MustCompile(`^[a-z]+-[a-z0-9.]+$`)

// parseBeadIDsFromText extracts bead IDs from bd children text output lines.
// Each non-empty line's first whitespace-delimited token is checked against beadIDPattern.
func parseBeadIDsFromText(output string) []string {
	var ids []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		token := strings.Fields(line)[0]
		// Strip leading status indicators (○, ◐, ●, ✓, ❄)
		token = strings.TrimLeft(token, "○◐●✓❄ \t")
		if beadIDPattern.MatchString(token) {
			ids = append(ids, token)
		}
	}
	sort.Strings(ids)
	return ids
}

// enumerateBeads returns the list of open bead IDs for an epic via bd children.
func enumerateBeads(epicID, bdCommand string) ([]string, error) {
	cmd := exec.Command(effectiveBDCommand(bdCommand), "children", epicID)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("bd children %s: %w", epicID, err)
	}
	return parseBeadIDsFromText(string(out)), nil
}

// runBeadWorker executes one bead's implementation in an isolated worker session.
// On failure it retries up to opts.MaxAttempts times with a fresh worker each time.
func runBeadWorker(ctx context.Context, beadID, runID string, bw *beadWorker, root string, opts orchOpts) error {
	bw.BeadID = beadID
	bw.Status = beadPending

	for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		workerID := generateWorkerID()
		bw.WorkerID = workerID
		bw.Attempts = attempt
		bw.StartedAt = time.Now().UTC().Format(time.RFC3339)
		bw.Status = beadRunning

		_, _ = appendRPIC2Event(root, rpiC2EventInput{
			RunID:    runID,
			WorkerID: workerID,
			Type:     "worker.bead.spawned",
			Message:  fmt.Sprintf("bead %s attempt %d started", beadID, attempt),
			Details:  map[string]any{"bead_id": beadID},
		})

		prompt := "/implement " + beadID
		err := spawnRuntimeDirectImpl(opts.RuntimeCommand, prompt, root, 2, opts.PhaseTimeout)
		if err == nil {
			bw.Status = beadDone
			bw.DoneAt = time.Now().UTC().Format(time.RFC3339)
			_, _ = appendRPIC2Event(root, rpiC2EventInput{
				RunID:    runID,
				WorkerID: workerID,
				Type:     "worker.bead.done",
				Message:  fmt.Sprintf("bead %s completed on attempt %d", beadID, attempt),
				Details:  map[string]any{"bead_id": beadID},
			})
			return nil
		}

		bw.Error = err.Error()
		_, _ = appendRPIC2Event(root, rpiC2EventInput{
			RunID:    runID,
			WorkerID: workerID,
			Type:     "worker.bead.failed",
			Message:  fmt.Sprintf("bead %s attempt %d failed: %v", beadID, attempt, err),
			Details:  map[string]any{"bead_id": beadID, "error": err.Error()},
		})
	}

	bw.Status = beadFailed
	return fmt.Errorf("bead %s failed after %d attempts", beadID, opts.MaxAttempts)
}

// dispatchBeadWorkers runs one isolated worker per bead in parallel via WaitGroup.
// Uses indexed slice writes to avoid mutex (each goroutine owns its index).
func dispatchBeadWorkers(ctx context.Context, state *orchState, root string, opts orchOpts) error {
	beadIDs, err := enumerateBeads(state.EpicID, opts.BDCommand)
	if err != nil {
		return fmt.Errorf("enumerate beads: %w", err)
	}
	if len(beadIDs) == 0 {
		fmt.Printf("orchestrate: no beads found for epic %s, skipping implementation phase\n", state.EpicID)
		return nil
	}

	beads := make([]beadWorker, len(beadIDs))
	errs := make([]error, len(beadIDs))
	var wg sync.WaitGroup

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, id := range beadIDs {
		i, id := i, id // capture by value
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = runBeadWorker(cancelCtx, id, state.RunID, &beads[i], root, opts)
			if errs[i] != nil {
				cancel() // cancel remaining workers on first failure
			}
		}()
	}

	wg.Wait()
	// Aggregate results regardless of error.
	state.Beads = beads
	if saveErr := saveOrchState(root, state.RunID, state); saveErr != nil {
		VerbosePrintf("Warning: could not save orch state after bead dispatch: %v\n", saveErr)
	}

	// Return first error encountered.
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

// spawnPhaseWorker spawns a fresh Claude session for a single RPI phase (discovery or validation).
func spawnPhaseWorker(_ context.Context, phaseNum int, phaseName, _ string, root string, workerID string, state *orchState, opts orchOpts) error {
	var prompt string
	switch phaseNum {
	case 1:
		prompt = "You are in RPI discovery phase. Goal: " + state.Goal + " Run /research, /plan, /pre-mortem"
	case 3:
		prompt = "You are in RPI validation phase. Epic: " + state.EpicID + " Run /vibe recent and /post-mortem " + state.EpicID
	default:
		return fmt.Errorf("unknown phase number %d for phase worker", phaseNum)
	}

	_, _ = appendRPIC2Event(root, rpiC2EventInput{
		RunID:    state.RunID,
		WorkerID: workerID,
		Phase:    phaseNum,
		Type:     "worker.phase.spawned",
		Message:  fmt.Sprintf("phase %d (%s) worker started", phaseNum, phaseName),
	})

	err := spawnRuntimeDirectImpl(opts.RuntimeCommand, prompt, root, phaseNum, opts.PhaseTimeout)

	evType := "worker.phase.done"
	msg := fmt.Sprintf("phase %d (%s) worker completed", phaseNum, phaseName)
	if err != nil {
		evType = "worker.phase.failed"
		msg = fmt.Sprintf("phase %d (%s) worker failed: %v", phaseNum, phaseName, err)
	}
	_, _ = appendRPIC2Event(root, rpiC2EventInput{
		RunID:    state.RunID,
		WorkerID: workerID,
		Phase:    phaseNum,
		Type:     evType,
		Message:  msg,
	})
	return err
}

// runRPIOrchestration drives the full RPI lifecycle with per-phase and per-bead worker isolation.
// It is the production orchestration engine invoked by ao rpi serve <goal>.
//
// State machine: INIT -> DISCOVERY -> IMPL -> VALIDATION -> DONE/FAILED
// Failure policy: each failure spawns a fresh worker (new workerID, fresh context).
func runRPIOrchestration(ctx context.Context, goal, runID, root string, opts orchOpts) error {
	state := &orchState{
		SchemaVersion: 1,
		RunID:         runID,
		Goal:          goal,
		Phase:         orchPhaseInit,
		Attempts:      map[string]int{},
		StartedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	if err := saveOrchState(root, runID, state); err != nil {
		return fmt.Errorf("init orch state: %w", err)
	}

	// --- Phase 1: Discovery ---
	state.Phase = orchPhaseDiscovery
	discoveryOK := false
	for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		state.Attempts["discovery"] = attempt
		if err := saveOrchState(root, runID, state); err != nil {
			VerbosePrintf("Warning: could not save state before discovery attempt %d: %v\n", attempt, err)
		}

		workerID := generateWorkerID()
		err := spawnPhaseWorker(ctx, 1, "discovery", runID, root, workerID, state, opts)
		if err == nil {
			discoveryOK = true
			break
		}
		fmt.Printf("orchestrate: discovery attempt %d/%d failed: %v\n", attempt, opts.MaxAttempts, err)
	}
	if !discoveryOK {
		state.Phase = orchPhaseFailed
		state.TerminalStatus = "failed"
		state.TerminalReason = fmt.Sprintf("discovery failed after %d attempts", opts.MaxAttempts)
		_ = saveOrchState(root, runID, state)
		return fmt.Errorf("RPI orchestration: %s", state.TerminalReason)
	}

	// Extract epic ID produced by the discovery worker.
	epicID, err := extractEpicID(opts.BDCommand)
	if err != nil {
		state.Phase = orchPhaseFailed
		state.TerminalStatus = "failed"
		state.TerminalReason = fmt.Sprintf("extract epic ID after discovery: %v", err)
		_ = saveOrchState(root, runID, state)
		return fmt.Errorf("RPI orchestration: %s", state.TerminalReason)
	}
	state.EpicID = epicID

	// --- Phase 2: Implementation (per-bead workers) ---
	state.Phase = orchPhaseImpl
	if err := saveOrchState(root, runID, state); err != nil {
		VerbosePrintf("Warning: could not save state at impl start: %v\n", err)
	}
	if err := dispatchBeadWorkers(ctx, state, root, opts); err != nil {
		state.Phase = orchPhaseFailed
		state.TerminalStatus = "failed"
		state.TerminalReason = fmt.Sprintf("implementation phase: %v", err)
		_ = saveOrchState(root, runID, state)
		return fmt.Errorf("RPI orchestration: %s", state.TerminalReason)
	}

	// --- Phase 3: Validation ---
	state.Phase = orchPhaseValidation
	validationOK := false
	for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		state.Attempts["validation"] = attempt
		if err := saveOrchState(root, runID, state); err != nil {
			VerbosePrintf("Warning: could not save state before validation attempt %d: %v\n", attempt, err)
		}

		workerID := generateWorkerID()
		err := spawnPhaseWorker(ctx, 3, "validation", runID, root, workerID, state, opts)
		if err == nil {
			validationOK = true
			break
		}
		fmt.Printf("orchestrate: validation attempt %d/%d failed: %v\n", attempt, opts.MaxAttempts, err)
	}
	if !validationOK {
		state.Phase = orchPhaseFailed
		state.TerminalStatus = "failed"
		state.TerminalReason = fmt.Sprintf("validation failed after %d attempts", opts.MaxAttempts)
		_ = saveOrchState(root, runID, state)
		return fmt.Errorf("RPI orchestration: %s", state.TerminalReason)
	}

	state.Phase = orchPhaseDone
	state.TerminalStatus = "done"
	_ = saveOrchState(root, runID, state)
	fmt.Printf("RPI orchestration complete. Goal: %s | Epic: %s\n", goal, epicID)
	return nil
}
