package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/storage"
	"github.com/boshu2/agentops/cli/internal/types"
)

// ---------------------------------------------------------------------------
// computeSigmaRho
// ---------------------------------------------------------------------------

func TestMetricsCov_computeSigmaRho(t *testing.T) {
	tests := []struct {
		name           string
		totalArtifacts int
		uniqueCited    int
		citationCount  int
		days           int
		wantSigma      float64
		wantRho        float64
	}{
		{"zero artifacts", 0, 0, 0, 7, 0, 0},
		{"no citations", 10, 0, 0, 7, 0, 0},
		{"half cited once per week", 10, 5, 5, 7, 0.5, 1.0},
		{"all cited", 4, 4, 8, 7, 1.0, 2.0},
		{"14-day period", 10, 5, 10, 14, 0.5, 1.0},
		{"zero days", 10, 5, 10, 0, 0.5, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sigma, rho := computeSigmaRho(tt.totalArtifacts, tt.uniqueCited, tt.citationCount, tt.days)
			if !floatApprox(sigma, tt.wantSigma, 0.01) {
				t.Errorf("sigma = %f, want ~%f", sigma, tt.wantSigma)
			}
			if !floatApprox(rho, tt.wantRho, 0.01) {
				t.Errorf("rho = %f, want ~%f", rho, tt.wantRho)
			}
		})
	}
}

func floatApprox(a, b, epsilon float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= epsilon
}

// ---------------------------------------------------------------------------
// filterCitationsForPeriod
// ---------------------------------------------------------------------------

func TestMetricsCov_filterCitationsForPeriod(t *testing.T) {
	now := time.Now()
	citations := []types.CitationEvent{
		{ArtifactPath: "a.md", CitedAt: now.AddDate(0, 0, -1)},
		{ArtifactPath: "b.md", CitedAt: now.AddDate(0, 0, -5)},
		{ArtifactPath: "c.md", CitedAt: now.AddDate(0, 0, -15)},
		{ArtifactPath: "d.md", CitedAt: now.AddDate(0, 0, -45)},
	}

	tests := []struct {
		name      string
		days      int
		wantCount int
	}{
		{"3-day window", 3, 1},
		{"7-day window", 7, 2},
		{"30-day window", 30, 3},
		{"60-day window", 60, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := now.AddDate(0, 0, -tt.days)
			stats := filterCitationsForPeriod(citations, start, now)
			if len(stats.citations) != tt.wantCount {
				t.Errorf("got %d citations, want %d", len(stats.citations), tt.wantCount)
			}
		})
	}
}

func TestMetricsCov_filterCitationsForPeriod_uniqueCited(t *testing.T) {
	now := time.Now()
	citations := []types.CitationEvent{
		{ArtifactPath: "a.md", CitedAt: now.AddDate(0, 0, -1)},
		{ArtifactPath: "a.md", CitedAt: now.AddDate(0, 0, -2)},
		{ArtifactPath: "b.md", CitedAt: now.AddDate(0, 0, -3)},
	}
	start := now.AddDate(0, 0, -7)
	stats := filterCitationsForPeriod(citations, start, now)
	if len(stats.uniqueCited) != 2 {
		t.Errorf("uniqueCited = %d, want 2", len(stats.uniqueCited))
	}
}

// ---------------------------------------------------------------------------
// normalizeArtifactPath / isRetrievableArtifactPath
// ---------------------------------------------------------------------------

func TestMetricsCov_isRetrievableArtifactPath(t *testing.T) {
	baseDir := "/tmp/repo"
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"learnings path", filepath.Join(baseDir, ".agents", "learnings", "test.md"), true},
		{"patterns path", filepath.Join(baseDir, ".agents", "patterns", "test.md"), true},
		{"research path", filepath.Join(baseDir, ".agents", "research", "test.md"), false},
		{"candidates path", filepath.Join(baseDir, ".agents", "candidates", "test.md"), false},
		{"random path", "/some/other/path.md", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetrievableArtifactPath(baseDir, tt.path)
			if got != tt.want {
				t.Errorf("isRetrievableArtifactPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// retrievableCitationStats
// ---------------------------------------------------------------------------

func TestMetricsCov_retrievableCitationStats(t *testing.T) {
	baseDir := "/tmp/repo"
	citations := []types.CitationEvent{
		{ArtifactPath: filepath.Join(baseDir, ".agents", "learnings", "a.md")},
		{ArtifactPath: filepath.Join(baseDir, ".agents", "learnings", "a.md")},
		{ArtifactPath: filepath.Join(baseDir, ".agents", "learnings", "b.md")},
		{ArtifactPath: filepath.Join(baseDir, ".agents", "patterns", "c.md")},
		{ArtifactPath: filepath.Join(baseDir, ".agents", "research", "d.md")},  // not retrievable
	}

	uniqueCount, citationCount := retrievableCitationStats(baseDir, citations)
	if citationCount != 4 {
		t.Errorf("citationCount = %d, want 4", citationCount)
	}
	if uniqueCount != 3 {
		t.Errorf("uniqueCount = %d, want 3", uniqueCount)
	}
}

func TestMetricsCov_retrievableCitationStats_empty(t *testing.T) {
	uniqueCount, citationCount := retrievableCitationStats("/tmp", nil)
	if uniqueCount != 0 || citationCount != 0 {
		t.Errorf("expected 0/0 for empty citations, got %d/%d", uniqueCount, citationCount)
	}
}

// ---------------------------------------------------------------------------
// countBypassCitations
// ---------------------------------------------------------------------------

func TestMetricsCov_countBypassCitations(t *testing.T) {
	citations := []types.CitationEvent{
		{CitationType: "reference", ArtifactPath: "a.md"},
		{CitationType: "bypass", ArtifactPath: "b.md"},
		{CitationType: "reference", ArtifactPath: "bypass:some-reason"},
		{CitationType: "applied", ArtifactPath: "c.md"},
	}
	got := countBypassCitations(citations)
	if got != 2 {
		t.Errorf("countBypassCitations = %d, want 2", got)
	}
}

func TestMetricsCov_countBypassCitations_none(t *testing.T) {
	citations := []types.CitationEvent{
		{CitationType: "reference"},
		{CitationType: "applied"},
	}
	got := countBypassCitations(citations)
	if got != 0 {
		t.Errorf("countBypassCitations = %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// countLoopMetrics
// ---------------------------------------------------------------------------

func TestMetricsCov_countLoopMetrics(t *testing.T) {
	baseDir := t.TempDir()
	learningsDir := filepath.Join(baseDir, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a recent learning
	if err := os.WriteFile(filepath.Join(learningsDir, "new.md"), []byte("# New"), 0o644); err != nil {
		t.Fatal(err)
	}

	periodStart := time.Now().Add(-24 * time.Hour)
	periodCitations := []types.CitationEvent{
		{ArtifactPath: filepath.Join(baseDir, ".agents", "learnings", "cited.md")},
		{ArtifactPath: filepath.Join(baseDir, ".agents", "research", "not-learning.md")},
		{ArtifactPath: filepath.Join(baseDir, ".agents", "learnings", "another.md")},
	}

	created, found := countLoopMetrics(baseDir, periodStart, periodCitations)
	if created != 1 {
		t.Errorf("created = %d, want 1", created)
	}
	if found != 2 {
		t.Errorf("found = %d, want 2", found)
	}
}

// ---------------------------------------------------------------------------
// countArtifacts
// ---------------------------------------------------------------------------

func TestMetricsCov_countArtifacts(t *testing.T) {
	baseDir := t.TempDir()
	dirs := map[string][]string{
		filepath.Join(baseDir, ".agents", "learnings"):  {"l1.md", "l2.jsonl"},
		filepath.Join(baseDir, ".agents", "patterns"):   {"p1.md"},
		filepath.Join(baseDir, ".agents", "candidates"): {"c1.md", "c2.md"},
		filepath.Join(baseDir, ".agents", "research"):   {"r1.md"},
		filepath.Join(baseDir, ".agents", "retros"):     {"retro1.md"},
		filepath.Join(baseDir, storage.DefaultBaseDir, storage.SessionsDir): {"s1.jsonl"},
	}
	for dir, files := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(dir, f), []byte("test"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	total, tierCounts, err := countArtifacts(baseDir)
	if err != nil {
		t.Fatalf("countArtifacts failed: %v", err)
	}

	if tierCounts["learning"] != 2 {
		t.Errorf("learning count = %d, want 2", tierCounts["learning"])
	}
	if tierCounts["pattern"] != 1 {
		t.Errorf("pattern count = %d, want 1", tierCounts["pattern"])
	}
	if tierCounts["observation"] != 3 { // 2 candidates + 1 research
		t.Errorf("observation count = %d, want 3", tierCounts["observation"])
	}
	if tierCounts["retro"] != 1 {
		t.Errorf("retro count = %d, want 1", tierCounts["retro"])
	}
	// Total: 2 learnings + 1 pattern + 2 candidates + 1 research + 1 retro + 1 session = 8
	if total != 8 {
		t.Errorf("total = %d, want 8", total)
	}
}

func TestMetricsCov_countArtifacts_emptyDir(t *testing.T) {
	total, tierCounts, err := countArtifacts(t.TempDir())
	if err != nil {
		t.Fatalf("countArtifacts failed: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if tierCounts["learning"] != 0 {
		t.Errorf("learning should be 0, got %d", tierCounts["learning"])
	}
}

// ---------------------------------------------------------------------------
// countNewArtifacts
// ---------------------------------------------------------------------------

func TestMetricsCov_countNewArtifacts(t *testing.T) {
	baseDir := t.TempDir()
	learningsDir := filepath.Join(baseDir, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create one old and one new file
	oldPath := filepath.Join(learningsDir, "old.md")
	newPath := filepath.Join(learningsDir, "new.md")
	if err := os.WriteFile(oldPath, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().AddDate(0, 0, -30)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	since := time.Now().AddDate(0, 0, -7)
	count, err := countNewArtifacts(baseDir, since)
	if err != nil {
		t.Fatalf("countNewArtifacts failed: %v", err)
	}
	if count != 1 {
		t.Errorf("new artifact count = %d, want 1", count)
	}
}

func TestMetricsCov_countNewArtifacts_missingDirs(t *testing.T) {
	count, err := countNewArtifacts(t.TempDir(), time.Now().AddDate(0, 0, -7))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for missing dirs, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// countNewArtifactsInDir
// ---------------------------------------------------------------------------

func TestMetricsCov_countNewArtifactsInDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "recent.md"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(dir, "old.md")
	if err := os.WriteFile(oldPath, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().AddDate(0, 0, -60)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	since := time.Now().AddDate(0, 0, -7)
	count, err := countNewArtifactsInDir(dir, since)
	if err != nil {
		t.Fatalf("countNewArtifactsInDir failed: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestMetricsCov_countNewArtifactsInDir_missingDir(t *testing.T) {
	count, err := countNewArtifactsInDir("/nonexistent/dir", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for missing dir, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// buildLastCitedMap
// ---------------------------------------------------------------------------

func TestMetricsCov_buildLastCitedMap(t *testing.T) {
	baseDir := "/tmp/test"
	now := time.Now()
	citations := []types.CitationEvent{
		{ArtifactPath: "/tmp/test/.agents/learnings/a.md", CitedAt: now.AddDate(0, 0, -10)},
		{ArtifactPath: "/tmp/test/.agents/learnings/a.md", CitedAt: now.AddDate(0, 0, -5)}, // later
		{ArtifactPath: "/tmp/test/.agents/learnings/b.md", CitedAt: now.AddDate(0, 0, -1)},
	}
	m := buildLastCitedMap(baseDir, citations)
	if len(m) != 2 {
		t.Errorf("expected 2 entries, got %d", len(m))
	}
}

func TestMetricsCov_buildLastCitedMap_empty(t *testing.T) {
	m := buildLastCitedMap("/tmp", nil)
	if len(m) != 0 {
		t.Errorf("expected 0 entries, got %d", len(m))
	}
}

// ---------------------------------------------------------------------------
// isKnowledgeFile
// ---------------------------------------------------------------------------

func TestMetricsCov_isKnowledgeFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"learning.md", true},
		{"learning.jsonl", true},
		{"data.json", false},
		{"script.sh", false},
		{"readme.txt", false},
		{".md", true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isKnowledgeFile(tt.path); got != tt.want {
				t.Errorf("isKnowledgeFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// isStaleArtifact
// ---------------------------------------------------------------------------

func TestMetricsCov_isStaleArtifact(t *testing.T) {
	baseDir := "/tmp/repo"
	now := time.Now()
	staleThreshold := now.AddDate(0, 0, -90)

	tests := []struct {
		name      string
		path      string
		modTime   time.Time
		lastCited map[string]time.Time
		want      bool
	}{
		{
			name:      "recent mod time",
			path:      "/tmp/repo/.agents/learnings/new.md",
			modTime:   now,
			lastCited: map[string]time.Time{},
			want:      false,
		},
		{
			name:      "old mod time no citation",
			path:      "/tmp/repo/.agents/learnings/old.md",
			modTime:   now.AddDate(0, 0, -120),
			lastCited: map[string]time.Time{},
			want:      true,
		},
		{
			name:    "old mod time with recent citation",
			path:    "/tmp/repo/.agents/learnings/old.md",
			modTime: now.AddDate(0, 0, -120),
			lastCited: map[string]time.Time{
				normalizeArtifactPath(baseDir, "/tmp/repo/.agents/learnings/old.md"): now.AddDate(0, 0, -30),
			},
			want: false,
		},
		{
			name:    "old mod time with old citation",
			path:    "/tmp/repo/.agents/learnings/old.md",
			modTime: now.AddDate(0, 0, -120),
			lastCited: map[string]time.Time{
				normalizeArtifactPath(baseDir, "/tmp/repo/.agents/learnings/old.md"): now.AddDate(0, 0, -100),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isStaleArtifact(baseDir, tt.path, tt.modTime, staleThreshold, tt.lastCited)
			if got != tt.want {
				t.Errorf("isStaleArtifact = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// countStaleInDir
// ---------------------------------------------------------------------------

func TestMetricsCov_countStaleInDir(t *testing.T) {
	baseDir := t.TempDir()
	dir := filepath.Join(baseDir, ".agents", "learnings")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Old file
	oldPath := filepath.Join(dir, "old.md")
	if err := os.WriteFile(oldPath, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().AddDate(0, 0, -120)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Recent file
	if err := os.WriteFile(filepath.Join(dir, "new.md"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Non-knowledge file
	if err := os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("skip"), 0o644); err != nil {
		t.Fatal(err)
	}

	staleThreshold := time.Now().AddDate(0, 0, -90)
	count := countStaleInDir(baseDir, dir, staleThreshold, map[string]time.Time{})
	if count != 1 {
		t.Errorf("stale count = %d, want 1", count)
	}
}

func TestMetricsCov_countStaleInDir_missingDir(t *testing.T) {
	count := countStaleInDir("/tmp", "/nonexistent", time.Now(), nil)
	if count != 0 {
		t.Errorf("expected 0 for missing dir, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// countStaleArtifacts
// ---------------------------------------------------------------------------

func TestMetricsCov_countStaleArtifacts(t *testing.T) {
	baseDir := t.TempDir()
	learningsDir := filepath.Join(baseDir, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	oldPath := filepath.Join(learningsDir, "stale.md")
	if err := os.WriteFile(oldPath, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().AddDate(0, 0, -120)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	count, err := countStaleArtifacts(baseDir, nil, 90)
	if err != nil {
		t.Fatalf("countStaleArtifacts failed: %v", err)
	}
	if count != 1 {
		t.Errorf("stale count = %d, want 1", count)
	}
}

// ---------------------------------------------------------------------------
// retroHasLearnings
// ---------------------------------------------------------------------------

func TestMetricsCov_retroHasLearnings(t *testing.T) {
	tmp := t.TempDir()

	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"has ## Learnings", "# Retro\n## Learnings\n- item\n", true},
		{"has ## Key Learnings", "# Retro\n## Key Learnings\n- item\n", true},
		{"has ### Learnings", "# Retro\n### Learnings\n- item\n", true},
		{"no learnings section", "# Retro\n## Summary\n- item\n", false},
		{"empty file", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmp, tt.name+".md")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			got := retroHasLearnings(path)
			if got != tt.want {
				t.Errorf("retroHasLearnings = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricsCov_retroHasLearnings_missingFile(t *testing.T) {
	got := retroHasLearnings("/nonexistent/file.md")
	if got {
		t.Error("expected false for missing file")
	}
}

// ---------------------------------------------------------------------------
// countRetros
// ---------------------------------------------------------------------------

func TestMetricsCov_countRetros(t *testing.T) {
	baseDir := t.TempDir()
	retrosDir := filepath.Join(baseDir, ".agents", "retros")
	if err := os.MkdirAll(retrosDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Recent retro with learnings
	if err := os.WriteFile(filepath.Join(retrosDir, "retro1.md"), []byte("# Retro\n## Learnings\n- L1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Recent retro without learnings
	if err := os.WriteFile(filepath.Join(retrosDir, "retro2.md"), []byte("# Retro\n## Summary\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Old retro (before since)
	oldPath := filepath.Join(retrosDir, "old-retro.md")
	if err := os.WriteFile(oldPath, []byte("# Old\n## Learnings\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().AddDate(0, 0, -60)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	since := time.Now().AddDate(0, 0, -7)
	total, withLearnings, err := countRetros(baseDir, since)
	if err != nil {
		t.Fatalf("countRetros failed: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if withLearnings != 1 {
		t.Errorf("withLearnings = %d, want 1", withLearnings)
	}
}

func TestMetricsCov_countRetros_missingDir(t *testing.T) {
	total, withLearnings, err := countRetros(t.TempDir(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 || withLearnings != 0 {
		t.Errorf("expected 0/0, got %d/%d", total, withLearnings)
	}
}

// ---------------------------------------------------------------------------
// computeUtilityStats
// ---------------------------------------------------------------------------

func TestMetricsCov_computeUtilityStats(t *testing.T) {
	tests := []struct {
		name      string
		utilities []float64
		wantMean  float64
		wantHigh  int
		wantLow   int
	}{
		{"empty", nil, 0, 0, 0},
		{"single high", []float64{0.8}, 0.8, 1, 0},
		{"single low", []float64{0.2}, 0.2, 0, 1},
		{"mixed", []float64{0.1, 0.5, 0.9}, 0.5, 1, 1},
		{"all mid-range", []float64{0.4, 0.5, 0.6}, 0.5, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := computeUtilityStats(tt.utilities)
			if !floatApprox(stats.mean, tt.wantMean, 0.01) {
				t.Errorf("mean = %f, want ~%f", stats.mean, tt.wantMean)
			}
			if stats.highCount != tt.wantHigh {
				t.Errorf("highCount = %d, want %d", stats.highCount, tt.wantHigh)
			}
			if stats.lowCount != tt.wantLow {
				t.Errorf("lowCount = %d, want %d", stats.lowCount, tt.wantLow)
			}
		})
	}
}

func TestMetricsCov_computeUtilityStats_stdDev(t *testing.T) {
	// All same values => stdDev = 0
	stats := computeUtilityStats([]float64{0.5, 0.5, 0.5})
	if stats.stdDev != 0 {
		t.Errorf("stdDev = %f, want 0 for identical values", stats.stdDev)
	}

	// Known spread
	stats2 := computeUtilityStats([]float64{0.0, 1.0})
	if stats2.stdDev < 0.49 || stats2.stdDev > 0.51 {
		t.Errorf("stdDev = %f, want ~0.5", stats2.stdDev)
	}
}

// ---------------------------------------------------------------------------
// collectUtilityValuesFromDir
// ---------------------------------------------------------------------------

func TestMetricsCov_collectUtilityValuesFromDir(t *testing.T) {
	dir := t.TempDir()

	// JSONL with utility
	jsonl := map[string]any{"id": "L1", "utility": 0.8}
	data, _ := json.Marshal(jsonl)
	if err := os.WriteFile(filepath.Join(dir, "l1.jsonl"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Markdown with utility
	if err := os.WriteFile(filepath.Join(dir, "l2.md"), []byte("---\nutility: 0.6\n---\n# L2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// File without utility (should not contribute)
	if err := os.WriteFile(filepath.Join(dir, "no-utility.md"), []byte("---\ntitle: test\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Non-matching extension
	if err := os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("skip"), 0o644); err != nil {
		t.Fatal(err)
	}

	values := collectUtilityValuesFromDir(dir)
	if len(values) != 2 {
		t.Errorf("expected 2 utility values, got %d", len(values))
	}
}

func TestMetricsCov_collectUtilityValuesFromDir_missingDir(t *testing.T) {
	values := collectUtilityValuesFromDir("/nonexistent/dir")
	if values != nil {
		t.Errorf("expected nil for missing dir, got %v", values)
	}
}

// ---------------------------------------------------------------------------
// parseUtilityFromFile / parseUtilityFromMarkdown / parseUtilityFromJSONL
// ---------------------------------------------------------------------------

func TestMetricsCov_parseUtilityFromMarkdown(t *testing.T) {
	tmp := t.TempDir()
	tests := []struct {
		name    string
		content string
		want    float64
	}{
		{"valid utility", "---\nutility: 0.73\n---\n# L\n", 0.73},
		{"no frontmatter", "# Just content\n", 0},
		{"no utility field", "---\ntitle: Test\n---\n", 0},
		{"empty file", "", 0},
		{"utility at zero", "---\nutility: 0.0\n---\n", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmp, tt.name+".md")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			got := parseUtilityFromMarkdown(path)
			if got != tt.want {
				t.Errorf("parseUtilityFromMarkdown = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestMetricsCov_parseUtilityFromJSONL(t *testing.T) {
	tmp := t.TempDir()
	tests := []struct {
		name    string
		content string
		want    float64
	}{
		{"valid", `{"utility":0.65}` + "\n", 0.65},
		{"no utility", `{"id":"L1"}` + "\n", 0},
		{"invalid JSON", "not json\n", 0},
		{"empty", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmp, tt.name+".jsonl")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			got := parseUtilityFromJSONL(path)
			if got != tt.want {
				t.Errorf("parseUtilityFromJSONL = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestMetricsCov_parseUtilityFromJSONL_missingFile(t *testing.T) {
	got := parseUtilityFromJSONL("/nonexistent/file.jsonl")
	if got != 0 {
		t.Errorf("expected 0 for missing file, got %f", got)
	}
}

func TestMetricsCov_parseUtilityFromFile_dispatch(t *testing.T) {
	tmp := t.TempDir()

	// .md file
	mdPath := filepath.Join(tmp, "test.md")
	if err := os.WriteFile(mdPath, []byte("---\nutility: 0.55\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// .jsonl file
	jsonlPath := filepath.Join(tmp, "test.jsonl")
	data, _ := json.Marshal(map[string]any{"utility": 0.77})
	if err := os.WriteFile(jsonlPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if got := parseUtilityFromFile(mdPath); got != 0.55 {
		t.Errorf("parseUtilityFromFile(.md) = %f, want 0.55", got)
	}
	if got := parseUtilityFromFile(jsonlPath); got != 0.77 {
		t.Errorf("parseUtilityFromFile(.jsonl) = %f, want 0.77", got)
	}
}

// ---------------------------------------------------------------------------
// printMetricsParameters / printMetricsDerived / printMetricsCounts /
// printMetricsLoopClosure / printMetricsUtility / printMetricsTable
// (smoke tests — just ensure no panic)
// ---------------------------------------------------------------------------

func TestMetricsCov_printMetricsParameters(t *testing.T) {
	m := &types.FlywheelMetrics{Delta: 0.17, Sigma: 0.5, Rho: 1.0}
	printMetricsParameters(m)
}

func TestMetricsCov_printMetricsDerived(t *testing.T) {
	tests := []struct {
		name string
		m    *types.FlywheelMetrics
	}{
		{"positive velocity", &types.FlywheelMetrics{SigmaRho: 0.5, Delta: 0.17, Velocity: 0.33, AboveEscapeVelocity: true}},
		{"negative velocity", &types.FlywheelMetrics{SigmaRho: 0.05, Delta: 0.17, Velocity: -0.12, AboveEscapeVelocity: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printMetricsDerived(tt.m)
		})
	}
}

func TestMetricsCov_printMetricsCounts(t *testing.T) {
	m := &types.FlywheelMetrics{
		TotalArtifacts:      10,
		CitationsThisPeriod: 5,
		UniqueCitedArtifacts: 3,
		NewArtifacts:        2,
		StaleArtifacts:      1,
		TierCounts:          map[string]int{"learning": 5, "pattern": 3, "observation": 2},
	}
	printMetricsCounts(m)
}

func TestMetricsCov_printMetricsCounts_emptyTiers(t *testing.T) {
	m := &types.FlywheelMetrics{TierCounts: map[string]int{}}
	printMetricsCounts(m)
}

func TestMetricsCov_printMetricsLoopClosure(t *testing.T) {
	tests := []struct {
		name string
		m    *types.FlywheelMetrics
	}{
		{"no loop data", &types.FlywheelMetrics{}},
		{"open loop", &types.FlywheelMetrics{LearningsCreated: 5, LearningsFound: 0, LoopClosureRatio: 0}},
		{"partial loop", &types.FlywheelMetrics{LearningsCreated: 5, LearningsFound: 3, LoopClosureRatio: 0.6}},
		{"closed loop", &types.FlywheelMetrics{LearningsCreated: 5, LearningsFound: 5, LoopClosureRatio: 1.0}},
		{"with retros", &types.FlywheelMetrics{LearningsCreated: 1, TotalRetros: 3, RetrosWithLearnings: 2}},
		{"with bypasses", &types.FlywheelMetrics{LearningsCreated: 1, PriorArtBypasses: 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printMetricsLoopClosure(tt.m)
		})
	}
}

func TestMetricsCov_printMetricsUtility(t *testing.T) {
	tests := []struct {
		name string
		m    *types.FlywheelMetrics
	}{
		{"no utility data", &types.FlywheelMetrics{}},
		{"healthy", &types.FlywheelMetrics{MeanUtility: 0.7, UtilityStdDev: 0.1, HighUtilityCount: 5, LowUtilityCount: 1}},
		{"neutral", &types.FlywheelMetrics{MeanUtility: 0.45, HighUtilityCount: 1}},
		{"needs review", &types.FlywheelMetrics{MeanUtility: 0.2, LowUtilityCount: 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printMetricsUtility(tt.m)
		})
	}
}

func TestMetricsCov_printMetricsTable(t *testing.T) {
	now := time.Now()
	m := &types.FlywheelMetrics{
		Timestamp:            now,
		PeriodStart:          now.AddDate(0, 0, -7),
		PeriodEnd:            now,
		Delta:                0.17,
		Sigma:                0.5,
		Rho:                  1.0,
		SigmaRho:             0.5,
		Velocity:             0.33,
		AboveEscapeVelocity:  true,
		TotalArtifacts:       10,
		CitationsThisPeriod:  5,
		UniqueCitedArtifacts: 3,
		TierCounts:           map[string]int{"learning": 5, "pattern": 3},
	}
	// Should not panic
	printMetricsTable(m)
}

// ---------------------------------------------------------------------------
// periodCitationStats struct
// ---------------------------------------------------------------------------

func TestMetricsCov_periodCitationStats(t *testing.T) {
	stats := periodCitationStats{
		citations:   []types.CitationEvent{{ArtifactPath: "a.md"}},
		uniqueCited: map[string]bool{"a.md": true},
	}
	if len(stats.citations) != 1 || len(stats.uniqueCited) != 1 {
		t.Error("periodCitationStats field access failed")
	}
}

// ---------------------------------------------------------------------------
// computeUtilityMetrics
// ---------------------------------------------------------------------------

func TestMetricsCov_computeUtilityMetrics(t *testing.T) {
	baseDir := t.TempDir()
	learningsDir := filepath.Join(baseDir, ".agents", "learnings")
	patternsDir := filepath.Join(baseDir, ".agents", "patterns")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(patternsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// JSONL learning
	data, _ := json.Marshal(map[string]any{"utility": 0.8})
	if err := os.WriteFile(filepath.Join(learningsDir, "l1.jsonl"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	// Markdown pattern
	if err := os.WriteFile(filepath.Join(patternsDir, "p1.md"), []byte("---\nutility: 0.6\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stats := computeUtilityMetrics(baseDir)
	if !floatApprox(stats.mean, 0.7, 0.01) {
		t.Errorf("mean = %f, want ~0.7", stats.mean)
	}
}

func TestMetricsCov_computeUtilityMetrics_emptyDirs(t *testing.T) {
	stats := computeUtilityMetrics(t.TempDir())
	if stats.mean != 0 {
		t.Errorf("mean = %f, want 0 for empty dirs", stats.mean)
	}
}
