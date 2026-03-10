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
		if isFindingArtifactPath(cwd, c.ArtifactPath) {
			path := normalizeArtifactPath(cwd, c.ArtifactPath)
			citedAt := c.CitedAt
			if citedAt.IsZero() {
				citedAt = time.Now()
			}
			if err := updateFindingCitationFields(path, citedAt); err != nil {
				skipped++
				continue
			}
			rewarded++
			continue
		}

		// Resolve the artifact to a file path
		learningID := extractLearningID(c.ArtifactPath)
		path, err := res.Resolve(learningID)
		if err != nil {
			skipped++
			continue
		}

		// Compute annealed alpha based on citation count
		rewardCount := getLearningRewardCount(path)
		alpha := annealedAlpha(types.DefaultAlpha, rewardCount)

		// Apply adaptive reward (transcript-derived or fallback)
		oldUtility, newUtility, err := updateLearningUtility(path, reward, alpha)
		if err != nil {
			skipped++
			continue
		}

		// Upgrade citation_type from "retrieved" to "applied" on positive feedback
		if reward >= 0.6 {
			upgradeCitationType(citations, c.ArtifactPath)
		}

		feedbackEvents = append(feedbackEvents, FeedbackEvent{
			SessionID:     sessionID,
			ArtifactPath:  path,
			Reward:        reward,
			UtilityBefore: oldUtility,
			UtilityAfter:  newUtility,
			Alpha:         alpha,
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

func updateFindingCitationFields(path string, citedAt time.Time) error {
	finding, err := parseFindingFile(path)
	if err != nil {
		return err
	}
	hitCount := finding.HitCount + 1
	return updateFindingFrontMatter(path, map[string]string{
		"hit_count":  fmt.Sprintf("%d", hitCount),
		"last_cited": citedAt.UTC().Format(time.RFC3339),
	})
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

// upgradeCitationType marks citations for the given artifact as "applied"
// when the session provided positive feedback (reward >= 0.6).
func upgradeCitationType(citations []types.CitationEvent, artifactPath string) {
	for i := range citations {
		if citations[i].ArtifactPath == artifactPath && citations[i].CitationType == "retrieved" {
			citations[i].CitationType = "applied"
		}
	}
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

// computeSessionRewardForCloseLoop checks for a binary session outcome file first,
// then falls back to transcript analysis.
func computeSessionRewardForCloseLoop(cwd string) (float64, error) {
	// Try session outcome file first (binary, reliable)
	outcomePath := filepath.Join(cwd, ".agents", "ao", "last-session-outcome.json")
	if data, err := os.ReadFile(outcomePath); err == nil {
		var outcome struct {
			Outcome string `json:"outcome"`
		}
		if json.Unmarshal(data, &outcome) == nil {
			switch outcome.Outcome {
			case "success":
				return 0.8, nil
			case "failure":
				return 0.2, nil
			case "abandoned":
				return 0.4, nil
			}
		}
	}
	// Fallback to transcript analysis (existing behavior)
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		return types.InitialUtility, nil
	}
	transcriptsDir := filepath.Join(homeDir, ".claude", "projects")
	transcriptPath := findMostRecentTranscript(transcriptsDir)
	if transcriptPath == "" {
		return types.InitialUtility, nil
	}
	outcome, err := analyzeTranscript(transcriptPath, "")
	if err != nil {
		return types.InitialUtility, nil
	}
	return outcome.Reward, nil
}

// promoteCitedLearnings reads the feedback log and attempts maturity promotion
// on each learning that received citation feedback. This ensures learnings whose
// utility was just bumped by citation feedback get promoted in the same close-loop
// cycle rather than waiting for the next run.
// Returns the number of learnings that transitioned.
func promoteCitedLearnings(cwd string, quiet bool) int {
	if GetDryRun() {
		return 0
	}

	feedbackPath := filepath.Join(cwd, FeedbackFilePath)
	data, err := os.ReadFile(feedbackPath)
	if err != nil {
		return 0
	}

	// Collect unique artifact paths from recent feedback events
	seen := make(map[string]bool)
	var paths []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var evt FeedbackEvent
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			continue
		}
		if evt.ArtifactPath == "" || seen[evt.ArtifactPath] {
			continue
		}
		seen[evt.ArtifactPath] = true
		paths = append(paths, evt.ArtifactPath)
	}

	promoted := 0
	for _, p := range paths {
		result, err := ratchet.ApplyMaturityTransition(p)
		if err != nil {
			continue
		}
		if result.Transitioned {
			promoted++
			if !quiet {
				fmt.Fprintf(os.Stderr, "  maturity: %s → %s (%s)\n", result.OldMaturity, result.NewMaturity, filepath.Base(p))
			}
		}
	}
	if promoted > 0 && !quiet {
		fmt.Fprintf(os.Stderr, "Promoted %d learnings\n", promoted)
	}
	return promoted
}
