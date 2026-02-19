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
