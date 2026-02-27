package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRPIC2EventAppendAndLoad(t *testing.T) {
	root := t.TempDir()
	runID := "run-c2-001"

	if _, err := appendRPIC2Event(root, rpiC2EventInput{
		RunID:   runID,
		Phase:   2,
		Backend: "stream",
		Source:  "runtime_stream",
		Type:    "phase.stream.started",
		Message: "started",
		Details: map[string]any{"attempt": 1},
	}); err != nil {
		t.Fatalf("appendRPIC2Event: %v", err)
	}

	events, err := loadRPIC2Events(root, runID)
	if err != nil {
		t.Fatalf("loadRPIC2Events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].Type != "phase.stream.started" {
		t.Fatalf("event type = %q", events[0].Type)
	}
	if events[0].RunID != runID {
		t.Fatalf("run_id = %q", events[0].RunID)
	}
}

func TestRPIC2Event_RequiresRunAndType(t *testing.T) {
	root := t.TempDir()
	if _, err := appendRPIC2Event(root, rpiC2EventInput{Type: "x"}); err == nil {
		t.Fatal("expected error for missing run_id")
	}
	if _, err := appendRPIC2Event(root, rpiC2EventInput{RunID: "run"}); err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestAppendRPIC2WorkerLogEvents(t *testing.T) {
	root := t.TempDir()
	runID := "run-c2-worker"
	runDir := rpiRunRegistryDir(root, runID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		t.Fatalf("mkdir run dir: %v", err)
	}
	logPath := filepath.Join(runDir, "phase-1-exit.w1.jsonl")
	payload := "{\"type\":\"rpi_worker_start\",\"worker\":\"1\"}\nplain line\n"
	if err := os.WriteFile(logPath, []byte(payload), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}

	if err := appendRPIC2WorkerLogEvents(root, runID, 1, "tmux", "1", logPath); err != nil {
		t.Fatalf("appendRPIC2WorkerLogEvents: %v", err)
	}

	events, err := loadRPIC2Events(root, runID)
	if err != nil {
		t.Fatalf("loadRPIC2Events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if events[0].Type != "worker.rpi_worker_start" {
		t.Fatalf("first event type = %q", events[0].Type)
	}
	if events[1].Type != "worker.log" {
		t.Fatalf("second event type = %q", events[1].Type)
	}
}

func TestMapStreamEventToRPIC2(t *testing.T) {
	input := mapStreamEventToRPIC2("run-1", 3, StreamEvent{
		Type:      EventTypeAssistant,
		Subtype:   "tool_use",
		ToolName:  "Read",
		Message:   "hello",
		SessionID: "sess-1",
	})
	if input.Type != "stream.assistant" {
		t.Fatalf("type = %q", input.Type)
	}
	if input.RunID != "run-1" {
		t.Fatalf("run_id = %q", input.RunID)
	}
	if input.Backend != "stream" {
		t.Fatalf("backend = %q", input.Backend)
	}
}
