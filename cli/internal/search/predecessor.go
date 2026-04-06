package search

import (
	"path/filepath"
	"strings"
)

// ExtractSections splits markdown content into header → body pairs.
func ExtractSections(content string) map[string]string {
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

// FindSection looks for the first matching section header (case-insensitive).
func FindSection(sections map[string]string, candidates ...string) string {
	for _, c := range candidates {
		for header, body := range sections {
			if strings.Contains(header, c) && body != "" {
				return TruncateText(body, 300)
			}
		}
	}
	return ""
}

// ExtractFirstParagraphs returns the first N non-empty paragraphs from content.
func ExtractFirstParagraphs(content string, maxParas int) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	paraCount := 0
	inPara := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			continue
		}
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

// DeriveTopicFromPath extracts a human-readable topic from a handoff filename.
func DeriveTopicFromPath(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	for _, prefix := range []string{"stop-", "auto-"} {
		base = strings.TrimPrefix(base, prefix)
	}
	if len(base) > 16 && base[8] == 'T' {
		if idx := strings.Index(base, "-"); idx > 0 && idx < 20 {
			base = base[idx+1:]
		}
	}
	base = strings.ReplaceAll(base, "-", " ")
	return base
}

// TruncatePredecessor ensures the total predecessor context fits a character budget.
func TruncatePredecessor(ctx *PredecessorContext, maxChars int) {
	total := len(ctx.WorkingOn) + len(ctx.Progress) + len(ctx.Blocker) + len(ctx.NextStep) + len(ctx.RawSummary)
	if total <= maxChars {
		return
	}

	ctx.WorkingOn = TruncateText(ctx.WorkingOn, 100)
	ctx.Progress = TruncateText(ctx.Progress, 250)
	ctx.Blocker = TruncateText(ctx.Blocker, 200)
	ctx.NextStep = TruncateText(ctx.NextStep, 150)
	ctx.RawSummary = TruncateText(ctx.RawSummary, 300)
}
