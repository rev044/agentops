package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

// FarmMeta stores farm state for resume and status.
type FarmMeta struct {
	FarmID       string    `json:"farm_id"`
	EpicID       string    `json:"epic_id"`
	AgentCount   int       `json:"agent_count"`
	TmuxSession  string    `json:"tmux_session"`
	WitnessSession string  `json:"witness_session"`
	WitnessPID   int       `json:"witness_pid"`
	StartedAt    time.Time `json:"started_at"`
	AgentPIDs    []int     `json:"agent_pids"`
	Status       string    `json:"status"` // running, stopped, completed
}

var (
	farmAgents    int
	farmEpic      string
	farmStagger   int
	farmReason    string
	farmNoWitness bool
)

var farmCmd = &cobra.Command{
	Use:   "farm",
	Short: "Manage Agent Farm for parallel issue execution",
	Long: `Agent Farm spawns multiple Claude agents to work on issues in parallel.

The farm pattern:
  1. Mayor session runs /farm to spawn agents
  2. Agents claim and work on issues via beads
  3. Witness monitors progress and sends summaries
  4. Mayor checks inbox for updates
  5. Farm completes when all issues done

Commands:
  start     Spawn agents and witness in tmux
  stop      Graceful shutdown of farm
  status    Show running farm state
  validate  Pre-flight checks before starting
  resume    Recover from disconnected session

Examples:
  ao farm start --agents 5
  ao farm status
  ao farm stop
  ao farm validate`,
}

var farmStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Agent Farm",
	Long: `Spawn agents in tmux to work on issues in parallel.

Performs pre-flight validation, then spawns agents serially with stagger delay.
Also spawns a witness in a separate tmux session to monitor progress.

Examples:
  ao farm start --agents 5
  ao farm start --agents 3 --epic gt-100
  ao farm start --agents 5 --stagger 45
  ao farm start --agents 3 --no-witness`,
	RunE: runFarmStart,
}

var farmStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Agent Farm",
	Long: `Graceful shutdown of running Agent Farm.

Sends SIGTERM to witness and all agents, waits for cleanup,
then kills any remaining processes.

Examples:
  ao farm stop
  ao farm stop --reason "manual stop"`,
	RunE: runFarmStop,
}

var farmStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show farm status",
	Long: `Display current Agent Farm status.

Shows agent count, witness health, issues in progress, and uptime.

Examples:
  ao farm status`,
	RunE: runFarmStatus,
}

var farmValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Pre-flight validation",
	Long: `Run pre-flight checks before starting farm.

Validates:
  - beads.jsonl exists and is valid JSON
  - Ready issues > 0
  - No circular dependencies
  - Disk space > 5GB

Examples:
  ao farm validate`,
	RunE: runFarmValidate,
}

var farmResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume disconnected farm",
	Long: `Recover from a disconnected session.

Finds orphaned polecats and in_progress issues, reconciles state,
and optionally restarts agents.

Examples:
  ao farm resume`,
	RunE: runFarmResume,
}

func init() {
	rootCmd.AddCommand(farmCmd)

	// Add subcommands
	farmCmd.AddCommand(farmStartCmd)
	farmCmd.AddCommand(farmStopCmd)
	farmCmd.AddCommand(farmStatusCmd)
	farmCmd.AddCommand(farmValidateCmd)
	farmCmd.AddCommand(farmResumeCmd)

	// Start command flags
	farmStartCmd.Flags().IntVar(&farmAgents, "agents", 5, "Number of agents to spawn (max 10)")
	farmStartCmd.Flags().StringVar(&farmEpic, "epic", "", "Epic ID to work on")
	farmStartCmd.Flags().IntVar(&farmStagger, "stagger", 30, "Seconds between agent spawns")
	farmStartCmd.Flags().BoolVar(&farmNoWitness, "no-witness", false, "Skip witness spawn")

	// Stop command flags
	farmStopCmd.Flags().StringVar(&farmReason, "reason", "manual", "Reason for stopping")
}

func runFarmStart(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Pre-flight validation
	fmt.Println("Running pre-flight validation...")
	if err := validatePreFlight(cwd); err != nil {
		return fmt.Errorf("pre-flight failed: %w", err)
	}
	fmt.Println("Pre-flight: OK")

	// Check ready issues
	readyCount, err := countReadyIssues(cwd)
	if err != nil {
		return fmt.Errorf("count ready issues: %w", err)
	}

	if readyCount == 0 {
		return fmt.Errorf("no ready issues found - run /plan first or check dependencies")
	}

	// Cap agents at ready count and max 10
	agents := farmAgents
	if agents > readyCount {
		agents = readyCount
		VerbosePrintf("Capping agents to %d (ready issue count)\n", agents)
	}
	if agents > 10 {
		agents = 10
		VerbosePrintf("Capping agents to 10 (max limit)\n")
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would start farm with %d agents\n", agents)
		fmt.Printf("[dry-run] Stagger: %d seconds\n", farmStagger)
		fmt.Printf("[dry-run] Witness: %v\n", !farmNoWitness)
		return nil
	}

	// Generate farm ID
	farmID := generateFarmID()
	projectName := filepath.Base(cwd)
	tmuxSession := fmt.Sprintf("ao-farm-%s", projectName)
	witnessSession := fmt.Sprintf("ao-farm-witness-%s", projectName)

	// Check if already running
	if isTmuxSessionRunning(tmuxSession) {
		return fmt.Errorf("farm already running in session %s - use 'ao farm stop' first", tmuxSession)
	}

	// Create farm metadata
	meta := FarmMeta{
		FarmID:         farmID,
		EpicID:         farmEpic,
		AgentCount:     agents,
		TmuxSession:    tmuxSession,
		WitnessSession: witnessSession,
		StartedAt:      time.Now(),
		Status:         "running",
	}

	// Spawn agents serially with stagger
	// First agent creates the session, subsequent agents split
	fmt.Printf("Spawning %d agents (stagger: %ds)...\n", agents, farmStagger)
	agentPIDs := make([]int, 0, agents)

	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for i := 1; i <= agents; i++ {
		select {
		case <-sigChan:
			fmt.Println("\nInterrupted - cleaning up...")
			cleanupFarm(meta)
			return fmt.Errorf("interrupted during spawn")
		default:
		}

		agentName := fmt.Sprintf("agent-%d", i)
		fmt.Printf("  Spawning %s...\n", agentName)

		pid, err := spawnAgent(tmuxSession, agentName, i, cwd)
		if err != nil {
			fmt.Printf("  Warning: failed to spawn %s: %v\n", agentName, err)
			// Circuit breaker: if >50% fail, stop
			if len(agentPIDs) < agents/2 && i > 2 {
				fmt.Println("Circuit breaker: >50% agents failed, stopping")
				cleanupFarm(meta)
				return fmt.Errorf("circuit breaker triggered")
			}
			continue
		}

		agentPIDs = append(agentPIDs, pid)

		// Stagger delay (except for last agent)
		if i < agents {
			time.Sleep(time.Duration(farmStagger) * time.Second)
		}
	}

	meta.AgentPIDs = agentPIDs

	// Spawn witness (unless --no-witness)
	if !farmNoWitness {
		fmt.Printf("Spawning witness in session: %s\n", witnessSession)
		witnessPID, err := spawnWitness(witnessSession, cwd, tmuxSession)
		if err != nil {
			fmt.Printf("Warning: failed to spawn witness: %v\n", err)
		} else {
			meta.WitnessPID = witnessPID

			// Write witness PID to file
			pidPath := filepath.Join(cwd, ".witness.pid")
			if err := os.WriteFile(pidPath, []byte(strconv.Itoa(witnessPID)), 0600); err != nil {
				VerbosePrintf("Warning: failed to write witness PID: %v\n", err)
			}
		}
	}

	// Save farm metadata
	metaPath := filepath.Join(cwd, ".farm.meta")
	if err := saveFarmMeta(metaPath, &meta); err != nil {
		VerbosePrintf("Warning: failed to save farm meta: %v\n", err)
	}

	// Print summary
	fmt.Println()
	fmt.Printf("Farm started: %d agents, %d witness\n", len(agentPIDs), boolToInt(!farmNoWitness))
	fmt.Printf("Session: %s\n", tmuxSession)
	if !farmNoWitness {
		fmt.Printf("Witness: %s\n", witnessSession)
	}
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  ao inbox          - Check messages")
	fmt.Println("  ao farm status    - Show agent states")
	fmt.Println("  ao farm stop      - Graceful shutdown")
	fmt.Printf("  tmux attach -t %s  - View agents\n", tmuxSession)

	return nil
}

func runFarmStop(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Load farm metadata
	metaPath := filepath.Join(cwd, ".farm.meta")
	meta, err := loadFarmMeta(metaPath)
	if err != nil {
		// Try to find and kill sessions by convention
		projectName := filepath.Base(cwd)
		tmuxSession := fmt.Sprintf("ao-farm-%s", projectName)
		witnessSession := fmt.Sprintf("ao-farm-witness-%s", projectName)

		fmt.Printf("No farm metadata found, attempting cleanup by session names...\n")

		if isTmuxSessionRunning(tmuxSession) {
			killTmuxSession(tmuxSession)
			fmt.Printf("Killed session: %s\n", tmuxSession)
		}
		if isTmuxSessionRunning(witnessSession) {
			killTmuxSession(witnessSession)
			fmt.Printf("Killed session: %s\n", witnessSession)
		}

		// Clean up PID files
		os.Remove(filepath.Join(cwd, ".witness.pid"))
		os.Remove(metaPath)

		fmt.Println("Cleanup complete")
		return nil
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would stop farm %s\n", meta.FarmID)
		fmt.Printf("[dry-run] Reason: %s\n", farmReason)
		return nil
	}

	fmt.Printf("Stopping farm %s (reason: %s)...\n", meta.FarmID, farmReason)

	// Graceful shutdown
	cleanupFarm(*meta)

	// Update metadata
	meta.Status = "stopped"
	if err := saveFarmMeta(metaPath, meta); err != nil {
		VerbosePrintf("Warning: failed to update farm meta: %v\n", err)
	}

	fmt.Println("Farm stopped")
	return nil
}

func runFarmStatus(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	metaPath := filepath.Join(cwd, ".farm.meta")
	meta, err := loadFarmMeta(metaPath)
	if err != nil {
		// Check if sessions exist anyway
		projectName := filepath.Base(cwd)
		tmuxSession := fmt.Sprintf("ao-farm-%s", projectName)

		if isTmuxSessionRunning(tmuxSession) {
			fmt.Printf("Farm session found: %s (no metadata)\n", tmuxSession)
			fmt.Println("Run 'ao farm resume' to reconcile state")
			return nil
		}

		fmt.Println("No farm running")
		return nil
	}

	// Output based on format
	switch GetOutput() {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(meta)

	default:
		// Table format
		fmt.Println()
		fmt.Printf("Farm: %s\n", meta.FarmID)
		fmt.Println("═══════════════════════════════════")

		uptime := time.Since(meta.StartedAt).Round(time.Second)
		fmt.Printf("Status:   %s\n", meta.Status)
		fmt.Printf("Uptime:   %s\n", uptime)
		fmt.Printf("Epic:     %s\n", orDefault(meta.EpicID, "(none)"))
		fmt.Println()

		// Check session health
		fmt.Println("Sessions:")
		agentStatus := "running"
		if !isTmuxSessionRunning(meta.TmuxSession) {
			agentStatus = "DEAD"
		}
		fmt.Printf("  Agents:  %s (%s)\n", meta.TmuxSession, agentStatus)

		witnessStatus := "running"
		if meta.WitnessPID > 0 {
			if !isProcessRunning(meta.WitnessPID) {
				witnessStatus = "DEAD"
			}
		} else {
			witnessStatus = "not spawned"
		}
		fmt.Printf("  Witness: %s (%s)\n", meta.WitnessSession, witnessStatus)
		fmt.Println()

		// Agent breakdown
		fmt.Printf("Agents: %d spawned\n", meta.AgentCount)
		liveCount := 0
		for _, pid := range meta.AgentPIDs {
			if isProcessRunning(pid) {
				liveCount++
			}
		}
		fmt.Printf("  Live: %d, Dead: %d\n", liveCount, len(meta.AgentPIDs)-liveCount)
		fmt.Println()

		// Show ready/in-progress issues
		readyCount, _ := countReadyIssues(cwd)
		inProgressCount, _ := countInProgressIssues(cwd)
		fmt.Printf("Issues: %d ready, %d in progress\n", readyCount, inProgressCount)
	}

	return nil
}

func runFarmValidate(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	fmt.Println("Running pre-flight validation...")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	allPassed := true

	// Check 1: beads.jsonl exists
	beadsPath := filepath.Join(cwd, ".beads", "issues.jsonl")
	if _, err := os.Stat(beadsPath); os.IsNotExist(err) {
		fmt.Fprintf(w, "Beads file:\tFAIL\t%s not found\n", beadsPath)
		allPassed = false
	} else {
		// Validate JSON lines
		if err := validateJSONL(beadsPath); err != nil {
			fmt.Fprintf(w, "Beads file:\tFAIL\tinvalid JSONL: %v\n", err)
			allPassed = false
		} else {
			fmt.Fprintf(w, "Beads file:\tOK\t%s\n", beadsPath)
		}
	}

	// Check 2: Ready issues > 0
	readyCount, err := countReadyIssues(cwd)
	if err != nil {
		fmt.Fprintf(w, "Ready issues:\tFAIL\t%v\n", err)
		allPassed = false
	} else if readyCount == 0 {
		fmt.Fprintf(w, "Ready issues:\tFAIL\t0 issues ready (all blocked?)\n")
		allPassed = false
	} else {
		fmt.Fprintf(w, "Ready issues:\tOK\t%d available\n", readyCount)
	}

	// Check 3: Circular dependencies
	if hasCycles, cycle := detectCycles(cwd); hasCycles {
		fmt.Fprintf(w, "Dependencies:\tFAIL\tcircular: %s\n", strings.Join(cycle, " -> "))
		allPassed = false
	} else {
		fmt.Fprintf(w, "Dependencies:\tOK\tno cycles\n")
	}

	// Check 4: Disk space > 5GB
	available, err := getAvailableDiskSpace(cwd)
	if err != nil {
		fmt.Fprintf(w, "Disk space:\tWARN\tcouldn't check: %v\n", err)
	} else if available < 5*1024*1024*1024 { // 5GB in bytes
		fmt.Fprintf(w, "Disk space:\tWARN\t%.1f GB available (< 5GB)\n", float64(available)/(1024*1024*1024))
	} else {
		fmt.Fprintf(w, "Disk space:\tOK\t%.1f GB available\n", float64(available)/(1024*1024*1024))
	}

	// Check 5: tmux available
	if _, err := exec.LookPath("tmux"); err != nil {
		fmt.Fprintf(w, "tmux:\tFAIL\tnot found in PATH\n")
		allPassed = false
	} else {
		fmt.Fprintf(w, "tmux:\tOK\tavailable\n")
	}

	// Check 6: claude CLI available
	if _, err := exec.LookPath("claude"); err != nil {
		fmt.Fprintf(w, "claude:\tFAIL\tnot found in PATH\n")
		allPassed = false
	} else {
		fmt.Fprintf(w, "claude:\tOK\tavailable\n")
	}

	w.Flush()
	fmt.Println()

	if allPassed {
		fmt.Println("All checks passed. Ready to start farm.")
		return nil
	}

	return fmt.Errorf("pre-flight validation failed")
}

func runFarmResume(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	projectName := filepath.Base(cwd)
	tmuxSession := fmt.Sprintf("ao-farm-%s", projectName)
	witnessSession := fmt.Sprintf("ao-farm-witness-%s", projectName)

	fmt.Println("Checking for orphaned farm state...")
	fmt.Println()

	// Check for running sessions
	agentSessionLive := isTmuxSessionRunning(tmuxSession)
	witnessSessionLive := isTmuxSessionRunning(witnessSession)

	// Check for in_progress issues
	inProgressCount, _ := countInProgressIssues(cwd)

	// Load metadata if exists
	metaPath := filepath.Join(cwd, ".farm.meta")
	meta, metaErr := loadFarmMeta(metaPath)

	fmt.Printf("Agent session (%s): %s\n", tmuxSession, statusString(agentSessionLive))
	fmt.Printf("Witness session (%s): %s\n", witnessSession, statusString(witnessSessionLive))
	fmt.Printf("In-progress issues: %d\n", inProgressCount)

	if metaErr == nil {
		fmt.Printf("Farm metadata: found (started %s)\n", meta.StartedAt.Format(time.RFC3339))
	} else {
		fmt.Println("Farm metadata: not found")
	}
	fmt.Println()

	if GetDryRun() {
		fmt.Println("[dry-run] Would reconcile farm state")
		return nil
	}

	// Reconciliation logic
	if !agentSessionLive && !witnessSessionLive && inProgressCount == 0 {
		fmt.Println("No orphaned state found. Use 'ao farm start' to begin.")
		return nil
	}

	// If sessions dead but issues in progress, requeue them
	if !agentSessionLive && inProgressCount > 0 {
		fmt.Printf("Found %d orphaned in_progress issues. Requeuing...\n", inProgressCount)
		if err := requeueOrphanedIssues(cwd); err != nil {
			return fmt.Errorf("requeue issues: %w", err)
		}
		fmt.Println("Issues requeued to ready state")
	}

	// If sessions live, just update metadata
	if agentSessionLive {
		fmt.Println("Agent session still running - nothing to resume")
		if metaErr != nil {
			// Recreate metadata
			newMeta := FarmMeta{
				FarmID:         generateFarmID(),
				TmuxSession:    tmuxSession,
				WitnessSession: witnessSession,
				StartedAt:      time.Now(),
				Status:         "running",
			}
			saveFarmMeta(metaPath, &newMeta)
			fmt.Println("Recreated farm metadata")
		}
	}

	// Clean up stale witness PID file
	pidPath := filepath.Join(cwd, ".witness.pid")
	if data, err := os.ReadFile(pidPath); err == nil {
		pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		if pid > 0 && !isProcessRunning(pid) {
			os.Remove(pidPath)
			fmt.Println("Cleaned up stale witness PID file")
		}
	}

	fmt.Println()
	fmt.Println("Resume complete. Run 'ao farm status' to check state.")

	return nil
}

// Helper functions

func generateFarmID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("farm-%d", time.Now().Unix())
	}
	return fmt.Sprintf("farm-%s", hex.EncodeToString(b))
}

func validatePreFlight(cwd string) error {
	// Quick validation - full validation in runFarmValidate
	beadsPath := filepath.Join(cwd, ".beads", "issues.jsonl")
	if _, err := os.Stat(beadsPath); os.IsNotExist(err) {
		return fmt.Errorf("beads not found at %s", beadsPath)
	}

	readyCount, err := countReadyIssues(cwd)
	if err != nil {
		return fmt.Errorf("count ready issues: %w", err)
	}
	if readyCount == 0 {
		return fmt.Errorf("no ready issues")
	}

	return nil
}

func countReadyIssues(cwd string) (int, error) {
	// Try bd CLI first
	cmd := exec.Command("bd", "ready")
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		count := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				count++
			}
		}
		return count, nil
	}

	// Fallback: parse issues.jsonl
	return countIssuesByStatus(cwd, "ready")
}

func countInProgressIssues(cwd string) (int, error) {
	return countIssuesByStatus(cwd, "in_progress")
}

func countIssuesByStatus(cwd, status string) (int, error) {
	beadsPath := filepath.Join(cwd, ".beads", "issues.jsonl")
	file, err := os.Open(beadsPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var issue map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &issue); err != nil {
			continue
		}
		if s, ok := issue["status"].(string); ok && s == status {
			count++
		}
	}
	return count, scanner.Err()
}

func validateJSONL(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		var obj map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &obj); err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}
	}
	return scanner.Err()
}

func detectCycles(cwd string) (bool, []string) {
	// Try bd validate first
	cmd := exec.Command("bd", "validate", "--check-cycles")
	cmd.Dir = cwd
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if output contains cycle info
		outStr := string(output)
		if strings.Contains(outStr, "cycle") || strings.Contains(outStr, "circular") {
			// Extract cycle from output (simplified)
			return true, []string{"detected (see bd validate output)"}
		}
	}

	// Fallback: basic cycle detection from issues.jsonl
	// This is a simplified version - full implementation would use graph algorithms
	return false, nil
}

func getAvailableDiskSpace(path string) (uint64, error) {
	cmd := exec.Command("df", "-k", path)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return 0, fmt.Errorf("unexpected df output format")
	}

	// Field 3 is available space in KB
	kb, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return 0, err
	}

	return kb * 1024, nil // Convert to bytes
}

func isTmuxSessionRunning(session string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", session)
	return cmd.Run() == nil
}

func createTmuxSession(session string) error {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", session)
	return cmd.Run()
}

func killTmuxSession(session string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", session)
	return cmd.Run()
}

func spawnAgent(session, agentName string, index int, cwd string) (int, error) {
	// Build startup command using exec env pattern (replaces shell so tmux can detect claude)
	// This matches the Gastown/daedalus pattern
	startupCmd := fmt.Sprintf("exec env AO_AGENT_NAME=%s AO_FARM_SESSION=%s claude --dangerously-skip-permissions",
		agentName, session)

	var paneID string

	if index == 1 {
		// First agent: create the session with command
		cmd := exec.Command("tmux", "new-session", "-d", "-s", session, "-c", cwd, startupCmd)
		if err := cmd.Run(); err != nil {
			return 0, fmt.Errorf("create session: %w", err)
		}
		paneID = fmt.Sprintf("%s:0", session)
	} else {
		// Subsequent agents: split window with command
		cmd := exec.Command("tmux", "split-window", "-t", session, "-h", "-c", cwd, startupCmd)
		if err := cmd.Run(); err != nil {
			return 0, fmt.Errorf("split window: %w", err)
		}
		// Rebalance panes
		exec.Command("tmux", "select-layout", "-t", session, "tiled").Run()
		paneID = fmt.Sprintf("%s:%d", session, index-1)
	}

	// Wait for Claude to start (poll for ~30s)
	started := waitForClaudeStart(paneID, 30*time.Second)
	if !started {
		return 0, fmt.Errorf("claude did not start within timeout")
	}

	// Accept the bypass permissions dialog
	// Claude shows "bypass permissions on" which is already selected - just press Enter
	time.Sleep(1 * time.Second) // Wait for UI to render
	acceptCmd := exec.Command("tmux", "send-keys", "-t", paneID, "Enter")
	if err := acceptCmd.Run(); err != nil {
		VerbosePrintf("Warning: failed to accept bypass dialog for %s: %v\n", agentName, err)
	}

	// Wait for bypass to be accepted and prompt to appear
	time.Sleep(2 * time.Second)

	// Now send the agent prompt as a nudge
	agentPrompt := fmt.Sprintf("You are %s in the Agent Farm. Your task: Run /implement on the next ready beads issue. Claim atomically with 'bd update <id> --status in_progress'. After completing, close with 'bd update <id> --status closed' and commit. Repeat until 'bd ready' returns no issues. Send completion messages via 'ao mail send --to mayor --body \"Completed <issue-id>\"'. Start now.", agentName)

	// Clear any partial input with Escape
	exec.Command("tmux", "send-keys", "-t", paneID, "Escape").Run()
	time.Sleep(100 * time.Millisecond)

	// Send the prompt text and Enter to submit
	// Use -l for literal text, then Enter separately
	sendTextCmd := exec.Command("tmux", "send-keys", "-t", paneID, "-l", agentPrompt)
	if err := sendTextCmd.Run(); err != nil {
		VerbosePrintf("Warning: failed to send prompt text to %s: %v\n", agentName, err)
	}
	time.Sleep(100 * time.Millisecond)

	// Send Enter to submit
	sendEnterCmd := exec.Command("tmux", "send-keys", "-t", paneID, "Enter")
	if err := sendEnterCmd.Run(); err != nil {
		VerbosePrintf("Warning: failed to send Enter to %s: %v\n", agentName, err)
	}

	// Return a placeholder PID
	return os.Getpid(), nil
}

// waitForClaudeStart polls tmux pane until claude appears to be running.
func waitForClaudeStart(paneID string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Check pane_current_command
		cmd := exec.Command("tmux", "display-message", "-t", paneID, "-p", "#{pane_current_command}")
		output, err := cmd.Output()
		if err == nil {
			cmdName := strings.TrimSpace(string(output))
			// Claude runs as node process, or shows version number (e.g., "2.1.20")
			// The pane_current_command can show "node", "claude", or the version number
			if strings.Contains(cmdName, "node") || strings.Contains(cmdName, "claude") {
				return true
			}
			// Check for version pattern (digits.digits.digits)
			if len(cmdName) > 0 && cmdName[0] >= '0' && cmdName[0] <= '9' && strings.Contains(cmdName, ".") {
				return true
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return false
}

func spawnWitness(session, cwd, agentSession string) (int, error) {
	// Create witness session with claude startup command
	// Using exec env pattern (matches Gastown/daedalus)
	startupCmd := fmt.Sprintf("exec env AO_AGENT_NAME=witness AO_FARM_SESSION=%s claude --dangerously-skip-permissions",
		agentSession)

	cmd := exec.Command("tmux", "new-session", "-d", "-s", session, "-c", cwd, startupCmd)
	if err := cmd.Run(); err != nil {
		return 0, err
	}

	// Wait for Claude to start
	paneID := fmt.Sprintf("%s:0", session)
	started := waitForClaudeStart(paneID, 30*time.Second)
	if !started {
		return 0, fmt.Errorf("witness claude did not start within timeout")
	}

	// Accept the bypass permissions dialog
	time.Sleep(1 * time.Second)
	acceptCmd := exec.Command("tmux", "send-keys", "-t", paneID, "Enter")
	if err := acceptCmd.Run(); err != nil {
		VerbosePrintf("Warning: failed to accept bypass dialog for witness: %v\n", err)
	}

	// Wait for bypass to be accepted
	time.Sleep(2 * time.Second)

	// Send witness prompt as nudge
	witnessPrompt := fmt.Sprintf(`You are the Witness monitoring Agent Farm session '%s'.

Your tasks (run continuously):
1. Every 60s: Check agent states with 'tmux capture-pane -t %s -p | tail -50'
2. Every 60s: Check progress with 'bd ready | wc -l' and 'bd list --status in_progress | wc -l'
3. Every 5m: Send summary via 'ao mail send --to mayor --body "Progress: X/Y issues done"'
4. Immediately: Escalate blockers via 'ao mail send --to mayor --type blocker --body "BLOCKER: ..."'
5. On completion: Send 'ao mail send --to mayor --type farm_complete --body "FARM COMPLETE: N issues in M min"'

Farm is complete when: bd ready returns 0, bd list --status in_progress returns 0, and agents are idle.
Write heartbeat every poll: echo $(date +%%s) > .witness.heartbeat

Start monitoring now.`, agentSession, agentSession)

	// Clear any partial input with Escape
	exec.Command("tmux", "send-keys", "-t", paneID, "Escape").Run()
	time.Sleep(100 * time.Millisecond)

	// Send the prompt text with -l for literal
	sendTextCmd := exec.Command("tmux", "send-keys", "-t", paneID, "-l", witnessPrompt)
	if err := sendTextCmd.Run(); err != nil {
		VerbosePrintf("Warning: failed to send prompt text to witness: %v\n", err)
	}
	time.Sleep(100 * time.Millisecond)

	// Send Enter to submit
	sendEnterCmd := exec.Command("tmux", "send-keys", "-t", paneID, "Enter")
	if err := sendEnterCmd.Run(); err != nil {
		VerbosePrintf("Warning: failed to send Enter to witness: %v\n", err)
	}

	return os.Getpid(), nil
}

func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds; we need to send signal 0
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func cleanupFarm(meta FarmMeta) {
	fmt.Println("Cleaning up farm resources...")

	// 1. SIGTERM to witness
	if meta.WitnessPID > 0 && isProcessRunning(meta.WitnessPID) {
		fmt.Printf("  Stopping witness (PID %d)...\n", meta.WitnessPID)
		syscall.Kill(meta.WitnessPID, syscall.SIGTERM)
	}

	// 2. Kill witness session
	if meta.WitnessSession != "" && isTmuxSessionRunning(meta.WitnessSession) {
		fmt.Printf("  Killing witness session: %s\n", meta.WitnessSession)
		killTmuxSession(meta.WitnessSession)
	}

	// 3. Kill agent session
	if meta.TmuxSession != "" && isTmuxSessionRunning(meta.TmuxSession) {
		fmt.Printf("  Killing agent session: %s\n", meta.TmuxSession)
		killTmuxSession(meta.TmuxSession)
	}

	// 4. Wait a bit for graceful exit
	time.Sleep(2 * time.Second)

	// 5. SIGKILL stragglers
	for _, pid := range meta.AgentPIDs {
		if isProcessRunning(pid) {
			syscall.Kill(pid, syscall.SIGKILL)
		}
	}

	fmt.Println("Cleanup complete")
}

func saveFarmMeta(path string, meta *FarmMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func loadFarmMeta(path string) (*FarmMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta FarmMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func requeueOrphanedIssues(cwd string) error {
	// Try bd CLI
	cmd := exec.Command("bd", "list", "--status", "in_progress", "-o", "json")
	cmd.Dir = cwd
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("list in_progress issues: %w", err)
	}

	var issues []map[string]interface{}
	if err := json.Unmarshal(output, &issues); err != nil {
		return fmt.Errorf("parse issues: %w", err)
	}

	for _, issue := range issues {
		id, ok := issue["id"].(string)
		if !ok {
			continue
		}
		cmd := exec.Command("bd", "update", id, "--status", "ready")
		cmd.Dir = cwd
		if err := cmd.Run(); err != nil {
			VerbosePrintf("Warning: failed to requeue %s: %v\n", id, err)
		}
	}

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func statusString(running bool) string {
	if running {
		return "running"
	}
	return "not running"
}
