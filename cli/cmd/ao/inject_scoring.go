package main

import "math"

// scorable is the interface for items that participate in MemRL Two-Phase
// composite scoring. Both learning and pattern implement this interface.
type scorable interface {
	getFreshness() float64
	getUtility() float64
	getMaturity() string
	setComposite(float64)
}

func (l *learning) getFreshness() float64  { return l.FreshnessScore }
func (l *learning) getUtility() float64    { return l.Utility }
func (l *learning) getMaturity() string    { return l.Maturity }
func (l *learning) setComposite(v float64) { l.CompositeScore = v }
func (f *finding) getFreshness() float64   { return f.FreshnessScore }
func (f *finding) getUtility() float64     { return f.Utility }
func (f *finding) getMaturity() string     { return "" }
func (f *finding) setComposite(v float64)  { f.CompositeScore = v }
func (p *pattern) getFreshness() float64   { return p.FreshnessScore }
func (p *pattern) getUtility() float64     { return p.Utility }
func (p *pattern) getMaturity() string     { return "" }
func (p *pattern) setComposite(v float64)  { p.CompositeScore = v }

// injectMaturityWeights maps CASS maturity levels to scoring multipliers.
// Slightly softer than search.go weights (1.5/0.3) since inject has smaller pools.
var injectMaturityWeights = map[string]float64{
	"established":  1.3,
	"candidate":    1.1,
	"provisional":  1.0,
	"anti-pattern": 0.4,
}

// maturityWeight returns the scoring multiplier for a CASS maturity level.
func maturityWeight(maturity string) float64 {
	if w, found := injectMaturityWeights[maturity]; found {
		return w
	}
	return 1.0
}

// freshnessScore calculates decay-adjusted score: exp(-ageWeeks * decayRate)
// Based on knowledge decay rate δ = 0.17/week (Darr et al.)
func freshnessScore(ageWeeks float64) float64 {
	const decayRate = 0.17
	score := math.Exp(-ageWeeks * decayRate)
	// Clamp to [0.1, 1.0] - old knowledge still has some value
	if score < 0.1 {
		return 0.1
	}
	return score
}

// applyCompositeScoringTo implements MemRL Two-Phase scoring for any scorable slice.
// Score = (z_norm(freshness) + λ × z_norm(utility)) × maturityWeight
// This combines recency (Phase A) with learned utility (Phase B) and CASS maturity.
func applyCompositeScoringTo(items []scorable, lambda float64) {
	if len(items) == 0 {
		return
	}

	// With fewer than 3 items, z-normalization is statistically meaningless
	// (produces near-zero scores). Use raw weighted scores instead.
	if len(items) < 3 {
		for _, item := range items {
			score := item.getFreshness() + lambda*item.getUtility()
			item.setComposite(score * maturityWeight(item.getMaturity()))
		}
		return
	}

	// Calculate means for z-normalization
	var sumF, sumU float64
	for _, item := range items {
		sumF += item.getFreshness()
		sumU += item.getUtility()
	}
	n := float64(len(items))
	meanF := sumF / n
	meanU := sumU / n

	// Calculate standard deviations
	var varF, varU float64
	for _, item := range items {
		f := item.getFreshness()
		u := item.getUtility()
		varF += (f - meanF) * (f - meanF)
		varU += (u - meanU) * (u - meanU)
	}
	stdF := math.Sqrt(varF / n)
	stdU := math.Sqrt(varU / n)

	// Avoid division by zero - use minimum of 0.001
	if stdF < 0.001 {
		stdF = 0.001
	}
	if stdU < 0.001 {
		stdU = 0.001
	}

	// Apply z-normalization, then maturity weight
	for _, item := range items {
		zFresh := (item.getFreshness() - meanF) / stdF
		zUtility := (item.getUtility() - meanU) / stdU
		score := zFresh + lambda*zUtility
		item.setComposite(score * maturityWeight(item.getMaturity()))
	}
}
