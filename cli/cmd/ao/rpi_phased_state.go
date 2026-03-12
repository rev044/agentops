package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// --- Phase result artifacts ---

// phaseResultFileFmt is the filename pattern for per-phase result artifacts.
// Each phase writes "phase-{N}-result.json" to .agents/rpi/.
// Contract: docs/contracts/rpi-phase-result.schema.json
const phaseResultFileFmt = "phase-%d-result.json"

// phaseResult is a structured artifact written after each phase completes or fails.
// Schema: docs/contracts/rpi-phase-result.schema.json
type phaseResult struct {
	SchemaVersion   int               `json:"schema_version"`
	RunID           string            `json:"run_id"`
	Phase           int               `json:"phase"`
	PhaseName       string            `json:"phase_name"`
	Status          string            `json:"status"`
	Retries         int               `json:"retries,omitempty"`
	Error           string            `json:"error,omitempty"`
	Backend         string            `json:"backend,omitempty"`
	Artifacts       map[string]string `json:"artifacts,omitempty"`
	Verdicts        map[string]string `json:"verdicts,omitempty"`
	StartedAt       string            `json:"started_at"`
	CompletedAt     string            `json:"completed_at,omitempty"`
	DurationSeconds float64           `json:"duration_seconds,omitempty"`
}

// writePhaseResult writes a phase-result.json artifact (named phase-{N}-result.json) atomically (write to .tmp, rename).
func writePhaseResult(cwd string, result *phaseResult) error {
	stateDir := filepath.Join(cwd, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0750); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal phase result: %w", err)
	}

	finalPath := filepath.Join(stateDir, fmt.Sprintf(phaseResultFileFmt, result.Phase))
	tmpPath := finalPath + ".tmp"

	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("write phase result tmp: %w", err)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("rename phase result: %w", err)
	}

	VerbosePrintf("Phase result written to %s\n", finalPath)
	return nil
}

// validatePriorPhaseResult checks that phase-{expectedPhase}-result.json exists
// and has a continuable status. Called at the start of phases 2 and 3.
// Both "completed" and "time_boxed" are treated as continuation signals —
// time_boxed means the phase ran but did not finish within its budget,
// and the next phase should proceed with whatever was accomplished.
func validatePriorPhaseResult(cwd string, expectedPhase int) error {
	resultPath := filepath.Join(cwd, ".agents", "rpi", fmt.Sprintf(phaseResultFileFmt, expectedPhase))
	data, err := os.ReadFile(resultPath)
	if err != nil {
		return fmt.Errorf("prior phase %d result not found at %s: %w", expectedPhase, resultPath, err)
	}

	var result phaseResult
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("prior phase %d result is malformed: %w", expectedPhase, err)
	}

	if result.Status != "completed" && result.Status != "time_boxed" {
		return fmt.Errorf("prior phase %d has status %q (expected %q or %q)", expectedPhase, result.Status, "completed", "time_boxed")
	}

	return nil
}

// --- State persistence ---

const phasedStateFile = "phased-state.json"

// rpiRunRegistryDir returns the per-run registry directory path.
// All per-run artifacts (state, heartbeat) are written here so the registry
// survives interruption and supports resume/status lookup.
// Path: .agents/rpi/runs/<run-id>/
func rpiRunRegistryDir(cwd, runID string) string {
	if runID == "" {
		return ""
	}
	return filepath.Join(cwd, ".agents", "rpi", "runs", runID)
}

// savePhasedState writes orchestrator state to disk atomically.
// The state is written to two locations:
//  1. .agents/rpi/phased-state.json (legacy flat path for backward compatibility)
//  2. .agents/rpi/runs/<run-id>/state.json (per-run registry directory)
//
// Both writes use the tmp+rename pattern to prevent corrupt partial writes.
func savePhasedState(cwd string, state *phasedState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	data = append(data, '\n')

	roots := artifactRootsForState(cwd, state)
	for i, root := range roots {
		if writeErr := writePhasedStateData(root, state.RunID, data); writeErr != nil {
			if i == 0 {
				return writeErr
			}
			VerbosePrintf("Warning: mirror state write skipped for %s: %v\n", root, writeErr)
		}
	}

	// Mirror state to additional roots discovered via mirrorRootsForEvent
	// and emit C2 tracking events for observability.
	mirrorStateToPeers(cwd, state, data)

	return nil
}

// mirrorStateToPeers writes the serialised state to every mirror root returned
// by mirrorRootsForEvent that is NOT the primary cwd.  For each mirror root it
// emits a "state.mirrored" C2 event on success or "state.mirror.failed" on error.
func mirrorStateToPeers(cwd string, state *phasedState, data []byte) {
	if state == nil || state.RunID == "" {
		return
	}
	primaryClean := filepath.Clean(cwd)
	for _, mirrorRoot := range mirrorRootsForEvent(cwd, state.RunID) {
		if filepath.Clean(mirrorRoot) == primaryClean {
			continue
		}
		if writeErr := writePhasedStateData(mirrorRoot, state.RunID, data); writeErr != nil {
			VerbosePrintf("Warning: mirror state write failed for %s: %v\n", mirrorRoot, writeErr)
			if _, evErr := appendRPIC2Event(cwd, rpiC2EventInput{
				RunID: state.RunID, Source: "orchestrator",
				Type:    "state.mirror.failed",
				Message: fmt.Sprintf("mirror state write failed for %s", mirrorRoot),
				Details: map[string]any{"mirror_root": mirrorRoot, "error": writeErr.Error()},
			}); evErr != nil {
				VerbosePrintf("Warning: could not emit state.mirror.failed event: %v\n", evErr)
			}
			continue
		}
		if _, evErr := appendRPIC2Event(cwd, rpiC2EventInput{
			RunID: state.RunID, Source: "orchestrator",
			Type:    "state.mirrored",
			Message: fmt.Sprintf("state mirrored to %s", mirrorRoot),
			Details: map[string]any{"mirror_root": mirrorRoot, "file": phasedStateFile},
		}); evErr != nil {
			VerbosePrintf("Warning: could not emit state.mirrored event: %v\n", evErr)
		}
	}
}

func writePhasedStateData(root, runID string, data []byte) error {
	stateDir := filepath.Join(root, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0750); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	flatPath := filepath.Join(stateDir, phasedStateFile)
	if err := writePhasedStateAtomic(flatPath, data); err != nil {
		return fmt.Errorf("write state: %w", err)
	}

	if runID != "" {
		runDir := rpiRunRegistryDir(root, runID)
		if mkErr := os.MkdirAll(runDir, 0750); mkErr != nil {
			VerbosePrintf("Warning: create run registry dir: %v\n", mkErr)
		} else {
			registryPath := filepath.Join(runDir, phasedStateFile)
			if wErr := writePhasedStateAtomic(registryPath, data); wErr != nil {
				VerbosePrintf("Warning: write run registry state: %v\n", wErr)
			}
		}
	}

	VerbosePrintf("State saved to %s\n", flatPath)
	return nil
}

// artifactRootsForState returns persistence roots for state artifacts.
// Primary root is always cwd; when running inside an isolated worktree, a
// mirror root is added for the original repo so supervisors can observe run
// state without traversing worktree directories.
func artifactRootsForState(cwd string, state *phasedState) []string {
	roots := []string{cwd}
	if state == nil || state.WorktreePath == "" {
		return roots
	}

	if filepath.Clean(state.WorktreePath) != filepath.Clean(cwd) {
		return roots
	}

	if repoRoot := inferSupervisorRepoRoot(cwd, state.RunID); repoRoot != "" && repoRoot != filepath.Clean(cwd) {
		roots = append(roots, repoRoot)
	}
	return roots
}

func inferSupervisorRepoRoot(cwd, runID string) string {
	ctx, cancel := context.WithTimeout(context.Background(), worktreeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-common-dir")
	cmd.Dir = cwd
	if out, err := cmd.Output(); err == nil {
		commonDir := filepath.Clean(strings.TrimSpace(string(out)))
		if commonDir != "" {
			if !filepath.IsAbs(commonDir) {
				commonDir = filepath.Clean(filepath.Join(cwd, commonDir))
			}
			repoRoot := filepath.Dir(commonDir)
			if repoRoot != "" {
				return filepath.Clean(repoRoot)
			}
		}
	}

	if strings.TrimSpace(runID) == "" {
		return ""
	}
	suffix := "-rpi-" + runID
	base := filepath.Base(cwd)
	if !strings.HasSuffix(base, suffix) {
		return ""
	}
	repoBase := strings.TrimSuffix(base, suffix)
	if repoBase == "" {
		return ""
	}
	candidate := filepath.Join(filepath.Dir(cwd), repoBase)
	if info, err := os.Stat(candidate); err == nil && info.IsDir() {
		return filepath.Clean(candidate)
	}
	return ""
}

// artifactRootsForRun infers persistence roots from stored run state.
// Used by heartbeat updates, where only cwd and runID are available.
func artifactRootsForRun(cwd, runID string) []string {
	if runID == "" {
		return []string{cwd}
	}
	statePath := filepath.Join(rpiRunRegistryDir(cwd, runID), phasedStateFile)
	data, err := os.ReadFile(statePath)
	if err != nil {
		return []string{cwd}
	}
	state, err := parsePhasedState(data)
	if err != nil {
		return []string{cwd}
	}
	return artifactRootsForState(cwd, state)
}

// writePhasedStateAtomic writes data to path using a tmp-file+rename pattern.
// This ensures readers never observe a partial write.
func writePhasedStateAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".phased-state-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create tmp file: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		_ = tmp.Close()
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync tmp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close tmp: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename tmp: %w", err)
	}
	cleanup = false
	return nil
}

// loadPhasedState reads orchestrator state from disk.
// It first tries the per-run registry directory (most recent run), then falls
// back to the flat .agents/rpi/phased-state.json path for backward compatibility.
func loadPhasedState(cwd string) (*phasedState, error) {
	flatPath := filepath.Join(cwd, ".agents", "rpi", phasedStateFile)

	// Try to find the most recently modified state in any run registry directory.
	// This allows resume when the worktree only has the registry (not the flat file).
	runState, runErr := loadLatestRunRegistryState(cwd)
	if runErr == nil && runState != nil {
		// Prefer registry state only when it is newer than (or the same as) the flat file.
		flatInfo, flatStatErr := os.Stat(flatPath)
		if flatStatErr != nil {
			// Flat file does not exist — use registry state.
			return runState, nil
		}
		registryPath := filepath.Join(rpiRunRegistryDir(cwd, runState.RunID), phasedStateFile)
		registryInfo, regStatErr := os.Stat(registryPath)
		if regStatErr == nil && !registryInfo.ModTime().Before(flatInfo.ModTime()) {
			return runState, nil
		}
	}

	// Fall back to flat path.
	data, err := os.ReadFile(flatPath)
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}
	return parsePhasedState(data)
}

// loadLatestRunRegistryState scans .agents/rpi/runs/ and returns the state
// from the most recently modified run directory, or nil if none exists.
func loadLatestRunRegistryState(cwd string) (*phasedState, error) {
	runsDir := filepath.Join(cwd, ".agents", "rpi", "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		return nil, err
	}

	var latestModTime int64
	var latestData []byte

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		statePath := filepath.Join(runsDir, entry.Name(), phasedStateFile)
		info, err := os.Stat(statePath)
		if err != nil {
			continue
		}
		if info.ModTime().UnixNano() > latestModTime {
			data, readErr := os.ReadFile(statePath)
			if readErr != nil {
				continue
			}
			latestModTime = info.ModTime().UnixNano()
			latestData = data
		}
	}

	if latestData == nil {
		return nil, os.ErrNotExist
	}
	return parsePhasedState(latestData)
}

// parsePhasedState parses JSON bytes into a phasedState with nil-safe maps.
func parsePhasedState(data []byte) (*phasedState, error) {
	var state phasedState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}
	if strings.TrimSpace(state.Goal) == "" {
		state.Goal = "unknown-goal"
	}
	// Backward compatibility: older/partial states may have zero values.
	// Normalize to safe defaults instead of hard-failing recovery paths.
	if state.Phase <= 0 {
		state.Phase = 1
	}
	if state.Cycle <= 0 {
		state.Cycle = 1
	}
	// Backward compatibility: older states may omit start_phase.
	if state.StartPhase == 0 {
		state.StartPhase = state.Phase
	}
	if state.StartPhase < 1 || state.StartPhase > len(phases) {
		state.StartPhase = state.Phase
	}

	// Ensure maps are never nil after deserialization.
	if state.Verdicts == nil {
		state.Verdicts = make(map[string]string)
	}
	if state.Attempts == nil {
		state.Attempts = make(map[string]int)
	}

	return &state, nil
}

// updateRunHeartbeat writes the current UTC timestamp to
// .agents/rpi/runs/<run-id>/heartbeat.txt atomically.
// It is called during phase execution to signal the run is alive.
// Failures are logged but do not abort the phase.
func updateRunHeartbeat(cwd, runID string) {
	if runID == "" {
		return
	}
	ts := time.Now().UTC().Format(time.RFC3339Nano) + "\n"
	for _, root := range artifactRootsForRun(cwd, runID) {
		runDir := rpiRunRegistryDir(root, runID)
		if err := os.MkdirAll(runDir, 0750); err != nil {
			VerbosePrintf("Warning: create run dir for heartbeat: %v\n", err)
			continue
		}
		heartbeatPath := filepath.Join(runDir, "heartbeat.txt")
		if err := writePhasedStateAtomic(heartbeatPath, []byte(ts)); err != nil {
			VerbosePrintf("Warning: update heartbeat: %v\n", err)
		}
	}
}

// readRunHeartbeat returns the last heartbeat timestamp for a run, or zero
// time if the heartbeat file does not exist or cannot be parsed.
func readRunHeartbeat(cwd, runID string) time.Time {
	if runID == "" {
		return time.Time{}
	}
	heartbeatPath := filepath.Join(rpiRunRegistryDir(cwd, runID), "heartbeat.txt")
	data, err := os.ReadFile(heartbeatPath)
	if err != nil {
		return time.Time{}
	}
	for _, line := range strings.Split(string(data), "\n") {
		candidate := strings.TrimSpace(line)
		if candidate == "" {
			continue
		}
		if ts, err := time.Parse(time.RFC3339Nano, candidate); err == nil {
			return ts
		}
		if ts, err := time.Parse(time.RFC3339, candidate); err == nil {
			return ts
		}
		break
	}
	return time.Time{}
}

// runHeartbeatAge returns the age of the most recent heartbeat for a run.
// If the heartbeat file is missing or unparseable, it returns -1 and false.
func runHeartbeatAge(cwd, runID string) (time.Duration, bool) {
	ts := readRunHeartbeat(cwd, runID)
	if ts.IsZero() {
		return -1, false
	}
	return time.Since(ts), true
}
