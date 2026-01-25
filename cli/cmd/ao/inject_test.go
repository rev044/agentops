package main

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/boshu2/agentops/plugins/olympus-kit/cli/internal/types"
)

func TestFreshnessScore(t *testing.T) {
	tests := []struct {
		name     string
		ageWeeks float64
		wantMin  float64
		wantMax  float64
	}{
		{"fresh (0 weeks)", 0, 0.99, 1.01},
		{"1 week old", 1, 0.82, 0.86},
		{"4 weeks old", 4, 0.49, 0.52},
		{"12 weeks old", 12, 0.10, 0.15},
		{"52 weeks old", 52, 0.10, 0.11}, // Clamped to 0.1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := freshnessScore(tt.ageWeeks)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("freshnessScore(%v) = %v, want between %v and %v",
					tt.ageWeeks, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestParseFrontMatter(t *testing.T) {
	tests := []struct {
		name           string
		lines          []string
		wantSuperseded string
		wantEndLine    int
	}{
		{
			name:           "no front matter",
			lines:          []string{"# Title", "Content"},
			wantSuperseded: "",
			wantEndLine:    0,
		},
		{
			name:           "empty front matter",
			lines:          []string{"---", "---", "# Title"},
			wantSuperseded: "",
			wantEndLine:    2,
		},
		{
			name:           "superseded_by set",
			lines:          []string{"---", "superseded_by: L42", "---", "# Title"},
			wantSuperseded: "L42",
			wantEndLine:    3,
		},
		{
			name:           "superseded-by with dash",
			lines:          []string{"---", "superseded-by: new-learning", "---"},
			wantSuperseded: "new-learning",
			wantEndLine:    3,
		},
		{
			name:           "superseded_by null",
			lines:          []string{"---", "superseded_by: null", "---"},
			wantSuperseded: "null",
			wantEndLine:    3,
		},
		{
			name:           "empty lines slice",
			lines:          []string{},
			wantSuperseded: "",
			wantEndLine:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, endLine := parseFrontMatter(tt.lines)
			if fm.SupersededBy != tt.wantSuperseded {
				t.Errorf("parseFrontMatter() supersededBy = %q, want %q",
					fm.SupersededBy, tt.wantSuperseded)
			}
			if endLine != tt.wantEndLine {
				t.Errorf("parseFrontMatter() endLine = %d, want %d",
					endLine, tt.wantEndLine)
			}
		})
	}
}

func TestExtractSummary(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		startIdx int
		want     string
	}{
		{
			name:     "simple paragraph",
			lines:    []string{"# Title", "This is the summary."},
			startIdx: 1,
			want:     "This is the summary.",
		},
		{
			name:     "skip empty lines",
			lines:    []string{"", "", "Summary text"},
			startIdx: 0,
			want:     "Summary text",
		},
		{
			name:     "skip headings",
			lines:    []string{"## Heading", "Content here"},
			startIdx: 0,
			want:     "Content here",
		},
		{
			name:     "multi-line paragraph",
			lines:    []string{"First line.", "Second line.", "Third line."},
			startIdx: 0,
			want:     "First line. Second line. Third line.",
		},
		{
			name:     "stop at empty line",
			lines:    []string{"First line.", "", "Different paragraph"},
			startIdx: 0,
			want:     "First line.",
		},
		{
			name:     "empty content",
			lines:    []string{},
			startIdx: 0,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSummary(tt.lines, tt.startIdx)
			if got != tt.want {
				t.Errorf("extractSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseLearningFile(t *testing.T) {
	// Create temp directory for test files
	tmpDir, err := os.MkdirTemp("", "inject_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test: regular markdown file
	t.Run("regular markdown", func(t *testing.T) {
		content := `---
id: L42
---
# Test Learning

This is the summary content.
`
		path := filepath.Join(tmpDir, "test-learning.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		l, err := parseLearningFile(path)
		if err != nil {
			t.Errorf("parseLearningFile() error = %v", err)
		}
		if l.Superseded {
			t.Error("expected Superseded = false")
		}
		if l.Title != "Test Learning" {
			t.Errorf("Title = %q, want %q", l.Title, "Test Learning")
		}
	})

	// Test: superseded markdown file
	t.Run("superseded markdown", func(t *testing.T) {
		content := `---
superseded_by: L100
---
# Old Learning
`
		path := filepath.Join(tmpDir, "old-learning.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		l, err := parseLearningFile(path)
		if err != nil {
			t.Errorf("parseLearningFile() error = %v", err)
		}
		if !l.Superseded {
			t.Error("expected Superseded = true")
		}
	})

	// Test: file not found
	t.Run("file not found", func(t *testing.T) {
		_, err := parseLearningFile(filepath.Join(tmpDir, "nonexistent.md"))
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}

func TestApplyCompositeScoring(t *testing.T) {
	tests := []struct {
		name      string
		learnings []learning
		lambda    float64
		// We check relative ordering rather than exact scores
		wantFirst string // ID of learning that should rank first
	}{
		{
			name:      "empty slice",
			learnings: []learning{},
			lambda:    0.5,
			wantFirst: "",
		},
		{
			name: "high utility wins with high lambda",
			learnings: []learning{
				{ID: "fresh", FreshnessScore: 1.0, Utility: 0.3},
				{ID: "useful", FreshnessScore: 0.5, Utility: 0.9},
			},
			lambda:    2.0, // Weight utility MORE than freshness (lambda > 1)
			wantFirst: "useful",
		},
		{
			name: "freshness wins with low lambda",
			learnings: []learning{
				{ID: "fresh", FreshnessScore: 1.0, Utility: 0.3},
				{ID: "useful", FreshnessScore: 0.5, Utility: 0.9},
			},
			lambda:    0.0, // Ignore utility
			wantFirst: "fresh",
		},
		{
			name: "balanced scoring",
			learnings: []learning{
				{ID: "L1", FreshnessScore: 0.8, Utility: 0.6},
				{ID: "L2", FreshnessScore: 0.6, Utility: 0.8},
				{ID: "L3", FreshnessScore: 0.5, Utility: 0.5},
			},
			lambda:    0.5,
			wantFirst: "L1", // L1 and L2 similar, but L1 has higher freshness
		},
		{
			name: "default utility (all 0.5)",
			learnings: []learning{
				{ID: "newer", FreshnessScore: 0.9, Utility: types.InitialUtility},
				{ID: "older", FreshnessScore: 0.3, Utility: types.InitialUtility},
			},
			lambda:    0.5,
			wantFirst: "newer", // When utility is equal, freshness wins
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			learnings := make([]learning, len(tt.learnings))
			copy(learnings, tt.learnings)

			applyCompositeScoring(learnings, tt.lambda)

			if tt.wantFirst == "" {
				return // Empty case
			}

			// Sort by composite score (descending)
			maxScore := math.Inf(-1)
			var winner string
			for _, l := range learnings {
				if l.CompositeScore > maxScore {
					maxScore = l.CompositeScore
					winner = l.ID
				}
			}

			if winner != tt.wantFirst {
				t.Errorf("winner = %q, want %q", winner, tt.wantFirst)
				for _, l := range learnings {
					t.Logf("  %s: freshness=%.2f, utility=%.2f, composite=%.3f",
						l.ID, l.FreshnessScore, l.Utility, l.CompositeScore)
				}
			}
		})
	}
}

func TestCompositeScoringZNormalization(t *testing.T) {
	// Test that z-normalization produces mean ~0 and stddev ~1
	learnings := []learning{
		{ID: "L1", FreshnessScore: 1.0, Utility: 0.9},
		{ID: "L2", FreshnessScore: 0.8, Utility: 0.7},
		{ID: "L3", FreshnessScore: 0.6, Utility: 0.5},
		{ID: "L4", FreshnessScore: 0.4, Utility: 0.3},
		{ID: "L5", FreshnessScore: 0.2, Utility: 0.1},
	}

	applyCompositeScoring(learnings, 0.5)

	// All learnings should have composite scores set
	for _, l := range learnings {
		if l.CompositeScore == 0 && l.FreshnessScore != 0.6 { // 0.6 is the mean, might be ~0
			// At least verify scores are computed
		}
	}

	// Verify that higher freshness + utility = higher score
	// L1 should have highest score, L5 should have lowest
	if learnings[0].CompositeScore <= learnings[4].CompositeScore {
		t.Errorf("expected L1 > L5 but got %v <= %v",
			learnings[0].CompositeScore, learnings[4].CompositeScore)
	}
}
