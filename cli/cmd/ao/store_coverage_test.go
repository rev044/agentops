package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// createIndexEntry
// ---------------------------------------------------------------------------

func TestStoreCoverage_CreateIndexEntry(t *testing.T) {
	tmp := t.TempDir()

	t.Run("basic markdown file", func(t *testing.T) {
		content := "# Test Document\n\nSome content with pattern: stuff.\n"
		path := filepath.Join(tmp, "doc.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		entry, err := createIndexEntry(path, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry.Path != path {
			t.Errorf("Path = %q, want %q", entry.Path, path)
		}
		if entry.ID != "doc.md" {
			t.Errorf("ID = %q, want %q", entry.ID, "doc.md")
		}
		if entry.Title != "Test Document" {
			t.Errorf("Title = %q, want %q", entry.Title, "Test Document")
		}
		if entry.Content != content {
			t.Errorf("Content mismatch")
		}
		if entry.IndexedAt.IsZero() {
			t.Error("IndexedAt should not be zero")
		}
		if entry.ModifiedAt.IsZero() {
			t.Error("ModifiedAt should not be zero")
		}
	})

	t.Run("with categorize", func(t *testing.T) {
		content := "---\ncategory: testing\ntags: [go, unit]\n---\n# Title\n\nContent.\n"
		path := filepath.Join(tmp, "categorized.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		entry, err := createIndexEntry(path, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry.Category != "testing" {
			t.Errorf("Category = %q, want %q", entry.Category, "testing")
		}
		if len(entry.Tags) == 0 {
			t.Error("expected tags to be populated")
		}
	})

	t.Run("without categorize has no category", func(t *testing.T) {
		content := "---\ncategory: testing\n---\n# Title\n"
		path := filepath.Join(tmp, "nocat.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		entry, err := createIndexEntry(path, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry.Category != "" {
			t.Errorf("Category = %q, want empty when categorize=false", entry.Category)
		}
	})

	t.Run("with utility and maturity", func(t *testing.T) {
		content := "# Doc\n\n**Utility**: 0.8\n**Maturity**: established\n"
		path := filepath.Join(tmp, "meta.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		entry, err := createIndexEntry(path, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry.Utility < 0.79 || entry.Utility > 0.81 {
			t.Errorf("Utility = %v, want ~0.8", entry.Utility)
		}
		if entry.Maturity != "established" {
			t.Errorf("Maturity = %q, want %q", entry.Maturity, "established")
		}
	})

	t.Run("type from path", func(t *testing.T) {
		dir := filepath.Join(tmp, "learnings")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "lesson.md")
		if err := os.WriteFile(path, []byte("# Lesson\n"), 0644); err != nil {
			t.Fatal(err)
		}
		entry, err := createIndexEntry(path, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry.Type != "learning" {
			t.Errorf("Type = %q, want %q", entry.Type, "learning")
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		_, err := createIndexEntry(filepath.Join(tmp, "nope.md"), false)
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}

// ---------------------------------------------------------------------------
// appendToIndex
// ---------------------------------------------------------------------------

func TestStoreCoverage_AppendToIndex(t *testing.T) {
	tmp := t.TempDir()

	entry := &IndexEntry{
		Path:       "/fake/path.md",
		ID:         "test-id",
		Type:       "learning",
		Title:      "Test Title",
		Content:    "Some content",
		IndexedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	t.Run("creates index dir and file", func(t *testing.T) {
		if err := appendToIndex(tmp, entry); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		indexPath := filepath.Join(tmp, IndexDir, IndexFileName)
		data, err := os.ReadFile(indexPath)
		if err != nil {
			t.Fatalf("read index: %v", err)
		}
		if !strings.Contains(string(data), "test-id") {
			t.Error("expected index to contain entry ID")
		}
	})

	t.Run("appends multiple entries", func(t *testing.T) {
		entry2 := &IndexEntry{
			Path:       "/fake/path2.md",
			ID:         "test-id-2",
			Type:       "pattern",
			Title:      "Second",
			Content:    "More content",
			IndexedAt:  time.Now(),
			ModifiedAt: time.Now(),
		}
		if err := appendToIndex(tmp, entry2); err != nil {
			t.Fatalf("append second entry: %v", err)
		}

		indexPath := filepath.Join(tmp, IndexDir, IndexFileName)
		data, err := os.ReadFile(indexPath)
		if err != nil {
			t.Fatalf("read index: %v", err)
		}
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) < 2 {
			t.Errorf("expected at least 2 lines, got %d", len(lines))
		}
	})
}

// ---------------------------------------------------------------------------
// searchIndex
// ---------------------------------------------------------------------------

func TestStoreCoverage_SearchIndex(t *testing.T) {
	tmp := t.TempDir()

	// Populate index
	entries := []*IndexEntry{
		{Path: "/a.md", ID: "a", Type: "learning", Title: "Mutex Pattern", Content: "Use mutex for shared state", Keywords: []string{"mutex", "concurrency"}, Utility: 0.8, IndexedAt: time.Now(), ModifiedAt: time.Now()},
		{Path: "/b.md", ID: "b", Type: "pattern", Title: "Error Handling", Content: "Always check errors in Go", Keywords: []string{"error", "go"}, Utility: 0.6, IndexedAt: time.Now(), ModifiedAt: time.Now()},
		{Path: "/c.md", ID: "c", Type: "research", Title: "Database Pooling", Content: "Connection pool setup guide", Keywords: []string{"database"}, Utility: 0.9, IndexedAt: time.Now(), ModifiedAt: time.Now()},
	}
	for _, e := range entries {
		if err := appendToIndex(tmp, e); err != nil {
			t.Fatalf("append: %v", err)
		}
	}

	t.Run("finds matching entries", func(t *testing.T) {
		results, err := searchIndex(tmp, "mutex", 10)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(results) == 0 {
			t.Error("expected results for 'mutex'")
		}
		if results[0].Entry.ID != "a" {
			t.Errorf("first result ID = %q, want %q", results[0].Entry.ID, "a")
		}
		if results[0].Score <= 0 {
			t.Error("expected positive score")
		}
	})

	t.Run("limit works", func(t *testing.T) {
		results, err := searchIndex(tmp, "the", 1)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(results) > 1 {
			t.Errorf("expected at most 1 result, got %d", len(results))
		}
	})

	t.Run("no matches returns empty", func(t *testing.T) {
		results, err := searchIndex(tmp, "zzzznonexistent", 10)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("no index returns error", func(t *testing.T) {
		_, err := searchIndex(filepath.Join(tmp, "nope"), "test", 10)
		if err == nil {
			t.Error("expected error for missing index")
		}
	})

	t.Run("results sorted by score descending", func(t *testing.T) {
		results, err := searchIndex(tmp, "error go", 10)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		for i := 1; i < len(results); i++ {
			if results[i].Score > results[i-1].Score {
				t.Errorf("results not sorted: score[%d]=%v > score[%d]=%v", i, results[i].Score, i-1, results[i-1].Score)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// computeIndexStats
// ---------------------------------------------------------------------------

func TestStoreCoverage_ComputeIndexStats(t *testing.T) {
	t.Run("with entries", func(t *testing.T) {
		tmp := t.TempDir()

		now := time.Now()
		earlier := now.Add(-24 * time.Hour)
		entries := []*IndexEntry{
			{Path: "/a.md", ID: "a", Type: "learning", Utility: 0.8, IndexedAt: earlier, ModifiedAt: earlier},
			{Path: "/b.md", ID: "b", Type: "pattern", Utility: 0.6, IndexedAt: now, ModifiedAt: now},
			{Path: "/c.md", ID: "c", Type: "learning", Utility: 0.0, IndexedAt: now, ModifiedAt: now},
		}
		for _, e := range entries {
			if err := appendToIndex(tmp, e); err != nil {
				t.Fatalf("append: %v", err)
			}
		}

		stats, err := computeIndexStats(tmp)
		if err != nil {
			t.Fatalf("computeIndexStats: %v", err)
		}
		if stats.TotalEntries != 3 {
			t.Errorf("TotalEntries = %d, want 3", stats.TotalEntries)
		}
		if stats.ByType["learning"] != 2 {
			t.Errorf("ByType[learning] = %d, want 2", stats.ByType["learning"])
		}
		if stats.ByType["pattern"] != 1 {
			t.Errorf("ByType[pattern] = %d, want 1", stats.ByType["pattern"])
		}
		// MeanUtility should be (0.8+0.6)/2 = 0.7 (zero utility not counted)
		if stats.MeanUtility < 0.69 || stats.MeanUtility > 0.71 {
			t.Errorf("MeanUtility = %v, want ~0.7", stats.MeanUtility)
		}
	})

	t.Run("no index file returns empty stats", func(t *testing.T) {
		tmp := t.TempDir()
		stats, err := computeIndexStats(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats.TotalEntries != 0 {
			t.Errorf("TotalEntries = %d, want 0", stats.TotalEntries)
		}
	})
}

// ---------------------------------------------------------------------------
// walkIndexableFiles
// ---------------------------------------------------------------------------

func TestStoreCoverage_WalkIndexableFiles(t *testing.T) {
	tmp := t.TempDir()

	// Create various files
	files := map[string]string{
		"a.md":        "test",
		"b.jsonl":     "test",
		"c.txt":       "test",
		"d.go":        "test",
		"sub/e.md":    "test",
		"sub/f.jsonl": "test",
	}
	for name, content := range files {
		path := filepath.Join(tmp, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := walkIndexableFiles(tmp)
	if err != nil {
		t.Fatalf("walkIndexableFiles: %v", err)
	}
	// Should find a.md, b.jsonl, sub/e.md, sub/f.jsonl (4 files)
	if len(got) != 4 {
		t.Errorf("found %d files, want 4; got %v", len(got), got)
	}
}

// ---------------------------------------------------------------------------
// collectArtifactFiles
// ---------------------------------------------------------------------------

func TestStoreCoverage_CollectArtifactFiles(t *testing.T) {
	tmp := t.TempDir()

	// Create .agents subdirectories
	for _, sub := range []string{"learnings", "patterns", "research", "retros", "candidates"} {
		dir := filepath.Join(tmp, ".agents", sub)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "test.md"), []byte("# Test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	files := collectArtifactFiles(tmp)
	if len(files) != 5 {
		t.Errorf("found %d files, want 5; got %v", len(files), files)
	}
}

// ---------------------------------------------------------------------------
// indexFiles
// ---------------------------------------------------------------------------

func TestStoreCoverage_IndexFiles(t *testing.T) {
	tmp := t.TempDir()

	// Create a couple of indexable files
	dir := filepath.Join(tmp, "learnings")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.md", "b.md"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("# Test\n\nSome content.\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	files := []string{
		filepath.Join(dir, "a.md"),
		filepath.Join(dir, "b.md"),
		filepath.Join(dir, "nonexistent.md"), // should be skipped
	}

	count := indexFiles(tmp, files, false)
	if count != 2 {
		t.Errorf("indexFiles returned %d, want 2", count)
	}

	// Verify index was written
	indexPath := filepath.Join(tmp, IndexDir, IndexFileName)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 index lines, got %d", len(lines))
	}
}

// ---------------------------------------------------------------------------
// printSearchResults (smoke test — just ensure no panic)
// ---------------------------------------------------------------------------

func TestStoreCoverage_PrintSearchResults(t *testing.T) {
	t.Run("empty results", func(t *testing.T) {
		// Should not panic
		printSearchResults("test query", nil)
	})

	t.Run("with results", func(t *testing.T) {
		results := []SearchResult{
			{
				Entry:   IndexEntry{Title: "Test", Type: "learning", Path: "/test.md", Utility: 0.8},
				Score:   3.5,
				Snippet: "some snippet",
			},
		}
		printSearchResults("test query", results)
	})

	t.Run("result without snippet", func(t *testing.T) {
		results := []SearchResult{
			{
				Entry: IndexEntry{Title: "Test", Type: "pattern", Path: "/test.md"},
				Score: 1.0,
			},
		}
		printSearchResults("query", results)
	})
}

// ---------------------------------------------------------------------------
// printIndexStats (smoke test — just ensure no panic)
// ---------------------------------------------------------------------------

func TestStoreCoverage_PrintIndexStats(t *testing.T) {
	t.Run("empty stats", func(t *testing.T) {
		stats := &IndexStats{
			ByType:    make(map[string]int),
			IndexPath: "/fake/path",
		}
		printIndexStats(stats)
	})

	t.Run("with entries and times", func(t *testing.T) {
		stats := &IndexStats{
			TotalEntries: 5,
			ByType:       map[string]int{"learning": 3, "pattern": 2},
			MeanUtility:  0.75,
			OldestEntry:  time.Now().Add(-24 * time.Hour),
			NewestEntry:  time.Now(),
			IndexPath:    "/fake/path",
		}
		printIndexStats(stats)
	})

	t.Run("with zero oldest entry", func(t *testing.T) {
		stats := &IndexStats{
			TotalEntries: 0,
			ByType:       map[string]int{},
			IndexPath:    "/fake/path",
		}
		printIndexStats(stats)
	})
}

// ---------------------------------------------------------------------------
// searchIndex JSONL line parse resilience
// ---------------------------------------------------------------------------

func TestStoreCoverage_SearchIndexMalformedLines(t *testing.T) {
	tmp := t.TempDir()

	// Write index with one valid and one malformed line
	indexDir := filepath.Join(tmp, IndexDir)
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		t.Fatal(err)
	}
	valid := IndexEntry{
		Path: "/a.md", ID: "a", Type: "learning", Title: "Mutex",
		Content: "Use mutex", IndexedAt: time.Now(), ModifiedAt: time.Now(),
	}
	validJSON, _ := json.Marshal(valid)
	indexPath := filepath.Join(indexDir, IndexFileName)
	content := "NOT VALID JSON\n" + string(validJSON) + "\n"
	if err := os.WriteFile(indexPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := searchIndex(tmp, "mutex", 10)
	if err != nil {
		t.Fatalf("searchIndex with malformed lines: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result (skipping malformed), got %d", len(results))
	}
}
