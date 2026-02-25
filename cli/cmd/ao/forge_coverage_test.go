package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/parser"
	"github.com/boshu2/agentops/cli/internal/storage"
	"github.com/boshu2/agentops/cli/internal/types"
)

// ---------------------------------------------------------------------------
// extractSnippet
// ---------------------------------------------------------------------------

func TestForgeCoverage_extractSnippet_NegativeStart(t *testing.T) {
	content := "hello world"
	got := extractSnippet(content, -10, 100)
	if got != content {
		t.Errorf("expected full content for negative start, got %q", got)
	}
}

func TestForgeCoverage_extractSnippet_PastEnd(t *testing.T) {
	got := extractSnippet("hello", 100, 50)
	if got != "" {
		t.Errorf("expected empty for past-end, got %q", got)
	}
}

func TestForgeCoverage_extractSnippet_ExactLength(t *testing.T) {
	content := "abcdef"
	got := extractSnippet(content, 0, 100)
	if got != content {
		t.Errorf("expected full content when maxLen exceeds content, got %q", got)
	}
}

func TestForgeCoverage_extractSnippet_Truncated(t *testing.T) {
	content := "the quick brown fox jumps over the lazy dog and keeps on running"
	got := extractSnippet(content, 0, 30)
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipsis at end of truncated snippet, got %q", got)
	}
	if len(got) > 35 { // 30 + "..." + some tolerance for word boundary
		t.Errorf("snippet too long: %d chars", len(got))
	}
}

func TestForgeCoverage_extractSnippet_FromMiddle(t *testing.T) {
	content := "abcdefghij klmnopqrst uvwxyz"
	got := extractSnippet(content, 11, 10)
	if !strings.HasPrefix(got, "klmnopqrst") {
		t.Errorf("expected snippet starting from index 11, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// lastSpaceIndex
// ---------------------------------------------------------------------------

func TestForgeCoverage_lastSpaceIndex(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"empty string", "", -1},
		{"no spaces", "abcdef", -1},
		{"trailing space", "abc ", 3},
		{"middle space", "ab cd ef", 5},
		{"leading space", " abc", 0},
		{"multiple spaces", "a b c d", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lastSpaceIndex(tt.s)
			if got != tt.want {
				t.Errorf("lastSpaceIndex(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// extractIssueIDs
// ---------------------------------------------------------------------------

func TestForgeCoverage_extractIssueIDs(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "no issues",
			content: "no issue ids here",
			want:    nil,
		},
		{
			name:    "single issue",
			content: "fixed issue ol-0001",
			want:    []string{"ol-0001"},
		},
		{
			name:    "multiple issues",
			content: "working on ag-abc and ol-def",
			want:    []string{"ag-abc", "ol-def"},
		},
		{
			name:    "issue with hyphen suffix",
			content: "see gt-abc-def for details",
			want:    []string{"gt-abc-def"},
		},
		{
			name:    "empty content",
			content: "",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractIssueIDs(tt.content)
			if len(got) != len(tt.want) {
				t.Errorf("extractIssueIDs(%q) returned %d issues, want %d", tt.content, len(got), len(tt.want))
				return
			}
			for i, id := range got {
				if id != tt.want[i] {
					t.Errorf("extractIssueIDs(%q)[%d] = %q, want %q", tt.content, i, id, tt.want[i])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// truncateString
// ---------------------------------------------------------------------------

func TestForgeCoverage_truncateString(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"very short max", "hello", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.s, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// dedup
// ---------------------------------------------------------------------------

func TestForgeCoverage_dedup(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		want  []string
	}{
		{"empty", []string{}, []string{}},
		{"no duplicates", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"with duplicates", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"all same", []string{"x", "x", "x"}, []string{"x"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dedup(tt.items)
			if len(got) != len(tt.want) {
				t.Errorf("dedup() returned %d items, want %d", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("dedup()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// countLines
// ---------------------------------------------------------------------------

func TestForgeCoverage_countLines(t *testing.T) {
	tmp := t.TempDir()

	tests := []struct {
		name    string
		content string
		want    int
	}{
		{"empty file", "", 0},
		{"single line no newline", "hello", 0},
		{"single line with newline", "hello\n", 1},
		{"three lines", "a\nb\nc\n", 3},
		{"multiple lines no trailing newline", "a\nb\nc", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmp, tt.name+".txt")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			got := countLines(path)
			if got != tt.want {
				t.Errorf("countLines() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestForgeCoverage_countLines_NonexistentFile(t *testing.T) {
	got := countLines("/nonexistent/file.txt")
	if got != 0 {
		t.Errorf("expected 0 for nonexistent file, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// generateSummary
// ---------------------------------------------------------------------------

func TestForgeCoverage_generateSummary(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		decisions []string
		knowledge []string
		wantHas   string
	}{
		{
			name:      "with decisions",
			decisions: []string{"decided to use Go"},
			knowledge: []string{"learned testing"},
			wantHas:   "decided to use Go",
		},
		{
			name:      "no decisions with knowledge",
			decisions: []string{},
			knowledge: []string{"learned testing patterns"},
			wantHas:   "learned testing",
		},
		{
			name:      "no decisions no knowledge",
			decisions: []string{},
			knowledge: []string{},
			wantHas:   "2026-01-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSummary(tt.decisions, tt.knowledge, date)
			if !strings.Contains(got, tt.wantHas) {
				t.Errorf("generateSummary() = %q, expected to contain %q", got, tt.wantHas)
			}
		})
	}
}

func TestForgeCoverage_generateSummary_LongDecision(t *testing.T) {
	long := strings.Repeat("x", 200)
	got := generateSummary([]string{long}, nil, time.Now())
	if len(got) > SummaryMaxLength {
		t.Errorf("summary too long: %d > %d", len(got), SummaryMaxLength)
	}
}

// ---------------------------------------------------------------------------
// initSession
// ---------------------------------------------------------------------------

func TestForgeCoverage_initSession(t *testing.T) {
	session := initSession("/path/to/transcript.jsonl")
	if session.TranscriptPath != "/path/to/transcript.jsonl" {
		t.Errorf("unexpected transcript path: %s", session.TranscriptPath)
	}
	if session.ToolCalls == nil {
		t.Error("expected ToolCalls map to be initialized")
	}
	if session.ID != "" {
		t.Error("expected empty session ID")
	}
}

// ---------------------------------------------------------------------------
// updateSessionMeta
// ---------------------------------------------------------------------------

func TestForgeCoverage_updateSessionMeta(t *testing.T) {
	session := &storage.Session{}
	now := time.Now()

	// First message sets ID and date
	msg1 := types.TranscriptMessage{
		SessionID: "sess-123",
		Timestamp: now,
	}
	updateSessionMeta(session, msg1)
	if session.ID != "sess-123" {
		t.Errorf("expected session ID 'sess-123', got %q", session.ID)
	}
	if session.Date != now {
		t.Error("expected date to be set from first message")
	}

	// Second message with earlier timestamp should update date
	earlier := now.Add(-time.Hour)
	msg2 := types.TranscriptMessage{
		SessionID: "sess-456",
		Timestamp: earlier,
	}
	updateSessionMeta(session, msg2)
	if session.ID != "sess-123" {
		t.Error("session ID should not change once set")
	}
	if session.Date != earlier {
		t.Error("date should be updated to earlier timestamp")
	}

	// Message with zero timestamp should not update date
	msg3 := types.TranscriptMessage{
		SessionID: "",
	}
	updateSessionMeta(session, msg3)
	if session.Date != earlier {
		t.Error("date should not change for zero timestamp")
	}
}

func TestForgeCoverage_updateSessionMeta_EmptySessionID(t *testing.T) {
	session := &storage.Session{}
	msg := types.TranscriptMessage{
		SessionID: "",
		Timestamp: time.Now(),
	}
	updateSessionMeta(session, msg)
	if session.ID != "" {
		t.Error("session ID should remain empty when message has no session ID")
	}
}

// ---------------------------------------------------------------------------
// extractMessageKnowledge
// ---------------------------------------------------------------------------

func TestForgeCoverage_extractMessageKnowledge_EmptyContent(t *testing.T) {
	extractor := parser.NewExtractor()
	state := &transcriptState{
		seenFiles:  make(map[string]bool),
		seenIssues: make(map[string]bool),
	}
	msg := types.TranscriptMessage{Content: ""}
	extractMessageKnowledge(msg, extractor, state)
	// Should be a no-op
	if len(state.decisions) != 0 || len(state.knowledge) != 0 {
		t.Error("expected no extractions from empty content")
	}
}

func TestForgeCoverage_extractMessageKnowledge_WithContent(t *testing.T) {
	extractor := parser.NewExtractor()
	state := &transcriptState{
		seenFiles:  make(map[string]bool),
		seenIssues: make(map[string]bool),
	}
	// Use content with patterns the extractor can find
	msg := types.TranscriptMessage{
		Content: "We decided to use PostgreSQL because it supports JSON indexing. The solution was to add a retry loop around the API calls.",
	}
	extractMessageKnowledge(msg, extractor, state)
	// The extractor may or may not find patterns depending on implementation
	// This test primarily verifies no panics occur
}

// ---------------------------------------------------------------------------
// extractMessageRefs
// ---------------------------------------------------------------------------

func TestForgeCoverage_extractMessageRefs(t *testing.T) {
	session := &storage.Session{ToolCalls: make(map[string]int)}
	state := &transcriptState{
		seenFiles:  make(map[string]bool),
		seenIssues: make(map[string]bool),
	}
	msg := types.TranscriptMessage{
		Content: "working on ag-abc and ol-def",
		Tools: []types.ToolCall{
			{
				Name:  "Write",
				Input: map[string]any{"file_path": "/test/file.go"},
			},
		},
	}
	extractMessageRefs(msg, session, state)
	if len(state.issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(state.issues))
	}
	if len(state.filesChanged) != 1 {
		t.Errorf("expected 1 file changed, got %d", len(state.filesChanged))
	}
}

// ---------------------------------------------------------------------------
// extractToolRefs
// ---------------------------------------------------------------------------

func TestForgeCoverage_extractToolRefs(t *testing.T) {
	session := &storage.Session{ToolCalls: make(map[string]int)}
	state := &transcriptState{
		seenFiles:  make(map[string]bool),
		seenIssues: make(map[string]bool),
	}
	tools := []types.ToolCall{
		{Name: "Write", Input: map[string]any{"file_path": "/a.go"}},
		{Name: "Read", Input: map[string]any{"path": "/b.go"}},
		{Name: "Write", Input: map[string]any{"file_path": "/c.go"}},
		{Name: "tool_result", Input: nil}, // should be skipped for counting
		{Name: "", Input: nil},            // empty name skipped
	}
	extractToolRefs(tools, session, state)
	if session.ToolCalls["Write"] != 2 {
		t.Errorf("expected Write=2, got %d", session.ToolCalls["Write"])
	}
	if session.ToolCalls["Read"] != 1 {
		t.Errorf("expected Read=1, got %d", session.ToolCalls["Read"])
	}
	if _, ok := session.ToolCalls["tool_result"]; ok {
		t.Error("tool_result should not be counted")
	}
	if len(state.filesChanged) != 3 {
		t.Errorf("expected 3 files changed, got %d", len(state.filesChanged))
	}
}

// ---------------------------------------------------------------------------
// extractFilePathsFromTool
// ---------------------------------------------------------------------------

func TestForgeCoverage_extractFilePathsFromTool(t *testing.T) {
	tests := []struct {
		name      string
		tool      types.ToolCall
		wantFiles int
	}{
		{
			name:      "nil input",
			tool:      types.ToolCall{Input: nil},
			wantFiles: 0,
		},
		{
			name: "file_path key",
			tool: types.ToolCall{
				Input: map[string]any{"file_path": "/test.go"},
			},
			wantFiles: 1,
		},
		{
			name: "path key",
			tool: types.ToolCall{
				Input: map[string]any{"path": "/test.go"},
			},
			wantFiles: 1,
		},
		{
			name: "both keys same file",
			tool: types.ToolCall{
				Input: map[string]any{"file_path": "/same.go", "path": "/same.go"},
			},
			wantFiles: 1, // deduped by seenFiles
		},
		{
			name: "both keys different files",
			tool: types.ToolCall{
				Input: map[string]any{"file_path": "/a.go", "path": "/b.go"},
			},
			wantFiles: 2,
		},
		{
			name: "non-string file_path",
			tool: types.ToolCall{
				Input: map[string]any{"file_path": 123},
			},
			wantFiles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &transcriptState{
				seenFiles:  make(map[string]bool),
				seenIssues: make(map[string]bool),
			}
			extractFilePathsFromTool(tt.tool, state)
			if len(state.filesChanged) != tt.wantFiles {
				t.Errorf("expected %d files, got %d", tt.wantFiles, len(state.filesChanged))
			}
		})
	}
}

func TestForgeCoverage_extractFilePathsFromTool_Dedup(t *testing.T) {
	state := &transcriptState{
		seenFiles:  make(map[string]bool),
		seenIssues: make(map[string]bool),
	}
	tool := types.ToolCall{
		Input: map[string]any{"file_path": "/same.go"},
	}
	extractFilePathsFromTool(tool, state)
	extractFilePathsFromTool(tool, state)
	if len(state.filesChanged) != 1 {
		t.Errorf("expected 1 file (deduped), got %d", len(state.filesChanged))
	}
}

// ---------------------------------------------------------------------------
// extractIssueRefs
// ---------------------------------------------------------------------------

func TestForgeCoverage_extractIssueRefs(t *testing.T) {
	state := &transcriptState{
		seenFiles:  make(map[string]bool),
		seenIssues: make(map[string]bool),
	}
	extractIssueRefs("working on ag-abc and ol-def", state)
	if len(state.issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(state.issues))
	}
	// Calling again should dedup
	extractIssueRefs("also see ag-abc", state)
	if len(state.issues) != 2 {
		t.Errorf("expected 2 issues after dedup, got %d", len(state.issues))
	}
}

func TestForgeCoverage_extractIssueRefs_NoIssues(t *testing.T) {
	state := &transcriptState{
		seenFiles:  make(map[string]bool),
		seenIssues: make(map[string]bool),
	}
	extractIssueRefs("no issue ids here", state)
	if len(state.issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(state.issues))
	}
}

// ---------------------------------------------------------------------------
// splitMarkdownSections
// ---------------------------------------------------------------------------

func TestForgeCoverage_splitMarkdownSections(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantLen  int
		wantDesc string
	}{
		{
			name:    "no headings",
			content: "just some text\nno headings here",
			wantLen: 1,
		},
		{
			name:    "single h1",
			content: "# Title\ncontent here",
			wantLen: 1,
		},
		{
			name:    "two h2 sections",
			content: "## Section 1\ncontent 1\n## Section 2\ncontent 2",
			wantLen: 2,
		},
		{
			name:    "preamble plus h2",
			content: "preamble text\n## Section 1\ncontent 1\n## Section 2\ncontent 2",
			wantLen: 3,
		},
		{
			name:    "h1 and h2 mixed",
			content: "# Title\nintro\n## Section 1\ncontent\n# Another\nmore",
			wantLen: 3,
		},
		{
			name:    "empty content",
			content: "",
			wantLen: 1, // empty content returns [""]
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitMarkdownSections(tt.content)
			if len(got) != tt.wantLen {
				t.Errorf("splitMarkdownSections() returned %d sections, want %d. Sections: %v", len(got), tt.wantLen, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// noFilesError
// ---------------------------------------------------------------------------

func TestForgeCoverage_noFilesError(t *testing.T) {
	// Quiet mode returns nil
	err := noFilesError(true, "no files found")
	if err != nil {
		t.Errorf("expected nil error in quiet mode, got %v", err)
	}

	// Non-quiet returns error
	err = noFilesError(false, "no files found")
	if err == nil {
		t.Error("expected error in non-quiet mode")
	}
	if err.Error() != "no files found" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// forgeWarnf
// ---------------------------------------------------------------------------

func TestForgeCoverage_forgeWarnf_Quiet(t *testing.T) {
	// Should be a no-op when quiet, no panic
	forgeWarnf(true, "test warning: %s\n", "detail")
}

func TestForgeCoverage_forgeWarnf_NotQuiet(t *testing.T) {
	// Should write to stderr, no panic
	forgeWarnf(false, "test warning: %s\n", "detail")
}

// ---------------------------------------------------------------------------
// forgeTotals.addSession
// ---------------------------------------------------------------------------

func TestForgeCoverage_forgeTotals_addSession(t *testing.T) {
	totals := forgeTotals{}
	session := &storage.Session{
		Decisions: []string{"d1", "d2"},
		Knowledge: []string{"k1"},
	}
	totals.addSession(session)
	if totals.sessions != 1 {
		t.Errorf("expected 1 session, got %d", totals.sessions)
	}
	if totals.decisions != 2 {
		t.Errorf("expected 2 decisions, got %d", totals.decisions)
	}
	if totals.knowledge != 1 {
		t.Errorf("expected 1 knowledge, got %d", totals.knowledge)
	}

	// Add another session
	session2 := &storage.Session{
		Decisions: []string{"d3"},
		Knowledge: []string{"k2", "k3"},
	}
	totals.addSession(session2)
	if totals.sessions != 2 {
		t.Errorf("expected 2 sessions, got %d", totals.sessions)
	}
	if totals.decisions != 3 {
		t.Errorf("expected 3 decisions, got %d", totals.decisions)
	}
	if totals.knowledge != 3 {
		t.Errorf("expected 3 knowledge, got %d", totals.knowledge)
	}
}

// ---------------------------------------------------------------------------
// handleForgeDryRun
// ---------------------------------------------------------------------------

func TestForgeCoverage_handleForgeDryRun(t *testing.T) {
	oldDryRun := dryRun
	defer func() { dryRun = oldDryRun }()

	t.Run("not dry run", func(t *testing.T) {
		dryRun = false
		var buf bytes.Buffer
		result := handleForgeDryRun(&buf, false, []string{"a.jsonl", "b.jsonl"}, "file(s)")
		if result {
			t.Error("expected false when not in dry-run mode")
		}
	})

	t.Run("dry run active", func(t *testing.T) {
		dryRun = true
		var buf bytes.Buffer
		result := handleForgeDryRun(&buf, false, []string{"a.jsonl", "b.jsonl"}, "file(s)")
		if !result {
			t.Error("expected true when in dry-run mode")
		}
		output := buf.String()
		if !strings.Contains(output, "2 file(s)") {
			t.Errorf("expected file count in output, got %q", output)
		}
		if !strings.Contains(output, "a.jsonl") {
			t.Errorf("expected file name in output, got %q", output)
		}
	})

	t.Run("quiet suppresses output", func(t *testing.T) {
		dryRun = true
		var buf bytes.Buffer
		result := handleForgeDryRun(&buf, true, []string{"a.jsonl"}, "file(s)")
		if result {
			t.Error("expected false when quiet is true")
		}
	})
}

// ---------------------------------------------------------------------------
// printForgeSummary
// ---------------------------------------------------------------------------

func TestForgeCoverage_printForgeSummary(t *testing.T) {
	var buf bytes.Buffer
	totals := forgeTotals{sessions: 3, decisions: 5, knowledge: 10}
	printForgeSummary(&buf, totals, "/path/to/output", "session(s)")
	output := buf.String()
	if !strings.Contains(output, "3 session(s)") {
		t.Errorf("expected session count in output, got %q", output)
	}
	if !strings.Contains(output, "Decisions: 5") {
		t.Errorf("expected decisions count, got %q", output)
	}
	if !strings.Contains(output, "Knowledge: 10") {
		t.Errorf("expected knowledge count, got %q", output)
	}
	if !strings.Contains(output, "/path/to/output") {
		t.Errorf("expected base dir in output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// reportProgress
// ---------------------------------------------------------------------------

func TestForgeCoverage_reportProgress(t *testing.T) {
	t.Run("quiet mode", func(t *testing.T) {
		var buf bytes.Buffer
		lastProgress := 0
		reportProgress(true, &buf, 2000, 5000, &lastProgress)
		if buf.Len() > 0 {
			t.Error("expected no output in quiet mode")
		}
	})

	t.Run("not enough lines since last", func(t *testing.T) {
		var buf bytes.Buffer
		lastProgress := 500
		reportProgress(false, &buf, 600, 5000, &lastProgress)
		if buf.Len() > 0 {
			t.Error("expected no output when less than 1000 lines since last progress")
		}
	})

	t.Run("enough lines to report", func(t *testing.T) {
		var buf bytes.Buffer
		lastProgress := 0
		reportProgress(false, &buf, 1000, 5000, &lastProgress)
		if buf.Len() == 0 {
			t.Error("expected progress output")
		}
		if lastProgress != 1000 {
			t.Errorf("expected lastProgress=1000, got %d", lastProgress)
		}
	})

	t.Run("zero total lines", func(t *testing.T) {
		var buf bytes.Buffer
		lastProgress := 0
		reportProgress(false, &buf, 1000, 0, &lastProgress)
		output := buf.String()
		if !strings.Contains(output, "0%") {
			t.Errorf("expected 0%% in output for zero total, got %q", output)
		}
	})
}

// ---------------------------------------------------------------------------
// drainParseErrors
// ---------------------------------------------------------------------------

func TestForgeCoverage_drainParseErrors(t *testing.T) {
	t.Run("empty channel", func(t *testing.T) {
		ch := make(chan error, 1)
		err := drainParseErrors(ch)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("channel with error", func(t *testing.T) {
		ch := make(chan error, 1)
		ch <- os.ErrNotExist
		err := drainParseErrors(ch)
		if err != os.ErrNotExist {
			t.Errorf("expected os.ErrNotExist, got %v", err)
		}
	})

	t.Run("channel with nil", func(t *testing.T) {
		ch := make(chan error, 1)
		ch <- nil
		err := drainParseErrors(ch)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// finalizeTranscriptSession
// ---------------------------------------------------------------------------

func TestForgeCoverage_finalizeTranscriptSession(t *testing.T) {
	session := &storage.Session{}
	state := &transcriptState{
		decisions:    []string{"d1", "d1", "d2"}, // has duplicates
		knowledge:    []string{"k1", "k2", "k1"}, // has duplicates
		filesChanged: []string{"/a.go", "/b.go"},
		issues:       []string{"ag-abc"},
	}

	finalizeTranscriptSession(session, state, 4000)

	if len(session.Decisions) != 2 {
		t.Errorf("expected 2 decisions after dedup, got %d", len(session.Decisions))
	}
	if len(session.Knowledge) != 2 {
		t.Errorf("expected 2 knowledge after dedup, got %d", len(session.Knowledge))
	}
	if len(session.FilesChanged) != 2 {
		t.Errorf("expected 2 files, got %d", len(session.FilesChanged))
	}
	if len(session.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(session.Issues))
	}
	if session.Tokens.Total != 1000 { // 4000 / CharsPerToken(4)
		t.Errorf("expected 1000 tokens, got %d", session.Tokens.Total)
	}
	if !session.Tokens.Estimated {
		t.Error("expected estimated=true")
	}
	if session.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

// ---------------------------------------------------------------------------
// transcriptState
// ---------------------------------------------------------------------------

func TestForgeCoverage_transcriptState_Init(t *testing.T) {
	state := &transcriptState{
		seenFiles:  make(map[string]bool),
		seenIssues: make(map[string]bool),
	}
	if len(state.decisions) != 0 {
		t.Error("expected empty decisions")
	}
	if len(state.knowledge) != 0 {
		t.Error("expected empty knowledge")
	}
	if len(state.filesChanged) != 0 {
		t.Error("expected empty filesChanged")
	}
	if len(state.issues) != 0 {
		t.Error("expected empty issues")
	}
}

// ---------------------------------------------------------------------------
// collectFilesFromPatterns
// ---------------------------------------------------------------------------

func TestForgeCoverage_collectFilesFromPatterns(t *testing.T) {
	tmp := t.TempDir()

	// Create test files
	for _, name := range []string{"a.jsonl", "b.jsonl", "c.md", "d.txt"} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("glob pattern", func(t *testing.T) {
		files, err := collectFilesFromPatterns([]string{filepath.Join(tmp, "*.jsonl")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 2 {
			t.Errorf("expected 2 files, got %d", len(files))
		}
	})

	t.Run("literal path", func(t *testing.T) {
		files, err := collectFilesFromPatterns([]string{filepath.Join(tmp, "c.md")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 1 {
			t.Errorf("expected 1 file, got %d", len(files))
		}
	})

	t.Run("with filter", func(t *testing.T) {
		pattern := filepath.Join(tmp, "*")
		files, err := collectFilesFromPatterns([]string{pattern}, func(path string) bool {
			return filepath.Ext(path) == ".md"
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 1 {
			t.Errorf("expected 1 file (filtered to .md), got %d", len(files))
		}
	})

	t.Run("nonexistent literal path", func(t *testing.T) {
		files, err := collectFilesFromPatterns([]string{filepath.Join(tmp, "nonexistent.jsonl")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 0 {
			t.Errorf("expected 0 files for nonexistent path, got %d", len(files))
		}
	})

	t.Run("invalid glob pattern", func(t *testing.T) {
		_, err := collectFilesFromPatterns([]string{"[invalid"}, nil)
		if err == nil {
			t.Error("expected error for invalid glob pattern")
		}
	})
}

// ---------------------------------------------------------------------------
// resolveMarkdownFiles
// ---------------------------------------------------------------------------

func TestForgeCoverage_resolveMarkdownFiles(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{"a.md", "b.md", "c.txt"} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := resolveMarkdownFiles([]string{filepath.Join(tmp, "*")})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 markdown files, got %d", len(files))
	}
}

// ---------------------------------------------------------------------------
// resolveTranscriptFiles
// ---------------------------------------------------------------------------

func TestForgeCoverage_resolveTranscriptFiles_WithArgs(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{"a.jsonl", "b.jsonl"} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	oldLastSession := forgeLastSession
	defer func() { forgeLastSession = oldLastSession }()
	forgeLastSession = false

	files, err := resolveTranscriptFiles([]string{filepath.Join(tmp, "*.jsonl")}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestForgeCoverage_resolveTranscriptFiles_LastSessionQuiet(t *testing.T) {
	oldLastSession := forgeLastSession
	defer func() { forgeLastSession = oldLastSession }()
	forgeLastSession = true

	t.Setenv("HOME", t.TempDir())

	// Should return nil, nil in quiet mode when no sessions found
	files, err := resolveTranscriptFiles(nil, true)
	if err != nil {
		t.Fatalf("expected nil error in quiet mode, got %v", err)
	}
	if files != nil {
		t.Error("expected nil files in quiet mode when no sessions")
	}
}

// ---------------------------------------------------------------------------
// isTranscriptCandidate
// ---------------------------------------------------------------------------

func TestForgeCoverage_isTranscriptCandidate(t *testing.T) {
	tmp := t.TempDir()
	projectsDir := tmp

	// Create a valid candidate
	path := filepath.Join(tmp, "session.jsonl")
	content := strings.Repeat("x", 200)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)
	if !isTranscriptCandidate(path, info, projectsDir) {
		t.Error("expected valid .jsonl file to be a candidate")
	}

	// Directory is not a candidate
	dirPath := filepath.Join(tmp, "subdir")
	os.MkdirAll(dirPath, 0755)
	dirInfo, _ := os.Stat(dirPath)
	if isTranscriptCandidate(dirPath, dirInfo, projectsDir) {
		t.Error("expected directory to not be a candidate")
	}

	// Non-jsonl file
	txtPath := filepath.Join(tmp, "file.txt")
	os.WriteFile(txtPath, []byte(content), 0644)
	txtInfo, _ := os.Stat(txtPath)
	if isTranscriptCandidate(txtPath, txtInfo, projectsDir) {
		t.Error("expected .txt file to not be a candidate")
	}

	// Too small file
	smallPath := filepath.Join(tmp, "small.jsonl")
	os.WriteFile(smallPath, []byte("x"), 0644)
	smallInfo, _ := os.Stat(smallPath)
	if isTranscriptCandidate(smallPath, smallInfo, projectsDir) {
		t.Error("expected small file to not be a candidate")
	}
}

// ---------------------------------------------------------------------------
// collectTranscriptCandidates
// ---------------------------------------------------------------------------

func TestForgeCoverage_collectTranscriptCandidates(t *testing.T) {
	tmp := t.TempDir()
	content := strings.Repeat("x", 200)

	// Create valid candidates
	if err := os.WriteFile(filepath.Join(tmp, "session1.jsonl"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "session2.jsonl"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	// Non-candidate
	if err := os.WriteFile(filepath.Join(tmp, "readme.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a subagents directory (should be skipped)
	subagentsDir := filepath.Join(tmp, "subagents")
	os.MkdirAll(subagentsDir, 0755)
	os.WriteFile(filepath.Join(subagentsDir, "sub.jsonl"), []byte(content), 0644)

	candidates, err := collectTranscriptCandidates(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Errorf("expected 2 candidates (subagents skipped), got %d", len(candidates))
	}
}

func TestForgeCoverage_collectTranscriptCandidates_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	candidates, err := collectTranscriptCandidates(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(candidates))
	}
}

// ---------------------------------------------------------------------------
// findLastSession
// ---------------------------------------------------------------------------

func TestForgeCoverage_findLastSession_NoProjectsDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	_, err := findLastSession()
	if err == nil {
		t.Error("expected error when no projects dir")
	}
}

func TestForgeCoverage_findLastSession_EmptyProjectsDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	os.MkdirAll(filepath.Join(home, ".claude", "projects"), 0755)
	_, err := findLastSession()
	if err == nil {
		t.Error("expected error when no transcript files")
	}
}

func TestForgeCoverage_findLastSession_FindsMostRecent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectsDir := filepath.Join(home, ".claude", "projects")
	os.MkdirAll(projectsDir, 0755)

	content := strings.Repeat("x", 200)

	// Create two transcript files with different modification times
	oldPath := filepath.Join(projectsDir, "old.jsonl")
	newPath := filepath.Join(projectsDir, "new.jsonl")

	os.WriteFile(oldPath, []byte(content), 0644)
	oldTime := time.Now().Add(-time.Hour)
	os.Chtimes(oldPath, oldTime, oldTime)

	os.WriteFile(newPath, []byte(content), 0644)
	// newPath will have the current time

	result, err := findLastSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != newPath {
		t.Errorf("expected most recent file %s, got %s", newPath, result)
	}
}

// ---------------------------------------------------------------------------
// fileWithTime type
// ---------------------------------------------------------------------------

func TestForgeCoverage_fileWithTime(t *testing.T) {
	f := fileWithTime{
		path:    "/test/path.jsonl",
		modTime: time.Now(),
	}
	if f.path != "/test/path.jsonl" {
		t.Error("unexpected path")
	}
}

// ---------------------------------------------------------------------------
// queueForExtraction
// ---------------------------------------------------------------------------

func TestForgeCoverage_queueForExtraction(t *testing.T) {
	tmp := t.TempDir()
	session := &storage.Session{
		ID:        "test-session-123",
		Summary:   "test summary",
		Decisions: []string{"d1"},
		Knowledge: []string{"k1"},
	}

	err := queueForExtraction(session, "/path/to/session.md", "/path/to/transcript.jsonl", tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify pending.jsonl was created
	pendingPath := filepath.Join(tmp, storage.DefaultBaseDir, "pending.jsonl")
	data, err := os.ReadFile(pendingPath)
	if err != nil {
		t.Fatalf("pending file not created: %v", err)
	}

	// Verify content is valid JSONL
	var parsed map[string]any
	if err := json.Unmarshal(data[:len(data)-1], &parsed); err != nil { // trim trailing newline
		t.Fatalf("invalid JSONL: %v", err)
	}
	if parsed["session_id"] != "test-session-123" {
		t.Errorf("unexpected session_id: %v", parsed["session_id"])
	}
}

func TestForgeCoverage_queueForExtraction_Append(t *testing.T) {
	tmp := t.TempDir()
	session1 := &storage.Session{ID: "sess-001", Summary: "first"}
	session2 := &storage.Session{ID: "sess-002", Summary: "second"}

	_ = queueForExtraction(session1, "path1", "trans1", tmp)
	_ = queueForExtraction(session2, "path2", "trans2", tmp)

	pendingPath := filepath.Join(tmp, storage.DefaultBaseDir, "pending.jsonl")
	data, err := os.ReadFile(pendingPath)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines in pending file, got %d", len(lines))
	}
}

// ---------------------------------------------------------------------------
// writeSessionIndex
// ---------------------------------------------------------------------------

func TestForgeCoverage_writeSessionIndex(t *testing.T) {
	tmp := t.TempDir()
	baseDir := filepath.Join(tmp, ".agents", "ao")
	fs := storage.NewFileStorage(storage.WithBaseDir(baseDir))
	if err := fs.Init(); err != nil {
		t.Fatal(err)
	}

	session := &storage.Session{
		ID:      "test-123",
		Date:    time.Now(),
		Summary: "test session",
	}
	err := writeSessionIndex(fs, session, "sessions/test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// writeSessionProvenance
// ---------------------------------------------------------------------------

func TestForgeCoverage_writeSessionProvenance(t *testing.T) {
	tmp := t.TempDir()
	baseDir := filepath.Join(tmp, ".agents", "ao")
	fs := storage.NewFileStorage(storage.WithBaseDir(baseDir))
	if err := fs.Init(); err != nil {
		t.Fatal(err)
	}

	err := writeSessionProvenance(fs, "test-session-id", "sessions/test.md", "/path/to/transcript.jsonl", "transcript", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestForgeCoverage_writeSessionProvenance_NoSessionID(t *testing.T) {
	tmp := t.TempDir()
	baseDir := filepath.Join(tmp, ".agents", "ao")
	fs := storage.NewFileStorage(storage.WithBaseDir(baseDir))
	if err := fs.Init(); err != nil {
		t.Fatal(err)
	}

	err := writeSessionProvenance(fs, "abcdefg", "sessions/test.md", "/path/to/file.md", "markdown", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// updateSearchIndexForFile
// ---------------------------------------------------------------------------

func TestForgeCoverage_updateSearchIndexForFile_NoIndex(t *testing.T) {
	tmp := t.TempDir()
	// No index file — should be a no-op
	updateSearchIndexForFile(tmp, filepath.Join(tmp, "test.md"), false)
}

// ---------------------------------------------------------------------------
// consumeTranscriptMessages
// ---------------------------------------------------------------------------

func TestForgeCoverage_consumeTranscriptMessages(t *testing.T) {
	session := initSession("/test.jsonl")
	extractor := parser.NewExtractor()
	state := &transcriptState{
		seenFiles:  make(map[string]bool),
		seenIssues: make(map[string]bool),
	}

	msgCh := make(chan types.TranscriptMessage, 3)
	msgCh <- types.TranscriptMessage{
		SessionID: "s1",
		Content:   "hello world",
		Timestamp: time.Now(),
	}
	msgCh <- types.TranscriptMessage{
		Content: "working on ag-abc",
		Tools: []types.ToolCall{
			{Name: "Write", Input: map[string]any{"file_path": "/test.go"}},
		},
	}
	msgCh <- types.TranscriptMessage{
		Content: "another message",
	}
	close(msgCh)

	var buf bytes.Buffer
	consumeTranscriptMessages(msgCh, session, extractor, state, true, &buf, 100)

	if session.ID != "s1" {
		t.Errorf("expected session ID 's1', got %q", session.ID)
	}
	if len(state.filesChanged) != 1 {
		t.Errorf("expected 1 file changed, got %d", len(state.filesChanged))
	}
}

// ---------------------------------------------------------------------------
// processMarkdown
// ---------------------------------------------------------------------------

func TestForgeCoverage_processMarkdown_EmptyFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.md")
	os.WriteFile(path, []byte(""), 0644)
	extractor := parser.NewExtractor()
	_, err := processMarkdown(path, extractor, true)
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestForgeCoverage_processMarkdown_ValidFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "valid.md")
	content := "# Title\n\nSome content here.\n\n## Section 2\n\nMore content about ag-xyz."
	os.WriteFile(path, []byte(content), 0644)

	extractor := parser.NewExtractor()
	session, err := processMarkdown(path, extractor, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("expected non-nil session")
	}
	if !strings.HasPrefix(session.ID, "md-") {
		t.Errorf("expected ID to start with 'md-', got %q", session.ID)
	}
	if session.TranscriptPath != path {
		t.Errorf("unexpected transcript path: %s", session.TranscriptPath)
	}
}

func TestForgeCoverage_processMarkdown_NonexistentFile(t *testing.T) {
	extractor := parser.NewExtractor()
	_, err := processMarkdown("/nonexistent/file.md", extractor, true)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// ---------------------------------------------------------------------------
// processMarkdown deterministic ID
// ---------------------------------------------------------------------------

func TestForgeCoverage_processMarkdown_DeterministicID(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.md")
	os.WriteFile(path, []byte("# Test\ncontent"), 0644)

	extractor := parser.NewExtractor()
	s1, _ := processMarkdown(path, extractor, true)
	s2, _ := processMarkdown(path, extractor, true)

	if s1.ID != s2.ID {
		t.Errorf("expected deterministic IDs to match, got %q and %q", s1.ID, s2.ID)
	}
}

// ---------------------------------------------------------------------------
// issueIDPattern
// ---------------------------------------------------------------------------

func TestForgeCoverage_issueIDPattern(t *testing.T) {
	tests := []struct {
		input   string
		matches []string
	}{
		{"ag-abc", []string{"ag-abc"}},
		{"ol-0001", []string{"ol-0001"}},
		{"gt-abc-def", []string{"gt-abc-def"}},
		{"no match here", nil},
		{"a-bc", nil},        // prefix too short
		{"abcd-ab", nil},     // prefix too long (4 chars)
		{"ag-ab", nil},       // suffix too short (2 chars)
		{"ag-abcdefgh", nil}, // suffix too long (8 chars)
	}

	for _, tt := range tests {
		matches := issueIDPattern.FindAllString(tt.input, -1)
		if len(matches) != len(tt.matches) {
			t.Errorf("issueIDPattern on %q: got %v, want %v", tt.input, matches, tt.matches)
		}
	}
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

func TestForgeCoverage_Constants(t *testing.T) {
	if SnippetMaxLength != 200 {
		t.Errorf("SnippetMaxLength = %d, want 200", SnippetMaxLength)
	}
	if SummaryMaxLength != 100 {
		t.Errorf("SummaryMaxLength = %d, want 100", SummaryMaxLength)
	}
	if CharsPerToken != 4 {
		t.Errorf("CharsPerToken = %d, want 4", CharsPerToken)
	}
}
