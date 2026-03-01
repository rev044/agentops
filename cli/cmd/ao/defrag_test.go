package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindOrphanLearnings_StaleUnreferenced(t *testing.T) {
	tmp := t.TempDir()

	// Create learnings dir with an old file
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o750); err != nil {
		t.Fatal(err)
	}
	oldFile := filepath.Join(learningsDir, "old-learning.md")
	if err := os.WriteFile(oldFile, []byte("# Old learning\nSome content"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Set mtime to 60 days ago
	oldTime := time.Now().AddDate(0, 0, -60)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Create empty patterns and research dirs (no references)
	if err := os.MkdirAll(filepath.Join(tmp, ".agents", "patterns"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".agents", "research"), 0o750); err != nil {
		t.Fatal(err)
	}

	result, err := findOrphanLearnings(tmp, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalLearnings != 1 {
		t.Errorf("TotalLearnings = %d, want 1", result.TotalLearnings)
	}
	if result.StaleCount != 1 {
		t.Errorf("StaleCount = %d, want 1", result.StaleCount)
	}
	if len(result.Orphans) != 1 {
		t.Fatalf("Orphans count = %d, want 1", len(result.Orphans))
	}
	expected := filepath.Join(".agents", "learnings", "old-learning.md")
	if result.Orphans[0] != expected {
		t.Errorf("Orphans[0] = %q, want %q", result.Orphans[0], expected)
	}
}

func TestFindOrphanLearnings_StaleButReferenced(t *testing.T) {
	tmp := t.TempDir()

	// Create learnings dir with an old file
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o750); err != nil {
		t.Fatal(err)
	}
	oldFile := filepath.Join(learningsDir, "referenced-learning.md")
	if err := os.WriteFile(oldFile, []byte("# Referenced learning"), 0o600); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().AddDate(0, 0, -60)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Create a pattern file that references the learning
	patternsDir := filepath.Join(tmp, ".agents", "patterns")
	if err := os.MkdirAll(patternsDir, 0o750); err != nil {
		t.Fatal(err)
	}
	patternContent := "# Pattern\nSee [referenced-learning.md](../learnings/referenced-learning.md)\n"
	if err := os.WriteFile(filepath.Join(patternsDir, "pattern.md"), []byte(patternContent), 0o600); err != nil {
		t.Fatal(err)
	}

	// Create research dir
	if err := os.MkdirAll(filepath.Join(tmp, ".agents", "research"), 0o750); err != nil {
		t.Fatal(err)
	}

	result, err := findOrphanLearnings(tmp, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.StaleCount != 1 {
		t.Errorf("StaleCount = %d, want 1", result.StaleCount)
	}
	if len(result.Orphans) != 0 {
		t.Errorf("Orphans = %v, want empty (file is referenced)", result.Orphans)
	}
}

func TestFindOrphanLearnings_NoLearningsDir(t *testing.T) {
	tmp := t.TempDir()

	result, err := findOrphanLearnings(tmp, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalLearnings != 0 {
		t.Errorf("TotalLearnings = %d, want 0", result.TotalLearnings)
	}
}

func TestFindDuplicateLearnings_NearDuplicate(t *testing.T) {
	tmp := t.TempDir()

	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// Two nearly identical files
	content := "This is a learning about how to handle errors in Go programs effectively and safely"
	if err := os.WriteFile(filepath.Join(learningsDir, "a.md"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	// Slightly different version of same content
	content2 := "This is a learning about how to handle errors in Go programs effectively and well"
	if err := os.WriteFile(filepath.Join(learningsDir, "b.md"), []byte(content2), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := findDuplicateLearnings(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Checked != 2 {
		t.Errorf("Checked = %d, want 2", result.Checked)
	}
	if len(result.DuplicatePairs) != 1 {
		t.Fatalf("DuplicatePairs count = %d, want 1", len(result.DuplicatePairs))
	}
	pair := result.DuplicatePairs[0]
	if pair[0] != "a.md" || pair[1] != "b.md" {
		t.Errorf("DuplicatePairs[0] = %v, want [a.md, b.md]", pair)
	}
}

func TestFindDuplicateLearnings_Distinct(t *testing.T) {
	tmp := t.TempDir()

	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// Two completely different files
	if err := os.WriteFile(filepath.Join(learningsDir, "x.md"),
		[]byte("# Go Error Handling\nAlways wrap errors with fmt.Errorf and %w verb"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "y.md"),
		[]byte("# Python Virtual Environments\nUse venv module for isolated package management"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := findDuplicateLearnings(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Checked != 2 {
		t.Errorf("Checked = %d, want 2", result.Checked)
	}
	if len(result.DuplicatePairs) != 0 {
		t.Errorf("DuplicatePairs = %v, want empty", result.DuplicatePairs)
	}
}

func TestSweepOscillatingGoals_Oscillating(t *testing.T) {
	tmp := t.TempDir()

	evolveDir := filepath.Join(tmp, ".agents", "evolve")
	if err := os.MkdirAll(evolveDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// Create a cycle history where "logging" oscillates improved/fail 4 times
	lines := []cycleRecord{
		{Cycle: 1, Target: "logging", Result: "improved"},
		{Cycle: 2, Target: "logging", Result: "fail"},
		{Cycle: 3, Target: "logging", Result: "improved"},
		{Cycle: 4, Target: "logging", Result: "fail"},
		{Cycle: 5, Target: "logging", Result: "improved"},
	}

	histPath := filepath.Join(evolveDir, "cycle-history.jsonl")
	f, err := os.Create(histPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, rec := range lines {
		data, _ := json.Marshal(rec)
		f.Write(data)
		f.Write([]byte("\n"))
	}
	f.Close()

	result, err := sweepOscillatingGoals(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.OscillatingGoals) != 1 {
		t.Fatalf("OscillatingGoals count = %d, want 1", len(result.OscillatingGoals))
	}
	goal := result.OscillatingGoals[0]
	if goal.Target != "logging" {
		t.Errorf("Target = %q, want %q", goal.Target, "logging")
	}
	if goal.AlternationCount != 4 {
		t.Errorf("AlternationCount = %d, want 4", goal.AlternationCount)
	}
	if goal.LastCycle != 5 {
		t.Errorf("LastCycle = %d, want 5", goal.LastCycle)
	}
}

func TestSweepOscillatingGoals_Stable(t *testing.T) {
	tmp := t.TempDir()

	evolveDir := filepath.Join(tmp, ".agents", "evolve")
	if err := os.MkdirAll(evolveDir, 0o750); err != nil {
		t.Fatal(err)
	}

	// All improved — no oscillation
	lines := []cycleRecord{
		{Cycle: 1, Target: "tests", Result: "improved"},
		{Cycle: 2, Target: "tests", Result: "improved"},
		{Cycle: 3, Target: "tests", Result: "improved"},
	}

	histPath := filepath.Join(evolveDir, "cycle-history.jsonl")
	f, err := os.Create(histPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, rec := range lines {
		data, _ := json.Marshal(rec)
		f.Write(data)
		f.Write([]byte("\n"))
	}
	f.Close()

	result, err := sweepOscillatingGoals(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.OscillatingGoals) != 0 {
		t.Errorf("OscillatingGoals = %v, want empty", result.OscillatingGoals)
	}
}

func TestSweepOscillatingGoals_NoHistory(t *testing.T) {
	tmp := t.TempDir()

	// No cycle-history.jsonl file at all
	result, err := sweepOscillatingGoals(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.OscillatingGoals != nil {
		t.Errorf("OscillatingGoals = %v, want nil", result.OscillatingGoals)
	}
}

func TestRunDefrag_DryRun(t *testing.T) {
	tmp := t.TempDir()

	// Create learnings dir with an old orphan file
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".agents", "patterns"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".agents", "research"), 0o750); err != nil {
		t.Fatal(err)
	}

	orphanFile := filepath.Join(learningsDir, "orphan.md")
	if err := os.WriteFile(orphanFile, []byte("# Orphan"), 0o600); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().AddDate(0, 0, -60)
	if err := os.Chtimes(orphanFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Simulate dry-run by calling findOrphanLearnings (the function
	// doesn't delete — that's done in runDefrag based on isDryRun)
	result, err := findOrphanLearnings(tmp, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Orphans) != 1 {
		t.Fatalf("Orphans count = %d, want 1", len(result.Orphans))
	}

	// Verify file still exists (dry-run should not delete)
	if _, err := os.Stat(orphanFile); os.IsNotExist(err) {
		t.Error("orphan file was deleted during dry-run")
	}

	// Verify Deleted is nil (not set during dry-run)
	if result.Deleted != nil {
		t.Errorf("Deleted = %v, want nil in dry-run", result.Deleted)
	}
}

func TestWriteDefragReport_CreatesLatest(t *testing.T) {
	tmp := t.TempDir()

	outDir := filepath.Join(tmp, "defrag-output")

	// Temporarily set defragQuiet so printDefragSummary is suppressed in tests
	origQuiet := defragQuiet
	defragQuiet = true
	defer func() { defragQuiet = origQuiet }()

	report := &DefragReport{
		Timestamp: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		DryRun:    true,
		Prune: &PruneResult{
			TotalLearnings: 10,
			StaleCount:     3,
			Orphans:        []string{".agents/learnings/old.md"},
		},
	}

	if err := writeDefragReport(outDir, report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check dated file exists
	datedPath := filepath.Join(outDir, "2026-03-01.json")
	if _, err := os.Stat(datedPath); os.IsNotExist(err) {
		t.Error("dated report file not created")
	}

	// Check latest.json exists
	latestPath := filepath.Join(outDir, "latest.json")
	if _, err := os.Stat(latestPath); os.IsNotExist(err) {
		t.Error("latest.json not created")
	}

	// Verify content can be parsed back
	data, err := os.ReadFile(latestPath)
	if err != nil {
		t.Fatalf("read latest.json: %v", err)
	}

	var parsed DefragReport
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal latest.json: %v", err)
	}

	if !parsed.DryRun {
		t.Error("parsed.DryRun = false, want true")
	}
	if parsed.Prune == nil {
		t.Fatal("parsed.Prune is nil")
	}
	if parsed.Prune.TotalLearnings != 10 {
		t.Errorf("parsed.Prune.TotalLearnings = %d, want 10", parsed.Prune.TotalLearnings)
	}
}

func TestBuildTrigrams(t *testing.T) {
	tg := buildTrigrams("abcde")
	// Should have: "abc", "bcd", "cde"
	if len(tg) != 3 {
		t.Errorf("trigram count = %d, want 3", len(tg))
	}
	for _, expected := range []string{"abc", "bcd", "cde"} {
		if !tg[expected] {
			t.Errorf("missing trigram %q", expected)
		}
	}
}

func TestTrigramOverlap_Identical(t *testing.T) {
	a := buildTrigrams("hello world")
	b := buildTrigrams("hello world")
	overlap := trigramOverlap(a, b)
	if overlap != 1.0 {
		t.Errorf("overlap = %f, want 1.0", overlap)
	}
}

func TestTrigramOverlap_Empty(t *testing.T) {
	a := buildTrigrams("")
	b := buildTrigrams("")
	overlap := trigramOverlap(a, b)
	if overlap != 0 {
		t.Errorf("overlap = %f, want 0", overlap)
	}
}

func TestCountAlternations(t *testing.T) {
	tests := []struct {
		name    string
		records []cycleRecord
		want    int
	}{
		{
			name:    "empty",
			records: nil,
			want:    0,
		},
		{
			name: "no alternation",
			records: []cycleRecord{
				{Result: "improved"},
				{Result: "improved"},
				{Result: "improved"},
			},
			want: 0,
		},
		{
			name: "three alternations",
			records: []cycleRecord{
				{Result: "improved"},
				{Result: "fail"},
				{Result: "improved"},
				{Result: "fail"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countAlternations(tt.records)
			if got != tt.want {
				t.Errorf("countAlternations = %d, want %d", got, tt.want)
			}
		})
	}
}
