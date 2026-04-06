//go:build !windows

package goals

import (
	"os/exec"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestRunGoals_GoroutineLeak(t *testing.T) {
	// The bug: runGoals spawns a goroutine listening on sigCh but never
	// closes sigCh after signal.Stop, so the goroutine leaks on every call.
	// Before the fix, each Measure() call leaks 1 goroutine.
	// Snapshot current goroutines so we only detect leaks from THIS test,
	// not from parallel tests that may have in-flight subprocess cleanup.
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	gf := &GoalFile{
		Version: 2,
		Goals: []Goal{
			{ID: "leak-test", Check: "true", Weight: 1, Type: GoalTypeHealth},
		},
	}
	for i := 0; i < 5; i++ {
		Measure(gf, 5*time.Second)
	}
}

func TestConfigureProcGroup_NilProcess(t *testing.T) {
	// The bug: configureProcGroup sets cmd.Cancel to a closure that
	// dereferences cmd.Process.Pid. If the command hasn't started,
	// cmd.Process is nil and Cancel panics.
	cmd := exec.Command("true")
	configureProcGroup(cmd)

	// cmd.Process is nil because we haven't called cmd.Start().
	if cmd.Process != nil {
		t.Fatal("expected cmd.Process to be nil before Start()")
	}

	// The Cancel function should handle nil Process gracefully.
	// Before the fix, this panics with nil pointer dereference.
	if cmd.Cancel != nil {
		err := cmd.Cancel()
		if err != nil {
			// A non-nil error is acceptable (e.g. "process not started").
			// A panic is not.
			t.Logf("Cancel returned error (OK): %v", err)
		}
	}
}
