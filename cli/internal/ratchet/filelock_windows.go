//go:build windows

package ratchet

import "os"

// Windows builds do not support syscall.Flock. For now, ratchet chain writes
// use process-local append semantics on Windows runners.
func lockFile(_ *os.File) error {
	return nil
}

func unlockFile(_ *os.File) error {
	return nil
}

