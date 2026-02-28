package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// contradict.go — if result.Contradictions > 0 branch (lines 197-203, 5 stmts)
// ---------------------------------------------------------------------------

func TestCov21_runContradict_withContradictions(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir learnings: %v", err)
	}

	// File A: positive assertion — no negation words.
	// File B: same topic with "never" → negation asymmetry.
	// Both share 14 tokens (≥3 chars) → Jaccard ≥ 0.93 > 0.4.
	bodyA := "use this pattern for your code architecture design build test deploy system database service"
	bodyB := "never use this pattern for your code architecture design build test deploy system database service"

	if err := os.WriteFile(filepath.Join(learningsDir, "2026-01-01-pattern-a.md"), []byte(bodyA), 0644); err != nil {
		t.Fatalf("write A: %v", err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "2026-01-01-pattern-b.md"), []byte(bodyB), 0644); err != nil {
		t.Fatalf("write B: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "" // text mode

	err := runContradict(contradictCmd, nil)
	if err != nil {
		t.Fatalf("runContradict with contradictions: %v", err)
	}
}

// ---------------------------------------------------------------------------
// vibe_check.go — parseDuration invalid format errors (lines 118, 128; 2 stmts)
// ---------------------------------------------------------------------------

func TestCov21_parseDuration_invalidDays(t *testing.T) {
	// "xd" has suffix "d" but "x" is not a valid integer → fmt.Sscanf fails.
	_, err := parseDuration("xd")
	if err == nil {
		t.Fatal("parseDuration('xd'): expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid days format") {
		t.Errorf("parseDuration('xd'): expected 'invalid days format' in error, got: %v", err)
	}
}

func TestCov21_parseDuration_invalidWeeks(t *testing.T) {
	// "xw" has suffix "w" but "x" is not a valid integer → fmt.Sscanf fails.
	_, err := parseDuration("xw")
	if err == nil {
		t.Fatal("parseDuration('xw'): expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid weeks format") {
		t.Errorf("parseDuration('xw'): expected 'invalid weeks format' in error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// batch_forge.go — non-dry-run path (lines 123-150; 14 stmts)
// ---------------------------------------------------------------------------

func TestCov21_runForgeBatch_notDryRun_validTranscript(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	transcriptsDir := filepath.Join(tmp, "sessions")
	if err := os.MkdirAll(transcriptsDir, 0755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}

	// Valid JSONL transcript — parseable by the forge pipeline, > 100 bytes.
	validContent := strings.Join([]string{
		`{"type":"summary","sessionId":"sess-cov21","timestamp":"2024-01-01T00:00:00Z"}`,
		`{"type":"assistant","role":"assistant","content":"Implemented feature Y for test coverage","sessionId":"sess-cov21","timestamp":"2024-01-01T00:01:00Z"}`,
	}, "\n") + "\n"

	transcriptPath := filepath.Join(transcriptsDir, "session-cov21.jsonl")
	if err := os.WriteFile(transcriptPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}

	origBatchDir := batchDir
	origDryRun := dryRun
	origBatchMax := batchMax
	defer func() {
		batchDir = origBatchDir
		dryRun = origDryRun
		batchMax = origBatchMax
	}()
	batchDir = transcriptsDir
	dryRun = false
	batchMax = 0

	cmd := &cobra.Command{}
	err := runForgeBatch(cmd, nil)
	if err != nil {
		t.Fatalf("runForgeBatch non-dry-run valid transcript: %v", err)
	}
}
