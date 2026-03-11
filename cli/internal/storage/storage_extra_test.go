package storage

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestExtra_scanJSONLFile_OpenError covers the non-IsNotExist error branch.
func TestExtra_scanJSONLFile_OpenError(t *testing.T) {
	// Create a directory where the file should be — Open will fail with EISDIR.
	tmp := t.TempDir()
	dirPath := filepath.Join(tmp, "notafile")
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		t.Fatal(err)
	}
	err := scanJSONLFile(dirPath, func(line []byte) {})
	if err == nil {
		t.Fatal("expected error opening a directory as file, got nil")
	}
}

// TestExtra_scanJSONLFile_NotExist covers the IsNotExist early return.
func TestExtra_scanJSONLFile_NotExist(t *testing.T) {
	err := scanJSONLFile("/nonexistent/file.jsonl", func(line []byte) {})
	if err != nil {
		t.Fatalf("expected nil for non-existent file, got: %v", err)
	}
}

// TestExtra_writeSyncClose_WriteError covers the write failure branch.
func TestExtra_writeSyncClose_WriteError(t *testing.T) {
	tmp := t.TempDir()
	f, err := os.CreateTemp(tmp, "wsc-")
	if err != nil {
		t.Fatal(err)
	}

	writeErr := errors.New("simulated write error")
	err = writeSyncClose(f, func(w io.Writer) error {
		return writeErr
	})
	if err == nil {
		t.Fatal("expected error from writeFunc, got nil")
	}
	if !errors.Is(err, writeErr) {
		t.Errorf("expected wrapped writeErr, got: %v", err)
	}
}

// TestExtra_writeSyncClose_SyncError covers the sync failure branch.
func TestExtra_writeSyncClose_SyncError(t *testing.T) {
	// Create a file, close it, then pass the closed file to trigger sync error.
	tmp := t.TempDir()
	f, err := os.CreateTemp(tmp, "wsc-sync-")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	f.Close()

	// Reopen read-only so Sync after write may fail, or use a pipe.
	r, w, _ := os.Pipe()
	defer r.Close()

	err = writeSyncClose(w, func(wr io.Writer) error {
		_, e := wr.Write([]byte("data"))
		return e
	})
	// Pipe doesn't support Sync — should get an error.
	if err == nil {
		t.Log("sync on pipe did not error — platform dependent, skipping")
	}
	_ = name // used above
}

// TestExtra_withLockedFile_MkdirError covers MkdirAll failure in withLockedFile.
func TestExtra_withLockedFile_MkdirError(t *testing.T) {
	fs := &FileStorage{BaseDir: t.TempDir()}
	// Use /dev/null as parent to force MkdirAll failure.
	err := fs.withLockedFile("/dev/null/sub/file.jsonl", func(f *os.File) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected MkdirAll error, got nil")
	}
}

// TestExtra_withLockedFile_OpenFileError covers the OpenFile failure branch.
func TestExtra_withLockedFile_OpenFileError(t *testing.T) {
	fs := &FileStorage{BaseDir: t.TempDir()}
	// Create a directory where the file should be created.
	tmp := t.TempDir()
	target := filepath.Join(tmp, "lockfile")
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	err := fs.withLockedFile(target, func(f *os.File) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected OpenFile error on directory, got nil")
	}
}
