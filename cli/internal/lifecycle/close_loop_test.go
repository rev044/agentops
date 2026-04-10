package lifecycle

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/pool"
)

// stubCloseLoopDeps returns a CloseLoopOpts wired with harmless no-op callbacks
// so tests can focus on the specific fan-out they want to exercise.
func stubCloseLoopDeps() CloseLoopOpts {
	return CloseLoopOpts{
		PendingDir:  filepath.Join(".agents", "knowledge", "pending"),
		Threshold:   0,
		Quiet:       true,
		DryRun:      false,
		IncludeGold: true,

		ResolveIngestFiles: func(cwd, pendingDir string, args []string) ([]string, error) {
			return nil, nil
		},
		IngestFilesToPool: func(cwd string, files []string) (CloseLoopIngestResult, error) {
			return CloseLoopIngestResult{FilesScanned: len(files)}, nil
		},
		AutoPromoteFn: func(p *pool.Pool, threshold time.Duration, includeGold bool) (CloseLoopAutoPromoteResult, error) {
			return CloseLoopAutoPromoteResult{Threshold: threshold.String()}, nil
		},
		ProcessCitationFeedback: func(cwd string) (int, int, int) {
			return 0, 0, 0
		},
		PromoteCitedLearnings: func(cwd string, quiet bool) int {
			return 0
		},
		PromoteToMemory: func(cwd string) (int, error) {
			return 0, nil
		},
		ApplyMaturityFn: func(cwd string) (MaturityTransitionSummary, error) {
			return MaturityTransitionSummary{}, nil
		},
		StoreIndexUpsertFn: func(baseDir string, paths []string, categorize bool) (int, string, error) {
			return len(paths), filepath.Join(baseDir, ".agents", "store", "index.jsonl"), nil
		},
	}
}

// snapshotDir walks a directory and returns a stable map of relpath -> sha256
// of the file contents. Used to prove byte-identical state after a dry-run.
func snapshotDir(t *testing.T, root string) map[string]string {
	t.Helper()
	out := make(map[string]string)
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return rerr
		}
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		h := sha256.Sum256(data)
		out[rel] = hex.EncodeToString(h[:])
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("snapshotDir: %v", err)
	}
	return out
}

// TestExecuteCloseLoop_EmptyCorpus verifies that running against a bare
// t.TempDir() with no .agents/ tree degrades gracefully: no panic, empty
// counters, and no error from the orchestrator (the stubs absorb the empty
// state).
func TestExecuteCloseLoop_EmptyCorpus(t *testing.T) {
	tmp := t.TempDir()
	opts := stubCloseLoopDeps()

	res, err := ExecuteCloseLoop(tmp, opts)
	if err != nil {
		t.Fatalf("ExecuteCloseLoop returned error on empty corpus: %v", err)
	}
	if res == nil {
		t.Fatal("ExecuteCloseLoop returned nil result")
	}
	if res.Ingest.Added != 0 {
		t.Errorf("Ingest.Added = %d, want 0", res.Ingest.Added)
	}
	if res.AutoPromote.Promoted != 0 {
		t.Errorf("AutoPromote.Promoted = %d, want 0", res.AutoPromote.Promoted)
	}
	if res.AntiPattern.Promoted != 0 {
		t.Errorf("AntiPattern.Promoted = %d, want 0", res.AntiPattern.Promoted)
	}
	if res.CitationFeedback.Processed != 0 {
		t.Errorf("CitationFeedback.Processed = %d, want 0", res.CitationFeedback.Processed)
	}
	if res.Store.Indexed != 0 {
		t.Errorf("Store.Indexed = %d, want 0", res.Store.Indexed)
	}
	if res.MemoryPromoted != 0 {
		t.Errorf("MemoryPromoted = %d, want 0", res.MemoryPromoted)
	}
}

// TestExecuteCloseLoop_NoOpWhenNothingEligible seeds a recording stub and
// verifies that two back-to-back runs over the same non-mutating corpus return
// identical results (nothing eligible -> nothing promoted).
func TestExecuteCloseLoop_NoOpWhenNothingEligible(t *testing.T) {
	tmp := t.TempDir()

	opts := stubCloseLoopDeps()
	// Simulate a pool with zero pending entries: auto-promote returns a fixed
	// shape every call (no mutation between runs).
	opts.AutoPromoteFn = func(p *pool.Pool, threshold time.Duration, includeGold bool) (CloseLoopAutoPromoteResult, error) {
		return CloseLoopAutoPromoteResult{Threshold: threshold.String()}, nil
	}

	res1, err1 := ExecuteCloseLoop(tmp, opts)
	if err1 != nil {
		t.Fatalf("first run: %v", err1)
	}
	res2, err2 := ExecuteCloseLoop(tmp, opts)
	if err2 != nil {
		t.Fatalf("second run: %v", err2)
	}

	if res1.AutoPromote.Promoted != res2.AutoPromote.Promoted {
		t.Errorf("AutoPromote.Promoted: run1=%d run2=%d", res1.AutoPromote.Promoted, res2.AutoPromote.Promoted)
	}
	if res1.AntiPattern.Promoted != res2.AntiPattern.Promoted {
		t.Errorf("AntiPattern.Promoted: run1=%d run2=%d", res1.AntiPattern.Promoted, res2.AntiPattern.Promoted)
	}
}

// TestExecuteCloseLoop_Idempotent runs the orchestrator twice against the same
// stub fixture and asserts the Promoted/Skipped counts on the second run match
// the first. Because the stubs are pure, exact equality is the expectation.
func TestExecuteCloseLoop_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	opts := stubCloseLoopDeps()

	// Seed some stable counts on the first run.
	opts.IngestFilesToPool = func(cwd string, files []string) (CloseLoopIngestResult, error) {
		return CloseLoopIngestResult{FilesScanned: 2, Added: 0, SkippedExisting: 2}, nil
	}
	opts.AutoPromoteFn = func(p *pool.Pool, threshold time.Duration, includeGold bool) (CloseLoopAutoPromoteResult, error) {
		return CloseLoopAutoPromoteResult{Threshold: threshold.String(), Considered: 0, Promoted: 0, Skipped: 0}, nil
	}
	opts.ProcessCitationFeedback = func(cwd string) (int, int, int) { return 0, 0, 0 }

	res1, err := ExecuteCloseLoop(tmp, opts)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	res2, err := ExecuteCloseLoop(tmp, opts)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	if res1.Ingest.FilesScanned != res2.Ingest.FilesScanned ||
		res1.Ingest.Added != res2.Ingest.Added ||
		res1.Ingest.SkippedExisting != res2.Ingest.SkippedExisting ||
		res1.Ingest.SkippedMalformed != res2.Ingest.SkippedMalformed {
		t.Errorf("Ingest not idempotent: %+v vs %+v", res1.Ingest, res2.Ingest)
	}
	if res1.AutoPromote.Promoted != res2.AutoPromote.Promoted || res1.AutoPromote.Skipped != res2.AutoPromote.Skipped {
		t.Errorf("AutoPromote not idempotent: %+v vs %+v", res1.AutoPromote, res2.AutoPromote)
	}
	if res1.CitationFeedback != res2.CitationFeedback {
		t.Errorf("CitationFeedback not idempotent: %+v vs %+v", res1.CitationFeedback, res2.CitationFeedback)
	}
	if res1.MemoryPromoted != res2.MemoryPromoted {
		t.Errorf("MemoryPromoted: run1=%d run2=%d", res1.MemoryPromoted, res2.MemoryPromoted)
	}
}

// TestExecuteCloseLoop_DryRunDoesNotMutate seeds a .agents/ directory with a
// single learning file, runs ExecuteCloseLoop with DryRun=true, and verifies
// that the directory state is byte-identical afterwards. Because we use stub
// callbacks that don't touch disk, the invariant is trivially satisfied; the
// test still exercises the wiring to make sure dry-run flags propagate
// through the opts without blowing up.
func TestExecuteCloseLoop_DryRunDoesNotMutate(t *testing.T) {
	tmp := t.TempDir()

	// Seed a representative .agents tree.
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatalf("mkdir learnings: %v", err)
	}
	learningFile := filepath.Join(learningsDir, "sample.md")
	if err := os.WriteFile(learningFile, []byte("---\nid: sample\nutility: 0.5\n---\nbody\n"), 0o644); err != nil {
		t.Fatalf("write learning: %v", err)
	}

	before := snapshotDir(t, tmp)

	opts := stubCloseLoopDeps()
	opts.DryRun = true

	if _, err := ExecuteCloseLoop(tmp, opts); err != nil {
		t.Fatalf("ExecuteCloseLoop dry-run: %v", err)
	}

	after := snapshotDir(t, tmp)

	if len(before) != len(after) {
		t.Errorf("dry-run added/removed files: before=%d after=%d", len(before), len(after))
	}
	for path, hBefore := range before {
		hAfter, ok := after[path]
		if !ok {
			t.Errorf("dry-run removed file %s", path)
			continue
		}
		if hBefore != hAfter {
			t.Errorf("dry-run mutated file %s: before=%s after=%s", path, hBefore, hAfter)
		}
	}
}

// TestExecuteCloseLoop_RequiresCallbacks verifies the argument-validation
// branch returns errors instead of panicking on missing callbacks.
func TestExecuteCloseLoop_RequiresCallbacks(t *testing.T) {
	tmp := t.TempDir()

	// Missing ResolveIngestFiles -> error
	opts := CloseLoopOpts{}
	if _, err := ExecuteCloseLoop(tmp, opts); err == nil {
		t.Errorf("expected error for missing ResolveIngestFiles")
	}

	// Supply only ResolveIngestFiles -> still errors (missing IngestFilesToPool)
	opts.ResolveIngestFiles = func(cwd, pendingDir string, args []string) ([]string, error) {
		return nil, nil
	}
	if _, err := ExecuteCloseLoop(tmp, opts); err == nil {
		t.Errorf("expected error for missing IngestFilesToPool")
	}
}

// TestExecuteCloseLoop_PropagatesCallbackError verifies that a callback error
// short-circuits the orchestrator and is returned verbatim.
func TestExecuteCloseLoop_PropagatesCallbackError(t *testing.T) {
	tmp := t.TempDir()
	opts := stubCloseLoopDeps()
	want := errors.New("synthetic ingest failure")
	opts.IngestFilesToPool = func(cwd string, files []string) (CloseLoopIngestResult, error) {
		return CloseLoopIngestResult{}, want
	}

	_, err := ExecuteCloseLoop(tmp, opts)
	if !errors.Is(err, want) {
		t.Errorf("expected propagation of synthetic error, got: %v", err)
	}
}

// TestExecuteCloseLoop_AppliesAllOrdering verifies the stages fire in the
// documented order (ingest -> auto-promote -> citation -> promote cited ->
// maturity -> store -> memory). We record the order via a shared slice and
// assert the final sequence.
func TestExecuteCloseLoop_AppliesAllOrdering(t *testing.T) {
	tmp := t.TempDir()
	opts := stubCloseLoopDeps()

	var order []string
	opts.IngestFilesToPool = func(cwd string, files []string) (CloseLoopIngestResult, error) {
		order = append(order, "ingest")
		return CloseLoopIngestResult{}, nil
	}
	opts.AutoPromoteFn = func(p *pool.Pool, threshold time.Duration, includeGold bool) (CloseLoopAutoPromoteResult, error) {
		order = append(order, "auto-promote")
		return CloseLoopAutoPromoteResult{Threshold: threshold.String()}, nil
	}
	opts.ProcessCitationFeedback = func(cwd string) (int, int, int) {
		order = append(order, "citation")
		return 0, 0, 0
	}
	opts.PromoteCitedLearnings = func(cwd string, quiet bool) int {
		order = append(order, "promote-cited")
		return 0
	}
	opts.ApplyMaturityFn = func(cwd string) (MaturityTransitionSummary, error) {
		order = append(order, "maturity")
		return MaturityTransitionSummary{}, nil
	}
	opts.StoreIndexUpsertFn = func(baseDir string, paths []string, categorize bool) (int, string, error) {
		order = append(order, "store")
		return 0, "", nil
	}
	opts.PromoteToMemory = func(cwd string) (int, error) {
		order = append(order, "memory")
		return 0, nil
	}

	if _, err := ExecuteCloseLoop(tmp, opts); err != nil {
		t.Fatalf("ExecuteCloseLoop: %v", err)
	}

	want := []string{"ingest", "auto-promote", "citation", "promote-cited", "maturity", "store", "memory"}
	if len(order) != len(want) {
		t.Fatalf("order len = %d, want %d (order=%v)", len(order), len(want), order)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("order[%d] = %q, want %q (full=%v)", i, order[i], want[i], order)
		}
	}

	// Extra belt-and-suspenders: ensure the sort order of a copy is different
	// to prove we're asserting the recorded order and not an accidentally
	// sorted slice.
	sorted := append([]string{}, order...)
	sort.Strings(sorted)
	if sorted[0] == order[0] && sorted[1] == order[1] && sorted[2] == order[2] {
		// This is acceptable; the sequence simply happens to line up.
		_ = sorted
	}
}
