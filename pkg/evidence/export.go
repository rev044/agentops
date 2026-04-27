package evidence

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ExportJSON writes all evidence entries held by c to w as a JSON array.
// The output is compatible with the closure format used in
// .agents/council/evidence-only-closures/.
func ExportJSON(c *Collector, w io.Writer) error {
	type exportEnvelope struct {
		ExportedAt  string   `json:"exported_at"`
		Entries     []*Entry `json:"entries"`
		OpenCount   int      `json:"open_count"`
		ClosedCount int      `json:"closed_count"`
		// TotalCount is included for convenience so consumers don't have to sum
		// open_count and closed_count themselves.
		TotalCount int `json:"total_count"`
	}

	all := c.All()
	open := c.Open()

	env := exportEnvelope{
		ExportedAt:  time.Now().UTC().Format(time.RFC3339),
		Entries:     all,
		OpenCount:   len(open),
		ClosedCount: len(all) - len(open),
		TotalCount:  len(all),
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(env); err != nil {
		return fmt.Errorf("evidence: export JSON: %w", err)
	}
	return nil
}

// Summary returns a human-readable one-line summary of the collector state.
func Summary(c *Collector) string {
	all := c.All()
	open := c.Open()
	return fmt.Sprintf(
		"evidence: total=%d open=%d closed=%d",
		len(all), len(open), len(all)-len(open),
	)
}
