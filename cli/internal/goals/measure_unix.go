//go:build !windows

package goals

import (
	"os/exec"
	"syscall"
)

// configureProcGroup sets up process-group isolation on POSIX systems so that
// child processes spawned by bash are killed together on timeout.
func configureProcGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		// Kill the entire process group, not just the parent.
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
