package main

import (
	"encoding/json"
	"testing"
)

// --- issueTypeFromMap ---

func TestIssueTypeFromMap_NilMap(t *testing.T) {
	isEpic, ok := issueTypeFromMap(nil)
	if ok {
		t.Error("expected ok=false for nil map")
	}
	if isEpic {
		t.Error("expected isEpic=false for nil map")
	}
}

func TestIssueTypeFromMap_EpicBoolTrue(t *testing.T) {
	m := map[string]any{"epic": true}
	isEpic, ok := issueTypeFromMap(m)
	if !ok {
		t.Error("expected ok=true when epic field is present")
	}
	if !isEpic {
		t.Error("expected isEpic=true")
	}
}

func TestIssueTypeFromMap_EpicBoolFalse(t *testing.T) {
	m := map[string]any{"epic": false}
	isEpic, ok := issueTypeFromMap(m)
	if !ok {
		t.Error("expected ok=true when epic field is present")
	}
	if isEpic {
		t.Error("expected isEpic=false")
	}
}

func TestIssueTypeFromMap_TypeFieldEpic(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]any
		want    bool
	}{
		{name: "type=epic", payload: map[string]any{"type": "epic"}, want: true},
		{name: "type=Epic", payload: map[string]any{"type": "Epic"}, want: true},
		{name: "type=EPIC", payload: map[string]any{"type": "EPIC"}, want: true},
		{name: "type=task", payload: map[string]any{"type": "task"}, want: false},
		{name: "issue_type=epic", payload: map[string]any{"issue_type": "epic"}, want: true},
		{name: "kind=epic", payload: map[string]any{"kind": "epic"}, want: true},
		{name: "kind=bug", payload: map[string]any{"kind": "bug"}, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isEpic, ok := issueTypeFromMap(tc.payload)
			if !ok {
				t.Error("expected ok=true")
			}
			if isEpic != tc.want {
				t.Errorf("isEpic = %v, want %v", isEpic, tc.want)
			}
		})
	}
}

func TestIssueTypeFromMap_NestedIssueField(t *testing.T) {
	m := map[string]any{
		"issue": map[string]any{
			"type": "epic",
		},
	}
	isEpic, ok := issueTypeFromMap(m)
	if !ok {
		t.Error("expected ok=true for nested issue field")
	}
	if !isEpic {
		t.Error("expected isEpic=true for nested issue.type=epic")
	}
}

func TestIssueTypeFromMap_NestedIssueField_NonEpic(t *testing.T) {
	m := map[string]any{
		"issue": map[string]any{
			"type": "task",
		},
	}
	isEpic, ok := issueTypeFromMap(m)
	if !ok {
		t.Error("expected ok=true for nested issue field")
	}
	if isEpic {
		t.Error("expected isEpic=false for nested issue.type=task")
	}
}

func TestIssueTypeFromMap_NoTypeFields(t *testing.T) {
	m := map[string]any{"title": "some issue", "status": "open"}
	_, ok := issueTypeFromMap(m)
	if ok {
		t.Error("expected ok=false when no type-related fields exist")
	}
}

func TestIssueTypeFromMap_EpicFieldNonBool(t *testing.T) {
	// When "epic" field is a non-bool (e.g. string), fall through to type/kind fields
	m := map[string]any{"epic": "yes", "type": "epic"}
	isEpic, ok := issueTypeFromMap(m)
	if !ok {
		t.Error("expected ok=true (should fall through to type field)")
	}
	if !isEpic {
		t.Error("expected isEpic=true from type field")
	}
}

// --- parseIssueTypeFromShowJSON ---

func TestParseIssueTypeFromShowJSON_SingleObject(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]any
		want    bool
		wantErr bool
	}{
		{
			name:    "epic=true",
			payload: map[string]any{"id": "ag-1", "epic": true},
			want:    true,
		},
		{
			name:    "epic=false",
			payload: map[string]any{"id": "ag-2", "epic": false},
			want:    false,
		},
		{
			name:    "type=epic",
			payload: map[string]any{"id": "ag-3", "type": "epic"},
			want:    true,
		},
		{
			name:    "type=task",
			payload: map[string]any{"id": "ag-4", "type": "task"},
			want:    false,
		},
		{
			name:    "no type info",
			payload: map[string]any{"id": "ag-5", "title": "no type"},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, _ := json.Marshal(tc.payload)
			got, err := parseIssueTypeFromShowJSON(data)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("isEpic = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseIssueTypeFromShowJSON_Array(t *testing.T) {
	arr := []map[string]any{
		{"id": "ag-1", "title": "no type"},
		{"id": "ag-2", "type": "epic"},
	}
	data, _ := json.Marshal(arr)
	got, err := parseIssueTypeFromShowJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected isEpic=true from second array element")
	}
}

func TestParseIssueTypeFromShowJSON_ArrayNoType(t *testing.T) {
	arr := []map[string]any{
		{"id": "ag-1", "title": "no type"},
		{"id": "ag-2", "title": "also no type"},
	}
	data, _ := json.Marshal(arr)
	_, err := parseIssueTypeFromShowJSON(data)
	if err == nil {
		t.Error("expected error when no array entry has type info")
	}
}

func TestParseIssueTypeFromShowJSON_InvalidJSON(t *testing.T) {
	_, err := parseIssueTypeFromShowJSON([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// --- parseLatestEpicIDFromJSON ---

func TestParseLatestEpicIDFromJSON_MultipleEntries(t *testing.T) {
	entries := []struct{ ID string }{
		{ID: "ag-100"},
		{ID: "ag-200"},
		{ID: "ag-300"},
	}
	data, _ := json.Marshal(entries)
	got, err := parseLatestEpicIDFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return the LAST non-empty entry.
	if got != "ag-300" {
		t.Errorf("got %q, want %q", got, "ag-300")
	}
}

func TestParseLatestEpicIDFromJSON_SkipsEmptyIDs(t *testing.T) {
	entries := []struct{ ID string }{
		{ID: "ag-100"},
		{ID: ""},
		{ID: "  "},
	}
	data, _ := json.Marshal(entries)
	got, err := parseLatestEpicIDFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ag-100" {
		t.Errorf("got %q, want %q", got, "ag-100")
	}
}

func TestParseLatestEpicIDFromJSON_AllEmpty(t *testing.T) {
	entries := []struct{ ID string }{
		{ID: ""},
		{ID: "  "},
	}
	data, _ := json.Marshal(entries)
	_, err := parseLatestEpicIDFromJSON(data)
	if err == nil {
		t.Error("expected error when all IDs are empty")
	}
}

// --- parseLatestEpicIDFromText ---

func TestParseLatestEpicIDFromText_MultipleLines(t *testing.T) {
	output := `ag-100 epic: first
ag-200 epic: second
ag-300 epic: latest`
	got, err := parseLatestEpicIDFromText(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ag-300" {
		t.Errorf("got %q, want %q", got, "ag-300")
	}
}

func TestParseLatestEpicIDFromText_CustomPrefixes(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{name: "ag prefix", output: "ag-1a some description\n", want: "ag-1a"},
		{name: "bd prefix", output: "bd-xyz description\n", want: "bd-xyz"},
		{name: "cm prefix", output: "[cm-42] task title\n", want: "cm-42"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseLatestEpicIDFromText(tc.output)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseLatestEpicIDFromText_NoMatch(t *testing.T) {
	_, err := parseLatestEpicIDFromText("no issues found\n")
	if err == nil {
		t.Error("expected error when no epic ID matches")
	}
}

// --- parseFastPath ---

func TestParseFastPath_MicroEpic(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{name: "empty output", output: "", want: true},
		{name: "single issue", output: "ag-1 open Fix bug\n", want: true},
		{name: "two issues", output: "ag-1 open Fix bug\nag-2 open Add feature\n", want: true},
		{name: "three issues (not micro)", output: "ag-1 open\nag-2 open\nag-3 open\n", want: false},
		{name: "one blocked", output: "ag-1 blocked Fix bug\n", want: false},
		{name: "two issues one blocked", output: "ag-1 open\nag-2 BLOCKED\n", want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseFastPath(tc.output)
			if got != tc.want {
				t.Errorf("parseFastPath = %v, want %v", got, tc.want)
			}
		})
	}
}

// --- parseCrankCompletion ---

func TestParseCrankCompletion_AllClosed(t *testing.T) {
	output := "ag-1 closed Fix\nag-2 closed Add\n"
	got := parseCrankCompletion(output)
	if got != "DONE" {
		t.Errorf("parseCrankCompletion = %q, want %q", got, "DONE")
	}
}

func TestParseCrankCompletion_HasBlocked(t *testing.T) {
	output := "ag-1 closed\nag-2 blocked\n"
	got := parseCrankCompletion(output)
	if got != "BLOCKED" {
		t.Errorf("parseCrankCompletion = %q, want %q", got, "BLOCKED")
	}
}

func TestParseCrankCompletion_Partial(t *testing.T) {
	output := "ag-1 closed\nag-2 open\n"
	got := parseCrankCompletion(output)
	if got != "PARTIAL" {
		t.Errorf("parseCrankCompletion = %q, want %q", got, "PARTIAL")
	}
}

func TestParseCrankCompletion_EmptyOutput(t *testing.T) {
	got := parseCrankCompletion("")
	if got != "DONE" {
		t.Errorf("parseCrankCompletion (empty) = %q, want %q", got, "DONE")
	}
}

func TestParseCrankCompletion_CheckmarkAsClosed(t *testing.T) {
	output := "ag-1 \u2713 Fix\nag-2 \u2713 Add\n"
	got := parseCrankCompletion(output)
	if got != "DONE" {
		t.Errorf("parseCrankCompletion with checkmarks = %q, want %q", got, "DONE")
	}
}
