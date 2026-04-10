package overnight

import "os/exec"

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
