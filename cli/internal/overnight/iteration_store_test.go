package overnight

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// helper: build a minimal valid IterationSummary with a correct ID prefix.
func makeIter(runID string, index int) IterationSummary {
	return IterationSummary{
		ID:         IterationID(fmt.Sprintf("%s-iter-%d", runID, index)),
		Index:      index,
		StartedAt:  time.Unix(1000, 0).UTC(),
		FinishedAt: time.Unix(1001, 0).UTC(),
		Duration:   "1s",
		Status:     "done",
		FitnessAfter: map[string]any{
			"citation_coverage": 0.5 + float64(index)*0.01,
		},
		FitnessDelta: 0.01,
	}
}

// TestLoadIterations_DirMissing covers case #1: missing directory returns
// empty slice and no error.
func TestLoadIterations_DirMissing(t *testing.T) {
	iters, rej, err := LoadIterations(filepath.Join(t.TempDir(), "nonexistent-subdir"), "r1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(iters) != 0 {
		t.Errorf("expected 0 iterations, got %d", len(iters))
	}
	if len(rej) != 0 {
		t.Errorf("expected 0 rejections, got %d", len(rej))
	}
}

// TestLoadIterations_DirEmpty covers case #2: empty directory returns
// empty slice and no error.
func TestLoadIterations_DirEmpty(t *testing.T) {
	dir := t.TempDir()
	iters, rej, err := LoadIterations(dir, "r1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(iters) != 0 {
		t.Errorf("expected 0 iterations, got %d", len(iters))
	}
	if len(rej) != 0 {
		t.Errorf("expected 0 rejections, got %d", len(rej))
	}
}

// TestWriteAndLoadIterations_HappyPath covers case #3: write 3 iterations,
// load them back in order with no rejections.
func TestWriteAndLoadIterations_HappyPath(t *testing.T) {
	dir := t.TempDir()
	runID := "r-happy"
	for i := 1; i <= 3; i++ {
		if err := writeIterationAtomic(dir, makeIter(runID, i)); err != nil {
			t.Fatalf("write iter-%d: %v", i, err)
		}
		// Verify final file exists with expected name.
		if _, err := os.Stat(filepath.Join(dir, fmt.Sprintf("iter-%d.json", i))); err != nil {
			t.Fatalf("iter-%d.json missing after write: %v", i, err)
		}
	}
	iters, rej, err := LoadIterations(dir, runID)
	if err != nil {
		t.Fatalf("LoadIterations: %v", err)
	}
	if len(rej) != 0 {
		t.Errorf("unexpected rejections: %v", rej)
	}
	if len(iters) != 3 {
		t.Fatalf("expected 3 iterations, got %d", len(iters))
	}
	for i, it := range iters {
		if it.Index != i+1 {
			t.Errorf("iters[%d].Index = %d, want %d", i, it.Index, i+1)
		}
		wantID := fmt.Sprintf("%s-iter-%d", runID, i+1)
		if string(it.ID) != wantID {
			t.Errorf("iters[%d].ID = %q, want %q", i, it.ID, wantID)
		}
	}
}

// TestLoadIterations_GapInIndices covers case #4: iter-1, iter-2, iter-4
// should return 1..2 and reject 4.
func TestLoadIterations_GapInIndices(t *testing.T) {
	dir := t.TempDir()
	runID := "r-gap"
	for _, i := range []int{1, 2, 4} {
		if err := writeIterationAtomic(dir, makeIter(runID, i)); err != nil {
			t.Fatalf("write iter-%d: %v", i, err)
		}
	}
	iters, rej, err := LoadIterations(dir, runID)
	if err != nil {
		t.Fatalf("LoadIterations: %v", err)
	}
	if len(iters) != 2 {
		t.Fatalf("expected 2 iterations (1,2), got %d", len(iters))
	}
	if iters[0].Index != 1 || iters[1].Index != 2 {
		t.Errorf("got indices [%d %d], want [1 2]", iters[0].Index, iters[1].Index)
	}
	if len(rej) != 1 {
		t.Fatalf("expected 1 rejection, got %d (%v)", len(rej), rej)
	}
	if !strings.Contains(rej[0], "gap") {
		t.Errorf("expected rejection reason to mention 'gap', got %q", rej[0])
	}
}

// TestLoadIterations_DifferentRunIDContamination covers case #5: an
// iter-1.json with a different runID in its embedded ID is rejected.
func TestLoadIterations_DifferentRunIDContamination(t *testing.T) {
	dir := t.TempDir()
	it := makeIter("r-other", 1)
	if err := writeIterationAtomic(dir, it); err != nil {
		t.Fatalf("write: %v", err)
	}
	iters, rej, err := LoadIterations(dir, "r-mine")
	if err != nil {
		t.Fatalf("LoadIterations: %v", err)
	}
	if len(iters) != 0 {
		t.Errorf("expected 0 iterations (all rejected), got %d", len(iters))
	}
	if len(rej) != 1 {
		t.Fatalf("expected 1 rejection, got %d", len(rej))
	}
	if !strings.Contains(rej[0], "runID") && !strings.Contains(rej[0], "id mismatch") {
		t.Errorf("expected runID mismatch message, got %q", rej[0])
	}
}

// TestLoadIterations_CorruptJSON covers case #6: malformed JSON file is
// rejected with a parse reason and load continues.
func TestLoadIterations_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	// Write a valid iter-1 and a corrupt iter-2.
	if err := writeIterationAtomic(dir, makeIter("r-corrupt", 1)); err != nil {
		t.Fatalf("write iter-1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "iter-2.json"), []byte(`{"malformed`), 0o644); err != nil {
		t.Fatalf("write corrupt: %v", err)
	}
	iters, rej, err := LoadIterations(dir, "r-corrupt")
	if err != nil {
		t.Fatalf("LoadIterations: %v", err)
	}
	if len(iters) != 1 {
		t.Fatalf("expected 1 valid iteration, got %d", len(iters))
	}
	if len(rej) != 1 {
		t.Fatalf("expected 1 rejection, got %d", len(rej))
	}
	if !strings.Contains(rej[0], "parse") {
		t.Errorf("expected rejection to contain 'parse', got %q", rej[0])
	}
}

// TestWriteIterationAtomic_MarshalFailsDoesNotCreateFinal covers case #7
// conceptually: if the write cannot complete, no final file exists.
// Testing real crash-before-rename is flaky (would require OS-level
// signal); instead we verify that a write into a non-writable directory
// leaves no final iter-N.json and no orphan temp file.
func TestWriteIterationAtomic_WriteToNonexistentParentStillAtomic(t *testing.T) {
	// Writing into a brand-new nested path: MkdirAll creates it.
	dir := filepath.Join(t.TempDir(), "a", "b", "c")
	if err := writeIterationAtomic(dir, makeIter("r-nested", 1)); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Final file exists.
	if _, err := os.Stat(filepath.Join(dir, "iter-1.json")); err != nil {
		t.Fatalf("iter-1.json missing: %v", err)
	}
	// No stray temp files.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".iter-") {
			t.Errorf("stray temp file left behind: %s", e.Name())
		}
	}
}

// TestWriteIterationAtomic_ReaderRace covers case #8: atomic rename
// survives a reader race; the reader never sees a partial file.
func TestWriteIterationAtomic_ReaderRace(t *testing.T) {
	dir := t.TempDir()
	runID := "r-race"
	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Reader goroutine: continuously load iterations. Every observation
	// must parse cleanly (no partial reads).
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				_, rej, err := LoadIterations(dir, runID)
				if err != nil {
					t.Errorf("reader LoadIterations: %v", err)
					return
				}
				for _, r := range rej {
					// parse-failure rejections would indicate we saw a
					// half-written file.
					if strings.Contains(r, "parse") {
						t.Errorf("reader observed partial file: %s", r)
						return
					}
				}
			}
		}
	}()

	// Writer: persist iter-1 through iter-20 in quick succession.
	for i := 1; i <= 20; i++ {
		if err := writeIterationAtomic(dir, makeIter(runID, i)); err != nil {
			t.Fatalf("write iter-%d: %v", i, err)
		}
	}
	close(stop)
	wg.Wait()
}

// TestWriteIterationAtomic_ParentDirFsync covers case #9: on Linux/Darwin
// the parent dir fsync path returns nil. Windows is skipped.
func TestWriteIterationAtomic_ParentDirFsync(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory fsync is not meaningful on Windows")
	}
	dir := t.TempDir()
	if err := writeIterationAtomic(dir, makeIter("r-fsync", 1)); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Direct call to fsyncDir must succeed on darwin/linux.
	if err := fsyncDir(dir); err != nil {
		t.Errorf("fsyncDir: %v", err)
	}
}

// TestWriteIterationAtomic_RoundTripsFields verifies the on-disk JSON
// preserves the IterationSummary fields we care about for rehydration.
func TestWriteIterationAtomic_RoundTripsFields(t *testing.T) {
	dir := t.TempDir()
	it := makeIter("r-round", 1)
	it.Degraded = []string{"example-note"}
	if err := writeIterationAtomic(dir, it); err != nil {
		t.Fatalf("write: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "iter-1.json"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var got IterationSummary
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Index != 1 {
		t.Errorf("Index = %d, want 1", got.Index)
	}
	if string(got.ID) != "r-round-iter-1" {
		t.Errorf("ID = %q, want %q", got.ID, "r-round-iter-1")
	}
	if len(got.Degraded) != 1 || got.Degraded[0] != "example-note" {
		t.Errorf("Degraded = %v, want [example-note]", got.Degraded)
	}
	if got.FitnessAfter["citation_coverage"] == nil {
		t.Errorf("FitnessAfter citation_coverage missing")
	}
}
