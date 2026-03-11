package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/types"
)

func TestExtractLearningID_RelativePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{".agents/learnings/2026-02-24-abc.md", "2026-02-24-abc.md"},
		{".agents/patterns/auth-pattern.md", "auth-pattern.md"},
		{".agents/learnings/nested/deep.md", "nested/deep.md"},
		{"random/path/file.md", "file.md"}, // fallback to Base
	}
	for _, tt := range tests {
		got := extractLearningID(tt.input)
		if got != tt.want {
			t.Errorf("extractLearningID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractLearningID_AbsolutePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/home/user/repo/.agents/learnings/2026-02-24-abc.md", "2026-02-24-abc.md"},
		{"/tmp/work/.agents/patterns/go-pattern.md", "go-pattern.md"},
	}
	for _, tt := range tests {
		got := extractLearningID(tt.input)
		if got != tt.want {
			t.Errorf("extractLearningID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMarkCitationsFeedbackGiven_WritesAllTrue(t *testing.T) {
	tmp := t.TempDir()
	citationsPath := filepath.Join(tmp, "citations.jsonl")

	citations := []types.CitationEvent{
		{ArtifactPath: ".agents/learnings/a.md", FeedbackGiven: false},
		{ArtifactPath: ".agents/learnings/b.md", FeedbackGiven: false},
		{ArtifactPath: ".agents/learnings/c.md", FeedbackGiven: true}, // already processed
	}

	markCitationsFeedbackGiven(citationsPath, citations)

	data, err := os.ReadFile(citationsPath)
	if err != nil {
		t.Fatalf("failed to read citations: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		var c types.CitationEvent
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			t.Fatalf("line %d: unmarshal error: %v", i, err)
		}
		if !c.FeedbackGiven {
			t.Errorf("line %d: expected FeedbackGiven=true, got false", i)
		}
	}
}

func TestMarkCitationsFeedbackGiven_EmptyList(t *testing.T) {
	tmp := t.TempDir()
	citationsPath := filepath.Join(tmp, "citations.jsonl")

	markCitationsFeedbackGiven(citationsPath, nil)

	data, err := os.ReadFile(citationsPath)
	if err != nil {
		t.Fatalf("failed to read citations: %v", err)
	}

	// Should just have a trailing newline
	if strings.TrimSpace(string(data)) != "" {
		t.Errorf("expected empty content, got %q", string(data))
	}
}

func TestProcessCitationFeedback_NoCitationsFile(t *testing.T) {
	tmp := t.TempDir()
	total, rewarded, skipped := processCitationFeedback(tmp)
	if total != 0 || rewarded != 0 || skipped != 0 {
		t.Errorf("expected (0,0,0), got (%d,%d,%d)", total, rewarded, skipped)
	}
}

func TestProcessCitationFeedback_AllAlreadyProcessed(t *testing.T) {
	tmp := t.TempDir()
	aoDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(aoDir, 0755); err != nil {
		t.Fatal(err)
	}

	citations := []types.CitationEvent{
		{ArtifactPath: ".agents/learnings/a.md", FeedbackGiven: true},
		{ArtifactPath: ".agents/learnings/b.md", FeedbackGiven: true},
	}

	var lines []string
	for _, c := range citations {
		data, _ := json.Marshal(c)
		lines = append(lines, string(data))
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(aoDir, "citations.jsonl"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	total, rewarded, skipped := processCitationFeedback(tmp)
	if total != 0 || rewarded != 0 || skipped != 0 {
		t.Errorf("expected (0,0,0) for all-processed, got (%d,%d,%d)", total, rewarded, skipped)
	}
}

func TestProcessCitationFeedback_UsesAdaptiveReward(t *testing.T) {
	// Setup: temp dir with citations and a learning file.
	// Override HOME so computeSessionRewardForCloseLoop finds no transcripts,
	// causing reward to fall back to 0.5 (InitialUtility).
	// With EMA: new = (1-0.1)*0.5 + 0.1*0.5 = 0.5 (unchanged).
	// If the old hardcoded 1.0 were used: new = (1-0.1)*0.5 + 0.1*1.0 = 0.55.
	tmp := t.TempDir()

	// Isolate from real transcripts by overriding HOME
	fakeHome := filepath.Join(tmp, "fakehome")
	if err := os.MkdirAll(fakeHome, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)

	// Create .agents/ao/ for citations
	aoDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(aoDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create .agents/learnings/ with a learning file (JSONL format)
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	learningPath := filepath.Join(learningsDir, "test-learning.jsonl")
	learningContent := `{"id":"test-learning","title":"Test Learning","utility":0.5}`
	if err := os.WriteFile(learningPath, []byte(learningContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write an unprocessed citation pointing at the learning
	citations := []types.CitationEvent{
		{ArtifactPath: ".agents/learnings/test-learning.jsonl", FeedbackGiven: false},
	}
	var citationLines []string
	for _, c := range citations {
		data, _ := json.Marshal(c)
		citationLines = append(citationLines, string(data))
	}
	citationsContent := strings.Join(citationLines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(aoDir, "citations.jsonl"), []byte(citationsContent), 0600); err != nil {
		t.Fatal(err)
	}

	total, rewarded, skipped := processCitationFeedback(tmp)
	if total != 1 {
		t.Errorf("expected total=1, got %d", total)
	}
	if rewarded != 1 {
		t.Errorf("expected rewarded=1, got %d", rewarded)
	}
	if skipped != 0 {
		t.Errorf("expected skipped=0, got %d", skipped)
	}

	// Verify the learning's utility was updated with fallback reward (0.5), NOT 1.0.
	// EMA with reward=0.5, alpha=0.1, old=0.5: new = 0.9*0.5 + 0.1*0.5 = 0.5
	updatedData, err := os.ReadFile(learningPath)
	if err != nil {
		t.Fatalf("failed to read updated learning: %v", err)
	}

	var parsed map[string]any
	firstLine := strings.Split(string(updatedData), "\n")[0]
	if err := json.Unmarshal([]byte(firstLine), &parsed); err != nil {
		t.Fatalf("failed to parse updated learning: %v", err)
	}

	utility, ok := parsed["utility"].(float64)
	if !ok {
		t.Fatal("utility field not found in updated learning")
	}
	// With fallback reward 0.5, utility should remain 0.5 (unchanged).
	// With old hardcoded 1.0, it would be 0.55.
	if utility > 0.501 || utility < 0.499 {
		t.Errorf("utility = %f, want ~0.5 (fallback adaptive reward); if 0.55, hardcoded 1.0 is still in use", utility)
	}
}

func TestProcessCitationFeedback_WritesFeedbackEvents(t *testing.T) {
	// Verify that processCitationFeedback writes FeedbackEvent entries to feedback.jsonl.
	tmp := t.TempDir()

	// Isolate from real transcripts by overriding HOME
	fakeHome := filepath.Join(tmp, "fakehome")
	if err := os.MkdirAll(fakeHome, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)

	// Create .agents/ao/ for citations
	aoDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(aoDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a learning file
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	learningPath := filepath.Join(learningsDir, "fb-test.jsonl")
	if err := os.WriteFile(learningPath, []byte(`{"id":"fb-test","title":"Feedback Test","utility":0.6}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Write an unprocessed citation
	citations := []types.CitationEvent{
		{ArtifactPath: ".agents/learnings/fb-test.jsonl", FeedbackGiven: false},
	}
	var citationLines []string
	for _, c := range citations {
		data, _ := json.Marshal(c)
		citationLines = append(citationLines, string(data))
	}
	if err := os.WriteFile(filepath.Join(aoDir, "citations.jsonl"), []byte(strings.Join(citationLines, "\n")+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	total, rewarded, _ := processCitationFeedback(tmp)
	if total != 1 || rewarded != 1 {
		t.Fatalf("expected (1,1,_), got (%d,%d,_)", total, rewarded)
	}

	// Verify feedback.jsonl was written
	feedbackPath := filepath.Join(tmp, ".agents", "ao", "feedback.jsonl")
	feedbackData, err := os.ReadFile(feedbackPath)
	if err != nil {
		t.Fatalf("feedback.jsonl not created: %v", err)
	}

	feedbackLines := strings.Split(strings.TrimSpace(string(feedbackData)), "\n")
	if len(feedbackLines) == 0 {
		t.Fatal("feedback.jsonl is empty")
	}

	var event FeedbackEvent
	if err := json.Unmarshal([]byte(feedbackLines[0]), &event); err != nil {
		t.Fatalf("failed to parse FeedbackEvent: %v", err)
	}

	if event.SessionID == "" {
		t.Error("FeedbackEvent.SessionID is empty")
	}
	if event.ArtifactPath == "" {
		t.Error("FeedbackEvent.ArtifactPath is empty")
	}
	if event.UtilityBefore < 0.001 {
		t.Errorf("FeedbackEvent.UtilityBefore = %f, expected non-zero", event.UtilityBefore)
	}
	// For a new learning (0 reward_count), annealed alpha = DefaultAlpha * 3.0
	expectedAlpha := annealedAlpha(types.DefaultAlpha, 0)
	if event.Alpha != expectedAlpha {
		t.Errorf("FeedbackEvent.Alpha = %f, want %f (annealed)", event.Alpha, expectedAlpha)
	}
	if event.RecordedAt.IsZero() {
		t.Error("FeedbackEvent.RecordedAt is zero")
	}
	// Reward should be the fallback (0.5) since no transcript exists with fake HOME
	if event.Reward < 0.499 || event.Reward > 0.501 {
		t.Errorf("FeedbackEvent.Reward = %f, want ~0.5 (fallback)", event.Reward)
	}
}

func TestUpgradeCitationType_PositiveReward(t *testing.T) {
	citations := []types.CitationEvent{
		{ArtifactPath: "/repo/.agents/learnings/a.md", CitationType: "retrieved"},
		{ArtifactPath: "/repo/.agents/learnings/b.md", CitationType: "retrieved"},
	}
	upgradeCitationType(citations, "/repo/.agents/learnings/a.md")
	if citations[0].CitationType != "applied" {
		t.Errorf("expected 'applied', got %q", citations[0].CitationType)
	}
	if citations[1].CitationType != "retrieved" {
		t.Errorf("expected b to remain 'retrieved', got %q", citations[1].CitationType)
	}
}

func TestUpgradeCitationType_AlreadyApplied(t *testing.T) {
	citations := []types.CitationEvent{
		{ArtifactPath: "/repo/.agents/learnings/a.md", CitationType: "applied"},
	}
	upgradeCitationType(citations, "/repo/.agents/learnings/a.md")
	if citations[0].CitationType != "applied" {
		t.Errorf("expected 'applied' unchanged, got %q", citations[0].CitationType)
	}
}

func TestUpgradeCitationType_ReferenceUnchanged(t *testing.T) {
	citations := []types.CitationEvent{
		{ArtifactPath: "/repo/.agents/learnings/a.md", CitationType: "reference"},
	}
	upgradeCitationType(citations, "/repo/.agents/learnings/a.md")
	if citations[0].CitationType != "reference" {
		t.Errorf("expected 'reference' unchanged, got %q", citations[0].CitationType)
	}
}

func TestUpgradeCitationType_NoMatch(t *testing.T) {
	citations := []types.CitationEvent{
		{ArtifactPath: "/repo/.agents/learnings/a.md", CitationType: "retrieved"},
	}
	upgradeCitationType(citations, "/repo/.agents/learnings/other.md")
	if citations[0].CitationType != "retrieved" {
		t.Errorf("expected 'retrieved' unchanged, got %q", citations[0].CitationType)
	}
}

func TestComputeSessionRewardForCloseLoop_NoTranscript(t *testing.T) {
	// Override HOME to an empty temp dir so no transcripts are found.
	fakeHome := filepath.Join(t.TempDir(), "emptyhome")
	if err := os.MkdirAll(fakeHome, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", fakeHome)

	// With no outcome file and no transcript, should return InitialUtility (neutral fallback)
	reward, err := computeSessionRewardForCloseLoop(t.TempDir())
	if err != nil {
		t.Errorf("expected nil error for graceful fallback, got: %v", err)
	}
	if reward != 0.5 {
		t.Errorf("expected InitialUtility (0.5) as fallback, got: %v", reward)
	}
}

func TestFlywheelCitationFeedback_FindingUpdatesCitationFields(t *testing.T) {
	tmp := t.TempDir()
	aoDir := filepath.Join(tmp, ".agents", "ao")
	findingsDir := filepath.Join(tmp, ".agents", SectionFindings)
	if err := os.MkdirAll(aoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(findingsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	findingPath := filepath.Join(findingsDir, "f-cited.md")
	if err := os.WriteFile(findingPath, []byte(`---
id: f-cited
title: Cited finding
status: active
hit_count: 1
---

# Cited finding

Summary.
`), 0o644); err != nil {
		t.Fatal(err)
	}

	citation := types.CitationEvent{
		ArtifactPath: filepath.Join(".agents", SectionFindings, "f-cited.md"),
		SessionID:    "session-1",
		CitedAt:      time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC),
	}
	data, err := json.Marshal(citation)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aoDir, "citations.jsonl"), append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}

	total, rewarded, skipped := processCitationFeedback(tmp)
	if total != 1 || rewarded != 1 || skipped != 0 {
		t.Fatalf("expected (1,1,0), got (%d,%d,%d)", total, rewarded, skipped)
	}

	content, err := os.ReadFile(findingPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	if !strings.Contains(text, "hit_count: 2") {
		t.Fatalf("expected updated hit_count, got:\n%s", text)
	}
	if !strings.Contains(text, "last_cited: 2026-03-10T01:00:00Z") {
		t.Fatalf("expected updated last_cited, got:\n%s", text)
	}
}

func TestPromoteCitedLearnings_NoFeedbackFile(t *testing.T) {
	dir := t.TempDir()
	count := promoteCitedLearnings(dir, true)
	if count != 0 {
		t.Errorf("expected 0 promoted for missing feedback file, got %d", count)
	}
}

func TestPromoteCitedLearnings_DryRunReturnsZero(t *testing.T) {
	dir := t.TempDir()
	feedbackDir := filepath.Join(dir, ".agents", "ao")
	if err := os.MkdirAll(feedbackDir, 0o755); err != nil {
		t.Fatal(err)
	}

	evt := FeedbackEvent{
		ArtifactPath: filepath.Join(dir, ".agents", "learnings", "test.md"),
		Reward:       1.0,
	}
	data, _ := json.Marshal(evt)
	if err := os.WriteFile(filepath.Join(feedbackDir, "feedback.jsonl"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	count := promoteCitedLearnings(dir, true)
	if count != 0 {
		t.Errorf("expected 0 in dry-run mode, got %d", count)
	}
}

func TestPromoteCitedLearnings_EmptyFeedbackFile(t *testing.T) {
	dir := t.TempDir()
	feedbackDir := filepath.Join(dir, ".agents", "ao")
	if err := os.MkdirAll(feedbackDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(feedbackDir, "feedback.jsonl"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	count := promoteCitedLearnings(dir, true)
	if count != 0 {
		t.Errorf("expected 0 for empty feedback, got %d", count)
	}
}

func TestPromoteCitedLearnings_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	feedbackDir := filepath.Join(dir, ".agents", "ao")
	if err := os.MkdirAll(feedbackDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(feedbackDir, "feedback.jsonl"), []byte("not json\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	count := promoteCitedLearnings(dir, true)
	if count != 0 {
		t.Errorf("expected 0 for malformed JSON, got %d", count)
	}
}
