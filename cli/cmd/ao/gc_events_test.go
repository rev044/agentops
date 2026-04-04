package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGCEmitEvent_NoGC(t *testing.T) {
	// When gc is not available, gcEmitEvent should silently return nil
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", t.TempDir())
	defer os.Setenv("PATH", origPath)

	err := gcEmitEvent("", "ao:test", map[string]any{"test": true})
	if err != nil {
		t.Errorf("gcEmitEvent should return nil when gc not available, got: %v", err)
	}
}

func TestGCEmitPhaseEvent_ArgsFormat(t *testing.T) {
	// Test that phase event data is properly structured
	data := map[string]any{
		"phase":  1,
		"status": "complete",
		"run_id": "test-123",
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	args := gcEventEmitArgs(GCEventAOPhase, string(dataJSON))
	if args[0] != "event" || args[1] != "emit" || args[2] != "ao:phase" {
		t.Errorf("unexpected args prefix: %v", args[:3])
	}
}

func TestGCEmitGateEvent_ArgsFormat(t *testing.T) {
	data := map[string]any{
		"gate":    "pre-mortem",
		"verdict": "PASS",
		"run_id":  "test-456",
	}
	dataJSON, _ := json.Marshal(data)
	args := gcEventEmitArgs(GCEventAOGate, string(dataJSON))
	if args[2] != "ao:gate" {
		t.Errorf("event type = %q, want %q", args[2], "ao:gate")
	}
}

func TestGCEmitFailureEvent_ArgsFormat(t *testing.T) {
	data := map[string]any{
		"error":  "phase timeout",
		"phase":  2,
		"run_id": "test-789",
	}
	dataJSON, _ := json.Marshal(data)
	args := gcEventEmitArgs(GCEventAOFailure, string(dataJSON))
	if args[2] != "ao:failure" {
		t.Errorf("event type = %q, want %q", args[2], "ao:failure")
	}
}

func TestGCEventsAvailable_NoGC(t *testing.T) {
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", t.TempDir())
	defer os.Setenv("PATH", origPath)

	if gcEventsAvailable("") {
		t.Error("gcEventsAvailable should return false when gc not available")
	}
}

func TestGCEventConstants(t *testing.T) {
	if GCEventAOPhase != "ao:phase" {
		t.Errorf("GCEventAOPhase = %q, want %q", GCEventAOPhase, "ao:phase")
	}
	if GCEventAOGate != "ao:gate" {
		t.Errorf("GCEventAOGate = %q, want %q", GCEventAOGate, "ao:gate")
	}
	if GCEventAOFailure != "ao:failure" {
		t.Errorf("GCEventAOFailure = %q, want %q", GCEventAOFailure, "ao:failure")
	}
	if GCEventAOMetric != "ao:metric" {
		t.Errorf("GCEventAOMetric = %q, want %q", GCEventAOMetric, "ao:metric")
	}
}
