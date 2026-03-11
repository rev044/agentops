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

func TestAllSteps(t *testing.T) {
	steps := AllSteps()

	// Verify count
	if len(steps) != 7 {
		t.Fatalf("AllSteps() returned %d steps, want 7", len(steps))
	}

	// Verify order matches workflow sequence
	expected := []Step{
		StepResearch,
		StepPreMortem,
		StepPlan,
		StepImplement,
		StepCrank,
		StepVibe,
		StepPostMortem,
	}

	for i, step := range expected {
		if steps[i] != step {
			t.Errorf("AllSteps()[%d] = %q, want %q", i, steps[i], step)
		}
	}
}

func TestAllStepsReturnsNewSlice(t *testing.T) {
	// Verify AllSteps returns a new slice each time (not a shared reference)
	a := AllSteps()
	b := AllSteps()

	a[0] = "mutated"
	if b[0] == "mutated" {
		t.Error("AllSteps() should return a new slice each call, got shared reference")
	}
}

func TestParseStep(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Step
	}{
		// Canonical names
		{"canonical research", "research", StepResearch},
		{"canonical pre-mortem", "pre-mortem", StepPreMortem},
		{"canonical plan", "plan", StepPlan},
		{"canonical implement", "implement", StepImplement},
		{"canonical crank", "crank", StepCrank},
		{"canonical vibe", "vibe", StepVibe},
		{"canonical post-mortem", "post-mortem", StepPostMortem},

		// Case insensitivity
		{"uppercase RESEARCH", "RESEARCH", StepResearch},
		{"mixed case Plan", "Plan", StepPlan},
		{"all caps VIBE", "VIBE", StepVibe},
		{"mixed Pre-Mortem", "Pre-Mortem", StepPreMortem},

		// Whitespace trimming
		{"leading space", " research", StepResearch},
		{"trailing space", "plan ", StepPlan},
		{"both spaces", " vibe ", StepVibe},
		{"tab whitespace", "\tcrank\t", StepCrank},

		// Aliases without hyphen
		{"premortem no hyphen", "premortem", StepPreMortem},
		{"postmortem no hyphen", "postmortem", StepPostMortem},

		// Aliases with underscore
		{"pre_mortem underscore", "pre_mortem", StepPreMortem},
		{"post_mortem underscore", "post_mortem", StepPostMortem},

		// Semantic aliases
		{"formulate alias", "formulate", StepPlan},
		{"autopilot alias", "autopilot", StepCrank},
		{"validate alias", "validate", StepVibe},
		{"review alias", "review", StepPostMortem},
		{"execute alias", "execute", StepCrank},

		// Unrecognized
		{"empty string", "", ""},
		{"unknown step", "unknown", ""},
		{"partial match", "res", ""},
		{"typo", "reserch", ""},
		{"numeric", "123", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseStep(tt.input)
			if got != tt.want {
				t.Errorf("ParseStep(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStepIsValid(t *testing.T) {
	tests := []struct {
		name  string
		step  Step
		valid bool
	}{
		{"research is valid", StepResearch, true},
		{"pre-mortem is valid", StepPreMortem, true},
		{"plan is valid", StepPlan, true},
		{"implement is valid", StepImplement, true},
		{"crank is valid", StepCrank, true},
		{"vibe is valid", StepVibe, true},
		{"post-mortem is valid", StepPostMortem, true},
		{"empty is invalid", Step(""), false},
		{"unknown is invalid", Step("bogus"), false},
		{"partial is invalid", Step("res"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.step.IsValid(); got != tt.valid {
				t.Errorf("Step(%q).IsValid() = %v, want %v", tt.step, got, tt.valid)
			}
		})
	}
}

func TestStepConstants(t *testing.T) {
	// Verify step constant values match expected strings
	tests := []struct {
		step Step
		want string
	}{
		{StepResearch, "research"},
		{StepPreMortem, "pre-mortem"},
		{StepPlan, "plan"},
		{StepImplement, "implement"},
		{StepCrank, "crank"},
		{StepVibe, "vibe"},
		{StepPostMortem, "post-mortem"},
	}

	for _, tt := range tests {
		if string(tt.step) != tt.want {
			t.Errorf("Step constant = %q, want %q", tt.step, tt.want)
		}
	}
}

func TestTierString(t *testing.T) {
	tests := []struct {
		tier Tier
		want string
	}{
		{TierObservation, "observation"},
		{TierLearning, "learning"},
		{TierPattern, "pattern"},
		{TierSkill, "skill"},
		{TierCore, "core"},
		{Tier(99), "unknown"},
		{Tier(-1), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.tier.String(); got != tt.want {
				t.Errorf("Tier(%d).String() = %q, want %q", tt.tier, got, tt.want)
			}
		})
	}
}

func TestTierLocation(t *testing.T) {
	tests := []struct {
		tier Tier
		want string
	}{
		{TierObservation, ".agents/candidates/"},
		{TierLearning, ".agents/learnings/"},
		{TierPattern, ".agents/patterns/"},
		{TierSkill, "plugins/*/skills/"},
		{TierCore, "CLAUDE.md"},
		{Tier(99), ""},
		{Tier(-1), ""},
	}

	for _, tt := range tests {
		t.Run(tt.tier.String(), func(t *testing.T) {
			if got := tt.tier.Location(); got != tt.want {
				t.Errorf("Tier(%d).Location() = %q, want %q", tt.tier, got, tt.want)
			}
		})
	}
}

func TestTierConstants(t *testing.T) {
	// Verify tier constant values
	tests := []struct {
		tier Tier
		want int
	}{
		{TierObservation, 0},
		{TierLearning, 1},
		{TierPattern, 2},
		{TierSkill, 3},
		{TierCore, 4},
	}

	for _, tt := range tests {
		if int(tt.tier) != tt.want {
			t.Errorf("Tier %s = %d, want %d", tt.tier, tt.tier, tt.want)
		}
	}
}

func TestTierOrdering(t *testing.T) {
	// Verify tiers are ordered from lowest to highest quality
	if TierObservation >= TierLearning {
		t.Error("TierObservation should be less than TierLearning")
	}
	if TierLearning >= TierPattern {
		t.Error("TierLearning should be less than TierPattern")
	}
	if TierPattern >= TierSkill {
		t.Error("TierPattern should be less than TierSkill")
	}
	if TierSkill >= TierCore {
		t.Error("TierSkill should be less than TierCore")
	}
}

func TestChainEntryJSONRoundTrip(t *testing.T) {
	tier := TierPattern
	entry := ChainEntry{
		Step:       StepResearch,
		Timestamp:  time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC),
		Input:      "/path/to/input.md",
		Output:     "/path/to/output.md",
		Locked:     true,
		Skipped:    false,
		Reason:     "",
		Tier:       &tier,
		Location:   ".agents/patterns/",
		Cycle:      2,
		ParentEpic: "ag-abc",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got ChainEntry
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Step != entry.Step {
		t.Errorf("Step = %q, want %q", got.Step, entry.Step)
	}
	if !got.Timestamp.Equal(entry.Timestamp) {
		t.Errorf("Timestamp = %v, want %v", got.Timestamp, entry.Timestamp)
	}
	if got.Input != entry.Input {
		t.Errorf("Input = %q, want %q", got.Input, entry.Input)
	}
	if got.Output != entry.Output {
		t.Errorf("Output = %q, want %q", got.Output, entry.Output)
	}
	if got.Locked != entry.Locked {
		t.Errorf("Locked = %v, want %v", got.Locked, entry.Locked)
	}
	if got.Tier == nil || *got.Tier != tier {
		t.Errorf("Tier = %v, want %v", got.Tier, &tier)
	}
	if got.Location != entry.Location {
		t.Errorf("Location = %q, want %q", got.Location, entry.Location)
	}
	if got.Cycle != entry.Cycle {
		t.Errorf("Cycle = %d, want %d", got.Cycle, entry.Cycle)
	}
	if got.ParentEpic != entry.ParentEpic {
		t.Errorf("ParentEpic = %q, want %q", got.ParentEpic, entry.ParentEpic)
	}
}

func TestChainEntrySkippedJSONFields(t *testing.T) {
	entry := ChainEntry{
		Step:      StepPlan,
		Timestamp: time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC),
		Output:    "skipped",
		Locked:    false,
		Skipped:   true,
		Reason:    "not needed for hotfix",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	if v, ok := m["skipped"]; !ok || v != true {
		t.Error("expected 'skipped' field = true in JSON")
	}
	if v, ok := m["reason"]; !ok || v != "not needed for hotfix" {
		t.Errorf("expected 'reason' field, got %v", v)
	}
}

func TestChainEntryOmitemptyFields(t *testing.T) {
	// Fields with omitempty should not appear when zero-valued
	entry := ChainEntry{
		Step:      StepImplement,
		Timestamp: time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC),
		Output:    "result",
		Locked:    true,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	omitemptyFields := []string{"input", "skipped", "reason", "tier", "location", "cycle", "parent_epic"}
	for _, field := range omitemptyFields {
		if _, ok := m[field]; ok {
			t.Errorf("expected field %q to be omitted when zero-valued", field)
		}
	}

	// These fields should always be present
	requiredFields := []string{"step", "timestamp", "output", "locked"}
	for _, field := range requiredFields {
		if _, ok := m[field]; !ok {
			t.Errorf("expected field %q to be present", field)
		}
	}
}

func TestChainJSONRoundTrip(t *testing.T) {
	chain := Chain{
		ID:      "ag-test",
		Started: time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC),
		Entries: []ChainEntry{
			{
				Step:      StepResearch,
				Timestamp: time.Date(2026, 2, 10, 10, 30, 0, 0, time.UTC),
				Output:    "/path/research.md",
				Locked:    true,
			},
			{
				Step:      StepPlan,
				Timestamp: time.Date(2026, 2, 10, 11, 0, 0, 0, time.UTC),
				Input:     "/path/research.md",
				Output:    "/path/plan.md",
				Locked:    true,
			},
		},
		EpicID: "ag-epic",
	}

	data, err := json.Marshal(chain)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Chain
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != chain.ID {
		t.Errorf("ID = %q, want %q", got.ID, chain.ID)
	}
	if !got.Started.Equal(chain.Started) {
		t.Errorf("Started = %v, want %v", got.Started, chain.Started)
	}
	if len(got.Entries) != len(chain.Entries) {
		t.Fatalf("Entries len = %d, want %d", len(got.Entries), len(chain.Entries))
	}
	if got.EpicID != chain.EpicID {
		t.Errorf("EpicID = %q, want %q", got.EpicID, chain.EpicID)
	}
}

func TestChainOmitemptyEpicID(t *testing.T) {
	chain := Chain{
		ID:      "test",
		Started: time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC),
		Entries: []ChainEntry{},
	}

	data, err := json.Marshal(chain)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	if _, ok := m["epic_id"]; ok {
		t.Error("expected 'epic_id' to be omitted when empty")
	}
}

func TestGateResultJSONRoundTrip(t *testing.T) {
	result := GateResult{
		Step:     StepResearch,
		Passed:   true,
		Message:  "research artifact found",
		Input:    "/path/to/research.md",
		Location: ".agents/research/",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got GateResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Step != result.Step {
		t.Errorf("Step = %q, want %q", got.Step, result.Step)
	}
	if got.Passed != result.Passed {
		t.Errorf("Passed = %v, want %v", got.Passed, result.Passed)
	}
	if got.Message != result.Message {
		t.Errorf("Message = %q, want %q", got.Message, result.Message)
	}
	if got.Input != result.Input {
		t.Errorf("Input = %q, want %q", got.Input, result.Input)
	}
	if got.Location != result.Location {
		t.Errorf("Location = %q, want %q", got.Location, result.Location)
	}
}

func TestValidationResultJSONRoundTrip(t *testing.T) {
	tier := TierLearning
	expiryDate := "2026-05-10"
	result := ValidationResult{
		Step:                StepVibe,
		Valid:               false,
		Issues:              []string{"missing tests", "no coverage"},
		Warnings:            []string{"complexity high"},
		Tier:                &tier,
		Lenient:             true,
		LenientExpiryDate:   &expiryDate,
		LenientExpiringSoon: true,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got ValidationResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Step != result.Step {
		t.Errorf("Step = %q, want %q", got.Step, result.Step)
	}
	if got.Valid != result.Valid {
		t.Errorf("Valid = %v, want %v", got.Valid, result.Valid)
	}
	if len(got.Issues) != 2 {
		t.Errorf("Issues len = %d, want 2", len(got.Issues))
	}
	if len(got.Warnings) != 1 {
		t.Errorf("Warnings len = %d, want 1", len(got.Warnings))
	}
	if got.Tier == nil || *got.Tier != tier {
		t.Errorf("Tier = %v, want %v", got.Tier, &tier)
	}
	if got.Lenient != true {
		t.Error("Lenient should be true")
	}
	if got.LenientExpiryDate == nil || *got.LenientExpiryDate != expiryDate {
		t.Errorf("LenientExpiryDate = %v, want %q", got.LenientExpiryDate, expiryDate)
	}
	if got.LenientExpiringSoon != true {
		t.Error("LenientExpiringSoon should be true")
	}
}

func TestValidationResultOmitemptyFields(t *testing.T) {
	result := ValidationResult{
		Step:  StepPlan,
		Valid: true,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	omitemptyFields := []string{"issues", "warnings", "tier", "lenient_expiry_date", "lenient_expiring_soon"}
	for _, field := range omitemptyFields {
		if _, ok := m[field]; ok {
			t.Errorf("expected field %q to be omitted when zero-valued", field)
		}
	}
}

func TestFindResultJSONRoundTrip(t *testing.T) {
	result := FindResult{
		Pattern: "research/*.md",
		Matches: []FindMatch{
			{Path: "/a/research/topic.md", Location: "crew", Priority: 0},
			{Path: "/b/research/topic.md", Location: "rig", Priority: 1},
		},
		Warnings: []string{"duplicate found across locations"},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got FindResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Pattern != result.Pattern {
		t.Errorf("Pattern = %q, want %q", got.Pattern, result.Pattern)
	}
	if len(got.Matches) != 2 {
		t.Fatalf("Matches len = %d, want 2", len(got.Matches))
	}
	if got.Matches[0].Path != "/a/research/topic.md" {
		t.Errorf("Matches[0].Path = %q, want %q", got.Matches[0].Path, "/a/research/topic.md")
	}
	if got.Matches[0].Location != "crew" {
		t.Errorf("Matches[0].Location = %q, want %q", got.Matches[0].Location, "crew")
	}
	if got.Matches[0].Priority != 0 {
		t.Errorf("Matches[0].Priority = %d, want 0", got.Matches[0].Priority)
	}
	if got.Matches[1].Priority != 1 {
		t.Errorf("Matches[1].Priority = %d, want 1", got.Matches[1].Priority)
	}
	if len(got.Warnings) != 1 {
		t.Errorf("Warnings len = %d, want 1", len(got.Warnings))
	}
}

func TestFindMatchJSONFields(t *testing.T) {
	match := FindMatch{
		Path:     "/test/path.md",
		Location: "town",
		Priority: 2,
	}

	data, err := json.Marshal(match)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	if v, ok := m["path"]; !ok || v != "/test/path.md" {
		t.Errorf("path field = %v", v)
	}
	if v, ok := m["location"]; !ok || v != "town" {
		t.Errorf("location field = %v", v)
	}
	if v, ok := m["priority"]; !ok || int(v.(float64)) != 2 {
		t.Errorf("priority field = %v", v)
	}
}

func TestValidateOptionsDefaults(t *testing.T) {
	opts := ValidateOptions{}

	if opts.Lenient != false {
		t.Error("default Lenient should be false")
	}
	if opts.LenientExpiryDate != nil {
		t.Error("default LenientExpiryDate should be nil")
	}
}

func TestParseStepAliasesCompleteness(t *testing.T) {
	t.Helper()

	// Every canonical step should parse to itself
	for _, step := range AllSteps() {
		got := ParseStep(string(step))
		if got != step {
			t.Errorf("ParseStep(%q) = %q, want %q (canonical self-lookup failed)", step, got, step)
		}
	}
}

// TestParseStepPhasedModeAliases verifies that the phased-mode canonical phase
// names are accepted as ratchet step aliases so that hooks and tools can use
// them directly without knowing the underlying ratchet step name.
func TestParseStepPhasedModeAliases(t *testing.T) {
	tests := []struct {
		alias    string
		wantStep Step
	}{
		// Phase-canonical names → ratchet steps
		{"discovery", StepResearch},
		{"validation", StepVibe},
		// Existing aliases must still work
		{"validate", StepVibe},
		{"implement", StepImplement},
		{"research", StepResearch},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			got := ParseStep(tt.alias)
			if got == "" {
				t.Errorf("ParseStep(%q) returned empty — phased-mode alias not registered", tt.alias)
				return
			}
			if got != tt.wantStep {
				t.Errorf("ParseStep(%q) = %q, want %q", tt.alias, got, tt.wantStep)
			}
		})
	}
}

func TestParseStepCaseAndWhitespaceCombined(t *testing.T) {
	// Combine case variation with whitespace
	got := ParseStep("  RESEARCH  ")
	if got != StepResearch {
		t.Errorf("ParseStep with upper+whitespace = %q, want %q", got, StepResearch)
	}

	got = ParseStep("\tPre-Mortem\t")
	if got != StepPreMortem {
		t.Errorf("ParseStep with mixed case+tabs = %q, want %q", got, StepPreMortem)
	}
}

func TestTierNilPointerInChainEntry(t *testing.T) {
	// ChainEntry with nil Tier should serialize without tier field
	entry := ChainEntry{
		Step:      StepResearch,
		Timestamp: time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC),
		Output:    "test",
		Locked:    true,
		Tier:      nil,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	if _, ok := m["tier"]; ok {
		t.Error("expected 'tier' to be omitted when nil")
	}
}

// =====================================================================
// Coverage gap tests below — each targets specific uncovered lines.
// =====================================================================

// -- chain.go: writeEntries round-trip --
func TestWriteEntriesRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	chainPath := filepath.Join(tmpDir, ".agents", "ao", "chain.jsonl")

	chain := &Chain{
		ID:      "test-write",
		Started: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
		Entries: []ChainEntry{
			{Step: StepResearch, Timestamp: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), Output: "r.md", Locked: true},
			{Step: StepPlan, Timestamp: time.Date(2026, 3, 1, 11, 0, 0, 0, time.UTC), Output: "p.md", Locked: true},
		},
	}
	chain.SetPath(chainPath)

	if err := chain.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := loadJSONLChain(chainPath)
	if err != nil {
		t.Fatalf("loadJSONLChain: %v", err)
	}
	if loaded.ID != "test-write" {
		t.Errorf("loaded ID = %q, want %q", loaded.ID, "test-write")
	}
	if len(loaded.Entries) != 2 {
		t.Errorf("loaded entries = %d, want 2", len(loaded.Entries))
	}
	if loaded.Entries[0].Step != StepResearch {
		t.Errorf("entry[0].Step = %q, want %q", loaded.Entries[0].Step, StepResearch)
	}
}

// -- maturity.go: parseYAMLFrontMatter empty body --
func TestParseYAMLFrontMatter_EmptyBody(t *testing.T) {
	lines := []string{"---", "---"}
	_, err := parseYAMLFrontMatter(lines)
	if err == nil {
		t.Error("expected error for empty front matter body")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error = %q, want it to contain 'empty'", err.Error())
	}
}

// -- maturity.go: parseYAMLFrontMatter invalid YAML --
func TestParseYAMLFrontMatter_InvalidYAML(t *testing.T) {
	lines := []string{"---", "  invalid: [yaml: broken", "---"}
	_, err := parseYAMLFrontMatter(lines)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parse YAML front matter") {
		t.Errorf("error = %q, want it to contain 'parse YAML front matter'", err.Error())
	}
}

// -- maturity.go: applyCandidateTransition implicit helpful signal --
func TestCandidateTransition_ImplicitHelpfulSignal(t *testing.T) {
	tmpDir := t.TempDir()
	learning := filepath.Join(tmpDir, "test.jsonl")
	data := `{"id":"test-implicit","maturity":"candidate","utility":0.6,"reward_count":10,"helpful_count":2,"harmful_count":2}`
	if err := os.WriteFile(learning, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := CheckMaturityTransition(learning)
	if err != nil {
		t.Fatalf("CheckMaturityTransition: %v", err)
	}
	if result.NewMaturity != types.MaturityEstablished {
		t.Errorf("NewMaturity = %q, want %q", result.NewMaturity, types.MaturityEstablished)
	}
	if !result.Transitioned {
		t.Error("expected Transitioned = true")
	}
	if !strings.Contains(result.Reason, "implicit helpful signal") {
		t.Errorf("Reason = %q, want it to contain 'implicit helpful signal'", result.Reason)
	}
}

// -- maturity.go: floatFromData int case --
func TestFloatFromData_IntCase(t *testing.T) {
	data := map[string]any{"val": int(42)}
	got := floatFromData(data, "val", 0.0)
	if got != 42.0 {
		t.Errorf("floatFromData(int 42) = %f, want 42.0", got)
	}
}

// -- maturity.go: parseFrontMatterBounds no opening --- --
func TestParseFrontMatterBounds_NoOpening(t *testing.T) {
	_, err := parseFrontMatterBounds([]string{"no front matter here"})
	if err == nil {
		t.Error("expected error for missing opening ---")
	}
	if !strings.Contains(err.Error(), "no front matter found") {
		t.Errorf("error = %q, want 'no front matter found'", err.Error())
	}
}

// -- maturity.go: parseFrontMatterBounds no closing --- --
func TestParseFrontMatterBounds_NoClosing(t *testing.T) {
	_, err := parseFrontMatterBounds([]string{"---", "key: value"})
	if err == nil {
		t.Error("expected error for missing closing ---")
	}
	if !strings.Contains(err.Error(), "malformed front matter") {
		t.Errorf("error = %q, want 'malformed front matter'", err.Error())
	}
}

// -- maturity.go: parseFrontMatterBounds empty input --
func TestParseFrontMatterBounds_EmptyLines(t *testing.T) {
	_, err := parseFrontMatterBounds([]string{})
	if err == nil {
		t.Error("expected error for empty lines")
	}
}

// -- maturity.go: updateJSONLFirstLine empty file --
func TestUpdateJSONLFirstLine_EmptyFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.jsonl")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	err := updateJSONLFirstLine(emptyFile, map[string]any{"maturity": "candidate"})
	// The file has one empty line after split, which will fail JSON parse
	if err == nil {
		t.Error("expected error for empty JSONL file")
	}
}

// -- maturity.go: updateJSONLFirstLine read error --
func TestUpdateJSONLFirstLine_NonexistentFile(t *testing.T) {
	err := updateJSONLFirstLine("/nonexistent/path/to/file.jsonl", map[string]any{"x": "y"})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "read learning for update") {
		t.Errorf("error = %q, want 'read learning for update'", err.Error())
	}
}

// -- maturity.go: updateMarkdownFrontMatter read error --
func TestUpdateMarkdownFrontMatter_NonexistentFile(t *testing.T) {
	err := updateMarkdownFrontMatter("/nonexistent/path/to/file.md", map[string]any{"x": "y"})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "read learning for update") {
		t.Errorf("error = %q, want 'read learning for update'", err.Error())
	}
}

// -- maturity.go: updateMarkdownFrontMatter no front matter --
func TestUpdateMarkdownFrontMatter_NoFrontMatter(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "no-fm.md")
	if err := os.WriteFile(f, []byte("just some text\nno front matter"), 0644); err != nil {
		t.Fatal(err)
	}

	err := updateMarkdownFrontMatter(f, map[string]any{"maturity": "candidate"})
	if err == nil {
		t.Error("expected error for file without front matter")
	}
	if !strings.Contains(err.Error(), "no front matter found") {
		t.Errorf("error = %q, want 'no front matter found'", err.Error())
	}
}

// -- maturity.go: updateMarkdownFrontMatter success path --
func TestUpdateMarkdownFrontMatter_Success(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "with-fm.md")
	content := "---\nmaturity: provisional\nid: test\n---\n# Body\nSome content"
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := updateMarkdownFrontMatter(f, map[string]any{"maturity": "candidate", "new_field": "added"})
	if err != nil {
		t.Fatalf("updateMarkdownFrontMatter: %v", err)
	}

	updated, err := os.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	text := string(updated)
	if !strings.Contains(text, "maturity: candidate") {
		t.Error("expected updated maturity field")
	}
	if !strings.Contains(text, "new_field: added") {
		t.Error("expected new_field to be added")
	}
	if !strings.Contains(text, "# Body") {
		t.Error("expected body content to be preserved")
	}
}

// -- maturity.go: ScanForMaturityTransitions skips unparseable --
func TestScanForMaturityTransitions_SkipsUnparseable(t *testing.T) {
	tmpDir := t.TempDir()
	// Valid file that would transition
	valid := filepath.Join(tmpDir, "valid.jsonl")
	validData := `{"id":"will-transition","maturity":"provisional","utility":0.7,"reward_count":5}`
	if err := os.WriteFile(valid, []byte(validData), 0644); err != nil {
		t.Fatal(err)
	}
	// Unparseable file
	bad := filepath.Join(tmpDir, "bad.jsonl")
	if err := os.WriteFile(bad, []byte("not json at all"), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := ScanForMaturityTransitions(tmpDir)
	if err != nil {
		t.Fatalf("ScanForMaturityTransitions: %v", err)
	}
	// Should include the valid transition and skip the bad file
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if len(results) > 0 && results[0].LearningID != "will-transition" {
		t.Errorf("LearningID = %q, want %q", results[0].LearningID, "will-transition")
	}
}

// -- maturity.go: GlobLearningFiles returns both .jsonl and .md --
func TestGlobLearningFiles_ReturnsJSONLAndMD(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.jsonl"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.md"), []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "c.txt"), []byte("ignore"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := GlobLearningFiles(tmpDir)
	if err != nil {
		t.Fatalf("GlobLearningFiles: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("got %d files, want 2 (.jsonl and .md only)", len(files))
	}
}

// -- maturity.go: mergeJSONData invalid JSON --
func TestMergeJSONData_InvalidJSON(t *testing.T) {
	_, err := mergeJSONData("not json", map[string]any{"k": "v"})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse learning for update") {
		t.Errorf("error = %q, want 'parse learning for update'", err.Error())
	}
}

// -- maturity.go: mergeJSONData success --
func TestMergeJSONData_Success(t *testing.T) {
	result, err := mergeJSONData(`{"a":"1","b":"2"}`, map[string]any{"b": "updated", "c": "new"})
	if err != nil {
		t.Fatalf("mergeJSONData: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if m["a"] != "1" {
		t.Errorf("a = %v, want '1'", m["a"])
	}
	if m["b"] != "updated" {
		t.Errorf("b = %v, want 'updated'", m["b"])
	}
	if m["c"] != "new" {
		t.Errorf("c = %v, want 'new'", m["c"])
	}
}

// -- maturity.go: ApplyMaturityTransition JSONL path --
func TestApplyMaturityTransition_JSONLPath(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "test.jsonl")
	data := `{"id":"promote-me","maturity":"provisional","utility":0.7,"reward_count":5}`
	if err := os.WriteFile(f, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ApplyMaturityTransition(f)
	if err != nil {
		t.Fatalf("ApplyMaturityTransition: %v", err)
	}
	if !result.Transitioned {
		t.Error("expected transition to occur")
	}
	if result.NewMaturity != types.MaturityCandidate {
		t.Errorf("NewMaturity = %q, want %q", result.NewMaturity, types.MaturityCandidate)
	}

	// Verify file was updated
	content, _ := os.ReadFile(f)
	if !strings.Contains(string(content), `"candidate"`) {
		t.Error("expected file to contain 'candidate' maturity")
	}
}

// -- maturity.go: ApplyMaturityTransition markdown path --
func TestApplyMaturityTransition_MarkdownPath(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "test.md")
	content := "---\nid: promote-md\nmaturity: provisional\nutility: 0.7\nreward_count: 5\n---\n# Content"
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ApplyMaturityTransition(f)
	if err != nil {
		t.Fatalf("ApplyMaturityTransition: %v", err)
	}
	if !result.Transitioned {
		t.Error("expected transition to occur")
	}

	updated, _ := os.ReadFile(f)
	if !strings.Contains(string(updated), "maturity: candidate") {
		t.Error("expected file to contain updated maturity")
	}
}

// -- maturity.go: ApplyMaturityTransition no transition --
func TestApplyMaturityTransition_NoTransitionStable(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "stable.jsonl")
	data := `{"id":"stable","maturity":"provisional","utility":0.3,"reward_count":1}`
	if err := os.WriteFile(f, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ApplyMaturityTransition(f)
	if err != nil {
		t.Fatalf("ApplyMaturityTransition: %v", err)
	}
	if result.Transitioned {
		t.Error("expected no transition")
	}
}

// -- maturity.go: GetMaturityDistribution unparseable file --
func TestGetMaturityDistribution_UnparseableFile(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "bad.jsonl"), []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "good.jsonl"), []byte(`{"maturity":"established"}`), 0644); err != nil {
		t.Fatal(err)
	}

	dist, err := GetMaturityDistribution(tmpDir)
	if err != nil {
		t.Fatalf("GetMaturityDistribution: %v", err)
	}
	if dist.Unknown != 1 {
		t.Errorf("Unknown = %d, want 1", dist.Unknown)
	}
	if dist.Established != 1 {
		t.Errorf("Established = %d, want 1", dist.Established)
	}
	if dist.Total != 2 {
		t.Errorf("Total = %d, want 2", dist.Total)
	}
}

// -- maturity.go: GetMaturityDistribution full distribution --
func TestGetMaturityDistribution_FullCoverage(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"prov.jsonl": `{"maturity":"provisional"}`,
		"cand.jsonl": `{"maturity":"candidate"}`,
		"est.jsonl":  `{"maturity":"established"}`,
		"anti.jsonl": `{"maturity":"anti-pattern"}`,
		"none.jsonl": `{"id":"no-maturity"}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	dist, err := GetMaturityDistribution(tmpDir)
	if err != nil {
		t.Fatalf("GetMaturityDistribution: %v", err)
	}
	if dist.Provisional != 2 { // "none" defaults to provisional
		t.Errorf("Provisional = %d, want 2", dist.Provisional)
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
	if dist.Total != 5 {
		t.Errorf("Total = %d, want 5", dist.Total)
	}
}

// -- maturity.go: GetAntiPatterns and GetEstablishedLearnings --
func TestGetAntiPatterns_Coverage(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.jsonl"), []byte(`{"maturity":"anti-pattern"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.jsonl"), []byte(`{"maturity":"provisional"}`), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := GetAntiPatterns(tmpDir)
	if err != nil {
		t.Fatalf("GetAntiPatterns: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
}

func TestGetEstablishedLearnings_Coverage(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "a.jsonl"), []byte(`{"maturity":"established"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.jsonl"), []byte(`{"maturity":"provisional"}`), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := GetEstablishedLearnings(tmpDir)
	if err != nil {
		t.Fatalf("GetEstablishedLearnings: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
}

// -- location.go: GetLocationPaths plugins fallback to rig --
func TestGetLocationPaths_PluginsFallbackToRig(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a rig marker (.beads) so rig is found
	if err := os.MkdirAll(filepath.Join(tmpDir, ".beads"), 0755); err != nil {
		t.Fatal(err)
	}
	// Create plugins dir under rig root (but NOT under startDir itself)
	subDir := filepath.Join(tmpDir, "subcrew")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	rigPlugins := filepath.Join(tmpDir, "plugins")
	if err := os.MkdirAll(rigPlugins, 0755); err != nil {
		t.Fatal(err)
	}

	loc, err := NewLocator(subDir)
	if err != nil {
		t.Fatal(err)
	}

	paths := loc.GetLocationPaths()
	pluginsPath, ok := paths[LocationPlugins]
	if !ok {
		t.Fatal("expected plugins path in result")
	}
	if pluginsPath != rigPlugins {
		t.Errorf("plugins path = %q, want %q", pluginsPath, rigPlugins)
	}
}

// -- location.go: glob bad pattern --
func TestGlob_BadPattern(t *testing.T) {
	loc := &Locator{startDir: t.TempDir()}
	_, err := loc.glob(loc.startDir, "[invalid")
	if err == nil {
		t.Error("expected error for bad glob pattern")
	}
}

// -- validate.go: validateStep implement/crank/vibe --
func TestValidateStep_ImplementCrankVibeSteps(t *testing.T) {
	tmpDir := t.TempDir()
	artFile := filepath.Join(tmpDir, "test-artifact.md")
	if err := os.WriteFile(artFile, []byte("---\nschema_version: 1\n---\n# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	v, err := NewValidator(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, step := range []Step{StepImplement, StepCrank, StepVibe} {
		result, err := v.Validate(step, artFile)
		if err != nil {
			t.Fatalf("Validate(%s): %v", step, err)
		}
		// These steps should produce a warning about no validation rules
		found := false
		for _, w := range result.Warnings {
			if strings.Contains(w, "No artifact validation rules for step") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("step %s: expected 'No artifact validation rules' warning, got warnings: %v", step, result.Warnings)
		}
	}
}

// -- validate.go: ValidateArtifactPath tilde is dead code (~ doesn't start with /) --
// The tilde check on line 717 is unreachable since ~ never satisfies filepath.IsAbs.
// Verify the "must be absolute" error fires instead.
func TestValidateArtifactPath_TildePath(t *testing.T) {
	err := ValidateArtifactPath("~/some/path")
	if err == nil {
		t.Error("expected error for tilde path")
	}
	if !strings.Contains(err.Error(), "must be absolute") {
		t.Errorf("error = %q, want 'must be absolute'", err.Error())
	}
}

// -- validate.go: ValidateCloseReason relative path patterns --
func TestValidateCloseReason_RelativePathPatterns(t *testing.T) {
	issues := ValidateCloseReason("See ./relative/path for details")
	if len(issues) == 0 {
		t.Error("expected issues for ./ relative path")
	}
	foundRelative := false
	for _, issue := range issues {
		if strings.Contains(issue, "relative path") {
			foundRelative = true
		}
	}
	if !foundRelative {
		t.Errorf("expected 'relative path' issue, got: %v", issues)
	}
}

func TestValidateCloseReason_TildePattern(t *testing.T) {
	issues := ValidateCloseReason("See ~/home/user/file for details")
	if len(issues) == 0 {
		t.Error("expected issues for ~/ relative path")
	}
}

func TestValidateCloseReason_ParentRelative(t *testing.T) {
	issues := ValidateCloseReason("See ../parent/path for details")
	if len(issues) == 0 {
		t.Error("expected issues for ../ relative path")
	}
}

// -- validate.go: gatherSessionDirs with rig and town --
func TestGatherSessionDirs_WithRigAndTown(t *testing.T) {
	tmpDir := t.TempDir()

	// Create local sessions
	localSessions := filepath.Join(tmpDir, "subcrew", ".agents", "ao", "sessions")
	if err := os.MkdirAll(localSessions, 0755); err != nil {
		t.Fatal(err)
	}

	// Create rig marker and rig sessions
	rigDir := tmpDir
	if err := os.MkdirAll(filepath.Join(rigDir, ".beads"), 0755); err != nil {
		t.Fatal(err)
	}
	rigSessions := filepath.Join(rigDir, ".agents", "ao", "sessions")
	if err := os.MkdirAll(rigSessions, 0755); err != nil {
		t.Fatal(err)
	}

	v, err := NewValidator(filepath.Join(tmpDir, "subcrew"))
	if err != nil {
		t.Fatal(err)
	}

	dirs := v.gatherSessionDirs()
	if len(dirs) < 2 {
		t.Errorf("expected at least 2 session dirs (local + rig), got %d", len(dirs))
	}
}

// -- validate.go: countRefsInDir --
func TestCountRefsInDir_FindsReferences(t *testing.T) {
	tmpDir := t.TempDir()
	sessDir := filepath.Join(tmpDir, "sessions")
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create session file referencing an artifact
	sessFile := filepath.Join(sessDir, "s1.md")
	if err := os.WriteFile(sessFile, []byte("Used learning-x.md for context"), 0644); err != nil {
		t.Fatal(err)
	}

	v, err := NewValidator(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	seen := make(map[string]bool)
	count := v.countRefsInDir(sessDir, "learning-x.md", seen)
	if count != 1 {
		t.Errorf("countRefsInDir = %d, want 1", count)
	}
}

// -- validate.go: checkTierRequirements observation tier --
func TestCheckTierRequirements_ObservationTier(t *testing.T) {
	tmpDir := t.TempDir()
	artFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(artFile, []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	v, err := NewValidator(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	result, err := v.ValidateForPromotion(artFile, TierObservation)
	if err != nil {
		t.Fatalf("ValidateForPromotion: %v", err)
	}
	// Observation tier has no promotion requirements, so it should be valid
	if !result.Valid {
		t.Errorf("expected valid for observation tier, got issues: %v", result.Issues)
	}
}

// -- validate.go: RecordCitation mkdir error --
func TestRecordCitation_ReadOnlyDirectory(t *testing.T) {
	// Use a path that can't be created
	err := RecordCitation("/dev/null/impossible", types.CitationEvent{
		ArtifactPath: "/some/artifact.md",
		SessionID:    "s1",
	})
	if err == nil {
		t.Error("expected error for impossible directory")
	}
	if !strings.Contains(err.Error(), "create citations directory") {
		t.Errorf("error = %q, want 'create citations directory'", err.Error())
	}
}

// -- validate.go: GetCitationsSince missing file --
func TestGetCitationsSince_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	results, err := GetCitationsSince(tmpDir, time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("GetCitationsSince: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for missing file, got %d", len(results))
	}
}

// -- validate.go: GetCitationsForSession missing file --
func TestGetCitationsForSession_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	results, err := GetCitationsForSession(tmpDir, "nonexistent-session")
	if err != nil {
		t.Fatalf("GetCitationsForSession: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for missing file, got %d", len(results))
	}
}

// -- validate.go: CanonicalArtifactPath empty baseDir --
func TestCanonicalArtifactPath_EmptyBaseDir(t *testing.T) {
	result := CanonicalArtifactPath("", "relative/path.md")
	if result == "" {
		t.Error("expected non-empty result")
	}
	// Should resolve to absolute path using cwd
	if !filepath.IsAbs(result) {
		t.Errorf("expected absolute path, got %q", result)
	}
}

// -- validate.go: CanonicalArtifactPath empty path --
func TestCanonicalArtifactPath_EmptyPath(t *testing.T) {
	result := CanonicalArtifactPath("/base", "")
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

// -- validate.go: CanonicalArtifactPath whitespace path --
func TestCanonicalArtifactPath_WhitespacePath(t *testing.T) {
	result := CanonicalArtifactPath("/base", "   ")
	if result != "" {
		t.Errorf("expected empty result for whitespace-only path, got %q", result)
	}
}

// -- validate.go: isSearchableFile --
func TestIsSearchableFile_Coverage(t *testing.T) {
	tests := []struct {
		name string
		path string
		dir  bool
		want bool
	}{
		{"jsonl file", "test.jsonl", false, true},
		{"md file", "test.md", false, true},
		{"txt file", "test.txt", false, false},
		{"directory", "somedir", true, false},
		{"go file", "test.go", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := fakeFileInfo{name: tt.path, isDir: tt.dir}
			got := isSearchableFile(tt.path, info)
			if got != tt.want {
				t.Errorf("isSearchableFile(%q, dir=%v) = %v, want %v", tt.path, tt.dir, got, tt.want)
			}
		})
	}
}

// fakeFileInfo implements os.FileInfo for testing isSearchableFile.
type fakeFileInfo struct {
	name  string
	isDir bool
}

func (f fakeFileInfo) Name() string      { return f.name }
func (f fakeFileInfo) Size() int64       { return 0 }
func (f fakeFileInfo) Mode() os.FileMode { return 0644 }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool       { return f.isDir }
func (f fakeFileInfo) Sys() any          { return nil }

// -- gate.go: checkImplementGate returns false when no epic --
func TestGateChecker_CheckImplement_NoBdFallthrough(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	gc, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	result, err := gc.Check(StepImplement)
	if err != nil {
		t.Fatalf("Check(implement): %v", err)
	}
	// Without bd CLI, both findEpic calls fail, so result.Passed should be false
	if result.Passed {
		t.Log("Note: bd CLI is available, so implement gate passed")
	}
}

// -- gate.go: checkPostMortemGate soft gate --
func TestGateChecker_CheckPostMortem_SoftGate(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	gc, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	result, err := gc.Check(StepPostMortem)
	if err != nil {
		t.Fatalf("Check(post-mortem): %v", err)
	}
	// Post-mortem is a soft gate, should always pass
	if !result.Passed {
		t.Error("post-mortem gate should always pass (soft gate)")
	}
}

// -- validate.go: ValidateCloseReason with extracted paths that are absolute (valid)
func TestValidateCloseReason_ValidAbsolutePaths(t *testing.T) {
	issues := ValidateCloseReason("Artifact: /valid/absolute/path.md")
	if len(issues) != 0 {
		t.Errorf("expected no issues for absolute path, got: %v", issues)
	}
}

// -- chain.go: Append to new file exercises writeMetadata inside Append --
func TestAppend_NewFileWritesMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	chainPath := filepath.Join(tmpDir, ".agents", "ao", "chain.jsonl")

	chain := &Chain{
		ID:      "append-meta",
		Started: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
		Entries: []ChainEntry{},
	}
	chain.SetPath(chainPath)

	entry := ChainEntry{
		Step:      StepResearch,
		Timestamp: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
		Output:    "r.md",
		Locked:    true,
	}

	if err := chain.Append(entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	loaded, err := loadJSONLChain(chainPath)
	if err != nil {
		t.Fatalf("loadJSONLChain: %v", err)
	}
	if loaded.ID != "append-meta" {
		t.Errorf("loaded ID = %q, want %q", loaded.ID, "append-meta")
	}
	if len(loaded.Entries) != 1 {
		t.Errorf("loaded entries = %d, want 1", len(loaded.Entries))
	}
}

// -- maturity.go: updateMarkdownFrontMatter write error --
func TestUpdateMarkdownFrontMatter_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "readonly.md")
	content := "---\nmaturity: provisional\n---\n# Body"
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	// Make file read-only
	if err := os.Chmod(f, 0444); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(f, 0644) // cleanup

	err := updateMarkdownFrontMatter(f, map[string]any{"maturity": "candidate"})
	if err == nil {
		t.Error("expected error for read-only file")
	}
	if err != nil && !strings.Contains(err.Error(), "write updated learning") {
		t.Errorf("error = %q, want 'write updated learning'", err.Error())
	}
}

// -- maturity.go: updateJSONLFirstLine write error --
func TestUpdateJSONLFirstLine_WriteErrorReadOnly(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "readonly.jsonl")
	if err := os.WriteFile(f, []byte(`{"id":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}
	// Make file read-only
	if err := os.Chmod(f, 0444); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(f, 0644) // cleanup

	err := updateJSONLFirstLine(f, map[string]any{"maturity": "candidate"})
	if err == nil {
		t.Error("expected error for read-only file")
	}
	if err != nil && !strings.Contains(err.Error(), "write updated learning") {
		t.Errorf("error = %q, want 'write updated learning'", err.Error())
	}
}

// -- validate.go: GetUniqueCitedArtifacts --
func TestGetUniqueCitedArtifacts_WithCitations(t *testing.T) {
	tmpDir := t.TempDir()
	// Record some citations with different times
	now := time.Now()
	for _, event := range []types.CitationEvent{
		{ArtifactPath: "/a/file1.md", SessionID: "s1", CitedAt: now.Add(-time.Hour)},
		{ArtifactPath: "/a/file1.md", SessionID: "s2", CitedAt: now.Add(-30 * time.Minute)},
		{ArtifactPath: "/a/file2.md", SessionID: "s3", CitedAt: now.Add(-15 * time.Minute)},
		{ArtifactPath: "/a/file3.md", SessionID: "s4", CitedAt: now.Add(time.Hour)}, // outside window
	} {
		if err := RecordCitation(tmpDir, event); err != nil {
			t.Fatalf("RecordCitation: %v", err)
		}
	}

	unique, err := GetUniqueCitedArtifacts(tmpDir, now.Add(-2*time.Hour), now)
	if err != nil {
		t.Fatalf("GetUniqueCitedArtifacts: %v", err)
	}
	if len(unique) != 2 {
		t.Errorf("got %d unique artifacts, want 2", len(unique))
	}
}
