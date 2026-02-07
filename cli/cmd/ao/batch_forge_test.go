package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNormalizeForDedup(t *testing.T) {
	tests := []struct {
		name     string
		a, b     string
		wantSame bool
	}{
		{
			name:     "exact duplicates",
			a:        "Lead-only commit eliminates merge conflicts",
			b:        "Lead-only commit eliminates merge conflicts",
			wantSame: true,
		},
		{
			name:     "case difference",
			a:        "Lead-Only Commit Pattern",
			b:        "lead-only commit pattern",
			wantSame: true,
		},
		{
			name:     "whitespace difference",
			a:        "  Lead-only  commit   pattern  ",
			b:        "Lead-only commit pattern",
			wantSame: true,
		},
		{
			name:     "trailing ellipsis stripped",
			a:        "Workers write files but never commit...",
			b:        "Workers write files but never commit",
			wantSame: true,
		},
		{
			name:     "distinct content with same 80-char prefix",
			a:        "Topological wave decomposition extracts parallelism from dependency graphs by grouping leaves — this is about WAVES",
			b:        "Topological wave decomposition extracts parallelism from dependency graphs by grouping leaves — this is about SORTING",
			wantSame: false,
		},
		{
			name:     "short similar strings are distinct",
			a:        "Use content hashing for dedup",
			b:        "Use content hashing for dedup detection",
			wantSame: false,
		},
		{
			name:     "completely different",
			a:        "Workers should never commit",
			b:        "Wave sizing follows dependency graph",
			wantSame: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyA := normalizeForDedup(tt.a)
			keyB := normalizeForDedup(tt.b)
			gotSame := keyA == keyB
			if gotSame != tt.wantSame {
				t.Errorf("normalizeForDedup(%q) == normalizeForDedup(%q): got %v, want %v\n  keyA=%s\n  keyB=%s",
					tt.a, tt.b, gotSame, tt.wantSame, keyA, keyB)
			}
		})
	}
}

func TestDedupSimilar(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		wantCount int
	}{
		{
			name:      "nil input",
			input:     nil,
			wantCount: 0,
		},
		{
			name:      "empty input",
			input:     []string{},
			wantCount: 0,
		},
		{
			name:      "no duplicates",
			input:     []string{"alpha", "beta", "gamma"},
			wantCount: 3,
		},
		{
			name:      "exact duplicates removed",
			input:     []string{"alpha", "beta", "alpha", "gamma", "beta"},
			wantCount: 3,
		},
		{
			name: "case-insensitive dedup",
			input: []string{
				"Lead-only commit pattern",
				"lead-only commit pattern",
				"LEAD-ONLY COMMIT PATTERN",
			},
			wantCount: 1,
		},
		{
			name: "preserves distinct long strings",
			input: []string{
				"Topological wave decomposition extracts parallelism from dependency graphs by grouping leaves — approach A",
				"Topological wave decomposition extracts parallelism from dependency graphs by grouping leaves — approach B",
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dedupSimilar(tt.input)
			if len(result) != tt.wantCount {
				t.Errorf("dedupSimilar() returned %d items, want %d. Items: %v", len(result), tt.wantCount, result)
			}
		})
	}
}

func TestFindPendingTranscripts(t *testing.T) {
	// Create temp directory with test transcript files
	tmpDir, err := os.MkdirTemp("", "batch_forge_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some test JSONL files (must be > 100 bytes to pass the size filter)
	for _, name := range []string{"session1.jsonl", "session2.jsonl"} {
		content := []byte(`{"role":"user","content":"hello world, this is a test message with enough content"}` + "\n" +
			`{"role":"assistant","content":"hi there, this is a sufficiently long response to exceed the 100 byte minimum"}` + "\n")
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a file too small to be a transcript
	if err := os.WriteFile(filepath.Join(tmpDir, "tiny.jsonl"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a non-JSONL file
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# Hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a subagents directory that should be skipped
	subDir := filepath.Join(tmpDir, "subagents")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "sub.jsonl"), []byte(`{"role":"user","content":"skip me please"}`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	candidates, err := findPendingTranscripts(tmpDir)
	if err != nil {
		t.Fatalf("findPendingTranscripts: %v", err)
	}

	if len(candidates) != 2 {
		t.Errorf("got %d candidates, want 2", len(candidates))
		for _, c := range candidates {
			t.Logf("  candidate: %s (size=%d)", c.path, c.size)
		}
	}

	// Verify sorted by mod time (oldest first)
	if len(candidates) >= 2 {
		if candidates[0].modTime.After(candidates[1].modTime) {
			t.Error("candidates not sorted by modification time (oldest first)")
		}
	}
}

func TestFindPendingTranscriptsEmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "batch_forge_empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	candidates, err := findPendingTranscripts(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(candidates) != 0 {
		t.Errorf("got %d candidates, want 0", len(candidates))
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
	}

	for _, tt := range tests {
		got := humanSize(tt.bytes)
		if got != tt.want {
			t.Errorf("humanSize(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestTranscriptCandidateFields(t *testing.T) {
	now := time.Now()
	c := transcriptCandidate{
		path:    "/tmp/test.jsonl",
		modTime: now,
		size:    1234,
	}

	if c.path != "/tmp/test.jsonl" {
		t.Error("path mismatch")
	}
	if c.size != 1234 {
		t.Error("size mismatch")
	}
	if !c.modTime.Equal(now) {
		t.Error("modTime mismatch")
	}
}
