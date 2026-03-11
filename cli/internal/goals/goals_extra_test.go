package goals

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- DetectFormat (64.3%) ---

func TestExtra_DetectFormat_ExplicitMD(t *testing.T) {
	got := DetectFormat("GOALS.md")
	if got != "md" {
		t.Errorf("DetectFormat(GOALS.md) = %q, want %q", got, "md")
	}
}

func TestExtra_DetectFormat_YAMLNoMDSibling(t *testing.T) {
	// Create a temp dir with only a YAML file (no GOALS.md sibling).
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "GOALS.yaml")
	if err := os.WriteFile(yamlPath, []byte("version: 2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got := DetectFormat(yamlPath)
	if got != "yaml" {
		t.Errorf("DetectFormat(yaml, no md sibling) = %q, want %q", got, "yaml")
	}
}

func TestExtra_DetectFormat_YAMLWithMDSibling(t *testing.T) {
	// When GOALS.md exists alongside GOALS.yaml, prefer md.
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "GOALS.yaml")
	mdPath := filepath.Join(dir, "GOALS.md")
	os.WriteFile(yamlPath, []byte("version: 2\n"), 0o600)
	os.WriteFile(mdPath, []byte("# Goals\n"), 0o600)

	got := DetectFormat(yamlPath)
	if got != "md" {
		t.Errorf("DetectFormat(yaml with md sibling) = %q, want %q", got, "md")
	}
}

func TestExtra_DetectFormat_YMLExtension(t *testing.T) {
	dir := t.TempDir()
	ymlPath := filepath.Join(dir, "GOALS.yml")
	os.WriteFile(ymlPath, []byte("version: 2\n"), 0o600)

	got := DetectFormat(ymlPath)
	if got != "yaml" {
		t.Errorf("DetectFormat(.yml) = %q, want %q", got, "yaml")
	}
}

func TestExtra_DetectFormat_NoExtensionNoMD(t *testing.T) {
	dir := t.TempDir()
	noExtPath := filepath.Join(dir, "GOALS")
	os.WriteFile(noExtPath, []byte("stuff\n"), 0o600)

	got := DetectFormat(noExtPath)
	if got != "yaml" {
		t.Errorf("DetectFormat(no ext, no md) = %q, want %q", got, "yaml")
	}
}

func TestExtra_DetectFormat_NoExtensionWithMD(t *testing.T) {
	dir := t.TempDir()
	noExtPath := filepath.Join(dir, "GOALS")
	mdPath := filepath.Join(dir, "GOALS.md")
	os.WriteFile(noExtPath, []byte("stuff\n"), 0o600)
	os.WriteFile(mdPath, []byte("# Goals\n"), 0o600)

	got := DetectFormat(noExtPath)
	if got != "md" {
		t.Errorf("DetectFormat(no ext, md exists) = %q, want %q", got, "md")
	}
}

// --- LoadGoals (87.5%) ---

func TestExtra_LoadGoals_UnsupportedVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "GOALS.yaml")
	os.WriteFile(path, []byte("version: 99\ngoals: []\n"), 0o600)

	_, err := LoadGoals(path)
	if err == nil {
		t.Fatal("expected error for unsupported version, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported version") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExtra_LoadGoals_Version1Warning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "GOALS.yaml")
	content := `version: 1
goals:
  - id: test-goal
    description: A goal
    check: echo ok
    weight: 5
`
	os.WriteFile(path, []byte(content), 0o600)

	gf, err := LoadGoals(path)
	if err != nil {
		t.Fatalf("LoadGoals v1: %v", err)
	}
	if gf.Format != "yaml" {
		t.Errorf("Format = %q, want %q", gf.Format, "yaml")
	}
	if gf.Version != 1 {
		t.Errorf("Version = %d, want 1", gf.Version)
	}
}

func TestExtra_LoadGoals_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "GOALS.yaml")
	os.WriteFile(path, []byte(":::invalid yaml\n"), 0o600)

	_, err := LoadGoals(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestExtra_LoadGoals_FileNotFound(t *testing.T) {
	_, err := LoadGoals("/nonexistent/GOALS.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestExtra_LoadGoals_MDFileNotFound(t *testing.T) {
	// DetectFormat returns "md" when GOALS.md exists, but if we pass
	// a .md path that doesn't exist, ReadFile should fail.
	_, err := LoadGoals("/nonexistent/GOALS.md")
	if err == nil {
		t.Fatal("expected error for missing md file")
	}
}

func TestExtra_LoadGoals_InvalidMarkdown(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "GOALS.md")
	// Empty file triggers "empty goals file" error.
	os.WriteFile(path, []byte(""), 0o600)

	_, err := LoadGoals(path)
	if err == nil {
		t.Fatal("expected error for empty md file")
	}
	if !strings.Contains(err.Error(), "empty goals file") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- AppendHistory (83.3%) ---

func TestExtra_AppendHistory_CreatesFileAndAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")

	entry := HistoryEntry{
		Timestamp:    time.Now().Format(time.RFC3339),
		GoalsPassing: 3,
		GoalsTotal:   5,
		Score:        60.0,
		GitSHA:       "abc1234",
	}

	if err := AppendHistory(entry, path); err != nil {
		t.Fatalf("AppendHistory (create): %v", err)
	}

	// Append a second entry.
	entry2 := HistoryEntry{
		Timestamp:    time.Now().Format(time.RFC3339),
		GoalsPassing: 4,
		GoalsTotal:   5,
		Score:        80.0,
		GitSHA:       "def5678",
	}
	if err := AppendHistory(entry2, path); err != nil {
		t.Fatalf("AppendHistory (append): %v", err)
	}

	// Verify both entries exist.
	entries, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("got %d entries, want 2", len(entries))
	}
	if entries[0].GoalsPassing != 3 {
		t.Errorf("first entry GoalsPassing = %d, want 3", entries[0].GoalsPassing)
	}
	if entries[1].GoalsPassing != 4 {
		t.Errorf("second entry GoalsPassing = %d, want 4", entries[1].GoalsPassing)
	}
}

func TestExtra_AppendHistory_BadPath(t *testing.T) {
	err := AppendHistory(HistoryEntry{}, "/nonexistent/dir/file.jsonl")
	if err == nil {
		t.Fatal("expected error for bad path")
	}
}

// --- ParseMarkdownGoals (88.2%) ---

func TestExtra_ParseMarkdownGoals_Empty(t *testing.T) {
	_, err := ParseMarkdownGoals([]byte(""))
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestExtra_ParseMarkdownGoals_WhitespaceOnly(t *testing.T) {
	_, err := ParseMarkdownGoals([]byte("   \n  \n  "))
	if err == nil {
		t.Fatal("expected error for whitespace-only input")
	}
}

func TestExtra_ParseMarkdownGoals_NoGatesSection(t *testing.T) {
	md := `# Goals
My mission statement

## North Stars
- Star one
`
	gf, err := ParseMarkdownGoals([]byte(md))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gf.Goals) != 0 {
		t.Errorf("expected 0 goals without Gates section, got %d", len(gf.Goals))
	}
	if gf.Mission != "My mission statement" {
		t.Errorf("Mission = %q, want %q", gf.Mission, "My mission statement")
	}
}

// --- parseMission (92.3%) ---

func TestExtra_parseMission_NoH1(t *testing.T) {
	lines := []string{"No heading here", "just text"}
	got := parseMission(lines)
	if got != "" {
		t.Errorf("parseMission(no H1) = %q, want empty", got)
	}
}

func TestExtra_parseMission_H1FollowedByH2(t *testing.T) {
	// H1 immediately followed by another heading (no mission paragraph).
	lines := []string{"# Goals", "", "## Section"}
	got := parseMission(lines)
	if got != "" {
		t.Errorf("parseMission(H1 then H2) = %q, want empty", got)
	}
}

func TestExtra_parseMission_H1FollowedByH3(t *testing.T) {
	lines := []string{"# Goals", "### Sub-heading"}
	got := parseMission(lines)
	if got != "" {
		t.Errorf("parseMission(H1 then H3) = %q, want empty", got)
	}
}

// --- parseListSection (95.5%) ---

func TestExtra_parseListSection_NonBulletLinesIgnored(t *testing.T) {
	lines := []string{
		"## North Stars",
		"This is not a bullet",
		"- Real bullet",
		"plain text again",
		"* Another bullet",
		"## Next Section",
	}
	items := parseListSection(lines, "North Stars")
	if len(items) != 2 {
		t.Errorf("got %d items, want 2", len(items))
	}
	if items[0] != "Real bullet" {
		t.Errorf("item[0] = %q, want %q", items[0], "Real bullet")
	}
	if items[1] != "Another bullet" {
		t.Errorf("item[1] = %q, want %q", items[1], "Another bullet")
	}
}

func TestExtra_parseListSection_EndedByH1(t *testing.T) {
	lines := []string{
		"## Anti Stars",
		"- Bad thing",
		"# Top Level Heading",
	}
	items := parseListSection(lines, "Anti Stars")
	if len(items) != 1 {
		t.Errorf("got %d items, want 1", len(items))
	}
}

func TestExtra_parseListSection_EndedByH3(t *testing.T) {
	lines := []string{
		"## Anti Stars",
		"- Bad thing",
		"### Sub Section",
		"- Should not appear",
	}
	items := parseListSection(lines, "Anti Stars")
	if len(items) != 1 {
		t.Errorf("got %d items, want 1", len(items))
	}
}

func TestExtra_parseListSection_EmptyBulletSkipped(t *testing.T) {
	lines := []string{
		"## North Stars",
		"- ",
		"- Real item",
	}
	items := parseListSection(lines, "North Stars")
	if len(items) != 1 {
		t.Errorf("got %d items, want 1 (empty bullet skipped)", len(items))
	}
}

func TestExtra_parseListSection_SectionNotFound(t *testing.T) {
	lines := []string{"## Other Section", "- item"}
	items := parseListSection(lines, "Missing Section")
	if len(items) != 0 {
		t.Errorf("got %d items for missing section, want 0", len(items))
	}
}

// --- parseGatesTable (96.9%) ---

func TestExtra_parseGatesTable_NoGatesSection(t *testing.T) {
	lines := []string{"## Other", "| a | b |"}
	goals, err := parseGatesTable(lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if goals != nil {
		t.Errorf("expected nil goals for no Gates section, got %v", goals)
	}
}

func TestExtra_parseGatesTable_EmptyRowSkipped(t *testing.T) {
	lines := []string{
		"## Gates",
		"| ID | Check | Weight | Description |",
		"|---|---|---|---|",
		"| | `echo hi` | 5 | empty id |",
		"| valid-id | `echo ok` | 3 | real goal |",
	}
	goals, err := parseGatesTable(lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The row with empty ID should be skipped.
	if len(goals) != 1 {
		t.Errorf("got %d goals, want 1 (empty ID row skipped)", len(goals))
	}
	if goals[0].ID != "valid-id" {
		t.Errorf("goal ID = %q, want %q", goals[0].ID, "valid-id")
	}
}

func TestExtra_parseGatesTable_InvalidWeightDefaultsTo5(t *testing.T) {
	lines := []string{
		"## Gates",
		"| ID | Check | Weight | Description |",
		"|---|---|---|---|",
		"| my-goal | `echo ok` | abc | test desc |",
	}
	goals, err := parseGatesTable(lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(goals) != 1 {
		t.Fatalf("got %d goals, want 1", len(goals))
	}
	if goals[0].Weight != 5 {
		t.Errorf("Weight = %d, want 5 (default for invalid)", goals[0].Weight)
	}
}

func TestExtra_parseGatesTable_OutOfRangeWeightDefaultsTo5(t *testing.T) {
	lines := []string{
		"## Gates",
		"| ID | Check | Weight | Description |",
		"|---|---|---|---|",
		"| my-goal | `echo ok` | 99 | test |",
	}
	goals, _ := parseGatesTable(lines)
	if goals[0].Weight != 5 {
		t.Errorf("Weight = %d, want 5 (default for out-of-range)", goals[0].Weight)
	}
}

func TestExtra_parseGatesTable_DescriptionFallsBackToID(t *testing.T) {
	lines := []string{
		"## Gates",
		"| ID | Check | Weight | Description |",
		"|---|---|---|---|",
		"| my-goal | `echo ok` | 3 | |",
	}
	goals, _ := parseGatesTable(lines)
	if goals[0].Description != "my-goal" {
		t.Errorf("Description = %q, want %q (fallback to ID)", goals[0].Description, "my-goal")
	}
}

// --- MeasureOne (82.6%) --- tests the Start error path

func TestExtra_MeasureOne_BadCommand(t *testing.T) {
	g := Goal{
		ID:     "bad-cmd",
		Check:  "/nonexistent/binary/that/does/not/exist",
		Weight: 3,
	}
	m := MeasureOne(g, 5*time.Second)
	if m.Result != "fail" {
		t.Errorf("Result = %q, want %q for bad command", m.Result, "fail")
	}
	if m.GoalID != "bad-cmd" {
		t.Errorf("GoalID = %q, want %q", m.GoalID, "bad-cmd")
	}
	if m.Weight != 3 {
		t.Errorf("Weight = %d, want 3", m.Weight)
	}
}

func TestExtra_MeasureOne_TimeoutSkip(t *testing.T) {
	g := Goal{
		ID:     "slow-cmd",
		Check:  "sleep 60",
		Weight: 2,
	}
	m := MeasureOne(g, 100*time.Millisecond)
	if m.Result != "skip" {
		t.Errorf("Result = %q, want %q for timeout", m.Result, "skip")
	}
}

func TestExtra_MeasureOne_ContinuousMetric(t *testing.T) {
	g := Goal{
		ID:     "metric-goal",
		Check:  "echo 95.5",
		Weight: 5,
		Continuous: &ContinuousMetric{
			Metric:    "coverage",
			Threshold: 90.0,
		},
	}
	m := MeasureOne(g, 5*time.Second)
	if m.Result != "pass" {
		t.Errorf("Result = %q, want pass", m.Result)
	}
	if m.Value == nil {
		t.Fatal("Value is nil, want parsed float")
	}
	if *m.Value != 95.5 {
		t.Errorf("Value = %f, want 95.5", *m.Value)
	}
	if m.Threshold == nil {
		t.Fatal("Threshold is nil, want 90.0")
	}
	if *m.Threshold != 90.0 {
		t.Errorf("Threshold = %f, want 90.0", *m.Threshold)
	}
}

// --- killAllChildren (80.0%) ---

func TestExtra_killAllChildren_EmptyPids(t *testing.T) {
	// Calling killAllChildren with no tracked children should not panic.
	// Clear any lingering state.
	childGroups.mu.Lock()
	childGroups.pids = make(map[int]struct{})
	childGroups.mu.Unlock()

	killAllChildren()

	childGroups.mu.Lock()
	defer childGroups.mu.Unlock()
	if len(childGroups.pids) != 0 {
		t.Errorf("pids should be empty after killAllChildren, got %d", len(childGroups.pids))
	}
}

func TestExtra_killAllChildren_WithStalePid(t *testing.T) {
	// Track a PID that doesn't exist (e.g., 999999999). killAllChildren
	// should not panic even if kill fails.
	childGroups.mu.Lock()
	childGroups.pids = map[int]struct{}{999999999: {}}
	childGroups.mu.Unlock()

	killAllChildren() // Should not panic.

	childGroups.mu.Lock()
	defer childGroups.mu.Unlock()
	if len(childGroups.pids) != 0 {
		t.Errorf("pids should be cleared after killAllChildren, got %d", len(childGroups.pids))
	}
}

// --- SaveSnapshot (90.0%) ---

func TestExtra_SaveSnapshot_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "snapshots")
	s := &Snapshot{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		GitSHA:    "abc",
		Goals:     []Measurement{{GoalID: "g1", Result: "pass", Weight: 5}},
		Summary:   SnapshotSummary{Total: 1, Passing: 1, Score: 100},
	}

	path, err := SaveSnapshot(s, dir)
	if err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}
	if !strings.HasSuffix(path, ".json") {
		t.Errorf("path %q should end with .json", path)
	}

	// Verify we can load it back.
	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if loaded.GitSHA != "abc" {
		t.Errorf("GitSHA = %q, want %q", loaded.GitSHA, "abc")
	}
	if len(loaded.Goals) != 1 {
		t.Errorf("Goals count = %d, want 1", len(loaded.Goals))
	}
}

func TestExtra_SaveSnapshot_BadDir(t *testing.T) {
	// Try to save to a path where dir creation will fail.
	s := &Snapshot{}
	_, err := SaveSnapshot(s, "/dev/null/impossible")
	if err == nil {
		t.Fatal("expected error for impossible directory")
	}
}

// --- runGoals (80.0%) ---

func TestExtra_runGoals_MetaAndNonMeta(t *testing.T) {
	goals := []Goal{
		{ID: "meta-g", Check: "echo meta", Weight: 3, Type: GoalTypeMeta},
		{ID: "health-g", Check: "echo health", Weight: 5, Type: GoalTypeHealth},
	}
	measurements := runGoals(goals, 5*time.Second)
	if len(measurements) != 2 {
		t.Fatalf("got %d measurements, want 2", len(measurements))
	}
	// Meta should run first.
	if measurements[0].GoalID != "meta-g" {
		t.Errorf("first measurement = %q, want %q (meta first)", measurements[0].GoalID, "meta-g")
	}
	if measurements[1].GoalID != "health-g" {
		t.Errorf("second measurement = %q, want %q", measurements[1].GoalID, "health-g")
	}
}

func TestExtra_runGoals_OnlyMeta(t *testing.T) {
	goals := []Goal{
		{ID: "m1", Check: "echo m1", Weight: 1, Type: GoalTypeMeta},
	}
	measurements := runGoals(goals, 5*time.Second)
	if len(measurements) != 1 {
		t.Fatalf("got %d measurements, want 1", len(measurements))
	}
	if measurements[0].GoalID != "m1" {
		t.Errorf("GoalID = %q, want %q", measurements[0].GoalID, "m1")
	}
}

func TestExtra_runGoals_ExclusiveExecution(t *testing.T) {
	// Goals containing "go test" trigger exclusive execution.
	goals := []Goal{
		{ID: "test-g", Check: "echo 'go test placeholder'", Weight: 5, Type: GoalTypeHealth},
		{ID: "normal-g", Check: "echo normal", Weight: 3, Type: GoalTypeHealth},
	}
	measurements := runGoals(goals, 5*time.Second)
	if len(measurements) != 2 {
		t.Fatalf("got %d measurements, want 2", len(measurements))
	}
	for _, m := range measurements {
		if m.Result != "pass" {
			t.Errorf("goal %q Result = %q, want pass", m.GoalID, m.Result)
		}
	}
}

// --- computeSummary ---

func TestExtra_computeSummary_AllSkipped(t *testing.T) {
	ms := []Measurement{
		{GoalID: "g1", Result: "skip", Weight: 5},
		{GoalID: "g2", Result: "skip", Weight: 3},
	}
	s := computeSummary(ms)
	if s.Skipped != 2 {
		t.Errorf("Skipped = %d, want 2", s.Skipped)
	}
	if s.Score != 0 {
		t.Errorf("Score = %f, want 0 (all skipped)", s.Score)
	}
}
