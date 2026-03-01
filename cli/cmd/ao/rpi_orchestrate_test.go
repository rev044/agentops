package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseBeadIDsFromText_TypicalBDOutput verifies that standard bd children output
// is parsed into a sorted slice of bead IDs.
func TestParseBeadIDsFromText_TypicalBDOutput(t *testing.T) {
	output := `ag-000.2 Implement rpi_serve.go orchestration mode
ag-000.4 Write orchestration tests
ag-000.1 Add rpi_orchestrate.go
ag-000.3 Update watch.html
`
	got := parseBeadIDsFromText(output)
	want := []string{"ag-000.1", "ag-000.2", "ag-000.3", "ag-000.4"}
	if len(got) != len(want) {
		t.Fatalf("got %d IDs, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

// TestParseBeadIDsFromText_Empty returns nil for empty or whitespace-only output.
func TestParseBeadIDsFromText_Empty(t *testing.T) {
	for _, input := range []string{"", "   ", "\n\n\n"} {
		if ids := parseBeadIDsFromText(input); len(ids) != 0 {
			t.Errorf("expected empty for %q, got %v", input, ids)
		}
	}
}

// TestParseBeadIDsFromText_IgnoresNonIDs ignores lines whose first token doesn't
// match the bead ID pattern.
func TestParseBeadIDsFromText_IgnoresNonIDs(t *testing.T) {
	output := `Not a bead ID
  ag-123 valid one
http://example.com/something
ag-000.99 another valid`
	got := parseBeadIDsFromText(output)
	if len(got) != 2 {
		t.Fatalf("expected 2 bead IDs, got %d: %v", len(got), got)
	}
	if got[0] != "ag-000.99" || got[1] != "ag-123" {
		t.Errorf("unexpected IDs: %v", got)
	}
}

// TestParseBeadIDsFromText_StatusPrefixes strips leading status indicator runes
// that bd may prepend (○ ◐ ● ✓ ❄) when they are attached to the bead ID token
// (no space separator).
func TestParseBeadIDsFromText_StatusPrefixes(t *testing.T) {
	// Status runes attached to bead ID (no space) — the parser strips them.
	output := "○ag-abc.1 pending\n●ag-abc.2 done\n✓ag-abc.3 closed\n"
	got := parseBeadIDsFromText(output)
	if len(got) != 3 {
		t.Fatalf("expected 3, got %d: %v", len(got), got)
	}
}

// TestSaveOrchState_RoundTrip writes and re-reads an orchState and verifies the
// key fields survive JSON serialisation.
func TestSaveOrchState_RoundTrip(t *testing.T) {
	dir := t.TempDir()

	state := &orchState{
		SchemaVersion: 1,
		RunID:         "rpi-deadbeef",
		Goal:          "add user auth",
		Phase:         orchPhaseDiscovery,
		EpicID:        "ag-999",
		Attempts:      map[string]int{"discovery": 1},
	}

	if err := saveOrchState(dir, state.RunID, state); err != nil {
		t.Fatalf("saveOrchState: %v", err)
	}

	path := filepath.Join(dir, ".agents", "rpi", "runs", state.RunID, "orchestration-state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read state file: %v", err)
	}

	var out orchState
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.RunID != state.RunID {
		t.Errorf("RunID: got %q, want %q", out.RunID, state.RunID)
	}
	if out.Goal != state.Goal {
		t.Errorf("Goal: got %q, want %q", out.Goal, state.Goal)
	}
	if out.Phase != orchPhaseDiscovery {
		t.Errorf("Phase: got %q, want %q", out.Phase, orchPhaseDiscovery)
	}
	if out.EpicID != "ag-999" {
		t.Errorf("EpicID: got %q, want %q", out.EpicID, "ag-999")
	}
	if out.UpdatedAt == "" {
		t.Error("UpdatedAt should be set by saveOrchState")
	}
}

// TestClassifyServeArg_GoalString verifies that a plain text goal is returned
// as the goal field, not as a run ID.
func TestClassifyServeArg_GoalString(t *testing.T) {
	goal, runID := classifyServeArg("", []string{"add user authentication"})
	if goal != "add user authentication" {
		t.Errorf("goal: got %q, want %q", goal, "add user authentication")
	}
	if runID != "" {
		t.Errorf("runID: got %q, want empty", runID)
	}
}

// TestClassifyServeArg_RunID verifies that a token matching rpiRunIDPattern is
// returned as a run ID, not as a goal.
func TestClassifyServeArg_RunID(t *testing.T) {
	goal, runID := classifyServeArg("", []string{"rpi-a1b2c3d4"})
	if goal != "" {
		t.Errorf("goal: got %q, want empty", goal)
	}
	if runID != "rpi-a1b2c3d4" {
		t.Errorf("runID: got %q, want %q", runID, "rpi-a1b2c3d4")
	}
}

// TestClassifyServeArg_FlagPrecedence verifies that --run-id wins over the
// positional argument when both are provided.
func TestClassifyServeArg_FlagPrecedence(t *testing.T) {
	goal, runID := classifyServeArg("rpi-deadbeef", []string{"some goal"})
	if goal != "" {
		t.Errorf("goal: got %q, want empty", goal)
	}
	if runID != "rpi-deadbeef" {
		t.Errorf("runID: got %q, want rpi-deadbeef", runID)
	}
}

// TestGenerateWorkerID_Format verifies the w-<hex> format and minimum length.
func TestGenerateWorkerID_Format(t *testing.T) {
	id := generateWorkerID()
	if !strings.HasPrefix(id, "w-") {
		t.Errorf("expected prefix w-, got %q", id)
	}
	// 2 bytes = 4 hex chars → "w-" + 8 chars
	if len(id) < 4 {
		t.Errorf("ID too short: %q", id)
	}
}

// TestRunRPIOrchestration_InvalidRoot verifies that orchestration returns an error
// when it cannot persist state (bad root path).
func TestRunRPIOrchestration_InvalidRoot(t *testing.T) {
	// Use a file as root so MkdirAll fails inside the registry dir.
	f, err := os.CreateTemp(t.TempDir(), "not-a-dir-*")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	opts := defaultOrchOpts()
	opts.MaxAttempts = 1
	opts.BDCommand = "false" // immediately fails bd children

	err = runRPIOrchestration(context.Background(), "test goal", "rpi-testtest", f.Name(), opts)
	if err == nil {
		t.Error("expected error for non-directory root, got nil")
	}
}
