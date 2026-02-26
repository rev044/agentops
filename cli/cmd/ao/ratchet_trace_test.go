package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/ratchet"
)

func TestBuildTrace_EmptyChain(t *testing.T) {
	chain := &ratchet.Chain{
		ID:      "test-chain",
		Started: time.Now(),
		Entries: []ratchet.ChainEntry{},
	}

	trace := buildTrace("some-artifact.md", chain)

	if trace.Artifact != "some-artifact.md" {
		t.Errorf("Artifact = %q, want %q", trace.Artifact, "some-artifact.md")
	}
	if len(trace.Chain) != 0 {
		t.Errorf("expected empty trace chain, got %d entries", len(trace.Chain))
	}
}

func TestBuildTrace_SingleEntry(t *testing.T) {
	now := time.Now()
	chain := &ratchet.Chain{
		ID:      "test",
		Started: now,
		Entries: []ratchet.ChainEntry{
			{
				Step:      ratchet.StepResearch,
				Timestamp: now,
				Input:     "",
				Output:    ".agents/research/findings.md",
				Locked:    true,
			},
		},
	}

	trace := buildTrace(".agents/research/findings.md", chain)

	if len(trace.Chain) != 1 {
		t.Fatalf("expected 1 trace entry, got %d", len(trace.Chain))
	}

	entry := trace.Chain[0]
	if entry.Step != ratchet.StepResearch {
		t.Errorf("Step = %q, want %q", entry.Step, ratchet.StepResearch)
	}
	if entry.Output != ".agents/research/findings.md" {
		t.Errorf("Output = %q", entry.Output)
	}
}

func TestBuildTrace_MultiStepProvenance(t *testing.T) {
	now := time.Now()
	chain := &ratchet.Chain{
		ID:      "test",
		Started: now,
		Entries: []ratchet.ChainEntry{
			{
				Step:      ratchet.StepResearch,
				Timestamp: now,
				Input:     "",
				Output:    "research.md",
				Locked:    true,
			},
			{
				Step:      ratchet.StepPreMortem,
				Timestamp: now.Add(time.Hour),
				Input:     "research.md",
				Output:    "spec-v2.md",
				Locked:    true,
			},
			{
				Step:      ratchet.StepPlan,
				Timestamp: now.Add(2 * time.Hour),
				Input:     "spec-v2.md",
				Output:    "epic:ol-0001",
				Locked:    true,
			},
		},
	}

	trace := buildTrace("epic:ol-0001", chain)

	if len(trace.Chain) != 3 {
		t.Fatalf("expected 3 trace entries, got %d", len(trace.Chain))
	}

	// Verify chain order (oldest first)
	if trace.Chain[0].Step != ratchet.StepResearch {
		t.Errorf("first entry step = %q, want research", trace.Chain[0].Step)
	}
	if trace.Chain[1].Step != ratchet.StepPreMortem {
		t.Errorf("second entry step = %q, want pre-mortem", trace.Chain[1].Step)
	}
	if trace.Chain[2].Step != ratchet.StepPlan {
		t.Errorf("third entry step = %q, want plan", trace.Chain[2].Step)
	}
}

func TestBuildTrace_SuffixMatching(t *testing.T) {
	now := time.Now()
	chain := &ratchet.Chain{
		ID:      "test",
		Started: now,
		Entries: []ratchet.ChainEntry{
			{
				Step:      ratchet.StepResearch,
				Timestamp: now,
				Output:    ".agents/research/findings.md",
				Locked:    true,
			},
		},
	}

	// buildTrace uses HasSuffix matching
	trace := buildTrace("findings.md", chain)

	if len(trace.Chain) != 1 {
		t.Errorf("suffix match should find entry, got %d", len(trace.Chain))
	}
}

func TestBuildTrace_NoMatchingArtifact(t *testing.T) {
	now := time.Now()
	chain := &ratchet.Chain{
		ID:      "test",
		Started: now,
		Entries: []ratchet.ChainEntry{
			{
				Step:      ratchet.StepResearch,
				Timestamp: now,
				Output:    "something-else.md",
				Locked:    true,
			},
		},
	}

	trace := buildTrace("nonexistent.md", chain)

	if len(trace.Chain) != 0 {
		t.Errorf("expected empty trace, got %d entries", len(trace.Chain))
	}
}

func TestBuildTrace_TimeFormatting(t *testing.T) {
	ts := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	chain := &ratchet.Chain{
		ID:      "test",
		Started: ts,
		Entries: []ratchet.ChainEntry{
			{
				Step:      ratchet.StepResearch,
				Timestamp: ts,
				Output:    "output.md",
				Locked:    true,
			},
		},
	}

	trace := buildTrace("output.md", chain)

	if len(trace.Chain) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(trace.Chain))
	}

	// Time should be RFC3339
	if !strings.Contains(trace.Chain[0].Time, "2025-01-15") {
		t.Errorf("Time = %q, expected RFC3339 with date 2025-01-15", trace.Chain[0].Time)
	}
}

func TestOutputTrace_JSON(t *testing.T) {
	origOutput := output
	output = "json"
	defer func() { output = origOutput }()

	trace := traceResult{
		Artifact: "test.md",
		Chain: []traceEntry{
			{
				Step:   ratchet.StepResearch,
				Input:  "",
				Output: "test.md",
				Time:   "2025-01-15T10:00:00Z",
			},
		},
	}

	// outputTrace writes to os.Stdout, test the JSON encoding logic
	data, err := json.Marshal(trace)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	if !bytes.Contains(data, []byte("test.md")) {
		t.Error("JSON output missing artifact name")
	}
	if !bytes.Contains(data, []byte("research")) {
		t.Error("JSON output missing step name")
	}
}

func TestOutputTraceText_EmptyChain(t *testing.T) {
	// Just verify it doesn't panic
	trace := traceResult{
		Artifact: "missing.md",
		Chain:    []traceEntry{},
	}

	// outputTraceText writes to stdout; capture is not straightforward
	// but we can at least verify it doesn't error
	err := outputTraceText(trace)
	if err != nil {
		t.Errorf("outputTraceText(empty) error = %v", err)
	}
}

func TestTraceEntry_Structure(t *testing.T) {
	entry := traceEntry{
		Step:   ratchet.StepPlan,
		Input:  "spec-v2.md",
		Output: "epic:ol-0001",
		Time:   "2025-01-15T10:00:00Z",
	}

	if entry.Step != ratchet.StepPlan {
		t.Errorf("Step = %q", entry.Step)
	}
	if entry.Input != "spec-v2.md" {
		t.Errorf("Input = %q", entry.Input)
	}
	if entry.Output != "epic:ol-0001" {
		t.Errorf("Output = %q", entry.Output)
	}
}

func TestTraceResult_JSONSerialization(t *testing.T) {
	trace := traceResult{
		Artifact: "retro.md",
		Chain: []traceEntry{
			{Step: ratchet.StepResearch, Output: "findings.md", Time: "2025-01-15T10:00:00Z"},
			{Step: ratchet.StepPlan, Input: "findings.md", Output: "epic:ol-001", Time: "2025-01-15T11:00:00Z"},
		},
	}

	data, err := json.MarshalIndent(trace, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}

	var decoded traceResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Artifact != trace.Artifact {
		t.Errorf("Artifact = %q, want %q", decoded.Artifact, trace.Artifact)
	}
	if len(decoded.Chain) != len(trace.Chain) {
		t.Errorf("Chain length = %d, want %d", len(decoded.Chain), len(trace.Chain))
	}
}
