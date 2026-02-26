package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuickstart_CommandExists(t *testing.T) {
	if quickstartCmd == nil {
		t.Fatal("quickstartCmd should not be nil")
	}
	if quickstartCmd.Use != "quick-start" {
		t.Errorf("quickstartCmd.Use = %q, want %q", quickstartCmd.Use, "quick-start")
	}
	if quickstartCmd.GroupID != "start" {
		t.Errorf("quickstartCmd.GroupID = %q, want %q", quickstartCmd.GroupID, "start")
	}
}

func TestQuickstart_HasFlags(t *testing.T) {
	if quickstartCmd.Flags().Lookup("no-beads") == nil {
		t.Error("quick-start should have --no-beads flag")
	}
	if quickstartCmd.Flags().Lookup("minimal") == nil {
		t.Error("quick-start should have --minimal flag")
	}
}

func TestQuickstart_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "quick-start" {
			found = true
			break
		}
	}
	if !found {
		t.Error("quickstartCmd should be registered on rootCmd")
	}
}

func TestQuickstart_CreateProjectClaudeMd_Content(t *testing.T) {
	tmp := t.TempDir()
	err := createProjectClaudeMd(tmp)
	if err != nil {
		t.Fatalf("createProjectClaudeMd: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(data)

	// Should contain key sections.
	for _, want := range []string{"Quick Start", "Session Protocol", "JIT Loading"} {
		if !strings.Contains(content, want) {
			t.Errorf("CLAUDE.md should contain %q", want)
		}
	}

	// Should use directory name as title.
	dirName := filepath.Base(tmp)
	if !strings.Contains(content, dirName) {
		t.Errorf("CLAUDE.md should contain directory name %q", dirName)
	}
}

func TestQuickstart_CreateTasksFile_ValidJSON(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	createTasksFile(tmp)

	data, err := os.ReadFile(filepath.Join(tmp, ".agents", "tasks.json"))
	if err != nil {
		t.Fatalf("read tasks.json: %v", err)
	}
	if !strings.Contains(string(data), "tasks") {
		t.Errorf("tasks.json should contain 'tasks' field, got: %s", string(data))
	}
	if !strings.Contains(string(data), "Beads-optional") {
		t.Errorf("tasks.json should contain note about beads-optional mode")
	}
}

func TestQuickstart_ShowNextSteps_WithBeads(t *testing.T) {
	// Should not panic.
	showNextSteps(true)
}

func TestQuickstart_ShowNextSteps_WithoutBeads(t *testing.T) {
	// Should not panic.
	showNextSteps(false)
}

func TestQuickstart_CreateStarterPack_CreatesPatterns(t *testing.T) {
	tmp := t.TempDir()
	for _, dir := range []string{".agents/patterns", ".agents/learnings"} {
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}

	err := createStarterPack(tmp)
	if err != nil {
		t.Fatalf("createStarterPack: %v", err)
	}

	expectedFiles := []string{
		".agents/patterns/context-boundaries.md",
		".agents/patterns/pre-mortem-first.md",
		".agents/learnings/session-hygiene.md",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(tmp, f)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("expected %s to exist", f)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("expected %s to have content", f)
		}
	}
}

func TestQuickstart_CreateStarterPack_PatternContent(t *testing.T) {
	tmp := t.TempDir()
	for _, dir := range []string{".agents/patterns", ".agents/learnings"} {
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}

	err := createStarterPack(tmp)
	if err != nil {
		t.Fatalf("createStarterPack: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, ".agents/patterns/context-boundaries.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "Fresh Context Per Phase") {
		t.Error("context-boundaries.md should contain 'Fresh Context Per Phase'")
	}
	if !strings.Contains(content, "40% Rule") {
		t.Error("context-boundaries.md should contain '40% Rule'")
	}
}
