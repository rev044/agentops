package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ===========================================================================
// fire.go — runFireIteration (zero coverage)
// ===========================================================================

func TestCov3_fire_runFireIteration_emptyEpic(t *testing.T) {
	// runFireIteration calls findPhase which calls bdReady, bdListByStatus, etc.
	// With no bd on PATH, findPhase returns an error, and runFireIteration wraps it.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("PATH", tmp) // no bd on PATH

	cfg := FireConfig{
		EpicID:       "test-epic",
		Rig:          "test-rig",
		MaxPolecats:  2,
		PollInterval: time.Second,
		MaxRetries:   3,
		BackoffBase:  time.Second,
	}
	retryQueue := make(map[string]*RetryInfo)

	done, err := runFireIteration(cfg, retryQueue)
	if done {
		t.Error("expected done=false when findPhase fails")
	}
	if err == nil {
		t.Fatal("expected error when bd is not available")
	}
}



// ===========================================================================
// fire.go — ignitePhase (zero coverage)
// ===========================================================================

func TestCov3_fire_ignitePhase_atCapacity(t *testing.T) {
	state := &FireState{
		Burning: []string{"a", "b", "c", "d"},
	}
	cfg := FireConfig{MaxPolecats: 4, Rig: "test"}
	retryQueue := make(map[string]*RetryInfo)

	ignited, err := ignitePhase(state, cfg, retryQueue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ignited) != 0 {
		t.Errorf("expected no ignitions at capacity, got %v", ignited)
	}
}

func TestCov3_fire_ignitePhase_noReadyNoDue(t *testing.T) {
	state := &FireState{
		Burning: []string{"a"},
	}
	cfg := FireConfig{MaxPolecats: 4, Rig: "test"}
	retryQueue := make(map[string]*RetryInfo)

	ignited, err := ignitePhase(state, cfg, retryQueue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ignited) != 0 {
		t.Errorf("expected no ignitions with no ready/due, got %v", ignited)
	}
}

func TestCov3_fire_ignitePhase_withReadyButNoGt(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("PATH", tmp) // no gt on PATH

	state := &FireState{
		Ready:   []string{"issue-1"},
		Burning: []string{},
	}
	cfg := FireConfig{MaxPolecats: 4, Rig: "test"}
	retryQueue := make(map[string]*RetryInfo)

	// ignitePhase calls slingIssues which calls gtSling; without gt, sling fails
	// but ignitePhase returns nil error (failures are logged, not returned)
	ignited, err := ignitePhase(state, cfg, retryQueue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No items should be successfully ignited since gt is missing
	if len(ignited) != 0 {
		t.Errorf("expected 0 ignited without gt, got %d", len(ignited))
	}
}

func TestCov3_fire_ignitePhase_withDueRetries(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("PATH", tmp) // no gt

	state := &FireState{
		Burning: []string{},
	}
	cfg := FireConfig{MaxPolecats: 4, Rig: "test"}
	retryQueue := map[string]*RetryInfo{
		"retry-issue": {
			IssueID:     "retry-issue",
			Attempt:     1,
			NextAttempt: time.Now().Add(-time.Minute), // already due
		},
	}

	ignited, err := ignitePhase(state, cfg, retryQueue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// retryQueue entry should have been consumed even though sling failed
	if _, exists := retryQueue["retry-issue"]; exists {
		t.Error("expected retry-issue to be removed from retry queue")
	}
	_ = ignited
}

// ===========================================================================
// fire.go — reapPhase (zero coverage)
// ===========================================================================

func TestCov3_fire_reapPhase_noBurning(t *testing.T) {
	state := &FireState{Burning: []string{}}
	reaped, failures, err := reapPhase(state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reaped) != 0 {
		t.Errorf("expected no reaped, got %v", reaped)
	}
	if len(failures) != 0 {
		t.Errorf("expected no failures, got %v", failures)
	}
}

func TestCov3_fire_reapPhase_withBurningNoBd(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("PATH", tmp) // no bd

	state := &FireState{Burning: []string{"issue-1", "issue-2"}}
	reaped, failures, err := reapPhase(state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// bdShowStatus fails for all issues -> all skipped (continue in switch)
	if len(reaped) != 0 {
		t.Errorf("expected 0 reaped without bd, got %d", len(reaped))
	}
	if len(failures) != 0 {
		t.Errorf("expected 0 failures without bd, got %d", len(failures))
	}
}









// ===========================================================================
// fire.go — sendMail (zero coverage)
// ===========================================================================

func TestCov3_fire_sendMail(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	err := sendMail("mayor", "test message body", "blocker")
	if err != nil {
		t.Fatalf("sendMail failed: %v", err)
	}

	// Verify the file was created
	messagesPath := filepath.Join(tmp, ".agents", "mail", "messages.jsonl")
	if _, err := os.Stat(messagesPath); os.IsNotExist(err) {
		t.Fatal("messages.jsonl was not created")
	}

	content, err := os.ReadFile(messagesPath)
	if err != nil {
		t.Fatalf("read messages.jsonl: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("messages.jsonl is empty")
	}
}

// ===========================================================================
// fire.go — printState (zero coverage)
// ===========================================================================

func TestCov3_fire_printState_doesNotPanic(t *testing.T) {
	// printState prints to stdout; verify it doesn't panic
	state := &FireState{
		Ready:   []string{"a", "b"},
		Burning: []string{"c"},
		Reaped:  []string{"d", "e", "f"},
		Blocked: []string{"g"},
	}
	// Should not panic
	printState(state)
}

func TestCov3_fire_printState_emptyState(t *testing.T) {
	state := &FireState{}
	printState(state)
}

