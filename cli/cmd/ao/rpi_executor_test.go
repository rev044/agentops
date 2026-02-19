package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExecutorInterface verifies all three executor types satisfy PhaseExecutor.
func TestExecutorInterface(t *testing.T) {
	var _ PhaseExecutor = &directExecutor{}
	var _ PhaseExecutor = &ntmExecutor{ntmPath: "/fake/ntm"}
	var _ PhaseExecutor = &streamExecutor{statusPath: "/tmp/status.md", allPhases: nil}

	// Verify Name() returns expected identifiers.
	tests := []struct {
		exec PhaseExecutor
		want string
	}{
		{&directExecutor{}, "direct"},
		{&ntmExecutor{ntmPath: "/fake/ntm"}, "ntm"},
		{&streamExecutor{}, "stream"},
	}
	for _, tt := range tests {
		if got := tt.exec.Name(); got != tt.want {
			t.Errorf("Name() = %q, want %q", got, tt.want)
		}
	}
}

// TestBackendSelectionDirect verifies that direct is selected when ntm is absent
// and live-status is disabled.
func TestBackendSelectionDirect(t *testing.T) {
	caps := backendCapabilities{
		LiveStatusEnabled: false,
		InAgentSession:    false,
		NtmPath:           "", // ntm not available
	}
	exec, reason := selectExecutorFromCaps(caps, "", nil)
	if exec.Name() != "direct" {
		t.Errorf("expected direct executor, got %q", exec.Name())
	}
	if !strings.Contains(reason, "ntm not found") {
		t.Errorf("reason should explain ntm absence, got %q", reason)
	}
}

// TestBackendSelectionNtm verifies that ntm is selected when it is on PATH
// and the session is not an agent session.
func TestBackendSelectionNtm(t *testing.T) {
	caps := backendCapabilities{
		LiveStatusEnabled: false,
		InAgentSession:    false,
		NtmPath:           "/usr/local/bin/ntm",
	}
	exec, reason := selectExecutorFromCaps(caps, "", nil)
	if exec.Name() != "ntm" {
		t.Errorf("expected ntm executor, got %q", exec.Name())
	}
	if !strings.Contains(reason, "/usr/local/bin/ntm") {
		t.Errorf("reason should include ntm path, got %q", reason)
	}
}

// TestBackendSelectionStream verifies that stream is selected when live-status is enabled,
// regardless of ntm availability.
func TestBackendSelectionStream(t *testing.T) {
	caps := backendCapabilities{
		LiveStatusEnabled: true,
		InAgentSession:    false,
		NtmPath:           "/usr/local/bin/ntm", // ntm available but should NOT win
	}
	exec, reason := selectExecutorFromCaps(caps, "/tmp/status.md", nil)
	if exec.Name() != "stream" {
		t.Errorf("expected stream executor, got %q", exec.Name())
	}
	if !strings.Contains(reason, "live-status") {
		t.Errorf("reason should mention live-status, got %q", reason)
	}
}

// TestBackendSelectionAgentSession verifies that agent sessions always use direct,
// even when ntm is on PATH (ntm requires interactive tmux which is unavailable in agents).
func TestBackendSelectionAgentSession(t *testing.T) {
	caps := backendCapabilities{
		LiveStatusEnabled: false,
		InAgentSession:    true,
		NtmPath:           "/usr/local/bin/ntm", // ntm found but suppressed
	}
	exec, reason := selectExecutorFromCaps(caps, "", nil)
	if exec.Name() != "direct" {
		t.Errorf("expected direct executor in agent session, got %q", exec.Name())
	}
	if !strings.Contains(reason, "agent session") {
		t.Errorf("reason should explain agent session suppression, got %q", reason)
	}
}

// TestFallbackDirectWhenNtmAbsent verifies the fallback to direct is deterministic
// and that the reason string makes the fallback explicit.
func TestFallbackDirectWhenNtmAbsent(t *testing.T) {
	caps := backendCapabilities{
		LiveStatusEnabled: false,
		InAgentSession:    false,
		NtmPath:           "",
	}
	exec, reason := selectExecutorFromCaps(caps, "", nil)
	if exec.Name() != "direct" {
		t.Fatalf("fallback should always be direct, got %q", exec.Name())
	}
	if reason == "" {
		t.Error("fallback reason must be non-empty for log traceability")
	}
}

// TestFallbackAgentSessionOverridesNtm ensures that being in an agent session
// causes fallback to direct even when ntm is available — no silent override.
func TestFallbackAgentSessionOverridesNtm(t *testing.T) {
	caps := backendCapabilities{
		LiveStatusEnabled: false,
		InAgentSession:    true,
		NtmPath:           "/opt/homebrew/bin/ntm",
	}
	exec, reason := selectExecutorFromCaps(caps, "", nil)
	if exec.Name() != "direct" {
		t.Fatalf("agent session should always fall back to direct, got %q", exec.Name())
	}
	if !strings.Contains(strings.ToLower(reason), "ntm suppressed") {
		t.Errorf("reason should state ntm was suppressed, got %q", reason)
	}
}

// TestProbeBackendCapabilities_LiveStatus verifies probeBackendCapabilities
// correctly propagates the liveStatus argument.
func TestProbeBackendCapabilities_LiveStatus(t *testing.T) {
	caps := probeBackendCapabilities(true)
	if !caps.LiveStatusEnabled {
		t.Error("LiveStatusEnabled should be true when liveStatus=true")
	}

	caps = probeBackendCapabilities(false)
	if caps.LiveStatusEnabled {
		t.Error("LiveStatusEnabled should be false when liveStatus=false")
	}
}

// TestProbeBackendCapabilities_AgentSession verifies that agent session detection
// works correctly for both CLAUDECODE and CLAUDE_CODE_ENTRYPOINT.
func TestProbeBackendCapabilities_AgentSession(t *testing.T) {
	// Save and restore env vars.
	origCC := os.Getenv("CLAUDECODE")
	origCCE := os.Getenv("CLAUDE_CODE_ENTRYPOINT")
	defer func() {
		os.Setenv("CLAUDECODE", origCC)
		os.Setenv("CLAUDE_CODE_ENTRYPOINT", origCCE)
	}()

	// No agent session env vars.
	os.Unsetenv("CLAUDECODE")
	os.Unsetenv("CLAUDE_CODE_ENTRYPOINT")
	caps := probeBackendCapabilities(false)
	if caps.InAgentSession {
		t.Error("InAgentSession should be false when neither env var is set")
	}

	// CLAUDECODE set.
	os.Setenv("CLAUDECODE", "1")
	caps = probeBackendCapabilities(false)
	if !caps.InAgentSession {
		t.Error("InAgentSession should be true when CLAUDECODE is set")
	}
	// NtmPath should NOT be probed when in agent session.
	// We can only check caps.NtmPath — it will be "" since ntm probe is skipped.
	// (We cannot inject lookPath here without mutating global state, so just check InAgentSession.)
	os.Unsetenv("CLAUDECODE")

	// CLAUDE_CODE_ENTRYPOINT set.
	os.Setenv("CLAUDE_CODE_ENTRYPOINT", "claude-code")
	caps = probeBackendCapabilities(false)
	if !caps.InAgentSession {
		t.Error("InAgentSession should be true when CLAUDE_CODE_ENTRYPOINT is set")
	}
}

// TestProbeBackendCapabilities_NtmProbe verifies that ntm path probing uses lookPath.
func TestProbeBackendCapabilities_NtmProbe(t *testing.T) {
	// Save and restore env vars and lookPath.
	origCC := os.Getenv("CLAUDECODE")
	origCCE := os.Getenv("CLAUDE_CODE_ENTRYPOINT")
	origLookPath := lookPath
	defer func() {
		os.Setenv("CLAUDECODE", origCC)
		os.Setenv("CLAUDE_CODE_ENTRYPOINT", origCCE)
		lookPath = origLookPath
	}()

	os.Unsetenv("CLAUDECODE")
	os.Unsetenv("CLAUDE_CODE_ENTRYPOINT")

	// Simulate ntm present.
	lookPath = func(name string) (string, error) {
		if name == "ntm" {
			return "/usr/local/bin/ntm", nil
		}
		return "", fmt.Errorf("not found: %s", name)
	}
	caps := probeBackendCapabilities(false)
	if caps.NtmPath != "/usr/local/bin/ntm" {
		t.Errorf("NtmPath = %q, want /usr/local/bin/ntm", caps.NtmPath)
	}

	// Simulate ntm absent.
	lookPath = func(name string) (string, error) {
		return "", fmt.Errorf("not found: %s", name)
	}
	caps = probeBackendCapabilities(false)
	if caps.NtmPath != "" {
		t.Errorf("NtmPath should be empty when ntm not found, got %q", caps.NtmPath)
	}
}

// TestSelectExecutorWithLog_LogsSelection verifies that selectExecutorWithLog
// appends a backend-selection entry to the orchestration log.
func TestSelectExecutorWithLog_LogsSelection(t *testing.T) {
	origLookPath := lookPath
	origLiveStatus := phasedLiveStatus
	defer func() {
		lookPath = origLookPath
		phasedLiveStatus = origLiveStatus
	}()

	// Ensure reproducible selection: no ntm, no live-status, no agent session.
	os.Unsetenv("CLAUDECODE")
	os.Unsetenv("CLAUDE_CODE_ENTRYPOINT")
	phasedLiveStatus = false
	lookPath = func(name string) (string, error) {
		return "", fmt.Errorf("not found: %s", name)
	}

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, ".agents", "rpi", "phased-orchestration.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		t.Fatal(err)
	}

	exec := selectExecutorWithLog("", nil, logPath, "test-run-id")
	if exec.Name() != "direct" {
		t.Errorf("expected direct, got %q", exec.Name())
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log file not written: %v", err)
	}
	logContent := string(data)
	if !strings.Contains(logContent, "backend-selection") {
		t.Errorf("log should contain backend-selection entry, got: %q", logContent)
	}
	if !strings.Contains(logContent, "direct") {
		t.Errorf("log should record selected backend (direct), got: %q", logContent)
	}
}

// TestSelectExecutorWithLog_NoLogPath verifies that selectExecutorWithLog works
// correctly when logPath is empty (no log writing, no panic).
func TestSelectExecutorWithLog_NoLogPath(t *testing.T) {
	origLookPath := lookPath
	origLiveStatus := phasedLiveStatus
	defer func() {
		lookPath = origLookPath
		phasedLiveStatus = origLiveStatus
	}()

	phasedLiveStatus = false
	lookPath = func(name string) (string, error) {
		return "", fmt.Errorf("not found: %s", name)
	}

	// Should not panic with empty logPath.
	exec := selectExecutorWithLog("", nil, "", "")
	if exec == nil {
		t.Fatal("executor should not be nil")
	}
}

// TestBackendSelectionPrecedence verifies the complete selection priority order:
// stream > ntm > direct.
func TestBackendSelectionPrecedence(t *testing.T) {
	tests := []struct {
		name        string
		caps        backendCapabilities
		wantBackend string
	}{
		{
			name: "stream wins over ntm",
			caps: backendCapabilities{
				LiveStatusEnabled: true,
				InAgentSession:    false,
				NtmPath:           "/usr/local/bin/ntm",
			},
			wantBackend: "stream",
		},
		{
			name: "stream wins over agent-session direct",
			caps: backendCapabilities{
				LiveStatusEnabled: true,
				InAgentSession:    true,
				NtmPath:           "",
			},
			wantBackend: "stream",
		},
		{
			name: "ntm wins over direct (no agent session)",
			caps: backendCapabilities{
				LiveStatusEnabled: false,
				InAgentSession:    false,
				NtmPath:           "/usr/local/bin/ntm",
			},
			wantBackend: "ntm",
		},
		{
			name: "direct when agent session and ntm present",
			caps: backendCapabilities{
				LiveStatusEnabled: false,
				InAgentSession:    true,
				NtmPath:           "/usr/local/bin/ntm",
			},
			wantBackend: "direct",
		},
		{
			name: "direct when nothing available",
			caps: backendCapabilities{
				LiveStatusEnabled: false,
				InAgentSession:    false,
				NtmPath:           "",
			},
			wantBackend: "direct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec, reason := selectExecutorFromCaps(tt.caps, "/tmp/status.md", nil)
			if exec.Name() != tt.wantBackend {
				t.Errorf("backend = %q, want %q (reason: %q)", exec.Name(), tt.wantBackend, reason)
			}
			if reason == "" {
				t.Error("reason must always be non-empty for log traceability")
			}
		})
	}
}
