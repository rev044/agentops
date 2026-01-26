package orchestrator

import (
	"sort"
	"strings"
)

// ConflictResolver handles conflicting findings across agents and pods.
type ConflictResolver struct {
	// consensus rules to apply.
	consensus *Consensus
}

// NewConflictResolver creates a conflict resolver.
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{
		consensus: NewConsensus(),
	}
}

// ResolveWithinPod resolves conflicts within a single pod.
// Uses single-veto rule: ANY CRITICAL from ANY agent = pod CRITICAL.
func (r *ConflictResolver) ResolveWithinPod(agentResults []AgentResult) PodResult {
	var allFindings []Finding
	var totalDuration int64

	for _, agent := range agentResults {
		allFindings = append(allFindings, agent.Findings...)
		totalDuration += agent.Duration.Nanoseconds()
	}

	// Deduplicate findings within the pod
	merged := r.consensus.Deduplicate(allFindings)

	// Note: Single-veto is applied at the cluster level in ResolveAcrossPods
	// The pod result just collects findings for cluster-level synthesis

	return PodResult{
		Findings:     merged,
		AgentResults: agentResults,
	}
}

// ResolveAcrossPods resolves conflicts across multiple pods.
// Uses quorum rule: 70% agreement needed for HIGH findings.
func (r *ConflictResolver) ResolveAcrossPods(podResults []PodResult) ClusterResult {
	var allFindings []Finding

	for _, pod := range podResults {
		allFindings = append(allFindings, pod.Findings...)
	}

	// Group similar findings
	groups := r.groupSimilarFindings(allFindings)

	var merged []Finding
	for _, group := range groups {
		resolved := r.resolveGroup(group, len(podResults))
		if resolved != nil {
			merged = append(merged, *resolved)
		}
	}

	// Sort by severity
	sort.Slice(merged, func(i, j int) bool {
		return SeverityOrder[merged[i].Severity] > SeverityOrder[merged[j].Severity]
	})

	// Determine highest severity
	highest := SeverityPass
	for _, f := range merged {
		if SeverityOrder[f.Severity] > SeverityOrder[highest] {
			highest = f.Severity
		}
	}

	var podNames []string
	for _, pod := range podResults {
		podNames = append(podNames, pod.Config.Name)
	}

	return ClusterResult{
		PodNames:        podNames,
		MergedFindings:  merged,
		HighestSeverity: highest,
	}
}

// ResolveFinal produces the final verdict from cluster results.
func (r *ConflictResolver) ResolveFinal(clusters []ClusterResult) FinalResult {
	var allFindings []Finding

	for _, cluster := range clusters {
		allFindings = append(allFindings, cluster.MergedFindings...)
	}

	// Final deduplication
	merged := r.consensus.Deduplicate(allFindings)

	// Apply final quorum (70% of clusters must agree for HIGH)
	merged = r.applyClusterQuorum(merged, clusters)

	// Calculate verdict
	verdict := r.consensus.VerdictFromFindings(merged)

	// Count severities
	criticalCount := 0
	highCount := 0
	for _, f := range merged {
		switch f.Severity {
		case SeverityCritical:
			criticalCount++
		case SeverityHigh:
			highCount++
		}
	}

	return FinalResult{
		Verdict:        verdict,
		Findings:       merged,
		CriticalCount:  criticalCount,
		HighCount:      highCount,
		ClusterResults: clusters,
	}
}

// groupSimilarFindings groups findings by semantic similarity.
func (r *ConflictResolver) groupSimilarFindings(findings []Finding) [][]Finding {
	groups := make(map[string][]Finding)

	for _, f := range findings {
		key := r.similarityKey(f)
		groups[key] = append(groups[key], f)
	}

	var result [][]Finding
	for _, group := range groups {
		result = append(result, group)
	}

	return result
}

// similarityKey generates a key for similarity grouping.
func (r *ConflictResolver) similarityKey(f Finding) string {
	// Normalize: lowercase, remove common words, keep category
	title := strings.ToLower(f.Title)
	// Remove common prefixes
	title = strings.TrimPrefix(title, "potential ")
	title = strings.TrimPrefix(title, "possible ")
	title = strings.TrimPrefix(title, "missing ")

	return f.Category + ":" + title
}

// resolveGroup resolves a group of similar findings.
func (r *ConflictResolver) resolveGroup(group []Finding, totalPods int) *Finding {
	if len(group) == 0 {
		return nil
	}

	// CRITICAL always passes (single-veto)
	for _, f := range group {
		if f.Severity == SeverityCritical {
			merged := r.mergeFindingGroup(group)
			merged.Severity = SeverityCritical
			return &merged
		}
	}

	// Calculate agreement
	agreement := float64(len(group)) / float64(totalPods)

	// HIGH needs 70% agreement
	hasHigh := false
	for _, f := range group {
		if f.Severity == SeverityHigh {
			hasHigh = true
			break
		}
	}
	if hasHigh && agreement < 0.7 {
		// Downgrade to MEDIUM if not enough agreement
		merged := r.mergeFindingGroup(group)
		merged.Severity = SeverityMedium
		return &merged
	}

	// MEDIUM needs 50% agreement
	hasMedium := false
	for _, f := range group {
		if f.Severity == SeverityMedium {
			hasMedium = true
			break
		}
	}
	if hasMedium && agreement < 0.5 {
		// Downgrade to LOW
		merged := r.mergeFindingGroup(group)
		merged.Severity = SeverityLow
		return &merged
	}

	// LOW needs at least 2 sources
	if len(group) < 2 {
		return nil // Drop findings with only 1 source
	}

	merged := r.mergeFindingGroup(group)
	return &merged
}

// mergeFindingGroup combines multiple findings into one.
func (r *ConflictResolver) mergeFindingGroup(group []Finding) Finding {
	if len(group) == 0 {
		return Finding{}
	}

	// Start with highest severity finding
	var best Finding
	for _, f := range group {
		if SeverityOrder[f.Severity] > SeverityOrder[best.Severity] {
			best = f
		}
	}

	// Merge all files
	fileSet := make(map[string]bool)
	for _, f := range group {
		for _, file := range f.Files {
			fileSet[file] = true
		}
	}
	best.Files = nil
	for file := range fileSet {
		best.Files = append(best.Files, file)
	}
	sort.Strings(best.Files)

	// Merge all lines
	lineSet := make(map[int]bool)
	for _, f := range group {
		for _, line := range f.Lines {
			lineSet[line] = true
		}
	}
	best.Lines = nil
	for line := range lineSet {
		best.Lines = append(best.Lines, line)
	}
	sort.Ints(best.Lines)

	// Average confidence
	totalConf := 0.0
	for _, f := range group {
		totalConf += f.Confidence
	}
	best.Confidence = totalConf / float64(len(group))

	// Merge sources
	sources := make(map[string]bool)
	for _, f := range group {
		if f.Source != "" {
			sources[f.Source] = true
		}
	}
	if len(sources) > 0 {
		var sourceList []string
		for s := range sources {
			sourceList = append(sourceList, s)
		}
		best.Source = strings.Join(sourceList, ", ")
	}

	return best
}

// applyClusterQuorum filters findings based on cluster agreement.
func (r *ConflictResolver) applyClusterQuorum(findings []Finding, clusters []ClusterResult) []Finding {
	if len(clusters) == 0 {
		return findings
	}

	var result []Finding

	for _, f := range findings {
		// CRITICAL always passes
		if f.Severity == SeverityCritical {
			result = append(result, f)
			continue
		}

		// Count clusters that have this finding
		count := 0
		for _, cluster := range clusters {
			for _, cf := range cluster.MergedFindings {
				if r.similarityKey(f) == r.similarityKey(cf) {
					count++
					break
				}
			}
		}

		agreement := float64(count) / float64(len(clusters))

		// HIGH needs 70% of clusters
		if f.Severity == SeverityHigh {
			if agreement >= 0.7 {
				result = append(result, f)
			}
			continue
		}

		// MEDIUM needs 50% of clusters
		if f.Severity == SeverityMedium {
			if agreement >= 0.5 {
				result = append(result, f)
			}
			continue
		}

		// LOW passes if in any cluster
		if count > 0 {
			result = append(result, f)
		}
	}

	return result
}

// ConflictStats tracks conflict resolution statistics.
type ConflictStats struct {
	TotalFindings    int
	MergedFindings   int
	DroppedByQuorum  int
	UpgradedSeverity int
	DowngradedSeverity int
}

// GetStats returns statistics about conflict resolution.
func (r *ConflictResolver) GetStats(original, resolved []Finding) ConflictStats {
	stats := ConflictStats{
		TotalFindings:  len(original),
		MergedFindings: len(resolved),
	}

	// Count drops and severity changes
	originalKeys := make(map[string]Severity)
	for _, f := range original {
		key := r.similarityKey(f)
		if existing, ok := originalKeys[key]; ok {
			if SeverityOrder[f.Severity] > SeverityOrder[existing] {
				originalKeys[key] = f.Severity
			}
		} else {
			originalKeys[key] = f.Severity
		}
	}

	resolvedKeys := make(map[string]Severity)
	for _, f := range resolved {
		resolvedKeys[r.similarityKey(f)] = f.Severity
	}

	for key, origSev := range originalKeys {
		if resSev, ok := resolvedKeys[key]; ok {
			if SeverityOrder[resSev] > SeverityOrder[origSev] {
				stats.UpgradedSeverity++
			} else if SeverityOrder[resSev] < SeverityOrder[origSev] {
				stats.DowngradedSeverity++
			}
		} else {
			stats.DroppedByQuorum++
		}
	}

	return stats
}
