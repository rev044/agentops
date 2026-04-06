package rpi

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Finding represents a single evaluator finding with description, fix, and reference.
type Finding struct {
	Description string `json:"description"`
	Fix         string `json:"fix,omitempty"`
	Ref         string `json:"ref,omitempty"`
}

// PhaseNameForNumber maps a 1-based phase number to its canonical name.
func PhaseNameForNumber(phaseNum int) string {
	switch phaseNum {
	case 1:
		return "discovery"
	case 2:
		return "implementation"
	case 3:
		return "validation"
	default:
		return fmt.Sprintf("phase-%d", phaseNum)
	}
}

// PhaseEvaluatorVerdict computes the evaluator verdict from gate verdict,
// transcript reward, tracker mode, and phase number.
func PhaseEvaluatorVerdict(phaseNum int, trackerMode, gateVerdict string, hasTranscript bool, reward float64) string {
	normalized := strings.ToUpper(strings.TrimSpace(gateVerdict))
	switch normalized {
	case "FAIL", "BLOCKED":
		return "FAIL"
	}
	if hasTranscript && reward < 0.25 {
		return "FAIL"
	}
	if normalized == "WARN" || normalized == "PARTIAL" || normalized == "SKIP" {
		return "WARN"
	}
	if phaseNum == 1 && trackerMode == "tasklist" {
		return "WARN"
	}
	if hasTranscript && reward < 0.55 {
		return "WARN"
	}
	return "PASS"
}

// PhaseEvaluatorSummary builds a human-readable summary line for a phase evaluator.
func PhaseEvaluatorSummary(phaseNum int, trackerMode, gateVerdict string, hasTranscript bool, reward float64, findingCount int) string {
	verdict := PhaseEvaluatorVerdict(phaseNum, trackerMode, gateVerdict, hasTranscript, reward)
	parts := []string{
		fmt.Sprintf("%s evaluator marked the phase %s", PhaseNameForNumber(phaseNum), verdict),
	}
	if gate := strings.ToUpper(strings.TrimSpace(gateVerdict)); gate != "" {
		parts = append(parts, fmt.Sprintf("gate=%s", gate))
	}
	if hasTranscript {
		parts = append(parts, fmt.Sprintf("reward=%.2f", reward))
	}
	if findingCount > 0 {
		parts = append(parts, fmt.Sprintf("findings=%d", findingCount))
	}
	if phaseNum == 1 && trackerMode == "tasklist" {
		parts = append(parts, "tracker degraded -> tasklist fallback")
	}
	return strings.Join(parts, " · ")
}

// DefaultEvaluatorFindings produces the standard findings based on gate verdict,
// tracker mode, and transcript reward.
func DefaultEvaluatorFindings(phaseNum int, trackerMode, gateVerdict string, hasTranscript bool, reward float64, transcriptPath string, evidenceRef string) []Finding {
	var findings []Finding

	switch strings.ToUpper(strings.TrimSpace(gateVerdict)) {
	case "BLOCKED":
		findings = append(findings, Finding{
			Description: "Implementation phase ended blocked",
			Fix:         "Unblock the remaining execution path before advancing validation.",
			Ref:         evidenceRef,
		})
	case "PARTIAL":
		findings = append(findings, Finding{
			Description: "Implementation phase ended partial",
			Fix:         "Complete the remaining execution work before validation claims success.",
			Ref:         evidenceRef,
		})
	case "FAIL":
		findings = append(findings, Finding{
			Description: fmt.Sprintf("%s gate returned FAIL", PhaseNameForNumber(phaseNum)),
			Fix:         "Resolve the failing report findings and rerun the phase gate.",
			Ref:         evidenceRef,
		})
	}

	if phaseNum == 1 && trackerMode == "tasklist" {
		findings = append(findings, Finding{
			Description: "Tracker degraded during discovery",
			Fix:         "Use the execution packet and plan artifact as the objective spine until tracker health is restored.",
			Ref:         evidenceRef,
		})
	}

	if hasTranscript {
		switch {
		case reward < 0.25:
			findings = append(findings, Finding{
				Description: fmt.Sprintf("Transcript-derived reward %.2f indicates a failing session outcome", reward),
				Fix:         "Inspect the transcript signals and resolve the failing test/error/push conditions before retrying.",
				Ref:         transcriptPath,
			})
		case reward < 0.55:
			findings = append(findings, Finding{
				Description: fmt.Sprintf("Transcript-derived reward %.2f indicates weak completion quality", reward),
				Fix:         "Tighten verification and closeout before treating the phase as complete.",
				Ref:         transcriptPath,
			})
		}
	}

	return findings
}

// UniqueFindings deduplicates findings by (Description, Fix, Ref), dropping empty entries.
func UniqueFindings(items []Finding) []Finding {
	seen := make(map[string]struct{}, len(items))
	out := make([]Finding, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Description) + "\x00" + strings.TrimSpace(item.Fix) + "\x00" + strings.TrimSpace(item.Ref)
		if strings.TrimSpace(item.Description) == "" && strings.TrimSpace(item.Fix) == "" && strings.TrimSpace(item.Ref) == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

// SessionIDFromEventDetails extracts a session_id from a JSON details blob.
func SessionIDFromEventDetails(details json.RawMessage) string {
	if len(details) == 0 {
		return ""
	}
	var d map[string]any
	if err := json.Unmarshal(details, &d); err != nil {
		return ""
	}
	if raw, ok := d["session_id"].(string); ok {
		return strings.TrimSpace(raw)
	}
	return ""
}
