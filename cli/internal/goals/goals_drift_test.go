package goals_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/goals"
)

func TestGoalsDrift_NoBaseline_CreatesInitialSnapshot(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create a simple GOALS.md with a passing gate
	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| pass-gate | ` + "`exit 0`" + ` | 5 | Always passes |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	snapDir := filepath.Join(dir, ".agents/ao/goals/baselines")

	var stdout, stderr bytes.Buffer
	err := goals.RunDrift(goals.DriftOptions{
		GoalsFile: goalsPath,
		Timeout:   10 * time.Second,
		JSON:      false,
		SnapDir:   snapDir,
		Stdout:    &stdout,
		Stderr:    &stderr,
	})
	if err != nil {
		t.Fatalf("drift returned error: %v", err)
	}

	// Verify a snapshot was created
	entries, err := os.ReadDir(snapDir)
	if err != nil {
		t.Fatalf("could not read snapshot dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected at least one snapshot file to be created")
	}
}

func TestGoalsDrift_WithBaseline_ComparesSnapshots(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create a GOALS.md with a passing gate
	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| pass-gate | ` + "`exit 0`" + ` | 5 | Always passes |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a baseline snapshot
	snapDir := filepath.Join(dir, ".agents/ao/goals/baselines")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatal(err)
	}

	baseline := &goals.Snapshot{
		Timestamp: "2025-01-01T00:00:00Z",
		Goals: []goals.Measurement{
			{GoalID: "pass-gate", Result: "pass", Weight: 5},
		},
		Summary: goals.SnapshotSummary{
			Total:   1,
			Passing: 1,
			Score:   100.0,
		},
	}
	data, _ := json.MarshalIndent(baseline, "", "  ")
	if err := os.WriteFile(filepath.Join(snapDir, "2025-01-01T00-00-00.000.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := goals.RunDrift(goals.DriftOptions{
		GoalsFile: goalsPath,
		Timeout:   10 * time.Second,
		JSON:      false,
		SnapDir:   snapDir,
		Stdout:    &stdout,
		Stderr:    &stderr,
	})
	if err != nil {
		t.Fatalf("drift returned error: %v", err)
	}

	// Should have created a new snapshot
	entries, err := os.ReadDir(snapDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 2 {
		t.Errorf("expected at least 2 snapshot files (baseline + current), got %d", len(entries))
	}
}

func TestGoalsDrift_JSONOutput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| test-gate | ` + "`exit 0`" + ` | 5 | Test gate |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create baseline
	snapDir := filepath.Join(dir, ".agents/ao/goals/baselines")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	baseline := &goals.Snapshot{
		Timestamp: "2025-01-01T00:00:00Z",
		Goals: []goals.Measurement{
			{GoalID: "test-gate", Result: "pass", Weight: 5},
		},
		Summary: goals.SnapshotSummary{Total: 1, Passing: 1, Score: 100.0},
	}
	data, _ := json.MarshalIndent(baseline, "", "  ")
	if err := os.WriteFile(filepath.Join(snapDir, "2025-01-01T00-00-00.000.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := goals.RunDrift(goals.DriftOptions{
		GoalsFile: goalsPath,
		Timeout:   10 * time.Second,
		JSON:      true,
		SnapDir:   snapDir,
		Stdout:    &stdout,
		Stderr:    &stderr,
	})
	if err != nil {
		t.Fatalf("drift returned error: %v", err)
	}

	var drifts []goals.DriftResult
	if err := json.Unmarshal(stdout.Bytes(), &drifts); err != nil {
		t.Fatalf("failed to decode JSON output: %v (raw: %s)", err, stdout.String())
	}
}

func TestGoalsDrift_MissingGoalsFile(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := goals.RunDrift(goals.DriftOptions{
		GoalsFile: "/nonexistent/GOALS.md",
		Timeout:   10 * time.Second,
		SnapDir:   "/nonexistent/snaps",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})
	if err == nil {
		t.Fatal("expected error for missing goals file")
	}
}
