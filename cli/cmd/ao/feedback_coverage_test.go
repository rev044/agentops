package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// printFeedbackJSON (0%)
// ---------------------------------------------------------------------------

func TestFeedbackCov_PrintFeedbackJSON(t *testing.T) {
	// printFeedbackJSON writes to os.Stdout. Capture it.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	printErr := printFeedbackJSON("L001", "/path/to/L001.jsonl", "helpful", 0.5, 0.55, 1.0, 0.1)

	_ = w.Close()
	os.Stdout = origStdout

	if printErr != nil {
		t.Fatalf("printFeedbackJSON() error = %v", printErr)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\nOutput: %s", err, output)
	}
	if result["learning_id"] != "L001" {
		t.Errorf("learning_id = %v, want L001", result["learning_id"])
	}
	if result["feedback_type"] != "helpful" {
		t.Errorf("feedback_type = %v, want helpful", result["feedback_type"])
	}
	if result["old_utility"].(float64) != 0.5 {
		t.Errorf("old_utility = %v, want 0.5", result["old_utility"])
	}
	if result["new_utility"].(float64) != 0.55 {
		t.Errorf("new_utility = %v, want 0.55", result["new_utility"])
	}
}

// ---------------------------------------------------------------------------
// runFeedback (0%) — dry-run path
// ---------------------------------------------------------------------------

func TestFeedbackCov_RunFeedback_DryRun(t *testing.T) {
	tmp := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	// Create a learning file
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "L001.jsonl"), []byte(`{"id":"L001","utility":0.5}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Save/restore globals
	origDryRun := dryRun
	origReward := feedbackReward
	origAlpha := feedbackAlpha
	origHelpful := feedbackHelpful
	origHarmful := feedbackHarmful
	dryRun = true
	feedbackReward = -1
	feedbackAlpha = 0.1
	feedbackHelpful = true
	feedbackHarmful = false
	t.Cleanup(func() {
		dryRun = origDryRun
		feedbackReward = origReward
		feedbackAlpha = origAlpha
		feedbackHelpful = origHelpful
		feedbackHarmful = origHarmful
	})

	err = runFeedback(feedbackCmd, []string{"L001"})
	if err != nil {
		t.Fatalf("runFeedback() dry-run error = %v", err)
	}
}

// ---------------------------------------------------------------------------
// runFeedback (0%) — actual update path (text output)
// ---------------------------------------------------------------------------

func TestFeedbackCov_RunFeedback_ActualUpdate(t *testing.T) {
	tmp := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "L002.jsonl"), []byte(`{"id":"L002","utility":0.5}`), 0644); err != nil {
		t.Fatal(err)
	}

	origDryRun := dryRun
	origOutput := output
	origReward := feedbackReward
	origAlpha := feedbackAlpha
	origHelpful := feedbackHelpful
	origHarmful := feedbackHarmful
	dryRun = false
	output = "table"
	feedbackReward = -1
	feedbackAlpha = 0.1
	feedbackHelpful = false
	feedbackHarmful = true
	t.Cleanup(func() {
		dryRun = origDryRun
		output = origOutput
		feedbackReward = origReward
		feedbackAlpha = origAlpha
		feedbackHelpful = origHelpful
		feedbackHarmful = origHarmful
	})

	err = runFeedback(feedbackCmd, []string{"L002"})
	if err != nil {
		t.Fatalf("runFeedback() error = %v", err)
	}

	// Verify the file was updated
	raw, err := os.ReadFile(filepath.Join(learningsDir, "L002.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if data["utility"] == nil {
		t.Error("expected utility field after update")
	}
}

// ---------------------------------------------------------------------------
// runFeedback — validation error (both helpful + harmful)
// ---------------------------------------------------------------------------

func TestFeedbackCov_RunFeedback_ValidationError(t *testing.T) {
	origReward := feedbackReward
	origAlpha := feedbackAlpha
	origHelpful := feedbackHelpful
	origHarmful := feedbackHarmful
	feedbackReward = -1
	feedbackAlpha = 0.1
	feedbackHelpful = true
	feedbackHarmful = true
	t.Cleanup(func() {
		feedbackReward = origReward
		feedbackAlpha = origAlpha
		feedbackHelpful = origHelpful
		feedbackHarmful = origHarmful
	})

	err := runFeedback(feedbackCmd, []string{"L999"})
	if err == nil {
		t.Error("expected error for both helpful+harmful")
	}
	if !strings.Contains(err.Error(), "cannot use both") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// runFeedback — learning not found
// ---------------------------------------------------------------------------

func TestFeedbackCov_RunFeedback_LearningNotFound(t *testing.T) {
	tmp := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	origReward := feedbackReward
	origAlpha := feedbackAlpha
	origHelpful := feedbackHelpful
	origHarmful := feedbackHarmful
	feedbackReward = 1.0
	feedbackAlpha = 0.1
	feedbackHelpful = false
	feedbackHarmful = false
	t.Cleanup(func() {
		feedbackReward = origReward
		feedbackAlpha = origAlpha
		feedbackHelpful = origHelpful
		feedbackHarmful = origHarmful
	})

	err = runFeedback(feedbackCmd, []string{"NONEXISTENT"})
	if err == nil {
		t.Error("expected error for missing learning")
	}
	if !strings.Contains(err.Error(), "find learning") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// runMigrate (0%)
// ---------------------------------------------------------------------------

func TestFeedbackCov_RunMigrate_UnknownMigration(t *testing.T) {
	err := runMigrate(migrateCmd, []string{"unknown"})
	if err == nil {
		t.Error("expected error for unknown migration")
	}
	if !strings.Contains(err.Error(), "unknown migration") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFeedbackCov_RunMigrate_MemRL_NoLearnings(t *testing.T) {
	tmp := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	err = runMigrate(migrateCmd, []string{"memrl"})
	if err != nil {
		t.Fatalf("runMigrate() error = %v", err)
	}
}

func TestFeedbackCov_RunMigrate_MemRL_WithFiles(t *testing.T) {
	tmp := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// File without utility (should be migrated)
	if err := os.WriteFile(filepath.Join(learningsDir, "L001.jsonl"), []byte(`{"id":"L001"}`), 0644); err != nil {
		t.Fatal(err)
	}
	// File with utility (should be skipped)
	if err := os.WriteFile(filepath.Join(learningsDir, "L002.jsonl"), []byte(`{"id":"L002","utility":0.7}`), 0644); err != nil {
		t.Fatal(err)
	}

	origDryRun := dryRun
	dryRun = false
	t.Cleanup(func() { dryRun = origDryRun })

	err = runMigrate(migrateCmd, []string{"memrl"})
	if err != nil {
		t.Fatalf("runMigrate() error = %v", err)
	}

	// Verify L001 was migrated
	raw, readErr := os.ReadFile(filepath.Join(learningsDir, "L001.jsonl"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if _, ok := data["utility"]; !ok {
		t.Error("expected utility field added to L001.jsonl")
	}
}

func TestFeedbackCov_RunMigrate_MemRL_DryRun(t *testing.T) {
	tmp := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "L001.jsonl"), []byte(`{"id":"L001"}`), 0644); err != nil {
		t.Fatal(err)
	}

	origDryRun := dryRun
	dryRun = true
	t.Cleanup(func() { dryRun = origDryRun })

	err = runMigrate(migrateCmd, []string{"memrl"})
	if err != nil {
		t.Fatalf("runMigrate() dry-run error = %v", err)
	}

	// Verify file was NOT changed
	raw, readErr := os.ReadFile(filepath.Join(learningsDir, "L001.jsonl"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if _, ok := data["utility"]; ok {
		t.Error("expected utility NOT added in dry-run mode")
	}
}

// ---------------------------------------------------------------------------
// migrateJSONLFiles — exercise dry-run and error paths
// ---------------------------------------------------------------------------

func TestFeedbackCov_MigrateJSONLFiles(t *testing.T) {
	tmp := t.TempDir()

	// Valid file needing migration
	f1 := filepath.Join(tmp, "needs.jsonl")
	if err := os.WriteFile(f1, []byte(`{"id":"L1"}`), 0644); err != nil {
		t.Fatal(err)
	}
	// Valid file already migrated
	f2 := filepath.Join(tmp, "done.jsonl")
	if err := os.WriteFile(f2, []byte(`{"id":"L2","utility":0.8}`), 0644); err != nil {
		t.Fatal(err)
	}
	// Invalid JSON file
	f3 := filepath.Join(tmp, "bad.jsonl")
	if err := os.WriteFile(f3, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("actual migration", func(t *testing.T) {
		migrated, skipped := migrateJSONLFiles([]string{f1, f2, f3}, false)
		if migrated != 1 {
			t.Errorf("migrated = %d, want 1", migrated)
		}
		if skipped != 1 {
			t.Errorf("skipped = %d, want 1", skipped)
		}
	})

	t.Run("dry-run migration", func(t *testing.T) {
		// Reset f1 to need migration again
		if err := os.WriteFile(f1, []byte(`{"id":"L1"}`), 0644); err != nil {
			t.Fatal(err)
		}
		migrated, skipped := migrateJSONLFiles([]string{f1, f2}, true)
		if migrated != 1 {
			t.Errorf("migrated = %d, want 1 in dry-run", migrated)
		}
		if skipped != 1 {
			t.Errorf("skipped = %d, want 1", skipped)
		}
	})
}

// ---------------------------------------------------------------------------
// resolveReward — edge cases
// ---------------------------------------------------------------------------

func TestFeedbackCov_ResolveReward_NoInput(t *testing.T) {
	_, err := resolveReward(false, false, -1, 0.1)
	if err == nil {
		t.Error("expected error when no reward/helpful/harmful")
	}
}

func TestFeedbackCov_ResolveReward_TooHigh(t *testing.T) {
	_, err := resolveReward(false, false, 1.5, 0.1)
	if err == nil {
		t.Error("expected error for reward > 1")
	}
}

func TestFeedbackCov_ResolveReward_BadAlpha(t *testing.T) {
	_, err := resolveReward(false, false, 0.5, 0.0)
	if err == nil {
		t.Error("expected error for alpha=0")
	}
	_, err = resolveReward(false, false, 0.5, 1.5)
	if err == nil {
		t.Error("expected error for alpha>1")
	}
}

// ---------------------------------------------------------------------------
// classifyFeedbackType
// ---------------------------------------------------------------------------

func TestFeedbackCov_ClassifyFeedbackType(t *testing.T) {
	if got := classifyFeedbackType(true, false); got != "helpful" {
		t.Errorf("got %q, want helpful", got)
	}
	if got := classifyFeedbackType(false, true); got != "harmful" {
		t.Errorf("got %q, want harmful", got)
	}
	if got := classifyFeedbackType(false, false); got != "custom" {
		t.Errorf("got %q, want custom", got)
	}
}

// ---------------------------------------------------------------------------
// parseJSONLFirstLine — error paths
// ---------------------------------------------------------------------------

func TestFeedbackCov_ParseJSONLFirstLine_Empty(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.jsonl")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	_, _, err := parseJSONLFirstLine(path)
	if err == nil {
		t.Error("expected error for empty JSONL file")
	}
}

func TestFeedbackCov_ParseJSONLFirstLine_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.jsonl")
	if err := os.WriteFile(path, []byte("not json\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, _, err := parseJSONLFirstLine(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFeedbackCov_ParseJSONLFirstLine_NonexistentFile(t *testing.T) {
	_, _, err := parseJSONLFirstLine("/nonexistent/path/file.jsonl")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// ---------------------------------------------------------------------------
// parseFrontMatterUtility — malformed
// ---------------------------------------------------------------------------

func TestFeedbackCov_ParseFrontMatterUtility_NoClosure(t *testing.T) {
	lines := []string{"---", "utility: 0.5", "no closing"}
	_, _, err := parseFrontMatterUtility(lines)
	if err == nil {
		t.Error("expected error for malformed front matter without closing ---")
	}
}

// ---------------------------------------------------------------------------
// rebuildWithFrontMatter
// ---------------------------------------------------------------------------

func TestFeedbackCov_RebuildWithFrontMatter(t *testing.T) {
	fm := []string{"id: test", "utility: 0.5"}
	body := []string{"# Test", "", "Content here."}
	result := rebuildWithFrontMatter(fm, body)
	if !strings.HasPrefix(result, "---\n") {
		t.Error("expected front matter opening ---")
	}
	if !strings.Contains(result, "id: test") {
		t.Error("expected id field")
	}
	if !strings.Contains(result, "# Test") {
		t.Error("expected body content")
	}
}

// ---------------------------------------------------------------------------
// needsUtilityMigration — empty file
// ---------------------------------------------------------------------------

func TestFeedbackCov_NeedsUtilityMigration_EmptyFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.jsonl")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := needsUtilityMigration(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("empty file should not need migration")
	}
}
