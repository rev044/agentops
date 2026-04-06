package lifecycle

import (
	"github.com/boshu2/agentops/cli/internal/types"
)

// StatusToMaturity maps a task status string to a CASS maturity level.
func StatusToMaturity(status string) types.Maturity {
	switch status {
	case "completed":
		return types.MaturityEstablished
	case "in_progress":
		return types.MaturityCandidate
	default: // "pending"
		return types.MaturityProvisional
	}
}

// CloneStringAnyMap returns a shallow copy of input, or nil if input is empty.
func CloneStringAnyMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(input))
	for k, v := range input {
		cloned[k] = v
	}
	return cloned
}
