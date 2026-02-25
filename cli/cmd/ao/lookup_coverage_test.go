package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// lookup.go — matchesID
// ---------------------------------------------------------------------------

func TestCov3_lookup_matchesID(t *testing.T) {
	tests := []struct {
		name     string
		itemID   string
		filePath string
		searchID string
		want     bool
	}{
		{
			name:     "exact ID match",
			itemID:   "learn-2026-01-20-cross-lang",
			searchID: "learn-2026-01-20-cross-lang",
			want:     true,
		},
		{
			name:     "case insensitive ID match",
			itemID:   "LEARN-ABC",
			searchID: "learn-abc",
			want:     true,
		},
		{
			name:     "filename match without extension",
			itemID:   "some-id",
			filePath: "/tmp/learn-2026-01-20-cross-lang.md",
			searchID: "learn-2026-01-20-cross-lang",
			want:     true,
		},
		{
			name:     "partial filename match",
			itemID:   "some-id",
			filePath: "/tmp/learn-2026-01-20-cross-lang.md",
			searchID: "cross-lang",
			want:     true,
		},
		{
			name:     "no match at all",
			itemID:   "id-alpha",
			filePath: "/tmp/beta.md",
			searchID: "gamma",
			want:     false,
		},
		{
			name:     "empty file path only checks ID",
			itemID:   "target-id",
			filePath: "",
			searchID: "target-id",
			want:     true,
		},
		{
			name:     "empty file path no match",
			itemID:   "alpha",
			filePath: "",
			searchID: "beta",
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesID(tc.itemID, tc.filePath, tc.searchID)
			if got != tc.want {
				t.Errorf("matchesID(%q, %q, %q) = %v, want %v",
					tc.itemID, tc.filePath, tc.searchID, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// lookup.go — filterByBead
// ---------------------------------------------------------------------------

func TestCov3_lookup_filterByBead(t *testing.T) {
	learnings := []learning{
		{ID: "l1", SourceBead: "ag-abc"},
		{ID: "l2", SourceBead: "ag-def"},
		{ID: "l3", SourceBead: "ag-abc"},
		{ID: "l4", SourceBead: ""},
	}

	t.Run("filters matching bead", func(t *testing.T) {
		filtered := filterByBead(learnings, "ag-abc")
		if len(filtered) != 2 {
			t.Errorf("expected 2 matches, got %d", len(filtered))
		}
		for _, l := range filtered {
			if l.SourceBead != "ag-abc" {
				t.Errorf("unexpected bead %q in results", l.SourceBead)
			}
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		filtered := filterByBead(learnings, "AG-ABC")
		if len(filtered) != 2 {
			t.Errorf("expected 2 matches (case insensitive), got %d", len(filtered))
		}
	})

	t.Run("no matches returns empty", func(t *testing.T) {
		filtered := filterByBead(learnings, "ag-xyz")
		if len(filtered) != 0 {
			t.Errorf("expected 0 matches, got %d", len(filtered))
		}
	})
}

// ---------------------------------------------------------------------------
// lookup.go — formatLookupAge
// ---------------------------------------------------------------------------

func TestCov3_lookup_formatLookupAge(t *testing.T) {
	tests := []struct {
		ageWeeks float64
		want     string
	}{
		{0.0, "<1d"},
		{0.13, "<1d"},
		{0.15, "1d"},
		{0.5, "4d"},
		{1.0, "1w"},
		{3.5, "4w"},
		{5.0, "1mo"},
		{8.57, "2mo"},
	}

	for _, tc := range tests {
		got := formatLookupAge(tc.ageWeeks)
		if got != tc.want {
			t.Errorf("formatLookupAge(%.2f) = %q, want %q", tc.ageWeeks, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// lookup.go — relPath
// ---------------------------------------------------------------------------

func TestCov3_lookup_relPath(t *testing.T) {
	cwd := "/Users/test/project"

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "subpath made relative",
			path: "/Users/test/project/src/file.go",
			want: "src/file.go",
		},
		{
			name: "unrelated path returns original",
			path: "/var/log/syslog",
			want: "../../../var/log/syslog",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := relPath(cwd, tc.path)
			if got != tc.want {
				t.Errorf("relPath(%q, %q) = %q, want %q", cwd, tc.path, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// lookup.go — outputResults
// ---------------------------------------------------------------------------

func TestCov3_lookup_outputResults_noResults(t *testing.T) {
	// Save and restore the module-level flag
	oldLookupJSON := lookupJSON
	lookupJSON = false
	defer func() { lookupJSON = oldLookupJSON }()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputResults("/tmp", nil, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputResults: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "No matching artifacts found") {
		t.Errorf("expected 'No matching artifacts found', got: %s", out)
	}
}

func TestCov3_lookup_outputResults_withLearnings(t *testing.T) {
	oldLookupJSON := lookupJSON
	lookupJSON = false
	defer func() { lookupJSON = oldLookupJSON }()

	learnings := []learning{
		{
			ID:             "l-test-1",
			Title:          "Test Learning",
			Summary:        "A summary",
			Source:         "/tmp/src.md",
			Utility:        0.75,
			AgeWeeks:       2.0,
			CompositeScore: 0.80,
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputResults("/tmp", learnings, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputResults: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, "l-test-1") {
		t.Errorf("expected learning ID in output, got: %s", out)
	}
	if !strings.Contains(out, "Test Learning") {
		t.Errorf("expected learning title in output, got: %s", out)
	}
}

func TestCov3_lookup_outputResults_jsonMode(t *testing.T) {
	oldLookupJSON := lookupJSON
	lookupJSON = true
	defer func() { lookupJSON = oldLookupJSON }()

	learnings := []learning{
		{ID: "l-json-1", Title: "JSON Test"},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputResults("/tmp", learnings, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputResults: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	if !strings.Contains(out, `"learnings"`) {
		t.Errorf("expected JSON with 'learnings' key, got: %s", out)
	}
}

// ---------------------------------------------------------------------------
// lookup.go — outputLearning (text mode, no-cite)
// ---------------------------------------------------------------------------

func TestCov3_lookup_outputLearning_textMode(t *testing.T) {
	oldLookupJSON := lookupJSON
	lookupJSON = false
	defer func() { lookupJSON = oldLookupJSON }()

	oldLookupNoCite := lookupNoCite
	lookupNoCite = true
	defer func() { lookupNoCite = oldLookupNoCite }()

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "test-learning.md")
	if err := os.WriteFile(srcFile, []byte("# Full Content\nDetails here."), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	l := learning{
		ID:             "learn-output-test",
		Title:          "Output Test Learning",
		Summary:        "A test summary",
		Source:         srcFile,
		SourceBead:     "ag-test",
		SourcePhase:    "implement",
		Utility:        0.80,
		AgeWeeks:       1.0,
		CompositeScore: 0.85,
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputLearning(tmpDir, l)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("outputLearning: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	_ = r.Close()
	out := string(buf[:n])

	checks := []string{
		"learn-output-test",
		"Output Test Learning",
		"A test summary",
		"Full Content",
		"ag-test",
		"implement",
	}
	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("expected output to contain %q, got:\n%s", check, out)
		}
	}
}
