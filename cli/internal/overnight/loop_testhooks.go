package overnight

// Test-only hooks that production code must NEVER set. These are package-
// private and live in a dedicated file so a reviewer grepping RunLoopOptions
// never sees a field they'd have to mentally filter as "production? test?".
//
// CONCURRENCY CONTRACT: these globals are NOT safe for parallel tests.
// Any test that calls SetFaultInjectionAfterIter MUST NOT call t.Parallel(),
// and MUST call t.Cleanup(func() { SetFaultInjectionAfterIter(0) }) to
// restore the default before the next test runs. Tests that forget the
// cleanup will bleed state into sibling tests. (Pre-mortem judge-3 B4 catch.)

import "sync/atomic"

// faultInjectionAfterIter, when non-zero, causes RunLoop to panic AFTER
// persisting iter-<N>.json but BEFORE updating prevSnapshot. This lets
// TestRunLoop_CrashAtIter2_ResumeRehydrates exercise the resume path
// deterministically. Stored as atomic.Int32 for race-detector cleanliness
// when multiple tests run serially in the same package.
var faultInjectionAfterIter atomic.Int32

// SetFaultInjectionAfterIter is the test-only setter. Calling it from
// non-test code is a bug; nothing in the production build path should
// reference this symbol. MUST be paired with t.Cleanup to restore 0.
func SetFaultInjectionAfterIter(n int) {
	faultInjectionAfterIter.Store(int32(n))
}

// getFaultInjectionAfterIter is the package-internal reader used by RunLoop.
func getFaultInjectionAfterIter() int {
	return int(faultInjectionAfterIter.Load())
}
