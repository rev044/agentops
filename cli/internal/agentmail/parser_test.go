package agentmail

import (
	"testing"
	"time"
)

func TestParser_ParseBeadAccepted(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:         "msg-001",
		SenderName: "demigod-ol-527-1",
		To:         "crank-ol527",
		Subject:    "BEAD_ACCEPTED",
		BodyMD: `Accepted bead: ol-527.1
Title: Add authentication middleware
Starting implementation at: 2026-01-31T10:00:00Z`,
		ThreadID:    "ol-527.1",
		Timestamp:   "2026-01-31T10:00:00Z",
		AckRequired: false,
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeBeadAccepted {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeBeadAccepted)
	}

	if msg.Parsed.BeadID != "ol-527.1" {
		t.Errorf("BeadID = %q, want %q", msg.Parsed.BeadID, "ol-527.1")
	}

	if msg.Parsed.Title != "Add authentication middleware" {
		t.Errorf("Title = %q, want %q", msg.Parsed.Title, "Add authentication middleware")
	}

	if msg.From != "demigod-ol-527-1" {
		t.Errorf("From = %q, want %q", msg.From, "demigod-ol-527-1")
	}
}

func TestParser_ParseBeadAcceptedWithBracketSubject(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:         "msg-002",
		SenderName: "demigod-ol-527-1",
		To:         "crank-ol527",
		Subject:    "[ol-527.1] BEAD_ACCEPTED",
		BodyMD:     "Starting work on authentication middleware",
		ThreadID:   "ol-527.1",
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeBeadAccepted {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeBeadAccepted)
	}

	// Should extract bead ID from subject
	if msg.Parsed.BeadID != "ol-527.1" {
		t.Errorf("BeadID = %q, want %q", msg.Parsed.BeadID, "ol-527.1")
	}
}

func TestParser_ParseProgress(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:         "msg-003",
		SenderName: "demigod-ol-527-1",
		To:         "crank-ol527",
		Subject:    "PROGRESS",
		BodyMD: `Bead: ol-527.1
Step: Step 4 - implementing auth middleware
Status: Writing middleware handler
Context usage: 45%
Files touched: src/auth.py, tests/test_auth.py`,
		ThreadID: "ol-527.1",
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeProgress {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeProgress)
	}

	if msg.Parsed.BeadID != "ol-527.1" {
		t.Errorf("BeadID = %q, want %q", msg.Parsed.BeadID, "ol-527.1")
	}

	if msg.Parsed.Step != "Step 4 - implementing auth middleware" {
		t.Errorf("Step = %q, want %q", msg.Parsed.Step, "Step 4 - implementing auth middleware")
	}

	if msg.Parsed.ContextUsage != 45 {
		t.Errorf("ContextUsage = %d, want %d", msg.Parsed.ContextUsage, 45)
	}

	if len(msg.Parsed.FilesTouched) != 2 {
		t.Errorf("FilesTouched count = %d, want %d", len(msg.Parsed.FilesTouched), 2)
	}
}

func TestParser_ParseHelpRequest(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:         "msg-004",
		SenderName: "demigod-ol-527-2",
		To:         "chiron@olympus",
		Subject:    "HELP_REQUEST",
		BodyMD: `Bead: ol-527.2
Issue Type: STUCK

## Problem
Cannot find the existing auth module to integrate with.

## What I Tried
- Searched for auth.py, authentication.py
- Checked imports in main.py
- Looked in lib/ directory

## Files Touched
- src/main.py
- lib/utils.py

## Question
Where is the existing authentication module located?`,
		ThreadID:    "ol-527.2",
		AckRequired: true,
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeHelpRequest {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeHelpRequest)
	}

	if msg.Parsed.BeadID != "ol-527.2" {
		t.Errorf("BeadID = %q, want %q", msg.Parsed.BeadID, "ol-527.2")
	}

	if msg.Parsed.IssueType != HelpRequestIssueTypeStuck {
		t.Errorf("IssueType = %q, want %q", msg.Parsed.IssueType, HelpRequestIssueTypeStuck)
	}

	if msg.Parsed.Problem == "" {
		t.Error("Problem should not be empty")
	}

	if msg.Parsed.Question == "" {
		t.Error("Question should not be empty")
	}

	if len(msg.Parsed.FilesTouched) != 2 {
		t.Errorf("FilesTouched count = %d, want %d", len(msg.Parsed.FilesTouched), 2)
	}

	if !msg.AckRequired {
		t.Error("AckRequired should be true")
	}
}

func TestParser_ParseHelpRequestVariants(t *testing.T) {
	tests := []struct {
		name      string
		issueType string
		want      HelpRequestIssueType
	}{
		{"STUCK", "STUCK", HelpRequestIssueTypeStuck},
		{"SPEC_UNCLEAR", "SPEC_UNCLEAR", HelpRequestIssueTypeSpecUnclear},
		{"BLOCKED", "BLOCKED", HelpRequestIssueTypeBlocked},
		{"TECHNICAL", "TECHNICAL", HelpRequestIssueTypeTechnical},
	}

	p := NewParser()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw := &RawMessage{
				Subject: "HELP_REQUEST",
				BodyMD:  "Bead: test-123\nIssue Type: " + tc.issueType + "\n\n## Problem\nTest",
			}

			msg, err := p.Parse(raw)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if msg.Parsed.IssueType != tc.want {
				t.Errorf("IssueType = %q, want %q", msg.Parsed.IssueType, tc.want)
			}
		})
	}
}

func TestParser_ParseOfferingReady(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:         "msg-005",
		SenderName: "demigod-ol-527-1",
		To:         "crank-ol527",
		Subject:    "OFFERING_READY",
		BodyMD: `Bead: ol-527.1
Status: DONE

## Changes
- Commit: abc123def
- Files: src/auth.py, tests/test_auth.py, src/middleware.py

## Self-Validation
- Tests: PASS
- Lint: PASS
- Build: PASS

## Summary
Added JWT authentication middleware with token validation and refresh logic.`,
		ThreadID:    "ol-527.1",
		AckRequired: true,
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeOfferingReady {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeOfferingReady)
	}

	if msg.Parsed.BeadID != "ol-527.1" {
		t.Errorf("BeadID = %q, want %q", msg.Parsed.BeadID, "ol-527.1")
	}

	if msg.Parsed.CommitSHA != "abc123def" {
		t.Errorf("CommitSHA = %q, want %q", msg.Parsed.CommitSHA, "abc123def")
	}

	if !msg.Parsed.TestsPass {
		t.Error("TestsPass should be true")
	}

	if !msg.Parsed.LintPass {
		t.Error("LintPass should be true")
	}

	if !msg.Parsed.BuildPass {
		t.Error("BuildPass should be true")
	}

	if len(msg.Parsed.Files) != 3 {
		t.Errorf("Files count = %d, want %d", len(msg.Parsed.Files), 3)
	}

	if msg.Parsed.Summary == "" {
		t.Error("Summary should not be empty")
	}

	if !msg.Type.IsCompletionType() {
		t.Error("OFFERING_READY should be a completion type")
	}

	if !msg.Type.IsSuccessType() {
		t.Error("OFFERING_READY should be a success type")
	}
}

func TestParser_ParseDone(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:         "msg-006",
		SenderName: "demigod-ol-527-3",
		To:         "crank-ol527",
		Subject:    "[ol-527.3] DONE",
		BodyMD: `Bead: ol-527.3
Status: DONE

## Changes
- Commit: def456abc
- Files: docs/api.md

## Self-Validation
- Tests: PASS
- Lint: PASS
- Build: PASS

## Summary
Updated API documentation with new endpoints.`,
		ThreadID: "ol-527.3",
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeDone {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeDone)
	}

	if !msg.Type.IsCompletionType() {
		t.Error("DONE should be a completion type")
	}

	if !msg.Type.IsSuccessType() {
		t.Error("DONE should be a success type")
	}
}

func TestParser_ParseFailed(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:         "msg-007",
		SenderName: "demigod-ol-527-4",
		To:         "crank-ol527",
		Subject:    "FAILED",
		BodyMD: `Bead: ol-527.4
Status: FAILED

## Failure
Type: TESTS_FAIL
Reason: Integration tests fail due to missing database connection
Internal Attempts: 3

## Partial Progress
- Commit: ghi789
- Files: src/db.py, tests/test_db.py

## Recommendation
Check database configuration and ensure test fixtures are properly set up.`,
		ThreadID: "ol-527.4",
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeFailed {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeFailed)
	}

	if msg.Parsed.BeadID != "ol-527.4" {
		t.Errorf("BeadID = %q, want %q", msg.Parsed.BeadID, "ol-527.4")
	}

	if msg.Parsed.FailureType != FailureTypeTestsFail {
		t.Errorf("FailureType = %q, want %q", msg.Parsed.FailureType, FailureTypeTestsFail)
	}

	if msg.Parsed.Reason == "" {
		t.Error("Reason should not be empty")
	}

	if msg.Parsed.InternalAttempts != 3 {
		t.Errorf("InternalAttempts = %d, want %d", msg.Parsed.InternalAttempts, 3)
	}

	if msg.Parsed.Recommendation == "" {
		t.Error("Recommendation should not be empty")
	}

	if msg.Type.IsSuccessType() {
		t.Error("FAILED should not be a success type")
	}

	if !msg.Type.IsCompletionType() {
		t.Error("FAILED should be a completion type")
	}
}

func TestParser_ParseFailedVariants(t *testing.T) {
	tests := []struct {
		name        string
		failureType string
		want        FailureType
	}{
		{"TESTS_FAIL", "TESTS_FAIL", FailureTypeTestsFail},
		{"BUILD_FAIL", "BUILD_FAIL", FailureTypeBuildFail},
		{"SPEC_IMPOSSIBLE", "SPEC_IMPOSSIBLE", FailureTypeSpecImpossible},
		{"CONTEXT_HIGH", "CONTEXT_HIGH", FailureTypeContextHigh},
		{"ERROR", "ERROR", FailureTypeError},
	}

	p := NewParser()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw := &RawMessage{
				Subject: "FAILED",
				BodyMD:  "Bead: test-123\n\n## Failure\nType: " + tc.failureType + "\nReason: Test failure",
			}

			msg, err := p.Parse(raw)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if msg.Parsed.FailureType != tc.want {
				t.Errorf("FailureType = %q, want %q", msg.Parsed.FailureType, tc.want)
			}
		})
	}
}

func TestParser_ParseCheckpoint(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:         "msg-008",
		SenderName: "demigod-ol-527-5",
		To:         "crank-ol527",
		Subject:    "CHECKPOINT",
		BodyMD: `Bead: ol-527.5
Reason: CONTEXT_HIGH

## Progress
- Commit: jkl012
- Description: Steps 1-3 complete, Step 4 in progress
- Context usage: 85%

## Next Steps for Successor
1. Complete Step 4 - finish implementing the rate limiter
2. Run remaining tests
3. Update documentation`,
		ThreadID: "ol-527.5",
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeCheckpoint {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeCheckpoint)
	}

	if msg.Parsed.BeadID != "ol-527.5" {
		t.Errorf("BeadID = %q, want %q", msg.Parsed.BeadID, "ol-527.5")
	}

	if msg.Parsed.CheckpointReason != CheckpointReasonContextHigh {
		t.Errorf("CheckpointReason = %q, want %q", msg.Parsed.CheckpointReason, CheckpointReasonContextHigh)
	}

	if msg.Parsed.PartialCommitSHA != "jkl012" {
		t.Errorf("PartialCommitSHA = %q, want %q", msg.Parsed.PartialCommitSHA, "jkl012")
	}

	if msg.Parsed.ContextUsage != 85 {
		t.Errorf("ContextUsage = %d, want %d", msg.Parsed.ContextUsage, 85)
	}

	if msg.Parsed.NextSteps == "" {
		t.Error("NextSteps should not be empty")
	}

	if !msg.Type.IsCompletionType() {
		t.Error("CHECKPOINT should be a completion type")
	}

	if msg.Type.IsSuccessType() {
		t.Error("CHECKPOINT should not be a success type")
	}
}

func TestParser_ParseSpawnRequest(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:         "msg-009",
		SenderName: "crank-ol527",
		To:         "spawner",
		Subject:    "SPAWN_REQUEST",
		BodyMD: `Issue: ol-527.6
Resume: true
Checkpoint: jkl012
Orchestrator: crank-ol527`,
		AckRequired: true,
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeSpawnRequest {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeSpawnRequest)
	}

	if msg.Parsed.IssueID != "ol-527.6" {
		t.Errorf("IssueID = %q, want %q", msg.Parsed.IssueID, "ol-527.6")
	}

	if !msg.Parsed.Resume {
		t.Error("Resume should be true")
	}

	if msg.Parsed.PartialCommitSHA != "jkl012" {
		t.Errorf("PartialCommitSHA = %q, want %q", msg.Parsed.PartialCommitSHA, "jkl012")
	}

	if msg.Parsed.Orchestrator != "crank-ol527" {
		t.Errorf("Orchestrator = %q, want %q", msg.Parsed.Orchestrator, "crank-ol527")
	}
}

func TestParser_ParseSpawnAck(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:         "msg-010",
		SenderName: "spawner",
		To:         "crank-ol527",
		Subject:    "SPAWN_ACK",
		BodyMD: `Issue: ol-527.6
Status: spawned
Session: demigod-ol-527-6`,
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeSpawnAck {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeSpawnAck)
	}

	if msg.Parsed.IssueID != "ol-527.6" {
		t.Errorf("IssueID = %q, want %q", msg.Parsed.IssueID, "ol-527.6")
	}

	if msg.Parsed.Status != "spawned" {
		t.Errorf("Status = %q, want %q", msg.Parsed.Status, "spawned")
	}
}

func TestParser_ParseUnknownType(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:      "msg-011",
		Subject: "RANDOM_MESSAGE",
		BodyMD:  "Some content",
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeUnknown {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeUnknown)
	}
}

func TestParser_ParseNilMessage(t *testing.T) {
	p := NewParser()

	_, err := p.Parse(nil)
	if err == nil {
		t.Error("Expected error for nil message")
	}
}

func TestParser_ParseTimestamp(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		Subject:   "PROGRESS",
		BodyMD:    "Bead: test-123",
		Timestamp: "2026-01-31T10:30:00Z",
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Timestamp.IsZero() {
		t.Error("Timestamp should be parsed")
	}

	expected := time.Date(2026, 1, 31, 10, 30, 0, 0, time.UTC)
	if !msg.Timestamp.Equal(expected) {
		t.Errorf("Timestamp = %v, want %v", msg.Timestamp, expected)
	}
}

func TestParser_ParseBatch(t *testing.T) {
	p := NewParser()

	raws := []*RawMessage{
		{Subject: "BEAD_ACCEPTED", BodyMD: "Accepted bead: test-1"},
		{Subject: "PROGRESS", BodyMD: "Bead: test-1\nStep: Working"},
		{Subject: "OFFERING_READY", BodyMD: "Bead: test-1\nStatus: DONE"},
	}

	msgs, errs := p.ParseBatch(raws)

	if len(errs) > 0 {
		t.Errorf("Unexpected errors: %v", errs)
	}

	if len(msgs) != 3 {
		t.Fatalf("Messages count = %d, want 3", len(msgs))
	}

	if msgs[0].Type != MessageTypeBeadAccepted {
		t.Errorf("First message type = %q, want %q", msgs[0].Type, MessageTypeBeadAccepted)
	}

	if msgs[1].Type != MessageTypeProgress {
		t.Errorf("Second message type = %q, want %q", msgs[1].Type, MessageTypeProgress)
	}

	if msgs[2].Type != MessageTypeOfferingReady {
		t.Errorf("Third message type = %q, want %q", msgs[2].Type, MessageTypeOfferingReady)
	}
}

func TestFilterByType(t *testing.T) {
	msgs := []*Message{
		{Type: MessageTypeBeadAccepted},
		{Type: MessageTypeProgress},
		{Type: MessageTypeProgress},
		{Type: MessageTypeOfferingReady},
	}

	filtered := FilterByType(msgs, MessageTypeProgress)

	if len(filtered) != 2 {
		t.Errorf("Filtered count = %d, want 2", len(filtered))
	}
}

func TestFilterByBeadID(t *testing.T) {
	msgs := []*Message{
		{Type: MessageTypeProgress, Parsed: ParsedContent{BeadID: "test-1"}},
		{Type: MessageTypeProgress, Parsed: ParsedContent{BeadID: "test-2"}},
		{Type: MessageTypeOfferingReady, Parsed: ParsedContent{BeadID: "test-1"}},
	}

	filtered := FilterByBeadID(msgs, "test-1")

	if len(filtered) != 2 {
		t.Errorf("Filtered count = %d, want 2", len(filtered))
	}
}

func TestFilterCompletions(t *testing.T) {
	msgs := []*Message{
		{Type: MessageTypeProgress},
		{Type: MessageTypeOfferingReady},
		{Type: MessageTypeDone},
		{Type: MessageTypeFailed},
		{Type: MessageTypeCheckpoint},
		{Type: MessageTypeBeadAccepted},
	}

	filtered := FilterCompletions(msgs)

	if len(filtered) != 4 {
		t.Errorf("Filtered count = %d, want 4", len(filtered))
	}
}

func TestFilterSuccesses(t *testing.T) {
	msgs := []*Message{
		{Type: MessageTypeOfferingReady},
		{Type: MessageTypeDone},
		{Type: MessageTypeFailed},
		{Type: MessageTypeCheckpoint},
	}

	filtered := FilterSuccesses(msgs)

	if len(filtered) != 2 {
		t.Errorf("Filtered count = %d, want 2", len(filtered))
	}
}

func TestFilterPending(t *testing.T) {
	msgs := []*Message{
		{Type: MessageTypeOfferingReady, AckRequired: true, Acknowledged: false},
		{Type: MessageTypeProgress, AckRequired: false, Acknowledged: false},
		{Type: MessageTypeHelpRequest, AckRequired: true, Acknowledged: true},
		{Type: MessageTypeDone, AckRequired: true, Acknowledged: false},
	}

	filtered := FilterPending(msgs)

	if len(filtered) != 2 {
		t.Errorf("Filtered count = %d, want 2", len(filtered))
	}
}

func TestParser_ExtractBeadIDFromSubject(t *testing.T) {
	tests := []struct {
		subject string
		want    string
	}{
		{"[ol-527.1] PROGRESS", "ol-527.1"},
		{"[test-123] DONE", "test-123"},
		{"[ag-p43.13] OFFERING_READY", "ag-p43.13"},
		{"ol-527.1: PROGRESS", "ol-527.1"},
		{"PROGRESS", ""},
		{"Just a message", ""},
	}

	p := NewParser()

	for _, tc := range tests {
		t.Run(tc.subject, func(t *testing.T) {
			got := p.extractBeadIDFromSubject(tc.subject)
			if got != tc.want {
				t.Errorf("extractBeadIDFromSubject(%q) = %q, want %q", tc.subject, got, tc.want)
			}
		})
	}
}

func TestParser_ParseValidationFailures(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		Subject: "OFFERING_READY",
		BodyMD: `Bead: test-fail
Status: DONE

## Self-Validation
- Tests: FAIL
- Lint: PASS
- Build: FAIL`,
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Parsed.TestsPass {
		t.Error("TestsPass should be false")
	}

	if !msg.Parsed.LintPass {
		t.Error("LintPass should be true")
	}

	if msg.Parsed.BuildPass {
		t.Error("BuildPass should be false")
	}
}

func TestParser_ParseFileListFormats(t *testing.T) {
	p := NewParser()

	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantItem string
	}{
		{
			name:     "comma separated",
			input:    "src/auth.py, tests/test_auth.py, lib/utils.py",
			wantLen:  3,
			wantItem: "src/auth.py",
		},
		{
			name:     "newline separated",
			input:    "src/auth.py\ntests/test_auth.py",
			wantLen:  2,
			wantItem: "src/auth.py",
		},
		{
			name:     "markdown list",
			input:    "- src/auth.py\n- tests/test_auth.py",
			wantLen:  2,
			wantItem: "src/auth.py",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			files := p.parseFileList(tc.input)
			if len(files) != tc.wantLen {
				t.Errorf("parseFileList() len = %d, want %d", len(files), tc.wantLen)
			}
			if len(files) > 0 && files[0] != tc.wantItem {
				t.Errorf("parseFileList() first = %q, want %q", files[0], tc.wantItem)
			}
		})
	}
}

func TestParser_HelpResponse(t *testing.T) {
	p := NewParser()

	raw := &RawMessage{
		ID:         "msg-012",
		SenderName: "chiron@olympus",
		To:         "demigod-ol-527-2",
		Subject:    "HELP_RESPONSE",
		BodyMD:     "The authentication module is located at lib/auth/middleware.py",
		ThreadID:   "ol-527.2",
	}

	msg, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if msg.Type != MessageTypeHelpResponse {
		t.Errorf("Type = %q, want %q", msg.Type, MessageTypeHelpResponse)
	}
}
