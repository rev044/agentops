package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDoctor_Integration_HealthyState(t *testing.T) {
	dir := chdirTemp(t)
	setupAgentsDir(t, dir)

	// Create learnings file so knowledge check passes
	writeFile(t, dir+"/.agents/learnings/test-learning.md", "# Test Learning\nSome content here.\n")

	out, err := captureStdout(t, func() error {
		rootCmd.SetArgs([]string{"doctor"})
		return rootCmd.Execute()
	})

	// Doctor may return error if required checks fail (e.g., missing hooks in temp dir)
	// but it should always produce output
	if out == "" {
		t.Fatal("expected doctor output, got empty string")
	}

	// Should contain the header
	if !strings.Contains(out, "ao doctor") {
		t.Errorf("expected output to contain 'ao doctor' header, got:\n%s", out)
	}

	// Should contain the ao CLI check (always passes)
	if !strings.Contains(out, "ao CLI") {
		t.Errorf("expected output to contain 'ao CLI' check, got:\n%s", out)
	}

	// Should contain a summary line with check counts
	hasSummary := strings.Contains(out, "checks passed") || strings.Contains(out, "HEALTHY") || strings.Contains(out, "DEGRADED") || strings.Contains(out, "UNHEALTHY")
	if !hasSummary {
		t.Errorf("expected output to contain a summary (checks passed / HEALTHY / DEGRADED / UNHEALTHY), got:\n%s", out)
	}

	_ = err // doctor may error on missing optional deps; we care about output structure
}

func TestDoctor_Integration_JSONOutput(t *testing.T) {
	dir := chdirTemp(t)
	setupAgentsDir(t, dir)
	writeFile(t, dir+"/.agents/learnings/test-learning.md", "# Learning\nContent.\n")

	out, _ := captureStdout(t, func() error {
		rootCmd.SetArgs([]string{"doctor", "--json"})
		return rootCmd.Execute()
	})

	if out == "" {
		t.Fatal("expected JSON output, got empty string")
	}

	// Parse as JSON to validate structure
	var result doctorOutput
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("expected valid JSON output, got parse error: %v\nraw output:\n%s", err, out)
	}

	// Must have checks array
	if len(result.Checks) == 0 {
		t.Error("expected at least one check in JSON output")
	}

	// Result must be one of the valid statuses
	validResults := map[string]bool{"HEALTHY": true, "DEGRADED": true, "UNHEALTHY": true}
	if !validResults[result.Result] {
		t.Errorf("expected result to be HEALTHY/DEGRADED/UNHEALTHY, got %q", result.Result)
	}

	// First check should be ao CLI
	if result.Checks[0].Name != "ao CLI" {
		t.Errorf("expected first check to be 'ao CLI', got %q", result.Checks[0].Name)
	}
	if result.Checks[0].Status != "pass" {
		t.Errorf("expected ao CLI check to pass, got %q", result.Checks[0].Status)
	}
}

func TestDoctor_Integration_DegradedState(t *testing.T) {
	dir := chdirTemp(t)

	// Create a minimal .agents/ without learnings dir — should trigger warnings
	writeFile(t, dir+"/.agents/ao/sessions/.gitkeep", "")
	// Deliberately skip .agents/learnings/ to trigger degraded state

	out, _ := captureStdout(t, func() error {
		rootCmd.SetArgs([]string{"doctor", "--json"})
		return rootCmd.Execute()
	})

	if out == "" {
		t.Fatal("expected JSON output, got empty string")
	}

	var result doctorOutput
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("JSON parse error: %v\nraw:\n%s", err, out)
	}

	// With missing learnings dir, should have at least one non-pass check
	hasNonPass := false
	for _, c := range result.Checks {
		if c.Status == "warn" || c.Status == "fail" {
			hasNonPass = true
			break
		}
	}
	if !hasNonPass {
		t.Error("expected at least one warn/fail check in degraded state")
	}
}

func TestDoctor_Integration_NoAgentsDir(t *testing.T) {
	// Completely empty directory — no .agents/ at all
	chdirTemp(t)

	out, _ := captureStdout(t, func() error {
		rootCmd.SetArgs([]string{"doctor", "--json"})
		return rootCmd.Execute()
	})

	if out == "" {
		t.Fatal("expected JSON output, got empty string")
	}

	var result doctorOutput
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("JSON parse error: %v\nraw:\n%s", err, out)
	}

	// ao CLI check must always pass regardless of directory state
	if result.Checks[0].Name != "ao CLI" {
		t.Errorf("expected first check 'ao CLI', got %q", result.Checks[0].Name)
	}
	if result.Checks[0].Status != "pass" {
		t.Errorf("expected ao CLI to pass even without .agents/, got %q", result.Checks[0].Status)
	}
}
