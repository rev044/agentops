package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRPIStream_JSONAndSSE(t *testing.T) {
	tmp := t.TempDir()
	runID := "stream-run-1"
	writeTestRunState(t, tmp, runID, 2)

	if _, err := appendRPIC2Event(tmp, rpiC2EventInput{
		RunID:    runID,
		Phase:    2,
		Backend:  "stream",
		Source:   "runtime_stream",
		Type:     "phase.stream.started",
		Message:  "started",
		WorkerID: "1",
	}); err != nil {
		t.Fatalf("appendRPIC2Event: %v", err)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldCwd)
	})

	jsonOut, err := executeCommand("rpi", "stream", "--run-id", runID, "--format", "json")
	if err != nil {
		t.Fatalf("rpi stream json: %v", err)
	}
	if !strings.Contains(jsonOut, `"type":"phase.stream.started"`) {
		t.Fatalf("json output missing event type: %s", jsonOut)
	}

	sseOut, err := executeCommand("rpi", "stream", "--run-id", runID, "--format", "sse")
	if err != nil {
		t.Fatalf("rpi stream sse: %v", err)
	}
	if !strings.Contains(sseOut, "event: phase.stream.started") {
		t.Fatalf("sse output missing event line: %s", sseOut)
	}
}

func TestRPIStream_InvalidFormat(t *testing.T) {
	tmp := t.TempDir()
	runID := "stream-run-invalid"
	writeTestRunState(t, tmp, runID, 1)

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldCwd)
	})

	_, err = executeCommand("rpi", "stream", "--run-id", runID, "--format", "yaml")
	if err == nil {
		t.Fatal("expected invalid format error")
	}
}

func TestRPIWorkers_JSONProjection(t *testing.T) {
	tmp := t.TempDir()
	runID := "workers-run-1"
	writeTestRunState(t, tmp, runID, 2)
	updateRunHeartbeat(tmp, runID)

	now := time.Now().UTC()
	if _, err := appendRPIC2Event(tmp, rpiC2EventInput{
		RunID:     runID,
		Phase:     2,
		Backend:   "tmux",
		Source:    "tmux_worker_log",
		WorkerID:  "1",
		Type:      "worker.rpi_worker_start",
		Message:   "worker started",
		Timestamp: now,
	}); err != nil {
		t.Fatalf("append worker 1 event: %v", err)
	}
	if _, err := appendRPIC2Event(tmp, rpiC2EventInput{
		RunID:     runID,
		Phase:     2,
		Backend:   "tmux",
		Source:    "tmux_worker_log",
		WorkerID:  "2",
		Type:      "worker.rpi_worker_start",
		Message:   "worker started",
		Timestamp: now.Add(-2 * time.Hour),
	}); err != nil {
		t.Fatalf("append worker 2 event: %v", err)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldCwd)
	})

	raw, err := executeCommand("rpi", "workers", "--run-id", runID, "--json")
	if err != nil {
		t.Fatalf("rpi workers --json: %v", err)
	}
	var output rpiWorkersOutput
	if err := json.Unmarshal([]byte(raw), &output); err != nil {
		t.Fatalf("unmarshal workers output: %v\nraw:\n%s", err, raw)
	}
	if len(output.Workers) != 2 {
		t.Fatalf("workers len = %d, want 2", len(output.Workers))
	}
	if output.Workers[0].WorkerID != "1" || output.Workers[0].Health != "healthy" {
		t.Fatalf("worker 1 unexpected: %#v", output.Workers[0])
	}
	if output.Workers[1].WorkerID != "2" || output.Workers[1].Health != "stale" {
		t.Fatalf("worker 2 unexpected: %#v", output.Workers[1])
	}
}

func writeTestRunState(t *testing.T, root, runID string, phase int) {
	t.Helper()
	runDir := rpiRunRegistryDir(root, runID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatalf("mkdir run dir: %v", err)
	}
	state := phasedState{
		SchemaVersion: 1,
		Goal:          "test run",
		RunID:         runID,
		Phase:         phase,
		StartedAt:     time.Now().UTC().Format(time.RFC3339Nano),
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(runDir, phasedStateFile), data, 0o644); err != nil {
		t.Fatalf("write state: %v", err)
	}
}
