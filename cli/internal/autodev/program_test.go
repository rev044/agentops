package autodev

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validProgramMarkdown = `# PROGRAM.md

## Objective

Ship a repo-native autonomous development loop with bounded experiments.

## Mutable Scope

- cli/cmd/ao/**
- cli/internal/autodev/**

## Immutable Scope

- hooks/**
- docs/contracts/**

## Experiment Unit

One bounded change set that updates tests first, then implementation, then validation.

## Validation Commands

- ` + "`cd cli && go test ./cmd/ao/... ./internal/autodev/...`" + `
- ` + "`bash scripts/check-worktree-disposition.sh`" + `

## Decision Policy

- Keep changes only when the validation bundle is green.
- Prefer simpler diffs when outcomes tie.
- Revert non-improving or out-of-scope experiments.

## Escalation Rules

Open a bead when required work escapes mutable scope or requires a cross-cutting rewrite.

## Stop Conditions

- Validation bundle is green.
- No unhandled out-of-scope findings remain.
`

func TestParseMarkdownProgram(t *testing.T) {
	prog, err := ParseMarkdownProgram([]byte(validProgramMarkdown))
	if err != nil {
		t.Fatalf("ParseMarkdownProgram() error = %v", err)
	}

	if !strings.Contains(prog.Objective, "repo-native autonomous development loop") {
		t.Fatalf("Objective = %q", prog.Objective)
	}
	if len(prog.MutableScope) != 2 || prog.MutableScope[0] != "cli/cmd/ao/**" {
		t.Fatalf("MutableScope = %#v", prog.MutableScope)
	}
	if len(prog.ImmutableScope) != 2 || prog.ImmutableScope[0] != "hooks/**" {
		t.Fatalf("ImmutableScope = %#v", prog.ImmutableScope)
	}
	if len(prog.ValidationCommands) != 2 || prog.ValidationCommands[0] != "cd cli && go test ./cmd/ao/... ./internal/autodev/..." {
		t.Fatalf("ValidationCommands = %#v", prog.ValidationCommands)
	}
	if len(prog.DecisionPolicy) != 3 {
		t.Fatalf("DecisionPolicy len = %d, want 3", len(prog.DecisionPolicy))
	}
	if len(prog.StopConditions) != 2 {
		t.Fatalf("StopConditions len = %d, want 2", len(prog.StopConditions))
	}
	if !strings.Contains(prog.ExperimentUnit, "tests first") {
		t.Fatalf("ExperimentUnit = %q", prog.ExperimentUnit)
	}
	if !strings.Contains(prog.EscalationRules, "Open a bead") {
		t.Fatalf("EscalationRules = %q", prog.EscalationRules)
	}
}

func TestValidateProgram(t *testing.T) {
	prog, err := ParseMarkdownProgram([]byte(validProgramMarkdown))
	if err != nil {
		t.Fatalf("ParseMarkdownProgram() error = %v", err)
	}
	if errs := ValidateProgram(prog); len(errs) != 0 {
		t.Fatalf("ValidateProgram(valid) = %v, want no errors", errs)
	}

	invalid := strings.ReplaceAll(validProgramMarkdown, "## Stop Conditions\n\n- Validation bundle is green.\n- No unhandled out-of-scope findings remain.\n", "## Stop Conditions\n")
	prog, err = ParseMarkdownProgram([]byte(invalid))
	if err != nil {
		t.Fatalf("ParseMarkdownProgram(invalid) error = %v", err)
	}
	errs := ValidateProgram(prog)
	if len(errs) == 0 {
		t.Fatal("ValidateProgram(invalid) = no errors, want at least one")
	}
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "stop conditions") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ValidateProgram(invalid) errors = %v, want stop conditions error", errs)
	}
}

func TestResolveAndLoadProgramPath(t *testing.T) {
	tmp := t.TempDir()

	if got := ResolveProgramPath(tmp); got != "" {
		t.Fatalf("ResolveProgramPath(no files) = %q, want empty", got)
	}

	if err := os.WriteFile(filepath.Join(tmp, "AUTODEV.md"), []byte(validProgramMarkdown), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := ResolveProgramPath(tmp); got != "AUTODEV.md" {
		t.Fatalf("ResolveProgramPath(AUTODEV.md only) = %q, want AUTODEV.md", got)
	}

	if err := os.WriteFile(filepath.Join(tmp, "PROGRAM.md"), []byte(validProgramMarkdown), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := ResolveProgramPath(tmp); got != "PROGRAM.md" {
		t.Fatalf("ResolveProgramPath(PROGRAM.md preferred) = %q, want PROGRAM.md", got)
	}

	prog, path, err := LoadProgram(filepath.Join(tmp, "PROGRAM.md"))
	if err != nil {
		t.Fatalf("LoadProgram(PROGRAM.md) error = %v", err)
	}
	if path != filepath.Join(tmp, "PROGRAM.md") {
		t.Fatalf("LoadProgram path = %q", path)
	}
	if !strings.Contains(prog.Objective, "bounded experiments") {
		t.Fatalf("LoadProgram objective = %q", prog.Objective)
	}
}
