package ratchet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/types"
)

// --- loadJSONLChain (92.3%) ---

func TestExtra_loadJSONLChain_MalformedMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chain.jsonl")
	os.WriteFile(path, []byte("not valid json\n"), 0o600)

	_, err := loadJSONLChain(path)
	if err == nil {
		t.Fatal("expected error for malformed metadata line")
	}
}

func TestExtra_loadJSONLChain_MalformedEntrySkipped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chain.jsonl")
	meta := `{"id":"c1","started":"2025-01-01T00:00:00Z"}`
	badEntry := `not json`
	goodEntry := `{"step":"research","output":"out.md","timestamp":"2025-01-01T00:00:00Z"}`
	content := meta + "\n" + badEntry + "\n" + goodEntry + "\n"
	os.WriteFile(path, []byte(content), 0o600)

	chain, err := loadJSONLChain(path)
	if err != nil {
		t.Fatalf("loadJSONLChain: %v", err)
	}
	if len(chain.Entries) != 1 {
		t.Errorf("got %d entries, want 1 (malformed skipped)", len(chain.Entries))
	}
}

func TestExtra_loadJSONLChain_FileNotFound(t *testing.T) {
	_, err := loadJSONLChain("/nonexistent/chain.jsonl")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// --- Save (83.3%) ---

func TestExtra_Save_NoPath(t *testing.T) {
	c := &Chain{ID: "test", Entries: []ChainEntry{}}
	err := c.Save()
	if err != ErrChainNoPath {
		t.Errorf("Save() = %v, want ErrChainNoPath", err)
	}
}

func TestExtra_Save_WritesAndReloads(t *testing.T) {
	dir := t.TempDir()
	chainDir := filepath.Join(dir, ".agents", "ao")
	os.MkdirAll(chainDir, 0o700)
	path := filepath.Join(chainDir, "chain.jsonl")

	c := &Chain{
		ID:      "c-save",
		Started: time.Now(),
		Entries: []ChainEntry{
			{Step: StepResearch, Output: "research.md", Timestamp: time.Now(), Locked: true},
		},
		path: path,
	}

	if err := c.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := loadJSONLChain(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if loaded.ID != "c-save" {
		t.Errorf("ID = %q, want %q", loaded.ID, "c-save")
	}
	if len(loaded.Entries) != 1 {
		t.Errorf("Entries = %d, want 1", len(loaded.Entries))
	}
}

// --- Append (78.6%) ---

func TestExtra_Append_NoPath(t *testing.T) {
	c := &Chain{ID: "test", Entries: []ChainEntry{}}
	err := c.Append(ChainEntry{Step: StepResearch, Output: "out.md"})
	if err != ErrChainNoPath {
		t.Errorf("Append() = %v, want ErrChainNoPath", err)
	}
}

func TestExtra_Append_CreatesFileAndAddsEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chain.jsonl")

	c := &Chain{
		ID:      "c-append",
		Started: time.Now(),
		Entries: []ChainEntry{},
		path:    path,
	}

	entry := ChainEntry{
		Step:      StepResearch,
		Output:    "research.md",
		Timestamp: time.Now(),
	}
	if err := c.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if len(c.Entries) != 1 {
		t.Errorf("in-memory entries = %d, want 1", len(c.Entries))
	}

	// Append a second entry.
	entry2 := ChainEntry{
		Step:      StepPlan,
		Output:    "plan.md",
		Timestamp: time.Now(),
	}
	if err := c.Append(entry2); err != nil {
		t.Fatalf("Append second: %v", err)
	}
	if len(c.Entries) != 2 {
		t.Errorf("in-memory entries = %d, want 2", len(c.Entries))
	}

	// Reload and verify.
	loaded, err := loadJSONLChain(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(loaded.Entries) != 2 {
		t.Errorf("reloaded entries = %d, want 2", len(loaded.Entries))
	}
}

// --- withLockedFile (92.3%) ---

func TestExtra_withLockedFile_BadDirectory(t *testing.T) {
	c := &Chain{path: "/dev/null/impossible/chain.jsonl"}
	err := c.withLockedFile(os.O_RDWR|os.O_CREATE, func(f *os.File) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for impossible directory")
	}
}

// --- writeMetadata / writeEntries (71.4%) ---

func TestExtra_writeMetadata_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "meta.jsonl")

	c := &Chain{ID: "meta-test", Started: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC), EpicID: "ep-1"}

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.writeMetadata(f); err != nil {
		f.Close()
		t.Fatalf("writeMetadata: %v", err)
	}
	f.Close()

	data, _ := os.ReadFile(path)
	var meta map[string]any
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	if meta["id"] != "meta-test" {
		t.Errorf("id = %v, want meta-test", meta["id"])
	}
	if meta["epic_id"] != "ep-1" {
		t.Errorf("epic_id = %v, want ep-1", meta["epic_id"])
	}
}

func TestExtra_writeEntries_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "entries.jsonl")

	c := &Chain{
		Entries: []ChainEntry{
			{Step: StepResearch, Output: "r.md", Timestamp: time.Now()},
			{Step: StepPlan, Output: "p.md", Timestamp: time.Now()},
		},
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.writeEntries(f); err != nil {
		f.Close()
		t.Fatalf("writeEntries: %v", err)
	}
	f.Close()

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("got %d lines, want 2", len(lines))
	}
}

// --- NewGateChecker (75.0%) ---

func TestExtra_NewGateChecker_ValidDir(t *testing.T) {
	gc, err := NewGateChecker(t.TempDir())
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}
	if gc == nil {
		t.Fatal("GateChecker is nil")
	}
}

// --- checkImplementGate / checkPostMortemGate / findEpic (various) ---
// These depend on external `bd` CLI. We test the parse helper instead.

func TestExtra_parseFirstEpicID_ValidOutput(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want string
	}{
		{"normal", "ep-001  open  My Epic\n", "ep-001"},
		{"with comments", "# epics\nep-002  open  Title\n", "ep-002"},
		{"empty output", "", ""},
		{"only comments", "# nothing\n# here\n", ""},
		{"blank lines", "\n\nep-003\n", "ep-003"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFirstEpicID([]byte(tt.out))
			if got != tt.want {
				t.Errorf("parseFirstEpicID(%q) = %q, want %q", tt.out, got, tt.want)
			}
		})
	}
}

// --- NewLocator (71.4%) ---

func TestExtra_NewLocator_ResolvesAbsPath(t *testing.T) {
	dir := t.TempDir()
	loc, err := NewLocator(dir)
	if err != nil {
		t.Fatalf("NewLocator: %v", err)
	}
	if loc.startDir != dir {
		t.Errorf("startDir = %q, want %q", loc.startDir, dir)
	}
	if loc.home == "" {
		t.Error("home should not be empty")
	}
}

// --- glob (88.9%) ---

func TestExtra_glob_AbsolutePathExists(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.md")
	os.WriteFile(f, []byte("content"), 0o600)

	loc, _ := NewLocator(dir)
	matches, err := loc.glob(dir, f)
	if err != nil {
		t.Fatalf("glob abs: %v", err)
	}
	if len(matches) != 1 || matches[0] != f {
		t.Errorf("glob abs = %v, want [%s]", matches, f)
	}
}

func TestExtra_glob_AbsolutePathNotExists(t *testing.T) {
	loc, _ := NewLocator(t.TempDir())
	matches, err := loc.glob("/tmp", "/nonexistent/file.md")
	if err != nil {
		t.Fatalf("glob abs missing: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected empty, got %v", matches)
	}
}

func TestExtra_glob_RelativePattern(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("a"), 0o600)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o600)

	loc, _ := NewLocator(dir)
	matches, err := loc.glob(dir, "*.md")
	if err != nil {
		t.Fatalf("glob relative: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("got %d matches, want 1", len(matches))
	}
}

// --- GetLocationPaths (90.9%) ---

func TestExtra_GetLocationPaths_ContainsCrewAndTown(t *testing.T) {
	dir := t.TempDir()
	loc, _ := NewLocator(dir)
	paths := loc.GetLocationPaths()

	if _, ok := paths[LocationCrew]; !ok {
		t.Error("missing LocationCrew in paths")
	}
	if _, ok := paths[LocationTown]; !ok {
		t.Error("missing LocationTown in paths")
	}
}

func TestExtra_GetLocationPaths_PluginsFromRig(t *testing.T) {
	// Create a rig-like structure with plugins.
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".beads"), 0o700)
	os.MkdirAll(filepath.Join(dir, "plugins"), 0o700)
	subDir := filepath.Join(dir, "crew", "nami")
	os.MkdirAll(subDir, 0o700)

	loc, _ := NewLocator(subDir)
	paths := loc.GetLocationPaths()

	if p, ok := paths[LocationPlugins]; ok {
		if !strings.Contains(p, "plugins") {
			t.Errorf("plugins path %q should contain 'plugins'", p)
		}
	}
}

// --- parseYAMLFrontMatter (81.8%) ---

func TestExtra_parseYAMLFrontMatter_Valid(t *testing.T) {
	lines := []string{"---", "maturity: candidate", "utility: 0.8", "---", "body text"}
	data, err := parseYAMLFrontMatter(lines)
	if err != nil {
		t.Fatalf("parseYAMLFrontMatter: %v", err)
	}
	if data["maturity"] != "candidate" {
		t.Errorf("maturity = %v, want candidate", data["maturity"])
	}
}

func TestExtra_parseYAMLFrontMatter_Empty(t *testing.T) {
	lines := []string{"---", "---"}
	_, err := parseYAMLFrontMatter(lines)
	if err == nil {
		t.Fatal("expected error for empty front matter")
	}
}

func TestExtra_parseYAMLFrontMatter_NoClosing(t *testing.T) {
	lines := []string{"---", "maturity: candidate"}
	// No closing --- means we read all lines as YAML (valid but unusual).
	data, err := parseYAMLFrontMatter(lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["maturity"] != "candidate" {
		t.Errorf("maturity = %v, want candidate", data["maturity"])
	}
}

// --- applyCandidateTransition (93.3%) ---

func TestExtra_applyCandidateTransition_ImplicitHelpful(t *testing.T) {
	result := &MaturityTransitionResult{
		OldMaturity:  types.MaturityCandidate,
		NewMaturity:  types.MaturityCandidate,
		Utility:      0.8,
		RewardCount:  12,
		HelpfulCount: 2,
		HarmfulCount: 5, // harmful > helpful, but reward >= 10
	}
	applyCandidateTransition(result)
	if result.NewMaturity != types.MaturityEstablished {
		t.Errorf("NewMaturity = %q, want %q (implicit helpful)", result.NewMaturity, types.MaturityEstablished)
	}
	if !result.Transitioned {
		t.Error("Transitioned should be true")
	}
	if !strings.Contains(result.Reason, "implicit helpful signal") {
		t.Errorf("Reason = %q, want mention of implicit helpful signal", result.Reason)
	}
}

func TestExtra_applyCandidateTransition_Demotion(t *testing.T) {
	result := &MaturityTransitionResult{
		OldMaturity: types.MaturityCandidate,
		NewMaturity: types.MaturityCandidate,
		Utility:     0.1,
		RewardCount: 1,
	}
	applyCandidateTransition(result)
	if result.NewMaturity != types.MaturityProvisional {
		t.Errorf("NewMaturity = %q, want %q (demotion)", result.NewMaturity, types.MaturityProvisional)
	}
}

// --- floatFromData (75.0%) ---

func TestExtra_floatFromData_IntValue(t *testing.T) {
	data := map[string]any{"val": 42}
	got := floatFromData(data, "val", 0.0)
	if got != 42.0 {
		t.Errorf("floatFromData(int) = %f, want 42.0", got)
	}
}

func TestExtra_floatFromData_StringFallback(t *testing.T) {
	data := map[string]any{"val": "not a number"}
	got := floatFromData(data, "val", 9.9)
	if got != 9.9 {
		t.Errorf("floatFromData(string) = %f, want 9.9 (default)", got)
	}
}

func TestExtra_floatFromData_MissingKey(t *testing.T) {
	data := map[string]any{}
	got := floatFromData(data, "missing", 1.5)
	if got != 1.5 {
		t.Errorf("floatFromData(missing) = %f, want 1.5", got)
	}
}

// --- GlobLearningFiles (71.4%) ---

func TestExtra_GlobLearningFiles_MixedTypes(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.jsonl"), []byte(`{"id":"a"}`+"\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "b.md"), []byte("---\nid: b\n---\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "c.txt"), []byte("ignored"), 0o600)

	files, err := GlobLearningFiles(dir)
	if err != nil {
		t.Fatalf("GlobLearningFiles: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("got %d files, want 2 (jsonl + md only)", len(files))
	}
}

func TestExtra_GlobLearningFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	files, err := GlobLearningFiles(dir)
	if err != nil {
		t.Fatalf("GlobLearningFiles empty: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("got %d files, want 0", len(files))
	}
}

// --- mergeJSONData (88.9%) ---

func TestExtra_mergeJSONData_InvalidJSON(t *testing.T) {
	_, err := mergeJSONData("not json", map[string]any{"key": "val"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestExtra_mergeJSONData_MergesFields(t *testing.T) {
	input := `{"id":"test","maturity":"provisional"}`
	result, err := mergeJSONData(input, map[string]any{"maturity": "candidate", "new_field": "value"})
	if err != nil {
		t.Fatalf("mergeJSONData: %v", err)
	}
	var data map[string]any
	json.Unmarshal(result, &data)
	if data["maturity"] != "candidate" {
		t.Errorf("maturity = %v, want candidate", data["maturity"])
	}
	if data["new_field"] != "value" {
		t.Errorf("new_field = %v, want value", data["new_field"])
	}
}

// --- updateJSONLFirstLine (92.3%) ---

func TestExtra_updateJSONLFirstLine_UpdatesFirstLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "learning.jsonl")
	os.WriteFile(path, []byte(`{"id":"l1","maturity":"provisional"}`+"\n"+`{"event":"feedback"}`+"\n"), 0o600)

	err := updateJSONLFirstLine(path, map[string]any{"maturity": "candidate"})
	if err != nil {
		t.Fatalf("updateJSONLFirstLine: %v", err)
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(string(data), "\n")
	var first map[string]any
	json.Unmarshal([]byte(lines[0]), &first)
	if first["maturity"] != "candidate" {
		t.Errorf("maturity = %v, want candidate", first["maturity"])
	}
}

func TestExtra_updateJSONLFirstLine_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.jsonl")
	os.WriteFile(path, []byte(""), 0o600)

	err := updateJSONLFirstLine(path, map[string]any{"key": "val"})
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

// --- parseFrontMatterBounds (66.7%) ---

func TestExtra_parseFrontMatterBounds_Valid(t *testing.T) {
	lines := []string{"---", "key: value", "---", "body"}
	idx, err := parseFrontMatterBounds(lines)
	if err != nil {
		t.Fatalf("parseFrontMatterBounds: %v", err)
	}
	if idx != 2 {
		t.Errorf("endIdx = %d, want 2", idx)
	}
}

func TestExtra_parseFrontMatterBounds_NoOpeningDelimiter(t *testing.T) {
	lines := []string{"no front matter", "---"}
	_, err := parseFrontMatterBounds(lines)
	if err == nil {
		t.Fatal("expected error for no opening ---")
	}
	if !strings.Contains(err.Error(), "no front matter") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExtra_parseFrontMatterBounds_NoClosingDelimiter(t *testing.T) {
	lines := []string{"---", "key: value", "no closing"}
	_, err := parseFrontMatterBounds(lines)
	if err == nil {
		t.Fatal("expected error for no closing ---")
	}
	if !strings.Contains(err.Error(), "malformed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExtra_parseFrontMatterBounds_EmptyLines(t *testing.T) {
	_, err := parseFrontMatterBounds([]string{})
	if err == nil {
		t.Fatal("expected error for empty lines")
	}
}

// --- updateMarkdownFrontMatter (78.6%) ---

func TestExtra_updateMarkdownFrontMatter_UpdatesExistingField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "learning.md")
	content := "---\nmaturity: provisional\nutility: 0.5\n---\n# Body\nSome text\n"
	os.WriteFile(path, []byte(content), 0o600)

	err := updateMarkdownFrontMatter(path, map[string]any{"maturity": "candidate"})
	if err != nil {
		t.Fatalf("updateMarkdownFrontMatter: %v", err)
	}

	data, _ := os.ReadFile(path)
	text := string(data)
	if !strings.Contains(text, "maturity: candidate") {
		t.Error("expected maturity to be updated to candidate")
	}
	if !strings.Contains(text, "utility: 0.5") {
		t.Error("expected utility to remain unchanged")
	}
	if !strings.Contains(text, "# Body") {
		t.Error("expected body to be preserved")
	}
}

func TestExtra_updateMarkdownFrontMatter_AddsNewField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "learning.md")
	content := "---\nmaturity: provisional\n---\nBody\n"
	os.WriteFile(path, []byte(content), 0o600)

	err := updateMarkdownFrontMatter(path, map[string]any{"new_field": "new_value"})
	if err != nil {
		t.Fatalf("updateMarkdownFrontMatter: %v", err)
	}

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "new_field: new_value") {
		t.Error("expected new_field to be added")
	}
}

func TestExtra_updateMarkdownFrontMatter_MissingFile(t *testing.T) {
	err := updateMarkdownFrontMatter("/nonexistent/file.md", map[string]any{"key": "val"})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestExtra_updateMarkdownFrontMatter_NoFrontMatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no-fm.md")
	os.WriteFile(path, []byte("# No front matter here\n"), 0o600)

	err := updateMarkdownFrontMatter(path, map[string]any{"key": "val"})
	if err == nil {
		t.Fatal("expected error for no front matter")
	}
}

// --- ScanForMaturityTransitions (90.9%) ---

func TestExtra_ScanForMaturityTransitions_SkipsUnparseable(t *testing.T) {
	dir := t.TempDir()
	// Create a valid learning that would transition.
	os.WriteFile(filepath.Join(dir, "good.jsonl"),
		[]byte(`{"id":"g1","maturity":"provisional","utility":0.8,"reward_count":5}`+"\n"), 0o600)
	// Create an unparseable file.
	os.WriteFile(filepath.Join(dir, "bad.jsonl"), []byte("garbage\n"), 0o600)

	results, err := ScanForMaturityTransitions(dir)
	if err != nil {
		t.Fatalf("ScanForMaturityTransitions: %v", err)
	}
	// The good one should transition provisional -> candidate.
	found := false
	for _, r := range results {
		if r.LearningID == "g1" && r.Transitioned {
			found = true
		}
	}
	if !found {
		t.Error("expected g1 to appear as transitioned")
	}
}

// --- filterLearningsByMaturity (87.5%) ---

func TestExtra_filterLearningsByMaturity_FiltersCorrectly(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.jsonl"),
		[]byte(`{"id":"a","maturity":"candidate"}`+"\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "b.jsonl"),
		[]byte(`{"id":"b","maturity":"provisional"}`+"\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "c.jsonl"),
		[]byte(`{"id":"c","maturity":"candidate"}`+"\n"), 0o600)

	files, err := filterLearningsByMaturity(dir, types.MaturityCandidate)
	if err != nil {
		t.Fatalf("filterLearningsByMaturity: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("got %d files, want 2 candidates", len(files))
	}
}

// --- GetMaturityDistribution (85.7%) ---

func TestExtra_GetMaturityDistribution_AllLevels(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "p.jsonl"),
		[]byte(`{"id":"p","maturity":"provisional"}`+"\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "c.jsonl"),
		[]byte(`{"id":"c","maturity":"candidate"}`+"\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "e.jsonl"),
		[]byte(`{"id":"e","maturity":"established"}`+"\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "a.jsonl"),
		[]byte(`{"id":"a","maturity":"anti-pattern"}`+"\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "u.jsonl"),
		[]byte("garbage\n"), 0o600)

	dist, err := GetMaturityDistribution(dir)
	if err != nil {
		t.Fatalf("GetMaturityDistribution: %v", err)
	}
	if dist.Provisional != 1 {
		t.Errorf("Provisional = %d, want 1", dist.Provisional)
	}
	if dist.Candidate != 1 {
		t.Errorf("Candidate = %d, want 1", dist.Candidate)
	}
	if dist.Established != 1 {
		t.Errorf("Established = %d, want 1", dist.Established)
	}
	if dist.AntiPattern != 1 {
		t.Errorf("AntiPattern = %d, want 1", dist.AntiPattern)
	}
	if dist.Unknown != 1 {
		t.Errorf("Unknown = %d, want 1", dist.Unknown)
	}
	if dist.Total != 5 {
		t.Errorf("Total = %d, want 5", dist.Total)
	}
}

// --- NewValidator (75.0%) ---

func TestExtra_NewValidator_ValidDir(t *testing.T) {
	v, err := NewValidator(t.TempDir())
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}
	if v == nil {
		t.Fatal("Validator is nil")
	}
	if v.metrics == nil {
		t.Fatal("metrics is nil")
	}
}

// --- validateStep (85.7%) ---

func TestExtra_validateStep_UnknownStep(t *testing.T) {
	v, _ := NewValidator(t.TempDir())
	result := &ValidationResult{Valid: true, Issues: []string{}, Warnings: []string{}}

	dir := t.TempDir()
	f := filepath.Join(dir, "artifact.md")
	os.WriteFile(f, []byte("---\nschema_version: 1\n---\n# Content\n"), 0o600)

	v.validateStep(Step("unknown-step"), f, result)
	if len(result.Warnings) == 0 {
		t.Error("expected warning for unknown step")
	}
}

func TestExtra_validateStep_ImplementStep(t *testing.T) {
	v, _ := NewValidator(t.TempDir())
	result := &ValidationResult{Valid: true, Issues: []string{}, Warnings: []string{}}

	dir := t.TempDir()
	f := filepath.Join(dir, "artifact.md")
	os.WriteFile(f, []byte("content"), 0o600)

	v.validateStep(StepImplement, f, result)
	// Should get "no artifact validation rules" warning.
	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "No artifact validation rules") {
			found = true
		}
	}
	if !found {
		t.Error("expected 'No artifact validation rules' warning for implement step")
	}
}

// --- countCitations (83.3%) ---

func TestExtra_countCitations_CountsBacklinks(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.md")
	os.WriteFile(target, []byte("# Target\n"), 0o600)
	// Create files that reference target.
	os.WriteFile(filepath.Join(dir, "ref1.md"), []byte("See target.md for details\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "ref2.md"), []byte("No reference here\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "ref3.md"), []byte("Also see target.md\n"), 0o600)

	v, _ := NewValidator(dir)
	count := v.countCitations(target)
	if count != 2 {
		t.Errorf("countCitations = %d, want 2", count)
	}
}

// --- gatherSessionDirs (75.0%) ---

func TestExtra_gatherSessionDirs_LocalSessionsExist(t *testing.T) {
	dir := t.TempDir()
	sessDir := filepath.Join(dir, ".agents", "ao", "sessions")
	os.MkdirAll(sessDir, 0o700)

	v, _ := NewValidator(dir)
	dirs := v.gatherSessionDirs()
	found := false
	for _, d := range dirs {
		if d == sessDir {
			found = true
		}
	}
	if !found {
		t.Errorf("expected %q in gathered dirs, got %v", sessDir, dirs)
	}
}

// --- countRefsInDir (90.0%) ---

func TestExtra_countRefsInDir_CountsRefs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "s1.jsonl"), []byte(`{"artifact":"target.md"}`+"\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "s2.md"), []byte("References target.md here\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "s3.md"), []byte("No match\n"), 0o600)

	v, _ := NewValidator(t.TempDir())
	seen := make(map[string]bool)
	count := v.countRefsInDir(dir, "target.md", seen)
	if count != 2 {
		t.Errorf("countRefsInDir = %d, want 2", count)
	}
}

// --- ValidateArtifactPath (85.7%) ---

func TestExtra_ValidateArtifactPath_RelativePath(t *testing.T) {
	err := ValidateArtifactPath("relative/path.md")
	if err == nil {
		t.Fatal("expected error for relative path")
	}
	if !strings.Contains(err.Error(), "must be absolute") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExtra_ValidateArtifactPath_TildePath(t *testing.T) {
	// Tilde paths are not absolute, so they fail the absolute check first.
	// Test with an absolute-looking tilde path to exercise tilde check.
	err := ValidateArtifactPath("~/path.md")
	if err == nil {
		t.Fatal("expected error for tilde path")
	}
	// The path is not absolute, so the "must be absolute" error fires.
	if !strings.Contains(err.Error(), "must be absolute") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExtra_ValidateArtifactPath_EmptyIsValid(t *testing.T) {
	err := ValidateArtifactPath("")
	if err != nil {
		t.Errorf("empty path should be valid, got: %v", err)
	}
}

// --- ValidateCloseReason (90.9%) ---

func TestExtra_ValidateCloseReason_RelativePatterns(t *testing.T) {
	issues := ValidateCloseReason("See ./relative/path for details")
	if len(issues) == 0 {
		t.Error("expected issue for ./ relative path")
	}
}

func TestExtra_ValidateCloseReason_TildePattern(t *testing.T) {
	issues := ValidateCloseReason("Artifact: ~/some/path.md")
	// Should catch both the extracted path (tilde) and the relative pattern.
	if len(issues) == 0 {
		t.Error("expected issues for ~/ path")
	}
}

// --- RecordCitation (88.2%) ---

func TestExtra_RecordCitation_WritesAndLoads(t *testing.T) {
	dir := t.TempDir()

	event := types.CitationEvent{
		ArtifactPath: "learnings/test.md",
		SessionID:    "sess-001",
		CitationType: "reference",
	}

	if err := RecordCitation(dir, event); err != nil {
		t.Fatalf("RecordCitation: %v", err)
	}

	citations, err := LoadCitations(dir)
	if err != nil {
		t.Fatalf("LoadCitations: %v", err)
	}
	if len(citations) != 1 {
		t.Fatalf("got %d citations, want 1", len(citations))
	}
	if citations[0].SessionID != "sess-001" {
		t.Errorf("SessionID = %q, want %q", citations[0].SessionID, "sess-001")
	}
}

func TestExtra_RecordCitation_BadDir(t *testing.T) {
	event := types.CitationEvent{ArtifactPath: "test.md"}
	err := RecordCitation("/dev/null/impossible", event)
	if err == nil {
		t.Fatal("expected error for impossible base dir")
	}
}

// --- GetCitationsSince (87.5%) ---

func TestExtra_GetCitationsSince_FiltersCorrectly(t *testing.T) {
	dir := t.TempDir()

	old := types.CitationEvent{
		ArtifactPath: "old.md",
		SessionID:    "s1",
		CitedAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	recent := types.CitationEvent{
		ArtifactPath: "recent.md",
		SessionID:    "s2",
		CitedAt:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	RecordCitation(dir, old)
	RecordCitation(dir, recent)

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	filtered, err := GetCitationsSince(dir, since)
	if err != nil {
		t.Fatalf("GetCitationsSince: %v", err)
	}
	if len(filtered) != 1 {
		t.Errorf("got %d citations, want 1 (only recent)", len(filtered))
	}
}

// --- GetUniqueCitedArtifacts (91.7%) ---

func TestExtra_GetUniqueCitedArtifacts_DeduplicatesAndFilters(t *testing.T) {
	dir := t.TempDir()

	e1 := types.CitationEvent{ArtifactPath: "a.md", SessionID: "s1", CitedAt: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)}
	e2 := types.CitationEvent{ArtifactPath: "a.md", SessionID: "s2", CitedAt: time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)}
	e3 := types.CitationEvent{ArtifactPath: "b.md", SessionID: "s3", CitedAt: time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)}
	e4 := types.CitationEvent{ArtifactPath: "c.md", SessionID: "s4", CitedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)} // outside range

	for _, e := range []types.CitationEvent{e1, e2, e3, e4} {
		RecordCitation(dir, e)
	}

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	unique, err := GetUniqueCitedArtifacts(dir, since, until)
	if err != nil {
		t.Fatalf("GetUniqueCitedArtifacts: %v", err)
	}
	if len(unique) != 2 {
		t.Errorf("got %d unique artifacts, want 2 (a.md deduped, c.md outside range)", len(unique))
	}
}

// --- GetCitationsForSession (87.5%) ---

func TestExtra_GetCitationsForSession_FiltersCorrectly(t *testing.T) {
	dir := t.TempDir()

	RecordCitation(dir, types.CitationEvent{ArtifactPath: "a.md", SessionID: "target-sess", CitedAt: time.Now()})
	RecordCitation(dir, types.CitationEvent{ArtifactPath: "b.md", SessionID: "other-sess", CitedAt: time.Now()})
	RecordCitation(dir, types.CitationEvent{ArtifactPath: "c.md", SessionID: "target-sess", CitedAt: time.Now()})

	filtered, err := GetCitationsForSession(dir, "target-sess")
	if err != nil {
		t.Fatalf("GetCitationsForSession: %v", err)
	}
	if len(filtered) != 2 {
		t.Errorf("got %d citations, want 2 for target-sess", len(filtered))
	}
}
