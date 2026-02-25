package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/ratchet"
	"github.com/boshu2/agentops/cli/internal/types"
)

// ===========================================================================
// feedback_loop.go — loadSessionCitations (zero coverage)
// ===========================================================================

func TestCov3_feedbackLoop_loadSessionCitations_noCitations(t *testing.T) {
	tmp := t.TempDir()

	citations, err := loadSessionCitations(tmp, "session-test", "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(citations) != 0 {
		t.Errorf("expected 0 citations, got %d", len(citations))
	}
}

func TestCov3_feedbackLoop_loadSessionCitations_filterByType(t *testing.T) {
	tmp := t.TempDir()
	sessionID := "session-20260201-120000"

	// Write citations with different types
	for _, entry := range []types.CitationEvent{
		{ArtifactPath: filepath.Join(tmp, "L1.md"), SessionID: sessionID, CitedAt: time.Now(), CitationType: "retrieved"},
		{ArtifactPath: filepath.Join(tmp, "L2.md"), SessionID: sessionID, CitedAt: time.Now(), CitationType: "applied"},
		{ArtifactPath: filepath.Join(tmp, "L3.md"), SessionID: sessionID, CitedAt: time.Now(), CitationType: "retrieved"},
	} {
		if err := ratchet.RecordCitation(tmp, entry); err != nil {
			t.Fatalf("record citation: %v", err)
		}
	}

	// Filter by "retrieved"
	retrieved, err := loadSessionCitations(tmp, sessionID, "retrieved")
	if err != nil {
		t.Fatalf("loadSessionCitations (retrieved): %v", err)
	}
	if len(retrieved) != 2 {
		t.Errorf("expected 2 retrieved citations, got %d", len(retrieved))
	}

	// Filter by "all"
	all, err := loadSessionCitations(tmp, sessionID, "all")
	if err != nil {
		t.Fatalf("loadSessionCitations (all): %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 total citations, got %d", len(all))
	}
}

func TestCov3_feedbackLoop_loadSessionCitations_wrongSession(t *testing.T) {
	tmp := t.TempDir()

	entry := types.CitationEvent{
		ArtifactPath: filepath.Join(tmp, "L1.md"),
		SessionID:    "session-other",
		CitedAt:      time.Now(),
		CitationType: "retrieved",
	}
	if err := ratchet.RecordCitation(tmp, entry); err != nil {
		t.Fatalf("record citation: %v", err)
	}

	citations, err := loadSessionCitations(tmp, "session-different", "all")
	if err != nil {
		t.Fatalf("loadSessionCitations: %v", err)
	}
	if len(citations) != 0 {
		t.Errorf("expected 0 citations for wrong session, got %d", len(citations))
	}
}

// ===========================================================================
// feedback_loop.go — computeRewardFromTranscript (zero coverage)
// ===========================================================================

func TestCov3_feedbackLoop_computeRewardFromTranscript_noTranscript(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	_, err := computeRewardFromTranscript("", "nonexistent-session")
	if err == nil {
		t.Fatal("expected error when no transcript found")
	}
}

func TestCov3_feedbackLoop_computeRewardFromTranscript_explicitPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Create a minimal transcript
	transcriptPath := filepath.Join(tmp, "transcript.jsonl")
	content := `{"type":"user","sessionId":"test-session"}
{"type":"assistant","message":{"content":"hello"}}
`
	if err := os.WriteFile(transcriptPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	reward, err := computeRewardFromTranscript(transcriptPath, "test-session")
	if err != nil {
		t.Fatalf("computeRewardFromTranscript: %v", err)
	}
	// Reward should be between 0 and 1
	if reward < 0 || reward > 1 {
		t.Errorf("reward %f out of range [0, 1]", reward)
	}
}

// ===========================================================================
// feedback_loop.go — processUniqueCitations (zero coverage)
// ===========================================================================

func TestCov3_feedbackLoop_processUniqueCitations_noFiles(t *testing.T) {
	tmp := t.TempDir()

	citations := []types.CitationEvent{
		{ArtifactPath: filepath.Join(tmp, "nonexistent.md"), SessionID: "s1", CitedAt: time.Now()},
	}

	events, updated, failed := processUniqueCitations(tmp, "s1", "", citations, 0.8, 0.1)
	if updated != 0 {
		t.Errorf("expected 0 updated, got %d", updated)
	}
	if failed != 1 {
		t.Errorf("expected 1 failed, got %d", failed)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestCov3_feedbackLoop_processUniqueCitations_withLearning(t *testing.T) {
	tmp := t.TempDir()

	// Create a learning file
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	learningPath := filepath.Join(learningsDir, "L-test-001.jsonl")
	data := map[string]any{"id": "L-test-001", "utility": 0.5}
	jsonData, _ := json.Marshal(data)
	if err := os.WriteFile(learningPath, append(jsonData, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	// Save and restore feedback flags for counter update in updateJSONLUtility
	origHelpful := feedbackHelpful
	origHarmful := feedbackHarmful
	defer func() {
		feedbackHelpful = origHelpful
		feedbackHarmful = origHarmful
	}()
	feedbackHelpful = false
	feedbackHarmful = false

	citations := []types.CitationEvent{
		{ArtifactPath: learningPath, SessionID: "s1", CitedAt: time.Now()},
	}

	events, updated, failed := processUniqueCitations(tmp, "s1", "", citations, 0.9, 0.1)
	if updated != 1 {
		t.Errorf("expected 1 updated, got %d", updated)
	}
	if failed != 0 {
		t.Errorf("expected 0 failed, got %d", failed)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if len(events) > 0 {
		if events[0].Reward != 0.9 {
			t.Errorf("event reward = %f, want 0.9", events[0].Reward)
		}
	}
}

// ===========================================================================
// feedback_loop.go — resolveFeedbackReward (zero coverage)
// ===========================================================================

func TestCov3_feedbackLoop_resolveFeedbackReward_explicit(t *testing.T) {
	reward, err := resolveFeedbackReward(0.75, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reward != 0.75 {
		t.Errorf("reward = %f, want 0.75", reward)
	}
}

func TestCov3_feedbackLoop_resolveFeedbackReward_zero(t *testing.T) {
	reward, err := resolveFeedbackReward(0.0, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reward != 0.0 {
		t.Errorf("reward = %f, want 0.0", reward)
	}
}

func TestCov3_feedbackLoop_resolveFeedbackReward_one(t *testing.T) {
	reward, err := resolveFeedbackReward(1.0, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reward != 1.0 {
		t.Errorf("reward = %f, want 1.0", reward)
	}
}

func TestCov3_feedbackLoop_resolveFeedbackReward_autoCompute_noTranscript(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	_, err := resolveFeedbackReward(-1, "", "no-such-session")
	if err == nil {
		t.Fatal("expected error when no transcript available for auto-compute")
	}
}

// ===========================================================================
// feedback_loop.go — outputFeedbackSummary (zero coverage)
// ===========================================================================

func TestCov3_feedbackLoop_outputFeedbackSummary_tableFormat(t *testing.T) {
	origOutput := output
	defer func() { output = origOutput }()
	output = "table"

	events := []FeedbackEvent{
		{SessionID: "s1", Reward: 0.8, ArtifactPath: "/a.md", UtilityBefore: 0.5, UtilityAfter: 0.53},
	}
	err := outputFeedbackSummary("s1", 0.8, 3, 2, 1, 0, events)
	if err != nil {
		t.Fatalf("outputFeedbackSummary (table): %v", err)
	}
}

func TestCov3_feedbackLoop_outputFeedbackSummary_tableWithFailed(t *testing.T) {
	origOutput := output
	defer func() { output = origOutput }()
	output = "table"

	err := outputFeedbackSummary("s1", 0.5, 5, 3, 2, 1, nil)
	if err != nil {
		t.Fatalf("outputFeedbackSummary (table with failed): %v", err)
	}
}

func TestCov3_feedbackLoop_outputFeedbackSummary_jsonFormat(t *testing.T) {
	origOutput := output
	defer func() { output = origOutput }()
	output = "json"

	events := []FeedbackEvent{
		{SessionID: "s1", Reward: 0.9, ArtifactPath: "/a.md"},
	}
	err := outputFeedbackSummary("s1", 0.9, 2, 1, 1, 0, events)
	if err != nil {
		t.Fatalf("outputFeedbackSummary (json): %v", err)
	}
}

// ===========================================================================
// feedback_loop.go — discoverUnprocessedSessions (zero coverage)
// ===========================================================================

func TestCov3_feedbackLoop_discoverUnprocessedSessions_noCitations(t *testing.T) {
	tmp := t.TempDir()

	origDays := batchFeedbackDays
	defer func() { batchFeedbackDays = origDays }()
	batchFeedbackDays = 7

	sessions, latest, err := discoverUnprocessedSessions(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
	if len(latest) != 0 {
		t.Errorf("expected 0 latest, got %d", len(latest))
	}
}

func TestCov3_feedbackLoop_discoverUnprocessedSessions_withCitations(t *testing.T) {
	tmp := t.TempDir()

	origDays := batchFeedbackDays
	defer func() { batchFeedbackDays = origDays }()
	batchFeedbackDays = 7

	// Create some citations
	for _, entry := range []types.CitationEvent{
		{ArtifactPath: "/a.md", SessionID: "session-1", CitedAt: time.Now(), CitationType: "retrieved"},
		{ArtifactPath: "/b.md", SessionID: "session-1", CitedAt: time.Now(), CitationType: "retrieved"},
		{ArtifactPath: "/c.md", SessionID: "session-2", CitedAt: time.Now(), CitationType: "retrieved"},
	} {
		if err := ratchet.RecordCitation(tmp, entry); err != nil {
			t.Fatalf("record citation: %v", err)
		}
	}

	sessions, latest, err := discoverUnprocessedSessions(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
	if len(latest) != 2 {
		t.Errorf("expected 2 latest entries, got %d", len(latest))
	}
}

func TestCov3_feedbackLoop_discoverUnprocessedSessions_excludesProcessed(t *testing.T) {
	tmp := t.TempDir()

	origDays := batchFeedbackDays
	defer func() { batchFeedbackDays = origDays }()
	batchFeedbackDays = 7

	// Create a citation
	entry := types.CitationEvent{
		ArtifactPath: "/a.md",
		SessionID:    "session-processed",
		CitedAt:      time.Now(),
		CitationType: "retrieved",
	}
	if err := ratchet.RecordCitation(tmp, entry); err != nil {
		t.Fatalf("record citation: %v", err)
	}

	// Create matching feedback (marks session as processed)
	feedbackDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(feedbackDir, 0755); err != nil {
		t.Fatal(err)
	}
	fbEvent := FeedbackEvent{SessionID: "session-processed", Reward: 0.8}
	fbData, _ := json.Marshal(fbEvent)
	fbPath := filepath.Join(feedbackDir, "feedback.jsonl")
	if err := os.WriteFile(fbPath, append(fbData, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	sessions, _, err := discoverUnprocessedSessions(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 unprocessed sessions (already processed), got %d", len(sessions))
	}
}

// ===========================================================================
// feedback_loop.go — reportBatchFeedbackDryRun (zero coverage)
// ===========================================================================

func TestCov3_feedbackLoop_reportBatchFeedbackDryRun(t *testing.T) {
	sessionIDs := []string{"s1", "s2"}
	sessionCitations := map[string][]types.CitationEvent{
		"s1": {{ArtifactPath: "/a.md"}},
		"s2": {{ArtifactPath: "/b.md"}, {ArtifactPath: "/c.md"}},
	}
	// Should not panic, prints dry-run report
	reportBatchFeedbackDryRun(sessionIDs, 3, sessionCitations)
}

func TestCov3_feedbackLoop_reportBatchFeedbackDryRun_empty(t *testing.T) {
	reportBatchFeedbackDryRun(nil, 0, nil)
}
