package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	// BeadScoreMatchDirect is the multiplicative boost for direct bead ID match.
	BeadScoreMatchDirect = 1.5

	// BeadScoreMatchLabel is the multiplicative boost per matching label.
	BeadScoreMatchLabel = 1.2

	// BeadScoreMatchCategory is the multiplicative boost for category match.
	BeadScoreMatchCategory = 1.1

	// beadContextCacheFile is the path (relative to .agents/ao/) for cached bead metadata.
	beadContextCacheFile = "bead-context.json"
)

// BeadContext holds resolved metadata about the bead being worked on.
// Populated from env vars or a cache file — never calls bd CLI directly
// to avoid blowing the 5-second hook timeout.
type BeadContext struct {
	ID       string   `json:"id"`
	Title    string   `json:"title,omitempty"`
	Labels   []string `json:"labels,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
	Phase    string   `json:"phase,omitempty"`
}

// resolveBeadContext builds a BeadContext from available sources.
// Priority: env vars > cache file > bead ID only (minimal context).
// Never shells out to bd — all sources are pre-resolved by Gas Town hooks.
func resolveBeadContext(beadID, cwd string) *BeadContext {
	if beadID == "" {
		return nil
	}

	ctx := &BeadContext{ID: beadID}

	// Try env vars first (set by Gas Town's gt prime --hook)
	if title := os.Getenv("HOOK_BEAD_TITLE"); title != "" {
		ctx.Title = title
	}
	if labels := os.Getenv("HOOK_BEAD_LABELS"); labels != "" {
		ctx.Labels = splitLabels(labels)
	}
	if phase := os.Getenv("HOOK_BEAD_PHASE"); phase != "" {
		ctx.Phase = phase
	}

	// If env vars gave us metadata, build keywords and return
	if ctx.Title != "" || len(ctx.Labels) > 0 {
		ctx.Keywords = buildKeywords(ctx)
		return ctx
	}

	// Try cache file as fallback
	if cached := readBeadCache(cwd, beadID); cached != nil {
		return cached
	}

	// Minimal context: just the bead ID for direct matching
	ctx.Keywords = buildKeywords(ctx)
	return ctx
}

// readBeadCache reads bead metadata from .agents/ao/bead-context.json.
// Written by gt prime --hook or a pre-inject step.
func readBeadCache(cwd, beadID string) *BeadContext {
	cachePath := filepath.Join(cwd, ".agents", "ao", beadContextCacheFile)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil
	}

	var ctx BeadContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil
	}

	// Verify cache is for the right bead
	if ctx.ID != beadID {
		return nil
	}

	ctx.Keywords = buildKeywords(&ctx)
	return &ctx
}

// applyBeadBoost adjusts a learning's composite score based on bead context.
// Uses multiplicative boosts to preserve z-norm distribution.
func applyBeadBoost(l *learning, bead *BeadContext) {
	if bead == nil {
		return
	}

	// Direct bead match: learning was produced under this exact bead
	if l.SourceBead != "" && l.SourceBead == bead.ID {
		l.CompositeScore *= BeadScoreMatchDirect
		return // Direct match is the strongest signal, skip label matching
	}

	// Label and keyword boosts intentionally stack (max 1.2 * 1.1 = 1.32x).
	// This rewards learnings relevant on multiple dimensions without reaching
	// the direct-match threshold (1.5x), preserving the hierarchy:
	//   direct > label+keyword > label-only > keyword-only > unmatched

	// Label matching: learning's category matches bead labels
	if len(bead.Labels) > 0 {
		titleLower := strings.ToLower(l.Title + " " + l.Summary)
		for _, label := range bead.Labels {
			if strings.Contains(titleLower, strings.ToLower(label)) {
				l.CompositeScore *= BeadScoreMatchLabel
				break // One label match is enough
			}
		}
	}

	// Keyword matching from bead title
	if len(bead.Keywords) > 0 {
		titleLower := strings.ToLower(l.Title + " " + l.Summary)
		for _, kw := range bead.Keywords {
			if strings.Contains(titleLower, kw) {
				l.CompositeScore *= BeadScoreMatchCategory
				break
			}
		}
	}
}

// splitLabels splits a comma-separated label string into a slice.
func splitLabels(s string) []string {
	parts := strings.Split(s, ",")
	labels := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			labels = append(labels, p)
		}
	}
	return labels
}

// buildKeywords extracts searchable keywords from bead context.
func buildKeywords(ctx *BeadContext) []string {
	seen := make(map[string]bool)
	var keywords []string

	add := func(word string) {
		w := strings.ToLower(strings.TrimSpace(word))
		if w != "" && len(w) >= 2 && !seen[w] {
			seen[w] = true
			keywords = append(keywords, w)
		}
	}

	// Keywords from title
	for _, word := range strings.Fields(ctx.Title) {
		add(word)
	}

	// Labels as keywords
	for _, label := range ctx.Labels {
		add(label)
	}

	return keywords
}
