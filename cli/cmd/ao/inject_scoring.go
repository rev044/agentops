package main

import "github.com/boshu2/agentops/cli/internal/search"

// scorable is a package-level alias for search.Scorable — kept for test compat.
type scorable = search.Scorable

// injectMaturityWeights kept as package-level alias for test compatibility.
var injectMaturityWeights = search.InjectMaturityWeights

// Thin wrappers — delegate to search package, kept for test compatibility.
func maturityWeight(maturity string) float64 { return search.MaturityWeight(maturity) }
func freshnessScore(ageWeeks float64) float64 { return search.FreshnessScore(ageWeeks) }
func tokenizeWords(text string) []string      { return search.TokenizeWords(text) }

func computeAdjacency(tokens []string, words []string) float64 {
	return search.ComputeAdjacency(tokens, words)
}

func weightedSectionScore(tokens []string, heading, content string, sectionIndex, totalSections int) float64 {
	return search.WeightedSectionScore(tokens, heading, content, sectionIndex, totalSections)
}

// Scoring weight constants — kept for test compatibility.
const (
	wSubstringCoverage = search.WSubstringCoverage
	wHeadingMatch      = search.WHeadingMatch
	wExactToken        = search.WExactToken
	wAdjacency         = search.WAdjacency
	wSectionProximity  = search.WSectionProximity
)

// applyCompositeScoringTo delegates to search.ApplyCompositeScoringTo.
func applyCompositeScoringTo(items []search.Scorable, lambda float64) {
	search.ApplyCompositeScoringTo(items, lambda)
}
