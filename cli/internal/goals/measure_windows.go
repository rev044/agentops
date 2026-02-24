//go:build windows

package goals

import "os/exec"

// configureProcGroup is a no-op on Windows where POSIX process groups are
// unavailable. The default exec.CommandContext cancel behaviour (Process.Kill)
// is used instead.
func configureProcGroup(_ *exec.Cmd) {}
