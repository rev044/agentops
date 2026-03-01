package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServeRPIIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	serveRPIIndex(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html Content-Type, got %q", ct)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "mission control") {
		t.Errorf("expected 'mission control' in body")
	}
	if !strings.Contains(body, "EventSource") {
		t.Errorf("expected EventSource JS in body")
	}
}

func TestServeRPIRuns_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	req := httptest.NewRequest(http.MethodGet, "/runs", nil)
	rr := httptest.NewRecorder()
	serveRPIRuns(rr, req, dir)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected JSON Content-Type, got %q", ct)
	}
}

func TestServeRPIEvents_NoRun(t *testing.T) {
	dir := t.TempDir()

	// Use a cancellable context to simulate client disconnect.
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/events?run-id=nonexistent", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		serveRPIEvents(rr, req, dir, "nonexistent")
		close(done)
	}()

	// Give the handler one poll cycle, then cancel.
	time.Sleep(600 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("serveRPIEvents did not return after context cancel")
	}

	// SSE headers must be set even with no events.
	if ct := rr.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected text/event-stream, got %q", ct)
	}
}

func TestWriteSSEEvent(t *testing.T) {
	rr := httptest.NewRecorder()
	ev := RPIC2Event{
		SchemaVersion: 1,
		EventID:       "evt-abc",
		RunID:         "rpi-test",
		Type:          "phase.stream.started",
		Message:       "hello",
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
	}

	if err := writeSSEEvent(rr, ev); err != nil {
		t.Fatalf("writeSSEEvent: %v", err)
	}
	body := rr.Body.String()
	if !strings.HasPrefix(body, "data: ") {
		t.Errorf("expected SSE 'data: ' prefix, got %q", body)
	}
	if !strings.HasSuffix(body, "\n\n") {
		t.Errorf("expected SSE double-newline suffix, got %q", body)
	}
	// Verify payload is valid JSON.
	jsonPart := strings.TrimPrefix(body, "data: ")
	jsonPart = strings.TrimSpace(jsonPart)
	var parsed RPIC2Event
	if err := json.Unmarshal([]byte(jsonPart), &parsed); err != nil {
		t.Errorf("SSE payload is not valid JSON: %v\nbody: %q", err, body)
	}
	if parsed.EventID != ev.EventID {
		t.Errorf("event_id: want %q, got %q", ev.EventID, parsed.EventID)
	}
}

func TestBuildServeMux(t *testing.T) {
	dir := t.TempDir()
	mux := buildServeMux(dir, "rpi-test")
	if mux == nil {
		t.Fatal("buildServeMux returned nil")
	}

	// Verify / route returns HTML.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("/ returned %d, want 200", rr.Code)
	}

	// Verify /runs route returns JSON array.
	req = httptest.NewRequest(http.MethodGet, "/runs", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("/runs returned %d, want 200", rr.Code)
	}
}

func TestSetCORSHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	setCORSHeaders(rr)
	if v := rr.Header().Get("Access-Control-Allow-Origin"); v != "*" {
		t.Errorf("CORS origin: want *, got %q", v)
	}
}
