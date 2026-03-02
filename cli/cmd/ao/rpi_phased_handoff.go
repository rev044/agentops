package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// phaseHandoff is the structured artifact that carries context between phases.
// Written by the orchestrator after each phase completes.
// Read by buildPromptForPhase() when constructing the next phase's prompt.
type phaseHandoff struct {
	SchemaVersion int               `json:"schema_version"`
	RunID         string            `json:"run_id"`
	Phase         int               `json:"phase"`
	PhaseName     string            `json:"phase_name"`
	Status        string            `json:"status"` // completed, time_boxed, failed

	// Context for next phase
	Goal    string            `json:"goal"`
	EpicID  string            `json:"epic_id,omitempty"`
	Verdicts map[string]string `json:"verdicts"`

	// What happened
	ArtifactsProduced []string `json:"artifacts_produced"`
	DecisionsMade     []string `json:"decisions_made"`
	OpenRisks         []string `json:"open_risks"`

	// Metrics
	DurationSeconds float64 `json:"duration_seconds"`
	CostUSD         float64 `json:"cost_usd,omitempty"`
	ToolCalls       int     `json:"tool_calls"`

	// Narrative (from Claude's phase-N-summary.md, capped)
	Narrative string `json:"narrative,omitempty"`

	// Timestamp
	CompletedAt string `json:"completed_at"`
}

// writePhaseHandoff atomically writes a phase handoff to .agents/rpi/phase-N-handoff.json.
func writePhaseHandoff(cwd string, h *phaseHandoff) error {
	dir := filepath.Join(cwd, ".agents", "rpi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create rpi dir: %w", err)
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal handoff: %w", err)
	}

	target := filepath.Join(dir, fmt.Sprintf("phase-%d-handoff.json", h.Phase))
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, target); err != nil {
		// Fallback: direct write if rename fails (cross-device)
		_ = os.Remove(tmp)
		return os.WriteFile(target, data, 0o644)
	}
	return nil
}

// readPhaseHandoff reads a single phase handoff. Falls back to reading
// phase-N-summary.md if the structured handoff doesn't exist (backward compat).
func readPhaseHandoff(cwd string, phaseNum int) (*phaseHandoff, error) {
	dir := filepath.Join(cwd, ".agents", "rpi")
	jsonPath := filepath.Join(dir, fmt.Sprintf("phase-%d-handoff.json", phaseNum))

	data, err := os.ReadFile(jsonPath)
	if err == nil {
		var h phaseHandoff
		if parseErr := json.Unmarshal(data, &h); parseErr != nil {
			return nil, fmt.Errorf("parse handoff: %w", parseErr)
		}
		return &h, nil
	}

	// Fallback: construct minimal handoff from legacy summary
	return readLegacySummaryAsHandoff(cwd, phaseNum)
}

// readLegacySummaryAsHandoff reads a phase-N-summary*.md file and wraps it
// in a minimal phaseHandoff struct for backward compatibility.
func readLegacySummaryAsHandoff(cwd string, phaseNum int) (*phaseHandoff, error) {
	dir := filepath.Join(cwd, ".agents", "rpi")
	// Find phase-N-summary files (may have date suffix)
	pattern := filepath.Join(dir, fmt.Sprintf("phase-%d-summary*.md", phaseNum))
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return nil, fmt.Errorf("no handoff or summary found for phase %d", phaseNum)
	}

	// Use the most recent match
	latest := matches[len(matches)-1]
	content, err := os.ReadFile(latest)
	if err != nil {
		return nil, fmt.Errorf("read summary: %w", err)
	}

	phaseNames := map[int]string{1: "discovery", 2: "implementation", 3: "validation"}
	return &phaseHandoff{
		SchemaVersion: 1,
		Phase:         phaseNum,
		PhaseName:     phaseNames[phaseNum],
		Status:        "completed",
		Narrative:     string(content),
	}, nil
}

// readAllHandoffs reads handoffs for phases 1 through upToPhase-1.
// Missing phases are silently skipped.
func readAllHandoffs(cwd string, upToPhase int) ([]*phaseHandoff, error) {
	var handoffs []*phaseHandoff
	for i := 1; i < upToPhase; i++ {
		h, err := readPhaseHandoff(cwd, i)
		if err != nil {
			continue // skip missing phases
		}
		handoffs = append(handoffs, h)
	}
	if len(handoffs) == 0 {
		return nil, fmt.Errorf("no handoffs found for phases 1..%d", upToPhase-1)
	}
	return handoffs, nil
}

// buildHandoffContext formats handoffs for prompt injection into the next phase.
func buildHandoffContext(handoffs []*phaseHandoff) string {
	if len(handoffs) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("--- RPI Context (structured handoffs from prior phases) ---\n")

	// Use goal from latest handoff
	for i := len(handoffs) - 1; i >= 0; i-- {
		if handoffs[i].Goal != "" {
			sb.WriteString(fmt.Sprintf("Goal: %s\n\n", handoffs[i].Goal))
			break
		}
	}

	for _, h := range handoffs {
		sb.WriteString(fmt.Sprintf("[Phase %d: %s — %s (source: phase-%d-handoff.json)", h.Phase, h.PhaseName, h.Status, h.Phase))
		if h.DurationSeconds > 0 {
			sb.WriteString(fmt.Sprintf(" in %.0fs", h.DurationSeconds))
		}
		sb.WriteString("]\n")

		// Verdicts
		if len(h.Verdicts) > 0 {
			var parts []string
			for k, v := range h.Verdicts {
				parts = append(parts, fmt.Sprintf("%s %s", k, v))
			}
			sb.WriteString(fmt.Sprintf("Verdict: %s\n", strings.Join(parts, ", ")))
		}

		// Epic ID
		if h.EpicID != "" {
			sb.WriteString(fmt.Sprintf("Epic: %s\n", h.EpicID))
		}

		// Artifacts
		if len(h.ArtifactsProduced) > 0 {
			sb.WriteString(fmt.Sprintf("Artifacts: %s\n", strings.Join(h.ArtifactsProduced, ", ")))
		}

		// Decisions
		if len(h.DecisionsMade) > 0 {
			sb.WriteString(fmt.Sprintf("Decisions: %s\n", strings.Join(h.DecisionsMade, ", ")))
		}

		// Risks
		if len(h.OpenRisks) > 0 {
			sb.WriteString(fmt.Sprintf("Risks: %s\n", strings.Join(h.OpenRisks, ", ")))
		}

		// Narrative (capped at 1000 chars)
		if h.Narrative != "" {
			narrative := h.Narrative
			if len(narrative) > 1000 {
				narrative = narrative[:1000] + "..."
			}
			sb.WriteString(fmt.Sprintf("Narrative (from phase-%d-summary): %s\n", h.Phase, narrative))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// buildPhaseHandoffFromState constructs a handoff from existing state + phase result + summary.
func buildPhaseHandoffFromState(state *phasedState, phaseNum int, cwd string) *phaseHandoff {
	phaseNames := map[int]string{1: "discovery", 2: "implementation", 3: "validation"}

	h := &phaseHandoff{
		SchemaVersion: 1,
		RunID:         state.RunID,
		Phase:         phaseNum,
		PhaseName:     phaseNames[phaseNum],
		Status:        "completed",
		Goal:          state.Goal,
		EpicID:        state.EpicID,
		Verdicts:      make(map[string]string),
		CompletedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	// Copy accumulated verdicts
	for k, v := range state.Verdicts {
		h.Verdicts[k] = v
	}

	// Read phase result for metrics if available
	resultPath := filepath.Join(cwd, ".agents", "rpi", fmt.Sprintf("phase-%d-result.json", phaseNum))
	if data, err := os.ReadFile(resultPath); err == nil {
		var pr phaseResult
		if json.Unmarshal(data, &pr) == nil {
			h.DurationSeconds = pr.DurationSeconds
			h.Status = pr.Status
			if pr.Status == "" {
				h.Status = "completed"
			}
		}
	}

	// Read narrative from summary file
	summaryDir := filepath.Join(cwd, ".agents", "rpi")
	pattern := filepath.Join(summaryDir, fmt.Sprintf("phase-%d-summary*.md", phaseNum))
	if matches, err := filepath.Glob(pattern); err == nil && len(matches) > 0 {
		if content, err := os.ReadFile(matches[len(matches)-1]); err == nil {
			narrative := string(content)
			if len(narrative) > 2000 {
				narrative = narrative[:2000]
			}
			h.Narrative = narrative
		}
	}

	// Scan for artifacts produced during this phase
	h.ArtifactsProduced = discoverPhaseArtifacts(cwd, phaseNum)

	return h
}

// discoverPhaseArtifacts finds key artifacts produced during a phase.
func discoverPhaseArtifacts(cwd string, phaseNum int) []string {
	var artifacts []string
	rpiDir := filepath.Join(cwd, ".agents", "rpi")

	// Check for phase-specific artifacts
	summaryPattern := filepath.Join(rpiDir, fmt.Sprintf("phase-%d-summary*.md", phaseNum))
	if matches, _ := filepath.Glob(summaryPattern); len(matches) > 0 {
		for _, m := range matches {
			rel, _ := filepath.Rel(cwd, m)
			if rel != "" {
				artifacts = append(artifacts, rel)
			}
		}
	}

	// Phase-specific artifact discovery
	switch phaseNum {
	case 1: // discovery — look for plans and council reports
		planPattern := filepath.Join(cwd, ".agents", "plans", "*.md")
		if matches, _ := filepath.Glob(planPattern); len(matches) > 0 {
			rel, _ := filepath.Rel(cwd, matches[len(matches)-1])
			if rel != "" {
				artifacts = append(artifacts, rel)
			}
		}
		councilPattern := filepath.Join(cwd, ".agents", "council", "*pre-mortem*.md")
		if matches, _ := filepath.Glob(councilPattern); len(matches) > 0 {
			rel, _ := filepath.Rel(cwd, matches[len(matches)-1])
			if rel != "" {
				artifacts = append(artifacts, rel)
			}
		}
	case 2: // implementation — note crank artifacts
		// Crank artifacts are tracked by beads, not file scanning
	case 3: // validation — look for vibe/post-mortem reports
		vibePattern := filepath.Join(cwd, ".agents", "council", "*vibe*.md")
		if matches, _ := filepath.Glob(vibePattern); len(matches) > 0 {
			rel, _ := filepath.Rel(cwd, matches[len(matches)-1])
			if rel != "" {
				artifacts = append(artifacts, rel)
			}
		}
	}

	return artifacts
}
