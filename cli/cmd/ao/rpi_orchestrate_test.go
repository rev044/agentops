package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// TestRunRPIOrchestration_EndToEnd exercises the full INIT→DISCOVERY→IMPL→VALIDATION→DONE
// state machine with hermetic shell stubs substituted for the bd CLI and the runtime
// (claude). The stubs exit 0 immediately, so the test is fast and free of network or
// API dependencies.
//
// bd stub behaviour:
//   - "bd list --type epic --status open --json" → JSON array with one epic
//   - "bd children <id>"                         → two bead IDs
//   - anything else                               → exit 0 (no-op)
//
// runtime stub: accepts any arguments and exits 0 (simulates a completed claude session).
func TestRunRPIOrchestration_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e orchestration test in short mode")
	}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// Write bd stub.
	bdStub := filepath.Join(binDir, "bd")
	bdScript := "#!/bin/sh\n" +
		"if [ \"$1\" = \"list\" ]; then\n" +
		"    echo '[{\"id\":\"ag-e2e\"}]'\n" +
		"elif [ \"$1\" = \"children\" ]; then\n" +
		"    echo \"ag-e2e.1 first bead\"\n" +
		"    echo \"ag-e2e.2 second bead\"\n" +
		"fi\n" +
		"exit 0\n"
	if err := os.WriteFile(bdStub, []byte(bdScript), 0o750); err != nil {
		t.Fatal(err)
	}

	// Write runtime stub (simulates claude -p "...").
	runtimeStub := filepath.Join(binDir, "claude-stub")
	if err := os.WriteFile(runtimeStub, []byte("#!/bin/sh\nexit 0\n"), 0o750); err != nil {
		t.Fatal(err)
	}

	opts := defaultOrchOpts()
	opts.MaxAttempts = 1
	opts.BDCommand = bdStub
	opts.RuntimeCommand = runtimeStub
	opts.PhaseTimeout = 5 * time.Second

	runID := generateRunID()

	if err := runRPIOrchestration(context.Background(), "e2e test goal", runID, dir, opts); err != nil {
		t.Fatalf("runRPIOrchestration: %v", err)
	}

	// Verify final persisted orchestration state.
	stateDir := filepath.Join(dir, ".agents", "rpi", "runs", runID)
	stateData, err := os.ReadFile(filepath.Join(stateDir, "orchestration-state.json"))
	if err != nil {
		t.Fatalf("read orchestration-state.json: %v", err)
	}
	var state orchState
	if err := json.Unmarshal(stateData, &state); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if state.Phase != orchPhaseDone {
		t.Errorf("Phase = %q, want %q", state.Phase, orchPhaseDone)
	}
	if state.TerminalStatus != "done" {
		t.Errorf("TerminalStatus = %q, want %q", state.TerminalStatus, "done")
	}
	if state.EpicID != "ag-e2e" {
		t.Errorf("EpicID = %q, want %q", state.EpicID, "ag-e2e")
	}
	if len(state.Beads) != 2 {
		t.Errorf("len(Beads) = %d, want 2", len(state.Beads))
	}
	for _, bw := range state.Beads {
		if bw.Status != beadDone {
			t.Errorf("bead %q: Status = %q, want %q", bw.BeadID, bw.Status, beadDone)
		}
	}

	// Verify C2 events were written for each lifecycle milestone.
	events, err := loadRPIC2Events(dir, runID)
	if err != nil {
		t.Fatalf("loadRPIC2Events: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected C2 events, got none")
	}
	typeSet := make(map[string]bool, len(events))
	for _, ev := range events {
		typeSet[ev.Type] = true
	}
	for _, want := range []string{
		"worker.phase.spawned",
		"worker.phase.done",
		"worker.bead.spawned",
		"worker.bead.done",
	} {
		if !typeSet[want] {
			t.Errorf("C2 event type %q missing; found types: %v", want, collectKeys(typeSet))
		}
	}
}

// collectKeys returns a sorted slice of map keys for diagnostic messages.
func collectKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// writeBDStub writes a bd shell stub that returns a single epic + one bead.
// epicID and beadID are embedded directly in the script.
func writeBDStub(t *testing.T, path, epicID, beadID string) {
	t.Helper()
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"list\" ]; then\n" +
		"    echo '[{\"id\":\"" + epicID + "\"}]'\n" +
		"elif [ \"$1\" = \"children\" ]; then\n" +
		"    echo \"" + beadID + " bead title\"\n" +
		"fi\n" +
		"exit 0\n"
	if err := os.WriteFile(path, []byte(script), 0o750); err != nil {
		t.Fatalf("write bd stub: %v", err)
	}
}

// TestRunRPIOrchestration_DiscoveryFails verifies that when all discovery attempts fail
// the engine persists phase=failed with TerminalStatus="failed" and returns an error.
func TestRunRPIOrchestration_DiscoveryFails(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e orchestration test in short mode")
	}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		t.Fatal(err)
	}

	bdStub := filepath.Join(binDir, "bd")
	if err := os.WriteFile(bdStub, []byte("#!/bin/sh\nexit 0\n"), 0o750); err != nil {
		t.Fatal(err)
	}

	runtimeStub := filepath.Join(binDir, "claude-stub")
	if err := os.WriteFile(runtimeStub, []byte("#!/bin/sh\nexit 1\n"), 0o750); err != nil {
		t.Fatal(err)
	}

	opts := defaultOrchOpts()
	opts.MaxAttempts = 2
	opts.BDCommand = bdStub
	opts.RuntimeCommand = runtimeStub
	opts.PhaseTimeout = 5 * time.Second

	runID := generateRunID()
	err := runRPIOrchestration(context.Background(), "fail goal", runID, dir, opts)
	if err == nil {
		t.Fatal("expected error from discovery failure, got nil")
	}

	stateData, readErr := os.ReadFile(filepath.Join(dir, ".agents", "rpi", "runs", runID, "orchestration-state.json"))
	if readErr != nil {
		t.Fatalf("read state: %v", readErr)
	}
	var state orchState
	if err := json.Unmarshal(stateData, &state); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if state.Phase != orchPhaseFailed {
		t.Errorf("Phase = %q, want %q", state.Phase, orchPhaseFailed)
	}
	if state.TerminalStatus != "failed" {
		t.Errorf("TerminalStatus = %q, want %q", state.TerminalStatus, "failed")
	}
	if state.Attempts["discovery"] != 2 {
		t.Errorf("Attempts[discovery] = %d, want 2", state.Attempts["discovery"])
	}
}

// TestRunRPIOrchestration_DiscoveryRetriesAndSucceeds verifies that a transient
// discovery failure triggers a retry and the run completes as orchPhaseDone.
func TestRunRPIOrchestration_DiscoveryRetriesAndSucceeds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e orchestration test in short mode")
	}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		t.Fatal(err)
	}

	bdStub := filepath.Join(binDir, "bd")
	writeBDStub(t, bdStub, "ag-retry", "ag-retry.1")

	// fail-once stub: uses TMPDIR counter file — exits 1 on first call, 0 thereafter.
	runtimeStub := filepath.Join(binDir, "claude-stub")
	failOnceScript := "#!/bin/sh\n" +
		"COUNTER=\"${TMPDIR:-/tmp}/stub-call-count\"\n" +
		"if [ ! -f \"$COUNTER\" ]; then\n" +
		"    touch \"$COUNTER\"\n" +
		"    exit 1\n" +
		"fi\n" +
		"exit 0\n"
	if err := os.WriteFile(runtimeStub, []byte(failOnceScript), 0o750); err != nil {
		t.Fatal(err)
	}

	// t.Setenv sets process-wide TMPDIR for subprocess isolation; must NOT use t.Parallel().
	t.Setenv("TMPDIR", dir)

	opts := defaultOrchOpts()
	opts.MaxAttempts = 2
	opts.BDCommand = bdStub
	opts.RuntimeCommand = runtimeStub
	opts.PhaseTimeout = 5 * time.Second

	runID := generateRunID()
	if err := runRPIOrchestration(context.Background(), "retry goal", runID, dir, opts); err != nil {
		t.Fatalf("runRPIOrchestration: %v", err)
	}

	stateData, err := os.ReadFile(filepath.Join(dir, ".agents", "rpi", "runs", runID, "orchestration-state.json"))
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var state orchState
	if err := json.Unmarshal(stateData, &state); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if state.Phase != orchPhaseDone {
		t.Errorf("Phase = %q, want %q", state.Phase, orchPhaseDone)
	}
	if state.Attempts["discovery"] != 2 {
		t.Errorf("Attempts[discovery] = %d, want 2", state.Attempts["discovery"])
	}
}

// TestRunRPIOrchestration_ValidationFails verifies that exhausted validation retries
// terminate with phase=failed and an error message containing "validation failed".
func TestRunRPIOrchestration_ValidationFails(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e orchestration test in short mode")
	}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		t.Fatal(err)
	}

	bdStub := filepath.Join(binDir, "bd")
	writeBDStub(t, bdStub, "ag-vfail", "ag-vfail.1")

	// Stub: fails for validation-phase prompts, succeeds for discovery and beads.
	runtimeStub := filepath.Join(binDir, "claude-stub")
	script := "#!/bin/sh\n" +
		"# $1=-p $2=prompt\n" +
		"if echo \"$2\" | grep -q \"validation phase\"; then\n" +
		"    exit 1\n" +
		"fi\n" +
		"exit 0\n"
	if err := os.WriteFile(runtimeStub, []byte(script), 0o750); err != nil {
		t.Fatal(err)
	}

	opts := defaultOrchOpts()
	opts.MaxAttempts = 2
	opts.BDCommand = bdStub
	opts.RuntimeCommand = runtimeStub
	opts.PhaseTimeout = 5 * time.Second

	runID := generateRunID()
	err := runRPIOrchestration(context.Background(), "validation fail goal", runID, dir, opts)
	if err == nil {
		t.Fatal("expected error from validation failure, got nil")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("error %q does not contain 'validation failed'", err.Error())
	}

	stateData, readErr := os.ReadFile(filepath.Join(dir, ".agents", "rpi", "runs", runID, "orchestration-state.json"))
	if readErr != nil {
		t.Fatalf("read state: %v", readErr)
	}
	var state orchState
	if err := json.Unmarshal(stateData, &state); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if state.Phase != orchPhaseFailed {
		t.Errorf("Phase = %q, want %q", state.Phase, orchPhaseFailed)
	}
}

// TestRunRPIOrchestration_ValidationRetriesAndSucceeds verifies that a single
// validation failure triggers a retry and the run completes as orchPhaseDone.
func TestRunRPIOrchestration_ValidationRetriesAndSucceeds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e orchestration test in short mode")
	}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		t.Fatal(err)
	}

	bdStub := filepath.Join(binDir, "bd")
	writeBDStub(t, bdStub, "ag-vretry", "ag-vretry.1")

	// Stub: fails once for validation-phase prompts using a per-test counter file.
	runtimeStub := filepath.Join(binDir, "claude-stub")
	script := "#!/bin/sh\n" +
		"COUNTER=\"${TMPDIR:-/tmp}/validation-fail\"\n" +
		"if echo \"$2\" | grep -q \"validation phase\"; then\n" +
		"    if [ ! -f \"$COUNTER\" ]; then\n" +
		"        touch \"$COUNTER\"\n" +
		"        exit 1\n" +
		"    fi\n" +
		"fi\n" +
		"exit 0\n"
	if err := os.WriteFile(runtimeStub, []byte(script), 0o750); err != nil {
		t.Fatal(err)
	}

	// t.Setenv sets process-wide TMPDIR for subprocess isolation; must NOT use t.Parallel().
	t.Setenv("TMPDIR", dir)

	opts := defaultOrchOpts()
	opts.MaxAttempts = 2
	opts.BDCommand = bdStub
	opts.RuntimeCommand = runtimeStub
	opts.PhaseTimeout = 5 * time.Second

	runID := generateRunID()
	if err := runRPIOrchestration(context.Background(), "val retry goal", runID, dir, opts); err != nil {
		t.Fatalf("runRPIOrchestration: %v", err)
	}

	stateData, err := os.ReadFile(filepath.Join(dir, ".agents", "rpi", "runs", runID, "orchestration-state.json"))
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var state orchState
	if err := json.Unmarshal(stateData, &state); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if state.Phase != orchPhaseDone {
		t.Errorf("Phase = %q, want %q", state.Phase, orchPhaseDone)
	}
	if state.Attempts["validation"] != 2 {
		t.Errorf("Attempts[validation] = %d, want 2", state.Attempts["validation"])
	}
}

// TestRunRPIOrchestration_NoBeads_SkipsImplGoesToValidation verifies that when
// bd children returns no beads the implementation phase is skipped and the run
// still completes as orchPhaseDone.
func TestRunRPIOrchestration_NoBeads_SkipsImplGoesToValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e orchestration test in short mode")
	}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// bd stub: list returns an epic, children returns empty (no beads).
	bdStub := filepath.Join(binDir, "bd")
	bdScript := "#!/bin/sh\n" +
		"if [ \"$1\" = \"list\" ]; then\n" +
		"    echo '[{\"id\":\"ag-nobead\"}]'\n" +
		"fi\n" +
		"exit 0\n"
	if err := os.WriteFile(bdStub, []byte(bdScript), 0o750); err != nil {
		t.Fatal(err)
	}

	runtimeStub := filepath.Join(binDir, "claude-stub")
	if err := os.WriteFile(runtimeStub, []byte("#!/bin/sh\nexit 0\n"), 0o750); err != nil {
		t.Fatal(err)
	}

	opts := defaultOrchOpts()
	opts.MaxAttempts = 1
	opts.BDCommand = bdStub
	opts.RuntimeCommand = runtimeStub
	opts.PhaseTimeout = 5 * time.Second

	runID := generateRunID()
	if err := runRPIOrchestration(context.Background(), "no beads goal", runID, dir, opts); err != nil {
		t.Fatalf("runRPIOrchestration: %v", err)
	}

	stateData, err := os.ReadFile(filepath.Join(dir, ".agents", "rpi", "runs", runID, "orchestration-state.json"))
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var state orchState
	if err := json.Unmarshal(stateData, &state); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	if state.Phase != orchPhaseDone {
		t.Errorf("Phase = %q, want %q", state.Phase, orchPhaseDone)
	}
	if len(state.Beads) != 0 {
		t.Errorf("len(Beads) = %d, want 0", len(state.Beads))
	}
}

// TestRunRPIOrchestration_ContextCancel verifies that a pre-cancelled context causes
// runRPIOrchestration to return a non-nil error without blocking.
func TestRunRPIOrchestration_ContextCancel(t *testing.T) {
	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// Stub content is irrelevant — context is pre-cancelled so the stub never executes.
	runtimeStub := filepath.Join(binDir, "claude-stub")
	if err := os.WriteFile(runtimeStub, []byte("#!/bin/sh\nexit 0\n"), 0o750); err != nil {
		t.Fatal(err)
	}

	opts := defaultOrchOpts()
	opts.MaxAttempts = 3
	opts.BDCommand = "false"
	opts.RuntimeCommand = runtimeStub
	opts.PhaseTimeout = 5 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before calling — exercises the ctx.Err() guard at loop entry

	runID := generateRunID()
	err := runRPIOrchestration(ctx, "cancel goal", runID, dir, opts)
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

// TestRunBeadWorker_RetriesAndSucceeds calls runBeadWorker directly and verifies
// that a fail-once runtime stub triggers a retry, landing at beadDone with Attempts==2.
func TestRunBeadWorker_RetriesAndSucceeds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e bead worker test in short mode")
	}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// fail-once stub: exits 1 first call, 0 thereafter.
	runtimeStub := filepath.Join(binDir, "claude-stub")
	failOnceScript := "#!/bin/sh\n" +
		"COUNTER=\"${TMPDIR:-/tmp}/bead-stub-count\"\n" +
		"if [ ! -f \"$COUNTER\" ]; then\n" +
		"    touch \"$COUNTER\"\n" +
		"    exit 1\n" +
		"fi\n" +
		"exit 0\n"
	if err := os.WriteFile(runtimeStub, []byte(failOnceScript), 0o750); err != nil {
		t.Fatal(err)
	}

	// t.Setenv sets process-wide TMPDIR for subprocess isolation; must NOT use t.Parallel().
	t.Setenv("TMPDIR", dir)

	opts := defaultOrchOpts()
	opts.MaxAttempts = 2
	opts.RuntimeCommand = runtimeStub
	opts.PhaseTimeout = 5 * time.Second

	runID := generateRunID()
	var bw beadWorker
	if err := runBeadWorker(context.Background(), "ag-bw.1", runID, &bw, dir, opts); err != nil {
		t.Fatalf("runBeadWorker: %v", err)
	}
	if bw.Status != beadDone {
		t.Errorf("Status = %q, want %q", bw.Status, beadDone)
	}
	if bw.Attempts != 2 {
		t.Errorf("Attempts = %d, want 2", bw.Attempts)
	}
	if bw.BeadID != "ag-bw.1" {
		t.Errorf("BeadID = %q, want %q", bw.BeadID, "ag-bw.1")
	}
}
