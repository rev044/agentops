package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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
		}
	}

	stateDir := filepath.Join(cwd, ".agents", "rpi")
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
	logPhaseTransition(logPath, "start", fmt.Sprintf("goal=%q from=%s", state.Goal, phasedFrom))

	// Execute phases sequentially
	for i := startPhase; i <= 6; i++ {
		p := phases[i-1]
		fmt.Printf("\n--- Phase %d: %s ---\n", p.Num, p.Name)
		state.Phase = i

		prompt, err := buildPromptForPhase(cwd, i, state, nil)
		if err != nil {
			return fmt.Errorf("build prompt for phase %d: %w", i, err)
		}

		if GetDryRun() {
			fmt.Printf("[dry-run] Would spawn: claude -p '%s'\n", prompt)
			logPhaseTransition(logPath, p.Name, "dry-run")
			continue
		}

		// Spawn phase session
		fmt.Printf("Spawning: claude -p '%s'\n", prompt)
		start := time.Now()

		if err := spawnClaudePhase(prompt); err != nil {
			logPhaseTransition(logPath, p.Name, fmt.Sprintf("FAILED: %v", err))
			return fmt.Errorf("phase %d (%s) failed: %w", i, p.Name, err)
		}

		elapsed := time.Since(start).Round(time.Second)
		fmt.Printf("Phase %d completed in %s\n", i, elapsed)
		logPhaseTransition(logPath, p.Name, fmt.Sprintf("completed in %s", elapsed))

		// Post-phase processing
		if err := postPhaseProcessing(cwd, state, i, logPath); err != nil {
			// Check if it's a gate failure that needs retry
			if retryErr, ok := err.(*gateFailError); ok {
				retried, retryErr2 := handleGateRetry(cwd, state, i, retryErr, logPath)
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
		if handoffDetected(cwd, i) {
			fmt.Printf("Phase %d: handoff detected — phase reported context degradation\n", i)
			logPhaseTransition(logPath, p.Name, "HANDOFF detected — context degradation")
			// Continue to next phase (fresh session will pick up from handoff)
		}

		// Write phase summary for next phase's context
		writePhaseSummary(cwd, state, i)

		// Record ratchet checkpoint
		recordRatchetCheckpoint(p.Step)

		// Save state
		if err := savePhasedState(cwd, state); err != nil {
			VerbosePrintf("Warning: could not save state: %v\n", err)
		}
	}

	// Final report
	fmt.Printf("\n=== RPI Phased Complete ===\n")
	fmt.Printf("Goal: %s\n", state.Goal)
	if state.EpicID != "" {
		fmt.Printf("Epic: %s\n", state.EpicID)
	}
	fmt.Printf("Verdicts: %v\n", state.Verdicts)
	logPhaseTransition(logPath, "complete", fmt.Sprintf("epic=%s verdicts=%v", state.EpicID, state.Verdicts))

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
		report, err := findLatestCouncilReport(cwd, "pre-mortem")
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
		report, err := findLatestCouncilReport(cwd, "vibe")
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
func handleGateRetry(cwd string, state *phasedState, phaseNum int, gateErr *gateFailError, logPath string) (bool, error) {
	phaseName := phases[phaseNum-1].Name
	attemptKey := fmt.Sprintf("phase_%d", phaseNum)

	state.Attempts[attemptKey]++
	attempt := state.Attempts[attemptKey]

	if attempt >= phasedMaxRetries {
		msg := fmt.Sprintf("%s failed %d times. Last report: %s. Manual intervention needed.",
			phaseName, phasedMaxRetries, gateErr.Report)
		fmt.Println(msg)
		logPhaseTransition(logPath, phaseName, msg)
		return false, nil
	}

	fmt.Printf("%s: %s (attempt %d/%d) — retrying\n", phaseName, gateErr.Verdict, attempt, phasedMaxRetries)
	logPhaseTransition(logPath, phaseName, fmt.Sprintf("RETRY attempt %d/%d", attempt+1, phasedMaxRetries))

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
	if err := spawnClaudePhase(retryPrompt); err != nil {
		return false, fmt.Errorf("retry failed: %w", err)
	}

	// Re-run the original phase after retry
	rerunPrompt, err := buildPromptForPhase(cwd, phaseNum, state, nil)
	if err != nil {
		return false, fmt.Errorf("build rerun prompt: %w", err)
	}

	fmt.Printf("Re-running phase %d after retry\n", phaseNum)
	if err := spawnClaudePhase(rerunPrompt); err != nil {
		return false, fmt.Errorf("rerun failed: %w", err)
	}

	// Check gate again
	if err := postPhaseProcessing(cwd, state, phaseNum, logPath); err != nil {
		if _, ok := err.(*gateFailError); ok {
			// Still failing — recurse
			return handleGateRetry(cwd, state, phaseNum, err.(*gateFailError), logPath)
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
func spawnClaudePhase(prompt string) error {
	cmd := exec.Command("claude", "-p", prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err == nil {
		return nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		code := exitErr.ExitCode()
		return fmt.Errorf("claude exited with code %d: %w", code, err)
	}
	return fmt.Errorf("claude execution failed: %w", err)
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
func findLatestCouncilReport(cwd string, pattern string) (string, error) {
	councilDir := filepath.Join(cwd, ".agents", "council")
	entries, err := os.ReadDir(councilDir)
	if err != nil {
		return "", fmt.Errorf("read council directory: %w", err)
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.Contains(name, pattern) && strings.HasSuffix(name, ".md") {
			matches = append(matches, filepath.Join(councilDir, name))
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no council report matching %q found", pattern)
	}

	sort.Strings(matches)

	return matches[len(matches)-1], nil
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
func logPhaseTransition(logPath, phase, details string) {
	entry := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format(time.RFC3339), phase, details)

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
