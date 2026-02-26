package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boshu2/agentops/cli/internal/goals"
)

func TestGoalsAdd_RejectsNonKebabID(t *testing.T) {
	dir := t.TempDir()
	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| existing | ` + "`echo ok`" + ` | 5 | Existing goal |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	defer func() { goalsFile = oldFile }()
	goalsFile = goalsPath

	tests := []struct {
		name string
		id   string
	}{
		{"uppercase", "MyGoal"},
		{"underscore", "my_goal"},
		{"spaces", "my goal"},
		{"starts with dash", "-my-goal"},
		{"ends with dash", "my-goal-"},
		{"special chars", "my@goal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := goalsAddCmd.RunE(goalsAddCmd, []string{tt.id, "echo test"})
			if err == nil {
				t.Error("expected error for non-kebab-case ID")
			}
			if err != nil && !strings.Contains(err.Error(), "kebab-case") {
				t.Errorf("error = %q, want it to mention kebab-case", err.Error())
			}
		})
	}
}

func TestGoalsAdd_RejectsDuplicateID(t *testing.T) {
	dir := t.TempDir()
	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| build-ok | ` + "`echo ok`" + ` | 5 | Build passes |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	defer func() { goalsFile = oldFile }()
	goalsFile = goalsPath

	err := goalsAddCmd.RunE(goalsAddCmd, []string{"build-ok", "echo test"})
	if err == nil {
		t.Fatal("expected error for duplicate goal ID")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want 'already exists'", err.Error())
	}
}

func TestGoalsAdd_InvalidType(t *testing.T) {
	dir := t.TempDir()
	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| existing | ` + "`echo ok`" + ` | 5 | Existing |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldType := goalsAddType
	oldDryRun := dryRun
	defer func() {
		goalsFile = oldFile
		goalsAddType = oldType
		dryRun = oldDryRun
	}()
	goalsFile = goalsPath
	goalsAddType = "bogus"
	dryRun = true // skip check validation

	err := goalsAddCmd.RunE(goalsAddCmd, []string{"new-goal", "echo ok"})
	if err == nil {
		t.Fatal("expected error for invalid goal type")
	}
	if !strings.Contains(err.Error(), "invalid type") {
		t.Errorf("error = %q, want 'invalid type'", err.Error())
	}
}

func TestGoalsAdd_DefaultsTypeToHealth(t *testing.T) {
	dir := t.TempDir()
	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| existing | ` + "`echo ok`" + ` | 5 | Existing |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldType := goalsAddType
	oldWeight := goalsAddWeight
	oldDesc := goalsAddDescription
	oldDryRun := dryRun
	defer func() {
		goalsFile = oldFile
		goalsAddType = oldType
		goalsAddWeight = oldWeight
		goalsAddDescription = oldDesc
		dryRun = oldDryRun
	}()

	goalsFile = goalsPath
	goalsAddType = ""
	goalsAddWeight = 5
	goalsAddDescription = "A new test goal"
	dryRun = true // skip command execution

	err := goalsAddCmd.RunE(goalsAddCmd, []string{"new-goal", "echo ok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Reload and verify the goal was added with default type
	gf, err := goals.LoadGoals(goalsPath)
	if err != nil {
		t.Fatalf("LoadGoals: %v", err)
	}

	var found *goals.Goal
	for i := range gf.Goals {
		if gf.Goals[i].ID == "new-goal" {
			found = &gf.Goals[i]
			break
		}
	}
	if found == nil {
		t.Fatal("new-goal not found in goals file")
	}
	if found.Type != goals.GoalTypeHealth {
		t.Errorf("Type = %q, want %q", found.Type, goals.GoalTypeHealth)
	}
}

func TestGoalsAdd_DescriptionFallsBackToID(t *testing.T) {
	dir := t.TempDir()
	md := `# Goals

Mission.

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| existing | ` + "`echo ok`" + ` | 5 | Existing |
`
	goalsPath := filepath.Join(dir, "GOALS.md")
	if err := os.WriteFile(goalsPath, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}

	oldFile := goalsFile
	oldType := goalsAddType
	oldWeight := goalsAddWeight
	oldDesc := goalsAddDescription
	oldDryRun := dryRun
	defer func() {
		goalsFile = oldFile
		goalsAddType = oldType
		goalsAddWeight = oldWeight
		goalsAddDescription = oldDesc
		dryRun = oldDryRun
	}()

	goalsFile = goalsPath
	goalsAddType = ""
	goalsAddWeight = 3
	goalsAddDescription = "" // empty description
	dryRun = true

	err := goalsAddCmd.RunE(goalsAddCmd, []string{"fallback-id", "echo ok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gf, err := goals.LoadGoals(goalsPath)
	if err != nil {
		t.Fatalf("LoadGoals: %v", err)
	}

	var found *goals.Goal
	for i := range gf.Goals {
		if gf.Goals[i].ID == "fallback-id" {
			found = &gf.Goals[i]
			break
		}
	}
	if found == nil {
		t.Fatal("fallback-id not found in goals file")
	}
	if found.Description != "fallback-id" {
		t.Errorf("Description = %q, want %q (fallback to ID)", found.Description, "fallback-id")
	}
}

func TestGoalsAdd_MissingGoalsFile(t *testing.T) {
	oldFile := goalsFile
	defer func() { goalsFile = oldFile }()

	goalsFile = "/nonexistent/path/GOALS.md"

	err := goalsAddCmd.RunE(goalsAddCmd, []string{"new-goal", "echo test"})
	if err == nil {
		t.Fatal("expected error for missing goals file")
	}
	if !strings.Contains(err.Error(), "loading goals") {
		t.Errorf("error = %q, want 'loading goals'", err.Error())
	}
}

func TestGoalsAdd_CobraArgsValidation(t *testing.T) {
	if goalsAddCmd.Args == nil {
		t.Fatal("goalsAddCmd.Args is nil, expected ExactArgs(2)")
	}
	// ExactArgs(2) should reject 0, 1, and 3 args
	if err := goalsAddCmd.Args(goalsAddCmd, []string{}); err == nil {
		t.Error("expected error for 0 args")
	}
	if err := goalsAddCmd.Args(goalsAddCmd, []string{"one"}); err == nil {
		t.Error("expected error for 1 arg")
	}
	if err := goalsAddCmd.Args(goalsAddCmd, []string{"one", "two", "three"}); err == nil {
		t.Error("expected error for 3 args")
	}
	if err := goalsAddCmd.Args(goalsAddCmd, []string{"id", "check"}); err != nil {
		t.Errorf("unexpected error for 2 args: %v", err)
	}
}
