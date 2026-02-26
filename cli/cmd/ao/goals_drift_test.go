package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/boshu2/agentops/cli/internal/goals"
)

func TestGoalsDrift_NoBaseline_CreatesInitialSnapshot(t *testing.T) {
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

	// Set up working directory and flag state
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldJSON := goalsJSON
	oldTimeout := goalsTimeout
	defer func() {
		goalsFile = oldFile
		goalsJSON = oldJSON
		goalsTimeout = oldTimeout
	}()
	goalsFile = goalsPath
	goalsJSON = false
	goalsTimeout = 10

	// Run drift with no existing snapshots
	err := goalsDriftCmd.RunE(goalsDriftCmd, nil)
	if err != nil {
		t.Fatalf("drift returned error: %v", err)
	}

	// Verify a snapshot was created
	snapDir := filepath.Join(dir, ".agents/ao/goals/baselines")
	entries, err := os.ReadDir(snapDir)
	if err != nil {
		t.Fatalf("could not read snapshot dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected at least one snapshot file to be created")
	}
}

func TestGoalsDrift_WithBaseline_ComparesSnapshots(t *testing.T) {
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

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldJSON := goalsJSON
	oldTimeout := goalsTimeout
	defer func() {
		goalsFile = oldFile
		goalsJSON = oldJSON
		goalsTimeout = oldTimeout
	}()
	goalsFile = goalsPath
	goalsJSON = false
	goalsTimeout = 10

	// Run drift — should compare against the baseline
	err := goalsDriftCmd.RunE(goalsDriftCmd, nil)
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

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldJSON := goalsJSON
	oldTimeout := goalsTimeout
	defer func() {
		goalsFile = oldFile
		goalsJSON = oldJSON
		goalsTimeout = oldTimeout
	}()
	goalsFile = goalsPath
	goalsJSON = true
	goalsTimeout = 10

	// Redirect stdout to capture JSON
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := goalsDriftCmd.RunE(goalsDriftCmd, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("drift returned error: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)

	var drifts []goals.DriftResult
	if err := json.Unmarshal(buf[:n], &drifts); err != nil {
		t.Fatalf("failed to decode JSON output: %v (raw: %s)", err, string(buf[:n]))
	}
}

func TestGoalsDrift_MissingGoalsFile(t *testing.T) {
	oldFile := goalsFile
	defer func() { goalsFile = oldFile }()

	goalsFile = "/nonexistent/GOALS.md"

	err := goalsDriftCmd.RunE(goalsDriftCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing goals file")
	}
}
