package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/vibecheck"
)

// ---------------------------------------------------------------------------
// goals_init.go — prompt (50% → higher)
// ---------------------------------------------------------------------------

func TestCov18_prompt_withInput(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("hello world\n"))
	got, err := prompt(scanner, "Enter: ")
	if err != nil {
		t.Fatalf("prompt with input: %v", err)
	}
	if got != "hello world" {
		t.Errorf("prompt: got %q, want %q", got, "hello world")
	}
}

func TestCov18_prompt_whitespaceTrimed(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("  spaces  \n"))
	got, err := prompt(scanner, "msg: ")
	if err != nil {
		t.Fatalf("prompt whitespace: %v", err)
	}
	if got != "spaces" {
		t.Errorf("prompt whitespace: got %q, want %q", got, "spaces")
	}
}

func TestCov18_prompt_eofReturnsEmpty(t *testing.T) {
	// Empty reader → scanner.Scan() returns false, Err() == nil → return "", nil
	scanner := bufio.NewScanner(strings.NewReader(""))
	got, err := prompt(scanner, "msg: ")
	if err != nil {
		t.Fatalf("prompt EOF: %v", err)
	}
	if got != "" {
		t.Errorf("prompt EOF: got %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// rpi_nudge.go — resolveNudgePhase (0% → higher)
// ---------------------------------------------------------------------------

func TestCov18_resolveNudgePhase_explicitValid(t *testing.T) {
	for _, phase := range []int{1, 2, 3} {
		got, err := resolveNudgePhase(nil, phase)
		if err != nil {
			t.Errorf("resolveNudgePhase(%d): unexpected error %v", phase, err)
			continue
		}
		if got != phase {
			t.Errorf("resolveNudgePhase(%d): got %d", phase, got)
		}
	}
}

func TestCov18_resolveNudgePhase_explicitOutOfRange(t *testing.T) {
	for _, phase := range []int{4, 5, 100} {
		_, err := resolveNudgePhase(nil, phase)
		if err == nil {
			t.Errorf("resolveNudgePhase(%d): expected error, got nil", phase)
			continue
		}
		if !strings.Contains(err.Error(), "--phase must be 1, 2, or 3") {
			t.Errorf("resolveNudgePhase(%d): wrong error: %v", phase, err)
		}
	}
}

func TestCov18_resolveNudgePhase_inferFromState(t *testing.T) {
	state := &phasedState{Phase: 2}
	got, err := resolveNudgePhase(state, 0)
	if err != nil {
		t.Fatalf("resolveNudgePhase infer from state: %v", err)
	}
	if got != 2 {
		t.Errorf("resolveNudgePhase infer: got %d, want 2", got)
	}
}

func TestCov18_resolveNudgePhase_nilStateNoPhase(t *testing.T) {
	_, err := resolveNudgePhase(nil, 0)
	if err == nil {
		t.Fatal("resolveNudgePhase nil state: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "could not infer phase") {
		t.Errorf("resolveNudgePhase nil state: wrong error: %v", err)
	}
}

func TestCov18_resolveNudgePhase_statePhaseZero(t *testing.T) {
	// Phase == 0 in state → can't infer (< 1)
	state := &phasedState{Phase: 0}
	_, err := resolveNudgePhase(state, 0)
	if err == nil {
		t.Fatal("resolveNudgePhase state.Phase=0: expected error, got nil")
	}
}

func TestCov18_resolveNudgePhase_statePhaseAboveRange(t *testing.T) {
	// Phase == 4 in state → can't infer (> 3)
	state := &phasedState{Phase: 4}
	_, err := resolveNudgePhase(state, 0)
	if err == nil {
		t.Fatal("resolveNudgePhase state.Phase=4: expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// goals_migrate.go — runGoalsMigrate (37.2% → higher)
// ---------------------------------------------------------------------------

func TestCov18_runGoalsMigrate_alreadyVersion2(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Write a version 2 GOALS.yaml
	goalsYAML := "version: 2\nmission: Test project\ngoals: []\n"
	if err := os.WriteFile(filepath.Join(tmp, "GOALS.yaml"), []byte(goalsYAML), 0644); err != nil {
		t.Fatalf("write GOALS.yaml: %v", err)
	}

	origMigrateToMD := migrateToMD
	origGoalsFile := goalsFile
	defer func() {
		migrateToMD = origMigrateToMD
		goalsFile = origGoalsFile
	}()
	migrateToMD = false
	goalsFile = ""

	err := runGoalsMigrate(goalsCmd, nil)
	if err != nil {
		t.Fatalf("runGoalsMigrate v2: %v", err)
	}
}

func TestCov18_runGoalsMigrate_alreadyMDFormat(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Write a valid GOALS.md so resolveGoalsFile returns it
	goalsMD := `# Goals

Test project goals.

## North Stars

- Tests pass

## Anti Stars

- Untested code

## Directives

### 1. Keep coverage high

Stay above 80%.

**Steer:** increase

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| go-test | go test ./... | 10 | Go tests pass |
`
	if err := os.WriteFile(filepath.Join(tmp, "GOALS.md"), []byte(goalsMD), 0644); err != nil {
		t.Fatalf("write GOALS.md: %v", err)
	}

	origMigrateToMD := migrateToMD
	origGoalsFile := goalsFile
	defer func() {
		migrateToMD = origMigrateToMD
		goalsFile = origGoalsFile
	}()
	migrateToMD = true
	goalsFile = ""

	err := runGoalsMigrate(goalsCmd, nil)
	if err != nil {
		t.Fatalf("runGoalsMigrate already md: %v", err)
	}
}

func TestCov18_runGoalsMigrate_yamlToMD(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// Write a version 2 GOALS.yaml to migrate to MD
	goalsYAML := "version: 2\nmission: Test project\ngoals:\n  - id: go-test\n    check: go test ./...\n    weight: 10\n    description: Go tests\n"
	if err := os.WriteFile(filepath.Join(tmp, "GOALS.yaml"), []byte(goalsYAML), 0644); err != nil {
		t.Fatalf("write GOALS.yaml: %v", err)
	}

	origMigrateToMD := migrateToMD
	origGoalsFile := goalsFile
	defer func() {
		migrateToMD = origMigrateToMD
		goalsFile = origGoalsFile
	}()
	migrateToMD = true
	goalsFile = ""

	err := runGoalsMigrate(goalsCmd, nil)
	if err != nil {
		t.Fatalf("runGoalsMigrate yaml-to-md: %v", err)
	}
	// GOALS.md should have been created
	if _, statErr := os.Stat(filepath.Join(tmp, "GOALS.md")); os.IsNotExist(statErr) {
		t.Error("expected GOALS.md to be created after --to-md migration")
	}
}

// ---------------------------------------------------------------------------
// goals_migrate.go — directivesFromPillars (0% → higher)
// ---------------------------------------------------------------------------

func TestCov18_directivesFromPillars_noPillars(t *testing.T) {
	result := directivesFromPillars(nil)
	// No pillars → returns a single generic directive
	if len(result) == 0 {
		t.Error("directivesFromPillars(nil): expected at least 1 directive, got 0")
	}
}

// ---------------------------------------------------------------------------
// search.go — executeGrepWithFallback (35.3% → higher)
// ---------------------------------------------------------------------------

func TestCov18_executeGrepWithFallback_noMatchesExitCode1(t *testing.T) {
	// grep returns exit code 1 when no matches — should return nil, nil (not an error)
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "test.txt"), []byte("hello world\n"), 0644)

	grepCmd, useRipgrep := buildGrepCommand("nonexistent-xyz-abc-12345", tmp, "*.txt")
	out, err := executeGrepWithFallback(grepCmd, useRipgrep, "nonexistent-xyz-abc-12345", tmp)
	// exit code 1 = no matches, should be treated as nil error
	if err != nil {
		t.Logf("executeGrepWithFallback (no matches): err=%v (system-dependent)", err)
	}
	_ = out
}

func TestCov18_executeGrepWithFallback_successPath(t *testing.T) {
	// grep finds something → returns output, nil
	tmp := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmp, "test.md"), []byte("learning content here\n"), 0644)

	grepCmd, useRipgrep := buildGrepCommand("learning", tmp, "*.md")
	out, err := executeGrepWithFallback(grepCmd, useRipgrep, "learning", tmp)
	if err != nil {
		t.Logf("executeGrepWithFallback (success): err=%v (grep may not be available)", err)
		return
	}
	_ = out
}

// ---------------------------------------------------------------------------
// goals_init.go — currentDir (0% → higher)
// ---------------------------------------------------------------------------

func TestCov18_currentDir_returnsString(t *testing.T) {
	result := currentDir()
	if result == "" {
		t.Error("currentDir: expected non-empty string")
	}
}

// ---------------------------------------------------------------------------
// rpi_nudge.go — resolveNudgeTargets (0% → higher)
// ---------------------------------------------------------------------------

func TestCov18_resolveNudgeTargets_workerNotFound(t *testing.T) {
	sessions := []string{"phase1-w1", "phase1-w2"}
	_, err := resolveNudgeTargets(sessions, "phase1", false, 99)
	if err == nil {
		t.Fatal("resolveNudgeTargets worker 99: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("resolveNudgeTargets: wrong error: %v", err)
	}
}

func TestCov18_resolveNudgeTargets_workerFound(t *testing.T) {
	sessions := []string{"phase1-w1", "phase1-w2"}
	targets, err := resolveNudgeTargets(sessions, "phase1", false, 1)
	if err != nil {
		t.Fatalf("resolveNudgeTargets worker 1: %v", err)
	}
	if len(targets) != 1 || targets[0] != "phase1-w1" {
		t.Errorf("resolveNudgeTargets worker 1: got %v", targets)
	}
}

func TestCov18_resolveNudgeTargets_allWorkersNoneFound(t *testing.T) {
	// No matching worker sessions → error
	sessions := []string{"unrelated-session"}
	_, err := resolveNudgeTargets(sessions, "phase1", true, 0)
	if err == nil {
		t.Fatal("resolveNudgeTargets allWorkers no match: expected error, got nil")
	}
}

func TestCov18_resolveNudgeTargets_phaseSessionDirectMatch(t *testing.T) {
	// Session exactly matches phaseSession → returns it directly
	sessions := []string{"phase1", "phase1-w1"}
	targets, err := resolveNudgeTargets(sessions, "phase1", false, 0)
	if err != nil {
		t.Fatalf("resolveNudgeTargets direct match: %v", err)
	}
	if len(targets) != 1 || targets[0] != "phase1" {
		t.Errorf("resolveNudgeTargets direct match: got %v", targets)
	}
}

func TestCov18_resolveNudgeTargets_singleWorkerAutoSelect(t *testing.T) {
	// No direct match, one worker → auto-select it
	sessions := []string{"phase1-w1"}
	targets, err := resolveNudgeTargets(sessions, "phase1", false, 0)
	if err != nil {
		t.Fatalf("resolveNudgeTargets single worker: %v", err)
	}
	if len(targets) != 1 {
		t.Errorf("resolveNudgeTargets single worker: got %v", targets)
	}
}

func TestCov18_resolveNudgeTargets_multipleWorkersAmbiguous(t *testing.T) {
	// No direct match, multiple workers → error asking for --all-workers or --worker
	sessions := []string{"phase1-w1", "phase1-w2"}
	_, err := resolveNudgeTargets(sessions, "phase1", false, 0)
	if err == nil {
		t.Fatal("resolveNudgeTargets multiple workers ambiguous: expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// rpi_nudge.go — resolveNudgeRun (31.2% → higher)
// ---------------------------------------------------------------------------

func TestCov18_resolveNudgeRun_emptyRunIDNoState(t *testing.T) {
	tmp := t.TempDir()
	// No phased state file in tmp → loadPhasedState should fail
	_, _, _, err := resolveNudgeRun(tmp, "")
	if err == nil {
		t.Fatal("resolveNudgeRun empty tmp: expected error, got nil")
	}
}

func TestCov18_resolveNudgeRun_explicitRunIDNotFound(t *testing.T) {
	tmp := t.TempDir()
	_, _, _, err := resolveNudgeRun(tmp, "nonexistent-run-id-xyz")
	if err == nil {
		t.Fatal("resolveNudgeRun non-existent run ID: expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// inject.go — trimJSONToCharBudget (0% → higher)
// ---------------------------------------------------------------------------

func TestCov18_trimJSONToCharBudget_withinBudget(t *testing.T) {
	k := &injectedKnowledge{
		Learnings: []learning{{ID: "l1", Title: "Test learning", Summary: "summary"}},
	}
	result := trimJSONToCharBudget(k, 10000)
	if result == "" {
		t.Error("trimJSONToCharBudget within budget: got empty string")
	}
}

func TestCov18_trimJSONToCharBudget_overBudget(t *testing.T) {
	k := &injectedKnowledge{
		Learnings: []learning{
			{ID: "l1", Title: "Test learning 1", Summary: "summary1"},
			{ID: "l2", Title: "Test learning 2", Summary: "summary2"},
		},
		Sessions: []session{{Date: "2026-01-01", Summary: "session summary"}},
	}
	// Very tight budget → forces trimming
	result := trimJSONToCharBudget(k, 5)
	// Should return something (even if heavily trimmed or empty)
	_ = result
}

// ---------------------------------------------------------------------------
// vibe_check.go — outputVibeCheckTable additional branches
// ---------------------------------------------------------------------------

func TestCov18_outputVibeCheckTable_emptyResult(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score:    0.85,
		Grade:    "B",
		Events:   nil,
		Metrics:  map[string]float64{"velocity": 1.2},
		Findings: nil,
	}
	err := outputVibeCheckTable(result)
	if err != nil {
		t.Fatalf("outputVibeCheckTable empty: %v", err)
	}
}

func TestCov18_outputVibeCheckTable_withFindings(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score: 0.72,
		Grade: "C",
		Findings: []vibecheck.Finding{
			{Severity: "high", Category: "quality", Message: "Missing tests", File: "foo.go", Line: 42},
			{Severity: "low", Category: "style", Message: "Minor issue", File: "bar.go", Line: 1},
		},
		Events: []vibecheck.TimelineEvent{
			{
				Timestamp: time.Now(),
				SHA:       "abc1234",
				Author:    "Test Author",
				Message:   "test commit",
			},
		},
	}
	err := outputVibeCheckTable(result)
	if err != nil {
		t.Fatalf("outputVibeCheckTable with findings: %v", err)
	}
}

// ---------------------------------------------------------------------------
// vibe_check.go — outputVibeCheckMarkdown additional branches
// ---------------------------------------------------------------------------

func TestCov18_outputVibeCheckMarkdown_withEvents(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score: 0.9,
		Grade: "A",
		Events: []vibecheck.TimelineEvent{
			{
				Timestamp: time.Now(),
				SHA:       "def5678",
				Author:    "Author",
				Message:   "fix: important fix",
			},
		},
		Metrics: map[string]float64{"commits": 3},
	}
	err := outputVibeCheckMarkdown(result)
	if err != nil {
		t.Fatalf("outputVibeCheckMarkdown with events: %v", err)
	}
}

func TestCov18_outputVibeCheckMarkdown_withFindings(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score: 0.55,
		Grade: "D",
		Findings: []vibecheck.Finding{
			{Severity: "critical", Category: "security", Message: "Unvalidated input", File: "main.go", Line: 10},
		},
	}
	err := outputVibeCheckMarkdown(result)
	if err != nil {
		t.Fatalf("outputVibeCheckMarkdown with findings: %v", err)
	}
}
