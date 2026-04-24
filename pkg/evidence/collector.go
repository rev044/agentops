package evidence

import (
	"sync"
	"time"
)

// Entry represents a single piece of evidence collected during an agent run.
type Entry struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Kind      string            `json:"kind"`
	Payload   map[string]any    `json:"payload"`
	Tags      []string          `json:"tags"`
	Closed    bool              `json:"closed"`
}

// Collector accumulates evidence entries in memory and supports closure marking.
type Collector struct {
	mu      sync.RWMutex
	entries []*Entry
}

// NewCollector returns an initialised Collector.
func NewCollector() *Collector {
	return &Collector{}
}

// Add inserts a new evidence entry and returns its index.
func (c *Collector) Add(kind string, payload map[string]any, tags ...string) *Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := &Entry{
		ID:        generateID(),
		Timestamp: time.Now().UTC(),
		Kind:      kind,
		Payload:   payload,
		Tags:      tags,
	}
	c.entries = append(c.entries, e)
	return e
}

// Close marks an entry as closed (evidence resolved / actioned).
func (c *Collector) Close(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, e := range c.entries {
		if e.ID == id {
			e.Closed = true
			return true
		}
	}
	return false
}

// All returns a snapshot of all entries.
func (c *Collector) All() []*Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*Entry, len(c.entries))
	copy(out, c.entries)
	return out
}

// Open returns only entries that have not been closed.
func (c *Collector) Open() []*Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []*Entry
	for _, e := range c.entries {
		if !e.Closed {
			out = append(out, e)
		}
	}
	return out
}
