package pool

import (
	"cmp"
	"regexp"
	"strings"
)

// LearningBlock is a parsed markdown learning block ready for scoring/ingest.
type LearningBlock struct {
	Title      string
	ID         string
	Category   string
	Confidence string
	Body       string
}

var (
	reLearningHeader = regexp.MustCompile(`(?m)^# Learning:\s*(.+)\s*$`)
	reIDLine         = regexp.MustCompile(`(?m)^\*\*ID:?\*\*:?\s*(.+)\s*$`)
	reCategoryLine   = regexp.MustCompile(`(?m)^\*\*Category:?\*\*:?\s*(.+)\s*$`)
	reConfidenceLine = regexp.MustCompile(`(?m)^\*\*Confidence:?\*\*:?\s*(.+)\s*$`)
	reFrontmatter    = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
)

// ParseLearningBlocks extracts one or more LearningBlock entries from markdown.
// Falls back to a single legacy frontmatter-based block when no "# Learning:"
// header is present.
func ParseLearningBlocks(md string) []LearningBlock {
	locs := reLearningHeader.FindAllStringSubmatchIndex(md, -1)
	if len(locs) == 0 {
		if legacy, ok := ParseLegacyFrontmatterLearning(md); ok {
			return []LearningBlock{legacy}
		}
		return nil
	}

	var blocks []LearningBlock
	for i, loc := range locs {
		start := loc[0]
		end := len(md)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		title := strings.TrimSpace(md[loc[2]:loc[3]])
		body := strings.TrimSpace(md[start:end])

		b := LearningBlock{Title: title, Body: body}
		if m := reIDLine.FindStringSubmatch(body); len(m) == 2 {
			b.ID = strings.TrimSpace(m[1])
		}
		if m := reCategoryLine.FindStringSubmatch(body); len(m) == 2 {
			b.Category = strings.TrimSpace(m[1])
		}
		if m := reConfidenceLine.FindStringSubmatch(body); len(m) == 2 {
			b.Confidence = strings.TrimSpace(m[1])
		}
		blocks = append(blocks, b)
	}
	return blocks
}

// ParseYAMLFrontmatter parses a raw YAML frontmatter block into a string map.
func ParseYAMLFrontmatter(raw string) map[string]string {
	fm := make(map[string]string)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		fm[key] = strings.Trim(val, `"'`)
	}
	return fm
}

// ExtractFirstHeadingText returns the first non-empty text line with leading
// '#' characters trimmed.
func ExtractFirstHeadingText(body string) string {
	for _, line := range strings.Split(body, "\n") {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		l = strings.TrimSpace(strings.TrimPrefix(l, "#"))
		if l != "" {
			return l
		}
	}
	return ""
}

// ParseLegacyFrontmatterLearning parses a single legacy /learn-style markdown
// file (YAML frontmatter + body) into a LearningBlock.
func ParseLegacyFrontmatterLearning(md string) (LearningBlock, bool) {
	fmMatch := reFrontmatter.FindStringSubmatchIndex(md)
	if len(fmMatch) < 4 {
		return LearningBlock{}, false
	}
	fmRaw := md[fmMatch[2]:fmMatch[3]]
	body := strings.TrimSpace(md[fmMatch[1]:])
	if body == "" {
		return LearningBlock{}, false
	}
	frontmatter := ParseYAMLFrontmatter(fmRaw)
	category := strings.TrimSpace(frontmatter["type"])
	if category == "" {
		return LearningBlock{}, false
	}
	title := ExtractFirstHeadingText(body)
	if title == "" {
		return LearningBlock{}, false
	}
	return LearningBlock{
		Title:      title,
		ID:         cmp.Or(strings.TrimSpace(frontmatter["id"]), "legacy"),
		Category:   category,
		Confidence: cmp.Or(strings.TrimSpace(frontmatter["confidence"]), "medium"),
		Body:       body,
	}, true
}

// IsSlugAlphanumeric reports whether the rune is kept as-is in a slug.
func IsSlugAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

// Slugify converts a string to a dash-separated lowercase slug.
func Slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		if IsSlugAlphanumeric(r) {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	return cmp.Or(strings.Trim(b.String(), "-"), "cand")
}

// ConfidenceToScore maps high/medium/low confidence strings to a numeric score.
func ConfidenceToScore(s string) float64 {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "high":
		return 0.9
	case "medium":
		return 0.7
	case "low":
		return 0.5
	default:
		return 0.6
	}
}

// ComputeSpecificityScore scores how specific a learning body is.
func ComputeSpecificityScore(body, lower string) float64 {
	spec := 0.4
	if strings.Contains(body, "`") || strings.Contains(body, "```") {
		spec += 0.2
	}
	if regexp.MustCompile(`\d`).MatchString(body) {
		spec += 0.2
	}
	if regexp.MustCompile(`\b[a-zA-Z0-9_./-]+\.(go|ts|js|py|sh|yaml|yml|json|md)\b`).MatchString(body) {
		spec += 0.2
	}
	if strings.Contains(lower, "line ") {
		spec += 0.1
	}
	if spec > 1.0 {
		spec = 1.0
	}
	return spec
}

// ComputeActionabilityScore scores how actionable a learning body is.
func ComputeActionabilityScore(body string) float64 {
	act := 0.4
	if regexp.MustCompile(`(?m)^\s*[-*]\s+`).MatchString(body) {
		act += 0.2
	}
	if regexp.MustCompile(`(?i)\b(run|add|remove|use|ensure|check|grep|rg|fix|avoid|prefer|must|should)\b`).MatchString(body) {
		act += 0.2
	}
	if strings.Contains(body, "```") {
		act += 0.2
	}
	if act > 1.0 {
		act = 1.0
	}
	return act
}

// ComputeNoveltyScore scores novelty based on body length heuristics.
func ComputeNoveltyScore(body string) float64 {
	nov := 0.5
	if len(body) > 800 {
		nov += 0.1
	}
	if len(body) < 250 {
		nov -= 0.1
	}
	if nov > 1.0 {
		nov = 1.0
	}
	if nov < 0.0 {
		nov = 0.0
	}
	return nov
}

// ComputeContextScore scores how well the body documents its context.
func ComputeContextScore(lower string) float64 {
	ctx := 0.5
	if strings.Contains(lower, "## source") || strings.Contains(lower, "**source**") {
		ctx += 0.2
	}
	if strings.Contains(lower, "## why it matters") {
		ctx += 0.1
	}
	if ctx > 1.0 {
		ctx = 1.0
	}
	return ctx
}
