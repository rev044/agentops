package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePatternFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("full pattern file", func(t *testing.T) {
		content := `# Mutex Guard Pattern

Always acquire mutex before accessing shared state.
Release in defer to prevent deadlocks.

## Example
...
`
		path := filepath.Join(tmpDir, "mutex-guard.md")
		os.WriteFile(path, []byte(content), 0644)

		p, err := parsePatternFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "Mutex Guard Pattern" {
			t.Errorf("Name = %q, want %q", p.Name, "Mutex Guard Pattern")
		}
		if p.Description == "" {
			t.Error("expected non-empty description")
		}
		if p.FilePath != path {
			t.Errorf("FilePath = %q, want %q", p.FilePath, path)
		}
	})

	t.Run("no title uses filename", func(t *testing.T) {
		content := `Some description without a heading.
`
		path := filepath.Join(tmpDir, "no-title.md")
		os.WriteFile(path, []byte(content), 0644)

		p, err := parsePatternFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "no-title" {
			t.Errorf("Name = %q, want %q", p.Name, "no-title")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := filepath.Join(tmpDir, "empty.md")
		os.WriteFile(path, []byte(""), 0644)

		p, err := parsePatternFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "empty" {
			t.Errorf("Name = %q, want %q", p.Name, "empty")
		}
		if p.Description != "" {
			t.Errorf("Description = %q, want empty", p.Description)
		}
	})

	t.Run("title with description below", func(t *testing.T) {
		content := `# My Pattern

The actual description starts here.
`
		path := filepath.Join(tmpDir, "titled.md")
		os.WriteFile(path, []byte(content), 0644)

		p, err := parsePatternFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "My Pattern" {
			t.Errorf("Name = %q, want %q", p.Name, "My Pattern")
		}
		if p.Description == "" {
			t.Error("expected non-empty description")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := parsePatternFile(filepath.Join(tmpDir, "nope.md"))
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}

func TestCollectPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create patterns directory
	patternsDir := filepath.Join(tmpDir, ".agents", "patterns")
	os.MkdirAll(patternsDir, 0755)

	os.WriteFile(filepath.Join(patternsDir, "mutex.md"), []byte("# Mutex Pattern\n\nUse mutex for shared state."), 0644)
	os.WriteFile(filepath.Join(patternsDir, "pool.md"), []byte("# Connection Pooling\n\nPool database connections."), 0644)

	t.Run("collects all patterns", func(t *testing.T) {
		got, err := collectPatterns(tmpDir, "", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d patterns, want 2", len(got))
		}
	})

	t.Run("filters by query", func(t *testing.T) {
		got, err := collectPatterns(tmpDir, "mutex", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Errorf("got %d patterns for 'mutex', want 1", len(got))
		}
	})

	t.Run("respects limit", func(t *testing.T) {
		got, err := collectPatterns(tmpDir, "", 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) > 1 {
			t.Errorf("got %d patterns, want at most 1", len(got))
		}
	})

	t.Run("no patterns directory", func(t *testing.T) {
		emptyDir := t.TempDir()
		got, err := collectPatterns(emptyDir, "", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
}
