package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// writePromptAuditTrail writes the fully-rendered prompt to the run's audit directory.
// Path: .agents/rpi/runs/<run-id>/phase-<N>-prompt.md
func writePromptAuditTrail(cwd, runID string, phaseNum int, prompt string) error {
	if runID == "" {
		return nil // no-op for empty run ID
	}
	dir := filepath.Join(cwd, ".agents", "rpi", "runs", runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, fmt.Sprintf("phase-%d-prompt.md", phaseNum))
	return os.WriteFile(path, []byte(prompt), 0o644)
}
