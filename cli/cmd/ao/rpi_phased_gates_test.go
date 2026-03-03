package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/boshu2/agentops/cli/internal/types"
)

// --- legacyGateAction ---

func TestLegacyGateAction_RetryWhenBelowMax(t *testing.T) {
	got := legacyGateAction(1, 3)
	if got != types.MemRLActionRetry {
		t.Errorf("legacyGateAction(1,3) = %q, want %q", got, types.MemRLActionRetry)
	}
}

func TestLegacyGateAction_EscalateWhenAtMax(t *testing.T) {
	got := legacyGateAction(3, 3)
	if got != types.MemRLActionEscalate {
		t.Errorf("legacyGateAction(3,3) = %q, want %q", got, types.MemRLActionEscalate)
	}
}

func TestLegacyGateAction_EscalateWhenAboveMax(t *testing.T) {
	got := legacyGateAction(5, 3)
	if got != types.MemRLActionEscalate {
		t.Errorf("legacyGateAction(5,3) = %q, want %q", got, types.MemRLActionEscalate)
	}
}

// --- classifyGateFailureClass ---

func TestClassifyGateFailureClass_NilError(t *testing.T) {
	got := classifyGateFailureClass(1, nil)
	if got != "" {
		t.Errorf("classifyGateFailureClass(1, nil) = %q, want empty", got)
	}
}

func TestClassifyGateFailureClass_PreMortemFail(t *testing.T) {
	gateErr := &gateFailError{Phase: 1, Verdict: "FAIL"}
	got := classifyGateFailureClass(1, gateErr)
	if got != types.MemRLFailureClassPreMortemFail {
		t.Errorf("got %q, want %q", got, types.MemRLFailureClassPreMortemFail)
	}
}

func TestClassifyGateFailureClass_CrankBlocked(t *testing.T) {
	gateErr := &gateFailError{Phase: 2, Verdict: "BLOCKED"}
	got := classifyGateFailureClass(2, gateErr)
	if got != types.MemRLFailureClassCrankBlocked {
		t.Errorf("got %q, want %q", got, types.MemRLFailureClassCrankBlocked)
	}
}

func TestClassifyGateFailureClass_CrankPartial(t *testing.T) {
	gateErr := &gateFailError{Phase: 2, Verdict: "PARTIAL"}
	got := classifyGateFailureClass(2, gateErr)
	if got != types.MemRLFailureClassCrankPartial {
		t.Errorf("got %q, want %q", got, types.MemRLFailureClassCrankPartial)
	}
}

func TestClassifyGateFailureClass_VibeFail(t *testing.T) {
	gateErr := &gateFailError{Phase: 3, Verdict: "FAIL"}
	got := classifyGateFailureClass(3, gateErr)
	if got != types.MemRLFailureClassVibeFail {
		t.Errorf("got %q, want %q", got, types.MemRLFailureClassVibeFail)
	}
}

func TestClassifyGateFailureClass_WhitespaceHandling(t *testing.T) {
	gateErr := &gateFailError{Phase: 1, Verdict: "  FAIL  "}
	got := classifyGateFailureClass(1, gateErr)
	if got != types.MemRLFailureClassPreMortemFail {
		t.Errorf("got %q (whitespace should be trimmed), want %q", got, types.MemRLFailureClassPreMortemFail)
	}
}

// --- resolveGateRetryAction ---

func TestResolveGateRetryAction_ModeOffUsesLegacy(t *testing.T) {
	// When memrl mode is off, the legacy action should be returned.
	t.Setenv("MEMRL_MODE", "off")
	state := newTestPhasedState().WithMaxRetries(3)
	gateErr := &gateFailError{Phase: 1, Verdict: "FAIL"}

	action, _ := resolveGateRetryAction(state, 1, gateErr, 1)
	// Legacy: attempt 1 < maxRetries 3 => retry
	if action != types.MemRLActionRetry {
		t.Errorf("action = %q, want %q", action, types.MemRLActionRetry)
	}
}

func TestResolveGateRetryAction_ModeOffEscalatesAtMax(t *testing.T) {
	t.Setenv("MEMRL_MODE", "off")
	state := newTestPhasedState().WithMaxRetries(2)
	gateErr := &gateFailError{Phase: 1, Verdict: "FAIL"}

	action, _ := resolveGateRetryAction(state, 1, gateErr, 2)
	// Legacy: attempt 2 >= maxRetries 2 => escalate
	if action != types.MemRLActionEscalate {
		t.Errorf("action = %q, want %q", action, types.MemRLActionEscalate)
	}
}

// --- logGateRetryMemRL ---

func TestLogGateRetryMemRL_OffModeSkipsLog(t *testing.T) {
	// When mode is off, no log entry should be written.
	// We just verify it does not panic with empty args.
	decision := types.MemRLPolicyDecision{Mode: types.MemRLModeOff}
	logGateRetryMemRL("", "", "test", decision, types.MemRLActionRetry)
	// No assertion needed — just verifying no panic.
}

// --- executeWithStatus ---

func TestExecuteWithStatus_SuccessPath(t *testing.T) {
	executor := &fakeExecutor{err: nil}
	state := newTestPhasedState()
	state.Opts.LiveStatus = false
	allPhases := buildAllPhases(phases)
	statusPath := ""

	err := executeWithStatus(context.Background(), executor, state, statusPath, allPhases, 1, 0, "test prompt", t.TempDir(), "running", "failed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteWithStatus_FailurePath(t *testing.T) {
	executor := &fakeExecutor{err: errFakeExecFailure}
	state := newTestPhasedState()
	state.Opts.LiveStatus = false
	allPhases := buildAllPhases(phases)

	err := executeWithStatus(context.Background(), executor, state, "", allPhases, 1, 0, "test prompt", t.TempDir(), "running", "failed")
	if err == nil {
		t.Fatal("expected error from failing executor")
	}
}

// --- maybeUpdateLiveStatus ---

func TestMaybeUpdateLiveStatus_DisabledDoesNotPanic(t *testing.T) {
	state := newTestPhasedState()
	state.Opts.LiveStatus = false
	allPhases := buildAllPhases(phases)
	// Should not panic even with empty statusPath when disabled.
	maybeUpdateLiveStatus(state, "", allPhases, 1, "test", 0, "")
}

// --- postPhaseProcessing ---

func TestPostPhaseProcessing_UnknownPhase(t *testing.T) {
	state := newTestPhasedState()
	err := postPhaseProcessing(t.TempDir(), state, 99, "")
	if err != nil {
		t.Errorf("unknown phase should return nil, got: %v", err)
	}
}

// --- gateFailError ---

func TestGateFailError_ErrorString(t *testing.T) {
	err := &gateFailError{Phase: 2, Verdict: "BLOCKED", Report: "/path/to/report.md"}
	got := err.Error()
	want := "gate FAIL at phase 2: BLOCKED (report: /path/to/report.md)"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestGateFailError_WithFindingsSlice(t *testing.T) {
	err := &gateFailError{
		Phase:   3,
		Verdict: "FAIL",
		Findings: []finding{
			{Description: "test error", Fix: "fix it", Ref: "file.go:10"},
		},
		Report: "report.md",
	}
	if len(err.Findings) != 1 {
		t.Errorf("expected 1 finding, got %d", len(err.Findings))
	}
}

// --- shouldForceEscalation ---

func TestShouldForceEscalation_BelowCeiling(t *testing.T) {
	for i := 1; i <= maxGateRetryDepth; i++ {
		if shouldForceEscalation(i) {
			t.Errorf("shouldForceEscalation(%d) = true, want false (maxGateRetryDepth=%d)", i, maxGateRetryDepth)
		}
	}
}

func TestShouldForceEscalation_AboveCeiling(t *testing.T) {
	for _, attempt := range []int{maxGateRetryDepth + 1, maxGateRetryDepth + 10, 100} {
		if !shouldForceEscalation(attempt) {
			t.Errorf("shouldForceEscalation(%d) = false, want true", attempt)
		}
	}
}

// --- C2 event emission tests ---

func TestGateDiscoveryVerdictC2Event(t *testing.T) {
	root := t.TempDir()
	runID := "run-disc-c2"

	// Create council report with PASS verdict
	councilDir := filepath.Join(root, ".agents", "council")
	if err := os.MkdirAll(councilDir, 0o755); err != nil {
		t.Fatal(err)
	}
	reportName := "2026-03-01-pre-mortem-test.md"
	reportPath := filepath.Join(councilDir, reportName)
	if err := os.WriteFile(reportPath, []byte("# Pre-mortem\n## Council Verdict: PASS\nAll good.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set up state with a plan-file epic to bypass bd
	state := newTestPhasedState().WithRunID(runID).WithEpicID(planFileEpicPrefix + "plan.md")
	state.Verdicts = make(map[string]string)

	logDir := filepath.Join(root, ".agents", "rpi")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(logDir, "phased-orchestration.log")

	err := processDiscoveryPhase(root, state, logPath)
	if err != nil {
		t.Fatalf("processDiscoveryPhase: %v", err)
	}

	events, err := loadRPIC2Events(root, runID)
	if err != nil {
		t.Fatalf("loadRPIC2Events: %v", err)
	}
	found := false
	for _, ev := range events {
		if ev.Type == "gate.discovery.verdict" {
			found = true
			if ev.Phase != 1 {
				t.Errorf("event phase = %d, want 1", ev.Phase)
			}
			if ev.RunID != runID {
				t.Errorf("event run_id = %q, want %q", ev.RunID, runID)
			}
		}
	}
	if !found {
		t.Errorf("gate.discovery.verdict event not found among %d events", len(events))
	}
}

func TestGateValidationVerdictC2Event(t *testing.T) {
	root := t.TempDir()
	runID := "run-val-c2"

	// Create vibe council report with PASS verdict
	councilDir := filepath.Join(root, ".agents", "council")
	if err := os.MkdirAll(councilDir, 0o755); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(councilDir, "2026-03-01-vibe-test.md")
	if err := os.WriteFile(reportPath, []byte("# Vibe\n## Council Verdict: PASS\nAll good.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create prior phase result to pass validation
	rpiDir := filepath.Join(root, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rpiDir, "phase-2-result.json"), []byte(`{"status":"ok"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	state := newTestPhasedState().WithRunID(runID).WithPhase(3).WithStartPhase(3)
	state.Verdicts = make(map[string]string)

	logPath := filepath.Join(rpiDir, "phased-orchestration.log")

	err := processValidationPhase(root, state, 3, logPath)
	if err != nil {
		t.Fatalf("processValidationPhase: %v", err)
	}

	events, err := loadRPIC2Events(root, runID)
	if err != nil {
		t.Fatalf("loadRPIC2Events: %v", err)
	}
	found := false
	for _, ev := range events {
		if ev.Type == "gate.validation.verdict" {
			found = true
			if ev.Phase != 3 {
				t.Errorf("event phase = %d, want 3", ev.Phase)
			}
			if ev.RunID != runID {
				t.Errorf("event run_id = %q, want %q", ev.RunID, runID)
			}
		}
	}
	if !found {
		t.Errorf("gate.validation.verdict event not found among %d events", len(events))
	}
}

func TestGateEscalationC2Event(t *testing.T) {
	root := t.TempDir()
	runID := "run-esc-c2"

	// Set up logPath at root/.agents/rpi/phased-orchestration.log
	rpiDir := filepath.Join(root, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(rpiDir, "phased-orchestration.log")
	statusPath := filepath.Join(rpiDir, "live-status.md")

	state := newTestPhasedState().WithRunID(runID).WithMaxRetries(3)
	gateErr := &gateFailError{Phase: 1, Verdict: "FAIL", Report: "report.md"}
	allPhases := buildAllPhases(phases)

	performGateEscalation(state, 1, 3, gateErr,
		types.MemRLPolicyDecision{}, types.MemRLActionEscalate,
		"discovery", logPath, statusPath, allPhases)

	// escalationRoot = filepath.Dir(filepath.Dir(logPath)) = root/.agents
	escalationRoot := filepath.Dir(filepath.Dir(logPath))
	events, err := loadRPIC2Events(escalationRoot, runID)
	if err != nil {
		t.Fatalf("loadRPIC2Events: %v", err)
	}
	found := false
	for _, ev := range events {
		if ev.Type == "gate.escalation" {
			found = true
			if ev.Phase != 1 {
				t.Errorf("event phase = %d, want 1", ev.Phase)
			}
			if ev.RunID != runID {
				t.Errorf("event run_id = %q, want %q", ev.RunID, runID)
			}
		}
	}
	if !found {
		t.Errorf("gate.escalation event not found among %d events", len(events))
	}
}

func TestGateRetryAttemptC2Event(t *testing.T) {
	root := t.TempDir()
	runID := "run-retry-c2"

	rpiDir := filepath.Join(root, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(rpiDir, "phased-orchestration.log")
	statusPath := filepath.Join(rpiDir, "live-status.md")

	state := newTestPhasedState().WithRunID(runID).WithMaxRetries(3)
	state.Attempts = map[string]int{}
	gateErr := &gateFailError{Phase: 1, Verdict: "FAIL", Report: "report.md"}
	allPhases := buildAllPhases(phases)

	// Enable dry-run to avoid actually spawning
	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	t.Setenv("MEMRL_MODE", "off")

	_, _ = handleGateRetry(context.Background(), root, state, 1, gateErr, logPath, root, statusPath, allPhases, &fakeExecutor{})

	events, err := loadRPIC2Events(root, runID)
	if err != nil {
		t.Fatalf("loadRPIC2Events: %v", err)
	}
	found := false
	for _, ev := range events {
		if ev.Type == "gate.retry.attempt" {
			found = true
			if ev.Phase != 1 {
				t.Errorf("event phase = %d, want 1", ev.Phase)
			}
			if ev.RunID != runID {
				t.Errorf("event run_id = %q, want %q", ev.RunID, runID)
			}
		}
	}
	if !found {
		t.Errorf("gate.retry.attempt event not found among %d events", len(events))
	}
}

func TestGateImplementationVerdictC2Event(t *testing.T) {
	root := t.TempDir()
	runID := "run-impl-c2"

	// Directly test the event shape by calling appendRPIC2Event with the same input
	// pattern used in processImplementationPhase. Testing processImplementationPhase
	// end-to-end requires bd CLI which is not available in unit tests.
	status := "COMPLETE"
	epicID := "ag-test"
	_, err := appendRPIC2Event(root, rpiC2EventInput{
		RunID:   runID,
		Phase:   2,
		Type:    "gate.implementation.verdict",
		Message: fmt.Sprintf("Crank status: %s", status),
		Details: map[string]any{"status": status, "epic_id": epicID},
	})
	if err != nil {
		t.Fatalf("appendRPIC2Event: %v", err)
	}

	events, err := loadRPIC2Events(root, runID)
	if err != nil {
		t.Fatalf("loadRPIC2Events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].Type != "gate.implementation.verdict" {
		t.Errorf("event type = %q, want gate.implementation.verdict", events[0].Type)
	}
	if events[0].Phase != 2 {
		t.Errorf("event phase = %d, want 2", events[0].Phase)
	}
}

// --- fakeExecutor for tests ---

type fakeExecutor struct {
	err      error
	executed bool
}

var errFakeExecFailure = fmt.Errorf("fake execution failure")

func (f *fakeExecutor) Name() string { return "fake" }
func (f *fakeExecutor) Execute(_ context.Context, prompt, cwd, runID string, phaseNum int) error {
	f.executed = true
	return f.err
}
