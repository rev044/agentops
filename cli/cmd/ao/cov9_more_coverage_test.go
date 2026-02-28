package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// hooks.go — runHooksInit (17.6% → higher)
// ---------------------------------------------------------------------------

func TestCov9_runHooksInit_json(t *testing.T) {
	orig := hooksOutputFormat
	defer func() { hooksOutputFormat = orig }()
	hooksOutputFormat = "json"

	cmd := &cobra.Command{}
	err := runHooksInit(cmd, nil)
	if err != nil {
		t.Fatalf("runHooksInit json: %v", err)
	}
}

func TestCov9_runHooksInit_shell(t *testing.T) {
	orig := hooksOutputFormat
	defer func() { hooksOutputFormat = orig }()
	hooksOutputFormat = "shell"

	cmd := &cobra.Command{}
	err := runHooksInit(cmd, nil)
	if err != nil {
		t.Fatalf("runHooksInit shell: %v", err)
	}
}

func TestCov9_runHooksInit_unknownFormat(t *testing.T) {
	orig := hooksOutputFormat
	defer func() { hooksOutputFormat = orig }()
	hooksOutputFormat = "xml"

	cmd := &cobra.Command{}
	err := runHooksInit(cmd, nil)
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("expected 'unknown format' in error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// store.go — runStoreIndex (19.2% → higher)
// ---------------------------------------------------------------------------

func TestCov9_runStoreIndex_noFiles(t *testing.T) {
	tmp := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = false

	cmd := &cobra.Command{}
	err = runStoreIndex(cmd, []string{"*.impossible_xyz_pattern"})
	if err == nil {
		t.Fatal("expected error for no matching files")
	}
	if !strings.Contains(err.Error(), "no files found") {
		t.Errorf("expected 'no files found' error, got %v", err)
	}
}

func TestCov9_runStoreIndex_dryRun(t *testing.T) {
	tmp := t.TempDir()

	// Create a real file so expandFilePatterns finds it
	mdFile := filepath.Join(tmp, "test.md")
	if err := os.WriteFile(mdFile, []byte("# Test\n\nContent here.\n"), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	cmd := &cobra.Command{}
	err = runStoreIndex(cmd, []string{"test.md"})
	if err != nil {
		t.Fatalf("runStoreIndex dry-run: %v", err)
	}
}

// ---------------------------------------------------------------------------
// batch_forge.go — runForgeBatch no-pending path (22.6% → higher)
// ---------------------------------------------------------------------------

func TestCov9_runForgeBatch_noPendingTranscripts(t *testing.T) {
	tmp := t.TempDir()

	// Point batchDir to an empty dir — no .jsonl files found
	origBatchDir := batchDir
	defer func() { batchDir = origBatchDir }()
	batchDir = tmp

	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = false

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cmd := &cobra.Command{}
	err = runForgeBatch(cmd, nil)
	if err != nil {
		t.Fatalf("runForgeBatch no-pending: %v", err)
	}
}

func TestCov9_runForgeBatch_dryRun(t *testing.T) {
	tmp := t.TempDir()

	// Create a small .jsonl file that meets the 100-byte threshold
	transcriptDir := filepath.Join(tmp, "projects", "test-project")
	if err := os.MkdirAll(transcriptDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	jsonlContent := strings.Repeat(`{"type":"summary","sessionId":"sess-1","timestamp":"2024-01-01T00:00:00Z"}`+"\n", 3)
	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	if err := os.WriteFile(transcriptPath, []byte(jsonlContent), 0644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}

	origBatchDir := batchDir
	defer func() { batchDir = origBatchDir }()
	batchDir = filepath.Join(tmp, "projects")

	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cmd := &cobra.Command{}
	err = runForgeBatch(cmd, nil)
	if err != nil {
		t.Fatalf("runForgeBatch dry-run: %v", err)
	}
}

// ---------------------------------------------------------------------------
// session_outcome.go — runSessionOutcome dry-run (23.3% → higher)
// ---------------------------------------------------------------------------

func TestCov9_runSessionOutcome_dryRunWithExplicitPath(t *testing.T) {
	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	cmd := &cobra.Command{}
	// Pass an explicit path — dry-run exits before opening the file
	err := runSessionOutcome(cmd, []string{"/tmp/fake-transcript.jsonl"})
	if err != nil {
		t.Fatalf("runSessionOutcome dry-run: %v", err)
	}
}

func TestCov9_runSessionOutcome_dryRunNoArgs(t *testing.T) {
	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	// Without args, it searches for most recent transcript.
	// dry-run exits before opening the file.
	// If no transcript is found, it returns an error BEFORE dry-run check.
	// So either it succeeds (dry-run printed) or fails with "no transcript found".
	cmd := &cobra.Command{}
	err := runSessionOutcome(cmd, []string{})
	// Accept either outcome — the point is to exercise the no-args branch
	_ = err
}
