package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/boshu2/agentops/cli/internal/goals"
	"github.com/boshu2/agentops/cli/internal/pool"
	"github.com/boshu2/agentops/cli/internal/types"
)

// ---------------------------------------------------------------------------
// curate.go — generateArtifactID (50% → higher)
// ---------------------------------------------------------------------------

func TestCov8_generateArtifactID_decision(t *testing.T) {
	got := generateArtifactID("decision", "2026-01-01", "some content")
	if !strings.HasPrefix(got, "decis-") {
		t.Errorf("generateArtifactID(decision,...) = %q, want 'decis-' prefix", got)
	}
}

func TestCov8_generateArtifactID_failure(t *testing.T) {
	got := generateArtifactID("failure", "2026-01-01", "some content")
	if !strings.HasPrefix(got, "fail-") {
		t.Errorf("generateArtifactID(failure,...) = %q, want 'fail-' prefix", got)
	}
}

func TestCov8_generateArtifactID_pattern(t *testing.T) {
	got := generateArtifactID("pattern", "2026-01-01", "some content")
	if !strings.HasPrefix(got, "patt-") {
		t.Errorf("generateArtifactID(pattern,...) = %q, want 'patt-' prefix", got)
	}
}

func TestCov8_generateArtifactID_unknown(t *testing.T) {
	// Unknown type falls through all else-if branches → empty prefix
	got := generateArtifactID("other", "2026-01-01", "content")
	if got == "" {
		t.Error("generateArtifactID(other,...) returned empty string")
	}
	// prefix is empty so result starts with "-"
	if !strings.HasPrefix(got, "-") {
		t.Errorf("generateArtifactID(other,...) = %q, expected prefix '-'", got)
	}
}

// ---------------------------------------------------------------------------
// curate.go — curateArtifactDir (66.7% → 100%)
// ---------------------------------------------------------------------------

func TestCov8_curateArtifactDir_pattern(t *testing.T) {
	got := curateArtifactDir("pattern")
	if got != ".agents/patterns" {
		t.Errorf("curateArtifactDir(pattern) = %q, want .agents/patterns", got)
	}
}

func TestCov8_curateArtifactDir_learning(t *testing.T) {
	got := curateArtifactDir("learning")
	if got != ".agents/learnings" {
		t.Errorf("curateArtifactDir(learning) = %q, want .agents/learnings", got)
	}
}

// ---------------------------------------------------------------------------
// goals_migrate.go — directivesFromPillars (0% → 100%)
// ---------------------------------------------------------------------------

func TestCov8_directivesFromPillars_empty(t *testing.T) {
	dirs := directivesFromPillars(nil)
	if len(dirs) != 1 {
		t.Fatalf("directivesFromPillars(nil) returned %d directives, want 1", len(dirs))
	}
	if dirs[0].Number != 1 {
		t.Errorf("default directive number = %d, want 1", dirs[0].Number)
	}
	if dirs[0].Title != "Improve project quality" {
		t.Errorf("default directive title = %q, unexpected", dirs[0].Title)
	}
}

func TestCov8_directivesFromPillars_noPillarField(t *testing.T) {
	// Goals with empty Pillar field are skipped → falls back to default directive
	gs := []goals.Goal{
		{ID: "g1", Description: "goal 1"},
		{ID: "g2", Description: "goal 2"},
	}
	dirs := directivesFromPillars(gs)
	if len(dirs) != 1 {
		t.Fatalf("directivesFromPillars(no-pillar) returned %d directives, want 1", len(dirs))
	}
}

func TestCov8_directivesFromPillars_withPillars(t *testing.T) {
	gs := []goals.Goal{
		{ID: "g1", Pillar: "quality"},
		{ID: "g2", Pillar: "coverage"},
		{ID: "g3", Pillar: "quality"}, // duplicate — should be deduped
	}
	dirs := directivesFromPillars(gs)
	if len(dirs) != 2 {
		t.Fatalf("directivesFromPillars(2 unique pillars) returned %d directives, want 2", len(dirs))
	}
	if dirs[0].Title != "Strengthen quality" {
		t.Errorf("directive[0].Title = %q, want 'Strengthen quality'", dirs[0].Title)
	}
	if dirs[0].Number != 1 {
		t.Errorf("directive[0].Number = %d, want 1", dirs[0].Number)
	}
	if dirs[1].Number != 2 {
		t.Errorf("directive[1].Number = %d, want 2", dirs[1].Number)
	}
}

// ---------------------------------------------------------------------------
// temper.go — runTemperValidate dry-run path (30.4% → higher)
// ---------------------------------------------------------------------------

func TestCov8_runTemperValidate_dryRun(t *testing.T) {
	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	cmd := &cobra.Command{}
	var buf strings.Builder
	cmd.SetOut(&buf)

	err := runTemperValidate(cmd, []string{"file1.md", "file2.md"})
	if err != nil {
		t.Fatalf("runTemperValidate dry-run: %v", err)
	}
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected dry-run message, got %q", buf.String())
	}
}

func TestCov8_runTemperValidate_dryRunZeroArgs(t *testing.T) {
	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	cmd := &cobra.Command{}
	cmd.SetOut(&strings.Builder{})

	err := runTemperValidate(cmd, []string{})
	if err != nil {
		t.Fatalf("runTemperValidate dry-run 0 args: %v", err)
	}
}

// ---------------------------------------------------------------------------
// temper.go — runTemperLock (21.7% → higher)
// ---------------------------------------------------------------------------

func TestCov8_runTemperLock_noFiles(t *testing.T) {
	// Need a temp dir with no matching files, then chdir into it
	tmp := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = false

	cmd := &cobra.Command{}
	// "nonexistent_xyz_*.md" won't glob-match anything in the empty tmp dir
	err = runTemperLock(cmd, []string{"nonexistent_xyz_*.md"})
	if err == nil {
		t.Fatal("expected error for no matching files")
	}
	if !strings.Contains(err.Error(), "no files found") {
		t.Errorf("expected 'no files found' error, got %v", err)
	}
}

func TestCov8_runTemperLock_dryRun(t *testing.T) {
	tmp := t.TempDir()

	// Create a real .md file so expandFilePatterns returns it
	mdFile := filepath.Join(tmp, "artifact.md")
	content := "# Test\n\nSome content.\n"
	if err := os.WriteFile(mdFile, []byte(content), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	cmd := &cobra.Command{}
	err = runTemperLock(cmd, []string{"artifact.md"})
	if err != nil {
		t.Fatalf("runTemperLock dry-run: %v", err)
	}
}

// ---------------------------------------------------------------------------
// batch_promote.go — promoteEntry dry-run (25% → higher)
// ---------------------------------------------------------------------------

func TestCov8_promoteEntry_dryRun(t *testing.T) {
	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	entry := pool.PoolEntry{
		AgeString: "2h",
	}
	// Manually set the embedded Candidate fields
	entry.Candidate = types.Candidate{
		ID:   "cand-test-id",
		Tier: types.TierBronze,
	}

	result := &batchPromoteResult{}
	err := promoteEntry(nil, entry, result) // pool is nil — fine for dry-run
	if err != nil {
		t.Fatalf("promoteEntry dry-run: %v", err)
	}
	if result.Promoted != 1 {
		t.Errorf("result.Promoted = %d, want 1", result.Promoted)
	}
}

// ---------------------------------------------------------------------------
// ratchet_find.go — runRatchetFind no-matches + json output (44% → higher)
// ---------------------------------------------------------------------------

func TestCov8_runRatchetFind_noMatches(t *testing.T) {
	origOutput := output
	defer func() { output = origOutput }()
	output = "" // text mode

	cmd := &cobra.Command{}
	// Use a pattern that won't match anything
	err := runRatchetFind(cmd, []string{"zzz_no_match_ever_xyz_*.impossible"})
	if err != nil {
		t.Fatalf("runRatchetFind no-matches: %v", err)
	}
}

func TestCov8_runRatchetFind_jsonOutput(t *testing.T) {
	origOutput := output
	defer func() { output = origOutput }()
	output = "json"

	cmd := &cobra.Command{}
	err := runRatchetFind(cmd, []string{"zzz_no_match_xyz_*.impossible"})
	if err != nil {
		t.Fatalf("runRatchetFind json output: %v", err)
	}
}

// ---------------------------------------------------------------------------
// inject.go — gatherLearnings / gatherPatterns with empty dirs (50% → higher)
// ---------------------------------------------------------------------------

func TestCov8_gatherLearnings_emptyDir(t *testing.T) {
	tmp := t.TempDir()

	origNoCite := injectNoCite
	defer func() { injectNoCite = origNoCite }()
	injectNoCite = true // skip citation recording branch

	got := gatherLearnings(tmp, "test query", "sess-test", "", 1.0)
	// No learnings in an empty dir — should return nil or empty slice
	if len(got) != 0 {
		t.Errorf("gatherLearnings(empty dir) returned %d items, want 0", len(got))
	}
}

func TestCov8_gatherPatterns_emptyDir(t *testing.T) {
	tmp := t.TempDir()

	origNoCite := injectNoCite
	defer func() { injectNoCite = origNoCite }()
	injectNoCite = true

	got := gatherPatterns(tmp, "test query", "sess-test", "", 1.0)
	if len(got) != 0 {
		t.Errorf("gatherPatterns(empty dir) returned %d items, want 0", len(got))
	}
}

// ---------------------------------------------------------------------------
// temper.go — runTemperStatus with chdir to empty dir (50% → higher)
// ---------------------------------------------------------------------------

func TestCov8_runTemperStatus_emptyDir(t *testing.T) {
	tmp := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "" // default text output

	cmd := &cobra.Command{}
	err = runTemperStatus(cmd, nil)
	if err != nil {
		t.Fatalf("runTemperStatus empty dir: %v", err)
	}
}

func TestCov8_runTemperStatus_jsonOutput(t *testing.T) {
	tmp := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origOutput := output
	defer func() { output = origOutput }()
	output = "json"

	cmd := &cobra.Command{}
	err = runTemperStatus(cmd, nil)
	if err != nil {
		t.Fatalf("runTemperStatus json: %v", err)
	}
}
