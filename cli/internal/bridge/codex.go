package bridge

import (
	"path/filepath"
	"strings"
)

// NormalizeCodexLifecyclePath cleans and absolutizes a lifecycle file path.
func NormalizeCodexLifecyclePath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if abs, err := filepath.Abs(trimmed); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(trimmed)
}

// FirstNonEmptyTrimmed returns the first non-empty trimmed value from the variadic list.
func FirstNonEmptyTrimmed(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// CodexStopAlreadyClosed determines whether a codex stop has already been recorded
// for the given session/transcript combination. It compares the last-stop state
// against the current session ID and transcript path.
func CodexStopAlreadyClosed(lastStopSessionID, lastStopTranscriptPath, sessionID, transcriptPath string) bool {
	lastSessionID := strings.TrimSpace(lastStopSessionID)
	lastTranscript := NormalizeCodexLifecyclePath(lastStopTranscriptPath)
	currentTranscript := NormalizeCodexLifecyclePath(transcriptPath)

	if currentTranscript != "" && lastTranscript != "" {
		if currentTranscript != lastTranscript {
			return false
		}
		if sessionID != "" && lastSessionID != "" && strings.TrimSpace(sessionID) != lastSessionID {
			return false
		}
		return true
	}

	if sessionID == "" || lastSessionID == "" {
		return false
	}
	return strings.TrimSpace(sessionID) == lastSessionID
}

// EnsureStopReason returns a human-readable reason for an ensure-stop result.
func EnsureStopReason(status string) string {
	if status == "already_closed" {
		return "closeout already recorded for this Codex thread"
	}
	return "closeout recorded for current Codex thread"
}
