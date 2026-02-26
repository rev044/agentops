package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/boshu2/agentops/cli/internal/goals"
)

func TestGoalsExport_WithExistingSnapshot(t *testing.T) {
	dir := t.TempDir()

	// Create GOALS.md
	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| test-gate | ` + "`exit 0`" + ` | 5 | Test |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a snapshot
	snapDir := filepath.Join(dir, ".agents/ao/goals/baselines")
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		t.Fatal(err)
	}
	snap := &goals.Snapshot{
		Timestamp: "2025-06-01T12:00:00Z",
		GitSHA:    "abc1234",
		Goals: []goals.Measurement{
			{GoalID: "test-gate", Result: "pass", Weight: 5, Duration: 0.1},
		},
		Summary: goals.SnapshotSummary{Total: 1, Passing: 1, Score: 100.0},
	}
	data, _ := json.MarshalIndent(snap, "", "  ")
	if err := os.WriteFile(filepath.Join(snapDir, "2025-06-01T12-00-00.000.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldTimeout := goalsTimeout
	defer func() {
		goalsFile = oldFile
		goalsTimeout = oldTimeout
	}()
	goalsFile = goalsPath
	goalsTimeout = 10

	// Capture stdout
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := goalsExportCmd.RunE(goalsExportCmd, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("export returned error: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)

	var exported goals.Snapshot
	if err := json.Unmarshal(buf[:n], &exported); err != nil {
		t.Fatalf("failed to decode JSON output: %v (raw: %s)", err, string(buf[:n]))
	}
	if exported.Timestamp != "2025-06-01T12:00:00Z" {
		t.Errorf("Timestamp = %q, want 2025-06-01T12:00:00Z", exported.Timestamp)
	}
	if exported.Summary.Score != 100.0 {
		t.Errorf("Score = %f, want 100.0", exported.Summary.Score)
	}
}

func TestGoalsExport_NoSnapshot_MeasuresFresh(t *testing.T) {
	dir := t.TempDir()

	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| quick-gate | ` + "`exit 0`" + ` | 5 | Quick test |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldTimeout := goalsTimeout
	defer func() {
		goalsFile = oldFile
		goalsTimeout = oldTimeout
	}()
	goalsFile = goalsPath
	goalsTimeout = 10

	// Capture stdout
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := goalsExportCmd.RunE(goalsExportCmd, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("export returned error: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)

	var exported goals.Snapshot
	if err := json.Unmarshal(buf[:n], &exported); err != nil {
		t.Fatalf("failed to decode JSON output: %v (raw: %s)", err, string(buf[:n]))
	}
	if exported.Summary.Total != 1 {
		t.Errorf("Total = %d, want 1", exported.Summary.Total)
	}
}

func TestGoalsExport_MissingGoalsFileAndNoSnapshots(t *testing.T) {
	dir := t.TempDir()

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	defer func() { goalsFile = oldFile }()
	goalsFile = filepath.Join(dir, "GOALS.md") // does not exist

	err := goalsExportCmd.RunE(goalsExportCmd, nil)
	if err == nil {
		t.Fatal("expected error when both snapshots and goals file are missing")
	}
}

func TestGoalsExport_CmdAttributes(t *testing.T) {
	if goalsExportCmd.Use != "export" {
		t.Errorf("Use = %q, want export", goalsExportCmd.Use)
	}
	if goalsExportCmd.GroupID != "analysis" {
		t.Errorf("GroupID = %q, want analysis", goalsExportCmd.GroupID)
	}
	found := false
	for _, a := range goalsExportCmd.Aliases {
		if a == "e" {
			found = true
		}
	}
	if !found {
		t.Error("expected alias 'e' for export command")
	}
}
