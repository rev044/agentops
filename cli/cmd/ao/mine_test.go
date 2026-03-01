package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseMineWindow_ValidDurations(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"26h", 26 * time.Hour},
		{"7d", 168 * time.Hour},
		{"30m", 30 * time.Minute},
		{"1h", 1 * time.Hour},
		{"1d", 24 * time.Hour},
		{"5m", 5 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseMineWindow(tt.input)
			if err != nil {
				t.Fatalf("parseMineWindow(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseMineWindow(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseMineWindow_Invalid(t *testing.T) {
	tests := []string{
		"foo",
		"",
		"0h",
		"-1d",
		"abc",
		"7x",
		"d",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseMineWindow(input)
			if err == nil {
				t.Errorf("parseMineWindow(%q) expected error, got nil", input)
			}
		})
	}
}

func TestSplitSources_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"git,agents,code", []string{"git", "agents", "code"}},
		{"git,agents", []string{"git", "agents"}},
		{"code", []string{"code"}},
		{"git", []string{"git"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := splitSources(tt.input)
			if err != nil {
				t.Fatalf("splitSources(%q) error: %v", tt.input, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("splitSources(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("splitSources(%q)[%d] = %q, want %q", tt.input, i, g, tt.want[i])
				}
			}
		})
	}
}

func TestSplitSources_UnknownSource(t *testing.T) {
	tests := []string{
		"git,fake",
		"unknown",
		"git,agents,xyz",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := splitSources(input)
			if err == nil {
				t.Errorf("splitSources(%q) expected error, got nil", input)
			}
		})
	}
}

func TestMineAgentsDir_OrphanDetection(t *testing.T) {
	tmp := t.TempDir()

	// Create .agents/research/ with two files
	researchDir := filepath.Join(tmp, ".agents", "research")
	if err := os.MkdirAll(researchDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(researchDir, "topic-a.md"), []byte("# Topic A\nSome research."), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(researchDir, "topic-b.md"), []byte("# Topic B\nMore research."), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create .agents/learnings/ with a file that references topic-a.md only
	learningsDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learningsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learningsDir, "learning-1.md"), []byte("# Learning\nBased on topic-a.md research."), 0o644); err != nil {
		t.Fatal(err)
	}

	findings, err := mineAgentsDir(tmp)
	if err != nil {
		t.Fatalf("mineAgentsDir error: %v", err)
	}

	if findings.TotalResearch != 2 {
		t.Errorf("TotalResearch = %d, want 2", findings.TotalResearch)
	}

	if len(findings.OrphanedResearch) != 1 {
		t.Fatalf("OrphanedResearch count = %d, want 1", len(findings.OrphanedResearch))
	}
	if findings.OrphanedResearch[0] != "topic-b.md" {
		t.Errorf("OrphanedResearch[0] = %q, want %q", findings.OrphanedResearch[0], "topic-b.md")
	}
}

func TestMineAgentsDir_NoResearchDir(t *testing.T) {
	tmp := t.TempDir()
	findings, err := mineAgentsDir(tmp)
	if err != nil {
		t.Fatalf("mineAgentsDir error: %v", err)
	}
	if findings.TotalResearch != 0 {
		t.Errorf("TotalResearch = %d, want 0", findings.TotalResearch)
	}
	if len(findings.OrphanedResearch) != 0 {
		t.Errorf("OrphanedResearch count = %d, want 0", len(findings.OrphanedResearch))
	}
}

func TestMineGitLog_NoGit(t *testing.T) {
	tmp := t.TempDir()
	findings, err := mineGitLog(tmp, 26*time.Hour)
	if err != nil {
		t.Fatalf("mineGitLog error: %v", err)
	}
	// No git repo → empty findings, no error
	if findings.CommitCount != 0 {
		t.Errorf("CommitCount = %d, want 0", findings.CommitCount)
	}
	if len(findings.CoChangeClusters) != 0 {
		t.Errorf("CoChangeClusters count = %d, want 0", len(findings.CoChangeClusters))
	}
	if len(findings.RecurringFixes) != 0 {
		t.Errorf("RecurringFixes count = %d, want 0", len(findings.RecurringFixes))
	}
}

func TestWriteMineReport_CreatesLatest(t *testing.T) {
	tmp := t.TempDir()
	outDir := filepath.Join(tmp, "mine-output")

	report := &MineReport{
		Timestamp:    time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC),
		SinceSeconds: 93600,
		Sources:      []string{"git"},
		Git: &GitFindings{
			CommitCount: 5,
		},
	}

	if err := writeMineReport(outDir, report); err != nil {
		t.Fatalf("writeMineReport error: %v", err)
	}

	// Check dated file exists
	datedPath := filepath.Join(outDir, "2026-03-01-14.json")
	if _, err := os.Stat(datedPath); os.IsNotExist(err) {
		t.Errorf("dated file not created: %s", datedPath)
	}

	// Check latest.json exists
	latestPath := filepath.Join(outDir, "latest.json")
	data, err := os.ReadFile(latestPath)
	if err != nil {
		t.Fatalf("read latest.json: %v", err)
	}

	var decoded MineReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal latest.json: %v", err)
	}
	if decoded.Git.CommitCount != 5 {
		t.Errorf("decoded CommitCount = %d, want 5", decoded.Git.CommitCount)
	}
	if decoded.SinceSeconds != 93600 {
		t.Errorf("decoded SinceSeconds = %d, want 93600", decoded.SinceSeconds)
	}
}

func TestRunMine_DryRun(t *testing.T) {
	// Save and restore global state
	oldDryRun := dryRun
	oldSources := mineSourcesFlag
	oldSince := mineSince
	oldOutput := mineOutputDir
	defer func() {
		dryRun = oldDryRun
		mineSourcesFlag = oldSources
		mineSince = oldSince
		mineOutputDir = oldOutput
	}()

	tmp := t.TempDir()
	dryRun = true
	mineSourcesFlag = "git,agents"
	mineSince = "26h"
	mineOutputDir = filepath.Join(tmp, "mine-output")

	var buf bytes.Buffer
	cmd := mineCmd
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := runMine(cmd, nil); err != nil {
		t.Fatalf("runMine dry-run error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected dry-run output, got empty string")
	}

	// Verify no files were written
	if _, err := os.Stat(filepath.Join(tmp, "mine-output")); !os.IsNotExist(err) {
		t.Error("dry-run should not create output directory")
	}
}

func TestPrintMineDryRun(t *testing.T) {
	var buf bytes.Buffer
	sources := []string{"git", "agents"}
	window := 26 * time.Hour

	if err := printMineDryRun(&buf, sources, window); err != nil {
		t.Fatalf("printMineDryRun error: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Error("expected output, got empty string")
	}
	if !bytes.Contains([]byte(out), []byte("dry-run")) {
		t.Error("expected output to contain 'dry-run'")
	}
	if !bytes.Contains([]byte(out), []byte("git")) {
		t.Error("expected output to contain 'git'")
	}
}

func TestReadDirContent_MissingDir(t *testing.T) {
	tmp := t.TempDir()
	_, err := readDirContent(filepath.Join(tmp, "nonexistent"))
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}
