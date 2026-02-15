package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReadUnconsumedItems_NoFile(t *testing.T) {
	items, err := readUnconsumedItems("/nonexistent/path/next-work.jsonl", "")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestReadUnconsumedItems_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "next-work.jsonl")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := readUnconsumedItems(path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestReadUnconsumedItems_ConsumedOnly(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "next-work.jsonl")

	entry := nextWorkEntry{
		SourceEpic: "ag-test",
		Timestamp:  "2026-02-10T00:00:00Z",
		Items: []nextWorkItem{
			{Title: "Should be skipped", Severity: "high"},
		},
		Consumed: true,
	}
	data, _ := json.Marshal(entry)
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := readUnconsumedItems(path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items from consumed entry, got %d", len(items))
	}
}

func TestReadUnconsumedItems_UnconsumedWithItems(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "next-work.jsonl")

	entry := nextWorkEntry{
		SourceEpic: "ag-test",
		Timestamp:  "2026-02-10T00:00:00Z",
		Items: []nextWorkItem{
			{Title: "Item A", Severity: "high"},
			{Title: "Item B", Severity: "low"},
		},
		Consumed: false,
	}
	data, _ := json.Marshal(entry)
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := readUnconsumedItems(path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "Item A" {
		t.Errorf("expected first item 'Item A', got %q", items[0].Title)
	}
}

func TestReadUnconsumedItems_EmptyItemsArray(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "next-work.jsonl")

	entry := nextWorkEntry{
		SourceEpic: "ag-empty",
		Timestamp:  "2026-02-10T00:00:00Z",
		Items:      []nextWorkItem{},
		Consumed:   false,
	}
	data, _ := json.Marshal(entry)
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := readUnconsumedItems(path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items from empty items array, got %d", len(items))
	}
}

func TestReadUnconsumedItems_MultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "next-work.jsonl")

	consumed := nextWorkEntry{
		SourceEpic: "ag-old",
		Timestamp:  "2026-02-10T00:00:00Z",
		Items:      []nextWorkItem{{Title: "Old item", Severity: "low"}},
		Consumed:   true,
	}
	unconsumed := nextWorkEntry{
		SourceEpic: "ag-new",
		Timestamp:  "2026-02-10T01:00:00Z",
		Items:      []nextWorkItem{{Title: "New item", Severity: "medium"}},
		Consumed:   false,
	}

	d1, _ := json.Marshal(consumed)
	d2, _ := json.Marshal(unconsumed)
	content := string(d1) + "\n" + string(d2) + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := readUnconsumedItems(path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item (only unconsumed), got %d", len(items))
	}
	if items[0].Title != "New item" {
		t.Errorf("expected 'New item', got %q", items[0].Title)
	}
}

func TestSelectHighestSeverityItem(t *testing.T) {
	tests := []struct {
		name     string
		items    []nextWorkItem
		expected string
	}{
		{
			name:     "empty",
			items:    nil,
			expected: "",
		},
		{
			name: "single item",
			items: []nextWorkItem{
				{Title: "Only one", Severity: "low"},
			},
			expected: "Only one",
		},
		{
			name: "high beats medium and low",
			items: []nextWorkItem{
				{Title: "Low item", Severity: "low"},
				{Title: "High item", Severity: "high"},
				{Title: "Medium item", Severity: "medium"},
			},
			expected: "High item",
		},
		{
			name: "medium beats low",
			items: []nextWorkItem{
				{Title: "Low item", Severity: "low"},
				{Title: "Medium item", Severity: "medium"},
			},
			expected: "Medium item",
		},
		{
			name: "unknown severity ranks lowest",
			items: []nextWorkItem{
				{Title: "Unknown", Severity: "critical"},
				{Title: "Low item", Severity: "low"},
			},
			expected: "Low item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := selectHighestSeverityItem(tt.items)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSeverityRank(t *testing.T) {
	tests := []struct {
		severity string
		rank     int
	}{
		{"high", 3},
		{"medium", 2},
		{"low", 1},
		{"unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			if got := severityRank(tt.severity); got != tt.rank {
				t.Errorf("severityRank(%q) = %d, want %d", tt.severity, got, tt.rank)
			}
		})
	}
}

func TestReadUnconsumedItems_MalformedLines(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "next-work.jsonl")

	entry := nextWorkEntry{
		SourceEpic: "ag-valid",
		Timestamp:  "2026-02-10T00:00:00Z",
		Items:      []nextWorkItem{{Title: "Valid", Severity: "high"}},
		Consumed:   false,
	}
	data, _ := json.Marshal(entry)
	content := "not json at all\n" + string(data) + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := readUnconsumedItems(path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item (skip malformed), got %d", len(items))
	}
	if items[0].Title != "Valid" {
		t.Errorf("expected 'Valid', got %q", items[0].Title)
	}
}

func TestReadUnconsumedItems_RepoFilter_Match(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "next-work.jsonl")

	entry := nextWorkEntry{
		SourceEpic: "ag-repo",
		Timestamp:  "2026-02-10T00:00:00Z",
		Items: []nextWorkItem{
			{Title: "For agentops", Severity: "high", TargetRepo: "agentops"},
			{Title: "For olympus", Severity: "medium", TargetRepo: "olympus"},
		},
		Consumed: false,
	}
	data, _ := json.Marshal(entry)
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := readUnconsumedItems(path, "agentops")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item matching repo filter, got %d", len(items))
	}
	if items[0].Title != "For agentops" {
		t.Errorf("expected 'For agentops', got %q", items[0].Title)
	}
}

func TestReadUnconsumedItems_RepoFilter_Exclude(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "next-work.jsonl")

	entry := nextWorkEntry{
		SourceEpic: "ag-repo",
		Timestamp:  "2026-02-10T00:00:00Z",
		Items: []nextWorkItem{
			{Title: "For olympus only", Severity: "high", TargetRepo: "olympus"},
		},
		Consumed: false,
	}
	data, _ := json.Marshal(entry)
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := readUnconsumedItems(path, "agentops")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items (filtered out), got %d", len(items))
	}
}

func TestReadUnconsumedItems_RepoFilter_Wildcard(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "next-work.jsonl")

	entry := nextWorkEntry{
		SourceEpic: "ag-repo",
		Timestamp:  "2026-02-10T00:00:00Z",
		Items: []nextWorkItem{
			{Title: "For all repos", Severity: "high", TargetRepo: "*"},
			{Title: "For olympus", Severity: "low", TargetRepo: "olympus"},
		},
		Consumed: false,
	}
	data, _ := json.Marshal(entry)
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := readUnconsumedItems(path, "agentops")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item (wildcard passes, olympus excluded), got %d", len(items))
	}
	if items[0].Title != "For all repos" {
		t.Errorf("expected 'For all repos', got %q", items[0].Title)
	}
}

func TestReadUnconsumedItems_RepoFilter_Legacy(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "next-work.jsonl")

	// Legacy items have no target_repo field (empty string after deserialization)
	entry := nextWorkEntry{
		SourceEpic: "ag-legacy",
		Timestamp:  "2026-02-10T00:00:00Z",
		Items: []nextWorkItem{
			{Title: "Legacy item", Severity: "medium"},
		},
		Consumed: false,
	}
	data, _ := json.Marshal(entry)
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	// Legacy items (no target_repo) should pass any filter
	items, err := readUnconsumedItems(path, "agentops")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item (legacy passes all filters), got %d", len(items))
	}
	if items[0].Title != "Legacy item" {
		t.Errorf("expected 'Legacy item', got %q", items[0].Title)
	}
}

func TestReadUnconsumedItems_RepoFilter_EmptyFilter(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "next-work.jsonl")

	entry := nextWorkEntry{
		SourceEpic: "ag-repo",
		Timestamp:  "2026-02-10T00:00:00Z",
		Items: []nextWorkItem{
			{Title: "For agentops", Severity: "high", TargetRepo: "agentops"},
			{Title: "For olympus", Severity: "medium", TargetRepo: "olympus"},
			{Title: "Legacy", Severity: "low"},
		},
		Consumed: false,
	}
	data, _ := json.Marshal(entry)
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	// Empty filter means no filtering - all items pass
	items, err := readUnconsumedItems(path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items (no filter), got %d", len(items))
	}
}
