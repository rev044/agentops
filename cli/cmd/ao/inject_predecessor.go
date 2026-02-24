package main

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	// maxPredecessorTokens is the token budget reserved for predecessor context.
	maxPredecessorTokens = 200
	// maxPredecessorChars is the character budget (tokens * 4).
	maxPredecessorChars = maxPredecessorTokens * InjectCharsPerToken
)

// predecessorContext holds structured context from a predecessor agent's handoff.
type predecessorContext struct {
	WorkingOn  string `json:"working_on,omitempty"`
	Progress   string `json:"progress,omitempty"`
	Blocker    string `json:"blocker,omitempty"`
	NextStep   string `json:"next_step,omitempty"`
	SessionAge string `json:"session_age,omitempty"`
	RawSummary string `json:"raw_summary,omitempty"` // Fallback when no structured headers found
}

// parsePredecessorFile reads a handoff file and extracts structured predecessor context.
// Supports three handoff formats:
//   - Explicit handoff (/handoff skill): has ## headers like Accomplishments, Pause Point, Blockers
//   - Auto-handoff (stop hook): prose summary, may have ## headers
//   - Pre-compact snapshot: has Branch, Status, Teams sections
//
// Falls back to first-N-paragraphs when no recognized headers are found.
func parsePredecessorFile(path string) *predecessorContext {
	data, err := os.ReadFile(path)
	if err != nil {
		VerbosePrintf("Warning: failed to read predecessor file: %v\n", err)
		return nil
	}

	content := string(data)
	if strings.TrimSpace(content) == "" {
		return nil
	}

	ctx := &predecessorContext{}

	// Extract structured sections by header
	sections := extractSections(content)

	// Map sections to context fields
	ctx.WorkingOn = findSection(sections, "working on", "topic", "task", "hook", "bead")
	ctx.Progress = findSection(sections, "accomplishments", "progress", "completed", "done", "what happened")
	ctx.Blocker = findSection(sections, "blockers", "blocker", "blocked", "issues", "problems")
	ctx.NextStep = findSection(sections, "next step", "next steps", "continuation", "pause point", "resume")

	// If no structured headers found, use first paragraphs as raw summary
	if ctx.Progress == "" && ctx.NextStep == "" && ctx.Blocker == "" {
		ctx.RawSummary = extractFirstParagraphs(content, 3)
	}

	// Derive working-on from filename if not in content
	if ctx.WorkingOn == "" {
		ctx.WorkingOn = deriveTopicFromPath(path)
	}

	// Derive session age from file modification time
	if info, err := os.Stat(path); err == nil {
		ctx.SessionAge = formatAge(info.ModTime())
	}

	// Truncate all fields to fit token budget
	truncatePredecessor(ctx)

	return ctx
}

// extractSections splits markdown content into header → body pairs.
func extractSections(content string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(content, "\n")

	var currentHeader string
	var currentBody strings.Builder

	flush := func() {
		if currentHeader != "" {
			sections[strings.ToLower(currentHeader)] = strings.TrimSpace(currentBody.String())
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			flush()
			currentHeader = strings.TrimPrefix(trimmed, "## ")
			currentBody.Reset()
		} else if strings.HasPrefix(trimmed, "# ") {
			// Top-level header — use as working-on hint
			flush()
			currentHeader = strings.TrimPrefix(trimmed, "# ")
			currentBody.Reset()
		} else if currentHeader != "" {
			currentBody.WriteString(line)
			currentBody.WriteString("\n")
		}
	}
	flush()

	return sections
}

// findSection looks for the first matching section header (case-insensitive).
func findSection(sections map[string]string, candidates ...string) string {
	for _, c := range candidates {
		for header, body := range sections {
			if strings.Contains(header, c) && body != "" {
				return truncateText(body, 300)
			}
		}
	}
	return ""
}

// extractFirstParagraphs returns the first N non-empty paragraphs from content.
func extractFirstParagraphs(content string, maxParas int) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	paraCount := 0
	inPara := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip front matter
		if trimmed == "---" {
			continue
		}
		// Skip headers
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		if trimmed == "" {
			if inPara {
				paraCount++
				inPara = false
				if paraCount >= maxParas {
					break
				}
				result.WriteString("\n")
			}
			continue
		}

		inPara = true
		result.WriteString(trimmed)
		result.WriteString(" ")
	}

	return strings.TrimSpace(result.String())
}

// deriveTopicFromPath extracts a human-readable topic from a handoff filename.
func deriveTopicFromPath(path string) string {
	base := filepath.Base(path)
	// Strip extensions
	base = strings.TrimSuffix(base, filepath.Ext(base))
	// Strip common prefixes: stop-, auto-, YYYYMMDDTHHMMSSZ-
	for _, prefix := range []string{"stop-", "auto-"} {
		base = strings.TrimPrefix(base, prefix)
	}
	// Strip timestamp prefix (20260224T103000Z-)
	if len(base) > 16 && base[8] == 'T' {
		if idx := strings.Index(base, "-"); idx > 0 && idx < 20 {
			base = base[idx+1:]
		}
	}
	// Convert hyphens to spaces
	base = strings.ReplaceAll(base, "-", " ")
	return base
}

// truncatePredecessor ensures the total predecessor context fits the token budget.
func truncatePredecessor(ctx *predecessorContext) {
	total := len(ctx.WorkingOn) + len(ctx.Progress) + len(ctx.Blocker) + len(ctx.NextStep) + len(ctx.RawSummary)
	if total <= maxPredecessorChars {
		return
	}

	// Truncate longest fields first
	ctx.WorkingOn = truncateText(ctx.WorkingOn, 100)
	ctx.Progress = truncateText(ctx.Progress, 250)
	ctx.Blocker = truncateText(ctx.Blocker, 200)
	ctx.NextStep = truncateText(ctx.NextStep, 150)
	ctx.RawSummary = truncateText(ctx.RawSummary, 300)
}
