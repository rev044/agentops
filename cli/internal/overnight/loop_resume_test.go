package overnight

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/overnight/fixture"
)

// TestRunLoop_CrashAtIter2_ResumeRehydrates exercises the Micro-epic 2
// persistence + rehydration path end-to-end:
//
//  1. Seed a corpus fixture.
//  2. Run RunLoop with fault injection to panic after iter 2 persists.
//  3. Confirm iter-1.json and iter-2.json exist on disk under
//     <outputDir>/<runID>/iterations/.
//  4. Run RunLoop again with the same RunID and confirm the returned
//     result.Iterations contains all 5 iterations in ascending index
//     order with RunID-prefixed IDs.
//
// The fixture is the shared GenerateFixture(DefaultOpts()) used by the
// existing L2 test. HOME is isolated so harvest.Promote does not touch
// the real hub.
func TestRunLoop_CrashAtIter2_ResumeRehydrates(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	restore := stubInjectRefresh(t)
	defer restore()

	dir := t.TempDir()
	if err := fixture.GenerateFixture(dir, fixture.DefaultOpts()); err != nil {
		t.Fatalf("GenerateFixture: %v", err)
	}

	outputDir := filepath.Join(dir, ".agents", "overnight", "crash-test")
	runID := "test-run-crash-1"
	iterDir := filepath.Join(outputDir, runID, "iterations")

	opts := RunLoopOptions{
		Cwd:            dir,
		OutputDir:      outputDir,
		RunID:          runID,
		RunTimeout:     30 * time.Second,
		MaxIterations:  5,
		PlateauEpsilon: 0.01,
		PlateauWindowK: 2,
		WarnOnly:       true,
		LogWriter:      io.Discard,
	}

	// Phase 1: run to iter 2, then "crash" via fault injection.
	SetFaultInjectionAfterIter(2)
	t.Cleanup(func() { SetFaultInjectionAfterIter(0) })

	func() {
		defer func() { _ = recover() }() // swallow the injected panic
		_, _ = RunLoop(context.Background(), opts)
	}()

	// After the crash, iter-1.json and iter-2.json must exist on disk.
	assertFileExists(t, filepath.Join(iterDir, "iter-1.json"))
	assertFileExists(t, filepath.Join(iterDir, "iter-2.json"))

	// Sanity: each persisted file has the correct RunID prefix in its ID.
	for i := 1; i <= 2; i++ {
		raw, err := os.ReadFile(filepath.Join(iterDir, fmt.Sprintf("iter-%d.json", i)))
		if err != nil {
			t.Fatalf("read iter-%d.json: %v", i, err)
		}
		var it IterationSummary
		if err := json.Unmarshal(raw, &it); err != nil {
			t.Fatalf("unmarshal iter-%d: %v", i, err)
		}
		wantID := fmt.Sprintf("%s-iter-%d", runID, i)
		if string(it.ID) != wantID {
			t.Errorf("persisted iter-%d.ID = %q, want %q", i, it.ID, wantID)
		}
	}

	// Phase 2: clear fault, resume. Expect 5 iterations total.
	SetFaultInjectionAfterIter(0)
	result, err := RunLoop(context.Background(), opts)
	if err != nil {
		t.Fatalf("resume RunLoop: %v", err)
	}
	if result == nil {
		t.Fatal("resume RunLoop returned nil result")
	}
	if len(result.Iterations) != 5 {
		t.Fatalf("expected 5 iterations after resume, got %d", len(result.Iterations))
	}
	for i, it := range result.Iterations {
		if it.Index != i+1 {
			t.Errorf("iter[%d].Index = %d, want %d", i, it.Index, i+1)
		}
		wantPrefix := runID + "-iter-"
		if !strings.HasPrefix(string(it.ID), wantPrefix) {
			t.Errorf("iter[%d].ID = %q, want prefix %q", i, it.ID, wantPrefix)
		}
	}
}

// TestRunLoop_IterationID_UsesRunIDPrefix is the regression guard for
// the IterationID contract-drift fix (types.go line 9 documents
// "<run-id>-iter-<N>" format but loop.go was generating "iter-N"
// without the run-id prefix).
func TestRunLoop_IterationID_UsesRunIDPrefix(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	restore := stubInjectRefresh(t)
	defer restore()

	dir := t.TempDir()
	if err := fixture.GenerateFixture(dir, fixture.DefaultOpts()); err != nil {
		t.Fatalf("GenerateFixture: %v", err)
	}

	opts := RunLoopOptions{
		Cwd:            dir,
		OutputDir:      filepath.Join(dir, ".agents", "overnight", "id-test"),
		RunID:          "contract-id-run",
		RunTimeout:     30 * time.Second,
		MaxIterations:  2,
		PlateauEpsilon: 0.01,
		PlateauWindowK: 2,
		WarnOnly:       true,
		LogWriter:      io.Discard,
	}
	result, err := RunLoop(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunLoop: %v", err)
	}
	if result == nil || len(result.Iterations) == 0 {
		t.Fatal("expected at least 1 iteration")
	}
	for _, it := range result.Iterations {
		want := fmt.Sprintf("%s-iter-%d", opts.RunID, it.Index)
		if string(it.ID) != want {
			t.Errorf("iter.ID = %q, want %q", it.ID, want)
		}
	}
}

// TestRunLoop_RequiresRunID asserts the normalize/validation path
// returns an error when RunID is empty.
func TestRunLoop_RequiresRunID(t *testing.T) {
	dir := t.TempDir()
	opts := RunLoopOptions{
		Cwd:       dir,
		OutputDir: filepath.Join(dir, "out"),
		// RunID deliberately omitted.
		RunTimeout:    5 * time.Second,
		MaxIterations: 1,
		LogWriter:     io.Discard,
	}
	_, err := RunLoop(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error when RunID is empty, got nil")
	}
	if !strings.Contains(err.Error(), "RunID") {
		t.Errorf("error should mention RunID, got %v", err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s to exist: %v", path, err)
	}
}
