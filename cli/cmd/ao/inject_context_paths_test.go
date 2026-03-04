package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContextArtifactDir_WithRunID(t *testing.T) {
	got := contextArtifactDir("run-abc123")
	want := filepath.Join(".agents", "context", "run-abc123")
	if got != want {
		t.Errorf("contextArtifactDir(\"run-abc123\") = %q, want %q", got, want)
	}
}

func TestContextArtifactDir_Empty(t *testing.T) {
	got := contextArtifactDir("")
	prefix := filepath.Join(".agents", "context", "adhoc-")
	if !strings.HasPrefix(got, prefix) {
		t.Errorf("contextArtifactDir(\"\") = %q, want prefix %q", got, prefix)
	}
	// Verify the suffix after "adhoc-" is numeric
	suffix := strings.TrimPrefix(got, prefix)
	if suffix == "" {
		t.Errorf("contextArtifactDir(\"\") suffix is empty, expected numeric timestamp")
	}
	for _, c := range suffix {
		if c < '0' || c > '9' {
			t.Errorf("contextArtifactDir(\"\") suffix %q contains non-numeric character %q", suffix, string(c))
			break
		}
	}
}

func TestEnsureContextDir_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	got, err := ensureContextDir(tmpDir, "test-run")
	if err != nil {
		t.Fatalf("ensureContextDir(%q, \"test-run\") error: %v", tmpDir, err)
	}
	wantSuffix := filepath.Join(".agents", "context", "test-run")
	if !strings.HasSuffix(got, wantSuffix) {
		t.Errorf("ensureContextDir returned %q, want suffix %q", got, wantSuffix)
	}
	info, err := os.Stat(got)
	if err != nil {
		t.Fatalf("os.Stat(%q) error: %v", got, err)
	}
	if !info.IsDir() {
		t.Errorf("%q is not a directory", got)
	}
}
