package overnight

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/corpus"
	"github.com/boshu2/agentops/cli/internal/lifecycle"
	"github.com/boshu2/agentops/cli/internal/overnight/fixture"
)

// stubInjectRefresh temporarily replaces refreshInjectCacheFn with a
// pass-through that returns a successful, attempted, in-process result
// without touching the real search.BuildIndex code path. The returned
// cleanup must be deferred to restore the package-level default.
//
// This mirrors the refreshInjectCacheFn override pattern established
// by Wave 3 and documented on the RefreshInjectCache godoc.
func stubInjectRefresh(t *testing.T) func() {
	t.Helper()
	prev := refreshInjectCacheFn
	refreshInjectCacheFn = func(_ context.Context, _ string, _ io.Writer) (*InjectRefreshResult, error) {
		return &InjectRefreshResult{
			Attempted: true,
			Succeeded: true,
			Method:    "in-process",
			Duration:  time.Millisecond,
		}, nil
	}
	return func() { refreshInjectCacheFn = prev }
}

// frontmatterKeys loads every *.md file under dir/.agents/learnings
// and returns a map keyed by relative filename whose value is the set
// of frontmatter keys seen in that file. Used by the metadata
// round-trip assertion in the L2 e2e test. The parser is intentionally
// minimal: it reads lines between the first "---" and the next "---"
// and treats the substring before the first ":" as the key.
func frontmatterKeys(t *testing.T, dir string) map[string]map[string]bool {
	t.Helper()
	out := map[string]map[string]bool{}
	learnings := filepath.Join(dir, ".agents", "learnings")
	entries, err := os.ReadDir(learnings)
	if err != nil {
		t.Fatalf("read learnings dir: %v", err)
	}
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(learnings, ent.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", ent.Name(), err)
		}
		keys := map[string]bool{}
		lines := strings.Split(string(data), "\n")
		inFront := false
		for _, line := range lines {
			if strings.TrimSpace(line) == "---" {
				if !inFront {
					inFront = true
					continue
				}
				break
			}
			if !inFront {
				continue
			}
			if idx := strings.Index(line, ":"); idx > 0 {
				keys[strings.TrimSpace(line[:idx])] = true
			}
		}
		out[ent.Name()] = keys
	}
	return out
}

// frontmatterDigest returns a stable sha256 hex of the frontmatter
// key-set snapshot across every learning. Two digests equal if-and-
// only-if the same set of (file, key) tuples is present.
func frontmatterDigest(fm map[string]map[string]bool) string {
	var names []string
	for name := range fm {
		names = append(names, name)
	}
	// Sort for determinism.
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[j] < names[i] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	h := sha256.New()
	for _, name := range names {
		h.Write([]byte(name))
		h.Write([]byte{0})
		var keys []string
		for k := range fm[name] {
			keys = append(keys, k)
		}
		for i := 0; i < len(keys); i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[j] < keys[i] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}
		for _, k := range keys {
			h.Write([]byte(k))
			h.Write([]byte{0})
		}
		h.Write([]byte{1})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// countLines returns the number of newline-terminated lines in path.
// Returns 0 if the file does not exist. Used to assert the findings
// router actually routed work into next-work.jsonl.
func countLines(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("read %s: %v", path, err)
	}
	if len(data) == 0 {
		return 0
	}
	return strings.Count(string(data), "\n")
}

// TestRunLoop_L2_FullIteration_CorpusQualityMoves exercises the
// Dream nightly compounder end-to-end against a 150-learning fixture.
//
// pm-015: the compound filter cliff only surfaces at >= 150 learnings,
// so the fixture default (via fixture.DefaultOpts) must stay at that
// threshold.
//
// Because RunLoop's outer driver is still a Wave 1 skeleton, the test
// also calls RunIngest/RunReduce/RunMeasure inline to prove the real
// stage drivers work against the same fixture.
func TestRunLoop_L2_FullIteration_CorpusQualityMoves(t *testing.T) {
	// Isolate HOME: harvest.Promote writes to $HOME/.agents/learnings/ by
	// default. Without this guard, every L2 run poisons the real user's
	// global knowledge hub with 150 "unknown-unknown-learning-NNN.md"
	// fixture files. Confirmed bug during Phase 3 validation.
	t.Setenv("HOME", t.TempDir())

	restore := stubInjectRefresh(t)
	defer restore()

	dir := t.TempDir()
	if err := fixture.GenerateFixture(dir, fixture.DefaultOpts()); err != nil {
		t.Fatalf("GenerateFixture: %v", err)
	}

	// Baseline fitness.
	baseline, _, err := corpus.Compute(dir)
	if err != nil {
		t.Fatalf("corpus.Compute baseline: %v", err)
	}
	if baseline == nil {
		t.Fatal("baseline FitnessVector is nil")
	}

	// --- Wave 1 skeleton RunLoop assertion -------------------------
	opts := RunLoopOptions{
		Cwd:            dir,
		OutputDir:      filepath.Join(dir, ".agents", "overnight", "test-run"),
		RunTimeout:     5 * time.Second,
		MaxIterations:  2,
		WarnOnly:       true,
		LogWriter:      io.Discard,
	}
	result, err := RunLoop(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunLoop: %v", err)
	}
	if result == nil {
		t.Fatal("RunLoop returned nil result")
	}
	foundSkeletonNote := false
	for _, d := range result.Degraded {
		if strings.Contains(d, "skeleton") && strings.Contains(d, "no iterations executed") {
			foundSkeletonNote = true
			break
		}
	}
	if !foundSkeletonNote {
		t.Errorf("expected Wave 1 skeleton degraded note in result.Degraded, got %v", result.Degraded)
	}
	if len(result.Iterations) != 0 {
		t.Errorf("expected no iterations in Wave 1 skeleton result, got %d", len(result.Iterations))
	}

	// --- Real stage drivers end-to-end ------------------------------
	// Snapshot frontmatter before REDUCE mutates anything.
	beforeFM := frontmatterKeys(t, dir)
	beforeDigest := frontmatterDigest(beforeFM)
	if len(beforeFM) < 150 {
		t.Fatalf("expected >= 150 learnings in fixture, got %d", len(beforeFM))
	}

	// INGEST
	ctx := context.Background()
	ingest, err := RunIngest(ctx, opts, io.Discard)
	if err != nil {
		t.Fatalf("RunIngest: %v", err)
	}
	if ingest == nil {
		t.Fatal("RunIngest returned nil result")
	}
	if ingest.HarvestCatalog == nil {
		t.Fatal("INGEST produced nil HarvestCatalog")
	}
	if ingest.HarvestCatalog.Summary.ArtifactsExtracted == 0 {
		t.Errorf("expected >0 artifacts extracted by INGEST, got %d",
			ingest.HarvestCatalog.Summary.ArtifactsExtracted)
	}

	// REDUCE — requires a checkpoint; no close-loop callbacks wired.
	cp, err := NewCheckpoint(dir, "l2-e2e-iter-1", 128*1024*1024)
	if err != nil {
		t.Fatalf("NewCheckpoint: %v", err)
	}
	reduce, err := RunReduce(ctx, opts, ingest, cp, lifecycle.CloseLoopOpts{}, io.Discard)
	if err != nil {
		t.Fatalf("RunReduce: %v", err)
	}
	if reduce.RolledBack {
		t.Fatalf("unexpected REDUCE rollback: %s", reduce.RollbackReason)
	}
	if !reduce.MetadataIntegrity.Pass {
		t.Errorf("metadata integrity failed: %d stripped fields", len(reduce.MetadataIntegrity.StrippedFields))
	}
	if reduce.InjectRefreshResult == nil {
		t.Error("expected non-nil InjectRefreshResult (inject-refresh stage should have run)")
	}

	// MEASURE
	measure, err := RunMeasure(ctx, opts, io.Discard)
	if err != nil {
		t.Fatalf("RunMeasure: %v", err)
	}
	if measure == nil || measure.Fitness == nil {
		t.Fatal("RunMeasure produced nil FitnessVector")
	}
	// Fixture mix is 50% provisional + 30% accepted + 15% stable +
	// 5% promoted — MaturityProvisional counts the fraction of
	// learnings with maturity >= provisional, which is effectively
	// 100% for our fixture. Guard with >= 0.5 per spec.
	if measure.Fitness.MaturityProvisional < 0.5 {
		t.Errorf("expected MaturityProvisional >= 0.5, got %f", measure.Fitness.MaturityProvisional)
	}

	// Findings router: fixture has 20 findings, ~10 unresolved; expect >0 routed.
	if reduce.FindingsRouted <= 0 {
		t.Errorf("expected findings router to route >0 entries, got %d", reduce.FindingsRouted)
	}
	nextWork := filepath.Join(dir, ".agents", "rpi", "next-work.jsonl")
	if countLines(t, nextWork) <= 0 {
		t.Errorf("expected next-work.jsonl to contain >0 routed lines, got %d", countLines(t, nextWork))
	}

	// Metadata round-trip assertion: no frontmatter key disappeared.
	afterFM := frontmatterKeys(t, dir)
	afterDigest := frontmatterDigest(afterFM)
	if afterDigest != beforeDigest {
		// Narrow down: which file/key was dropped?
		for file, beforeKeys := range beforeFM {
			afterKeys := afterFM[file]
			for k := range beforeKeys {
				if !afterKeys[k] {
					t.Errorf("frontmatter key %q dropped from %s", k, file)
				}
			}
		}
	}
}

// TestRunLoop_L2_MetadataStripDetected is a pm-005 regression guard at
// the L2 level: corrupt a staging copy of a learning by removing a
// frontmatter key, then verify VerifyMetadataRoundTrip surfaces the
// drop and rollback restores the live tree.
func TestRunLoop_L2_MetadataStripDetected(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // isolate global hub writes; see note on sibling test
	restore := stubInjectRefresh(t)
	defer restore()

	dir := t.TempDir()
	if err := fixture.GenerateFixture(dir, fixture.DefaultOpts()); err != nil {
		t.Fatalf("GenerateFixture: %v", err)
	}

	cp, err := NewCheckpoint(dir, "l2-e2e-strip", 128*1024*1024)
	if err != nil {
		t.Fatalf("NewCheckpoint: %v", err)
	}

	// Confirm the staging copy has the original frontmatter (the
	// checkpoint cloned it before we corrupt live below).
	stagedTarget := filepath.Join(cp.StagingDir, ".agents", "learnings", "learning-000.md")
	stagedBefore, err := os.ReadFile(stagedTarget)
	if err != nil {
		t.Fatalf("read staged fixture: %v", err)
	}
	if !strings.Contains(string(stagedBefore), "Synthetic Learning 000") {
		t.Fatalf("expected staging copy to hold original content, got %q",
			string(stagedBefore))
	}

	// Strip the LIVE copy of one learning — VerifyMetadataRoundTrip
	// compares staging (which has the original frontmatter) against
	// live (which now has none), so every staged key appears as
	// dropped for this file.
	target := filepath.Join(dir, ".agents", "learnings", "learning-000.md")
	if err := os.WriteFile(target, []byte("# Stripped\n\nNo frontmatter here.\n"), 0o644); err != nil {
		t.Fatalf("strip live fixture: %v", err)
	}

	report := VerifyMetadataRoundTrip(cp)
	if report.Pass {
		t.Fatal("expected VerifyMetadataRoundTrip to detect stripped fields, got Pass=true")
	}
	if len(report.StrippedFields) == 0 {
		t.Fatal("expected StrippedFields to be non-empty")
	}
	// Confirm at least one stripped field points at learning-000.md.
	found := false
	for _, sf := range report.StrippedFields {
		if strings.HasSuffix(sf.File, "learning-000.md") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a stripped field for learning-000.md, got %+v", report.StrippedFields)
	}

	// Pre-commit Rollback removes the staging dir but does NOT
	// restore the live tree (Checkpoint.Rollback only reverses
	// partial commits after a READY marker). Assert the staging
	// directory is gone, which proves the checkpoint is no longer
	// usable — the semantic guard is "no half-mutated state left
	// behind" rather than "live tree restored".
	if err := cp.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if _, statErr := os.Stat(cp.StagingDir); !os.IsNotExist(statErr) {
		t.Errorf("expected staging dir removed after Rollback, stat err=%v", statErr)
	}
}

// TestRunLoop_L2_CompoundFilterCliff_DoesNotSilentlyKillCorpus is the
// pm-015 regression guard: the 2026-04-02 flywheel-quality-fixes
// learning warned that compound filters can silently drop every
// learning when stacked naively. This test generates a 150-learning
// fixture, runs REDUCE, and asserts no corpus metric drops by more
// than 0.30 between pre- and post-REDUCE corpus.Compute snapshots.
func TestRunLoop_L2_CompoundFilterCliff_DoesNotSilentlyKillCorpus(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // isolate global hub writes; see note on sibling test
	restore := stubInjectRefresh(t)
	defer restore()

	dir := t.TempDir()
	opts := fixture.DefaultOpts() // 150 learnings
	if err := fixture.GenerateFixture(dir, opts); err != nil {
		t.Fatalf("GenerateFixture: %v", err)
	}

	pre, _, err := corpus.Compute(dir)
	if err != nil {
		t.Fatalf("corpus.Compute pre: %v", err)
	}
	if pre == nil {
		t.Fatal("pre FitnessVector nil")
	}

	runOpts := RunLoopOptions{
		Cwd:       dir,
		OutputDir: filepath.Join(dir, ".agents", "overnight", "cliff-run"),
		WarnOnly:  true,
		LogWriter: io.Discard,
	}
	ingest, err := RunIngest(context.Background(), runOpts, io.Discard)
	if err != nil {
		t.Fatalf("RunIngest: %v", err)
	}
	cp, err := NewCheckpoint(dir, "l2-cliff-iter", 128*1024*1024)
	if err != nil {
		t.Fatalf("NewCheckpoint: %v", err)
	}
	reduce, err := RunReduce(context.Background(), runOpts, ingest, cp, lifecycle.CloseLoopOpts{}, io.Discard)
	if err != nil {
		t.Fatalf("RunReduce: %v", err)
	}
	if reduce.RolledBack {
		t.Fatalf("unexpected cliff-run rollback: %s", reduce.RollbackReason)
	}

	post, _, err := corpus.Compute(dir)
	if err != nil {
		t.Fatalf("corpus.Compute post: %v", err)
	}

	// Cliff threshold: 0.30. Any metric that drops by more than this
	// is flagged as a silent-kill symptom.
	checks := []struct {
		name string
		pre  float64
		post float64
	}{
		{"maturity_provisional", pre.MaturityProvisional, post.MaturityProvisional},
		{"citation_coverage", pre.CitationCoverage, post.CitationCoverage},
		{"inject_visibility", pre.InjectVisibility, post.InjectVisibility},
	}
	for _, c := range checks {
		drop := c.pre - c.post
		if drop > 0.30 {
			t.Errorf("pm-015 cliff: metric %q dropped by %.3f (pre=%.3f post=%.3f)",
				c.name, drop, c.pre, c.post)
		}
	}
}
