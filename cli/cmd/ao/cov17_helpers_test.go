package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// dedup.go — pure/helper functions (0% → higher)
// ---------------------------------------------------------------------------

func TestCov17_extractMarkdownBody_noFrontmatter(t *testing.T) {
	text := "# Hello\nThis is content.\n"
	got := extractMarkdownBody(text)
	if got != text {
		t.Errorf("extractMarkdownBody (no frontmatter): got %q, want %q", got, text)
	}
}

func TestCov17_extractMarkdownBody_withFrontmatter(t *testing.T) {
	text := "---\ntitle: test\n---\n# Body here\n"
	got := extractMarkdownBody(text)
	if !strings.Contains(got, "# Body here") {
		t.Errorf("extractMarkdownBody: expected body after frontmatter, got %q", got)
	}
	if strings.Contains(got, "title: test") {
		t.Errorf("extractMarkdownBody: frontmatter leaked into body: %q", got)
	}
}

func TestCov17_extractMarkdownBody_noClosingDelimiter(t *testing.T) {
	text := "---\ntitle: test\n# No closing delimiter\n"
	got := extractMarkdownBody(text)
	// No closing ---, entire content returned
	if got != text {
		t.Errorf("extractMarkdownBody (no closing): got %q, want %q", got, text)
	}
}

func TestCov17_extractJSONLBody_contentField(t *testing.T) {
	text := `{"content":"This is the content","title":"unused"}` + "\n"
	got := extractJSONLBody(text)
	if got != "This is the content" {
		t.Errorf("extractJSONLBody content: got %q", got)
	}
}

func TestCov17_extractJSONLBody_titleFallback(t *testing.T) {
	text := `{"title":"Use the title","other":"ignored"}` + "\n"
	got := extractJSONLBody(text)
	if got != "Use the title" {
		t.Errorf("extractJSONLBody title fallback: got %q", got)
	}
}

func TestCov17_extractJSONLBody_empty(t *testing.T) {
	text := `{"other":"no content or title"}` + "\n"
	got := extractJSONLBody(text)
	if got != "" {
		t.Errorf("extractJSONLBody empty: got %q", got)
	}
}

func TestCov17_extractJSONLBody_invalidJSON(t *testing.T) {
	text := "not json at all\n"
	got := extractJSONLBody(text)
	if got != "" {
		t.Errorf("extractJSONLBody invalid JSON: got %q", got)
	}
}

func TestCov17_hashNormalizedContent_basic(t *testing.T) {
	h1 := hashNormalizedContent("Hello World")
	h2 := hashNormalizedContent("hello world")
	h3 := hashNormalizedContent("**Hello** World")
	// Normalized: lowercase + strip markdown → all produce same hash
	if h1 != h2 {
		t.Errorf("hashNormalizedContent: case should be normalized, got h1=%s h2=%s", h1, h2)
	}
	if h1 != h3 {
		t.Errorf("hashNormalizedContent: markdown stripping should normalize, got h1=%s h3=%s", h1, h3)
	}
}

func TestCov17_readUtilityFromFrontmatter_noFrontmatter(t *testing.T) {
	text := "# Hello\nNo frontmatter.\n"
	got := readUtilityFromFrontmatter(text, 0.5)
	if got != 0.5 {
		t.Errorf("readUtilityFromFrontmatter (no frontmatter): got %.2f, want 0.5", got)
	}
}

func TestCov17_readUtilityFromFrontmatter_withUtility(t *testing.T) {
	text := "---\nutility: 0.9\ntitle: test\n---\n# Content\n"
	got := readUtilityFromFrontmatter(text, 0.5)
	if got < 0.89 || got > 0.91 {
		t.Errorf("readUtilityFromFrontmatter: got %.4f, want ~0.9", got)
	}
}

func TestCov17_readUtilityFromFrontmatter_noUtilityField(t *testing.T) {
	text := "---\ntitle: test\ndate: 2026-01-01\n---\n# Content\n"
	got := readUtilityFromFrontmatter(text, 0.5)
	if got != 0.5 {
		t.Errorf("readUtilityFromFrontmatter (no utility): got %.2f, want 0.5", got)
	}
}

func TestCov17_readUtilityFromJSONL_float64(t *testing.T) {
	text := `{"utility":0.8,"title":"test"}` + "\n"
	got := readUtilityFromJSONL(text, 0.5)
	if got < 0.79 || got > 0.81 {
		t.Errorf("readUtilityFromJSONL float64: got %.4f, want ~0.8", got)
	}
}

func TestCov17_readUtilityFromJSONL_stringValue(t *testing.T) {
	text := `{"utility":"0.75","title":"test"}` + "\n"
	got := readUtilityFromJSONL(text, 0.5)
	if got < 0.74 || got > 0.76 {
		t.Errorf("readUtilityFromJSONL string: got %.4f, want ~0.75", got)
	}
}

func TestCov17_readUtilityFromJSONL_noUtility(t *testing.T) {
	text := `{"title":"no utility field"}` + "\n"
	got := readUtilityFromJSONL(text, 0.5)
	if got != 0.5 {
		t.Errorf("readUtilityFromJSONL (no utility): got %.2f, want 0.5", got)
	}
}

func TestCov17_readUtilityFromJSONL_invalidJSON(t *testing.T) {
	text := "not valid json\n"
	got := readUtilityFromJSONL(text, 0.5)
	if got != 0.5 {
		t.Errorf("readUtilityFromJSONL (invalid JSON): got %.2f, want 0.5", got)
	}
}

func TestCov17_extractLearningBody_mdFile(t *testing.T) {
	tmp := t.TempDir()
	content := "---\ntitle: test\n---\n# Body content here\n"
	path := filepath.Join(tmp, "learning.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := extractLearningBody(path)
	if !strings.Contains(got, "Body content here") {
		t.Errorf("extractLearningBody md: got %q", got)
	}
}

func TestCov17_extractLearningBody_jsonlFile(t *testing.T) {
	tmp := t.TempDir()
	content := `{"content":"Learning content","title":"test"}` + "\n"
	path := filepath.Join(tmp, "learning.jsonl")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := extractLearningBody(path)
	if got != "Learning content" {
		t.Errorf("extractLearningBody jsonl: got %q", got)
	}
}

func TestCov17_extractLearningBody_nonexistentFile(t *testing.T) {
	got := extractLearningBody("/nonexistent/path/learning.md")
	if got != "" {
		t.Errorf("extractLearningBody (nonexistent): got %q, want empty", got)
	}
}

func TestCov17_collectDedupFiles_noDirs(t *testing.T) {
	tmp := t.TempDir()
	files, err := collectDedupFiles(tmp)
	if err != nil {
		t.Fatalf("collectDedupFiles noDirs: %v", err)
	}
	if files != nil {
		t.Errorf("collectDedupFiles (no dirs): got %v, want nil", files)
	}
}

func TestCov17_collectDedupFiles_withLearnings(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create one .md and one .jsonl
	if err := os.WriteFile(filepath.Join(learningsDir, "test.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "test.jsonl"), []byte(`{"title":"test"}`+"\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	files, err := collectDedupFiles(tmp)
	if err != nil {
		t.Fatalf("collectDedupFiles withLearnings: %v", err)
	}
	if len(files) < 2 {
		t.Errorf("collectDedupFiles: got %d files, want >= 2", len(files))
	}
}

func TestCov17_buildDedupResult_withDuplicates(t *testing.T) {
	h1 := hashNormalizedContent("content alpha")
	h2 := hashNormalizedContent("content beta")
	hashToFiles := map[string][]string{
		h1: {"/a/file1.md", "/a/file2.md"},
		h2: {"/a/file3.md"},
	}
	result := buildDedupResult(hashToFiles, 3, "/a")
	if result.TotalFiles != 3 {
		t.Errorf("TotalFiles: got %d, want 3", result.TotalFiles)
	}
	if result.UniqueContent != 2 {
		t.Errorf("UniqueContent: got %d, want 2", result.UniqueContent)
	}
	if result.DuplicateGroups != 1 {
		t.Errorf("DuplicateGroups: got %d, want 1", result.DuplicateGroups)
	}
	if result.DuplicateFiles != 2 {
		t.Errorf("DuplicateFiles: got %d, want 2", result.DuplicateFiles)
	}
}

func TestCov17_runDedup_emptyDir(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	cmd := &cobra.Command{}
	err := runDedup(cmd, nil)
	if err != nil {
		t.Fatalf("runDedup empty dir: %v", err)
	}
}

func TestCov17_runDedup_withFiles(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Two files with identical content → duplicates
	content := "---\ntitle: same\n---\n# Same body content here\n"
	if err := os.WriteFile(filepath.Join(learningsDir, "a.md"), []byte(content), 0644); err != nil {
		t.Fatalf("write a.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "b.md"), []byte(content), 0644); err != nil {
		t.Fatalf("write b.md: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "" // text mode

	cmd := &cobra.Command{}
	err := runDedup(cmd, nil)
	if err != nil {
		t.Fatalf("runDedup with files: %v", err)
	}
}

func TestCov17_runDedup_jsonOutput(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "a.md"), []byte("# Only one file\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "json"

	cmd := &cobra.Command{}
	err := runDedup(cmd, nil)
	if err != nil {
		t.Fatalf("runDedup json output: %v", err)
	}
}

// ---------------------------------------------------------------------------
// inject.go — helper functions (0% → higher)
// ---------------------------------------------------------------------------

func TestCov17_resortLearnings_basic(t *testing.T) {
	learnings := []learning{
		{ID: "a", CompositeScore: 0.3},
		{ID: "b", CompositeScore: 0.9},
		{ID: "c", CompositeScore: 0.6},
	}
	resortLearnings(learnings)
	if learnings[0].ID != "b" {
		t.Errorf("resortLearnings: first should be 'b' (highest), got %q", learnings[0].ID)
	}
	if learnings[2].ID != "a" {
		t.Errorf("resortLearnings: last should be 'a' (lowest), got %q", learnings[2].ID)
	}
}

func TestCov17_trimToCharBudget_withinBudget(t *testing.T) {
	text := "Short text"
	got := trimToCharBudget(text, 1000)
	if got != text {
		t.Errorf("trimToCharBudget (within): got %q, want %q", got, text)
	}
}

func TestCov17_trimToCharBudget_exceedsBudget(t *testing.T) {
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = strings.Repeat("x", 50)
	}
	text := strings.Join(lines, "\n")
	budget := 200
	got := trimToCharBudget(text, budget)
	if len(got) > budget+100 { // some slack for truncation marker
		t.Errorf("trimToCharBudget: output too long: got %d chars, budget %d", len(got), budget)
	}
	if !strings.Contains(got, "truncated") {
		t.Errorf("trimToCharBudget: expected truncation marker in output")
	}
}

func TestCov17_atomicWriteFile_success(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "output.txt")
	data := []byte("hello atomic")
	if err := atomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("atomicWriteFile: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after atomicWriteFile: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("atomicWriteFile: got %q, want %q", got, data)
	}
}

func TestCov17_atomicWriteFile_invalidDir(t *testing.T) {
	// Writing to a nonexistent directory should fail
	err := atomicWriteFile("/nonexistent/directory/output.txt", []byte("data"), 0644)
	if err == nil {
		t.Error("atomicWriteFile: expected error for nonexistent directory, got nil")
	}
}

func TestCov17_findAgentsSubdir_noMatch(t *testing.T) {
	tmp := t.TempDir()
	result := findAgentsSubdir(tmp, "learnings")
	if result != "" {
		t.Errorf("findAgentsSubdir (no match): got %q, want empty", result)
	}
}

func TestCov17_findAgentsSubdir_withMatch(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	result := findAgentsSubdir(tmp, "learnings")
	if result == "" {
		t.Error("findAgentsSubdir (with match): expected non-empty result")
	}
}

func TestCov17_writeLearningsSection_empty(t *testing.T) {
	var sb strings.Builder
	writeLearningsSection(&sb, nil)
	if sb.Len() != 0 {
		t.Errorf("writeLearningsSection empty: expected nothing written, got %q", sb.String())
	}
}

func TestCov17_writeLearningsSection_withItems(t *testing.T) {
	var sb strings.Builder
	learnings := []learning{
		{ID: "learn-001", Title: "Test Learning", Summary: "Summary here"},
	}
	writeLearningsSection(&sb, learnings)
	if !strings.Contains(sb.String(), "learn-001") {
		t.Errorf("writeLearningsSection: expected ID in output, got %q", sb.String())
	}
}

func TestCov17_writePatternsSection_empty(t *testing.T) {
	var sb strings.Builder
	writePatternsSection(&sb, nil)
	if sb.Len() != 0 {
		t.Errorf("writePatternsSection empty: expected nothing written")
	}
}

func TestCov17_writePatternsSection_withItems(t *testing.T) {
	var sb strings.Builder
	patterns := []pattern{
		{Name: "my-pattern", Description: "A great pattern"},
		{Name: "no-desc-pattern"}, // No description branch
	}
	writePatternsSection(&sb, patterns)
	if !strings.Contains(sb.String(), "my-pattern") {
		t.Errorf("writePatternsSection: expected name in output, got %q", sb.String())
	}
}

func TestCov17_writeSessionsSection_empty(t *testing.T) {
	var sb strings.Builder
	writeSessionsSection(&sb, nil)
	if sb.Len() != 0 {
		t.Errorf("writeSessionsSection empty: expected nothing written")
	}
}

func TestCov17_writeSessionsSection_withItems(t *testing.T) {
	var sb strings.Builder
	sessions := []session{
		{Date: "2026-01-01", Summary: "Test session"},
	}
	writeSessionsSection(&sb, sessions)
	if !strings.Contains(sb.String(), "2026-01-01") {
		t.Errorf("writeSessionsSection: expected date in output, got %q", sb.String())
	}
}

func TestCov17_writePredecessorSection_nil(t *testing.T) {
	var sb strings.Builder
	writePredecessorSection(&sb, nil)
	if sb.Len() != 0 {
		t.Errorf("writePredecessorSection nil: expected nothing written")
	}
}

func TestCov17_writePredecessorSection_allFields(t *testing.T) {
	var sb strings.Builder
	pred := &predecessorContext{
		SessionAge: "2h",
		WorkingOn:  "coverage sprint",
		Progress:   "84.3%",
		Blocker:    "none",
		NextStep:   "write cov17",
	}
	writePredecessorSection(&sb, pred)
	out := sb.String()
	if !strings.Contains(out, "coverage sprint") {
		t.Errorf("writePredecessorSection: expected WorkingOn in output, got %q", out)
	}
}

func TestCov17_writePredecessorSection_rawSummaryFallback(t *testing.T) {
	var sb strings.Builder
	pred := &predecessorContext{
		RawSummary: "Raw summary text",
		// No Progress → RawSummary fallback executes
	}
	writePredecessorSection(&sb, pred)
	if !strings.Contains(sb.String(), "Raw summary text") {
		t.Errorf("writePredecessorSection RawSummary: expected raw summary in output, got %q", sb.String())
	}
}

// ---------------------------------------------------------------------------
// lookup.go — pure helper functions (0% → higher)
// ---------------------------------------------------------------------------

func TestCov17_matchesID_exactMatch(t *testing.T) {
	if !matchesID("learn-001", "", "learn-001") {
		t.Error("matchesID exact: expected true")
	}
}

func TestCov17_matchesID_caseInsensitive(t *testing.T) {
	if !matchesID("LEARN-001", "", "learn-001") {
		t.Error("matchesID case insensitive: expected true")
	}
}

func TestCov17_matchesID_filenameMatch(t *testing.T) {
	if !matchesID("", "/path/to/learn-001.md", "learn-001") {
		t.Error("matchesID filename: expected true")
	}
}

func TestCov17_matchesID_containsMatch(t *testing.T) {
	if !matchesID("", "/path/to/2026-01-01-learn-001-test.md", "learn-001") {
		t.Error("matchesID contains: expected true")
	}
}

func TestCov17_matchesID_noMatch(t *testing.T) {
	if matchesID("other-id", "/path/other.md", "learn-001") {
		t.Error("matchesID no match: expected false")
	}
}

func TestCov17_filterByBead_matching(t *testing.T) {
	learnings := []learning{
		{ID: "l1", SourceBead: "ag-abc"},
		{ID: "l2", SourceBead: "ag-xyz"},
		{ID: "l3", SourceBead: "ag-abc"},
	}
	result := filterByBead(learnings, "ag-abc")
	if len(result) != 2 {
		t.Errorf("filterByBead: got %d, want 2", len(result))
	}
}

func TestCov17_filterByBead_noMatch(t *testing.T) {
	learnings := []learning{
		{ID: "l1", SourceBead: "ag-xyz"},
	}
	result := filterByBead(learnings, "ag-abc")
	if len(result) != 0 {
		t.Errorf("filterByBead no match: got %d, want 0", len(result))
	}
}

func TestCov17_formatLookupAge_lessThanDay(t *testing.T) {
	got := formatLookupAge(0.05)
	if got != "<1d" {
		t.Errorf("formatLookupAge <1d: got %q, want '<1d'", got)
	}
}

func TestCov17_formatLookupAge_days(t *testing.T) {
	got := formatLookupAge(0.5) // ~3.5 days → days format
	if !strings.HasSuffix(got, "d") {
		t.Errorf("formatLookupAge days: got %q, expected suffix 'd'", got)
	}
}

func TestCov17_formatLookupAge_weeks(t *testing.T) {
	got := formatLookupAge(2.0) // 2 weeks → weeks format
	if !strings.HasSuffix(got, "w") {
		t.Errorf("formatLookupAge weeks: got %q, expected suffix 'w'", got)
	}
}

func TestCov17_formatLookupAge_months(t *testing.T) {
	got := formatLookupAge(8.0) // 8 weeks = ~56 days → months format
	if !strings.HasSuffix(got, "mo") {
		t.Errorf("formatLookupAge months: got %q, expected suffix 'mo'", got)
	}
}

func TestCov17_runLookup_noArgsNoFlags(t *testing.T) {
	origQuery := lookupQuery
	origBead := lookupBead
	defer func() {
		lookupQuery = origQuery
		lookupBead = origBead
	}()
	lookupQuery = ""
	lookupBead = ""

	cmd := lookupCmd
	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("runLookup no args: expected error, got nil")
	}
}

func TestCov17_outputResults_empty(t *testing.T) {
	origJSON := lookupJSON
	defer func() { lookupJSON = origJSON }()
	lookupJSON = false

	err := outputResults("/tmp", nil, nil)
	if err != nil {
		t.Fatalf("outputResults empty: %v", err)
	}
}

func TestCov17_outputResults_withLearnings(t *testing.T) {
	origJSON := lookupJSON
	defer func() { lookupJSON = origJSON }()
	lookupJSON = false

	learnings := []learning{
		{ID: "l1", Title: "Test Learning", Summary: "A test summary", Source: "/tmp/test.md"},
	}
	err := outputResults("/tmp", learnings, nil)
	if err != nil {
		t.Fatalf("outputResults with learnings: %v", err)
	}
}

func TestCov17_outputResults_jsonMode(t *testing.T) {
	origJSON := lookupJSON
	defer func() { lookupJSON = origJSON }()
	lookupJSON = true

	learnings := []learning{
		{ID: "l1", Title: "Test"},
	}
	err := outputResults("/tmp", learnings, nil)
	if err != nil {
		t.Fatalf("outputResults json: %v", err)
	}
}
