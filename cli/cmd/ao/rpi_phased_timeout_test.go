package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFakeClaude(t *testing.T, script string) string {
	t.Helper()

	binDir := t.TempDir()
	path := filepath.Join(binDir, "claude")
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("write fake claude: %v", err)
	}
	return binDir
}

func TestSpawnClaudeDirectImpl_TimesOut(t *testing.T) {
	origTimeout := phasedPhaseTimeout
	defer func() { phasedPhaseTimeout = origTimeout }()

	phasedPhaseTimeout = 150 * time.Millisecond

	binDir := writeFakeClaude(t, "#!/bin/sh\nsleep 5\n")
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	err := spawnClaudeDirectImpl("test prompt", t.TempDir(), 2)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out after") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func TestSpawnClaudePhaseWithStream_TimesOut(t *testing.T) {
	origTimeout := phasedPhaseTimeout
	defer func() { phasedPhaseTimeout = origTimeout }()

	phasedPhaseTimeout = 200 * time.Millisecond

	binDir := writeFakeClaude(t, "#!/bin/sh\necho '{\"type\":\"init\",\"session_id\":\"s1\",\"model\":\"m\"}'\nsleep 5\n")
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	tmpDir := t.TempDir()
	statusPath := filepath.Join(tmpDir, "live-status.md")
	allPhases := []PhaseProgress{{Name: "discovery", CurrentAction: "starting"}}

	err := spawnClaudePhaseWithStream("test prompt", tmpDir, "run-1", 1, statusPath, allPhases)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out after") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

// TestSpawnClaudePhaseWithStream_StallDetected verifies that the stream executor
// fires the stall watchdog when no stream events are received within stallTimeout.
func TestSpawnClaudePhaseWithStream_StallDetected(t *testing.T) {
	origStall := phasedStallTimeout
	origCheck := stallCheckInterval
	origTimeout := phasedPhaseTimeout
	defer func() {
		phasedStallTimeout = origStall
		stallCheckInterval = origCheck
		phasedPhaseTimeout = origTimeout
	}()

	// Use a very short stall timeout with a matching check interval so the
	// watchdog fires almost immediately after one tick.
	phasedStallTimeout = 100 * time.Millisecond
	stallCheckInterval = 50 * time.Millisecond
	phasedPhaseTimeout = 0 // disable hard phase timeout so stall fires first

	// Fake claude: emit one init event then hang — no further activity.
	binDir := writeFakeClaude(t, "#!/bin/sh\necho '{\"type\":\"init\",\"session_id\":\"s1\",\"model\":\"m\"}'\nsleep 10\n")
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	tmpDir := t.TempDir()
	statusPath := filepath.Join(tmpDir, "live-status.md")
	allPhases := []PhaseProgress{{Name: "discovery", CurrentAction: "starting"}}

	err := spawnClaudePhaseWithStream("test prompt", tmpDir, "run-stall", 1, statusPath, allPhases)
	if err == nil {
		t.Fatal("expected stall error")
	}
	if !strings.Contains(err.Error(), string(failReasonStall)) {
		t.Fatalf("expected stall failure reason in error, got: %v", err)
	}
}

// TestSpawnClaudePhaseNtm_TimesOut verifies the ntm executor returns a timeout
// error when phasedPhaseTimeout is exceeded.  We use a fake ntm that succeeds
// spawn/send so the polling loop is entered, then a very short phase timeout
// and a shortened poll interval so the timer fires quickly.
func TestSpawnClaudePhaseNtm_TimesOut(t *testing.T) {
	origTimeout := phasedPhaseTimeout
	origPoll := ntmPollInterval
	origDirect := spawnDirectFn
	defer func() {
		phasedPhaseTimeout = origTimeout
		ntmPollInterval = origPoll
		spawnDirectFn = origDirect
	}()

	phasedPhaseTimeout = 150 * time.Millisecond
	ntmPollInterval = 50 * time.Millisecond

	// spawnDirectFn must not be called — if the ntm path falls back we'd miss the test.
	spawnDirectFn = func(prompt, cwd string, phaseNum int) error {
		t.Error("unexpected fallback to spawnDirectFn")
		return nil
	}

	// Write fake ntm binary: spawn and send succeed; kill is also accepted.
	// The session is never removed from tmux so the polling loop keeps running
	// until the phase timeout fires.
	tmpBin := t.TempDir()
	fakentm := filepath.Join(tmpBin, "ntm")
	script := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(fakentm, []byte(script), 0755); err != nil {
		t.Fatalf("write fake ntm: %v", err)
	}

	// Write fake tmux binary: has-session always succeeds (session "exists"),
	// capture-pane emits static content (triggering stall counter, but we rely
	// on the phase timeout here, so stall detection is disabled).
	faketmux := filepath.Join(tmpBin, "tmux")
	tmuxScript := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(faketmux, []byte(tmuxScript), 0755); err != nil {
		t.Fatalf("write fake tmux: %v", err)
	}

	phasedStallTimeout = 0 // disable stall so only phase timeout fires
	t.Setenv("PATH", tmpBin+":"+os.Getenv("PATH"))

	err := spawnClaudePhaseNtm(fakentm, "test prompt", t.TempDir(), "run-ntm-timeout", 2)
	if err == nil {
		t.Fatal("expected timeout error from ntm executor")
	}
	if !strings.Contains(err.Error(), "timed out after") {
		t.Fatalf("expected timeout in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), string(failReasonTimeout)) {
		t.Fatalf("expected failReasonTimeout in error, got: %v", err)
	}
}

// TestSpawnClaudePhaseNtm_StallDetected verifies the ntm executor returns a stall
// error when pane content is static for longer than phasedStallTimeout.
func TestSpawnClaudePhaseNtm_StallDetected(t *testing.T) {
	origStall := phasedStallTimeout
	origPoll := ntmPollInterval
	origTimeout := phasedPhaseTimeout
	origDirect := spawnDirectFn
	defer func() {
		phasedStallTimeout = origStall
		ntmPollInterval = origPoll
		phasedPhaseTimeout = origTimeout
		spawnDirectFn = origDirect
	}()

	phasedStallTimeout = 80 * time.Millisecond
	ntmPollInterval = 40 * time.Millisecond
	phasedPhaseTimeout = 0 // disable hard timeout so stall fires first

	spawnDirectFn = func(prompt, cwd string, phaseNum int) error {
		t.Error("unexpected fallback to spawnDirectFn")
		return nil
	}

	tmpBin := t.TempDir()

	fakentm := filepath.Join(tmpBin, "ntm")
	if err := os.WriteFile(fakentm, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("write fake ntm: %v", err)
	}

	// fake tmux: has-session succeeds (keeps loop alive), capture-pane returns
	// a fixed string so content never changes — this triggers stall detection.
	faketmux := filepath.Join(tmpBin, "tmux")
	tmuxScript := `#!/bin/sh
case "$1" in
  has-session) exit 0 ;;
  capture-pane) echo "static pane content" ;;
  *) exit 0 ;;
esac
`
	if err := os.WriteFile(faketmux, []byte(tmuxScript), 0755); err != nil {
		t.Fatalf("write fake tmux: %v", err)
	}

	t.Setenv("PATH", tmpBin+":"+os.Getenv("PATH"))

	err := spawnClaudePhaseNtm(fakentm, "test prompt", t.TempDir(), "run-ntm-stall", 3)
	if err == nil {
		t.Fatal("expected stall error from ntm executor")
	}
	if !strings.Contains(err.Error(), string(failReasonStall)) {
		t.Fatalf("expected stall failure reason in error, got: %v", err)
	}
}

// TestStallTimeoutClassification verifies that failure reasons are distinct
// string constants and that each error path embeds the correct reason.
func TestStallTimeoutClassification(t *testing.T) {
	if failReasonTimeout == failReasonStall {
		t.Error("failReasonTimeout and failReasonStall must be distinct")
	}
	if failReasonTimeout == failReasonExit {
		t.Error("failReasonTimeout and failReasonExit must be distinct")
	}
	if failReasonStall == failReasonExit {
		t.Error("failReasonStall and failReasonExit must be distinct")
	}

	// Verify the error strings embed the expected reason tokens.
	timeoutMsg := string(failReasonTimeout)
	stallMsg := string(failReasonStall)

	if !strings.Contains("phase 1 (timeout) timed out after 30m0s", timeoutMsg) {
		t.Errorf("expected %q in timeout error format", timeoutMsg)
	}
	if !strings.Contains("phase 1 (stall): stall detected: no pane activity for 5m0s", stallMsg) {
		t.Errorf("expected %q in stall error format", stallMsg)
	}
}
