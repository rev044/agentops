package main

import (
	"bufio"
	"context"
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

Uses the run registry at .agents/rpi/runs/ as the primary source of truth.
Heartbeat files determine liveness (alive = heartbeat within last 5 minutes).
Tmux sessions are only probed for runs that lack a recent heartbeat, with a
bounded timeout to prevent blocking.

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

// --- rpiRunInfo: state-file-based run data ---

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
	// Liveness metadata (not shown in table, used for categorisation)
	IsActive      bool      `json:"is_active"`
	LastHeartbeat time.Time `json:"last_heartbeat,omitempty"`
}

type rpiStatusOutput struct {
	Active       []rpiRunInfo         `json:"active"`
	Historical   []rpiRunInfo         `json:"historical"`
	Runs         []rpiRunInfo         `json:"runs"`         // combined, kept for back-compat
	LogRuns      []rpiRun             `json:"log_runs,omitempty"`
	LiveStatuses []liveStatusSnapshot `json:"live_statuses,omitempty"`
	Count        int                  `json:"count"`
}

type liveStatusSnapshot struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// heartbeatLiveThreshold is the maximum age of a heartbeat for a run to be
// considered alive without probing tmux.
const heartbeatLiveThreshold = 5 * time.Minute

// tmuxProbeTimeout is the maximum time we will wait for a single tmux probe.
const tmuxProbeTimeout = 2 * time.Second

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

	active, historical := discoverRPIRunsRegistryFirst(cwd)
	allRuns := append(active, historical...)

	// Parse orchestration logs for enriched data
	logRuns := discoverLogRuns(cwd)
	liveStatuses := discoverLiveStatuses(cwd)

	output := rpiStatusOutput{
		Active:       active,
		Historical:   historical,
		Runs:         allRuns,
		LogRuns:      logRuns,
		LiveStatuses: liveStatuses,
		Count:        len(allRuns),
	}

	if GetOutput() == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	// Table output: state-file runs
	if len(allRuns) == 0 && len(logRuns) == 0 && len(liveStatuses) == 0 {
		fmt.Println("No active RPI runs found.")
		return nil
	}

	// Active runs section
	if len(active) > 0 {
		fmt.Println("Active Runs")
		fmt.Printf("%-14s %-30s %-14s %-10s %s\n", "RUN-ID", "GOAL", "PHASE", "STATUS", "ELAPSED")
		fmt.Println(strings.Repeat("─", 82))
		for _, r := range active {
			goal := r.Goal
			if len(goal) > 28 {
				goal = goal[:25] + "..."
			}
			fmt.Printf("%-14s %-30s %-14s %-10s %s\n",
				r.RunID, goal, r.PhaseName, r.Status, r.Elapsed)
		}
		fmt.Printf("\n%d active run(s) found.\n", len(active))
	}

	// Historical runs section
	if len(historical) > 0 {
		if len(active) > 0 {
			fmt.Println()
		}
		fmt.Println("Historical Runs")
		fmt.Printf("%-14s %-30s %-14s %-10s %s\n", "RUN-ID", "GOAL", "PHASE", "STATUS", "ELAPSED")
		fmt.Println(strings.Repeat("─", 82))
		for _, r := range historical {
			goal := r.Goal
			if len(goal) > 28 {
				goal = goal[:25] + "..."
			}
			fmt.Printf("%-14s %-30s %-14s %-10s %s\n",
				r.RunID, goal, r.PhaseName, r.Status, r.Elapsed)
		}
		fmt.Printf("\n%d historical run(s) found.\n", len(historical))
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

	if len(liveStatuses) > 0 {
		fmt.Println("\nLive Status Files")
		fmt.Println(strings.Repeat("─", 100))
		for _, ls := range liveStatuses {
			path := ls.Path
			if rel, err := filepath.Rel(cwd, ls.Path); err == nil {
				path = rel
			}
			fmt.Printf("\n[%s]\n%s\n", path, strings.TrimSpace(ls.Content))
		}
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
			// Check for terminal failure details.
			if strings.HasPrefix(details, "FAILED:") || strings.HasPrefix(details, "FATAL:") {
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
			// Extract inline verdicts from phase details (legacy + consolidated formats).
			if v := extractInlineVerdict(details); v != "" {
				lphase := strings.ToLower(phaseName)
				ldetails := strings.ToLower(details)
				switch {
				case strings.Contains(lphase, "pre-mortem") || strings.Contains(ldetails, "pre-mortem verdict"):
					run.Verdicts["pre_mortem"] = v
				case strings.Contains(lphase, "vibe") || strings.Contains(ldetails, "vibe verdict"):
					run.Verdicts["vibe"] = v
				case strings.Contains(lphase, "post-mortem") || strings.Contains(ldetails, "post-mortem verdict"):
					run.Verdicts["post_mortem"] = v
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

func discoverLiveStatuses(cwd string) []liveStatusSnapshot {
	var snapshots []liveStatusSnapshot
	seen := make(map[string]struct{})

	add := func(path string) {
		if _, ok := seen[path]; ok {
			return
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}
		seen[path] = struct{}{}
		snapshots = append(snapshots, liveStatusSnapshot{
			Path:    path,
			Content: string(data),
		})
	}

	// Current directory live-status.
	add(filepath.Join(cwd, ".agents", "rpi", "live-status.md"))

	// Sibling worktree live-status files.
	parent := filepath.Dir(cwd)
	pattern := filepath.Join(parent, "*-rpi-*", ".agents", "rpi", "live-status.md")
	matches, err := filepath.Glob(pattern)
	if err == nil {
		for _, match := range matches {
			add(match)
		}
	}

	return snapshots
}

// --- Registry-first run discovery ---

// discoverRPIRunsRegistryFirst is the primary discovery path.
// It scans .agents/rpi/runs/ for all run directories, reads state and heartbeat
// files, and uses heartbeat age to separate active from historical runs.
// Tmux is only probed for runs that lack a recent heartbeat, with a bounded
// per-probe timeout.
//
// Returns (active, historical) slices.
func discoverRPIRunsRegistryFirst(cwd string) (active, historical []rpiRunInfo) {
	// Collect all search roots: cwd + sibling worktrees.
	roots := collectSearchRoots(cwd)

	seen := make(map[string]struct{})
	for _, root := range roots {
		runs := scanRegistryRuns(root)
		for _, r := range runs {
			if _, ok := seen[r.RunID]; ok {
				continue
			}
			seen[r.RunID] = struct{}{}
			if r.IsActive {
				active = append(active, r)
			} else {
				historical = append(historical, r)
			}
		}
	}
	return active, historical
}

// discoverRPIRuns is the legacy discovery function kept for backward
// compatibility with existing tests.  It returns all runs (active + historical)
// discovered via the registry-first path, falling back to the flat state file
// when the registry is empty.
func discoverRPIRuns(cwd string) []rpiRunInfo {
	active, historical := discoverRPIRunsRegistryFirst(cwd)
	all := append(active, historical...)
	if len(all) > 0 {
		return all
	}

	// Fallback: flat phased-state.json (backward compatibility for pre-registry runs)
	var fallback []rpiRunInfo
	if run, ok := loadRPIRun(cwd); ok {
		fallback = append(fallback, run)
	}
	parent := filepath.Dir(cwd)
	pattern := filepath.Join(parent, "*-rpi-*", ".agents", "rpi", "phased-state.json")
	matches, err := filepath.Glob(pattern)
	if err == nil {
		for _, match := range matches {
			wtDir := filepath.Dir(filepath.Dir(filepath.Dir(match)))
			if wtDir == cwd {
				continue
			}
			if run, ok := loadRPIRun(wtDir); ok {
				fallback = append(fallback, run)
			}
		}
	}
	return fallback
}

// collectSearchRoots returns the cwd plus any sibling worktree directories
// that match the *-rpi-* naming convention.
func collectSearchRoots(cwd string) []string {
	roots := []string{cwd}
	parent := filepath.Dir(cwd)
	pattern := filepath.Join(parent, "*-rpi-*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return roots
	}
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil || !info.IsDir() {
			continue
		}
		if m == cwd {
			continue
		}
		roots = append(roots, m)
	}
	return roots
}

// scanRegistryRuns reads all run directories under <root>/.agents/rpi/runs/
// and returns rpiRunInfo for each valid run.
func scanRegistryRuns(root string) []rpiRunInfo {
	runsDir := filepath.Join(root, ".agents", "rpi", "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		// Directory may not exist yet; fall through silently.
		return nil
	}

	var runs []rpiRunInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		runID := entry.Name()
		statePath := filepath.Join(runsDir, runID, phasedStateFile)
		data, err := os.ReadFile(statePath)
		if err != nil {
			continue
		}
		state, err := parsePhasedState(data)
		if err != nil || state.RunID == "" {
			continue
		}

		// Determine liveness from heartbeat first, tmux as fallback.
		isActive, lastHB := determineRunLiveness(root, state)

		phaseName := displayPhaseName(*state)
		status := classifyRunStatus(*state, isActive)

		elapsed := ""
		if state.StartedAt != "" {
			if t, err := time.Parse(time.RFC3339, state.StartedAt); err == nil {
				elapsed = time.Since(t).Truncate(time.Second).String()
			}
		}

		runs = append(runs, rpiRunInfo{
			RunID:         state.RunID,
			Goal:          state.Goal,
			Phase:         state.Phase,
			PhaseName:     phaseName,
			Status:        status,
			EpicID:        state.EpicID,
			Worktree:      root,
			StartedAt:     state.StartedAt,
			Elapsed:       elapsed,
			IsActive:      isActive,
			LastHeartbeat: lastHB,
		})
	}
	return runs
}

// determineRunLiveness decides whether a run is alive.
//
// Priority:
//  1. If heartbeat file exists and is recent (< heartbeatLiveThreshold), the run
//     is alive without any tmux probe.
//  2. If heartbeat is absent or stale, probe tmux with a bounded timeout.
//  3. If neither heartbeat nor tmux session is found, the run is historical.
//
// Returns (isActive bool, lastHeartbeat time.Time).
func determineRunLiveness(cwd string, state *phasedState) (bool, time.Time) {
	hb := readRunHeartbeat(cwd, state.RunID)
	if !hb.IsZero() && time.Since(hb) < heartbeatLiveThreshold {
		// Recent heartbeat — alive without tmux probe.
		return true, hb
	}

	// Heartbeat absent or stale: probe tmux with bounded timeout.
	if checkTmuxSessionAlive(state.RunID) {
		return true, hb
	}

	return false, hb
}

// classifyRunStatus derives a human-readable status string.
// Uses liveness information and phase number.
func classifyRunStatus(state phasedState, isActive bool) string {
	if isActive {
		return "running"
	}
	if state.Phase >= completedPhaseNumber(state) {
		return "completed"
	}
	return "unknown"
}

// --- State-file based discovery (legacy, kept for backward compat) ---

func loadRPIRun(dir string) (rpiRunInfo, bool) {
	// Try registry-first: scan .agents/rpi/runs/ for the most recent run.
	runs := scanRegistryRuns(dir)
	if len(runs) > 0 {
		// Return the most recently started run.
		best := runs[0]
		for _, r := range runs[1:] {
			if r.StartedAt > best.StartedAt {
				best = r
			}
		}
		return best, true
	}

	// Fallback: flat phased-state.json for backward compatibility.
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

	phaseName := displayPhaseName(state)

	// Determine status via heartbeat + tmux session liveness.
	isActive, lastHB := determineRunLiveness(dir, &state)
	status := classifyRunStatus(state, isActive)

	elapsed := ""
	if state.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339, state.StartedAt); err == nil {
			elapsed = time.Since(t).Truncate(time.Second).String()
		}
	}

	return rpiRunInfo{
		RunID:         state.RunID,
		Goal:          state.Goal,
		Phase:         state.Phase,
		PhaseName:     phaseName,
		Status:        status,
		EpicID:        state.EpicID,
		Worktree:      dir,
		StartedAt:     state.StartedAt,
		Elapsed:       elapsed,
		IsActive:      isActive,
		LastHeartbeat: lastHB,
	}, true
}

// determineRunStatus checks if a tmux session ao-rpi-<runID>-* exists.
// Returns "running" if a matching tmux session is alive, "completed" if the
// state file indicates all phases are done, or "unknown" otherwise.
func determineRunStatus(state phasedState) string {
	isActive, _ := determineRunLiveness("", &state)
	return classifyRunStatus(state, isActive)
}

func completedPhaseNumber(state phasedState) int {
	// Schema v1+ uses consolidated phased orchestration: 1=discovery, 2=implementation, 3=validation.
	if state.SchemaVersion >= 1 {
		return 3
	}
	// Legacy phased state used six steps.
	return 6
}

func displayPhaseName(state phasedState) string {
	if state.SchemaVersion >= 1 {
		phaseNames := map[int]string{
			1: "discovery",
			2: "implementation",
			3: "validation",
		}
		if phaseName := phaseNames[state.Phase]; phaseName != "" {
			return phaseName
		}
		return fmt.Sprintf("phase-%d", state.Phase)
	}

	// Legacy fallback (pre-consolidation).
	legacyPhaseNames := map[int]string{
		1: "research",
		2: "plan",
		3: "pre-mortem",
		4: "crank",
		5: "vibe",
		6: "post-mortem",
	}
	if phaseName := legacyPhaseNames[state.Phase]; phaseName != "" {
		return phaseName
	}
	return fmt.Sprintf("phase-%d", state.Phase)
}

// checkTmuxSessionAlive checks if any tmux session matching ao-rpi-<runID>-* exists.
// Each probe is bounded by tmuxProbeTimeout to prevent blocking indefinitely.
func checkTmuxSessionAlive(runID string) bool {
	if runID == "" {
		return false
	}
	// Try phases 1-6 for tmux session naming convention ao-rpi-<runID>-p<N>
	for i := 1; i <= 6; i++ {
		sessionName := fmt.Sprintf("ao-rpi-%s-p%d", runID, i)
		ctx, cancel := context.WithTimeout(context.Background(), tmuxProbeTimeout)
		cmd := exec.CommandContext(ctx, "tmux", "has-session", "-t", sessionName)
		err := cmd.Run()
		cancel()
		if err == nil {
			return true
		}
	}
	return false
}

// locateRunMetadata finds the phasedState for a given run ID.
// It searches the run registry across cwd and sibling directories, then falls
// back to the flat phased-state.json. This is used by resume to locate a run
// without relying on cwd heuristics alone.
func locateRunMetadata(cwd, runID string) (*phasedState, string, error) {
	roots := collectSearchRoots(cwd)
	for _, root := range roots {
		registryDir := rpiRunRegistryDir(root, runID)
		if registryDir == "" {
			continue
		}
		statePath := filepath.Join(registryDir, phasedStateFile)
		data, err := os.ReadFile(statePath)
		if err != nil {
			continue
		}
		state, err := parsePhasedState(data)
		if err != nil || state.RunID != runID {
			continue
		}
		return state, root, nil
	}

	// Fallback: flat phased-state.json in cwd (backward compatibility).
	flatPath := filepath.Join(cwd, ".agents", "rpi", phasedStateFile)
	data, err := os.ReadFile(flatPath)
	if err != nil {
		return nil, "", fmt.Errorf("run %s not found in registry or flat state", runID)
	}
	state, err := parsePhasedState(data)
	if err != nil {
		return nil, "", fmt.Errorf("parse flat state for run %s: %w", runID, err)
	}
	if state.RunID != runID {
		return nil, "", fmt.Errorf("run %s not found (flat state contains run %s)", runID, state.RunID)
	}
	return state, cwd, nil
}
