package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRPI_Integration_StatusNoRuns(t *testing.T) {
	dir := chdirTemp(t)
	setupAgentsDir(t, dir)
	initMinimalGitRepo(t, dir)

	oldOutput := output
	oldJsonFlag := jsonFlag
	output = "table"
	jsonFlag = false
	t.Cleanup(func() { output = oldOutput; jsonFlag = oldJsonFlag })

	out, err := captureStdout(t, func() error {
		rootCmd.SetArgs([]string{"rpi", "status"})
		return rootCmd.Execute()
	})

	if err != nil {
		t.Fatalf("expected rpi status to succeed, got error: %v", err)
	}

	if !strings.Contains(out, "No active RPI runs") {
		t.Errorf("expected 'No active RPI runs' message, got:\n%s", out)
	}
}

func TestRPI_Integration_StatusWithRunRegistry(t *testing.T) {
	dir := chdirTemp(t)
	setupAgentsDir(t, dir)
	initMinimalGitRepo(t, dir)

	oldOutput := output
	oldJsonFlag := jsonFlag
	output = "table"
	jsonFlag = false
	t.Cleanup(func() { output = oldOutput; jsonFlag = oldJsonFlag })

	// Create a run registry entry as a subdirectory with phased-state.json
	runDir := filepath.Join(dir, ".agents", "rpi", "runs", "rpi-abcd1234ef01")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	stateJSON := `{
		"schema_version": 1,
		"run_id": "rpi-abcd1234ef01",
		"goal": "test integration goal",
		"phase": 3,
		"start_phase": 1,
		"cycle": 1,
		"verdicts": {},
		"attempts": {},
		"started_at": "2026-04-01T10:00:00Z"
	}`
	writeFile(t, filepath.Join(runDir, "phased-state.json"), stateJSON)

	out, err := captureStdout(t, func() error {
		rootCmd.SetArgs([]string{"rpi", "status"})
		return rootCmd.Execute()
	})

	if err != nil {
		t.Fatalf("expected rpi status to succeed, got error: %v\noutput:\n%s", err, out)
	}

	// Should show the run (either in active or historical section)
	if !strings.Contains(out, "rpi-abcd1234ef01") && !strings.Contains(out, "test integration goal") {
		t.Errorf("expected run ID or goal in output, got:\n%s", out)
	}
}

func TestRPI_Integration_StatusJSON(t *testing.T) {
	dir := chdirTemp(t)
	setupAgentsDir(t, dir)
	initMinimalGitRepo(t, dir)

	oldOutput := output
	output = "json"
	t.Cleanup(func() { output = oldOutput })

	out, err := captureStdout(t, func() error {
		rootCmd.SetArgs([]string{"rpi", "status"})
		return rootCmd.Execute()
	})

	if err != nil {
		t.Fatalf("expected rpi status --json to succeed, got error: %v", err)
	}

	var result rpiStatusOutput
	if jsonErr := json.Unmarshal([]byte(out), &result); jsonErr != nil {
		t.Fatalf("expected valid JSON, got parse error: %v\nraw:\n%s", jsonErr, out)
	}

	// Count should be 0 in an empty state
	if result.Count != 0 {
		t.Errorf("expected count=0 with no runs, got %d", result.Count)
	}
}

func TestRPI_Integration_StatusJSONWithRun(t *testing.T) {
	dir := chdirTemp(t)
	setupAgentsDir(t, dir)
	initMinimalGitRepo(t, dir)

	runsDir := filepath.Join(dir, ".agents", "rpi", "runs")
	if err := os.MkdirAll(runsDir, 0755); err != nil {
		t.Fatal(err)
	}
	runDir := filepath.Join(runsDir, "rpi-aabb11223344")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(runDir, "phased-state.json"), `{
		"schema_version": 1,
		"run_id": "rpi-aabb11223344",
		"goal": "JSON test goal",
		"phase": 3,
		"start_phase": 1,
		"cycle": 1,
		"verdicts": {},
		"attempts": {},
		"started_at": "2026-04-01T10:00:00Z"
	}`)

	oldOutput := output
	output = "json"
	t.Cleanup(func() { output = oldOutput })

	out, err := captureStdout(t, func() error {
		rootCmd.SetArgs([]string{"rpi", "status"})
		return rootCmd.Execute()
	})

	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	var result rpiStatusOutput
	if jsonErr := json.Unmarshal([]byte(out), &result); jsonErr != nil {
		t.Fatalf("JSON parse error: %v\nraw:\n%s", jsonErr, out)
	}

	if result.Count != 1 {
		t.Errorf("expected count=1, got %d", result.Count)
	}
}

func TestRPI_Integration_PhasedDryRun(t *testing.T) {
	dir := chdirTemp(t)
	setupAgentsDir(t, dir)
	initMinimalGitRepo(t, dir)

	oldDryRun := dryRun
	dryRun = true
	t.Cleanup(func() { dryRun = oldDryRun })

	out, err := captureStdout(t, func() error {
		rootCmd.SetArgs([]string{"rpi", "phased", "--dry-run", "test dry run goal"})
		return rootCmd.Execute()
	})

	// Dry-run should produce output about what would happen without spawning sessions
	if out == "" && err == nil {
		t.Error("expected some output from dry-run phased command")
	}

	// We don't assert specific output since dry-run behavior varies,
	// but the command should not panic
	_ = err
}
