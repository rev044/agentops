package main

import "testing"

func TestMaturityWeight_KnownLevels(t *testing.T) {
	tests := []struct {
		maturity string
		want     float64
	}{
		{"established", 1.3},
		{"candidate", 1.1},
		{"provisional", 1.0},
		{"anti-pattern", 0.4},
	}
	for _, tt := range tests {
		got := maturityWeight(tt.maturity)
		if got != tt.want {
			t.Errorf("maturityWeight(%q) = %v, want %v", tt.maturity, got, tt.want)
		}
	}
}

func TestMaturityWeight_Unknown(t *testing.T) {
	tests := []string{"", "unknown", "experimental"}
	for _, m := range tests {
		got := maturityWeight(m)
		if got != 1.0 {
			t.Errorf("maturityWeight(%q) = %v, want 1.0", m, got)
		}
	}
}

func TestApplyCompositeScoring_WithMaturity(t *testing.T) {
	// Two learnings with identical freshness/utility but different maturity.
	// Established should outrank provisional.
	established := &learning{
		FreshnessScore: 0.9,
		Utility:        0.7,
		Maturity:       "established",
	}
	provisional := &learning{
		FreshnessScore: 0.9,
		Utility:        0.7,
		Maturity:       "provisional",
	}
	antiPattern := &learning{
		FreshnessScore: 0.9,
		Utility:        0.7,
		Maturity:       "anti-pattern",
	}

	items := []scorable{established, provisional, antiPattern}
	applyCompositeScoringTo(items, 0.5)

	if established.CompositeScore <= provisional.CompositeScore {
		t.Errorf("established (%v) should outrank provisional (%v)",
			established.CompositeScore, provisional.CompositeScore)
	}
	if provisional.CompositeScore <= antiPattern.CompositeScore {
		t.Errorf("provisional (%v) should outrank anti-pattern (%v)",
			provisional.CompositeScore, antiPattern.CompositeScore)
	}
}

func TestKnowledgeFindingScorableInterface(t *testing.T) {
	f := &knowledgeFinding{
		FreshnessScore: 0.75,
		Utility:        0.6,
	}
	if got := f.getFreshness(); got != 0.75 {
		t.Errorf("getFreshness() = %v, want 0.75", got)
	}
	if got := f.getUtility(); got != 0.6 {
		t.Errorf("getUtility() = %v, want 0.6", got)
	}
	if got := f.getMaturity(); got != "" {
		t.Errorf("getMaturity() = %q, want empty", got)
	}
	f.setComposite(1.5)
	if f.CompositeScore != 1.5 {
		t.Errorf("setComposite(1.5) -> CompositeScore = %v, want 1.5", f.CompositeScore)
	}
}

func TestApplyCompositeScoring_NoMaturity(t *testing.T) {
	// Patterns don't have maturity — should get weight 1.0 (no change).
	p1 := &pattern{FreshnessScore: 0.8, Utility: 0.6}
	p2 := &pattern{FreshnessScore: 0.5, Utility: 0.4}
	p3 := &pattern{FreshnessScore: 0.3, Utility: 0.2}

	items := []scorable{p1, p2, p3}
	applyCompositeScoringTo(items, 0.5)

	// Fresher + higher utility should rank higher
	if p1.CompositeScore <= p2.CompositeScore {
		t.Errorf("p1 (%v) should outrank p2 (%v)", p1.CompositeScore, p2.CompositeScore)
	}
	if p2.CompositeScore <= p3.CompositeScore {
		t.Errorf("p2 (%v) should outrank p3 (%v)", p2.CompositeScore, p3.CompositeScore)
	}
}
