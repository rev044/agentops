// Package bench provides pure helpers for retrieval benchmark scoring
// and corpus section analysis. Functions here have no dependencies on
// the CLI command layer and can be reused/tested in isolation.
package bench

import (
	"strings"
)

// NormalizeSplit lowercases and trims a benchmark split name.
func NormalizeSplit(split string) string {
	return strings.ToLower(strings.TrimSpace(split))
}

// NormalizeSection normalizes a markdown section heading for comparison.
func NormalizeSection(section string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimLeft(section, "#")))
}

// StripFrontMatter removes a YAML front matter block from markdown content.
func StripFrontMatter(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return content
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[i+1:], "\n")
		}
	}
	return content
}

// SectionHeading returns the first markdown heading text found in a section.
func SectionHeading(section string) string {
	lines := strings.Split(section, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			return strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		}
	}
	return ""
}

// ScoreResults computes Precision@K and MRR given a slice of result IDs,
// the set of expected IDs, a single "best" ID for MRR, and K.
func ScoreResults(resultIDs []string, expectedSet map[string]bool, bestID string, k int) (float64, float64) {
	if k <= 0 || len(resultIDs) == 0 {
		return 0, 0
	}
	n := k
	if n > len(resultIDs) {
		n = len(resultIDs)
	}
	hits := 0
	for _, id := range resultIDs[:n] {
		if expectedSet[id] {
			hits++
		}
	}
	pAtK := float64(hits) / float64(k)

	mrr := 0.0
	for i, id := range resultIDs {
		if id == bestID {
			mrr = 1.0 / float64(i+1)
			break
		}
	}
	return pAtK, mrr
}
