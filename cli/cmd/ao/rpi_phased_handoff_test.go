package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWritePhaseHandoff_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	h := &phaseHandoff{
		SchemaVersion:     1,
		RunID:             "test-run",
		Phase:             1,
		PhaseName:         "discovery",
		Status:            "completed",
		Goal:              "test goal",
		EpicID:            "ag-123",
		Verdicts:          map[string]string{"pre_mortem": "PASS"},
		ArtifactsProduced: []string{"plan.md"},
		DecisionsMade:     []string{"use JWT"},
		OpenRisks:         []string{"migration downtime"},
		DurationSeconds:   312.5,
		ToolCalls:         42,
		Narrative:         "Discovery completed successfully.",
		CompletedAt:       "2026-03-02T12:00:00Z",
	}

	if err := writePhaseHandoff(dir, h); err != nil {
		t.Fatalf("writePhaseHandoff: %v", err)
	}

	got, err := readPhaseHandoff(dir, 1)
	if err != nil {
		t.Fatalf("readPhaseHandoff: %v", err)
	}

	if got.RunID != h.RunID {
		t.Errorf("RunID = %q, want %q", got.RunID, h.RunID)
	}
	if got.Goal != h.Goal {
		t.Errorf("Goal = %q, want %q", got.Goal, h.Goal)
	}
	if got.EpicID != h.EpicID {
		t.Errorf("EpicID = %q, want %q", got.EpicID, h.EpicID)
	}
	if got.Status != h.Status {
		t.Errorf("Status = %q, want %q", got.Status, h.Status)
	}
	if got.DurationSeconds != h.DurationSeconds {
		t.Errorf("DurationSeconds = %f, want %f", got.DurationSeconds, h.DurationSeconds)
	}
	if v, ok := got.Verdicts["pre_mortem"]; !ok || v != "PASS" {
		t.Errorf("Verdicts[pre_mortem] = %q, want PASS", v)
	}
	if len(got.ArtifactsProduced) != 1 || got.ArtifactsProduced[0] != "plan.md" {
		t.Errorf("ArtifactsProduced = %v, want [plan.md]", got.ArtifactsProduced)
	}
	if len(got.DecisionsMade) != 1 || got.DecisionsMade[0] != "use JWT" {
		t.Errorf("DecisionsMade = %v, want [use JWT]", got.DecisionsMade)
	}
	if len(got.OpenRisks) != 1 || got.OpenRisks[0] != "migration downtime" {
		t.Errorf("OpenRisks = %v, want [migration downtime]", got.OpenRisks)
	}
	if got.ToolCalls != h.ToolCalls {
		t.Errorf("ToolCalls = %d, want %d", got.ToolCalls, h.ToolCalls)
	}
	if got.Narrative != h.Narrative {
		t.Errorf("Narrative = %q, want %q", got.Narrative, h.Narrative)
	}
	if got.CompletedAt != h.CompletedAt {
		t.Errorf("CompletedAt = %q, want %q", got.CompletedAt, h.CompletedAt)
	}
}

func TestReadPhaseHandoff_MissingSummary(t *testing.T) {
	dir := t.TempDir()
	// No files exist
	_, err := readPhaseHandoff(dir, 1)
	if err == nil {
		t.Fatal("expected error for missing handoff and summary")
	}
}

func TestReadAllHandoffs_Mixed(t *testing.T) {
	dir := t.TempDir()
	rpiDir := filepath.Join(dir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write handoff for phase 1
	h1 := &phaseHandoff{SchemaVersion: 1, Phase: 1, PhaseName: "discovery", Status: "completed", Goal: "test"}
	if err := writePhaseHandoff(dir, h1); err != nil {
		t.Fatal(err)
	}

	// Write handoff for phase 2
	h2 := &phaseHandoff{SchemaVersion: 1, Phase: 2, PhaseName: "implementation", Status: "completed", Goal: "test"}
	if err := writePhaseHandoff(dir, h2); err != nil {
		t.Fatal(err)
	}

	// Phase 3 missing — should skip

	handoffs, err := readAllHandoffs(dir, 4)
	if err != nil {
		t.Fatalf("readAllHandoffs: %v", err)
	}
	if len(handoffs) != 2 {
		t.Errorf("got %d handoffs, want 2", len(handoffs))
	}
}

func TestReadAllHandoffs_NoHandoffs(t *testing.T) {
	dir := t.TempDir()
	_, err := readAllHandoffs(dir, 3)
	if err == nil {
		t.Fatal("expected error when no handoffs exist")
	}
}

func TestReadPhaseHandoff_LegacyFallback(t *testing.T) {
	dir := t.TempDir()
	rpiDir := filepath.Join(dir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a legacy summary file (no handoff.json)
	summaryPath := filepath.Join(rpiDir, "phase-1-summary.md")
	if err := os.WriteFile(summaryPath, []byte("# Phase 1 Summary\nDiscovery completed."), 0o644); err != nil {
		t.Fatal(err)
	}

	h, err := readPhaseHandoff(dir, 1)
	if err != nil {
		t.Fatalf("readPhaseHandoff with legacy fallback: %v", err)
	}
	if h.Phase != 1 {
		t.Errorf("Phase = %d, want 1", h.Phase)
	}
	if h.PhaseName != "discovery" {
		t.Errorf("PhaseName = %q, want discovery", h.PhaseName)
	}
	if h.Narrative == "" {
		t.Error("expected narrative from legacy summary")
	}
	if h.Status != "completed" {
		t.Errorf("Status = %q, want completed", h.Status)
	}
	if h.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", h.SchemaVersion)
	}
}

func TestBuildHandoffContext_Formatting(t *testing.T) {
	handoffs := []*phaseHandoff{
		{
			Phase: 1, PhaseName: "discovery", Status: "completed",
			Goal: "add auth", EpicID: "ag-123",
			DurationSeconds:   312,
			Verdicts:          map[string]string{"pre_mortem": "PASS"},
			ArtifactsProduced: []string{"plan.md"},
			DecisionsMade:     []string{"use JWT"},
			OpenRisks:         []string{"migration downtime"},
			Narrative:         "Discovery done.",
		},
	}

	allFieldsManifest := phaseManifest{Phase: 2, NarrativeCap: 1000}
	ctx := buildHandoffContext(handoffs, allFieldsManifest)
	if ctx == "" {
		t.Fatal("expected non-empty context")
	}

	// Check required sections
	checks := []string{
		"RPI Context",
		"Goal: add auth",
		"Phase 1: discovery",
		"ag-123",
		"pre_mortem PASS",
		"plan.md",
		"use JWT",
		"migration downtime",
		"Discovery done.",
	}
	for _, check := range checks {
		if !strings.Contains(ctx, check) {
			t.Errorf("context missing %q\ngot:\n%s", check, ctx)
		}
	}
}

func TestBuildHandoffContext_MultiPhase(t *testing.T) {
	handoffs := []*phaseHandoff{
		{
			Phase: 1, PhaseName: "discovery", Status: "completed",
			Goal: "add auth",
		},
		{
			Phase: 2, PhaseName: "implementation", Status: "time_boxed",
			Goal:            "add auth",
			DurationSeconds: 5400,
		},
	}

	allFieldsManifest := phaseManifest{Phase: 2, NarrativeCap: 1000}
	ctx := buildHandoffContext(handoffs, allFieldsManifest)
	if !strings.Contains(ctx, "Phase 1: discovery") {
		t.Error("missing phase 1")
	}
	if !strings.Contains(ctx, "Phase 2: implementation") {
		t.Error("missing phase 2")
	}
	if !strings.Contains(ctx, "time_boxed") {
		t.Error("missing time_boxed status")
	}
	if !strings.Contains(ctx, "5400s") {
		t.Error("missing duration")
	}
}

func TestBuildPhaseHandoffFromState(t *testing.T) {
	dir := t.TempDir()
	rpiDir := filepath.Join(dir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a summary file
	if err := os.WriteFile(filepath.Join(rpiDir, "phase-1-summary.md"), []byte("Phase 1 done."), 0o644); err != nil {
		t.Fatal(err)
	}

	state := &phasedState{
		RunID:    "test-run",
		Goal:     "test goal",
		EpicID:   "ag-456",
		Verdicts: map[string]string{"pre_mortem": "WARN"},
	}

	h := buildPhaseHandoffFromState(state, 1, dir)
	if h.RunID != "test-run" {
		t.Errorf("RunID = %q, want test-run", h.RunID)
	}
	if h.Goal != "test goal" {
		t.Errorf("Goal = %q, want test goal", h.Goal)
	}
	if h.EpicID != "ag-456" {
		t.Errorf("EpicID = %q, want ag-456", h.EpicID)
	}
	if v := h.Verdicts["pre_mortem"]; v != "WARN" {
		t.Errorf("Verdicts[pre_mortem] = %q, want WARN", v)
	}
	if h.Narrative != "Phase 1 done." {
		t.Errorf("Narrative = %q, want 'Phase 1 done.'", h.Narrative)
	}
	if h.CompletedAt == "" {
		t.Error("expected CompletedAt to be set")
	}
	if h.Status != "completed" {
		t.Errorf("Status = %q, want completed", h.Status)
	}
}

func TestBuildPhaseHandoffFromState_WithPhaseResult(t *testing.T) {
	dir := t.TempDir()
	rpiDir := filepath.Join(dir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a phase result with time_boxed status
	pr := &phaseResult{
		SchemaVersion:   1,
		RunID:           "test-run",
		Phase:           2,
		PhaseName:       "implementation",
		Status:          "time_boxed",
		DurationSeconds: 5400,
	}
	if err := writePhaseResult(dir, pr); err != nil {
		t.Fatal(err)
	}

	state := &phasedState{
		RunID:    "test-run",
		Goal:     "test goal",
		Verdicts: map[string]string{},
	}

	h := buildPhaseHandoffFromState(state, 2, dir)
	if h.Status != "time_boxed" {
		t.Errorf("Status = %q, want time_boxed", h.Status)
	}
	if h.DurationSeconds != 5400 {
		t.Errorf("DurationSeconds = %f, want 5400", h.DurationSeconds)
	}
}

func TestWritePhaseHandoff_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	h := &phaseHandoff{SchemaVersion: 1, Phase: 2, PhaseName: "implementation", Status: "completed"}
	if err := writePhaseHandoff(dir, h); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Verify no .tmp file remains
	tmpPath := filepath.Join(dir, ".agents", "rpi", "phase-2-handoff.json.tmp")
	if _, err := os.Stat(tmpPath); err == nil {
		t.Error("tmp file should not remain after atomic write")
	}

	// Verify file is valid JSON
	data, err := os.ReadFile(filepath.Join(dir, ".agents", "rpi", "phase-2-handoff.json"))
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var parsed phaseHandoff
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed.Phase != 2 {
		t.Errorf("parsed Phase = %d, want 2", parsed.Phase)
	}
}

func TestBuildHandoffContext_Empty(t *testing.T) {
	ctx := buildHandoffContext(nil, phaseManifest{})
	if ctx != "" {
		t.Errorf("expected empty context for nil handoffs, got %q", ctx)
	}
}

func TestBuildHandoffContext_EmptySlice(t *testing.T) {
	ctx := buildHandoffContext([]*phaseHandoff{}, phaseManifest{})
	if ctx != "" {
		t.Errorf("expected empty context for empty handoffs, got %q", ctx)
	}
}

func TestBuildHandoffContext_NarrativeTruncation(t *testing.T) {
	longNarrative := strings.Repeat("x", 2000)
	handoffs := []*phaseHandoff{
		{
			Phase: 1, PhaseName: "discovery", Status: "completed",
			Narrative: longNarrative,
		},
	}

	ctx := buildHandoffContext(handoffs, phaseManifest{NarrativeCap: 1000})
	// Narrative should be capped at 1000 chars + "..."
	if !strings.Contains(ctx, "...") {
		t.Error("expected truncation marker for long narrative")
	}
	// The full 2000-char narrative should NOT appear
	if strings.Contains(ctx, longNarrative) {
		t.Error("full 2000-char narrative should be truncated")
	}
}

func TestBuildHandoffContext_DeterministicVerdictOrder(t *testing.T) {
	handoffs := []*phaseHandoff{
		{
			Phase: 1, PhaseName: "discovery", Status: "completed",
			Verdicts: map[string]string{
				"vibe":       "PASS",
				"pre_mortem": "WARN",
				"crank":      "PASS",
			},
		},
	}

	manifest := phaseManifest{Phase: 2, NarrativeCap: 0}
	// Run 10 times to catch non-determinism
	first := buildHandoffContext(handoffs, manifest)
	for i := 0; i < 10; i++ {
		got := buildHandoffContext(handoffs, manifest)
		if got != first {
			t.Fatalf("non-deterministic output on iteration %d:\nfirst:\n%s\ngot:\n%s", i, first, got)
		}
	}
	// Verify sorted order: crank, pre_mortem, vibe
	if !strings.Contains(first, "crank PASS, pre_mortem WARN, vibe PASS") {
		t.Errorf("verdicts not in sorted order\ngot:\n%s", first)
	}
}

func TestRenderHandoffField_StringSlice(t *testing.T) {
	got := renderHandoffField("Artifacts", []string{"plan.md", "auth.go"})
	if got != "Artifacts: plan.md, auth.go\n" {
		t.Errorf("renderHandoffField = %q, want 'Artifacts: plan.md, auth.go\\n'", got)
	}
}

func TestRenderHandoffField_EmptySlice(t *testing.T) {
	got := renderHandoffField("Artifacts", []string{})
	if got != "" {
		t.Errorf("renderHandoffField for empty slice = %q, want empty", got)
	}
}

func TestRenderHandoffField_String(t *testing.T) {
	got := renderHandoffField("Epic", "ag-123")
	if got != "Epic: ag-123\n" {
		t.Errorf("renderHandoffField = %q, want 'Epic: ag-123\\n'", got)
	}
}

func TestRenderHandoffField_EmptyString(t *testing.T) {
	got := renderHandoffField("Epic", "")
	if got != "" {
		t.Errorf("renderHandoffField for empty string = %q, want empty", got)
	}
}

func TestFormatVerdicts_Sorted(t *testing.T) {
	verdicts := map[string]string{
		"zebra": "FAIL",
		"alpha": "PASS",
		"mid":   "WARN",
	}
	got := formatVerdicts(verdicts)
	want := "Verdict: alpha PASS, mid WARN, zebra FAIL\n"
	if got != want {
		t.Errorf("formatVerdicts = %q, want %q", got, want)
	}
}

func TestFormatVerdicts_Empty(t *testing.T) {
	got := formatVerdicts(nil)
	if got != "" {
		t.Errorf("formatVerdicts(nil) = %q, want empty", got)
	}
	got = formatVerdicts(map[string]string{})
	if got != "" {
		t.Errorf("formatVerdicts({}) = %q, want empty", got)
	}
}

func TestDiscoverPhaseArtifacts_Discovery(t *testing.T) {
	dir := t.TempDir()
	rpiDir := filepath.Join(dir, ".agents", "rpi")
	plansDir := filepath.Join(dir, ".agents", "plans")
	councilDir := filepath.Join(dir, ".agents", "council")
	for _, d := range []string{rpiDir, plansDir, councilDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Create phase summary
	os.WriteFile(filepath.Join(rpiDir, "phase-1-summary.md"), []byte("summary"), 0o644)
	// Create plan
	os.WriteFile(filepath.Join(plansDir, "plan.md"), []byte("plan"), 0o644)
	// Create pre-mortem report
	os.WriteFile(filepath.Join(councilDir, "2026-03-02-pre-mortem-auth.md"), []byte("report"), 0o644)

	artifacts := discoverPhaseArtifacts(dir, 1)
	if len(artifacts) < 3 {
		t.Errorf("expected at least 3 artifacts, got %d: %v", len(artifacts), artifacts)
	}
}

func TestDiscoverPhaseArtifacts_Validation(t *testing.T) {
	dir := t.TempDir()
	rpiDir := filepath.Join(dir, ".agents", "rpi")
	councilDir := filepath.Join(dir, ".agents", "council")
	for _, d := range []string{rpiDir, councilDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Create vibe report
	os.WriteFile(filepath.Join(councilDir, "2026-03-02-vibe-recent.md"), []byte("vibe"), 0o644)

	artifacts := discoverPhaseArtifacts(dir, 3)
	found := false
	for _, a := range artifacts {
		if strings.Contains(a, "vibe") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected vibe artifact, got %v", artifacts)
	}
}

func TestDiscoverPhaseArtifacts_NoArtifacts(t *testing.T) {
	dir := t.TempDir()
	artifacts := discoverPhaseArtifacts(dir, 1)
	if len(artifacts) != 0 {
		t.Errorf("expected no artifacts, got %v", artifacts)
	}
}

func TestReadPhaseHandoff_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	rpiDir := filepath.Join(dir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write invalid JSON
	jsonPath := filepath.Join(rpiDir, "phase-1-handoff.json")
	if err := os.WriteFile(jsonPath, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := readPhaseHandoff(dir, 1)
	if err == nil {
		t.Fatal("expected error for corrupt JSON")
	}
	if !strings.Contains(err.Error(), "parse handoff") {
		t.Errorf("expected 'parse handoff' error, got: %v", err)
	}
}

func TestWritePhaseHandoff_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	// .agents/rpi/ does not exist yet
	h := &phaseHandoff{SchemaVersion: 1, Phase: 1, PhaseName: "discovery", Status: "completed"}
	if err := writePhaseHandoff(dir, h); err != nil {
		t.Fatalf("writePhaseHandoff should create directories: %v", err)
	}

	// Verify file exists
	jsonPath := filepath.Join(dir, ".agents", "rpi", "phase-1-handoff.json")
	if _, err := os.Stat(jsonPath); err != nil {
		t.Errorf("handoff file not created: %v", err)
	}
}

func TestReadPhaseHandoff_PrefersJSONOverLegacy(t *testing.T) {
	dir := t.TempDir()
	rpiDir := filepath.Join(dir, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write both JSON handoff and legacy summary
	h := &phaseHandoff{
		SchemaVersion: 1, Phase: 1, PhaseName: "discovery",
		Status: "completed", Goal: "from-json",
	}
	if err := writePhaseHandoff(dir, h); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rpiDir, "phase-1-summary.md"), []byte("from-legacy"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := readPhaseHandoff(dir, 1)
	if err != nil {
		t.Fatalf("readPhaseHandoff: %v", err)
	}
	// Should prefer JSON
	if got.Goal != "from-json" {
		t.Errorf("Goal = %q, want from-json (JSON should take precedence)", got.Goal)
	}
}
