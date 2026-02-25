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
// batch_forge.go — dedupSimilar
// ---------------------------------------------------------------------------

func TestCov3_batchForge_dedupSimilar(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		want  int
	}{
		{
			name:  "nil input",
			items: nil,
			want:  0,
		},
		{
			name:  "empty slice",
			items: []string{},
			want:  0,
		},
		{
			name:  "no duplicates",
			items: []string{"alpha", "beta", "gamma"},
			want:  3,
		},
		{
			name:  "exact duplicates",
			items: []string{"alpha", "beta", "alpha", "gamma", "beta"},
			want:  3,
		},
		{
			name:  "case-insensitive duplicates",
			items: []string{"Hello World", "hello world", "HELLO WORLD"},
			want:  1,
		},
		{
			name:  "whitespace normalization",
			items: []string{"hello  world", "hello world", "  hello world  "},
			want:  1,
		},
		{
			name:  "ellipsis stripping",
			items: []string{"some long text...", "some long text"},
			want:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := dedupSimilar(tc.items)
			if tc.want == 0 && got != nil && len(got) != 0 {
				t.Errorf("expected nil/empty, got %d items", len(got))
			} else if tc.want > 0 && len(got) != tc.want {
				t.Errorf("dedupSimilar() returned %d items, want %d", len(got), tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// batch_forge.go — normalizeForDedup
// ---------------------------------------------------------------------------

func TestCov3_batchForge_normalizeForDedup(t *testing.T) {
	// Same input should always produce same hash
	hash1 := normalizeForDedup("Hello World")
	hash2 := normalizeForDedup("hello world")
	hash3 := normalizeForDedup("  HELLO   WORLD  ")

	if hash1 != hash2 {
		t.Errorf("case-insensitive normalization failed: %q != %q", hash1, hash2)
	}
	if hash1 != hash3 {
		t.Errorf("whitespace normalization failed: %q != %q", hash1, hash3)
	}

	// Different content should produce different hashes
	hashA := normalizeForDedup("alpha content")
	hashB := normalizeForDedup("beta content")
	if hashA == hashB {
		t.Errorf("expected different hashes for different content")
	}

	// Ellipsis stripping
	hashWithEllipsis := normalizeForDedup("some text...")
	hashWithout := normalizeForDedup("some text")
	if hashWithEllipsis != hashWithout {
		t.Errorf("ellipsis stripping failed: %q != %q", hashWithEllipsis, hashWithout)
	}
}

// ---------------------------------------------------------------------------
// batch_forge.go — humanSize
// ---------------------------------------------------------------------------

func TestCov3_batchForge_humanSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range tests {
		got := humanSize(tc.bytes)
		if got != tc.want {
			t.Errorf("humanSize(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// batch_forge.go — loadForgedIndex / appendForgedRecord
// ---------------------------------------------------------------------------

func TestCov3_batchForge_loadForgedIndex(t *testing.T) {
	t.Run("file does not exist returns empty set", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "nonexistent.jsonl")

		set, err := loadForgedIndex(path)
		if err != nil {
			t.Fatalf("loadForgedIndex: %v", err)
		}
		if len(set) != 0 {
			t.Errorf("expected empty set, got %d entries", len(set))
		}
	})

	t.Run("reads valid records", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "forged.jsonl")

		records := []ForgedRecord{
			{Path: "/tmp/a.jsonl", ForgedAt: time.Now(), Session: "sess-1"},
			{Path: "/tmp/b.jsonl", ForgedAt: time.Now(), Session: "sess-2"},
		}

		var lines []byte
		for _, r := range records {
			data, _ := json.Marshal(r)
			lines = append(lines, data...)
			lines = append(lines, '\n')
		}
		if err := os.WriteFile(path, lines, 0644); err != nil {
			t.Fatalf("write: %v", err)
		}

		set, err := loadForgedIndex(path)
		if err != nil {
			t.Fatalf("loadForgedIndex: %v", err)
		}
		if len(set) != 2 {
			t.Errorf("expected 2 entries, got %d", len(set))
		}
		if !set["/tmp/a.jsonl"] {
			t.Error("expected /tmp/a.jsonl in set")
		}
	})

	t.Run("skips malformed lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "forged.jsonl")

		content := `{"path":"/good.jsonl","forged_at":"2026-01-01T00:00:00Z"}
not-json-line
{"path":"/also-good.jsonl","forged_at":"2026-01-02T00:00:00Z"}
`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}

		set, err := loadForgedIndex(path)
		if err != nil {
			t.Fatalf("loadForgedIndex: %v", err)
		}
		if len(set) != 2 {
			t.Errorf("expected 2 valid entries, got %d", len(set))
		}
	})
}

func TestCov3_batchForge_appendForgedRecord(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "forged.jsonl")

	record := ForgedRecord{
		Path:     "/tmp/test.jsonl",
		ForgedAt: time.Now(),
		Session:  "sess-abc",
	}

	// Should create directory and file
	err := appendForgedRecord(path, record)
	if err != nil {
		t.Fatalf("appendForgedRecord: %v", err)
	}

	// Verify file exists and contains the record
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if !strings.Contains(string(data), "/tmp/test.jsonl") {
		t.Error("expected record path in file")
	}

	// Append another record
	record2 := ForgedRecord{
		Path:     "/tmp/test2.jsonl",
		ForgedAt: time.Now(),
		Session:  "sess-def",
	}
	if err := appendForgedRecord(path, record2); err != nil {
		t.Fatalf("appendForgedRecord second: %v", err)
	}

	// Load back and verify both
	set, err := loadForgedIndex(path)
	if err != nil {
		t.Fatalf("loadForgedIndex: %v", err)
	}
	if len(set) != 2 {
		t.Errorf("expected 2 records, got %d", len(set))
	}
}

// ---------------------------------------------------------------------------
// batch_forge.go — batchForgeAccumulator
// ---------------------------------------------------------------------------

func TestCov3_batchForge_accumulator(t *testing.T) {
	var acc batchForgeAccumulator

	// Accumulate a success
	acc.accumulate(true, []string{"dec1", "dec2"}, []string{"know1"}, "/tmp/a.jsonl")

	if acc.processed != 1 {
		t.Errorf("processed: got %d, want 1", acc.processed)
	}
	if acc.totalDecisions != 2 {
		t.Errorf("totalDecisions: got %d, want 2", acc.totalDecisions)
	}
	if acc.totalKnowledge != 1 {
		t.Errorf("totalKnowledge: got %d, want 1", acc.totalKnowledge)
	}

	// Accumulate a failure
	acc.accumulate(false, nil, nil, "")

	if acc.failed != 1 {
		t.Errorf("failed: got %d, want 1", acc.failed)
	}
	if acc.processed != 1 {
		t.Errorf("processed should still be 1, got %d", acc.processed)
	}

	// Accumulate another success
	acc.accumulate(true, []string{"dec3"}, []string{"know2", "know3"}, "/tmp/b.jsonl")

	if acc.processed != 2 {
		t.Errorf("processed: got %d, want 2", acc.processed)
	}
	if acc.totalDecisions != 3 {
		t.Errorf("totalDecisions: got %d, want 3", acc.totalDecisions)
	}
	if acc.totalKnowledge != 3 {
		t.Errorf("totalKnowledge: got %d, want 3", acc.totalKnowledge)
	}
	if len(acc.processedPaths) != 2 {
		t.Errorf("processedPaths: got %d, want 2", len(acc.processedPaths))
	}
}
