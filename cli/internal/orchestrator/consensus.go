package orchestrator

import (
	"sort"
)

// Consensus implements consensus rules for multi-agent validation.
type Consensus struct {
	// QuorumThreshold is the minimum agreement needed (0-1).
	QuorumThreshold float64

	// VetoOnCritical enables single-veto rule for CRITICAL.
	VetoOnCritical bool

	// StreamResults enables streaming partial results.
	StreamResults bool

	// DropLowOnContextPressure drops LOW findings when context is tight.
	DropLowOnContextPressure bool
}

// NewConsensus creates a Consensus with default settings.
func NewConsensus() *Consensus {
	return &Consensus{
		QuorumThreshold:          0.70,
		VetoOnCritical:           true,
		StreamResults:            true,
		DropLowOnContextPressure: true,
	}
}

// ApplySingleVeto returns CRITICAL if ANY finding is CRITICAL.
func (c *Consensus) ApplySingleVeto(findings []Finding) Severity {
	if !c.VetoOnCritical {
		return c.calculateMaxSeverity(findings)
	}

	for _, f := range findings {
		if f.Severity == SeverityCritical {
			return SeverityCritical
		}
	}

	return c.calculateMaxSeverity(findings)
}

// ApplyQuorum filters findings based on agreement threshold.
func (c *Consensus) ApplyQuorum(findings []Finding, totalAgents int) []Finding {
	if totalAgents == 0 {
		return findings
	}

	// Group findings by similarity
	groups := make(map[string][]Finding)
	for _, f := range findings {
		key := c.findingKey(f)
		groups[key] = append(groups[key], f)
	}

	var result []Finding
	for _, group := range groups {
		agreement := float64(len(group)) / float64(totalAgents)

		// CRITICAL always passes (single-veto)
		if group[0].Severity == SeverityCritical {
			result = append(result, c.mergeFindingGroup(group))
			continue
		}

		// HIGH needs quorum
		if group[0].Severity == SeverityHigh && agreement >= c.QuorumThreshold {
			result = append(result, c.mergeFindingGroup(group))
			continue
		}

		// MEDIUM needs lower quorum (50%)
		if group[0].Severity == SeverityMedium && agreement >= 0.5 {
			result = append(result, c.mergeFindingGroup(group))
			continue
		}

		// LOW passes if at least 2 agents agree
		if group[0].Severity == SeverityLow && len(group) >= 2 {
			result = append(result, c.mergeFindingGroup(group))
		}
	}

	// Sort by severity
	sort.Slice(result, func(i, j int) bool {
		return SeverityOrder[result[i].Severity] > SeverityOrder[result[j].Severity]
	})

	return result
}

// Deduplicate removes semantically similar findings.
func (c *Consensus) Deduplicate(findings []Finding) []Finding {
	if len(findings) == 0 {
		return findings
	}

	// Group by category and normalized title
	groups := make(map[string][]Finding)
	for _, f := range findings {
		key := c.findingKey(f)
		groups[key] = append(groups[key], f)
	}

	var result []Finding
	for _, group := range groups {
		result = append(result, c.mergeFindingGroup(group))
	}

	// Sort by severity (highest first)
	sort.Slice(result, func(i, j int) bool {
		return SeverityOrder[result[i].Severity] > SeverityOrder[result[j].Severity]
	})

	return result
}

// FilterByContextBudget removes LOW findings when context is constrained.
func (c *Consensus) FilterByContextBudget(findings []Finding, contextUsage float64) []Finding {
	if !c.DropLowOnContextPressure {
		return findings
	}

	// At 80% context, drop LOW and MEDIUM (check first - more restrictive)
	if contextUsage >= 0.8 {
		var result []Finding
		for _, f := range findings {
			if f.Severity == SeverityCritical || f.Severity == SeverityHigh {
				result = append(result, f)
			}
		}
		return result
	}

	// At 60% context, drop LOW
	if contextUsage >= 0.6 {
		var result []Finding
		for _, f := range findings {
			if f.Severity != SeverityLow {
				result = append(result, f)
			}
		}
		return result
	}

	return findings
}

// CheckEarlyTermination determines if we should stop early.
func (c *Consensus) CheckEarlyTermination(podResults []PodResult) (bool, string) {
	// Count CRITICAL findings across pods
	criticalCounts := make(map[string]int)

	for _, pod := range podResults {
		for _, f := range pod.Findings {
			if f.Severity == SeverityCritical {
				key := c.findingKey(f)
				criticalCounts[key]++
			}
		}
	}

	// Early termination if 3+ pods agree on a CRITICAL
	for key, count := range criticalCounts {
		if count >= 3 {
			return true, key
		}
	}

	return false, ""
}

// calculateMaxSeverity returns the highest severity in findings.
func (c *Consensus) calculateMaxSeverity(findings []Finding) Severity {
	max := SeverityPass
	for _, f := range findings {
		if SeverityOrder[f.Severity] > SeverityOrder[max] {
			max = f.Severity
		}
	}
	return max
}

// findingKey generates a key for grouping similar findings.
func (c *Consensus) findingKey(f Finding) string {
	// Combine category and normalized title
	return f.Category + ":" + f.Title
}

// mergeFindingGroup combines a group of similar findings.
func (c *Consensus) mergeFindingGroup(group []Finding) Finding {
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

	// Merge files from all findings
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

	// Merge lines
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
	if len(group) > 0 {
		totalConf := 0.0
		for _, f := range group {
			totalConf += f.Confidence
		}
		best.Confidence = totalConf / float64(len(group))
	}

	return best
}

// VerdictFromFindings calculates overall verdict from findings.
func (c *Consensus) VerdictFromFindings(findings []Finding) Severity {
	if len(findings) == 0 {
		return SeverityPass
	}

	// Single-veto: any CRITICAL = CRITICAL verdict
	if c.VetoOnCritical {
		for _, f := range findings {
			if f.Severity == SeverityCritical {
				return SeverityCritical
			}
		}
	}

	return c.calculateMaxSeverity(findings)
}
