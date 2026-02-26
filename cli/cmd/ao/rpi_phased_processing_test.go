package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/boshu2/agentops/cli/internal/types"
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
		name     string
		payload  map[string]any
		want     bool
		wantErr  bool
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

// --- classifyByPhase ---

func TestClassifyByPhase_UnknownPhase(t *testing.T) {
	got := classifyByPhase(99, "FAIL")
	if got != "" {
		t.Errorf("classifyByPhase(99, FAIL) = %q, want empty", got)
	}
}

func TestClassifyByPhase_Phase2NonBlockedOrPartial(t *testing.T) {
	got := classifyByPhase(2, "FAIL")
	if got != "" {
		t.Errorf("classifyByPhase(2, FAIL) = %q, want empty (only BLOCKED/PARTIAL handled for phase 2)", got)
	}
}

// --- classifyByVerdict ---

func TestClassifyByVerdict_Timeout(t *testing.T) {
	got := classifyByVerdict(string(failReasonTimeout))
	if got != types.MemRLFailureClassPhaseTimeout {
		t.Errorf("got %q, want %q", got, types.MemRLFailureClassPhaseTimeout)
	}
}

func TestClassifyByVerdict_Stall(t *testing.T) {
	got := classifyByVerdict(string(failReasonStall))
	if got != types.MemRLFailureClassPhaseStall {
		t.Errorf("got %q, want %q", got, types.MemRLFailureClassPhaseStall)
	}
}

func TestClassifyByVerdict_ExitError(t *testing.T) {
	got := classifyByVerdict(string(failReasonExit))
	if got != types.MemRLFailureClassPhaseExitError {
		t.Errorf("got %q, want %q", got, types.MemRLFailureClassPhaseExitError)
	}
}

func TestClassifyByVerdict_UnknownVerdict(t *testing.T) {
	got := classifyByVerdict("SOMETHING_ELSE")
	if got != types.MemRLFailureClass("something_else") {
		t.Errorf("got %q, want lowercase version", got)
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

	err := executeWithStatus(executor, state, statusPath, allPhases, 1, 0, "test prompt", t.TempDir(), "running", "failed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteWithStatus_FailurePath(t *testing.T) {
	executor := &fakeExecutor{err: errFakeExecFailure}
	state := newTestPhasedState()
	state.Opts.LiveStatus = false
	allPhases := buildAllPhases(phases)

	err := executeWithStatus(executor, state, "", allPhases, 1, 0, "test prompt", t.TempDir(), "running", "failed")
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

// --- deriveRepoRootFromRPIOrchestrationLog ---

func TestDeriveRepoRoot_ValidLogPath(t *testing.T) {
	root, ok := deriveRepoRootFromRPIOrchestrationLog("/home/user/project/.agents/rpi/phased-orchestration.log")
	if !ok {
		t.Fatal("expected ok=true for valid path")
	}
	if root != "/home/user/project" {
		t.Errorf("root = %q, want %q", root, "/home/user/project")
	}
}

func TestDeriveRepoRoot_InvalidLogPaths(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "not in rpi dir", path: "/home/user/.agents/logs/log.txt"},
		{name: "not in .agents", path: "/home/user/rpi/log.txt"},
		{name: "empty", path: ""},
		{name: "just rpi", path: "/rpi/log.txt"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := deriveRepoRootFromRPIOrchestrationLog(tc.path)
			if ok {
				t.Errorf("expected ok=false for path %q", tc.path)
			}
		})
	}
}

// --- ledgerActionFromDetails ---

func TestLedgerActionFromDetails_Prefixes(t *testing.T) {
	tests := []struct {
		details string
		want    string
	}{
		{"started", "started"},
		{"completed in 5s", "completed"},
		{"failed: phase error", "failed"},
		{"FATAL: crash", "fatal"},
		{"retry attempt 2/3", "retry"},
		{"dry-run complete", "dry-run"},
		{"HANDOFF detected", "handoff"},
		{"epic=ag-1 verdicts=map[]", "summary"},
		{"", "event"},
		{"some random text", "some"},
	}
	for _, tc := range tests {
		t.Run(tc.details, func(t *testing.T) {
			got := ledgerActionFromDetails(tc.details)
			if got != tc.want {
				t.Errorf("ledgerActionFromDetails(%q) = %q, want %q", tc.details, got, tc.want)
			}
		})
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
func (f *fakeExecutor) Execute(prompt, cwd, runID string, phaseNum int) error {
	f.executed = true
	return f.err
}
