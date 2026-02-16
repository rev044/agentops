package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

var (
	phasedFrom        string
	phasedTestFirst   bool
	phasedFastPath    bool
	phasedInteractive bool
	phasedMaxRetries  int
	phasedNoWorktree  bool
	phasedLiveStatus  bool
)

func init() {
	phasedCmd := &cobra.Command{
		Use:   "phased <goal>",
		Short: "Run RPI with fresh Claude session per phase",
		Long: `Orchestrate the full RPI lifecycle by spawning a fresh Claude session per phase.

Each phase gets its own context window (Ralph Wiggum pattern):
  1. Research  — explore codebase and prior knowledge
  2. Plan      — decompose into trackable issues
  3. Pre-mortem — validate plan with council
  4. Crank     — autonomous implementation
  5. Vibe      — validate with council
  6. Post-mortem — extract learnings

Between phases, the CLI reads filesystem artifacts, constructs prompts
via templates, and spawns the next session. Retry loops for gate failures
are handled across session boundaries.

Examples:
  ao rpi phased "add user authentication"       # full lifecycle
  ao rpi phased --from=crank "add auth"          # skip to crank (needs epic)
  ao rpi phased --from=vibe                      # just validation + post-mortem
  ao rpi phased --dry-run "add auth"             # show prompts without spawning
  ao rpi phased --fast-path "fix typo"           # force --quick for gates`,
		Args: cobra.MaximumNArgs(1),
		RunE: runRPIPhased,
	}

	phasedCmd.Flags().StringVar(&phasedFrom, "from", "research", "Start from phase (research, plan, pre-mortem, crank, vibe, post-mortem)")
	phasedCmd.Flags().BoolVar(&phasedTestFirst, "test-first", false, "Pass --test-first to /crank for spec-first TDD")
	phasedCmd.Flags().BoolVar(&phasedFastPath, "fast-path", false, "Force fast path (--quick for gates)")
	phasedCmd.Flags().BoolVar(&phasedInteractive, "interactive", false, "Enable human gates at research and plan phases")
	phasedCmd.Flags().IntVar(&phasedMaxRetries, "max-retries", 3, "Maximum retry attempts per gate (default: 3)")
	phasedCmd.Flags().BoolVar(&phasedNoWorktree, "no-worktree", false, "Disable worktree isolation (run in current directory)")
	phasedCmd.Flags().BoolVar(&phasedLiveStatus, "live-status", false, "Stream phase progress to a live-status.md file")

	rpiCmd.AddCommand(phasedCmd)
}

// Phase represents an RPI phase with its index and name.
type phase struct {
	Num  int
	Name string
	Step string // ratchet step name
}

var phases = []phase{
	{1, "research", "research"},
	{2, "plan", "plan"},
	{3, "pre-mortem", "pre-mortem"},
	{4, "crank", "implement"},
	{5, "vibe", "vibe"},
	{6, "post-mortem", "post-mortem"},
}

// phasedState persists orchestrator state between phase spawns.
type phasedState struct {
	SchemaVersion int               `json:"schema_version"`
	Goal          string            `json:"goal"`
	EpicID        string            `json:"epic_id,omitempty"`
	Phase         int               `json:"phase"`
	Cycle         int               `json:"cycle"`
	ParentEpic    string            `json:"parent_epic,omitempty"`
	FastPath      bool              `json:"fast_path"`
	TestFirst     bool              `json:"test_first"`
	Verdicts      map[string]string `json:"verdicts"`
	Attempts      map[string]int    `json:"attempts"`
	StartedAt     string            `json:"started_at"`
	WorktreePath  string            `json:"worktree_path,omitempty"`
	RunID         string            `json:"run_id,omitempty"`
}

// retryContext holds context for retrying a failed gate.
type retryContext struct {
	Attempt  int
	Findings []finding
	Verdict  string
}

// finding represents a structured finding from a council report.
type finding struct {
	Description string `json:"description"`
	Fix         string `json:"fix"`
	Ref         string `json:"ref"`
}

// phaseSummaryInstruction is prepended to each phase prompt so Claude writes a rich summary.
// Placed first so it survives context compaction (early instructions persist longer).
const phaseSummaryInstruction = `PHASE SUMMARY CONTRACT: Before finishing this session, write a concise summary (max 500 tokens) to .agents/rpi/phase-{{.PhaseNum}}-summary.md covering key insights, tradeoffs considered, and risks for subsequent phases. This file is read by the next phase.

`

// contextDisciplineInstruction is prepended to every phase prompt to prevent compaction.
// CONTEXT DISCIPLINE: This constant exists so the CLI can enforce context-aware behavior.
const contextDisciplineInstruction = `CONTEXT DISCIPLINE: You are running inside ao rpi phased (phase {{.PhaseNum}} of 6). Each phase gets a FRESH context window. Stay disciplined:
- Do NOT accumulate large file contents in context. Read files with the Read tool JIT and extract only what you need.
- Do NOT explore broadly when narrow exploration suffices. Be surgical.
- Write findings, plans, and results to DISK (files in .agents/), not just in conversation.
- If you are delegating to workers or spawning agents, do NOT accumulate their full output. Read their result files from disk.
- If you notice context degradation (forgetting earlier instructions, repeating yourself, losing track of the goal), IMMEDIATELY write a handoff to .agents/rpi/phase-{{.PhaseNum}}-handoff.md with: (1) what you accomplished, (2) what remains, (3) key context. Then finish cleanly.
{{.ContextBudget}}
`

// phaseContextBudgets provides phase-specific context guidance.
var phaseContextBudgets = map[int]string{
	1: "BUDGET: Limit codebase exploration to ~15 file reads. Write research findings to .agents/research/, not into conversation context.",
	2: "BUDGET: Plan decomposition is lightweight. Write the plan document to .agents/plans/. Keep conversation focused on issue creation.",
	3: "BUDGET: Pre-mortem invokes /council which manages its own agents. Your job: invoke, read the verdict file, done. Minimal context.",
	4: "BUDGET (CRITICAL): Crank is the highest-risk phase for context. /crank spawns workers internally. Do NOT re-read worker output into your context. Trust /crank to manage its waves. Read only the completion status.",
	5: "BUDGET: Vibe invokes /council. Your job: invoke, read the verdict file, done. Minimal context.",
	6: "BUDGET: Post-mortem invokes /council + /retro. Your job: invoke both, read their output files, write summary. Minimal context.",
}

// phasePrompts defines Go templates for each phase's Claude invocation.
var phasePrompts = map[int]string{
	1: `/research "{{.Goal}}"{{if not .Interactive}} --auto{{end}}`,
	2: `/plan "{{.Goal}}"{{if not .Interactive}} --auto{{end}}`,
	3: `/pre-mortem{{if .FastPath}} --quick{{end}}`,
	4: `/crank {{.EpicID}}{{if .TestFirst}} --test-first{{end}}`,
	5: `/vibe{{if .FastPath}} --quick{{end}} recent`,
	6: `/post-mortem{{if .FastPath}} --quick{{end}} {{.EpicID}}`,
}

// retryPrompts defines templates for retry invocations with feedback context.
var retryPrompts = map[int]string{
	// Pre-mortem FAIL → re-plan with feedback
	3: `/plan "{{.Goal}}" --auto` + "\n\n" +
		`Pre-mortem FAIL (attempt {{.RetryAttempt}}/{{.MaxRetries}}). Address these findings:` + "\n" +
		`{{range .Findings}}FINDING: {{.Description}} | FIX: {{.Fix}} | REF: {{.Ref}}` + "\n" + `{{end}}`,
	// Vibe FAIL → re-crank with feedback
	5: `/crank {{.EpicID}}{{if .TestFirst}} --test-first{{end}}` + "\n\n" +
		`Vibe FAIL (attempt {{.RetryAttempt}}/{{.MaxRetries}}). Address these findings:` + "\n" +
		`{{range .Findings}}FINDING: {{.Description}} | FIX: {{.Fix}} | REF: {{.Ref}}` + "\n" + `{{end}}`,
}

func runRPIPhased(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Pre-flight: check claude on PATH
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude CLI not found on PATH (required for spawning phase sessions)")
	}

	originalCwd := cwd
	// spawnCwd tracks the directory for spawning claude sessions.
	// When worktree is active, this is the worktree path; otherwise, it's cwd.
	spawnCwd := cwd

	// Parse goal
	goal := ""
	if len(args) > 0 {
		goal = args[0]
	}

	// Determine start phase
	startPhase := phaseNameToNum(phasedFrom)
	if startPhase == 0 {
		return fmt.Errorf("unknown phase: %q (valid: research, plan, pre-mortem, crank, vibe, post-mortem)", phasedFrom)
	}

	// For crank/vibe/post-mortem without goal, we need an epic ID
	if startPhase >= 4 && goal == "" {
		// Try to extract epic from existing state
		state, err := loadPhasedState(cwd)
		if err == nil && state.EpicID != "" {
			goal = state.Goal
		}
	}

	if goal == "" && startPhase <= 2 {
		return fmt.Errorf("goal is required (provide as argument)")
	}

	// Initialize state
	state := &phasedState{
		SchemaVersion: 1,
		Goal:          goal,
		Phase:         startPhase,
		Cycle:         1,
		FastPath:      phasedFastPath,
		TestFirst:     phasedTestFirst,
		Verdicts:      make(map[string]string),
		Attempts:      make(map[string]int),
		StartedAt:     time.Now().Format(time.RFC3339),
	}

	// Try loading existing state for resume
	if startPhase > 1 {
		existing, err := loadPhasedState(cwd)
		if err == nil {
			state.EpicID = existing.EpicID
			state.FastPath = existing.FastPath || phasedFastPath
			if existing.Verdicts != nil {
				state.Verdicts = existing.Verdicts
			}
			if existing.Attempts != nil {
				state.Attempts = existing.Attempts
			}
			if goal == "" {
				state.Goal = existing.Goal
			}
			// Resume: reuse existing worktree if still present.
			if !phasedNoWorktree && existing.WorktreePath != "" {
				if _, statErr := os.Stat(existing.WorktreePath); statErr == nil {
					spawnCwd = existing.WorktreePath
					state.WorktreePath = existing.WorktreePath
					state.RunID = existing.RunID
					fmt.Printf("Resuming in existing worktree: %s\n", spawnCwd)
				} else {
					return fmt.Errorf("worktree %s from previous run no longer exists (was it removed?)", existing.WorktreePath)
				}
			}
		}
	}

	// Create worktree for isolation (unless resuming into existing one, or opted out).
	cleanupSuccess := false
	var worktreeRunID string
	if !phasedNoWorktree && !GetDryRun() && state.WorktreePath == "" {
		worktreePath, runID, wtErr := createWorktree(cwd)
		if wtErr != nil {
			return fmt.Errorf("create worktree: %w", wtErr)
		}
		spawnCwd = worktreePath
		worktreeRunID = runID
		state.WorktreePath = worktreePath
		state.RunID = runID
		fmt.Printf("Worktree created: %s (branch: rpi/%s)\n", worktreePath, runID)

		// Signal handler: preserve worktree on interruption.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			if sig, ok := <-sigCh; ok {
				fmt.Fprintf(os.Stderr, "\nInterrupted (%v). Worktree preserved at: %s\n", sig, worktreePath)
				os.Exit(1)
			}
		}()

		defer func() {
			signal.Stop(sigCh)
			close(sigCh)
			if cleanupSuccess {
				if mergeErr := mergeWorktree(originalCwd, worktreeRunID); mergeErr != nil {
					fmt.Fprintf(os.Stderr, "Merge failed: %v\nWorktree preserved at: %s\n", mergeErr, worktreePath)
				} else {
					if rmErr := removeWorktree(originalCwd, worktreePath, worktreeRunID); rmErr != nil {
						fmt.Fprintf(os.Stderr, "Cleanup warning: %v\n", rmErr)
					}
				}
			} else {
				fmt.Fprintf(os.Stderr, "Worktree preserved for debugging: %s\n", worktreePath)
			}
		}()
	}

	stateDir := filepath.Join(spawnCwd, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	logPath := filepath.Join(stateDir, "phased-orchestration.log")

	// Clean stale phase summaries from prior runs (only on fresh start)
	if startPhase == 1 {
		cleanPhaseSummaries(stateDir)
	}

	fmt.Printf("\n=== RPI Phased: %s ===\n", state.Goal)
	fmt.Printf("Starting from phase %d (%s)\n", startPhase, phases[startPhase-1].Name)
	logPhaseTransition(logPath, state.RunID, "start", fmt.Sprintf("goal=%q from=%s", state.Goal, phasedFrom))

	// Register with agent mail for observability
	registerRPIAgent(state.RunID)

	// Execute phases sequentially
	for i := startPhase; i <= 6; i++ {
		p := phases[i-1]
		fmt.Printf("\n--- Phase %d: %s ---\n", p.Num, p.Name)
		state.Phase = i

		prompt, err := buildPromptForPhase(spawnCwd, i, state, nil)
		if err != nil {
			return fmt.Errorf("build prompt for phase %d: %w", i, err)
		}

		emitRPIStatus(state.RunID, p.Name, "started")

		if GetDryRun() {
			fmt.Printf("[dry-run] Would spawn: claude -p '%s'\n", prompt)
			if !phasedNoWorktree && i == startPhase {
				runID := generateRunID()
				fmt.Printf("[dry-run] Would create worktree: ../%s-rpi-%s/ (branch: rpi/%s)\n",
					filepath.Base(cwd), runID, runID)
			}
			logPhaseTransition(logPath, state.RunID, p.Name, "dry-run")
			continue
		}

		// Spawn phase session
		fmt.Printf("Spawning: claude -p '%s'\n", prompt)
		start := time.Now()

		var spawnErr error
		if phasedLiveStatus {
			statusPath := filepath.Join(stateDir, "live-status.md")
			allPhases := buildAllPhases(phases)
			spawnErr = spawnClaudePhaseWithStream(prompt, spawnCwd, state.RunID, i, statusPath, allPhases)
		} else {
			spawnErr = spawnClaudePhase(prompt, spawnCwd, state.RunID, i)
		}
		if err := spawnErr; err != nil {
			logPhaseTransition(logPath, state.RunID, p.Name, fmt.Sprintf("FAILED: %v", err))
			return fmt.Errorf("phase %d (%s) failed: %w", i, p.Name, err)
		}

		elapsed := time.Since(start).Round(time.Second)
		fmt.Printf("Phase %d completed in %s\n", i, elapsed)
		logPhaseTransition(logPath, state.RunID, p.Name, fmt.Sprintf("completed in %s", elapsed))
		emitRPIStatus(state.RunID, p.Name, "completed")

		// Post-phase processing
		if err := postPhaseProcessing(spawnCwd, state, i, logPath); err != nil {
			// Check if it's a gate failure that needs retry
			if retryErr, ok := err.(*gateFailError); ok {
				retried, retryErr2 := handleGateRetry(spawnCwd, state, i, retryErr, logPath, spawnCwd)
				if retryErr2 != nil {
					return retryErr2
				}
				if !retried {
					return fmt.Errorf("phase %d (%s): gate failed after max retries", i, p.Name)
				}
				// Retry succeeded, continue to next phase
				continue
			}
			return err
		}

		// Check if phase triggered a handoff (context degradation detected)
		if handoffDetected(spawnCwd, i) {
			fmt.Printf("Phase %d: handoff detected — phase reported context degradation\n", i)
			logPhaseTransition(logPath, state.RunID, p.Name, "HANDOFF detected — context degradation")
			// Continue to next phase (fresh session will pick up from handoff)
		}

		// Write phase summary for next phase's context
		writePhaseSummary(spawnCwd, state, i)

		// Record ratchet checkpoint
		recordRatchetCheckpoint(p.Step)

		// Save state
		if err := savePhasedState(spawnCwd, state); err != nil {
			VerbosePrintf("Warning: could not save state: %v\n", err)
		}
	}

	// All phases completed — mark worktree for merge+cleanup.
	cleanupSuccess = true

	// Final report
	fmt.Printf("\n=== RPI Phased Complete ===\n")
	fmt.Printf("Goal: %s\n", state.Goal)
	if state.EpicID != "" {
		fmt.Printf("Epic: %s\n", state.EpicID)
	}
	fmt.Printf("Verdicts: %v\n", state.Verdicts)
	logPhaseTransition(logPath, state.RunID, "complete", fmt.Sprintf("epic=%s verdicts=%v", state.EpicID, state.Verdicts))

	// Deregister from agent mail
	deregisterRPIAgent(state.RunID)

	return nil
}

// gateFailError signals a gate check failure that may be retried.
type gateFailError struct {
	Phase    int
	Verdict  string
	Findings []finding
	Report   string
}

func (e *gateFailError) Error() string {
	return fmt.Sprintf("gate FAIL at phase %d: %s (report: %s)", e.Phase, e.Verdict, e.Report)
}

// postPhaseProcessing handles phase-specific post-processing.
func postPhaseProcessing(cwd string, state *phasedState, phaseNum int, logPath string) error {
	switch phaseNum {
	case 2: // Plan — extract epic ID and detect fast path
		epicID, err := extractEpicID()
		if err != nil {
			return fmt.Errorf("plan phase: could not extract epic ID (crank needs this): %w", err)
		}
		state.EpicID = epicID
		fmt.Printf("Epic ID: %s\n", epicID)

		if !phasedFastPath {
			fast, err := detectFastPath(state.EpicID)
			if err != nil {
				VerbosePrintf("Warning: fast-path detection failed (continuing without): %v\n", err)
			} else if fast {
				state.FastPath = true
				fmt.Println("Micro-epic detected — using fast path (--quick for gates)")
			}
		}

	case 3: // Pre-mortem — check verdict
		report, err := findLatestCouncilReport(cwd, "pre-mortem", time.Time{}, state.EpicID)
		if err != nil {
			return fmt.Errorf("pre-mortem phase: council report not found (phase may not have completed): %w", err)
		}
		verdict, err := extractCouncilVerdict(report)
		if err != nil {
			return fmt.Errorf("pre-mortem phase: could not extract verdict from %s: %w", report, err)
		}
		state.Verdicts["pre_mortem"] = verdict
		fmt.Printf("Pre-mortem verdict: %s\n", verdict)

		if verdict == "FAIL" {
			findings, _ := extractCouncilFindings(report, 5)
			return &gateFailError{Phase: 3, Verdict: verdict, Findings: findings, Report: report}
		}

	case 4: // Crank — check completion via bd children
		if state.EpicID != "" {
			status, err := checkCrankCompletion(state.EpicID)
			if err != nil {
				VerbosePrintf("Warning: could not check crank completion (continuing to vibe): %v\n", err)
			} else {
				fmt.Printf("Crank status: %s\n", status)
				if status == "BLOCKED" || status == "PARTIAL" {
					return &gateFailError{Phase: 4, Verdict: status, Report: "bd children " + state.EpicID}
				}
			}
		}

	case 5: // Vibe — check verdict
		report, err := findLatestCouncilReport(cwd, "vibe", time.Time{}, state.EpicID)
		if err != nil {
			return fmt.Errorf("vibe phase: council report not found (phase may not have completed): %w", err)
		}
		verdict, err := extractCouncilVerdict(report)
		if err != nil {
			return fmt.Errorf("vibe phase: could not extract verdict from %s: %w", report, err)
		}
		state.Verdicts["vibe"] = verdict
		fmt.Printf("Vibe verdict: %s\n", verdict)

		if verdict == "FAIL" {
			findings, _ := extractCouncilFindings(report, 5)
			return &gateFailError{Phase: 5, Verdict: verdict, Findings: findings, Report: report}
		}
	}

	return nil
}

// handleGateRetry manages retry logic for failed gates.
// spawnCwd is the working directory for spawned claude sessions (may be worktree).
func handleGateRetry(cwd string, state *phasedState, phaseNum int, gateErr *gateFailError, logPath string, spawnCwd string) (bool, error) {
	phaseName := phases[phaseNum-1].Name
	attemptKey := fmt.Sprintf("phase_%d", phaseNum)

	state.Attempts[attemptKey]++
	attempt := state.Attempts[attemptKey]

	if attempt >= phasedMaxRetries {
		msg := fmt.Sprintf("%s failed %d times. Last report: %s. Manual intervention needed.",
			phaseName, phasedMaxRetries, gateErr.Report)
		fmt.Println(msg)
		logPhaseTransition(logPath, state.RunID, phaseName, msg)
		return false, nil
	}

	fmt.Printf("%s: %s (attempt %d/%d) — retrying\n", phaseName, gateErr.Verdict, attempt, phasedMaxRetries)
	logPhaseTransition(logPath, state.RunID, phaseName, fmt.Sprintf("RETRY attempt %d/%d", attempt+1, phasedMaxRetries))

	// Build retry prompt
	retryCtx := &retryContext{
		Attempt:  attempt + 1,
		Findings: gateErr.Findings,
		Verdict:  gateErr.Verdict,
	}

	retryPrompt, err := buildRetryPrompt(cwd, phaseNum, state, retryCtx)
	if err != nil {
		return false, fmt.Errorf("build retry prompt: %w", err)
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would spawn retry: claude -p '%s'\n", retryPrompt)
		return false, nil
	}

	// Spawn retry session
	fmt.Printf("Spawning retry: claude -p '%s'\n", retryPrompt)
	if phasedLiveStatus {
		statusPath := filepath.Join(cwd, ".agents", "rpi", "live-status.md")
		allPhases := buildAllPhases(phases)
		if err := spawnClaudePhaseWithStream(retryPrompt, spawnCwd, state.RunID, phaseNum, statusPath, allPhases); err != nil {
			return false, fmt.Errorf("retry failed: %w", err)
		}
	} else {
		if err := spawnClaudePhase(retryPrompt, spawnCwd, state.RunID, phaseNum); err != nil {
			return false, fmt.Errorf("retry failed: %w", err)
		}
	}

	// Re-run the original phase after retry
	rerunPrompt, err := buildPromptForPhase(cwd, phaseNum, state, nil)
	if err != nil {
		return false, fmt.Errorf("build rerun prompt: %w", err)
	}

	fmt.Printf("Re-running phase %d after retry\n", phaseNum)
	if phasedLiveStatus {
		statusPath := filepath.Join(cwd, ".agents", "rpi", "live-status.md")
		allPhases := buildAllPhases(phases)
		if err := spawnClaudePhaseWithStream(rerunPrompt, spawnCwd, state.RunID, phaseNum, statusPath, allPhases); err != nil {
			return false, fmt.Errorf("rerun failed: %w", err)
		}
	} else {
		if err := spawnClaudePhase(rerunPrompt, spawnCwd, state.RunID, phaseNum); err != nil {
			return false, fmt.Errorf("rerun failed: %w", err)
		}
	}

	// Check gate again
	if err := postPhaseProcessing(cwd, state, phaseNum, logPath); err != nil {
		if _, ok := err.(*gateFailError); ok {
			// Still failing — recurse
			return handleGateRetry(cwd, state, phaseNum, err.(*gateFailError), logPath, spawnCwd)
		}
		return false, err
	}

	return true, nil
}

// buildPromptForPhase constructs the Claude invocation prompt for a phase.
func buildPromptForPhase(cwd string, phaseNum int, state *phasedState, _ *retryContext) (string, error) {
	tmplStr, ok := phasePrompts[phaseNum]
	if !ok {
		return "", fmt.Errorf("no prompt template for phase %d", phaseNum)
	}

	tmpl, err := template.New("phase").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	// Get phase-specific context budget guidance
	budget := phaseContextBudgets[phaseNum]

	data := struct {
		Goal          string
		EpicID        string
		FastPath      bool
		TestFirst     bool
		Interactive   bool
		PhaseNum      int
		ContextBudget string
	}{
		Goal:          state.Goal,
		EpicID:        state.EpicID,
		FastPath:      state.FastPath,
		TestFirst:     state.TestFirst,
		Interactive:   phasedInteractive,
		PhaseNum:      phaseNum,
		ContextBudget: budget,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	skillInvocation := buf.String()

	// Build prompt: summary contract first, then context, then skill invocation.
	// Early instructions survive compaction better than trailing ones.
	var prompt strings.Builder

	// 1. Context discipline instruction (first — survives compaction)
	disciplineTmpl, err := template.New("discipline").Parse(contextDisciplineInstruction)
	if err == nil {
		if err := disciplineTmpl.Execute(&prompt, data); err != nil {
			VerbosePrintf("Warning: could not render context discipline instruction: %v\n", err)
		}
	}

	// 2. Summary instruction
	summaryTmpl, err := template.New("summary").Parse(phaseSummaryInstruction)
	if err == nil {
		if err := summaryTmpl.Execute(&prompt, data); err != nil {
			VerbosePrintf("Warning: could not render summary instruction: %v\n", err)
		}
	}

	// 3. Cross-phase context for phases 3+ (goal, verdicts, prior summaries)
	if phaseNum >= 3 {
		ctx := buildPhaseContext(cwd, state, phaseNum)
		if ctx != "" {
			prompt.WriteString(ctx)
			prompt.WriteString("\n\n")
		}
	}

	// 4. Skill invocation (last — the actual command)
	prompt.WriteString(skillInvocation)

	return prompt.String(), nil
}

// buildPhaseContext constructs a context block from goal, verdicts, and prior phase summaries.
func buildPhaseContext(cwd string, state *phasedState, phaseNum int) string {
	var parts []string

	// Always include the goal
	if state.Goal != "" {
		parts = append(parts, fmt.Sprintf("Goal: %s", state.Goal))
	}

	// Include prior verdicts
	for key, verdict := range state.Verdicts {
		parts = append(parts, fmt.Sprintf("%s verdict: %s", strings.ReplaceAll(key, "_", "-"), verdict))
	}

	// Include prior phase summaries (read from disk)
	if cwd != "" {
		summaries := readPhaseSummaries(cwd, phaseNum)
		if summaries != "" {
			parts = append(parts, summaries)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return "--- RPI Context (from prior phases) ---\n" + strings.Join(parts, "\n")
}

// readPhaseSummaries reads all phase summary files prior to the given phase.
func readPhaseSummaries(cwd string, currentPhase int) string {
	var summaries []string
	rpiDir := filepath.Join(cwd, ".agents", "rpi")

	for i := 1; i < currentPhase; i++ {
		path := filepath.Join(rpiDir, fmt.Sprintf("phase-%d-summary.md", i))
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		content := strings.TrimSpace(string(data))
		if content == "" {
			continue
		}
		// Cap each summary to prevent context bloat
		if len(content) > 2000 {
			content = content[:2000] + "..."
		}
		phaseName := "unknown"
		if i > 0 && i <= len(phases) {
			phaseName = phases[i-1].Name
		}
		summaries = append(summaries, fmt.Sprintf("[Phase %d: %s]\n%s", i, phaseName, content))
	}

	if len(summaries) == 0 {
		return ""
	}
	return strings.Join(summaries, "\n\n")
}

// buildRetryPrompt constructs a retry prompt with feedback context.
func buildRetryPrompt(cwd string, phaseNum int, state *phasedState, retryCtx *retryContext) (string, error) {
	tmplStr, ok := retryPrompts[phaseNum]
	if !ok {
		// No retry template — fall back to normal prompt
		return buildPromptForPhase(cwd, phaseNum, state, retryCtx)
	}

	tmpl, err := template.New("retry").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse retry template: %w", err)
	}

	data := struct {
		Goal         string
		EpicID       string
		FastPath     bool
		TestFirst    bool
		RetryAttempt int
		MaxRetries   int
		Findings     []finding
	}{
		Goal:         state.Goal,
		EpicID:       state.EpicID,
		FastPath:     state.FastPath,
		TestFirst:    state.TestFirst,
		RetryAttempt: retryCtx.Attempt,
		MaxRetries:   phasedMaxRetries,
		Findings:     retryCtx.Findings,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute retry template: %w", err)
	}

	return buf.String(), nil
}

// Exit codes for phased orchestration.
const (
	ExitGateFail  = 10 // Council gate returned FAIL
	ExitUserAbort = 20 // User cancelled the session
	ExitCLIError  = 30 // Claude CLI error (not found, config issue)
)

// spawnClaudePhase spawns a fresh Claude session for a single phase.
// When ntm is available in PATH, wraps the session in a named tmux pane
// (ao-rpi-<runID>-p<N>) for live observability via ntm attach.
// Falls back to direct exec when ntm is unavailable.
// Strips CLAUDECODE env var so the child session doesn't trigger the
// nesting guard — these are independent sequential sessions, not nested.
// cmd.Dir is set to cwd for worktree isolation. GIT_DIR/GIT_WORK_TREE are
// NOT set — git auto-discovers the repo from the working directory.
func spawnClaudePhase(prompt, cwd, runID string, phaseNum int) error {
	// Check if ntm is available for observable sessions
	ntmPath, ntmErr := lookPath("ntm")
	if ntmErr == nil {
		return spawnClaudePhaseNtm(ntmPath, prompt, cwd, runID, phaseNum)
	}
	return spawnDirectFn(prompt, cwd)
}

// spawnClaudeDirectImpl runs claude -p directly (fallback when ntm unavailable).
func spawnClaudeDirectImpl(prompt, cwd string) error {
	cmd := exec.Command("claude", "-p", prompt)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = cleanEnvNoClaude()
	err := cmd.Run()
	if err == nil {
		return nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("claude exited with code %d: %w", exitErr.ExitCode(), err)
	}
	return fmt.Errorf("claude execution failed: %w", err)
}

// spawnClaudePhaseNtm wraps a claude session inside an ntm-managed tmux pane.
// Session name: ao-rpi-<runID>-p<phaseNum>. Attach with: ntm attach <name>.
func spawnClaudePhaseNtm(ntmPath, prompt, cwd, runID string, phaseNum int) error {
	sessionName := fmt.Sprintf("ao-rpi-%s-p%d", runID, phaseNum)
	fmt.Printf("ntm session: %s (attach with: ntm attach %s)\n", sessionName, sessionName)

	// Spawn ntm session with one claude agent
	spawnCmd := exec.Command(ntmPath, "spawn", sessionName, "--cc=1", "--no-user-pane", "--dir", cwd)
	spawnCmd.Env = cleanEnvNoClaude()
	if out, err := spawnCmd.CombinedOutput(); err != nil {
		fmt.Printf("ntm spawn failed, falling back to direct exec: %s\n", string(out))
		return spawnDirectFn(prompt, cwd)
	}

	// Send the prompt to the claude agent
	sendCmd := exec.Command(ntmPath, "send", sessionName, prompt)
	sendCmd.Env = cleanEnvNoClaude()
	if out, err := sendCmd.CombinedOutput(); err != nil {
		fmt.Printf("ntm send failed: %s\n", string(out))
		// Clean up session on failure
		_ = exec.Command(ntmPath, "kill", sessionName).Run()
		return fmt.Errorf("ntm send failed: %w", err)
	}

	// Poll for session completion (agent exits when prompt completes)
	for {
		time.Sleep(5 * time.Second)
		checkCmd := exec.Command("tmux", "has-session", "-t", sessionName)
		if err := checkCmd.Run(); err != nil {
			// Session gone — agent completed
			break
		}
	}

	// Clean up tmux session
	_ = exec.Command(ntmPath, "kill", sessionName).Run()
	return nil
}

// spawnClaudePhaseWithStream spawns a Claude session using --output-format stream-json
// and feeds stdout through ParseStreamEvents for live progress tracking.
// An onUpdate callback calls WriteLiveStatus after every parsed event so that
// external watchers (e.g. ao status) can tail the status file.
// Stderr is passed through to os.Stderr for real-time error visibility.
func spawnClaudePhaseWithStream(prompt, cwd, runID string, phaseNum int, statusPath string, allPhases []PhaseProgress) error {
	cmd := exec.Command("claude", "-p", prompt, "--output-format", "stream-json", "--verbose")
	cmd.Dir = cwd
	cmd.Stderr = os.Stderr
	cmd.Env = cleanEnvNoClaude()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start claude: %w", err)
	}

	// phaseIdx is 0-based for allPhases slice.
	phaseIdx := phaseNum - 1

	onUpdate := func(p PhaseProgress) {
		if phaseIdx >= 0 && phaseIdx < len(allPhases) {
			allPhases[phaseIdx] = p
		}
		if writeErr := WriteLiveStatus(statusPath, allPhases, phaseIdx); writeErr != nil {
			VerbosePrintf("Warning: could not write live status: %v\n", writeErr)
		}
	}

	_, parseErr := ParseStreamEvents(stdout, onUpdate)
	waitErr := cmd.Wait()

	// Prefer wait error (exit code) over parse error.
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			return fmt.Errorf("claude exited with code %d: %w", exitErr.ExitCode(), waitErr)
		}
		return fmt.Errorf("claude execution failed: %w", waitErr)
	}
	if parseErr != nil {
		return fmt.Errorf("stream parse error: %w", parseErr)
	}
	return nil
}

// buildAllPhases constructs a []PhaseProgress with Name fields populated
// from the global phases slice, used as the initial state for live status tracking.
func buildAllPhases(phaseDefs []phase) []PhaseProgress {
	all := make([]PhaseProgress, len(phaseDefs))
	for i, p := range phaseDefs {
		all[i] = PhaseProgress{Name: p.Name}
	}
	return all
}

// cleanEnvNoClaude builds a clean env without CLAUDECODE to avoid nesting guard.
func cleanEnvNoClaude() []string {
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "CLAUDECODE=") {
			env = append(env, e)
		}
	}
	return env
}

// --- Verdict extraction helpers ---

// extractCouncilVerdict reads a council report and returns the verdict (PASS/WARN/FAIL).
func extractCouncilVerdict(reportPath string) (string, error) {
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return "", fmt.Errorf("read report: %w", err)
	}

	re := regexp.MustCompile(`(?m)^## Council Verdict:\s*(PASS|WARN|FAIL)`)
	matches := re.FindSubmatch(data)
	if len(matches) < 2 {
		return "", fmt.Errorf("no verdict found in %s", reportPath)
	}
	return string(matches[1]), nil
}

// findLatestCouncilReport finds the most recent council report matching a pattern.
// When epicID is non-empty, reports whose filename contains the epicID are preferred.
// If no epic-scoped report is found, all pattern-matching reports are used as fallback.
func findLatestCouncilReport(cwd string, pattern string, notBefore time.Time, epicID string) (string, error) {
	councilDir := filepath.Join(cwd, ".agents", "council")
	entries, err := os.ReadDir(councilDir)
	if err != nil {
		return "", fmt.Errorf("read council directory: %w", err)
	}

	var matches []string
	var epicMatches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.Contains(name, pattern) && strings.HasSuffix(name, ".md") {
			if !notBefore.IsZero() {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				if info.ModTime().Before(notBefore) {
					continue
				}
			}
			fullPath := filepath.Join(councilDir, name)
			matches = append(matches, fullPath)
			if epicID != "" && strings.Contains(name, epicID) {
				epicMatches = append(epicMatches, fullPath)
			}
		}
	}

	// Prefer epic-scoped matches when available.
	selected := matches
	if len(epicMatches) > 0 {
		selected = epicMatches
	}

	if len(selected) == 0 {
		return "", fmt.Errorf("no council report matching %q found", pattern)
	}

	sort.Strings(selected)

	return selected[len(selected)-1], nil
}

// extractCouncilFindings extracts structured findings from a council report.
func extractCouncilFindings(reportPath string, max int) ([]finding, error) {
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, fmt.Errorf("read report: %w", err)
	}

	// Look for structured findings: FINDING: ... | FIX: ... | REF: ...
	re := regexp.MustCompile(`(?m)FINDING:\s*(.+?)\s*\|\s*FIX:\s*(.+?)\s*\|\s*REF:\s*(.+?)$`)
	allMatches := re.FindAllSubmatch(data, -1)

	var findings []finding
	for i, m := range allMatches {
		if i >= max {
			break
		}
		findings = append(findings, finding{
			Description: string(m[1]),
			Fix:         string(m[2]),
			Ref:         string(m[3]),
		})
	}

	// Fallback: if no structured findings, extract from "## Shared Findings" section
	if len(findings) == 0 {
		re2 := regexp.MustCompile(`(?m)^\d+\.\s+\*\*(.+?)\*\*\s*[—–-]\s*(.+)$`)
		allMatches2 := re2.FindAllSubmatch(data, -1)
		for i, m := range allMatches2 {
			if i >= max {
				break
			}
			findings = append(findings, finding{
				Description: string(m[1]) + ": " + string(m[2]),
				Fix:         "See council report",
				Ref:         reportPath,
			})
		}
	}

	return findings, nil
}

// --- Epic and completion helpers ---

// extractEpicID finds the most recent open epic ID via bd CLI.
func extractEpicID() (string, error) {
	cmd := exec.Command("bd", "list", "--type", "epic", "--status", "open")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("bd list: %w", err)
	}

	re := regexp.MustCompile(`(ag-[a-z0-9]+)`)
	matches := re.FindSubmatch(out)
	if len(matches) < 2 {
		return "", fmt.Errorf("no epic found in bd list output")
	}
	return string(matches[1]), nil
}

// detectFastPath checks if an epic is a micro-epic (≤2 issues, no blockers).
func detectFastPath(epicID string) (bool, error) {
	cmd := exec.Command("bd", "children", epicID)
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("bd children: %w", err)
	}
	return parseFastPath(string(out)), nil
}

// parseFastPath determines if bd children output indicates a micro-epic.
func parseFastPath(output string) bool {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	issueCount := 0
	blockedCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		issueCount++
		if strings.Contains(strings.ToLower(line), "blocked") {
			blockedCount++
		}
	}
	return issueCount <= 2 && blockedCount == 0
}

// checkCrankCompletion checks epic completion via bd children statuses.
// Returns "DONE", "BLOCKED", or "PARTIAL".
func checkCrankCompletion(epicID string) (string, error) {
	cmd := exec.Command("bd", "children", epicID)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("bd children: %w", err)
	}
	return parseCrankCompletion(string(out)), nil
}

// parseCrankCompletion determines completion status from bd children output.
func parseCrankCompletion(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	total := 0
	closed := 0
	blocked := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		total++
		lower := strings.ToLower(line)
		if strings.Contains(lower, "closed") || strings.Contains(lower, "✓") {
			closed++
		}
		if strings.Contains(lower, "blocked") {
			blocked++
		}
	}

	if total == 0 {
		return "DONE"
	}
	if closed == total {
		return "DONE"
	}
	if blocked > 0 {
		return "BLOCKED"
	}
	return "PARTIAL"
}

// --- Phase summaries ---

// writePhaseSummary writes a fallback summary only if Claude didn't write one.
func writePhaseSummary(cwd string, state *phasedState, phaseNum int) {
	rpiDir := filepath.Join(cwd, ".agents", "rpi")
	path := filepath.Join(rpiDir, fmt.Sprintf("phase-%d-summary.md", phaseNum))

	// If Claude already wrote a summary, keep it (it's richer than our mechanical one)
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Phase %d: Claude-written summary found\n", phaseNum)
		return
	}
	fmt.Printf("Phase %d: no Claude summary found, writing fallback\n", phaseNum)

	if err := os.MkdirAll(rpiDir, 0755); err != nil {
		VerbosePrintf("Warning: could not create rpi dir for summary: %v\n", err)
		return
	}

	summary := generatePhaseSummary(state, phaseNum)
	if summary == "" {
		return
	}

	if err := os.WriteFile(path, []byte(summary), 0644); err != nil {
		VerbosePrintf("Warning: could not write phase summary: %v\n", err)
	}
}

// generatePhaseSummary produces a concise summary of what a phase accomplished.
func generatePhaseSummary(state *phasedState, phaseNum int) string {
	switch phaseNum {
	case 1: // Research
		return fmt.Sprintf("Research completed for goal: %s\nSee .agents/research/ for findings.", state.Goal)
	case 2: // Plan
		summary := fmt.Sprintf("Plan completed. Epic: %s", state.EpicID)
		if state.FastPath {
			summary += " (micro-epic, fast path)"
		}
		return summary
	case 3: // Pre-mortem
		verdict := state.Verdicts["pre_mortem"]
		if verdict == "" {
			verdict = "unknown"
		}
		return fmt.Sprintf("Pre-mortem verdict: %s\nSee .agents/council/*pre-mortem*.md for details.", verdict)
	case 4: // Crank
		return fmt.Sprintf("Crank completed for epic %s.\nCheck bd children %s for issue statuses.", state.EpicID, state.EpicID)
	case 5: // Vibe
		verdict := state.Verdicts["vibe"]
		if verdict == "" {
			verdict = "unknown"
		}
		return fmt.Sprintf("Vibe verdict: %s\nSee .agents/council/*vibe*.md for details.", verdict)
	case 6: // Post-mortem
		return fmt.Sprintf("Post-mortem completed for epic %s.\nSee .agents/council/*post-mortem*.md and .agents/learnings/ for extracted knowledge.", state.EpicID)
	}
	return ""
}

// handoffDetected checks if a phase wrote a handoff file (context degradation signal).
func handoffDetected(cwd string, phaseNum int) bool {
	path := filepath.Join(cwd, ".agents", "rpi", fmt.Sprintf("phase-%d-handoff.md", phaseNum))
	_, err := os.Stat(path)
	return err == nil
}

// cleanPhaseSummaries removes stale phase summaries and handoffs from a prior run.
func cleanPhaseSummaries(stateDir string) {
	for i := 1; i <= 6; i++ {
		path := filepath.Join(stateDir, fmt.Sprintf("phase-%d-summary.md", i))
		os.Remove(path) //nolint:errcheck
		handoffPath := filepath.Join(stateDir, fmt.Sprintf("phase-%d-handoff.md", i))
		os.Remove(handoffPath) //nolint:errcheck
	}
}

// --- Worktree isolation ---

// worktreeTimeout is the timeout for git worktree operations (matches Olympus DefaultTimeout).
const worktreeTimeout = 30 * time.Second

// generateRunID returns a 12-char lowercase hex string from crypto/rand.
func generateRunID() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based if crypto/rand fails (shouldn't happen).
		return fmt.Sprintf("%012x", time.Now().UnixNano()&0xffffffffffff)
	}
	return hex.EncodeToString(b)
}

// getCurrentBranch returns the current branch name, or error if detached HEAD.
func getCurrentBranch(repoRoot string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), worktreeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("git rev-parse timed out after %s", worktreeTimeout)
		}
		return "", fmt.Errorf("get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(out))
	if branch == "HEAD" {
		return "", fmt.Errorf("detached HEAD: worktree requires a named branch")
	}
	return branch, nil
}

// getRepoRoot returns the git repository root directory.
func getRepoRoot() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), worktreeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("git rev-parse timed out after %s", worktreeTimeout)
		}
		return "", fmt.Errorf("get repo root: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// createWorktree creates a sibling git worktree for isolated RPI execution.
// Path: ../<repo-basename>-rpi-<runID>/
// Branch: rpi/<runID>
func createWorktree(cwd string) (worktreePath, runID string, err error) {
	repoRoot, err := getRepoRoot()
	if err != nil {
		return "", "", err
	}

	currentBranch, err := getCurrentBranch(repoRoot)
	if err != nil {
		return "", "", err
	}

	// Retry up to 3 times in case of branch collision (astronomically unlikely with crypto/rand).
	for attempt := 0; attempt < 3; attempt++ {
		runID = generateRunID()
		repoBasename := filepath.Base(repoRoot)
		worktreePath = filepath.Join(filepath.Dir(repoRoot), repoBasename+"-rpi-"+runID)
		branchName := "rpi/" + runID

		ctx, cancel := context.WithTimeout(context.Background(), worktreeTimeout)
		cmd := exec.CommandContext(ctx, "git", "worktree", "add", "-b", branchName, worktreePath, currentBranch)
		cmd.Dir = repoRoot
		output, cmdErr := cmd.CombinedOutput()
		cancel()

		if cmdErr == nil {
			// Create .agents/rpi/ inside worktree for state files.
			if mkErr := os.MkdirAll(filepath.Join(worktreePath, ".agents", "rpi"), 0755); mkErr != nil {
				// Non-fatal: phase session will create it if needed.
				VerbosePrintf("Warning: could not create .agents/rpi/ in worktree: %v\n", mkErr)
			}
			return worktreePath, runID, nil
		}

		// Check if branch already exists (collision) — retry with new ID.
		if strings.Contains(string(output), "already exists") {
			VerbosePrintf("Worktree branch collision on %s, retrying (%d/3)\n", branchName, attempt+1)
			continue
		}

		if ctx.Err() == context.DeadlineExceeded {
			return "", "", fmt.Errorf("git worktree add timed out after %s", worktreeTimeout)
		}
		return "", "", fmt.Errorf("git worktree add failed: %w (output: %s)", cmdErr, string(output))
	}
	return "", "", fmt.Errorf("failed to create unique worktree branch after 3 attempts")
}

// mergeWorktree merges the RPI worktree branch back into the original branch.
// Retries the pre-merge dirty check with backoff to handle the race where
// another parallel run is mid-merge (repo momentarily dirty).
func mergeWorktree(repoRoot, runID string) error {
	// Retry dirty check up to 5 times with 2s backoff.
	// Another parallel run's merge takes <1s, so 10s total wait is generous.
	var dirtyErr error
	for attempt := 0; attempt < 5; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), worktreeTimeout)
		checkCmd := exec.CommandContext(ctx, "git", "diff-index", "--quiet", "HEAD")
		checkCmd.Dir = repoRoot
		dirtyErr = checkCmd.Run()
		cancel()

		if dirtyErr == nil {
			break // Repo is clean, proceed to merge.
		}
		if attempt < 4 {
			VerbosePrintf("Repo dirty (another merge in progress?), retrying in 2s (%d/5)\n", attempt+1)
			time.Sleep(2 * time.Second)
		}
	}
	if dirtyErr != nil {
		return fmt.Errorf("original repo has uncommitted changes after 5 retries: commit or stash before merge")
	}

	// Merge the worktree branch.
	ctx, cancel := context.WithTimeout(context.Background(), worktreeTimeout)
	defer cancel()

	branchName := "rpi/" + runID
	mergeMsg := fmt.Sprintf("Merge %s (ao rpi phased worktree)", branchName)
	mergeCmd := exec.CommandContext(ctx, "git", "merge", "--no-ff", "-m", mergeMsg, branchName)
	mergeCmd.Dir = repoRoot
	if err := mergeCmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("git merge timed out after %s", worktreeTimeout)
		}
		// Detect conflict files.
		conflictCmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
		conflictCmd.Dir = repoRoot
		conflictOut, _ := conflictCmd.Output()
		// Abort the merge to leave repo clean.
		abortCmd := exec.Command("git", "merge", "--abort")
		abortCmd.Dir = repoRoot
		_ = abortCmd.Run() //nolint:errcheck
		files := strings.TrimSpace(string(conflictOut))
		if files != "" {
			return fmt.Errorf("merge conflict in %s.\nConflicting files:\n%s\nResolve manually: cd %s && git merge %s",
				branchName, files, repoRoot, branchName)
		}
		return fmt.Errorf("git merge failed: %w", err)
	}
	return nil
}

// removeWorktree removes a worktree directory and its branch.
// Modeled on Olympus internal/git/worktree.go Remove().
func removeWorktree(repoRoot, worktreePath, runID string) error {
	// Structural path validation: worktree must be a sibling of repoRoot
	// with basename matching <repoBasename>-rpi-<runID>.
	// Use EvalSymlinks to resolve macOS /var → /private/var and similar.
	absPath, err := filepath.EvalSymlinks(worktreePath)
	if err != nil {
		// Path may already be removed; try Abs as fallback.
		absPath, err = filepath.Abs(worktreePath)
		if err != nil {
			return fmt.Errorf("invalid worktree path: %w", err)
		}
	}
	resolvedRoot, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		resolvedRoot = repoRoot
	}
	expectedBasename := filepath.Base(resolvedRoot) + "-rpi-" + runID
	expectedPath := filepath.Join(filepath.Dir(resolvedRoot), expectedBasename)
	if absPath != expectedPath {
		return fmt.Errorf("refusing to remove %s: expected %s (path validation failed)", absPath, expectedPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), worktreeTimeout)
	defer cancel()

	// Remove worktree via git.
	cmd := exec.CommandContext(ctx, "git", "worktree", "remove", absPath, "--force")
	cmd.Dir = repoRoot
	if _, err := cmd.CombinedOutput(); err != nil {
		// Fallback: direct removal if git fails.
		_ = os.RemoveAll(absPath) //nolint:errcheck
	}

	// Delete branch (only rpi/* prefix for safety).
	branchName := "rpi/" + runID
	branchCmd := exec.CommandContext(ctx, "git", "branch", "-D", branchName)
	branchCmd.Dir = repoRoot
	_ = branchCmd.Run() //nolint:errcheck — branch may not exist

	return nil
}

// --- State persistence ---

const phasedStateFile = "phased-state.json"

// savePhasedState writes orchestrator state to disk.
func savePhasedState(cwd string, state *phasedState) error {
	stateDir := filepath.Join(cwd, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	path := filepath.Join(stateDir, phasedStateFile)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}

	VerbosePrintf("State saved to %s\n", path)
	return nil
}

// loadPhasedState reads orchestrator state from disk.
func loadPhasedState(cwd string) (*phasedState, error) {
	path := filepath.Join(cwd, ".agents", "rpi", phasedStateFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}

	var state phasedState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}

	// Ensure maps are never nil after deserialization
	if state.Verdicts == nil {
		state.Verdicts = make(map[string]string)
	}
	if state.Attempts == nil {
		state.Attempts = make(map[string]int)
	}

	return &state, nil
}

// --- Ratchet and logging ---

// recordRatchetCheckpoint records a ratchet checkpoint for a phase.
func recordRatchetCheckpoint(step string) {
	cmd := exec.Command("ao", "ratchet", "record", step)
	if err := cmd.Run(); err != nil {
		VerbosePrintf("Warning: ratchet record %s: %v\n", step, err)
	}
}

// logPhaseTransition appends a log entry to the orchestration log.
func logPhaseTransition(logPath, runID, phase, details string) {
	var entry string
	if runID != "" {
		entry = fmt.Sprintf("[%s] [%s] %s: %s\n", time.Now().Format(time.RFC3339), runID, phase, details)
	} else {
		entry = fmt.Sprintf("[%s] %s: %s\n", time.Now().Format(time.RFC3339), phase, details)
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		VerbosePrintf("Warning: could not write orchestration log: %v\n", err)
		return
	}
	defer f.Close() //nolint:errcheck

	if _, err := f.WriteString(entry); err != nil {
		VerbosePrintf("Warning: could not write log entry: %v\n", err)
	}
}

// --- Agent mail observability ---

// lookPath is the function used to resolve binary paths. Package-level for testability.
var lookPath = exec.LookPath

// spawnDirectFn is the function used to spawn claude directly. Package-level for testability.
var spawnDirectFn = spawnClaudeDirectImpl

// gtPath caches the resolved path to the gt binary, or empty string if not found.
var gtPath string

func init() {
	gtPath, _ = lookPath("gt")
}

// registerRPIAgent registers this RPI run with agent mail for observability.
// Fails silently if gt is not on PATH or the command errors.
func registerRPIAgent(runID string) {
	if gtPath == "" {
		return
	}
	_ = exec.Command(gtPath, "mail", "register", "rpi-"+runID).Run() //nolint:errcheck
}

// emitRPIStatus sends a status message to mayor via agent mail.
// Fails silently if gt is not on PATH or the command errors.
func emitRPIStatus(runID, phaseName, status string) {
	if gtPath == "" {
		return
	}
	msg := fmt.Sprintf("rpi-%s: %s %s", runID, phaseName, status)
	_ = exec.Command(gtPath, "mail", "send", "mayor", msg).Run() //nolint:errcheck
}

// deregisterRPIAgent deregisters this RPI run from agent mail.
// Fails silently if gt is not on PATH or the command errors.
func deregisterRPIAgent(runID string) {
	if gtPath == "" {
		return
	}
	_ = exec.Command(gtPath, "mail", "deregister", "rpi-"+runID).Run() //nolint:errcheck
}

// --- Phase name helpers ---

// phaseNameToNum converts a phase name to its number (1-6).
func phaseNameToNum(name string) int {
	normalized := strings.ToLower(strings.TrimSpace(name))
	aliases := map[string]int{
		"research":    1,
		"plan":        2,
		"pre-mortem":  3,
		"premortem":   3,
		"pre_mortem":  3,
		"crank":       4,
		"implement":   4,
		"vibe":        5,
		"validate":    5,
		"post-mortem": 6,
		"postmortem":  6,
		"post_mortem": 6,
	}
	return aliases[normalized]
}
