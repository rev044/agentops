package context

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExtra_BudgetTracker_Save_MkdirError covers MkdirAll failure in Save.
func TestExtra_BudgetTracker_Save_MkdirError(t *testing.T) {
	bt := &BudgetTracker{SessionID: "test-session"}
	// Use /dev/null as base to force MkdirAll failure.
	err := bt.Save("/dev/null")
	if err == nil {
		t.Fatal("expected MkdirAll error, got nil")
	}
}

// TestExtra_BudgetTracker_Save_Success covers the happy path.
func TestExtra_BudgetTracker_Save_Success(t *testing.T) {
	bt := &BudgetTracker{SessionID: "test-save"}
	tmp := t.TempDir()
	if err := bt.Save(tmp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(tmp, ".agents", "ao", "context", "budget-test-save.json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected budget file to exist: %v", err)
	}
}

// TestExtra_Summarizer_SaveState_MkdirError covers MkdirAll failure in SaveState.
func TestExtra_Summarizer_SaveState_MkdirError(t *testing.T) {
	s := &Summarizer{}
	state := SummarizeState{SessionID: "test-session"}
	err := s.SaveState("/dev/null", state)
	if err == nil {
		t.Fatal("expected MkdirAll error, got nil")
	}
}

// TestExtra_Summarizer_SaveState_Success covers the happy path.
func TestExtra_Summarizer_SaveState_Success(t *testing.T) {
	s := &Summarizer{}
	state := SummarizeState{
		SessionID:      "test-state",
		CompletedTasks: []string{"task1"},
		Notes:          "some notes",
	}
	tmp := t.TempDir()
	if err := s.SaveState(tmp, state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(tmp, ".agents", "ao", "context", "state-test-state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected state file to exist: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty state file")
	}
}
