//go:build windows

package goals

import (
	"fmt"
	"os/exec"
	"strconv"
	"syscall"
)

// configureProcGroup sets up process-tree cleanup on Windows.
// Uses CREATE_NEW_PROCESS_GROUP so child processes are grouped, and overrides
// Cancel to kill the entire process tree via taskkill /T /F.
func configureProcGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	cmd.Cancel = func() error {
		return killProcessTree(cmd.Process.Pid)
	}
}

// killProcessTree terminates a process and all its descendants on Windows.
func killProcessTree(pid int) error {
	kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid))
	if out, err := kill.CombinedOutput(); err != nil {
		return fmt.Errorf("taskkill PID %d: %w (%s)", pid, err, out)
	}
	return nil
}

// killAllChildren terminates all tracked child process trees on Windows.
func killAllChildren() {
	childGroups.mu.Lock()
	defer childGroups.mu.Unlock()
	for pid := range childGroups.pids {
		_ = killProcessTree(pid)
	}
	childGroups.pids = nil
}
