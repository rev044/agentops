package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/types"
	"github.com/spf13/cobra"
)

func setupHealthTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// Create standard directories
	for _, sub := range []string{
		".agents/learnings",
		".agents/patterns",
		".agents/constraints",
		".agents/ao",
	} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func writeHealthCitations(t *testing.T, dir string, events []types.CitationEvent) {
	t.Helper()
	citDir := filepath.Join(dir, ".agents", "ao")
	if err := os.MkdirAll(citDir, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(citDir, "citations.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range events {
		if err := enc.Encode(e); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMetricsHealth_EmptyRepo(t *testing.T) {
	dir := t.TempDir()
	// No .agents/ directory at all

	hm, err := computeHealthMetrics(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hm.Sigma != 0 {
		t.Errorf("expected sigma=0, got %f", hm.Sigma)
	}
	if hm.Rho != 0 {
		t.Errorf("expected rho=0, got %f", hm.Rho)
	}
	if hm.Delta != 0 {
		t.Errorf("expected delta=0, got %f", hm.Delta)
	}
	if hm.EscapeVelocity {
		t.Error("expected escape_velocity=false for empty repo")
	}
	if hm.KnowledgeStock.Total != 0 {
		t.Errorf("expected knowledge_stock.total=0, got %d", hm.KnowledgeStock.Total)
	}
	if hm.LoopDominance.Dominant != "B1" {
		t.Errorf("expected loop_dominance.dominant=B1, got %s", hm.LoopDominance.Dominant)
	}
}

func TestMetricsHealth_WithCitations(t *testing.T) {
	dir := setupHealthTestDir(t)
	now := time.Now()

	// Create some learnings files
	learningsDir := filepath.Join(dir, ".agents", "learnings")
	for _, name := range []string{"a.md", "b.md", "c.md", "d.md"} {
		if err := os.WriteFile(filepath.Join(learningsDir, name), []byte("# Learning"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create citations: 2 of 4 learnings cited across 2 sessions
	citations := []types.CitationEvent{
		{
			ArtifactPath: filepath.Join(dir, ".agents", "learnings", "a.md"),
			SessionID:    "session-1",
			CitedAt:      now.AddDate(0, 0, -2),
			CitationType: "reference",
		},
		{
			ArtifactPath: filepath.Join(dir, ".agents", "learnings", "a.md"),
			SessionID:    "session-2",
			CitedAt:      now.AddDate(0, 0, -1),
			CitationType: "applied",
		},
		{
			ArtifactPath: filepath.Join(dir, ".agents", "learnings", "b.md"),
			SessionID:    "session-1",
			CitedAt:      now.AddDate(0, 0, -2),
			CitationType: "reference",
		},
	}
	writeHealthCitations(t, dir, citations)

	hm, err := computeHealthMetrics(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// sigma = unique cited (a.md, b.md = 2) / total retrievable (4 learnings + 0 patterns = 4) = 0.5
	if hm.Sigma < 0.49 || hm.Sigma > 0.51 {
		t.Errorf("expected sigma~0.5, got %f", hm.Sigma)
	}

	// rho = citation count (3) / total retrievable (4) = 0.75
	if hm.Rho < 0.74 || hm.Rho > 0.76 {
		t.Errorf("expected rho~0.75, got %f", hm.Rho)
	}

	// Knowledge stock
	if hm.KnowledgeStock.Learnings != 4 {
		t.Errorf("expected 4 learnings, got %d", hm.KnowledgeStock.Learnings)
	}
}

func TestMetricsHealth_EscapeVelocity_Positive(t *testing.T) {
	dir := setupHealthTestDir(t)
	now := time.Now()

	// Create 2 very recent learnings (low delta => low threshold)
	learningsDir := filepath.Join(dir, ".agents", "learnings")
	for _, name := range []string{"x.md", "y.md"} {
		if err := os.WriteFile(filepath.Join(learningsDir, name), []byte("# L"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Cite both heavily across multiple sessions => high sigma, high rho
	var citations []types.CitationEvent
	for i := 0; i < 5; i++ {
		citations = append(citations, types.CitationEvent{
			ArtifactPath: filepath.Join(dir, ".agents", "learnings", "x.md"),
			SessionID:    "s" + string(rune('1'+i)),
			CitedAt:      now.Add(-time.Duration(i) * time.Hour),
		})
		citations = append(citations, types.CitationEvent{
			ArtifactPath: filepath.Join(dir, ".agents", "learnings", "y.md"),
			SessionID:    "s" + string(rune('1'+i)),
			CitedAt:      now.Add(-time.Duration(i) * time.Hour),
		})
	}
	writeHealthCitations(t, dir, citations)

	hm, err := computeHealthMetrics(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// sigma should be 1.0 (both cited / both exist)
	// rho should be high (10 citations / 2 artifacts = 5.0)
	// delta should be very low (files just created, ~0 days)
	// So sigma*rho (1.0 * 5.0 = 5.0) > delta/100 (~0) => compounding
	if !hm.EscapeVelocity {
		t.Errorf("expected escape_velocity=true (compounding), got false; sigma=%f rho=%f delta=%f sigmaRho=%f threshold=%f",
			hm.Sigma, hm.Rho, hm.Delta, hm.Sigma*hm.Rho, hm.Delta/100.0)
	}
}

func TestMetricsHealth_EscapeVelocity_Negative(t *testing.T) {
	dir := setupHealthTestDir(t)

	// Create 10 old learnings with no citations at all => high delta, zero sigma*rho
	learningsDir := filepath.Join(dir, ".agents", "learnings")
	for i := 0; i < 10; i++ {
		name := filepath.Join(learningsDir, "old-"+string(rune('a'+i))+".md")
		if err := os.WriteFile(name, []byte("# Old Learning"), 0o644); err != nil {
			t.Fatal(err)
		}
		// Backdate the file to 60 days ago
		oldTime := time.Now().AddDate(0, 0, -60)
		if err := os.Chtimes(name, oldTime, oldTime); err != nil {
			t.Fatal(err)
		}
	}

	// No citations => sigma=0, rho=0
	hm, err := computeHealthMetrics(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// sigma*rho = 0 < delta/100 (~0.6) => decaying
	if hm.EscapeVelocity {
		t.Errorf("expected escape_velocity=false (decaying), got true; sigma=%f rho=%f delta=%f",
			hm.Sigma, hm.Rho, hm.Delta)
	}
	if hm.Delta < 50 {
		t.Errorf("expected delta >= 50 days for 60-day-old files, got %f", hm.Delta)
	}
}

func TestMetricsHealth_JSONOutput(t *testing.T) {
	dir := setupHealthTestDir(t)
	now := time.Now()

	// Create a learning and cite it
	learningsDir := filepath.Join(dir, ".agents", "learnings")
	if err := os.WriteFile(filepath.Join(learningsDir, "test.md"), []byte("# Test"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeHealthCitations(t, dir, []types.CitationEvent{
		{
			ArtifactPath: filepath.Join(dir, ".agents", "learnings", "test.md"),
			SessionID:    "s1",
			CitedAt:      now.AddDate(0, 0, -1),
		},
	})

	// Save and restore global state
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	oldOutput := output
	output = "json"
	defer func() { output = oldOutput }()

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := runMetricsHealth(cmd, nil); err != nil {
		t.Fatalf("runMetricsHealth failed: %v", err)
	}

	// Parse JSON output
	var parsed healthMetrics
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		t.Fatalf("expected valid JSON, got: %q (%v)", out.String(), err)
	}

	// Verify schema fields are present
	if parsed.Sigma < 0 {
		t.Error("expected sigma >= 0")
	}
	if parsed.KnowledgeStock.Total < 1 {
		t.Errorf("expected knowledge_stock.total >= 1, got %d", parsed.KnowledgeStock.Total)
	}
	if parsed.LoopDominance.Dominant == "" {
		t.Error("expected loop_dominance.dominant to be non-empty")
	}

	// Verify escape_velocity field exists (JSON will have it even if false)
	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(out.Bytes(), &raw); err != nil {
		t.Fatalf("failed to parse raw JSON: %v", err)
	}
	if _, ok := raw["escape_velocity"]; !ok {
		t.Error("expected escape_velocity field in JSON output")
	}
	if _, ok := raw["knowledge_stock"]; !ok {
		t.Error("expected knowledge_stock field in JSON output")
	}
	if _, ok := raw["loop_dominance"]; !ok {
		t.Error("expected loop_dominance field in JSON output")
	}
}
