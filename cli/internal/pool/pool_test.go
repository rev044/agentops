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

	// Bulk approve with 0 threshold (approve all)
	approved, err := p.BulkApprove(0, "bulk-tester", false)
	if err != nil {
		t.Fatalf("BulkApprove failed: %v", err)
	}
	if len(approved) != 3 {
		t.Errorf("expected 3 approved, got %d", len(approved))
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
