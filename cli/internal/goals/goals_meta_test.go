package goals_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/goals"
)

func TestGoalsMeta_NoMetaGoals(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| health-gate | ` + "`exit 0`" + ` | 5 | Health gate |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	// No meta-goals — should print message and return nil
	err := goals.RunMeta(goals.MetaOptions{
		GoalsFile: goalsPath,
		JSON:      false,
		Timeout:   10 * time.Second,
		Stdout:    &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("meta returned error for no meta-goals: %v", err)
	}
}

func TestGoalsMeta_AllMetaPass(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	yaml := `version: 2
mission: Test
goals:
  - id: meta-one
    description: Meta goal one
    check: "exit 0"
    weight: 5
    type: meta
  - id: meta-two
    description: Meta goal two
    check: "exit 0"
    weight: 3
    type: meta
  - id: health-one
    description: Health goal
    check: "exit 0"
    weight: 5
    type: health
`
	goalsPath := filepath.Join(dir, "GOALS.yaml")
	if err := os.WriteFile(goalsPath, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	err := goals.RunMeta(goals.MetaOptions{
		GoalsFile: goalsPath,
		JSON:      false,
		Timeout:   10 * time.Second,
		Stdout:    &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("meta returned error: %v", err)
	}
}

func TestGoalsMeta_MetaFailure_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	yaml := `version: 2
mission: Test
goals:
  - id: meta-fail
    description: Failing meta goal
    check: "exit 1"
    weight: 5
    type: meta
`
	goalsPath := filepath.Join(dir, "GOALS.yaml")
	if err := os.WriteFile(goalsPath, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	err := goals.RunMeta(goals.MetaOptions{
		GoalsFile: goalsPath,
		JSON:      false,
		Timeout:   10 * time.Second,
		Stdout:    &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error when meta-goals fail")
	}
	if !strings.Contains(err.Error(), "meta-goal failures") {
		t.Errorf("error = %q, want 'meta-goal failures'", err.Error())
	}
}

func TestGoalsMeta_JSONOutput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	yaml := `version: 2
mission: Test
goals:
  - id: meta-json
    description: JSON meta
    check: "exit 0"
    weight: 5
    type: meta
`
	goalsPath := filepath.Join(dir, "GOALS.yaml")
	if err := os.WriteFile(goalsPath, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := goals.RunMeta(goals.MetaOptions{
		GoalsFile: goalsPath,
		JSON:      true,
		Timeout:   10 * time.Second,
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("meta returned error: %v", err)
	}

	var snap goals.Snapshot
	if err := json.Unmarshal(buf.Bytes(), &snap); err != nil {
		t.Fatalf("failed to decode JSON: %v (raw: %s)", err, buf.String())
	}

	if snap.Summary.Total != 1 {
		t.Errorf("Total = %d, want 1 (only meta-goals)", snap.Summary.Total)
	}
}

func TestGoalsMeta_FiltersOutNonMeta(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	yaml := `version: 2
mission: Test
goals:
  - id: meta-g
    description: Meta
    check: "exit 0"
    weight: 5
    type: meta
  - id: health-g
    description: Health
    check: "exit 0"
    weight: 5
    type: health
  - id: quality-g
    description: Quality
    check: "exit 0"
    weight: 5
    type: quality
`
	goalsPath := filepath.Join(dir, "GOALS.yaml")
	if err := os.WriteFile(goalsPath, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := goals.RunMeta(goals.MetaOptions{
		GoalsFile: goalsPath,
		JSON:      true,
		Timeout:   10 * time.Second,
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("meta returned error: %v", err)
	}

	var snap goals.Snapshot
	if err := json.Unmarshal(buf.Bytes(), &snap); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	// Should only have the meta goal
	if snap.Summary.Total != 1 {
		t.Errorf("Total = %d, want 1 (only meta-goals)", snap.Summary.Total)
	}
	if len(snap.Goals) != 1 || snap.Goals[0].GoalID != "meta-g" {
		t.Error("expected only meta-g in results")
	}
}

func TestGoalsMeta_MissingGoalsFile(t *testing.T) {
	t.Parallel()

	err := goals.RunMeta(goals.MetaOptions{
		GoalsFile: "/nonexistent/GOALS.yaml",
		Stdout:    &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error for missing goals file")
	}
}
