package main

// testutil_test.go consolidates test helper functions that are shared across
// multiple test files in package main. Each helper was originally defined in a
// single file but called from others; moving them here avoids duplication and
// makes cross-file dependencies explicit.

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
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

// cov3W2CaptureStdout redirects os.Stdout, calls fn, and returns the output.
// Uses a fixed 64KB buffer.
// Origin: maturity_deep_coverage_test.go
func cov3W2CaptureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 65536)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
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

// chdirTo changes to the specified directory and returns the previous working
// directory. Caller is responsible for restoring it.
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
	return prev
}

// cov3W2ChdirTemp changes to dir and registers cleanup to restore.
// Origin: maturity_deep_coverage_test.go
func cov3W2ChdirTemp(t *testing.T, dir string) {
	t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })
}

// setupTempWorkdir creates a temp directory, chdir's into it, and registers
// cleanup. Same pattern as chdirTemp but originated from cobra_commands_test.go.
// Origin: cobra_commands_test.go
func setupTempWorkdir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
	return tmp
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

// initTestRepo creates a temp directory with a git repo containing one commit.
// Origin: rpi_phased_worktree_test.go
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git init setup (%v): %v\n%s", args, err, out)
		}
	}
	// Create a file and commit so HEAD exists.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", "Initial commit"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit setup (%v): %v\n%s", args, err, out)
		}
	}
	return dir
}
