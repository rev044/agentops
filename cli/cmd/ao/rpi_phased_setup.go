package main

import (
	"cmp"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func preflightRuntimeAvailability(runtimeCommand string) error {
	if GetDryRun() {
		return nil
	}
	command := cmp.Or(strings.TrimSpace(runtimeCommand), "claude")
	executable, _ := splitRuntimeCommand(command)
	if executable == "" {
		return fmt.Errorf("runtime command %q is empty", command)
	}
	if _, err := lookPath(executable); err != nil {
		return fmt.Errorf("runtime executable %q (from %q) not found on PATH (required for spawning phase sessions)", executable, command)
	}
	return nil
}

func resolveGoalAndStartPhase(opts phasedEngineOptions, args []string, cwd string) (string, int, error) {
	goal := ""
	if len(args) > 0 {
		goal = args[0]
	}

	startPhase := phaseNameToNum(opts.From)
	if startPhase == 0 {
		return "", 0, fmt.Errorf("unknown phase: %q (valid: discovery, implementation, validation)", opts.From)
	}

	if startPhase >= 2 && goal == "" {
		existing, err := loadPhasedState(cwd)
		if err == nil && existing.EpicID != "" {
			goal = existing.Goal
		}
	}
	if goal == "" && startPhase <= 1 {
		return "", 0, fmt.Errorf("goal is required (provide as argument)")
	}

	return goal, startPhase, nil
}

func newPhasedState(opts phasedEngineOptions, startPhase int, goal string) *phasedState {
	s := &phasedState{
		SchemaVersion: 1,
		Goal:          goal,
		Phase:         startPhase,
		StartPhase:    startPhase,
		Cycle:         1,
		FastPath:      opts.FastPath,
		TestFirst:     opts.TestFirst,
		SwarmFirst:    opts.SwarmFirst,
		Verdicts:      make(map[string]string),
		Attempts:      make(map[string]int),
		StartedAt:     time.Now().Format(time.RFC3339),
		Opts:          opts,
	}
	if opts.RunID != "" {
		s.RunID = opts.RunID
	}
	return s
}

// mergeExistingStateFields copies relevant fields from a previous run's state
// into the current state for phase resumption.
func mergeExistingStateFields(state *phasedState, existing *phasedState, opts phasedEngineOptions, goal string) {
	state.EpicID = existing.EpicID
	state.FastPath = existing.FastPath || opts.FastPath
	state.SwarmFirst = existing.SwarmFirst || opts.SwarmFirst
	if existing.Verdicts != nil {
		state.Verdicts = existing.Verdicts
	}
	if existing.Attempts != nil {
		state.Attempts = existing.Attempts
	}
	if goal == "" {
		state.Goal = existing.Goal
	}
}

// resolveExistingWorktree checks whether the previous run's worktree still exists
// and returns the worktree path to use as spawnCwd.  Returns ("", nil) when no
// worktree should be used (NoWorktree flag or no prior worktree path).
func resolveExistingWorktree(state *phasedState, existing *phasedState, opts phasedEngineOptions) (string, error) {
	if opts.NoWorktree || existing.WorktreePath == "" {
		return "", nil
	}
	if _, statErr := os.Stat(existing.WorktreePath); statErr != nil {
		return "", fmt.Errorf("worktree %s from previous run no longer exists (was it removed?)", existing.WorktreePath)
	}
	state.WorktreePath = existing.WorktreePath
	state.RunID = existing.RunID
	fmt.Printf("Resuming in existing worktree: %s\n", existing.WorktreePath)
	return existing.WorktreePath, nil
}

func resumePhasedStateIfNeeded(cwd string, opts phasedEngineOptions, startPhase int, goal string, state *phasedState) (string, error) {
	if startPhase <= 1 {
		return cwd, nil
	}

	existing, err := loadPhasedState(cwd)
	if err != nil {
		return cwd, nil
	}

	mergeExistingStateFields(state, existing, opts, goal)

	wtPath, err := resolveExistingWorktree(state, existing, opts)
	if err != nil {
		return "", err
	}
	if wtPath != "" {
		return wtPath, nil
	}
	return cwd, nil
}

func setupWorktreeLifecycle(cwd, originalCwd string, opts phasedEngineOptions, state *phasedState) (string, func(success bool, logPath string) error, error) {
	spawnCwd := cwd
	noopCleanup := func(bool, string) error { return nil }

	if opts.NoWorktree || GetDryRun() || state.WorktreePath != "" {
		return spawnCwd, noopCleanup, nil
	}

	worktreePath, runID, err := createWorktree(cwd)
	if err != nil {
		return "", noopCleanup, fmt.Errorf("create worktree: %w", err)
	}

	spawnCwd = worktreePath
	state.WorktreePath = worktreePath
	if state.RunID == "" {
		state.RunID = runID
	}
	fmt.Printf("Worktree created: %s (detached)\n", worktreePath)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if sig, ok := <-sigCh; ok {
			fmt.Fprintf(os.Stderr, "\nInterrupted (%v). Worktree preserved at: %s\n", sig, worktreePath)
			// Write terminal metadata so `ao rpi status` shows "interrupted" instead of "running".
			state.TerminalStatus = "interrupted"
			state.TerminalReason = fmt.Sprintf("signal: %v", sig)
			state.TerminatedAt = time.Now().Format(time.RFC3339)
			_ = savePhasedState(spawnCwd, state)
			os.Exit(1)
		}
	}()

	cleanup := func(success bool, logPath string) error {
		signal.Stop(sigCh)
		close(sigCh)

		if !success {
			fmt.Fprintf(os.Stderr, "Worktree preserved for debugging: %s\n", worktreePath)
			return nil
		}

		if mergeErr := mergeWorktree(originalCwd, worktreePath, runID); mergeErr != nil {
			fmt.Fprintf(os.Stderr, "Merge failed: %v\nWorktree preserved at: %s\n", mergeErr, worktreePath)
			return fmt.Errorf("worktree merge failed: %w", mergeErr)
		}
		if rmErr := removeWorktree(originalCwd, worktreePath, runID); rmErr != nil {
			fmt.Fprintf(os.Stderr, "Cleanup failed: %v\nWorktree may require manual removal: %s\n", rmErr, worktreePath)
			logFailureContext(logPath, state.RunID, "cleanup", rmErr)
			return fmt.Errorf("worktree cleanup failed: %w", rmErr)
		}
		return nil
	}

	return spawnCwd, cleanup, nil
}

func ensureStateRunID(state *phasedState) {
	if state.RunID != "" {
		return
	}
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	state.RunID = hex.EncodeToString(b)
}

func initializeRunArtifacts(spawnCwd string, startPhase int, state *phasedState, opts phasedEngineOptions) (string, string, string, []PhaseProgress, error) {
	stateDir := filepath.Join(spawnCwd, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0750); err != nil {
		return "", "", "", nil, fmt.Errorf("create state directory: %w", err)
	}

	logPath := filepath.Join(stateDir, "phased-orchestration.log")
	statusPath := filepath.Join(stateDir, "live-status.md")
	var allPhases []PhaseProgress

	if startPhase == 1 {
		cleanPhaseSummaries(stateDir)
	}

	fmt.Printf("\n=== RPI Phased: %s ===\n", state.Goal)
	fmt.Printf("Starting from phase %d (%s)\n", startPhase, phases[startPhase-1].Name)
	if opts.NoDashboard || GetDryRun() {
		fmt.Println("Monitor: ao rpi status --watch  or  ao rpi serve " + state.RunID)
	}

	if opts.LiveStatus {
		allPhases = buildAllPhases(phases)
		fmt.Printf("Live phase status file: %s\n", statusPath)
		if err := WriteLiveStatus(statusPath, allPhases, startPhase-1); err != nil {
			VerbosePrintf("Warning: could not initialize live status: %v\n", err)
		}
	}

	return stateDir, logPath, statusPath, allPhases, nil
}
