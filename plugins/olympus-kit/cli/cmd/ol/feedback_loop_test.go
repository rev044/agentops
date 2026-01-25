package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boshu2/agentops/plugins/olympus-kit/cli/internal/ratchet"
	"github.com/boshu2/agentops/plugins/olympus-kit/cli/internal/types"
)

func TestWriteFeedbackEvents(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "feedback-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck // test cleanup

	events := []FeedbackEvent{
		{
			SessionID:     "session-20260125-120000",
			ArtifactPath:  "/path/to/learning1.jsonl",
			Reward:        0.85,
			UtilityBefore: 0.5,
			UtilityAfter:  0.535,
			Alpha:         0.1,
			RecordedAt:    time.Now(),
		},
		{
			SessionID:     "session-20260125-120000",
			ArtifactPath:  "/path/to/learning2.jsonl",
			Reward:        0.85,
			UtilityBefore: 0.6,
			UtilityAfter:  0.625,
			Alpha:         0.1,
			RecordedAt:    time.Now(),
		},
	}

	// Write events
	if err := writeFeedbackEvents(tempDir, events); err != nil {
		t.Fatalf("write feedback events: %v", err)
	}

	// Verify file exists
	feedbackPath := filepath.Join(tempDir, FeedbackFilePath)
	if _, err := os.Stat(feedbackPath); os.IsNotExist(err) {
		t.Fatal("feedback file not created")
	}

	// Read and verify content
	loaded, err := loadFeedbackEvents(tempDir)
	if err != nil {
		t.Fatalf("load feedback events: %v", err)
	}

	if len(loaded) != len(events) {
		t.Errorf("got %d events, want %d", len(loaded), len(events))
	}

	// Verify first event
	if loaded[0].SessionID != events[0].SessionID {
		t.Errorf("session ID mismatch: got %s, want %s", loaded[0].SessionID, events[0].SessionID)
	}
	if loaded[0].Reward != events[0].Reward {
		t.Errorf("reward mismatch: got %.2f, want %.2f", loaded[0].Reward, events[0].Reward)
	}
}

func TestLoadFeedbackEventsEmpty(t *testing.T) {
	// Create temp directory without feedback file
	tempDir, err := os.MkdirTemp("", "feedback-empty-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck // test cleanup

	// Load from non-existent file should return empty slice
	events, err := loadFeedbackEvents(tempDir)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
	if len(events) != 0 {
		t.Errorf("expected empty slice, got %d events", len(events))
	}
}

func TestCanonicalSessionID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		pattern  string // regex pattern to match if exact match not expected
	}{
		{
			name:    "empty generates timestamp",
			input:   "",
			pattern: `^session-\d{8}-\d{6}$`,
		},
		{
			name:     "already canonical",
			input:    "session-20260125-120000",
			expected: "session-20260125-120000",
		},
		{
			name:    "UUID format",
			input:   "2d608ace-e8e4-4649-8ac0-70aeba0dcfee",
			pattern: `^session-\d{8}-\d{6}$`,
		},
		{
			name:     "custom ID preserved",
			input:    "my-custom-session",
			expected: "my-custom-session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canonicalSessionID(tt.input)

			if tt.expected != "" {
				if result != tt.expected {
					t.Errorf("got %q, want %q", result, tt.expected)
				}
			} else if tt.pattern != "" {
				// Pattern match for generated IDs
				if !matchPattern(result, tt.pattern) {
					t.Errorf("result %q doesn't match pattern %s", result, tt.pattern)
				}
			}
		})
	}
}

func matchPattern(s, pattern string) bool {
	// Simple pattern match without regexp for testing
	// Just check basic format
	if pattern == `^session-\d{8}-\d{6}$` {
		if len(s) != 23 { // "session-YYYYMMDD-HHMMSS" = 23 chars
			return false
		}
		if s[:8] != "session-" {
			return false
		}
		return true
	}
	return false
}

func TestIntegrationFeedbackLoop(t *testing.T) {
	// Create temp directory with full structure
	tempDir, err := os.MkdirTemp("", "feedback-integration-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck // test cleanup

	// Create .agents/learnings/ directory
	learningsDir := filepath.Join(tempDir, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("create learnings dir: %v", err)
	}

	// Create .agents/olympus/ directory for citations
	olympusDir := filepath.Join(tempDir, ".agents", "olympus")
	if err := os.MkdirAll(olympusDir, 0755); err != nil {
		t.Fatalf("create olympus dir: %v", err)
	}

	// Create a test learning
	learningData := map[string]interface{}{
		"id":      "L-test-001",
		"title":   "Test Learning",
		"content": "This is a test learning",
		"utility": 0.5,
	}
	learningJSON, _ := json.Marshal(learningData)
	learningPath := filepath.Join(learningsDir, "L-test-001.jsonl")
	if err := os.WriteFile(learningPath, learningJSON, 0644); err != nil {
		t.Fatalf("write learning: %v", err)
	}

	// Create a citation for the learning
	sessionID := "session-20260125-120000"
	citation := types.CitationEvent{
		ArtifactPath: learningPath,
		SessionID:    sessionID,
		CitedAt:      time.Now(),
		CitationType: "retrieved",
	}
	if err := ratchet.RecordCitation(tempDir, citation); err != nil {
		t.Fatalf("record citation: %v", err)
	}

	// Verify citation was recorded
	citations, err := ratchet.LoadCitations(tempDir)
	if err != nil {
		t.Fatalf("load citations: %v", err)
	}
	if len(citations) != 1 {
		t.Fatalf("expected 1 citation, got %d", len(citations))
	}

	// Verify the citation has correct session ID
	if citations[0].SessionID != sessionID {
		t.Errorf("citation session ID mismatch: got %s, want %s", citations[0].SessionID, sessionID)
	}
}

func TestFeedbackFilePath(t *testing.T) {
	// Verify the feedback file path is correct
	expected := ".agents/olympus/feedback.jsonl"
	if FeedbackFilePath != expected {
		t.Errorf("FeedbackFilePath = %q, want %q", FeedbackFilePath, expected)
	}
}
