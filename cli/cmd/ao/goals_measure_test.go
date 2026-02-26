package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boshu2/agentops/cli/internal/goals"
)

func TestGoalsMeasure_BasicRun(t *testing.T) {
	dir := t.TempDir()

	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| pass-gate | ` + "`exit 0`" + ` | 5 | Always passes |
| fail-gate | ` + "`exit 1`" + ` | 3 | Always fails |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldJSON := goalsJSON
	oldTimeout := goalsTimeout
	oldGoalID := goalsMeasureGoalID
	oldDirectives := goalsMeasureDirectives
	defer func() {
		goalsFile = oldFile
		goalsJSON = oldJSON
		goalsTimeout = oldTimeout
		goalsMeasureGoalID = oldGoalID
		goalsMeasureDirectives = oldDirectives
	}()

	goalsFile = goalsPath
	goalsJSON = true
	goalsTimeout = 10
	goalsMeasureGoalID = ""
	goalsMeasureDirectives = false

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := goalsMeasureCmd.RunE(goalsMeasureCmd, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("measure returned error: %v", err)
	}

	buf := make([]byte, 16384)
	n, _ := r.Read(buf)

	var snap goals.Snapshot
	if err := json.Unmarshal(buf[:n], &snap); err != nil {
		t.Fatalf("failed to decode JSON: %v (raw: %s)", err, string(buf[:n]))
	}

	if snap.Summary.Total != 2 {
		t.Errorf("Total = %d, want 2", snap.Summary.Total)
	}
	if snap.Summary.Passing != 1 {
		t.Errorf("Passing = %d, want 1", snap.Summary.Passing)
	}
	if snap.Summary.Failing != 1 {
		t.Errorf("Failing = %d, want 1", snap.Summary.Failing)
	}
}

func TestGoalsMeasure_SingleGoalFilter(t *testing.T) {
	dir := t.TempDir()

	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| target-goal | ` + "`exit 0`" + ` | 5 | Target |
| other-goal | ` + "`exit 0`" + ` | 5 | Other |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldJSON := goalsJSON
	oldTimeout := goalsTimeout
	oldGoalID := goalsMeasureGoalID
	oldDirectives := goalsMeasureDirectives
	defer func() {
		goalsFile = oldFile
		goalsJSON = oldJSON
		goalsTimeout = oldTimeout
		goalsMeasureGoalID = oldGoalID
		goalsMeasureDirectives = oldDirectives
	}()

	goalsFile = goalsPath
	goalsJSON = true
	goalsTimeout = 10
	goalsMeasureGoalID = "target-goal"
	goalsMeasureDirectives = false

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := goalsMeasureCmd.RunE(goalsMeasureCmd, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("measure returned error: %v", err)
	}

	buf := make([]byte, 16384)
	n, _ := r.Read(buf)

	var snap goals.Snapshot
	if err := json.Unmarshal(buf[:n], &snap); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if snap.Summary.Total != 1 {
		t.Errorf("Total = %d, want 1 (filtered to single goal)", snap.Summary.Total)
	}
	if len(snap.Goals) != 1 || snap.Goals[0].GoalID != "target-goal" {
		t.Errorf("expected only target-goal in results")
	}
}

func TestGoalsMeasure_GoalNotFound(t *testing.T) {
	dir := t.TempDir()

	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| only-goal | ` + "`exit 0`" + ` | 5 | Only |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldGoalID := goalsMeasureGoalID
	oldDirectives := goalsMeasureDirectives
	oldTimeout := goalsTimeout
	defer func() {
		goalsFile = oldFile
		goalsMeasureGoalID = oldGoalID
		goalsMeasureDirectives = oldDirectives
		goalsTimeout = oldTimeout
	}()

	goalsFile = goalsPath
	goalsMeasureGoalID = "nonexistent-goal"
	goalsMeasureDirectives = false
	goalsTimeout = 10

	err := goalsMeasureCmd.RunE(goalsMeasureCmd, nil)
	if err == nil {
		t.Fatal("expected error for nonexistent goal ID")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err.Error())
	}
}

func TestGoalsMeasure_DirectivesAndGoalConflict(t *testing.T) {
	dir := t.TempDir()

	md := `# Goals

Mission.

## Directives

### 1. First

Body.

**Steer:** increase

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| g1 | ` + "`exit 0`" + ` | 5 | Goal |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldGoalID := goalsMeasureGoalID
	oldDirectives := goalsMeasureDirectives
	defer func() {
		goalsFile = oldFile
		goalsMeasureGoalID = oldGoalID
		goalsMeasureDirectives = oldDirectives
	}()

	goalsFile = goalsPath
	goalsMeasureGoalID = "g1"
	goalsMeasureDirectives = true

	err := goalsMeasureCmd.RunE(goalsMeasureCmd, nil)
	if err == nil {
		t.Fatal("expected error when --directives and --goal are both set")
	}
	if !strings.Contains(err.Error(), "cannot be combined") {
		t.Errorf("error = %q, want 'cannot be combined'", err.Error())
	}
}

func TestGoalsMeasure_DirectivesOutput(t *testing.T) {
	dir := t.TempDir()

	md := `# Goals

Mission.

## Directives

### 1. Ship fast

Deploy continuously.

**Steer:** increase

### 2. Stay secure

No vulnerabilities.

**Steer:** hold

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| g1 | ` + "`exit 0`" + ` | 5 | Goal |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldGoalID := goalsMeasureGoalID
	oldDirectives := goalsMeasureDirectives
	defer func() {
		goalsFile = oldFile
		goalsMeasureGoalID = oldGoalID
		goalsMeasureDirectives = oldDirectives
	}()

	goalsFile = goalsPath
	goalsMeasureGoalID = ""
	goalsMeasureDirectives = true

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := goalsMeasureCmd.RunE(goalsMeasureCmd, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("measure --directives returned error: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)

	var dirs []goals.Directive
	if err := json.Unmarshal(buf[:n], &dirs); err != nil {
		t.Fatalf("failed to decode directives JSON: %v", err)
	}

	if len(dirs) != 2 {
		t.Fatalf("expected 2 directives, got %d", len(dirs))
	}
	if dirs[0].Title != "Ship fast" {
		t.Errorf("dirs[0].Title = %q, want 'Ship fast'", dirs[0].Title)
	}
	if dirs[1].Steer != "hold" {
		t.Errorf("dirs[1].Steer = %q, want 'hold'", dirs[1].Steer)
	}
}

func TestGoalsMeasure_DirectivesOnYAML(t *testing.T) {
	dir := t.TempDir()

	yaml := `version: 2
mission: Test
goals:
  - id: g1
    description: Goal one
    check: "exit 0"
    weight: 5
`
	goalsPath := filepath.Join(dir, "GOALS.yaml")
	if err := os.WriteFile(goalsPath, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldGoalID := goalsMeasureGoalID
	oldDirectives := goalsMeasureDirectives
	defer func() {
		goalsFile = oldFile
		goalsMeasureGoalID = oldGoalID
		goalsMeasureDirectives = oldDirectives
	}()

	goalsFile = goalsPath
	goalsMeasureGoalID = ""
	goalsMeasureDirectives = true

	// Should not error, just warn on stderr and return nil
	err := goalsMeasureCmd.RunE(goalsMeasureCmd, nil)
	if err != nil {
		t.Fatalf("expected nil error for YAML + --directives, got: %v", err)
	}
}

func TestGoalsMeasure_MissingGoalsFile(t *testing.T) {
	oldFile := goalsFile
	oldDirectives := goalsMeasureDirectives
	defer func() {
		goalsFile = oldFile
		goalsMeasureDirectives = oldDirectives
	}()

	goalsFile = "/nonexistent/GOALS.md"
	goalsMeasureDirectives = false

	err := goalsMeasureCmd.RunE(goalsMeasureCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing goals file")
	}
}

func TestGoalsMeasure_CmdAttributes(t *testing.T) {
	if goalsMeasureCmd.Use != "measure" {
		t.Errorf("Use = %q, want measure", goalsMeasureCmd.Use)
	}
	if goalsMeasureCmd.GroupID != "measurement" {
		t.Errorf("GroupID = %q, want measurement", goalsMeasureCmd.GroupID)
	}
	found := false
	for _, a := range goalsMeasureCmd.Aliases {
		if a == "m" {
			found = true
		}
	}
	if !found {
		t.Error("expected alias 'm' for measure command")
	}
}
