package main

import (
	"strings"
	"testing"
)

func TestSelectHandoffFields_SubsetSelection(t *testing.T) {
	h := &phaseHandoff{
		Goal:              "add auth",
		EpicID:            "ag-123",
		Verdicts:          map[string]string{"pre_mortem": "PASS"},
		ArtifactsProduced: []string{"plan.md"},
		DecisionsMade:     []string{"use JWT"},
		OpenRisks:         []string{"migration downtime"},
	}

	result := selectHandoffFields(h, []string{"goal", "verdicts"})

	if result["goal"] != "add auth" {
		t.Errorf("goal = %v, want 'add auth'", result["goal"])
	}
	if result["verdicts"] == nil {
		t.Error("verdicts should be present")
	}
	// Fields NOT requested should be absent
	if _, ok := result["epic_id"]; ok {
		t.Error("epic_id should not be present in subset selection")
	}
	if _, ok := result["artifacts_produced"]; ok {
		t.Error("artifacts_produced should not be present in subset selection")
	}
	if _, ok := result["decisions_made"]; ok {
		t.Error("decisions_made should not be present in subset selection")
	}
	if _, ok := result["open_risks"]; ok {
		t.Error("open_risks should not be present in subset selection")
	}
}

func TestSelectHandoffFields_AllFields(t *testing.T) {
	h := &phaseHandoff{
		Goal:              "add auth",
		EpicID:            "ag-123",
		Verdicts:          map[string]string{"pre_mortem": "PASS"},
		ArtifactsProduced: []string{"plan.md"},
		DecisionsMade:     []string{"use JWT"},
		OpenRisks:         []string{"migration downtime"},
	}

	// nil fields = return all
	result := selectHandoffFields(h, nil)
	expected := []string{"goal", "epic_id", "verdicts", "artifacts_produced", "decisions_made", "open_risks"}
	for _, field := range expected {
		if _, ok := result[field]; !ok {
			t.Errorf("field %q missing from all-fields result", field)
		}
	}

	// empty slice = return all
	result2 := selectHandoffFields(h, []string{})
	for _, field := range expected {
		if _, ok := result2[field]; !ok {
			t.Errorf("field %q missing from empty-slice result", field)
		}
	}
}

func TestBuildHandoffContext_WithManifest_Phase2(t *testing.T) {
	handoffs := []*phaseHandoff{
		{
			Phase: 1, PhaseName: "discovery", Status: "completed",
			Goal:              "add auth",
			EpicID:            "ag-123",
			Verdicts:          map[string]string{"pre_mortem": "PASS"},
			ArtifactsProduced: []string{"plan.md"},
			DecisionsMade:     []string{"use JWT"},
			OpenRisks:         []string{"migration downtime"},
			Narrative:         "Discovery completed.",
		},
	}

	// Phase 2 manifest: has decisions + risks, no artifacts
	manifest := defaultPhaseManifests[2]
	ctx := buildHandoffContext(handoffs, manifest)

	if !strings.Contains(ctx, "use JWT") {
		t.Errorf("phase 2 manifest should include decisions\ngot:\n%s", ctx)
	}
	if !strings.Contains(ctx, "migration downtime") {
		t.Errorf("phase 2 manifest should include risks\ngot:\n%s", ctx)
	}
	// artifacts_produced is NOT in phase 2 manifest HandoffFields
	if strings.Contains(ctx, "plan.md") {
		t.Errorf("phase 2 manifest should NOT include artifacts\ngot:\n%s", ctx)
	}
}

func TestBuildHandoffContext_WithManifest_Phase3(t *testing.T) {
	handoffs := []*phaseHandoff{
		{
			Phase: 2, PhaseName: "implementation", Status: "completed",
			Goal:              "add auth",
			EpicID:            "ag-456",
			Verdicts:          map[string]string{"crank": "PASS"},
			ArtifactsProduced: []string{"auth.go", "auth_test.go"},
			DecisionsMade:     []string{"use JWT"},
			OpenRisks:         []string{"migration downtime"},
			Narrative:         "Implementation completed.",
		},
	}

	// Phase 3 manifest: has artifacts, no decisions or risks
	manifest := defaultPhaseManifests[3]
	ctx := buildHandoffContext(handoffs, manifest)

	if !strings.Contains(ctx, "auth.go") {
		t.Errorf("phase 3 manifest should include artifacts\ngot:\n%s", ctx)
	}
	// decisions_made is NOT in phase 3 manifest HandoffFields
	if strings.Contains(ctx, "use JWT") {
		t.Errorf("phase 3 manifest should NOT include decisions\ngot:\n%s", ctx)
	}
	// open_risks is NOT in phase 3 manifest HandoffFields
	if strings.Contains(ctx, "migration downtime") {
		t.Errorf("phase 3 manifest should NOT include risks\ngot:\n%s", ctx)
	}
}

func TestBuildHandoffContext_NarrativeCap(t *testing.T) {
	longNarrative := strings.Repeat("a", 1000)
	handoffs := []*phaseHandoff{
		{
			Phase: 1, PhaseName: "discovery", Status: "completed",
			Narrative: longNarrative,
		},
	}

	// NarrativeCap=500 should truncate at 500, not 1000
	manifest := phaseManifest{NarrativeCap: 500}
	ctx := buildHandoffContext(handoffs, manifest)

	if !strings.Contains(ctx, "...") {
		t.Error("expected truncation marker for narrative exceeding cap")
	}

	// Extract the narrative line and verify length
	// The truncated narrative should be 500 chars of 'a' + "..."
	truncated := strings.Repeat("a", 500) + "..."
	if !strings.Contains(ctx, truncated) {
		t.Errorf("expected narrative truncated to 500 chars + '...'\ngot:\n%s", ctx)
	}

	// Full 1000-char narrative should NOT appear
	if strings.Contains(ctx, longNarrative) {
		t.Error("full narrative should be truncated at NarrativeCap=500")
	}
}
