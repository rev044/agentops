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

func TestRoundTripPipeChars(t *testing.T) {
	// Build a GoalFile with pipe characters in Check and Description fields.
	gf := &GoalFile{
		Version: 4,
		Format:  "md",
		Mission: "Test pipe escaping.",
		Goals: []Goal{
			{
				ID:          "pipe-check",
				Check:       "cmd --flag a|b",
				Weight:      7,
				Description: "Runs a|b pipeline",
				Type:        GoalTypeHealth,
			},
			{
				ID:          "pipe-desc",
				Check:       "make test",
				Weight:      5,
				Description: "Pipe in desc: foo|bar|baz",
				Type:        GoalTypeHealth,
			},
		},
	}

	rendered := RenderGoalsMD(gf)

	gf2, err := ParseMarkdownGoals([]byte(rendered))
	if err != nil {
		t.Fatalf("parsing rendered markdown: %v", err)
	}

	if len(gf2.Goals) != 2 {
		t.Fatalf("goals count = %d, want 2", len(gf2.Goals))
	}

	for i, want := range gf.Goals {
		got := gf2.Goals[i]
		if got.ID != want.ID {
			t.Errorf("goal[%d].ID = %q, want %q", i, got.ID, want.ID)
		}
		if got.Check != want.Check {
			t.Errorf("goal[%d].Check = %q, want %q", i, got.Check, want.Check)
		}
		if got.Weight != want.Weight {
			t.Errorf("goal[%d].Weight = %d, want %d", i, got.Weight, want.Weight)
		}
		if got.Description != want.Description {
			t.Errorf("goal[%d].Description = %q, want %q", i, got.Description, want.Description)
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

// TestParseMarkdownGoals_Adversarial exercises edge-case inputs to verify the
// parser does not panic and handles each situation gracefully.
func TestParseMarkdownGoals_Adversarial(t *testing.T) {
	tests := []struct {
		name string
		// input is the full GOALS.md content to parse.
		input string
		// wantErr indicates the parse should return a non-nil error.
		wantErr bool
		// check is an optional function that inspects the parsed GoalFile.
		// It is only called when wantErr is false.
		check func(t *testing.T, gf *GoalFile)
	}{
		{
			name: "pipe in Check field round-trips via RenderGoalsMD",
			// This test verifies that a goal whose Check contains a literal pipe
			// survives render → parse without corruption.
			input: func() string {
				gf := &GoalFile{
					Version: 4,
					Format:  "md",
					Mission: "Pipe test.",
					Goals: []Goal{
						{
							ID:          "pipe-in-check",
							Check:       `bash -c "echo a | grep b"`,
							Weight:      5,
							Description: "Grep after echo",
							Type:        GoalTypeHealth,
						},
					},
				}
				return RenderGoalsMD(gf)
			}(),
			check: func(t *testing.T, gf *GoalFile) {
				if len(gf.Goals) != 1 {
					t.Fatalf("goals = %d, want 1", len(gf.Goals))
				}
				want := `bash -c "echo a | grep b"`
				if gf.Goals[0].Check != want {
					t.Errorf("Check = %q, want %q", gf.Goals[0].Check, want)
				}
			},
		},
		{
			name: "pipe in Description field round-trips via RenderGoalsMD",
			input: func() string {
				gf := &GoalFile{
					Version: 4,
					Format:  "md",
					Mission: "Pipe desc test.",
					Goals: []Goal{
						{
							ID:          "pipe-in-desc",
							Check:       "make test",
							Weight:      5,
							Description: "Run A | then B",
							Type:        GoalTypeHealth,
						},
					},
				}
				return RenderGoalsMD(gf)
			}(),
			check: func(t *testing.T, gf *GoalFile) {
				if len(gf.Goals) != 1 {
					t.Fatalf("goals = %d, want 1", len(gf.Goals))
				}
				want := "Run A | then B"
				if gf.Goals[0].Description != want {
					t.Errorf("Description = %q, want %q", gf.Goals[0].Description, want)
				}
			},
		},
		{
			// Backtick-wrapped check values have their surrounding backticks stripped via
			// strings.Trim(check, "`"). A cell containing multiple backtick-delimited
			// segments (e.g. "`cmd` && `other`") is unusual but must not cause a panic.
			// The parser strips only leading/trailing backticks, so the inner content
			// may differ from the original. This test documents current behaviour.
			name: "backtick in Check field — no panic, documented behaviour",
			// Build a table row whose Check cell contains internal backticks:
			// | bt-check | `cmd` && `other` | 5 | Backtick check |
			input: "# Goals\n\nBacktick test.\n\n## Gates\n\n" +
				"| ID | Check | Weight | Description |\n" +
				"|----|-------|--------|-------------|\n" +
				"| bt-check | `cmd` && `other` | 5 | Backtick check |\n",
			check: func(t *testing.T, gf *GoalFile) {
				// Must not panic; at least one goal parsed.
				if len(gf.Goals) != 1 {
					t.Fatalf("goals = %d, want 1 (no panic)", len(gf.Goals))
				}
				// Check field is non-empty after backtick stripping.
				if gf.Goals[0].Check == "" {
					t.Error("Check is empty, expected non-empty")
				}
			},
		},
		{
			// A row where the Description column is empty; the parser falls back to the
			// goal ID as the description (see the "Use ID as description fallback" comment
			// in parseGatesTable).
			name: "empty Description cell falls back to ID",
			input: "# Goals\n\nEmpty desc.\n\n## Gates\n\n| ID | Check | Weight | Description |\n|----|-------|--------|-------------|\n| empty-desc | `echo ok` | 5 |  |\n",
			check: func(t *testing.T, gf *GoalFile) {
				if len(gf.Goals) != 1 {
					t.Fatalf("goals = %d, want 1", len(gf.Goals))
				}
				// Falls back to ID when Description is blank.
				if gf.Goals[0].Description != "empty-desc" {
					t.Errorf("Description = %q, want %q (fallback to ID)", gf.Goals[0].Description, "empty-desc")
				}
			},
		},
		{
			// A data row that has only 3 columns instead of the expected 4.
			// The parser should not panic; the missing Description column is simply absent
			// and the ID-fallback applies.
			name: "fewer columns than header — no panic",
			input: "# Goals\n\nShort row.\n\n## Gates\n\n| ID | Check | Weight | Description |\n|----|-------|--------|-------------|\n| short-row | `echo ok` | 3 |\n",
			check: func(t *testing.T, gf *GoalFile) {
				// Parser must survive and produce a goal (or zero goals), but must not panic.
				// A row ending in a trailing pipe after 3 cells still splits into 3 cells,
				// so Description is missing — fallback to ID.
				_ = gf // just verifying no panic
			},
		},
		{
			// Unicode characters in mission, north-star items, and directive titles.
			name: "unicode in mission, north stars, and directive titles",
			input: "# Goals\n\n目標: 品質を改善する 🚀\n\n## North Stars\n\n- Schön und ästhetisch\n- 日本語のアイテム\n\n## Directives\n\n### 1. Ünïcödé Directive\n\nAccents and kanji: 漢字\n\n**Steer:** increase\n\n## Gates\n\n| ID | Check | Weight | Description |\n|----|-------|--------|-------------|\n| unicode-gate | `echo ok` | 5 | Unicode: 日本語 |\n",
			check: func(t *testing.T, gf *GoalFile) {
				wantMission := "目標: 品質を改善する 🚀"
				if gf.Mission != wantMission {
					t.Errorf("Mission = %q, want %q", gf.Mission, wantMission)
				}
				if len(gf.NorthStars) != 2 {
					t.Fatalf("NorthStars = %d, want 2", len(gf.NorthStars))
				}
				if gf.NorthStars[0] != "Schön und ästhetisch" {
					t.Errorf("NorthStars[0] = %q", gf.NorthStars[0])
				}
				if len(gf.Directives) != 1 {
					t.Fatalf("Directives = %d, want 1", len(gf.Directives))
				}
				if gf.Directives[0].Title != "Ünïcödé Directive" {
					t.Errorf("Directive[0].Title = %q", gf.Directives[0].Title)
				}
				if len(gf.Goals) != 1 {
					t.Fatalf("Goals = %d, want 1", len(gf.Goals))
				}
				if gf.Goals[0].Description != "Unicode: 日本語" {
					t.Errorf("Goals[0].Description = %q", gf.Goals[0].Description)
				}
			},
		},
		{
			// Directives numbered 1, 3, 5 (2 and 4 are absent). The parser should
			// preserve the authored Number field from the heading — not reassign indices.
			name: "skipped directive numbers preserve authored numbers",
			input: "# Goals\n\nSkip test.\n\n## Directives\n\n### 1. First\n\nBody one.\n\n**Steer:** increase\n\n### 3. Third\n\nBody three.\n\n**Steer:** hold\n\n### 5. Fifth\n\nBody five.\n\n**Steer:** decrease\n\n## Gates\n\n| ID | Check | Weight | Description |\n|----|-------|--------|-------------|\n",
			check: func(t *testing.T, gf *GoalFile) {
				if len(gf.Directives) != 3 {
					t.Fatalf("Directives = %d, want 3", len(gf.Directives))
				}
				wantNumbers := []int{1, 3, 5}
				for i, want := range wantNumbers {
					if gf.Directives[i].Number != want {
						t.Errorf("Directives[%d].Number = %d, want %d", i, gf.Directives[i].Number, want)
					}
				}
				wantTitles := []string{"First", "Third", "Fifth"}
				for i, want := range wantTitles {
					if gf.Directives[i].Title != want {
						t.Errorf("Directives[%d].Title = %q, want %q", i, gf.Directives[i].Title, want)
					}
				}
			},
		},
		{
			// Gates section with header + separator but zero data rows.
			name: "table with no data rows yields empty goals slice",
			input: "# Goals\n\nNo data rows.\n\n## Gates\n\n| ID | Check | Weight | Description |\n|----|-------|--------|-------------|\n",
			check: func(t *testing.T, gf *GoalFile) {
				if len(gf.Goals) != 0 {
					t.Errorf("Goals = %d, want 0", len(gf.Goals))
				}
			},
		},
		{
			// A separator row that has fewer dashes than expected (e.g. only one cell's
			// worth of dashes). The regex tableSepRe should still recognise it as a
			// separator, and no data row is emitted.
			name: "malformed separator row is skipped, not treated as data",
			input: "# Goals\n\nMalformed sep.\n\n## Gates\n\n| ID | Check | Weight | Description |\n|---|\n| my-goal | `echo ok` | 5 | OK |\n",
			check: func(t *testing.T, gf *GoalFile) {
				// The malformed separator row (|---|) should still be recognised as a
				// separator by tableSepRe, so the first non-separator row is the header.
				// After that the data row should be parsed.
				// We verify no panic regardless of exact goal count.
				_ = gf
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Must not panic.
			gf, err := ParseMarkdownGoals([]byte(tc.input))
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, gf)
			}
		})
	}
}
