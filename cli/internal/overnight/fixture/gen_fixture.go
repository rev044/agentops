// Package fixture is a fixture generator for the Dream nightly
// compounder L2 e2e test. It produces a realistic .agents/ tree with
// enough scale to exercise compound filter effects that smaller
// fixtures miss (pm-015: the 2026-04-02 flywheel-quality-fixes
// learning warned that cliff effects only surface at 150+ entries).
//
// The generator is fully deterministic via an explicit random seed so
// the L2 test is reproducible across runs and across machines. It
// only ever writes inside the caller-supplied dir — it never touches
// the real repo or the user's home directory.
package fixture

import (
	"fmt"
	// This package is a TEST FIXTURE GENERATOR, not production code. It
	// uses math/rand (seeded, not crypto/rand) for byte-identical L2
	// e2e reproducibility: the same Seed MUST produce the same files
	// across runs and machines, which rules out crypto/rand. No
	// generated value is ever used in an authentication, authorization,
	// token, key, nonce, or other security-sensitive context — the
	// output is .md files inside a caller-supplied temp dir.
	// nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FixtureOpts knobs the generator's output. Defaults come from
// DefaultOpts; zero values are substituted back to defaults by
// GenerateFixture so callers can pass the zero FixtureOpts when they
// want "a reasonable L2 fixture".
type FixtureOpts struct {
	// Seed deterministically controls the randomized fields (maturity,
	// utility, mtime, body text). Default: 42.
	Seed int64

	// LearningCount is the number of learning files to emit. Default
	// 150. pm-015 requires >= 150 to surface the compound-filter cliff.
	LearningCount int

	// FindingCount is the number of finding files to emit. Default 20.
	FindingCount int

	// PatternCount is the number of pattern files to emit. Default 5.
	PatternCount int

	// KnowledgeCount is the number of knowledge files to emit. Default 3.
	KnowledgeCount int

	// WithSourceBead is the fraction of learnings that have the
	// source_bead frontmatter field populated. Default 0.6.
	WithSourceBead float64

	// MaturityMix is the fractional distribution of maturity values
	// across generated learnings. Must sum to 1.0 (or close to it;
	// floating-point slop is tolerated). Default:
	// provisional=0.5, accepted=0.3, stable=0.15, promoted=0.05.
	MaturityMix map[string]float64

	// UtilityFloor is the minimum utility value assigned to a
	// generated learning. Default 0.1.
	UtilityFloor float64

	// UtilityCeiling is the maximum utility value assigned to a
	// generated learning. Default 0.9.
	UtilityCeiling float64

	// DaysSpan is the width of the date/mtime window. Generated
	// artifacts are spread uniformly across now-DaysSpan..now.
	// Default 180.
	DaysSpan int
}

// DefaultOpts returns the documented default FixtureOpts used by the
// L2 e2e test (pm-015: minimum 150 learnings).
func DefaultOpts() FixtureOpts {
	return FixtureOpts{
		Seed:           42,
		LearningCount:  150,
		FindingCount:   20,
		PatternCount:   5,
		KnowledgeCount: 3,
		WithSourceBead: 0.6,
		MaturityMix: map[string]float64{
			"provisional": 0.5,
			"accepted":    0.3,
			"stable":      0.15,
			"promoted":    0.05,
		},
		UtilityFloor:   0.1,
		UtilityCeiling: 0.9,
		DaysSpan:       180,
	}
}

// normalize returns opts with zero values replaced by documented
// defaults. It never returns an error; it is the silent-substitution
// equivalent of RunLoopOptions.normalize in the parent package.
func (o FixtureOpts) normalize() FixtureOpts {
	d := DefaultOpts()
	if o.Seed == 0 {
		o.Seed = d.Seed
	}
	if o.LearningCount <= 0 {
		o.LearningCount = d.LearningCount
	}
	if o.FindingCount <= 0 {
		o.FindingCount = d.FindingCount
	}
	if o.PatternCount <= 0 {
		o.PatternCount = d.PatternCount
	}
	if o.KnowledgeCount <= 0 {
		o.KnowledgeCount = d.KnowledgeCount
	}
	if o.WithSourceBead <= 0 {
		o.WithSourceBead = d.WithSourceBead
	}
	if len(o.MaturityMix) == 0 {
		o.MaturityMix = d.MaturityMix
	}
	if o.UtilityFloor <= 0 {
		o.UtilityFloor = d.UtilityFloor
	}
	if o.UtilityCeiling <= 0 {
		o.UtilityCeiling = d.UtilityCeiling
	}
	if o.DaysSpan <= 0 {
		o.DaysSpan = d.DaysSpan
	}
	return o
}

// GenerateFixture writes a realistic .agents/ tree under dir. It
// creates learnings, findings, patterns, and knowledge files per
// opts, plus an empty rpi/next-work.jsonl the findings router can
// append to. It never touches dir/.agents/overnight/ — that is
// Dream's own domain.
//
// The generator is deterministic given opts.Seed: re-running with
// the same seed produces byte-identical files (modulo mtime, which
// is computed from the seed-derived date window).
func GenerateFixture(dir string, opts FixtureOpts) error {
	opts = opts.normalize()
	if dir == "" {
		return fmt.Errorf("testdata: GenerateFixture requires a non-empty dir")
	}

	root := filepath.Join(dir, ".agents")
	subdirs := []string{"learnings", "findings", "patterns", "knowledge", "rpi"}
	for _, sub := range subdirs {
		if err := os.MkdirAll(filepath.Join(root, sub), 0o755); err != nil {
			return fmt.Errorf("testdata: mkdir %s: %w", sub, err)
		}
	}

	// Create an empty next-work.jsonl so the findings router finds
	// the expected target and appends cleanly.
	nextWork := filepath.Join(root, "rpi", "next-work.jsonl")
	if err := os.WriteFile(nextWork, nil, 0o644); err != nil {
		return fmt.Errorf("testdata: write next-work.jsonl: %w", err)
	}

	rng := rand.New(rand.NewSource(opts.Seed)) // #nosec G404 -- deterministic fixture data needs reproducible pseudo-randomness, not secrecy.
	now := time.Now().UTC()

	// Build a deterministic maturity bucket list by repeating each
	// maturity per its fraction * LearningCount. Sort for determinism
	// in the presence of Go's map iteration order.
	maturities := buildMaturityBuckets(opts.MaturityMix, opts.LearningCount)

	for i := 0; i < opts.LearningCount; i++ {
		if err := writeLearning(root, i, maturities[i], opts, rng, now); err != nil {
			return err
		}
	}

	for i := 0; i < opts.FindingCount; i++ {
		if err := writeFinding(root, i, opts, rng, now); err != nil {
			return err
		}
	}

	// Build a reference line listing every learning filename. Pattern
	// files include this list so lifecycle.FindOrphanLearnings does
	// not flag fixture learnings as orphans during the REDUCE prune
	// stage — prune's link check is a plain substring scan over
	// patterns/ and research/ content, so a simple concatenated list
	// is sufficient to keep every learning anchored.
	learningNames := make([]string, 0, opts.LearningCount)
	for i := 0; i < opts.LearningCount; i++ {
		learningNames = append(learningNames, fmt.Sprintf("learning-%03d.md", i))
	}

	for i := 0; i < opts.PatternCount; i++ {
		if err := writePattern(root, i, opts, rng, now, learningNames); err != nil {
			return err
		}
	}

	for i := 0; i < opts.KnowledgeCount; i++ {
		if err := writeKnowledge(root, i, opts, rng, now); err != nil {
			return err
		}
	}

	return nil
}

// buildMaturityBuckets returns a slice of length count where the
// ordered values are drawn from the MaturityMix distribution. Any
// rounding slop is absorbed into the first maturity key (sorted
// alphabetically for determinism).
func buildMaturityBuckets(mix map[string]float64, count int) []string {
	keys := make([]string, 0, len(mix))
	for k := range mix {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]string, 0, count)
	remaining := count
	for idx, k := range keys {
		frac := mix[k]
		n := int(frac * float64(count))
		if idx == len(keys)-1 {
			n = remaining // absorb slop into the last bucket
		}
		if n > remaining {
			n = remaining
		}
		for i := 0; i < n; i++ {
			out = append(out, k)
		}
		remaining -= n
	}
	// Top up with the first key if we still have slots (defensive).
	for len(out) < count && len(keys) > 0 {
		out = append(out, keys[0])
	}
	return out[:count]
}

// writeLearning emits a single learning file plus sets its mtime.
func writeLearning(root string, idx int, maturity string, opts FixtureOpts, rng *rand.Rand, now time.Time) error {
	name := fmt.Sprintf("learning-%03d.md", idx)
	path := filepath.Join(root, "learnings", name)

	utility := opts.UtilityFloor + rng.Float64()*(opts.UtilityCeiling-opts.UtilityFloor)
	confidence := 0.5 + rng.Float64()*0.5
	ageDays := rng.Intn(opts.DaysSpan + 1)
	date := now.AddDate(0, 0, -ageDays)

	var sourceBeadLine string
	if rng.Float64() < opts.WithSourceBead {
		sourceBeadLine = fmt.Sprintf("source_bead: ao-%03d\n", idx)
	}

	// Generate a distinctive body so the dedup pass (80% trigram
	// overlap) does not flag fixture learnings as duplicates. Each
	// body is a sequence of pseudo-random hex tokens seeded by idx
	// and the fixture seed; trigram overlap between two such bodies
	// is far below the threshold.
	var bodySb strings.Builder
	fmt.Fprintf(&bodySb, "# Synthetic Learning %03d\n\n", idx)
	for line := 0; line < 12; line++ {
		fmt.Fprintf(&bodySb, "token-%d: %08x %08x %08x %08x\n",
			line, rng.Uint32(), rng.Uint32(), rng.Uint32(), rng.Uint32())
	}
	fmt.Fprintf(&bodySb, "\nStable anchor: fixture-learning-%03d\n", idx)

	content := fmt.Sprintf(`---
title: Synthetic Learning %03d
type: learning
maturity: %s
utility: %.3f
confidence: %.3f
%sdate: %s
---

%s`,
		idx, maturity, utility, confidence, sourceBeadLine,
		date.Format("2006-01-02"), bodySb.String())

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("testdata: write learning %d: %w", idx, err)
	}
	if err := os.Chtimes(path, date, date); err != nil {
		return fmt.Errorf("testdata: chtimes learning %d: %w", idx, err)
	}
	return nil
}

// writeFinding emits a single finding file named to match the router
// regex `f-YYYY-MM-NN-III.md`. Roughly half carry a **Resolved:**
// marker; the rest do not (so the router has unresolved work to route).
func writeFinding(root string, idx int, opts FixtureOpts, rng *rand.Rand, now time.Time) error {
	// Deterministic date-ish naming: fix the year/month so the regex
	// in findings_router matches, then use idx as the day/seq.
	day := (idx % 28) + 1
	name := fmt.Sprintf("f-2026-03-%02d-%03d.md", day, idx)
	path := filepath.Join(root, "findings", name)

	severities := []string{"low", "medium", "high"}
	severity := severities[idx%len(severities)]

	ageDays := rng.Intn(opts.DaysSpan + 1)
	date := now.AddDate(0, 0, -ageDays)

	var resolvedMarker string
	if idx%2 == 0 {
		// Half resolved, half unresolved.
		resolvedMarker = "\n**Resolved:** yes — closed in test fixture.\n"
	}

	content := fmt.Sprintf(`---
title: Synthetic Finding %03d
type: finding
severity: %s
date: %s
---

Synthetic finding %d, generated by the Dream nightly compounder L2
fixture generator. Used to exercise the findings router.
%s`,
		idx, severity, date.Format("2006-01-02"), idx, resolvedMarker)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("testdata: write finding %d: %w", idx, err)
	}
	if err := os.Chtimes(path, date, date); err != nil {
		return fmt.Errorf("testdata: chtimes finding %d: %w", idx, err)
	}
	return nil
}

// writePattern emits a pattern stub. learningNames is a slice of every
// learning filename in the fixture; the pattern file concatenates them
// into its body so lifecycle.FindOrphanLearnings' substring reference
// check considers every learning anchored and therefore non-orphan.
// Without this, the 180-day mtime span would cause prune to delete
// every generated learning in the REDUCE stage.
func writePattern(root string, idx int, opts FixtureOpts, rng *rand.Rand, now time.Time, learningNames []string) error {
	name := fmt.Sprintf("pattern-%03d.md", idx)
	path := filepath.Join(root, "patterns", name)

	ageDays := rng.Intn(opts.DaysSpan + 1)
	date := now.AddDate(0, 0, -ageDays)

	// Only the first pattern carries the full learning reference list;
	// the rest stay small so the fixture doesn't bloat.
	var refBlock string
	if idx == 0 {
		var sb strings.Builder
		sb.WriteString("\nReferenced learnings:\n")
		for _, ln := range learningNames {
			sb.WriteString("- ")
			sb.WriteString(ln)
			sb.WriteByte('\n')
		}
		refBlock = sb.String()
	}

	content := fmt.Sprintf(`---
title: Synthetic Pattern %03d
type: pattern
date: %s
---

Synthetic pattern %d (L2 fixture stub).
%s`, idx, date.Format("2006-01-02"), idx, refBlock)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("testdata: write pattern %d: %w", idx, err)
	}
	if err := os.Chtimes(path, date, date); err != nil {
		return fmt.Errorf("testdata: chtimes pattern %d: %w", idx, err)
	}
	return nil
}

// writeKnowledge emits a small knowledge stub.
func writeKnowledge(root string, idx int, opts FixtureOpts, rng *rand.Rand, now time.Time) error {
	name := fmt.Sprintf("knowledge-%03d.md", idx)
	path := filepath.Join(root, "knowledge", name)

	ageDays := rng.Intn(opts.DaysSpan + 1)
	date := now.AddDate(0, 0, -ageDays)

	content := fmt.Sprintf(`---
title: Synthetic Knowledge %03d
type: knowledge
date: %s
---

Synthetic knowledge doc %d (L2 fixture stub).
`, idx, date.Format("2006-01-02"), idx)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("testdata: write knowledge %d: %w", idx, err)
	}
	if err := os.Chtimes(path, date, date); err != nil {
		return fmt.Errorf("testdata: chtimes knowledge %d: %w", idx, err)
	}
	return nil
}
