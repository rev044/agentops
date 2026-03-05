package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"os/exec"
)

// ---------------------------------------------------------------------------
// runIDFromWorktreePath
// ---------------------------------------------------------------------------

func TestWorktree_runIDFromWorktreePath(t *testing.T) {
	tests := []struct {
		name         string
		repoRoot     string
		worktreePath string
		want         string
	}{
		{
			name:         "valid rpi worktree",
			repoRoot:     "/home/user/agentops",
			worktreePath: "/home/user/agentops-rpi-abc123",
			want:         "abc123",
		},
		{
			name:         "nested path",
			repoRoot:     "/home/user/my-project",
			worktreePath: "/home/user/my-project-rpi-run42",
			want:         "run42",
		},
		{
			name:         "non-matching prefix",
			repoRoot:     "/home/user/agentops",
			worktreePath: "/home/user/other-project-rpi-abc",
			want:         "",
		},
		{
			name:         "no run ID",
			repoRoot:     "/home/user/agentops",
			worktreePath: "/home/user/agentops-rpi-",
			want:         "",
		},
		{
			name:         "exact repo name, no rpi suffix",
			repoRoot:     "/home/user/agentops",
			worktreePath: "/home/user/agentops",
			want:         "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runIDFromWorktreePath(tt.repoRoot, tt.worktreePath)
			if got != tt.want {
				t.Errorf("runIDFromWorktreePath(%q, %q) = %q, want %q",
					tt.repoRoot, tt.worktreePath, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// worktreeReferenceTime
// ---------------------------------------------------------------------------

func TestWorktree_worktreeReferenceTime(t *testing.T) {
	t.Run("uses most recent modtime among candidates", func(t *testing.T) {
		tmp := t.TempDir()
		rpiDir := filepath.Join(tmp, ".agents", "rpi")
		if err := os.MkdirAll(rpiDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Create phased-state.json with a specific modtime
		statePath := filepath.Join(rpiDir, "phased-state.json")
		if err := os.WriteFile(statePath, []byte(`{}`), 0644); err != nil {
			t.Fatal(err)
		}
		targetTime := time.Now().Add(-2 * time.Hour)
		if err := os.Chtimes(statePath, targetTime, targetTime); err != nil {
			t.Fatal(err)
		}

		// The directory itself was just created, so it's newer
		got := worktreeReferenceTime(tmp)
		if got.IsZero() {
			t.Error("expected non-zero reference time")
		}
		// The reference time should be the modtime of the worktree dir (most recent)
		dirInfo, _ := os.Stat(tmp)
		if !got.Equal(dirInfo.ModTime()) {
			// Could be either the dir or the file, depending on creation order
			if got.Before(targetTime) {
				t.Error("reference time should be >= oldest candidate file time")
			}
		}
	})

	t.Run("returns epoch for nonexistent path", func(t *testing.T) {
		got := worktreeReferenceTime("/nonexistent/path")
		if !got.Equal(time.Unix(0, 0)) {
			t.Errorf("expected epoch time, got %v", got)
		}
	})

	t.Run("uses worktree dir itself if no rpi files", func(t *testing.T) {
		tmp := t.TempDir()
		got := worktreeReferenceTime(tmp)
		if got.IsZero() {
			t.Error("expected non-zero time from worktree dir stat")
		}
	})
}

// ---------------------------------------------------------------------------
// parseRPITmuxSessionRunID
// ---------------------------------------------------------------------------

func TestWorktree_parseRPITmuxSessionRunID(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		wantRunID   string
		wantOK      bool
	}{
		{"valid p1", "ao-rpi-abc123-p1", "abc123", true},
		{"valid p2", "ao-rpi-run-42-p2", "run-42", true},
		{"valid p3", "ao-rpi-xyz-p3", "xyz", true},
		{"wrong prefix", "tmux-rpi-abc-p1", "", false},
		{"no phase suffix", "ao-rpi-abc", "", false},
		{"wrong phase", "ao-rpi-abc-p4", "", false},
		{"empty", "", "", false},
		{"just prefix", "ao-rpi--p1", "", false},
		{"non-rpi session", "main", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRunID, gotOK := parseRPITmuxSessionRunID(tt.sessionName)
			if gotOK != tt.wantOK {
				t.Errorf("ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotRunID != tt.wantRunID {
				t.Errorf("runID = %q, want %q", gotRunID, tt.wantRunID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseTmuxSessionListOutput
// ---------------------------------------------------------------------------

func TestWorktree_parseTmuxSessionListOutput(t *testing.T) {
	t.Run("normal output", func(t *testing.T) {
		output := "ao-rpi-abc-p1\t1708000000\nao-rpi-abc-p2\t1708000100\nmain\t1708000200\n"
		sessions := parseTmuxSessionListOutput(output)
		if len(sessions) != 2 {
			t.Errorf("expected 2 RPI sessions, got %d", len(sessions))
		}
		if sessions[0].RunID != "abc" {
			t.Errorf("first session RunID = %q, want abc", sessions[0].RunID)
		}
		if sessions[0].Name != "ao-rpi-abc-p1" {
			t.Errorf("first session Name = %q", sessions[0].Name)
		}
	})

	t.Run("empty output", func(t *testing.T) {
		sessions := parseTmuxSessionListOutput("")
		if len(sessions) != 0 {
			t.Errorf("expected 0 sessions for empty output, got %d", len(sessions))
		}
	})

	t.Run("malformed lines ignored", func(t *testing.T) {
		output := "ao-rpi-x-p1\t1708000000\nbadline\nao-rpi-y-p2\tnot_a_number\n"
		sessions := parseTmuxSessionListOutput(output)
		if len(sessions) != 1 {
			t.Errorf("expected 1 valid session, got %d", len(sessions))
		}
	})

	t.Run("whitespace handling", func(t *testing.T) {
		output := "  ao-rpi-run1-p1 \t 1708000000 \n"
		sessions := parseTmuxSessionListOutput(output)
		if len(sessions) != 1 {
			t.Errorf("expected 1 session after whitespace trim, got %d", len(sessions))
		}
	})
}

// ---------------------------------------------------------------------------
// shouldCleanupRPITmuxSession
// ---------------------------------------------------------------------------

func TestWorktree_shouldCleanupRPITmuxSession(t *testing.T) {
	now := time.Now()
	staleAfter := 24 * time.Hour
	activeRuns := map[string]bool{"active-run": true}
	liveWorktreeRuns := map[string]bool{"wt-run": true}

	tests := []struct {
		name      string
		runID     string
		createdAt time.Time
		want      bool
	}{
		{"empty runID", "", now.Add(-48 * time.Hour), false},
		{"active run", "active-run", now.Add(-48 * time.Hour), false},
		{"live worktree run", "wt-run", now.Add(-48 * time.Hour), false},
		{"too recent", "old-run", now.Add(-12 * time.Hour), false},
		{"stale and orphaned", "orphan-run", now.Add(-48 * time.Hour), true},
		{"exactly at threshold", "border-run", now.Add(-24 * time.Hour), true},
		{"just past threshold", "past-run", now.Add(-25 * time.Hour), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldCleanupRPITmuxSession(tt.runID, tt.createdAt, now, staleAfter, activeRuns, liveWorktreeRuns)
			if got != tt.want {
				t.Errorf("shouldCleanupRPITmuxSession(%q, ...) = %v, want %v", tt.runID, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// worktreeGCCandidate struct
// ---------------------------------------------------------------------------

func TestWorktree_worktreeGCCandidate_Fields(t *testing.T) {
	c := worktreeGCCandidate{
		RunID:     "test-run",
		Path:      "/path/to/worktree",
		Reference: time.Now(),
		Dirty:     true,
	}
	if c.RunID != "test-run" {
		t.Error("RunID not set")
	}
	if c.Path != "/path/to/worktree" {
		t.Error("Path not set")
	}
	if !c.Dirty {
		t.Error("Dirty not set")
	}
}

// ---------------------------------------------------------------------------
// tmuxSessionMeta struct
// ---------------------------------------------------------------------------

func TestWorktree_tmuxSessionMeta_Fields(t *testing.T) {
	m := tmuxSessionMeta{
		Name:      "ao-rpi-abc-p1",
		RunID:     "abc",
		CreatedAt: time.Unix(1708000000, 0),
	}
	if m.Name != "ao-rpi-abc-p1" {
		t.Error("Name not set")
	}
	if m.RunID != "abc" {
		t.Error("RunID not set")
	}
}

// ---------------------------------------------------------------------------
// findRPISiblingWorktreePaths
// ---------------------------------------------------------------------------

func TestWorktree_findRPISiblingWorktreePaths(t *testing.T) {
	// Create a mock repo root with sibling worktrees
	parent := t.TempDir()
	repoRoot := filepath.Join(parent, "myproject")
	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		t.Fatal(err)
	}

	// Create sibling worktree directories
	for _, suffix := range []string{"rpi-run1", "rpi-run2"} {
		wtDir := filepath.Join(parent, "myproject-"+suffix)
		if err := os.MkdirAll(wtDir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create a non-matching sibling
	otherDir := filepath.Join(parent, "other-project")
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatal(err)
	}

	paths, err := findRPISiblingWorktreePaths(repoRoot)
	if err != nil {
		t.Fatalf("findRPISiblingWorktreePaths: %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("expected 2 worktree paths, got %d: %v", len(paths), paths)
	}

	for _, p := range paths {
		base := filepath.Base(p)
		if !strings.HasPrefix(base, "myproject-rpi-") {
			t.Errorf("unexpected worktree path: %s", p)
		}
	}
}

func TestWorktree_findRPISiblingWorktreePaths_NoSiblings(t *testing.T) {
	parent := t.TempDir()
	repoRoot := filepath.Join(parent, "lonely-project")
	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		t.Fatal(err)
	}

	paths, err := findRPISiblingWorktreePaths(repoRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(paths))
	}
}

// ---------------------------------------------------------------------------
// tmuxRPISessionPattern regex
// ---------------------------------------------------------------------------

func TestWorktree_tmuxRPISessionPattern(t *testing.T) {
	tests := []struct {
		name  string
		input string
		match bool
	}{
		{"valid p1", "ao-rpi-abc-p1", true},
		{"valid p2", "ao-rpi-long-run-id-p2", true},
		{"valid p3", "ao-rpi-x-p3", true},
		{"no match: wrong prefix", "tmux-rpi-abc-p1", false},
		{"no match: missing phase", "ao-rpi-abc", false},
		{"no match: p4", "ao-rpi-abc-p4", false},
		{"no match: empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tmuxRPISessionPattern.MatchString(tt.input)
			if got != tt.match {
				t.Errorf("tmuxRPISessionPattern.MatchString(%q) = %v, want %v", tt.input, got, tt.match)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// finalizeWorktreeGC (output only, verifying it does not panic)
// ---------------------------------------------------------------------------

func TestWorktree_finalizeWorktreeGC_DryRun(t *testing.T) {
	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	// Should not panic
	finalizeWorktreeGC("/fake/repo", 5, 3, 2, 4)
}

func TestWorktree_finalizeWorktreeGC_NotDryRun(t *testing.T) {
	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	// Save and restore worktreeGCPrune since we test with it off
	oldPrune := worktreeGCPrune
	worktreeGCPrune = false
	defer func() { worktreeGCPrune = oldPrune }()

	// Should not panic (prune disabled so no git command needed)
	finalizeWorktreeGC("/fake/repo", 5, 3, 2, 4)
}

// ---------------------------------------------------------------------------
// gcWorktreeCandidates (no candidates case)
// ---------------------------------------------------------------------------

func TestWorktree_gcWorktreeCandidates_Empty(t *testing.T) {
	liveRuns := make(map[string]bool)
	removed := gcWorktreeCandidates(nil, liveRuns, "/fake/repo", time.Now())
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestWorktree_gcWorktreeCandidates_DryRun(t *testing.T) {
	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	candidates := []worktreeGCCandidate{
		{RunID: "test-run", Path: "/fake/path", Reference: time.Now().Add(-48 * time.Hour)},
	}
	liveRuns := make(map[string]bool)
	removed := gcWorktreeCandidates(candidates, liveRuns, "/fake/repo", time.Now())
	if removed != 0 {
		t.Errorf("expected 0 removed in dry-run, got %d", removed)
	}
}

// NOTE: Tests for isWorktreeDirty, removeOrphanedWorktree, resolveRepoRoot,
// discoverActiveRPIRuns, and killTmuxSession require actual git repos or
// tmux processes and are better suited for integration tests.

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
