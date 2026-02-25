package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestMDLearning creates a markdown learning file with YAML front matter.
func writeTestMDLearning(t *testing.T, dir, filename string, fm map[string]string, body string) string {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("---\n")
	for k, v := range fm {
		sb.WriteString(k + ": " + v + "\n")
	}
	sb.WriteString("---\n")
	sb.WriteString(body)
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		t.Fatalf("write test learning: %v", err)
	}
	return path
}

func TestApplyAllMaturityTransitions_NoLearningsDir(t *testing.T) {
	tmp := t.TempDir()
	// No .agents/learnings/ directory exists
	summary, err := applyAllMaturityTransitions(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 0 {
		t.Errorf("Total = %d, want 0", summary.Total)
	}
	if summary.Applied != 0 {
		t.Errorf("Applied = %d, want 0", summary.Applied)
	}
	if len(summary.ChangedPaths) != 0 {
		t.Errorf("ChangedPaths = %v, want empty", summary.ChangedPaths)
	}
}

func TestApplyAllMaturityTransitions_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	summary, err := applyAllMaturityTransitions(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 0 {
		t.Errorf("Total = %d, want 0", summary.Total)
	}
	if summary.Applied != 0 {
		t.Errorf("Applied = %d, want 0", summary.Applied)
	}
}

func TestApplyAllMaturityTransitions_AppliesPromotion(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// provisional -> candidate requires utility >= 0.7 AND reward_count >= 3
	path := writeTestMDLearning(t, learningsDir, "test-learn.md", map[string]string{
		"id":           "test-learn",
		"maturity":     "provisional",
		"utility":      "0.85",
		"reward_count": "5",
		"confidence":   "0.8",
	}, "# Test Learning\nContent here.\n")

	summary, err := applyAllMaturityTransitions(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 1 {
		t.Errorf("Total = %d, want 1", summary.Total)
	}
	if summary.Applied != 1 {
		t.Errorf("Applied = %d, want 1", summary.Applied)
	}

	// Verify the file now has maturity: candidate
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "maturity: candidate") {
		t.Errorf("expected maturity: candidate in file, got:\n%s", content)
	}
}

func TestApplyAllMaturityTransitions_AppliesAntiPattern(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// anti-pattern requires utility <= 0.2 AND harmful_count >= 5
	path := writeTestMDLearning(t, learningsDir, "bad-learn.md", map[string]string{
		"id":            "bad-learn",
		"maturity":      "provisional",
		"utility":       "0.1",
		"harmful_count": "6",
		"confidence":    "0.5",
	}, "# Bad Learning\nThis is harmful.\n")

	summary, err := applyAllMaturityTransitions(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Total != 1 {
		t.Errorf("Total = %d, want 1", summary.Total)
	}
	if summary.Applied != 1 {
		t.Errorf("Applied = %d, want 1", summary.Applied)
	}

	// Verify the file now has maturity: anti-pattern
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "maturity: anti-pattern") {
		t.Errorf("expected maturity: anti-pattern in file, got:\n%s", content)
	}
}

func TestApplyAllMaturityTransitions_MixedTransitions(t *testing.T) {
	tmp := t.TempDir()
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// 1. Should promote: provisional -> candidate (utility >= 0.7, reward_count >= 3)
	writeTestMDLearning(t, learningsDir, "good-learn.md", map[string]string{
		"id":           "good-learn",
		"maturity":     "provisional",
		"utility":      "0.85",
		"reward_count": "5",
		"confidence":   "0.8",
	}, "# Good Learning\nHigh utility content.\n")

	// 2. Should stay: provisional, mid-range utility, not enough rewards
	writeTestMDLearning(t, learningsDir, "mid-learn.md", map[string]string{
		"id":           "mid-learn",
		"maturity":     "provisional",
		"utility":      "0.5",
		"reward_count": "1",
		"confidence":   "0.6",
	}, "# Mid Learning\nAverage content.\n")

	// 3. Should become anti-pattern: utility <= 0.2, harmful_count >= 5
	writeTestMDLearning(t, learningsDir, "bad-learn.md", map[string]string{
		"id":            "bad-learn",
		"maturity":      "provisional",
		"utility":       "0.1",
		"harmful_count": "7",
		"confidence":    "0.3",
	}, "# Bad Learning\nHarmful content.\n")

	summary, err := applyAllMaturityTransitions(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Total should be 2: good-learn (promotes) and bad-learn (anti-pattern)
	// mid-learn does NOT transition so it is NOT in the scan results
	if summary.Total != 2 {
		t.Errorf("Total = %d, want 2", summary.Total)
	}
	if summary.Applied != 2 {
		t.Errorf("Applied = %d, want 2", summary.Applied)
	}
	if len(summary.ChangedPaths) != 2 {
		t.Errorf("ChangedPaths count = %d, want 2", len(summary.ChangedPaths))
	}

	// Verify good-learn is now candidate
	goodData, err := os.ReadFile(filepath.Join(learningsDir, "good-learn.md"))
	if err != nil {
		t.Fatalf("read good-learn: %v", err)
	}
	if !strings.Contains(string(goodData), "maturity: candidate") {
		t.Errorf("good-learn should be candidate, got:\n%s", string(goodData))
	}

	// Verify mid-learn is still provisional
	midData, err := os.ReadFile(filepath.Join(learningsDir, "mid-learn.md"))
	if err != nil {
		t.Fatalf("read mid-learn: %v", err)
	}
	if !strings.Contains(string(midData), "maturity: provisional") {
		t.Errorf("mid-learn should still be provisional, got:\n%s", string(midData))
	}

	// Verify bad-learn is now anti-pattern
	badData, err := os.ReadFile(filepath.Join(learningsDir, "bad-learn.md"))
	if err != nil {
		t.Fatalf("read bad-learn: %v", err)
	}
	if !strings.Contains(string(badData), "maturity: anti-pattern") {
		t.Errorf("bad-learn should be anti-pattern, got:\n%s", string(badData))
	}
}
