package goals

import (
	"sync"
	"testing"
)

func TestChildGroupsInitialized(t *testing.T) {
	// Bug #7: childGroups.pids should be non-nil at package init time.
	// Before the fix, it starts nil and relies on lazy init in trackChild.
	if childGroups.pids == nil {
		t.Fatal("childGroups.pids is nil at package init; expected eager initialization")
	}
}

func TestTrackChild_ConcurrentAccess(t *testing.T) {
	// Bug #7: Verify trackChild/untrackChild are safe under concurrent access.
	// Must pass with -race flag.
	const goroutines = 10

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		pid := 10000 + i
		go func(p int) {
			defer wg.Done()
			trackChild(p)
		}(pid)
		go func(p int) {
			defer wg.Done()
			untrackChild(p)
		}(pid)
	}
	wg.Wait()

	// Clean up: remove any leftover tracked pids from this test.
	childGroups.mu.Lock()
	for pid := 10000; pid < 10000+goroutines; pid++ {
		delete(childGroups.pids, pid)
	}
	childGroups.mu.Unlock()
}
