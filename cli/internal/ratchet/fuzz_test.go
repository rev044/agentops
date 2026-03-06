package ratchet

import (
	"bufio"
	"strings"
	"testing"
)

// FuzzParseChainLines fuzzes the JSONL chain parser with arbitrary input.
// parseChainLines expects line 1 as chain metadata and subsequent lines as entries.
func FuzzParseChainLines(f *testing.F) {
	// Seed corpus with realistic chain formats (JSONL: line 1 = metadata, rest = entries)
	f.Add(`{"id":"chain-001","started":"2026-01-01T00:00:00Z","chain":[]}` + "\n" +
		`{"step":"research","timestamp":"2026-01-01T00:01:00Z","output":"plan.md","locked":true}`)
	f.Add(`{"id":"chain-002","started":"2026-01-01T00:00:00Z"}`)
	f.Add(``)
	f.Add(`not json at all`)
	f.Add(`{"id":""}` + "\n" + `{"step":"bad","timestamp":"not-a-time"}`)
	f.Add(`{"id":"chain-003"}` + "\n" + `malformed entry` + "\n" + `{"step":"plan","timestamp":"2026-01-01T00:00:00Z","output":"out.md","locked":false}`)
	f.Add(`{}`)
	f.Add("\n\n\n")
	f.Add(`{"id":"chain-004","started":"2026-01-01T00:00:00Z","epic_id":"epic-1"}` + "\n" +
		`{"step":"research","timestamp":"2026-01-01T00:00:00Z","input":"in.md","output":"out.md","locked":true}` + "\n" +
		`{"step":"plan","timestamp":"2026-01-01T00:01:00Z","input":"out.md","output":"plan.md","locked":false}`)

	f.Fuzz(func(t *testing.T, data string) {
		scanner := bufio.NewScanner(strings.NewReader(data))
		chain := &Chain{}
		// Must never panic — errors are acceptable
		_ = parseChainLines(scanner, chain)
	})
}
