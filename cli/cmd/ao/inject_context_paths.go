package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// contextArtifactDir returns the path for context-scoped artifacts.
// If runID is empty, generates an adhoc identifier from the current timestamp.
// Called automatically when --for is used; uses RPI_RUN_ID if set.
func contextArtifactDir(runID string, randReader io.Reader) string {
	if randReader == nil {
		randReader = rand.Reader
	}
	if runID == "" {
		runID = newAdhocContextRunID(time.Now(), randReader)
	}
	return filepath.Join(".agents", "context", runID)
}

func newAdhocContextRunID(now time.Time, r io.Reader) string {
	suffix := make([]byte, 2)
	if _, err := io.ReadFull(r, suffix); err != nil {
		return fmt.Sprintf("adhoc-%d-%04x", now.Unix(), uint16(now.UnixNano()))
	}
	return fmt.Sprintf("adhoc-%d-%s", now.Unix(), hex.EncodeToString(suffix))
}

// ensureContextDir creates the context artifact directory on disk.
func ensureContextDir(cwd, runID string, randReader io.Reader) (string, error) {
	dir := filepath.Join(cwd, contextArtifactDir(runID, randReader))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create context dir: %w", err)
	}
	return dir, nil
}
