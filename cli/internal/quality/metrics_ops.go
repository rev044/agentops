package quality

import (
	"math"
)

// UtilityStats holds computed utility statistics.
type UtilityStats struct {
	Mean      float64
	StdDev    float64
	HighCount int // utility > 0.7
	LowCount  int // utility < 0.3
}

// ComputeUtilityStats calculates statistics from a slice of utility values.
func ComputeUtilityStats(utilities []float64) UtilityStats {
	var stats UtilityStats
	if len(utilities) == 0 {
		return stats
	}
	var sum float64
	for _, u := range utilities {
		sum += u
	}
	stats.Mean = sum / float64(len(utilities))
	var variance float64
	for _, u := range utilities {
		variance += (u - stats.Mean) * (u - stats.Mean)
	}
	stats.StdDev = math.Sqrt(variance / float64(len(utilities)))
	for _, u := range utilities {
		if u > 0.7 {
			stats.HighCount++
		}
		if u < 0.3 {
			stats.LowCount++
		}
	}
	return stats
}

// ComputeOperationalSigmaRho calculates retrieval coverage (sigma) and
// evidence-backed influence rate (rho) from artifact and citation counts.
func ComputeOperationalSigmaRho(totalArtifacts, uniqueCited, evidenceBacked int) (sigma, rho float64) {
	if totalArtifacts > 0 {
		sigma = float64(uniqueCited) / float64(totalArtifacts)
		if sigma > 1.0 {
			sigma = 1.0
		}
	}
	if uniqueCited > 0 {
		rho = float64(evidenceBacked) / float64(uniqueCited)
		if rho > 1.0 {
			rho = 1.0
		}
	}
	return sigma, rho
}

// EscapeVelocityThreshold returns delta/100, the operational threshold above
// which sigma*rho indicates the knowledge flywheel is compounding.
func EscapeVelocityThreshold(delta float64) float64 {
	if delta <= 0 {
		return 0
	}
	return delta / 100.0
}
