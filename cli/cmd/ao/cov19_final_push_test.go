package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/boshu2/agentops/cli/internal/goals"
	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// goals_migrate.go — directivesFromPillars with actual pillars (0% → higher)
// ---------------------------------------------------------------------------

func TestCov19_directivesFromPillars_withPillars(t *testing.T) {
	gs := []goals.Goal{
		{ID: "g1", Pillar: "quality"},
		{ID: "g2", Pillar: "speed"},
		{ID: "g3", Pillar: "quality"}, // duplicate — should deduplicate
		{ID: "g4", Pillar: ""},        // no pillar — skipped
	}
	result := directivesFromPillars(gs)
	if len(result) != 2 {
		t.Fatalf("directivesFromPillars: got %d directives, want 2", len(result))
	}
	if result[0].Title != "Strengthen quality" {
		t.Errorf("first directive title: got %q, want 'Strengthen quality'", result[0].Title)
	}
	if result[1].Number != 2 {
		t.Errorf("second directive number: got %d, want 2", result[1].Number)
	}
}

func TestCov19_directivesFromPillars_noPillarSkipped(t *testing.T) {
	gs := []goals.Goal{
		{ID: "g1", Pillar: ""},
		{ID: "g2", Pillar: ""},
	}
	result := directivesFromPillars(gs)
	// All pillars empty → falls through to default single directive
	if len(result) != 1 {
		t.Fatalf("directivesFromPillars all-empty: got %d, want 1", len(result))
	}
}

// ---------------------------------------------------------------------------
// temper.go — runTemperValidate dry-run path (39.1% → higher)
// ---------------------------------------------------------------------------

func TestCov19_runTemperValidate_dryRun(t *testing.T) {
	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	cmd := &cobra.Command{}
	err := runTemperValidate(cmd, []string{"*.md"})
	if err != nil {
		t.Fatalf("runTemperValidate dry-run: %v", err)
	}
}

func TestCov19_runTemperValidate_dryRunNoArgs(t *testing.T) {
	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	cmd := &cobra.Command{}
	err := runTemperValidate(cmd, nil)
	if err != nil {
		t.Fatalf("runTemperValidate dry-run no args: %v", err)
	}
}

// ---------------------------------------------------------------------------
// session_outcome.go — runSessionOutcome dry-run path (36.7% → higher)
// ---------------------------------------------------------------------------

func TestCov19_runSessionOutcome_dryRunWithPath(t *testing.T) {
	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	// When len(args) > 0, the function sets transcriptPath = args[0]
	// then hits GetDryRun() → returns nil immediately.
	err := runSessionOutcome(sessionOutcomeCmd, []string{"/tmp/fake-session-transcript.jsonl"})
	if err != nil {
		t.Fatalf("runSessionOutcome dry-run with path: %v", err)
	}
}

// ---------------------------------------------------------------------------
// context.go — runContextStatus empty pool path (38.1% → higher)
// ---------------------------------------------------------------------------

func TestCov19_runContextStatus_emptyPool(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// No .agents/ao/context dir → collectTrackedSessionStatuses returns empty
	cmd := &cobra.Command{}
	origOutput := output
	defer func() { output = origOutput }()
	output = "table"

	err := runContextStatus(cmd, nil)
	if err != nil {
		t.Fatalf("runContextStatus empty pool: %v", err)
	}
}

func TestCov19_runContextStatus_jsonOutput(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cmd := &cobra.Command{}
	origOutput := output
	defer func() { output = origOutput }()
	output = "json"

	err := runContextStatus(cmd, nil)
	if err != nil {
		t.Fatalf("runContextStatus json output: %v", err)
	}
}

// ---------------------------------------------------------------------------
// goals_migrate.go — runGoalsMigrate v1 → v2 path (37.2% → higher)
// ---------------------------------------------------------------------------

func TestCov19_runGoalsMigrate_v1ToV2(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Write a v1 GOALS.yaml
	v1Content := `version: 1
goals:
  - id: test-goal
    description: "Test goal"
    check: "echo ok"
    weight: 5
`
	if err := os.WriteFile(filepath.Join(tmp, "GOALS.yaml"), []byte(v1Content), 0o644); err != nil {
		t.Fatalf("write GOALS.yaml: %v", err)
	}

	origMigrateToMD := migrateToMD
	defer func() { migrateToMD = origMigrateToMD }()
	migrateToMD = false

	cmd := &cobra.Command{}
	err := runGoalsMigrate(cmd, nil)
	if err != nil {
		t.Fatalf("runGoalsMigrate v1→v2: %v", err)
	}

	// Verify backup was created
	if _, statErr := os.Stat(filepath.Join(tmp, "GOALS.yaml.v1.bak")); os.IsNotExist(statErr) {
		t.Error("expected backup file GOALS.yaml.v1.bak to be created")
	}
}

func TestCov19_runGoalsMigrate_yamlToMD_withPillars(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Write a v2 GOALS.yaml with pillar groups
	v2Content := `version: 2
mission: "Test mission"
goals:
  - id: test-goal
    description: "Test goal"
    check: "echo ok"
    weight: 5
    pillar: "quality"
`
	if err := os.WriteFile(filepath.Join(tmp, "GOALS.yaml"), []byte(v2Content), 0o644); err != nil {
		t.Fatalf("write GOALS.yaml: %v", err)
	}

	origMigrateToMD := migrateToMD
	defer func() { migrateToMD = origMigrateToMD }()
	migrateToMD = true

	cmd := &cobra.Command{}
	err := runGoalsMigrate(cmd, nil)
	if err != nil {
		t.Fatalf("runGoalsMigrate yaml→md with pillars: %v", err)
	}

	// Verify GOALS.md was created
	if _, statErr := os.Stat(filepath.Join(tmp, "GOALS.md")); os.IsNotExist(statErr) {
		t.Error("expected GOALS.md to be created")
	}
}
