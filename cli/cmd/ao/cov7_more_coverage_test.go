package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// inject.go — atomicWriteFile (47.4% coverage)
// ---------------------------------------------------------------------------

func TestCov7_atomicWriteFile_happyPath(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "output.txt")
	data := []byte("hello world")

	if err := atomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("atomicWriteFile: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("atomicWriteFile wrote %q, want %q", got, data)
	}
}

func TestCov7_atomicWriteFile_noPermission(t *testing.T) {
	// Try writing to a nonexistent parent directory to trigger error
	err := atomicWriteFile("/nonexistent/path/file.txt", []byte("data"), 0644)
	if err == nil {
		t.Error("expected error writing to nonexistent directory")
	}
}

func TestCov7_atomicWriteFile_emptyData(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.txt")
	if err := atomicWriteFile(path, []byte{}, 0644); err != nil {
		t.Fatalf("atomicWriteFile empty: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("expected empty file, got size %d", info.Size())
	}
}

// ---------------------------------------------------------------------------
// context_assemble.go — extractIntelJSONContent (77.8% coverage)
// ---------------------------------------------------------------------------

func TestCov7_extractIntelJSONContent_invalidJSON(t *testing.T) {
	got := extractIntelJSONContent([]byte("not json at all"))
	if got != "not json at all" {
		t.Errorf("extractIntelJSONContent(invalid) = %q, want raw string", got)
	}
}

func TestCov7_extractIntelJSONContent_withContentKey(t *testing.T) {
	data, _ := json.Marshal(map[string]any{"content": "the content", "title": "ignored"})
	got := extractIntelJSONContent(data)
	if got != "the content" {
		t.Errorf("extractIntelJSONContent with content key = %q, want %q", got, "the content")
	}
}

func TestCov7_extractIntelJSONContent_withPatternKey(t *testing.T) {
	data, _ := json.Marshal(map[string]any{"pattern": "the pattern"})
	got := extractIntelJSONContent(data)
	if got != "the pattern" {
		t.Errorf("extractIntelJSONContent with pattern key = %q, want %q", got, "the pattern")
	}
}

func TestCov7_extractIntelJSONContent_withSummaryKey(t *testing.T) {
	data, _ := json.Marshal(map[string]any{"summary": "the summary"})
	got := extractIntelJSONContent(data)
	if got != "the summary" {
		t.Errorf("extractIntelJSONContent with summary key = %q, want %q", got, "the summary")
	}
}

func TestCov7_extractIntelJSONContent_noMatchingKey(t *testing.T) {
	// JSON with no matching keys falls back to the raw JSON string
	data, _ := json.Marshal(map[string]any{"foo": "bar"})
	got := extractIntelJSONContent(data)
	if got == "" {
		t.Error("expected non-empty fallback for JSON without matching keys")
	}
}

func TestCov7_extractIntelJSONContent_nullContent(t *testing.T) {
	// content key is present but null — should fall through to summary
	data, _ := json.Marshal(map[string]any{"content": nil, "summary": "fallback summary"})
	got := extractIntelJSONContent(data)
	if got != "fallback summary" {
		t.Errorf("extractIntelJSONContent with null content = %q, want %q", got, "fallback summary")
	}
}

// ---------------------------------------------------------------------------
// inject_predecessor.go — truncatePredecessor (37.5% coverage)
// ---------------------------------------------------------------------------

func TestCov7_truncatePredecessor_underBudget(t *testing.T) {
	ctx := &predecessorContext{
		WorkingOn: "short",
		Progress:  "quick note",
	}
	// Copy original
	origWorking := ctx.WorkingOn
	origProgress := ctx.Progress

	truncatePredecessor(ctx)

	// Should not be modified since total is small
	if ctx.WorkingOn != origWorking {
		t.Errorf("WorkingOn modified when under budget: %q -> %q", origWorking, ctx.WorkingOn)
	}
	if ctx.Progress != origProgress {
		t.Errorf("Progress modified when under budget: %q -> %q", origProgress, ctx.Progress)
	}
}

func TestCov7_truncatePredecessor_overBudget(t *testing.T) {
	// Create a predecessor that exceeds the budget (200 tokens * 4 chars = 800 chars)
	longText := strings.Repeat("x", 300)
	ctx := &predecessorContext{
		WorkingOn:  longText,
		Progress:   longText,
		Blocker:    longText,
		NextStep:   longText,
		RawSummary: longText,
	}

	truncatePredecessor(ctx)

	// All fields should have been truncated
	if len(ctx.WorkingOn) > 100+len("...") {
		t.Errorf("WorkingOn not truncated: len=%d", len(ctx.WorkingOn))
	}
	if len(ctx.Progress) > 250+len("...") {
		t.Errorf("Progress not truncated: len=%d", len(ctx.Progress))
	}
	if len(ctx.Blocker) > 200+len("...") {
		t.Errorf("Blocker not truncated: len=%d", len(ctx.Blocker))
	}
}

// ---------------------------------------------------------------------------
// forge.go — updateSearchIndexForFile (20% coverage)
// ---------------------------------------------------------------------------

func TestCov7_updateSearchIndexForFile_noIndex(t *testing.T) {
	tmp := t.TempDir()
	// No index file exists — function should return early with no error
	updateSearchIndexForFile(tmp, filepath.Join(tmp, "file.md"), true)
	// Verify no index file was created as a side effect
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty dir after no-index update, got %d entries", len(entries))
	}
}

// ---------------------------------------------------------------------------
// extract.go — outputExtractDryRun (38.5% coverage)
// ---------------------------------------------------------------------------

func TestCov7_outputExtractDryRun_textOutput(t *testing.T) {
	origOutput := output
	defer func() { output = origOutput }()
	output = "" // text mode

	pending := []PendingExtraction{
		{SessionID: "sess-1", Summary: "summary 1"},
		{SessionID: "sess-2", Summary: "summary 2"},
	}

	err := outputExtractDryRun(pending)
	if err != nil {
		t.Fatalf("outputExtractDryRun: %v", err)
	}
}

func TestCov7_outputExtractDryRun_jsonOutput(t *testing.T) {
	origOutput := output
	defer func() { output = origOutput }()
	output = "json"

	pending := []PendingExtraction{
		{SessionID: "sess-abc", Summary: "test summary"},
	}

	err := outputExtractDryRun(pending)
	if err != nil {
		t.Fatalf("outputExtractDryRun json: %v", err)
	}
}

func TestCov7_outputExtractDryRun_empty(t *testing.T) {
	origOutput := output
	defer func() { output = origOutput }()
	output = ""

	err := outputExtractDryRun(nil)
	if err != nil {
		t.Fatalf("outputExtractDryRun empty: %v", err)
	}
}

// ---------------------------------------------------------------------------
// inject.go — writePredecessorSection with SessionAge empty (covers missing branch)
// ---------------------------------------------------------------------------

func TestCov7_writePredecessorSection_noSessionAge(t *testing.T) {
	pred := &predecessorContext{
		WorkingOn: "task with no session age",
	}
	var sb strings.Builder
	writePredecessorSection(&sb, pred)
	out := sb.String()

	// Should contain "Predecessor Context" but no "(ago)" since SessionAge is empty
	if !strings.Contains(out, "Predecessor Context") {
		t.Errorf("expected 'Predecessor Context' in output, got %q", out)
	}
	if strings.Contains(out, "ago") {
		t.Errorf("unexpected 'ago' when SessionAge is empty, got %q", out)
	}
}
