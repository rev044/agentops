package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestTmuxExecutor_Name(t *testing.T) {
	exec := &tmuxExecutor{}
	if got := exec.Name(); got != "tmux" {
		t.Errorf("Name() = %q, want %q", got, "tmux")
	}
}

func TestTmuxSessionName_Format(t *testing.T) {
	got := tmuxSessionName("abcdef1234567890", 1)
	if got != "ao-rpi-abcdef12-p1" {
		t.Errorf("tmuxSessionName = %q", got)
	}
	got2 := tmuxSessionName("abc", 2)
	if got2 != "ao-rpi-abc-p2" {
		t.Errorf("tmuxSessionName(short) = %q", got2)
	}
}

func TestTmuxSessionName_ShellSafe(t *testing.T) {
	safePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	for _, in := range []string{"run.id", "run id", "run@#$", "path/to/run"} {
		got := tmuxSessionName(in, 1)
		if !safePattern.MatchString(got) {
			t.Fatalf("unsafe session name: %q", got)
		}
	}
}

func TestTmuxRuntimeInvocationTemplate(t *testing.T) {
	exe, args, err := tmuxRuntimeInvocationTemplate("claude")
	if err != nil {
		t.Fatalf("tmuxRuntimeInvocationTemplate(claude): %v", err)
	}
	if exe != "claude" {
		t.Fatalf("exe = %q", exe)
	}
	if len(args) != 5 || args[0] != "-p" || args[1] != "--output-format" || args[2] != "stream-json" || args[3] != "--include-partial-messages" || args[4] != "--verbose" {
		t.Fatalf("args = %#v", args)
	}

	exe2, args2, err := tmuxRuntimeInvocationTemplate("codex")
	if err != nil {
		t.Fatalf("tmuxRuntimeInvocationTemplate(codex): %v", err)
	}
	if exe2 != "codex" {
		t.Fatalf("exe2 = %q", exe2)
	}
	if len(args2) != 2 || args2[0] != "exec" || args2[1] != "--json" {
		t.Fatalf("args2 = %#v", args2)
	}
}

func TestSelectExecutorFromCaps_Tmux(t *testing.T) {
	opts := defaultPhasedEngineOptions()
	opts.RuntimeMode = "tmux"
	opts.TmuxWorkers = 2
	exec, reason := selectExecutorFromCaps(backendCapabilities{RuntimeMode: "tmux"}, "", nil, opts)
	if exec.Name() != "tmux" {
		t.Fatalf("executor = %q", exec.Name())
	}
	if reason != "runtime=tmux" {
		t.Fatalf("reason = %q", reason)
	}
	tmux, ok := exec.(*tmuxExecutor)
	if !ok {
		t.Fatalf("expected *tmuxExecutor, got %T", exec)
	}
	if tmux.workerCount != 2 {
		t.Fatalf("workerCount = %d", tmux.workerCount)
	}
}

func TestValidateRuntimeMode_Tmux(t *testing.T) {
	for _, mode := range []string{"tmux", "TMUX", "auto", "direct", "stream"} {
		if err := validateRuntimeMode(mode); err != nil {
			t.Fatalf("validateRuntimeMode(%q): %v", mode, err)
		}
	}
	if err := validateRuntimeMode("bogus"); err == nil {
		t.Fatal("expected error for bogus mode")
	}
}

func TestTmuxHelpers_FilterWorkers(t *testing.T) {
	base := "ao-rpi-abcd1234-p2"
	sessions := []string{
		base,
		base + "-w1",
		base + "-w2",
		"other",
	}
	got := filterTmuxWorkerSessions(sessions, base)
	if len(got) != 2 {
		t.Fatalf("expected 2 workers, got %d", len(got))
	}
}

func TestTmuxExecutorE2ENudgeTwoWorkers(t *testing.T) {
	tmuxBin, err := lookPath("tmux")
	if err != nil {
		t.Skipf("tmux not available: %v", err)
	}

	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "tmux-workers.log")

	runtimePath := filepath.Join(tmp, "mock-runtime.sh")
	runtimeScript := fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
log=%q
worker="${RPI_TMUX_WORKER_ID:-0}"
echo "start:${worker}" >> "$log"
while IFS= read -r line; do
  echo "in:${worker}:${line}" >> "$log"
  if [[ "$line" == *"change-direction"* ]]; then
    echo "nudged:${worker}" >> "$log"
  fi
  if [[ "$line" == *"complete-now"* ]]; then
    exit 0
  fi
done
`, logPath)
	if err := os.WriteFile(runtimePath, []byte(runtimeScript), 0o700); err != nil {
		t.Fatalf("write runtime script: %v", err)
	}

	execTmux := &tmuxExecutor{
		tmuxCommand:    tmuxBin,
		runtimeCommand: runtimePath,
		phaseTimeout:   20 * time.Second,
		pollInterval:   200 * time.Millisecond,
		workerCount:    2,
	}

	runID := "nudge1234"
	done := make(chan error, 1)
	go func() {
		done <- execTmux.Execute("test prompt", tmp, runID, 1)
	}()

	baseSession := tmuxSessionName(runID, 1)
	workerA := baseSession + "-w1"
	workerB := baseSession + "-w2"

	// Wait for worker sessions to appear.
	deadline := time.Now().Add(8 * time.Second)
	for {
		sessions, err := listTmuxSessions(tmuxBin)
		if err == nil {
			seenA := false
			seenB := false
			for _, s := range sessions {
				if s == workerA {
					seenA = true
				}
				if s == workerB {
					seenB = true
				}
			}
			if seenA && seenB {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("workers not ready before deadline")
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := sendTmuxNudge(tmuxBin, workerA, "change-direction worker-a"); err != nil {
		t.Fatalf("nudge workerA: %v", err)
	}
	if err := sendTmuxNudge(tmuxBin, workerB, "change-direction worker-b"); err != nil {
		t.Fatalf("nudge workerB: %v", err)
	}
	if err := sendTmuxNudge(tmuxBin, workerA, "complete-now"); err != nil {
		t.Fatalf("complete workerA: %v", err)
	}
	if err := sendTmuxNudge(tmuxBin, workerB, "complete-now"); err != nil {
		t.Fatalf("complete workerB: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("tmux executor failed: %v", err)
		}
	case <-time.After(12 * time.Second):
		t.Fatal("timeout waiting for tmux executor completion")
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	log := string(data)
	if !strings.Contains(log, "nudged:1") {
		t.Fatalf("missing nudge confirmation for worker 1:\n%s", log)
	}
	if !strings.Contains(log, "nudged:2") {
		t.Fatalf("missing nudge confirmation for worker 2:\n%s", log)
	}
}

func TestTmuxExecutorImplementsPhaseExecutor(t *testing.T) {
	var _ PhaseExecutor = &tmuxExecutor{}
}

func TestTmuxExecutor_MissingExitFileEmitsFailureEvent(t *testing.T) {
	tmp := t.TempDir()
	tmuxPath := filepath.Join(tmp, "fake-tmux.sh")
	fakeTmux := `#!/usr/bin/env bash
set -euo pipefail
cmd="${1:-}"
case "$cmd" in
  new-session) exit 0 ;;
  has-session) exit 1 ;;
  kill-session) exit 0 ;;
  *) exit 0 ;;
esac
`
	if err := os.WriteFile(tmuxPath, []byte(fakeTmux), 0o700); err != nil {
		t.Fatalf("write fake tmux: %v", err)
	}

	execTmux := &tmuxExecutor{
		tmuxCommand:    tmuxPath,
		runtimeCommand: "echo",
		phaseTimeout:   2 * time.Second,
		pollInterval:   50 * time.Millisecond,
		workerCount:    1,
	}
	runID := "exit-missing-run"
	err := execTmux.Execute("prompt", tmp, runID, 1)
	if err == nil {
		t.Fatal("expected missing exit-code file error")
	}
	if !strings.Contains(err.Error(), "exit-code file") {
		t.Fatalf("expected exit-code file error, got: %v", err)
	}

	events, loadErr := loadRPIC2Events(tmp, runID)
	if loadErr != nil {
		t.Fatalf("loadRPIC2Events: %v", loadErr)
	}
	foundFailure := false
	for _, ev := range events {
		if ev.Type == "phase.tmux.failed" && strings.Contains(ev.Message, "exit-code file") {
			foundFailure = true
			break
		}
	}
	if !foundFailure {
		t.Fatalf("expected phase.tmux.failed event with exit-code file message, events=%#v", events)
	}
}
