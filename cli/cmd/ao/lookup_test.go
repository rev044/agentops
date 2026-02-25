package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMatchesID_ExactMatch(t *testing.T) {
	if !matchesID("learn-001", "/path/to/learn-001.md", "learn-001") {
		t.Error("expected exact ID match")
	}
}

func TestMatchesID_CaseInsensitive(t *testing.T) {
	if !matchesID("Learn-001", "/path/to/file.md", "learn-001") {
		t.Error("expected case-insensitive match")
	}
}

func TestMatchesID_FilenameMatch(t *testing.T) {
	if !matchesID("some-other-id", "/path/to/2026-02-22-cross-language.md", "cross-language") {
		t.Error("expected filename substring match")
	}
}

func TestMatchesID_NoMatch(t *testing.T) {
	if matchesID("learn-001", "/path/to/learn-001.md", "learn-999") {
		t.Error("expected no match")
	}
}

func TestFilterByBead(t *testing.T) {
	learnings := []learning{
		{ID: "l1", SourceBead: "ag-mrr"},
		{ID: "l2", SourceBead: "ag-xyz"},
		{ID: "l3", SourceBead: "ag-mrr"},
		{ID: "l4", SourceBead: ""},
	}
	filtered := filterByBead(learnings, "ag-mrr")
	if len(filtered) != 2 {
		t.Errorf("expected 2 matches, got %d", len(filtered))
	}
	for _, l := range filtered {
		if l.SourceBead != "ag-mrr" {
			t.Errorf("unexpected bead: %s", l.SourceBead)
		}
	}
}

func TestFilterByBead_CaseInsensitive(t *testing.T) {
	learnings := []learning{
		{ID: "l1", SourceBead: "AG-MRR"},
	}
	filtered := filterByBead(learnings, "ag-mrr")
	if len(filtered) != 1 {
		t.Errorf("expected 1 match, got %d", len(filtered))
	}
}

func TestFilterByBead_Empty(t *testing.T) {
	filtered := filterByBead(nil, "ag-mrr")
	if len(filtered) != 0 {
		t.Errorf("expected 0 matches, got %d", len(filtered))
	}
}

func TestRelPath(t *testing.T) {
	cwd := "/Users/test/project"
	path := "/Users/test/project/.agents/learnings/test.md"
	got := relPath(cwd, path)
	if got != ".agents/learnings/test.md" {
		t.Errorf("relPath = %q, want .agents/learnings/test.md", got)
	}
}

func TestLookupByID_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	// Create empty learnings dir
	os.MkdirAll(filepath.Join(tmpDir, ".agents", "learnings"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, ".agents", "patterns"), 0755)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	err := lookupByID(tmpDir, "nonexistent-id", nil)
	if err == nil {
		t.Error("expected error for nonexistent ID")
	}
	if !strings.Contains(err.Error(), "no artifact found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFormatLookupAge(t *testing.T) {
	tests := []struct {
		weeks float64
		want  string
	}{
		{0.05, "<1d"},
		{0.3, "2d"},
		{1.0, "1w"},
		{1.5, "2w"},
		{4.0, "4w"},
		{8.0, "2mo"},
	}
	for _, tt := range tests {
		got := formatLookupAge(tt.weeks)
		if got != tt.want {
			t.Errorf("formatLookupAge(%v) = %q, want %q", tt.weeks, got, tt.want)
		}
	}
}
