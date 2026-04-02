package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// precisionAtK returns the fraction of top-K results whose IDs are in the expected set.
func precisionAtK(results []learning, expected map[string]bool, k int) float64 {
	if k <= 0 || len(results) == 0 {
		return 0
	}
	n := k
	if n > len(results) {
		n = len(results)
	}
	hits := 0
	for _, r := range results[:n] {
		if expected[r.ID] {
			hits++
		}
	}
	return float64(hits) / float64(k)
}

// meanReciprocalRank returns 1/rank of the first result matching expectedID, or 0 if not found.
func meanReciprocalRank(results []learning, expectedID string) float64 {
	for i, r := range results {
		if r.ID == expectedID {
			return 1.0 / float64(i+1)
		}
	}
	return 0
}

// seedCorpus copies test learning files from testdata/retrieval-bench/ into a temp dir
// structured as .agents/learnings/ for collectLearnings to discover.
func seedCorpus(t *testing.T) string {
	t.Helper()
	corpusDir := filepath.Join("testdata", "retrieval-bench")
	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Fatalf("reading corpus dir: %v", err)
	}

	tmpDir := t.TempDir()
	learningsDir := filepath.Join(tmpDir, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatalf("creating learnings dir: %v", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(corpusDir, e.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", e.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(learningsDir, e.Name()), data, 0o644); err != nil {
			t.Fatalf("writing %s: %v", e.Name(), err)
		}
	}
	return tmpDir
}

func TestRetrievalBench_PrecisionAtK(t *testing.T) {
	dir := seedCorpus(t)

	tests := []struct {
		query    string
		expected map[string]bool // IDs expected in top 3
		mustMiss map[string]bool // IDs that must NOT appear in top 3
	}{
		{
			query:    "CI pipeline",
			expected: map[string]bool{"ci-1.md": true, "ci-2.md": true, "ci-3.md": true, "db-1.md": true}, // db-1 mentions CI pipeline legitimately
			mustMiss: map[string]bool{"db-2.md": true, "db-3.md": true},
		},
		{
			query:    "session intelligence",
			expected: map[string]bool{"si-1.md": true, "si-2.md": true, "si-3.md": true},
			mustMiss: map[string]bool{"hook-1.md": true, "db-1.md": true},
		},
		{
			query:    "hook authoring",
			expected: map[string]bool{"hook-1.md": true, "hook-2.md": true, "hook-3.md": true},
			mustMiss: map[string]bool{"db-1.md": true, "si-1.md": true},
		},
		{
			query:    "database",
			expected: map[string]bool{"db-1.md": true, "db-2.md": true, "db-3.md": true},
			mustMiss: map[string]bool{"ci-1.md": true, "si-1.md": true},
		},
		{
			query:    "swarm",
			expected: map[string]bool{"swarm-1.md": true, "swarm-2.md": true},
			mustMiss: map[string]bool{"db-1": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results, err := collectLearnings(dir, tt.query, 10, "", 0)
			if err != nil {
				t.Fatalf("collectLearnings(%q): %v", tt.query, err)
			}
			if len(results) == 0 {
				t.Fatalf("collectLearnings(%q) returned 0 results", tt.query)
			}

			k := 3
			if len(tt.expected) < k {
				k = len(tt.expected)
			}
			p := precisionAtK(results, tt.expected, k)
			if p < 0.67 {
				ids := make([]string, 0, len(results))
				for _, r := range results {
					ids = append(ids, r.ID)
				}
				t.Errorf("P@%d = %.2f (want >= 0.67) for query %q; got IDs: %v", k, p, tt.query, ids)
			}

			// Verify must-miss items are not in top 3
			top := k
			if top > len(results) {
				top = len(results)
			}
			for _, r := range results[:top] {
				if tt.mustMiss[r.ID] {
					t.Errorf("query %q: unwanted ID %q found in top %d results", tt.query, r.ID, k)
				}
			}
		})
	}
}

func TestRetrievalBench_MRR(t *testing.T) {
	dir := seedCorpus(t)

	tests := []struct {
		query    string
		bestID   string  // the single best expected result
		minMRR   float64 // minimum acceptable MRR (1/rank)
	}{
		{"CI pipeline", "ci-1.md", 0.5},            // ci-1 is established + highest utility
		{"session intelligence", "si-1.md", 0.5},    // si-1 is candidate + highest utility
		{"hook authoring", "hook-1.md", 0.5},        // hook-1 is established + highest utility
		{"database", "db-1.md", 0.5},                // db-1 has highest utility in the db set
		{"swarm", "swarm-1.md", 0.5},                // swarm-1 is candidate + higher utility
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results, err := collectLearnings(dir, tt.query, 10, "", 0)
			if err != nil {
				t.Fatalf("collectLearnings(%q): %v", tt.query, err)
			}
			if len(results) == 0 {
				t.Fatalf("collectLearnings(%q) returned 0 results", tt.query)
			}

			mrr := meanReciprocalRank(results, tt.bestID)
			if mrr < tt.minMRR {
				ids := make([]string, 0, len(results))
				for _, r := range results {
					ids = append(ids, r.ID)
				}
				t.Errorf("MRR = %.2f (want >= %.2f) for query %q bestID=%q; got IDs: %v",
					mrr, tt.minMRR, tt.query, tt.bestID, ids)
			}
		})
	}
}

func TestRetrievalBench_FreshnessVsUtility(t *testing.T) {
	dir := t.TempDir()
	learningsDir := filepath.Join(dir, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Fresh learning with decent utility
	fresh := `---
type: learning
maturity: provisional
utility: 0.6
---
# Fresh Auth Pattern

A fresh learning about authentication patterns.
`
	// Old learning with high utility
	old := `---
type: learning
maturity: provisional
utility: 0.9
---
# Old Auth Pattern

An older learning about authentication patterns.
`

	os.WriteFile(filepath.Join(learningsDir, "fresh.md"), []byte(fresh), 0o644)
	os.WriteFile(filepath.Join(learningsDir, "old.md"), []byte(old), 0o644)

	// Make old.md appear old by backdating its mtime
	oldTime := filepath.Join(learningsDir, "old.md")
	oldMtime := mustParseTime(t, "2025-01-01T00:00:00Z")
	os.Chtimes(oldTime, oldMtime, oldMtime)

	results, err := collectLearnings(dir, "auth", 10, "", 0)
	if err != nil {
		t.Fatalf("collectLearnings: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Fresh learning should rank higher due to significant freshness advantage
	if results[0].ID != "fresh.md" {
		t.Errorf("expected fresh learning to rank first (freshness advantage), got %q with score %.3f vs fresh score %.3f",
			results[0].ID, results[0].CompositeScore, results[1].CompositeScore)
	}
}

func TestRetrievalBench_MaturityBoost(t *testing.T) {
	dir := t.TempDir()
	learningsDir := filepath.Join(dir, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	established := `---
type: learning
maturity: established
utility: 0.6
---
# Established Config Pattern

An established learning about config management patterns.
`
	provisional := `---
type: learning
maturity: provisional
utility: 0.7
---
# Provisional Config Pattern

A provisional learning about config management patterns.
`

	os.WriteFile(filepath.Join(learningsDir, "established.md"), []byte(established), 0o644)
	os.WriteFile(filepath.Join(learningsDir, "provisional.md"), []byte(provisional), 0o644)

	results, err := collectLearnings(dir, "config", 10, "", 0)
	if err != nil {
		t.Fatalf("collectLearnings: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Established (1.3x multiplier) with utility=0.6 should beat provisional (1.0x) with utility=0.7
	if results[0].ID != "established.md" {
		t.Errorf("expected established learning to rank first (maturity boost), got %q", results[0].ID)
	}
}

func TestRetrievalBench_GlobalLocalRanking(t *testing.T) {
	localDir := t.TempDir()
	globalDir := t.TempDir()

	localLearnings := filepath.Join(localDir, ".agents", "learnings")
	os.MkdirAll(localLearnings, 0o755)
	globalLearnings := filepath.Join(globalDir, "learnings")
	os.MkdirAll(globalLearnings, 0o755)

	content := `---
type: learning
maturity: candidate
utility: 0.7
---
# Deploy Pattern

A learning about deployment patterns.
`
	os.WriteFile(filepath.Join(localLearnings, "local-deploy.md"), []byte(content), 0o644)
	os.WriteFile(filepath.Join(globalLearnings, "global-deploy.md"), []byte(content), 0o644)

	results, err := collectLearnings(localDir, "deploy", 10, globalDir, 0.8)
	if err != nil {
		t.Fatalf("collectLearnings: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Local should rank above global due to 0.8 penalty on global
	if results[0].Global {
		t.Errorf("expected local learning to rank first, but got global (local score=%.3f, global score=%.3f)",
			results[1].CompositeScore, results[0].CompositeScore)
	}
}

func mustParseTime(t *testing.T, s string) (tm time.Time) {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse time %q: %v", s, err)
	}
	return tm
}
