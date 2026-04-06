package search

// InjectOptions holds all runtime configuration for an `ao inject` invocation.
// It replaces the prior constellation of package-level flag vars in cmd/ao.
type InjectOptions struct {
	MaxTokens         int
	Context           string
	Format            string
	SessionID         string
	NoCite            bool
	ApplyDecay        bool
	Bead              string
	Predecessor       string
	IndexOnly         bool
	QuarantineFlagged bool
	ForSkill          string
	SessionType       string
	Profile           bool

	// Query is the resolved query (positional arg or Context fallback).
	Query string
}
