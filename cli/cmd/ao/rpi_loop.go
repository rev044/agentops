package main

import (
	"bufio"
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	rpiMaxCycles             int
	rpiRepoFilter            string
	rpiSupervisor            bool
	rpiRalph                 bool
	rpiFailurePolicy         string
	rpiCycleRetries          int
	rpiRetryBackoff          time.Duration
	rpiCycleDelay            time.Duration
	rpiLease                 bool
	rpiLeasePath             string
	rpiLeaseTTL              time.Duration
	rpiDetachedHeal          bool
	rpiDetachedBranchPrefix  string
	rpiAutoClean             bool
	rpiAutoCleanStaleAfter   time.Duration
	rpiEnsureCleanup         bool
	rpiCleanupPruneWorktrees bool
	rpiCleanupPruneBranches  bool
	rpiGatePolicy            string
	rpiValidateFastScript    string
	rpiSecurityGateScript    string
	rpiLandingPolicy         string
	rpiLandingBranch         string
	rpiLandingCommitMessage  string
	rpiLandingLockPath       string
	rpiBDSyncPolicy          string
	rpiCommandTimeout        time.Duration
	rpiKillSwitchPath        string
	rpiAthena                bool
	rpiAthenaInterval        time.Duration
	rpiAthenaSince           string
	rpiAthenaDefrag          bool
)

var errQueueClaimConflict = errors.New("next-work item no longer available for this consumer")

type queuePreflightDecision struct {
	Consume bool
	Reason  string
}

var (
	queueFileTokenPattern        = regexp.MustCompile(`(?i)\b(?:[A-Za-z0-9_.-]+/)*[A-Za-z0-9_.-]+\.(?:go|py|ts|tsx|js|jsx|json|md|ya?ml|sh|rb|rs|java|c|cc|cpp|h|hpp)\b`)
	queueConsolidationVerb       = regexp.MustCompile(`(?i)\b(merge|merged|remove|delete|rename|fold|consolidate|dedupe|deduplicate|retire|archive|purge)\b`)
	preflightQueueSelectionFn    = preflightQueueSelection
	queuePreflightConsumedBy     = "ao-rpi-loop:preflight"
	defaultQueuePreflightTimeout = 30 * time.Second
)

func init() {
	loopCmd := &cobra.Command{
		Use:   "loop [goal]",
		Short: "Run continuous RPI cycles from next-work queue",
		Long: `Execute RPI cycles in a loop, consuming from next-work.jsonl.

Each cycle drives a queue item through the full phased RPI engine:
  1. Read unconsumed items from .agents/rpi/next-work.jsonl
  2. Pick highest-severity item as goal (or use explicit goal)
  3. Run: ao rpi phased "<goal>" (discovery → implementation → validation)
  4. Claim the queue item while it runs, then consume on success or release on failure
  5. Re-read next-work.jsonl (post-mortem may have harvested new items)
  6. Repeat until queue empty or max-cycles reached

Queue semantics:
  - An item is only marked consumed after the phased engine completes without error.
  - Queue items are claimed before execution and released back to available state
    on interruption or failure so harvested work can continue compounding.
  - Task failures record failed_at per item for retry ordering, but do not consume
    sibling items in the same harvested batch.
  - Already-consumed items and currently-claimed items are skipped (idempotent).

Examples:
  ao rpi loop                          # consume from queue until stable
  ao rpi loop "improve test coverage"  # run one cycle with explicit goal
  ao rpi loop --max-cycles 3           # cap at 3 iterations
  ao rpi loop --repo-filter agentops   # only process items targeting agentops
  ao rpi loop --dry-run                # show what would run`,
		Args: cobra.MaximumNArgs(1),
		RunE: runRPILoop,
	}

	loopCmd.Flags().IntVar(&rpiMaxCycles, "max-cycles", 0, "Maximum cycles (0 = unlimited, stop when queue empty)")
	loopCmd.Flags().StringVar(&rpiRepoFilter, "repo-filter", "", "Only process queue items targeting this repo (empty = all)")
	loopCmd.Flags().BoolVar(&rpiSupervisor, "supervisor", false, "Enable autonomous supervisor mode (lease lock, self-heal, retries, gates, cleanup)")
	loopCmd.Flags().BoolVar(&rpiRalph, "ralph", false, "Enable Ralph-mode preset for unattended external loop supervision (implies supervisor defaults with safe nonstop settings)")
	loopCmd.Flags().StringVar(&rpiFailurePolicy, "failure-policy", "stop", "Cycle failure policy: stop|continue")
	loopCmd.Flags().IntVar(&rpiCycleRetries, "cycle-retries", 0, "Automatic retry count per cycle after a failed attempt")
	loopCmd.Flags().DurationVar(&rpiRetryBackoff, "retry-backoff", 30*time.Second, "Backoff between cycle retry attempts")
	loopCmd.Flags().DurationVar(&rpiCycleDelay, "cycle-delay", 0, "Delay between completed cycles")
	loopCmd.Flags().BoolVar(&rpiLease, "lease", false, "Acquire a single-flight supervisor lease lock before running")
	loopCmd.Flags().StringVar(&rpiLeasePath, "lease-path", filepath.Join(".agents", "rpi", "supervisor.lock"), "Lease lock file path (absolute or repo-relative)")
	loopCmd.Flags().DurationVar(&rpiLeaseTTL, "lease-ttl", 2*time.Minute, "Lease heartbeat TTL for supervisor lock metadata")
	loopCmd.Flags().BoolVar(&rpiDetachedHeal, "detached-heal", false, "Auto-create/switch to a named branch when HEAD is detached")
	loopCmd.Flags().StringVar(&rpiDetachedBranchPrefix, "detached-branch-prefix", "codex/auto-rpi", "Branch prefix used by detached HEAD self-heal")
	loopCmd.Flags().BoolVar(&rpiAutoClean, "auto-clean", false, "Run stale RPI cleanup before each phased cycle")
	loopCmd.Flags().DurationVar(&rpiAutoCleanStaleAfter, "auto-clean-stale-after", 24*time.Hour, "Only auto-clean runs older than this age")
	loopCmd.Flags().BoolVar(&rpiEnsureCleanup, "ensure-cleanup", false, "Run stale-run cleanup after each cycle (cleanup guarantee)")
	loopCmd.Flags().BoolVar(&rpiCleanupPruneWorktrees, "cleanup-prune-worktrees", true, "Run git worktree prune during supervisor cleanup")
	loopCmd.Flags().BoolVar(&rpiCleanupPruneBranches, "cleanup-prune-branches", false, "Run legacy branch cleanup during supervisor cleanup")
	loopCmd.Flags().StringVar(&rpiGatePolicy, "gate-policy", "off", "Quality/security gate policy: off|best-effort|required")
	loopCmd.Flags().StringVar(&rpiValidateFastScript, "gate-fast-script", filepath.Join("scripts", "validate-go-fast.sh"), "Fast validation gate script path")
	loopCmd.Flags().StringVar(&rpiSecurityGateScript, "gate-security-script", filepath.Join("scripts", "security-gate.sh"), "Security gate script path")
	loopCmd.Flags().StringVar(&rpiLandingPolicy, "landing-policy", "off", "Landing policy after successful cycle: off|commit|sync-push")
	loopCmd.Flags().StringVar(&rpiLandingBranch, "landing-branch", "", "Landing target branch (empty resolves origin/HEAD, then current branch, then main)")
	loopCmd.Flags().StringVar(&rpiLandingCommitMessage, "landing-commit-message", "chore(rpi): autonomous cycle {{cycle}}", "Commit message template for landing policies that commit")
	loopCmd.Flags().StringVar(&rpiLandingLockPath, "landing-lock-path", filepath.Join(".agents", "rpi", "landing.lock"), "Landing lock file path for synchronized integration (absolute or repo-relative)")
	loopCmd.Flags().StringVar(&rpiBDSyncPolicy, "bd-sync-policy", "auto", "Legacy bd landing checkpoint policy: auto|always|never (auto/always run 'bd export -o /dev/null' on current bd releases)")
	loopCmd.Flags().DurationVar(&rpiCommandTimeout, "command-timeout", 20*time.Minute, "Timeout for supervisor external commands (git/bd/gate scripts)")
	loopCmd.Flags().StringVar(&rpiKillSwitchPath, "kill-switch-path", filepath.Join(".agents", "rpi", "KILL"), "Supervisor kill-switch file path checked at cycle boundaries (absolute or repo-relative)")
	loopCmd.Flags().BoolVar(&rpiAthena, "athena", false, "Enable Athena producer cadence before queue selection")
	loopCmd.Flags().DurationVar(&rpiAthenaInterval, "athena-interval", 30*time.Minute, "Minimum interval between Athena producer ticks (0 = every cycle)")
	loopCmd.Flags().StringVar(&rpiAthenaSince, "athena-since", "26h", "Lookback window for Athena mine producer")
	loopCmd.Flags().BoolVar(&rpiAthenaDefrag, "athena-defrag", false, "Run defrag sweep after Athena mine producer tick")

	rpiCmd.AddCommand(loopCmd)
}

// nextWorkEntry represents one line in next-work.jsonl.
type nextWorkEntry struct {
	SourceEpic  string         `json:"source_epic"`
	Timestamp   string         `json:"timestamp"`
	Items       []nextWorkItem `json:"items,omitempty"`
	Consumed    bool           `json:"consumed"`
	ClaimStatus string         `json:"claim_status,omitempty"`
	ClaimedBy   *string        `json:"claimed_by,omitempty"`
	ClaimedAt   *string        `json:"claimed_at,omitempty"`
	ConsumedBy  *string        `json:"consumed_by"`
	ConsumedAt  *string        `json:"consumed_at"`
	FailedAt    *string        `json:"failed_at,omitempty"`
	LegacyID    string         `json:"id,omitempty"`
	CreatedAt   string         `json:"created_at,omitempty"`
	Title       string         `json:"title,omitempty"`
	Type        string         `json:"type,omitempty"`
	Severity    string         `json:"severity,omitempty"`
	Source      string         `json:"source,omitempty"`
	Description string         `json:"description,omitempty"`
	Evidence    string         `json:"evidence,omitempty"`
	TargetRepo  string         `json:"target_repo,omitempty"`
	QueueIndex  int            `json:"-"`
}

// nextWorkItem represents a single harvested work item.
type nextWorkItem struct {
	Title       string  `json:"title"`
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Source      string  `json:"source"`
	Description string  `json:"description"`
	Evidence    string  `json:"evidence,omitempty"`
	TargetRepo  string  `json:"target_repo,omitempty"`
	Consumed    bool    `json:"consumed,omitempty"`
	ClaimStatus string  `json:"claim_status,omitempty"`
	ClaimedBy   *string `json:"claimed_by,omitempty"`
	ClaimedAt   *string `json:"claimed_at,omitempty"`
	ConsumedBy  *string `json:"consumed_by,omitempty"`
	ConsumedAt  *string `json:"consumed_at,omitempty"`
	FailedAt    *string `json:"failed_at,omitempty"`
}

// queueSelection holds the selected item together with its source entry index
// so the caller can mark the correct entry consumed/failed.
type queueSelection struct {
	Item       nextWorkItem
	EntryIndex int // 0-based index among parseable JSON entries in next-work.jsonl
	ItemIndex  int // index of the selected item within the entry
	ClaimedBy  string
}

var (
	runRPISupervisedCycleFn = runRPISupervisedCycle
	runAthenaProducerTickFn = runAthenaProducerTick
)

func runRPILoop(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	cfg, err := resolveLoopSupervisorConfig(cmd, cwd)
	if err != nil {
		return err
	}

	explicitGoal := ""
	if len(args) > 0 {
		explicitGoal = args[0]
	}

	nextWorkPath := filepath.Join(cwd, ".agents", "rpi", "next-work.jsonl")
	if err := os.MkdirAll(filepath.Join(cwd, ".agents", "rpi"), 0750); err != nil {
		return fmt.Errorf("ensure .agents/rpi directory: %w", err)
	}

	if cfg.LeaseEnabled && !GetDryRun() {
		runID := generateRunID()
		lease, leaseErr := acquireSupervisorLease(cwd, cfg.LeasePath, cfg.LeaseTTL, runID)
		if leaseErr != nil {
			return leaseErr
		}
		defer func() {
			if releaseErr := lease.Release(); releaseErr != nil {
				VerbosePrintf("Warning: could not release supervisor lease: %v\n", releaseErr)
			}
		}()
		fmt.Printf("Supervisor lease acquired: %s (run=%s)\n", lease.Path(), runID)
	}

	return executeLoopCycles(cwd, explicitGoal, nextWorkPath, cfg)
}

// loopCycleResult signals the loop iteration outcome.
type loopCycleResult int

const (
	loopContinue loopCycleResult = iota
	loopBreak
	loopReturn
)

// errKillSwitchActivated is a sentinel returned when the kill switch fires
// during cycle retries, signaling a clean early exit without queue mutation.
var errKillSwitchActivated = fmt.Errorf("kill switch activated")

// executeLoopCycles runs the main RPI loop consuming from the next-work queue.
func executeLoopCycles(cwd, explicitGoal, nextWorkPath string, cfg rpiLoopSupervisorConfig) error {
	cycle := 0
	executedCycles := 0
	athenaState := athenaProducerState{}
	for {
		cycle++

		if rpiMaxCycles > 0 && cycle > rpiMaxCycles {
			fmt.Printf("\nReached max cycles (%d). Stopping.\n", rpiMaxCycles)
			break
		}
		stop, err := applyCycleDelay(cycle, cfg)
		if err != nil {
			return err
		}
		if stop {
			break
		}

		fmt.Printf("\n=== RPI Loop: Cycle %d ===\n", cycle)
		if err := maybeRunAthenaProducerCadence(cwd, explicitGoal, cfg, &athenaState); err != nil {
			return err
		}

		goal, sel, action, err := resolveLoopGoal(cwd, explicitGoal, nextWorkPath, cfg)
		if err != nil {
			return err
		}
		if action == loopBreak {
			break
		}

		fmt.Printf("Running phased engine for: %q\n", goal)
		executedCycles++

		result, err := runCycleWithRetries(cwd, goal, cycle, executedCycles, nextWorkPath, sel, explicitGoal, cfg)
		if err != nil {
			return err
		}
		if result == loopBreak {
			break
		}
	}

	fmt.Printf("\nRPI loop finished after %d cycle(s).\n", executedCycles)
	return nil
}

type athenaProducerState struct {
	LastTick time.Time
}

func maybeRunAthenaProducerCadence(cwd, explicitGoal string, cfg rpiLoopSupervisorConfig, state *athenaProducerState) error {
	if explicitGoal != "" || !cfg.AthenaEnabled {
		return nil
	}
	if GetDryRun() {
		fmt.Println("[dry-run] Skipping Athena producer cadence.")
		return nil
	}
	if state != nil && cfg.AthenaInterval > 0 && !state.LastTick.IsZero() {
		if elapsed := time.Since(state.LastTick); elapsed < cfg.AthenaInterval {
			return nil
		}
	}
	if err := runAthenaProducerTickFn(cwd, cfg); err != nil {
		wrapped := wrapCycleFailure(cycleFailureInfrastructure, "athena producer", err)
		if cfg.ShouldContinueAfterFailure() {
			VerbosePrintf("Warning: %v\n", wrapped)
			return nil
		}
		return wrapped
	}
	if state != nil {
		state.LastTick = time.Now()
	}
	return nil
}

func runAthenaProducerTick(cwd string, cfg rpiLoopSupervisorConfig) error {
	aoCommand := cmp.Or(strings.TrimSpace(cfg.AOCommand), "ao")
	since := cmp.Or(strings.TrimSpace(cfg.AthenaSince), "26h")

	mineArgs := []string{"mine", "--emit-work-items", "--since", since, "--quiet"}
	fmt.Printf("Athena producer tick: %s %s\n", aoCommand, strings.Join(mineArgs, " "))
	if err := loopCommandRunner(cwd, cfg.CommandTimeout, aoCommand, mineArgs...); err != nil {
		return fmt.Errorf("athena mine producer failed: %w", err)
	}

	if !cfg.AthenaDefrag {
		return nil
	}

	defragArgs := []string{"defrag", "--prune", "--dedup", "--oscillation-sweep", "--quiet"}
	fmt.Printf("Athena producer defrag: %s %s\n", aoCommand, strings.Join(defragArgs, " "))
	if err := loopCommandRunner(cwd, cfg.CommandTimeout, aoCommand, defragArgs...); err != nil {
		return fmt.Errorf("athena defrag sweep failed: %w", err)
	}
	return nil
}

// applyCycleDelay handles inter-cycle delay and kill-switch checking.
// Returns (true, nil) when the loop should stop.
func applyCycleDelay(cycle int, cfg rpiLoopSupervisorConfig) (bool, error) {
	if cycle > 1 && cfg.CycleDelay > 0 {
		fmt.Printf("\nSleeping %s before next cycle...\n", cfg.CycleDelay.Round(time.Second))
		time.Sleep(cfg.CycleDelay)
	}
	killSwitchSet, killErr := isLoopKillSwitchSet(cfg)
	if killErr != nil {
		return false, killErr
	}
	if killSwitchSet {
		fmt.Printf("Kill switch detected (%s). Stopping loop.\n", cfg.KillSwitchPath)
		return true, nil
	}
	return false, nil
}

// resolveLoopGoal determines the goal and queue selection for a cycle.
// Returns the goal string, optional queue selection, and a loop action.
func resolveLoopGoal(cwd, explicitGoal, nextWorkPath string, cfg rpiLoopSupervisorConfig) (string, *queueSelection, loopCycleResult, error) {
	goal := explicitGoal
	var sel *queueSelection

	if goal == "" {
		for {
			entries, err := readQueueEntries(nextWorkPath)
			if err != nil {
				VerbosePrintf("Warning: %v\n", err)
			}
			sel = selectHighestSeverityEntry(entries, rpiRepoFilter)
			if sel == nil {
				fmt.Println("No unconsumed work in queue. Flywheel stable.")
				return "", nil, loopBreak, nil
			}
			if GetDryRun() {
				goal = sel.Item.Title
				fmt.Printf("From queue: %s\n", goal)
				break
			}
			decision, preflightErr := preflightQueueSelectionFn(cwd, sel, cfg)
			if preflightErr != nil {
				return "", nil, loopReturn, preflightErr
			}
			if !decision.Consume {
				goal = sel.Item.Title
				fmt.Printf("From queue: %s\n", goal)
				break
			}
			if err := markItemConsumed(nextWorkPath, sel.EntryIndex, sel.ItemIndex, queuePreflightConsumedBy); err != nil {
				if errors.Is(err, errQueueClaimConflict) {
					continue
				}
				return "", nil, loopReturn, fmt.Errorf("consume preflight queue item %q: %w", sel.Item.Title, err)
			}
			reason := strings.TrimSpace(decision.Reason)
			if reason == "" {
				reason = "repo preflight indicates the work is already satisfied"
			}
			fmt.Printf("Queue preflight consumed %q: %s\n", sel.Item.Title, reason)
		}
	}

	if goal == "" {
		fmt.Println("No goal and empty queue. Nothing to do.")
		return "", nil, loopBreak, nil
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would run phased engine for: %q\n", goal)
		if explicitGoal == "" {
			fmt.Println("[dry-run] Queue not consumed in dry-run. Showing first cycle only.")
		}
		return goal, sel, loopBreak, nil
	}

	return goal, sel, loopContinue, nil
}

func preflightQueueSelection(cwd string, sel *queueSelection, cfg rpiLoopSupervisorConfig) (queuePreflightDecision, error) {
	if sel == nil {
		return queuePreflightDecision{}, nil
	}
	text := strings.TrimSpace(strings.Join([]string{sel.Item.Title, sel.Item.Description, sel.Item.Evidence}, " "))
	if !queueConsolidationVerb.MatchString(text) {
		return queuePreflightDecision{}, nil
	}
	tokens := extractQueuePathTokens(text)
	if len(tokens) == 0 {
		return queuePreflightDecision{}, nil
	}
	timeout := cfg.CommandTimeout
	if timeout <= 0 {
		timeout = defaultQueuePreflightTimeout
	}
	exists, err := queueTokensExistInRepo(cwd, tokens, timeout)
	if err != nil || exists {
		return queuePreflightDecision{}, err
	}
	seenInHistory, err := queueTokensSeenInHistory(cwd, tokens, timeout)
	if err != nil || !seenInHistory {
		return queuePreflightDecision{}, err
	}
	return queuePreflightDecision{
		Consume: true,
		Reason:  fmt.Sprintf("all referenced file tokens are absent from the repo but still present in git history (%s)", strings.Join(tokens, ", ")),
	}, nil
}

func extractQueuePathTokens(text string) []string {
	matches := queueFileTokenPattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}
	tokens := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		token := strings.TrimSpace(match)
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		tokens = append(tokens, token)
	}
	return tokens
}

func queueTokensExistInRepo(cwd string, tokens []string, timeout time.Duration) (bool, error) {
	for _, token := range tokens {
		exists, err := queueTokenExistsInRepo(cwd, token, timeout)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

func queueTokenExistsInRepo(cwd, token string, timeout time.Duration) (bool, error) {
	if strings.Contains(token, "/") {
		if _, err := os.Stat(filepath.Join(cwd, token)); err == nil {
			return true, nil
		} else if !os.IsNotExist(err) {
			return false, err
		}
	}
	out, err := loopCommandOutputRunner(cwd, timeout, "git", append([]string{"ls-files", "--cached", "--others", "--exclude-standard", "--"}, queueGitPathspecs(token)...)...)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(out) != "", nil
}

func queueTokensSeenInHistory(cwd string, tokens []string, timeout time.Duration) (bool, error) {
	args := []string{"log", "--all", "--format=%H", "-n", "1", "--"}
	for _, token := range tokens {
		args = append(args, queueGitPathspecs(token)...)
	}
	out, err := loopCommandOutputRunner(cwd, timeout, "git", args...)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(out) != "", nil
}

func queueGitPathspecs(token string) []string {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	if strings.Contains(token, "/") {
		return []string{token}
	}
	return []string{token, ":(glob)**/" + token}
}

// runCycleWithRetries executes a single cycle with retry logic and handles
// success/failure queue marking.
func runCycleWithRetries(cwd, goal string, cycle, executedCycles int, nextWorkPath string, sel *queueSelection, explicitGoal string, cfg rpiLoopSupervisorConfig) (loopCycleResult, error) {
	if err := claimQueueSelection(nextWorkPath, sel, cycle); err != nil {
		if errors.Is(err, errQueueClaimConflict) && explicitGoal == "" {
			fmt.Printf("Queue contention for %q; another consumer won the claim. Continuing.\n", goal)
			return loopContinue, nil
		}
		return loopReturn, err
	}

	start := time.Now()
	cycleErr := executeCycleAttempts(cwd, goal, cycle, executedCycles, cfg)
	elapsed := time.Since(start).Round(time.Second)

	// Kill switch fired mid-retry: clean exit without queue mutation.
	if cycleErr == errKillSwitchActivated {
		releaseQueueSelection(nextWorkPath, sel, false)
		return loopBreak, nil
	}

	if cycleErr != nil {
		return handleCycleFailure(cycleErr, cycle, elapsed, nextWorkPath, sel, explicitGoal, cfg)
	}

	fmt.Printf("Cycle %d completed in %s\n", cycle, elapsed)
	markQueueEntryConsumed(nextWorkPath, sel)

	if explicitGoal != "" {
		fmt.Println("Explicit goal completed.")
		return loopBreak, nil
	}
	return loopContinue, nil
}

// executeCycleAttempts runs the phased engine with retry attempts, checking
// the kill switch before each attempt. Returns errKillSwitchActivated when
// the kill switch fires mid-retry (clean exit, no queue mutation).
func executeCycleAttempts(cwd, goal string, cycle, executedCycles int, cfg rpiLoopSupervisorConfig) error {
	maxAttempts := cfg.MaxCycleAttempts()
	var cycleErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		killSwitchSet, killErr := isLoopKillSwitchSet(cfg)
		if killErr != nil {
			return killErr
		}
		if killSwitchSet {
			fmt.Printf("Kill switch detected (%s). Stopping loop before cycle execution.\n", cfg.KillSwitchPath)
			fmt.Printf("\nRPI loop finished after %d cycle(s).\n", executedCycles)
			return errKillSwitchActivated
		}
		cycleErr = runRPISupervisedCycleFn(cwd, goal, cycle, attempt, cfg)
		if cycleErr == nil {
			return nil
		}
		if attempt >= maxAttempts {
			return cycleErr
		}
		fmt.Printf("Cycle %d attempt %d/%d failed: %v\n", cycle, attempt, maxAttempts, cycleErr)
		if cfg.RetryBackoff > 0 {
			fmt.Printf("Retrying in %s...\n", cfg.RetryBackoff.Round(time.Second))
			time.Sleep(cfg.RetryBackoff)
		}
	}
	return cycleErr
}

// handleCycleFailure processes a cycle failure by marking the queue entry and
// deciding whether to continue or stop the loop.
func handleCycleFailure(cycleErr error, cycle int, elapsed time.Duration, nextWorkPath string, sel *queueSelection, explicitGoal string, cfg rpiLoopSupervisorConfig) (loopCycleResult, error) {
	fmt.Printf("Cycle %d failed after %s: %v\n", cycle, elapsed, cycleErr)
	markQueueEntryFailed(nextWorkPath, sel, cycleErr)

	if cfg.ShouldContinueAfterFailure() && explicitGoal == "" {
		fmt.Printf("Failure policy %q: continuing to next queue item.\n", cfg.FailurePolicy)
		return loopContinue, nil
	}

	fmt.Println("Stopping loop due to failure policy.")
	return loopReturn, cycleErr
}

func claimQueueSelection(nextWorkPath string, sel *queueSelection, cycle int) error {
	if sel == nil {
		return nil
	}
	claimedBy := fmt.Sprintf("ao-rpi-loop:cycle-%d", cycle)
	if err := markItemClaimed(nextWorkPath, sel.EntryIndex, sel.ItemIndex, claimedBy); err != nil {
		return fmt.Errorf("claim queue item %q: %w", sel.Item.Title, err)
	}
	sel.ClaimedBy = claimedBy
	fmt.Printf("Queue item claimed: %q\n", sel.Item.Title)
	return nil
}

func releaseQueueSelection(nextWorkPath string, sel *queueSelection, failed bool) {
	if sel == nil {
		return
	}
	var markErr error
	if failed {
		markErr = markItemFailedOwned(nextWorkPath, sel.EntryIndex, sel.ItemIndex, sel.ClaimedBy)
	} else {
		markErr = releaseItemClaimOwned(nextWorkPath, sel.EntryIndex, sel.ItemIndex, sel.ClaimedBy)
	}
	if markErr != nil {
		VerbosePrintf("Warning: could not release queue item claim: %v\n", markErr)
		return
	}
	if failed {
		fmt.Printf("Queue item released for retry after task failure: %q\n", sel.Item.Title)
		return
	}
	fmt.Printf("Queue item released after interruption/transient failure: %q\n", sel.Item.Title)
}

// markQueueEntryFailed marks the queue entry as failed when appropriate.
func markQueueEntryFailed(nextWorkPath string, sel *queueSelection, cycleErr error) {
	if sel == nil {
		return
	}
	if shouldMarkQueueEntryFailed(cycleErr) {
		releaseQueueSelection(nextWorkPath, sel, true)
	} else {
		releaseQueueSelection(nextWorkPath, sel, false)
	}
}

// markQueueEntryConsumed marks the specific queue item as consumed after success.
// When all items in the entry are consumed, the entry itself is marked consumed.
func markQueueEntryConsumed(nextWorkPath string, sel *queueSelection) {
	if sel == nil {
		return
	}
	if markErr := markItemConsumedOwned(nextWorkPath, sel.EntryIndex, sel.ItemIndex, "ao-rpi-loop", sel.ClaimedBy); markErr != nil {
		VerbosePrintf("Warning: could not mark queue item as consumed: %v\n", markErr)
	} else {
		fmt.Printf("Queue item consumed: %q\n", sel.Item.Title)
	}
}

// readUnconsumedItems reads next-work.jsonl and returns all unconsumed items
// across all entries, flattened. When repoFilter is non-empty, items with a
// non-empty TargetRepo that is neither "*" nor equal to repoFilter are skipped.
// Items without a TargetRepo (legacy) or with TargetRepo=="*" always pass.
func readUnconsumedItems(path string, repoFilter string) ([]nextWorkItem, error) {
	entries, err := readQueueEntries(path)
	if err != nil {
		return nil, err
	}

	var items []nextWorkItem
	for _, entry := range entries {
		for _, item := range entry.Items {
			if !isQueueItemSelectable(item) {
				continue
			}
			if repoFilter != "" && item.TargetRepo != "" && item.TargetRepo != "*" && item.TargetRepo != repoFilter {
				continue
			}
			items = append(items, item)
		}
	}
	return items, nil
}

// readQueueEntries reads next-work.jsonl and returns entries with at least one
// selectable queue item (with their 0-based index preserved for later marking).
// Malformed
// lines are skipped with a verbose warning. Missing files return nil, nil.
func readQueueEntries(path string) ([]nextWorkEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open next-work.jsonl: %w", err)
	}
	defer f.Close()

	var entries []nextWorkEntry
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	parseableIndex := -1

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		entry, err := parseNextWorkEntryLine(line)
		if err != nil {
			VerbosePrintf("Skipping malformed line: %v\n", err)
			continue
		}
		parseableIndex++
		entry.QueueIndex = parseableIndex

		// Skip entries that are already consumed or use legacy failed-at suppression.
		if entry.Consumed || normalizeClaimStatus(entry.Consumed, entry.ClaimStatus) == "consumed" {
			continue
		}
		if shouldSkipLegacyFailedEntry(entry) {
			continue
		}
		if len(entry.Items) == 0 {
			continue
		}
		// Skip entries where all items are either consumed or currently claimed.
		hasSelectableItem := false
		for _, item := range entry.Items {
			if isQueueItemSelectable(item) {
				hasSelectableItem = true
				break
			}
		}
		if !hasSelectableItem {
			continue
		}

		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

func parseNextWorkEntryLine(line string) (nextWorkEntry, error) {
	var entry nextWorkEntry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		return nextWorkEntry{}, err
	}

	if entry.Timestamp == "" && entry.CreatedAt != "" {
		entry.Timestamp = entry.CreatedAt
	}

	if len(entry.Items) == 0 && hasLegacyFlatNextWorkItem(entry) {
		entry.Items = []nextWorkItem{{
			Title:       entry.Title,
			Type:        entry.Type,
			Severity:    entry.Severity,
			Source:      entry.Source,
			Description: entry.Description,
			Evidence:    entry.Evidence,
			TargetRepo:  entry.TargetRepo,
			Consumed:    entry.Consumed,
			ClaimStatus: normalizeClaimStatus(entry.Consumed, entry.ClaimStatus),
			ClaimedBy:   entry.ClaimedBy,
			ClaimedAt:   entry.ClaimedAt,
			ConsumedBy:  entry.ConsumedBy,
			ConsumedAt:  entry.ConsumedAt,
			FailedAt:    entry.FailedAt,
		}}
	}

	return entry, nil
}

func hasLegacyFlatNextWorkItem(entry nextWorkEntry) bool {
	return strings.TrimSpace(entry.Title) != "" ||
		strings.TrimSpace(entry.Type) != "" ||
		strings.TrimSpace(entry.Severity) != "" ||
		strings.TrimSpace(entry.Description) != "" ||
		strings.TrimSpace(entry.Evidence) != "" ||
		strings.TrimSpace(entry.TargetRepo) != "" ||
		strings.TrimSpace(entry.Source) != ""
}

// normalizeClaimStatus keeps omitted item `claim_status` semantically
// equivalent to available unless the item is already consumed.
func normalizeClaimStatus(consumed bool, claimStatus string) string {
	switch claimStatus {
	case "available", "in_progress", "consumed":
		if consumed && claimStatus != "in_progress" {
			return "consumed"
		}
		return claimStatus
	default:
		if consumed {
			return "consumed"
		}
		return "available"
	}
}

func isQueueItemSelectable(item nextWorkItem) bool {
	if item.Consumed || normalizeClaimStatus(item.Consumed, item.ClaimStatus) == "consumed" {
		return false
	}
	return normalizeClaimStatus(item.Consumed, item.ClaimStatus) != "in_progress"
}

func hasQueueItemLifecycleMetadata(item nextWorkItem) bool {
	return item.ClaimStatus != "" ||
		item.ClaimedBy != nil ||
		item.ClaimedAt != nil ||
		item.ConsumedBy != nil ||
		item.ConsumedAt != nil ||
		item.FailedAt != nil
}

func shouldSkipLegacyFailedEntry(entry nextWorkEntry) bool {
	if entry.FailedAt == nil {
		return false
	}
	if entry.ClaimStatus != "" || entry.ClaimedBy != nil || entry.ClaimedAt != nil {
		return false
	}
	for _, item := range entry.Items {
		if hasQueueItemLifecycleMetadata(item) {
			return false
		}
	}
	return true
}

// selectHighestSeverityEntry picks the best item across all eligible entries.
// It returns a queueSelection containing the winning item and its source entry
// parseable index in next-work.jsonl. Items filtered out by repoFilter are skipped.
// Returns nil if no eligible items exist.
func selectHighestSeverityEntry(entries []nextWorkEntry, repoFilter string) *queueSelection {
	type candidate struct {
		item       nextWorkItem
		entryIndex int
		itemIndex  int
		severity   int
		affinity   int
		freshness  int
		typeRank   int
	}

	var candidates []candidate
	for _, entry := range entries {
		for itemIdx, item := range entry.Items {
			if !isQueueItemSelectable(item) {
				continue
			}
			if repoFilter != "" && item.TargetRepo != "" && item.TargetRepo != "*" && item.TargetRepo != repoFilter {
				continue
			}
			candidates = append(candidates, candidate{
				item:       item,
				entryIndex: entry.QueueIndex,
				itemIndex:  itemIdx,
				severity:   severityRank(item.Severity),
				affinity:   repoAffinityRank(item, repoFilter),
				freshness:  freshnessRank(item),
				typeRank:   workTypeRank(item),
			})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	slices.SortFunc(candidates, func(a, b candidate) int {
		if diff := cmp.Compare(b.affinity, a.affinity); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(b.freshness, a.freshness); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(b.severity, a.severity); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(b.typeRank, a.typeRank); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(a.entryIndex, b.entryIndex); diff != 0 {
			return diff
		}
		return cmp.Compare(a.itemIndex, b.itemIndex)
	})

	best := candidates[0]
	return &queueSelection{Item: best.item, EntryIndex: best.entryIndex, ItemIndex: best.itemIndex}
}

func freshnessRank(item nextWorkItem) int {
	if item.FailedAt != nil {
		return 0
	}
	return 1
}

func repoAffinityRank(item nextWorkItem, repoFilter string) int {
	if repoFilter == "" {
		return 0
	}
	switch item.TargetRepo {
	case repoFilter:
		return 3
	case "*":
		return 2
	case "":
		return 1
	default:
		return 0
	}
}

func workTypeRank(item nextWorkItem) int {
	switch item.Type {
	case "feature", "improvement", "tech-debt", "pattern-fix", "bug", "task":
		return 2
	case "process-improvement":
		return 1
	default:
		return 0
	}
}

// rewriteNextWorkFile rewrites the JSONL file with updated entries applied via
// the transform function. The full read-modify-write runs under an exclusive
// flock so concurrent queue consumers cannot interleave updates. Entries that
// could not be parsed are preserved verbatim. If the file does not exist,
// rewriteNextWorkFile is a no-op.
func rewriteNextWorkFile(path string, transform func(idx int, entry *nextWorkEntry) error) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open next-work.jsonl: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	if err := flockLock(f); err != nil {
		return fmt.Errorf("lock next-work.jsonl: %w", err)
	}
	defer func() {
		_ = flockUnlock(f)
	}()
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek next-work.jsonl: %w", err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read next-work.jsonl: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var lines []string
	parseableIndex := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			lines = append(lines, line)
			continue
		}

		var entry nextWorkEntry
		if jsonErr := json.Unmarshal([]byte(line), &entry); jsonErr != nil {
			// Preserve malformed lines verbatim.
			lines = append(lines, line)
			continue
		}

		if err := transform(parseableIndex, &entry); err != nil {
			return err
		}
		rewritten, marshalErr := json.Marshal(entry)
		if marshalErr != nil {
			lines = append(lines, line)
		} else {
			lines = append(lines, string(rewritten))
		}
		parseableIndex++
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan next-work.jsonl: %w", err)
	}

	var out bytes.Buffer
	for _, l := range lines {
		out.WriteString(l)
		out.WriteByte('\n')
	}

	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("truncate next-work.jsonl: %w", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek next-work.jsonl for write: %w", err)
	}
	if _, err := f.Write(out.Bytes()); err != nil {
		return fmt.Errorf("write next-work.jsonl: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync next-work.jsonl: %w", err)
	}
	return nil
}

// markEntryConsumed sets Consumed=true and ConsumedAt on the entry at entryIndex.
// entryIndex is the 0-based index of the entry among parseable JSON entries in
// the file (blank/malformed lines do not receive an index).
//
// Returns an error when the file does not exist so callers can distinguish a
// missing-queue situation from a successful no-op.
func markEntryConsumed(path string, entryIndex int, consumedBy string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("next-work.jsonl not found: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	return rewriteNextWorkFile(path, func(idx int, entry *nextWorkEntry) error {
		if idx != entryIndex {
			return nil
		}
		entry.Consumed = true
		entry.ClaimStatus = "consumed"
		entry.ClaimedAt = nil
		entry.ClaimedBy = nil
		entry.ConsumedAt = &now
		entry.ConsumedBy = &consumedBy
		entry.FailedAt = nil
		return nil
	})
}

// markItemConsumed sets Consumed=true on a specific item within an entry.
// When all items in the entry are consumed, the entry itself is marked consumed.
func markItemConsumed(path string, entryIndex int, itemIndex int, consumedBy string) error {
	return markItemConsumedOwned(path, entryIndex, itemIndex, consumedBy, "")
}

func markItemConsumedOwned(path string, entryIndex int, itemIndex int, consumedBy string, expectedClaimedBy string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("next-work.jsonl not found: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	targetFound := false
	err := rewriteNextWorkFile(path, func(idx int, entry *nextWorkEntry) error {
		if idx != entryIndex {
			return nil
		}
		targetFound = true
		if len(entry.Items) == 0 && hasLegacyFlatNextWorkItem(*entry) {
			if err := requireQueueClaimOwner(entry.ClaimedBy, expectedClaimedBy); err != nil {
				return err
			}
			entry.Consumed = true
			entry.ClaimStatus = "consumed"
			entry.ClaimedAt = nil
			entry.ClaimedBy = nil
			entry.ConsumedAt = &now
			entry.ConsumedBy = &consumedBy
			entry.FailedAt = nil
			return nil
		}
		if itemIndex < 0 || itemIndex >= len(entry.Items) {
			return errQueueClaimConflict
		}
		if err := requireQueueClaimOwner(entry.Items[itemIndex].ClaimedBy, expectedClaimedBy); err != nil {
			return err
		}
		entry.Items[itemIndex].Consumed = true
		entry.Items[itemIndex].ClaimStatus = "consumed"
		entry.Items[itemIndex].ClaimedBy = nil
		entry.Items[itemIndex].ClaimedAt = nil
		entry.Items[itemIndex].ConsumedBy = &consumedBy
		entry.Items[itemIndex].ConsumedAt = &now
		entry.Items[itemIndex].FailedAt = nil
		recomputeEntryLifecycle(entry)
		return nil
	})
	if err != nil {
		return err
	}
	if !targetFound {
		return errQueueClaimConflict
	}
	return nil
}

// markEntryFailed records a FailedAt timestamp on the entry at entryIndex without
// setting Consumed. This leaves the entry recoverable: set consumed=false to retry.
func markEntryFailed(path string, entryIndex int) error {
	now := time.Now().UTC().Format(time.RFC3339)
	return rewriteNextWorkFile(path, func(idx int, entry *nextWorkEntry) error {
		if idx != entryIndex {
			return nil
		}
		entry.FailedAt = &now
		entry.ClaimStatus = "available"
		entry.ClaimedBy = nil
		entry.ClaimedAt = nil
		entry.Consumed = false
		return nil
	})
}

func markItemClaimed(path string, entryIndex int, itemIndex int, claimedBy string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("next-work.jsonl not found: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	targetFound := false
	err := rewriteNextWorkFile(path, func(idx int, entry *nextWorkEntry) error {
		if idx != entryIndex {
			return nil
		}
		targetFound = true
		if len(entry.Items) == 0 && hasLegacyFlatNextWorkItem(*entry) {
			if err := ensureQueueItemClaimable(normalizeClaimStatus(entry.Consumed, entry.ClaimStatus), entry.ClaimedBy, claimedBy); err != nil {
				return err
			}
			entry.ClaimStatus = "in_progress"
			entry.ClaimedBy = &claimedBy
			entry.ClaimedAt = &now
			entry.Consumed = false
			return nil
		}
		if itemIndex < 0 || itemIndex >= len(entry.Items) {
			return errQueueClaimConflict
		}
		if err := ensureQueueItemClaimable(normalizeClaimStatus(entry.Items[itemIndex].Consumed, entry.Items[itemIndex].ClaimStatus), entry.Items[itemIndex].ClaimedBy, claimedBy); err != nil {
			return err
		}
		entry.Items[itemIndex].ClaimStatus = "in_progress"
		entry.Items[itemIndex].ClaimedBy = &claimedBy
		entry.Items[itemIndex].ClaimedAt = &now
		entry.Items[itemIndex].Consumed = false
		recomputeEntryLifecycle(entry)
		return nil
	})
	if err != nil {
		return err
	}
	if !targetFound {
		return errQueueClaimConflict
	}
	return nil
}

func releaseItemClaim(path string, entryIndex int, itemIndex int) error {
	return releaseItemClaimOwned(path, entryIndex, itemIndex, "")
}

func markItemFailed(path string, entryIndex int, itemIndex int) error {
	return markItemFailedOwned(path, entryIndex, itemIndex, "")
}

func releaseItemClaimOwned(path string, entryIndex int, itemIndex int, expectedClaimedBy string) error {
	return releaseQueueItem(path, entryIndex, itemIndex, nil, expectedClaimedBy)
}

func markItemFailedOwned(path string, entryIndex int, itemIndex int, expectedClaimedBy string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	return releaseQueueItem(path, entryIndex, itemIndex, &now, expectedClaimedBy)
}

func releaseQueueItem(path string, entryIndex int, itemIndex int, failedAt *string, expectedClaimedBy string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("next-work.jsonl not found: %w", err)
	}
	targetFound := false
	err := rewriteNextWorkFile(path, func(idx int, entry *nextWorkEntry) error {
		if idx != entryIndex {
			return nil
		}
		targetFound = true
		if len(entry.Items) == 0 && hasLegacyFlatNextWorkItem(*entry) {
			if err := requireQueueClaimOwner(entry.ClaimedBy, expectedClaimedBy); err != nil {
				return err
			}
			entry.ClaimStatus = "available"
			entry.ClaimedBy = nil
			entry.ClaimedAt = nil
			entry.Consumed = false
			if failedAt != nil {
				entry.FailedAt = failedAt
			}
			return nil
		}
		if itemIndex < 0 || itemIndex >= len(entry.Items) {
			return errQueueClaimConflict
		}
		if err := requireQueueClaimOwner(entry.Items[itemIndex].ClaimedBy, expectedClaimedBy); err != nil {
			return err
		}
		entry.Items[itemIndex].ClaimStatus = "available"
		entry.Items[itemIndex].ClaimedBy = nil
		entry.Items[itemIndex].ClaimedAt = nil
		entry.Items[itemIndex].Consumed = false
		if failedAt != nil {
			entry.Items[itemIndex].FailedAt = failedAt
		}
		recomputeEntryLifecycle(entry)
		return nil
	})
	if err != nil {
		return err
	}
	if !targetFound {
		return errQueueClaimConflict
	}
	return nil
}

func ensureQueueItemClaimable(status string, currentClaimedBy *string, claimedBy string) error {
	if status == "consumed" {
		return errQueueClaimConflict
	}
	if status == "in_progress" && (currentClaimedBy == nil || *currentClaimedBy != claimedBy) {
		return errQueueClaimConflict
	}
	return nil
}

func requireQueueClaimOwner(currentClaimedBy *string, expectedClaimedBy string) error {
	if expectedClaimedBy == "" {
		return nil
	}
	if currentClaimedBy == nil || *currentClaimedBy != expectedClaimedBy {
		return errQueueClaimConflict
	}
	return nil
}

func recomputeEntryLifecycle(entry *nextWorkEntry) {
	if len(entry.Items) == 0 {
		return
	}

	allConsumed := true
	claimedIndex := -1
	var latestFailed *string
	var finalConsumedBy *string
	var finalConsumedAt *string

	for i := range entry.Items {
		status := normalizeClaimStatus(entry.Items[i].Consumed, entry.Items[i].ClaimStatus)
		entry.Items[i].ClaimStatus = status

		switch status {
		case "consumed":
			entry.Items[i].Consumed = true
			if entry.Items[i].ConsumedBy != nil {
				finalConsumedBy = entry.Items[i].ConsumedBy
			}
			if entry.Items[i].ConsumedAt != nil {
				finalConsumedAt = entry.Items[i].ConsumedAt
			}
		default:
			allConsumed = false
		}

		if status == "in_progress" && claimedIndex == -1 {
			claimedIndex = i
		}
		if entry.Items[i].FailedAt != nil {
			latestFailed = entry.Items[i].FailedAt
		}
	}

	entry.FailedAt = latestFailed
	if allConsumed {
		entry.Consumed = true
		entry.ClaimStatus = "consumed"
		entry.ClaimedBy = nil
		entry.ClaimedAt = nil
		entry.ConsumedBy = finalConsumedBy
		entry.ConsumedAt = finalConsumedAt
		return
	}

	entry.Consumed = false
	entry.ConsumedBy = nil
	entry.ConsumedAt = nil
	if claimedIndex >= 0 {
		entry.ClaimStatus = "in_progress"
		entry.ClaimedBy = entry.Items[claimedIndex].ClaimedBy
		entry.ClaimedAt = entry.Items[claimedIndex].ClaimedAt
		return
	}
	entry.ClaimStatus = "available"
	entry.ClaimedBy = nil
	entry.ClaimedAt = nil
}

// selectHighestSeverityItem returns the title of the highest-severity item.
// Severity order: high > medium > low.
func selectHighestSeverityItem(items []nextWorkItem) string {
	if len(items) == 0 {
		return ""
	}

	slices.SortFunc(items, func(a, b nextWorkItem) int {
		return cmp.Compare(severityRank(b.Severity), severityRank(a.Severity))
	})

	return items[0].Title
}

func severityRank(s string) int {
	switch s {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}
