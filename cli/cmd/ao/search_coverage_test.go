package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// displaySearchResults (0%)
// ---------------------------------------------------------------------------

func TestSearchCov_DisplaySearchResults_Basic(t *testing.T) {
	results := []searchResult{
		{Path: "/path/to/file1.md", Context: "line one\nline two", Type: "session"},
		{Path: "/path/to/file2.md", Context: "", Type: "learning"},
	}

	// Just ensure it doesn't panic. It writes to stdout.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	displaySearchResults("test query", results)

	_ = w.Close()
	os.Stdout = origStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := string(buf[:n])

	if !strings.Contains(out, "2 result(s)") {
		t.Errorf("expected '2 result(s)' in output, got: %s", out)
	}
	if !strings.Contains(out, "test query") {
		t.Errorf("expected 'test query' in output, got: %s", out)
	}
	if !strings.Contains(out, "file1.md") {
		t.Errorf("expected 'file1.md' in output, got: %s", out)
	}
}

func TestSearchCov_DisplaySearchResults_Empty(t *testing.T) {
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	displaySearchResults("empty", []searchResult{})

	_ = w.Close()
	os.Stdout = origStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := string(buf[:n])

	if !strings.Contains(out, "0 result(s)") {
		t.Errorf("expected '0 result(s)' in output, got: %s", out)
	}
}

func TestSearchCov_DisplaySearchResults_WithContext(t *testing.T) {
	results := []searchResult{
		{Path: "/path/to/file.md", Context: "context line 1\ncontext line 2\n", Type: "session"},
	}

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	displaySearchResults("query", results)

	_ = w.Close()
	os.Stdout = origStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := string(buf[:n])

	if !strings.Contains(out, "context line 1") {
		t.Errorf("expected context line in output, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// outputSearchResults — text mode (non-JSON)
// ---------------------------------------------------------------------------

func TestSearchCov_OutputSearchResults_TextMode(t *testing.T) {
	origOutput := output
	output = "table"
	t.Cleanup(func() { output = origOutput })

	results := []searchResult{
		{Path: "/test/file.md", Context: "some context", Type: "session"},
	}

	// Capture stdout
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	err = outputSearchResults("query", results)

	_ = w.Close()
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("outputSearchResults() error = %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := string(buf[:n])

	if !strings.Contains(out, "1 result(s)") {
		t.Errorf("expected '1 result(s)' in text output, got: %s", out)
	}
}

func TestSearchCov_OutputSearchResults_JSONMode(t *testing.T) {
	origOutput := output
	output = "json"
	t.Cleanup(func() { output = origOutput })

	results := []searchResult{
		{Path: "/test/file.md", Context: "ctx", Type: "session"},
	}

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	err = outputSearchResults("query", results)

	_ = w.Close()
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("outputSearchResults() error = %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	out := strings.TrimSpace(string(buf[:n]))

	var parsed []searchResult
	if jsonErr := json.Unmarshal([]byte(out), &parsed); jsonErr != nil {
		t.Fatalf("expected valid JSON, got error: %v\nOutput: %s", jsonErr, out)
	}
	if len(parsed) != 1 {
		t.Errorf("expected 1 result, got %d", len(parsed))
	}
}

// ---------------------------------------------------------------------------
// searchCASS (0%)
// ---------------------------------------------------------------------------

func TestSearchCov_SearchCASS(t *testing.T) {
	tmp := t.TempDir()

	// Create sessions dir
	sessDir := filepath.Join(tmp, "sessions")
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessDir, "session-1.md"), []byte("mutex pattern for concurrency"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create learnings dir with JSONL
	learningsDir := filepath.Join(tmp, "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	learning := map[string]any{
		"id":         "L1",
		"summary":    "mutex pattern for safe access",
		"maturity":   "established",
		"utility":    0.8,
		"confidence": 0.9,
	}
	line, _ := json.Marshal(learning)
	if err := os.WriteFile(filepath.Join(learningsDir, "L1.jsonl"), line, 0644); err != nil {
		t.Fatal(err)
	}

	// Create patterns dir
	patternsDir := filepath.Join(tmp, "patterns")
	if err := os.MkdirAll(patternsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(patternsDir, "mutex.md"), []byte("# Mutex Pattern\n\nUse mutex for concurrency."), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := searchCASS("mutex", sessDir, 10)
	if err != nil {
		t.Fatalf("searchCASS() error = %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result from searchCASS")
	}

	// Results should be sorted by score descending
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("results not sorted by score: [%d]=%f > [%d]=%f", i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

func TestSearchCov_SearchCASS_NoLearnings(t *testing.T) {
	tmp := t.TempDir()
	sessDir := filepath.Join(tmp, "sessions")
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessDir, "s1.md"), []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := searchCASS("hello", sessDir, 10)
	if err != nil {
		t.Fatalf("searchCASS() error = %v", err)
	}
	// Should still work even without learnings/patterns dirs
	if results == nil {
		t.Error("expected non-nil results")
	}
}

func TestSearchCov_SearchCASS_LimitEnforced(t *testing.T) {
	tmp := t.TempDir()
	sessDir := filepath.Join(tmp, "sessions")
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create many files
	for i := 0; i < 10; i++ {
		name := filepath.Join(sessDir, "session-"+string(rune('a'+i))+".md")
		if err := os.WriteFile(name, []byte("searchable content here"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	results, err := searchCASS("searchable", sessDir, 3)
	if err != nil {
		t.Fatalf("searchCASS() error = %v", err)
	}
	if len(results) > 3 {
		t.Errorf("expected at most 3 results, got %d", len(results))
	}
}

// ---------------------------------------------------------------------------
// searchLearningsWithMaturity (0%)
// ---------------------------------------------------------------------------

func TestSearchCov_SearchLearningsWithMaturity(t *testing.T) {
	tmp := t.TempDir()

	// Create JSONL learning
	learning := map[string]any{
		"id":         "L1",
		"summary":    "authentication pattern for services",
		"maturity":   "established",
		"utility":    0.9,
		"confidence": 0.8,
	}
	line, _ := json.Marshal(learning)
	if err := os.WriteFile(filepath.Join(tmp, "auth.jsonl"), line, 0644); err != nil {
		t.Fatal(err)
	}

	// Create MD learning
	if err := os.WriteFile(filepath.Join(tmp, "auth-notes.md"), []byte("authentication notes for services"), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := searchLearningsWithMaturity("authentication", tmp, 10)
	if err != nil {
		t.Fatalf("searchLearningsWithMaturity() error = %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result")
	}
}

func TestSearchCov_SearchLearningsWithMaturity_NoMatch(t *testing.T) {
	tmp := t.TempDir()
	learning := map[string]any{"id": "L1", "summary": "unrelated content"}
	line, _ := json.Marshal(learning)
	if err := os.WriteFile(filepath.Join(tmp, "other.jsonl"), line, 0644); err != nil {
		t.Fatal(err)
	}

	results, err := searchLearningsWithMaturity("nonexistent_xyz", tmp, 10)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchCov_SearchLearningsWithMaturity_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	results, err := searchLearningsWithMaturity("test", tmp, 10)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// ---------------------------------------------------------------------------
// truncateContext
// ---------------------------------------------------------------------------

func TestSearchCov_TruncateContext_Short(t *testing.T) {
	got := truncateContext("short")
	if got != "short" {
		t.Errorf("expected 'short', got %q", got)
	}
}

func TestSearchCov_TruncateContext_Long(t *testing.T) {
	long := strings.Repeat("x", ContextLineMaxLength+50)
	got := truncateContext(long)
	if len(got) != ContextLineMaxLength+3 { // +3 for "..."
		t.Errorf("expected length %d, got %d", ContextLineMaxLength+3, len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Error("expected '...' suffix")
	}
}

// ---------------------------------------------------------------------------
// parseLearningMatch
// ---------------------------------------------------------------------------

func TestSearchCov_ParseLearningMatch_Valid(t *testing.T) {
	data := map[string]any{
		"summary":  "test learning",
		"maturity": "candidate",
		"utility":  0.7,
	}
	line, _ := json.Marshal(data)

	result, ok := parseLearningMatch(string(line), "/path/to/file.jsonl")
	if !ok {
		t.Error("expected ok=true")
	}
	if result.Type != "learning" {
		t.Errorf("type = %q, want learning", result.Type)
	}
	if !strings.Contains(result.Context, "[candidate]") {
		t.Errorf("expected '[candidate]' in context, got: %s", result.Context)
	}
}

func TestSearchCov_ParseLearningMatch_InvalidJSON(t *testing.T) {
	_, ok := parseLearningMatch("not json", "/path/file.jsonl")
	if ok {
		t.Error("expected ok=false for invalid JSON")
	}
}

// ---------------------------------------------------------------------------
// extractLearningContext
// ---------------------------------------------------------------------------

func TestSearchCov_ExtractLearningContext_Summary(t *testing.T) {
	data := map[string]any{"summary": "test summary"}
	got := extractLearningContext(data)
	if got != "test summary" {
		t.Errorf("expected 'test summary', got %q", got)
	}
}

func TestSearchCov_ExtractLearningContext_Content(t *testing.T) {
	data := map[string]any{"content": "test content"}
	got := extractLearningContext(data)
	if got != "test content" {
		t.Errorf("expected 'test content', got %q", got)
	}
}

func TestSearchCov_ExtractLearningContext_Neither(t *testing.T) {
	data := map[string]any{"id": "L1"}
	got := extractLearningContext(data)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// maturityToWeight
// ---------------------------------------------------------------------------

func TestSearchCov_MaturityToWeight(t *testing.T) {
	tests := []struct {
		data map[string]any
		want float64
	}{
		{map[string]any{"maturity": "established"}, 1.5},
		{map[string]any{"maturity": "candidate"}, 1.2},
		{map[string]any{"maturity": "provisional"}, 1.0},
		{map[string]any{"maturity": "anti-pattern"}, 0.3},
		{map[string]any{"maturity": "unknown"}, 1.0},
		{map[string]any{}, 1.0},
	}
	for _, tt := range tests {
		got := maturityToWeight(tt.data)
		if got != tt.want {
			t.Errorf("maturityToWeight(%v) = %v, want %v", tt.data, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// parseJSONLMatch
// ---------------------------------------------------------------------------

func TestSearchCov_ParseJSONLMatch_WithSummary(t *testing.T) {
	data := map[string]any{"summary": "test summary", "id": "L1"}
	line, _ := json.Marshal(data)
	result, ok := parseJSONLMatch(string(line), "/path/file.jsonl")
	if !ok {
		t.Error("expected ok=true")
	}
	if result.Context != "test summary" {
		t.Errorf("expected 'test summary', got %q", result.Context)
	}
}

func TestSearchCov_ParseJSONLMatch_LongSummary(t *testing.T) {
	long := strings.Repeat("x", ContextLineMaxLength+50)
	data := map[string]any{"summary": long}
	line, _ := json.Marshal(data)
	result, ok := parseJSONLMatch(string(line), "/path/file.jsonl")
	if !ok {
		t.Error("expected ok=true")
	}
	if len(result.Context) > ContextLineMaxLength+3 {
		t.Errorf("expected truncated context, got length %d", len(result.Context))
	}
}

func TestSearchCov_ParseJSONLMatch_NoSummary(t *testing.T) {
	data := map[string]any{"id": "L1", "content": "some content"}
	line, _ := json.Marshal(data)
	result, ok := parseJSONLMatch(string(line), "/path/file.jsonl")
	if !ok {
		t.Error("expected ok=true")
	}
	if result.Context != "" {
		t.Errorf("expected empty context when no summary, got %q", result.Context)
	}
}

// ---------------------------------------------------------------------------
// selectAndSearch — file-based default path
// ---------------------------------------------------------------------------

func TestSearchCov_SelectAndSearch_FileBased(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "test.md"), []byte("searchable content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Reset search flags
	origUseSC := searchUseSC
	origUseCASS := searchUseCASS
	searchUseSC = false
	searchUseCASS = false
	t.Cleanup(func() {
		searchUseSC = origUseSC
		searchUseCASS = origUseCASS
	})

	results, err := selectAndSearch("searchable", tmp, 10)
	if err != nil {
		t.Fatalf("selectAndSearch() error = %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result")
	}
}
