package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- handleDryRunPhase ---

func TestHandleDryRunPhase_DryRunActive(t *testing.T) {
	origDryRun := dryRun
	dryRun = true
	defer func() { dryRun = origDryRun }()

	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "rpi.log")
	state := newTestPhasedState().WithRunID("dry-test")
	p := phases[0]
	opts := phasedEngineOptions{}

	result := handleDryRunPhase(tmp, state, 1, p, opts, "test prompt", logPath)
	if !result {
		t.Error("handleDryRunPhase should return true in dry-run mode")
	}

	// Verify log was written.
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}
	if !strings.Contains(string(data), "dry-run") {
		t.Errorf("log should contain 'dry-run', got: %s", string(data))
	}
}

func TestHandleDryRunPhase_DryRunInactive(t *testing.T) {
	origDryRun := dryRun
	dryRun = false
	defer func() { dryRun = origDryRun }()

	state := newTestPhasedState()
	p := phases[0]
	opts := phasedEngineOptions{}

	result := handleDryRunPhase(t.TempDir(), state, 1, p, opts, "test prompt", "")
	if result {
		t.Error("handleDryRunPhase should return false when dry-run is off")
	}
}

// --- maybeLiveStatus ---

func TestMaybeLiveStatus_Enabled(t *testing.T) {
	tmp := t.TempDir()
	statusPath := filepath.Join(tmp, "status.md")
	allPhases := buildAllPhases(phases)
	opts := phasedEngineOptions{LiveStatus: true}

	// Should not panic and should create/update the status file.
	maybeLiveStatus(opts, statusPath, allPhases, 1, "running", 0, "")

	if _, err := os.Stat(statusPath); os.IsNotExist(err) {
		t.Error("status file should be created when LiveStatus is enabled")
	}
}

func TestMaybeLiveStatus_Disabled(t *testing.T) {
	tmp := t.TempDir()
	statusPath := filepath.Join(tmp, "status.md")
	opts := phasedEngineOptions{LiveStatus: false}

	maybeLiveStatus(opts, statusPath, nil, 1, "running", 0, "")

	if _, err := os.Stat(statusPath); !os.IsNotExist(err) {
		t.Error("status file should NOT be created when LiveStatus is disabled")
	}
}

// --- writeFinalPhasedReport ---

func TestWriteFinalPhasedReport_BasicOutput(t *testing.T) {
	state := newTestPhasedState().WithGoal("test goal").WithEpicID("ag-99").
		WithVerdicts(map[string]string{"pre_mortem": "PASS", "vibe": "PASS"})

	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "rpi.log")

	// Capture stdout — we just verify no panic.
	writeFinalPhasedReport(state, logPath)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}
	if !strings.Contains(string(data), "complete") {
		t.Errorf("log should contain 'complete', got: %s", string(data))
	}
}

func TestWriteFinalPhasedReport_PlanFileEpic(t *testing.T) {
	state := newTestPhasedState().WithGoal("test").WithEpicID("plan:.agents/plans/feature.md")

	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "rpi.log")

	// Should not panic with plan-file epic.
	writeFinalPhasedReport(state, logPath)
}

func TestWriteFinalPhasedReport_EmptyEpicID(t *testing.T) {
	state := newTestPhasedState().WithGoal("test")

	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "rpi.log")

	// Should not panic with empty epic ID.
	writeFinalPhasedReport(state, logPath)
}

// --- logAndFailPhase ---

func TestLogAndFailPhase_SetsTerminalFields(t *testing.T) {
	tmp := t.TempDir()
	stateDir := filepath.Join(tmp, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(stateDir, "rpi.log")

	state := newTestPhasedState().WithRunID("fail-test")

	origErr := logAndFailPhase(state, "implementation", logPath, tmp, errFakeExecFailure)
	if origErr == nil {
		t.Fatal("logAndFailPhase should return the original error")
	}
	if state.TerminalStatus != "failed" {
		t.Errorf("TerminalStatus = %q, want %q", state.TerminalStatus, "failed")
	}
	if state.TerminalReason == "" {
		t.Error("TerminalReason should be set")
	}
	if !strings.Contains(state.TerminalReason, "implementation") {
		t.Errorf("TerminalReason = %q, should mention phase name", state.TerminalReason)
	}
	if state.TerminatedAt == "" {
		t.Error("TerminatedAt should be set")
	}
}

// --- runPhaseLoop ---

func TestRunPhaseLoop_FastPathSkipsPhase3(t *testing.T) {
	tmp := t.TempDir()
	stateDir := filepath.Join(tmp, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(stateDir, "rpi.log")
	statusPath := filepath.Join(stateDir, "status.md")

	origDryRun := dryRun
	dryRun = true
	defer func() { dryRun = origDryRun }()

	state := newTestPhasedState().
		WithGoal("trivial fix").
		WithRunID("fast-test").
		WithFastPath(true).
		WithEpicID("ag-fast-1")
	state.Complexity = ComplexityFast

	allPhases := buildAllPhases(phases)
	opts := phasedEngineOptions{NoWorktree: true}
	executor := &fakeExecutor{}

	// Run from phase 1 through all phases in dry-run mode.
	err := runPhaseLoop(tmp, tmp, state, 1, opts, statusPath, allPhases, logPath, executor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify phase 3 skip was logged.
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log: %v", err)
	}
	if !strings.Contains(string(data), "skipped") {
		t.Errorf("log should contain 'skipped' for phase 3, got: %s", string(data))
	}
}
