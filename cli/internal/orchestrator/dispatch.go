package orchestrator

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

// WaveDispatcher manages hierarchical agent dispatch.
type WaveDispatcher struct {
	Config DispatchConfig

	// ResultChan streams results as they arrive.
	ResultChan chan interface{}

	// stopChan for early termination.
	stopChan chan struct{}
}

// NewWaveDispatcher creates a dispatcher with the given config.
func NewWaveDispatcher(config DispatchConfig) *WaveDispatcher {
	return &WaveDispatcher{
		Config:     config,
		ResultChan: make(chan interface{}, 100),
		stopChan:   make(chan struct{}),
	}
}

// DispatchAll runs the full 3-wave validation.
func (d *WaveDispatcher) DispatchAll(files []string) (*FinalResult, error) {
	startTime := time.Now()

	// Determine if hierarchical dispatch is needed
	if len(files) < 20 {
		// Use simple flat dispatch for small reviews
		return d.dispatchFlat(files)
	}

	// Wave 1: Pod analysis
	podResults, err := d.DispatchWave1(files)
	if err != nil {
		return nil, fmt.Errorf("wave 1 failed: %w", err)
	}

	// Check for early termination
	if d.Config.EarlyTermination && d.hasUnanimousCritical(podResults) {
		return d.createEarlyTerminationResult(podResults, startTime), nil
	}

	// Wave 2: Cluster synthesis
	clusterResults, err := d.DispatchWave2(podResults)
	if err != nil {
		return nil, fmt.Errorf("wave 2 failed: %w", err)
	}

	// Wave 3: Final synthesis
	finalResult, err := d.DispatchWave3(clusterResults)
	if err != nil {
		return nil, fmt.Errorf("wave 3 failed: %w", err)
	}

	finalResult.PodResults = podResults
	finalResult.ClusterResults = clusterResults
	finalResult.Duration = time.Since(startTime)
	finalResult.CompletedAt = time.Now()

	return finalResult, nil
}

// DispatchWave1 runs parallel pod analysis.
func (d *WaveDispatcher) DispatchWave1(files []string) ([]PodResult, error) {
	// Partition files across pods
	pods := d.createPods(files)

	// In a real implementation, this would dispatch to actual agents
	// For now, we create a structure that can be consumed by Claude Code Task tool
	results := make([]PodResult, len(pods))

	for i, pod := range pods {
		result := PodResult{
			Config:   pod,
			Findings: []Finding{},
			Duration: 0, // Will be filled by actual execution
		}
		results[i] = result
	}

	return results, nil
}

// DispatchWave2 synthesizes pod results into clusters.
func (d *WaveDispatcher) DispatchWave2(pods []PodResult) ([]ClusterResult, error) {
	// Group pods into clusters (2-3 pods per cluster)
	clusters := d.createClusters(pods)

	results := make([]ClusterResult, len(clusters))

	for i, cluster := range clusters {
		// Merge findings with deduplication
		merged := d.mergeFindings(cluster.findings)

		// Apply single-veto rule
		highest := d.applySingleVeto(merged)

		results[i] = ClusterResult{
			ClusterID:       cluster.id,
			PodNames:        cluster.podNames,
			MergedFindings:  merged,
			HighestSeverity: highest,
		}
	}

	return results, nil
}

// DispatchWave3 produces the final synthesis.
func (d *WaveDispatcher) DispatchWave3(clusters []ClusterResult) (*FinalResult, error) {
	// Collect all findings
	var allFindings []Finding
	for _, cluster := range clusters {
		allFindings = append(allFindings, cluster.MergedFindings...)
	}

	// Final deduplication
	merged := d.mergeFindings(allFindings)

	// Apply quorum rule for HIGH findings
	merged = d.applyQuorum(merged, clusters)

	// Calculate verdict
	verdict := d.calculateVerdict(merged)

	// Count by severity
	criticalCount := 0
	highCount := 0
	for _, f := range merged {
		if f.Severity == SeverityCritical {
			criticalCount++
		} else if f.Severity == SeverityHigh {
			highCount++
		}
	}

	return &FinalResult{
		Verdict:       verdict,
		Findings:      merged,
		CriticalCount: criticalCount,
		HighCount:     highCount,
		Summary:       d.generateSummary(merged, verdict),
	}, nil
}

// createPods partitions files into validation pods.
func (d *WaveDispatcher) createPods(files []string) []PodConfig {
	// Use standard pod categories
	pods := make([]PodConfig, len(PodCategories))
	copy(pods, PodCategories)

	// Distribute files across pods based on category hints
	// In practice, all pods would see all files but focus on their area
	for i := range pods {
		pods[i].Files = files
		if pods[i].AgentCount == 0 {
			pods[i].AgentCount = d.Config.PodSize
		}
	}

	// Limit to MaxPods
	if len(pods) > d.Config.MaxPods {
		pods = pods[:d.Config.MaxPods]
	}

	return pods
}

// clusterInfo groups pods for synthesis.
type clusterInfo struct {
	id       string
	podNames []string
	findings []Finding
}

// createClusters groups pods into clusters.
func (d *WaveDispatcher) createClusters(pods []PodResult) []clusterInfo {
	// Group 2-3 pods per cluster
	clusterSize := 2
	if len(pods) > 6 {
		clusterSize = 3
	}

	var clusters []clusterInfo
	for i := 0; i < len(pods); i += clusterSize {
		end := i + clusterSize
		if end > len(pods) {
			end = len(pods)
		}

		cluster := clusterInfo{
			id: generateID(),
		}
		for j := i; j < end; j++ {
			cluster.podNames = append(cluster.podNames, pods[j].Config.Name)
			cluster.findings = append(cluster.findings, pods[j].Findings...)
		}
		clusters = append(clusters, cluster)
	}

	return clusters
}

// mergeFindings deduplicates similar findings.
func (d *WaveDispatcher) mergeFindings(findings []Finding) []Finding {
	if len(findings) == 0 {
		return findings
	}

	// Group by category and similarity
	groups := make(map[string][]Finding)
	for _, f := range findings {
		key := fmt.Sprintf("%s:%s", f.Category, normalizeTitle(f.Title))
		groups[key] = append(groups[key], f)
	}

	// Merge each group, keeping highest severity
	var merged []Finding
	for _, group := range groups {
		best := group[0]
		for _, f := range group[1:] {
			if SeverityOrder[f.Severity] > SeverityOrder[best.Severity] {
				best = f
			}
			// Merge files
			best.Files = mergeStrings(best.Files, f.Files)
		}
		merged = append(merged, best)
	}

	// Sort by severity (highest first)
	sort.Slice(merged, func(i, j int) bool {
		return SeverityOrder[merged[i].Severity] > SeverityOrder[merged[j].Severity]
	})

	return merged
}

// applySingleVeto returns the highest severity (single-veto rule).
func (d *WaveDispatcher) applySingleVeto(findings []Finding) Severity {
	highest := SeverityPass
	for _, f := range findings {
		if SeverityOrder[f.Severity] > SeverityOrder[highest] {
			highest = f.Severity
		}
	}
	return highest
}

// applyQuorum filters findings based on cross-cluster agreement.
func (d *WaveDispatcher) applyQuorum(findings []Finding, clusters []ClusterResult) []Finding {
	// CRITICAL findings always pass (single-veto)
	// HIGH findings need quorum agreement
	var result []Finding

	for _, f := range findings {
		if f.Severity == SeverityCritical {
			result = append(result, f)
			continue
		}

		if f.Severity == SeverityHigh {
			// Check if this finding appears in enough clusters
			count := 0
			for _, cluster := range clusters {
				for _, cf := range cluster.MergedFindings {
					if similarFindings(f, cf) {
						count++
						break
					}
				}
			}
			agreement := float64(count) / float64(len(clusters))
			if agreement >= d.Config.QuorumThreshold {
				result = append(result, f)
			}
			continue
		}

		// MEDIUM and LOW pass through
		result = append(result, f)
	}

	return result
}

// calculateVerdict determines the overall result.
func (d *WaveDispatcher) calculateVerdict(findings []Finding) Severity {
	if len(findings) == 0 {
		return SeverityPass
	}

	// Single-veto: any CRITICAL means CRITICAL verdict
	for _, f := range findings {
		if f.Severity == SeverityCritical {
			return SeverityCritical
		}
	}

	// Check for HIGH findings
	for _, f := range findings {
		if f.Severity == SeverityHigh {
			return SeverityHigh
		}
	}

	// Check for MEDIUM findings
	for _, f := range findings {
		if f.Severity == SeverityMedium {
			return SeverityMedium
		}
	}

	return SeverityLow
}

// generateSummary creates a summary of findings.
func (d *WaveDispatcher) generateSummary(findings []Finding, verdict Severity) string {
	if len(findings) == 0 {
		return "No issues found. Code passed all validation checks."
	}

	counts := make(map[Severity]int)
	for _, f := range findings {
		counts[f.Severity]++
	}

	return fmt.Sprintf("Found %d CRITICAL, %d HIGH, %d MEDIUM, %d LOW issues. Verdict: %s",
		counts[SeverityCritical],
		counts[SeverityHigh],
		counts[SeverityMedium],
		counts[SeverityLow],
		verdict)
}

// hasUnanimousCritical checks if multiple pods report the same CRITICAL.
func (d *WaveDispatcher) hasUnanimousCritical(pods []PodResult) bool {
	criticalCounts := make(map[string]int)

	for _, pod := range pods {
		for _, f := range pod.Findings {
			if f.Severity == SeverityCritical {
				key := normalizeTitle(f.Title)
				criticalCounts[key]++
			}
		}
	}

	// Need 3+ pods agreeing for early termination
	for _, count := range criticalCounts {
		if count >= 3 {
			return true
		}
	}

	return false
}

// createEarlyTerminationResult creates a result for early termination.
func (d *WaveDispatcher) createEarlyTerminationResult(pods []PodResult, startTime time.Time) *FinalResult {
	// Collect CRITICAL findings
	var criticals []Finding
	for _, pod := range pods {
		for _, f := range pod.Findings {
			if f.Severity == SeverityCritical {
				criticals = append(criticals, f)
			}
		}
	}

	merged := d.mergeFindings(criticals)

	return &FinalResult{
		Verdict:       SeverityCritical,
		Findings:      merged,
		CriticalCount: len(merged),
		Summary:       "EARLY TERMINATION: Multiple pods identified unanimous CRITICAL issues.",
		Duration:      time.Since(startTime),
		CompletedAt:   time.Now(),
		PodResults:    pods,
	}
}

// dispatchFlat handles simple validation without hierarchical dispatch.
func (d *WaveDispatcher) dispatchFlat(files []string) (*FinalResult, error) {
	startTime := time.Now()

	// For flat dispatch, we just create a single pod result structure
	// The actual validation would be done by a single agent
	return &FinalResult{
		Verdict:     SeverityPass,
		Findings:    []Finding{},
		Summary:     fmt.Sprintf("Flat validation of %d files (hierarchical not needed)", len(files)),
		Duration:    time.Since(startTime),
		CompletedAt: time.Now(),
	}, nil
}

// Stop signals early termination.
func (d *WaveDispatcher) Stop() {
	close(d.stopChan)
}

// Helper functions

func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based ID if crypto random fails
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func normalizeTitle(title string) string {
	// Simple normalization - in practice would use NLP
	return title
}

func mergeStrings(a, b []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range a {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func similarFindings(a, b Finding) bool {
	// Simple similarity check - in practice would use embeddings
	return a.Category == b.Category && normalizeTitle(a.Title) == normalizeTitle(b.Title)
}
