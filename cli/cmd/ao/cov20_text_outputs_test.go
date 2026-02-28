package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/boshu2/agentops/cli/internal/ratchet"
	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// session_outcome.go — text output path (lines 257-271, 7 stmts)
// ---------------------------------------------------------------------------

func TestCov20_runSessionOutcome_textOutput(t *testing.T) {
	tmp := t.TempDir()
	// Empty transcript — analyzeTranscript succeeds with zero lines
	transcriptFile := filepath.Join(tmp, "session.jsonl")
	if err := os.WriteFile(transcriptFile, []byte(""), 0644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}

	origDryRun := dryRun
	origOutput := sessionOutcomeOutput
	defer func() {
		dryRun = origDryRun
		sessionOutcomeOutput = origOutput
	}()
	dryRun = false
	sessionOutcomeOutput = "text" // triggers default: branch

	err := runSessionOutcome(sessionOutcomeCmd, []string{transcriptFile})
	if err != nil {
		t.Fatalf("runSessionOutcome text output: %v", err)
	}
}

func TestCov20_runSessionOutcome_textOutputWithLines(t *testing.T) {
	tmp := t.TempDir()
	// Transcript with a line to exercise the Signals loop
	content := `{"type":"assistant","message":{"content":"ok"},"sessionId":"test-session-abc"}` + "\n"
	transcriptFile := filepath.Join(tmp, "session2.jsonl")
	if err := os.WriteFile(transcriptFile, []byte(content), 0644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}

	origDryRun := dryRun
	origOutput := sessionOutcomeOutput
	defer func() {
		dryRun = origDryRun
		sessionOutcomeOutput = origOutput
	}()
	dryRun = false
	sessionOutcomeOutput = "text"

	err := runSessionOutcome(sessionOutcomeCmd, []string{transcriptFile})
	if err != nil {
		t.Fatalf("runSessionOutcome text output with lines: %v", err)
	}
}

// ---------------------------------------------------------------------------
// contradict.go — text output path (lines 189-206, 6 stmts)
// ---------------------------------------------------------------------------

func TestCov20_runContradict_textOutput(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Create .agents/learnings/ with a minimal learning file
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir learnings: %v", err)
	}
	learningContent := `---
type: pattern
confidence: high
---
# Test Learning

Some content that describes a pattern about always doing X.
`
	if err := os.WriteFile(filepath.Join(learningsDir, "2026-01-01-test.md"), []byte(learningContent), 0644); err != nil {
		t.Fatalf("write learning: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "" // non-json → text path

	err := runContradict(contradictCmd, nil)
	if err != nil {
		t.Fatalf("runContradict text output: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ratchet_record.go — dry-run path (lines 52-57, 5 stmts)
// ---------------------------------------------------------------------------

func TestCov20_runRatchetRecord_dryRun(t *testing.T) {
	cov4RatchetSetup(t)

	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	cmd := &cobra.Command{}
	err := runRatchetRecord(cmd, []string{"research"})
	if err != nil {
		t.Fatalf("runRatchetRecord dry-run: %v", err)
	}
}

func TestCov20_runRatchetRecord_dryRunPlan(t *testing.T) {
	cov4RatchetSetup(t)

	origDryRun := dryRun
	origInput := ratchetInput
	origOutput := ratchetOutput
	defer func() {
		dryRun = origDryRun
		ratchetInput = origInput
		ratchetOutput = origOutput
	}()
	dryRun = true
	ratchetInput = "PLAN.md"
	ratchetOutput = "plan-output.md"

	cmd := &cobra.Command{}
	err := runRatchetRecord(cmd, []string{"plan"})
	if err != nil {
		t.Fatalf("runRatchetRecord dry-run plan: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ratchet_status.go — entry details path (lines 68-75, 6 stmts)
// ---------------------------------------------------------------------------

func TestCov20_runRatchetStatus_withEntry(t *testing.T) {
	tmp := cov4RatchetSetup(t)

	// Append a chain entry so chain.GetLatest(step) != nil
	chain, err := ratchet.LoadChain(tmp)
	if err != nil {
		t.Fatalf("LoadChain: %v", err)
	}
	entry := ratchet.ChainEntry{
		Step:       ratchet.StepResearch,
		Output:     "research-output.md",
		Input:      "research-input.md",
		Location:   "research",
		Cycle:      1,
		ParentEpic: "ag-test",
	}
	if err := chain.Append(entry); err != nil {
		t.Fatalf("chain.Append: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	err = runRatchetStatus(cmd, nil)
	if err != nil {
		t.Fatalf("runRatchetStatus with entry: %v", err)
	}
}

// ---------------------------------------------------------------------------
// batch_forge.go — dry-run with unforged transcripts (lines 115-120, 5 stmts)
// ---------------------------------------------------------------------------

func TestCov20_runForgeBatch_dryRunWithTranscripts(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Create a .jsonl file > 100 bytes (isBatchTranscriptCandidate threshold)
	transcriptsDir := filepath.Join(tmp, "sessions")
	if err := os.MkdirAll(transcriptsDir, 0755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}
	// 150-byte content with valid JSON-ish line
	bigContent := bytes.Repeat([]byte{'x'}, 150)
	if err := os.WriteFile(filepath.Join(transcriptsDir, "session-abc123.jsonl"), bigContent, 0644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}

	origBatchDir := batchDir
	origDryRun := dryRun
	defer func() {
		batchDir = origBatchDir
		dryRun = origDryRun
	}()
	batchDir = transcriptsDir
	dryRun = true

	cmd := &cobra.Command{}
	err := runForgeBatch(cmd, nil)
	if err != nil {
		t.Fatalf("runForgeBatch dry-run with transcripts: %v", err)
	}
}

func TestCov20_runForgeBatch_dryRunMultipleTranscripts(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	transcriptsDir := filepath.Join(tmp, "multi-sessions")
	if err := os.MkdirAll(transcriptsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Two transcripts > 100 bytes each
	bigContent := bytes.Repeat([]byte{'y'}, 200)
	for _, name := range []string{"sess-001.jsonl", "sess-002.jsonl"} {
		if err := os.WriteFile(filepath.Join(transcriptsDir, name), bigContent, 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
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
	dryRun = true
	batchMax = 0 // no limit

	cmd := &cobra.Command{}
	err := runForgeBatch(cmd, nil)
	if err != nil {
		t.Fatalf("runForgeBatch dry-run multiple transcripts: %v", err)
	}
}
