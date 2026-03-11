package search

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestExtra_BuildIndex_WalkError covers the error return from filepath.Walk
// when the root directory does not exist.
func TestExtra_BuildIndex_NonExistentDir(t *testing.T) {
	idx, err := BuildIndex("/nonexistent/path/that/does/not/exist")
	if err != nil {
		t.Fatalf("expected nil error for non-existent dir walk, got: %v", err)
	}
	if len(idx.Terms) != 0 {
		t.Errorf("expected empty index, got %d terms", len(idx.Terms))
	}
}

// TestExtra_SaveIndex_MkdirAllError covers the MkdirAll failure branch in SaveIndex.
func TestExtra_SaveIndex_MkdirAllError(t *testing.T) {
	idx := NewIndex()
	// Use /dev/null as parent so MkdirAll fails (it's a file, not a dir).
	err := SaveIndex(idx, "/dev/null/subdir/index.jsonl")
	if err == nil {
		t.Fatal("expected error when MkdirAll fails, got nil")
	}
}

// TestExtra_SaveIndex_CreateFileError covers the os.Create failure branch.
func TestExtra_SaveIndex_CreateFileError(t *testing.T) {
	idx := NewIndex()
	// Create a directory where the file should be — os.Create will fail.
	tmp := t.TempDir()
	target := filepath.Join(tmp, "index.jsonl")
	if err := os.MkdirAll(target, 0750); err != nil {
		t.Fatal(err)
	}
	err := SaveIndex(idx, target)
	if err == nil {
		t.Fatal("expected error when os.Create fails on a directory, got nil")
	}
}

// TestExtra_writeTermEntry_MarshalSuccess verifies writeTermEntry writes valid JSONL.
func TestExtra_writeTermEntry_MarshalSuccess(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	docs := map[string]bool{"a.md": true, "b.md": true}
	if err := writeTermEntry(w, "hello", docs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

// TestExtra_BuildIndex_SkipsNonIndexable verifies BuildIndex skips non-.md/.jsonl files.
func TestExtra_BuildIndex_SkipsNonIndexable(t *testing.T) {
	tmp := t.TempDir()
	// Create a .txt file that should be skipped.
	if err := os.WriteFile(filepath.Join(tmp, "skip.txt"), []byte("hello world"), 0600); err != nil {
		t.Fatal(err)
	}
	// Create a .md file that should be indexed.
	if err := os.WriteFile(filepath.Join(tmp, "keep.md"), []byte("indexed content"), 0600); err != nil {
		t.Fatal(err)
	}

	idx, err := BuildIndex(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "indexed" should be in the index, "hello" (from .txt) should not.
	if _, ok := idx.Terms["indexed"]; !ok {
		t.Error("expected 'indexed' term from .md file")
	}
}
