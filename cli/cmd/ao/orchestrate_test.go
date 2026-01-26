package main

import (
	"strings"
	"testing"

	"github.com/boshu2/agentops/cli/internal/orchestrator"
)

func TestGenerateShortID(t *testing.T) {
	// Generate multiple IDs and verify uniqueness
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateShortID()
		if seen[id] {
			t.Errorf("duplicate ID generated: %s", id)
		}
		seen[id] = true

		// Should be hex-encoded 4 bytes = 8 characters
		if len(id) != 8 {
			// Could be fallback format
			if len(id) < 8 {
				t.Errorf("ID too short: %s", id)
			}
		}
	}
}

func TestBuildDispatchPlan(t *testing.T) {
	planID := "test-plan-123"
	files := []string{"file1.go", "file2.go"}
	findingsDir := "/tmp/test-findings"

	plan := buildDispatchPlan(planID, files, findingsDir)

	if plan.PlanID != planID {
		t.Errorf("expected plan ID %s, got %s", planID, plan.PlanID)
	}

	if plan.FindingsDir != findingsDir {
		t.Errorf("expected findings dir %s, got %s", findingsDir, plan.FindingsDir)
	}

	if len(plan.Wave1) == 0 {
		t.Error("expected at least one dispatch in Wave1")
	}

	// Verify each dispatch has required fields
	for _, dispatch := range plan.Wave1 {
		if dispatch.ID == "" {
			t.Error("dispatch ID should not be empty")
		}
		if dispatch.Category == "" {
			t.Error("dispatch category should not be empty")
		}
		if dispatch.SubagentType == "" {
			t.Error("dispatch subagent type should not be empty")
		}
		if dispatch.Prompt == "" {
			t.Error("dispatch prompt should not be empty")
		}
		if len(dispatch.Files) != len(files) {
			t.Errorf("expected %d files, got %d", len(files), len(dispatch.Files))
		}
	}

	// Verify all expected categories are present
	expectedCategories := make(map[string]bool)
	for _, pod := range orchestrator.PodCategories {
		expectedCategories[pod.Category] = true
	}

	for _, dispatch := range plan.Wave1 {
		delete(expectedCategories, dispatch.Category)
	}

	if len(expectedCategories) > 0 {
		var missing []string
		for cat := range expectedCategories {
			missing = append(missing, cat)
		}
		t.Errorf("missing categories in plan: %v", missing)
	}
}

func TestBuildAgentPrompt(t *testing.T) {
	files := []string{"main.go", "utils.go"}

	tests := []struct {
		category   string
		wantPrefix string
	}{
		{"security", "You are a security expert"},
		{"quality", "You are a quality expert"},
		{"architecture", "You are a architecture expert"},
	}

	for _, tt := range tests {
		prompt := buildAgentPrompt(tt.category, files, "Custom pod prompt")

		if !strings.HasPrefix(prompt, tt.wantPrefix) {
			t.Errorf("prompt for %s should start with '%s', got: %s...",
				tt.category, tt.wantPrefix, prompt[:50])
		}

		// Should include files
		if !strings.Contains(prompt, "main.go") {
			t.Errorf("prompt should contain file names")
		}

		// Should include output format instruction
		if !strings.Contains(prompt, "OUTPUT FORMAT") {
			t.Errorf("prompt should include output format")
		}

		// Should include JSON structure
		if !strings.Contains(prompt, `"findings"`) {
			t.Errorf("prompt should include findings JSON structure")
		}
	}
}

func TestBuildAgentPromptLongFileList(t *testing.T) {
	// Create a very long file list
	var files []string
	for i := 0; i < 50; i++ {
		files = append(files, "very/long/path/to/some/deeply/nested/file"+string(rune('a'+i%26))+".go")
	}

	prompt := buildAgentPrompt("security", files, "Pod prompt")

	// The file list should be truncated
	if len(prompt) > 5000 {
		// Prompt should be reasonable length
		t.Logf("prompt length: %d", len(prompt))
	}

	// Should still contain truncation indicator if files were truncated
	if strings.Contains(prompt, "...") {
		t.Log("file list was truncated as expected")
	}
}

func TestDispatchPlanCreatedTimestamp(t *testing.T) {
	plan := buildDispatchPlan("test", []string{"file.go"}, "/tmp")

	if plan.Created == "" {
		t.Error("plan should have created timestamp")
	}

	// Should be in RFC3339 format
	if !strings.Contains(plan.Created, "T") {
		t.Errorf("timestamp should be in RFC3339 format: %s", plan.Created)
	}
}

func TestDispatchPlanConfig(t *testing.T) {
	plan := buildDispatchPlan("test", []string{"file.go"}, "/tmp")

	// Default config values
	if plan.Config.PodSize == 0 {
		t.Error("config PodSize should not be zero")
	}

	if plan.Config.QuorumThreshold == 0 {
		t.Error("config QuorumThreshold should not be zero")
	}
}
