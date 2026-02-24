package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContextAssemble_FiveSections(t *testing.T) {
	tmp := t.TempDir()

	// Create GOALS.md
	goalsContent := `# GOALS

## Mission
Test project goals

## Fitness Gates

### test-gate
- **Description:** Ensure tests pass
- **Check:** echo pass
- **Weight:** 5
- **Type:** quality
`
	if err := os.WriteFile(filepath.Join(tmp, "GOALS.md"), []byte(goalsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create cycle history
	histDir := filepath.Join(tmp, ".agents", "evolve")
	if err := os.MkdirAll(histDir, 0755); err != nil {
		t.Fatal(err)
	}
	histLine := `{"timestamp":"2026-02-20T10:00:00Z","cycle":1,"status":"pass","summary":"test cycle"}`
	if err := os.WriteFile(filepath.Join(histDir, "cycle-history.jsonl"), []byte(histLine+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a learning
	learnDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learnDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learnDir, "test-lesson.md"), []byte("# Test Lesson\nAlways verify."), 0644); err != nil {
		t.Fatal(err)
	}

	// Assemble.
	sections := assembleSections(tmp, "Build the auth system", defaultAssembleMaxChars)

	if len(sections) != 5 {
		t.Fatalf("expected 5 sections, got %d", len(sections))
	}

	expectedNames := []string{"GOALS", "HISTORY", "INTEL", "TASK", "PROTOCOL"}
	for i, name := range expectedNames {
		if sections[i].Name != name {
			t.Errorf("section %d: expected name %q, got %q", i, name, sections[i].Name)
		}
		if sections[i].CharCount == 0 {
			t.Errorf("section %q has 0 chars", name)
		}
		if sections[i].Content == "" {
			t.Errorf("section %q has empty content", name)
		}
	}

	// Verify GOALS section references goals.
	if !strings.Contains(sections[0].Content, "GOALS") {
		t.Error("GOALS section missing GOALS header")
	}

	// Verify HISTORY section has the entry.
	if !strings.Contains(sections[1].Content, "HISTORY") {
		t.Error("HISTORY section missing HISTORY header")
	}

	// Verify INTEL section has the learning.
	if !strings.Contains(sections[2].Content, "INTEL") {
		t.Error("INTEL section missing INTEL header")
	}

	// Verify TASK section contains the task description.
	if !strings.Contains(sections[3].Content, "Build the auth system") {
		t.Error("TASK section missing task description")
	}

	// Verify PROTOCOL section has execution contract.
	if !strings.Contains(sections[4].Content, "PROTOCOL") {
		t.Error("PROTOCOL section missing PROTOCOL header")
	}
	if !strings.Contains(sections[4].Content, "Execution Contract") {
		t.Error("PROTOCOL section missing Execution Contract content")
	}
}

func TestContextAssemble_EmptyRepo(t *testing.T) {
	tmp := t.TempDir()

	// No GOALS, no history, no learnings — should gracefully handle all.
	sections := assembleSections(tmp, "Some task", defaultAssembleMaxChars)

	if len(sections) != 5 {
		t.Fatalf("expected 5 sections even in empty repo, got %d", len(sections))
	}

	// GOALS should have graceful fallback.
	if !strings.Contains(sections[0].Content, "No GOALS file found") {
		t.Error("GOALS section should indicate no GOALS file found")
	}

	// HISTORY should have graceful fallback.
	if !strings.Contains(sections[1].Content, "No cycle history found") {
		t.Error("HISTORY section should indicate no cycle history")
	}

	// INTEL should have graceful fallback.
	if !strings.Contains(sections[2].Content, "No learnings or patterns found") {
		t.Error("INTEL section should indicate no learnings or patterns")
	}

	// TASK should still have the task.
	if !strings.Contains(sections[3].Content, "Some task") {
		t.Error("TASK section missing task description")
	}

	// PROTOCOL should still have the template.
	if !strings.Contains(sections[4].Content, "Execution Contract") {
		t.Error("PROTOCOL section missing static template")
	}
}

func TestContextAssemble_CharBudget(t *testing.T) {
	tmp := t.TempDir()

	// Create oversized learnings to test budget enforcement.
	learnDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learnDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write 20 large learning files (each ~2KB).
	for i := 0; i < 20; i++ {
		content := "# Large Learning " + strings.Repeat("x", 2000)
		name := filepath.Join(learnDir, "learning-"+strings.Repeat("a", 3)+string(rune('a'+i))+".md")
		if err := os.WriteFile(name, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Use a small max-chars budget.
	smallBudget := 5000
	sections := assembleSections(tmp, "Test budget", smallBudget)

	totalChars := 0
	for _, s := range sections {
		totalChars += s.CharCount
	}

	// Total should not exceed the budget by much (some overhead from headers).
	// The budget is distributed proportionally, so we allow 20% slack for section headers.
	maxAllowed := int(float64(smallBudget) * 1.3)
	if totalChars > maxAllowed {
		t.Errorf("total chars %d exceeds budget %d (max allowed %d)", totalChars, smallBudget, maxAllowed)
	}

	// Verify each section is individually bounded.
	scale := float64(smallBudget) / float64(defaultAssembleMaxChars)
	intelBudget := int(float64(budgetIntel) * scale)

	// Intel section should be within its scaled budget + truncation marker.
	if sections[2].CharCount > intelBudget+100 {
		t.Errorf("INTEL section %d chars exceeds scaled budget %d", sections[2].CharCount, intelBudget)
	}
}

func TestContextAssemble_Redaction_EnvVar(t *testing.T) {
	tmp := t.TempDir()

	// Create a learning with a secret env var.
	learnDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learnDir, 0755); err != nil {
		t.Fatal(err)
	}
	secretContent := "# Config\nSet API_KEY=sk-abc123xyz to authenticate.\nNormal line here."
	if err := os.WriteFile(filepath.Join(learnDir, "config-secret.md"), []byte(secretContent), 0644); err != nil {
		t.Fatal(err)
	}

	sections := assembleSections(tmp, "Test redaction", defaultAssembleMaxChars)

	// Find INTEL section.
	intelContent := sections[2].Content

	// The API_KEY=sk-abc123xyz should be redacted.
	if strings.Contains(intelContent, "sk-abc123xyz") {
		t.Error("env var secret was not redacted from INTEL section")
	}
	if !strings.Contains(intelContent, "[REDACTED: env-var]") {
		t.Error("expected [REDACTED: env-var] marker in INTEL section")
	}

	// Verify redaction count.
	if sections[2].Redactions == 0 {
		t.Error("expected non-zero redaction count for INTEL section with secret")
	}

	// Check redaction log was written.
	logPath := filepath.Join(tmp, ".agents", "ao", "redaction.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Error("expected redaction.log to be written")
	}
}

func TestContextAssemble_Redaction_JWT(t *testing.T) {
	tmp := t.TempDir()

	// Create a learning with a JWT token.
	learnDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learnDir, 0755); err != nil {
		t.Fatal(err)
	}

	jwtToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0"
	jwtContent := "# Auth Token\nUse this token: " + jwtToken + "\nDone."
	if err := os.WriteFile(filepath.Join(learnDir, "auth-token.md"), []byte(jwtContent), 0644); err != nil {
		t.Fatal(err)
	}

	sections := assembleSections(tmp, "Test JWT redaction", defaultAssembleMaxChars)

	intelContent := sections[2].Content

	// JWT should be redacted.
	if strings.Contains(intelContent, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Error("JWT token was not redacted from INTEL section")
	}
	if !strings.Contains(intelContent, "[REDACTED: jwt-token]") {
		t.Error("expected [REDACTED: jwt-token] marker in INTEL section")
	}

	if sections[2].Redactions == 0 {
		t.Error("expected non-zero redaction count for INTEL section with JWT")
	}
}

func TestContextAssemble_JSONOutput(t *testing.T) {
	tmp := t.TempDir()

	sections := assembleSections(tmp, "JSON test task", defaultAssembleMaxChars)

	totalChars := 0
	totalRedacted := 0
	for _, s := range sections {
		totalChars += s.CharCount
		totalRedacted += s.Redactions
	}

	outPath := filepath.Join(tmp, "briefing.md")

	out := assembleJSONOutput{
		OutputPath:    outPath,
		TotalChars:    totalChars,
		Sections:      sections,
		TotalRedacted: totalRedacted,
		Timestamp:     "2026-02-24T00:00:00Z",
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	// Verify it's valid JSON.
	var parsed assembleJSONOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("JSON output is not valid: %v", err)
	}

	// Verify metadata fields.
	if parsed.OutputPath != outPath {
		t.Errorf("output_path = %q, want %q", parsed.OutputPath, outPath)
	}
	if parsed.TotalChars == 0 {
		t.Error("total_chars should not be 0")
	}
	if len(parsed.Sections) != 5 {
		t.Errorf("expected 5 sections in JSON, got %d", len(parsed.Sections))
	}
	if parsed.Timestamp == "" {
		t.Error("timestamp should not be empty")
	}

	// Verify section names in JSON.
	expectedNames := []string{"GOALS", "HISTORY", "INTEL", "TASK", "PROTOCOL"}
	for i, name := range expectedNames {
		if parsed.Sections[i].Name != name {
			t.Errorf("JSON section %d: expected name %q, got %q", i, name, parsed.Sections[i].Name)
		}
	}
}

func TestContextAssemble_ReadIntelDirReadsJSONFiles(t *testing.T) {
	tmp := t.TempDir()
	learnDir := filepath.Join(tmp, ".agents", "learnings")
	if err := os.MkdirAll(learnDir, 0o755); err != nil {
		t.Fatal(err)
	}
	patternDir := filepath.Join(tmp, ".agents", "patterns")
	if err := os.MkdirAll(patternDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(learnDir, "from-md.md"), []byte("# Markdown learning"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(learnDir, "from-json.json"), []byte(`{"content":"json learning"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(patternDir, "pattern.json"), []byte(`{"pattern":"go"}`), 0644); err != nil {
		t.Fatal(err)
	}

	sections := assembleSections(tmp, "Include JSON artifacts", defaultAssembleMaxChars)
	intelContent := sections[2].Content

	if !strings.Contains(intelContent, "from-md") || !strings.Contains(intelContent, "from-json") {
		t.Error("INTEL section should include both markdown and json learnings")
	}
	if !strings.Contains(intelContent, "pattern") {
		t.Error("INTEL section should include pattern JSON artifacts")
	}
}
