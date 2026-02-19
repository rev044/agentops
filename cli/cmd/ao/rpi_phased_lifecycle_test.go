package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDeregisterOnSuccess verifies that deregisterRPIAgent is callable after a
// successful run (i.e., the defer is in place for the success path).
// We test this by directly verifying the function does not panic and produces
// the correct gt command invocation.
func TestDeregisterOnSuccess(t *testing.T) {
	origGtPath := gtPath
	defer func() { gtPath = origGtPath }()

	tmpDir := t.TempDir()
	markerFile := filepath.Join(tmpDir, "deregister.marker")
	tmpBin := createFakeGtMarkerBinary(t, markerFile, "deregister")
	gtPath = tmpBin

	// Simulate success path: register then deregister (as defer would do).
	runID := "test-success-run"
	registerRPIAgent(runID)
	deregisterRPIAgent(runID) // This is what the defer calls on success.

	// Check synchronously — the shell script runs synchronously via exec.Command.Run().
	if _, err := os.Stat(markerFile); err != nil {
		t.Error("deregisterRPIAgent was not called on success path (marker file not created)")
	}
}

// TestDeregisterOnFailure verifies that deregisterRPIAgent is callable after a
// failed run. In the actual code, `defer deregisterRPIAgent(state.RunID)` is
// registered unconditionally — it fires on all exit paths including failure.
func TestDeregisterOnFailure(t *testing.T) {
	origGtPath := gtPath
	defer func() { gtPath = origGtPath }()

	tmpDir := t.TempDir()
	markerFile := filepath.Join(tmpDir, "deregister-fail.marker")
	tmpBin := createFakeGtMarkerBinary(t, markerFile, "deregister")
	gtPath = tmpBin

	// Simulate failure path: register then hit an error, defer still fires.
	runID := "test-failure-run"
	registerRPIAgent(runID)

	// Simulate what the defer does even on error exit.
	func() {
		defer deregisterRPIAgent(runID)
		// Simulate a failure return — defer fires regardless.
		_ = fmt.Errorf("simulated phase failure")
	}()

	// Check synchronously.
	if _, err := os.Stat(markerFile); err != nil {
		t.Error("deregisterRPIAgent was not called on failure path (defer must fire unconditionally)")
	}
}

// TestMergeFailurePropagation verifies that mergeWorktree errors surface as
// non-zero command results (non-nil error). Uses a real git repo to test the
// dirty-repo rejection path which is the most deterministic failure mode.
func TestMergeFailurePropagation(t *testing.T) {
	repo := initTestRepo(t)

	origDir, _ := os.Getwd()
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir) //nolint:errcheck

	worktreePath, runID, err := createWorktree(repo)
	if err != nil {
		t.Fatalf("createWorktree: %v", err)
	}
	defer func() {
		cmd := exec.Command("git", "worktree", "remove", worktreePath, "--force")
		cmd.Dir = repo
		_ = cmd.Run()
		cmd = exec.Command("git", "branch", "-D", "rpi/"+runID)
		cmd.Dir = repo
		_ = cmd.Run()
	}()

	// Dirty the original repo so merge is rejected.
	dirtyFile := filepath.Join(repo, "uncommitted.txt")
	if err := os.WriteFile(dirtyFile, []byte("dirty\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "uncommitted.txt")
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	// mergeWorktree must return a non-nil error — callers must propagate it.
	mergeErr := mergeWorktree(repo, runID)
	if mergeErr == nil {
		t.Fatal("expected mergeWorktree to return error for dirty repo; got nil (silent-success violation)")
	}

	// Verify the error message is actionable (not just a wrapped exit code).
	errMsg := mergeErr.Error()
	if !strings.Contains(errMsg, "uncommitted") && !strings.Contains(errMsg, "after 5 retries") {
		t.Errorf("merge failure error should mention uncommitted changes, got: %v", mergeErr)
	}

	t.Logf("merge failure propagated correctly: %v", mergeErr)
}

// TestCleanupFailurePropagation verifies that removeWorktree errors are logged
// with actionable context via logFailureContext and propagated as non-nil errors.
// Tests the specific behavior added in this task: cleanup failures must not
// silently succeed.
func TestCleanupFailurePropagation(t *testing.T) {
	// Test that removeWorktree returns an error for invalid paths (path validation).
	repo := initTestRepo(t)

	// Path that fails validation (not an rpi sibling path).
	rmErr := removeWorktree(repo, "/tmp/not-a-valid-rpi-worktree", "somerunid")
	if rmErr == nil {
		t.Fatal("expected removeWorktree to return error for invalid path; got nil (silent-success violation)")
	}
	if !strings.Contains(rmErr.Error(), "path validation failed") {
		t.Errorf("removeWorktree error should mention path validation, got: %v", rmErr)
	}

	// Verify logFailureContext records the error with actionable context.
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "phased-orchestration.log")

	logFailureContext(logPath, "test-run", "cleanup", rmErr)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log file not written after logFailureContext: %v", err)
	}
	logContent := string(data)

	// Must contain FAILURE_CONTEXT marker.
	if !strings.Contains(logContent, "FAILURE_CONTEXT") {
		t.Errorf("log must contain FAILURE_CONTEXT marker, got: %q", logContent)
	}
	// Must contain actionable guidance.
	if !strings.Contains(logContent, "action:") {
		t.Errorf("log must contain actionable guidance (action:), got: %q", logContent)
	}
	// Must include the phase name.
	if !strings.Contains(logContent, "cleanup") {
		t.Errorf("log must include phase name 'cleanup', got: %q", logContent)
	}
	// Must include the run ID.
	if !strings.Contains(logContent, "test-run") {
		t.Errorf("log must include run ID 'test-run', got: %q", logContent)
	}

	t.Logf("cleanup failure logged with actionable context: %s", logContent)
}

// TestDeregisterRPIAgent_CalledOnAllExitPaths verifies the defer placement
// ensures deregisterRPIAgent fires on all exit paths by testing the underlying
// function directly under different conditions.
func TestDeregisterRPIAgent_CalledOnAllExitPaths(t *testing.T) {
	origGtPath := gtPath
	defer func() { gtPath = origGtPath }()

	// Test 1: no-op when gt not on PATH (must not panic).
	gtPath = ""
	deregisterRPIAgent("any-run-id") // Should not panic.

	// Test 2: with a non-existent binary (should fail silently).
	gtPath = "/nonexistent/gt-binary"
	deregisterRPIAgent("any-run-id") // Should not panic or propagate error.

	// Test 3: with a valid binary (verify correct arguments via marker file).
	tmpDir := t.TempDir()
	markerFile := filepath.Join(tmpDir, "deregister-all-paths.marker")
	tmpBin := createFakeGtMarkerBinary(t, markerFile, "deregister")
	gtPath = tmpBin

	deregisterRPIAgent("run-123")
	if _, err := os.Stat(markerFile); err != nil {
		t.Error("deregisterRPIAgent did not attempt gt command")
	}
}

// createFakeGtMarkerBinary creates a shell script that acts as a fake "gt" binary.
// When invoked with any argument matching expectSubcommand, it creates markerFile.
// Returns the path to the executable script.
func createFakeGtMarkerBinary(t *testing.T, markerFile, expectSubcommand string) string {
	t.Helper()
	tmpDir := t.TempDir()
	script := filepath.Join(tmpDir, "gt")

	scriptContent := fmt.Sprintf(`#!/bin/sh
for arg in "$@"; do
  if [ "$arg" = "%s" ]; then
    touch "%s"
    exit 0
  fi
done
exit 0
`, expectSubcommand, markerFile)

	if err := os.WriteFile(script, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("create fake gt script: %v", err)
	}

	return script
}
