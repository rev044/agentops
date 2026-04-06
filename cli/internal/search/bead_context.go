package search

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

	// BeadContextCacheFile is the path (relative to .agents/ao/) for cached bead metadata.
	BeadContextCacheFile = "bead-context.json"
)

// BeadContext holds resolved metadata about the bead being worked on.
type BeadContext struct {
	ID       string   `json:"id"`
	Title    string   `json:"title,omitempty"`
	Labels   []string `json:"labels,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
	Phase    string   `json:"phase,omitempty"`
}

// ResolveBeadContext builds a BeadContext from available sources.
// Priority: env vars > cache file > bead ID only (minimal context).
// Never shells out to bd — all sources are pre-resolved by Gas Town hooks.
func ResolveBeadContext(beadID, cwd string) *BeadContext {
	if beadID == "" {
		return nil
	}

	ctx := &BeadContext{ID: beadID}

	// Try env vars first (set by Gas Town's gt prime --hook)
	if title := os.Getenv("HOOK_BEAD_TITLE"); title != "" {
		ctx.Title = title
	}
	if labels := os.Getenv("HOOK_BEAD_LABELS"); labels != "" {
		ctx.Labels = SplitLabels(labels)
	}
	if phase := os.Getenv("HOOK_BEAD_PHASE"); phase != "" {
		ctx.Phase = phase
	}

	// If env vars gave us metadata, build keywords and return
	if ctx.Title != "" || len(ctx.Labels) > 0 {
		ctx.Keywords = BuildKeywords(ctx)
		return ctx
	}

	// Try cache file as fallback
	if cached := ReadBeadCache(cwd, beadID); cached != nil {
		return cached
	}

	// Minimal context: just the bead ID for direct matching
	ctx.Keywords = BuildKeywords(ctx)
	return ctx
}

// ReadBeadCache reads bead metadata from .agents/ao/bead-context.json.
func ReadBeadCache(cwd, beadID string) *BeadContext {
	cachePath := filepath.Join(cwd, ".agents", "ao", BeadContextCacheFile)
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

	ctx.Keywords = BuildKeywords(&ctx)
	return &ctx
}

// SplitLabels splits a comma-separated label string into a slice.
func SplitLabels(s string) []string {
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

// BuildKeywords extracts searchable keywords from bead context.
func BuildKeywords(ctx *BeadContext) []string {
	seen := make(map[string]bool)
	var keywords []string

	add := func(word string) {
		w := strings.ToLower(strings.TrimSpace(word))
		if w != "" && len(w) >= 2 && !seen[w] {
			seen[w] = true
			keywords = append(keywords, w)
		}
	}

	for _, word := range strings.Fields(ctx.Title) {
		add(word)
	}

	for _, label := range ctx.Labels {
		add(label)
	}

	return keywords
}
