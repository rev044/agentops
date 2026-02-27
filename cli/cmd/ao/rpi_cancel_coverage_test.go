package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// --- runRPICancel tests ---

func TestCov3_rpiCancel_runRPICancel_noFlagsError(t *testing.T) {
	// Neither --all nor --run-id set
	oldAll := rpiCancelAll
	oldRunID := rpiCancelRunID
	rpiCancelAll = false
	rpiCancelRunID = ""
	defer func() {
		rpiCancelAll = oldAll
		rpiCancelRunID = oldRunID
	}()

	cmd := maturityCmd // reuse any cobra command; RunE is not called via cobra
	err := runRPICancel(cmd, nil)
	if err == nil {
		t.Fatal("expected error when neither --all nor --run-id set")
	}
	if !strings.Contains(err.Error(), "specify --all or --run-id") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCov3_rpiCancel_runRPICancel_badSignal(t *testing.T) {
	oldAll := rpiCancelAll
	oldSignal := rpiCancelSignal
	rpiCancelAll = true
	rpiCancelSignal = "BOGUS"
	defer func() {
		rpiCancelAll = oldAll
		rpiCancelSignal = oldSignal
	}()

	cmd := maturityCmd
	err := runRPICancel(cmd, nil)
	if err == nil {
		t.Fatal("expected error for bad signal")
	}
	if !strings.Contains(err.Error(), "unsupported signal") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- executeCancelTargets tests ---

func TestCov3_rpiCancel_executeCancelTargets_noTargets(t *testing.T) {
	failures := executeCancelTargets(nil, syscall.SIGTERM)
	if len(failures) != 0 {
		t.Fatalf("expected no failures for empty targets, got: %v", failures)
	}
}

func TestCov3_rpiCancel_executeCancelTargets_dryRun(t *testing.T) {
	oldDryRun := rpiCancelDryRun
	rpiCancelDryRun = true
	defer func() { rpiCancelDryRun = oldDryRun }()

	targets := []cancelTarget{
		{
			Kind:  "phased",
			RunID: "test-dry",
			PIDs:  []int{99999},
		},
	}

	captureJSONStdout(t, func() {
		failures := executeCancelTargets(targets, syscall.SIGTERM)
		if len(failures) != 0 {
			t.Fatalf("expected no failures in dry-run, got: %v", failures)
		}
	})
}

// --- cancelOneTarget tests ---

func TestCov3_rpiCancel_cancelOneTarget_dryRun(t *testing.T) {
	oldDryRun := rpiCancelDryRun
	rpiCancelDryRun = true
	defer func() { rpiCancelDryRun = oldDryRun }()

	target := cancelTarget{
		Kind:  "supervisor",
		RunID: "dry-target",
		PIDs:  []int{99999},
	}

	captureJSONStdout(t, func() {
		failures := cancelOneTarget(target, syscall.SIGTERM, os.Getpid())
		if len(failures) != 0 {
			t.Fatalf("expected no failures in dry-run, got: %v", failures)
		}
	})
}

func TestCov3_rpiCancel_cancelOneTarget_nonexistentPID(t *testing.T) {
	oldDryRun := rpiCancelDryRun
	rpiCancelDryRun = false
	defer func() { rpiCancelDryRun = oldDryRun }()

	// PID 99999999 should not exist; ESRCH error is silently ignored
	target := cancelTarget{
		Kind:  "phased",
		RunID: "no-pid",
		PIDs:  []int{99999999},
	}

	captureJSONStdout(t, func() {
		failures := cancelOneTarget(target, syscall.SIGTERM, os.Getpid())
		// ESRCH is ignored, so no failures
		if len(failures) != 0 {
			t.Fatalf("expected ESRCH to be ignored, got failures: %v", failures)
		}
	})
}

func TestCov3_rpiCancel_cancelOneTarget_withStateUpdate(t *testing.T) {
	oldDryRun := rpiCancelDryRun
	rpiCancelDryRun = false
	defer func() { rpiCancelDryRun = oldDryRun }()

	root := t.TempDir()
	runID := "cancel-state-test"
	statePath := filepath.Join(root, ".agents", "rpi", "runs", runID, phasedStateFile)
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		t.Fatal(err)
	}

	state := map[string]any{
		"schema_version": 1,
		"run_id":         runID,
		"goal":           "test",
		"phase":          1,
	}
	data, _ := json.MarshalIndent(state, "", "  ")
	data = append(data, '\n')
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		t.Fatal(err)
	}

	target := cancelTarget{
		Kind:      "phased",
		RunID:     runID,
		Root:      root,
		StatePath: statePath,
		PIDs:      []int{99999999}, // nonexistent, ESRCH ignored
	}

	captureJSONStdout(t, func() {
		failures := cancelOneTarget(target, syscall.SIGTERM, os.Getpid())
		if len(failures) != 0 {
			t.Fatalf("unexpected failures: %v", failures)
		}
	})

	// Verify state was updated
	updated, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(updated, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["terminal_status"] != "interrupted" {
		t.Fatalf("expected terminal_status=interrupted, got %v", raw["terminal_status"])
	}
}

// --- discoverCancelTargets tests ---

func TestCov3_rpiCancel_discoverCancelTargets_emptyRoots(t *testing.T) {
	targets := discoverCancelTargets(nil, "", nil)
	if len(targets) != 0 {
		t.Fatalf("expected no targets for nil roots, got %d", len(targets))
	}
}

func TestCov3_rpiCancel_discoverCancelTargets_sortOrder(t *testing.T) {
	// Create two roots, one with a supervisor lease and one with a run registry
	root1 := t.TempDir()
	root2 := t.TempDir()

	// Root1: supervisor lease with active process (self PID)
	leaseDir1 := filepath.Join(root1, ".agents", "rpi")
	if err := os.MkdirAll(leaseDir1, 0755); err != nil {
		t.Fatal(err)
	}
	meta1 := supervisorLeaseMetadata{
		RunID:     "z-lease",
		PID:       os.Getpid(),
		ExpiresAt: time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339),
	}
	data1, _ := json.Marshal(meta1)
	if err := os.WriteFile(filepath.Join(leaseDir1, "supervisor.lock"), data1, 0644); err != nil {
		t.Fatal(err)
	}

	procs := []processInfo{
		{PID: os.Getpid(), PPID: 1, Command: "test-process"},
	}

	targets := discoverCancelTargets([]string{root1, root2}, "", procs)
	// Should find at least the lease target
	for i := 1; i < len(targets); i++ {
		if targets[i-1].Kind > targets[i].Kind {
			t.Fatalf("targets not sorted by kind: %s > %s", targets[i-1].Kind, targets[i].Kind)
		}
	}
}

// --- loadActiveSupervisorLease tests ---

func TestCov3_rpiCancel_loadActiveSupervisorLease_missingFile(t *testing.T) {
	_, ok := loadActiveSupervisorLease("/nonexistent/path", "", nil)
	if ok {
		t.Fatal("expected false for missing file")
	}
}

func TestCov3_rpiCancel_loadActiveSupervisorLease_emptyRunID(t *testing.T) {
	tmp := t.TempDir()
	leasePath := filepath.Join(tmp, "supervisor.lock")
	meta := supervisorLeaseMetadata{
		RunID: "", // empty
		PID:   100,
	}
	data, _ := json.Marshal(meta)
	if err := os.WriteFile(leasePath, data, 0644); err != nil {
		t.Fatal(err)
	}
	_, ok := loadActiveSupervisorLease(leasePath, "", nil)
	if ok {
		t.Fatal("expected false for empty run ID")
	}
}

func TestCov3_rpiCancel_loadActiveSupervisorLease_runIDMismatch(t *testing.T) {
	tmp := t.TempDir()
	leasePath := filepath.Join(tmp, "supervisor.lock")
	meta := supervisorLeaseMetadata{
		RunID:     "run-A",
		PID:       os.Getpid(),
		ExpiresAt: time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(meta)
	if err := os.WriteFile(leasePath, data, 0644); err != nil {
		t.Fatal(err)
	}

	procs := []processInfo{{PID: os.Getpid(), PPID: 1, Command: "test"}}
	_, ok := loadActiveSupervisorLease(leasePath, "run-B", procs)
	if ok {
		t.Fatal("expected false for run ID mismatch")
	}
}

func TestCov3_rpiCancel_loadActiveSupervisorLease_processNotFound(t *testing.T) {
	tmp := t.TempDir()
	leasePath := filepath.Join(tmp, "supervisor.lock")
	meta := supervisorLeaseMetadata{
		RunID:     "run-C",
		PID:       99999999,
		ExpiresAt: time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(meta)
	if err := os.WriteFile(leasePath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Empty proc list: PID won't be found
	_, ok := loadActiveSupervisorLease(leasePath, "", nil)
	if ok {
		t.Fatal("expected false when process not in list")
	}
}
