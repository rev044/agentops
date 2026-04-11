package rpi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// NudgeRecord holds the audit trail for a single nudge operation.
type NudgeRecord struct {
	Timestamp string   `json:"timestamp"`
	RunID     string   `json:"run_id"`
	Phase     int      `json:"phase"`
	Targets   []string `json:"targets"`
	Message   string   `json:"message"`
}

// ResolveNudgePhase resolves the target phase for a nudge operation.
// If requestedPhase is positive, it validates and returns it.
// Otherwise it infers from the state's current phase.
func ResolveNudgePhase(statePhase int, requestedPhase int) (int, error) {
	if requestedPhase > 0 {
		if requestedPhase < 1 || requestedPhase > 3 {
			return 0, fmt.Errorf("--phase must be 1, 2, or 3 (got %d)", requestedPhase)
		}
		return requestedPhase, nil
	}
	if statePhase >= 1 && statePhase <= 3 {
		return statePhase, nil
	}
	return 0, fmt.Errorf("could not infer phase from state; pass --phase")
}

// ResolveNudgeTargets identifies which tmux sessions to target for a nudge.
func ResolveNudgeTargets(sessions []string, phaseSession string, allWorkers bool, worker int) ([]string, error) {
	sessionSet := make(map[string]struct{}, len(sessions))
	for _, s := range sessions {
		sessionSet[s] = struct{}{}
	}
	has := func(name string) bool {
		_, ok := sessionSet[name]
		return ok
	}

	if worker > 0 {
		target := fmt.Sprintf("%s-w%d", phaseSession, worker)
		if !has(target) {
			return nil, fmt.Errorf("worker session %q not found", target)
		}
		return []string{target}, nil
	}

	if allWorkers {
		targets := FilterTmuxWorkerSessions(sessions, phaseSession)
		sort.Strings(targets)
		if len(targets) == 0 {
			return nil, fmt.Errorf("no worker sessions found for %q", phaseSession)
		}
		return targets, nil
	}

	if has(phaseSession) {
		return []string{phaseSession}, nil
	}

	workers := FilterTmuxWorkerSessions(sessions, phaseSession)
	sort.Strings(workers)
	switch len(workers) {
	case 0:
		return nil, fmt.Errorf("no tmux session found for %q", phaseSession)
	case 1:
		return workers, nil
	default:
		return nil, fmt.Errorf("multiple worker sessions found for %q; use --all-workers or --worker", phaseSession)
	}
}

// FilterTmuxWorkerSessions returns sessions matching "<prefix>-w<N>" pattern.
func FilterTmuxWorkerSessions(sessions []string, prefix string) []string {
	var result []string
	workerPrefix := prefix + "-w"
	for _, s := range sessions {
		if strings.HasPrefix(s, workerPrefix) {
			result = append(result, s)
		}
	}
	return result
}

// AppendNudgeAudit writes a nudge record to the run's nudges.jsonl file.
func AppendNudgeAudit(runDir string, record NudgeRecord) error {
	if runDir == "" {
		return nil
	}
	if err := os.MkdirAll(runDir, 0o750); err != nil {
		return err
	}
	path := filepath.Join(runDir, "nudges.jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(file).Encode(record); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
