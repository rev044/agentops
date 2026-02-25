package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ===========================================================================
// Coverage tests for rpi_status.go — targeting zero-coverage functions
// ===========================================================================

// --- buildRPIStatusOutput ---

func TestRPIStatusCov_BuildRPIStatusOutput_WithRegistryRuns(t *testing.T) {
	tmpDir := t.TempDir()

	writeRegistryRun(t, tmpDir, registryRunSpec{
		runID:  "bso-active",
		phase:  1,
		schema: 1,
		goal:   "active goal",
		hbAge:  30 * time.Second,
	})
	writeRegistryRun(t, tmpDir, registryRunSpec{
		runID:  "bso-hist",
		phase:  3,
		schema: 1,
		goal:   "done goal",
		hbAge:  0,
	})

	output := buildRPIStatusOutput(tmpDir)
	if output.Count != 2 {
		t.Errorf("expected Count=2, got %d", output.Count)
	}
	if len(output.Active) != 1 {
		t.Errorf("expected 1 active run, got %d", len(output.Active))
	}
	if len(output.Historical) != 1 {
		t.Errorf("expected 1 historical run, got %d", len(output.Historical))
	}
	// Runs should be the union
	if len(output.Runs) != 2 {
		t.Errorf("expected 2 combined runs, got %d", len(output.Runs))
	}
}

func TestRPIStatusCov_BuildRPIStatusOutput_WithLogAndLiveStatus(t *testing.T) {
	tmpDir := t.TempDir()

	// Create orchestration log
	rpiDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0755); err != nil {
		t.Fatal(err)
	}
	logContent := `[2026-02-15T10:00:00Z] [log-run1] start: goal="log test" from=discovery
[2026-02-15T10:05:00Z] [log-run1] complete: epic=ag-logtest verdicts=map[]
`
	if err := os.WriteFile(filepath.Join(rpiDir, "phased-orchestration.log"), []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create live status
	if err := os.WriteFile(filepath.Join(rpiDir, "live-status.md"), []byte("# Phase 2 running"), 0644); err != nil {
		t.Fatal(err)
	}

	output := buildRPIStatusOutput(tmpDir)
	if len(output.LogRuns) == 0 {
		t.Error("expected LogRuns to be populated from orchestration log")
	}
	if len(output.LiveStatuses) == 0 {
		t.Error("expected LiveStatuses to be populated from live-status.md")
	}
}

// --- renderRPIStatusTable ---

func TestRPIStatusCov_RenderRPIStatusTable_WithHistoricalRuns(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = old
	}()

	output := rpiStatusOutput{
		Historical: []rpiRunInfo{
			{RunID: "h-run", Goal: "hist goal", PhaseName: "validation", Status: "completed", Elapsed: "10m0s"},
		},
		Runs:  []rpiRunInfo{{RunID: "h-run"}},
		Count: 1,
	}

	if err := renderRPIStatusTable(t.TempDir(), output); err != nil {
		t.Errorf("renderRPIStatusTable with historical runs: %v", err)
	}
}

func TestRPIStatusCov_RenderRPIStatusTable_WithLogRuns(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = old
	}()

	output := rpiStatusOutput{
		LogRuns: []rpiRun{
			{
				RunID:    "lr1",
				Goal:     "log run goal",
				Phases:   []rpiPhaseEntry{{Name: "discovery"}},
				Status:   "completed",
				Retries:  map[string]int{"validation": 1},
				Duration: 15 * time.Minute,
			},
		},
		Count: 0,
	}

	if err := renderRPIStatusTable(t.TempDir(), output); err != nil {
		t.Errorf("renderRPIStatusTable with log runs: %v", err)
	}
}

func TestRPIStatusCov_RenderRPIStatusTable_WithLiveStatuses(t *testing.T) {
	cwd := t.TempDir()
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = old
	}()

	output := rpiStatusOutput{
		LiveStatuses: []liveStatusSnapshot{
			{Path: filepath.Join(cwd, "live-status.md"), Content: "# Phase 2 running"},
		},
		Count: 0,
	}

	if err := renderRPIStatusTable(cwd, output); err != nil {
		t.Errorf("renderRPIStatusTable with live statuses: %v", err)
	}
}

func TestRPIStatusCov_RenderRPIStatusTable_ActiveAndHistorical(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = old
	}()

	output := rpiStatusOutput{
		Active: []rpiRunInfo{
			{RunID: "a-run", Goal: "active", PhaseName: "discovery", Status: "running"},
		},
		Historical: []rpiRunInfo{
			{RunID: "h-run", Goal: "done", PhaseName: "validation", Status: "completed"},
		},
		Runs:  []rpiRunInfo{{RunID: "a-run"}, {RunID: "h-run"}},
		Count: 2,
	}

	if err := renderRPIStatusTable(t.TempDir(), output); err != nil {
		t.Errorf("renderRPIStatusTable with both active and historical: %v", err)
	}
}

// --- renderStateRunsSection ---

func TestRPIStatusCov_RenderStateRunsSection_WithReason(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = old
	}()

	runs := []rpiRunInfo{
		{RunID: "stale-1", Goal: "stale goal", PhaseName: "implementation", Status: "stale", Reason: "worktree missing", Elapsed: "30m0s"},
		{RunID: "stale-2", Goal: "another stale", PhaseName: "discovery", Status: "stale", Reason: "", Elapsed: "5m0s"},
	}
	renderStateRunsSection("Stale Runs", runs, "stale", true)
}

func TestRPIStatusCov_RenderStateRunsSection_WithoutReason(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = old
	}()

	runs := []rpiRunInfo{
		{RunID: "r1", Goal: "goal 1", PhaseName: "discovery", Status: "running", Elapsed: "5m0s"},
	}
	renderStateRunsSection("Active Runs", runs, "active", false)
}

// --- renderLogRunsSection ---

func TestRPIStatusCov_RenderLogRunsSection(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = old
	}()

	logRuns := []rpiRun{
		{
			RunID:    "lr1",
			Goal:     "first log run",
			Phases:   []rpiPhaseEntry{{Name: "start"}, {Name: "discovery"}, {Name: "complete"}},
			Status:   "completed",
			Verdicts: map[string]string{"vibe": "PASS"},
			Retries:  map[string]int{"validation": 2},
			Duration: 25 * time.Minute,
		},
		{
			RunID:    "lr2",
			Goal:     "second log run with a really long goal name that exceeds maximum",
			Phases:   nil,
			Status:   "running",
			Verdicts: map[string]string{},
			Retries:  map[string]int{},
			Duration: 0,
		},
	}
	renderLogRunsSection(logRuns)
}

// --- renderLiveStatusesSection ---

func TestRPIStatusCov_RenderLiveStatusesSection(t *testing.T) {
	cwd := t.TempDir()
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = old
	}()

	// Absolute path that can be made relative to cwd
	statusPath := filepath.Join(cwd, "live-status.md")
	renderLiveStatusesSection(cwd, []liveStatusSnapshot{
		{Path: statusPath, Content: "# Phase 2\nImplementation in progress\n"},
	})
}

func TestRPIStatusCov_RenderLiveStatusesSection_NonRelativePath(t *testing.T) {
	cwd := t.TempDir()
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		_ = w.Close()
		os.Stdout = old
	}()

	// Path in a different directory (can't be relativized)
	otherPath := filepath.Join(t.TempDir(), "other-live-status.md")
	renderLiveStatusesSection(cwd, []liveStatusSnapshot{
		{Path: otherPath, Content: "status content"},
	})
}

// --- totalRetries ---

func TestRPIStatusCov_TotalRetries(t *testing.T) {
	tests := []struct {
		name    string
		retries map[string]int
		want    int
	}{
		{"nil map", nil, 0},
		{"empty map", map[string]int{}, 0},
		{"single entry", map[string]int{"validation": 3}, 3},
		{"multiple entries", map[string]int{"validation": 2, "discovery": 1}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := totalRetries(tt.retries)
			if got != tt.want {
				t.Errorf("totalRetries = %d, want %d", got, tt.want)
			}
		})
	}
}

// --- joinVerdicts ---

func TestRPIStatusCov_JoinVerdicts(t *testing.T) {
	tests := []struct {
		name     string
		verdicts map[string]string
		wantLen  int // check length of joined result (ordering not deterministic)
	}{
		{"nil map", nil, 0},
		{"empty map", map[string]string{}, 0},
		{"single entry", map[string]string{"vibe": "PASS"}, len("vibe=PASS")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinVerdicts(tt.verdicts)
			if len(got) != tt.wantLen {
				t.Errorf("joinVerdicts length = %d, want %d (got %q)", len(got), tt.wantLen, got)
			}
		})
	}
}

func TestRPIStatusCov_JoinVerdicts_MultipleEntries(t *testing.T) {
	verdicts := map[string]string{"pre_mortem": "PASS", "vibe": "WARN"}
	got := joinVerdicts(verdicts)
	// Should contain both entries separated by comma
	if !strings.Contains(got, "pre_mortem=PASS") {
		t.Errorf("expected pre_mortem=PASS in %q", got)
	}
	if !strings.Contains(got, "vibe=WARN") {
		t.Errorf("expected vibe=WARN in %q", got)
	}
	if !strings.Contains(got, ",") {
		t.Errorf("expected comma separator in %q", got)
	}
}

// --- formattedLogRunStatus ---

func TestRPIStatusCov_FormattedLogRunStatus_CompletedWithVerdicts(t *testing.T) {
	run := rpiRun{
		Status:   "completed",
		Verdicts: map[string]string{"vibe": "PASS"},
	}
	got := formattedLogRunStatus(run)
	if !strings.Contains(got, "completed") {
		t.Errorf("expected 'completed' in %q", got)
	}
	if !strings.Contains(got, "vibe=PASS") {
		t.Errorf("expected verdict in %q", got)
	}
}

// --- clearScreen ---

func TestRPIStatusCov_ClearScreen(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	// Just verify it doesn't panic
	clearScreen()

	_ = w.Close()
	os.Stdout = old
}

// --- orchestrationLogState methods ---

func TestRPIStatusCov_OrchestrationLogState_ResolveRunID(t *testing.T) {
	tests := []struct {
		name      string
		runID     string
		phaseName string
		wantAnon  bool
	}{
		{"explicit run ID", "run-abc", "discovery", false},
		{"anonymous start", "", "start", true},
		{"anonymous non-start", "", "discovery", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newOrchestrationLogState()
			got := state.resolveRunID(tt.runID, tt.phaseName)
			if tt.wantAnon {
				if !strings.HasPrefix(got, "anon-") {
					t.Errorf("expected anon- prefix, got %q", got)
				}
			} else {
				if got != tt.runID {
					t.Errorf("expected %q, got %q", tt.runID, got)
				}
			}
		})
	}
}

func TestRPIStatusCov_OrchestrationLogState_ResolveRunID_Sequential(t *testing.T) {
	state := newOrchestrationLogState()

	// First anonymous start
	id1 := state.resolveRunID("", "start")
	if id1 != "anon-1" {
		t.Errorf("first start: expected anon-1, got %s", id1)
	}

	// Non-start anonymous should use current counter
	id2 := state.resolveRunID("", "discovery")
	if id2 != "anon-1" {
		t.Errorf("non-start: expected anon-1, got %s", id2)
	}

	// Second start bumps counter
	id3 := state.resolveRunID("", "start")
	if id3 != "anon-2" {
		t.Errorf("second start: expected anon-2, got %s", id3)
	}
}

func TestRPIStatusCov_OrchestrationLogState_GetOrCreateRun(t *testing.T) {
	state := newOrchestrationLogState()

	run1 := state.getOrCreateRun("run-1")
	if run1.RunID != "run-1" {
		t.Errorf("expected run-1, got %s", run1.RunID)
	}
	if run1.Status != "running" {
		t.Errorf("expected initial status running, got %s", run1.Status)
	}

	// Second call returns the same run
	run1Again := state.getOrCreateRun("run-1")
	if run1Again != run1 {
		t.Error("expected same pointer for existing run")
	}

	// Different run creates new
	run2 := state.getOrCreateRun("run-2")
	if run2.RunID != "run-2" {
		t.Errorf("expected run-2, got %s", run2.RunID)
	}
}

func TestRPIStatusCov_OrchestrationLogState_OrderedRuns(t *testing.T) {
	state := newOrchestrationLogState()
	state.getOrCreateRun("first")
	state.getOrCreateRun("second")
	state.getOrCreateRun("third")

	runs := state.orderedRuns()
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(runs))
	}
	if runs[0].RunID != "first" || runs[1].RunID != "second" || runs[2].RunID != "third" {
		t.Errorf("unexpected order: %s, %s, %s", runs[0].RunID, runs[1].RunID, runs[2].RunID)
	}
}

// --- applyOrchestrationLogEntry ---

func TestRPIStatusCov_ApplyOrchestrationLogEntry_StartPhase(t *testing.T) {
	run := &rpiRun{
		RunID:    "test",
		Verdicts: make(map[string]string),
		Retries:  make(map[string]int),
		Status:   "running",
	}
	entry := orchestrationLogEntry{
		PhaseName: "start",
		Details:   `goal="my goal" from=discovery`,
		Timestamp: "2026-02-15T10:00:00Z",
		ParsedAt:  time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC),
		HasTime:   true,
	}

	applyOrchestrationLogEntry(run, entry)

	if run.Goal != "my goal" {
		t.Errorf("expected goal 'my goal', got %q", run.Goal)
	}
	if run.StartedAt.IsZero() {
		t.Error("expected non-zero StartedAt")
	}
}

func TestRPIStatusCov_ApplyOrchestrationLogEntry_CompletePhase(t *testing.T) {
	startTime := time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC)
	completeTime := time.Date(2026, 2, 15, 10, 30, 0, 0, time.UTC)
	run := &rpiRun{
		RunID:     "test",
		StartedAt: startTime,
		Verdicts:  make(map[string]string),
		Retries:   make(map[string]int),
		Status:    "running",
	}
	entry := orchestrationLogEntry{
		PhaseName: "complete",
		Details:   "epic=ag-xyz verdicts=map[pre_mortem:PASS vibe:WARN]",
		Timestamp: "2026-02-15T10:30:00Z",
		ParsedAt:  completeTime,
		HasTime:   true,
	}

	applyOrchestrationLogEntry(run, entry)

	if run.Status != "completed" {
		t.Errorf("expected completed, got %s", run.Status)
	}
	if run.EpicID != "ag-xyz" {
		t.Errorf("expected epic ag-xyz, got %s", run.EpicID)
	}
	if run.Duration != 30*time.Minute {
		t.Errorf("expected 30m, got %s", run.Duration)
	}
	if run.Verdicts["pre_mortem"] != "PASS" {
		t.Errorf("expected pre_mortem PASS, got %s", run.Verdicts["pre_mortem"])
	}
}

// --- applyCompletePhase ---

func TestRPIStatusCov_ApplyCompletePhase_NoTime(t *testing.T) {
	run := &rpiRun{
		RunID:    "test",
		Verdicts: make(map[string]string),
		Status:   "running",
	}
	entry := orchestrationLogEntry{
		PhaseName: "complete",
		Details:   "epic=ag-notime verdicts=map[]",
		HasTime:   false,
	}
	applyCompletePhase(run, entry)

	if run.Status != "completed" {
		t.Errorf("expected completed, got %s", run.Status)
	}
	if !run.FinishedAt.IsZero() {
		t.Error("expected zero FinishedAt without time")
	}
}

// --- applyNonTerminalPhase ---

func TestRPIStatusCov_ApplyNonTerminalPhase_FailedDetails(t *testing.T) {
	run := &rpiRun{
		RunID:    "test",
		Verdicts: make(map[string]string),
		Retries:  make(map[string]int),
		Status:   "running",
	}
	entry := orchestrationLogEntry{
		PhaseName: "implementation",
		Details:   "FAILED: exit code 1",
		HasTime:   false,
	}
	applyNonTerminalPhase(run, entry)

	if run.Status != "failed" {
		t.Errorf("expected failed, got %s", run.Status)
	}
}

func TestRPIStatusCov_ApplyNonTerminalPhase_FatalDetails(t *testing.T) {
	run := &rpiRun{
		RunID:    "test",
		Verdicts: make(map[string]string),
		Retries:  make(map[string]int),
		Status:   "running",
	}
	entry := orchestrationLogEntry{
		PhaseName: "discovery",
		Details:   "FATAL: build prompt failed",
		HasTime:   false,
	}
	applyNonTerminalPhase(run, entry)

	if run.Status != "failed" {
		t.Errorf("expected failed, got %s", run.Status)
	}
}

func TestRPIStatusCov_ApplyNonTerminalPhase_RetryDetails(t *testing.T) {
	run := &rpiRun{
		RunID:    "test",
		Verdicts: make(map[string]string),
		Retries:  make(map[string]int),
		Status:   "running",
	}
	entry := orchestrationLogEntry{
		PhaseName: "validation",
		Details:   "RETRY attempt 2/3",
		HasTime:   false,
	}
	applyNonTerminalPhase(run, entry)

	if run.Retries["validation"] != 1 {
		t.Errorf("expected 1 retry, got %d", run.Retries["validation"])
	}
}

// --- updateFailureStatus ---

func TestRPIStatusCov_UpdateFailureStatus(t *testing.T) {
	tests := []struct {
		name     string
		details  string
		wantFail bool
	}{
		{"FAILED prefix", "FAILED: something", true},
		{"FATAL prefix", "FATAL: something", true},
		{"normal details", "completed in 5m", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := &rpiRun{Status: "running"}
			updateFailureStatus(run, tt.details)
			if tt.wantFail && run.Status != "failed" {
				t.Errorf("expected failed, got %s", run.Status)
			}
			if !tt.wantFail && run.Status != "running" {
				t.Errorf("expected running, got %s", run.Status)
			}
		})
	}
}

// --- updateRetryCount ---

func TestRPIStatusCov_UpdateRetryCount(t *testing.T) {
	run := &rpiRun{Retries: make(map[string]int)}
	updateRetryCount(run, "validation", "RETRY attempt 1/3")
	updateRetryCount(run, "validation", "RETRY attempt 2/3")
	updateRetryCount(run, "discovery", "something else") // not a retry

	if run.Retries["validation"] != 2 {
		t.Errorf("expected 2 retries for validation, got %d", run.Retries["validation"])
	}
	if run.Retries["discovery"] != 0 {
		t.Errorf("expected 0 retries for discovery, got %d", run.Retries["discovery"])
	}
}

// --- updateFinishedAtFromCompletedDuration ---

func TestRPIStatusCov_UpdateFinishedAtFromCompletedDuration(t *testing.T) {
	now := time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		details  string
		hasTime  bool
		parsedAt time.Time
		wantSet  bool
	}{
		{"completed duration with time", "completed in 5m0s", true, now, true},
		{"completed duration without time", "completed in 5m0s", false, time.Time{}, false},
		{"non-completed prefix", "started phase 2", true, now, false},
		{"invalid duration", "completed in not-a-duration", true, now, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := &rpiRun{}
			entry := orchestrationLogEntry{
				Details:  tt.details,
				HasTime:  tt.hasTime,
				ParsedAt: tt.parsedAt,
			}
			updateFinishedAtFromCompletedDuration(run, entry)
			if tt.wantSet && run.FinishedAt.IsZero() {
				t.Error("expected FinishedAt to be set")
			}
			if !tt.wantSet && !run.FinishedAt.IsZero() {
				t.Error("expected FinishedAt to be zero")
			}
		})
	}
}

// --- updateInlineVerdicts ---

func TestRPIStatusCov_UpdateInlineVerdicts(t *testing.T) {
	tests := []struct {
		name      string
		phaseName string
		details   string
		wantKey   string
		wantVal   string
	}{
		{"pre-mortem phase", "pre-mortem", "verdict: PASS", "pre_mortem", "PASS"},
		{"vibe phase", "vibe", "verdict: WARN", "vibe", "WARN"},
		{"post-mortem verdict in details", "validation", "post-mortem verdict: FAIL", "post_mortem", "FAIL"},
		{"pre-mortem verdict in details", "discovery", "pre-mortem verdict: WARN", "pre_mortem", "WARN"},
		{"vibe verdict in details", "validation", "vibe verdict: PASS", "vibe", "PASS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := &rpiRun{Verdicts: make(map[string]string)}
			updateInlineVerdicts(run, tt.phaseName, tt.details)
			if run.Verdicts[tt.wantKey] != tt.wantVal {
				t.Errorf("expected %s=%s, got %s", tt.wantKey, tt.wantVal, run.Verdicts[tt.wantKey])
			}
		})
	}
}

func TestRPIStatusCov_UpdateInlineVerdicts_NoVerdict(t *testing.T) {
	run := &rpiRun{Verdicts: make(map[string]string)}
	updateInlineVerdicts(run, "discovery", "completed in 5m0s")
	if len(run.Verdicts) != 0 {
		t.Errorf("expected no verdicts, got %d", len(run.Verdicts))
	}
}

// --- discoverLogRuns ---

func TestRPIStatusCov_DiscoverLogRuns_NoLogs(t *testing.T) {
	runs := discoverLogRuns(t.TempDir())
	if len(runs) != 0 {
		t.Errorf("expected 0 runs for empty dir, got %d", len(runs))
	}
}

func TestRPIStatusCov_DiscoverLogRuns_WithSiblingLog(t *testing.T) {
	parent := t.TempDir()
	cwd := filepath.Join(parent, "repo")
	sibling := filepath.Join(parent, "repo-rpi-abc")

	for _, dir := range []string{cwd, sibling} {
		rpiDir := filepath.Join(dir, ".agents", "rpi")
		if err := os.MkdirAll(rpiDir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	logContent := `[2026-02-15T10:00:00Z] [sib1] start: goal="sibling" from=discovery
[2026-02-15T10:05:00Z] [sib1] complete: epic=ag-sib verdicts=map[]
`
	if err := os.WriteFile(filepath.Join(sibling, ".agents", "rpi", "phased-orchestration.log"), []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}

	runs := discoverLogRuns(cwd)
	if len(runs) != 1 {
		t.Errorf("expected 1 run from sibling, got %d", len(runs))
	}
}

// --- discoverLiveStatuses dedup ---

func TestRPIStatusCov_DiscoverLiveStatuses_Dedup(t *testing.T) {
	parent := t.TempDir()
	cwd := filepath.Join(parent, "repo")
	rpiDir := filepath.Join(cwd, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rpiDir, "live-status.md"), []byte("status"), 0644); err != nil {
		t.Fatal(err)
	}

	// Call twice — same file should not duplicate
	got := discoverLiveStatuses(cwd)
	if len(got) != 1 {
		t.Errorf("expected 1 deduped live status, got %d", len(got))
	}
}

// --- completedPhaseNumber ---

func TestRPIStatusCov_CompletedPhaseNumber(t *testing.T) {
	tests := []struct {
		schema int
		want   int
	}{
		{0, 6},
		{1, 3},
		{2, 3},
	}
	for _, tt := range tests {
		state := phasedState{SchemaVersion: tt.schema}
		got := completedPhaseNumber(state)
		if got != tt.want {
			t.Errorf("completedPhaseNumber(schema=%d) = %d, want %d", tt.schema, got, tt.want)
		}
	}
}

// --- displayPhaseName ---

func TestRPIStatusCov_DisplayPhaseName_LegacyUnknown(t *testing.T) {
	state := phasedState{SchemaVersion: 0, Phase: 99}
	got := displayPhaseName(state)
	if got != "phase-99" {
		t.Errorf("expected phase-99, got %s", got)
	}
}

// --- tryAddSearchRoot ---

func TestRPIStatusCov_TryAddSearchRoot(t *testing.T) {
	dir := t.TempDir()
	seen := make(map[string]struct{})
	roots := []string{}

	// Add valid directory
	tryAddSearchRoot(dir, seen, &roots)
	if len(roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(roots))
	}

	// Duplicate should not add
	tryAddSearchRoot(dir, seen, &roots)
	if len(roots) != 1 {
		t.Errorf("expected still 1 root after duplicate, got %d", len(roots))
	}

	// Empty path ignored
	tryAddSearchRoot("", seen, &roots)
	if len(roots) != 1 {
		t.Errorf("expected still 1 root after empty, got %d", len(roots))
	}

	// Non-existent path ignored
	tryAddSearchRoot("/nonexistent/path/xyz", seen, &roots)
	if len(roots) != 1 {
		t.Errorf("expected still 1 root after nonexistent, got %d", len(roots))
	}
}

// --- collectSearchRoots ---

func TestRPIStatusCov_CollectSearchRoots(t *testing.T) {
	parent := t.TempDir()
	cwd := filepath.Join(parent, "myrepo")
	sibling := filepath.Join(parent, "myrepo-rpi-xyz")

	for _, dir := range []string{cwd, sibling} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	roots := collectSearchRoots(cwd)
	if len(roots) < 1 {
		t.Error("expected at least cwd in roots")
	}

	// Verify cwd is in roots
	found := false
	for _, r := range roots {
		if r == cwd {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected cwd %s in roots %v", cwd, roots)
	}
}

// --- normalizeSearchRootPath ---

func TestRPIStatusCov_NormalizeSearchRootPath_Existing(t *testing.T) {
	dir := t.TempDir()
	got := normalizeSearchRootPath(dir)
	if got == "" {
		t.Error("expected non-empty path")
	}
}

func TestRPIStatusCov_NormalizeSearchRootPath_NonExistent(t *testing.T) {
	got := normalizeSearchRootPath("/tmp/nonexistent-coverage-test-xyz")
	if got == "" {
		t.Error("expected non-empty result")
	}
}

// --- scanRegistryRuns with corrupt state ---

func TestRPIStatusCov_ScanRegistryRuns_CorruptState(t *testing.T) {
	tmpDir := t.TempDir()
	runDir := filepath.Join(tmpDir, ".agents", "rpi", "runs", "corrupt-run")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, phasedStateFile), []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	runs := scanRegistryRuns(tmpDir)
	if len(runs) != 0 {
		t.Errorf("expected 0 runs for corrupt state, got %d", len(runs))
	}
}

func TestRPIStatusCov_ScanRegistryRuns_FileNotDir(t *testing.T) {
	tmpDir := t.TempDir()
	runsDir := filepath.Join(tmpDir, ".agents", "rpi", "runs")
	if err := os.MkdirAll(runsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write a file, not a directory, inside runs/
	if err := os.WriteFile(filepath.Join(runsDir, "not-a-dir"), []byte("file"), 0644); err != nil {
		t.Fatal(err)
	}

	runs := scanRegistryRuns(tmpDir)
	if len(runs) != 0 {
		t.Errorf("expected 0 runs for non-directory entry, got %d", len(runs))
	}
}

// --- discoverRPIRuns (legacy compat) ---

func TestRPIStatusCov_DiscoverRPIRuns_EmptyDir(t *testing.T) {
	runs := discoverRPIRuns(t.TempDir())
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

// --- writeRPIStatusJSON ---

func TestRPIStatusCov_WriteRPIStatusJSON_Fields(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	output := rpiStatusOutput{
		Active: []rpiRunInfo{{RunID: "a1"}},
		Count:  1,
	}
	_ = writeRPIStatusJSON(output)

	_ = w.Close()
	os.Stdout = old

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	var decoded rpiStatusOutput
	if err := json.Unmarshal(buf[:n], &decoded); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if decoded.Count != 1 {
		t.Errorf("expected count 1, got %d", decoded.Count)
	}
}

// --- parseOrchestrationLogLine ---

func TestRPIStatusCov_ParseOrchestrationLogLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantOK  bool
		runID   string
		phase   string
		details string
	}{
		{
			"new format",
			"[2026-02-15T10:00:00Z] [run1] start: goal=\"test\" from=discovery",
			true, "run1", "start", `goal="test" from=discovery`,
		},
		{
			"old format",
			"[2026-02-15T10:00:00Z] discovery: completed in 5m0s",
			true, "", "discovery", "completed in 5m0s",
		},
		{
			"garbage line",
			"this is not a log line",
			false, "", "", "",
		},
		{
			"empty line",
			"",
			false, "", "", "",
		},
		{
			"non-RFC3339 timestamp",
			"[not-a-time] [r1] phase: details",
			true, "r1", "phase", "details",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, ok := parseOrchestrationLogLine(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if entry.RunID != tt.runID {
				t.Errorf("runID = %q, want %q", entry.RunID, tt.runID)
			}
			if entry.PhaseName != tt.phase {
				t.Errorf("phase = %q, want %q", entry.PhaseName, tt.phase)
			}
			if entry.Details != tt.details {
				t.Errorf("details = %q, want %q", entry.Details, tt.details)
			}
		})
	}
}

func TestRPIStatusCov_ParseOrchestrationLogLine_RFC3339Time(t *testing.T) {
	entry, ok := parseOrchestrationLogLine("[2026-02-15T10:00:00Z] [r1] start: test")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if !entry.HasTime {
		t.Error("expected HasTime=true for RFC3339 timestamp")
	}
	if entry.ParsedAt.IsZero() {
		t.Error("expected non-zero ParsedAt")
	}
}

func TestRPIStatusCov_ParseOrchestrationLogLine_NonRFC3339Time(t *testing.T) {
	entry, ok := parseOrchestrationLogLine("[2026/02/15 10:00] [r1] start: test")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if entry.HasTime {
		t.Error("expected HasTime=false for non-RFC3339 timestamp")
	}
}

// --- newOrchestrationLogState ---

func TestRPIStatusCov_NewOrchestrationLogState(t *testing.T) {
	state := newOrchestrationLogState()
	if state.runMap == nil {
		t.Error("expected non-nil runMap")
	}
	if state.anonymousCounter != 0 {
		t.Errorf("expected 0 counter, got %d", state.anonymousCounter)
	}
}
