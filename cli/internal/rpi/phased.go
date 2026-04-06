package rpi

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PhaseFailureReason classifies why a phase spawn failed.
type PhaseFailureReason string

const (
	FailReasonTimeout PhaseFailureReason = "timeout"
	FailReasonStall   PhaseFailureReason = "stall"
	FailReasonExit    PhaseFailureReason = "exit_error"
	FailReasonUnknown PhaseFailureReason = "unknown"
)

// MinPositiveDuration returns the smaller of two durations, ignoring non-positive values.
// If both are non-positive, returns b.
func MinPositiveDuration(a, b time.Duration) time.Duration {
	if a <= 0 {
		return b
	}
	if b <= 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}

// AppendTimeBoxedMarker appends a [TIME-BOXED] marker to the phase summary file.
func AppendTimeBoxedMarker(spawnCwd string, phaseNum int, phaseName string, budget time.Duration) error {
	stateDir := filepath.Join(spawnCwd, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("create rpi state directory: %w", err)
	}

	summaryPath := filepath.Join(stateDir, fmt.Sprintf("phase-%d-summary.md", phaseNum))
	f, err := os.OpenFile(summaryPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open phase summary for marker: %w", err)
	}
	defer f.Close()

	marker := fmt.Sprintf("[TIME-BOXED] Phase %s time-boxed at %ds (budget: %ds)\n", phaseName, int(budget.Seconds()), int(budget.Seconds()))
	if _, err := f.WriteString(marker); err != nil {
		return fmt.Errorf("write time-box marker: %w", err)
	}
	return nil
}
