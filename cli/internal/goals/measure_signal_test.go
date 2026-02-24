//go:build !windows

package goals

import (
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestRunGoals_GoroutineLeak(t *testing.T) {
	// Call Measure multiple times and verify goroutine count doesn't grow.
	// The bug: runGoals spawns a goroutine listening on sigCh but never
	// closes sigCh after signal.Stop, so the goroutine leaks on every call.

	// Warm up — let runtime settle.
	gf := &GoalFile{
		Version: 2,
		Goals: []Goal{
			{ID: "leak-test", Check: "true", Weight: 1, Type: GoalTypeHealth},
		},
	}
	Measure(gf, 5*time.Second)
	runtime.GC()
	time.Sleep(50 * time.Millisecond)

	baseline := runtime.NumGoroutine()

	const iterations = 5
	for i := 0; i < iterations; i++ {
		Measure(gf, 5*time.Second)
	}

	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()

	// Allow 1 goroutine of slack for runtime jitter, but not 5.
	// Before the fix, each Measure() call leaks 1 goroutine.
	if after > baseline+2 {
		t.Errorf("goroutine leak: baseline=%d, after %d iterations=%d (delta=%d, want <=2)",
			baseline, iterations, after, after-baseline)
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
