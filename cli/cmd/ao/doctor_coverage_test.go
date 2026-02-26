package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// renderDoctorTable (0%)
// ---------------------------------------------------------------------------

func TestDoctorCov_RenderDoctorTable_Healthy(t *testing.T) {
	output := doctorOutput{
		Checks: []doctorCheck{
			{Name: "ao CLI", Status: "pass", Detail: "v2.0.0", Required: true},
			{Name: "Knowledge", Status: "pass", Detail: "ok", Required: true},
		},
		Result:  "HEALTHY",
		Summary: "2/2 checks passed",
	}
	var buf bytes.Buffer
	renderDoctorTable(&buf, output)
	rendered := buf.String()

	if !strings.Contains(rendered, "ao doctor") {
		t.Error("expected header 'ao doctor'")
	}
	if !strings.Contains(rendered, "ao CLI") {
		t.Error("expected check name 'ao CLI'")
	}
	if !strings.Contains(rendered, "2/2 checks passed") {
		t.Error("expected summary in output")
	}
}

func TestDoctorCov_RenderDoctorTable_WithWarnings(t *testing.T) {
	output := doctorOutput{
		Checks: []doctorCheck{
			{Name: "Pass Check", Status: "pass", Detail: "ok", Required: true},
			{Name: "Warn Check", Status: "warn", Detail: "degraded", Required: false},
			{Name: "Fail Check", Status: "fail", Detail: "broken", Required: true},
			{Name: "Unknown", Status: "unknown", Detail: "???", Required: false},
		},
		Result:  "UNHEALTHY",
		Summary: "1/4 checks passed, 1 warning, 1 failed",
	}
	var buf bytes.Buffer
	renderDoctorTable(&buf, output)
	rendered := buf.String()

	// All status icons should appear
	if !strings.Contains(rendered, "\u2713") {
		t.Error("expected pass icon")
	}
	if !strings.Contains(rendered, "!") {
		t.Error("expected warn icon")
	}
	if !strings.Contains(rendered, "\u2717") {
		t.Error("expected fail icon")
	}
	if !strings.Contains(rendered, "?") {
		t.Error("expected unknown icon")
	}
}

func TestDoctorCov_RenderDoctorTable_Empty(t *testing.T) {
	output := doctorOutput{
		Checks:  []doctorCheck{},
		Result:  "HEALTHY",
		Summary: "0/0 checks passed",
	}
	var buf bytes.Buffer
	renderDoctorTable(&buf, output)
	if buf.Len() == 0 {
		t.Error("expected non-empty output even with no checks")
	}
}

// ---------------------------------------------------------------------------
// doctorStatusIcon (already 100%, but test unknown branch explicitly)
// ---------------------------------------------------------------------------

func TestDoctorCov_DoctorStatusIcon_Unknown(t *testing.T) {
	got := doctorStatusIcon("bogus")
	if got != "?" {
		t.Errorf("doctorStatusIcon(bogus) = %q, want ?", got)
	}
}

// ---------------------------------------------------------------------------
// checkOptionalCLI (0%)
// ---------------------------------------------------------------------------

func TestDoctorCov_CheckOptionalCLI_Available(t *testing.T) {
	// "go" should be in PATH during tests
	result := checkOptionalCLI("go", "test builds")
	if result.Status != "pass" {
		t.Errorf("status=%q, want pass for 'go' CLI (detail: %s)", result.Status, result.Detail)
	}
	if result.Required {
		t.Error("optional CLI should not be required")
	}
	if !strings.Contains(result.Detail, "available") {
		t.Errorf("expected 'available' in detail, got %q", result.Detail)
	}
}

func TestDoctorCov_CheckOptionalCLI_NotAvailable(t *testing.T) {
	result := checkOptionalCLI("nonexistent_cli_xyz_999", "test feature")
	if result.Status != "warn" {
		t.Errorf("status=%q, want warn for missing CLI (detail: %s)", result.Status, result.Detail)
	}
	if !strings.Contains(result.Detail, "not found") {
		t.Errorf("expected 'not found' in detail, got %q", result.Detail)
	}
	if !strings.Contains(result.Detail, "test feature") {
		t.Errorf("expected reason in detail, got %q", result.Detail)
	}
}

// ---------------------------------------------------------------------------
// checkCLIDependencies (0%)
// ---------------------------------------------------------------------------

func TestDoctorCov_CheckCLIDependencies(t *testing.T) {
	// We can't guarantee gt/bd availability, but we can exercise the function.
	result := checkCLIDependencies()
	// It should return either pass or warn, never fail
	if result.Status != "pass" && result.Status != "warn" {
		t.Errorf("status=%q, expected pass or warn", result.Status)
	}
	if result.Name != "CLI Dependencies" {
		t.Errorf("name=%q, want 'CLI Dependencies'", result.Name)
	}
	if result.Required {
		t.Error("CLI Dependencies should not be required")
	}
}

// ---------------------------------------------------------------------------
// checkHookCoverage (0%) — needs a fake HOME with settings.json
// ---------------------------------------------------------------------------

func TestDoctorCov_CheckHookCoverage_NoHooksFile(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	result := checkHookCoverage()
	if result.Status != "warn" {
		t.Errorf("status=%q, want warn when no hooks files exist (detail: %s)", result.Status, result.Detail)
	}
}

func TestDoctorCov_CheckHookCoverage_SettingsWithHooks(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Build a settings.json with hooks containing all events
	hooks := make(map[string]any)
	for _, event := range AllEventNames() {
		hooks[event] = []any{
			map[string]any{
				"matcher": "",
				"hooks": []any{
					map[string]any{"type": "command", "command": "ao hook-dispatch " + event},
				},
			},
		}
	}
	settings := map[string]any{"hooks": hooks}
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	result := checkHookCoverage()
	// Should find hooks and report coverage
	if result.Status == "fail" {
		t.Errorf("status=fail unexpected (detail: %s)", result.Detail)
	}
}

func TestDoctorCov_CheckHookCoverage_FallbackHooksJSON(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// hooks.json with a single event (partial coverage)
	hooks := map[string]any{
		"SessionStart": []any{
			map[string]any{
				"matcher": "",
				"hooks": []any{
					map[string]any{"type": "command", "command": "ao hook-dispatch SessionStart"},
				},
			},
		},
	}
	data, err := json.Marshal(hooks)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "hooks.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	result := checkHookCoverage()
	// Should detect partial coverage
	if result.Status == "fail" {
		t.Errorf("status=fail unexpected for partial hooks.json (detail: %s)", result.Detail)
	}
}

func TestDoctorCov_UsesRuntimeManifestContract(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	settings := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"type": "command", "command": "ao inject --apply-decay"},
					},
				},
			},
		},
	}
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	hooksDir := filepath.Join(tmp, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	manifest := `{
		"hooks": {
			"SessionStart": [{"hooks": [{"type":"command","command":"ao inject --apply-decay"}]}]
		}
	}`
	if err := os.WriteFile(filepath.Join(hooksDir, "hooks.json"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	result := checkHookCoverage()
	if result.Status != "pass" {
		t.Fatalf("expected pass with 1/1 active manifest event, got %q (%s)", result.Status, result.Detail)
	}
	if !strings.Contains(result.Detail, "1/1") {
		t.Fatalf("expected active contract denominator in detail, got %q", result.Detail)
	}
}

func TestDoctorCov_FallbackReasonSurfaced(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	claudeDir := filepath.Join(fakeHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	settings := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"type": "command", "command": "ao inject --apply-decay"},
					},
				},
			},
		},
	}
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	hooksDir := filepath.Join(tmp, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "hooks.json"), []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	result := checkHookCoverage()
	if !strings.Contains(result.Detail, "coverage contract fallback:") {
		t.Fatalf("expected fallback reason in detail, got %q", result.Detail)
	}
	if !strings.Contains(result.Detail, "parse hooks manifest") {
		t.Fatalf("expected parse failure reason in detail, got %q", result.Detail)
	}
}

// ---------------------------------------------------------------------------
// checkSkills (0%) — needs a fake HOME
// ---------------------------------------------------------------------------

func TestDoctorCov_CheckSkills_NoSkills(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	result := checkSkills()
	if result.Status != "warn" {
		t.Errorf("status=%q, want warn when no skills installed (detail: %s)", result.Status, result.Detail)
	}
}

func TestDoctorCov_CheckSkills_WithSkills(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	// Create a fake skill directory with SKILL.md
	skillDir := filepath.Join(fakeHome, ".claude", "skills", "fake-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Fake Skill"), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkSkills()
	if result.Status != "pass" {
		t.Errorf("status=%q, want pass when skills are found (detail: %s)", result.Status, result.Detail)
	}
	if !strings.Contains(result.Detail, "1 skills found") {
		t.Errorf("expected '1 skills found' in detail, got %q", result.Detail)
	}
}

func TestDoctorCov_CheckSkills_AltPath(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	// Create skills under .agents/skills/ (alt path)
	skillDir := filepath.Join(fakeHome, ".agents", "skills", "alt-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Alt Skill"), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkSkills()
	if result.Status != "pass" {
		t.Errorf("status=%q, want pass for alt-path skills (detail: %s)", result.Status, result.Detail)
	}
}

// ---------------------------------------------------------------------------
// runDoctor (0%) — exercise via cobra command
// ---------------------------------------------------------------------------

func TestDoctorCov_RunDoctor_TableOutput(t *testing.T) {
	tmp := chdirTemp(t)
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	// Set up knowledge base so at least one check passes
	if err := os.MkdirAll(filepath.Join(tmp, ".agents", "ao"), 0755); err != nil {
		t.Fatal(err)
	}

	// Save and restore the global doctorJSON flag
	origJSON := doctorJSON
	doctorJSON = false
	t.Cleanup(func() { doctorJSON = origJSON })

	var buf bytes.Buffer
	doctorCmd.SetOut(&buf)
	doctorCmd.SetErr(&buf)

	// runDoctor may return error (required check failure) — we just ensure it runs
	_ = runDoctor(doctorCmd, nil)

	rendered := buf.String()
	if !strings.Contains(rendered, "ao doctor") {
		t.Error("expected 'ao doctor' header in table output")
	}
}

func TestDoctorCov_RunDoctor_JSONOutput(t *testing.T) {
	tmp := chdirTemp(t)
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	if err := os.MkdirAll(filepath.Join(tmp, ".agents", "ao"), 0755); err != nil {
		t.Fatal(err)
	}

	origJSON := doctorJSON
	doctorJSON = true
	t.Cleanup(func() { doctorJSON = origJSON })

	var buf bytes.Buffer
	doctorCmd.SetOut(&buf)
	doctorCmd.SetErr(&buf)

	_ = runDoctor(doctorCmd, nil)

	// Should be valid JSON
	var result doctorOutput
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("expected valid JSON output, got error: %v\nOutput: %s", err, buf.String())
	}
	if len(result.Checks) == 0 {
		t.Error("expected at least one check in JSON output")
	}
	if result.Result == "" {
		t.Error("expected non-empty result field")
	}
}

// ---------------------------------------------------------------------------
// newestFileModTime — branch for no regular files
// ---------------------------------------------------------------------------

func TestDoctorCov_NewestFileModTime_OnlyDirs(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	newest := newestFileModTime(entries)
	if !newest.IsZero() {
		t.Error("expected zero time for directory-only entries")
	}
}

// ---------------------------------------------------------------------------
// countEstablished
// ---------------------------------------------------------------------------

func TestDoctorCov_CountEstablished(t *testing.T) {
	tmp := t.TempDir()
	// Create some files with various names
	for _, name := range []string{
		"learning-established-001.md",
		"learning-promoted-002.md",
		"learning-provisional-003.md",
		"unrelated.jsonl",
	} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	got := countEstablished(tmp)
	if got != 2 {
		t.Errorf("countEstablished() = %d, want 2", got)
	}
}

func TestDoctorCov_CountEstablished_NonexistentDir(t *testing.T) {
	got := countEstablished("/nonexistent/path/xyz")
	if got != 0 {
		t.Errorf("countEstablished(nonexistent) = %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// findHealScript — exercise all branches
// ---------------------------------------------------------------------------

func TestDoctorCov_FindHealScript_NotFound(t *testing.T) {
	// Use a temp dir so in-repo heal.sh isn't found
	chdirTemp(t)
	t.Setenv("HOME", t.TempDir())

	path := findHealScript()
	if path != "" {
		t.Errorf("expected empty path, got %q", path)
	}
}

func TestDoctorCov_FindHealScript_InClaude(t *testing.T) {
	chdirTemp(t)
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	// Create heal.sh in ~/.claude/skills/heal-skill/scripts/
	healDir := filepath.Join(fakeHome, ".claude", "skills", "heal-skill", "scripts")
	if err := os.MkdirAll(healDir, 0755); err != nil {
		t.Fatal(err)
	}
	healPath := filepath.Join(healDir, "heal.sh")
	if err := os.WriteFile(healPath, []byte("#!/bin/bash\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	got := findHealScript()
	if got != healPath {
		t.Errorf("findHealScript() = %q, want %q", got, healPath)
	}
}

func TestDoctorCov_FindHealScript_InAgents(t *testing.T) {
	chdirTemp(t)
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	// Create heal.sh in ~/.agents/skills/heal-skill/scripts/
	healDir := filepath.Join(fakeHome, ".agents", "skills", "heal-skill", "scripts")
	if err := os.MkdirAll(healDir, 0755); err != nil {
		t.Fatal(err)
	}
	healPath := filepath.Join(healDir, "heal.sh")
	if err := os.WriteFile(healPath, []byte("#!/bin/bash\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	got := findHealScript()
	if got != healPath {
		t.Errorf("findHealScript() = %q, want %q", got, healPath)
	}
}

// ---------------------------------------------------------------------------
// gatherDoctorChecks (0%) — smoke test that it returns checks
// ---------------------------------------------------------------------------

func TestDoctorCov_GatherDoctorChecks(t *testing.T) {
	tmp := chdirTemp(t)
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	// Set up minimal environment
	if err := os.MkdirAll(filepath.Join(tmp, ".agents", "ao"), 0755); err != nil {
		t.Fatal(err)
	}

	checks := gatherDoctorChecks()
	if len(checks) == 0 {
		t.Error("expected at least one check from gatherDoctorChecks()")
	}

	// First check should be ao CLI
	if checks[0].Name != "ao CLI" {
		t.Errorf("first check name = %q, want 'ao CLI'", checks[0].Name)
	}
	if checks[0].Status != "pass" {
		t.Errorf("ao CLI check status = %q, want 'pass'", checks[0].Status)
	}
}

// ---------------------------------------------------------------------------
// extractHooksMap — test with invalid JSON
// ---------------------------------------------------------------------------

func TestDoctorCov_ExtractHooksMap_InvalidJSON(t *testing.T) {
	_, ok := extractHooksMap([]byte("not valid json"))
	if ok {
		t.Error("expected false for invalid JSON")
	}
}

func TestDoctorCov_ExtractHooksMap_NoHooksField(t *testing.T) {
	data := []byte(`{"foo": "bar"}`)
	_, ok := extractHooksMap(data)
	if ok {
		t.Error("expected false when no hooks field and no events found")
	}
}

// ---------------------------------------------------------------------------
// countHooksInMap — test with []any (array branch)
// ---------------------------------------------------------------------------

func TestDoctorCov_CountHooksInMap_Array(t *testing.T) {
	got := countHooksInMap([]any{"a", "b", "c"})
	if got != 3 {
		t.Errorf("countHooksInMap([]any) = %d, want 3", got)
	}
}

// ---------------------------------------------------------------------------
// checkStaleReferences — stale command reference detector
// ---------------------------------------------------------------------------

func TestDoctorCov_CheckStaleReferences_NoFiles(t *testing.T) {
	chdirTemp(t)
	result := checkStaleReferences()
	if result.Status != "pass" {
		t.Errorf("status=%q, want pass when no hooks/skills exist (detail: %s)", result.Status, result.Detail)
	}
	if result.Required {
		t.Error("Stale References should not be required")
	}
}

func TestDoctorCov_CheckStaleReferences_CleanFiles(t *testing.T) {
	tmp := chdirTemp(t)

	// Create a hooks file with only new-style commands
	hooksDir := filepath.Join(tmp, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "dispatch.sh"), []byte("#!/bin/bash\nao know forge transcript\nao work rpi start\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a skill file with only new-style commands
	skillDir := filepath.Join(tmp, "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Test Skill\nRun `ao know inject` to load context.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkStaleReferences()
	if result.Status != "pass" {
		t.Errorf("status=%q, want pass for clean files (detail: %s)", result.Status, result.Detail)
	}
}

func TestDoctorCov_CheckStaleReferences_StaleInHooks(t *testing.T) {
	tmp := chdirTemp(t)

	hooksDir := filepath.Join(tmp, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Use old-style "ao forge" instead of "ao know forge"
	if err := os.WriteFile(filepath.Join(hooksDir, "dispatch.sh"), []byte("#!/bin/bash\nao forge transcript\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkStaleReferences()
	if result.Status != "warn" {
		t.Errorf("status=%q, want warn for stale hooks reference (detail: %s)", result.Status, result.Detail)
	}
	if !strings.Contains(result.Detail, "stale reference") {
		t.Errorf("expected 'stale reference' in detail, got %q", result.Detail)
	}
}

func TestDoctorCov_CheckStaleReferences_StaleInSkills(t *testing.T) {
	tmp := chdirTemp(t)

	skillDir := filepath.Join(tmp, "skills", "test-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Use old-style "ao inject" instead of "ao know inject"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Test\nRun `ao inject` to load.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkStaleReferences()
	if result.Status != "warn" {
		t.Errorf("status=%q, want warn for stale skill reference (detail: %s)", result.Status, result.Detail)
	}
}

func TestDoctorCov_CheckStaleReferences_NoFalsePositiveOnNamespace(t *testing.T) {
	tmp := chdirTemp(t)

	// "ao know forge" contains "ao forge" as a substring — should NOT trigger
	hooksDir := filepath.Join(tmp, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "test.sh"), []byte("ao know forge transcript\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result := checkStaleReferences()
	if result.Status != "pass" {
		t.Errorf("status=%q, want pass (should not false-positive on 'ao know forge') (detail: %s)", result.Status, result.Detail)
	}
}

func TestDoctorCov_ScanFileForDeprecatedCommands(t *testing.T) {
	tmp := t.TempDir()

	t.Run("nonexistent file returns nil", func(t *testing.T) {
		refs := scanFileForDeprecatedCommands(filepath.Join(tmp, "nope.sh"))
		if len(refs) != 0 {
			t.Errorf("expected 0 refs for nonexistent file, got %d", len(refs))
		}
	})

	t.Run("file with multiple deprecated commands", func(t *testing.T) {
		f := filepath.Join(tmp, "multi.sh")
		content := "ao forge transcript\nao inject --apply-decay\nao know search query\n"
		if err := os.WriteFile(f, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		refs := scanFileForDeprecatedCommands(f)
		if len(refs) < 2 {
			t.Errorf("expected at least 2 stale refs, got %d", len(refs))
		}
	})
}

func TestDoctorCov_CountUniqueFiles(t *testing.T) {
	refs := []staleReference{
		{File: "a.sh", OldCommand: "ao forge", NewCommand: "ao know forge"},
		{File: "a.sh", OldCommand: "ao inject", NewCommand: "ao know inject"},
		{File: "b.md", OldCommand: "ao forge", NewCommand: "ao know forge"},
	}
	got := countUniqueFiles(refs)
	if got != 2 {
		t.Errorf("countUniqueFiles() = %d, want 2", got)
	}
}
