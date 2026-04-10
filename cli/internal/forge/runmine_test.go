package forge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunMinePass_MissingSessionsDir(t *testing.T) {
	cwd := t.TempDir()
	// No SessionsDir set — should error out with a descriptive message.
	report, err := RunMinePass(cwd, MineOpts{})
	if err == nil {
		t.Fatalf("expected error for missing SessionsDir, got nil")
	}
	if report == nil {
		t.Fatalf("expected non-nil report even on error, got nil")
	}
	if len(report.Learnings) != 0 {
		t.Errorf("expected empty learnings on error, got %d", len(report.Learnings))
	}
}

func TestRunMinePass_NonExistentSessionsDir(t *testing.T) {
	cwd := t.TempDir()
	opts := MineOpts{
		SessionsDir: filepath.Join(cwd, "does-not-exist"),
	}
	report, err := RunMinePass(cwd, opts)
	if err != nil {
		t.Fatalf("expected nil error for non-existent sessions dir (soft-fail), got %v", err)
	}
	if report == nil {
		t.Fatalf("expected non-nil report, got nil")
	}
	if len(report.Learnings) != 0 {
		t.Errorf("expected 0 learnings, got %d", len(report.Learnings))
	}
	if len(report.Degraded) == 0 {
		t.Errorf("expected a degraded note for missing dir, got none")
	}
}

func TestRunMinePass_EmptyCorpus(t *testing.T) {
	cwd := t.TempDir()
	sessionsDir := filepath.Join(cwd, ".agents", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}

	report, err := RunMinePass(cwd, MineOpts{SessionsDir: sessionsDir})
	if err != nil {
		t.Fatalf("RunMinePass empty corpus: %v", err)
	}
	if report.SessionsRead != 0 {
		t.Errorf("SessionsRead = %d, want 0", report.SessionsRead)
	}
	if len(report.Learnings) != 0 {
		t.Errorf("Learnings len = %d, want 0", len(report.Learnings))
	}
	if len(report.Degraded) != 0 {
		t.Errorf("Degraded len = %d, want 0 (empty dir is not degraded), got %v", len(report.Degraded), report.Degraded)
	}
}

func TestRunMinePass_MinimalSession(t *testing.T) {
	cwd := t.TempDir()
	sessionsDir := filepath.Join(cwd, ".agents", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}

	// Seed a single session file with 2 decisions + 1 knowledge record.
	sess := minedSessionFile{
		ID:      "sess-abc123",
		Date:    time.Now(),
		Summary: "test session summary",
		Decisions: []string{
			"decided to extract RunMinePass",
			"decided to use pure extraction (Option A)",
		},
		Knowledge: []string{
			"learned that forge helpers already live in internal/forge",
		},
	}
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		t.Fatalf("marshal session: %v", err)
	}
	sessionFile := filepath.Join(sessionsDir, "sess-abc123.json")
	if err := os.WriteFile(sessionFile, data, 0o600); err != nil {
		t.Fatalf("write session: %v", err)
	}

	report, err := RunMinePass(cwd, MineOpts{SessionsDir: sessionsDir})
	if err != nil {
		t.Fatalf("RunMinePass: %v", err)
	}

	if report.SessionsRead != 1 {
		t.Errorf("SessionsRead = %d, want 1", report.SessionsRead)
	}
	if len(report.Learnings) != 3 {
		t.Fatalf("Learnings len = %d, want 3", len(report.Learnings))
	}

	// Count kinds explicitly.
	var decisions, knowledge int
	for _, l := range report.Learnings {
		if l.Source != "sess-abc123.json" {
			t.Errorf("Learning.Source = %q, want sess-abc123.json", l.Source)
		}
		if l.Title != "test session summary" {
			t.Errorf("Learning.Title = %q, want 'test session summary'", l.Title)
		}
		switch l.Kind {
		case "decision":
			decisions++
		case "knowledge":
			knowledge++
		default:
			t.Errorf("unexpected kind %q", l.Kind)
		}
		if l.Body == "" {
			t.Errorf("Learning.Body is empty")
		}
		if l.Extracted.IsZero() {
			t.Errorf("Learning.Extracted is zero")
		}
	}
	if decisions != 2 {
		t.Errorf("decisions count = %d, want 2", decisions)
	}
	if knowledge != 1 {
		t.Errorf("knowledge count = %d, want 1", knowledge)
	}
}

func TestRunMinePass_SinceTimeFilter(t *testing.T) {
	cwd := t.TempDir()
	sessionsDir := filepath.Join(cwd, ".agents", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}

	// Create an "old" session and an implicitly newer session.
	oldSess := minedSessionFile{
		ID:        "old",
		Summary:   "old session",
		Decisions: []string{"old decision"},
	}
	oldData, _ := json.Marshal(oldSess)
	oldPath := filepath.Join(sessionsDir, "old.json")
	if err := os.WriteFile(oldPath, oldData, 0o600); err != nil {
		t.Fatalf("write old: %v", err)
	}
	// Set old mtime to 48h ago.
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes old: %v", err)
	}

	newSess := minedSessionFile{
		ID:        "new",
		Summary:   "new session",
		Knowledge: []string{"new knowledge"},
	}
	newData, _ := json.Marshal(newSess)
	newPath := filepath.Join(sessionsDir, "new.json")
	if err := os.WriteFile(newPath, newData, 0o600); err != nil {
		t.Fatalf("write new: %v", err)
	}

	// SinceTime 24h ago — should exclude the old session.
	sinceTime := time.Now().Add(-24 * time.Hour)
	report, err := RunMinePass(cwd, MineOpts{
		SessionsDir: sessionsDir,
		SinceTime:   sinceTime,
	})
	if err != nil {
		t.Fatalf("RunMinePass: %v", err)
	}
	if report.SessionsRead != 1 {
		t.Errorf("SessionsRead = %d, want 1 (only new session should be read)", report.SessionsRead)
	}
	if len(report.Learnings) != 1 {
		t.Fatalf("Learnings len = %d, want 1", len(report.Learnings))
	}
	if report.Learnings[0].Source != "new.json" {
		t.Errorf("Learnings[0].Source = %q, want new.json", report.Learnings[0].Source)
	}
}

func TestRunMinePass_MaxSessionsCap(t *testing.T) {
	cwd := t.TempDir()
	sessionsDir := filepath.Join(cwd, ".agents", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}

	// Write 5 session files, each with 1 decision.
	for i := 0; i < 5; i++ {
		sess := minedSessionFile{
			ID:        "s" + string(rune('0'+i)),
			Decisions: []string{"decision " + string(rune('0'+i))},
		}
		data, _ := json.Marshal(sess)
		path := filepath.Join(sessionsDir, "sess-"+string(rune('0'+i))+".json")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatalf("write sess %d: %v", i, err)
		}
	}

	report, err := RunMinePass(cwd, MineOpts{
		SessionsDir: sessionsDir,
		MaxSessions: 2,
	})
	if err != nil {
		t.Fatalf("RunMinePass: %v", err)
	}
	if report.SessionsRead != 2 {
		t.Errorf("SessionsRead = %d, want 2 (MaxSessions cap)", report.SessionsRead)
	}
	if len(report.Learnings) != 2 {
		t.Errorf("Learnings len = %d, want 2", len(report.Learnings))
	}
}
