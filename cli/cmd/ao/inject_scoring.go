package main

import "github.com/boshu2/agentops/cli/internal/search"

// scorable is the interface for items that participate in MemRL Two-Phase
// composite scoring. Bridge to search.Scorable — cmd/ao types implement
// the unexported interface, which maps to the exported search.Scorable.
type scorable interface {
	getFreshness() float64
	getUtility() float64
	getMaturity() string
	setComposite(float64)
}

func (l *learning) getFreshness() float64         { return l.FreshnessScore }
func (l *learning) getUtility() float64           { return l.Utility }
func (l *learning) getMaturity() string           { return l.Maturity }
func (l *learning) setComposite(v float64)        { l.CompositeScore = v }
func (l *learning) GetFreshness() float64         { return l.FreshnessScore }
func (l *learning) GetUtility() float64           { return l.Utility }
func (l *learning) GetMaturity() string           { return l.Maturity }
func (l *learning) SetComposite(v float64)        { l.CompositeScore = v }
func (f *knowledgeFinding) getFreshness() float64  { return f.FreshnessScore }
func (f *knowledgeFinding) getUtility() float64    { return f.Utility }
func (f *knowledgeFinding) getMaturity() string    { return "" }
func (f *knowledgeFinding) setComposite(v float64) { f.CompositeScore = v }
func (f *knowledgeFinding) GetFreshness() float64  { return f.FreshnessScore }
func (f *knowledgeFinding) GetUtility() float64    { return f.Utility }
func (f *knowledgeFinding) GetMaturity() string    { return "" }
func (f *knowledgeFinding) SetComposite(v float64) { f.CompositeScore = v }
func (p *pattern) getFreshness() float64           { return p.FreshnessScore }
func (p *pattern) getUtility() float64             { return p.Utility }
func (p *pattern) getMaturity() string             { return "" }
func (p *pattern) setComposite(v float64)          { p.CompositeScore = v }
func (p *pattern) GetFreshness() float64           { return p.FreshnessScore }
func (p *pattern) GetUtility() float64             { return p.Utility }
func (p *pattern) GetMaturity() string             { return "" }
func (p *pattern) SetComposite(v float64)          { p.CompositeScore = v }

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

// applyCompositeScoringTo bridges the cmd/ao scorable interface to search.Scorable.
func applyCompositeScoringTo(items []scorable, lambda float64) {
	exported := make([]search.Scorable, len(items))
	for i, item := range items {
		exported[i] = item.(search.Scorable)
	}
	search.ApplyCompositeScoringTo(exported, lambda)
}
