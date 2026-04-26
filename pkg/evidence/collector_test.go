package evidence

import (
	"testing"
)

func TestAdd(t *testing.T) {
	c := NewCollector()
	e := c.Add("observation", map[string]any{"detail": "test"}, "unit")
	if e == nil {
		t.Fatal("expected non-nil entry")
	}
	if e.Kind != "observation" {
		t.Errorf("kind: got %q, want %q", e.Kind, "observation")
	}
	if e.Closed {
		t.Error("new entry should not be closed")
	}
	if len(e.ID) == 0 {
		t.Error("ID must not be empty")
	}
}

func TestClose(t *testing.T) {
	c := NewCollector()
	e := c.Add("finding", nil)
	ok := c.Close(e.ID)
	if !ok {
		t.Fatal("Close returned false for valid ID")
	}
	if !e.Closed {
		t.Error("entry should be marked closed")
	}
	if c.Close("nonexistent") {
		t.Error("Close should return false for unknown ID")
	}
}

func TestOpenFiltering(t *testing.T) {
	c := NewCollector()
	a := c.Add("a", nil)
	b := c.Add("b", nil)
	c.Close(a.ID)

	open := c.Open()
	if len(open) != 1 {
		t.Fatalf("expected 1 open entry, got %d", len(open))
	}
	if open[0].ID != b.ID {
		t.Errorf("wrong open entry: got %q, want %q", open[0].ID, b.ID)
	}
}

func TestAllReturnsSnapshot(t *testing.T) {
	c := NewCollector()
	c.Add("x", nil)
	c.Add("y", nil)
	all := c.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

// TestConcurrentAdd verifies that the collector is safe for concurrent use.
// Bumped goroutine count from 100 to 200 to increase confidence in race safety.
func TestConcurrentAdd(t *testing.T) {
	c := NewCollector()
	const n = 200
	done := make(chan struct{})
	for i := 0; i < n; i++ {
		go func() {
			c.Add("concurrent", nil)
			done <- struct{}{}
		}()
	}
	for i := 0; i < n; i++ {
		<-done
	}
	if len(c.All()) != n {
		t.Errorf("expected %d entries after concurrent adds, got %d", n, len(c.All()))
	}
}
