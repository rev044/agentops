package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// --- runQuickstart tests ---

func TestCov3_quickstart_runQuickstart_minimal(t *testing.T) {
	tmp := t.TempDir()
	chdirTo(t, tmp)

	oldMinimal := minimal
	minimal = true
	defer func() { minimal = oldMinimal }()

	oldNoBeads := noBeads
	noBeads = true
	defer func() { noBeads = oldNoBeads }()

	cmd := &cobra.Command{}
	got := captureJSONStdout(t, func() {
		err := runQuickstart(cmd, nil)
		if err != nil {
			t.Fatalf("runQuickstart minimal: %v", err)
		}
	})

	if !strings.Contains(got, "Minimal setup complete") {
		t.Fatalf("expected minimal completion message, got: %s", got)
	}

	// Verify directories were created
	dirs := []string{
		".agents/research",
		".agents/synthesis",
		".agents/specs",
		".agents/learnings",
		".agents/patterns",
		".agents/retros",
		".agents/handoff",
	}
	for _, dir := range dirs {
		path := filepath.Join(tmp, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("expected directory %s to exist", dir)
		}
	}
}

func TestCov3_quickstart_runQuickstart_fullNoBeads(t *testing.T) {
	tmp := t.TempDir()
	chdirTo(t, tmp)

	oldMinimal := minimal
	minimal = false
	defer func() { minimal = oldMinimal }()

	oldNoBeads := noBeads
	noBeads = true
	defer func() { noBeads = oldNoBeads }()

	cmd := &cobra.Command{}
	got := captureJSONStdout(t, func() {
		err := runQuickstart(cmd, nil)
		if err != nil {
			t.Fatalf("runQuickstart full no-beads: %v", err)
		}
	})

	if !strings.Contains(got, "SETUP COMPLETE") {
		t.Fatalf("expected setup complete message, got: %s", got)
	}

	// Verify starter pack files
	if _, err := os.Stat(filepath.Join(tmp, ".agents", "patterns", "context-boundaries.md")); os.IsNotExist(err) {
		t.Fatal("expected context-boundaries.md to be created")
	}
	if _, err := os.Stat(filepath.Join(tmp, ".agents", "patterns", "pre-mortem-first.md")); os.IsNotExist(err) {
		t.Fatal("expected pre-mortem-first.md to be created")
	}
	if _, err := os.Stat(filepath.Join(tmp, ".agents", "learnings", "session-hygiene.md")); os.IsNotExist(err) {
		t.Fatal("expected session-hygiene.md to be created")
	}
}

func TestCov3_quickstart_runQuickstart_createsClaudeMd(t *testing.T) {
	tmp := t.TempDir()
	chdirTo(t, tmp)

	oldMinimal := minimal
	minimal = false
	defer func() { minimal = oldMinimal }()

	oldNoBeads := noBeads
	noBeads = true
	defer func() { noBeads = oldNoBeads }()

	cmd := &cobra.Command{}
	captureJSONStdout(t, func() {
		err := runQuickstart(cmd, nil)
		if err != nil {
			t.Fatalf("runQuickstart: %v", err)
		}
	})

	claudeMdPath := filepath.Join(tmp, "CLAUDE.md")
	if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
		t.Fatal("expected CLAUDE.md to be created")
	}

	content, err := os.ReadFile(claudeMdPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Quick Start") {
		t.Fatal("expected CLAUDE.md to contain Quick Start section")
	}
}

func TestCov3_quickstart_runQuickstart_existingClaudeMd(t *testing.T) {
	tmp := t.TempDir()
	chdirTo(t, tmp)

	// Pre-create CLAUDE.md
	claudeMdPath := filepath.Join(tmp, "CLAUDE.md")
	if err := os.WriteFile(claudeMdPath, []byte("# Existing\n"), 0644); err != nil {
		t.Fatal(err)
	}

	oldMinimal := minimal
	minimal = false
	defer func() { minimal = oldMinimal }()

	oldNoBeads := noBeads
	noBeads = true
	defer func() { noBeads = oldNoBeads }()

	cmd := &cobra.Command{}
	got := captureJSONStdout(t, func() {
		err := runQuickstart(cmd, nil)
		if err != nil {
			t.Fatalf("runQuickstart: %v", err)
		}
	})

	if !strings.Contains(got, "CLAUDE.md already exists") {
		t.Fatalf("expected 'already exists' message, got: %s", got)
	}

	// Verify original content preserved
	content, err := os.ReadFile(claudeMdPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "# Existing\n" {
		t.Fatal("CLAUDE.md should not have been overwritten")
	}
}

// --- quickstartBeadsStep tests ---

func TestCov3_quickstart_quickstartBeadsStep_noBeads(t *testing.T) {
	tmp := t.TempDir()

	oldNoBeads := noBeads
	noBeads = true
	defer func() { noBeads = oldNoBeads }()

	// Pre-create the .agents dir for tasks.json creation
	if err := os.MkdirAll(filepath.Join(tmp, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	got := captureJSONStdout(t, func() {
		quickstartBeadsStep(tmp)
	})

	if !strings.Contains(got, "Skipping beads") {
		t.Fatalf("expected skipping beads message, got: %s", got)
	}

	// Verify tasks.json was created
	tasksPath := filepath.Join(tmp, ".agents", "tasks.json")
	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		t.Fatal("expected tasks.json to be created")
	}
}

// --- quickstartClaudeMdStep tests ---

func TestCov3_quickstart_quickstartClaudeMdStep_creates(t *testing.T) {
	tmp := t.TempDir()

	got := captureJSONStdout(t, func() {
		quickstartClaudeMdStep(tmp)
	})

	if !strings.Contains(got, "Created CLAUDE.md") {
		t.Fatalf("expected creation message, got: %s", got)
	}

	claudeMdPath := filepath.Join(tmp, "CLAUDE.md")
	if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
		t.Fatal("expected CLAUDE.md to be created")
	}
}

func TestCov3_quickstart_quickstartClaudeMdStep_alreadyExists(t *testing.T) {
	tmp := t.TempDir()
	claudeMdPath := filepath.Join(tmp, "CLAUDE.md")
	if err := os.WriteFile(claudeMdPath, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	got := captureJSONStdout(t, func() {
		quickstartClaudeMdStep(tmp)
	})

	if !strings.Contains(got, "already exists") {
		t.Fatalf("expected 'already exists' message, got: %s", got)
	}
}

// --- createStarterPack tests ---

func TestCov3_quickstart_createStarterPack(t *testing.T) {
	tmp := t.TempDir()

	// Create needed directories
	dirs := []string{".agents/patterns", ".agents/learnings"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}

	captureJSONStdout(t, func() {
		err := createStarterPack(tmp)
		if err != nil {
			t.Fatalf("createStarterPack: %v", err)
		}
	})

	// Verify files exist
	expected := []string{
		".agents/patterns/context-boundaries.md",
		".agents/patterns/pre-mortem-first.md",
		".agents/learnings/session-hygiene.md",
	}
	for _, name := range expected {
		path := filepath.Join(tmp, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("expected %s to be created", name)
		}
	}
}

// --- createTasksFile tests ---

func TestCov3_quickstart_createTasksFile(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	captureJSONStdout(t, func() {
		createTasksFile(tmp)
	})

	tasksPath := filepath.Join(tmp, ".agents", "tasks.json")
	content, err := os.ReadFile(tasksPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "tasks") {
		t.Fatalf("expected tasks field in file, got: %s", string(content))
	}
}

// --- createProjectClaudeMd tests ---

func TestCov3_quickstart_createProjectClaudeMd(t *testing.T) {
	tmp := t.TempDir()

	err := createProjectClaudeMd(tmp)
	if err != nil {
		t.Fatalf("createProjectClaudeMd: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmp, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}

	// Should contain the directory name as the title
	dirName := filepath.Base(tmp)
	if !strings.Contains(string(content), dirName) {
		t.Fatalf("expected CLAUDE.md to contain dir name %q, got: %s", dirName, string(content))
	}
}

// --- showNextSteps tests ---

func TestCov3_quickstart_showNextSteps_withBeads(t *testing.T) {
	got := captureJSONStdout(t, func() {
		showNextSteps(true)
	})

	if !strings.Contains(got, "Create your first issue") {
		t.Fatalf("expected beads next steps, got: %s", got)
	}
}

func TestCov3_quickstart_showNextSteps_withoutBeads(t *testing.T) {
	got := captureJSONStdout(t, func() {
		showNextSteps(false)
	})

	if !strings.Contains(got, "Start Claude in your project") {
		t.Fatalf("expected no-beads next steps, got: %s", got)
	}
}
