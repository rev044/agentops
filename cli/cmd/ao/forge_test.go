package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/storage"
	"github.com/boshu2/agentops/cli/internal/types"
)

// ---------------------------------------------------------------------------
// collectFilesFromPatterns
// ---------------------------------------------------------------------------

func TestForge_collectFilesFromPatterns_LiteralPath(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "example.jsonl")
	if err := os.WriteFile(f, []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := collectFilesFromPatterns([]string{f}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0] != f {
		t.Errorf("expected [%s], got %v", f, files)
	}
}

func TestForge_collectFilesFromPatterns_GlobPattern(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{"a.jsonl", "b.jsonl", "c.txt"} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	pattern := filepath.Join(tmp, "*.jsonl")
	files, err := collectFilesFromPatterns([]string{pattern}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(files), files)
	}
}

func TestForge_collectFilesFromPatterns_WithFilter(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{"a.md", "b.md", "c.txt"} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("# hi"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	pattern := filepath.Join(tmp, "*")
	mdFilter := func(path string) bool {
		return filepath.Ext(path) == ".md"
	}

	files, err := collectFilesFromPatterns([]string{pattern}, mdFilter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 .md files, got %d: %v", len(files), files)
	}
}

func TestForge_collectFilesFromPatterns_InvalidPattern(t *testing.T) {
	_, err := collectFilesFromPatterns([]string{"[invalid"}, nil)
	if err == nil {
		t.Error("expected error for invalid glob pattern")
	}
}

func TestForge_collectFilesFromPatterns_NonexistentLiteralPath(t *testing.T) {
	files, err := collectFilesFromPatterns([]string{"/nonexistent/file.jsonl"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for nonexistent literal path, got %d", len(files))
	}
}

// ---------------------------------------------------------------------------
// resolveMarkdownFiles
// ---------------------------------------------------------------------------

func TestForge_resolveMarkdownFiles_FiltersMd(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{"notes.md", "data.jsonl", "readme.md"} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	pattern := filepath.Join(tmp, "*")
	files, err := resolveMarkdownFiles([]string{pattern})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 .md files, got %d: %v", len(files), files)
	}
	for _, f := range files {
		if filepath.Ext(f) != ".md" {
			t.Errorf("non-.md file included: %s", f)
		}
	}
}

// ---------------------------------------------------------------------------
// handleForgeDryRun
// ---------------------------------------------------------------------------

func TestForge_handleForgeDryRun_Active(t *testing.T) {
	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	var buf bytes.Buffer
	files := []string{"a.jsonl", "b.jsonl"}
	result := handleForgeDryRun(&buf, false, files, "file(s)")
	if !result {
		t.Error("expected true return when dry-run is active")
	}
	out := buf.String()
	if !strings.Contains(out, "dry-run") {
		t.Errorf("expected dry-run message, got %q", out)
	}
	if !strings.Contains(out, "2 file(s)") {
		t.Errorf("expected '2 file(s)' in output, got %q", out)
	}
}

func TestForge_handleForgeDryRun_Quiet(t *testing.T) {
	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	var buf bytes.Buffer
	result := handleForgeDryRun(&buf, true, []string{"a.jsonl"}, "file(s)")
	if result {
		t.Error("expected false return when quiet is true (even if dry-run)")
	}
}

func TestForge_handleForgeDryRun_NotDryRun(t *testing.T) {
	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	var buf bytes.Buffer
	result := handleForgeDryRun(&buf, false, []string{"a.jsonl"}, "file(s)")
	if result {
		t.Error("expected false return when dry-run is off")
	}
}

// ---------------------------------------------------------------------------
// noFilesError
// ---------------------------------------------------------------------------

func TestForge_noFilesError_Quiet(t *testing.T) {
	err := noFilesError(true, "some error")
	if err != nil {
		t.Errorf("expected nil error when quiet, got %v", err)
	}
}

func TestForge_noFilesError_NotQuiet(t *testing.T) {
	err := noFilesError(false, "no files found")
	if err == nil {
		t.Error("expected error when not quiet")
	}
	if err.Error() != "no files found" {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// forgeTotals
// ---------------------------------------------------------------------------

func TestForge_forgeTotals_addSession(t *testing.T) {
	totals := forgeTotals{}
	session := &storage.Session{
		Decisions: []string{"decision1", "decision2"},
		Knowledge: []string{"knowledge1"},
	}

	totals.addSession(session)
	if totals.sessions != 1 {
		t.Errorf("sessions = %d, want 1", totals.sessions)
	}
	if totals.decisions != 2 {
		t.Errorf("decisions = %d, want 2", totals.decisions)
	}
	if totals.knowledge != 1 {
		t.Errorf("knowledge = %d, want 1", totals.knowledge)
	}

	// Add another session
	totals.addSession(&storage.Session{
		Decisions: []string{"d3"},
		Knowledge: []string{"k2", "k3"},
	})
	if totals.sessions != 2 {
		t.Errorf("sessions = %d, want 2", totals.sessions)
	}
	if totals.decisions != 3 {
		t.Errorf("decisions = %d, want 3", totals.decisions)
	}
	if totals.knowledge != 3 {
		t.Errorf("knowledge = %d, want 3", totals.knowledge)
	}
}

// ---------------------------------------------------------------------------
// printForgeSummary
// ---------------------------------------------------------------------------

func TestForge_printForgeSummary(t *testing.T) {
	var buf bytes.Buffer
	totals := forgeTotals{sessions: 3, decisions: 5, knowledge: 7}
	printForgeSummary(&buf, totals, "/tmp/out", "session(s)")

	out := buf.String()
	if !strings.Contains(out, "3 session(s)") {
		t.Errorf("expected '3 session(s)' in output, got %q", out)
	}
	if !strings.Contains(out, "Decisions: 5") {
		t.Errorf("expected 'Decisions: 5' in output, got %q", out)
	}
	if !strings.Contains(out, "Knowledge: 7") {
		t.Errorf("expected 'Knowledge: 7' in output, got %q", out)
	}
	if !strings.Contains(out, "/tmp/out") {
		t.Errorf("expected output dir in summary, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// splitMarkdownSections
// ---------------------------------------------------------------------------

func TestForge_splitMarkdownSections(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantLen  int
		checkIdx int
		contains string
	}{
		{
			name:     "two h2 sections",
			content:  "## First\nSome content\n## Second\nMore content",
			wantLen:  2,
			checkIdx: 1,
			contains: "## Second",
		},
		{
			name:     "h1 and h2 mix",
			content:  "# Title\nIntro\n## Section\nBody",
			wantLen:  2,
			checkIdx: 0,
			contains: "# Title",
		},
		{
			name:     "no headings",
			content:  "Just plain text\nwith lines",
			wantLen:  1,
			checkIdx: 0,
			contains: "Just plain text",
		},
		{
			name:     "empty content",
			content:  "",
			wantLen:  1,
			checkIdx: 0,
			contains: "",
		},
		{
			name:     "heading at start followed by content then heading",
			content:  "## A\nline1\nline2\n## B\nline3",
			wantLen:  2,
			checkIdx: 0,
			contains: "line1",
		},
		{
			name:     "three sections",
			content:  "# One\na\n# Two\nb\n# Three\nc",
			wantLen:  3,
			checkIdx: 2,
			contains: "# Three",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := splitMarkdownSections(tt.content)
			if len(sections) != tt.wantLen {
				t.Errorf("got %d sections, want %d; sections=%v", len(sections), tt.wantLen, sections)
			}
			if tt.checkIdx < len(sections) && !strings.Contains(sections[tt.checkIdx], tt.contains) {
				t.Errorf("section[%d] = %q, want to contain %q", tt.checkIdx, sections[tt.checkIdx], tt.contains)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// initSession
// ---------------------------------------------------------------------------

func TestForge_initSession(t *testing.T) {
	session := initSession("/path/to/transcript.jsonl")
	if session.TranscriptPath != "/path/to/transcript.jsonl" {
		t.Errorf("TranscriptPath = %q, want %q", session.TranscriptPath, "/path/to/transcript.jsonl")
	}
	if session.ToolCalls == nil {
		t.Error("ToolCalls map should be initialized")
	}
	if session.ID != "" {
		t.Errorf("ID should be empty initially, got %q", session.ID)
	}
}

// ---------------------------------------------------------------------------
// updateSessionMeta
// ---------------------------------------------------------------------------

func TestForge_updateSessionMeta(t *testing.T) {
	t.Run("sets ID from first message", func(t *testing.T) {
		session := &storage.Session{}
		msg := types.TranscriptMessage{SessionID: "sess-123"}
		updateSessionMeta(session, msg)
		if session.ID != "sess-123" {
			t.Errorf("ID = %q, want %q", session.ID, "sess-123")
		}
	})

	t.Run("does not overwrite existing ID", func(t *testing.T) {
		session := &storage.Session{ID: "first-id"}
		msg := types.TranscriptMessage{SessionID: "second-id"}
		updateSessionMeta(session, msg)
		if session.ID != "first-id" {
			t.Errorf("ID = %q, want %q (should not overwrite)", session.ID, "first-id")
		}
	})

	t.Run("picks earliest timestamp", func(t *testing.T) {
		t1 := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
		session := &storage.Session{Date: t1}

		msg := types.TranscriptMessage{Timestamp: t2}
		updateSessionMeta(session, msg)
		if !session.Date.Equal(t2) {
			t.Errorf("Date = %v, want %v (earlier)", session.Date, t2)
		}
	})

	t.Run("ignores zero timestamp", func(t *testing.T) {
		t1 := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)
		session := &storage.Session{Date: t1}

		msg := types.TranscriptMessage{}
		updateSessionMeta(session, msg)
		if !session.Date.Equal(t1) {
			t.Errorf("Date should remain unchanged, got %v", session.Date)
		}
	})
}

// ---------------------------------------------------------------------------
// extractFilePathsFromTool
// ---------------------------------------------------------------------------

func TestForge_extractFilePathsFromTool(t *testing.T) {
	t.Run("file_path input", func(t *testing.T) {
		state := &transcriptState{
			seenFiles: make(map[string]bool),
		}
		tool := types.ToolCall{
			Name:  "Read",
			Input: map[string]any{"file_path": "/src/main.go"},
		}
		extractFilePathsFromTool(tool, state)
		if len(state.filesChanged) != 1 || state.filesChanged[0] != "/src/main.go" {
			t.Errorf("expected [/src/main.go], got %v", state.filesChanged)
		}
	})

	t.Run("path input", func(t *testing.T) {
		state := &transcriptState{
			seenFiles: make(map[string]bool),
		}
		tool := types.ToolCall{
			Name:  "Glob",
			Input: map[string]any{"path": "/src/"},
		}
		extractFilePathsFromTool(tool, state)
		if len(state.filesChanged) != 1 {
			t.Errorf("expected 1 file, got %d", len(state.filesChanged))
		}
	})

	t.Run("deduplicates paths", func(t *testing.T) {
		state := &transcriptState{
			seenFiles: make(map[string]bool),
		}
		tool := types.ToolCall{
			Name:  "Read",
			Input: map[string]any{"file_path": "/src/main.go"},
		}
		extractFilePathsFromTool(tool, state)
		extractFilePathsFromTool(tool, state)
		if len(state.filesChanged) != 1 {
			t.Errorf("expected 1 file after dedup, got %d", len(state.filesChanged))
		}
	})

	t.Run("nil input is safe", func(t *testing.T) {
		state := &transcriptState{seenFiles: make(map[string]bool)}
		tool := types.ToolCall{Name: "test", Input: nil}
		extractFilePathsFromTool(tool, state) // should not panic
	})
}

// ---------------------------------------------------------------------------
// extractIssueRefs
// ---------------------------------------------------------------------------

func TestForge_extractIssueRefs(t *testing.T) {
	state := &transcriptState{seenIssues: make(map[string]bool)}

	extractIssueRefs("Working on ol-0001 and ag-abc", state)
	if len(state.issues) != 2 {
		t.Errorf("expected 2 issues, got %d: %v", len(state.issues), state.issues)
	}

	// Calling again with same IDs should not duplicate
	extractIssueRefs("ol-0001 again", state)
	if len(state.issues) != 2 {
		t.Errorf("expected 2 issues after dedup, got %d", len(state.issues))
	}
}

// ---------------------------------------------------------------------------
// generateSummary
// ---------------------------------------------------------------------------

func TestForge_generateSummary(t *testing.T) {
	t.Run("from decisions", func(t *testing.T) {
		summary := generateSummary([]string{"Use middleware for auth"}, nil, time.Now())
		if summary != "Use middleware for auth" {
			t.Errorf("unexpected summary: %q", summary)
		}
	})

	t.Run("from knowledge when no decisions", func(t *testing.T) {
		summary := generateSummary(nil, []string{"Found a fix for race condition"}, time.Now())
		if summary != "Found a fix for race condition" {
			t.Errorf("unexpected summary: %q", summary)
		}
	})

	t.Run("date fallback when no content", func(t *testing.T) {
		date := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
		summary := generateSummary(nil, nil, date)
		if !strings.Contains(summary, "2026-02-10") {
			t.Errorf("expected date in summary, got %q", summary)
		}
	})

	t.Run("truncates long decisions", func(t *testing.T) {
		long := strings.Repeat("a", 200)
		summary := generateSummary([]string{long}, nil, time.Now())
		if len(summary) > SummaryMaxLength {
			t.Errorf("summary too long: %d > %d", len(summary), SummaryMaxLength)
		}
		if !strings.HasSuffix(summary, "...") {
			t.Error("expected ellipsis on truncated summary")
		}
	})
}

// ---------------------------------------------------------------------------
// countLines
// ---------------------------------------------------------------------------

func TestForge_countLines(t *testing.T) {
	tmp := t.TempDir()

	t.Run("normal file", func(t *testing.T) {
		path := filepath.Join(tmp, "lines.txt")
		if err := os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644); err != nil {
			t.Fatal(err)
		}
		count := countLines(path)
		if count != 3 {
			t.Errorf("countLines = %d, want 3", count)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := filepath.Join(tmp, "empty.txt")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		count := countLines(path)
		if count != 0 {
			t.Errorf("countLines = %d, want 0", count)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		count := countLines("/nonexistent/file.txt")
		if count != 0 {
			t.Errorf("countLines = %d, want 0 for nonexistent file", count)
		}
	})

	t.Run("no trailing newline", func(t *testing.T) {
		path := filepath.Join(tmp, "notail.txt")
		if err := os.WriteFile(path, []byte("line1\nline2"), 0644); err != nil {
			t.Fatal(err)
		}
		count := countLines(path)
		if count != 1 {
			t.Errorf("countLines = %d, want 1 (only one \\n)", count)
		}
	})
}

// ---------------------------------------------------------------------------
// extractIssueIDs
// ---------------------------------------------------------------------------

func TestForge_extractIssueIDs(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
	}{
		{"single issue", "Working on ol-0001", 1},
		{"multiple issues", "Fixed ag-abc and gs-def", 2},
		{"no issues", "No issues here", 0},
		{"issue in sentence", "The issue ol-1234 was resolved", 1},
		{"hyphenated issue ID", "Task gt-abc-def completed", 1},
		{"empty string", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := extractIssueIDs(tt.content)
			if len(ids) != tt.want {
				t.Errorf("extractIssueIDs(%q) returned %d IDs, want %d; ids=%v",
					tt.content, len(ids), tt.want, ids)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// lastSpaceIndex
// ---------------------------------------------------------------------------

func TestForge_lastSpaceIndex(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"has spaces", "hello world foo", 11},
		{"single space", "a b", 1},
		{"no spaces", "nospace", -1},
		{"empty", "", -1},
		{"trailing space", "hello ", 5},
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
// dedup
// ---------------------------------------------------------------------------

func TestForge_dedup(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want int
	}{
		{"no duplicates", []string{"a", "b", "c"}, 3},
		{"with duplicates", []string{"a", "b", "a", "c", "b"}, 3},
		{"all same", []string{"x", "x", "x"}, 1},
		{"empty", []string{}, 0},
		{"nil", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dedup(tt.in)
			if len(got) != tt.want {
				t.Errorf("dedup(%v) returned %d items, want %d", tt.in, len(got), tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// truncateString
// ---------------------------------------------------------------------------

func TestForge_truncateString(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{"short", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"truncated", "hello world!", 8, "hello..."},
		{"empty", "", 10, ""},
		{"maxLen 4", "abcdef", 4, "a..."},
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
// queueForExtraction
// ---------------------------------------------------------------------------

func TestForge_queueForExtraction(t *testing.T) {
	tmp := t.TempDir()

	session := &storage.Session{
		ID:        "test-session-1234567890",
		Summary:   "Test summary",
		Decisions: []string{"decision1"},
		Knowledge: []string{"knowledge1"},
	}

	err := queueForExtraction(session, "/path/session.md", "/path/transcript.jsonl", tmp)
	if err != nil {
		t.Fatalf("queueForExtraction failed: %v", err)
	}

	pendingPath := filepath.Join(tmp, storage.DefaultBaseDir, "pending.jsonl")
	content, err := os.ReadFile(pendingPath)
	if err != nil {
		t.Fatalf("read pending file: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "test-session-1234567890") {
		t.Error("expected session_id in pending output")
	}
	if !strings.Contains(text, "Test summary") {
		t.Error("expected summary in pending output")
	}
	if !strings.Contains(text, "decision1") {
		t.Error("expected decisions in pending output")
	}
}

// ---------------------------------------------------------------------------
// finalizeTranscriptSession
// ---------------------------------------------------------------------------

func TestForge_finalizeTranscriptSession(t *testing.T) {
	session := &storage.Session{}
	state := &transcriptState{
		decisions:    []string{"d1", "d2", "d1"}, // contains duplicate
		knowledge:    []string{"k1"},
		filesChanged: []string{"/a.go", "/b.go"},
		issues:       []string{"ol-001"},
	}

	finalizeTranscriptSession(session, state, 4000)

	if len(session.Decisions) != 2 {
		t.Errorf("expected 2 unique decisions, got %d", len(session.Decisions))
	}
	if len(session.Knowledge) != 1 {
		t.Errorf("expected 1 knowledge, got %d", len(session.Knowledge))
	}
	if len(session.FilesChanged) != 2 {
		t.Errorf("expected 2 files changed, got %d", len(session.FilesChanged))
	}
	if len(session.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(session.Issues))
	}
	if session.Tokens.Total != 4000/CharsPerToken {
		t.Errorf("Tokens.Total = %d, want %d", session.Tokens.Total, 4000/CharsPerToken)
	}
	if !session.Tokens.Estimated {
		t.Error("expected Tokens.Estimated = true")
	}
	if session.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

// ---------------------------------------------------------------------------
// isTranscriptCandidate
// ---------------------------------------------------------------------------

func TestForge_isTranscriptCandidate(t *testing.T) {
	tmp := t.TempDir()
	projectsDir := tmp

	t.Run("valid jsonl file", func(t *testing.T) {
		path := filepath.Join(tmp, "session.jsonl")
		if err := os.WriteFile(path, make([]byte, 200), 0644); err != nil {
			t.Fatal(err)
		}
		info, _ := os.Stat(path)
		if !isTranscriptCandidate(path, info, projectsDir) {
			t.Error("expected valid transcript candidate")
		}
	})

	t.Run("too small file", func(t *testing.T) {
		path := filepath.Join(tmp, "tiny.jsonl")
		if err := os.WriteFile(path, make([]byte, 10), 0644); err != nil {
			t.Fatal(err)
		}
		info, _ := os.Stat(path)
		if isTranscriptCandidate(path, info, projectsDir) {
			t.Error("expected tiny file to be rejected")
		}
	})

	t.Run("non-jsonl file", func(t *testing.T) {
		path := filepath.Join(tmp, "notes.txt")
		if err := os.WriteFile(path, make([]byte, 200), 0644); err != nil {
			t.Fatal(err)
		}
		info, _ := os.Stat(path)
		if isTranscriptCandidate(path, info, projectsDir) {
			t.Error("expected non-jsonl file to be rejected")
		}
	})

	t.Run("directory", func(t *testing.T) {
		subDir := filepath.Join(tmp, "subdir")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}
		info, _ := os.Stat(subDir)
		if isTranscriptCandidate(subDir, info, projectsDir) {
			t.Error("expected directory to be rejected")
		}
	})
}

// ---------------------------------------------------------------------------
// reportProgress
// ---------------------------------------------------------------------------

func TestForge_reportProgress(t *testing.T) {
	t.Run("quiet suppresses output", func(t *testing.T) {
		var buf bytes.Buffer
		last := 0
		reportProgress(true, &buf, 2000, 5000, &last)
		if buf.Len() != 0 {
			t.Error("expected no output in quiet mode")
		}
	})

	t.Run("skips when under threshold", func(t *testing.T) {
		var buf bytes.Buffer
		last := 0
		reportProgress(false, &buf, 500, 5000, &last)
		if buf.Len() != 0 {
			t.Error("expected no output when lineCount - lastProgress < 1000")
		}
	})

	t.Run("reports when over threshold", func(t *testing.T) {
		var buf bytes.Buffer
		last := 0
		reportProgress(false, &buf, 1000, 5000, &last)
		out := buf.String()
		if !strings.Contains(out, "Processing") {
			t.Errorf("expected progress message, got %q", out)
		}
		if last != 1000 {
			t.Errorf("lastProgress not updated: got %d, want 1000", last)
		}
	})
}

// ---------------------------------------------------------------------------
// drainParseErrors
// ---------------------------------------------------------------------------

func TestForge_drainParseErrors(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		ch := make(chan error, 1)
		err := drainParseErrors(ch)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("with error", func(t *testing.T) {
		ch := make(chan error, 1)
		ch <- os.ErrNotExist
		err := drainParseErrors(ch)
		if err == nil {
			t.Error("expected error")
		}
	})
}

// ---------------------------------------------------------------------------
// forgeWarnf
// ---------------------------------------------------------------------------

func TestForge_forgeWarnf_QuietSuppressesOutput(t *testing.T) {
	// forgeWarnf writes to os.Stderr, so we just verify it does not panic
	forgeWarnf(true, "should not appear: %s\n", "test")
}

// ---------------------------------------------------------------------------
// extractToolRefs
// ---------------------------------------------------------------------------

func TestForge_extractToolRefs(t *testing.T) {
	session := &storage.Session{
		ToolCalls: make(map[string]int),
	}
	state := &transcriptState{
		seenFiles: make(map[string]bool),
	}

	tools := []types.ToolCall{
		{Name: "Read", Input: map[string]any{"file_path": "/a.go"}},
		{Name: "Write", Input: map[string]any{"file_path": "/b.go"}},
		{Name: "Read", Input: map[string]any{"file_path": "/a.go"}}, // duplicate
		{Name: "tool_result"},                                        // should be skipped
	}

	extractToolRefs(tools, session, state)

	if session.ToolCalls["Read"] != 2 {
		t.Errorf("Read count = %d, want 2", session.ToolCalls["Read"])
	}
	if session.ToolCalls["Write"] != 1 {
		t.Errorf("Write count = %d, want 1", session.ToolCalls["Write"])
	}
	if _, exists := session.ToolCalls["tool_result"]; exists {
		t.Error("tool_result should not be counted")
	}
	if len(state.filesChanged) != 2 {
		t.Errorf("expected 2 unique files, got %d", len(state.filesChanged))
	}
}
