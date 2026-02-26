package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boshu2/agentops/cli/internal/goals"
)

func TestGoalsMeta_NoMetaGoals(t *testing.T) {
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

	oldFile := goalsFile
	oldJSON := goalsJSON
	oldTimeout := goalsTimeout
	defer func() {
		goalsFile = oldFile
		goalsJSON = oldJSON
		goalsTimeout = oldTimeout
	}()
	goalsFile = goalsPath
	goalsJSON = false
	goalsTimeout = 10

	// No meta-goals — should print message and return nil
	err := goalsMetaCmd.RunE(goalsMetaCmd, nil)
	if err != nil {
		t.Fatalf("meta returned error for no meta-goals: %v", err)
	}
}

func TestGoalsMeta_AllMetaPass(t *testing.T) {
	dir := t.TempDir()

	// Create a GOALS.md file, then manually build the GoalFile with meta type
	// since the markdown format doesn't have a type column. We test the RunE
	// by providing a YAML file that supports types.
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

	oldFile := goalsFile
	oldJSON := goalsJSON
	oldTimeout := goalsTimeout
	defer func() {
		goalsFile = oldFile
		goalsJSON = oldJSON
		goalsTimeout = oldTimeout
	}()
	goalsFile = goalsPath
	goalsJSON = false
	goalsTimeout = 10

	err := goalsMetaCmd.RunE(goalsMetaCmd, nil)
	if err != nil {
		t.Fatalf("meta returned error: %v", err)
	}
}

func TestGoalsMeta_MetaFailure_ReturnsError(t *testing.T) {
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

	oldFile := goalsFile
	oldJSON := goalsJSON
	oldTimeout := goalsTimeout
	defer func() {
		goalsFile = oldFile
		goalsJSON = oldJSON
		goalsTimeout = oldTimeout
	}()
	goalsFile = goalsPath
	goalsJSON = false
	goalsTimeout = 10

	err := goalsMetaCmd.RunE(goalsMetaCmd, nil)
	if err == nil {
		t.Fatal("expected error when meta-goals fail")
	}
	if !strings.Contains(err.Error(), "meta-goal failures") {
		t.Errorf("error = %q, want 'meta-goal failures'", err.Error())
	}
}

func TestGoalsMeta_JSONOutput(t *testing.T) {
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

	oldFile := goalsFile
	oldJSON := goalsJSON
	oldTimeout := goalsTimeout
	defer func() {
		goalsFile = oldFile
		goalsJSON = oldJSON
		goalsTimeout = oldTimeout
	}()
	goalsFile = goalsPath
	goalsJSON = true
	goalsTimeout = 10

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := goalsMetaCmd.RunE(goalsMetaCmd, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("meta returned error: %v", err)
	}

	buf := make([]byte, 16384)
	n, _ := r.Read(buf)

	var snap goals.Snapshot
	if err := json.Unmarshal(buf[:n], &snap); err != nil {
		t.Fatalf("failed to decode JSON: %v (raw: %s)", err, string(buf[:n]))
	}

	if snap.Summary.Total != 1 {
		t.Errorf("Total = %d, want 1 (only meta-goals)", snap.Summary.Total)
	}
}

func TestGoalsMeta_FiltersOutNonMeta(t *testing.T) {
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

	oldFile := goalsFile
	oldJSON := goalsJSON
	oldTimeout := goalsTimeout
	defer func() {
		goalsFile = oldFile
		goalsJSON = oldJSON
		goalsTimeout = oldTimeout
	}()
	goalsFile = goalsPath
	goalsJSON = true
	goalsTimeout = 10

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := goalsMetaCmd.RunE(goalsMetaCmd, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("meta returned error: %v", err)
	}

	buf := make([]byte, 16384)
	n, _ := r.Read(buf)

	var snap goals.Snapshot
	if err := json.Unmarshal(buf[:n], &snap); err != nil {
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
	oldFile := goalsFile
	defer func() { goalsFile = oldFile }()

	goalsFile = "/nonexistent/GOALS.yaml"

	err := goalsMetaCmd.RunE(goalsMetaCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing goals file")
	}
}

func TestGoalsMeta_CmdAttributes(t *testing.T) {
	if goalsMetaCmd.Use != "meta" {
		t.Errorf("Use = %q, want meta", goalsMetaCmd.Use)
	}
	if goalsMetaCmd.GroupID != "management" {
		t.Errorf("GroupID = %q, want management", goalsMetaCmd.GroupID)
	}
}
