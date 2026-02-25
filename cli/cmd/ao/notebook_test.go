package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFindMemoryFile_StandardPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Create the expected Claude Code project path structure
	cwd := "/Users/testuser/projects/myrepo"
	normalizedPath := strings.ReplaceAll(cwd, "/", "-")
	memoryDir := filepath.Join(tmp, ".claude", "projects", normalizedPath, "memory")
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		t.Fatal(err)
	}
	memoryFile := filepath.Join(memoryDir, "MEMORY.md")
	if err := os.WriteFile(memoryFile, []byte("# Test Memory\n"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := findMemoryFile(cwd)
	if err != nil {
		t.Fatalf("findMemoryFile() error: %v", err)
	}
	if got != memoryFile {
		t.Errorf("findMemoryFile() = %q, want %q", got, memoryFile)
	}
}

func TestFindMemoryFile_FallbackSearch(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Create a project dir that contains the last path component
	memoryDir := filepath.Join(tmp, ".claude", "projects", "-Users-someone-projects-myrepo", "memory")
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		t.Fatal(err)
	}
	memoryFile := filepath.Join(memoryDir, "MEMORY.md")
	if err := os.WriteFile(memoryFile, []byte("# Test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Use a different cwd that won't match exactly but has "myrepo" as last component
	got, err := findMemoryFile("/different/path/to/myrepo")
	if err != nil {
		t.Fatalf("findMemoryFile() error: %v", err)
	}
	if got != memoryFile {
		t.Errorf("findMemoryFile() = %q, want %q", got, memoryFile)
	}
}

func TestFindMemoryFile_NotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	// Create the projects dir but no memory files
	if err := os.MkdirAll(filepath.Join(tmp, ".claude", "projects"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := findMemoryFile("/nonexistent/repo")
	if err == nil {
		t.Error("expected error when no MEMORY.md exists, got nil")
	}
}

func TestParseNotebookSections(t *testing.T) {
	content := `# My Memory

## Section One
- item 1
- item 2

## Section Two
- item A
- item B

## Section Three
- item X
`

	sections := parseSectionsFromString(content)

	if len(sections) != 4 {
		t.Fatalf("expected 4 sections, got %d", len(sections))
	}

	if sections[0].Heading != "# My Memory" {
		t.Errorf("section 0 heading = %q, want %q", sections[0].Heading, "# My Memory")
	}
	if sections[1].Heading != "## Section One" {
		t.Errorf("section 1 heading = %q, want %q", sections[1].Heading, "## Section One")
	}
	if sections[2].Heading != "## Section Two" {
		t.Errorf("section 2 heading = %q, want %q", sections[2].Heading, "## Section Two")
	}
	if sections[3].Heading != "## Section Three" {
		t.Errorf("section 3 heading = %q, want %q", sections[3].Heading, "## Section Three")
	}
}

func TestUpdateLastSessionSection(t *testing.T) {
	sections := []notebookSection{
		{Heading: "# Memory", Lines: []string{""}},
		{Heading: "## Last Session", Lines: []string{"- old data", ""}},
		{Heading: "## Patterns", Lines: []string{"- pattern 1", ""}},
	}

	newSession := notebookSection{
		Heading: "## Last Session",
		Lines:   []string{"- **Date:** 2026-02-25", "- **Summary:** new summary", ""},
	}

	result := upsertLastSession(sections, newSession)

	if len(result) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(result))
	}
	if result[1].Lines[0] != "- **Date:** 2026-02-25" {
		t.Errorf("Last Session not updated: %v", result[1].Lines)
	}
}

func TestUpdateLastSessionSection_Insert(t *testing.T) {
	sections := []notebookSection{
		{Heading: "# Memory", Lines: []string{""}},
		{Heading: "## Patterns", Lines: []string{"- pattern 1", ""}},
	}

	newSession := notebookSection{
		Heading: "## Last Session",
		Lines:   []string{"- **Date:** 2026-02-25", ""},
	}

	result := upsertLastSession(sections, newSession)

	if len(result) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(result))
	}
	if result[1].Heading != "## Last Session" {
		t.Errorf("Last Session not inserted at position 1: heading = %q", result[1].Heading)
	}
	if result[2].Heading != "## Patterns" {
		t.Errorf("Patterns should be at position 2: heading = %q", result[2].Heading)
	}
}

func TestPruneNotebook_UnderLimit(t *testing.T) {
	sections := []notebookSection{
		{Heading: "# Memory", Lines: []string{""}},
		{Heading: "## Last Session", Lines: []string{"- date", ""}},
		{Heading: "## Lessons", Lines: []string{"- lesson 1", "- lesson 2", ""}},
	}

	result := pruneNotebook(sections, 190)

	// Should be unchanged — well under limit
	totalBefore := totalLines(sections)
	totalAfter := totalLines(result)
	if totalBefore != totalAfter {
		t.Errorf("pruneNotebook changed content when under limit: %d → %d", totalBefore, totalAfter)
	}
}

func TestPruneNotebook_OverLimit(t *testing.T) {
	// Create a notebook that's over the limit
	var longLines []string
	for i := 0; i < 100; i++ {
		longLines = append(longLines, "- bullet point "+strings.Repeat("x", 10))
	}

	sections := []notebookSection{
		{Heading: "# Memory", Lines: []string{""}},
		{Heading: "## Last Session", Lines: []string{"- date", ""}},
		{Heading: "## Big Section", Lines: longLines},
		{Heading: "## Small Section", Lines: []string{"- keep this", ""}},
	}

	maxLines := 20
	result := pruneNotebook(sections, maxLines)
	total := totalLines(result)

	if total > maxLines {
		t.Errorf("pruneNotebook result has %d lines, want <= %d", total, maxLines)
	}
}

func TestNotebookUpdate_Idempotent(t *testing.T) {
	tmp := t.TempDir()

	// Create a MEMORY.md
	memoryFile := filepath.Join(tmp, "MEMORY.md")
	initialContent := "# Test Memory\n\n## Lessons\n- lesson 1\n"
	if err := os.WriteFile(memoryFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a pending.jsonl with one entry
	aoDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(aoDir, 0755); err != nil {
		t.Fatal(err)
	}
	entry := pendingEntry{
		SessionID: "test-123",
		Summary:   "Did some work",
		Decisions: []string{"Used approach A"},
		Knowledge: []string{"Next: implement feature B"},
		QueuedAt:  time.Date(2026, 2, 25, 12, 0, 0, 0, time.UTC),
	}
	data, _ := json.Marshal(entry)
	if err := os.WriteFile(filepath.Join(aoDir, "pending.jsonl"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	// Run update twice
	sections1, _ := parseNotebookSections(memoryFile)
	sections1 = upsertLastSession(sections1, buildLastSessionSection(&entry))
	content1 := renderNotebook(sections1)
	if err := os.WriteFile(memoryFile, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}

	sections2, _ := parseNotebookSections(memoryFile)
	sections2 = upsertLastSession(sections2, buildLastSessionSection(&entry))
	content2 := renderNotebook(sections2)

	if content1 != content2 {
		t.Error("running update twice with same input produced different output")
		t.Logf("first:\n%s", content1)
		t.Logf("second:\n%s", content2)
	}
}

func TestNotebookUpdate_EmptyPending(t *testing.T) {
	tmp := t.TempDir()

	// No pending.jsonl exists
	entry, err := readLatestPendingEntry(tmp)
	if entry != nil {
		t.Errorf("expected nil entry, got %+v", entry)
	}
	if err == nil {
		// File doesn't exist — that's fine, err will be non-nil
	}
}

func TestNotebookUpdate_AtomicWrite(t *testing.T) {
	tmp := t.TempDir()
	memoryFile := filepath.Join(tmp, "MEMORY.md")

	content := "# Test Memory\n\n## Section\n- data\n"
	if err := atomicWriteFile(memoryFile, []byte(content), 0644); err != nil {
		t.Fatalf("atomicWriteFile failed: %v", err)
	}

	// Verify file is valid
	got, err := os.ReadFile(memoryFile)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != content {
		t.Errorf("content mismatch: got %q, want %q", string(got), content)
	}

	// Verify no temp files left behind
	entries, _ := os.ReadDir(tmp)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".ao-tmp-") {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestBuildLastSessionSection(t *testing.T) {
	entry := &pendingEntry{
		Summary:   "Implemented notebook command",
		Decisions: []string{"Use MEMORY.md as compound notebook"},
		Knowledge: []string{
			"Next: wire into SessionEnd hook",
			"Success: all tests passing",
			"Architecture uses atomic writes",
		},
		QueuedAt: time.Date(2026, 2, 25, 12, 0, 0, 0, time.UTC),
	}

	section := buildLastSessionSection(entry)

	if section.Heading != "## Last Session" {
		t.Errorf("heading = %q, want %q", section.Heading, "## Last Session")
	}

	content := strings.Join(section.Lines, "\n")

	if !strings.Contains(content, "2026-02-25") {
		t.Error("missing date")
	}
	if !strings.Contains(content, "Implemented notebook command") {
		t.Error("missing summary")
	}
	if !strings.Contains(content, "Key decisions") {
		t.Error("missing decisions")
	}
	if !strings.Contains(content, "Next:") {
		t.Error("missing next steps")
	}
}

func TestTotalLines(t *testing.T) {
	sections := []notebookSection{
		{Heading: "# Title", Lines: []string{""}},         // 2 lines
		{Heading: "## Section", Lines: []string{"a", "b"}}, // 3 lines
	}

	got := totalLines(sections)
	if got != 5 {
		t.Errorf("totalLines = %d, want 5", got)
	}
}

func TestReadLatestPendingEntry(t *testing.T) {
	tmp := t.TempDir()
	aoDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(aoDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write two entries — should read the last one
	entries := []pendingEntry{
		{SessionID: "old", Summary: "old session"},
		{SessionID: "new", Summary: "new session"},
	}
	var lines []string
	for _, e := range entries {
		data, _ := json.Marshal(e)
		lines = append(lines, string(data))
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(aoDir, "pending.jsonl"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := readLatestPendingEntry(tmp)
	if err != nil {
		t.Fatalf("readLatestPendingEntry error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil entry")
	}
	if got.SessionID != "new" {
		t.Errorf("SessionID = %q, want %q", got.SessionID, "new")
	}
}

func TestParseSectionsFromString_Empty(t *testing.T) {
	sections := parseSectionsFromString("")
	if len(sections) != 0 {
		t.Errorf("expected 0 sections for empty input, got %d", len(sections))
	}
}

func TestParseSectionsFromString_NoHeadings(t *testing.T) {
	sections := parseSectionsFromString("just text\nno headings here")
	if len(sections) != 1 {
		t.Fatalf("expected 1 preamble section, got %d", len(sections))
	}
	if sections[0].Heading != "" {
		t.Errorf("expected empty heading for preamble, got %q", sections[0].Heading)
	}
}

func TestUpsertLastSession_RemovesDuplicates(t *testing.T) {
	sections := []notebookSection{
		{Heading: "# Memory", Lines: []string{""}},
		{Heading: "## Last Session", Lines: []string{"- old data 1", ""}},
		{Heading: "## Patterns", Lines: []string{"- pattern", ""}},
		{Heading: "## Last Session", Lines: []string{"- old data 2 (duplicate)", ""}},
	}

	newSession := notebookSection{
		Heading: "## Last Session",
		Lines:   []string{"- new data", ""},
	}

	result := upsertLastSession(sections, newSession)

	// Should have 3 sections (no duplicate Last Session)
	count := 0
	for _, s := range result {
		if s.Heading == "## Last Session" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 Last Session section, got %d", count)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 sections total, got %d", len(result))
	}
}

func TestPruneNotebook_SmallSections(t *testing.T) {
	// Test that pruning works even when all sections have <= 2 lines
	sections := []notebookSection{
		{Heading: "# Title", Lines: []string{""}},
		{Heading: "## Last Session", Lines: []string{"- date", ""}},
		{Heading: "## A", Lines: []string{"- a1", "- a2"}},
		{Heading: "## B", Lines: []string{"- b1", "- b2"}},
		{Heading: "## C", Lines: []string{"- c1", "- c2"}},
	}

	result := pruneNotebook(sections, 10)
	total := totalLines(result)

	if total > 10 {
		t.Errorf("pruneNotebook result has %d lines, want <= 10", total)
	}
}
