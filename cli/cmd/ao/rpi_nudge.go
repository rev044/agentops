package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	rpiNudgeRunID      string
	rpiNudgePhase      int
	rpiNudgeAllWorkers bool
	rpiNudgeWorker     int
	rpiNudgeMessage    string
)

func init() {
	nudgeCmd := &cobra.Command{
		Use:   "nudge [message]",
		Short: "Send a mid-flight nudge into tmux-backed RPI phase sessions",
		Long: `Send a message to an active tmux-backed RPI phase session.

Targets may be the mayor session (default), one worker (--worker), or every
worker session for a phase (--all-workers).`,
		RunE: runRPINudge,
	}
	nudgeCmd.Flags().StringVar(&rpiNudgeRunID, "run-id", "", "Run ID to target (defaults to latest phased state)")
	nudgeCmd.Flags().IntVar(&rpiNudgePhase, "phase", 0, "Phase number to target (1-3; defaults to run's current phase)")
	nudgeCmd.Flags().BoolVar(&rpiNudgeAllWorkers, "all-workers", false, "Send nudge to every worker session for the phase")
	nudgeCmd.Flags().IntVar(&rpiNudgeWorker, "worker", 0, "Send nudge to one worker (for example: 1)")
	nudgeCmd.Flags().StringVar(&rpiNudgeMessage, "message", "", "Nudge message to send")
	rpiCmd.AddCommand(nudgeCmd)
}

type rpiNudgeRecord struct {
	Timestamp string   `json:"timestamp"`
	RunID     string   `json:"run_id"`
	Phase     int      `json:"phase"`
	Targets   []string `json:"targets"`
	Message   string   `json:"message"`
}

func runRPINudge(cmd *cobra.Command, args []string) error {
	if rpiNudgeAllWorkers && rpiNudgeWorker > 0 {
		return fmt.Errorf("use either --all-workers or --worker, not both")
	}

	message := strings.TrimSpace(rpiNudgeMessage)
	if message == "" && len(args) > 0 {
		message = strings.TrimSpace(strings.Join(args, " "))
	}
	if message == "" {
		return fmt.Errorf("provide a nudge message via --message or positional text")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	runID, state, root, err := resolveNudgeRun(cwd, strings.TrimSpace(rpiNudgeRunID))
	if err != nil {
		return err
	}

	phase, err := resolveNudgePhase(state, rpiNudgePhase)
	if err != nil {
		return err
	}

	toolchain, tcErr := resolveRPIToolchainDefaults()
	tmuxCommand := "tmux"
	if tcErr == nil && strings.TrimSpace(toolchain.TmuxCommand) != "" {
		tmuxCommand = strings.TrimSpace(toolchain.TmuxCommand)
	}
	tmuxBin, err := lookPath(tmuxCommand)
	if err != nil {
		return fmt.Errorf("tmux binary %q not found: %w", tmuxCommand, err)
	}

	phaseSession := tmuxSessionName(runID, phase)
	sessions, err := listTmuxSessions(tmuxBin)
	if err != nil {
		return fmt.Errorf("list tmux sessions: %w", err)
	}
	targets, err := resolveNudgeTargets(sessions, phaseSession, rpiNudgeAllWorkers, rpiNudgeWorker)
	if err != nil {
		return err
	}

	commandRecord, err := appendRPIC2Command(root, rpiC2CommandInput{
		RunID:    runID,
		Phase:    phase,
		Kind:     "nudge",
		Targets:  append([]string(nil), targets...),
		Message:  message,
		Deadline: time.Now().UTC().Add(30 * time.Second),
		Metadata: map[string]any{
			"phase_session": phaseSession,
			"all_workers":   rpiNudgeAllWorkers,
			"worker":        rpiNudgeWorker,
		},
	})
	if err != nil {
		return fmt.Errorf("append command log: %w", err)
	}
	commandID := commandRecord.CommandID

	for _, target := range targets {
		if err := sendTmuxNudge(tmuxBin, target, message); err != nil {
			emitErr := appendRPINudgeC2Event(root, runID, phase, commandID, target, "failed", err.Error())
			if emitErr != nil {
				VerbosePrintf("Warning: could not append failed nudge event: %v\n", emitErr)
			}
			return err
		}
		if err := appendRPINudgeC2Event(root, runID, phase, commandID, target, "ack", "nudge delivered"); err != nil {
			VerbosePrintf("Warning: could not append ack nudge event: %v\n", err)
		}
	}

	record := rpiNudgeRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		RunID:     runID,
		Phase:     phase,
		Targets:   append([]string(nil), targets...),
		Message:   message,
	}
	if err := appendRPINudgeAudit(root, runID, record); err != nil {
		VerbosePrintf("Warning: could not write nudge audit: %v\n", err)
	}

	fmt.Printf("Nudged %d session(s): %s\n", len(targets), strings.Join(targets, ", "))
	return nil
}

func appendRPINudgeC2Event(root, runID string, phase int, commandID, target, status, message string) error {
	eventType := "command.nudge." + strings.TrimSpace(status)
	_, err := appendRPIC2Event(root, rpiC2EventInput{
		RunID:     runID,
		CommandID: commandID,
		Phase:     phase,
		Backend:   "tmux",
		Source:    "rpi_nudge",
		Type:      eventType,
		Message:   message,
		Details: map[string]any{
			"target": target,
			"status": strings.TrimSpace(status),
		},
	})
	return err
}

func resolveNudgeRun(cwd, requestedRunID string) (runID string, state *phasedState, root string, err error) {
	runID = strings.TrimSpace(requestedRunID)
	if runID != "" {
		state, root, err = locateRunMetadata(cwd, runID)
		if err != nil {
			return "", nil, "", fmt.Errorf("locate run %s: %w", runID, err)
		}
		return runID, state, root, nil
	}

	state, err = loadPhasedState(cwd)
	if err != nil {
		return "", nil, "", fmt.Errorf("load latest phased state (or pass --run-id): %w", err)
	}
	runID = strings.TrimSpace(state.RunID)
	if runID == "" {
		return "", nil, "", fmt.Errorf("latest phased state has no run_id; pass --run-id")
	}
	state, root, err = locateRunMetadata(cwd, runID)
	if err != nil {
		return "", nil, "", fmt.Errorf("locate latest run %s: %w", runID, err)
	}
	return runID, state, root, nil
}

func resolveNudgePhase(state *phasedState, requestedPhase int) (int, error) {
	if requestedPhase > 0 {
		if requestedPhase < 1 || requestedPhase > 3 {
			return 0, fmt.Errorf("--phase must be 1, 2, or 3 (got %d)", requestedPhase)
		}
		return requestedPhase, nil
	}
	if state != nil && state.Phase >= 1 && state.Phase <= 3 {
		return state.Phase, nil
	}
	return 0, fmt.Errorf("could not infer phase from state; pass --phase")
}

func resolveNudgeTargets(sessions []string, phaseSession string, allWorkers bool, worker int) ([]string, error) {
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
		targets := filterTmuxWorkerSessions(sessions, phaseSession)
		sort.Strings(targets)
		if len(targets) == 0 {
			return nil, fmt.Errorf("no worker sessions found for %q", phaseSession)
		}
		return targets, nil
	}

	if has(phaseSession) {
		return []string{phaseSession}, nil
	}

	workers := filterTmuxWorkerSessions(sessions, phaseSession)
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

func appendRPINudgeAudit(root, runID string, record rpiNudgeRecord) error {
	runDir := rpiRunRegistryDir(root, runID)
	if runDir == "" {
		return nil
	}
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(runDir, "nudges.jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(record)
}
