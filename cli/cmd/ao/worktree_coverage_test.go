package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// ===========================================================================
// worktree.go — gcWorktreeCandidates (zero coverage)
// ===========================================================================

func TestCov3_worktree_gcWorktreeCandidates_empty(t *testing.T) {
	var candidates []worktreeGCCandidate
	liveRuns := make(map[string]bool)
	removed := gcWorktreeCandidates(candidates, liveRuns, "/fake/root", time.Now())
	if removed != 0 {
		t.Errorf("expected 0 removed for empty candidates, got %d", removed)
	}
}

func TestCov3_worktree_gcWorktreeCandidates_dryRun(t *testing.T) {
	// Save and restore dryRun
	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	candidates := []worktreeGCCandidate{
		{
			RunID:     "run-1",
			Path:      "/fake/path/repo-rpi-run-1",
			Reference: time.Now().Add(-2 * time.Hour),
			Dirty:     false,
		},
	}
	liveRuns := map[string]bool{"run-1": true}
	removed := gcWorktreeCandidates(candidates, liveRuns, "/fake/root", time.Now())
	if removed != 0 {
		t.Errorf("expected 0 removed in dry-run mode, got %d", removed)
	}
}

func TestCov3_worktree_gcWorktreeCandidates_realRemoveNonexistentPath(t *testing.T) {
	// Save and restore dryRun
	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = false

	tmp := t.TempDir()
	repoRoot := filepath.Join(tmp, "myrepo")
	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		t.Fatal(err)
	}

	// Candidate with nonexistent path — removeOrphanedWorktree will fail
	candidates := []worktreeGCCandidate{
		{
			RunID:     "run-1",
			Path:      filepath.Join(tmp, "myrepo-rpi-run-1"),
			Reference: time.Now().Add(-2 * time.Hour),
			Dirty:     false,
		},
	}
	liveRuns := map[string]bool{"run-1": true}
	removed := gcWorktreeCandidates(candidates, liveRuns, repoRoot, time.Now())
	// removeOrphanedWorktree may fail or succeed (path doesn't exist -> may error)
	_ = removed // just ensure no panic
}

// ===========================================================================
// worktree.go — isWorktreeDirty (zero coverage)
// ===========================================================================

func TestCov3_worktree_isWorktreeDirty_cleanRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmp := t.TempDir()

	// Init a git repo
	cmd := exec.Command("git", "init", tmp)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}

	// Configure git user for the test repo
	for _, args := range [][]string{
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	} {
		c := exec.Command("git", append([]string{"-C", tmp}, args...)...)
		if err := c.Run(); err != nil {
			t.Fatalf("git config: %v", err)
		}
	}

	// Create and commit a file so we have a clean worktree
	testFile := filepath.Join(tmp, "file.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	c := exec.Command("git", "-C", tmp, "add", ".")
	if err := c.Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	c = exec.Command("git", "-C", tmp, "commit", "-m", "init")
	if err := c.Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	dirty, err := isWorktreeDirty(tmp)
	if err != nil {
		t.Fatalf("isWorktreeDirty: %v", err)
	}
	if dirty {
		t.Error("expected clean repo to not be dirty")
	}
}

func TestCov3_worktree_isWorktreeDirty_dirtyRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmp := t.TempDir()

	cmd := exec.Command("git", "init", tmp)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}

	// Configure git user
	for _, args := range [][]string{
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	} {
		c := exec.Command("git", append([]string{"-C", tmp}, args...)...)
		if err := c.Run(); err != nil {
			t.Fatalf("git config: %v", err)
		}
	}

	// Create and commit, then modify
	testFile := filepath.Join(tmp, "file.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	c := exec.Command("git", "-C", tmp, "add", ".")
	if err := c.Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	c = exec.Command("git", "-C", tmp, "commit", "-m", "init")
	if err := c.Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	// Make dirty
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatal(err)
	}

	dirty, err := isWorktreeDirty(tmp)
	if err != nil {
		t.Fatalf("isWorktreeDirty: %v", err)
	}
	if !dirty {
		t.Error("expected modified repo to be dirty")
	}
}

func TestCov3_worktree_isWorktreeDirty_nonGitDir(t *testing.T) {
	tmp := t.TempDir()
	_, err := isWorktreeDirty(tmp)
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

// ===========================================================================
// worktree.go — resolveRepoRoot (zero coverage)
// ===========================================================================

func TestCov3_worktree_resolveRepoRoot_inGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmp := t.TempDir()
	cmd := exec.Command("git", "init", tmp)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}

	root, err := resolveRepoRoot(tmp)
	if err != nil {
		t.Fatalf("resolveRepoRoot: %v", err)
	}
	if root == "" {
		t.Fatal("expected non-empty repo root")
	}
}

func TestCov3_worktree_resolveRepoRoot_nonGitDir(t *testing.T) {
	tmp := t.TempDir()
	_, err := resolveRepoRoot(tmp)
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

// ===========================================================================
// worktree.go — findStaleRPITmuxSessions (zero coverage)
// ===========================================================================

func TestCov3_worktree_findStaleRPITmuxSessions_noTmux(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("PATH", tmp) // no tmux

	now := time.Now()
	activeRuns := make(map[string]bool)
	liveWorktreeRuns := make(map[string]bool)

	stale, err := findStaleRPITmuxSessions(now, time.Hour, activeRuns, liveWorktreeRuns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// listRPITmuxSessions returns nil when tmux not found
	if len(stale) != 0 {
		t.Errorf("expected 0 stale sessions without tmux, got %d", len(stale))
	}
}

// ===========================================================================
// worktree.go — listRPITmuxSessions (zero coverage)
// ===========================================================================

func TestCov3_worktree_listRPITmuxSessions_noTmux(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("PATH", tmp) // no tmux

	sessions, err := listRPITmuxSessions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions != nil {
		t.Errorf("expected nil sessions when tmux not found, got %v", sessions)
	}
}

// ===========================================================================
// worktree.go — finalizeWorktreeGC (zero coverage)
// ===========================================================================

func TestCov3_worktree_finalizeWorktreeGC_dryRun(t *testing.T) {
	origDryRun := dryRun
	defer func() { dryRun = origDryRun }()
	dryRun = true

	// Should not panic and should print dry-run message
	finalizeWorktreeGC("/fake/root", 3, 0, 0, 2)
}

func TestCov3_worktree_finalizeWorktreeGC_noPrune(t *testing.T) {
	origDryRun := dryRun
	origPrune := worktreeGCPrune
	defer func() {
		dryRun = origDryRun
		worktreeGCPrune = origPrune
	}()
	dryRun = false
	worktreeGCPrune = false

	// Should not panic
	finalizeWorktreeGC("/fake/root", 0, 0, 0, 0)
}

func TestCov3_worktree_finalizeWorktreeGC_withPrune(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	origDryRun := dryRun
	origPrune := worktreeGCPrune
	defer func() {
		dryRun = origDryRun
		worktreeGCPrune = origPrune
	}()
	dryRun = false
	worktreeGCPrune = true

	tmp := t.TempDir()
	cmd := exec.Command("git", "init", tmp)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}

	// finalizeWorktreeGC calls pruneWorktrees which runs git worktree prune
	finalizeWorktreeGC(tmp, 0, 0, 0, 0)
}

// ===========================================================================
// worktree.go — runWorktreeGC (zero coverage)
// ===========================================================================

func TestCov3_worktree_runWorktreeGC_invalidStaleAfter(t *testing.T) {
	origStaleAfter := worktreeGCStaleAfter
	defer func() { worktreeGCStaleAfter = origStaleAfter }()
	worktreeGCStaleAfter = 0

	err := runWorktreeGC(nil, nil)
	if err == nil {
		t.Fatal("expected error for stale-after <= 0")
	}
}

func TestCov3_worktree_runWorktreeGC_negativeStaleAfter(t *testing.T) {
	origStaleAfter := worktreeGCStaleAfter
	defer func() { worktreeGCStaleAfter = origStaleAfter }()
	worktreeGCStaleAfter = -time.Hour

	err := runWorktreeGC(nil, nil)
	if err == nil {
		t.Fatal("expected error for negative stale-after")
	}
}

func TestCov4_runWorktreeGC_dryRunInGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmp := t.TempDir()
	if err := exec.Command("git", "init", tmp).Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}

	// Save and restore working directory (process-global)
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	origDryRun := dryRun
	origStaleAfter := worktreeGCStaleAfter
	defer func() {
		dryRun = origDryRun
		worktreeGCStaleAfter = origStaleAfter
	}()

	dryRun = true
	worktreeGCStaleAfter = time.Hour

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := runWorktreeGC(nil, nil); err != nil {
		t.Fatalf("runWorktreeGC dry-run: %v", err)
	}
}

// ===========================================================================
// worktree.go — gcTmuxSessions (zero coverage)
// ===========================================================================

func TestCov3_worktree_gcTmuxSessions_noTmux(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("PATH", tmp) // no tmux

	origStaleAfter := worktreeGCStaleAfter
	defer func() { worktreeGCStaleAfter = origStaleAfter }()
	worktreeGCStaleAfter = time.Hour

	now := time.Now()
	activeRuns := make(map[string]bool)
	liveRuns := make(map[string]bool)

	killed, candidates := gcTmuxSessions(now, activeRuns, liveRuns)
	if killed != 0 {
		t.Errorf("expected 0 killed without tmux, got %d", killed)
	}
	if candidates != 0 {
		t.Errorf("expected 0 candidates without tmux, got %d", candidates)
	}
}

// ===========================================================================
// worktree.go — killTmuxSession (zero coverage)
// ===========================================================================

func TestCov3_worktree_killTmuxSession_noTmux(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("PATH", tmp) // no tmux

	err := killTmuxSession("ao-rpi-test-p1")
	if err == nil {
		t.Fatal("expected error when tmux is not on PATH")
	}
}

func TestCov3_worktree_killTmuxSession_nonexistentSession(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not available")
	}

	// Killing a non-existent session should return an error
	err := killTmuxSession("ao-rpi-nonexistent-session-xyz-p1")
	if err == nil {
		t.Fatal("expected error killing non-existent tmux session")
	}
}
