package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type phaseSessionOutcome struct {
	SessionID      string   `json:"session_id"`
	TranscriptPath string   `json:"transcript_path,omitempty"`
	Reward         float64  `json:"reward,omitempty"`
	Signals        []Signal `json:"signals,omitempty"`
	AnalyzedAt     string   `json:"analyzed_at,omitempty"`
}

type phaseEvaluatorArtifact struct {
	SchemaVersion  int                  `json:"schema_version"`
	RunID          string               `json:"run_id"`
	Phase          int                  `json:"phase"`
	PhaseName      string               `json:"phase_name"`
	GateVerdict    string               `json:"gate_verdict,omitempty"`
	Verdict        string               `json:"verdict"`
	Summary        string               `json:"summary"`
	Findings       []finding            `json:"findings,omitempty"`
	Evidence       []string             `json:"evidence,omitempty"`
	SessionOutcome *phaseSessionOutcome `json:"session_outcome,omitempty"`
	GeneratedAt    string               `json:"generated_at"`
}

func emitPhaseEvaluatorArtifact(cwd string, state *phasedState, phaseNum int, gateVerdict string, gateFindings []finding, evidence ...string) (*phaseEvaluatorArtifact, error) {
	if state == nil {
		return nil, fmt.Errorf("state is required")
	}

	outcome := collectPhaseSessionOutcome(cwd, state.RunID, phaseNum)
	evidence = collectPhaseEvaluatorEvidence(cwd, phaseNum, evidence...)
	findings := append([]finding{}, gateFindings...)
	findings = append(findings, defaultEvaluatorFindings(phaseNum, state, gateVerdict, outcome, evidence)...)
	findings = uniqueFindings(findings)

	artifact := &phaseEvaluatorArtifact{
		SchemaVersion:  1,
		RunID:          state.RunID,
		Phase:          phaseNum,
		PhaseName:      phaseNameForNumber(phaseNum),
		GateVerdict:    strings.ToUpper(strings.TrimSpace(gateVerdict)),
		Verdict:        phaseEvaluatorVerdict(phaseNum, state, gateVerdict, outcome),
		Summary:        phaseEvaluatorSummary(phaseNum, state, gateVerdict, outcome, findings),
		Findings:       findings,
		Evidence:       evidence,
		SessionOutcome: outcome,
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	if err := writePhaseEvaluatorArtifact(cwd, artifact); err != nil {
		return nil, err
	}

	if state.Verdicts == nil {
		state.Verdicts = make(map[string]string)
	}
	state.Verdicts[fmt.Sprintf("%s_evaluator", artifact.PhaseName)] = artifact.Verdict
	if _, err := appendRPIC2Event(cwd, rpiC2EventInput{
		RunID:   state.RunID,
		Phase:   phaseNum,
		Backend: state.Backend,
		Source:  "orchestrator",
		Type:    "phase.evaluator.updated",
		Message: fmt.Sprintf("Phase %d evaluator verdict: %s", phaseNum, artifact.Verdict),
		Details: map[string]any{
			"verdict":      artifact.Verdict,
			"gate_verdict": artifact.GateVerdict,
			"artifact":     filepath.ToSlash(filepath.Join(".agents", "rpi", fmt.Sprintf(phaseEvaluatorFileFmt, phaseNum))),
		},
	}); err != nil {
		VerbosePrintf("Warning: could not emit phase.evaluator.updated event: %v\n", err)
	}

	return artifact, nil
}

func writePhaseEvaluatorArtifact(cwd string, artifact *phaseEvaluatorArtifact) error {
	stateDir := filepath.Join(cwd, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0o750); err != nil {
		return fmt.Errorf("create evaluator directory: %w", err)
	}

	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal evaluator artifact: %w", err)
	}
	data = append(data, '\n')

	path := filepath.Join(stateDir, fmt.Sprintf(phaseEvaluatorFileFmt, artifact.Phase))
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write evaluator tmp: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename evaluator artifact: %w", err)
	}
	return nil
}

func collectPhaseEvaluatorEvidence(cwd string, phaseNum int, extra ...string) []string {
	var evidence []string
	evidence = append(evidence, latestRelativeArtifact(cwd, filepath.Join(".agents", "rpi", fmt.Sprintf(phaseResultFileFmt, phaseNum))))
	evidence = append(evidence, latestRelativeArtifact(cwd, filepath.Join(".agents", "rpi", fmt.Sprintf("phase-%d-summary*.md", phaseNum))))
	evidence = append(evidence, latestRelativeArtifact(cwd, filepath.Join(".agents", "rpi", "execution-packet.json")))
	for _, item := range extra {
		item = pathClean(item)
		if !isSafeArtifactRelPath(item) {
			continue
		}
		full := filepath.Join(cwd, filepath.FromSlash(item))
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			evidence = append(evidence, item)
		}
	}
	return uniqueStringsPreserveOrder(evidence)
}

func collectPhaseSessionOutcome(cwd, runID string, phaseNum int) *phaseSessionOutcome {
	sessionID := latestPhaseSessionID(cwd, runID, phaseNum)
	if sessionID == "" {
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return &phaseSessionOutcome{SessionID: sessionID}
	}
	transcriptPath := findTranscriptForSession(filepath.Join(homeDir, ".claude", "projects"), sessionID)
	if transcriptPath == "" {
		return &phaseSessionOutcome{SessionID: sessionID}
	}

	outcome, err := analyzeTranscript(transcriptPath, sessionID)
	if err != nil {
		return &phaseSessionOutcome{
			SessionID:      sessionID,
			TranscriptPath: transcriptPath,
		}
	}
	return &phaseSessionOutcome{
		SessionID:      outcome.SessionID,
		TranscriptPath: transcriptPath,
		Reward:         outcome.Reward,
		Signals:        outcome.Signals,
		AnalyzedAt:     outcome.AnalyzedAt.UTC().Format(time.RFC3339),
	}
}

func latestPhaseSessionID(cwd, runID string, phaseNum int) string {
	if strings.TrimSpace(runID) == "" {
		return ""
	}
	events, err := loadRPIC2Events(cwd, runID)
	if err != nil {
		return ""
	}
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Phase != phaseNum {
			continue
		}
		if sessionID := sessionIDFromRPIEvent(events[i]); sessionID != "" {
			return sessionID
		}
	}
	return ""
}

func sessionIDFromRPIEvent(ev RPIC2Event) string {
	if len(ev.Details) == 0 {
		return ""
	}
	var details map[string]any
	if err := json.Unmarshal(ev.Details, &details); err != nil {
		return ""
	}
	if raw, ok := details["session_id"].(string); ok {
		return strings.TrimSpace(raw)
	}
	return ""
}

func phaseEvaluatorVerdict(phaseNum int, state *phasedState, gateVerdict string, outcome *phaseSessionOutcome) string {
	normalized := strings.ToUpper(strings.TrimSpace(gateVerdict))
	switch normalized {
	case "FAIL", "BLOCKED":
		return "FAIL"
	}
	if outcome != nil && outcome.TranscriptPath != "" && outcome.Reward < 0.25 {
		return "FAIL"
	}
	if normalized == "WARN" || normalized == "PARTIAL" || normalized == "SKIP" {
		return "WARN"
	}
	if phaseNum == 1 && state != nil && state.TrackerMode == "tasklist" {
		return "WARN"
	}
	if outcome != nil && outcome.TranscriptPath != "" && outcome.Reward < 0.55 {
		return "WARN"
	}
	return "PASS"
}

func phaseEvaluatorSummary(phaseNum int, state *phasedState, gateVerdict string, outcome *phaseSessionOutcome, findings []finding) string {
	parts := []string{
		fmt.Sprintf("%s evaluator marked the phase %s", phaseNameForNumber(phaseNum), phaseEvaluatorVerdict(phaseNum, state, gateVerdict, outcome)),
	}
	if gate := strings.ToUpper(strings.TrimSpace(gateVerdict)); gate != "" {
		parts = append(parts, fmt.Sprintf("gate=%s", gate))
	}
	if outcome != nil && outcome.TranscriptPath != "" {
		parts = append(parts, fmt.Sprintf("reward=%.2f", outcome.Reward))
	}
	if len(findings) > 0 {
		parts = append(parts, fmt.Sprintf("findings=%d", len(findings)))
	}
	if phaseNum == 1 && state != nil && state.TrackerMode == "tasklist" {
		parts = append(parts, "tracker degraded -> tasklist fallback")
	}
	return strings.Join(parts, " · ")
}

func defaultEvaluatorFindings(phaseNum int, state *phasedState, gateVerdict string, outcome *phaseSessionOutcome, evidence []string) []finding {
	var findings []finding
	ref := ""
	if len(evidence) > 0 {
		ref = evidence[0]
	}

	switch strings.ToUpper(strings.TrimSpace(gateVerdict)) {
	case "BLOCKED":
		findings = append(findings, finding{
			Description: "Implementation phase ended blocked",
			Fix:         "Unblock the remaining execution path before advancing validation.",
			Ref:         ref,
		})
	case "PARTIAL":
		findings = append(findings, finding{
			Description: "Implementation phase ended partial",
			Fix:         "Complete the remaining execution work before validation claims success.",
			Ref:         ref,
		})
	case "FAIL":
		findings = append(findings, finding{
			Description: fmt.Sprintf("%s gate returned FAIL", phaseNameForNumber(phaseNum)),
			Fix:         "Resolve the failing report findings and rerun the phase gate.",
			Ref:         ref,
		})
	}

	if phaseNum == 1 && state != nil && state.TrackerMode == "tasklist" {
		findings = append(findings, finding{
			Description: "Tracker degraded during discovery",
			Fix:         "Use the execution packet and plan artifact as the objective spine until tracker health is restored.",
			Ref:         ref,
		})
	}

	if outcome != nil && outcome.TranscriptPath != "" {
		switch {
		case outcome.Reward < 0.25:
			findings = append(findings, finding{
				Description: fmt.Sprintf("Transcript-derived reward %.2f indicates a failing session outcome", outcome.Reward),
				Fix:         "Inspect the transcript signals and resolve the failing test/error/push conditions before retrying.",
				Ref:         outcome.TranscriptPath,
			})
		case outcome.Reward < 0.55:
			findings = append(findings, finding{
				Description: fmt.Sprintf("Transcript-derived reward %.2f indicates weak completion quality", outcome.Reward),
				Fix:         "Tighten verification and closeout before treating the phase as complete.",
				Ref:         outcome.TranscriptPath,
			})
		}
	}

	return findings
}

func uniqueFindings(items []finding) []finding {
	seen := make(map[string]struct{}, len(items))
	out := make([]finding, 0, len(items))
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

func phaseNameForNumber(phaseNum int) string {
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
