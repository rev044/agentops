package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// contextArtifactDir returns the path for context-scoped artifacts.
// If runID is empty, generates an adhoc identifier from the current timestamp.
// Called automatically when --for is used; uses RPI_RUN_ID if set.
func contextArtifactDir(runID string) string {
	if runID == "" {
		runID = fmt.Sprintf("adhoc-%d-%04x", time.Now().Unix(), rand.Intn(0x10000)) //nolint:gosec // non-cryptographic use
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
