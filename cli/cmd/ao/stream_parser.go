package main

import (
	"bufio"
	"io"
	"time"
)

// PhaseProgress tracks cumulative progress while parsing a stream of
// Claude Code JSON events.
type PhaseProgress struct {
	Name         string
	SessionID    string
	Model        string
	LastToolCall string
	ToolCount    int
	TurnCount    int
	Tokens       int
	CostUSD      float64
	Elapsed      time.Duration
}

// ParseStreamEvents reads newline-delimited JSON events from r, updating
// a PhaseProgress as it goes.  If onUpdate is non-nil it is called after
// every successfully parsed event.  The final PhaseProgress is returned
// along with the first non-EOF read error (malformed JSON lines are
// silently skipped so that a partial stream still yields useful data).
func ParseStreamEvents(r io.Reader, onUpdate func(PhaseProgress)) (PhaseProgress, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	var p PhaseProgress

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		ev, err := ParseStreamEvent(line)
		if err != nil {
			// Skip malformed lines.
			continue
		}

		switch ev.Type {
		case EventTypeInit:
			p.SessionID = ev.SessionID
			p.Model = ev.Model

		case EventTypeAssistant:
			if ev.ToolName != "" {
				p.ToolCount++
				p.LastToolCall = ev.ToolName
			}

		case EventTypeResult:
			p.CostUSD = ev.CostUSD
			p.TurnCount = ev.NumTurns
			if ev.DurationMS > 0 {
				p.Elapsed = time.Duration(ev.DurationMS * float64(time.Millisecond))
			}
		}

		if onUpdate != nil {
			onUpdate(p)
		}
	}

	return p, scanner.Err()
}
