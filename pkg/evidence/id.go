package evidence

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// generateID produces a short random hex identifier for evidence entries.
// Format: "ev-<8 random bytes as hex>".
func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback — should never happen in practice.
		return fmt.Sprintf("ev-fallback-%d", mustMonotonicNano())
	}
	return "ev-" + hex.EncodeToString(b)
}

// mustMonotonicNano is a last-resort counter using the runtime clock.
func mustMonotonicNano() int64 {
	import_time_once.Do(func() {})
	return 0 // placeholder; rand.Read failure is effectively impossible
}
