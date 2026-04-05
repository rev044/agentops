package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// gcEventType constants for ao events in the gc event bus.
const (
	GCEventAOPhase   = "ao:phase"
	GCEventAOGate    = "ao:gate"
	GCEventAOFailure = "ao:failure"
	GCEventAOMetric  = "ao:metric"
)

// gcEmitPhaseEvent emits an ao:phase event to the gc event bus.
func gcEmitPhaseEvent(cityPath string, phase int, status, runID string) error {
	data := map[string]any{
		"phase":     phase,
		"status":    status,
		"run_id":    runID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return gcEmitEvent(cityPath, GCEventAOPhase, data)
}

// gcEmitGateEvent emits an ao:gate event (pre-mortem, vibe, etc.) to the gc event bus.
func gcEmitGateEvent(cityPath string, gate, verdict, runID string) error {
	data := map[string]any{
		"gate":      gate,
		"verdict":   verdict,
		"run_id":    runID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return gcEmitEvent(cityPath, GCEventAOGate, data)
}

// gcEmitFailureEvent emits an ao:failure event to the gc event bus.
func gcEmitFailureEvent(cityPath string, errMsg, runID string, phase int) error {
	data := map[string]any{
		"error":     errMsg,
		"phase":     phase,
		"run_id":    runID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return gcEmitEvent(cityPath, GCEventAOFailure, data)
}

// gcEmitMetricEvent emits an ao:metric event to the gc event bus.
func gcEmitMetricEvent(cityPath string, metric string, value float64, runID string) error {
	data := map[string]any{
		"metric":    metric,
		"value":     value,
		"run_id":    runID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return gcEmitEvent(cityPath, GCEventAOMetric, data)
}

// gcEmitEvent is the low-level event emitter to the gc event bus.
func gcEmitEvent(cityPath, eventType string, data map[string]any) error {
	if !gcBridgeAvailable() {
		return nil // silently skip if gc not available
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal event data: %w", err)
	}
	args := gcEventEmitArgs(eventType, string(dataJSON))
	if cityPath != "" {
		args = append([]string{"--city", cityPath}, args...)
	}
	cmd := gcExecCommand("gc", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gc event emit %s: %w (output: %s)", eventType, err, string(out))
	}
	return nil
}

// gcEventsAvailable returns true if gc event system is accessible.
func gcEventsAvailable(cityPath string) bool {
	if !gcBridgeAvailable() {
		return false
	}
	args := []string{"events", "--help"}
	if cityPath != "" {
		args = append([]string{"--city", cityPath}, args...)
	}
	return gcExecCommand("gc", args...).Run() == nil
}
