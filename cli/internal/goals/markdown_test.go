package goals

import (
	"os"
	"strings"
	"testing"
)

func TestParseMarkdownGoals_ValidFixture(t *testing.T) {
	data, err := os.ReadFile(testdataPath("valid_goals.md"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	gf, err := ParseMarkdownGoals(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gf.Version != 4 {
		t.Errorf("version = %d, want 4", gf.Version)
	}
	if gf.Format != "md" {
		t.Errorf("format = %q, want %q", gf.Format, "md")
	}
	if gf.Mission != "Ship reliable, well-tested software that improves continuously." {
		t.Errorf("mission = %q", gf.Mission)
	}
	if len(gf.NorthStars) != 2 {
		t.Fatalf("north stars = %d, want 2", len(gf.NorthStars))
	}
	if len(gf.AntiStars) != 2 {
		t.Fatalf("anti stars = %d, want 2", len(gf.AntiStars))
	}
	if len(gf.Directives) != 2 {
		t.Fatalf("directives = %d, want 2", len(gf.Directives))
	}
	if len(gf.Goals) != 2 {
		t.Fatalf("goals = %d, want 2", len(gf.Goals))
	}
}

func TestParseMission(t *testing.T) {
	lines := strings.Split("# Goals\n\nBuild great software.\n\n## North Stars\n", "\n")
	mission := parseMission(lines)
	if mission != "Build great software." {
		t.Errorf("mission = %q, want %q", mission, "Build great software.")
	}
}

func TestParseMission_Empty(t *testing.T) {
	lines := strings.Split("# Goals\n\n## North Stars\n", "\n")
	mission := parseMission(lines)
	if mission != "" {
		t.Errorf("mission = %q, want empty", mission)
	}
}

func TestParseListSection_NorthStars(t *testing.T) {
	input := "## North Stars\n\n- First star\n- Second star\n\n## Anti Stars\n"
	items := parseListSection(strings.Split(input, "\n"), "North Stars")
	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}
	if items[0] != "First star" {
		t.Errorf("items[0] = %q", items[0])
	}
}

func TestParseListSection_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name    string
		heading string
	}{
		{"lowercase", "## north stars\n\n- Item\n"},
		{"uppercase", "## NORTH STARS\n\n- Item\n"},
		{"mixed", "## North stars\n\n- Item\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			items := parseListSection(strings.Split(tc.heading, "\n"), "North Stars")
			if len(items) != 1 {
				t.Errorf("items = %d, want 1", len(items))
			}
		})
	}
}

func TestParseListSection_AsteriskBullets(t *testing.T) {
	input := "## North Stars\n\n* First\n* Second\n"
	items := parseListSection(strings.Split(input, "\n"), "North Stars")
	if len(items) != 2 {
		t.Errorf("items = %d, want 2", len(items))
	}
}

func TestParseListSection_Missing(t *testing.T) {
	input := "## Something Else\n\n- Item\n"
	items := parseListSection(strings.Split(input, "\n"), "North Stars")
	if len(items) != 0 {
		t.Errorf("items = %d, want 0", len(items))
	}
}

func TestParseDirectives(t *testing.T) {
	input := `## Directives

### 1. First Directive

Some description here.

**Steer:** increase

### 2. Second Directive

Another description.

**Steer:** decrease

## Gates
`
	directives, err := parseDirectives(strings.Split(input, "\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(directives) != 2 {
		t.Fatalf("directives = %d, want 2", len(directives))
	}
	if directives[0].Number != 1 {
		t.Errorf("d[0].Number = %d, want 1", directives[0].Number)
	}
	if directives[0].Title != "First Directive" {
		t.Errorf("d[0].Title = %q", directives[0].Title)
	}
	if directives[0].Description != "Some description here." {
		t.Errorf("d[0].Description = %q", directives[0].Description)
	}
	if directives[0].Steer != "increase" {
		t.Errorf("d[0].Steer = %q", directives[0].Steer)
	}
	if directives[1].Number != 2 {
		t.Errorf("d[1].Number = %d, want 2", directives[1].Number)
	}
}

func TestParseDirectives_CaseInsensitive(t *testing.T) {
	input := "## DIRECTIVES\n\n### 1. Test\n\nBody.\n\n**Steer:** hold\n"
	directives, err := parseDirectives(strings.Split(input, "\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(directives) != 1 {
		t.Errorf("directives = %d, want 1", len(directives))
	}
}

func TestParseDirectives_Missing(t *testing.T) {
	input := "## Gates\n\n| ID | Check | Weight | Description |\n"
	directives, err := parseDirectives(strings.Split(input, "\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(directives) != 0 {
		t.Errorf("directives = %d, want 0", len(directives))
	}
}

func TestParseGatesTable(t *testing.T) {
	input := `## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| build-ok | ` + "`make build`" + ` | 8 | Build passes |
| test-ok | ` + "`make test`" + ` | 5 | Tests pass |

## Next Section
`
	goals, err := parseGatesTable(strings.Split(input, "\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(goals) != 2 {
		t.Fatalf("goals = %d, want 2", len(goals))
	}
	if goals[0].ID != "build-ok" {
		t.Errorf("goals[0].ID = %q", goals[0].ID)
	}
	if goals[0].Check != "make build" {
		t.Errorf("goals[0].Check = %q (backticks should be stripped)", goals[0].Check)
	}
	if goals[0].Weight != 8 {
		t.Errorf("goals[0].Weight = %d, want 8", goals[0].Weight)
	}
	if goals[0].Description != "Build passes" {
		t.Errorf("goals[0].Description = %q", goals[0].Description)
	}
}

func TestParseGatesTable_CaseInsensitive(t *testing.T) {
	input := "## GATES\n\n| ID | Check | Weight | Description |\n|----|-------|--------|-------------|\n| test | `echo ok` | 5 | Test |\n"
	goals, err := parseGatesTable(strings.Split(input, "\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(goals) != 1 {
		t.Errorf("goals = %d, want 1", len(goals))
	}
}

func TestParseGatesTable_ExtraWhitespace(t *testing.T) {
	input := "## Gates\n\n|  ID  |  Check  |  Weight  |  Description  |\n| --- | --- | --- | --- |\n|  spaced  |  `echo hi`  |  3  |  Spaced out  |\n"
	goals, err := parseGatesTable(strings.Split(input, "\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(goals) != 1 {
		t.Fatalf("goals = %d, want 1", len(goals))
	}
	if goals[0].ID != "spaced" {
		t.Errorf("goals[0].ID = %q, want %q", goals[0].ID, "spaced")
	}
}

func TestParseGatesTable_MissingWeight(t *testing.T) {
	input := "## Gates\n\n| ID | Check | Weight | Description |\n|----|-------|--------|-------------|\n| no-weight | `echo ok` | bad | Test |\n"
	goals, err := parseGatesTable(strings.Split(input, "\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(goals) != 1 {
		t.Fatalf("goals = %d, want 1", len(goals))
	}
	// Bad weight defaults to 5
	if goals[0].Weight != 5 {
		t.Errorf("goals[0].Weight = %d, want 5 (default)", goals[0].Weight)
	}
}

func TestParseGatesTable_Missing(t *testing.T) {
	input := "## Directives\n\n### 1. Only Directives\n"
	goals, err := parseGatesTable(strings.Split(input, "\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(goals) != 0 {
		t.Errorf("goals = %d, want 0", len(goals))
	}
}

func TestParseMarkdownGoals_Empty(t *testing.T) {
	_, err := ParseMarkdownGoals([]byte(""))
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestParseMarkdownGoals_MissingSections(t *testing.T) {
	// Minimal valid: just a heading
	data := []byte("# Goals\n\nJust a mission.\n")
	gf, err := ParseMarkdownGoals(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gf.Mission != "Just a mission." {
		t.Errorf("mission = %q", gf.Mission)
	}
	if len(gf.NorthStars) != 0 {
		t.Errorf("north stars should be empty, got %d", len(gf.NorthStars))
	}
	if len(gf.Directives) != 0 {
		t.Errorf("directives should be empty, got %d", len(gf.Directives))
	}
	if len(gf.Goals) != 0 {
		t.Errorf("goals should be empty, got %d", len(gf.Goals))
	}
}

func TestRoundTrip(t *testing.T) {
	data, err := os.ReadFile(testdataPath("valid_goals.md"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	gf1, err := ParseMarkdownGoals(data)
	if err != nil {
		t.Fatalf("first parse: %v", err)
	}

	rendered := RenderGoalsMD(gf1)

	gf2, err := ParseMarkdownGoals([]byte(rendered))
	if err != nil {
		t.Fatalf("second parse: %v", err)
	}

	// Compare key fields
	if gf1.Mission != gf2.Mission {
		t.Errorf("mission mismatch: %q vs %q", gf1.Mission, gf2.Mission)
	}
	if len(gf1.NorthStars) != len(gf2.NorthStars) {
		t.Errorf("north stars count: %d vs %d", len(gf1.NorthStars), len(gf2.NorthStars))
	}
	if len(gf1.AntiStars) != len(gf2.AntiStars) {
		t.Errorf("anti stars count: %d vs %d", len(gf1.AntiStars), len(gf2.AntiStars))
	}
	if len(gf1.Directives) != len(gf2.Directives) {
		t.Errorf("directives count: %d vs %d", len(gf1.Directives), len(gf2.Directives))
	}
	for i := range gf1.Directives {
		if gf1.Directives[i].Title != gf2.Directives[i].Title {
			t.Errorf("directive[%d] title: %q vs %q", i, gf1.Directives[i].Title, gf2.Directives[i].Title)
		}
		if gf1.Directives[i].Steer != gf2.Directives[i].Steer {
			t.Errorf("directive[%d] steer: %q vs %q", i, gf1.Directives[i].Steer, gf2.Directives[i].Steer)
		}
	}
	if len(gf1.Goals) != len(gf2.Goals) {
		t.Errorf("goals count: %d vs %d", len(gf1.Goals), len(gf2.Goals))
	}
	for i := range gf1.Goals {
		if gf1.Goals[i].ID != gf2.Goals[i].ID {
			t.Errorf("goal[%d] ID: %q vs %q", i, gf1.Goals[i].ID, gf2.Goals[i].ID)
		}
		if gf1.Goals[i].Check != gf2.Goals[i].Check {
			t.Errorf("goal[%d] Check: %q vs %q", i, gf1.Goals[i].Check, gf2.Goals[i].Check)
		}
		if gf1.Goals[i].Weight != gf2.Goals[i].Weight {
			t.Errorf("goal[%d] Weight: %d vs %d", i, gf1.Goals[i].Weight, gf2.Goals[i].Weight)
		}
	}
}

func TestSplitTableRow(t *testing.T) {
	row := "| foo | bar | baz |"
	cells := splitTableRow(row)
	if len(cells) != 3 {
		t.Fatalf("cells = %d, want 3", len(cells))
	}
	if cells[0] != "foo" {
		t.Errorf("cells[0] = %q", cells[0])
	}
}
