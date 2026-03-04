package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// contextArtifactDir returns the path for context-scoped artifacts.
// If runID is empty, generates an adhoc identifier from the current timestamp.
// NOTE: Not yet wired into the --for codepath — artifact path wiring deferred to Phase 2.
// These functions are available for orchestrators to call directly.
func contextArtifactDir(runID string) string {
	if runID == "" {
		runID = fmt.Sprintf("adhoc-%d", time.Now().Unix())
	}
	return filepath.Join(".agents", "context", runID)
}

// ensureContextDir creates the context artifact directory on disk.
func ensureContextDir(cwd, runID string) (string, error) {
	dir := filepath.Join(cwd, contextArtifactDir(runID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create context dir: %w", err)
	}
	return dir, nil
}
