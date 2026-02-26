package main

import (
	"encoding/json"
	"testing"
)

func TestStreamEvents_ParseStreamEvent_Init(t *testing.T) {
	data := `{"type":"init","session_id":"sess-abc","model":"claude-sonnet-4-20250514","tools":["Bash","Read","Grep"]}`
	ev, err := ParseStreamEvent([]byte(data))
	if err != nil {
		t.Fatalf("ParseStreamEvent error: %v", err)
	}
	if ev.Type != EventTypeInit {
		t.Errorf("Type = %q, want %q", ev.Type, EventTypeInit)
	}
	if ev.SessionID != "sess-abc" {
		t.Errorf("SessionID = %q, want %q", ev.SessionID, "sess-abc")
	}
	if ev.Model != "claude-sonnet-4-20250514" {
		t.Errorf("Model = %q, want %q", ev.Model, "claude-sonnet-4-20250514")
	}
	if len(ev.Tools) != 3 {
		t.Errorf("Tools count = %d, want 3", len(ev.Tools))
	}
}

func TestStreamEvents_ParseStreamEvent_Assistant(t *testing.T) {
	data := `{"type":"assistant","subtype":"tool_use","tool_name":"Bash","message":"ls -la"}`
	ev, err := ParseStreamEvent([]byte(data))
	if err != nil {
		t.Fatalf("ParseStreamEvent error: %v", err)
	}
	if ev.Type != EventTypeAssistant {
		t.Errorf("Type = %q, want %q", ev.Type, EventTypeAssistant)
	}
	if ev.Subtype != "tool_use" {
		t.Errorf("Subtype = %q, want %q", ev.Subtype, "tool_use")
	}
	if ev.ToolName != "Bash" {
		t.Errorf("ToolName = %q, want %q", ev.ToolName, "Bash")
	}
}

func TestStreamEvents_ParseStreamEvent_Result(t *testing.T) {
	data := `{"type":"result","cost_usd":0.042,"duration_ms":12500,"duration_api_ms":10000,"is_error":false,"num_turns":3}`
	ev, err := ParseStreamEvent([]byte(data))
	if err != nil {
		t.Fatalf("ParseStreamEvent error: %v", err)
	}
	if ev.Type != EventTypeResult {
		t.Errorf("Type = %q, want %q", ev.Type, EventTypeResult)
	}
	if ev.CostUSD != 0.042 {
		t.Errorf("CostUSD = %f, want 0.042", ev.CostUSD)
	}
	if ev.DurationMS != 12500 {
		t.Errorf("DurationMS = %f, want 12500", ev.DurationMS)
	}
	if ev.DurationAPIMS != 10000 {
		t.Errorf("DurationAPIMS = %f, want 10000", ev.DurationAPIMS)
	}
	if ev.IsError {
		t.Error("IsError should be false")
	}
	if ev.NumTurns != 3 {
		t.Errorf("NumTurns = %d, want 3", ev.NumTurns)
	}
}

func TestStreamEvents_ParseStreamEvent_ErrorResult(t *testing.T) {
	data := `{"type":"result","is_error":true,"message":"something went wrong"}`
	ev, err := ParseStreamEvent([]byte(data))
	if err != nil {
		t.Fatalf("ParseStreamEvent error: %v", err)
	}
	if !ev.IsError {
		t.Error("IsError should be true")
	}
	if ev.Message != "something went wrong" {
		t.Errorf("Message = %q, want %q", ev.Message, "something went wrong")
	}
}

func TestStreamEvents_ParseStreamEvent_ToolInput(t *testing.T) {
	data := `{"type":"assistant","subtype":"tool_use","tool_name":"Read","tool_use_id":"tu-1","tool_input":{"path":"/tmp/test.go"}}`
	ev, err := ParseStreamEvent([]byte(data))
	if err != nil {
		t.Fatalf("ParseStreamEvent error: %v", err)
	}
	if ev.ToolUseID != "tu-1" {
		t.Errorf("ToolUseID = %q, want %q", ev.ToolUseID, "tu-1")
	}
	if ev.ToolInput == nil {
		t.Fatal("ToolInput should not be nil")
	}
	var input map[string]string
	if err := json.Unmarshal(ev.ToolInput, &input); err != nil {
		t.Fatalf("unmarshal ToolInput: %v", err)
	}
	if input["path"] != "/tmp/test.go" {
		t.Errorf("ToolInput.path = %q, want %q", input["path"], "/tmp/test.go")
	}
}

func TestStreamEvents_ParseStreamEvent_InvalidJSON(t *testing.T) {
	_, err := ParseStreamEvent([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestStreamEvents_ParseStreamEvent_EmptyObject(t *testing.T) {
	ev, err := ParseStreamEvent([]byte("{}"))
	if err != nil {
		t.Fatalf("ParseStreamEvent error: %v", err)
	}
	if ev.Type != "" {
		t.Errorf("Type = %q, want empty", ev.Type)
	}
}

func TestStreamEvents_ParseStreamEvent_UnknownFields(t *testing.T) {
	// Permissive parsing: unknown fields should be silently ignored.
	data := `{"type":"system","message":"hello","unknown_field":"value","another":42}`
	ev, err := ParseStreamEvent([]byte(data))
	if err != nil {
		t.Fatalf("ParseStreamEvent error: %v", err)
	}
	if ev.Type != EventTypeSystem {
		t.Errorf("Type = %q, want %q", ev.Type, EventTypeSystem)
	}
	if ev.Message != "hello" {
		t.Errorf("Message = %q, want %q", ev.Message, "hello")
	}
}

func TestStreamEvents_EventTypeConstants(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "system", value: EventTypeSystem, want: "system"},
		{name: "assistant", value: EventTypeAssistant, want: "assistant"},
		{name: "user", value: EventTypeUser, want: "user"},
		{name: "result", value: EventTypeResult, want: "result"},
		{name: "init", value: EventTypeInit, want: "init"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value != tc.want {
				t.Errorf("%s = %q, want %q", tc.name, tc.value, tc.want)
			}
		})
	}
}

func TestStreamEvents_ParseStreamEvent_RoundTrip(t *testing.T) {
	original := StreamEvent{
		Type:      EventTypeResult,
		CostUSD:   0.05,
		NumTurns:  7,
		IsError:   false,
		Message:   "done",
		SessionID: "sess-rt",
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	parsed, err := ParseStreamEvent(data)
	if err != nil {
		t.Fatalf("ParseStreamEvent: %v", err)
	}
	if parsed.Type != original.Type {
		t.Errorf("Type = %q, want %q", parsed.Type, original.Type)
	}
	if parsed.CostUSD != original.CostUSD {
		t.Errorf("CostUSD = %f, want %f", parsed.CostUSD, original.CostUSD)
	}
	if parsed.NumTurns != original.NumTurns {
		t.Errorf("NumTurns = %d, want %d", parsed.NumTurns, original.NumTurns)
	}
	if parsed.SessionID != original.SessionID {
		t.Errorf("SessionID = %q, want %q", parsed.SessionID, original.SessionID)
	}
}
