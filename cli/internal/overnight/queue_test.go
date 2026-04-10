package overnight

import (
	"os"
	"path/filepath"
	"testing"
)

// writeQueue writes a fixture queue file under t.TempDir() and returns the
// absolute path for ParseQueue calls.
func writeQueue(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "queue.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write queue: %v", err)
	}
	return path
}

func TestParseQueue_EmptyPath_ReturnsNilNoError(t *testing.T) {
	items, err := ParseQueue("")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil items, got %v", items)
	}
}

func TestParseQueue_MissingFile_SoftFail(t *testing.T) {
	items, err := ParseQueue(filepath.Join(t.TempDir(), "does-not-exist.md"))
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil items, got %v", items)
	}
}

func TestParseQueue_SimpleBullets(t *testing.T) {
	path := writeQueue(t, `# Queue

- First item
- Second item
- Third item
`)
	items, err := ParseQueue(path)
	if err != nil {
		t.Fatalf("ParseQueue: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	want := []string{"First item", "Second item", "Third item"}
	for i, it := range items {
		if it.Order != i+1 {
			t.Fatalf("item %d: expected order %d, got %d", i, i+1, it.Order)
		}
		if it.Title != want[i] {
			t.Fatalf("item %d: expected title %q, got %q", i, want[i], it.Title)
		}
		if it.Description != "" {
			t.Fatalf("item %d: expected empty description, got %q", i, it.Description)
		}
	}
}

func TestParseQueue_BulletWithDescription(t *testing.T) {
	path := writeQueue(t, `- Bullet title
  A description paragraph
  continuing on next line.
`)
	items, err := ParseQueue(path)
	if err != nil {
		t.Fatalf("ParseQueue: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Bullet title" {
		t.Fatalf("unexpected title %q", items[0].Title)
	}
	want := "A description paragraph continuing on next line."
	if items[0].Description != want {
		t.Fatalf("expected description %q, got %q", want, items[0].Description)
	}
}

func TestParseQueue_TargetFileMarker(t *testing.T) {
	path := writeQueue(t, `- Fix bug [file: cli/foo.go]
`)
	items, err := ParseQueue(path)
	if err != nil {
		t.Fatalf("ParseQueue: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].TargetFile != "cli/foo.go" {
		t.Fatalf("expected TargetFile=cli/foo.go, got %q", items[0].TargetFile)
	}
	if items[0].Title != "Fix bug" {
		t.Fatalf("expected Title=Fix bug, got %q", items[0].Title)
	}
}

func TestParseQueue_SeverityMarker(t *testing.T) {
	path := writeQueue(t, `- Critical bug [severity: high]
`)
	items, err := ParseQueue(path)
	if err != nil {
		t.Fatalf("ParseQueue: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Severity != "high" {
		t.Fatalf("expected Severity=high, got %q", items[0].Severity)
	}
	if items[0].Title != "Critical bug" {
		t.Fatalf("expected Title=Critical bug, got %q", items[0].Title)
	}
}

func TestParseQueue_IgnoresHeadings(t *testing.T) {
	path := writeQueue(t, `# Top
## Section A

- Alpha

## Section B

- Beta
`)
	items, err := ParseQueue(path)
	if err != nil {
		t.Fatalf("ParseQueue: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "Alpha" || items[1].Title != "Beta" {
		t.Fatalf("unexpected titles: %+v", items)
	}
}

func TestParseQueue_IgnoresComments(t *testing.T) {
	path := writeQueue(t, `<!-- hidden operator note -->
- Visible item
<!-- another note -->
`)
	items, err := ParseQueue(path)
	if err != nil {
		t.Fatalf("ParseQueue: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Visible item" {
		t.Fatalf("unexpected title %q", items[0].Title)
	}
}

func TestParseQueue_PreservesOrder(t *testing.T) {
	path := writeQueue(t, `# Roadmap

## One
- first

## Two
- second
- third

## Three
- fourth
- fifth
`)
	items, err := ParseQueue(path)
	if err != nil {
		t.Fatalf("ParseQueue: %v", err)
	}
	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}
	wantTitles := []string{"first", "second", "third", "fourth", "fifth"}
	for i, it := range items {
		if it.Order != i+1 {
			t.Fatalf("item %d: expected order %d, got %d", i, i+1, it.Order)
		}
		if it.Title != wantTitles[i] {
			t.Fatalf("item %d: expected title %q, got %q", i, wantTitles[i], it.Title)
		}
	}
}

func TestParseQueue_BothMarkersOnSameBullet(t *testing.T) {
	path := writeQueue(t, `- Ship launcher [file: cli/cmd/ao/overnight.go] [severity: high]
`)
	items, err := ParseQueue(path)
	if err != nil {
		t.Fatalf("ParseQueue: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].TargetFile != "cli/cmd/ao/overnight.go" {
		t.Fatalf("expected TargetFile=cli/cmd/ao/overnight.go, got %q", items[0].TargetFile)
	}
	if items[0].Severity != "high" {
		t.Fatalf("expected Severity=high, got %q", items[0].Severity)
	}
	if items[0].Title != "Ship launcher" {
		t.Fatalf("expected Title=Ship launcher, got %q", items[0].Title)
	}
}
