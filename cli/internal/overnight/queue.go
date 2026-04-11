package overnight

import (
	"bufio"
	"errors"
	"os"
	"regexp"
	"strings"
)

// QueueItem is one operator-pinned nightly priority read from a markdown
// queue file. Format mirrors skills/evolve/references/pinned-queue.md.
type QueueItem struct {
	// Order is the 1-based index assigned by ParseQueue in the order items
	// appear in the source file.
	Order int
	// Title is the bullet text with inline markers stripped.
	Title string
	// Description is any body paragraph attached to the bullet. May be
	// empty.
	Description string
	// TargetFile is parsed from an inline "[file: <path>]" marker.
	TargetFile string
	// Severity is parsed from an inline "[severity: <level>]" marker.
	Severity string
}

// commentRe removes HTML comments (single- or multi-line within one line).
var commentRe = regexp.MustCompile(`(?s)<!--.*?-->`)

// markerRe matches "[key: value]" inline markers. Keys are matched
// case-insensitively; the first match per key wins.
var markerRe = regexp.MustCompile(`\[\s*([a-zA-Z_]+)\s*:\s*([^\]]+?)\s*\]`)

// bulletRe matches the start of a bullet (either "- " or "* ") with any
// leading whitespace. Indented bullets still count as items for parsing
// simplicity; nested hierarchy is not preserved.
var bulletRe = regexp.MustCompile(`^(\s*)[-*]\s+(.*)$`)

// ParseQueue reads the given path and returns ordered items.
//
// Soft-fail semantics:
//
//   - path == ""                → (nil, nil) — no queue configured.
//   - path does not exist       → (nil, nil) — caller may log degraded.
//   - path exists but unreadable → (nil, non-nil error).
//
// See queue.md docs and the evolve pinned-queue reference for the file
// format. This parser is intentionally forgiving: unknown markers are
// dropped, comments are stripped, headings and blank lines are ignored,
// and a bullet without a body becomes a Title-only item.
func ParseQueue(path string) ([]QueueItem, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var items []QueueItem
	var current *QueueItem
	var descLines []string

	flush := func() {
		if current == nil {
			return
		}
		if desc := strings.TrimSpace(strings.Join(descLines, " ")); desc != "" {
			current.Description = desc
		}
		items = append(items, *current)
		current = nil
		descLines = nil
	}

	for scanner.Scan() {
		raw := scanner.Text()
		stripped := commentRe.ReplaceAllString(raw, "")
		// Skip lines that are now entirely blank after comment removal only
		// when we're between items; mid-item blank lines terminate the
		// current description paragraph.
		trimmed := strings.TrimSpace(stripped)

		if strings.HasPrefix(trimmed, "#") {
			flush()
			continue
		}

		if match := bulletRe.FindStringSubmatch(stripped); match != nil {
			flush()
			title, tgt, sev := extractMarkers(match[2])
			current = &QueueItem{
				Order:      len(items) + 1,
				Title:      strings.TrimSpace(title),
				TargetFile: tgt,
				Severity:   sev,
			}
			continue
		}

		if current == nil {
			continue
		}

		if trimmed == "" {
			// A blank line ends the description paragraph but the item
			// remains current until flushed by the next bullet or heading
			// or EOF. We flush here so later stray text does not attach.
			if len(descLines) > 0 {
				flush()
			}
			continue
		}

		descLines = append(descLines, trimmed)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	flush()

	return items, nil
}

// extractMarkers pulls the "[file:...]" and "[severity:...]" markers out of
// a bullet body and returns the cleaned title plus parsed values. Unknown
// marker keys are left in the title untouched so operators can see them.
func extractMarkers(raw string) (title, targetFile, severity string) {
	title = raw
	matches := markerRe.FindAllStringSubmatchIndex(raw, -1)
	// Walk in reverse so we can strip by index without shifting earlier
	// match offsets.
	type hit struct {
		start, end     int
		key, value     string
		recognizedKeys bool
	}
	var hits []hit
	for _, m := range matches {
		if len(m) < 6 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(raw[m[2]:m[3]]))
		value := strings.TrimSpace(raw[m[4]:m[5]])
		recognized := key == "file" || key == "severity"
		hits = append(hits, hit{start: m[0], end: m[1], key: key, value: value, recognizedKeys: recognized})
	}
	for i := len(hits) - 1; i >= 0; i-- {
		h := hits[i]
		if !h.recognizedKeys {
			continue
		}
		switch h.key {
		case "file":
			if targetFile == "" {
				targetFile = h.value
			}
		case "severity":
			if severity == "" {
				severity = h.value
			}
		}
		title = title[:h.start] + title[h.end:]
	}
	title = strings.TrimSpace(title)
	return title, targetFile, severity
}
