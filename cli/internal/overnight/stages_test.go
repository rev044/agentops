package overnight

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/lifecycle"
	"github.com/boshu2/agentops/cli/internal/pool"
)

// writeLearning writes a minimal learning markdown file under
// <root>/.agents/learnings/<name>.md with the given frontmatter fields
// and body. All tests use this helper so the file shape stays in one
// place.
func writeLearning(t *testing.T, root, name string, fm map[string]string, body string) string {
	t.Helper()
	dir := filepath.Join(root, ".agents", "learnings")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir learnings: %v", err)
	}
	var b bytes.Buffer
	b.WriteString("---\n")
	for k, v := range fm {
		fmt.Fprintf(&b, "%s: %q\n", k, v)
	}
	b.WriteString("---\n\n")
	b.WriteString(body)
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		t.Fatalf("write learning %s: %v", name, err)
	}
	return path
}

// writeFindingRaw writes a canonical finding markdown file under
// <root>/.agents/findings/<name>.md with exact body contents. Used by
// MEASURE tests that seed unresolved findings. Distinct from
// findings_router_test.go's writeFinding helper which takes a
// title/summary pair.
func writeFindingRaw(t *testing.T, root, name, body string) string {
	t.Helper()
	dir := filepath.Join(root, ".agents", "findings")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir findings: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write finding %s: %v", name, err)
	}
	return path
}

// newTestOpts returns a RunLoopOptions scoped to a tempdir with sane
// zero values so each stage test can mutate the fields it cares about.
func newTestOpts(cwd string) RunLoopOptions {
	return RunLoopOptions{
		Cwd:       cwd,
		OutputDir: filepath.Join(cwd, ".agents", "overnight", "test-run"),
	}
}

// --- INGEST ------------------------------------------------------------

func TestRunIngest_HappyPath_HarvestCatalogBuilt(t *testing.T) {
	cwd := t.TempDir()
	writeLearning(t, cwd, "2026-04-09-first.md", map[string]string{
		"type":       "learning",
		"maturity":   "provisional",
		"confidence": "0.8",
	}, "# First\n\nBody text about the first learning.\n")
	writeLearning(t, cwd, "2026-04-09-second.md", map[string]string{
		"type":       "learning",
		"maturity":   "accepted",
		"confidence": "0.9",
	}, "# Second\n\nBody text about the second learning.\n")

	var logBuf bytes.Buffer
	res, err := RunIngest(context.Background(), newTestOpts(cwd), &logBuf)
	if err != nil {
		t.Fatalf("RunIngest error: %v", err)
	}
	if res.HarvestCatalog == nil {
		t.Fatal("expected HarvestCatalog to be non-nil")
	}
	if res.HarvestCatalog.Summary.ArtifactsExtracted < 2 {
		t.Errorf("expected >=2 artifacts, got %d", res.HarvestCatalog.Summary.ArtifactsExtracted)
	}
	if res.Duration <= 0 {
		t.Error("expected positive Duration")
	}
}

func TestRunIngest_HarvestDryRunDoesNotMutate(t *testing.T) {
	cwd := t.TempDir()
	writeLearning(t, cwd, "2026-04-09-static.md", map[string]string{
		"type":       "learning",
		"confidence": "0.9",
	}, "# Static\n\nUntouched body.\n")

	learningsDir := filepath.Join(cwd, ".agents", "learnings")
	before, err := os.ReadDir(learningsDir)
	if err != nil {
		t.Fatalf("read learnings before: %v", err)
	}

	if _, err := RunIngest(context.Background(), newTestOpts(cwd), io.Discard); err != nil {
		t.Fatalf("RunIngest error: %v", err)
	}

	after, err := os.ReadDir(learningsDir)
	if err != nil {
		t.Fatalf("read learnings after: %v", err)
	}
	if len(before) != len(after) {
		t.Errorf("learnings mutated: before=%d after=%d", len(before), len(after))
	}
}

func TestRunIngest_EmptyCorpus(t *testing.T) {
	// Isolate HOME to the tempdir so harvest.DiscoverRigs does not pick
	// up the operator's real ~/.agents/ global hub. DiscoverRigs adds
	// ~/.agents unconditionally; scoping HOME is the only portable way
	// to keep the walker from crossing the corpus we're trying to
	// measure as empty.
	home := t.TempDir()
	t.Setenv("HOME", home)

	cwd := t.TempDir()
	// Create .agents/ but no learnings/patterns/research.
	if err := os.MkdirAll(filepath.Join(cwd, ".agents"), 0o755); err != nil {
		t.Fatalf("mkdir .agents: %v", err)
	}

	res, err := RunIngest(context.Background(), newTestOpts(cwd), io.Discard)
	if err != nil {
		t.Fatalf("RunIngest error: %v", err)
	}
	if res.HarvestCatalog == nil {
		t.Fatal("expected HarvestCatalog to be non-nil even for empty corpus")
	}
	if res.HarvestCatalog.Summary.ArtifactsExtracted != 0 {
		t.Errorf("expected 0 artifacts, got %d", res.HarvestCatalog.Summary.ArtifactsExtracted)
	}
	if !containsSubstring(res.Degraded, "empty corpus") {
		t.Errorf("expected empty-corpus degraded note, got %v", res.Degraded)
	}
}

func TestRunIngest_ForgeProvenanceMineAreDegraded(t *testing.T) {
	cwd := t.TempDir()
	if err := os.MkdirAll(filepath.Join(cwd, ".agents"), 0o755); err != nil {
		t.Fatalf("mkdir .agents: %v", err)
	}
	res, err := RunIngest(context.Background(), newTestOpts(cwd), io.Discard)
	if err != nil {
		t.Fatalf("RunIngest error: %v", err)
	}
	for _, need := range []string{"forge-mine", "provenance-audit", "mine-findings"} {
		if !containsSubstring(res.Degraded, need) {
			t.Errorf("expected degraded note containing %q, got %v", need, res.Degraded)
		}
	}
}

// --- REDUCE ------------------------------------------------------------

// recordingCallbacks builds a CloseLoopOpts whose callbacks append their
// name to the shared order slice, so stage-order tests can assert the
// contract in one place.
type recorder struct {
	order []string
	// closeLoopErr, when non-nil, is returned from the first callback
	// that ExecuteCloseLoop invokes, forcing the stage to fail.
	closeLoopErr error
}

func (r *recorder) callbacks() lifecycle.CloseLoopOpts {
	return lifecycle.CloseLoopOpts{
		ResolveIngestFiles: func(cwd, pendingDir string, args []string) ([]string, error) {
			r.order = append(r.order, "close-loop")
			if r.closeLoopErr != nil {
				return nil, r.closeLoopErr
			}
			return nil, nil
		},
		IngestFilesToPool: func(cwd string, files []string) (lifecycle.CloseLoopIngestResult, error) {
			return lifecycle.CloseLoopIngestResult{}, nil
		},
		AutoPromoteFn: func(p *pool.Pool, threshold time.Duration, includeGold bool) (lifecycle.CloseLoopAutoPromoteResult, error) {
			return lifecycle.CloseLoopAutoPromoteResult{}, nil
		},
		ProcessCitationFeedback: func(cwd string) (int, int, int) {
			return 0, 0, 0
		},
		PromoteCitedLearnings: func(cwd string, quiet bool) int { return 0 },
		PromoteToMemory:       func(cwd string) (int, error) { return 0, nil },
		ApplyMaturityFn: func(cwd string) (lifecycle.MaturityTransitionSummary, error) {
			return lifecycle.MaturityTransitionSummary{}, nil
		},
	}
}

// newReduceFixture lays out a minimal .agents/ tree plus a fresh
// Checkpoint ready for RunReduce to drive.
func newReduceFixture(t *testing.T) (string, *Checkpoint, *IngestResult) {
	t.Helper()
	cwd := t.TempDir()
	// Minimal learnings so harvest has something to work with.
	writeLearning(t, cwd, "2026-04-09-fixture.md", map[string]string{
		"type":       "learning",
		"maturity":   "provisional",
		"confidence": "0.8",
	}, "# Fixture\n\nBody.\n")
	// Empty findings dir so the router has somewhere to look.
	if err := os.MkdirAll(filepath.Join(cwd, ".agents", "findings"), 0o755); err != nil {
		t.Fatalf("mkdir findings: %v", err)
	}

	cp, err := NewCheckpoint(cwd, "test-iter-1", 64*1024*1024)
	if err != nil {
		t.Fatalf("NewCheckpoint: %v", err)
	}

	ingest := &IngestResult{
		HarvestCatalog: nil, // harvest-promote stage no-ops when nil
	}
	return cwd, cp, ingest
}

func TestRunReduce_StageOrderEnforced(t *testing.T) {
	cwd, cp, ingest := newReduceFixture(t)
	rec := &recorder{}
	opts := newTestOpts(cwd)

	res, err := RunReduce(context.Background(), opts, ingest, cp, rec.callbacks(), io.Discard)
	if err != nil {
		t.Fatalf("RunReduce: %v", err)
	}
	if res.RolledBack {
		t.Fatalf("unexpected rollback: reason=%s", res.RollbackReason)
	}
	// close-loop is the only stage that records into rec.order via
	// a callback. The harvest-promote, dedup, maturity-temper,
	// defrag-prune, and findings-router stages execute real package
	// helpers (or the deferred stub) so their order is implicit in
	// RunReduce's in-source stage slice, not observable through
	// callbacks. Asserting close-loop got called at all confirms the
	// loop ran past its position in the stage sequence.
	if len(rec.order) != 1 || rec.order[0] != "close-loop" {
		t.Errorf("expected close-loop to be invoked exactly once, got %v", rec.order)
	}
}

func TestRunReduce_RollbackOnMetadataStrip(t *testing.T) {
	cwd, cp, ingest := newReduceFixture(t)

	// After NewCheckpoint clones the current live learnings into
	// staging, overwrite the live copy with a stripped version (no
	// frontmatter). VerifyMetadataRoundTrip will detect that every
	// staged frontmatter key is now missing from the live copy.
	livePath := filepath.Join(cwd, ".agents", "learnings", "2026-04-09-fixture.md")
	if err := os.WriteFile(livePath, []byte("# Fixture\n\nNo frontmatter here.\n"), 0o644); err != nil {
		t.Fatalf("strip live fixture: %v", err)
	}

	rec := &recorder{}
	_, err := RunReduce(context.Background(), newTestOpts(cwd), ingest, cp, rec.callbacks(), io.Discard)
	if err == nil {
		t.Fatal("expected RunReduce to fail integrity check")
	}
	// The result is captured even on error — re-run is unnecessary;
	// instead, assert the error surface contains the rollback reason.
	if !strings.Contains(err.Error(), "metadata integrity failed") {
		t.Errorf("expected metadata integrity error, got %v", err)
	}
}

func TestRunReduce_RollbackOnMetadataStrip_ResultState(t *testing.T) {
	cwd, cp, ingest := newReduceFixture(t)
	livePath := filepath.Join(cwd, ".agents", "learnings", "2026-04-09-fixture.md")
	if err := os.WriteFile(livePath, []byte("# Fixture\n\nNo frontmatter here.\n"), 0o644); err != nil {
		t.Fatalf("strip live fixture: %v", err)
	}

	rec := &recorder{}
	res, err := RunReduce(context.Background(), newTestOpts(cwd), ingest, cp, rec.callbacks(), io.Discard)
	if err == nil {
		t.Fatal("expected RunReduce to return error")
	}
	if !res.RolledBack {
		t.Errorf("expected RolledBack=true")
	}
	if res.RollbackReason == "" {
		t.Errorf("expected non-empty RollbackReason")
	}
	if res.MetadataIntegrity.Pass {
		t.Errorf("expected MetadataIntegrity.Pass=false")
	}
}

func TestRunReduce_CloseLoopCallbacksNil_Skipped(t *testing.T) {
	cwd, cp, ingest := newReduceFixture(t)
	var empty lifecycle.CloseLoopOpts
	res, err := RunReduce(context.Background(), newTestOpts(cwd), ingest, cp, empty, io.Discard)
	if err != nil {
		t.Fatalf("RunReduce: %v", err)
	}
	if !containsSubstring(res.Degraded, "close-loop") {
		t.Errorf("expected close-loop degraded note, got %v", res.Degraded)
	}
	if res.CloseLoopPromoted != 0 {
		t.Errorf("expected CloseLoopPromoted=0 when skipped, got %d", res.CloseLoopPromoted)
	}
}

func TestRunReduce_CloseLoopStageFailurePropagates(t *testing.T) {
	cwd, cp, ingest := newReduceFixture(t)
	rec := &recorder{closeLoopErr: errors.New("synthetic close-loop failure")}

	res, err := RunReduce(context.Background(), newTestOpts(cwd), ingest, cp, rec.callbacks(), io.Discard)
	if err == nil {
		t.Fatal("expected RunReduce to fail")
	}
	if !res.RolledBack {
		t.Error("expected RolledBack=true")
	}
	if res.RollbackReason == "" {
		t.Error("expected non-empty RollbackReason")
	}
	if _, ok := res.StageFailures["close-loop"]; !ok {
		t.Errorf("expected close-loop in StageFailures, got %v", res.StageFailures)
	}
}

// --- MEASURE -----------------------------------------------------------

func TestRunMeasure_CorpusComputeGoes(t *testing.T) {
	cwd := t.TempDir()
	writeLearning(t, cwd, "2026-04-09-m1.md", map[string]string{
		"type":        "learning",
		"maturity":    "provisional",
		"source_bead": "ao-1",
	}, "# M1\n\nBody.\n")

	res, err := RunMeasure(context.Background(), newTestOpts(cwd), io.Discard)
	if err != nil {
		t.Fatalf("RunMeasure: %v", err)
	}
	if res.Fitness == nil {
		t.Fatal("expected Fitness non-nil")
	}
	wantKeys := []string{
		"retrieval_precision",
		"retrieval_recall",
		"maturity_provisional_or_higher",
		"unresolved_findings",
		"citation_coverage",
		"inject_visibility",
		"cross_rig_dedup_ratio",
	}
	if len(res.FitnessSnapshot.Metrics) != len(wantKeys) {
		t.Errorf("expected %d metrics, got %d", len(wantKeys), len(res.FitnessSnapshot.Metrics))
	}
	for _, k := range wantKeys {
		if _, ok := res.FitnessSnapshot.Metrics[k]; !ok {
			t.Errorf("missing snapshot metric %s", k)
		}
	}
}

func TestRunMeasure_UnresolvedFindingsSignInverted(t *testing.T) {
	cwd := t.TempDir()
	// Seed .agents/ with a learning so corpus.Compute returns normally.
	writeLearning(t, cwd, "2026-04-09-m2.md", map[string]string{
		"type":     "learning",
		"maturity": "provisional",
	}, "# M2\n\nBody.\n")
	// Seed two unresolved findings.
	writeFindingRaw(t, cwd, "f-2026-04-09-001.md", "# First\n\nUnresolved.\n")
	writeFindingRaw(t, cwd, "f-2026-04-09-002.md", "# Second\n\nUnresolved.\n")

	res, err := RunMeasure(context.Background(), newTestOpts(cwd), io.Discard)
	if err != nil {
		t.Fatalf("RunMeasure: %v", err)
	}
	got, ok := res.FitnessSnapshot.Metrics["unresolved_findings"]
	if !ok {
		t.Fatal("missing unresolved_findings metric")
	}
	if got != -2.0 {
		t.Errorf("expected unresolved_findings=-2.0, got %v", got)
	}
}

func TestRunMeasure_MissingAgentsDir(t *testing.T) {
	cwd := t.TempDir()
	// No .agents/ directory at all.
	_, err := RunMeasure(context.Background(), newTestOpts(cwd), io.Discard)
	if err == nil {
		t.Fatal("expected error for missing .agents/")
	}
	if !strings.Contains(err.Error(), ".agents") {
		t.Errorf("expected error to mention .agents/, got %v", err)
	}
}

// --- helpers -----------------------------------------------------------

// containsSubstring returns true if any entry in haystack contains
// needle. Used by the degraded-note assertions so tests can match
// stable fragments of the degraded message without locking in the
// full human-readable wording.
func containsSubstring(haystack []string, needle string) bool {
	for _, s := range haystack {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}
