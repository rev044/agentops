package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/types"
)

// ===========================================================================
// Coverage tests for rpi_phased_processing.go — targeting zero-coverage functions
// ===========================================================================

// --- gateFailError ---

func TestPhasedCov_GateFailError_Error(t *testing.T) {
	err := &gateFailError{
		Phase:   2,
		Verdict: "BLOCKED",
		Report:  "bd children ag-xyz",
	}
	got := err.Error()
	if !strings.Contains(got, "phase 2") {
		t.Errorf("expected 'phase 2' in error, got %q", got)
	}
	if !strings.Contains(got, "BLOCKED") {
		t.Errorf("expected 'BLOCKED' in error, got %q", got)
	}
}

// --- isPlanFileEpic / planFileFromEpic ---

func TestPhasedCov_IsPlanFileEpic(t *testing.T) {
	tests := []struct {
		epicID string
		want   bool
	}{
		{"plan:.agents/plans/my-plan.md", true},
		{"ag-xyz", false},
		{"plan:", true},
		{"", false},
	}
	for _, tt := range tests {
		if got := isPlanFileEpic(tt.epicID); got != tt.want {
			t.Errorf("isPlanFileEpic(%q) = %v, want %v", tt.epicID, got, tt.want)
		}
	}
}

func TestPhasedCov_PlanFileFromEpic(t *testing.T) {
	tests := []struct {
		epicID string
		want   string
	}{
		{"plan:.agents/plans/my-plan.md", ".agents/plans/my-plan.md"},
		{"plan:", ""},
		{"ag-xyz", "ag-xyz"},
	}
	for _, tt := range tests {
		if got := planFileFromEpic(tt.epicID); got != tt.want {
			t.Errorf("planFileFromEpic(%q) = %q, want %q", tt.epicID, got, tt.want)
		}
	}
}

// --- discoverPlanFile ---

func TestPhasedCov_DiscoverPlanFile_Found(t *testing.T) {
	cwd := t.TempDir()
	plansDir := filepath.Join(cwd, ".agents", "plans")
	if err := os.MkdirAll(plansDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write two plan files with different mod times
	path1 := filepath.Join(plansDir, "plan-a.md")
	if err := os.WriteFile(path1, []byte("plan A"), 0644); err != nil {
		t.Fatal(err)
	}
	// Ensure second file has a later mod time
	time.Sleep(10 * time.Millisecond)
	path2 := filepath.Join(plansDir, "plan-b.md")
	if err := os.WriteFile(path2, []byte("plan B"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := discoverPlanFile(cwd)
	if err != nil {
		t.Fatalf("discoverPlanFile: %v", err)
	}
	if !strings.HasSuffix(got, "plan-b.md") {
		t.Errorf("expected most recent plan file, got %q", got)
	}
}

func TestPhasedCov_DiscoverPlanFile_NoPlans(t *testing.T) {
	cwd := t.TempDir()
	plansDir := filepath.Join(cwd, ".agents", "plans")
	if err := os.MkdirAll(plansDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := discoverPlanFile(cwd)
	if err == nil {
		t.Fatal("expected error for empty plans directory")
	}
}

func TestPhasedCov_DiscoverPlanFile_NoPlanDir(t *testing.T) {
	_, err := discoverPlanFile(t.TempDir())
	if err == nil {
		t.Fatal("expected error when plans dir doesn't exist")
	}
}

func TestPhasedCov_DiscoverPlanFile_SkipNonMD(t *testing.T) {
	cwd := t.TempDir()
	plansDir := filepath.Join(cwd, ".agents", "plans")
	if err := os.MkdirAll(plansDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Only non-md files
	if err := os.WriteFile(filepath.Join(plansDir, "notes.txt"), []byte("notes"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := discoverPlanFile(cwd)
	if err == nil {
		t.Fatal("expected error when no .md files exist")
	}
}

// --- extractCouncilVerdict ---

func TestPhasedCov_ExtractCouncilVerdict_MissingFile(t *testing.T) {
	_, err := extractCouncilVerdict("/nonexistent/report.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// --- findLatestCouncilReport ---

func TestPhasedCov_FindLatestCouncilReport_NoDir(t *testing.T) {
	_, err := findLatestCouncilReport(t.TempDir(), "vibe", time.Time{}, "")
	if err == nil {
		t.Fatal("expected error when council dir doesn't exist")
	}
}

func TestPhasedCov_FindLatestCouncilReport_NoMatches(t *testing.T) {
	cwd := t.TempDir()
	councilDir := filepath.Join(cwd, ".agents", "council")
	if err := os.MkdirAll(councilDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(councilDir, "unrelated.md"), []byte("nope"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := findLatestCouncilReport(cwd, "vibe", time.Time{}, "")
	if err == nil {
		t.Fatal("expected error when no matching reports")
	}
}

func TestPhasedCov_FindLatestCouncilReport_WithEpicID(t *testing.T) {
	cwd := t.TempDir()
	councilDir := filepath.Join(cwd, ".agents", "council")
	if err := os.MkdirAll(councilDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Generic vibe report
	if err := os.WriteFile(filepath.Join(councilDir, "2026-01-01-vibe-generic.md"), []byte("## Council Verdict: PASS"), 0644); err != nil {
		t.Fatal(err)
	}
	// Epic-scoped vibe report
	if err := os.WriteFile(filepath.Join(councilDir, "2026-01-02-vibe-ag-xyz.md"), []byte("## Council Verdict: WARN"), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := findLatestCouncilReport(cwd, "vibe", time.Time{}, "ag-xyz")
	if err != nil {
		t.Fatalf("findLatestCouncilReport: %v", err)
	}
	if !strings.Contains(report, "ag-xyz") {
		t.Errorf("expected epic-scoped report, got %q", report)
	}
}

func TestPhasedCov_FindLatestCouncilReport_NotBeforeFilter(t *testing.T) {
	cwd := t.TempDir()
	councilDir := filepath.Join(cwd, ".agents", "council")
	if err := os.MkdirAll(councilDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write a file with a specific mod time
	reportPath := filepath.Join(councilDir, "2026-02-01-vibe-test.md")
	if err := os.WriteFile(reportPath, []byte("## Council Verdict: PASS"), 0644); err != nil {
		t.Fatal(err)
	}

	// notBefore in the future should exclude it
	_, err := findLatestCouncilReport(cwd, "vibe", time.Now().Add(1*time.Hour), "")
	if err == nil {
		t.Fatal("expected error when notBefore excludes all reports")
	}
}

func TestPhasedCov_FindLatestCouncilReport_SkipDirs(t *testing.T) {
	cwd := t.TempDir()
	councilDir := filepath.Join(cwd, ".agents", "council")
	if err := os.MkdirAll(councilDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a directory that matches the pattern name
	if err := os.MkdirAll(filepath.Join(councilDir, "2026-01-01-vibe-subdir.md"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := findLatestCouncilReport(cwd, "vibe", time.Time{}, "")
	if err == nil {
		t.Fatal("expected error when only matching entry is a directory")
	}
}

// --- matchCouncilEntry ---

func TestPhasedCov_MatchCouncilEntry(t *testing.T) {
	cwd := t.TempDir()
	councilDir := filepath.Join(cwd, ".agents", "council")
	if err := os.MkdirAll(councilDir, 0755); err != nil {
		t.Fatal(err)
	}

	matchingFile := filepath.Join(councilDir, "2026-01-01-vibe-test.md")
	if err := os.WriteFile(matchingFile, []byte("report"), 0644); err != nil {
		t.Fatal(err)
	}
	nonMatchingFile := filepath.Join(councilDir, "2026-01-01-other.md")
	if err := os.WriteFile(nonMatchingFile, []byte("other"), 0644); err != nil {
		t.Fatal(err)
	}
	nonMDFile := filepath.Join(councilDir, "2026-01-01-vibe-test.txt")
	if err := os.WriteFile(nonMDFile, []byte("txt"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(councilDir)
	if err != nil {
		t.Fatal(err)
	}

	matched := 0
	for _, entry := range entries {
		if _, ok := matchCouncilEntry(entry, councilDir, "vibe", time.Time{}); ok {
			matched++
		}
	}
	if matched != 1 {
		t.Errorf("expected 1 match, got %d", matched)
	}
}

// --- extractCouncilFindings ---

func TestPhasedCov_ExtractCouncilFindings_StructuredFormat(t *testing.T) {
	cwd := t.TempDir()
	reportPath := filepath.Join(cwd, "report.md")
	content := `# Report
FINDING: Missing error handling | FIX: Add nil check | REF: cmd/ao/main.go:42
FINDING: Race condition | FIX: Add mutex | REF: cmd/ao/server.go:88
FINDING: Extra 1 | FIX: fix 1 | REF: ref1
FINDING: Extra 2 | FIX: fix 2 | REF: ref2
`
	if err := os.WriteFile(reportPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := extractCouncilFindings(reportPath, 3)
	if err != nil {
		t.Fatalf("extractCouncilFindings: %v", err)
	}
	if len(findings) != 3 {
		t.Errorf("expected 3 findings (max), got %d", len(findings))
	}
	if findings[0].Description != "Missing error handling" {
		t.Errorf("first finding = %q", findings[0].Description)
	}
}

func TestPhasedCov_ExtractCouncilFindings_FallbackFormat(t *testing.T) {
	cwd := t.TempDir()
	reportPath := filepath.Join(cwd, "report.md")
	content := `# Report

## Shared Findings

1. **Missing tests** — No unit tests for new function
2. **Duplicate code** — Same logic in two files
`
	if err := os.WriteFile(reportPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := extractCouncilFindings(reportPath, 5)
	if err != nil {
		t.Fatalf("extractCouncilFindings: %v", err)
	}
	if len(findings) != 2 {
		t.Errorf("expected 2 fallback findings, got %d", len(findings))
	}
	if !strings.Contains(findings[0].Description, "Missing tests") {
		t.Errorf("first finding = %q", findings[0].Description)
	}
}

func TestPhasedCov_ExtractCouncilFindings_MissingFile(t *testing.T) {
	_, err := extractCouncilFindings("/nonexistent/report.md", 5)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPhasedCov_ExtractCouncilFindings_NoFindings(t *testing.T) {
	reportPath := filepath.Join(t.TempDir(), "empty-report.md")
	if err := os.WriteFile(reportPath, []byte("# Report\nNo findings here."), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := extractCouncilFindings(reportPath, 5)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

// --- parseFastPath ---

func TestPhasedCov_ParseFastPath(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{"empty output", "", true},
		{"single issue", "ag-1 [open] task description", true},
		{"two issues", "ag-1 [open] first\nag-2 [open] second", true},
		{"three issues", "ag-1 [open] first\nag-2 [open] second\nag-3 [open] third", false},
		{"one blocked", "ag-1 [blocked] task", false},
		{"two with one blocked", "ag-1 [open] first\nag-2 [blocked] second", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFastPath(tt.output)
			if got != tt.want {
				t.Errorf("parseFastPath = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- parseCrankCompletion ---

func TestPhasedCov_ParseCrankCompletion(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{"empty", "", "DONE"},
		{"all closed", "ag-1 [closed] done\nag-2 [closed] done", "DONE"},
		{"all closed checkmark", "ag-1 ✓ done\nag-2 ✓ done", "DONE"},
		{"partial", "ag-1 [closed] done\nag-2 [open] wip", "PARTIAL"},
		{"blocked", "ag-1 [blocked] waiting\nag-2 [open] wip", "BLOCKED"},
		{"all open", "ag-1 [open] wip\nag-2 [open] wip", "PARTIAL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCrankCompletion(tt.output)
			if got != tt.want {
				t.Errorf("parseCrankCompletion = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- parseLatestEpicIDFromJSON ---

func TestPhasedCov_ParseLatestEpicIDFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    string
		wantErr bool
	}{
		{
			"single entry",
			`[{"id": "ag-abc"}]`,
			"ag-abc",
			false,
		},
		{
			"multiple entries returns last",
			`[{"id": "ag-old"}, {"id": "ag-new"}]`,
			"ag-new",
			false,
		},
		{
			"empty array",
			`[]`,
			"",
			true,
		},
		{
			"invalid JSON",
			`not json`,
			"",
			true,
		},
		{
			"entries with empty IDs",
			`[{"id": ""}, {"id": "  "}, {"id": "ag-valid"}]`,
			"ag-valid",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLatestEpicIDFromJSON([]byte(tt.data))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// --- parseLatestEpicIDFromText ---

func TestPhasedCov_ParseLatestEpicIDFromText(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    string
		wantErr bool
	}{
		{
			"standard ID",
			"ag-abc123 [open] Epic title",
			"ag-abc123",
			false,
		},
		{
			"multiple lines returns last",
			"ag-old [closed] Old epic\nag-new [open] New epic",
			"ag-new",
			false,
		},
		{
			"bracketed ID",
			"[ag-xyz] Some description",
			"ag-xyz",
			false,
		},
		{
			"no match",
			"no epic ID here",
			"",
			true,
		},
		{
			"empty output",
			"",
			"",
			true,
		},
		{
			"custom prefix",
			"bd-123 [open] custom prefix epic",
			"bd-123",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLatestEpicIDFromText(tt.output)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// --- generatePhaseSummary ---

func TestPhasedCov_GeneratePhaseSummary(t *testing.T) {
	tests := []struct {
		name     string
		state    *phasedState
		phaseNum int
		wantSub  string
	}{
		{
			"discovery phase",
			&phasedState{
				Goal:     "test goal",
				EpicID:   "ag-abc",
				Verdicts: map[string]string{"pre_mortem": "PASS"},
			},
			1,
			"Discovery completed",
		},
		{
			"discovery with fast path",
			&phasedState{
				Goal:     "test goal",
				EpicID:   "ag-abc",
				FastPath: true,
				Verdicts: map[string]string{},
			},
			1,
			"fast path",
		},
		{
			"discovery no epic",
			&phasedState{
				Goal:     "test goal",
				Verdicts: map[string]string{},
			},
			1,
			"Discovery completed",
		},
		{
			"implementation phase",
			&phasedState{EpicID: "ag-abc", Verdicts: map[string]string{}},
			2,
			"Crank completed",
		},
		{
			"validation phase with verdicts",
			&phasedState{
				EpicID:   "ag-abc",
				Verdicts: map[string]string{"vibe": "PASS", "post_mortem": "WARN"},
			},
			3,
			"Vibe verdict: PASS",
		},
		{
			"validation no verdicts",
			&phasedState{
				Verdicts: map[string]string{},
			},
			3,
			"learnings",
		},
		{
			"unknown phase",
			&phasedState{Verdicts: map[string]string{}},
			99,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generatePhaseSummary(tt.state, tt.phaseNum)
			if tt.wantSub != "" && !strings.Contains(got, tt.wantSub) {
				t.Errorf("generatePhaseSummary = %q, want to contain %q", got, tt.wantSub)
			}
			if tt.wantSub == "" && got != "" {
				t.Errorf("expected empty string for unknown phase, got %q", got)
			}
		})
	}
}

// --- writePhaseSummary ---

func TestPhasedCov_WritePhaseSummary_ClaudeWrote(t *testing.T) {
	cwd := t.TempDir()
	rpiDir := filepath.Join(cwd, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0755); err != nil {
		t.Fatal(err)
	}
	summaryPath := filepath.Join(rpiDir, "phase-1-summary.md")
	if err := os.WriteFile(summaryPath, []byte("Claude wrote this"), 0644); err != nil {
		t.Fatal(err)
	}

	state := &phasedState{
		Goal:     "test",
		Verdicts: map[string]string{},
	}
	writePhaseSummary(cwd, state, 1)

	// Verify Claude's summary was preserved
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Claude wrote this" {
		t.Error("Claude summary should not be overwritten")
	}
}

func TestPhasedCov_WritePhaseSummary_FallbackWritten(t *testing.T) {
	cwd := t.TempDir()
	state := &phasedState{
		Goal:     "test goal",
		EpicID:   "ag-xyz",
		Verdicts: map[string]string{"pre_mortem": "PASS"},
	}
	writePhaseSummary(cwd, state, 1)

	summaryPath := filepath.Join(cwd, ".agents", "rpi", "phase-1-summary.md")
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("expected summary to be written: %v", err)
	}
	if !strings.Contains(string(data), "Discovery completed") {
		t.Errorf("summary = %q, want 'Discovery completed'", string(data))
	}
}

// --- handoffDetected ---

func TestPhasedCov_HandoffDetected(t *testing.T) {
	cwd := t.TempDir()
	rpiDir := filepath.Join(cwd, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0755); err != nil {
		t.Fatal(err)
	}

	// No handoff file
	if handoffDetected(cwd, 1) {
		t.Error("expected false when no handoff file")
	}

	// Create handoff file (.json, not .md)
	if err := os.WriteFile(filepath.Join(rpiDir, "phase-1-handoff.json"), []byte(`{"phase":1}`), 0644); err != nil {
		t.Fatal(err)
	}
	if !handoffDetected(cwd, 1) {
		t.Error("expected true when handoff file exists")
	}
}

// --- cleanPhaseSummaries ---

func TestPhasedCov_CleanPhaseSummaries(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create summary, handoff, and result files for each phase
	for i := 1; i <= 3; i++ {
		for _, suffix := range []string{"summary.md", "handoff.md", "result.json"} {
			path := filepath.Join(stateDir, strings.Replace("phase-N-"+suffix, "N", string(rune('0'+i)), 1))
			if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	cleanPhaseSummaries(stateDir)

	// Verify all files are cleaned
	for i := 1; i <= 3; i++ {
		for _, suffix := range []string{"summary.md", "handoff.md", "result.json"} {
			path := filepath.Join(stateDir, strings.Replace("phase-N-"+suffix, "N", string(rune('0'+i)), 1))
			if _, err := os.Stat(path); err == nil {
				t.Errorf("expected %s to be removed", path)
			}
		}
	}
}

// --- writePhaseResult ---

func TestPhasedCov_WritePhaseResult(t *testing.T) {
	cwd := t.TempDir()
	result := &phaseResult{
		SchemaVersion: 1,
		RunID:         "test-run",
		Phase:         2,
		PhaseName:     "implementation",
		Status:        "completed",
		StartedAt:     time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		CompletedAt:   time.Now().Format(time.RFC3339),
		Artifacts:     map[string]string{"summary": "phase-2-summary.md"},
		Verdicts:      map[string]string{"vibe": "PASS"},
	}

	if err := writePhaseResult(cwd, result); err != nil {
		t.Fatalf("writePhaseResult: %v", err)
	}

	resultPath := filepath.Join(cwd, ".agents", "rpi", "phase-2-result.json")
	data, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}

	var readResult phaseResult
	if err := json.Unmarshal(data, &readResult); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if readResult.RunID != "test-run" {
		t.Errorf("RunID = %q, want test-run", readResult.RunID)
	}
	if readResult.Phase != 2 {
		t.Errorf("Phase = %d, want 2", readResult.Phase)
	}
}

// --- validatePriorPhaseResult ---

func TestPhasedCov_ValidatePriorPhaseResult_Completed(t *testing.T) {
	cwd := t.TempDir()
	result := &phaseResult{
		SchemaVersion: 1,
		RunID:         "test",
		Phase:         1,
		Status:        "completed",
		StartedAt:     time.Now().Format(time.RFC3339),
	}
	if err := writePhaseResult(cwd, result); err != nil {
		t.Fatal(err)
	}

	if err := validatePriorPhaseResult(cwd, 1); err != nil {
		t.Errorf("expected no error for completed phase, got: %v", err)
	}
}

func TestPhasedCov_ValidatePriorPhaseResult_Failed(t *testing.T) {
	cwd := t.TempDir()
	result := &phaseResult{
		SchemaVersion: 1,
		RunID:         "test",
		Phase:         1,
		Status:        "failed",
		StartedAt:     time.Now().Format(time.RFC3339),
	}
	if err := writePhaseResult(cwd, result); err != nil {
		t.Fatal(err)
	}

	err := validatePriorPhaseResult(cwd, 1)
	if err == nil {
		t.Fatal("expected error for failed phase")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("error = %q, want 'failed'", err.Error())
	}
}

func TestPhasedCov_ValidatePriorPhaseResult_Missing(t *testing.T) {
	err := validatePriorPhaseResult(t.TempDir(), 1)
	if err == nil {
		t.Fatal("expected error for missing result")
	}
}

func TestPhasedCov_ValidatePriorPhaseResult_Corrupt(t *testing.T) {
	cwd := t.TempDir()
	rpiDir := filepath.Join(cwd, ".agents", "rpi")
	if err := os.MkdirAll(rpiDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rpiDir, "phase-1-result.json"), []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	err := validatePriorPhaseResult(cwd, 1)
	if err == nil {
		t.Fatal("expected error for corrupt result")
	}
}

// --- savePhasedState / loadPhasedState / parsePhasedState ---

func TestPhasedCov_SaveAndLoadPhasedState(t *testing.T) {
	cwd := t.TempDir()
	state := &phasedState{
		SchemaVersion: 1,
		Goal:          "round trip test",
		Phase:         2,
		Cycle:         1,
		RunID:         "save-load-test",
		EpicID:        "ag-test",
		Verdicts:      map[string]string{"pre_mortem": "PASS"},
		Attempts:      map[string]int{"phase_1": 0},
		StartedAt:     time.Now().Format(time.RFC3339),
		StartPhase:    1,
	}

	if err := savePhasedState(cwd, state); err != nil {
		t.Fatalf("savePhasedState: %v", err)
	}

	loaded, err := loadPhasedState(cwd)
	if err != nil {
		t.Fatalf("loadPhasedState: %v", err)
	}
	if loaded.RunID != "save-load-test" {
		t.Errorf("RunID = %q, want save-load-test", loaded.RunID)
	}
	if loaded.Goal != "round trip test" {
		t.Errorf("Goal = %q", loaded.Goal)
	}
}

func TestPhasedCov_LoadPhasedState_Missing(t *testing.T) {
	_, err := loadPhasedState(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing state")
	}
}

func TestPhasedCov_ParsePhasedState_NilMaps(t *testing.T) {
	data := `{"schema_version": 1, "run_id": "test", "phase": 2}`
	state, err := parsePhasedState([]byte(data))
	if err != nil {
		t.Fatalf("parsePhasedState: %v", err)
	}
	if state.Verdicts == nil {
		t.Error("expected non-nil Verdicts")
	}
	if state.Attempts == nil {
		t.Error("expected non-nil Attempts")
	}
}

func TestPhasedCov_ParsePhasedState_EmptyGoal(t *testing.T) {
	data := `{"schema_version": 1, "run_id": "test", "goal": "  "}`
	state, err := parsePhasedState([]byte(data))
	if err != nil {
		t.Fatalf("parsePhasedState: %v", err)
	}
	if state.Goal != "unknown-goal" {
		t.Errorf("expected 'unknown-goal', got %q", state.Goal)
	}
}

func TestPhasedCov_ParsePhasedState_ZeroPhaseDefaults(t *testing.T) {
	data := `{"schema_version": 1, "run_id": "test", "goal": "test", "phase": 0, "cycle": 0}`
	state, err := parsePhasedState([]byte(data))
	if err != nil {
		t.Fatalf("parsePhasedState: %v", err)
	}
	if state.Phase != 1 {
		t.Errorf("expected Phase=1, got %d", state.Phase)
	}
	if state.Cycle != 1 {
		t.Errorf("expected Cycle=1, got %d", state.Cycle)
	}
}

func TestPhasedCov_ParsePhasedState_StartPhaseDefaults(t *testing.T) {
	data := `{"schema_version": 1, "run_id": "test", "goal": "test", "phase": 2}`
	state, err := parsePhasedState([]byte(data))
	if err != nil {
		t.Fatalf("parsePhasedState: %v", err)
	}
	if state.StartPhase != 2 {
		t.Errorf("expected StartPhase=2 (defaults to Phase), got %d", state.StartPhase)
	}
}

func TestPhasedCov_ParsePhasedState_InvalidStartPhase(t *testing.T) {
	data := `{"schema_version": 1, "run_id": "test", "goal": "test", "phase": 2, "start_phase": 99}`
	state, err := parsePhasedState([]byte(data))
	if err != nil {
		t.Fatalf("parsePhasedState: %v", err)
	}
	if state.StartPhase != 2 {
		t.Errorf("expected StartPhase=2 (invalid 99 clamped to Phase), got %d", state.StartPhase)
	}
}

func TestPhasedCov_ParsePhasedState_Invalid(t *testing.T) {
	_, err := parsePhasedState([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// --- rpiRunRegistryDir ---

func TestPhasedCov_RPIRunRegistryDir(t *testing.T) {
	got := rpiRunRegistryDir("/repo", "run-abc")
	if !strings.Contains(got, filepath.Join("runs", "run-abc")) {
		t.Errorf("expected runs/run-abc in path, got %q", got)
	}

	empty := rpiRunRegistryDir("/repo", "")
	if empty != "" {
		t.Errorf("expected empty for empty runID, got %q", empty)
	}
}

// --- updateRunHeartbeat / readRunHeartbeat ---

func TestPhasedCov_UpdateAndReadRunHeartbeat(t *testing.T) {
	cwd := t.TempDir()
	runID := "hb-test"

	updateRunHeartbeat(cwd, runID)

	hb := readRunHeartbeat(cwd, runID)
	if hb.IsZero() {
		t.Error("expected non-zero heartbeat after update")
	}
	if time.Since(hb) > 5*time.Second {
		t.Errorf("heartbeat too old: %v", time.Since(hb))
	}
}

func TestPhasedCov_UpdateRunHeartbeat_EmptyRunID(t *testing.T) {
	// Should not panic or create files
	updateRunHeartbeat(t.TempDir(), "")
}

func TestPhasedCov_ReadRunHeartbeat_EmptyRunID(t *testing.T) {
	hb := readRunHeartbeat(t.TempDir(), "")
	if !hb.IsZero() {
		t.Error("expected zero time for empty runID")
	}
}

func TestPhasedCov_ReadRunHeartbeat_Missing(t *testing.T) {
	hb := readRunHeartbeat(t.TempDir(), "nonexistent")
	if !hb.IsZero() {
		t.Error("expected zero time for missing heartbeat")
	}
}

func TestPhasedCov_ReadRunHeartbeat_RFC3339Format(t *testing.T) {
	cwd := t.TempDir()
	runID := "rfc3339-hb"
	runDir := filepath.Join(cwd, ".agents", "rpi", "runs", runID)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	if err := os.WriteFile(filepath.Join(runDir, "heartbeat.txt"), []byte(ts+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	hb := readRunHeartbeat(cwd, runID)
	if hb.IsZero() {
		t.Error("expected non-zero heartbeat for RFC3339 format")
	}
}

func TestPhasedCov_ReadRunHeartbeat_InvalidFormat(t *testing.T) {
	cwd := t.TempDir()
	runID := "bad-hb"
	runDir := filepath.Join(cwd, ".agents", "rpi", "runs", runID)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "heartbeat.txt"), []byte("not-a-timestamp\n"), 0644); err != nil {
		t.Fatal(err)
	}

	hb := readRunHeartbeat(cwd, runID)
	if !hb.IsZero() {
		t.Error("expected zero for invalid timestamp format")
	}
}

// --- loadLatestRunRegistryState ---

func TestPhasedCov_LoadLatestRunRegistryState_NoRunsDir(t *testing.T) {
	_, err := loadLatestRunRegistryState(t.TempDir())
	if err == nil {
		t.Fatal("expected error for non-existent runs dir")
	}
}

func TestPhasedCov_LoadLatestRunRegistryState_MultipleRuns(t *testing.T) {
	cwd := t.TempDir()
	runsDir := filepath.Join(cwd, ".agents", "rpi", "runs")

	// Create two runs; the second should be newer
	for _, spec := range []struct {
		id    string
		phase int
	}{
		{"old-run", 1},
		{"new-run", 2},
	} {
		runDir := filepath.Join(runsDir, spec.id)
		if err := os.MkdirAll(runDir, 0755); err != nil {
			t.Fatal(err)
		}
		state := map[string]any{
			"schema_version": 1,
			"run_id":         spec.id,
			"goal":           "test",
			"phase":          spec.phase,
		}
		data, _ := json.Marshal(state)
		if err := os.WriteFile(filepath.Join(runDir, phasedStateFile), data, 0644); err != nil {
			t.Fatal(err)
		}
		// Small sleep to ensure different mod times
		time.Sleep(10 * time.Millisecond)
	}

	state, err := loadLatestRunRegistryState(cwd)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if state.RunID != "new-run" {
		t.Errorf("expected newest run 'new-run', got %q", state.RunID)
	}
}

// --- writePhasedStateAtomic ---

func TestPhasedCov_WritePhasedStateAtomic(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test-state.json")

	data := []byte(`{"test": true}`)
	if err := writePhasedStateAtomic(path, data); err != nil {
		t.Fatalf("writePhasedStateAtomic: %v", err)
	}

	read, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(read) != string(data) {
		t.Errorf("expected %q, got %q", string(data), string(read))
	}
}

// --- logPhaseTransition ---

func TestPhasedCov_LogPhaseTransition_WithRunID(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logPhaseTransition(logPath, "run-abc", "discovery", "epic extracted")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "[run-abc]") {
		t.Errorf("expected [run-abc] in log, got %q", string(data))
	}
	if !strings.Contains(string(data), "discovery: epic extracted") {
		t.Errorf("expected phase:details in log, got %q", string(data))
	}
}

func TestPhasedCov_LogPhaseTransition_WithoutRunID(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logPhaseTransition(logPath, "", "start", "goal=\"test\"")

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "[]") {
		t.Errorf("should not have empty brackets for no runID: %q", content)
	}
	if !strings.Contains(content, "start: goal=") {
		t.Errorf("expected start phase in log, got %q", content)
	}
}

// --- logFailureContext ---

func TestPhasedCov_LogFailureContext(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "failure.log")

	logFailureContext(logPath, "run-xyz", "validation", os.ErrNotExist)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "FAILURE_CONTEXT") {
		t.Errorf("expected FAILURE_CONTEXT in log, got %q", content)
	}
}

// --- deriveRepoRootFromRPIOrchestrationLog ---

func TestPhasedCov_DeriveRepoRootFromRPIOrchestrationLog(t *testing.T) {
	tests := []struct {
		logPath string
		wantOK  bool
	}{
		{"/repo/.agents/rpi/phased-orchestration.log", true},
		{"/tmp/some/random/file.log", false},
		{"/repo/.agents/wrong/phased-orchestration.log", false},
	}
	for _, tt := range tests {
		root, ok := deriveRepoRootFromRPIOrchestrationLog(tt.logPath)
		if ok != tt.wantOK {
			t.Errorf("deriveRepoRoot(%q) ok=%v, want %v", tt.logPath, ok, tt.wantOK)
		}
		if ok && root == "" {
			t.Error("expected non-empty root")
		}
	}
}

// --- ledgerActionFromDetails ---

func TestPhasedCov_LedgerActionFromDetails(t *testing.T) {
	tests := []struct {
		details string
		want    string
	}{
		{"started goal=test", "started"},
		{"completed in 5m", "completed"},
		{"failed: exit code 1", "failed"},
		{"fatal: build failed", "fatal"},
		{"retry attempt 2/3", "retry"},
		{"dry-run: skipped", "dry-run"},
		{"handoff context degradation", "handoff"},
		{"epic=ag-xyz verdicts=map[]", "summary"},
		{"something random", "something"},
		{"", "event"},
		{"   ", "event"},
	}
	for _, tt := range tests {
		got := ledgerActionFromDetails(tt.details)
		if got != tt.want {
			t.Errorf("ledgerActionFromDetails(%q) = %q, want %q", tt.details, got, tt.want)
		}
	}
}

// --- classifyByPhase ---

func TestPhasedCov_ClassifyByPhase(t *testing.T) {
	tests := []struct {
		phase   int
		verdict string
		want    types.MemRLFailureClass
	}{
		{1, "FAIL", types.MemRLFailureClassPreMortemFail},
		{1, "WARN", ""},
		{2, "BLOCKED", types.MemRLFailureClassCrankBlocked},
		{2, "PARTIAL", types.MemRLFailureClassCrankPartial},
		{2, "DONE", ""},
		{3, "FAIL", types.MemRLFailureClassVibeFail},
		{3, "PASS", ""},
		{4, "FAIL", ""},
	}
	for _, tt := range tests {
		got := classifyByPhase(tt.phase, tt.verdict)
		if got != tt.want {
			t.Errorf("classifyByPhase(%d, %q) = %q, want %q", tt.phase, tt.verdict, got, tt.want)
		}
	}
}

// --- classifyByVerdict ---

func TestPhasedCov_ClassifyByVerdict(t *testing.T) {
	tests := []struct {
		verdict string
		want    types.MemRLFailureClass
	}{
		{string(failReasonTimeout), types.MemRLFailureClassPhaseTimeout},
		{string(failReasonStall), types.MemRLFailureClassPhaseStall},
		{string(failReasonExit), types.MemRLFailureClassPhaseExitError},
		{"CUSTOM", types.MemRLFailureClass("custom")},
		{"", types.MemRLFailureClass("")},
	}
	for _, tt := range tests {
		got := classifyByVerdict(tt.verdict)
		if got != tt.want {
			t.Errorf("classifyByVerdict(%q) = %q, want %q", tt.verdict, got, tt.want)
		}
	}
}

// --- postPhaseProcessing with unknown phase ---

func TestPhasedCov_PostPhaseProcessing_UnknownPhase(t *testing.T) {
	state := &phasedState{Verdicts: make(map[string]string), Attempts: make(map[string]int)}
	err := postPhaseProcessing(t.TempDir(), state, 99, filepath.Join(t.TempDir(), "test.log"))
	if err != nil {
		t.Errorf("expected nil for unknown phase, got: %v", err)
	}
}

// --- maybeUpdateLiveStatus ---

func TestPhasedCov_MaybeUpdateLiveStatus_Disabled(t *testing.T) {
	state := &phasedState{
		Opts: phasedEngineOptions{LiveStatus: false},
	}
	// Should not panic even with zero-value arguments
	maybeUpdateLiveStatus(state, "", nil, 1, "running", 0, "")
}
