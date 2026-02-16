package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// WriteLiveStatus writes a Markdown status table to path atomically.
// It writes to path+".tmp" first, then renames to path to avoid partial reads.
// currentPhase is the 0-based index into allPhases that is currently running.
// Phases before currentPhase are marked "done", after are "pending".
func WriteLiveStatus(path string, allPhases []PhaseProgress, currentPhase int) error {
	var b strings.Builder

	b.WriteString("# Live Status\n\n")
	b.WriteString("| Phase | Status | Elapsed | Tools | Tokens | Cost |\n")
	b.WriteString("|-------|--------|---------|-------|--------|------|\n")

	for i, p := range allPhases {
		var status string
		switch {
		case i < currentPhase:
			status = "done"
		case i == currentPhase:
			status = "running"
		default:
			status = "pending"
		}

		elapsed := p.Elapsed.Truncate(time.Second).String()
		cost := fmt.Sprintf("$%.4f", p.CostUSD)

		fmt.Fprintf(&b, "| %s | %s | %s | %d | %d | %s |\n",
			p.Name, status, elapsed, p.ToolCount, p.Tokens, cost)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
