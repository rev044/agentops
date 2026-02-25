package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boshu2/agentops/cli/internal/ratchet"
	"github.com/boshu2/agentops/cli/internal/resolver"
	"github.com/boshu2/agentops/cli/internal/types"
)

// processCitationFeedback reads unprocessed citations from .agents/ao/citations.jsonl,
// applies positive MemRL feedback for each cited learning, and marks them as processed.
// Returns (total processed, rewarded count, skipped count).
func processCitationFeedback(cwd string) (int, int, int) {
	citationsPath := filepath.Join(cwd, ratchet.CitationsFilePath)
	citations, err := ratchet.LoadCitations(cwd)
	if err != nil || len(citations) == 0 {
		return 0, 0, 0
	}

	// Deduplicate: one feedback event per artifact (most recent citation wins)
	seen := make(map[string]bool)
	var unique []types.CitationEvent
	for i := len(citations) - 1; i >= 0; i-- {
		c := citations[i]
		if c.FeedbackGiven {
			continue // Already processed
		}
		key := c.ArtifactPath
		if seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, c)
	}

	if len(unique) == 0 {
		return 0, 0, 0
	}

	// Compute adaptive reward from most recent transcript
	reward, err := computeSessionRewardForCloseLoop(cwd)
	if err != nil {
		reward = types.InitialUtility // Fallback to 0.5 (neutral), NOT 1.0
	}

	res := resolver.NewFileResolver(cwd)
	var rewarded, skipped int
	sessionID := canonicalSessionID("")
	var feedbackEvents []FeedbackEvent

	for _, c := range unique {
		// Resolve the artifact to a file path
		learningID := extractLearningID(c.ArtifactPath)
		path, err := res.Resolve(learningID)
		if err != nil {
			skipped++
			continue
		}

		// Apply adaptive reward (transcript-derived or fallback)
		oldUtility, newUtility, err := updateLearningUtility(path, reward, types.DefaultAlpha)
		if err != nil {
			skipped++
			continue
		}

		feedbackEvents = append(feedbackEvents, FeedbackEvent{
			SessionID:     sessionID,
			ArtifactPath:  path,
			Reward:        reward,
			UtilityBefore: oldUtility,
			UtilityAfter:  newUtility,
			Alpha:         types.DefaultAlpha,
			RecordedAt:    time.Now(),
		})
		rewarded++
	}

	// Write audit trail
	if len(feedbackEvents) > 0 {
		_ = writeFeedbackEvents(cwd, feedbackEvents)
	}

	// Mark all citations as feedback-given by rewriting the file
	markCitationsFeedbackGiven(citationsPath, citations)

	return len(unique), rewarded, skipped
}

// extractLearningID derives a learning ID from an artifact path.
// Handles both relative (".agents/learnings/abc.md") and absolute
// ("/home/user/repo/.agents/learnings/abc.md") paths.
func extractLearningID(artifactPath string) string {
	for _, marker := range []string{"/.agents/learnings/", "/.agents/patterns/", ".agents/learnings/", ".agents/patterns/"} {
		if idx := strings.Index(artifactPath, marker); idx >= 0 {
			return artifactPath[idx+len(marker):]
		}
	}
	return filepath.Base(artifactPath)
}

// markCitationsFeedbackGiven rewrites citations.jsonl with FeedbackGiven=true for all entries.
func markCitationsFeedbackGiven(citationsPath string, citations []types.CitationEvent) {
	if GetDryRun() {
		return
	}

	var lines []string
	for _, c := range citations {
		c.FeedbackGiven = true
		data, err := json.Marshal(c)
		if err != nil {
			continue
		}
		lines = append(lines, string(data))
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(citationsPath, []byte(content), 0600); err != nil {
		VerbosePrintf("Warning: failed to write citations feedback: %v\n", err)
	}
}

// computeSessionRewardForCloseLoop finds the most recent transcript and computes reward.
func computeSessionRewardForCloseLoop(cwd string) (float64, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, fmt.Errorf("get home directory: %w", err)
	}
	transcriptsDir := filepath.Join(homeDir, ".claude", "projects")
	transcriptPath := findMostRecentTranscript(transcriptsDir)
	if transcriptPath == "" {
		return 0, fmt.Errorf("no transcript found")
	}
	outcome, err := analyzeTranscript(transcriptPath, "")
	if err != nil {
		return 0, fmt.Errorf("analyze transcript: %w", err)
	}
	return outcome.Reward, nil
}
