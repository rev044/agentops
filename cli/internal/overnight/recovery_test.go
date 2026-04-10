package overnight

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeMarkerFile seeds a COMMIT-MARKER file on disk with the given body.
func writeMarkerFile(t *testing.T, overnightDir, iterationID, state string) string {
	t.Helper()
	if err := os.MkdirAll(overnightDir, 0o755); err != nil {
		t.Fatalf("mkdir overnight: %v", err)
	}
	body := markerBody{
		State:       state,
		IterationID: iterationID,
		StartedAt:   time.Now().UTC().Format(time.RFC3339Nano),
	}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal marker: %v", err)
	}
	path := filepath.Join(overnightDir, "COMMIT-MARKER."+iterationID)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	return path
}

func TestRecoverFromCrash_NoMarkerFile(t *testing.T) {
	cwd := t.TempDir()
	if err := os.MkdirAll(filepath.Join(cwd, ".agents", "overnight"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	actions, err := RecoverFromCrash(cwd)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(actions) != 0 {
		t.Fatalf("expected nil actions, got %v", actions)
	}
}

func TestRecoverFromCrash_NoOvernightDir(t *testing.T) {
	cwd := t.TempDir()
	actions, err := RecoverFromCrash(cwd)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(actions) != 0 {
		t.Fatalf("expected nil actions, got %v", actions)
	}
}

func TestRecoverFromCrash_DoneMarker_Cleanup(t *testing.T) {
	cwd := t.TempDir()
	overnightDir := filepath.Join(cwd, ".agents", "overnight")
	iter := "iter-done-1"
	writeMarkerFile(t, overnightDir, iter, markerStateDone)

	// Seed matching staging and prev dirs with contents.
	stagingDir := filepath.Join(overnightDir, "staging", iter)
	prevDir := filepath.Join(overnightDir, "prev."+iter)
	if err := os.MkdirAll(filepath.Join(stagingDir, ".agents", "learnings"), 0o755); err != nil {
		t.Fatalf("mkdir staging: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stagingDir, ".agents", "learnings", "x.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed staging file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(prevDir, "learnings"), 0o755); err != nil {
		t.Fatalf("mkdir prev: %v", err)
	}

	actions, err := RecoverFromCrash(cwd)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %v", actions)
	}
	if !strings.Contains(actions[0], "DONE") || !strings.Contains(actions[0], iter) {
		t.Errorf("action text missing DONE/iter: %q", actions[0])
	}

	if _, err := os.Stat(filepath.Join(overnightDir, "COMMIT-MARKER."+iter)); !os.IsNotExist(err) {
		t.Errorf("marker not removed: err=%v", err)
	}
	if _, err := os.Stat(stagingDir); !os.IsNotExist(err) {
		t.Errorf("staging not removed: err=%v", err)
	}
	if _, err := os.Stat(prevDir); !os.IsNotExist(err) {
		t.Errorf("prev not removed: err=%v", err)
	}
}

func TestRecoverFromCrash_ReadyMarker_ReverseSwap(t *testing.T) {
	cwd := t.TempDir()
	overnightDir := filepath.Join(cwd, ".agents", "overnight")
	liveDir := filepath.Join(cwd, ".agents")
	iter := "iter-ready-1"

	// Simulate mid-commit state:
	//   - live .agents/learnings/ is missing (rename-to-prev happened)
	//   - prev.iter/learnings/ contains the real content
	if err := os.MkdirAll(liveDir, 0o755); err != nil {
		t.Fatalf("mkdir live: %v", err)
	}
	prevLearnings := filepath.Join(overnightDir, "prev."+iter, "learnings")
	if err := os.MkdirAll(prevLearnings, 0o755); err != nil {
		t.Fatalf("mkdir prev learnings: %v", err)
	}
	content := []byte("real content")
	if err := os.WriteFile(filepath.Join(prevLearnings, "real.md"), content, 0o644); err != nil {
		t.Fatalf("seed prev file: %v", err)
	}

	// Staging dir exists (to be cleaned up).
	stagingDir := filepath.Join(overnightDir, "staging", iter)
	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		t.Fatalf("mkdir staging: %v", err)
	}

	writeMarkerFile(t, overnightDir, iter, markerStateReady)

	actions, err := RecoverFromCrash(cwd)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %v", actions)
	}
	if !strings.Contains(actions[0], "READY") {
		t.Errorf("expected READY in action: %q", actions[0])
	}

	// Live learnings should now contain the restored content.
	restored, err := os.ReadFile(filepath.Join(liveDir, "learnings", "real.md"))
	if err != nil {
		t.Fatalf("read restored: %v", err)
	}
	if string(restored) != string(content) {
		t.Errorf("expected %q, got %q", content, restored)
	}

	// Staging, prev, marker all gone.
	if _, err := os.Stat(stagingDir); !os.IsNotExist(err) {
		t.Errorf("staging not cleaned: %v", err)
	}
	if _, err := os.Stat(filepath.Join(overnightDir, "prev."+iter)); !os.IsNotExist(err) {
		t.Errorf("prev not cleaned: %v", err)
	}
	if _, err := os.Stat(filepath.Join(overnightDir, "COMMIT-MARKER."+iter)); !os.IsNotExist(err) {
		t.Errorf("marker not cleaned: %v", err)
	}
}

func TestRecoverFromCrash_MalformedMarker_Error(t *testing.T) {
	cwd := t.TempDir()
	overnightDir := filepath.Join(cwd, ".agents", "overnight")
	liveDir := filepath.Join(cwd, ".agents")
	if err := os.MkdirAll(overnightDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Seed .agents/learnings with content we want preserved untouched.
	learningsDir := filepath.Join(liveDir, "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatalf("mkdir learnings: %v", err)
	}
	origContent := []byte("keep me")
	origPath := filepath.Join(learningsDir, "keep.md")
	if err := os.WriteFile(origPath, origContent, 0o644); err != nil {
		t.Fatalf("seed learning: %v", err)
	}

	markerPath := filepath.Join(overnightDir, "COMMIT-MARKER.bad-iter")
	if err := os.WriteFile(markerPath, []byte("not json {{{"), 0o644); err != nil {
		t.Fatalf("seed marker: %v", err)
	}

	actions, err := RecoverFromCrash(cwd)
	if err == nil {
		t.Fatalf("expected error for malformed marker")
	}
	if !strings.Contains(err.Error(), "investigation") && !strings.Contains(err.Error(), "parse") {
		t.Errorf("error message should mention investigation/parse: %v", err)
	}
	if len(actions) == 0 {
		t.Errorf("expected degraded action entry, got none")
	}

	// .agents/learnings must be untouched.
	got, readErr := os.ReadFile(origPath)
	if readErr != nil {
		t.Fatalf("read original: %v", readErr)
	}
	if string(got) != string(origContent) {
		t.Errorf("live .agents was modified: %q", got)
	}
}

func TestRecoverFromCrash_MultipleMarkers_ProcessedInOrder(t *testing.T) {
	cwd := t.TempDir()
	overnightDir := filepath.Join(cwd, ".agents", "overnight")
	liveDir := filepath.Join(cwd, ".agents")
	if err := os.MkdirAll(liveDir, 0o755); err != nil {
		t.Fatalf("mkdir live: %v", err)
	}

	// iter-a: DONE; iter-b: READY.
	writeMarkerFile(t, overnightDir, "iter-a", markerStateDone)
	// Seed a prev for the READY marker so the swap actually reverses.
	prevLearnings := filepath.Join(overnightDir, "prev.iter-b", "learnings")
	if err := os.MkdirAll(prevLearnings, 0o755); err != nil {
		t.Fatalf("mkdir prev b: %v", err)
	}
	if err := os.WriteFile(filepath.Join(prevLearnings, "b.md"), []byte("b"), 0o644); err != nil {
		t.Fatalf("seed prev b: %v", err)
	}
	writeMarkerFile(t, overnightDir, "iter-b", markerStateReady)

	actions, err := RecoverFromCrash(cwd)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d: %v", len(actions), actions)
	}
	// Lexicographic order: COMMIT-MARKER.iter-a < COMMIT-MARKER.iter-b.
	if !strings.Contains(actions[0], "iter-a") {
		t.Errorf("expected iter-a first, got %q", actions[0])
	}
	if !strings.Contains(actions[1], "iter-b") {
		t.Errorf("expected iter-b second, got %q", actions[1])
	}
	if !strings.Contains(actions[0], "DONE") {
		t.Errorf("expected DONE in first: %q", actions[0])
	}
	if !strings.Contains(actions[1], "READY") {
		t.Errorf("expected READY in second: %q", actions[1])
	}

	// Verify reversal produced the learning file.
	if _, err := os.Stat(filepath.Join(liveDir, "learnings", "b.md")); err != nil {
		t.Errorf("expected restored b.md: %v", err)
	}
}

func TestLockIsStale_NoLockFile(t *testing.T) {
	dir := t.TempDir()
	stale, err := LockIsStale(filepath.Join(dir, "nope.lock"), time.Minute)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if stale {
		t.Error("expected not stale for missing file")
	}
}

func TestLockIsStale_FreshLock(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "fresh.lock")
	if err := WriteLockPID(lockPath); err != nil {
		t.Fatalf("write lock: %v", err)
	}
	stale, err := LockIsStale(lockPath, 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if stale {
		t.Error("expected fresh lock to be not stale")
	}
}

func TestLockIsStale_OldLockLivePID(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "old-live.lock")
	if err := WriteLockPID(lockPath); err != nil {
		t.Fatalf("write lock: %v", err)
	}
	// Backdate mtime to 24h ago.
	old := time.Now().Add(-24 * time.Hour)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	stale, err := LockIsStale(lockPath, time.Hour)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if stale {
		t.Error("expected not stale: pid is current (live) process")
	}
}

func TestLockIsStale_OldLockDeadPID(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "old-dead.lock")
	// Write a definitely-dead PID (impossibly large).
	if err := os.WriteFile(lockPath, []byte("2147480000\n"), 0o644); err != nil {
		t.Fatalf("write lock: %v", err)
	}
	old := time.Now().Add(-24 * time.Hour)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	stale, err := LockIsStale(lockPath, time.Hour)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !stale {
		t.Error("expected stale for old lock with dead PID")
	}
}

func TestReadLockPID_Missing(t *testing.T) {
	dir := t.TempDir()
	if pid := ReadLockPID(filepath.Join(dir, "nope.lock")); pid != 0 {
		t.Errorf("expected 0, got %d", pid)
	}
}

func TestReadLockPID_Malformed(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "bad.lock")
	if err := os.WriteFile(lockPath, []byte("not-a-pid"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if pid := ReadLockPID(lockPath); pid != 0 {
		t.Errorf("expected 0 for malformed, got %d", pid)
	}
}

func TestReadLockPID_Valid(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "valid.lock")
	if err := os.WriteFile(lockPath, []byte("12345\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if pid := ReadLockPID(lockPath); pid != 12345 {
		t.Errorf("expected 12345, got %d", pid)
	}
}

func TestWriteLockPID_CreatesAndOverwrites(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "sub", "pid.lock")

	if err := WriteLockPID(lockPath); err != nil {
		t.Fatalf("first write: %v", err)
	}
	first := ReadLockPID(lockPath)
	if first != os.Getpid() {
		t.Errorf("expected %d, got %d", os.Getpid(), first)
	}

	// Stuff some extra content in, then overwrite — must not append.
	if err := os.WriteFile(lockPath, []byte("111\nLEFTOVER\nMORE\n"), 0o644); err != nil {
		t.Fatalf("seed leftover: %v", err)
	}
	if err := WriteLockPID(lockPath); err != nil {
		t.Fatalf("second write: %v", err)
	}
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.Contains(string(data), "LEFTOVER") {
		t.Errorf("expected truncation, got %q", data)
	}
	if pid := ReadLockPID(lockPath); pid != os.Getpid() {
		t.Errorf("expected %d after rewrite, got %d", os.Getpid(), pid)
	}
}

func TestProcessAlive_CurrentProcess(t *testing.T) {
	if !ProcessAlive(os.Getpid()) {
		t.Error("current process should be alive")
	}
}

func TestProcessAlive_DefinitelyDead(t *testing.T) {
	// A PID of 0 is never a valid target.
	if ProcessAlive(0) {
		t.Error("pid 0 should not be alive")
	}
	if ProcessAlive(-1) {
		t.Error("negative pid should not be alive")
	}
	// Very large PID. Most systems cap at ~4M; 2_147_480_000 is safely dead.
	if ProcessAlive(2147480000) {
		t.Error("huge pid should not be alive")
	}
}
