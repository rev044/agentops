package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boshu2/agentops/cli/internal/types"
)

// ---------------------------------------------------------------------------
// readLearningData — md files
// ---------------------------------------------------------------------------

func TestMaturity_readLearningData_Markdown(t *testing.T) {
	tmp := t.TempDir()

	t.Run("reads frontmatter fields from md", func(t *testing.T) {
		path := filepath.Join(tmp, "learn.md")
		content := "---\nutility: 0.7500\nconfidence: 0.6000\nmaturity: candidate\n---\n# Content\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		data, ok := readLearningData(path)
		if !ok {
			t.Fatal("expected ok=true for valid md with frontmatter")
		}

		utility, ok := data["utility"].(float64)
		if !ok || utility != 0.75 {
			t.Errorf("utility = %v, want 0.75", data["utility"])
		}

		maturity, ok := data["maturity"].(string)
		if !ok || maturity != "candidate" {
			t.Errorf("maturity = %v, want 'candidate'", data["maturity"])
		}
	})

	t.Run("returns false for md with no parseable fields", func(t *testing.T) {
		path := filepath.Join(tmp, "plain.md")
		content := "# Just a heading\nSome content\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		_, ok := readLearningData(path)
		if ok {
			t.Error("expected ok=false for md file with no relevant frontmatter fields")
		}
	})

	t.Run("handles numeric strings as floats", func(t *testing.T) {
		path := filepath.Join(tmp, "numeric.md")
		content := "---\nutility: 0.4200\nreward_count: 5\n---\n# X\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		data, ok := readLearningData(path)
		if !ok {
			t.Fatal("expected ok=true")
		}

		// reward_count should parse as float
		if rc, ok := data["reward_count"].(float64); !ok || rc != 5 {
			t.Errorf("reward_count = %v, want 5.0", data["reward_count"])
		}
	})
}

func TestMaturity_readLearningData_JSONL(t *testing.T) {
	tmp := t.TempDir()

	t.Run("delegates to JSONL reader", func(t *testing.T) {
		path := filepath.Join(tmp, "learn.jsonl")
		firstLine := `{"id":"L-1","utility":0.8,"maturity":"established"}`
		if err := os.WriteFile(path, []byte(firstLine+"\n"), 0644); err != nil {
			t.Fatal(err)
		}

		data, ok := readLearningData(path)
		if !ok {
			t.Fatal("expected ok=true for valid JSONL")
		}

		if data["id"] != "L-1" {
			t.Errorf("id = %v, want 'L-1'", data["id"])
		}
	})
}

func TestMaturity_readLearningData_MissingFile(t *testing.T) {
	_, ok := readLearningData("/nonexistent/file.md")
	if ok {
		t.Error("expected ok=false for missing file")
	}
}

// ---------------------------------------------------------------------------
// runMaturityMigrateMd
// ---------------------------------------------------------------------------

func TestMaturity_runMaturityMigrateMd_NoBareMarkdown(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Already has utility - should NOT be migrated
	content := "---\nutility: 0.8000\nmaturity: established\n---\n# Already good\n"
	if err := os.WriteFile(filepath.Join(learningsDir, "good.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := runMaturityMigrateMd(learningsDir)
	if err != nil {
		t.Fatalf("runMaturityMigrateMd failed: %v", err)
	}

	// Verify the file was NOT modified
	data, _ := os.ReadFile(filepath.Join(learningsDir, "good.md"))
	if !strings.Contains(string(data), "0.8000") {
		t.Error("file with existing utility should not be modified")
	}
}

func TestMaturity_runMaturityMigrateMd_InjectsFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Has frontmatter but no utility
	content := "---\ntitle: Test\n---\n# Content\n"
	if err := os.WriteFile(filepath.Join(learningsDir, "needsmigration.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := runMaturityMigrateMd(learningsDir)
	if err != nil {
		t.Fatalf("runMaturityMigrateMd failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(learningsDir, "needsmigration.md"))
	text := string(data)
	if !strings.Contains(text, "utility:") {
		t.Error("expected utility field to be injected")
	}
	if !strings.Contains(text, "maturity: provisional") {
		t.Error("expected maturity field to be injected")
	}
}

func TestMaturity_runMaturityMigrateMd_PrependsFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// No frontmatter at all
	content := "# Plain markdown\nSome content here\n"
	if err := os.WriteFile(filepath.Join(learningsDir, "bare.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := runMaturityMigrateMd(learningsDir)
	if err != nil {
		t.Fatalf("runMaturityMigrateMd failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(learningsDir, "bare.md"))
	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		t.Error("expected frontmatter to be prepended")
	}
	if !strings.Contains(text, "utility:") {
		t.Error("expected utility in prepended frontmatter")
	}
	if !strings.Contains(text, "# Plain markdown") {
		t.Error("expected original content to be preserved")
	}
}

func TestMaturity_runMaturityMigrateMd_SkipsJSONL(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// JSONL file should be skipped
	jsonlContent := `{"id":"L-1","utility":0.5}` + "\n"
	if err := os.WriteFile(filepath.Join(learningsDir, "test.jsonl"), []byte(jsonlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// .md file to verify counting
	mdContent := "# No frontmatter\nContent\n"
	if err := os.WriteFile(filepath.Join(learningsDir, "needs.md"), []byte(mdContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := runMaturityMigrateMd(learningsDir)
	if err != nil {
		t.Fatalf("runMaturityMigrateMd failed: %v", err)
	}

	// JSONL should be unchanged
	jsonlData, _ := os.ReadFile(filepath.Join(learningsDir, "test.jsonl"))
	if string(jsonlData) != jsonlContent {
		t.Error("JSONL file should not be modified")
	}
}

// ---------------------------------------------------------------------------
// runMaturityRecalibrate
// ---------------------------------------------------------------------------

func TestMaturity_runMaturityRecalibrate_DryRun(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}

	jsonlContent := `{"id":"L-1","utility":0.3,"confidence":0.5}` + "\n"
	if err := os.WriteFile(filepath.Join(learningsDir, "test.jsonl"), []byte(jsonlContent), 0644); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	err := runMaturityRecalibrate(learningsDir)
	if err != nil {
		t.Fatalf("runMaturityRecalibrate dry-run failed: %v", err)
	}

	// Verify file was NOT modified in dry-run
	data, _ := os.ReadFile(filepath.Join(learningsDir, "test.jsonl"))
	if !strings.Contains(string(data), "0.3") {
		t.Error("file should not be modified in dry-run mode")
	}
}

// ---------------------------------------------------------------------------
// readLearningData dispatching
// ---------------------------------------------------------------------------

func TestMaturity_readLearningData_DispatchesByExtension(t *testing.T) {
	tmp := t.TempDir()

	// Create a JSONL file
	jsonlPath := filepath.Join(tmp, "test.jsonl")
	jsonlData := map[string]any{"id": "L-1", "utility": 0.75}
	b, _ := json.Marshal(jsonlData)
	if err := os.WriteFile(jsonlPath, append(b, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an MD file
	mdPath := filepath.Join(tmp, "test.md")
	mdContent := "---\nutility: 0.6000\nmaturity: candidate\n---\n# X\n"
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		t.Fatal(err)
	}

	// JSONL dispatch
	data, ok := readLearningData(jsonlPath)
	if !ok {
		t.Fatal("expected ok=true for JSONL")
	}
	if data["id"] != "L-1" {
		t.Errorf("JSONL id = %v, want L-1", data["id"])
	}

	// MD dispatch
	data, ok = readLearningData(mdPath)
	if !ok {
		t.Fatal("expected ok=true for MD")
	}
	if v, ok := data["maturity"].(string); !ok || v != "candidate" {
		t.Errorf("MD maturity = %v, want candidate", data["maturity"])
	}
}

// ---------------------------------------------------------------------------
// runMaturityMigrateMd edge cases
// ---------------------------------------------------------------------------

func TestMaturity_runMaturityMigrateMd_MalformedFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Opening --- but no closing ---
	content := "---\ntitle: Bad\nno closing delimiter\n"
	if err := os.WriteFile(filepath.Join(learningsDir, "malformed.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Should not panic or error
	err := runMaturityMigrateMd(learningsDir)
	if err != nil {
		t.Fatalf("runMaturityMigrateMd failed on malformed: %v", err)
	}
}

func TestMaturity_runMaturityMigrateMd_EmptyLearningsDir(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := runMaturityMigrateMd(learningsDir)
	if err != nil {
		t.Fatalf("runMaturityMigrateMd failed on empty: %v", err)
	}
}

// ---------------------------------------------------------------------------
// runMaturityRecalibrate — actual recalibration
// ---------------------------------------------------------------------------

func TestMaturity_runMaturityRecalibrate_ResetsUtility(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatal(err)
	}

	data := map[string]any{
		"id":         "L-1",
		"utility":    0.1,
		"confidence": 0.2,
		"maturity":   "provisional",
	}
	b, _ := json.Marshal(data)
	if err := os.WriteFile(filepath.Join(learningsDir, "low.jsonl"), append(b, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	err := runMaturityRecalibrate(learningsDir)
	if err != nil {
		t.Fatalf("runMaturityRecalibrate failed: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(learningsDir, "low.jsonl"))
	lines := strings.Split(string(content), "\n")
	if len(lines) < 1 || lines[0] == "" {
		t.Fatal("expected non-empty first line after recalibration")
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &result); err != nil {
		t.Fatalf("unmarshal recalibrated: %v", err)
	}

	// Utility should be reset to InitialUtility (0.5)
	utility, ok := result["utility"].(float64)
	if !ok {
		t.Fatal("utility not found in result")
	}
	if utility < types.InitialUtility-0.01 || utility > types.InitialUtility+0.01 {
		t.Errorf("utility = %.4f, want ~%.4f", utility, types.InitialUtility)
	}
}

// ---------------------------------------------------------------------------
// expiryCategory struct behavior
// ---------------------------------------------------------------------------

func TestMaturity_expiryCategory_AccumulatesCorrectly(t *testing.T) {
	cats := expiryCategory{}
	cats.active = append(cats.active, "a.md", "b.md")
	cats.neverExpiring = append(cats.neverExpiring, "c.md")
	cats.newlyExpired = append(cats.newlyExpired, "d.md")
	cats.alreadyArchived = append(cats.alreadyArchived, "e.md")

	total := len(cats.active) + len(cats.neverExpiring) + len(cats.newlyExpired) + len(cats.alreadyArchived)
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
}

// ---------------------------------------------------------------------------
// evictionCandidate JSON serialization
// ---------------------------------------------------------------------------

func TestMaturity_evictionCandidate_JSONSerialization(t *testing.T) {
	c := evictionCandidate{
		Path:       "/path/to/file.jsonl",
		Name:       "file.jsonl",
		Utility:    0.15,
		Confidence: 0.1,
		Maturity:   "provisional",
		LastCited:  "never",
	}

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded evictionCandidate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Name != "file.jsonl" {
		t.Errorf("Name = %q, want %q", decoded.Name, "file.jsonl")
	}
	if decoded.Utility != 0.15 {
		t.Errorf("Utility = %f, want 0.15", decoded.Utility)
	}
	if decoded.LastCited != "never" {
		t.Errorf("LastCited = %q, want %q", decoded.LastCited, "never")
	}
}
