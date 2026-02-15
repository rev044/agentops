package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRPIStatusDiscovery(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}

	state := map[string]interface{}{
		"run_id":        "abc123def456",
		"goal":          "test goal",
		"current_phase": 3,
		"epic_id":       "ag-test",
		"started_at":    time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
	}
	data, _ := json.Marshal(state)
	if err := os.WriteFile(filepath.Join(stateDir, "phased-state.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	run, ok := loadRPIRun(tmpDir, "")
	if !ok {
		t.Fatal("expected loadRPIRun to return a run")
	}

	if run.RunID != "abc123def456" {
		t.Errorf("expected run_id abc123def456, got %s", run.RunID)
	}
	if run.PhaseName != "pre-mortem" {
		t.Errorf("expected phase pre-mortem, got %s", run.PhaseName)
	}
	if run.EpicID != "ag-test" {
		t.Errorf("expected epic ag-test, got %s", run.EpicID)
	}
	if !run.Active {
		t.Error("expected active=true for recently modified state")
	}
	if run.Goal != "test goal" {
		t.Errorf("expected goal 'test goal', got %s", run.Goal)
	}
}

func TestRPIStatusMissingState(t *testing.T) {
	tmpDir := t.TempDir()
	_, ok := loadRPIRun(tmpDir, "")
	if ok {
		t.Fatal("expected loadRPIRun to return false for empty dir")
	}
}

func TestRPIStatusCorruptState(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "phased-state.json"), []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, ok := loadRPIRun(tmpDir, "")
	if ok {
		t.Fatal("expected loadRPIRun to return false for corrupt state")
	}
}

func TestRPIStatusPhaseNames(t *testing.T) {
	tests := []struct {
		phase    int
		expected string
	}{
		{1, "research"},
		{2, "plan"},
		{3, "pre-mortem"},
		{4, "crank"},
		{5, "vibe"},
		{6, "post-mortem"},
		{99, "phase-99"},
	}

	for _, tt := range tests {
		tmpDir := t.TempDir()
		stateDir := filepath.Join(tmpDir, ".agents", "rpi")
		if err := os.MkdirAll(stateDir, 0755); err != nil {
			t.Fatal(err)
		}

		state := map[string]interface{}{
			"run_id":        "test-run",
			"current_phase": tt.phase,
		}
		data, _ := json.Marshal(state)
		if err := os.WriteFile(filepath.Join(stateDir, "phased-state.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		run, ok := loadRPIRun(tmpDir, "")
		if !ok {
			t.Fatalf("expected run for phase %d", tt.phase)
		}
		if run.PhaseName != tt.expected {
			t.Errorf("phase %d: expected %s, got %s", tt.phase, tt.expected, run.PhaseName)
		}
	}
}

func TestRPIStatusEmptyRunID(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}

	state := map[string]interface{}{
		"goal":          "no run id",
		"current_phase": 1,
	}
	data, _ := json.Marshal(state)
	if err := os.WriteFile(filepath.Join(stateDir, "phased-state.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	_, ok := loadRPIRun(tmpDir, "")
	if ok {
		t.Fatal("expected loadRPIRun to return false when run_id is empty")
	}
}
