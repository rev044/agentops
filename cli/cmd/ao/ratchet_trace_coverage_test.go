package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/ratchet"
)

// ---------------------------------------------------------------------------
// ratchet_trace.go — buildTrace
// ---------------------------------------------------------------------------

func TestCov3_ratchetTrace_buildTrace_emptyChain(t *testing.T) {
	chain := &ratchet.Chain{
		ID:      "test-chain-1",
		Started: time.Now(),
		Entries: []ratchet.ChainEntry{},
	}

	trace := buildTrace("artifact.md", chain)

	if trace.Artifact != "artifact.md" {
		t.Errorf("expected artifact 'artifact.md', got %q", trace.Artifact)
	}
	if len(trace.Chain) != 0 {
		t.Errorf("expected empty chain, got %d entries", len(trace.Chain))
	}
}

func TestCov3_ratchetTrace_buildTrace_matchesOutput(t *testing.T) {
	now := time.Now()
	chain := &ratchet.Chain{
		ID:      "test-chain-2",
		Started: now,
		Entries: []ratchet.ChainEntry{
			{
				Step:      ratchet.StepResearch,
				Input:     "transcript.jsonl",
				Output:    "research-notes.md",
				Timestamp: now.Add(-2 * time.Hour),
			},
			{
				Step:      ratchet.StepPlan,
				Input:     "research-notes.md",
				Output:    "plan.md",
				Timestamp: now.Add(-1 * time.Hour),
			},
			{
				Step:      ratchet.StepImplement,
				Input:     "plan.md",
				Output:    "result.md",
				Timestamp: now,
			},
		},
	}

	// Trace from result.md backward
	trace := buildTrace("result.md", chain)

	if trace.Artifact != "result.md" {
		t.Errorf("expected artifact 'result.md', got %q", trace.Artifact)
	}

	// Should trace back: result.md <- plan.md <- research-notes.md
	if len(trace.Chain) != 3 {
		t.Errorf("expected 3 chain entries, got %d", len(trace.Chain))
	}

	// Verify chain order is forward (reversed from backward walk)
	if len(trace.Chain) >= 3 {
		if trace.Chain[0].Step != ratchet.StepResearch {
			t.Errorf("first entry step: got %q, want %q", trace.Chain[0].Step, ratchet.StepResearch)
		}
		if trace.Chain[2].Step != ratchet.StepImplement {
			t.Errorf("last entry step: got %q, want %q", trace.Chain[2].Step, ratchet.StepImplement)
		}
	}
}

func TestCov3_ratchetTrace_buildTrace_suffixMatch(t *testing.T) {
	now := time.Now()
	chain := &ratchet.Chain{
		ID:      "test-chain-3",
		Started: now,
		Entries: []ratchet.ChainEntry{
			{
				Step:      ratchet.StepVibe,
				Input:     "input.md",
				Output:    "/full/path/to/artifact.md",
				Timestamp: now,
			},
		},
	}

	// Should match by suffix
	trace := buildTrace("artifact.md", chain)

	if len(trace.Chain) != 1 {
		t.Errorf("expected 1 chain entry via suffix match, got %d", len(trace.Chain))
	}
}

// ---------------------------------------------------------------------------
// ratchet_trace.go — outputTrace (JSON mode)
// ---------------------------------------------------------------------------

func TestCov3_ratchetTrace_outputTrace_json(t *testing.T) {
	oldOutput := output
	output = "json"
	defer func() { output = oldOutput }()

	trace := traceResult{
		Artifact: "test.md",
		Chain: []traceEntry{
			{
				Step:   ratchet.StepResearch,
				Input:  "transcript.jsonl",
				Output: "test.md",
				Time:   time.Now().Format(time.RFC3339),
			},
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputTrace(trace)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputTrace: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, `"artifact"`) {
		t.Errorf("expected JSON with 'artifact' field, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// ratchet_trace.go — outputTraceText
// ---------------------------------------------------------------------------

func TestCov3_ratchetTrace_outputTraceText_emptyChain(t *testing.T) {
	trace := traceResult{
		Artifact: "missing.md",
		Chain:    nil,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputTraceText(trace)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputTraceText: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "No provenance chain found") {
		t.Errorf("expected 'No provenance chain found', got: %s", out)
	}
}

func TestCov3_ratchetTrace_outputTraceText_withEntries(t *testing.T) {
	trace := traceResult{
		Artifact: "output.md",
		Chain: []traceEntry{
			{
				Step:   ratchet.StepResearch,
				Input:  "transcript.jsonl",
				Output: "research.md",
				Time:   "2026-01-15T10:00:00Z",
			},
			{
				Step:   ratchet.StepPlan,
				Input:  "research.md",
				Output: "output.md",
				Time:   "2026-01-15T11:00:00Z",
			},
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputTraceText(trace)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputTraceText: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "Provenance Trace: output.md") {
		t.Errorf("expected header, got: %s", out)
	}
	if !strings.Contains(out, "1. research") {
		t.Errorf("expected step 1 research, got: %s", out)
	}
	if !strings.Contains(out, "2. plan") {
		t.Errorf("expected step 2 plan, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// ratchet_trace.go — runRatchetTrace error path
// ---------------------------------------------------------------------------

func TestCov3_ratchetTrace_runRatchetTrace_noAgentsDir(t *testing.T) {
	tmpDir := t.TempDir()
	prev, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(prev) }()

	// Create .agents dir so LoadChain doesn't fail but chain is empty
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents", "ao"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runRatchetTrace(nil, []string{"nonexistent.md"})

	_ = w.Close()
	os.Stdout = oldStdout

	// Should succeed with empty trace
	if err != nil {
		t.Fatalf("runRatchetTrace: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "No provenance chain found") {
		t.Errorf("expected no provenance message, got: %s", out)
	}
}
