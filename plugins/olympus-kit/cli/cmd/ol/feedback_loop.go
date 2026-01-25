package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/plugins/olympus-kit/cli/internal/ratchet"
	"github.com/boshu2/agentops/plugins/olympus-kit/cli/internal/types"
)

// FeedbackEvent records a feedback loop closure event.
type FeedbackEvent struct {
	SessionID      string    `json:"session_id"`
	ArtifactPath   string    `json:"artifact_path"`
	Reward         float64   `json:"reward"`
	UtilityBefore  float64   `json:"utility_before"`
	UtilityAfter   float64   `json:"utility_after"`
	Alpha          float64   `json:"alpha"`
	RecordedAt     time.Time `json:"recorded_at"`
	TranscriptPath string    `json:"transcript_path,omitempty"`
}

// FeedbackFilePath is the relative path to the feedback log.
const FeedbackFilePath = ".agents/olympus/feedback.jsonl"

var feedbackLoopCmd = &cobra.Command{
	Use:   "feedback-loop",
	Short: "Close the MemRL feedback loop for a session",
	Long: `Automatically close the MemRL feedback loop by updating utilities of cited learnings.

This command:
1. Reads citations for the session from .agents/olympus/citations.jsonl
2. Computes reward from session outcome (or uses --reward override)
3. Updates utility of each cited learning via EMA rule
4. Logs feedback events to .agents/olympus/feedback.jsonl

The feedback loop enables knowledge to compound:
- High-utility learnings surface more often
- Learnings that correlate with success get reinforced
- Learnings that don't help slowly decay

Examples:
  ol feedback-loop --session session-20260125-120000
  ol feedback-loop --session abc123 --reward 0.85
  ol feedback-loop --transcript ~/.claude/projects/*/abc.jsonl`,
	RunE: runFeedbackLoop,
}

var (
	feedbackLoopSessionID    string
	feedbackLoopReward       float64
	feedbackLoopTranscript   string
	feedbackLoopAlpha        float64
	feedbackLoopCitationType string
)

func init() {
	rootCmd.AddCommand(feedbackLoopCmd)
	feedbackLoopCmd.Flags().StringVar(&feedbackLoopSessionID, "session", "", "Session ID to process")
	feedbackLoopCmd.Flags().Float64Var(&feedbackLoopReward, "reward", -1, "Override reward value (0.0-1.0); -1 = compute from transcript")
	feedbackLoopCmd.Flags().StringVar(&feedbackLoopTranscript, "transcript", "", "Path to transcript for reward computation")
	feedbackLoopCmd.Flags().Float64Var(&feedbackLoopAlpha, "alpha", types.DefaultAlpha, "EMA learning rate")
	feedbackLoopCmd.Flags().StringVar(&feedbackLoopCitationType, "citation-type", "retrieved", "Filter citations by type (retrieved, applied, all)")
}

// loadSessionCitations loads and filters citations for a session.
func loadSessionCitations(cwd, sessionID, citationType string) ([]types.CitationEvent, error) {
	allCitations, err := ratchet.LoadCitations(cwd)
	if err != nil {
		return nil, fmt.Errorf("load citations: %w", err)
	}

	var sessionCitations []types.CitationEvent
	for _, c := range allCitations {
		if c.SessionID != sessionID {
			continue
		}
		if citationType != "all" && c.CitationType != citationType {
			continue
		}
		sessionCitations = append(sessionCitations, c)
	}
	return sessionCitations, nil
}

// computeRewardFromTranscript derives reward from transcript analysis.
func computeRewardFromTranscript(transcriptPath, sessionID string) (float64, error) {
	if transcriptPath == "" {
		homeDir, _ := os.UserHomeDir()
		transcriptsDir := filepath.Join(homeDir, ".claude", "projects")
		transcriptPath = findMostRecentTranscript(transcriptsDir)
	}
	if transcriptPath == "" {
		return 0, fmt.Errorf("no transcript found; use --reward to specify manually")
	}
	outcome, err := analyzeTranscript(transcriptPath, sessionID)
	if err != nil {
		return 0, fmt.Errorf("analyze transcript: %w", err)
	}
	VerbosePrintf("Computed reward %.2f from transcript %s\n", outcome.Reward, transcriptPath)
	return outcome.Reward, nil
}

// deduplicateCitations returns unique citations by artifact path.
func deduplicateCitations(citations []types.CitationEvent) []types.CitationEvent {
	seen := make(map[string]bool)
	var unique []types.CitationEvent
	for _, c := range citations {
		if !seen[c.ArtifactPath] {
			seen[c.ArtifactPath] = true
			unique = append(unique, c)
		}
	}
	return unique
}

// processUniqueCitations updates learning utilities and returns feedback events.
func processUniqueCitations(cwd, sessionID, transcriptPath string, citations []types.CitationEvent, reward, alpha float64) ([]FeedbackEvent, int, int) {
	var events []FeedbackEvent
	updatedCount, failedCount := 0, 0

	for _, citation := range citations {
		learningPath, err := findLearningFile(cwd, filepath.Base(citation.ArtifactPath))
		if err != nil {
			if _, statErr := os.Stat(citation.ArtifactPath); statErr == nil {
				learningPath = citation.ArtifactPath
			} else {
				VerbosePrintf("Warning: learning not found for %s: %v\n", citation.ArtifactPath, err)
				failedCount++
				continue
			}
		}

		oldUtility, newUtility, err := updateLearningUtility(learningPath, reward, alpha)
		if err != nil {
			VerbosePrintf("Warning: failed to update %s: %v\n", learningPath, err)
			failedCount++
			continue
		}

		event := FeedbackEvent{
			SessionID:      sessionID,
			ArtifactPath:   learningPath,
			Reward:         reward,
			UtilityBefore:  oldUtility,
			UtilityAfter:   newUtility,
			Alpha:          alpha,
			RecordedAt:     time.Now(),
			TranscriptPath: transcriptPath,
		}
		events = append(events, event)
		updatedCount++

		VerbosePrintf("Updated %s: %.3f → %.3f (reward=%.2f)\n",
			filepath.Base(learningPath), oldUtility, newUtility, reward)
	}

	return events, updatedCount, failedCount
}

func runFeedbackLoop(cmd *cobra.Command, args []string) error {
	if feedbackLoopSessionID == "" {
		return fmt.Errorf("--session is required")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	sessionID := canonicalSessionID(feedbackLoopSessionID)

	if GetDryRun() {
		fmt.Printf("[dry-run] Would process feedback loop for session: %s\n", sessionID)
		return nil
	}

	// Load and filter citations
	sessionCitations, err := loadSessionCitations(cwd, sessionID, feedbackLoopCitationType)
	if err != nil {
		return err
	}
	if len(sessionCitations) == 0 {
		fmt.Printf("No citations found for session %s\n", sessionID)
		return nil
	}

	// Determine reward
	reward := feedbackLoopReward
	if reward < 0 || reward > 1 {
		reward, err = computeRewardFromTranscript(feedbackLoopTranscript, sessionID)
		if err != nil {
			return err
		}
	}

	// Process citations
	uniqueCitations := deduplicateCitations(sessionCitations)
	feedbackEvents, updatedCount, failedCount := processUniqueCitations(
		cwd, sessionID, feedbackLoopTranscript, uniqueCitations, reward, feedbackLoopAlpha,
	)

	// Write feedback events to log
	if err := writeFeedbackEvents(cwd, feedbackEvents); err != nil {
		VerbosePrintf("Warning: failed to write feedback log: %v\n", err)
	}

	// Output summary
	return outputFeedbackSummary(sessionID, reward, len(sessionCitations), len(uniqueCitations), updatedCount, failedCount, feedbackEvents)
}

// outputFeedbackSummary outputs the feedback loop results.
func outputFeedbackSummary(sessionID string, reward float64, totalCitations, uniqueCount, updatedCount, failedCount int, events []FeedbackEvent) error {
	switch GetOutput() {
	case "json":
		result := map[string]interface{}{
			"session_id": sessionID,
			"reward":     reward,
			"citations":  totalCitations,
			"unique":     uniqueCount,
			"updated":    updatedCount,
			"failed":     failedCount,
			"feedback":   events,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)

	default:
		fmt.Printf("Feedback Loop Complete\n")
		fmt.Printf("======================\n")
		fmt.Printf("Session:     %s\n", sessionID)
		fmt.Printf("Reward:      %.2f\n", reward)
		fmt.Printf("Citations:   %d (%d unique)\n", totalCitations, uniqueCount)
		fmt.Printf("Updated:     %d\n", updatedCount)
		if failedCount > 0 {
			fmt.Printf("Failed:      %d\n", failedCount)
		}
	}

	return nil
}

// writeFeedbackEvents appends feedback events to the feedback log.
func writeFeedbackEvents(baseDir string, events []FeedbackEvent) error {
	if len(events) == 0 {
		return nil
	}

	feedbackPath := filepath.Join(baseDir, FeedbackFilePath)

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(feedbackPath), 0755); err != nil {
		return fmt.Errorf("create feedback directory: %w", err)
	}

	// Open for append
	f, err := os.OpenFile(feedbackPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open feedback file: %w", err)
	}
	defer f.Close() //nolint:errcheck // write-only file, Close error non-actionable

	// Write each event as JSONL
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}
		if _, err := f.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("write feedback event: %w", err)
		}
	}

	return nil
}

// batchFeedbackCmd processes feedback for multiple sessions.
var batchFeedbackCmd = &cobra.Command{
	Use:   "batch-feedback",
	Short: "Process feedback loop for all recent sessions",
	Long: `Process feedback loop for all sessions that have citations but no feedback.

Scans .agents/olympus/citations.jsonl for sessions without corresponding
entries in .agents/olympus/feedback.jsonl and processes them.

Examples:
  ol batch-feedback
  ol batch-feedback --days 7
  ol batch-feedback --dry-run`,
	RunE: runBatchFeedback,
}

var batchFeedbackDays int

func init() {
	rootCmd.AddCommand(batchFeedbackCmd)
	batchFeedbackCmd.Flags().IntVar(&batchFeedbackDays, "days", 7, "Process sessions from the last N days")
}

func runBatchFeedback(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Load all citations
	citations, err := ratchet.LoadCitations(cwd)
	if err != nil {
		return fmt.Errorf("load citations: %w", err)
	}

	// Load existing feedback
	existingFeedback, err := loadFeedbackEvents(cwd)
	if err != nil && !os.IsNotExist(err) {
		VerbosePrintf("Warning: failed to load feedback: %v\n", err)
	}

	// Build set of sessions that already have feedback
	processedSessions := make(map[string]bool)
	for _, f := range existingFeedback {
		processedSessions[f.SessionID] = true
	}

	// Find sessions with citations but no feedback
	since := time.Now().AddDate(0, 0, -batchFeedbackDays)
	sessionCitations := make(map[string][]types.CitationEvent)

	for _, c := range citations {
		if c.CitedAt.Before(since) {
			continue
		}
		if processedSessions[c.SessionID] {
			continue
		}
		sessionCitations[c.SessionID] = append(sessionCitations[c.SessionID], c)
	}

	if len(sessionCitations) == 0 {
		fmt.Println("No unprocessed sessions found")
		return nil
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would process %d sessions:\n", len(sessionCitations))
		for sessionID, citations := range sessionCitations {
			fmt.Printf("  - %s (%d citations)\n", sessionID, len(citations))
		}
		return nil
	}

	// Process each session
	processed := 0
	for sessionID := range sessionCitations {
		// Set flags and run feedback loop
		feedbackLoopSessionID = sessionID
		feedbackLoopReward = -1 // Compute from transcript

		fmt.Printf("Processing session %s...\n", sessionID)
		if err := runFeedbackLoop(cmd, nil); err != nil {
			VerbosePrintf("Warning: failed to process %s: %v\n", sessionID, err)
			continue
		}
		processed++
	}

	fmt.Printf("\nProcessed %d/%d sessions\n", processed, len(sessionCitations))
	return nil
}

// loadFeedbackEvents reads all feedback events from the log.
func loadFeedbackEvents(baseDir string) ([]FeedbackEvent, error) {
	feedbackPath := filepath.Join(baseDir, FeedbackFilePath)

	f, err := os.Open(feedbackPath)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck // read-only file, Close error non-actionable

	var events []FeedbackEvent
	decoder := json.NewDecoder(f)
	for decoder.More() {
		var event FeedbackEvent
		if err := decoder.Decode(&event); err != nil {
			continue // Skip malformed lines
		}
		events = append(events, event)
	}

	return events, nil
}
