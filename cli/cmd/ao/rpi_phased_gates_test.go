package main

import (
	"context"
	"fmt"
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
