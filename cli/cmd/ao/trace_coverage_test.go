package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/provenance"
	"github.com/boshu2/agentops/cli/internal/storage"
)

// ---------------------------------------------------------------------------
// trace.go — traceOneArtifact
// ---------------------------------------------------------------------------

func TestCov3_trace_traceOneArtifact_jsonOutput(t *testing.T) {
	// Set up a temp dir with a provenance graph file
	tmpDir := t.TempDir()
	provDir := filepath.Join(tmpDir, storage.DefaultBaseDir, storage.ProvenanceDir)
	if err := os.MkdirAll(provDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	record := provenance.Record{
		ID:           "prov-abc1234",
		ArtifactPath: "sessions/test-session.md",
		ArtifactType: "session",
		SourcePath:   "/tmp/transcript.jsonl",
		SourceType:   "transcript",
		SessionID:    "session-abc123",
		CreatedAt:    time.Now(),
	}
	data, _ := json.Marshal(record)

	provFile := filepath.Join(provDir, storage.ProvenanceFile)
	if err := os.WriteFile(provFile, append(data, '\n'), 0644); err != nil {
		t.Fatalf("write prov: %v", err)
	}

	graph, err := provenance.NewGraph(provFile)
	if err != nil {
		t.Fatalf("new graph: %v", err)
	}

	// Test JSON output mode
	oldOutput := output
	output = "json"
	defer func() { output = oldOutput }()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = traceOneArtifact(graph, "sessions/test-session.md")

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("traceOneArtifact: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "session-abc123") {
		t.Errorf("expected JSON output to contain session ID, got: %s", out)
	}
}

func TestCov3_trace_traceOneArtifact_noProvenance(t *testing.T) {
	// Create an empty provenance graph
	tmpDir := t.TempDir()
	provDir := filepath.Join(tmpDir, storage.DefaultBaseDir, storage.ProvenanceDir)
	if err := os.MkdirAll(provDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	provFile := filepath.Join(provDir, storage.ProvenanceFile)
	if err := os.WriteFile(provFile, []byte{}, 0644); err != nil {
		t.Fatalf("write prov: %v", err)
	}

	graph, err := provenance.NewGraph(provFile)
	if err != nil {
		t.Fatalf("new graph: %v", err)
	}

	// Test table output with no chain
	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = traceOneArtifact(graph, "nonexistent-artifact.md")

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("traceOneArtifact: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "No provenance found") {
		t.Errorf("expected 'No provenance found', got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// trace.go — printTraceGraph
// ---------------------------------------------------------------------------

func TestCov3_trace_printTraceGraph(t *testing.T) {
	result := &provenance.TraceResult{
		Artifact: "sessions/test.md",
		Chain: []provenance.Record{
			{
				ID:           "prov-001",
				ArtifactPath: "sessions/test.md",
				ArtifactType: "session",
				SourcePath:   "/tmp/transcript.jsonl",
				SourceType:   "transcript",
				SessionID:    "sess-1234567890ab",
				CreatedAt:    time.Now(),
			},
		},
		Sources: []string{"/tmp/transcript.jsonl"},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printTraceGraph(result)

	_ = w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	// Should contain graph symbols and session ID (truncated to 12)
	if !strings.Contains(out, "Provenance Graph for:") {
		t.Errorf("expected graph header, got: %s", out)
	}
	if !strings.Contains(out, "sess-1234567") {
		t.Errorf("expected truncated session ID, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// trace.go — repeatString, min
// ---------------------------------------------------------------------------

func TestCov3_trace_repeatString(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"=", 0, ""},
		{"=", 3, "==="},
		{"ab", 2, "abab"},
		{"", 5, ""},
	}
	for _, tc := range tests {
		got := repeatString(tc.s, tc.n)
		if got != tc.want {
			t.Errorf("repeatString(%q, %d) = %q, want %q", tc.s, tc.n, got, tc.want)
		}
	}
}

func TestCov3_trace_min(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{3, 5, 3},
		{5, 3, 3},
		{0, 0, 0},
		{-1, 1, -1},
	}
	for _, tc := range tests {
		got := min(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("min(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
