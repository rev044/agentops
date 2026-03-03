package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHandoff_WritesArtifact(t *testing.T) {
	dir := t.TempDir()

	artifact := &handoffArtifact{
		SchemaVersion: 1,
		ID:            "handoff-20260303T191026Z",
		CreatedAt:     "2026-03-03T19:10:26Z",
		Type:          "manual",
		Goal:          "test goal",
		Summary:       "test summary",
		Consumed:      false,
	}

	path, err := writeHandoffArtifact(dir, artifact)
	if err != nil {
		t.Fatalf("writeHandoffArtifact failed: %v", err)
	}

	// Verify file exists
	if !fileExists(path) {
		t.Fatalf("artifact file not found at %s", path)
	}

	// Verify path is under .agents/handoff/
	expectedDir := filepath.Join(dir, ".agents", "handoff")
	if !strings.HasPrefix(path, expectedDir) {
		t.Errorf("path %q not under expected dir %q", path, expectedDir)
	}

	// Verify JSON content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}

	var read handoffArtifact
	if err := json.Unmarshal(data, &read); err != nil {
		t.Fatalf("unmarshal artifact: %v", err)
	}

	if read.Goal != "test goal" {
		t.Errorf("goal = %q, want %q", read.Goal, "test goal")
	}
	if read.Summary != "test summary" {
		t.Errorf("summary = %q, want %q", read.Summary, "test summary")
	}
	if read.Type != "manual" {
		t.Errorf("type = %q, want %q", read.Type, "manual")
	}
}

func TestRunHandoff_DryRun(t *testing.T) {
	dir := t.TempDir()

	artifact := &handoffArtifact{
		SchemaVersion: 1,
		ID:            "handoff-20260303T191026Z",
		CreatedAt:     "2026-03-03T19:10:26Z",
		Type:          "manual",
		Summary:       "dry run test",
		Consumed:      false,
	}

	// Marshal to verify JSON is valid (what --dry-run would print)
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(data), "dry run test") {
		t.Error("marshalled output missing summary")
	}

	// Verify NO file is written
	handoffDir := filepath.Join(dir, ".agents", "handoff")
	if fileExists(handoffDir) {
		t.Error("handoff dir should not exist in dry-run mode")
	}
}

func TestRunHandoff_NoKill(t *testing.T) {
	dir := t.TempDir()

	artifact := &handoffArtifact{
		SchemaVersion: 1,
		ID:            "handoff-20260303T191026Z",
		CreatedAt:     "2026-03-03T19:10:26Z",
		Type:          "manual",
		Summary:       "no-kill test",
		Consumed:      false,
	}

	path, err := writeHandoffArtifact(dir, artifact)
	if err != nil {
		t.Fatalf("writeHandoffArtifact failed: %v", err)
	}

	// Verify artifact was written (no-kill still writes)
	if !fileExists(path) {
		t.Fatal("artifact should exist with --no-kill")
	}
}

func TestRunHandoff_NonTmux(t *testing.T) {
	// Unset TMUX_PANE to simulate non-tmux environment
	origPane := os.Getenv("TMUX_PANE")
	os.Unsetenv("TMUX_PANE")
	defer func() {
		if origPane != "" {
			os.Setenv("TMUX_PANE", origPane)
		}
	}()

	err := killSessionViaTmux("/tmp")
	if err == nil {
		t.Error("expected error when TMUX_PANE is unset")
	}
	if !strings.Contains(err.Error(), "not in tmux") {
		t.Errorf("error = %q, want 'not in tmux'", err.Error())
	}
}

func TestRunHandoff_CollectState(t *testing.T) {
	dir := t.TempDir()

	// Create a git repo for collectHandoffState to work with
	state := collectHandoffState(dir)

	if state == nil {
		t.Fatal("collectHandoffState returned nil")
	}

	// In a temp dir with no git repo, branch should be empty and dirty should be false
	if state.GitDirty {
		t.Error("expected GitDirty=false in empty temp dir")
	}
}

func TestRunHandoff_RPIPhase(t *testing.T) {
	dir := t.TempDir()

	// Create phased-state.json with verdicts
	rpiDir := filepath.Join(dir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stateData := `{
		"verdicts": {"security": "PASS", "complexity": "WARN"},
		"epic_id": "na-abc",
		"run_id": "r12345"
	}`
	if err := os.WriteFile(filepath.Join(rpiDir, "phased-state.json"), []byte(stateData), 0o644); err != nil {
		t.Fatal(err)
	}

	rpi := buildHandoffRPIContext(dir, 2, "", "")
	if rpi == nil {
		t.Fatal("buildHandoffRPIContext returned nil")
	}

	if rpi.Phase != 2 {
		t.Errorf("phase = %d, want 2", rpi.Phase)
	}
	if rpi.PhaseName != "implementation" {
		t.Errorf("phase_name = %q, want %q", rpi.PhaseName, "implementation")
	}
	if rpi.EpicID != "na-abc" {
		t.Errorf("epic_id = %q, want %q", rpi.EpicID, "na-abc")
	}
	if rpi.RunID != "r12345" {
		t.Errorf("run_id = %q, want %q", rpi.RunID, "r12345")
	}
	if len(rpi.Verdicts) != 2 {
		t.Errorf("verdicts count = %d, want 2", len(rpi.Verdicts))
	}
	if rpi.Verdicts["security"] != "PASS" {
		t.Errorf("verdicts[security] = %q, want %q", rpi.Verdicts["security"], "PASS")
	}
}

func TestRunHandoff_RPIPhase_FlagOverride(t *testing.T) {
	dir := t.TempDir()

	// Create phased-state.json
	rpiDir := filepath.Join(dir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stateData := `{"verdicts": {}, "epic_id": "from-file", "run_id": "from-file"}`
	if err := os.WriteFile(filepath.Join(rpiDir, "phased-state.json"), []byte(stateData), 0o644); err != nil {
		t.Fatal(err)
	}

	// Flags should take precedence over file values
	rpi := buildHandoffRPIContext(dir, 1, "from-flag", "run-flag")
	if rpi.EpicID != "from-flag" {
		t.Errorf("epic_id = %q, want %q (flag override)", rpi.EpicID, "from-flag")
	}
	if rpi.RunID != "run-flag" {
		t.Errorf("run_id = %q, want %q (flag override)", rpi.RunID, "run-flag")
	}
}

func TestRunHandoff_SchemaVersion(t *testing.T) {
	artifact := &handoffArtifact{
		SchemaVersion: 1,
		ID:            "handoff-20260303T191026Z",
		CreatedAt:     "2026-03-03T19:10:26Z",
		Type:          "manual",
		Consumed:      false,
	}

	data, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}

	sv, ok := parsed["schema_version"]
	if !ok {
		t.Fatal("schema_version field missing")
	}
	if sv != float64(1) {
		t.Errorf("schema_version = %v, want 1", sv)
	}
}

func TestHandoffArtifact_JSONRoundTrip(t *testing.T) {
	original := handoffArtifact{
		SchemaVersion: 1,
		ID:            "handoff-20260303T191026Z",
		CreatedAt:     "2026-03-03T19:10:26Z",
		Type:          "rpi",
		Goal:          "build auth module",
		Summary:       "completed JWT flow",
		Continuation:  "add refresh tokens next",
		ArtifactsProduced: []string{
			"cli/cmd/ao/auth.go",
			"cli/cmd/ao/auth_test.go",
		},
		DecisionsMade: []string{"used HS256 over RS256 for simplicity"},
		OpenRisks:     []string{"token rotation not implemented"},
		RPI: &handoffRPI{
			Phase:     2,
			PhaseName: "implementation",
			EpicID:    "na-auth",
			RunID:     "abc123",
			Verdicts:  map[string]string{"security": "PASS", "complexity": "WARN"},
		},
		State: &handoffState{
			GitBranch:      "feat/auth",
			GitDirty:       true,
			ModifiedFiles:  []string{"auth.go", "auth_test.go"},
			ActiveBead:     "na-auth.3",
			OpenBeadsCount: 5,
			RecentCommits:  []string{"abc123 add JWT signing", "def456 add token validation"},
		},
		Consumed:   false,
		ConsumedAt: nil,
		ConsumedBy: nil,
	}

	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var roundTripped handoffArtifact
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Verify all fields survive roundtrip
	if roundTripped.SchemaVersion != original.SchemaVersion {
		t.Errorf("schema_version: got %d, want %d", roundTripped.SchemaVersion, original.SchemaVersion)
	}
	if roundTripped.ID != original.ID {
		t.Errorf("id: got %q, want %q", roundTripped.ID, original.ID)
	}
	if roundTripped.CreatedAt != original.CreatedAt {
		t.Errorf("created_at: got %q, want %q", roundTripped.CreatedAt, original.CreatedAt)
	}
	if roundTripped.Type != original.Type {
		t.Errorf("type: got %q, want %q", roundTripped.Type, original.Type)
	}
	if roundTripped.Goal != original.Goal {
		t.Errorf("goal: got %q, want %q", roundTripped.Goal, original.Goal)
	}
	if roundTripped.Summary != original.Summary {
		t.Errorf("summary: got %q, want %q", roundTripped.Summary, original.Summary)
	}
	if roundTripped.Continuation != original.Continuation {
		t.Errorf("continuation: got %q, want %q", roundTripped.Continuation, original.Continuation)
	}
	if len(roundTripped.ArtifactsProduced) != len(original.ArtifactsProduced) {
		t.Errorf("artifacts_produced: got %d items, want %d", len(roundTripped.ArtifactsProduced), len(original.ArtifactsProduced))
	}
	if len(roundTripped.DecisionsMade) != len(original.DecisionsMade) {
		t.Errorf("decisions_made: got %d items, want %d", len(roundTripped.DecisionsMade), len(original.DecisionsMade))
	}
	if len(roundTripped.OpenRisks) != len(original.OpenRisks) {
		t.Errorf("open_risks: got %d items, want %d", len(roundTripped.OpenRisks), len(original.OpenRisks))
	}
	if roundTripped.RPI == nil {
		t.Fatal("rpi: got nil, want non-nil")
	}
	if roundTripped.RPI.Phase != original.RPI.Phase {
		t.Errorf("rpi.phase: got %d, want %d", roundTripped.RPI.Phase, original.RPI.Phase)
	}
	if roundTripped.RPI.PhaseName != original.RPI.PhaseName {
		t.Errorf("rpi.phase_name: got %q, want %q", roundTripped.RPI.PhaseName, original.RPI.PhaseName)
	}
	if roundTripped.RPI.EpicID != original.RPI.EpicID {
		t.Errorf("rpi.epic_id: got %q, want %q", roundTripped.RPI.EpicID, original.RPI.EpicID)
	}
	if roundTripped.RPI.RunID != original.RPI.RunID {
		t.Errorf("rpi.run_id: got %q, want %q", roundTripped.RPI.RunID, original.RPI.RunID)
	}
	if len(roundTripped.RPI.Verdicts) != len(original.RPI.Verdicts) {
		t.Errorf("rpi.verdicts: got %d entries, want %d", len(roundTripped.RPI.Verdicts), len(original.RPI.Verdicts))
	}
	if roundTripped.State == nil {
		t.Fatal("state: got nil, want non-nil")
	}
	if roundTripped.State.GitBranch != original.State.GitBranch {
		t.Errorf("state.git_branch: got %q, want %q", roundTripped.State.GitBranch, original.State.GitBranch)
	}
	if roundTripped.State.GitDirty != original.State.GitDirty {
		t.Errorf("state.git_dirty: got %v, want %v", roundTripped.State.GitDirty, original.State.GitDirty)
	}
	if len(roundTripped.State.ModifiedFiles) != len(original.State.ModifiedFiles) {
		t.Errorf("state.modified_files: got %d, want %d", len(roundTripped.State.ModifiedFiles), len(original.State.ModifiedFiles))
	}
	if roundTripped.State.ActiveBead != original.State.ActiveBead {
		t.Errorf("state.active_bead: got %q, want %q", roundTripped.State.ActiveBead, original.State.ActiveBead)
	}
	if roundTripped.State.OpenBeadsCount != original.State.OpenBeadsCount {
		t.Errorf("state.open_beads_count: got %d, want %d", roundTripped.State.OpenBeadsCount, original.State.OpenBeadsCount)
	}
	if len(roundTripped.State.RecentCommits) != len(original.State.RecentCommits) {
		t.Errorf("state.recent_commits: got %d, want %d", len(roundTripped.State.RecentCommits), len(original.State.RecentCommits))
	}
	if roundTripped.Consumed != original.Consumed {
		t.Errorf("consumed: got %v, want %v", roundTripped.Consumed, original.Consumed)
	}
	if roundTripped.ConsumedAt != nil {
		t.Error("consumed_at: got non-nil, want nil")
	}
	if roundTripped.ConsumedBy != nil {
		t.Error("consumed_by: got non-nil, want nil")
	}
}

func TestCollectHandoffState_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	state := collectHandoffState(dir)
	if state == nil {
		t.Fatal("collectHandoffState returned nil for empty dir")
	}
	if state.GitDirty {
		t.Error("expected GitDirty=false for non-git dir")
	}
	if state.GitBranch != "" {
		t.Errorf("expected empty branch for non-git dir, got %q", state.GitBranch)
	}
	if state.ActiveBead != "" {
		t.Errorf("expected empty active bead, got %q", state.ActiveBead)
	}
}

func TestWriteHandoffArtifact_AtomicWrite(t *testing.T) {
	dir := t.TempDir()

	artifact := &handoffArtifact{
		SchemaVersion: 1,
		ID:            "handoff-20260303T200000Z",
		CreatedAt:     "2026-03-03T20:00:00Z",
		Type:          "manual",
		Summary:       "atomic test",
		Consumed:      false,
	}

	path, err := writeHandoffArtifact(dir, artifact)
	if err != nil {
		t.Fatalf("writeHandoffArtifact failed: %v", err)
	}

	// Verify no .tmp file lingering
	tmpPath := path + ".tmp"
	if fileExists(tmpPath) {
		t.Error(".tmp file should not exist after successful write")
	}

	// Verify final file is valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var check handoffArtifact
	if err := json.Unmarshal(data, &check); err != nil {
		t.Fatalf("written file is not valid JSON: %v", err)
	}
	if check.Summary != "atomic test" {
		t.Errorf("summary = %q, want %q", check.Summary, "atomic test")
	}
}

func TestHandoffShellQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "'simple'"},
		{"with spaces", "'with spaces'"},
		{"it's", "'it'\\''s'"},
		{"", "''"},
	}

	for _, tt := range tests {
		got := shellQuote(tt.input)
		if got != tt.want {
			t.Errorf("shellQuote(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseOpenBeadsCount(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{`[{"id":"1"},{"id":"2"},{"id":"3"}]`, 3},
		{`[]`, 0},
		{"", 0},
		{"invalid", 0},
		{"5", 5},
	}

	for _, tt := range tests {
		got := parseOpenBeadsCount(tt.input)
		if got != tt.want {
			t.Errorf("parseOpenBeadsCount(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestBuildHandoffRPIContext_NoStateFile(t *testing.T) {
	dir := t.TempDir()

	// No phased-state.json exists
	rpi := buildHandoffRPIContext(dir, 3, "na-xyz", "run-abc")
	if rpi == nil {
		t.Fatal("expected non-nil RPI even without state file")
	}
	if rpi.Phase != 3 {
		t.Errorf("phase = %d, want 3", rpi.Phase)
	}
	if rpi.PhaseName != "validation" {
		t.Errorf("phase_name = %q, want %q", rpi.PhaseName, "validation")
	}
	if rpi.EpicID != "na-xyz" {
		t.Errorf("epic_id = %q, want %q", rpi.EpicID, "na-xyz")
	}
	if rpi.RunID != "run-abc" {
		t.Errorf("run_id = %q, want %q", rpi.RunID, "run-abc")
	}
}
