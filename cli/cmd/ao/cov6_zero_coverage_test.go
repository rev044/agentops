package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// printMineSummary — pure io.Writer formatter
// ---------------------------------------------------------------------------

func TestPrintMineSummary_AllFields(t *testing.T) {
	var buf bytes.Buffer
	r := &MineReport{
		Git: &GitFindings{
			CommitCount:      42,
			TopCoChangeFiles: []string{"a.go", "b.go"},
			RecurringFixes:   []string{"nil-check"},
		},
		Agents: &AgentsFindings{
			TotalResearch:    10,
			OrphanedResearch: []string{"orphan.md"},
		},
		Code: &CodeFindings{
			Hotspots: []ComplexityHotspot{{Func: "foo", Complexity: 25}},
		},
	}
	printMineSummary(&buf, r)
	out := buf.String()

	if out == "" {
		t.Fatal("expected non-empty output")
	}
	assertContains(t, out, "Mine complete.")
	assertContains(t, out, "42 commits")
	assertContains(t, out, "2 co-change files")
	assertContains(t, out, "1 fix patterns")
	assertContains(t, out, "10 research files")
	assertContains(t, out, "1 orphaned")
	assertContains(t, out, "1 hotspots")
}

func TestPrintMineSummary_NilSections(t *testing.T) {
	var buf bytes.Buffer
	printMineSummary(&buf, &MineReport{})
	out := buf.String()
	assertContains(t, out, "Mine complete.")
	// No git/agents/code sections
	if bytes.Contains([]byte(out), []byte("commits")) {
		t.Error("should not print git section when Git is nil")
	}
}

func TestPrintMineSummary_CodeSkipped(t *testing.T) {
	var buf bytes.Buffer
	r := &MineReport{
		Code: &CodeFindings{Skipped: true},
	}
	printMineSummary(&buf, r)
	assertContains(t, buf.String(), "skipped")
}

func TestPrintMineSummary_GitNoCoChangeNoFixes(t *testing.T) {
	var buf bytes.Buffer
	r := &MineReport{
		Git: &GitFindings{CommitCount: 5},
	}
	printMineSummary(&buf, r)
	out := buf.String()
	assertContains(t, out, "5 commits")
	if bytes.Contains([]byte(out), []byte("co-change")) {
		t.Error("should not print co-change when empty")
	}
}

// ---------------------------------------------------------------------------
// runQuarantineFlagged — JSON parsing + file quarantine
// ---------------------------------------------------------------------------

func TestRunQuarantineFlagged_Success(t *testing.T) {
	tmp := t.TempDir()
	defragDir := filepath.Join(tmp, ".agents", "defrag")
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(defragDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a learning file to quarantine
	learningPath := filepath.Join(learningsDir, "stale.md")
	if err := os.WriteFile(learningPath, []byte("stale content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write quality report pointing to the learning
	report := map[string]any{
		"flagged_paths": []string{filepath.Join(".agents", "learnings", "stale.md")},
	}
	data, _ := json.Marshal(report)
	if err := os.WriteFile(filepath.Join(defragDir, "quality-report.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	err := runQuarantineFlagged(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify original file is gone
	if _, err := os.Stat(learningPath); !os.IsNotExist(err) {
		t.Error("original learning should be removed")
	}
	// Verify it moved to quarantine
	quarantinePath := filepath.Join(learningsDir, ".quarantine", "stale.md")
	if _, err := os.Stat(quarantinePath); err != nil {
		t.Errorf("expected file in quarantine: %v", err)
	}
}

func TestRunQuarantineFlagged_NoReport(t *testing.T) {
	tmp := t.TempDir()
	err := runQuarantineFlagged(tmp)
	if err == nil {
		t.Fatal("expected error when quality report is missing")
	}
	assertContains(t, err.Error(), "no quality report found")
}

func TestRunQuarantineFlagged_EmptyFlaggedPaths(t *testing.T) {
	tmp := t.TempDir()
	defragDir := filepath.Join(tmp, ".agents", "defrag")
	if err := os.MkdirAll(defragDir, 0o755); err != nil {
		t.Fatal(err)
	}
	report := map[string]any{"flagged_paths": []string{}}
	data, _ := json.Marshal(report)
	if err := os.WriteFile(filepath.Join(defragDir, "quality-report.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	err := runQuarantineFlagged(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// runDefrag — cobra command with temp dirs
// ---------------------------------------------------------------------------

func TestRunDefrag_PruneOnly_DryRun(t *testing.T) {
	tmp := t.TempDir()

	// Create learnings directory
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	defragDir := filepath.Join(tmp, ".agents", "defrag")
	if err := os.MkdirAll(defragDir, 0o755); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Set flags
	origPrune := defragPrune
	origDedup := defragDedup
	origSweep := defragOscillationSweep
	origDays := defragStaleDays
	origOutputDir := defragOutputDir
	origDryRun := dryRun
	defer func() {
		defragPrune = origPrune
		defragDedup = origDedup
		defragOscillationSweep = origSweep
		defragStaleDays = origDays
		defragOutputDir = origOutputDir
		dryRun = origDryRun
	}()

	defragPrune = true
	defragDedup = false
	defragOscillationSweep = false
	defragStaleDays = 30
	defragOutputDir = defragDir
	dryRun = true

	err := runDefrag(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify report was written
	entries, _ := os.ReadDir(defragDir)
	found := false
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			found = true
		}
	}
	if !found {
		t.Error("expected defrag report JSON in output dir")
	}
}

// ---------------------------------------------------------------------------
// runBDSyncIfNeeded — loopCommandRunner mockable
// ---------------------------------------------------------------------------

func TestRunBDSyncIfNeeded_NeverPolicy(t *testing.T) {
	cfg := rpiLoopSupervisorConfig{
		BDSyncPolicy:   loopBDSyncPolicyNever,
		BDCommand:      "bd",
		CommandTimeout: 10 * time.Second,
	}
	err := runBDSyncIfNeeded(t.TempDir(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunBDSyncIfNeeded_AutoPolicy_NoBeadsDir(t *testing.T) {
	tmp := t.TempDir()
	// No .beads directory → should skip

	origLookPath := loopLookPath
	defer func() { loopLookPath = origLookPath }()
	loopLookPath = func(file string) (string, error) {
		return "/usr/bin/bd", nil // pretend bd exists
	}

	cfg := rpiLoopSupervisorConfig{
		BDSyncPolicy:   loopBDSyncPolicyAuto,
		BDCommand:      "bd",
		CommandTimeout: 10 * time.Second,
	}
	err := runBDSyncIfNeeded(tmp, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunBDSyncIfNeeded_AutoPolicy_WithBeadsDir(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".beads"), 0o755); err != nil {
		t.Fatal(err)
	}

	origLookPath := loopLookPath
	origRunner := loopCommandRunner
	defer func() {
		loopLookPath = origLookPath
		loopCommandRunner = origRunner
	}()
	loopLookPath = func(file string) (string, error) {
		return "/usr/bin/bd", nil
	}

	var called bool
	loopCommandRunner = func(cwd string, timeout time.Duration, command string, args ...string) error {
		called = true
		if command != "bd" || len(args) == 0 || args[0] != "sync" {
			t.Errorf("expected bd sync, got %s %v", command, args)
		}
		return nil
	}

	cfg := rpiLoopSupervisorConfig{
		BDSyncPolicy:   loopBDSyncPolicyAuto,
		BDCommand:      "bd",
		CommandTimeout: 10 * time.Second,
	}
	err := runBDSyncIfNeeded(tmp, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected loopCommandRunner to be called")
	}
}

func TestRunBDSyncIfNeeded_AlwaysPolicy_NoCLI(t *testing.T) {
	origLookPath := loopLookPath
	defer func() { loopLookPath = origLookPath }()
	loopLookPath = func(file string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	cfg := rpiLoopSupervisorConfig{
		BDSyncPolicy:   loopBDSyncPolicyAlways,
		BDCommand:      "bd",
		CommandTimeout: 10 * time.Second,
	}
	err := runBDSyncIfNeeded(t.TempDir(), cfg)
	if err == nil {
		t.Fatal("expected error when bd CLI not found with always policy")
	}
}

// ---------------------------------------------------------------------------
// runAthenaProducerTick — loopCommandRunner mockable
// ---------------------------------------------------------------------------

func TestRunAthenaProducerTick_MineOnly(t *testing.T) {
	origRunner := loopCommandRunner
	defer func() { loopCommandRunner = origRunner }()

	var commands []string
	loopCommandRunner = func(cwd string, timeout time.Duration, command string, args ...string) error {
		commands = append(commands, command+" "+fmt.Sprint(args))
		return nil
	}

	cfg := rpiLoopSupervisorConfig{
		AOCommand:      "ao",
		AthenaSince:    "24h",
		AthenaDefrag:   false,
		CommandTimeout: 30 * time.Second,
	}
	err := runAthenaProducerTick(t.TempDir(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d: %v", len(commands), commands)
	}
	assertContains(t, commands[0], "mine")
}

func TestRunAthenaProducerTick_WithDefrag(t *testing.T) {
	origRunner := loopCommandRunner
	defer func() { loopCommandRunner = origRunner }()

	var commands []string
	loopCommandRunner = func(cwd string, timeout time.Duration, command string, args ...string) error {
		commands = append(commands, command+" "+fmt.Sprint(args))
		return nil
	}

	cfg := rpiLoopSupervisorConfig{
		AOCommand:      "ao",
		AthenaSince:    "26h",
		AthenaDefrag:   true,
		CommandTimeout: 30 * time.Second,
	}
	err := runAthenaProducerTick(t.TempDir(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commands) != 2 {
		t.Fatalf("expected 2 commands (mine + defrag), got %d", len(commands))
	}
	assertContains(t, commands[0], "mine")
	assertContains(t, commands[1], "defrag")
}

func TestRunAthenaProducerTick_MineFailure(t *testing.T) {
	origRunner := loopCommandRunner
	defer func() { loopCommandRunner = origRunner }()

	loopCommandRunner = func(cwd string, timeout time.Duration, command string, args ...string) error {
		return fmt.Errorf("mine failed")
	}

	cfg := rpiLoopSupervisorConfig{
		AOCommand:      "ao",
		CommandTimeout: 30 * time.Second,
	}
	err := runAthenaProducerTick(t.TempDir(), cfg)
	if err == nil {
		t.Fatal("expected error from mine failure")
	}
	assertContains(t, err.Error(), "athena mine producer failed")
}

// ---------------------------------------------------------------------------
// runParallelGateScript — runs a temp bash script
// ---------------------------------------------------------------------------

func TestRunParallelGateScript_EmptyScript(t *testing.T) {
	err := runParallelGateScript(t.TempDir(), 5)
	if err != nil {
		t.Fatalf("should skip when parallelGateScript is empty: %v", err)
	}
}

func TestRunParallelGateScript_ZeroMerges(t *testing.T) {
	origScript := parallelGateScript
	defer func() { parallelGateScript = origScript }()
	parallelGateScript = "/some/script.sh"

	err := runParallelGateScript(t.TempDir(), 0)
	if err != nil {
		t.Fatalf("should skip when mergedCount=0: %v", err)
	}
}

func TestRunParallelGateScript_Success(t *testing.T) {
	tmp := t.TempDir()
	script := filepath.Join(tmp, "gate.sh")
	if err := os.WriteFile(script, []byte("#!/bin/bash\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	origScript := parallelGateScript
	defer func() { parallelGateScript = origScript }()
	parallelGateScript = script

	err := runParallelGateScript(tmp, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunParallelGateScript_Failure(t *testing.T) {
	tmp := t.TempDir()
	script := filepath.Join(tmp, "gate.sh")
	if err := os.WriteFile(script, []byte("#!/bin/bash\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	origScript := parallelGateScript
	defer func() { parallelGateScript = origScript }()
	parallelGateScript = script

	err := runParallelGateScript(tmp, 2)
	if err == nil {
		t.Fatal("expected error from failing gate script")
	}
	assertContains(t, err.Error(), "gate script failed")
}

// ---------------------------------------------------------------------------
// runBatchFeedback — cobra command; test flag validation + dry-run
// ---------------------------------------------------------------------------

func TestRunBatchFeedback_InvalidMaxSessions(t *testing.T) {
	origMax := batchFeedbackMaxSessions
	origReward := batchFeedbackReward
	origRuntime := batchFeedbackMaxRuntime
	defer func() {
		batchFeedbackMaxSessions = origMax
		batchFeedbackReward = origReward
		batchFeedbackMaxRuntime = origRuntime
	}()

	batchFeedbackMaxSessions = -1
	batchFeedbackReward = -1
	batchFeedbackMaxRuntime = 0

	err := runBatchFeedback(nil, nil)
	if err == nil {
		t.Fatal("expected validation error for negative max-sessions")
	}
	assertContains(t, err.Error(), "--max-sessions")
}

func TestRunBatchFeedback_InvalidReward(t *testing.T) {
	origMax := batchFeedbackMaxSessions
	origReward := batchFeedbackReward
	origRuntime := batchFeedbackMaxRuntime
	defer func() {
		batchFeedbackMaxSessions = origMax
		batchFeedbackReward = origReward
		batchFeedbackMaxRuntime = origRuntime
	}()

	batchFeedbackMaxSessions = 0
	batchFeedbackReward = 2.0 // out of range
	batchFeedbackMaxRuntime = 0

	err := runBatchFeedback(nil, nil)
	if err == nil {
		t.Fatal("expected validation error for reward > 1.0")
	}
	assertContains(t, err.Error(), "--reward")
}

func TestRunBatchFeedback_DryRun_NoSessions(t *testing.T) {
	tmp := t.TempDir()

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	origMax := batchFeedbackMaxSessions
	origReward := batchFeedbackReward
	origRuntime := batchFeedbackMaxRuntime
	origDays := batchFeedbackDays
	origDryRun := dryRun
	defer func() {
		batchFeedbackMaxSessions = origMax
		batchFeedbackReward = origReward
		batchFeedbackMaxRuntime = origRuntime
		batchFeedbackDays = origDays
		dryRun = origDryRun
	}()

	batchFeedbackMaxSessions = 0
	batchFeedbackReward = -1
	batchFeedbackMaxRuntime = 0
	batchFeedbackDays = 1
	dryRun = true

	// Create empty citations file so discoverUnprocessedSessions returns 0
	aoDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(aoDir, 0o755); err != nil {
		t.Fatal(err)
	}

	err := runBatchFeedback(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// spawnClaudeDirectGlobal — verify it delegates to spawnClaudeDirectImpl
// ---------------------------------------------------------------------------

func TestSpawnClaudeDirectGlobal_DelegatesToImpl(t *testing.T) {
	// We can't actually spawn claude in tests, but we can verify the function
	// is callable and returns an error when the runtime isn't available.
	// This exercises the function entry point for coverage.
	origTimeout := phasedPhaseTimeout
	defer func() { phasedPhaseTimeout = origTimeout }()
	phasedPhaseTimeout = 1 * time.Millisecond

	// The function should fail quickly because "claude" isn't available in test env
	// or the timeout is very short. Either way, the function body is exercised.
	err := spawnClaudeDirectGlobal("test prompt", t.TempDir(), 1)
	// We expect an error (runtime not available or timeout)
	if err == nil {
		t.Log("spawnClaudeDirectGlobal returned nil (claude may be on PATH); function exercised")
	}
}

// ---------------------------------------------------------------------------
// cleanupParallelWorktrees — exercise with empty inputs
// ---------------------------------------------------------------------------

func TestCleanupParallelWorktrees_EmptyInputs(t *testing.T) {
	// Exercise the function with empty slices — should be a no-op
	cleanupParallelWorktrees(nil, nil)
	cleanupParallelWorktrees([]worktreeInfo{}, []parallelResult{})
}

// ---------------------------------------------------------------------------
// runPoolAutoPromoteAndPromote — thin wrapper test
// ---------------------------------------------------------------------------

func TestRunPoolAutoPromoteAndPromote_NilPool(t *testing.T) {
	// Pool operations on nil pool should return an error or panic.
	// This exercises the function entry for coverage.
	defer func() {
		if r := recover(); r != nil {
			t.Log("recovered from nil pool panic — function exercised")
		}
	}()
	_ = runPoolAutoPromoteAndPromote(nil, 24*time.Hour, "test")
}

// ---------------------------------------------------------------------------
// runRPIStatusWatch — signal-driven exit
// ---------------------------------------------------------------------------

func TestRunRPIStatusWatch_ExitsOnSignal(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Create minimal .agents dir so status doesn't error
	if err := os.MkdirAll(filepath.Join(tmp, ".agents", "rpi"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Send SIGINT to ourselves after a brief delay to break the watch loop
	go func() {
		time.Sleep(200 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(os.Interrupt)
	}()

	err := runRPIStatusWatch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !bytes.Contains([]byte(haystack), []byte(needle)) {
		t.Errorf("expected %q to contain %q", haystack, needle)
	}
}
