package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/ratchet"
)

// ---------------------------------------------------------------------------
// parseFrontmatterFields
// ---------------------------------------------------------------------------

func TestMaturity_parseFrontmatterFields(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		fields     []string
		wantFields map[string]string
	}{
		{
			name:       "basic frontmatter",
			content:    "---\ntitle: My Learning\nvalid_until: 2099-12-31\n---\n# Body\n",
			fields:     []string{"title", "valid_until"},
			wantFields: map[string]string{"title": "My Learning", "valid_until": "2099-12-31"},
		},
		{
			name:       "quoted values stripped",
			content:    "---\ntitle: \"Quoted Title\"\nstatus: 'single'\n---\n",
			fields:     []string{"title", "status"},
			wantFields: map[string]string{"title": "Quoted Title", "status": "single"},
		},
		{
			name:       "field not present",
			content:    "---\ntitle: Test\n---\n",
			fields:     []string{"missing_field"},
			wantFields: map[string]string{},
		},
		{
			name:       "no frontmatter",
			content:    "# Just a heading\nSome content\n",
			fields:     []string{"title"},
			wantFields: map[string]string{},
		},
		{
			name:       "empty file",
			content:    "",
			fields:     []string{"title"},
			wantFields: map[string]string{},
		},
		{
			name:       "only opening delimiter",
			content:    "---\ntitle: Test\n",
			fields:     []string{"title"},
			wantFields: map[string]string{"title": "Test"},
		},
		{
			name:       "expiry_status field",
			content:    "---\nexpiry_status: archived\nvalid_until: 2020-01-01\n---\n",
			fields:     []string{"expiry_status", "valid_until"},
			wantFields: map[string]string{"expiry_status": "archived", "valid_until": "2020-01-01"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := filepath.Join(tmp, "test.md")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}

			got, err := parseFrontmatterFields(path, tt.fields...)
			if err != nil {
				t.Fatalf("parseFrontmatterFields failed: %v", err)
			}

			for k, want := range tt.wantFields {
				if got[k] != want {
					t.Errorf("field %q = %q, want %q", k, got[k], want)
				}
			}
			// Check no extra fields returned
			for k := range got {
				if _, ok := tt.wantFields[k]; !ok {
					t.Errorf("unexpected field %q = %q", k, got[k])
				}
			}
		})
	}
}

func TestMaturity_parseFrontmatterFields_fileNotFound(t *testing.T) {
	_, err := parseFrontmatterFields("/nonexistent/path.md", "title")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// ---------------------------------------------------------------------------
// classifyExpiryEntry
// ---------------------------------------------------------------------------

func TestMaturity_classifyExpiryEntry(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Active: future valid_until
	writeTestFrontmatter(t, learningsDir, "active.md", "valid_until: 2099-12-31")
	// Expired: past valid_until
	writeTestFrontmatter(t, learningsDir, "expired.md", "valid_until: 2020-01-01")
	// No expiry
	writeTestFrontmatter(t, learningsDir, "no-expiry.md", "title: Test")
	// Already archived
	writeTestFrontmatter(t, learningsDir, "archived.md", "expiry_status: archived\nvalid_until: 2020-01-01")
	// Malformed date
	writeTestFrontmatter(t, learningsDir, "bad-date.md", "valid_until: not-a-date")
	// RFC3339 date format
	writeTestFrontmatter(t, learningsDir, "rfc3339.md", "valid_until: 2099-12-31T23:59:59Z")

	entries, err := os.ReadDir(learningsDir)
	if err != nil {
		t.Fatal(err)
	}

	cats := expiryCategory{}
	for _, entry := range entries {
		if entry.IsDir() || !hasExtension(entry.Name(), ".md") {
			continue
		}
		classifyExpiryEntry(entry, learningsDir, &cats)
	}

	if len(cats.active) != 2 { // active.md + rfc3339.md
		t.Errorf("active count = %d, want 2 (got %v)", len(cats.active), cats.active)
	}
	if len(cats.newlyExpired) != 1 { // expired.md
		t.Errorf("newlyExpired count = %d, want 1 (got %v)", len(cats.newlyExpired), cats.newlyExpired)
	}
	if len(cats.alreadyArchived) != 1 { // archived.md
		t.Errorf("alreadyArchived count = %d, want 1 (got %v)", len(cats.alreadyArchived), cats.alreadyArchived)
	}
	if len(cats.neverExpiring) != 2 { // no-expiry.md + bad-date.md
		t.Errorf("neverExpiring count = %d, want 2 (got %v)", len(cats.neverExpiring), cats.neverExpiring)
	}
}

func hasExtension(name, ext string) bool {
	return len(name) > len(ext) && name[len(name)-len(ext):] == ext
}

func writeTestFrontmatter(t *testing.T, dir, name, frontmatter string) {
	t.Helper()
	content := "---\n" + frontmatter + "\n---\n# Content\n"
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// isEvictionEligible
// ---------------------------------------------------------------------------

func TestMaturity_isEvictionEligible(t *testing.T) {
	tests := []struct {
		name       string
		utility    float64
		confidence float64
		maturity   string
		want       bool
	}{
		{"established never eligible", 0.1, 0.1, "established", false},
		{"high utility not eligible", 0.5, 0.1, "provisional", false},
		{"high confidence not eligible", 0.1, 0.5, "provisional", false},
		{"all criteria met", 0.1, 0.1, "provisional", true},
		{"boundary utility 0.3", 0.3, 0.1, "provisional", false},
		{"boundary confidence 0.3", 0.1, 0.3, "provisional", false},
		{"candidate eligible", 0.1, 0.1, "candidate", true},
		{"anti-pattern eligible", 0.1, 0.1, "anti-pattern", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEvictionEligible(tt.utility, tt.confidence, tt.maturity)
			if got != tt.want {
				t.Errorf("isEvictionEligible(%f, %f, %q) = %v, want %v",
					tt.utility, tt.confidence, tt.maturity, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// evictionCitationStatus
// ---------------------------------------------------------------------------

func TestMaturity_evictionCitationStatus(t *testing.T) {
	now := time.Now()
	cutoff := now.AddDate(0, 0, -90)

	tests := []struct {
		name      string
		file      string
		lastCited map[string]time.Time
		wantStr   string
		wantOK    bool
	}{
		{
			name:      "never cited",
			file:      "/path/to/file",
			lastCited: map[string]time.Time{},
			wantStr:   "never",
			wantOK:    true,
		},
		{
			name:      "cited before cutoff",
			file:      "/path/to/file",
			lastCited: map[string]time.Time{"/path/to/file": now.AddDate(0, 0, -120)},
			wantStr:   now.AddDate(0, 0, -120).Format("2006-01-02"),
			wantOK:    true,
		},
		{
			name:      "cited after cutoff - not eligible",
			file:      "/path/to/file",
			lastCited: map[string]time.Time{"/path/to/file": now.AddDate(0, 0, -30)},
			wantStr:   "",
			wantOK:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotOK := evictionCitationStatus(tt.file, tt.lastCited, cutoff)
			if gotOK != tt.wantOK {
				t.Errorf("ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotStr != tt.wantStr {
				t.Errorf("str = %q, want %q", gotStr, tt.wantStr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// readLearningJSONLData
// ---------------------------------------------------------------------------

func TestMaturityCov_readLearningJSONLData(t *testing.T) {
	tmp := t.TempDir()

	tests := []struct {
		name    string
		content string
		wantOK  bool
	}{
		{"valid JSONL", `{"id":"L-1","utility":0.5}` + "\n", true},
		{"empty file", "", false},
		{"only whitespace", "   \n", false},
		{"invalid JSON", "not json\n", false},
		{"multi-line uses first", `{"id":"L-1"}` + "\n" + `{"id":"L-2"}` + "\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmp, tt.name+".jsonl")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			data, ok := readLearningJSONLData(path)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOK && data == nil {
				t.Error("expected non-nil data when ok=true")
			}
		})
	}
}

func TestMaturity_readLearningJSONLData_missingFile(t *testing.T) {
	_, ok := readLearningJSONLData("/nonexistent/file.jsonl")
	if ok {
		t.Error("expected ok=false for missing file")
	}
}

// ---------------------------------------------------------------------------
// filterTransitionsByNewMaturity
// ---------------------------------------------------------------------------

func TestMaturity_filterTransitionsByNewMaturity(t *testing.T) {
	results := []*ratchet.MaturityTransitionResult{
		{LearningID: "L1", NewMaturity: "anti-pattern"},
		{LearningID: "L2", NewMaturity: "established"},
		{LearningID: "L3", NewMaturity: "anti-pattern"},
		{LearningID: "L4", NewMaturity: "candidate"},
	}

	filtered := filterTransitionsByNewMaturity(results, "anti-pattern")
	if len(filtered) != 2 {
		t.Errorf("expected 2 anti-pattern results, got %d", len(filtered))
	}

	filtered2 := filterTransitionsByNewMaturity(results, "established")
	if len(filtered2) != 1 {
		t.Errorf("expected 1 established result, got %d", len(filtered2))
	}

	filtered3 := filterTransitionsByNewMaturity(results, "nonexistent")
	if len(filtered3) != 0 {
		t.Errorf("expected 0 results for nonexistent maturity, got %d", len(filtered3))
	}

	// nil input
	filtered4 := filterTransitionsByNewMaturity(nil, "anti-pattern")
	if len(filtered4) != 0 {
		t.Errorf("expected 0 results for nil input, got %d", len(filtered4))
	}
}

// ---------------------------------------------------------------------------
// archiveExpiredLearnings
// ---------------------------------------------------------------------------

func TestMaturityCov_archiveExpiredLearnings(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create test files
	for _, name := range []string{"expired1.md", "expired2.md"} {
		if err := os.WriteFile(filepath.Join(learningsDir, name), []byte("# Test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	expired := []string{"expired1.md", "expired2.md"}
	err := archiveExpiredLearnings(tmp, learningsDir, expired)
	if err != nil {
		t.Fatalf("archiveExpiredLearnings failed: %v", err)
	}

	archiveDir := filepath.Join(tmp, ".agents", "archive", "learnings")
	for _, name := range expired {
		if _, err := os.Stat(filepath.Join(archiveDir, name)); os.IsNotExist(err) {
			t.Errorf("expected %s in archive dir", name)
		}
		if _, err := os.Stat(filepath.Join(learningsDir, name)); !os.IsNotExist(err) {
			t.Errorf("expected %s removed from learnings dir", name)
		}
	}
}

func TestMaturity_archiveExpiredLearnings_dryRun(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "test.md"), []byte("# Test"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	err := archiveExpiredLearnings(tmp, learningsDir, []string{"test.md"})
	if err != nil {
		t.Fatalf("archiveExpiredLearnings dry-run failed: %v", err)
	}

	// File should still exist
	if _, err := os.Stat(filepath.Join(learningsDir, "test.md")); os.IsNotExist(err) {
		t.Error("file should not be moved in dry-run mode")
	}
}

// ---------------------------------------------------------------------------
// archiveEvictionCandidates
// ---------------------------------------------------------------------------

func TestMaturityCov_archiveEvictionCandidates(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(learningsDir, "evict-me.jsonl")
	if err := os.WriteFile(path, []byte(`{"id":"L-1"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	candidates := []evictionCandidate{
		{Path: path, Name: "evict-me.jsonl"},
	}
	err := archiveEvictionCandidates(tmp, candidates)
	if err != nil {
		t.Fatalf("archiveEvictionCandidates failed: %v", err)
	}

	archiveDir := filepath.Join(tmp, ".agents", "archive", "learnings")
	if _, err := os.Stat(filepath.Join(archiveDir, "evict-me.jsonl")); os.IsNotExist(err) {
		t.Error("expected file in archive dir")
	}
}

func TestMaturity_archiveEvictionCandidates_dryRun(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(learningsDir, "keep.jsonl")
	if err := os.WriteFile(path, []byte(`{"id":"L-1"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	candidates := []evictionCandidate{
		{Path: path, Name: "keep.jsonl"},
	}
	err := archiveEvictionCandidates(tmp, candidates)
	if err != nil {
		t.Fatalf("archiveEvictionCandidates dry-run failed: %v", err)
	}

	// File should still exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should not be moved in dry-run mode")
	}
}

// ---------------------------------------------------------------------------
// collectEvictionCandidates
// ---------------------------------------------------------------------------

func TestMaturityCov_collectEvictionCandidates(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Eligible: low utility, low confidence, provisional
	createTestLearningJSONL(t, learningsDir, "eligible.jsonl", map[string]any{
		"id": "L-eligible", "utility": 0.1, "confidence": 0.1, "maturity": "provisional",
	})
	// Not eligible: established
	createTestLearningJSONL(t, learningsDir, "established.jsonl", map[string]any{
		"id": "L-established", "utility": 0.1, "confidence": 0.1, "maturity": "established",
	})
	// Not eligible: high utility
	createTestLearningJSONL(t, learningsDir, "high-utility.jsonl", map[string]any{
		"id": "L-high", "utility": 0.8, "confidence": 0.1, "maturity": "provisional",
	})

	files, _ := filepath.Glob(filepath.Join(learningsDir, "*.jsonl"))
	cutoff := time.Now().AddDate(0, 0, -90)
	lastCited := make(map[string]time.Time) // no citations

	candidates := collectEvictionCandidates(tmp, files, lastCited, cutoff)
	if len(candidates) != 1 {
		t.Errorf("expected 1 candidate, got %d", len(candidates))
	}
	if len(candidates) > 0 && candidates[0].Name != "eligible.jsonl" {
		t.Errorf("expected eligible.jsonl, got %s", candidates[0].Name)
	}
}

// ---------------------------------------------------------------------------
// reportEvictionCandidates
// ---------------------------------------------------------------------------

func TestMaturity_reportEvictionCandidates_noCandidates(t *testing.T) {
	shouldArchive, err := reportEvictionCandidates([]string{"a.jsonl", "b.jsonl"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shouldArchive {
		t.Error("expected shouldArchive=false for no candidates")
	}
}

func TestMaturity_reportEvictionCandidates_withCandidates(t *testing.T) {
	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	candidates := []evictionCandidate{
		{Name: "test.jsonl", Utility: 0.1, Confidence: 0.1, Maturity: "provisional", LastCited: "never"},
	}
	shouldArchive, err := reportEvictionCandidates([]string{"test.jsonl"}, candidates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !shouldArchive {
		t.Error("expected shouldArchive=true when candidates exist (text mode)")
	}
}

func TestMaturity_reportEvictionCandidates_jsonOutput(t *testing.T) {
	oldOutput := output
	output = "json"
	defer func() { output = oldOutput }()

	candidates := []evictionCandidate{
		{Name: "test.jsonl", Utility: 0.1, Confidence: 0.1, Maturity: "provisional", LastCited: "never"},
	}
	shouldArchive, err := reportEvictionCandidates([]string{"test.jsonl"}, candidates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shouldArchive {
		t.Error("expected shouldArchive=false in JSON mode (JSON output handles it)")
	}
}

// ---------------------------------------------------------------------------
// displayMaturityDistribution (smoke test, output only)
// ---------------------------------------------------------------------------

func TestMaturityCov_displayMaturityDistribution(t *testing.T) {
	dist := &ratchet.MaturityDistribution{
		Provisional: 3,
		Candidate:   2,
		Established: 1,
		AntiPattern: 0,
		Total:       6,
	}
	// Should not panic
	displayMaturityDistribution(dist)
}

// ---------------------------------------------------------------------------
// displayMaturityResult (smoke test, output only)
// ---------------------------------------------------------------------------

func TestMaturityCov_displayMaturityResult(t *testing.T) {
	tests := []struct {
		name        string
		result      *ratchet.MaturityTransitionResult
		applied     bool
	}{
		{
			name: "no transition",
			result: &ratchet.MaturityTransitionResult{
				LearningID:   "L-1",
				OldMaturity:  "provisional",
				Transitioned: false,
				Utility:      0.5,
				Confidence:   0.6,
				RewardCount:  3,
				HelpfulCount: 2,
				HarmfulCount: 1,
				Reason:       "not enough feedback",
			},
			applied: false,
		},
		{
			name: "transition applied",
			result: &ratchet.MaturityTransitionResult{
				LearningID:   "L-2",
				OldMaturity:  "provisional",
				NewMaturity:  "candidate",
				Transitioned: true,
				Utility:      0.8,
				Confidence:   0.7,
				RewardCount:  5,
				HelpfulCount: 4,
				HarmfulCount: 1,
				Reason:       "sufficient positive feedback",
			},
			applied: true,
		},
		{
			name: "transition not yet applied",
			result: &ratchet.MaturityTransitionResult{
				LearningID:   "L-3",
				OldMaturity:  "candidate",
				NewMaturity:  "established",
				Transitioned: true,
				Utility:      0.9,
				Confidence:   0.8,
				RewardCount:  10,
				HelpfulCount: 8,
				HarmfulCount: 2,
				Reason:       "proven value",
			},
			applied: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			displayMaturityResult(tt.result, tt.applied)
		})
	}
}

// ---------------------------------------------------------------------------
// displayPendingTransitions
// ---------------------------------------------------------------------------

func TestMaturity_displayPendingTransitions_text(t *testing.T) {
	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	results := []*ratchet.MaturityTransitionResult{
		{LearningID: "L-1", OldMaturity: "provisional", NewMaturity: "candidate", Transitioned: true},
	}
	err := displayPendingTransitions(results)
	if err != nil {
		t.Fatalf("displayPendingTransitions text failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// displayAntiPatternCandidates (smoke test)
// ---------------------------------------------------------------------------

func TestMaturityCov_displayAntiPatternCandidates(t *testing.T) {
	results := []*ratchet.MaturityTransitionResult{
		{LearningID: "L-1", Utility: 0.1, HarmfulCount: 7},
		{LearningID: "L-2", Utility: 0.15, HarmfulCount: 5},
	}
	// Should not panic
	displayAntiPatternCandidates(results)
}

// ---------------------------------------------------------------------------
// evictionCandidate struct
// ---------------------------------------------------------------------------

func TestMaturity_evictionCandidateFields(t *testing.T) {
	c := evictionCandidate{
		Path:       "/path/to/file.jsonl",
		Name:       "file.jsonl",
		Utility:    0.15,
		Confidence: 0.1,
		Maturity:   "provisional",
		LastCited:  "never",
	}
	if c.Path == "" || c.Name == "" || c.Maturity == "" {
		t.Error("evictionCandidate fields should be set")
	}
}

// ---------------------------------------------------------------------------
// expiryCategory struct
// ---------------------------------------------------------------------------

func TestMaturity_expiryCategoryEmpty(t *testing.T) {
	cats := expiryCategory{}
	if len(cats.active) != 0 || len(cats.neverExpiring) != 0 ||
		len(cats.newlyExpired) != 0 || len(cats.alreadyArchived) != 0 {
		t.Error("empty expiryCategory should have all nil slices")
	}
}
