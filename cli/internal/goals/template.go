package goals

import (
	"fmt"
	"strings"
)

// RenderGoalsMD produces a well-formatted GOALS.md string from a GoalFile.
func RenderGoalsMD(gf *GoalFile) string {
	var b strings.Builder

	// Header + mission
	b.WriteString("# Goals\n\n")
	if gf.Mission != "" {
		b.WriteString(gf.Mission)
		b.WriteString("\n")
	}

	// North Stars
	if len(gf.NorthStars) > 0 {
		b.WriteString("\n## North Stars\n\n")
		for _, s := range gf.NorthStars {
			fmt.Fprintf(&b, "- %s\n", s)
		}
	}

	// Anti Stars
	if len(gf.AntiStars) > 0 {
		b.WriteString("\n## Anti Stars\n\n")
		for _, s := range gf.AntiStars {
			fmt.Fprintf(&b, "- %s\n", s)
		}
	}

	// Directives
	if len(gf.Directives) > 0 {
		b.WriteString("\n## Directives\n")
		for _, d := range gf.Directives {
			fmt.Fprintf(&b, "\n### %d. %s\n\n", d.Number, d.Title)
			if d.Description != "" {
				b.WriteString(d.Description)
				b.WriteString("\n\n")
			}
			if d.Steer != "" {
				fmt.Fprintf(&b, "**Steer:** %s\n", d.Steer)
			}
		}
	}

	// Gates table
	if len(gf.Goals) > 0 {
		b.WriteString("\n## Gates\n\n")
		b.WriteString("| ID | Check | Weight | Description |\n")
		b.WriteString("|----|-------|--------|-------------|\n")
		for _, g := range gf.Goals {
			fmt.Fprintf(&b, "| %s | `%s` | %d | %s |\n", g.ID, g.Check, g.Weight, g.Description)
		}
	}

	return b.String()
}
