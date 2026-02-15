package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show active RPI phased runs",
		Long: `Display active and recent RPI phased runs.

Discovers runs via git worktree list and phased-state.json files.
Shows phase progress, elapsed time, and process liveness.

Examples:
  ao rpi status
  ao rpi status -o json`,
		RunE: runRPIStatus,
	}
	rpiCmd.AddCommand(statusCmd)
}

type rpiRunInfo struct {
	RunID     string `json:"run_id"`
	Goal      string `json:"goal,omitempty"`
	Phase     int    `json:"current_phase"`
	PhaseName string `json:"phase_name"`
	EpicID    string `json:"epic_id,omitempty"`
	Worktree  string `json:"worktree,omitempty"`
	Active    bool   `json:"active"`
	StartedAt string `json:"started_at,omitempty"`
	Elapsed   string `json:"elapsed,omitempty"`
}

type rpiStatusOutput struct {
	Runs  []rpiRunInfo `json:"runs"`
	Count int          `json:"count"`
}

func runRPIStatus(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	runs := discoverRPIRuns(cwd)

	output := rpiStatusOutput{
		Runs:  runs,
		Count: len(runs),
	}

	if GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	// Table output
	if len(runs) == 0 {
		fmt.Println("No active RPI runs found.")
		return nil
	}

	fmt.Printf("%-14s %-8s %-14s %-8s %-10s %s\n", "RUN-ID", "PHASE", "EPIC", "ACTIVE", "ELAPSED", "GOAL")
	fmt.Println(strings.Repeat("─", 80))
	for _, r := range runs {
		active := "no"
		if r.Active {
			active = "yes"
		}
		goal := r.Goal
		if len(goal) > 30 {
			goal = goal[:27] + "..."
		}
		fmt.Printf("%-14s %-8s %-14s %-8s %-10s %s\n",
			r.RunID, r.PhaseName, r.EpicID, active, r.Elapsed, goal)
	}
	fmt.Printf("\n%d run(s) found.\n", len(runs))
	return nil
}

func discoverRPIRuns(cwd string) []rpiRunInfo {
	var runs []rpiRunInfo

	// Check main repo
	if run, ok := loadRPIRun(cwd, ""); ok {
		runs = append(runs, run)
	}

	// Check worktrees
	out, err := exec.Command("git", "worktree", "list", "--porcelain").Output()
	if err != nil {
		return runs
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			wtPath := strings.TrimPrefix(line, "worktree ")
			if wtPath == cwd {
				continue // already checked main
			}
			if run, ok := loadRPIRun(wtPath, wtPath); ok {
				runs = append(runs, run)
			}
		}
	}

	return runs
}

func loadRPIRun(dir string, worktree string) (rpiRunInfo, bool) {
	stateFile := filepath.Join(dir, ".agents", "rpi", "phased-state.json")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return rpiRunInfo{}, false
	}

	var state struct {
		RunID        string `json:"run_id"`
		Goal         string `json:"goal"`
		CurrentPhase int    `json:"current_phase"`
		EpicID       string `json:"epic_id"`
		StartedAt    string `json:"started_at"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return rpiRunInfo{}, false
	}

	if state.RunID == "" {
		return rpiRunInfo{}, false
	}

	phaseNames := map[int]string{
		1: "research", 2: "plan", 3: "pre-mortem",
		4: "crank", 5: "vibe", 6: "post-mortem",
	}
	phaseName := phaseNames[state.CurrentPhase]
	if phaseName == "" {
		phaseName = fmt.Sprintf("phase-%d", state.CurrentPhase)
	}

	// Check liveness via state file mtime
	info, _ := os.Stat(stateFile)
	active := false
	if info != nil {
		active = time.Since(info.ModTime()) < 10*time.Minute
	}

	elapsed := ""
	if state.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339, state.StartedAt); err == nil {
			elapsed = time.Since(t).Truncate(time.Second).String()
		}
	}

	return rpiRunInfo{
		RunID:     state.RunID,
		Goal:      state.Goal,
		Phase:     state.CurrentPhase,
		PhaseName: phaseName,
		EpicID:    state.EpicID,
		Worktree:  worktree,
		Active:    active,
		StartedAt: state.StartedAt,
		Elapsed:   elapsed,
	}, true
}
