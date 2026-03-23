package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleProgramMarkdown = `# PROGRAM.md

## Objective

Build a PROGRAM.md-driven autodev loop.

## Mutable Scope

- cli/cmd/ao/**

## Immutable Scope

- hooks/**

## Experiment Unit

One bounded TDD slice.

## Validation Commands

- ` + "`cd cli && go test ./cmd/ao/...`" + `

## Decision Policy

- Keep changes only when tests pass.

## Escalation Rules

Open a bead for cross-cutting work.

## Stop Conditions

- Tests are green.
`

func TestCobraAutodevInitCreatesProgram(t *testing.T) {
	tmp := chdirTemp(t)
	t.Setenv("HOME", tmp)

	out, err := executeCommand("autodev", "init", "Build a repo-native autodev loop")
	if err != nil {
		t.Fatalf("ao autodev init failed: %v\noutput: %s", err, out)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "PROGRAM.md"))
	if err != nil {
		t.Fatalf("PROGRAM.md not created: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "Build a repo-native autodev loop") {
		t.Fatalf("PROGRAM.md missing objective override:\n%s", text)
	}
	for _, heading := range []string{
		"## Objective",
		"## Mutable Scope",
		"## Immutable Scope",
		"## Experiment Unit",
		"## Validation Commands",
		"## Decision Policy",
		"## Escalation Rules",
		"## Stop Conditions",
	} {
		if !strings.Contains(text, heading) {
			t.Fatalf("PROGRAM.md missing %q:\n%s", heading, text)
		}
	}
}

func TestCobraAutodevValidateJSON(t *testing.T) {
	tmp := chdirTemp(t)
	t.Setenv("HOME", tmp)
	if err := os.WriteFile(filepath.Join(tmp, "PROGRAM.md"), []byte(sampleProgramMarkdown), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := executeCommand("autodev", "validate", "--json")
	if err != nil {
		t.Fatalf("ao autodev validate --json failed: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if valid, _ := result["valid"].(bool); !valid {
		t.Fatalf("validate JSON valid=false: %#v", result)
	}
	if path, _ := result["path"].(string); path != "PROGRAM.md" {
		t.Fatalf("validate JSON path = %q, want PROGRAM.md", path)
	}
}

func TestCobraAutodevShowJSON(t *testing.T) {
	tmp := chdirTemp(t)
	t.Setenv("HOME", tmp)
	if err := os.WriteFile(filepath.Join(tmp, "PROGRAM.md"), []byte(sampleProgramMarkdown), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := executeCommand("autodev", "show", "--json")
	if err != nil {
		t.Fatalf("ao autodev show --json failed: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if objective, _ := result["objective"].(string); !strings.Contains(objective, "PROGRAM.md-driven autodev loop") {
		t.Fatalf("objective = %q", objective)
	}
}
