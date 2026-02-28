package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// notebook.go — runNotebookUpdate no-MEMORY.md paths (26.8% → higher)
// ---------------------------------------------------------------------------

func TestCov11_runNotebookUpdate_noMemoryFileNotQuiet(t *testing.T) {
	origFile := notebookMemoryFile
	origQuiet := notebookQuiet
	origSessionID := notebookSessionID
	defer func() {
		notebookMemoryFile = origFile
		notebookQuiet = origQuiet
		notebookSessionID = origSessionID
	}()
	notebookMemoryFile = "" // force findMemoryFile path
	notebookQuiet = false
	notebookSessionID = ""

	// chdir to temp dir that has no MEMORY.md anywhere in its lineage
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cmd := &cobra.Command{}
	var buf strings.Builder
	cmd.SetOut(&buf)

	err := runNotebookUpdate(cmd, nil)
	if err != nil {
		t.Fatalf("runNotebookUpdate notQuiet no-file: %v", err)
	}
	// "No MEMORY.md found" is printed to stdout directly (not cmd.OutOrStdout()),
	// so we just verify no error was returned and the function exited early.
}

func TestCov11_runNotebookUpdate_noMemoryFileQuiet(t *testing.T) {
	origFile := notebookMemoryFile
	origQuiet := notebookQuiet
	origSessionID := notebookSessionID
	defer func() {
		notebookMemoryFile = origFile
		notebookQuiet = origQuiet
		notebookSessionID = origSessionID
	}()
	notebookMemoryFile = "" // force findMemoryFile path
	notebookQuiet = true    // suppress output
	notebookSessionID = ""

	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cmd := &cobra.Command{}
	err := runNotebookUpdate(cmd, nil)
	if err != nil {
		t.Fatalf("runNotebookUpdate quiet no-file: %v", err)
	}
}

func TestCov11_runNotebookUpdate_sessionIDNotFound(t *testing.T) {
	origFile := notebookMemoryFile
	origQuiet := notebookQuiet
	origSessionID := notebookSessionID
	defer func() {
		notebookMemoryFile = origFile
		notebookQuiet = origQuiet
		notebookSessionID = origSessionID
	}()

	tmp := t.TempDir()

	// Create a real MEMORY.md so findMemoryFile step is bypassed
	memPath := filepath.Join(tmp, "MEMORY.md")
	if err := os.WriteFile(memPath, []byte("# Memory\n"), 0644); err != nil {
		t.Fatalf("write MEMORY.md: %v", err)
	}
	notebookMemoryFile = memPath
	notebookQuiet = false
	notebookSessionID = "sess-nonexistent-xyzxyz" // won't be in sessions dir

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cmd := &cobra.Command{}
	err := runNotebookUpdate(cmd, nil)
	// Should return nil (graceful skip when session not found)
	if err != nil {
		t.Fatalf("runNotebookUpdate session-not-found: %v", err)
	}
}

// ---------------------------------------------------------------------------
// task_sync.go — runTaskFeedback no-processable + dry-run (31.8% → higher)
// ---------------------------------------------------------------------------

func TestCov11_runTaskFeedback_noPendingWithLearnings(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Create tasks file with only pending tasks (no completed + learning_id)
	tasksDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	taskLine := `{"task_id":"t1","subject":"pending task","status":"pending","session_id":"sess-1","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}` + "\n"
	if err := os.WriteFile(filepath.Join(tasksDir, "tasks.jsonl"), []byte(taskLine), 0644); err != nil {
		t.Fatalf("write tasks: %v", err)
	}

	origSessionID := taskFeedbackSessionID
	defer func() { taskFeedbackSessionID = origSessionID }()
	taskFeedbackSessionID = "" // no session filter

	cmd := &cobra.Command{}
	err := runTaskFeedback(cmd, nil)
	if err != nil {
		t.Fatalf("runTaskFeedback noPendingWithLearnings: %v", err)
	}
}

func TestCov11_runTaskFeedback_dryRunWithProcessable(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Create tasks file with a completed task that has learning_id
	tasksDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	taskLine := `{"task_id":"t2","subject":"done task","status":"completed","learning_id":"learn-abc","session_id":"sess-2","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}` + "\n"
	if err := os.WriteFile(filepath.Join(tasksDir, "tasks.jsonl"), []byte(taskLine), 0644); err != nil {
		t.Fatalf("write tasks: %v", err)
	}

	origDryRun := dryRun
	origSessionID := taskFeedbackSessionID
	defer func() {
		dryRun = origDryRun
		taskFeedbackSessionID = origSessionID
	}()
	dryRun = true
	taskFeedbackSessionID = "" // no session filter

	cmd := &cobra.Command{}
	err := runTaskFeedback(cmd, nil)
	if err != nil {
		t.Fatalf("runTaskFeedback dryRun: %v", err)
	}
}

// ---------------------------------------------------------------------------
// feedback_loop.go — runFeedbackLoop no-citations path (32% → higher)
// ---------------------------------------------------------------------------

func TestCov11_runFeedbackLoop_noCitations(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origSessionID := feedbackLoopSessionID
	origDryRun := dryRun
	defer func() {
		feedbackLoopSessionID = origSessionID
		dryRun = origDryRun
	}()
	feedbackLoopSessionID = "sess-no-citations-xyz"
	dryRun = false

	cmd := &cobra.Command{}
	err := runFeedbackLoop(cmd, nil)
	// No citations file → empty citations → prints "No citations found" → nil
	if err != nil {
		t.Fatalf("runFeedbackLoop noCitations: %v", err)
	}
}

// ---------------------------------------------------------------------------
// pool_ingest.go — runPoolIngest no-new-files path (29.6% → higher)
// ---------------------------------------------------------------------------

func TestCov11_runPoolIngest_noNewFiles(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origPoolIngestDir := poolIngestDir
	origDryRun := dryRun
	defer func() {
		poolIngestDir = origPoolIngestDir
		dryRun = origDryRun
	}()
	// Point to empty dir — no *.md files will match
	poolIngestDir = filepath.Join(".agents", "knowledge", "pending")
	dryRun = false

	cmd := &cobra.Command{}
	// No args → resolveIngestFiles uses poolIngestDir globs → no files found
	err := runPoolIngest(cmd, nil)
	if err != nil {
		t.Fatalf("runPoolIngest noNewFiles: %v", err)
	}
}

// ---------------------------------------------------------------------------
// plans.go — runPlansDiff no-manifest path (26.9% → higher)
// ---------------------------------------------------------------------------

func TestCov11_runPlansDiff_noManifest(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cmd := &cobra.Command{}
	err := runPlansDiff(cmd, nil)
	// No manifest file → IsNotExist → prints "No manifest found." → nil
	if err != nil {
		t.Fatalf("runPlansDiff noManifest: %v", err)
	}
}
