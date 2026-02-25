package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractMarkdownBody(t *testing.T) {
	input := "---\ntitle: Test Learning\ndate: 2026-01-15\n---\nThis is the body content.\nIt spans multiple lines."
	got := extractMarkdownBody(input)
	want := "This is the body content.\nIt spans multiple lines."
	if got != want {
		t.Errorf("extractMarkdownBody with frontmatter:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestExtractMarkdownBody_NoFrontMatter(t *testing.T) {
	input := "Just plain content.\nNo frontmatter here."
	got := extractMarkdownBody(input)
	if got != input {
		t.Errorf("extractMarkdownBody without frontmatter:\n  got:  %q\n  want: %q", got, input)
	}
}

func TestExtractJSONLBody(t *testing.T) {
	input := `{"content":"some content","title":"title"}`
	got := extractJSONLBody(input)
	want := "some content"
	if got != want {
		t.Errorf("extractJSONLBody with content field:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestExtractJSONLBody_TitleFallback(t *testing.T) {
	input := `{"title":"just title"}`
	got := extractJSONLBody(input)
	want := "just title"
	if got != want {
		t.Errorf("extractJSONLBody title fallback:\n  got:  %q\n  want: %q", got, want)
	}
}

func TestHashNormalizedContent_IdenticalContent(t *testing.T) {
	a := "  This is Some  Content  "
	b := "this is some content"
	hashA := hashNormalizedContent(a)
	hashB := hashNormalizedContent(b)
	if hashA != hashB {
		t.Errorf("expected identical hashes for whitespace/case variants:\n  a: %s\n  b: %s", hashA, hashB)
	}
}

func TestHashNormalizedContent_DifferentContent(t *testing.T) {
	a := "This is about mutexes and concurrency"
	b := "This is about database connections"
	hashA := hashNormalizedContent(a)
	hashB := hashNormalizedContent(b)
	if hashA == hashB {
		t.Errorf("expected different hashes for different content, both got: %s", hashA)
	}
}

func TestRunDedup_NoDuplicates(t *testing.T) {
	// Create temp directory structure
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatalf("creating learnings dir: %v", err)
	}

	// Write two .md files with different body content
	file1 := filepath.Join(learningsDir, "learn-01.md")
	file2 := filepath.Join(learningsDir, "learn-02.md")

	if err := os.WriteFile(file1, []byte("---\ntitle: First\n---\nContent about mutexes."), 0o644); err != nil {
		t.Fatalf("writing file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("---\ntitle: Second\n---\nContent about databases."), 0o644); err != nil {
		t.Fatalf("writing file2: %v", err)
	}

	// Change to temp dir for the command
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Force JSON output so we can parse the result
	origOutput := output
	output = "json"
	defer func() { output = origOutput }()

	// Capture stdout
	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	err := runDedup(nil, nil)

	_ = w.Close()
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("runDedup returned error: %v", err)
	}

	var result DedupResult
	if decErr := json.NewDecoder(r).Decode(&result); decErr != nil {
		t.Fatalf("decoding JSON output: %v", decErr)
	}

	if result.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want 2", result.TotalFiles)
	}
	if result.DuplicateGroups != 0 {
		t.Errorf("DuplicateGroups = %d, want 0", result.DuplicateGroups)
	}
	if result.DuplicateFiles != 0 {
		t.Errorf("DuplicateFiles = %d, want 0", result.DuplicateFiles)
	}
}

func TestRunDedup_FindsDuplicates(t *testing.T) {
	// Create temp directory structure
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatalf("creating learnings dir: %v", err)
	}

	// Write two .md files with identical body content but different frontmatter
	file1 := filepath.Join(learningsDir, "learn-a.md")
	file2 := filepath.Join(learningsDir, "learn-b.md")

	if err := os.WriteFile(file1, []byte("---\ntitle: First Copy\ndate: 2026-01-10\n---\nAlways use mutex for shared state."), 0o644); err != nil {
		t.Fatalf("writing file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("---\ntitle: Second Copy\ndate: 2026-02-15\n---\nAlways use mutex for shared state."), 0o644); err != nil {
		t.Fatalf("writing file2: %v", err)
	}

	// Change to temp dir for the command
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Force JSON output so we can parse the result
	origOutput := output
	output = "json"
	defer func() { output = origOutput }()

	// Capture stdout
	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w

	err := runDedup(nil, nil)

	_ = w.Close()
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("runDedup returned error: %v", err)
	}

	var result DedupResult
	if decErr := json.NewDecoder(r).Decode(&result); decErr != nil {
		t.Fatalf("decoding JSON output: %v", decErr)
	}

	if result.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want 2", result.TotalFiles)
	}
	if result.DuplicateGroups != 1 {
		t.Errorf("DuplicateGroups = %d, want 1", result.DuplicateGroups)
	}
	if result.DuplicateFiles != 2 {
		t.Errorf("DuplicateFiles = %d, want 2", result.DuplicateFiles)
	}
	if len(result.Groups) != 1 {
		t.Fatalf("len(Groups) = %d, want 1", len(result.Groups))
	}
	if result.Groups[0].Count != 2 {
		t.Errorf("Groups[0].Count = %d, want 2", result.Groups[0].Count)
	}
}
