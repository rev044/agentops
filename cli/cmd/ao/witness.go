package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// WitnessState tracks witness status.
type WitnessState struct {
	PID         int       `json:"pid"`
	Session     string    `json:"session"`
	StartedAt   time.Time `json:"started_at"`
	LastPoll    time.Time `json:"last_poll"`
	FarmSession string    `json:"farm_session"`
	Status      string    `json:"status"`
}

var (
	witnessFarmSession string
	witnessPollInterval int
	witnessSummaryInterval int
)

var witnessCmd = &cobra.Command{
	Use:   "witness",
	Short: "Manage witness agent",
	Long: `The witness monitors the Agent Farm and reports to the mayor.

Responsibilities:
  - Poll agent tmux panes every 60 seconds
  - Summarize progress to mayor every 5 minutes
  - Detect blockers and escalate immediately
  - Send "FARM COMPLETE" when all work done

Commands:
  start   Start witness in tmux session
  stop    Stop witness
  status  Show witness state

Examples:
  ao witness start --farm ao-farm-myproject
  ao witness status
  ao witness stop`,
}

var witnessStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start witness agent",
	Long: `Start the witness in a separate tmux session.

The witness runs as a Claude session that monitors the farm,
summarizes progress, and escalates blockers.

Examples:
  ao witness start --farm ao-farm-myproject
  ao witness start --poll 120 --summary 300`,
	RunE: runWitnessStart,
}

var witnessStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop witness agent",
	RunE:  runWitnessStop,
}

var witnessStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show witness status",
	RunE:  runWitnessStatus,
}

func init() {
	rootCmd.AddCommand(witnessCmd)

	witnessCmd.AddCommand(witnessStartCmd)
	witnessCmd.AddCommand(witnessStopCmd)
	witnessCmd.AddCommand(witnessStatusCmd)

	witnessStartCmd.Flags().StringVar(&witnessFarmSession, "farm", "", "Farm tmux session to monitor (required)")
	witnessStartCmd.Flags().IntVar(&witnessPollInterval, "poll", 60, "Seconds between agent polls")
	witnessStartCmd.Flags().IntVar(&witnessSummaryInterval, "summary", 300, "Seconds between progress summaries")

	_ = witnessStartCmd.MarkFlagRequired("farm")
}

func runWitnessStart(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectName := filepath.Base(cwd)
	witnessSession := fmt.Sprintf("ao-farm-witness-%s", projectName)

	// Check if farm session exists
	if !isTmuxSessionRunning(witnessFarmSession) {
		return fmt.Errorf("farm session %s not found - start farm first", witnessFarmSession)
	}

	// Check if witness already running
	if isTmuxSessionRunning(witnessSession) {
		return fmt.Errorf("witness already running in session %s", witnessSession)
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would start witness\n")
		fmt.Printf("[dry-run] Session: %s\n", witnessSession)
		fmt.Printf("[dry-run] Monitoring: %s\n", witnessFarmSession)
		fmt.Printf("[dry-run] Poll interval: %ds\n", witnessPollInterval)
		fmt.Printf("[dry-run] Summary interval: %ds\n", witnessSummaryInterval)
		return nil
	}

	// Create witness session with claude startup command
	fmt.Printf("Starting witness session: %s\n", witnessSession)

	// Use exec env pattern (matches Gastown/daedalus)
	startupCmd := fmt.Sprintf("exec env AO_AGENT_NAME=witness AO_FARM_SESSION=%s claude --dangerously-skip-permissions",
		witnessFarmSession)

	tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", witnessSession, "-c", cwd, startupCmd)
	if err := tmuxCmd.Run(); err != nil {
		return fmt.Errorf("create witness session: %w", err)
	}

	// Wait for Claude to start
	paneID := fmt.Sprintf("%s:0", witnessSession)
	started := waitForClaudeStart(paneID, 30*time.Second)
	if !started {
		return fmt.Errorf("witness claude did not start within timeout")
	}

	// Build witness prompt and send as nudge
	witnessPrompt := buildWitnessPrompt(cwd, witnessFarmSession, witnessPollInterval, witnessSummaryInterval)
	nudgeCmd := exec.Command("tmux", "send-keys", "-t", paneID, "Escape", witnessPrompt, "Enter")
	if err := nudgeCmd.Run(); err != nil {
		return fmt.Errorf("send witness prompt: %w", err)
	}

	// Save witness state
	state := WitnessState{
		PID:         os.Getpid(), // Placeholder - actual PID is inside tmux
		Session:     witnessSession,
		StartedAt:   time.Now(),
		FarmSession: witnessFarmSession,
		Status:      "running",
	}

	statePath := filepath.Join(cwd, ".witness.state")
	if err := saveWitnessState(statePath, &state); err != nil {
		VerbosePrintf("Warning: failed to save witness state: %v\n", err)
	}

	// Write PID file
	pidPath := filepath.Join(cwd, ".witness.pid")
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(state.PID)), 0600); err != nil {
		VerbosePrintf("Warning: failed to write witness PID: %v\n", err)
	}

	fmt.Println()
	fmt.Printf("Witness started\n")
	fmt.Printf("Session: %s\n", witnessSession)
	fmt.Printf("Monitoring: %s\n", witnessFarmSession)
	fmt.Printf("Poll interval: %ds\n", witnessPollInterval)
	fmt.Printf("Summary interval: %ds\n", witnessSummaryInterval)
	fmt.Println()
	fmt.Printf("View: tmux attach -t %s\n", witnessSession)

	return nil
}

func runWitnessStop(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectName := filepath.Base(cwd)
	witnessSession := fmt.Sprintf("ao-farm-witness-%s", projectName)

	if GetDryRun() {
		fmt.Printf("[dry-run] Would stop witness session: %s\n", witnessSession)
		return nil
	}

	// Kill session
	if isTmuxSessionRunning(witnessSession) {
		if err := killTmuxSession(witnessSession); err != nil {
			return fmt.Errorf("kill witness session: %w", err)
		}
		fmt.Printf("Stopped witness session: %s\n", witnessSession)
	} else {
		fmt.Println("Witness not running")
	}

	// Clean up state files
	os.Remove(filepath.Join(cwd, ".witness.state"))
	os.Remove(filepath.Join(cwd, ".witness.pid"))

	return nil
}

func runWitnessStatus(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectName := filepath.Base(cwd)
	witnessSession := fmt.Sprintf("ao-farm-witness-%s", projectName)

	// Check session
	sessionRunning := isTmuxSessionRunning(witnessSession)

	// Load state
	statePath := filepath.Join(cwd, ".witness.state")
	state, stateErr := loadWitnessState(statePath)

	// Check PID
	pidPath := filepath.Join(cwd, ".witness.pid")
	pidData, pidErr := os.ReadFile(pidPath)
	pid := 0
	pidRunning := false
	if pidErr == nil {
		pid, _ = strconv.Atoi(strings.TrimSpace(string(pidData)))
		if pid > 0 {
			pidRunning = isProcessRunning(pid)
		}
	}

	switch GetOutput() {
	case "json":
		result := map[string]interface{}{
			"session":         witnessSession,
			"session_running": sessionRunning,
			"pid":             pid,
			"pid_running":     pidRunning,
		}
		if stateErr == nil {
			result["state"] = state
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)

	default:
		fmt.Println()
		fmt.Println("Witness Status")
		fmt.Println("══════════════")
		fmt.Println()

		if sessionRunning {
			fmt.Printf("Session: %s (running)\n", witnessSession)
		} else {
			fmt.Printf("Session: %s (not running)\n", witnessSession)
		}

		if pid > 0 {
			if pidRunning {
				fmt.Printf("PID: %d (alive)\n", pid)
			} else {
				fmt.Printf("PID: %d (dead)\n", pid)
			}
		}

		if stateErr == nil {
			fmt.Printf("Farm: %s\n", state.FarmSession)
			fmt.Printf("Started: %s\n", state.StartedAt.Format(time.RFC3339))
			if !state.LastPoll.IsZero() {
				fmt.Printf("Last poll: %s\n", state.LastPoll.Format(time.RFC3339))
			}
		}

		fmt.Println()

		if !sessionRunning {
			fmt.Println("Witness is not running. Start with 'ao witness start --farm <session>'")
		}
	}

	return nil
}

func buildWitnessPrompt(cwd, farmSession string, pollInterval, summaryInterval int) string {
	return fmt.Sprintf(`You are the Witness - a monitoring agent for the Agent Farm.

WORKING DIRECTORY: %s
FARM SESSION: %s
POLL INTERVAL: %d seconds
SUMMARY INTERVAL: %d seconds

YOUR RESPONSIBILITIES:

1. MONITOR AGENT HEALTH (every %d seconds)
   - Capture output from farm tmux session
   - Check if agents are active or idle
   - Detect hung or crashed agents

2. SUMMARIZE PROGRESS (every %d seconds)
   - Count completed vs remaining issues
   - Note any blockers or failures
   - Send summary to mayor via: ao mail send --to mayor --body "Progress: X/Y issues done"

3. DETECT AND ESCALATE BLOCKERS
   - If agent stuck for >5 minutes: escalate immediately
   - If circular dependency detected: escalate
   - If >50%% agents failed: escalate with "CIRCUIT BREAKER" alert

4. SIGNAL COMPLETION
   - When ALL agents are idle AND no ready issues remain:
   - Send: ao mail send --to mayor --body "FARM COMPLETE: N issues closed in M minutes"

MONITORING LOOP:

while true:
  1. Check agent panes: tmux capture-pane -t %s -p | tail -50
  2. Count ready issues: bd ready | wc -l
  3. Count in_progress: bd list --status in_progress | wc -l
  4. Write heartbeat: echo $(date +%%s) > .witness.heartbeat
  5. If (ready == 0 && in_progress == 0 && agents_idle): send FARM COMPLETE
  6. Wait %d seconds

HEARTBEAT FILE: .witness.heartbeat
- Write current timestamp every poll
- Mayor checks this to verify witness is alive

START MONITORING NOW. Execute the loop immediately.`, cwd, farmSession, pollInterval, summaryInterval, pollInterval, summaryInterval, farmSession, pollInterval)
}

func escapeForShell(s string) string {
	// Simple escaping - replace single quotes with escaped version
	return strings.ReplaceAll(s, "'", "'\\''")
}

func saveWitnessState(path string, state *WitnessState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func loadWitnessState(path string) (*WitnessState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var state WitnessState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}
