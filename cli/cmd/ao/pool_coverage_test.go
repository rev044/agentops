package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/pool"
	"github.com/boshu2/agentops/cli/internal/types"
)

// ---------------------------------------------------------------------------
// pool.go — outputPoolList
// ---------------------------------------------------------------------------

func TestCov3_pool_outputPoolList_emptyTable(t *testing.T) {
	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputPoolList(nil, 0, 50, 0)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputPoolList: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "No pool entries found") {
		t.Errorf("expected 'No pool entries found', got: %s", out)
	}
}

func TestCov3_pool_outputPoolList_jsonMode(t *testing.T) {
	oldOutput := output
	output = "json"
	defer func() { output = oldOutput }()

	entries := []pool.PoolEntry{
		{
			PoolEntry: types.PoolEntry{
				Candidate: types.Candidate{
					ID:         "cand-test-001",
					Tier:       types.TierSilver,
					Utility:    0.75,
					Confidence: 0.90,
				},
				Status: types.PoolStatusPending,
			},
			AgeString: "2h",
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputPoolList(entries, 0, 50, 1)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputPoolList: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "cand-test-001") {
		t.Errorf("expected JSON output with candidate ID, got: %s", out)
	}
}

func TestCov3_pool_outputPoolList_paginationMessage(t *testing.T) {
	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	// Temporarily widen to avoid truncation issues
	oldPoolWide := poolWide
	poolWide = true
	defer func() { poolWide = oldPoolWide }()

	entries := []pool.PoolEntry{
		{
			PoolEntry: types.PoolEntry{
				Candidate: types.Candidate{
					ID:         "cand-page-001",
					Tier:       types.TierBronze,
					Utility:    0.60,
					Confidence: 0.55,
				},
				Status: types.PoolStatusPending,
			},
			AgeString: "1h",
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// total > len(entries) triggers pagination message
	err := outputPoolList(entries, 0, 50, 10)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputPoolList: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "Showing") {
		t.Errorf("expected pagination message with 'Showing', got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// pool.go — outputPoolShow
// ---------------------------------------------------------------------------

func TestCov3_pool_outputPoolShow_textMode(t *testing.T) {
	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	entry := &pool.PoolEntry{
		PoolEntry: types.PoolEntry{
			Candidate: types.Candidate{
				ID:          "show-test-001",
				Type:        types.KnowledgeTypeLearning,
				Tier:        types.TierGold,
				Content:     "Test learning content for show",
				Utility:     0.85,
				Confidence:  0.92,
				Maturity:    types.MaturityEstablished,
				RewardCount: 5,
				Source: types.Source{
					SessionID:      "session-show-test",
					TranscriptPath: "/tmp/show.jsonl",
					MessageIndex:   10,
				},
			},
			ScoringResult: types.Scoring{
				RawScore: 0.87,
				Rubric: types.RubricScores{
					Specificity:   0.90,
					Actionability: 0.85,
					Novelty:       0.80,
					Context:       0.88,
					Confidence:    0.91,
				},
			},
			Status: types.PoolStatusStaged,
		},
		AgeString: "3d",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputPoolShow(entry)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputPoolShow: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	checks := []string{
		"show-test-001",
		"gold",
		"staged",
		"Test learning content for show",
		"MemRL Metrics:",
		"Provenance:",
	}
	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, out)
		}
	}
}

func TestCov3_pool_outputPoolShow_withHumanReview(t *testing.T) {
	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	reviewedAt := time.Now()
	entry := &pool.PoolEntry{
		PoolEntry: types.PoolEntry{
			Candidate: types.Candidate{
				ID:      "review-show-001",
				Tier:    types.TierBronze,
				Content: "Reviewed content",
				Source:  types.Source{SessionID: "s1", TranscriptPath: "/t.jsonl"},
			},
			Status: types.PoolStatusPending,
			HumanReview: &types.HumanReview{
				Reviewed:   true,
				Approved:   true,
				Reviewer:   "test-user",
				Notes:      "Looks good",
				ReviewedAt: reviewedAt,
			},
			ScoringResult: types.Scoring{},
		},
		AgeString: "1h",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputPoolShow(entry)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputPoolShow: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "Human Review:") {
		t.Errorf("expected 'Human Review:' section, got:\n%s", out)
	}
	if !strings.Contains(out, "test-user") {
		t.Errorf("expected reviewer name, got:\n%s", out)
	}
}
