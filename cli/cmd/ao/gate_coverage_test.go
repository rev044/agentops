package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/pool"
	"github.com/boshu2/agentops/cli/internal/types"
)

// ---------------------------------------------------------------------------
// gate.go — entryUrgency
// ---------------------------------------------------------------------------

func TestCov3_gate_entryUrgency(t *testing.T) {
	tests := []struct {
		name  string
		entry pool.PoolEntry
		want  string
	}{
		{
			name: "approaching auto promote is HIGH",
			entry: pool.PoolEntry{
				ApproachingAutoPromote: true,
				Age:                   23 * time.Hour,
			},
			want: "HIGH (approaching 24h)",
		},
		{
			name: "over 12h is MEDIUM",
			entry: pool.PoolEntry{
				ApproachingAutoPromote: false,
				Age:                   13 * time.Hour,
			},
			want: "MEDIUM",
		},
		{
			name: "under 12h is LOW",
			entry: pool.PoolEntry{
				ApproachingAutoPromote: false,
				Age:                   6 * time.Hour,
			},
			want: "LOW",
		},
		{
			name: "zero age is LOW",
			entry: pool.PoolEntry{
				ApproachingAutoPromote: false,
				Age:                   0,
			},
			want: "LOW",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := entryUrgency(tc.entry)
			if got != tc.want {
				t.Errorf("entryUrgency() = %q, want %q", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// gate.go — outputGatePending (table mode)
// ---------------------------------------------------------------------------

func TestCov3_gate_outputGatePending_table(t *testing.T) {
	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	t.Run("empty list", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputGatePending(nil)

		_ = w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Fatalf("outputGatePending: %v", err)
		}

		buf := make([]byte, 4096)
		n, _ := r.Read(buf)
		_ = r.Close()
		out := string(buf[:n])

		if !strings.Contains(out, "No pending reviews") {
			t.Errorf("expected 'No pending reviews', got: %s", out)
		}
	})

	t.Run("with entries", func(t *testing.T) {
		entries := []pool.PoolEntry{
			{
				PoolEntry: types.PoolEntry{
					Candidate: types.Candidate{
						ID:      "gate-entry-1",
						Tier:    types.TierBronze,
						Utility: 0.55,
					},
				},
				AgeString:              "18h",
				Age:                    18 * time.Hour,
				ApproachingAutoPromote: false,
			},
			{
				PoolEntry: types.PoolEntry{
					Candidate: types.Candidate{
						ID:      "gate-entry-2",
						Tier:    types.TierBronze,
						Utility: 0.60,
					},
				},
				AgeString:              "23h",
				Age:                    23 * time.Hour,
				ApproachingAutoPromote: true,
			},
		}

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputGatePending(entries)

		_ = w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Fatalf("outputGatePending: %v", err)
		}

		buf := make([]byte, 8192)
		n, _ := r.Read(buf)
		_ = r.Close()
		out := string(buf[:n])

		if !strings.Contains(out, "Pending Reviews (2)") {
			t.Errorf("expected 'Pending Reviews (2)', got: %s", out)
		}
		if !strings.Contains(out, "approaching 24h") {
			t.Errorf("expected auto-promote warning, got: %s", out)
		}
	})
}

// ---------------------------------------------------------------------------
// gate.go — outputGatePending (JSON mode)
// ---------------------------------------------------------------------------

func TestCov3_gate_outputGatePending_json(t *testing.T) {
	oldOutput := output
	output = "json"
	defer func() { output = oldOutput }()

	entries := []pool.PoolEntry{
		{
			PoolEntry: types.PoolEntry{
				Candidate: types.Candidate{
					ID:   "gate-json-1",
					Tier: types.TierBronze,
				},
				Status: types.PoolStatusPending,
			},
			AgeString: "5h",
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputGatePending(entries)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputGatePending: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "gate-json-1") {
		t.Errorf("expected JSON with candidate ID, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// gate.go — printAutoPromoteWarning
// ---------------------------------------------------------------------------

func TestCov3_gate_printAutoPromoteWarning(t *testing.T) {
	t.Run("no approaching entries", func(t *testing.T) {
		entries := []pool.PoolEntry{
			{ApproachingAutoPromote: false},
			{ApproachingAutoPromote: false},
		}

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		printAutoPromoteWarning(entries)

		_ = w.Close()
		os.Stdout = oldStdout

		buf := make([]byte, 4096)
		n, _ := r.Read(buf)
		_ = r.Close()
		out := string(buf[:n])

		if strings.Contains(out, "approaching 24h") {
			t.Errorf("should not show warning when no entries approaching, got: %s", out)
		}
	})

	t.Run("some approaching entries", func(t *testing.T) {
		entries := []pool.PoolEntry{
			{ApproachingAutoPromote: false},
			{ApproachingAutoPromote: true},
			{ApproachingAutoPromote: true},
		}

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		printAutoPromoteWarning(entries)

		_ = w.Close()
		os.Stdout = oldStdout

		buf := make([]byte, 4096)
		n, _ := r.Read(buf)
		_ = r.Close()
		out := string(buf[:n])

		if !strings.Contains(out, "2 candidate(s) approaching 24h") {
			t.Errorf("expected warning about 2 approaching candidates, got: %s", out)
		}
	})
}
