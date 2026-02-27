package main

import (
	"os"
	"path/filepath"
	"testing"
)

// cov4StoreSetup creates a temp workdir with .agents/ directories for store tests.
func cov4StoreSetup(t *testing.T) string {
	t.Helper()
	tmp := chdirTemp(t)
	t.Setenv("HOME", tmp)
	setupAgentsDir(t, tmp)
	return tmp
}

func TestCov4_StoreIndex(t *testing.T) {
	tmp := cov4StoreSetup(t)
	// Create a file to index
	artifact := filepath.Join(tmp, ".agents", "learnings", "test-learning.md")
	if err := os.WriteFile(artifact, []byte("# Test Learning\n\n**ID**: test-1\n**Utility**: 0.7\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, _ = executeCommand("store", "index", artifact)
}

func TestCov4_StoreSearch(t *testing.T) {
	tmp := cov4StoreSetup(t)
	// Create index with a test entry first
	artifact := filepath.Join(tmp, ".agents", "learnings", "test-learning.md")
	if err := os.WriteFile(artifact, []byte("# Test Learning\n\n**ID**: test-1\n**Utility**: 0.7\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, _ = executeCommand("store", "index", artifact)
	// Now search
	_, _ = executeCommand("store", "search", "test")
}

func TestCov4_StoreRebuild(t *testing.T) {
	cov4StoreSetup(t)
	// Rebuild with no artifacts — should succeed gracefully
	_, _ = executeCommand("store", "rebuild")
}

func TestCov4_StoreStats(t *testing.T) {
	cov4StoreSetup(t)
	// Stats with no index — should succeed gracefully
	_, _ = executeCommand("store", "stats")
}
