package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const rpiProgramMarkdown = `# PROGRAM.md

## Objective

Drive bounded autodev experiments through RPI.

## Mutable Scope

- cli/cmd/ao/**
- cli/internal/autodev/**

## Immutable Scope

- hooks/**

## Experiment Unit

One bounded TDD slice that updates tests, implementation, and docs.

## Validation Commands

- ` + "`cd cli && go test ./cmd/ao/... ./internal/autodev/...`" + `
- ` + "`bash scripts/check-worktree-disposition.sh`" + `

## Decision Policy

- Keep changes only when validation commands pass.
- Revert non-improving or out-of-scope work.

## Escalation Rules

Open a bead when required work escapes mutable scope.

## Stop Conditions

- Validation bundle is green.
- No out-of-scope findings remain.
`

func TestInitPhasedState_AttachesProgramPath(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "PROGRAM.md"), []byte(rpiProgramMarkdown), 0o644); err != nil {
		t.Fatal(err)
	}

	state, _, _, err := initPhasedState(tmp, defaultPhasedEngineOptions(), []string{"build the loop"})
	if err != nil {
		t.Fatalf("initPhasedState() error = %v", err)
	}
	if state.ProgramPath != "PROGRAM.md" {
		t.Fatalf("ProgramPath = %q, want PROGRAM.md", state.ProgramPath)
	}
}

func TestInitPhasedState_InvalidProgramFails(t *testing.T) {
	tmp := t.TempDir()
	badProgram := strings.ReplaceAll(rpiProgramMarkdown, "## Mutable Scope\n\n- cli/cmd/ao/**\n- cli/internal/autodev/**\n\n", "## Mutable Scope\n\n")
	if err := os.WriteFile(filepath.Join(tmp, "PROGRAM.md"), []byte(badProgram), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, _, err := initPhasedState(tmp, defaultPhasedEngineOptions(), []string{"build the loop"})
	if err == nil {
		t.Fatal("initPhasedState() error = nil, want invalid PROGRAM.md failure")
	}
	if !strings.Contains(err.Error(), "PROGRAM.md") {
		t.Fatalf("error = %v, want mention of PROGRAM.md", err)
	}
}

func TestBuildPromptForPhase_IncludesProgramContract(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "PROGRAM.md"), []byte(rpiProgramMarkdown), 0o644); err != nil {
		t.Fatal(err)
	}

	state := &phasedState{
		Goal:        "build the loop",
		EpicID:      "ag-123",
		ProgramPath: "PROGRAM.md",
		Opts:        defaultPhasedEngineOptions(),
	}

	prompt, err := buildPromptForPhase(tmp, 1, state, nil)
	if err != nil {
		t.Fatalf("buildPromptForPhase() error = %v", err)
	}
	if !strings.Contains(prompt, "AUTODEV PROGRAM CONTRACT") {
		t.Fatalf("prompt missing program contract:\n%s", prompt)
	}
	if !strings.Contains(prompt, "Read PROGRAM.md before any other repo exploration") {
		t.Fatalf("prompt missing read instruction:\n%s", prompt)
	}
}

func TestWriteExecutionPacketSeed_UsesProgramContract(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "PROGRAM.md"), []byte(rpiProgramMarkdown), 0o644); err != nil {
		t.Fatal(err)
	}

	state := newTestPhasedState().WithGoal("build the loop")
	state.Complexity = ComplexityStandard
	state.ProgramPath = "PROGRAM.md"

	if err := writeExecutionPacketSeed(tmp, state); err != nil {
		t.Fatalf("writeExecutionPacketSeed() error = %v", err)
	}

	packetPath := filepath.Join(tmp, ".agents", "rpi", "execution-packet.json")
	data, err := os.ReadFile(packetPath)
	if err != nil {
		t.Fatalf("read execution packet: %v", err)
	}

	var packet map[string]any
	if err := json.Unmarshal(data, &packet); err != nil {
		t.Fatalf("unmarshal execution packet: %v", err)
	}
	if objective, _ := packet["objective"].(string); objective != "build the loop" {
		t.Fatalf("objective = %q", objective)
	}
	contracts, _ := packet["contract_surfaces"].([]any)
	if len(contracts) == 0 {
		t.Fatalf("contract_surfaces = %#v, want non-empty", packet["contract_surfaces"])
	}
	foundProgram := false
	for _, item := range contracts {
		if item == "PROGRAM.md" {
			foundProgram = true
			break
		}
	}
	if !foundProgram {
		t.Fatalf("contract_surfaces = %#v, want PROGRAM.md", packet["contract_surfaces"])
	}
	doneCriteria, _ := packet["done_criteria"].([]any)
	if len(doneCriteria) < 2 {
		t.Fatalf("done_criteria = %#v, want >= 2", packet["done_criteria"])
	}
}
