package rpi

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ShouldFallbackToDirect returns true when a stream error should trigger direct fallback.
func ShouldFallbackToDirect(err error, failReasonStall string) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "stream startup timeout") ||
		strings.Contains(msg, "stream parse error") ||
		strings.Contains(msg, "does not support stream-json") ||
		(strings.Contains(msg, failReasonStall) && strings.Contains(msg, "no stream activity"))
}

// RuntimeBinaryName extracts the lowercase base binary name from a runtime command string.
func RuntimeBinaryName(command string) string {
	executable, _ := SplitRuntimeCommand(command)
	if executable == "" {
		return ""
	}
	base := strings.ToLower(filepath.Base(executable))
	return strings.TrimSuffix(base, ".exe")
}

// SplitRuntimeCommand splits a runtime command into executable and prefix args.
func SplitRuntimeCommand(command string) (string, []string) {
	fields := strings.Fields(strings.TrimSpace(command))
	if len(fields) == 0 {
		return "", nil
	}
	return fields[0], fields[1:]
}

// RuntimeDirectCommandArgs builds the argument list for direct runtime execution.
func RuntimeDirectCommandArgs(command, prompt string) []string {
	_, prefixArgs := SplitRuntimeCommand(command)
	args := append([]string{}, prefixArgs...)
	if RuntimeBinaryName(command) == "codex" {
		return append(args, "exec", prompt)
	}
	return append(args, "-p", prompt)
}

// RuntimeStreamCommandArgs builds the argument list for stream-json runtime execution.
func RuntimeStreamCommandArgs(command, prompt string) ([]string, error) {
	if RuntimeBinaryName(command) == "codex" {
		return nil, fmt.Errorf("runtime %q does not support stream-json mode", command)
	}
	_, prefixArgs := SplitRuntimeCommand(command)
	args := append([]string{}, prefixArgs...)
	args = append(args, "-p", prompt, "--output-format", "stream-json", "--verbose")
	return args, nil
}

// FormatRuntimePromptInvocation formats a human-readable runtime invocation string.
func FormatRuntimePromptInvocation(command, prompt string) string {
	executable, _ := SplitRuntimeCommand(command)
	if executable == "" {
		executable = command
	}
	args := RuntimeDirectCommandArgs(command, prompt)
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, executable)
	for _, arg := range args {
		parts = append(parts, fmt.Sprintf("%q", arg))
	}
	return strings.Join(parts, " ")
}

// NormalizeCheckInterval returns checkInterval or a 1s default.
func NormalizeCheckInterval(checkInterval time.Duration) time.Duration {
	if checkInterval <= 0 {
		return 1 * time.Second
	}
	return checkInterval
}

// MergePhaseProgress updates dst with non-zero fields from src.
// PhaseProgress fields: Name, CurrentAction, RetryCount, LastError.
type PhaseProgressUpdate struct {
	Name          string
	CurrentAction string
	RetryCount    int
	LastError     string
}

// MergePhaseProgressFields merges non-zero fields from src into dst.
func MergePhaseProgressFields(dst, src *PhaseProgressUpdate) {
	if src.Name != "" {
		dst.Name = src.Name
	}
	if src.CurrentAction != "" {
		dst.CurrentAction = src.CurrentAction
	}
	if src.RetryCount != 0 {
		dst.RetryCount = src.RetryCount
	}
	if src.LastError != "" {
		dst.LastError = src.LastError
	}
}

// ClassifyStreamResult examines the context, wait error, and parse error to
// produce the appropriate error for a completed stream-json phase.
func ClassifyStreamResult(ctx, stallCtx context.Context, command string, phaseNum int, phaseTimeout time.Duration, waitErr, parseErr error, eventCount int64, failReasonTimeout, failReasonStall, failReasonExit, failReasonUnknown string) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("phase %d (%s) timed out after %s (set --phase-timeout to increase)", phaseNum, failReasonTimeout, phaseTimeout)
	}
	stallErr := stallCtx.Err()
	if stallErr != nil && ctx.Err() == nil {
		if cause := context.Cause(stallCtx); cause != nil {
			return fmt.Errorf("phase %d (%s): %w", phaseNum, failReasonStall, cause)
		}
	}
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			return fmt.Errorf("%s exited with code %d (%s): %w", command, exitErr.ExitCode(), failReasonExit, waitErr)
		}
		return fmt.Errorf("%s execution failed (%s): %w", command, failReasonUnknown, waitErr)
	}
	if parseErr != nil {
		return fmt.Errorf("stream parse error: %w", parseErr)
	}
	if eventCount == 0 {
		return fmt.Errorf("stream startup timeout: stream completed without parseable events")
	}
	return nil
}

// BuildStreamPhaseContext creates a context with optional timeout for a stream phase.
func BuildStreamPhaseContext(parent context.Context, phaseTimeout time.Duration) (context.Context, context.CancelFunc) {
	if phaseTimeout > 0 {
		return context.WithTimeout(parent, phaseTimeout)
	}
	return context.WithCancel(parent)
}

// BackendCapabilities probes the runtime environment for executor prerequisites.
type BackendCapabilities struct {
	LiveStatusEnabled bool
	RuntimeMode       string
}

// ProbeBackendCapabilities detects available backends in the current environment.
// Pure function — no side effects.
func ProbeBackendCapabilities(liveStatus bool, runtimeMode string, normalizeModeFn func(string) string) BackendCapabilities {
	return BackendCapabilities{
		LiveStatusEnabled: liveStatus,
		RuntimeMode:       normalizeModeFn(runtimeMode),
	}
}

// BuildAllPhaseNames constructs phase name/action pairs from phase definitions.
type PhaseNameDef struct {
	Name          string
	CurrentAction string
}

// BuildAllPhaseProgress constructs initial phase progress entries from name definitions.
func BuildAllPhaseProgress(phaseDefs []PhaseNameDef) []PhaseProgressUpdate {
	all := make([]PhaseProgressUpdate, len(phaseDefs))
	for i, p := range phaseDefs {
		all[i] = PhaseProgressUpdate{Name: p.Name, CurrentAction: "pending"}
	}
	return all
}
