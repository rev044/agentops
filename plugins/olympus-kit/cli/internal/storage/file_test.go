package storage

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileStorage_Init(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, ".agents/olympus")

	fs := NewFileStorage(WithBaseDir(baseDir))

	if err := fs.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Verify directories were created
	dirs := []string{
		filepath.Join(baseDir, SessionsDir),
		filepath.Join(baseDir, IndexDir),
		filepath.Join(baseDir, ProvenanceDir),
	}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Init() did not create directory %s", dir)
		}
	}
}

func TestFileStorage_WriteIndex_Dedup(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, ".agents/olympus")

	fs := NewFileStorage(WithBaseDir(baseDir))
	if err := fs.Init(); err != nil {
		t.Fatal(err)
	}

	entry := &IndexEntry{
		SessionID:   "test-session-123",
		Date:        time.Now(),
		SessionPath: "/path/to/session.md",
		Summary:     "Test session",
	}

	// Write same entry twice
	if err := fs.WriteIndex(entry); err != nil {
		t.Fatalf("WriteIndex() first call error = %v", err)
	}
	if err := fs.WriteIndex(entry); err != nil {
		t.Fatalf("WriteIndex() second call error = %v", err)
	}

	// Verify only one entry exists
	entries, err := fs.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry after dedup, got %d", len(entries))
	}
}

func TestFileStorage_WriteProvenance(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, ".agents/olympus")

	fs := NewFileStorage(WithBaseDir(baseDir))
	if err := fs.Init(); err != nil {
		t.Fatal(err)
	}

	record := &ProvenanceRecord{
		ID:           "prov-123",
		ArtifactPath: "/output/session.md",
		ArtifactType: "session",
		SourcePath:   "/input/transcript.jsonl",
		SourceType:   "transcript",
		SessionID:    "session-456",
		CreatedAt:    time.Now(),
	}

	if err := fs.WriteProvenance(record); err != nil {
		t.Fatalf("WriteProvenance() error = %v", err)
	}

	// Query provenance
	records, err := fs.QueryProvenance("/output/session.md")
	if err != nil {
		t.Fatalf("QueryProvenance() error = %v", err)
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 provenance record, got %d", len(records))
	}
	if records[0].ID != "prov-123" {
		t.Errorf("Expected ID 'prov-123', got %s", records[0].ID)
	}
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "session"},
		{"Hello World", "hello-world"},
		{"Test 123", "test-123"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"Special!@#$%^&*()Characters", "special-characters"},
		{"UPPERCASE", "uppercase"},
		{"A very long slug that exceeds the maximum allowed length for slugs which is fifty characters", "a-very-long-slug-that-exceeds-the-maximum-allowed"},
	}

	for _, tt := range tests {
		got := generateSlug(tt.input)
		if got != tt.want {
			t.Errorf("generateSlug(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFileStorage_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, ".agents/olympus")

	fs := NewFileStorage(WithBaseDir(baseDir))
	if err := fs.Init(); err != nil {
		t.Fatal(err)
	}

	testPath := filepath.Join(baseDir, "test.txt")
	content := "test content"

	err := fs.atomicWrite(testPath, func(w io.Writer) error {
		_, err := w.Write([]byte(content))
		return err
	})
	if err != nil {
		t.Fatalf("atomicWrite() error = %v", err)
	}

	// Verify content
	data, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != content {
		t.Errorf("Expected content %q, got %q", content, string(data))
	}

	// Verify no temp files left behind
	files, _ := filepath.Glob(filepath.Join(baseDir, ".tmp-*"))
	if len(files) > 0 {
		t.Errorf("Temp files left behind: %v", files)
	}
}

func TestFileStorage_ListSessions_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, ".agents/olympus")

	fs := NewFileStorage(WithBaseDir(baseDir))
	if err := fs.Init(); err != nil {
		t.Fatal(err)
	}

	// List from empty index
	entries, err := fs.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected empty/nil entries, got %v", entries)
	}
}
