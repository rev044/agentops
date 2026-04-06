package lifecycle

import "math"

// AnnealedAlpha computes a decaying learning rate for the MemRL feedback loop.
// It starts at 3*baseAlpha for new learnings and decays toward baseAlpha/10 as
// citation count grows.
func AnnealedAlpha(baseAlpha float64, citationCount int) float64 {
	alpha := (baseAlpha * 3.0) * math.Exp(-float64(citationCount)*0.1)
	floor := baseAlpha / 10.0
	if alpha < floor {
		return floor
	}
	return alpha
}
