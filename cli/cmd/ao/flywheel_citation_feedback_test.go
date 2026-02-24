package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boshu2/agentops/cli/internal/types"
)

func TestExtractLearningID_RelativePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{".agents/learnings/2026-02-24-abc.md", "2026-02-24-abc.md"},
		{".agents/patterns/auth-pattern.md", "auth-pattern.md"},
		{".agents/learnings/nested/deep.md", "nested/deep.md"},
		{"random/path/file.md", "file.md"}, // fallback to Base
	}
	for _, tt := range tests {
		got := extractLearningID(tt.input)
		if got != tt.want {
			t.Errorf("extractLearningID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractLearningID_AbsolutePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/home/user/repo/.agents/learnings/2026-02-24-abc.md", "2026-02-24-abc.md"},
		{"/tmp/work/.agents/patterns/go-pattern.md", "go-pattern.md"},
	}
	for _, tt := range tests {
		got := extractLearningID(tt.input)
		if got != tt.want {
			t.Errorf("extractLearningID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMarkCitationsFeedbackGiven_WritesAllTrue(t *testing.T) {
	tmp := t.TempDir()
	citationsPath := filepath.Join(tmp, "citations.jsonl")

	citations := []types.CitationEvent{
		{ArtifactPath: ".agents/learnings/a.md", FeedbackGiven: false},
		{ArtifactPath: ".agents/learnings/b.md", FeedbackGiven: false},
		{ArtifactPath: ".agents/learnings/c.md", FeedbackGiven: true}, // already processed
	}

	markCitationsFeedbackGiven(citationsPath, citations)

	data, err := os.ReadFile(citationsPath)
	if err != nil {
		t.Fatalf("failed to read citations: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		var c types.CitationEvent
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			t.Fatalf("line %d: unmarshal error: %v", i, err)
		}
		if !c.FeedbackGiven {
			t.Errorf("line %d: expected FeedbackGiven=true, got false", i)
		}
	}
}

func TestMarkCitationsFeedbackGiven_EmptyList(t *testing.T) {
	tmp := t.TempDir()
	citationsPath := filepath.Join(tmp, "citations.jsonl")

	markCitationsFeedbackGiven(citationsPath, nil)

	data, err := os.ReadFile(citationsPath)
	if err != nil {
		t.Fatalf("failed to read citations: %v", err)
	}

	// Should just have a trailing newline
	if strings.TrimSpace(string(data)) != "" {
		t.Errorf("expected empty content, got %q", string(data))
	}
}

func TestProcessCitationFeedback_NoCitationsFile(t *testing.T) {
	tmp := t.TempDir()
	total, rewarded, skipped := processCitationFeedback(tmp)
	if total != 0 || rewarded != 0 || skipped != 0 {
		t.Errorf("expected (0,0,0), got (%d,%d,%d)", total, rewarded, skipped)
	}
}

func TestProcessCitationFeedback_AllAlreadyProcessed(t *testing.T) {
	tmp := t.TempDir()
	aoDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(aoDir, 0755); err != nil {
		t.Fatal(err)
	}

	citations := []types.CitationEvent{
		{ArtifactPath: ".agents/learnings/a.md", FeedbackGiven: true},
		{ArtifactPath: ".agents/learnings/b.md", FeedbackGiven: true},
	}

	var lines []string
	for _, c := range citations {
		data, _ := json.Marshal(c)
		lines = append(lines, string(data))
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(aoDir, "citations.jsonl"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	total, rewarded, skipped := processCitationFeedback(tmp)
	if total != 0 || rewarded != 0 || skipped != 0 {
		t.Errorf("expected (0,0,0) for all-processed, got (%d,%d,%d)", total, rewarded, skipped)
	}
}
