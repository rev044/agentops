package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/boshu2/agentops/cli/internal/types"
)

// computeGoldenSignals calculates the four golden signals that distinguish
// knowledge compounding from noise accumulation.
func computeGoldenSignals(baseDir string, days int) (*types.GoldenSignals, error) {
	gs := &types.GoldenSignals{}

	// Signal 1: Compounding Velocity Trend
	trend7d, trend30d, trendVerdict, err := computeVelocityTrend(baseDir)
	if err == nil {
		gs.VelocityTrend7d = trend7d
		gs.VelocityTrend30d = trend30d
		gs.TrendVerdict = trendVerdict
	}

	// Signal 2: Citation Utility Pipeline
	highPct, medDelta, appliedRatio, pipeVerdict := computeCitationPipeline(baseDir, days)
	gs.HighUtilityCitationPct = highPct
	gs.MedianUtilityDelta = medDelta
	gs.AppliedToRetrievedRatio = appliedRatio
	gs.PipelineVerdict = pipeVerdict

	// Signal 3: Research-to-Learning Closure
	orphanCount, orphanPct, avgAge, closureVerdict := computeResearchClosure(baseDir)
	gs.OrphanedResearchCount = orphanCount
	gs.OrphanedResearchPct = orphanPct
	gs.AvgOrphanAgeDays = avgAge
	gs.ClosureVerdict = closureVerdict

	// Signal 4: Knowledge Reuse Concentration
	gini, activePct, topBottom, concVerdict := computeReuseConcentration(baseDir, days)
	gs.CitationGini = gini
	gs.ActivePoolPct = activePct
	gs.Top10BottomRatio = topBottom
	gs.ConcentrationVerdict = concVerdict

	// Overall verdict
	gs.OverallVerdict = computeOverallVerdict(gs)

	return gs, nil
}

// computeVelocityTrend reads metric baselines and computes the slope of
// velocity over rolling 7d and 30d windows.
func computeVelocityTrend(baseDir string) (trend7d, trend30d float64, verdict string, err error) {
	metricsDir := filepath.Join(baseDir, ".agents", "ao", "metrics")
	entries, err := os.ReadDir(metricsDir)
	if err != nil {
		return 0, 0, "stagnant", nil // no baselines yet
	}

	type velocityPoint struct {
		dayOffset float64
		velocity  float64
	}

	var points []velocityPoint
	now := time.Now()

	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "baseline-") || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(metricsDir, e.Name()))
		if err != nil {
			continue
		}
		var m types.FlywheelMetrics
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		dayOffset := now.Sub(m.Timestamp).Hours() / 24.0
		points = append(points, velocityPoint{dayOffset: dayOffset, velocity: m.Velocity})
	}

	if len(points) < 3 {
		return 0, 0, "stagnant", nil // insufficient data
	}

	// Sort by dayOffset ascending (most recent = smallest dayOffset first)
	sort.Slice(points, func(i, j int) bool {
		return points[i].dayOffset < points[j].dayOffset
	})

	// Compute trend for 7d window
	var xs7, ys7 []float64
	for _, p := range points {
		if p.dayOffset <= 7 {
			xs7 = append(xs7, p.dayOffset)
			ys7 = append(ys7, p.velocity)
		}
	}
	if len(xs7) >= 2 {
		trend7d = linearRegressionSlope(xs7, ys7)
		// Negate: lower dayOffset = more recent, positive slope in
		// dayOffset means velocity decreasing over time
		trend7d = -trend7d
	}

	// Compute trend for 30d window
	var xs30, ys30 []float64
	for _, p := range points {
		if p.dayOffset <= 30 {
			xs30 = append(xs30, p.dayOffset)
			ys30 = append(ys30, p.velocity)
		}
	}
	if len(xs30) >= 2 {
		trend30d = linearRegressionSlope(xs30, ys30)
		trend30d = -trend30d
	}

	// Verdict based on 30d trend (more stable than 7d)
	trendRef := trend30d
	if len(xs30) < 3 {
		trendRef = trend7d
	}
	switch {
	case trendRef > 0.01:
		verdict = "compounding"
	case trendRef < -0.01:
		verdict = "decaying"
	default:
		verdict = "stagnant"
	}

	return trend7d, trend30d, verdict, nil
}

// computeCitationPipeline measures whether citations are delivering value.
func computeCitationPipeline(baseDir string, days int) (highPct, medianDelta, appliedRatio float64, verdict string) {
	citationsPath := filepath.Join(baseDir, ".agents", "ao", "citations.jsonl")
	feedbackPath := filepath.Join(baseDir, ".agents", "ao", "feedback.jsonl")
	now := time.Now()
	cutoff := now.AddDate(0, 0, -days)

	// Read citations for the period
	var applied, retrieved int
	citations := readJSONLFile(citationsPath)
	for _, raw := range citations {
		var c struct {
			CitedAt      time.Time `json:"cited_at"`
			CitationType string    `json:"citation_type"`
		}
		if json.Unmarshal(raw, &c) != nil || c.CitedAt.Before(cutoff) {
			continue
		}
		switch c.CitationType {
		case "applied":
			applied++
		case "retrieved":
			retrieved++
		}
	}
	total := applied + retrieved
	if total > 0 {
		appliedRatio = float64(applied) / float64(total)
	}

	// Read feedback for utility deltas
	var deltas []float64
	var highUtility int
	feedback := readJSONLFile(feedbackPath)
	for _, raw := range feedback {
		var f struct {
			RecordedAt    time.Time `json:"recorded_at"`
			Reward        float64   `json:"reward"`
			UtilityBefore float64   `json:"utility_before"`
			UtilityAfter  float64   `json:"utility_after"`
		}
		if json.Unmarshal(raw, &f) != nil || f.RecordedAt.Before(cutoff) {
			continue
		}
		deltas = append(deltas, f.UtilityAfter-f.UtilityBefore)
		if f.Reward > 0.6 {
			highUtility++
		}
	}

	if len(deltas) == 0 {
		return 0, 0, appliedRatio, "insufficient-data"
	}

	highPct = float64(highUtility) / float64(len(deltas)) * 100.0
	sort.Float64s(deltas)
	n := len(deltas)
	if n%2 == 0 {
		medianDelta = (deltas[n/2-1] + deltas[n/2]) / 2.0
	} else {
		medianDelta = deltas[n/2]
	}

	// Verdict
	switch {
	case highPct > 60:
		verdict = "reinforcing"
	case highPct > 30:
		verdict = "inert"
	default:
		verdict = "degrading"
	}

	return highPct, medianDelta, appliedRatio, verdict
}

// computeResearchClosure measures whether research is being mined into learnings.
// A research file is "closed" if any learning's YAML frontmatter source field
// references the research file path, or if there's date+keyword overlap.
func computeResearchClosure(baseDir string) (orphanCount int, orphanPct float64, avgAgeDays float64, verdict string) {
	researchDir := filepath.Join(baseDir, ".agents", "research")
	learningsDir := filepath.Join(baseDir, ".agents", "learnings")

	researchFiles, err := os.ReadDir(researchDir)
	if err != nil || len(researchFiles) == 0 {
		return 0, 0, 0, "starved" // no research at all
	}

	// Build set of learning source references
	learningSourceRefs := make(map[string]bool)
	learningEntries, _ := os.ReadDir(learningsDir)
	for _, le := range learningEntries {
		if le.IsDir() || !strings.HasSuffix(le.Name(), ".md") {
			continue
		}
		path := filepath.Join(learningsDir, le.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := string(data)
		// Extract source field from frontmatter
		if idx := strings.Index(content, "source:"); idx != -1 {
			line := content[idx:]
			if nl := strings.IndexByte(line, '\n'); nl != -1 {
				line = line[:nl]
			}
			learningSourceRefs[line] = true
		}
	}

	// Check each research file for closure
	now := time.Now()
	var totalAgeDays float64
	var mdCount int

	for _, re := range researchFiles {
		if re.IsDir() || !strings.HasSuffix(re.Name(), ".md") {
			continue
		}
		mdCount++

		// Check if any learning references this research file
		closed := false
		researchName := re.Name()
		for ref := range learningSourceRefs {
			if strings.Contains(ref, researchName) || strings.Contains(ref, strings.TrimSuffix(researchName, ".md")) {
				closed = true
				break
			}
		}

		if !closed {
			orphanCount++
			info, err := re.Info()
			if err == nil {
				totalAgeDays += now.Sub(info.ModTime()).Hours() / 24.0
			}
		}
	}

	if mdCount == 0 {
		return 0, 0, 0, "starved"
	}

	orphanPct = float64(orphanCount) / float64(mdCount) * 100.0
	if orphanCount > 0 {
		avgAgeDays = totalAgeDays / float64(orphanCount)
	}

	switch {
	case orphanPct <= 10:
		verdict = "mining"
	case orphanPct < 50:
		verdict = "hoarding"
	default:
		verdict = "unmined"
	}

	return orphanCount, orphanPct, avgAgeDays, verdict
}

// countCitationsInPeriod reads a JSONL citations file and returns per-artifact
// citation counts for entries within the cutoff window.
func countCitationsInPeriod(citationsPath string, cutoff time.Time) map[string]int {
	counts := make(map[string]int)
	for _, raw := range readJSONLFile(citationsPath) {
		var c struct {
			ArtifactPath string    `json:"artifact_path"`
			CitedAt      time.Time `json:"cited_at"`
		}
		if json.Unmarshal(raw, &c) != nil || c.CitedAt.Before(cutoff) {
			continue
		}
		counts[filepath.Base(c.ArtifactPath)]++
	}
	return counts
}

// learningCitationCounts builds a float64 slice of citation counts for each
// .md file in learningEntries, using citationCounts for lookup.
// Returns the counts slice, the number of cited (non-zero) learnings, and total learnings.
func learningCitationCounts(entries []os.DirEntry, citationCounts map[string]int) (counts []float64, cited, total int) {
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		total++
		c := citationCounts[e.Name()]
		if c > 0 {
			cited++
		}
		counts = append(counts, float64(c))
	}
	return counts, cited, total
}

// top10BottomRatio computes the ratio of the sum of the top 10% values
// to the sum of the bottom 90% in a sorted slice. Requires len >= 10.
func top10BottomRatio(sorted []float64) float64 {
	n := len(sorted)
	top10Start := n - n/10
	if top10Start == n {
		top10Start = n - 1
	}
	var top10Sum, bottomSum float64
	for i, v := range sorted {
		if i >= top10Start {
			top10Sum += v
		} else {
			bottomSum += v
		}
	}
	if bottomSum > 0 {
		return top10Sum / bottomSum
	}
	if top10Sum > 0 {
		return 999.0 // sentinel: all citations in top 10%, maximum concentration
	}
	return 0
}

// computeReuseConcentration measures whether the whole knowledge pool is
// active or if only a few items get all the citations.
func computeReuseConcentration(baseDir string, days int) (gini, activePct, topBottomRatio float64, verdict string) {
	citationsPath := filepath.Join(baseDir, ".agents", "ao", "citations.jsonl")
	learningsDir := filepath.Join(baseDir, ".agents", "learnings")
	cutoff := time.Now().AddDate(0, 0, -days)

	citationCounts := countCitationsInPeriod(citationsPath, cutoff)

	learningEntries, err := os.ReadDir(learningsDir)
	if err != nil {
		return 0, 0, 0, "dormant"
	}

	counts, cited, total := learningCitationCounts(learningEntries, citationCounts)
	if total == 0 {
		return 0, 0, 0, "dormant"
	}

	activePct = float64(cited) / float64(total) * 100.0

	if len(counts) > 1 {
		gini = giniCoefficient(counts)
	}

	if len(counts) >= 10 {
		sort.Float64s(counts)
		topBottomRatio = top10BottomRatio(counts)
	}

	switch {
	case gini < 0.4 && activePct > 30:
		verdict = "distributed"
	case gini > 0.7 || activePct < 10:
		verdict = "dormant"
	default:
		verdict = "concentrated"
	}

	return gini, activePct, topBottomRatio, verdict
}

// computeOverallVerdict combines the four signal verdicts into one.
func computeOverallVerdict(gs *types.GoldenSignals) string {
	positive := 0
	negative := 0

	switch gs.TrendVerdict {
	case "compounding":
		positive++
	case "decaying":
		negative++
	}
	switch gs.PipelineVerdict {
	case "reinforcing":
		positive++
	case "degrading":
		negative++
	}
	switch gs.ClosureVerdict {
	case "mining":
		positive++
	case "hoarding", "unmined":
		negative++
	}
	switch gs.ConcentrationVerdict {
	case "distributed":
		positive++
	case "dormant":
		negative++
	}

	switch {
	case positive >= 3:
		return "compounding"
	case negative >= 3:
		return "decaying"
	default:
		return "accumulating"
	}
}

// linearRegressionSlope computes the least-squares slope for the given points.
func linearRegressionSlope(xs, ys []float64) float64 {
	n := float64(len(xs))
	if n < 2 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2 float64
	for i := range xs {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumX2 += xs[i] * xs[i]
	}
	denom := n*sumX2 - sumX*sumX
	if math.Abs(denom) < 1e-12 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

// giniCoefficient computes the Gini coefficient for a set of values.
// 0 = perfect equality, 1 = maximum inequality.
func giniCoefficient(values []float64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}
	sorted := make([]float64, n)
	copy(sorted, values)
	sort.Float64s(sorted)

	var sum float64
	for _, v := range sorted {
		sum += v
	}
	if sum == 0 {
		return 0
	}

	// Standard formula: G = (2 * sum(i*y_i) / (n * sum(y_i))) - (n+1)/n
	var weightedSum float64
	for i, v := range sorted {
		weightedSum += float64(i+1) * v
	}
	g := (2*weightedSum)/(float64(n)*sum) - float64(n+1)/float64(n)
	if g < 0 {
		g = 0
	}
	return g
}

// fprintGoldenSignals displays golden signals in a formatted table to the given writer.
func fprintGoldenSignals(w io.Writer, gs *types.GoldenSignals) {
	if gs == nil {
		fmt.Fprintln(w, "\nGolden Signals: insufficient data")
		return
	}

	fmt.Fprintln(w, "\n  GOLDEN SIGNALS:")
	fmt.Fprintln(w, "  ───────────────────────────────")
	fmt.Fprintf(w, "  1. Velocity Trend:    %s\n", gs.TrendVerdict)
	fmt.Fprintf(w, "     7d slope: %+.4f   30d slope: %+.4f\n", gs.VelocityTrend7d, gs.VelocityTrend30d)
	fmt.Fprintf(w, "  2. Citation Pipeline: %s\n", gs.PipelineVerdict)
	fmt.Fprintf(w, "     High-util: %.1f%%   Med delta: %+.4f   Applied: %.1f%%\n",
		gs.HighUtilityCitationPct, gs.MedianUtilityDelta, gs.AppliedToRetrievedRatio*100)
	fmt.Fprintf(w, "  3. Research Closure:  %s\n", gs.ClosureVerdict)
	fmt.Fprintf(w, "     Orphans: %d (%.1f%%)   Avg age: %.0fd\n",
		gs.OrphanedResearchCount, gs.OrphanedResearchPct, gs.AvgOrphanAgeDays)
	fmt.Fprintf(w, "  4. Reuse Concentr.:   %s\n", gs.ConcentrationVerdict)
	fmt.Fprintf(w, "     Gini: %.3f   Active: %.1f%%   Top10/Bot: %.1fx\n",
		gs.CitationGini, gs.ActivePoolPct, gs.Top10BottomRatio)
	fmt.Fprintln(w, "  ───────────────────────────────")
	fmt.Fprintf(w, "  OVERALL: %s\n", gs.OverallVerdict)
	fmt.Fprintln(w)
}

// readJSONLFile reads a JSONL file and returns raw JSON messages.
func readJSONLFile(path string) []json.RawMessage {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var results []json.RawMessage
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		cp := make([]byte, len(line))
		copy(cp, line)
		results = append(results, json.RawMessage(cp))
	}
	return results
}
