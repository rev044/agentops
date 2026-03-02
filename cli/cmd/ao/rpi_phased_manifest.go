package main

// phaseManifest declares what context a phase needs from prior handoffs.
// This enables least-privilege loading — each phase gets only what it needs.
type phaseManifest struct {
	Phase         int      `json:"phase"`
	HandoffFields []string `json:"handoff_fields"` // which phaseHandoff fields to include
	NarrativeCap  int      `json:"narrative_cap"`  // 0 = omit narrative entirely
	MaxTokens     int      `json:"max_tokens"`     // total token budget (0 = unlimited)
}

// defaultPhaseManifests returns the built-in manifest for each phase.
// Phase 1 has no prior context. Phase 2 gets decisions/risks. Phase 3 gets artifacts.
var defaultPhaseManifests = map[int]phaseManifest{
	1: {Phase: 1, HandoffFields: nil, NarrativeCap: 0, MaxTokens: 0},
	2: {Phase: 2, HandoffFields: []string{"goal", "epic_id", "verdicts", "decisions_made", "open_risks"}, NarrativeCap: 500, MaxTokens: 2500},
	3: {Phase: 3, HandoffFields: []string{"goal", "epic_id", "verdicts", "artifacts_produced"}, NarrativeCap: 1000, MaxTokens: 2500},
}

