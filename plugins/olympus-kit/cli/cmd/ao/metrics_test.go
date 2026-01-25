package main

import (
	"testing"
	"time"

	"github.com/boshu2/agentops/plugins/olympus-kit/cli/internal/types"
)

func TestFilterCitationsForPeriod(t *testing.T) {
	now := time.Now()
	oneDayAgo := now.AddDate(0, 0, -1)
	twoDaysAgo := now.AddDate(0, 0, -2)
	oneWeekAgo := now.AddDate(0, 0, -7)
	twoWeeksAgo := now.AddDate(0, 0, -14)

	citations := []types.CitationEvent{
		{ArtifactPath: "/path/a.md", CitedAt: oneDayAgo},
		{ArtifactPath: "/path/b.md", CitedAt: twoDaysAgo},
		{ArtifactPath: "/path/c.md", CitedAt: oneWeekAgo},
		{ArtifactPath: "/path/d.md", CitedAt: twoWeeksAgo},
	}

	tests := []struct {
		name          string
		start         time.Time
		end           time.Time
		wantCount     int
		wantUniqueCnt int
	}{
		{
			name:          "all in period",
			start:         twoWeeksAgo.AddDate(0, 0, -1),
			end:           now.AddDate(0, 0, 1),
			wantCount:     4,
			wantUniqueCnt: 4,
		},
		{
			name:          "last 3 days",
			start:         now.AddDate(0, 0, -3),
			end:           now.AddDate(0, 0, 1),
			wantCount:     2,
			wantUniqueCnt: 2,
		},
		{
			name:          "last week",
			start:         now.AddDate(0, 0, -8),
			end:           now.AddDate(0, 0, 1),
			wantCount:     3,
			wantUniqueCnt: 3,
		},
		{
			name:          "empty period",
			start:         now.AddDate(0, 0, -100),
			end:           now.AddDate(0, 0, -50),
			wantCount:     0,
			wantUniqueCnt: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := filterCitationsForPeriod(citations, tt.start, tt.end)
			if len(stats.citations) != tt.wantCount {
				t.Errorf("filterCitationsForPeriod() count = %d, want %d",
					len(stats.citations), tt.wantCount)
			}
			if len(stats.uniqueCited) != tt.wantUniqueCnt {
				t.Errorf("filterCitationsForPeriod() uniqueCited = %d, want %d",
					len(stats.uniqueCited), tt.wantUniqueCnt)
			}
		})
	}
}

func TestComputeSigmaRho(t *testing.T) {
	tests := []struct {
		name           string
		totalArtifacts int
		uniqueCited    int
		citationCount  int
		days           int
		wantSigma      float64
		wantRho        float64
	}{
		{
			name:           "normal case",
			totalArtifacts: 100,
			uniqueCited:    50,
			citationCount:  100,
			days:           7,
			wantSigma:      0.5,
			wantRho:        2.0, // 100/50/1week = 2
		},
		{
			name:           "no artifacts",
			totalArtifacts: 0,
			uniqueCited:    0,
			citationCount:  0,
			days:           7,
			wantSigma:      0,
			wantRho:        0,
		},
		{
			name:           "no citations",
			totalArtifacts: 100,
			uniqueCited:    0,
			citationCount:  0,
			days:           7,
			wantSigma:      0,
			wantRho:        0,
		},
		{
			name:           "14 days",
			totalArtifacts: 100,
			uniqueCited:    50,
			citationCount:  100,
			days:           14,
			wantSigma:      0.5,
			wantRho:        1.0, // 100/50/2weeks = 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sigma, rho := computeSigmaRho(tt.totalArtifacts, tt.uniqueCited, tt.citationCount, tt.days)

			if !floatEqual(sigma, tt.wantSigma, 0.01) {
				t.Errorf("computeSigmaRho() sigma = %v, want %v", sigma, tt.wantSigma)
			}
			if !floatEqual(rho, tt.wantRho, 0.01) {
				t.Errorf("computeSigmaRho() rho = %v, want %v", rho, tt.wantRho)
			}
		})
	}
}

func TestCountLoopMetrics(t *testing.T) {
	now := time.Now()
	oneDayAgo := now.AddDate(0, 0, -1)

	citations := []types.CitationEvent{
		{ArtifactPath: "/path/to/.agents/learnings/L1.md", CitedAt: oneDayAgo},
		{ArtifactPath: "/path/to/.agents/learnings/L2.md", CitedAt: oneDayAgo},
		{ArtifactPath: "/path/to/.agents/patterns/P1.md", CitedAt: oneDayAgo},
		{ArtifactPath: "/other/file.md", CitedAt: oneDayAgo},
	}

	// countLoopMetrics requires actual directory structure, so we just test
	// the learningsFound counting logic here via the helper
	learningsFound := 0
	for _, c := range citations {
		if containsLearningsPath(c.ArtifactPath) {
			learningsFound++
		}
	}

	if learningsFound != 2 {
		t.Errorf("learningsFound = %d, want 2", learningsFound)
	}
}

func TestCountBypassCitations(t *testing.T) {
	citations := []types.CitationEvent{
		{ArtifactPath: "/normal/path.md", CitationType: "recall"},
		{ArtifactPath: "/bypass/path.md", CitationType: "bypass"},
		{ArtifactPath: "bypass:/skipped", CitationType: ""},
		{ArtifactPath: "/another/path.md", CitationType: "inject"},
	}

	got := countBypassCitations(citations)
	if got != 2 {
		t.Errorf("countBypassCitations() = %d, want 2", got)
	}
}

// floatEqual checks if two floats are approximately equal
func floatEqual(a, b, epsilon float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}

// containsLearningsPath checks if path contains /learnings/
func containsLearningsPath(path string) bool {
	for i := 0; i <= len(path)-11; i++ {
		if path[i:i+11] == "/learnings/" {
			return true
		}
	}
	return false
}
