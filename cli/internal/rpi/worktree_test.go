package rpi

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEnsureAttachedBranch_DetachedHEAD(t *testing.T) {
	repo := initGitRepo(t)
	initialBranch, err := GetCurrentBranch(repo, 30*time.Second)
	if err != nil {
		t.Fatalf("GetCurrentBranch initial: %v", err)
	}

	sha := runGitOutput(t, repo, "rev-parse", "HEAD")
	runGit(t, repo, "checkout", strings.TrimSpace(sha))

	branch, healed, err := EnsureAttachedBranch(repo, 30*time.Second, "codex/auto-rpi")
	if err != nil {
		t.Fatalf("EnsureAttachedBranch: %v", err)
	}
	if !healed {
		t.Fatal("expected detached HEAD to be healed")
	}
	if branch != "codex/auto-rpi-recovery" {
		t.Fatalf("unexpected healed branch: %q", branch)
	}

	runGit(t, repo, "checkout", "--detach", strings.TrimSpace(sha))

	branch, healed, err = EnsureAttachedBranch(repo, 30*time.Second, "codex/auto-rpi")
	if err != nil {
		t.Fatalf("EnsureAttachedBranch: %v", err)
	}
	if !healed {
		t.Fatal("expected second detached heal to reuse stable branch")
	}
	if branch != "codex/auto-rpi-recovery" {
		t.Fatalf("unexpected healed branch on second run: %q", branch)
	}

	currentBranch, err := GetCurrentBranch(repo, 30*time.Second)
	if err != nil {
		t.Fatalf("GetCurrentBranch after heal: %v", err)
	}
	if currentBranch == "" {
		t.Fatal("expected current branch after named checkout")
	}
	if currentBranch != "codex/auto-rpi-recovery" {
		t.Fatalf("expected recovery branch, got %q", currentBranch)
	}
	baseBranch := initialBranch

	branches := listBranches(t, repo, "codex/auto-rpi-*")
	if len(branches) != 1 {
		t.Fatalf("expected one recovery branch, found %d (%v)", len(branches), branches)
	}
	if branches[0] != "codex/auto-rpi-recovery" {
		t.Fatalf("expected only codex/auto-rpi-recovery, got %q", branches[0])
	}

	runGit(t, repo, "checkout", baseBranch)
	currentBranch, err = GetCurrentBranch(repo, 30*time.Second)
	if err != nil {
		t.Fatalf("GetCurrentBranch after checkout %s: %v", baseBranch, err)
	}
	if currentBranch != baseBranch {
		t.Fatalf("expected %s after checkout, got %q", baseBranch, currentBranch)
	}
}

func TestEnsureAttachedBranch_NoopOnNamedBranch(t *testing.T) {
	repo := initGitRepo(t)

	current, err := GetCurrentBranch(repo, 30*time.Second)
	if err != nil {
		t.Fatalf("GetCurrentBranch: %v", err)
	}

	branch, healed, err := EnsureAttachedBranch(repo, 30*time.Second, "codex/auto-rpi")
	if err != nil {
		t.Fatalf("EnsureAttachedBranch: %v", err)
	}
	if healed {
		t.Fatal("expected no heal on named branch")
	}
	if branch != current {
		t.Fatalf("branch mismatch: got %q want %q", branch, current)
	}
}

func TestEnsureAttachedBranch_DetachedHEAD_WorktreeConflictFallsBackDetached(t *testing.T) {
	repo := initGitRepo(t)

	worktreeRoot := t.TempDir()
	conflictingBranch := "codex/auto-rpi-recovery"
	runGit(t, repo, "branch", "-f", conflictingBranch, "HEAD")

	conflictPath := filepath.Join(worktreeRoot, "conflict")
	if err := os.MkdirAll(conflictPath, 0755); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "worktree", "add", conflictPath, conflictingBranch)
	defer runGitIgnoreError(t, repo, "worktree", "remove", "--force", conflictPath)

	runGit(t, repo, "checkout", "--detach", "HEAD")

	branch, healed, err := EnsureAttachedBranch(repo, 30*time.Second, "codex/auto-rpi")
	if err != nil {
		t.Fatalf("EnsureAttachedBranch: %v", err)
	}
	if healed {
		t.Fatal("expected no recovery branch switch when branch is used by another worktree")
	}
	if branch != "" {
		t.Fatalf("expected detached path with no switch, got %q", branch)
	}

	if _, err := GetCurrentBranch(repo, 30*time.Second); err == nil {
		t.Fatal("expected repository to remain detached when recovery branch is unavailable")
	}

	branches := listBranches(t, repo, "codex/auto-rpi-*")
	if len(branches) != 1 {
		t.Fatalf("expected one recovery branch pattern entry, found %d (%v)", len(branches), branches)
	}
}

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")

	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "initial")
	return dir
}

func runGit(t *testing.T, cwd string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}

func runGitOutput(t *testing.T, cwd string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s output failed: %v", strings.Join(args, " "), err)
	}
	return string(out)
}

func listBranches(t *testing.T, cwd string, pattern string) []string {
	t.Helper()
	cmd := exec.Command("git", "branch", "--list", pattern)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git branch --list %q failed: %v\n%s", pattern, err, string(out))
	}
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "* ")
		branches = append(branches, line)
	}
	return branches
}

func runGitIgnoreError(t *testing.T, cwd string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	_ = cmd.Run()
}

func TestClassifyWorktreeError_AlreadyExists(t *testing.T) {
	output := []byte("fatal: '/path/foo' already exists")
	retryable, err := classifyWorktreeError(output, nil, nil, 30*time.Second)
	if !retryable {
		t.Error("expected retryable for 'already exists' output")
	}
	if err != nil {
		t.Errorf("expected nil error for retryable case, got: %v", err)
	}
}

func TestClassifyWorktreeError_Timeout(t *testing.T) {
	output := []byte("signal: killed")
	retryable, err := classifyWorktreeError(output, context.DeadlineExceeded, nil, 30*time.Second)
	if retryable {
		t.Error("expected non-retryable for timeout")
	}
	if err == nil {
		t.Fatal("expected error for timeout")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected 'timed out' in error, got: %v", err)
	}
}

func TestClassifyWorktreeError_GenericFailure(t *testing.T) {
	cmdErr := os.ErrPermission
	output := []byte("fatal: unable to create")
	retryable, err := classifyWorktreeError(output, nil, cmdErr, 30*time.Second)
	if retryable {
		t.Error("expected non-retryable for generic failure")
	}
	if err == nil {
		t.Fatal("expected error for generic failure")
	}
	if !strings.Contains(err.Error(), "git worktree add failed") {
		t.Errorf("expected 'git worktree add failed' in error, got: %v", err)
	}
}

func TestInitWorktreeAgentsDir_Success(t *testing.T) {
	tmpDir := t.TempDir()
	// Should create the directory without error
	initWorktreeAgentsDir(tmpDir, nil)

	agentsDir := filepath.Join(tmpDir, ".agents", "rpi")
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		t.Error("expected .agents/rpi directory to be created")
	}
}

func TestInitWorktreeAgentsDir_WarningLogged(t *testing.T) {
	// Use a read-only path to trigger warning
	var logged bool
	verbosef := func(format string, args ...any) {
		logged = true
	}
	// A path that cannot be created (nested under a file)
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "blocker")
	if err := os.WriteFile(filePath, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	initWorktreeAgentsDir(filePath, verbosef)
	if !logged {
		t.Error("expected warning to be logged when MkdirAll fails")
	}
}

func TestAcquireMergeLock_Basic(t *testing.T) {
	tmp := t.TempDir()
	lock, err := acquireMergeLock(tmp)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	defer releaseMergeLock(lock)

	lockPath := filepath.Join(tmp, ".git", "agentops", "merge.lock")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("expected merge.lock file to exist")
	}
}

func TestAcquireMergeLock_CreatesDir(t *testing.T) {
	tmp := t.TempDir()
	// .agents/rpi/ doesn't exist yet
	lock, err := acquireMergeLock(tmp)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	releaseMergeLock(lock)

	dir := filepath.Join(tmp, ".git", "agentops")
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected .git/agentops/ to be created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected .git/agentops/ to be a directory")
	}
}

func TestReleaseMergeLock_NilSafe(t *testing.T) {
	// Should not panic
	releaseMergeLock(nil)
}

func TestTryCreateWorktree_InvalidRepoRoot(t *testing.T) {
	repo := t.TempDir()

	worktreePath, runID, err := tryCreateWorktree(repo, "HEAD", 200*time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected error for non-git repo root")
	}
	if worktreePath != "" || runID != "" {
		t.Fatalf("expected empty outputs on error, got path=%q runID=%q", worktreePath, runID)
	}
}

func TestWaitForCleanRepo_CleanRepo(t *testing.T) {
	repo := initGitRepo(t)
	if err := waitForCleanRepo(repo, time.Second, nil); err != nil {
		t.Fatalf("waitForCleanRepo() error = %v, want nil", err)
	}
}

func TestResolveMergeSource_InvalidPath(t *testing.T) {
	_, err := resolveMergeSource(filepath.Join(t.TempDir(), "missing"), 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected error for invalid worktree path")
	}
	if !strings.Contains(err.Error(), "resolve worktree merge source") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAttemptBranchHeal_EmptyOutputError(t *testing.T) {
	// Exercise attemptBranchHeal where branch create fails with empty output,
	// reaching the bare ErrDetachedSelfHealFailed return.
	emptyRepo := t.TempDir()
	cmd := exec.Command("git", "init", emptyRepo)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = emptyRepo
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = emptyRepo
	_ = cmd.Run()

	_, _, err := attemptBranchHeal(emptyRepo, 5*time.Second, "test-recovery")
	if err == nil {
		t.Fatal("expected error from attemptBranchHeal on empty repo")
	}
}

func TestAttemptBranchSwitch_WorktreeBusy(t *testing.T) {
	// Exercise attemptBranchSwitch where switch fails with "used by worktree".
	repo := initGitRepo(t)

	branchName := "switch-busy-test"
	runGit(t, repo, "branch", branchName)

	wtDir := t.TempDir()
	wtPath := filepath.Join(wtDir, "wt")
	runGit(t, repo, "worktree", "add", wtPath, branchName)
	defer runGitIgnoreError(t, repo, "worktree", "remove", "--force", wtPath)

	branch, healed, err := attemptBranchSwitch(repo, 5*time.Second, branchName)
	if err != nil {
		t.Fatalf("expected nil error for worktree-busy fallback, got: %v", err)
	}
	if healed {
		t.Error("expected healed=false for worktree-busy case")
	}
	if branch != "" {
		t.Errorf("expected empty branch for worktree-busy case, got %q", branch)
	}
}

func TestHandleMergeFailure_Timeout(t *testing.T) {
	// Exercise handleMergeFailure with context deadline exceeded.
	repo := initGitRepo(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond)

	err := handleMergeFailure(repo, "abc123", "abc1", ctx, os.ErrPermission, 5*time.Second)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected 'timed out' in error, got: %v", err)
	}
}

func TestHandleMergeFailure_ConflictListFails(t *testing.T) {
	// Exercise handleMergeFailure where git diff fails (non-git dir) and
	// conflict files are empty, returning "git merge failed".
	nonGitDir := t.TempDir()
	ctx := context.Background()

	err := handleMergeFailure(nonGitDir, "abc123def456", "abc123def456", ctx, os.ErrPermission, 5*time.Second)
	if err == nil {
		t.Fatal("expected error from handleMergeFailure")
	}
	if !strings.Contains(err.Error(), "git merge failed") {
		t.Errorf("expected 'git merge failed' in error, got: %v", err)
	}
}

func TestAcquireMergeLock_MkdirAllFails(t *testing.T) {
	// Exercise acquireMergeLock where MkdirAll fails.
	tmp := t.TempDir()
	gitDir := filepath.Join(tmp, ".git")
	if err := os.WriteFile(gitDir, []byte("not a dir"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := acquireMergeLock(tmp)
	if err == nil {
		t.Fatal("expected error when lock dir cannot be created")
	}
	if !strings.Contains(err.Error(), "create lock dir") {
		t.Errorf("expected 'create lock dir' error, got: %v", err)
	}
}

func TestAcquireMergeLock_OpenFileFails(t *testing.T) {
	// Exercise acquireMergeLock where os.OpenFile fails.
	tmp := t.TempDir()
	lockDir := filepath.Join(tmp, ".git", "agentops")
	if err := os.MkdirAll(lockDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(lockDir, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(lockDir, 0750) })

	_, err := acquireMergeLock(tmp)
	if err == nil {
		t.Fatal("expected error when lock file cannot be opened")
	}
	if !strings.Contains(err.Error(), "open merge lock") {
		t.Errorf("expected 'open merge lock' error, got: %v", err)
	}
}

func TestMergeWorktree_LockDirBlocked(t *testing.T) {
	// Exercise MergeWorktree where acquireMergeLock fails.
	repo := initGitRepo(t)

	agentopsFile := filepath.Join(repo, ".git", "agentops")
	if err := os.WriteFile(agentopsFile, []byte("blocker"), 0600); err != nil {
		t.Fatal(err)
	}

	err := MergeWorktree(repo, "/fake/path", "fakerun", 5*time.Second, nil)
	if err == nil {
		t.Fatal("expected error for merge with blocked lock dir")
	}
	if !strings.Contains(err.Error(), "merge lock") {
		t.Errorf("expected 'merge lock' in error, got: %v", err)
	}
}

func TestWaitForCleanRepo_GitStatusTimeout(t *testing.T) {
	// Exercise waitForCleanRepo timeout path.
	repo := initGitRepo(t)
	err := waitForCleanRepo(repo, 1*time.Nanosecond, nil)
	if err == nil {
		t.Skip("git status completed faster than 1ns timeout")
	}
	if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "uncommitted") {
		t.Errorf("expected timeout or unclean error, got: %v", err)
	}
}

func TestResolveAbsPath_EvalSymlinksFails_AbsSucceeds(t *testing.T) {
	// Exercise resolveAbsPath where EvalSymlinks fails but filepath.Abs succeeds.
	brokenLink := filepath.Join(t.TempDir(), "broken-link")
	if err := os.Symlink("/nonexistent/target/path", brokenLink); err != nil {
		t.Fatal(err)
	}
	deepPath := filepath.Join(brokenLink, "sub", "dir")

	absPath, err := resolveAbsPath(deepPath)
	if err != nil {
		t.Fatalf("resolveAbsPath should succeed via Abs fallback, got: %v", err)
	}
	if !filepath.IsAbs(absPath) {
		t.Errorf("expected absolute path, got: %q", absPath)
	}
}

func TestAttemptBranchHeal_BranchCreateTimeout(t *testing.T) {
	// Exercise attemptBranchHeal with very short timeout to potentially reach
	// empty output error path.
	repo := initGitRepo(t)
	sha := strings.TrimSpace(runGitOutput(t, repo, "rev-parse", "HEAD"))
	runGit(t, repo, "checkout", sha)

	_, _, err := attemptBranchHeal(repo, 1*time.Nanosecond, "timeout-test-recovery")
	if err == nil {
		t.Skip("git branch completed faster than 1ns timeout")
	}
}

func TestLockFile_ClosedFd(t *testing.T) {
	// Exercise lockFile with a closed file descriptor.
	tmp := t.TempDir()
	lockPath := filepath.Join(tmp, "test.lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	lockErr := lockFile(f)
	if lockErr == nil {
		t.Skip("lockFile succeeded on closed fd (OS dependent)")
	}
}
