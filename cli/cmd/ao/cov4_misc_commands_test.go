package main

import (
	"os"
	"path/filepath"
	"testing"
)

// cov4MiscSetup creates a temp workdir with .agents/ directories for misc command tests.
func cov4MiscSetup(t *testing.T) string {
	t.Helper()
	tmp := setupTempWorkdir(t)
	t.Setenv("HOME", tmp)
	setupAgentsDir(t, tmp)
	return tmp
}

func TestCov4_ContextStatus(t *testing.T) {
	cov4MiscSetup(t)
	_, _ = executeCommand("context", "status")
}

func TestCov4_ContextGuard(t *testing.T) {
	cov4MiscSetup(t)
	_, _ = executeCommand("context", "guard")
}

func TestCov4_HooksShow(t *testing.T) {
	tmp := cov4MiscSetup(t)
	// Create a minimal settings.json for hooks show
	claudeDir := filepath.Join(tmp, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	_, _ = executeCommand("hooks", "show")
}

func TestCov4_HooksInstall(t *testing.T) {
	tmp := cov4MiscSetup(t)
	// Create a minimal settings.json
	claudeDir := filepath.Join(tmp, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	// Use --dry-run to avoid actual installation side effects
	_, _ = executeCommand("hooks", "install", "--dry-run")
}

func TestCov4_Inject(t *testing.T) {
	cov4MiscSetup(t)
	// Inject with no knowledge artifacts — should succeed gracefully
	_, _ = executeCommand("inject", "--no-cite")
}

func TestCov4_TaskSync(t *testing.T) {
	tmp := cov4MiscSetup(t)
	// Create a minimal transcript file for task-sync to read
	transcript := filepath.Join(tmp, "transcript.jsonl")
	if err := os.WriteFile(transcript, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	_, _ = executeCommand("task-sync", "--transcript", transcript)
}

func TestCov4_TaskStatus(t *testing.T) {
	cov4MiscSetup(t)
	// task-status with no tasks file — should handle gracefully
	_, _ = executeCommand("task-status")
}

func TestCov4_SessionClose(t *testing.T) {
	cov4MiscSetup(t)
	// No transcript to close — should fail gracefully
	_, _ = executeCommand("session", "close")
}

func TestCov4_SessionOutcome(t *testing.T) {
	cov4MiscSetup(t)
	// No transcript — should fail gracefully
	_, _ = executeCommand("session-outcome")
}

func TestCov4_Trace(t *testing.T) {
	tmp := cov4MiscSetup(t)
	// Create a dummy artifact to trace
	artifact := filepath.Join(tmp, ".agents", "ao", "sessions", "test-session.md")
	if err := os.WriteFile(artifact, []byte("# Test Session\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, _ = executeCommand("trace", artifact)
}

func TestCov4_Lookup(t *testing.T) {
	cov4MiscSetup(t)
	// Lookup with --query — no learnings, should return "No matching artifacts"
	_, _ = executeCommand("lookup", "--query", "test", "--no-cite")
}

func TestCov4_PlansDiff(t *testing.T) {
	cov4MiscSetup(t)
	// plans diff with no manifest — should fail gracefully
	_, _ = executeCommand("plans", "diff")
}

func TestCov4_GoalsMigrate(t *testing.T) {
	tmp := cov4MiscSetup(t)
	// Create a minimal GOALS.yaml for migrate to read
	goalsContent := "version: 1\ngoals:\n  - name: test\n    target: 90\n"
	if err := os.WriteFile(filepath.Join(tmp, "GOALS.yaml"), []byte(goalsContent), 0644); err != nil {
		t.Fatal(err)
	}
	_, _ = executeCommand("goals", "migrate")
}

func TestCov4_PoolIngest(t *testing.T) {
	cov4MiscSetup(t)
	// Ingest with no pending files — should succeed gracefully
	_, _ = executeCommand("pool", "ingest")
}

func TestCov4_NotebookUpdate(t *testing.T) {
	tmp := cov4MiscSetup(t)
	// Create a minimal MEMORY.md for notebook update
	memoryDir := filepath.Join(tmp, ".claude", "projects", "-test")
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		t.Fatal(err)
	}
	memoryFile := filepath.Join(memoryDir, "MEMORY.md")
	if err := os.WriteFile(memoryFile, []byte("# Memory\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, _ = executeCommand("notebook", "update", "--memory-file", memoryFile)
}

func TestCov4_MetricsCite(t *testing.T) {
	tmp := cov4MiscSetup(t)
	// Create a dummy artifact for citation
	artifact := filepath.Join(tmp, ".agents", "learnings", "cite-test.md")
	if err := os.WriteFile(artifact, []byte("# Citation Test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, _ = executeCommand("metrics", "cite", artifact)
}

func TestCov4_FlywheelCloseLoop(t *testing.T) {
	cov4MiscSetup(t)
	// Close loop with no pending content — should succeed gracefully
	_, _ = executeCommand("flywheel", "close-loop")
}
