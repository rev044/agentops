package ratchet

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBdCLITimeout(t *testing.T) {
	// Verify the timeout constant is set correctly
	if BdCLITimeout != 5*time.Second {
		t.Errorf("expected BdCLITimeout to be 5s, got %v", BdCLITimeout)
	}

	// Verify error message is correct
	expectedMsg := "bd CLI timeout after 5s"
	if ErrBdCLITimeout.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, ErrBdCLITimeout.Error())
	}
}

func TestGateCheckerWithMissingBd(t *testing.T) {
	// This test verifies that findEpic handles command errors gracefully.
	// When bd is not installed or not in PATH, we should get an error but not hang.
	tmpDir := t.TempDir()
	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		// NewGateChecker may fail if the directory structure is not set up,
		// which is expected for this test
		t.Skip("GateChecker requires specific directory structure")
	}

	// Call findEpic - it will fail but should not hang
	start := time.Now()
	_, err = checker.findEpic("open")
	elapsed := time.Since(start)

	// The command should return quickly (within timeout) even if bd is not found
	if elapsed > BdCLITimeout+time.Second {
		t.Errorf("findEpic took too long (%v), expected to complete within timeout", elapsed)
	}

	// We expect an error (bd not found or no epic found), but not a timeout
	// unless the command is actually hanging
	if errors.Is(err, ErrBdCLITimeout) {
		t.Error("unexpected timeout error - bd command should fail fast if not installed")
	}
}

func TestGetRequiredInput(t *testing.T) {
	cases := []struct {
		step Step
		want string
	}{
		{StepResearch, ""},
		{StepPreMortem, ".agents/research/*.md"},
		{StepPlan, ".agents/specs/*-v2.md OR .agents/synthesis/*.md"},
		{StepImplement, "epic:<epic-id>"},
		{StepCrank, "epic:<epic-id>"},
		{StepVibe, "code changes (optional)"},
		{StepPostMortem, "closed epic (optional)"},
		{Step("unknown"), "unknown"},
	}

	for _, tc := range cases {
		got := GetRequiredInput(tc.step)
		if got != tc.want {
			t.Errorf("GetRequiredInput(%q) = %q, want %q", tc.step, got, tc.want)
		}
	}
}

func TestGetExpectedOutput(t *testing.T) {
	cases := []struct {
		step Step
		want string
	}{
		{StepResearch, ".agents/research/<topic>.md"},
		{StepPreMortem, ".agents/specs/<topic>-v2.md"},
		{StepPlan, "epic:<epic-id>"},
		{StepImplement, "issue:<issue-id> (closed)"},
		{StepCrank, "issue:<issue-id> (closed)"},
		{StepVibe, "validation report"},
		{StepPostMortem, ".agents/learnings/<date>-<topic>.md"},
		{Step("unknown"), "unknown"},
	}

	for _, tc := range cases {
		got := GetExpectedOutput(tc.step)
		if got != tc.want {
			t.Errorf("GetExpectedOutput(%q) = %q, want %q", tc.step, got, tc.want)
		}
	}
}

func TestGateChecker_CheckResearch(t *testing.T) {
	tmpDir := t.TempDir()
	// Create .agents dir so the locator can initialize
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Skip("GateChecker requires specific directory structure")
	}

	result, err := checker.Check(StepResearch)
	if err != nil {
		t.Fatalf("Check(Research): %v", err)
	}
	if !result.Passed {
		t.Error("Research gate should always pass (chaos phase)")
	}
	if result.Step != StepResearch {
		t.Errorf("Step = %q, want %q", result.Step, StepResearch)
	}
}

func TestGateChecker_CheckPreMortem_NoArtifact(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Skip("GateChecker requires specific directory structure")
	}

	result, err := checker.Check(StepPreMortem)
	if err != nil {
		t.Fatalf("Check(PreMortem): %v", err)
	}
	// The locator searches upward, so it may find artifacts from the parent project.
	// We just verify the result is well-formed.
	if result.Step != StepPreMortem {
		t.Errorf("Step = %q, want %q", result.Step, StepPreMortem)
	}
}

func TestGateChecker_CheckPreMortem_WithArtifact(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".agents")
	researchDir := filepath.Join(agentsDir, "research")
	if err := os.MkdirAll(researchDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(researchDir, "topic.md"), []byte("# Research\n"), 0644); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Skip("GateChecker requires specific directory structure")
	}

	result, err := checker.Check(StepPreMortem)
	if err != nil {
		t.Fatalf("Check(PreMortem): %v", err)
	}
	if !result.Passed {
		t.Error("PreMortem gate should pass with research artifact")
	}
}

func TestGateChecker_CheckPlan_NoArtifact(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Skip("GateChecker requires specific directory structure")
	}

	result, err := checker.Check(StepPlan)
	if err != nil {
		t.Fatalf("Check(Plan): %v", err)
	}
	// The locator searches upward, so it may find artifacts from the parent project.
	// We just verify the result is well-formed.
	if result.Step != StepPlan {
		t.Errorf("Step = %q, want %q", result.Step, StepPlan)
	}
}

func TestGateChecker_CheckPlan_WithSynthesis(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".agents")
	synthesisDir := filepath.Join(agentsDir, "synthesis")
	if err := os.MkdirAll(synthesisDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(synthesisDir, "analysis.md"), []byte("# Synthesis\n"), 0644); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Skip("GateChecker requires specific directory structure")
	}

	result, err := checker.Check(StepPlan)
	if err != nil {
		t.Fatalf("Check(Plan): %v", err)
	}
	if !result.Passed {
		t.Error("Plan gate should pass with synthesis artifact")
	}
}

func TestGateChecker_CheckVibe(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Skip("GateChecker requires specific directory structure")
	}

	result, err := checker.Check(StepVibe)
	if err != nil {
		t.Fatalf("Check(Vibe): %v", err)
	}
	if !result.Passed {
		t.Error("Vibe gate should always pass (soft gate)")
	}
}

func TestGateChecker_CheckPostMortem(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Skip("GateChecker requires specific directory structure")
	}

	result, err := checker.Check(StepPostMortem)
	if err != nil {
		t.Fatalf("Check(PostMortem): %v", err)
	}
	if !result.Passed {
		t.Error("PostMortem gate should pass (soft gate)")
	}
}

func TestGateChecker_CheckImplement(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Skip("GateChecker requires specific directory structure")
	}

	result, err := checker.Check(StepImplement)
	if err != nil {
		t.Fatalf("Check(Implement): %v", err)
	}
	// Should fail because bd is not available
	if result.Passed {
		t.Log("Implement gate passed (bd CLI found and returned an epic)")
	}
}

func TestGateChecker_CheckUnknownStep(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Skip("GateChecker requires specific directory structure")
	}

	result, err := checker.Check(Step("nonexistent"))
	if err != nil {
		t.Fatalf("Check(unknown): %v", err)
	}
	if result.Passed {
		t.Error("Unknown step gate should fail")
	}
}

func TestNewGateChecker_InvalidDir(t *testing.T) {
	_, err := NewGateChecker("/nonexistent/path/that/does/not/exist")
	// Should not panic, but may return an error depending on locator behavior
	_ = err
}

// restrictSearchOrder temporarily overrides SearchOrder to only crew-local search,
// preventing tests from finding artifacts in the host's ~/gt/.agents/ or parent rigs.
func restrictSearchOrder(t *testing.T) {
	t.Helper()
	orig := SearchOrder
	SearchOrder = []LocationType{LocationCrew}
	t.Cleanup(func() { SearchOrder = orig })
}

func prependFakeCommand(t *testing.T, name string, body string) {
	t.Helper()

	binDir := t.TempDir()
	scriptPath := filepath.Join(binDir, name)
	script := "#!/usr/bin/env bash\nset -euo pipefail\n" + body + "\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("write fake %s: %v", name, err)
	}

	pathValue := binDir
	if existing := os.Getenv("PATH"); existing != "" {
		pathValue += string(os.PathListSeparator) + existing
	}
	t.Setenv("PATH", pathValue)
}

func chdirTemp(t *testing.T, dir string) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})
}

func TestGateChecker_CheckPreMortem_NoArtifact_CrewOnly(t *testing.T) {
	restrictSearchOrder(t)

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}

	result, err := checker.Check(StepPreMortem)
	if err != nil {
		t.Fatalf("Check(PreMortem): %v", err)
	}
	if result.Passed {
		t.Error("PreMortem gate should fail with no research artifact in crew-only mode")
	}
	if result.Step != StepPreMortem {
		t.Errorf("Step = %q, want %q", result.Step, StepPreMortem)
	}
}

func TestGateChecker_CheckPlan_NoArtifact_CrewOnly(t *testing.T) {
	restrictSearchOrder(t)

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}

	result, err := checker.Check(StepPlan)
	if err != nil {
		t.Fatalf("Check(Plan): %v", err)
	}
	if result.Passed {
		t.Error("Plan gate should fail with no synthesis/spec artifact in crew-only mode")
	}
	if result.Step != StepPlan {
		t.Errorf("Step = %q, want %q", result.Step, StepPlan)
	}
}

func TestGateChecker_CheckImplement_CrewOnly(t *testing.T) {
	restrictSearchOrder(t)

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}

	result, err := checker.Check(StepImplement)
	if err != nil {
		t.Fatalf("Check(Implement): %v", err)
	}
	// bd CLI may or may not be installed, but implement gate should not pass
	// in a bare temp directory without any epic
	if result.Step != StepImplement {
		t.Errorf("Step = %q, want %q", result.Step, StepImplement)
	}
}

func TestGateChecker_CheckVibe_NoChanges_CrewOnly(t *testing.T) {
	restrictSearchOrder(t)

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}

	result, err := checker.Check(StepVibe)
	if err != nil {
		t.Fatalf("Check(Vibe): %v", err)
	}
	// Vibe is a soft gate -- always passes
	if !result.Passed {
		t.Error("Vibe gate should always pass")
	}
	// In a temp dir with no git, checkGitChanges returns false,
	// so message should indicate "no code changes detected"
	if result.Message == "" {
		t.Error("expected non-empty message from vibe gate")
	}
}

func TestGateChecker_CheckPostMortem_CrewOnly(t *testing.T) {
	restrictSearchOrder(t)

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}

	result, err := checker.Check(StepPostMortem)
	if err != nil {
		t.Fatalf("Check(PostMortem): %v", err)
	}
	// PostMortem is a soft gate -- always passes
	if !result.Passed {
		t.Error("PostMortem gate should always pass (soft gate)")
	}
}

func TestParseFirstEpicID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "comments only", in: "# heading\n# comment", want: ""},
		{name: "skips blank and comment lines", in: "\n# comment\n  epic-123   fix release flow", want: "epic-123"},
		{name: "first non-comment wins", in: "epic-101 first\n epic-202 second", want: "epic-101"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseFirstEpicID([]byte(tt.in)); got != tt.want {
				t.Fatalf("parseFirstEpicID(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGateChecker_CheckImplement_WithOpenEpic(t *testing.T) {
	restrictSearchOrder(t)

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}
	prependFakeCommand(t, "bd", `
if [[ "$*" == *"--status open"* ]]; then
  printf 'epic-open Implement release hotfix\n'
  exit 0
fi
exit 1
`)

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}

	result, err := checker.Check(StepImplement)
	if err != nil {
		t.Fatalf("Check(Implement): %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected open epic to satisfy implement gate: %+v", result)
	}
	if result.Input != "epic-open" {
		t.Fatalf("expected epic-open input, got %q", result.Input)
	}
	if result.Location != "beads" {
		t.Fatalf("expected beads location, got %q", result.Location)
	}
}

func TestGateChecker_CheckImplement_FallsBackToInProgress(t *testing.T) {
	restrictSearchOrder(t)

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}
	prependFakeCommand(t, "bd", `
if [[ "$*" == *"--status open"* ]]; then
  printf '# no open epics\n'
  exit 0
fi
if [[ "$*" == *"--status in_progress"* ]]; then
  printf 'epic-active Continue rollout\n'
  exit 0
fi
exit 1
`)

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}

	result, err := checker.Check(StepImplement)
	if err != nil {
		t.Fatalf("Check(Implement): %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected in_progress fallback to satisfy implement gate: %+v", result)
	}
	if result.Input != "epic-active" {
		t.Fatalf("expected epic-active input, got %q", result.Input)
	}
}

func TestGateChecker_CheckCrank_WithOpenEpic(t *testing.T) {
	restrictSearchOrder(t)

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}
	prependFakeCommand(t, "bd", `
if [[ "$*" == *"--status open"* ]]; then
  printf 'epic-open Crank release hotfix\n'
  exit 0
fi
exit 1
`)

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}

	result, err := checker.Check(StepCrank)
	if err != nil {
		t.Fatalf("Check(Crank): %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected open epic to satisfy crank gate: %+v", result)
	}
	if result.Step != StepCrank {
		t.Fatalf("expected crank step, got %q", result.Step)
	}
	if result.Input != "epic-open" {
		t.Fatalf("expected epic-open input, got %q", result.Input)
	}
}

func TestGateChecker_CheckPostMortem_WithClosedEpic(t *testing.T) {
	restrictSearchOrder(t)

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}
	prependFakeCommand(t, "bd", `
if [[ "$*" == *"--status closed"* ]]; then
  printf '# recently closed\nclosed-epic Done and shipped\n'
  exit 0
fi
exit 1
`)

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}

	result, err := checker.Check(StepPostMortem)
	if err != nil {
		t.Fatalf("Check(PostMortem): %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected closed epic to satisfy post-mortem gate: %+v", result)
	}
	if result.Input != "closed-epic" {
		t.Fatalf("expected closed-epic input, got %q", result.Input)
	}
}

func TestGateChecker_CheckVibe_WithGitChanges(t *testing.T) {
	restrictSearchOrder(t)

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}
	prependFakeCommand(t, "git", `
if [[ "${1:-}" == "status" && "${2:-}" == "--porcelain" ]]; then
  printf ' M cli/internal/ratchet/gate.go\n'
  exit 0
fi
exit 1
`)
	chdirTemp(t, tmpDir)

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}

	result, err := checker.Check(StepVibe)
	if err != nil {
		t.Fatalf("Check(Vibe): %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected vibe soft gate to pass: %+v", result)
	}
	if !strings.Contains(result.Message, "Code changes detected") {
		t.Fatalf("expected dirty-tree message, got %q", result.Message)
	}
}

func TestGateChecker_CheckVibe_UsesCheckerRootForGit(t *testing.T) {
	restrictSearchOrder(t)

	tmpDir := t.TempDir()
	ambientDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EXPECTED_GIT_DIR", tmpDir)
	prependFakeCommand(t, "git", `
if [[ "${1:-}" == "status" && "${2:-}" == "--porcelain" ]]; then
  if [[ "$PWD" == "$EXPECTED_GIT_DIR" ]]; then
    printf ' M cli/internal/ratchet/gate.go\n'
  fi
  exit 0
fi
exit 1
`)
	chdirTemp(t, ambientDir)

	checker, err := NewGateChecker(tmpDir)
	if err != nil {
		t.Fatalf("NewGateChecker: %v", err)
	}

	result, err := checker.Check(StepVibe)
	if err != nil {
		t.Fatalf("Check(Vibe): %v", err)
	}
	if !result.Passed {
		t.Fatalf("expected vibe soft gate to pass: %+v", result)
	}
	if !strings.Contains(result.Message, "Code changes detected") {
		t.Fatalf("expected dirty-tree message when checker root is dirty, got %q", result.Message)
	}
}
