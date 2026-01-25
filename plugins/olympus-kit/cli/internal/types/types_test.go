package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTranscriptMessageJSONRoundTrip(t *testing.T) {
	original := TranscriptMessage{
		Type:         "assistant",
		Timestamp:    time.Date(2026, 1, 24, 10, 30, 0, 0, time.UTC),
		Role:         "assistant",
		Content:      "Let me help you with that.",
		SessionID:    "session-123",
		MessageIndex: 5,
		Tools: []ToolCall{
			{
				Name:   "Read",
				Input:  map[string]interface{}{"file_path": "/tmp/test.go"},
				Output: "file contents",
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded TranscriptMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Type != original.Type {
		t.Errorf("Type mismatch: got %q, want %q", decoded.Type, original.Type)
	}
	if decoded.Role != original.Role {
		t.Errorf("Role mismatch: got %q, want %q", decoded.Role, original.Role)
	}
	if decoded.Content != original.Content {
		t.Errorf("Content mismatch: got %q, want %q", decoded.Content, original.Content)
	}
	if decoded.SessionID != original.SessionID {
		t.Errorf("SessionID mismatch: got %q, want %q", decoded.SessionID, original.SessionID)
	}
	if decoded.MessageIndex != original.MessageIndex {
		t.Errorf("MessageIndex mismatch: got %d, want %d", decoded.MessageIndex, original.MessageIndex)
	}
	if len(decoded.Tools) != 1 {
		t.Fatalf("Tools length mismatch: got %d, want 1", len(decoded.Tools))
	}
	if decoded.Tools[0].Name != "Read" {
		t.Errorf("Tool name mismatch: got %q, want %q", decoded.Tools[0].Name, "Read")
	}
}

func TestCandidateJSONRoundTrip(t *testing.T) {
	original := Candidate{
		ID:      "ol-cand-abc123",
		Type:    KnowledgeTypeDecision,
		Content: "Use context.WithCancel for graceful shutdown",
		Context: "When implementing Go services that need cleanup",
		Source: Source{
			TranscriptPath: "/home/user/.claude/sessions/abc.jsonl",
			MessageIndex:   42,
			Timestamp:      time.Date(2026, 1, 24, 10, 30, 0, 0, time.UTC),
			SessionID:      "session-123",
		},
		RawScore:      0.87,
		Tier:          TierGold,
		ProvenanceIDs: []string{"prov-1", "prov-2"},
		ExtractedAt:   time.Date(2026, 1, 24, 10, 35, 0, 0, time.UTC),
		Metadata: map[string]interface{}{
			"extractor": "transcript-forge",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Candidate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type mismatch: got %q, want %q", decoded.Type, original.Type)
	}
	if decoded.Content != original.Content {
		t.Errorf("Content mismatch: got %q, want %q", decoded.Content, original.Content)
	}
	if decoded.RawScore != original.RawScore {
		t.Errorf("RawScore mismatch: got %f, want %f", decoded.RawScore, original.RawScore)
	}
	if decoded.Tier != original.Tier {
		t.Errorf("Tier mismatch: got %q, want %q", decoded.Tier, original.Tier)
	}
	if decoded.Source.TranscriptPath != original.Source.TranscriptPath {
		t.Errorf("Source.TranscriptPath mismatch: got %q, want %q",
			decoded.Source.TranscriptPath, original.Source.TranscriptPath)
	}
	if len(decoded.ProvenanceIDs) != len(original.ProvenanceIDs) {
		t.Errorf("ProvenanceIDs length mismatch: got %d, want %d",
			len(decoded.ProvenanceIDs), len(original.ProvenanceIDs))
	}
}

func TestScoringJSONRoundTrip(t *testing.T) {
	original := Scoring{
		RawScore:       0.82,
		TierAssignment: TierSilver,
		Rubric: RubricScores{
			Specificity:   0.85,
			Actionability: 0.80,
			Novelty:       0.75,
			Context:       0.90,
			Confidence:    0.70,
		},
		GateRequired: false,
		ScoredAt:     time.Date(2026, 1, 24, 10, 40, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Scoring
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.RawScore != original.RawScore {
		t.Errorf("RawScore mismatch: got %f, want %f", decoded.RawScore, original.RawScore)
	}
	if decoded.TierAssignment != original.TierAssignment {
		t.Errorf("TierAssignment mismatch: got %q, want %q", decoded.TierAssignment, original.TierAssignment)
	}
	if decoded.Rubric.Specificity != original.Rubric.Specificity {
		t.Errorf("Rubric.Specificity mismatch: got %f, want %f",
			decoded.Rubric.Specificity, original.Rubric.Specificity)
	}
	if decoded.GateRequired != original.GateRequired {
		t.Errorf("GateRequired mismatch: got %v, want %v", decoded.GateRequired, original.GateRequired)
	}
}

func TestPoolEntryJSONRoundTrip(t *testing.T) {
	reviewTime := time.Date(2026, 1, 24, 11, 0, 0, 0, time.UTC)
	original := PoolEntry{
		Candidate: Candidate{
			ID:      "ol-cand-xyz789",
			Type:    KnowledgeTypeSolution,
			Content: "Fix rate limiting with backoff",
		},
		ScoringResult: Scoring{
			RawScore:       0.65,
			TierAssignment: TierBronze,
			GateRequired:   true,
		},
		HumanReview: &HumanReview{
			Reviewed:   true,
			Approved:   true,
			Reviewer:   "boden",
			Notes:      "Good insight",
			ReviewedAt: reviewTime,
		},
		Status:    PoolStatusArchived,
		AddedAt:   time.Date(2026, 1, 24, 10, 45, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 24, 11, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded PoolEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Status != original.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, original.Status)
	}
	if decoded.HumanReview == nil {
		t.Fatal("HumanReview is nil")
	}
	if decoded.HumanReview.Approved != original.HumanReview.Approved {
		t.Errorf("HumanReview.Approved mismatch: got %v, want %v",
			decoded.HumanReview.Approved, original.HumanReview.Approved)
	}
	if decoded.HumanReview.Reviewer != original.HumanReview.Reviewer {
		t.Errorf("HumanReview.Reviewer mismatch: got %q, want %q",
			decoded.HumanReview.Reviewer, original.HumanReview.Reviewer)
	}
}

func TestKnowledgeTypeValues(t *testing.T) {
	types := []KnowledgeType{
		KnowledgeTypeDecision,
		KnowledgeTypeSolution,
		KnowledgeTypeLearning,
		KnowledgeTypeFailure,
		KnowledgeTypeReference,
	}

	expected := []string{"decision", "solution", "learning", "failure", "reference"}

	for i, kt := range types {
		if string(kt) != expected[i] {
			t.Errorf("KnowledgeType value mismatch: got %q, want %q", kt, expected[i])
		}
	}
}

func TestTierValues(t *testing.T) {
	tiers := []Tier{TierGold, TierSilver, TierBronze, TierDiscard}
	expected := []string{"gold", "silver", "bronze", "discard"}

	for i, tier := range tiers {
		if string(tier) != expected[i] {
			t.Errorf("Tier value mismatch: got %q, want %q", tier, expected[i])
		}
	}
}

func TestPoolStatusValues(t *testing.T) {
	statuses := []PoolStatus{
		PoolStatusPending,
		PoolStatusStaged,
		PoolStatusArchived,
		PoolStatusRejected,
	}
	expected := []string{"pending", "staged", "archived", "rejected"}

	for i, status := range statuses {
		if string(status) != expected[i] {
			t.Errorf("PoolStatus value mismatch: got %q, want %q", status, expected[i])
		}
	}
}

// --- Supersession tests (ol-a46.1.4) ---

func TestSupersede(t *testing.T) {
	older := &Candidate{
		ID:                "L1",
		Type:              KnowledgeTypeLearning,
		Content:           "Original learning",
		IsCurrent:         true,
		SupersessionDepth: 0,
	}

	newer := &Candidate{
		ID:      "L2",
		Type:    KnowledgeTypeLearning,
		Content: "Updated learning",
	}

	err := Supersede(older, newer)
	if err != nil {
		t.Fatalf("Supersede failed: %v", err)
	}

	// Check older candidate
	if older.SupersededBy != "L2" {
		t.Errorf("older.SupersededBy: got %q, want %q", older.SupersededBy, "L2")
	}
	if older.IsCurrent {
		t.Error("older.IsCurrent should be false")
	}

	// Check newer candidate
	if newer.Supersedes != "L1" {
		t.Errorf("newer.Supersedes: got %q, want %q", newer.Supersedes, "L1")
	}
	if !newer.IsCurrent {
		t.Error("newer.IsCurrent should be true")
	}
	if newer.SupersessionDepth != 1 {
		t.Errorf("newer.SupersessionDepth: got %d, want 1", newer.SupersessionDepth)
	}
}

func TestSupersede_MaxDepth(t *testing.T) {
	// Create a chain at max depth
	older := &Candidate{
		ID:                "L3",
		Type:              KnowledgeTypeLearning,
		Content:           "Learning at max depth",
		IsCurrent:         true,
		SupersessionDepth: MaxSupersessionDepth, // Already at max
	}

	newer := &Candidate{
		ID:      "L4",
		Type:    KnowledgeTypeLearning,
		Content: "Would exceed max depth",
	}

	err := Supersede(older, newer)
	if err == nil {
		t.Fatal("Expected error for exceeding max depth, got nil")
	}

	supersessionErr, ok := err.(*SupersessionError)
	if !ok {
		t.Fatalf("Expected SupersessionError, got %T", err)
	}

	if supersessionErr.Depth != MaxSupersessionDepth+1 {
		t.Errorf("Error depth: got %d, want %d", supersessionErr.Depth, MaxSupersessionDepth+1)
	}
}

func TestValidateSupersessionDepth(t *testing.T) {
	tests := []struct {
		name    string
		depth   int
		wantErr bool
	}{
		{"depth 0", 0, false},
		{"depth 1", 1, false},
		{"depth 2", 2, false},
		{"depth 3 (max)", 3, false},
		{"depth 4 (exceeds)", 4, true},
		{"depth 10 (way over)", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Candidate{
				ID:                "test",
				SupersessionDepth: tt.depth,
			}

			err := ValidateSupersessionDepth(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSupersessionDepth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCandidate_IsSuperseded(t *testing.T) {
	notSuperseded := &Candidate{ID: "L1"}
	if notSuperseded.IsSuperseded() {
		t.Error("Expected IsSuperseded() = false for empty SupersededBy")
	}

	superseded := &Candidate{ID: "L1", SupersededBy: "L2"}
	if !superseded.IsSuperseded() {
		t.Error("Expected IsSuperseded() = true when SupersededBy is set")
	}
}

func TestCandidateSupersessionJSONRoundTrip(t *testing.T) {
	original := Candidate{
		ID:                "ol-cand-learning-001",
		Type:              KnowledgeTypeLearning,
		Content:           "Updated learning about context usage",
		IsCurrent:         true,
		Supersedes:        "ol-cand-learning-000",
		SupersededBy:      "",
		SupersessionDepth: 1,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Candidate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.IsCurrent != original.IsCurrent {
		t.Errorf("IsCurrent mismatch: got %v, want %v", decoded.IsCurrent, original.IsCurrent)
	}
	if decoded.Supersedes != original.Supersedes {
		t.Errorf("Supersedes mismatch: got %q, want %q", decoded.Supersedes, original.Supersedes)
	}
	if decoded.SupersededBy != original.SupersededBy {
		t.Errorf("SupersededBy mismatch: got %q, want %q", decoded.SupersededBy, original.SupersededBy)
	}
	if decoded.SupersessionDepth != original.SupersessionDepth {
		t.Errorf("SupersessionDepth mismatch: got %d, want %d", decoded.SupersessionDepth, original.SupersessionDepth)
	}
}
