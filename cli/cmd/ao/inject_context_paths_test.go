package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("boom")
}

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
	// Verify the suffix after "adhoc-" matches <timestamp>-<4hex>
	suffix := strings.TrimPrefix(got, prefix)
	if suffix == "" {
		t.Errorf("contextArtifactDir(\"\") suffix is empty, expected timestamp-hex")
	}
	parts := strings.SplitN(suffix, "-", 2)
	if len(parts) != 2 {
		t.Errorf("contextArtifactDir(\"\") suffix %q expected format <timestamp>-<hex>, got %d parts", suffix, len(parts))
	} else {
		for _, c := range parts[0] {
			if c < '0' || c > '9' {
				t.Errorf("contextArtifactDir(\"\") timestamp part %q contains non-numeric character %q", parts[0], string(c))
				break
			}
		}
		if len(parts[1]) != 4 {
			t.Errorf("contextArtifactDir(\"\") hex suffix %q expected 4 characters", parts[1])
		}
	}
}

func TestNewAdhocContextRunID_UsesCryptoSuffix(t *testing.T) {
	got := newAdhocContextRunID(time.Unix(1234, 0), strings.NewReader("\xab\xcd"))
	if got != "adhoc-1234-abcd" {
		t.Fatalf("newAdhocContextRunID() = %q, want %q", got, "adhoc-1234-abcd")
	}
}

func TestNewAdhocContextRunID_FallsBackToTimeBits(t *testing.T) {
	now := time.Unix(1234, 0).Add(0x1234)
	got := newAdhocContextRunID(now, errReader{})
	want := "adhoc-1234-c634"
	if got != want {
		t.Fatalf("newAdhocContextRunID() fallback = %q, want %q", got, want)
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
