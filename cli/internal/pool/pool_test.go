package pool

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/types"
)

func TestNewPool(t *testing.T) {
	p := NewPool("/tmp/test")
	if p.BaseDir != "/tmp/test" {
		t.Errorf("expected BaseDir /tmp/test, got %s", p.BaseDir)
	}
	if p.PoolPath != "/tmp/test/.agents/pool" {
		t.Errorf("expected PoolPath /tmp/test/.agents/pool, got %s", p.PoolPath)
	}
}

func TestPoolInit(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	if err := p.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Check directories were created
	dirs := []string{
		filepath.Join(p.PoolPath, PendingDir),
		filepath.Join(p.PoolPath, StagedDir),
		filepath.Join(p.PoolPath, RejectedDir),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("directory not created: %s", dir)
		}
	}
}

func TestPoolAddAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	candidate := types.Candidate{
		ID:         "test-candidate-1",
		Type:       types.KnowledgeTypeLearning,
		Tier:       types.TierSilver,
		Content:    "Test learning content",
		Utility:    0.75,
		Confidence: 0.8,
		Maturity:   types.MaturityCandidate,
		Source: types.Source{
			SessionID:      "session-123",
			TranscriptPath: "/path/to/transcript.jsonl",
		},
	}

	scoring := types.Scoring{
		RawScore: 0.72,
		Rubric: types.RubricScores{
			Specificity:   0.8,
			Actionability: 0.7,
			Novelty:       0.6,
			Context:       0.75,
			Confidence:    0.8,
		},
	}

	// Add candidate
	if err := p.Add(candidate, scoring); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Get candidate
	entry, err := p.Get("test-candidate-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if entry.Candidate.ID != "test-candidate-1" {
		t.Errorf("expected ID test-candidate-1, got %s", entry.Candidate.ID)
	}
	if entry.Candidate.Tier != types.TierSilver {
		t.Errorf("expected tier silver, got %s", entry.Candidate.Tier)
	}
}

func TestPoolList(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	// Add test candidates
	candidates := []types.Candidate{
		{ID: "gold-1", Tier: types.TierGold, Content: "Gold content"},
		{ID: "silver-1", Tier: types.TierSilver, Content: "Silver content"},
		{ID: "bronze-1", Tier: types.TierBronze, Content: "Bronze content"},
	}

	for _, c := range candidates {
		if err := p.Add(c, types.Scoring{}); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// List all
	entries, err := p.List(ListOptions{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// List by tier
	goldEntries, err := p.List(ListOptions{Tier: types.TierGold})
	if err != nil {
		t.Fatalf("List gold failed: %v", err)
	}
	if len(goldEntries) != 1 {
		t.Errorf("expected 1 gold entry, got %d", len(goldEntries))
	}
}

func TestPoolStageAndPromote(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	candidate := types.Candidate{
		ID:       "promote-test",
		Tier:     types.TierSilver,
		Type:     types.KnowledgeTypeLearning,
		Content:  "Promotable learning",
		Maturity: types.MaturityCandidate,
	}

	if err := p.Add(candidate, types.Scoring{}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Stage
	if err := p.Stage("promote-test", types.TierBronze); err != nil {
		t.Fatalf("Stage failed: %v", err)
	}

	// Verify staged
	entry, err := p.Get("promote-test")
	if err != nil {
		t.Fatalf("Get after stage failed: %v", err)
	}
	if entry.Status != types.PoolStatusStaged {
		t.Errorf("expected status staged, got %s", entry.Status)
	}

	// Promote
	artifactPath, err := p.Promote("promote-test")
	if err != nil {
		t.Fatalf("Promote failed: %v", err)
	}
	if artifactPath == "" {
		t.Error("expected artifact path, got empty")
	}

	// Verify artifact exists
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		t.Errorf("artifact not created: %s", artifactPath)
	}
}

func TestPoolReject(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	candidate := types.Candidate{
		ID:      "reject-test",
		Tier:    types.TierBronze,
		Content: "Rejectable content",
	}

	if err := p.Add(candidate, types.Scoring{}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Reject
	if err := p.Reject("reject-test", "Too vague", "tester"); err != nil {
		t.Fatalf("Reject failed: %v", err)
	}

	// Verify rejected
	entry, err := p.Get("reject-test")
	if err != nil {
		t.Fatalf("Get after reject failed: %v", err)
	}
	if entry.Status != types.PoolStatusRejected {
		t.Errorf("expected status rejected, got %s", entry.Status)
	}
	if entry.HumanReview == nil || entry.HumanReview.Notes != "Too vague" {
		t.Error("rejection reason not recorded")
	}
}

func TestPoolRejectPreventsPromotion(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	candidate := types.Candidate{
		ID:      "reject-promote-test",
		Tier:    types.TierSilver,
		Type:    types.KnowledgeTypeLearning,
		Content: "Rejectable content",
	}

	if err := p.Add(candidate, types.Scoring{}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Reject the candidate
	if err := p.Reject("reject-promote-test", "Too vague", "tester"); err != nil {
		t.Fatalf("Reject failed: %v", err)
	}

	// Attempt to promote rejected candidate should fail
	_, err := p.Promote("reject-promote-test")
	if err == nil {
		t.Error("expected error when promoting rejected candidate")
	}
	if err.Error() != "cannot promote rejected candidate" {
		t.Errorf("unexpected error message: %v", err)
	}

	// Attempt to stage rejected candidate should also fail
	err = p.Stage("reject-promote-test", types.TierBronze)
	if err == nil {
		t.Error("expected error when staging rejected candidate")
	}
	if err.Error() != "cannot stage rejected candidate" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPoolBulkApprove(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	// Add old silver candidates
	for i := 0; i < 3; i++ {
		candidate := types.Candidate{
			ID:      string(rune('a'+i)) + "-silver",
			Tier:    types.TierSilver,
			Content: "Silver content",
		}
		if err := p.Add(candidate, types.Scoring{}); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
	}

	// Bulk approve with minimum valid threshold (1 hour)
	// Candidates were just added, so they won't match the threshold,
	// but this tests the function doesn't error.
	approved, err := p.BulkApprove(time.Hour, "bulk-tester", false)
	if err != nil {
		t.Fatalf("BulkApprove failed: %v", err)
	}
	// Candidates were just added, so none should be older than 1 hour
	if len(approved) != 0 {
		t.Errorf("expected 0 approved (none old enough), got %d", len(approved))
	}
}

func TestPoolBulkApproveThresholdTooLow(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	// Threshold below minimum should return error
	_, err := p.BulkApprove(0, "bulk-tester", false)
	if err != ErrThresholdTooLow {
		t.Errorf("expected ErrThresholdTooLow, got %v", err)
	}

	// Just under 1 hour should also fail
	_, err = p.BulkApprove(59*time.Minute, "bulk-tester", false)
	if err != ErrThresholdTooLow {
		t.Errorf("expected ErrThresholdTooLow for 59m, got %v", err)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{30 * time.Minute, "30m"},
		{2 * time.Hour, "2h"},
		{48 * time.Hour, "2d"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.d)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %s, expected %s", tt.d, result, tt.expected)
		}
	}
}

func TestIsAboveThreshold(t *testing.T) {
	tests := []struct {
		tier     types.Tier
		minTier  types.Tier
		expected bool
	}{
		{types.TierGold, types.TierBronze, true},
		{types.TierSilver, types.TierSilver, true},
		{types.TierBronze, types.TierSilver, false},
		{types.TierGold, types.TierGold, true},
	}

	for _, tt := range tests {
		result := isAboveThreshold(tt.tier, tt.minTier)
		if result != tt.expected {
			t.Errorf("isAboveThreshold(%s, %s) = %v, expected %v",
				tt.tier, tt.minTier, result, tt.expected)
		}
	}
}

func TestPoolApprove(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	candidate := types.Candidate{
		ID:      "approve-test",
		Tier:    types.TierBronze,
		Content: "Content to approve",
	}

	if err := p.Add(candidate, types.Scoring{GateRequired: true}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// First approval should succeed
	if err := p.Approve("approve-test", "First approval", "first-reviewer"); err != nil {
		t.Fatalf("First Approve failed: %v", err)
	}

	// Verify review was recorded
	entry, err := p.Get("approve-test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if entry.HumanReview == nil || !entry.HumanReview.Reviewed {
		t.Error("HumanReview not recorded after approval")
	}
	if entry.HumanReview.Reviewer != "first-reviewer" {
		t.Errorf("expected reviewer first-reviewer, got %s", entry.HumanReview.Reviewer)
	}
}

func TestPoolApproveAlreadyReviewed(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	candidate := types.Candidate{
		ID:      "already-reviewed",
		Tier:    types.TierBronze,
		Content: "Already reviewed content",
	}

	if err := p.Add(candidate, types.Scoring{GateRequired: true}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// First approval
	if err := p.Approve("already-reviewed", "First note", "first-reviewer"); err != nil {
		t.Fatalf("First Approve failed: %v", err)
	}

	// Second approval should fail with "already reviewed by X"
	err := p.Approve("already-reviewed", "Second note", "second-reviewer")
	if err == nil {
		t.Fatal("Expected error for already-reviewed candidate")
	}

	expectedMsg := "already reviewed by first-reviewer"
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestPoolListPendingReview(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	// Add bronze candidates (only bronze should appear in pending review)
	bronzeCandidate := types.Candidate{
		ID:      "bronze-pending",
		Tier:    types.TierBronze,
		Content: "Bronze content",
	}
	silverCandidate := types.Candidate{
		ID:      "silver-no-review",
		Tier:    types.TierSilver,
		Content: "Silver content",
	}

	if err := p.Add(bronzeCandidate, types.Scoring{GateRequired: true}); err != nil {
		t.Fatalf("Add bronze failed: %v", err)
	}
	if err := p.Add(silverCandidate, types.Scoring{GateRequired: false}); err != nil {
		t.Fatalf("Add silver failed: %v", err)
	}

	pending, err := p.ListPendingReview()
	if err != nil {
		t.Fatalf("ListPendingReview failed: %v", err)
	}

	// Should only return bronze candidates awaiting review
	if len(pending) != 1 {
		t.Errorf("expected 1 pending review (bronze only), got %d", len(pending))
	}

	if len(pending) > 0 && pending[0].Candidate.ID != "bronze-pending" {
		t.Errorf("expected bronze-pending, got %s", pending[0].Candidate.ID)
	}
}

func TestPoolRejectReasonTooLong(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	candidate := types.Candidate{
		ID:      "reason-length-test",
		Tier:    types.TierBronze,
		Content: "Content to reject with long reason",
	}

	if err := p.Add(candidate, types.Scoring{GateRequired: true}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Create a reason that exceeds MaxReasonLength (1000 chars)
	longReason := make([]byte, MaxReasonLength+1)
	for i := range longReason {
		longReason[i] = 'x'
	}

	err := p.Reject("reason-length-test", string(longReason), "reviewer")
	if err != ErrReasonTooLong {
		t.Errorf("expected ErrReasonTooLong, got %v", err)
	}

	// Exactly at max should succeed
	exactReason := make([]byte, MaxReasonLength)
	for i := range exactReason {
		exactReason[i] = 'x'
	}
	err = p.Reject("reason-length-test", string(exactReason), "reviewer")
	if err != nil {
		t.Errorf("expected nil error for reason at max length, got %v", err)
	}
}

func TestPoolApproveNoteTooLong(t *testing.T) {
	tmpDir := t.TempDir()
	p := NewPool(tmpDir)

	candidate := types.Candidate{
		ID:      "note-length-test",
		Tier:    types.TierBronze,
		Content: "Content to approve with long note",
	}

	if err := p.Add(candidate, types.Scoring{GateRequired: true}); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	// Create a note that exceeds MaxReasonLength (1000 chars)
	longNote := make([]byte, MaxReasonLength+1)
	for i := range longNote {
		longNote[i] = 'x'
	}

	err := p.Approve("note-length-test", string(longNote), "reviewer")
	if err != ErrReasonTooLong {
		t.Errorf("expected ErrReasonTooLong, got %v", err)
	}

	// Exactly at max should succeed
	exactNote := make([]byte, MaxReasonLength)
	for i := range exactNote {
		exactNote[i] = 'x'
	}
	err = p.Approve("note-length-test", string(exactNote), "reviewer")
	if err != nil {
		t.Errorf("expected nil error for note at max length, got %v", err)
	}
}

func TestTruncateAtWordBoundary(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		limit    int
		expected string
	}{
		{
			name:     "short string no truncation",
			input:    "hello world",
			limit:    77,
			expected: "hello world",
		},
		{
			name:     "truncate at word boundary",
			input:    "This is a very long string that needs to be truncated at word boundary properly",
			limit:    40,
			expected: "This is a very long string that needs",
		},
		{
			name:     "no spaces in truncation zone",
			input:    "superlongwordwithoutspaces and more",
			limit:    25,
			expected: "superlongwordwithoutspace",
		},
		{
			name:     "truncate respects last space",
			input:    "word1 word2 word3 word4 word5",
			limit:    15,
			expected: "word1 word2",
		},
		{
			name:     "exact limit equals length",
			input:    "hello",
			limit:    5,
			expected: "hello",
		},
		{
			name:     "single word longer than limit",
			input:    "supercalifragilistic",
			limit:    10,
			expected: "supercalif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateAtWordBoundary(tt.input, tt.limit)
			if result != tt.expected {
				t.Errorf("truncateAtWordBoundary(%q, %d) = %q, expected %q",
					tt.input, tt.limit, result, tt.expected)
			}
		})
	}
}
