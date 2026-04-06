// Package rpi parallel helpers.
//
// NOTE: cmd/ao/rpi_parallel.go is deprecated in favor of the gc sling + convoy
// path. These helpers are extracted for reuse / testability only; no new
// capabilities should be added here.
package rpi

import (
	"path/filepath"
	"strings"
)

// GoalSlug creates a short filesystem-safe name from a goal string by taking
// the first 3 significant words (alphanumeric only, stopwords skipped).
// Returns empty string if no significant words remain.
func GoalSlug(goal string) string {
	words := strings.Fields(strings.ToLower(goal))
	skip := map[string]bool{
		"add": true, "the": true, "a": true, "an": true, "to": true,
		"for": true, "and": true, "with": true, "in": true, "on": true,
	}
	var sig []string
	for _, w := range words {
		clean := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, w)
		if clean == "" || skip[clean] {
			continue
		}
		sig = append(sig, clean)
		if len(sig) >= 3 {
			break
		}
	}
	if len(sig) == 0 {
		return ""
	}
	return strings.Join(sig, "-")
}

// ShellQuote wraps a string in single quotes, escaping embedded single quotes
// so the result is safe to pass to a POSIX shell.
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// ParallelWorktreeRoot returns the root directory where parallel epic
// worktrees should be created for a given base cwd and runtime command.
func ParallelWorktreeRoot(baseCwd, runtimeCmd string) string {
	runtimeName := filepath.Base(runtimeCmd)
	if runtimeName == "." || runtimeName == string(filepath.Separator) || runtimeName == "" {
		runtimeName = "claude"
	}
	return filepath.Join(baseCwd, "."+runtimeName, "worktrees")
}

// ResolveMergeOrderByNames returns indices into epicNames in the order
// specified by mergeOrder (comma-separated names). Unknown names are skipped.
func ResolveMergeOrderByNames(epicNames []string, mergeOrder string) []int {
	names := strings.Split(mergeOrder, ",")
	indices := make([]int, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		for i, n := range epicNames {
			if n == name {
				indices = append(indices, i)
				break
			}
		}
	}
	return indices
}

// ResolveMergeOrderByField sorts indices by the mergeOrder int field.
// Uses insertion sort (stable, small N).
func ResolveMergeOrderByField(mergeOrders []int) []int {
	type indexed struct {
		idx   int
		order int
	}
	items := make([]indexed, len(mergeOrders))
	for i, o := range mergeOrders {
		items[i] = indexed{idx: i, order: o}
	}
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].order < items[j-1].order; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
	indices := make([]int, len(items))
	for i, item := range items {
		indices[i] = item.idx
	}
	return indices
}
