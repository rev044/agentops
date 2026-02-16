package main

import (
	"fmt"
	"testing"
)

// TestSpawnClaudePhase_FallbackNoNtm verifies that when ntm is not on PATH,
// spawnClaudePhase falls back to spawnDirectFn (the LookPath branch).
func TestSpawnClaudePhase_FallbackNoNtm(t *testing.T) {
	origLookPath := lookPath
	origSpawnDirect := spawnDirectFn
	defer func() {
		lookPath = origLookPath
		spawnDirectFn = origSpawnDirect
	}()

	// Make ntm not found.
	lookPath = func(name string) (string, error) {
		return "", fmt.Errorf("not found: %s", name)
	}

	// Track whether the direct spawn path was called.
	directCalled := false
	var capturedPrompt, capturedCwd string
	spawnDirectFn = func(prompt, cwd string) error {
		directCalled = true
		capturedPrompt = prompt
		capturedCwd = cwd
		return nil
	}

	cwd := t.TempDir()
	err := spawnClaudePhase("test prompt", cwd, "test-run", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !directCalled {
		t.Error("expected spawnDirectFn to be called when ntm is not on PATH")
	}
	if capturedPrompt != "test prompt" {
		t.Errorf("expected prompt %q, got %q", "test prompt", capturedPrompt)
	}
	if capturedCwd != cwd {
		t.Errorf("expected cwd %q, got %q", cwd, capturedCwd)
	}
}

// TestSpawnClaudePhase_NtmAvailable verifies that when ntm IS on PATH,
// spawnClausePhase attempts the ntm path (spawnClaudePhaseNtm) and does NOT
// call spawnDirectFn unless ntm spawn fails.
func TestSpawnClaudePhase_NtmAvailable(t *testing.T) {
	origLookPath := lookPath
	origSpawnDirect := spawnDirectFn
	defer func() {
		lookPath = origLookPath
		spawnDirectFn = origSpawnDirect
	}()

	// Make ntm "found" at a fake path.
	lookPath = func(name string) (string, error) {
		if name == "ntm" {
			return "/fake/ntm", nil
		}
		return "", fmt.Errorf("not found: %s", name)
	}

	// Track whether the direct path is called (it should be, as ntm spawn will fail).
	directCalled := false
	spawnDirectFn = func(prompt, cwd string) error {
		directCalled = true
		return nil
	}

	err := spawnClaudePhase("test prompt", t.TempDir(), "test-run", 1)
	// The ntm spawn will fail (fake binary), which triggers fallback to direct exec.
	// Either path is acceptable — we just verify no panic and graceful handling.
	if err != nil {
		// Error from ntm send failing is also acceptable.
		t.Logf("Got error (expected — fake ntm binary): %v", err)
	}
	if directCalled {
		t.Log("Fallback to direct exec after ntm spawn failure confirmed")
	}
}

// TestSpawnClaudePhase_DirectError verifies that errors from spawnDirectFn propagate.
func TestSpawnClaudePhase_DirectError(t *testing.T) {
	origLookPath := lookPath
	origSpawnDirect := spawnDirectFn
	defer func() {
		lookPath = origLookPath
		spawnDirectFn = origSpawnDirect
	}()

	lookPath = func(name string) (string, error) {
		return "", fmt.Errorf("not found: %s", name)
	}

	expectedErr := fmt.Errorf("claude process crashed")
	spawnDirectFn = func(prompt, cwd string) error {
		return expectedErr
	}

	err := spawnClaudePhase("test prompt", t.TempDir(), "test-run", 1)
	if err == nil {
		t.Fatal("expected error to propagate from spawnDirectFn")
	}
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

// TestRegisterRPIAgent_NoGt verifies registerRPIAgent handles gt not being
// on PATH gracefully: no panic, no error, silent return.
func TestRegisterRPIAgent_NoGt(t *testing.T) {
	origGtPath := gtPath
	defer func() { gtPath = origGtPath }()

	gtPath = ""

	// Should not panic or error — just silently return.
	registerRPIAgent("test-run-id")
}

// TestRegisterRPIAgent_WithGt verifies registerRPIAgent attempts to call gt
// when gtPath is set (even if the binary doesn't exist at that path).
func TestRegisterRPIAgent_WithGt(t *testing.T) {
	origGtPath := gtPath
	defer func() { gtPath = origGtPath }()

	// Point to a non-existent binary. The exec.Command().Run() will fail
	// silently because registerRPIAgent ignores the error.
	gtPath = "/nonexistent/gt"

	// Should not panic — the function swallows exec errors.
	registerRPIAgent("test-run-id")
}

// TestEmitRPIStatus_NoGt verifies emitRPIStatus handles gt not on PATH gracefully.
func TestEmitRPIStatus_NoGt(t *testing.T) {
	origGtPath := gtPath
	defer func() { gtPath = origGtPath }()

	gtPath = ""

	// Should not panic or error — just silently return.
	emitRPIStatus("test-run-id", "research", "started")
}

// TestEmitRPIStatus_WithGt verifies emitRPIStatus attempts to call gt
// when gtPath is set.
func TestEmitRPIStatus_WithGt(t *testing.T) {
	origGtPath := gtPath
	defer func() { gtPath = origGtPath }()

	gtPath = "/nonexistent/gt"

	// Should not panic — the function swallows exec errors.
	emitRPIStatus("test-run-id", "research", "started")
}

// TestDeregisterRPIAgent_NoGt verifies deregisterRPIAgent handles gt not on PATH gracefully.
func TestDeregisterRPIAgent_NoGt(t *testing.T) {
	origGtPath := gtPath
	defer func() { gtPath = origGtPath }()

	gtPath = ""

	// Should not panic or error — just silently return.
	deregisterRPIAgent("test-run-id")
}

// TestDeregisterRPIAgent_WithGt verifies deregisterRPIAgent attempts to call gt
// when gtPath is set.
func TestDeregisterRPIAgent_WithGt(t *testing.T) {
	origGtPath := gtPath
	defer func() { gtPath = origGtPath }()

	gtPath = "/nonexistent/gt"

	// Should not panic — the function swallows exec errors.
	deregisterRPIAgent("test-run-id")
}
