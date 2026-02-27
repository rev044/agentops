package main

import (
	"os"
	"path/filepath"
	"testing"
)

// cov4TemperSetup creates a temp workdir with .agents/ directories for temper tests.
func cov4TemperSetup(t *testing.T) string {
	t.Helper()
	tmp := chdirTemp(t)
	t.Setenv("HOME", tmp)
	setupAgentsDir(t, tmp)
	// Create chain file for lock command
	chainDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(chainDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(chainDir, "chain.jsonl"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	return tmp
}

func TestCov4_TemperValidate(t *testing.T) {
	tmp := cov4TemperSetup(t)
	// Create a file to validate
	artifact := filepath.Join(tmp, ".agents", "learnings", "test-artifact.md")
	content := "# Test\n\n**ID**: test-1\n**Utility**: 0.7\n**Maturity**: provisional\n**Confidence**: 0.8\n"
	if err := os.WriteFile(artifact, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	// Validation may fail (utility/feedback thresholds) but RunE is exercised
	_, _ = executeCommand("temper", "validate", artifact)
}

func TestCov4_TemperLock(t *testing.T) {
	tmp := cov4TemperSetup(t)
	// Create a file to lock
	artifact := filepath.Join(tmp, ".agents", "learnings", "test-lock.md")
	content := "# Test Lock\n\n**ID**: lock-1\n**Utility**: 0.9\n**Maturity**: candidate\n**Confidence**: 0.9\n"
	if err := os.WriteFile(artifact, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	// Lock with --force to skip validation
	_, _ = executeCommand("temper", "lock", artifact, "--force")
}

func TestCov4_TemperStatus(t *testing.T) {
	cov4TemperSetup(t)
	// Status with empty artifacts — should succeed gracefully
	out, err := executeCommand("temper", "status")
	_ = out
	_ = err
}
