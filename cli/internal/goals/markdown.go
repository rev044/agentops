package goals

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// directiveHeadingRe matches "### N. Title" where N is the directive number.
var directiveHeadingRe = regexp.MustCompile(`^###\s+(\d+)\.\s+(.+)$`)

// tableRowRe matches a markdown table row (non-separator).
var tableRowRe = regexp.MustCompile(`^\s*\|(.+)\|\s*$`)

// tableSepRe matches a markdown table separator row.
var tableSepRe = regexp.MustCompile(`^\s*\|[\s:|-]+\|\s*$`)

// ParseMarkdownGoals parses a GOALS.md file into a GoalFile.
// Sets Version=4 and Format="md".
func ParseMarkdownGoals(data []byte) (*GoalFile, error) {
	content := string(data)
	lines := strings.Split(content, "\n")

	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("empty goals file")
	}

	gf := &GoalFile{
		Version: 4,
		Format:  "md",
	}

	gf.Mission = parseMission(lines)
	gf.NorthStars = parseListSection(lines, "North Stars")
	gf.AntiStars = parseListSection(lines, "Anti Stars")

	directives, err := parseDirectives(lines)
	if err != nil {
		return nil, fmt.Errorf("parsing directives: %w", err)
	}
	gf.Directives = directives

	goals, err := parseGatesTable(lines)
	if err != nil {
		return nil, fmt.Errorf("parsing gates table: %w", err)
	}
	gf.Goals = goals

	return gf, nil
}

// parseMission extracts the mission statement — the first non-empty paragraph
// after a line starting with "# " (case-insensitive match on "Goals").
func parseMission(lines []string) string {
	inHeader := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for "# Goals" or "# <anything>" as the H1
		if strings.HasPrefix(trimmed, "# ") {
			inHeader = true
			continue
		}

		if inHeader {
			// Skip empty lines after the heading
			if trimmed == "" {
				continue
			}
			// Skip if this line is a heading (section start)
			if strings.HasPrefix(trimmed, "#") {
				return ""
			}
			// First non-empty, non-heading line is the mission
			return trimmed
		}
	}
	return ""
}

// parseListSection extracts bullet items from a section matching the given heading.
// Matching is case-insensitive.
func parseListSection(lines []string, heading string) []string {
	var items []string
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for H2 heading match (case-insensitive)
		if strings.HasPrefix(trimmed, "## ") {
			sectionTitle := strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			if strings.EqualFold(sectionTitle, heading) {
				inSection = true
				continue
			}
			if inSection {
				// Hit another H2 — done with this section
				break
			}
			continue
		}

		// Check for any heading at H1/H3+ level — also ends the section
		if inSection && strings.HasPrefix(trimmed, "#") {
			break
		}

		if inSection {
			if trimmed == "" {
				continue
			}
			// Parse bullet items (- or *)
			if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
				item := strings.TrimSpace(trimmed[2:])
				if item != "" {
					items = append(items, item)
				}
			}
		}
	}

	return items
}

// parseDirectives extracts numbered H3 directives with optional body text and **Steer:** line.
func parseDirectives(lines []string) ([]Directive, error) {
	// Find the Directives section (case-insensitive H2 match)
	directiveStart := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			sectionTitle := strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			if strings.EqualFold(sectionTitle, "Directives") {
				directiveStart = i + 1
				break
			}
		}
	}

	if directiveStart < 0 {
		return nil, nil // No directives section — not an error
	}

	var directives []Directive
	var current *Directive
	var bodyLines []string

	flushCurrent := func() {
		if current != nil {
			current.Description = strings.TrimSpace(strings.Join(bodyLines, "\n"))
			directives = append(directives, *current)
		}
	}

	for i := directiveStart; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		// End of directives section at next H2 or H1
		if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "# ") {
			break
		}

		// Check for directive heading: ### N. Title
		if m := directiveHeadingRe.FindStringSubmatch(trimmed); m != nil {
			flushCurrent()
			num, _ := strconv.Atoi(m[1])
			current = &Directive{
				Number: num,
				Title:  strings.TrimSpace(m[2]),
			}
			bodyLines = nil
			continue
		}

		if current != nil {
			// Check for **Steer:** line
			if strings.HasPrefix(trimmed, "**Steer:**") {
				steerVal := strings.TrimSpace(strings.TrimPrefix(trimmed, "**Steer:**"))
				current.Steer = steerVal
				continue
			}
			if trimmed != "" {
				bodyLines = append(bodyLines, trimmed)
			}
		}
	}

	flushCurrent()
	return directives, nil
}

// parseGatesTable extracts goals from a markdown table under the "Gates" section.
// Expected columns: ID | Check | Weight | Description (order may vary if header is present).
func parseGatesTable(lines []string) ([]Goal, error) {
	// Find the Gates section (case-insensitive H2 match)
	gatesStart := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			sectionTitle := strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			if strings.EqualFold(sectionTitle, "Gates") {
				gatesStart = i + 1
				break
			}
		}
	}

	if gatesStart < 0 {
		return nil, nil // No gates section — not an error
	}

	var goals []Goal
	headerFound := false
	colMap := map[string]int{"id": 0, "check": 1, "weight": 2, "description": 3}

	for i := gatesStart; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		// End of gates section at next heading
		if strings.HasPrefix(trimmed, "#") {
			break
		}

		if trimmed == "" {
			continue
		}

		// Must be a table row
		if !tableRowRe.MatchString(trimmed) {
			continue
		}

		// Skip separator rows
		if tableSepRe.MatchString(trimmed) {
			continue
		}

		cells := splitTableRow(trimmed)

		// First non-separator row is the header
		if !headerFound {
			headerFound = true
			// Build column index map from header
			for j, cell := range cells {
				lower := strings.ToLower(strings.TrimSpace(cell))
				switch {
				case lower == "id":
					colMap["id"] = j
				case lower == "check":
					colMap["check"] = j
				case lower == "weight":
					colMap["weight"] = j
				case lower == "description":
					colMap["description"] = j
				}
			}
			continue
		}

		// Parse data row
		g := Goal{Type: GoalTypeHealth}

		if idx, ok := colMap["id"]; ok && idx < len(cells) {
			g.ID = strings.TrimSpace(cells[idx])
		}
		if idx, ok := colMap["check"]; ok && idx < len(cells) {
			check := strings.TrimSpace(cells[idx])
			// Strip backticks
			check = strings.Trim(check, "`")
			g.Check = check
		}
		if idx, ok := colMap["weight"]; ok && idx < len(cells) {
			w, err := strconv.Atoi(strings.TrimSpace(cells[idx]))
			if err != nil {
				// Default weight if unparseable
				w = 5
			}
			g.Weight = w
		}
		if idx, ok := colMap["description"]; ok && idx < len(cells) {
			g.Description = strings.TrimSpace(cells[idx])
		}

		// Use ID as description fallback
		if g.Description == "" {
			g.Description = g.ID
		}

		if g.ID != "" {
			goals = append(goals, g)
		}
	}

	return goals, nil
}

// splitTableRow splits a markdown table row into cells, stripping outer pipes.
func splitTableRow(row string) []string {
	// Remove leading/trailing whitespace and pipes
	row = strings.TrimSpace(row)
	row = strings.Trim(row, "|")
	parts := strings.Split(row, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}
