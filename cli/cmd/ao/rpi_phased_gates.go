package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/boshu2/agentops/cli/internal/types"
)

// maxGateRetryDepth is the hard ceiling for gate retry attempts.
// Attempts 1 through maxGateRetryDepth proceed normally; attempt
// maxGateRetryDepth+1 forces escalation regardless of MemRL policy.
// Set higher than default MaxRetries (3) to catch runaway MemRL
// policy overrides without capping normal retries.
const maxGateRetryDepth = 5

// shouldForceEscalation returns true when the attempt count exceeds the hard ceiling.
func shouldForceEscalation(attempt int) bool {
	return attempt > maxGateRetryDepth
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
	case 1:
		return processDiscoveryPhase(cwd, state, logPath)
	case 2:
		return processImplementationPhase(cwd, state, phaseNum, logPath)
	case 3:
		return processValidationPhase(cwd, state, phaseNum, logPath)
	}
	return nil
}

// processDiscoveryPhase handles post-processing for the discovery phase.
// It extracts the epic ID, detects fast path, and checks the pre-mortem verdict.
func processDiscoveryPhase(cwd string, state *phasedState, logPath string) error {
	epicID, err := extractEpicID(state.Opts.BDCommand)
	if err != nil {
		// Fallback 1: discover plan file when bd has no epic
		planPath, planErr := discoverPlanFile(cwd)
		if planErr == nil {
			epicID = planFileEpicPrefix + planPath
			fmt.Printf("Plan-file fallback: using %s as epic ID\n", planPath)
		} else {
			// Fallback 2: any open issue (handles small-scope tasks that aren't epics)
			issueID, issueErr := extractAnyOpenIssueID(state.Opts.BDCommand)
			if issueErr != nil {
				return fmt.Errorf("discovery phase: could not find epic, plan file, or open issue: %w", err)
			}
			epicID = issueID
			fmt.Printf("Single-issue fallback: using %s (not an epic)\n", epicID)
		}
	}
	state.EpicID = epicID
	fmt.Printf("Epic ID: %s\n", epicID)
	logPhaseTransition(logPath, state.RunID, "discovery", fmt.Sprintf("extracted epic: %s", epicID))

	if !state.Opts.FastPath && !isPlanFileEpic(epicID) {
		fast, err := detectFastPath(state.EpicID, state.Opts.BDCommand)
		if err != nil {
			VerbosePrintf("Warning: fast-path detection failed (continuing without): %v\n", err)
		} else if fast {
			state.FastPath = true
			fmt.Println("Micro-epic detected — using fast path (--quick for gates)")
		}
	}

	// For plan-file epics, don't try to match epic ID in council report filenames
	councilEpicID := state.EpicID
	if isPlanFileEpic(state.EpicID) {
		councilEpicID = ""
	}
	report, err := findLatestCouncilReport(cwd, "pre-mortem", time.Time{}, councilEpicID)
	if err != nil {
		// Pre-mortem may not have run if the session handled retries internally
		// and ultimately gave up. Check if council report exists at all.
		VerbosePrintf("Warning: pre-mortem council report not found (session may have handled retries internally): %v\n", err)
		return nil
	}
	verdict, err := extractCouncilVerdict(report)
	if err != nil {
		VerbosePrintf("Warning: could not extract pre-mortem verdict: %v\n", err)
		return nil
	}
	state.Verdicts["pre_mortem"] = verdict
	fmt.Printf("Pre-mortem verdict: %s\n", verdict)
	_, _ = appendRPIC2Event(cwd, rpiC2EventInput{
		RunID:   state.RunID,
		Phase:   1,
		Type:    "gate.discovery.verdict",
		Message: fmt.Sprintf("Pre-mortem verdict: %s", verdict),
		Details: map[string]any{"verdict": verdict, "report": report},
	})
	logPhaseTransition(logPath, state.RunID, "discovery", fmt.Sprintf("pre-mortem verdict: %s report=%s", verdict, report))

	if verdict == "FAIL" {
		// Discovery session was instructed to retry internally.
		// If we still see FAIL here, it means all retries failed.
		findings, _ := extractCouncilFindings(report, 5)
		return &gateFailError{Phase: 1, Verdict: verdict, Findings: findings, Report: report}
	}
	return nil
}

// processImplementationPhase handles post-processing for the implementation phase.
// It validates the prior phase result and checks crank completion status.
func processImplementationPhase(cwd string, state *phasedState, phaseNum int, logPath string) error {
	if state.StartPhase <= 1 {
		if err := validatePriorPhaseResult(cwd, 1); err != nil {
			return fmt.Errorf("phase %d prerequisite not met: %w", phaseNum, err)
		}
	}
	if state.EpicID == "" {
		return nil
	}
	// Plan-file mode: skip bd-dependent completion check
	if isPlanFileEpic(state.EpicID) {
		return nil
	}
	if isEpic, err := isEpicIssue(state.EpicID, state.Opts.BDCommand); err == nil && !isEpic {
		fmt.Printf("Crank status: SKIP (non-epic issue %s)\n", state.EpicID)
		logPhaseTransition(logPath, state.RunID, "implementation", fmt.Sprintf("crank status: SKIP (non-epic issue %s)", state.EpicID))
		return nil
	} else if err != nil {
		VerbosePrintf("Warning: could not determine issue type for %s (continuing with crank completion check): %v\n", state.EpicID, err)
	}
	status, err := checkCrankCompletion(state.EpicID, state.Opts.BDCommand)
	if err != nil {
		VerbosePrintf("Warning: could not check crank completion (continuing to validation): %v\n", err)
		return nil
	}
	fmt.Printf("Crank status: %s\n", status)
	_, _ = appendRPIC2Event(cwd, rpiC2EventInput{
		RunID:   state.RunID,
		Phase:   2,
		Type:    "gate.implementation.verdict",
		Message: fmt.Sprintf("Crank status: %s", status),
		Details: map[string]any{"status": status, "epic_id": state.EpicID},
	})
	logPhaseTransition(logPath, state.RunID, "implementation", fmt.Sprintf("crank status: %s", status))
	if status == "BLOCKED" || status == "PARTIAL" {
		return &gateFailError{Phase: 2, Verdict: status, Report: "bd children " + state.EpicID}
	}
	return nil
}

// processValidationPhase handles post-processing for the validation phase.
// It validates the prior phase result, checks the vibe verdict, and optionally extracts
// the post-mortem verdict.
func processValidationPhase(cwd string, state *phasedState, phaseNum int, logPath string) error {
	if state.StartPhase <= 2 {
		if err := validatePriorPhaseResult(cwd, 2); err != nil {
			return fmt.Errorf("phase %d prerequisite not met: %w", phaseNum, err)
		}
	}
	report, err := findLatestCouncilReport(cwd, "vibe", time.Time{}, state.EpicID)
	if err != nil {
		return fmt.Errorf("validation phase: vibe report not found (phase may not have completed): %w", err)
	}
	verdict, err := extractCouncilVerdict(report)
	if err != nil {
		return fmt.Errorf("validation phase: could not extract vibe verdict from %s: %w", report, err)
	}
	state.Verdicts["vibe"] = verdict
	fmt.Printf("Vibe verdict: %s\n", verdict)
	_, _ = appendRPIC2Event(cwd, rpiC2EventInput{
		RunID:   state.RunID,
		Phase:   3,
		Type:    "gate.validation.verdict",
		Message: fmt.Sprintf("Vibe verdict: %s", verdict),
		Details: map[string]any{"verdict": verdict, "report": report},
	})
	logPhaseTransition(logPath, state.RunID, "validation", fmt.Sprintf("vibe verdict: %s report=%s", verdict, report))

	if verdict == "FAIL" {
		findings, _ := extractCouncilFindings(report, 5)
		return &gateFailError{Phase: 3, Verdict: verdict, Findings: findings, Report: report}
	}

	// Also extract post-mortem verdict if available (non-blocking)
	pmReport, err := findLatestCouncilReport(cwd, "post-mortem", time.Time{}, state.EpicID)
	if err == nil {
		pmVerdict, err := extractCouncilVerdict(pmReport)
		if err == nil {
			state.Verdicts["post_mortem"] = pmVerdict
			fmt.Printf("Post-mortem verdict: %s\n", pmVerdict)
			logPhaseTransition(logPath, state.RunID, "validation", fmt.Sprintf("post-mortem verdict: %s report=%s", pmVerdict, pmReport))
		}
	}
	return nil
}

func legacyGateAction(attempt, maxRetries int) types.MemRLAction {
	if attempt >= maxRetries {
		return types.MemRLActionEscalate
	}
	return types.MemRLActionRetry
}

func classifyGateFailureClass(phaseNum int, gateErr *gateFailError) types.MemRLFailureClass {
	if gateErr == nil {
		return ""
	}
	verdict := strings.ToUpper(strings.TrimSpace(gateErr.Verdict))
	if fc := classifyByPhase(phaseNum, verdict); fc != "" {
		return fc
	}
	return classifyByVerdict(verdict)
}

func resolveGateRetryAction(state *phasedState, phaseNum int, gateErr *gateFailError, attempt int) (types.MemRLAction, types.MemRLPolicyDecision) {
	mode := types.GetMemRLMode()
	failureClass := classifyGateFailureClass(phaseNum, gateErr)
	metadataPresent := gateErr != nil && strings.TrimSpace(gateErr.Verdict) != ""

	decision := types.EvaluateDefaultMemRLPolicy(types.MemRLPolicyInput{
		Mode:            mode,
		FailureClass:    failureClass,
		Attempt:         attempt,
		MaxAttempts:     state.Opts.MaxRetries,
		MetadataPresent: metadataPresent,
	})

	legacy := legacyGateAction(attempt, state.Opts.MaxRetries)
	if mode == types.MemRLModeEnforce {
		return decision.Action, decision
	}
	return legacy, decision
}

// handleGateRetry manages retry logic for failed gates.
// spawnCwd is the working directory for spawned claude sessions (may be worktree).
func handleGateRetry(ctx context.Context, cwd string, state *phasedState, phaseNum int, gateErr *gateFailError, logPath string, spawnCwd string, statusPath string, allPhases []PhaseProgress, executor PhaseExecutor) (bool, error) {
	phaseName := phases[phaseNum-1].Name
	attemptKey := fmt.Sprintf("phase_%d", phaseNum)

	state.Attempts[attemptKey]++
	attempt := state.Attempts[attemptKey]

	// Hard ceiling: force escalation regardless of MemRL policy
	if shouldForceEscalation(attempt) {
		return performGateEscalation(state, phaseNum, attempt, gateErr,
			types.MemRLPolicyDecision{}, types.MemRLActionEscalate,
			phaseName, logPath, statusPath, allPhases)
	}

	maybeUpdateLiveStatus(state, statusPath, allPhases, phaseNum, "retrying after "+gateErr.Verdict, attempt, "")

	action, decision := resolveGateRetryAction(state, phaseNum, gateErr, attempt)
	logGateRetryMemRL(logPath, state.RunID, phaseName, decision, action)

	if action == types.MemRLActionEscalate {
		return performGateEscalation(state, phaseNum, attempt, gateErr, decision, action, phaseName, logPath, statusPath, allPhases)
	}

	fmt.Printf("%s: %s (attempt %d/%d) — retrying\n", phaseName, gateErr.Verdict, attempt, state.Opts.MaxRetries)
	_, _ = appendRPIC2Event(cwd, rpiC2EventInput{
		RunID:   state.RunID,
		Phase:   phaseNum,
		Type:    "gate.retry.attempt",
		Message: fmt.Sprintf("%s retry attempt %d/%d", phaseName, attempt, state.Opts.MaxRetries),
		Details: map[string]any{"attempt": attempt, "max_retries": state.Opts.MaxRetries, "verdict": gateErr.Verdict},
	})
	logPhaseTransition(logPath, state.RunID, phaseName, fmt.Sprintf("RETRY attempt %d/%d verdict=%s report=%s", attempt, state.Opts.MaxRetries, gateErr.Verdict, gateErr.Report))

	retryCtx := &retryContext{
		Attempt:  attempt,
		Findings: gateErr.Findings,
		Verdict:  gateErr.Verdict,
	}

	retryPrompt, err := buildRetryPrompt(cwd, phaseNum, state, retryCtx)
	if err != nil {
		return false, fmt.Errorf("build retry prompt: %w", err)
	}

	if GetDryRun() {
		fmt.Printf("[dry-run] Would spawn retry: %s\n", formatRuntimePromptInvocation(effectiveRuntimeCommand(state.Opts.RuntimeCommand), retryPrompt))
		return false, nil
	}

	if err := executeWithStatus(ctx, executor, state, statusPath, allPhases, phaseNum, attempt, retryPrompt, spawnCwd, "running retry prompt", "retry failed"); err != nil {
		return false, fmt.Errorf("retry failed: %w", err)
	}

	rerunPrompt, err := buildPromptForPhase(cwd, phaseNum, state, nil)
	if err != nil {
		return false, fmt.Errorf("build rerun prompt: %w", err)
	}

	fmt.Printf("Re-running phase %d after retry\n", phaseNum)
	if err := executeWithStatus(ctx, executor, state, statusPath, allPhases, phaseNum, attempt, rerunPrompt, spawnCwd, "re-running phase", "rerun failed"); err != nil {
		return false, fmt.Errorf("rerun failed: %w", err)
	}

	return verifyGateAfterRetry(ctx, cwd, state, phaseNum, logPath, spawnCwd, statusPath, allPhases, executor, attempt)
}

func maybeUpdateLiveStatus(state *phasedState, statusPath string, allPhases []PhaseProgress, phaseNum int, status string, attempt int, errMsg string) {
	if state.Opts.LiveStatus {
		updateLivePhaseStatus(statusPath, allPhases, phaseNum, status, attempt, errMsg)
	}
}

func executeWithStatus(ctx context.Context, executor PhaseExecutor, state *phasedState, statusPath string, allPhases []PhaseProgress, phaseNum, attempt int, prompt, spawnCwd, runningMsg, failedMsg string) error {
	maybeUpdateLiveStatus(state, statusPath, allPhases, phaseNum, runningMsg, attempt, "")
	if err := executor.Execute(ctx, prompt, spawnCwd, state.RunID, phaseNum); err != nil {
		maybeUpdateLiveStatus(state, statusPath, allPhases, phaseNum, failedMsg, attempt, err.Error())
		return err
	}
	return nil
}

// logGateRetryMemRL logs the MemRL policy decision for a gate retry, if mode is not off.
func logGateRetryMemRL(logPath, runID, phaseName string, decision types.MemRLPolicyDecision, action types.MemRLAction) {
	if decision.Mode == types.MemRLModeOff {
		return
	}
	logPhaseTransition(
		logPath,
		runID,
		phaseName,
		fmt.Sprintf(
			"memrl policy mode=%s failure_class=%s attempt_bucket=%s policy_action=%s selected_action=%s rule=%s",
			decision.Mode,
			decision.FailureClass,
			decision.AttemptBucket,
			decision.Action,
			action,
			decision.RuleID,
		),
	)
}

// performGateEscalation handles the escalation path when the retry action is escalate.
// Returns (false, nil) to signal escalation without error (caller will handle reporting).
func performGateEscalation(state *phasedState, phaseNum, attempt int, gateErr *gateFailError, decision types.MemRLPolicyDecision, action types.MemRLAction, phaseName, logPath, statusPath string, allPhases []PhaseProgress) (bool, error) {
	msg := fmt.Sprintf(
		"%s escalated (mode=%s, action=%s, rule=%s, attempt=%d/%d). Last report: %s. Manual intervention needed.",
		phaseName,
		decision.Mode,
		action,
		decision.RuleID,
		attempt,
		state.Opts.MaxRetries,
		gateErr.Report,
	)
	fmt.Println(msg)
	escalationRoot := filepath.Dir(filepath.Dir(logPath))
	_, _ = appendRPIC2Event(escalationRoot, rpiC2EventInput{
		RunID:   state.RunID,
		Phase:   phaseNum,
		Type:    "gate.escalation",
		Message: msg,
		Details: map[string]any{"attempt": attempt, "report": gateErr.Report},
	})
	if state.Opts.LiveStatus {
		updateLivePhaseStatus(statusPath, allPhases, phaseNum, "failed after retries", attempt, gateErr.Report)
	}
	logPhaseTransition(logPath, state.RunID, phaseName, msg)
	return false, nil
}

// verifyGateAfterRetry re-checks the gate after a retry session completes.
// If the gate still fails, it recurses into handleGateRetry.
func verifyGateAfterRetry(ctx context.Context, cwd string, state *phasedState, phaseNum int, logPath, spawnCwd, statusPath string, allPhases []PhaseProgress, executor PhaseExecutor, attempt int) (bool, error) {
	if err := postPhaseProcessing(cwd, state, phaseNum, logPath); err != nil {
		var gateErr *gateFailError
		if errors.As(err, &gateErr) {
			// Still failing — recurse
			return handleGateRetry(ctx, cwd, state, phaseNum, gateErr, logPath, spawnCwd, statusPath, allPhases, executor)
		}
		return false, err
	}
	if state.Opts.LiveStatus {
		updateLivePhaseStatus(statusPath, allPhases, phaseNum, "retry succeeded", attempt, "")
	}
	return true, nil
}
