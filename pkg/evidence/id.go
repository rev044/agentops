package evidence

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// generateID produces a short random hex identifier for evidence entries.
// Format: "ev-<12 random bytes as hex>".
// Using 12 bytes (96 bits) for better collision resistance, especially if
// evidence entries grow in volume over time.
func generateID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		// Fallback — should never happen in practice.
		return fmt.Sprintf("ev-fallback-%d", mustMonotonicNano())
	}
	return "ev-" + hex.EncodeToString(b)
}

// mustMonotonicNano is a last-resort counter using the runtime clock.
func mustMonotonicNano() int64 {
	return time.Now().UnixNano()
}
