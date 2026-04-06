package goals_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/goals"
)

func TestGoalsAdd_RejectsNonKebabID(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			err := goals.RunAdd(context.Background(), goals.AddOptions{
				ID:        tt.id,
				Check:     "echo test",
				GoalsFile: goalsPath,
				Weight:    5,
				Timeout:   10 * time.Second,
				DryRun:    true,
				Stdout:    &bytes.Buffer{},
			})
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
	t.Parallel()
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

	err := goals.RunAdd(context.Background(), goals.AddOptions{
		ID:        "build-ok",
		Check:     "echo test",
		GoalsFile: goalsPath,
		Weight:    5,
		Timeout:   10 * time.Second,
		DryRun:    true,
		Stdout:    &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error for duplicate goal ID")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want 'already exists'", err.Error())
	}
}

func TestGoalsAdd_InvalidType(t *testing.T) {
	t.Parallel()
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

	err := goals.RunAdd(context.Background(), goals.AddOptions{
		ID:        "new-goal",
		Check:     "echo ok",
		Type:      "bogus",
		GoalsFile: goalsPath,
		Weight:    5,
		Timeout:   10 * time.Second,
		DryRun:    true,
		Stdout:    &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error for invalid goal type")
	}
	if !strings.Contains(err.Error(), "invalid type") {
		t.Errorf("error = %q, want 'invalid type'", err.Error())
	}
}

func TestGoalsAdd_DefaultsTypeToHealth(t *testing.T) {
	t.Parallel()
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

	err := goals.RunAdd(context.Background(), goals.AddOptions{
		ID:          "new-goal",
		Check:       "echo ok",
		Type:        "",
		Description: "A new test goal",
		GoalsFile:   goalsPath,
		Weight:      5,
		Timeout:     10 * time.Second,
		DryRun:      true,
		Stdout:      &bytes.Buffer{},
	})
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
	t.Parallel()
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

	err := goals.RunAdd(context.Background(), goals.AddOptions{
		ID:          "fallback-id",
		Check:       "echo ok",
		Type:        "",
		Description: "",
		GoalsFile:   goalsPath,
		Weight:      3,
		Timeout:     10 * time.Second,
		DryRun:      true,
		Stdout:      &bytes.Buffer{},
	})
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
	t.Parallel()

	err := goals.RunAdd(context.Background(), goals.AddOptions{
		ID:        "new-goal",
		Check:     "echo test",
		GoalsFile: "/nonexistent/path/GOALS.md",
		Weight:    5,
		Timeout:   10 * time.Second,
		DryRun:    true,
		Stdout:    &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error for missing goals file")
	}
	if !strings.Contains(err.Error(), "loading goals") {
		t.Errorf("error = %q, want 'loading goals'", err.Error())
	}
}
