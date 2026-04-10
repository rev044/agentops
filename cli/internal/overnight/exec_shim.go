package overnight

import (
	"context"
	"os/exec"
)

// ExecCommand is the package-level exec.Command hook used by every
// overnight stage that needs to shell out to an external tool. Tests
// swap this variable to intercept forbidden subprocess invocations
// (git, /rpi, etc.) without mocking individual call sites.
//
// The boundary tests in boundary_test.go install a panicking shim on
// this variable to enforce the "Dream never invokes /rpi" and
// "Dream never touches git" anti-goals mechanically.
//
// Default: exec.Command from the standard library.
var ExecCommand = exec.Command

// ExecCommandContext is the context-aware companion to ExecCommand.
// inject_refresh.go uses this for the subprocess fallback when the
// in-process inject-cache rebuild is unavailable. Boundary tests can
// intercept it to catch forbidden subprocess invocations even when
// they go through the context-aware API.
//
// Default: exec.CommandContext from the standard library.
var ExecCommandContext = exec.CommandContext

// ExecLookPath is the package-level exec.LookPath hook. inject_refresh.go
// uses this to locate the `ao` binary before the subprocess fallback.
// Boundary tests can intercept it to simulate a missing binary without
// mutating the real $PATH.
//
// Default: exec.LookPath from the standard library.
var ExecLookPath = exec.LookPath

// execCommandContextShim is a small wrapper that keeps the
// ExecCommandContext type signature the same as exec.CommandContext so
// callers can swap it in and out cleanly. It exists for symmetry with
// ExecCommand; tests substitute ExecCommandContext directly.
var _ = func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return ExecCommandContext(ctx, name, args...)
}
