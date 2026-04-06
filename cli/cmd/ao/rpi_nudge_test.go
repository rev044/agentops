package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolveNudgeTargets_AllWorkers(t *testing.T) {
	base := "ao-rpi-run12345-p1"
	sessions := []string{base, base + "-w2", base + "-w1", "other"}

	got, err := resolveNudgeTargets(sessions, base, true, 0)
	if err != nil {
		t.Fatalf("resolveNudgeTargets(all workers): %v", err)
	}
	if len(got) != 2 || got[0] != base+"-w1" || got[1] != base+"-w2" {
		t.Fatalf("unexpected targets: %#v", got)
	}
}

func TestResolveNudgeTargets_DefaultMayor(t *testing.T) {
	base := "ao-rpi-run12345-p2"
	sessions := []string{base}

	got, err := resolveNudgeTargets(sessions, base, false, 0)
	if err != nil {
		t.Fatalf("resolveNudgeTargets(default mayor): %v", err)
	}
	if len(got) != 1 || got[0] != base {
		t.Fatalf("unexpected targets: %#v", got)
	}
}

func TestResolveNudgeTargets_DefaultAmbiguousWorkers(t *testing.T) {
	base := "ao-rpi-run12345-p3"
	sessions := []string{base + "-w1", base + "-w2"}

	_, err := resolveNudgeTargets(sessions, base, false, 0)
	if err == nil || !strings.Contains(err.Error(), "multiple worker sessions") {
		t.Fatalf("expected ambiguous worker error, got: %v", err)
	}
}

func TestResolveNudgeTargets_OneWorker(t *testing.T) {
	base := "ao-rpi-run12345-p1"
	sessions := []string{base, base + "-w1", base + "-w2"}

	got, err := resolveNudgeTargets(sessions, base, false, 2)
	if err != nil {
		t.Fatalf("resolveNudgeTargets(worker): %v", err)
	}
	if len(got) != 1 || got[0] != base+"-w2" {
		t.Fatalf("unexpected targets: %#v", got)
	}
}

func TestResolveNudgePhase(t *testing.T) {
	phase, err := resolveNudgePhase(&phasedState{Phase: 2}, 0)
	if err != nil {
		t.Fatalf("resolveNudgePhase(state): %v", err)
	}
	if phase != 2 {
		t.Fatalf("phase = %d, want 2", phase)
	}

	phase, err = resolveNudgePhase(&phasedState{Phase: 2}, 3)
	if err != nil {
		t.Fatalf("resolveNudgePhase(flag): %v", err)
	}
	if phase != 3 {
		t.Fatalf("phase = %d, want 3", phase)
	}

	if _, err := resolveNudgePhase(&phasedState{Phase: 2}, 9); err == nil {
		t.Fatal("expected error for invalid phase")
	}
}

func TestAppendRPINudgeAudit(t *testing.T) {
	root := t.TempDir()
	runID := "audit-run-1"
	record := rpiNudgeRecord{
		Timestamp: "2026-02-26T00:00:00Z",
		RunID:     runID,
		Phase:     2,
		Targets:   []string{"ao-rpi-auditrun-p2-w1", "ao-rpi-auditrun-p2-w2"},
		Message:   "change direction",
	}

	if err := appendRPINudgeAudit(root, runID, record); err != nil {
		t.Fatalf("appendRPINudgeAudit: %v", err)
	}

	path := filepath.Join(rpiRunRegistryDir(root, runID), "nudges.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read nudges.jsonl: %v", err)
	}

	var got rpiNudgeRecord
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal audit record: %v", err)
	}
	if got.RunID != runID || got.Phase != 2 || got.Message != "change direction" {
		t.Fatalf("unexpected record: %#v", got)
	}
	if len(got.Targets) != 2 {
		t.Fatalf("unexpected targets: %#v", got.Targets)
	}
}

func TestRPINudgeCommand_E2EAllWorkers(t *testing.T) {
	tmuxBin, err := defaultLookPath(nil)("tmux")
	if err != nil {
		t.Skipf("tmux not available: %v", err)
	}

	tmp := t.TempDir()
	runID := "nudge-e2e-01"
	phase := 1
	baseSession := tmuxSessionName(runID, phase)
	worker1 := baseSession + "-w1"
	worker2 := baseSession + "-w2"
	logPath := filepath.Join(tmp, "nudge-e2e.log")

	t.Cleanup(func() {
		_ = exec.Command(tmuxBin, "kill-session", "-t", worker1).Run()
		_ = exec.Command(tmuxBin, "kill-session", "-t", worker2).Run()
	})

	statePath := filepath.Join(tmp, ".agents", "rpi", "runs", runID, phasedStateFile)
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	state := &phasedState{
		SchemaVersion: 1,
		Goal:          "test nudge command",
		Phase:         phase,
		RunID:         runID,
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal state: %v", err)
	}
	if err := os.WriteFile(statePath, data, 0o644); err != nil {
		t.Fatalf("write state: %v", err)
	}

	workerScriptPath := filepath.Join(tmp, "worker.sh")
	workerScript := `#!/usr/bin/env bash
set -euo pipefail
worker="$1"
log="$2"
echo "ready:${worker}" >> "$log"
while true; do
  if IFS= read -r line; then
    echo "in:${worker}:${line}" >> "$log"
    if [[ "$line" == *"stop-now"* ]]; then
      exit 0
    fi
  else
    sleep 0.05
  fi
done
`
	if err := os.WriteFile(workerScriptPath, []byte(workerScript), 0o700); err != nil {
		t.Fatalf("write worker script: %v", err)
	}

	if err := exec.Command(tmuxBin, "new-session", "-d", "-s", worker1, "-c", tmp, workerScriptPath, "1", logPath).Run(); err != nil {
		t.Fatalf("spawn worker1: %v", err)
	}
	if err := exec.Command(tmuxBin, "new-session", "-d", "-s", worker2, "-c", tmp, workerScriptPath, "2", logPath).Run(); err != nil {
		t.Fatalf("spawn worker2: %v", err)
	}

	readyDeadline := time.Now().Add(5 * time.Second)
	for {
		raw, _ := os.ReadFile(logPath)
		log := string(raw)
		if strings.Contains(log, "ready:1") && strings.Contains(log, "ready:2") {
			break
		}
		if time.Now().After(readyDeadline) {
			t.Fatalf("workers did not become ready:\n%s", log)
		}
		time.Sleep(100 * time.Millisecond)
	}

	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldCwd)
	})

	if _, err := executeCommand(
		"rpi",
		"nudge",
		"--run-id", runID,
		"--phase", "1",
		"--all-workers",
		"--message", "change-direction now",
	); err != nil {
		t.Fatalf("execute ao rpi nudge: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for {
		raw, _ := os.ReadFile(logPath)
		log := string(raw)
		if strings.Contains(log, "in:1:change-direction now") && strings.Contains(log, "in:2:change-direction now") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("did not observe nudges in both workers log:\n%s", log)
		}
		time.Sleep(100 * time.Millisecond)
	}

	commands, err := loadRPIC2Commands(tmp, runID)
	if err != nil {
		t.Fatalf("loadRPIC2Commands: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("len(commands) = %d, want 1", len(commands))
	}
	commandID := commands[0].CommandID
	if commandID == "" {
		t.Fatal("expected non-empty command_id")
	}

	events, err := loadRPIC2Events(tmp, runID)
	if err != nil {
		t.Fatalf("loadRPIC2Events: %v", err)
	}
	ackCount := 0
	for _, ev := range events {
		if ev.CommandID != commandID {
			continue
		}
		if ev.Type == "command.nudge.ack" {
			ackCount++
		}
	}
	if ackCount < 2 {
		t.Fatalf("expected >=2 ack events for command %s, got %d", commandID, ackCount)
	}

	_ = sendTmuxNudge(tmuxBin, worker1, "stop-now")
	_ = sendTmuxNudge(tmuxBin, worker2, "stop-now")
}
