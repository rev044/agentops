package corpus

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/boshu2/agentops/cli/internal/harvest"
)

// FitnessVector is the canonical corpus-quality fitness capture used by
// Dream's MEASURE stage. All fields are deterministic functions of the
// current .agents/ corpus state — Compute must produce the same vector
// on the same corpus with zero side effects.
//
// The seven metrics are chosen so each moves independently under typical
// Dream iterations:
//
//   - RetrievalPrecision: pAtK on the live retrieval-bench eval set
//   - RetrievalRecall: recall@K on the same eval set
//   - MaturityProvisional: fraction of .agents/learnings whose frontmatter
//     `maturity` is at or above "provisional"
//   - UnresolvedFindings: count of .agents/findings/f-*.md whose body does
//     not contain an explicit resolution marker
//   - CitationCoverage: fraction of learnings with non-empty `source_bead`
//   - InjectVisibility: fraction of promoted learnings in .agents/learnings
//     that would be visible to `ao inject` (rough proxy: has frontmatter,
//     not superseded). See pm-FEAS-06 — this is a proxy metric until the
//     inject subsystem exposes a probe entry point.
//   - CrossRigDedupRatio: unique / total of harvestable artifacts across
//     rigs (0 when harvest catalog is unavailable).
type FitnessVector struct {
	RetrievalPrecision  float64   `json:"retrieval_precision"`
	RetrievalRecall     float64   `json:"retrieval_recall"`
	MaturityProvisional float64   `json:"maturity_provisional_or_higher"`
	UnresolvedFindings  int       `json:"unresolved_findings"`
	CitationCoverage    float64   `json:"citation_coverage"`
	InjectVisibility    float64   `json:"inject_visibility"`
	CrossRigDedupRatio  float64   `json:"cross_rig_dedup_ratio"`
	ComputedAt          time.Time `json:"computed_at"`
}

// maturityRanks maps the accepted maturity tokens to a numeric rank so
// comparisons are case-insensitive and tolerant to historical spelling.
// Anything less than provisional (rank 1) does not count against the
// "provisional or higher" fraction.
var maturityRanks = map[string]int{
	"provisional": 1,
	"accepted":    2,
	"stable":      3,
	"promoted":    4,
}

// Compute walks the corpus under cwd and returns a FitnessVector.
//
// The function is IO-bound but deterministic:
//   - Walks .agents/learnings/*.md to compute maturity, citation, and
//     inject-visibility fractions.
//   - Walks .agents/findings/f-*.md to count unresolved findings.
//   - Retrieval precision/recall and cross-rig dedup are populated only
//     when the relevant source data is available; otherwise they are
//     left as 0 and noted in the degraded list returned alongside.
//
// On missing .agents/ entirely, Compute returns a zero-value vector with
// a clear error. On individual missing subdirectories (e.g., no
// findings/), it returns a partial vector with 0 for the affected
// metrics and notes the degradation.
func Compute(cwd string) (*FitnessVector, []string, error) {
	agentsDir := filepath.Join(cwd, ".agents")
	info, err := os.Stat(agentsDir)
	if err != nil || !info.IsDir() {
		return &FitnessVector{ComputedAt: time.Now().UTC()}, nil,
			fmt.Errorf("corpus: .agents directory not found at %s", agentsDir)
	}

	vec := &FitnessVector{ComputedAt: time.Now().UTC()}
	var degraded []string

	// Learnings-derived metrics: maturity, citation, inject-visibility.
	learningsDir := filepath.Join(agentsDir, "learnings")
	lm, ldegraded := computeLearningsMetrics(learningsDir)
	vec.MaturityProvisional = lm.maturityFraction
	vec.CitationCoverage = lm.citationFraction
	vec.InjectVisibility = lm.injectVisibilityFraction
	degraded = append(degraded, ldegraded...)

	// Findings-derived metric: unresolved count.
	findingsDir := filepath.Join(agentsDir, "findings")
	unresolved, fdegraded := computeUnresolvedFindings(findingsDir)
	vec.UnresolvedFindings = unresolved
	degraded = append(degraded, fdegraded...)

	// Retrieval metrics: gated on a bench manifest we can consume.
	benchManifest := filepath.Join(cwd, "testdata", "retrieval-bench", "manifest.json")
	if _, berr := os.Stat(benchManifest); berr == nil {
		// The current bench package exposes only scoring helpers, not a
		// one-shot manifest runner. Until bench wiring lands we leave the
		// metrics at 0 and note the degradation so downstream Dream
		// logic can decide whether to block on it.
		degraded = append(degraded, "retrieval metrics deferred to bench package wiring")
	} else {
		degraded = append(degraded, "retrieval metrics unavailable: testdata/retrieval-bench/manifest.json missing")
	}

	// Cross-rig dedup ratio: best-effort via the harvest package. Scope
	// discovery to cwd so Compute stays hermetic — we are reporting the
	// fitness of *this* corpus, not the global workspace.
	ratio, cdegraded := computeCrossRigDedup(cwd)
	vec.CrossRigDedupRatio = ratio
	degraded = append(degraded, cdegraded...)

	// Sort degraded notes for deterministic output.
	sort.Strings(degraded)

	return vec, degraded, nil
}

// learningsMetrics holds the three learnings-derived fractions so
// computeLearningsMetrics can walk the directory once.
type learningsMetrics struct {
	maturityFraction         float64
	citationFraction         float64
	injectVisibilityFraction float64
}

// computeLearningsMetrics walks a learnings directory and computes the
// maturity, citation coverage, and inject-visibility fractions.
// Returns zeroed metrics and a degraded note when the directory is
// missing or empty.
func computeLearningsMetrics(dir string) (learningsMetrics, []string) {
	var degraded []string
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		degraded = append(degraded, "learnings metrics unavailable: .agents/learnings missing")
		return learningsMetrics{}, degraded
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		degraded = append(degraded, fmt.Sprintf("learnings metrics unavailable: %v", err))
		return learningsMetrics{}, degraded
	}

	var total, matured, cited, visible int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		fm, ok := parseFrontmatter(data)
		total++
		if !ok {
			// No frontmatter: not visible to inject, not matured, not cited.
			continue
		}

		if maturityOk(fm) {
			matured++
		}
		if citationOk(fm) {
			cited++
		}
		if injectVisible(fm) {
			visible++
		}
	}

	if total == 0 {
		degraded = append(degraded, "learnings metrics unavailable: no .md files under .agents/learnings")
		return learningsMetrics{}, degraded
	}

	return learningsMetrics{
		maturityFraction:         float64(matured) / float64(total),
		citationFraction:         float64(cited) / float64(total),
		injectVisibilityFraction: float64(visible) / float64(total),
	}, degraded
}

// parseFrontmatter extracts a YAML frontmatter block from a markdown file.
// Returns the parsed map and true when a valid leading `---` ... `---`
// block is found.
func parseFrontmatter(data []byte) (map[string]any, bool) {
	// Must start with "---" on its own line.
	if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
		return nil, false
	}
	// Skip the opening delimiter.
	rest := data
	if idx := bytes.IndexByte(rest, '\n'); idx >= 0 {
		rest = rest[idx+1:]
	}
	// Find the closing delimiter.
	closeMarker := bytes.Index(rest, []byte("\n---"))
	if closeMarker < 0 {
		return nil, false
	}
	block := rest[:closeMarker]
	var out map[string]any
	if err := yaml.Unmarshal(block, &out); err != nil {
		return nil, false
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, true
}

// maturityOk returns true when the frontmatter's maturity field is at or
// above "provisional" (case-insensitive, whitespace-trimmed).
func maturityOk(fm map[string]any) bool {
	raw, ok := fm["maturity"]
	if !ok {
		return false
	}
	s, ok := raw.(string)
	if !ok {
		return false
	}
	rank, ok := maturityRanks[strings.ToLower(strings.TrimSpace(s))]
	if !ok {
		return false
	}
	return rank >= 1
}

// citationOk returns true when the frontmatter has a non-empty
// source_bead field.
func citationOk(fm map[string]any) bool {
	raw, ok := fm["source_bead"]
	if !ok {
		return false
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case nil:
		return false
	default:
		// Any non-nil, non-string value (e.g., list, number) counts as
		// a populated citation.
		return true
	}
}

// injectVisible returns true when a learning has a frontmatter block and
// is NOT explicitly superseded. This is a proxy metric — see the
// FitnessVector doc for context (pm-FEAS-06).
func injectVisible(fm map[string]any) bool {
	raw, ok := fm["superseded"]
	if !ok {
		return true
	}
	switch v := raw.(type) {
	case bool:
		return !v
	case string:
		return strings.ToLower(strings.TrimSpace(v)) != "true"
	default:
		return true
	}
}

// computeUnresolvedFindings counts .agents/findings/f-*.md files whose
// body does not carry an explicit resolution marker.
func computeUnresolvedFindings(dir string) (int, []string) {
	var degraded []string
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		degraded = append(degraded, "unresolved findings unavailable: .agents/findings missing")
		return 0, degraded
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		degraded = append(degraded, fmt.Sprintf("unresolved findings unavailable: %v", err))
		return 0, degraded
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "f-") || !strings.HasSuffix(name, ".md") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if findingResolved(data) {
			continue
		}
		count++
	}
	return count, degraded
}

// findingResolved returns true when a finding body shows any of the
// accepted resolution markers: the prose `**Resolved:**` token, the
// `<!-- RESOLVED -->` HTML comment, or a frontmatter `resolved: true`
// field.
func findingResolved(data []byte) bool {
	if bytes.Contains(data, []byte("**Resolved:**")) {
		return true
	}
	if bytes.Contains(data, []byte("<!-- RESOLVED -->")) {
		return true
	}
	fm, ok := parseFrontmatter(data)
	if !ok {
		return false
	}
	raw, ok := fm["resolved"]
	if !ok {
		return false
	}
	switch v := raw.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(strings.TrimSpace(v)) == "true"
	default:
		return false
	}
}

// computeCrossRigDedup calls the harvest package to build a catalog over
// cwd and returns the unique-over-total ratio. When fewer than two rigs
// are discovered, or the catalog cannot be built, returns 0 and a
// degraded note. Scoping to cwd keeps Compute hermetic in tests and
// reports the fitness of *this* corpus rather than the global workspace.
func computeCrossRigDedup(cwd string) (float64, []string) {
	var degraded []string
	opts := harvest.DefaultWalkOptions()
	opts.Roots = []string{cwd}
	rigs, err := harvest.DiscoverRigs(opts)
	if err != nil {
		degraded = append(degraded, fmt.Sprintf("cross-rig dedup unavailable: %v", err))
		return 0, degraded
	}
	if len(rigs) < 2 {
		degraded = append(degraded, "cross-rig dedup unavailable: fewer than 2 rigs discovered")
		return 0, degraded
	}

	var all []harvest.Artifact
	for _, rig := range rigs {
		arts, _ := harvest.ExtractArtifacts(rig, opts)
		all = append(all, arts...)
	}
	if len(all) == 0 {
		degraded = append(degraded, "cross-rig dedup unavailable: no artifacts extracted")
		return 0, degraded
	}

	cat := harvest.BuildCatalog(all, 0.5)
	if cat == nil || cat.Summary.ArtifactsExtracted == 0 {
		degraded = append(degraded, "cross-rig dedup unavailable: empty catalog")
		return 0, degraded
	}

	total := cat.Summary.ArtifactsExtracted
	unique := cat.Summary.UniqueArtifacts
	if total <= 0 {
		return 0, degraded
	}
	return float64(unique) / float64(total), degraded
}
