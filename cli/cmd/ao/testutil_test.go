package main

// testutil_test.go consolidates test helper functions that are shared across
// multiple test files in package main. Each helper was originally defined in a
// single file but called from others; moving them here avoids duplication and
// makes cross-file dependencies explicit.

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Stdout capture helpers
// ---------------------------------------------------------------------------

// captureStdout redirects os.Stdout to a pipe, calls fn, and returns everything
// written to stdout along with fn's error.
// Origin: rpi_verify_test.go
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	runErr := fn()

	_ = w.Close()
	os.Stdout = oldStdout
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	return string(out), runErr
}

// captureJSONStdout redirects os.Stdout, calls fn (no return value), and
// returns everything written. Useful for commands that print JSON.
// Origin: json_validity_test.go
func captureJSONStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

// ---------------------------------------------------------------------------
// Working-directory helpers
// ---------------------------------------------------------------------------

// chdirTemp creates a temp directory, chdir's into it, and registers cleanup
// to restore the original working directory.
// Origin: doctor_test.go
func chdirTemp(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
	return tmp
}

// chdirTo changes to the specified directory, registers a cleanup to restore
// the previous working directory, and returns the previous working directory
// for backward compatibility.
// Origin: constraint_cmd_test.go
func chdirTo(t *testing.T, wd string) string {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
	return prev
}

// ---------------------------------------------------------------------------
// File / directory setup helpers
// ---------------------------------------------------------------------------

// setupAgentsDir creates the .agents/ao directory structure in the given dir.
// Origin: cobra_commands_test.go
func setupAgentsDir(t *testing.T, dir string) {
	t.Helper()
	dirs := []string{
		".agents/ao/sessions",
		".agents/ao/index",
		".agents/ao/provenance",
		".agents/ao/metrics",
		".agents/learnings",
		".agents/research",
		".agents/patterns",
		".agents/retros",
		".agents/plans",
		".agents/council",
		".agents/knowledge/pending",
		".agents/rpi",
		".agents/constraints",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}
}

// writeFile creates parent directories if needed and writes content to path.
// Origin: helpers2_test.go
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// Git helpers
// ---------------------------------------------------------------------------

type gitCommitFixture struct {
	Path      string
	Content   string
	Message   string
	Timestamp time.Time
}

func initGitHistoryFixtureRepo(t *testing.T, commits []gitCommitFixture) string {
	t.Helper()
	dir := t.TempDir()
	initHistoryFixtureGitRepo(t, dir)
	for i, commit := range commits {
		path := commit.Path
		if path == "" {
			path = fmt.Sprintf("fixture-%d.txt", i)
		}
		content := commit.Content
		if content == "" {
			content = fmt.Sprintf("%s at %s\n", commit.Message, commit.Timestamp.Format(time.RFC3339))
		}
		writeFile(t, filepath.Join(dir, path), content)
		runFixtureGit(
			t,
			dir,
			nil,
			"add",
			path,
		)
		runFixtureGit(
			t,
			dir,
			[]string{
				"GIT_AUTHOR_DATE=" + commit.Timestamp.Format(time.RFC3339),
				"GIT_COMMITTER_DATE=" + commit.Timestamp.Format(time.RFC3339),
			},
			"commit",
			"-m",
			commit.Message,
		)
	}
	return dir
}

// initTestRepo creates a temp directory with a git repo containing one commit.
// Origin: rpi_phased_worktree_test.go
func initTestRepo(t *testing.T) string {
	t.Helper()
	return initGitHistoryFixtureRepo(t, []gitCommitFixture{{
		Path:      "README.md",
		Content:   "# Test\n",
		Message:   "Initial commit",
		Timestamp: time.Now().Add(-1 * time.Hour).UTC(),
	}})
}

func initHistoryFixtureGitRepo(t *testing.T, dir string) {
	t.Helper()
	runFixtureGit(t, dir, nil, "init")
	runFixtureGit(t, dir, nil, "config", "user.email", "test@test.com")
	runFixtureGit(t, dir, nil, "config", "user.name", "Test")
}

func runFixtureGit(t *testing.T, dir string, extraEnv []string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return string(out)
}
