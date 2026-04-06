package rpi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// --- Phase result artifacts ---

// PhaseResultFileFmt is the filename pattern for per-phase result artifacts.
// Each phase writes "phase-{N}-result.json" to .agents/rpi/.
// Contract: docs/contracts/rpi-phase-result.schema.json
const PhaseResultFileFmt = "phase-%d-result.json"

// PhaseResult is a structured artifact written after each phase completes or fails.
// Schema: docs/contracts/rpi-phase-result.schema.json
type PhaseResult struct {
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

// ValidatePriorPhaseResult checks that phase-{expectedPhase}-result.json exists
// and has a continuable status. Called at the start of phases 2 and 3.
// Both "completed" and "time_boxed" are treated as continuation signals —
// time_boxed means the phase ran but did not finish within its budget,
// and the next phase should proceed with whatever was accomplished.
func ValidatePriorPhaseResult(cwd string, expectedPhase int) error {
	resultPath := filepath.Join(cwd, ".agents", "rpi", fmt.Sprintf(PhaseResultFileFmt, expectedPhase))
	data, err := os.ReadFile(resultPath)
	if err != nil {
		return fmt.Errorf("prior phase %d result not found at %s: %w", expectedPhase, resultPath, err)
	}

	var result PhaseResult
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("prior phase %d result is malformed: %w", expectedPhase, err)
	}

	if result.Status != "completed" && result.Status != "time_boxed" {
		return fmt.Errorf("prior phase %d has status %q (expected %q or %q)", expectedPhase, result.Status, "completed", "time_boxed")
	}

	return nil
}

// --- State persistence ---

// PhasedStateFile is the filename for orchestrator state.
const PhasedStateFile = "phased-state.json"

// RPIRunRegistryDir returns the per-run registry directory path.
// All per-run artifacts (state, heartbeat) are written here so the registry
// survives interruption and supports resume/status lookup.
// Path: .agents/rpi/runs/<run-id>/
func RPIRunRegistryDir(cwd, runID string) string {
	if runID == "" {
		return ""
	}
	return filepath.Join(cwd, ".agents", "rpi", "runs", runID)
}

// WritePhasedStateAtomic writes data to path using a tmp-file+rename pattern.
// This ensures readers never observe a partial write.
func WritePhasedStateAtomic(path string, data []byte) error {
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

// ParsePhasedStateJSON parses JSON bytes into a map with nil-safe defaults.
// Returns the parsed state as a map along with normalized fields.
// The numPhases parameter is used to validate StartPhase bounds.
//
// This is the pure parsing core; the caller is responsible for unmarshalling
// into the full phasedState struct (which lives in cmd/ao).
func ParsePhasedStateJSON(data []byte) (map[string]json.RawMessage, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}
	return raw, nil
}

// NormalizeParsedState applies backward-compatible defaults to parsed state values.
// goal defaults to "unknown-goal", phase/cycle default to 1, and empty maps
// are initialised. numPhases is the total phase count for StartPhase bounds.
func NormalizeParsedState(goal *string, phase, cycle, startPhase *int, numPhases int,
	verdicts *map[string]string, attempts *map[string]int) {
	if strings.TrimSpace(*goal) == "" {
		*goal = "unknown-goal"
	}
	if *phase <= 0 {
		*phase = 1
	}
	if *cycle <= 0 {
		*cycle = 1
	}
	if *startPhase == 0 {
		*startPhase = *phase
	}
	if *startPhase < 1 || *startPhase > numPhases {
		*startPhase = *phase
	}
	if *verdicts == nil {
		*verdicts = make(map[string]string)
	}
	if *attempts == nil {
		*attempts = make(map[string]int)
	}
}

// LoadLatestRunRegistryState scans .agents/rpi/runs/ and returns the raw JSON
// bytes and run ID from the most recently modified run directory, or an error.
func LoadLatestRunRegistryState(cwd string) ([]byte, error) {
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
		statePath := filepath.Join(runsDir, entry.Name(), PhasedStateFile)
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
	return latestData, nil
}

// ReadRunHeartbeat returns the last heartbeat timestamp for a run, or zero
// time if the heartbeat file does not exist or cannot be parsed.
func ReadRunHeartbeat(cwd, runID string) time.Time {
	if runID == "" {
		return time.Time{}
	}
	heartbeatPath := filepath.Join(RPIRunRegistryDir(cwd, runID), "heartbeat.txt")
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

// RunHeartbeatAge returns the age of the most recent heartbeat for a run.
// If the heartbeat file is missing or unparseable, it returns -1 and false.
func RunHeartbeatAge(cwd, runID string) (time.Duration, bool) {
	ts := ReadRunHeartbeat(cwd, runID)
	if ts.IsZero() {
		return -1, false
	}
	return time.Since(ts), true
}
