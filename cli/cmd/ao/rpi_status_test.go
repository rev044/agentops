package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRPIStatusDiscovery(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}

	state := map[string]interface{}{
		"run_id":     "abc123def456",
		"goal":       "test goal",
		"phase":      3,
		"epic_id":    "ag-test",
		"started_at": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
	}
	data, _ := json.Marshal(state)
	if err := os.WriteFile(filepath.Join(stateDir, "phased-state.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	run, ok := loadRPIRun(tmpDir)
	if !ok {
		t.Fatal("expected loadRPIRun to return a run")
	}

	if run.RunID != "abc123def456" {
		t.Errorf("expected run_id abc123def456, got %s", run.RunID)
	}
	if run.PhaseName != "pre-mortem" {
		t.Errorf("expected phase pre-mortem, got %s", run.PhaseName)
	}
	if run.EpicID != "ag-test" {
		t.Errorf("expected epic ag-test, got %s", run.EpicID)
	}
	if run.Goal != "test goal" {
		t.Errorf("expected goal 'test goal', got %s", run.Goal)
	}
	// Without tmux, status should be "unknown" (phase 3 < 6, no tmux session)
	if run.Status != "unknown" {
		t.Errorf("expected status 'unknown', got %s", run.Status)
	}
}

func TestRPIStatusMissingState(t *testing.T) {
	tmpDir := t.TempDir()
	_, ok := loadRPIRun(tmpDir)
	if ok {
		t.Fatal("expected loadRPIRun to return false for empty dir")
	}
}

func TestRPIStatusCorruptState(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "phased-state.json"), []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, ok := loadRPIRun(tmpDir)
	if ok {
		t.Fatal("expected loadRPIRun to return false for corrupt state")
	}
}

func TestRPIStatusPhaseNames(t *testing.T) {
	tests := []struct {
		phase    int
		expected string
	}{
		{1, "research"},
		{2, "plan"},
		{3, "pre-mortem"},
		{4, "crank"},
		{5, "vibe"},
		{6, "post-mortem"},
		{99, "phase-99"},
	}

	for _, tt := range tests {
		tmpDir := t.TempDir()
		stateDir := filepath.Join(tmpDir, ".agents", "rpi")
		if err := os.MkdirAll(stateDir, 0755); err != nil {
			t.Fatal(err)
		}

		state := map[string]interface{}{
			"run_id": "test-run",
			"phase":  tt.phase,
		}
		data, _ := json.Marshal(state)
		if err := os.WriteFile(filepath.Join(stateDir, "phased-state.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		run, ok := loadRPIRun(tmpDir)
		if !ok {
			t.Fatalf("expected run for phase %d", tt.phase)
		}
		if run.PhaseName != tt.expected {
			t.Errorf("phase %d: expected %s, got %s", tt.phase, tt.expected, run.PhaseName)
		}
	}
}

func TestRPIStatusEmptyRunID(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}

	state := map[string]interface{}{
		"goal":  "no run id",
		"phase": 1,
	}
	data, _ := json.Marshal(state)
	if err := os.WriteFile(filepath.Join(stateDir, "phased-state.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	_, ok := loadRPIRun(tmpDir)
	if ok {
		t.Fatal("expected loadRPIRun to return false when run_id is empty")
	}
}

func TestRPIStatusCompletedPhase6(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}

	state := map[string]interface{}{
		"run_id":     "completed-run",
		"goal":       "finished goal",
		"phase":      6,
		"started_at": time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
	}
	data, _ := json.Marshal(state)
	if err := os.WriteFile(filepath.Join(stateDir, "phased-state.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	run, ok := loadRPIRun(tmpDir)
	if !ok {
		t.Fatal("expected loadRPIRun to return a run")
	}
	if run.Status != "completed" {
		t.Errorf("expected status 'completed' for phase 6, got %s", run.Status)
	}
}

func TestRPIStatusDetermineRunStatus(t *testing.T) {
	// No tmux in test environment, so all sessions are "not alive"
	tests := []struct {
		name     string
		phase    int
		expected string
	}{
		{"phase 1 no tmux", 1, "unknown"},
		{"phase 3 no tmux", 3, "unknown"},
		{"phase 6 completed", 6, "completed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := phasedState{
				RunID: "test-run",
				Phase: tt.phase,
			}
			status := determineRunStatus(state)
			if status != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, status)
			}
		})
	}
}

func TestRPIStatusSiblingDiscovery(t *testing.T) {
	// Create a parent directory with cwd and a sibling worktree
	parent := t.TempDir()
	cwd := filepath.Join(parent, "myrepo")
	sibling := filepath.Join(parent, "myrepo-rpi-abc123")

	for _, dir := range []string{cwd, sibling} {
		stateDir := filepath.Join(dir, ".agents", "rpi")
		if err := os.MkdirAll(stateDir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Write state in cwd
	cwdState := map[string]interface{}{
		"run_id": "main-run",
		"goal":   "main goal",
		"phase":  2,
	}
	cwdData, _ := json.Marshal(cwdState)
	if err := os.WriteFile(filepath.Join(cwd, ".agents", "rpi", "phased-state.json"), cwdData, 0644); err != nil {
		t.Fatal(err)
	}

	// Write state in sibling
	sibState := map[string]interface{}{
		"run_id": "sibling-run",
		"goal":   "sibling goal",
		"phase":  4,
	}
	sibData, _ := json.Marshal(sibState)
	if err := os.WriteFile(filepath.Join(sibling, ".agents", "rpi", "phased-state.json"), sibData, 0644); err != nil {
		t.Fatal(err)
	}

	runs := discoverRPIRuns(cwd)
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}

	// Verify both runs are present
	foundMain, foundSibling := false, false
	for _, r := range runs {
		if r.RunID == "main-run" {
			foundMain = true
		}
		if r.RunID == "sibling-run" {
			foundSibling = true
		}
	}
	if !foundMain {
		t.Error("expected to find main-run")
	}
	if !foundSibling {
		t.Error("expected to find sibling-run")
	}
}

// --- Log parser tests ---

func TestRPIStatusParseLogNewFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "phased-orchestration.log")

	logContent := `[2026-02-15T10:00:00Z] [abc123] start: goal="add user auth" from=research
[2026-02-15T10:05:00Z] [abc123] research: completed in 5m0s
[2026-02-15T10:10:00Z] [abc123] plan: completed in 5m0s
[2026-02-15T10:15:00Z] [abc123] pre-mortem: completed in 5m0s
[2026-02-15T10:25:00Z] [abc123] crank: completed in 10m0s
[2026-02-15T10:30:00Z] [abc123] vibe: completed in 5m0s
[2026-02-15T10:35:00Z] [abc123] post-mortem: completed in 5m0s
[2026-02-15T10:35:00Z] [abc123] complete: epic=ag-test verdicts=map[pre_mortem:PASS vibe:PASS]
`
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}

	runs, err := parseOrchestrationLog(logPath)
	if err != nil {
		t.Fatalf("parseOrchestrationLog failed: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	run := runs[0]
	if run.RunID != "abc123" {
		t.Errorf("expected run_id abc123, got %s", run.RunID)
	}
	if run.Goal != "add user auth" {
		t.Errorf("expected goal 'add user auth', got %q", run.Goal)
	}
	if run.Status != "completed" {
		t.Errorf("expected status completed, got %s", run.Status)
	}
	if run.EpicID != "ag-test" {
		t.Errorf("expected epic ag-test, got %s", run.EpicID)
	}
	if len(run.Phases) != 8 {
		t.Errorf("expected 8 phase entries, got %d", len(run.Phases))
	}
	// Check verdicts
	if run.Verdicts["pre_mortem"] != "PASS" {
		t.Errorf("expected pre_mortem verdict PASS, got %q", run.Verdicts["pre_mortem"])
	}
	if run.Verdicts["vibe"] != "PASS" {
		t.Errorf("expected vibe verdict PASS, got %q", run.Verdicts["vibe"])
	}
	// Check duration (35 min from start to complete)
	expectedDur := 35 * time.Minute
	if run.Duration != expectedDur {
		t.Errorf("expected duration %s, got %s", expectedDur, run.Duration)
	}
}

func TestRPIStatusParseLogOldFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "phased-orchestration.log")

	logContent := `[2026-02-15T09:00:00Z] start: goal="fix typo" from=research
[2026-02-15T09:02:00Z] research: completed in 2m0s
[2026-02-15T09:04:00Z] plan: completed in 2m0s
[2026-02-15T09:06:00Z] complete: epic=ag-typo verdicts=map[vibe:PASS]
`
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}

	runs, err := parseOrchestrationLog(logPath)
	if err != nil {
		t.Fatalf("parseOrchestrationLog failed: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	run := runs[0]
	if run.Goal != "fix typo" {
		t.Errorf("expected goal 'fix typo', got %q", run.Goal)
	}
	if run.Status != "completed" {
		t.Errorf("expected status completed, got %s", run.Status)
	}
	if run.EpicID != "ag-typo" {
		t.Errorf("expected epic ag-typo, got %s", run.EpicID)
	}
}

func TestRPIStatusParseLogMultipleRuns(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "phased-orchestration.log")

	logContent := `[2026-02-15T10:00:00Z] [run1] start: goal="first goal" from=research
[2026-02-15T10:05:00Z] [run1] research: completed in 5m0s
[2026-02-15T10:05:00Z] [run2] start: goal="second goal" from=research
[2026-02-15T10:10:00Z] [run1] complete: epic=ag-first verdicts=map[]
[2026-02-15T10:12:00Z] [run2] research: completed in 7m0s
[2026-02-15T10:15:00Z] [run2] complete: epic=ag-second verdicts=map[vibe:WARN]
`
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}

	runs, err := parseOrchestrationLog(logPath)
	if err != nil {
		t.Fatalf("parseOrchestrationLog failed: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}

	if runs[0].RunID != "run1" || runs[0].Goal != "first goal" {
		t.Errorf("run1 mismatch: id=%s goal=%q", runs[0].RunID, runs[0].Goal)
	}
	if runs[1].RunID != "run2" || runs[1].Goal != "second goal" {
		t.Errorf("run2 mismatch: id=%s goal=%q", runs[1].RunID, runs[1].Goal)
	}
	if runs[0].Status != "completed" {
		t.Errorf("run1 expected completed, got %s", runs[0].Status)
	}
	if runs[1].Status != "completed" {
		t.Errorf("run2 expected completed, got %s", runs[1].Status)
	}
	if runs[1].Verdicts["vibe"] != "WARN" {
		t.Errorf("run2 expected vibe verdict WARN, got %q", runs[1].Verdicts["vibe"])
	}
}

func TestRPIStatusParseLogRetries(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "phased-orchestration.log")

	logContent := `[2026-02-15T10:00:00Z] [abc] start: goal="retry test" from=research
[2026-02-15T10:05:00Z] [abc] research: completed in 5m0s
[2026-02-15T10:10:00Z] [abc] pre-mortem: RETRY attempt 2/3
[2026-02-15T10:15:00Z] [abc] pre-mortem: RETRY attempt 3/3
[2026-02-15T10:20:00Z] [abc] pre-mortem: completed in 10m0s
[2026-02-15T10:25:00Z] [abc] complete: epic=ag-retry verdicts=map[pre_mortem:PASS]
`
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}

	runs, err := parseOrchestrationLog(logPath)
	if err != nil {
		t.Fatalf("parseOrchestrationLog failed: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	run := runs[0]
	if run.Retries["pre-mortem"] != 2 {
		t.Errorf("expected 2 retries for pre-mortem, got %d", run.Retries["pre-mortem"])
	}
	if run.Status != "completed" {
		t.Errorf("expected status completed, got %s", run.Status)
	}
}

func TestRPIStatusParseLogFailed(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "phased-orchestration.log")

	logContent := `[2026-02-15T10:00:00Z] [fail1] start: goal="failing run" from=research
[2026-02-15T10:05:00Z] [fail1] research: completed in 5m0s
[2026-02-15T10:10:00Z] [fail1] crank: FAILED: claude exited with code 1
`
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}

	runs, err := parseOrchestrationLog(logPath)
	if err != nil {
		t.Fatalf("parseOrchestrationLog failed: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	if runs[0].Status != "failed" {
		t.Errorf("expected status failed, got %s", runs[0].Status)
	}
}

func TestRPIStatusParseLogMissingFile(t *testing.T) {
	_, err := parseOrchestrationLog("/nonexistent/path/log")
	if err == nil {
		t.Fatal("expected error for missing log file")
	}
}

func TestRPIStatusParseLogEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "empty.log")
	if err := os.WriteFile(logPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	runs, err := parseOrchestrationLog(logPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs for empty log, got %d", len(runs))
	}
}

func TestRPIStatusParseLogGarbageLines(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "garbage.log")

	logContent := `this is not a log line
also not valid
[2026-02-15T10:00:00Z] [good] start: goal="valid" from=research
random noise here
[2026-02-15T10:05:00Z] [good] complete: epic=ag-good verdicts=map[]
more garbage
`
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}

	runs, err := parseOrchestrationLog(logPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Goal != "valid" {
		t.Errorf("expected goal 'valid', got %q", runs[0].Goal)
	}
}

func TestRPIStatusExtractGoalFromDetails(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`goal="add auth" from=research`, "add auth"},
		{`goal="" from=research`, ""},
		{`just plain text`, "just plain text"},
	}
	for _, tt := range tests {
		got := extractGoalFromDetails(tt.input)
		if got != tt.expected {
			t.Errorf("extractGoalFromDetails(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRPIStatusExtractEpicFromDetails(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`epic=ag-test verdicts=map[]`, "ag-test"},
		{`epic=ag-1234 verdicts=map[vibe:PASS]`, "ag-1234"},
		{`no epic here`, ""},
	}
	for _, tt := range tests {
		got := extractEpicFromDetails(tt.input)
		if got != tt.expected {
			t.Errorf("extractEpicFromDetails(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRPIStatusExtractVerdictsFromDetails(t *testing.T) {
	verdicts := make(map[string]string)
	extractVerdictsFromDetails("epic=ag-test verdicts=map[pre_mortem:PASS vibe:WARN]", verdicts)
	if verdicts["pre_mortem"] != "PASS" {
		t.Errorf("expected pre_mortem=PASS, got %q", verdicts["pre_mortem"])
	}
	if verdicts["vibe"] != "WARN" {
		t.Errorf("expected vibe=WARN, got %q", verdicts["vibe"])
	}
}

func TestRPIStatusExtractVerdictsEmpty(t *testing.T) {
	verdicts := make(map[string]string)
	extractVerdictsFromDetails("epic=ag-test verdicts=map[]", verdicts)
	if len(verdicts) != 0 {
		t.Errorf("expected 0 verdicts, got %d", len(verdicts))
	}
}

func TestRPIStatusExtractInlineVerdict(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"completed with PASS", "PASS"},
		{"gate FAIL after retry", "FAIL"},
		{"verdict WARN for complexity", "WARN"},
		{"nothing here", ""},
	}
	for _, tt := range tests {
		got := extractInlineVerdict(tt.input)
		if got != tt.expected {
			t.Errorf("extractInlineVerdict(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRPIStatusParseLogRunningStatus(t *testing.T) {
	// A run with start but no complete entry should be "running"
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "running.log")

	logContent := `[2026-02-15T10:00:00Z] [active1] start: goal="in progress" from=research
[2026-02-15T10:05:00Z] [active1] research: completed in 5m0s
[2026-02-15T10:10:00Z] [active1] plan: completed in 5m0s
`
	if err := os.WriteFile(logPath, []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}

	runs, err := parseOrchestrationLog(logPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Status != "running" {
		t.Errorf("expected status 'running' for incomplete run, got %s", runs[0].Status)
	}
}
