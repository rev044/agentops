package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var rpiStatusWatch bool

func init() {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show active RPI phased runs",
		Long: `Display active and recent RPI phased runs.

Scans for phased-state.json files in the current directory and sibling
worktree directories. Cross-references tmux sessions for liveness.
Also parses orchestration logs for phase history, durations, and verdicts.

Examples:
  ao rpi status
  ao rpi status -o json
  ao rpi status --watch`,
		RunE: runRPIStatus,
	}
	statusCmd.Flags().BoolVar(&rpiStatusWatch, "watch", false, "Poll every 5s and redraw (Ctrl-C to exit)")
	rpiCmd.AddCommand(statusCmd)
}

// --- rpiRun: log-parsed run data ---

// rpiRun represents a single orchestration run parsed from the log file.
type rpiRun struct {
	RunID      string            `json:"run_id"`
	Goal       string            `json:"goal,omitempty"`
	Phases     []rpiPhaseEntry   `json:"phases"`
	StartedAt  time.Time         `json:"started_at"`
	FinishedAt time.Time         `json:"finished_at,omitempty"`
	Duration   time.Duration     `json:"duration,omitempty"`
	Verdicts   map[string]string `json:"verdicts,omitempty"`
	Retries    map[string]int    `json:"retries,omitempty"`
	Status     string            `json:"status"` // running, completed, failed
	EpicID     string            `json:"epic_id,omitempty"`
}

// rpiPhaseEntry represents a single phase log entry within a run.
type rpiPhaseEntry struct {
	Name    string `json:"name"`
	Details string `json:"details"`
	Time    string `json:"time"`
}

// --- rpiRunInfo: state-file-based run data (existing) ---

type rpiRunInfo struct {
	RunID     string `json:"run_id"`
	Goal      string `json:"goal,omitempty"`
	Phase     int    `json:"phase"`
	PhaseName string `json:"phase_name"`
	Status    string `json:"status"`
	EpicID    string `json:"epic_id,omitempty"`
	Worktree  string `json:"worktree,omitempty"`
	StartedAt string `json:"started_at,omitempty"`
	Elapsed   string `json:"elapsed,omitempty"`
}

type rpiStatusOutput struct {
	Runs    []rpiRunInfo `json:"runs"`
	LogRuns []rpiRun     `json:"log_runs,omitempty"`
	Count   int          `json:"count"`
}

func runRPIStatus(cmd *cobra.Command, args []string) error {
	if rpiStatusWatch {
		return runRPIStatusWatch()
	}
	return runRPIStatusOnce()
}

func runRPIStatusOnce() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	runs := discoverRPIRuns(cwd)

	// Parse orchestration logs for enriched data
	logRuns := discoverLogRuns(cwd)

	output := rpiStatusOutput{
		Runs:    runs,
		LogRuns: logRuns,
		Count:   len(runs),
	}

	if GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	// Table output: state-file runs
	if len(runs) == 0 && len(logRuns) == 0 {
		fmt.Println("No active RPI runs found.")
		return nil
	}

	if len(runs) > 0 {
		fmt.Printf("%-14s %-30s %-12s %-10s %s\n", "RUN-ID", "GOAL", "PHASE", "STATUS", "ELAPSED")
		fmt.Println(strings.Repeat("─", 80))
		for _, r := range runs {
			goal := r.Goal
			if len(goal) > 28 {
				goal = goal[:25] + "..."
			}
			fmt.Printf("%-14s %-30s %-12s %-10s %s\n",
				r.RunID, goal, r.PhaseName, r.Status, r.Elapsed)
		}
		fmt.Printf("\n%d active run(s) found.\n", len(runs))
	}

	// Log history section
	if len(logRuns) > 0 {
		fmt.Printf("\n%-14s %-30s %-12s %-10s %-10s %s\n", "RUN-ID", "GOAL", "LAST-PHASE", "STATUS", "RETRIES", "DURATION")
		fmt.Println(strings.Repeat("─", 100))
		for _, lr := range logRuns {
			goal := lr.Goal
			if len(goal) > 28 {
				goal = goal[:25] + "..."
			}
			lastPhase := ""
			if len(lr.Phases) > 0 {
				lastPhase = lr.Phases[len(lr.Phases)-1].Name
			}
			totalRetries := 0
			for _, v := range lr.Retries {
				totalRetries += v
			}
			dur := ""
			if lr.Duration > 0 {
				dur = lr.Duration.Truncate(time.Second).String()
			}
			retryStr := fmt.Sprintf("%d", totalRetries)
			verdictStr := ""
			for k, v := range lr.Verdicts {
				if verdictStr != "" {
					verdictStr += ","
				}
				verdictStr += k + "=" + v
			}
			status := lr.Status
			if verdictStr != "" && status == "completed" {
				status += " [" + verdictStr + "]"
			}
			fmt.Printf("%-14s %-30s %-12s %-10s %-10s %s\n",
				lr.RunID, goal, lastPhase, status, retryStr, dur)
		}
		fmt.Printf("\n%d log run(s) found.\n", len(logRuns))
	}

	return nil
}

// runRPIStatusWatch polls every 5s and redraws the display.
func runRPIStatusWatch() error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Draw immediately on first invocation
	clearScreen()
	if err := runRPIStatusOnce(); err != nil {
		return err
	}
	fmt.Printf("\n[watch mode — polling every 5s, Ctrl-C to exit]")

	for {
		select {
		case <-sigCh:
			fmt.Println("\nExiting watch mode.")
			return nil
		case <-ticker.C:
			clearScreen()
			if err := runRPIStatusOnce(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			fmt.Printf("\n[watch mode — polling every 5s, Ctrl-C to exit]")
		}
	}
}

// clearScreen emits ANSI escape sequences to clear the terminal and move cursor to top.
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

// --- Log parsing ---

// logLineRegex matches both old format and new format log lines.
// Old: [timestamp] phase: details
// New: [timestamp] [runID] phase: details
var logLineRegex = regexp.MustCompile(
	`^\[([^\]]+)\]\s+(?:\[([^\]]+)\]\s+)?([^:]+):\s+(.*)$`,
)

// parseOrchestrationLog reads the orchestration log file and returns parsed runs.
// Handles both old format (no run-ID) and new format (with [runID] bracket).
// Groups entries by run-ID or start->complete blocks for old format.
func parseOrchestrationLog(logPath string) ([]rpiRun, error) {
	f, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("open log: %w", err)
	}
	defer f.Close() //nolint:errcheck

	// Map runID -> *rpiRun for grouping
	runMap := make(map[string]*rpiRun)
	// Order of first appearance
	var runOrder []string
	// Counter for old-format entries without a runID
	anonymousCounter := 0

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		matches := logLineRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		timestamp := matches[1]
		runID := matches[2] // may be empty for old format
		phaseName := strings.TrimSpace(matches[3])
		details := strings.TrimSpace(matches[4])

		// For old format without runID, group by start->complete blocks
		if runID == "" {
			if phaseName == "start" {
				anonymousCounter++
				runID = fmt.Sprintf("anon-%d", anonymousCounter)
			} else {
				// Attach to the current anonymous run
				runID = fmt.Sprintf("anon-%d", anonymousCounter)
				if anonymousCounter == 0 {
					anonymousCounter = 1
					runID = "anon-1"
				}
			}
		}

		run, exists := runMap[runID]
		if !exists {
			run = &rpiRun{
				RunID:    runID,
				Verdicts: make(map[string]string),
				Retries:  make(map[string]int),
				Status:   "running",
			}
			runMap[runID] = run
			runOrder = append(runOrder, runID)
		}

		// Parse timestamp
		t, tErr := time.Parse(time.RFC3339, timestamp)
		if tErr == nil {
			if run.StartedAt.IsZero() {
				run.StartedAt = t
			}
		}

		// Add phase entry
		entry := rpiPhaseEntry{
			Name:    phaseName,
			Details: details,
			Time:    timestamp,
		}
		run.Phases = append(run.Phases, entry)

		// Extract structured data from details
		switch phaseName {
		case "start":
			run.Goal = extractGoalFromDetails(details)
		case "complete":
			run.Status = "completed"
			if tErr == nil {
				run.FinishedAt = t
				if !run.StartedAt.IsZero() {
					run.Duration = run.FinishedAt.Sub(run.StartedAt)
				}
			}
			run.EpicID = extractEpicFromDetails(details)
			extractVerdictsFromDetails(details, run.Verdicts)
		default:
			// Check for FAILED
			if strings.HasPrefix(details, "FAILED:") {
				run.Status = "failed"
			}
			// Check for RETRY
			if strings.HasPrefix(details, "RETRY") {
				run.Retries[phaseName]++
			}
			// Check for completed with duration
			if strings.HasPrefix(details, "completed in ") {
				durStr := strings.TrimPrefix(details, "completed in ")
				if d, dErr := time.ParseDuration(durStr); dErr == nil {
					// Update last-known finish time based on duration
					if tErr == nil {
						run.FinishedAt = t
						_ = d // duration per phase, total computed from start->complete
					}
				}
			}
			// Extract verdict from phase details (e.g., "verdict: PASS")
			if strings.Contains(phaseName, "pre-mortem") || strings.Contains(phaseName, "vibe") {
				if v := extractInlineVerdict(details); v != "" {
					run.Verdicts[phaseName] = v
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan log: %w", err)
	}

	// Build ordered result
	result := make([]rpiRun, 0, len(runOrder))
	for _, id := range runOrder {
		result = append(result, *runMap[id])
	}

	return result, nil
}

// extractGoalFromDetails extracts goal from "goal=\"...\" from=..." format.
func extractGoalFromDetails(details string) string {
	re := regexp.MustCompile(`goal="([^"]*)"`)
	m := re.FindStringSubmatch(details)
	if len(m) >= 2 {
		return m[1]
	}
	return details
}

// extractEpicFromDetails extracts epic ID from "epic=ag-xxx verdicts=..." format.
func extractEpicFromDetails(details string) string {
	re := regexp.MustCompile(`epic=(\S+)`)
	m := re.FindStringSubmatch(details)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// extractVerdictsFromDetails extracts verdicts from "verdicts=map[key:val ...]" format.
func extractVerdictsFromDetails(details string, verdicts map[string]string) {
	re := regexp.MustCompile(`verdicts=map\[([^\]]*)\]`)
	m := re.FindStringSubmatch(details)
	if len(m) < 2 {
		return
	}
	pairs := strings.Fields(m[1])
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			verdicts[parts[0]] = parts[1]
		}
	}
}

// extractInlineVerdict looks for PASS/WARN/FAIL in a details string.
func extractInlineVerdict(details string) string {
	for _, v := range []string{"PASS", "WARN", "FAIL"} {
		if strings.Contains(details, v) {
			return v
		}
	}
	return ""
}

// discoverLogRuns finds and parses orchestration logs in cwd and siblings.
func discoverLogRuns(cwd string) []rpiRun {
	var allRuns []rpiRun

	// Check current directory
	logPath := filepath.Join(cwd, ".agents", "rpi", "phased-orchestration.log")
	if runs, err := parseOrchestrationLog(logPath); err == nil {
		allRuns = append(allRuns, runs...)
	}

	// Check sibling worktree directories
	parent := filepath.Dir(cwd)
	pattern := filepath.Join(parent, "*-rpi-*", ".agents", "rpi", "phased-orchestration.log")
	matches, err := filepath.Glob(pattern)
	if err == nil {
		for _, match := range matches {
			// Skip if same as cwd log
			if match == logPath {
				continue
			}
			if runs, err := parseOrchestrationLog(match); err == nil {
				allRuns = append(allRuns, runs...)
			}
		}
	}

	return allRuns
}

// --- State-file based discovery (existing) ---

func discoverRPIRuns(cwd string) []rpiRunInfo {
	var runs []rpiRunInfo

	// 1. Check current directory: .agents/rpi/phased-state.json
	if run, ok := loadRPIRun(cwd); ok {
		runs = append(runs, run)
	}

	// 2. Check sibling worktree directories: ../*-rpi-*/.agents/rpi/phased-state.json
	parent := filepath.Dir(cwd)
	pattern := filepath.Join(parent, "*-rpi-*", ".agents", "rpi", "phased-state.json")
	matches, err := filepath.Glob(pattern)
	if err == nil {
		for _, match := range matches {
			// Derive the worktree dir (3 levels up from phased-state.json)
			wtDir := filepath.Dir(filepath.Dir(filepath.Dir(match)))
			if wtDir == cwd {
				continue // already checked
			}
			if run, ok := loadRPIRun(wtDir); ok {
				runs = append(runs, run)
			}
		}
	}

	return runs
}

func loadRPIRun(dir string) (rpiRunInfo, bool) {
	stateFile := filepath.Join(dir, ".agents", "rpi", "phased-state.json")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return rpiRunInfo{}, false
	}

	var state phasedState
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
	phaseName := phaseNames[state.Phase]
	if phaseName == "" {
		phaseName = fmt.Sprintf("phase-%d", state.Phase)
	}

	// Determine status via tmux session liveness
	status := determineRunStatus(state)

	elapsed := ""
	if state.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339, state.StartedAt); err == nil {
			elapsed = time.Since(t).Truncate(time.Second).String()
		}
	}

	return rpiRunInfo{
		RunID:     state.RunID,
		Goal:      state.Goal,
		Phase:     state.Phase,
		PhaseName: phaseName,
		Status:    status,
		EpicID:    state.EpicID,
		Worktree:  dir,
		StartedAt: state.StartedAt,
		Elapsed:   elapsed,
	}, true
}

// determineRunStatus checks if a tmux session ao-rpi-<runID>-* exists.
// Returns "running" if a matching tmux session is alive, "completed" if the
// state file indicates all phases are done (phase 6), or "unknown" otherwise.
func determineRunStatus(state phasedState) string {
	if checkTmuxSessionAlive(state.RunID) {
		return "running"
	}
	if state.Phase >= 6 {
		return "completed"
	}
	return "unknown"
}

// checkTmuxSessionAlive checks if any tmux session matching ao-rpi-<runID>-* exists.
func checkTmuxSessionAlive(runID string) bool {
	if runID == "" {
		return false
	}
	// Try phases 1-6 for tmux session naming convention ao-rpi-<runID>-p<N>
	for i := 1; i <= 6; i++ {
		sessionName := fmt.Sprintf("ao-rpi-%s-p%d", runID, i)
		cmd := exec.Command("tmux", "has-session", "-t", sessionName)
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	return false
}
