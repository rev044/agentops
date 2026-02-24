package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCurate_Catalog_Learning catalogs a learning artifact with correct
// structure and verifies it gets written to .agents/learnings/.
func TestCurate_Catalog_Learning(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// Create a test artifact file with YAML frontmatter
	artifactContent := `---
type: learning
date: 2026-02-24
---
Agents forget context between sessions. Use .agents/ for persistence.
`
	inputPath := filepath.Join(tmpDir, "test-learning.md")
	if err := os.WriteFile(inputPath, []byte(artifactContent), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	// Run catalog
	if err := runCurateCatalog(nil, []string{inputPath}); err != nil {
		t.Fatalf("runCurateCatalog: %v", err)
	}

	// Verify artifact was written to .agents/learnings/
	learningsDir := filepath.Join(tmpDir, ".agents", "learnings")
	entries, err := os.ReadDir(learningsDir)
	if err != nil {
		t.Fatalf("reading learnings dir: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 artifact in learnings dir, got %d", len(entries))
	}

	// Read and validate the artifact
	data, err := os.ReadFile(filepath.Join(learningsDir, entries[0].Name()))
	if err != nil {
		t.Fatalf("reading artifact: %v", err)
	}

	var artifact curateArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fatalf("unmarshal artifact: %v", err)
	}

	if artifact.Type != "learning" {
		t.Errorf("expected type 'learning', got %q", artifact.Type)
	}
	if artifact.Date != "2026-02-24" {
		t.Errorf("expected date '2026-02-24', got %q", artifact.Date)
	}
	if artifact.SchemaVersion != 1 {
		t.Errorf("expected schema_version 1, got %d", artifact.SchemaVersion)
	}
	if artifact.CuratedAt == "" {
		t.Error("expected curated_at to be set")
	}
	if artifact.ID == "" {
		t.Error("expected auto-generated ID")
	}
	if artifact.Content == "" {
		t.Error("expected non-empty content")
	}
}

// TestCurate_Catalog_InvalidType rejects artifacts with unknown type.
func TestCurate_Catalog_InvalidType(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// Create artifact with invalid type
	artifactContent := `---
type: opinion
date: 2026-02-24
---
This is not a valid type.
`
	inputPath := filepath.Join(tmpDir, "bad-type.md")
	if err := os.WriteFile(inputPath, []byte(artifactContent), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	err = runCurateCatalog(nil, []string{inputPath})
	if err == nil {
		t.Fatal("expected error for unknown artifact type, got nil")
	}

	expectedMsg := `unknown artifact type "opinion": must be one of learning, decision, failure, pattern`
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}

	// Verify nothing was written
	learningsDir := filepath.Join(tmpDir, ".agents", "learnings")
	if _, statErr := os.Stat(learningsDir); !os.IsNotExist(statErr) {
		t.Error("learnings dir should not exist after rejected artifact")
	}
}

// TestCurate_Verify_AllPass returns verified=true when all gates pass.
// This test mocks the gate checks by creating a baseline with all passing gates
// and verifying that no regressions are detected.
func TestCurate_Verify_AllPass(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// Create a minimal GOALS.yaml with a trivially passing check
	goalsContent := `version: 3
goals:
  - id: always-pass
    description: A gate that always passes
    check: "true"
    weight: 1
`
	if err := os.WriteFile(filepath.Join(tmpDir, "GOALS.yaml"), []byte(goalsContent), 0o644); err != nil {
		t.Fatalf("write GOALS.yaml: %v", err)
	}

	// Create a baseline snapshot where everything passes
	baselineDir := filepath.Join(tmpDir, ".agents", "ao", "baselines")
	if err := os.MkdirAll(baselineDir, 0o755); err != nil {
		t.Fatalf("mkdir baselines: %v", err)
	}

	baselineSnap := map[string]interface{}{
		"timestamp": "2026-02-23T10:00:00Z",
		"git_sha":   "abc123",
		"goals": []map[string]interface{}{
			{"goal_id": "always-pass", "result": "pass", "duration_s": 0.1, "weight": 1},
		},
		"summary": map[string]interface{}{
			"total": 1, "passing": 1, "failing": 0, "skipped": 0, "score": 100.0,
		},
	}
	baselineData, _ := json.MarshalIndent(baselineSnap, "", "  ")
	if err := os.WriteFile(filepath.Join(baselineDir, "2026-02-23T10-00-00.000.json"), baselineData, 0o600); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	// Force JSON output so we can parse the result
	origOutput := output
	output = "json"
	t.Cleanup(func() { output = origOutput })

	// Capture stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCurateVerify(nil, nil)

	w.Close()
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("runCurateVerify: %v", err)
	}

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	r.Close()

	var result curateVerifyResult
	if err := json.Unmarshal(buf[:n], &result); err != nil {
		t.Fatalf("unmarshal result: %v (raw: %s)", err, string(buf[:n]))
	}

	if !result.Verified {
		t.Error("expected verified=true")
	}
	if result.GatesPassed != 1 {
		t.Errorf("expected 1 gate passed, got %d", result.GatesPassed)
	}
	if result.GatesFailed != 0 {
		t.Errorf("expected 0 gates failed, got %d", result.GatesFailed)
	}
	if len(result.Regressions) != 0 {
		t.Errorf("expected 0 regressions, got %v", result.Regressions)
	}
}

// TestCurate_Verify_Regression detects regression when a gate that previously
// passed now fails.
func TestCurate_Verify_Regression(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// Create a GOALS.yaml with a gate that fails
	goalsContent := `version: 3
goals:
  - id: now-fails
    description: A gate that now fails
    check: "false"
    weight: 1
`
	if err := os.WriteFile(filepath.Join(tmpDir, "GOALS.yaml"), []byte(goalsContent), 0o644); err != nil {
		t.Fatalf("write GOALS.yaml: %v", err)
	}

	// Create a baseline where it previously passed
	baselineDir := filepath.Join(tmpDir, ".agents", "ao", "baselines")
	if err := os.MkdirAll(baselineDir, 0o755); err != nil {
		t.Fatalf("mkdir baselines: %v", err)
	}

	baselineSnap := map[string]interface{}{
		"timestamp": "2026-02-23T10:00:00Z",
		"git_sha":   "abc123",
		"goals": []map[string]interface{}{
			{"goal_id": "now-fails", "result": "pass", "duration_s": 0.1, "weight": 1},
		},
		"summary": map[string]interface{}{
			"total": 1, "passing": 1, "failing": 0, "skipped": 0, "score": 100.0,
		},
	}
	baselineData, _ := json.MarshalIndent(baselineSnap, "", "  ")
	if err := os.WriteFile(filepath.Join(baselineDir, "2026-02-23T10-00-00.000.json"), baselineData, 0o600); err != nil {
		t.Fatalf("write baseline: %v", err)
	}

	// Force JSON output
	origOutput := output
	output = "json"
	t.Cleanup(func() { output = origOutput })

	// Capture stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCurateVerify(nil, nil)

	w.Close()
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("runCurateVerify: %v", err)
	}

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	r.Close()

	var result curateVerifyResult
	if err := json.Unmarshal(buf[:n], &result); err != nil {
		t.Fatalf("unmarshal result: %v (raw: %s)", err, string(buf[:n]))
	}

	if result.Verified {
		t.Error("expected verified=false due to regression")
	}
	if result.GatesFailed != 1 {
		t.Errorf("expected 1 gate failed, got %d", result.GatesFailed)
	}
	if len(result.Regressions) != 1 {
		t.Fatalf("expected 1 regression, got %d", len(result.Regressions))
	}
	if result.Regressions[0] != "now-fails" {
		t.Errorf("expected regression on 'now-fails', got %q", result.Regressions[0])
	}
}

// TestCurate_Status_EmptyRepo returns zero counts gracefully when no artifacts exist.
func TestCurate_Status_EmptyRepo(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	// Force JSON output
	origOutput := output
	output = "json"
	t.Cleanup(func() { output = origOutput })

	// Capture stdout
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runCurateStatus(nil, nil)

	w.Close()
	os.Stdout = origStdout

	if err != nil {
		t.Fatalf("runCurateStatus: %v", err)
	}

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	r.Close()

	var result curateStatusResult
	if err := json.Unmarshal(buf[:n], &result); err != nil {
		t.Fatalf("unmarshal result: %v (raw: %s)", err, string(buf[:n]))
	}

	if result.Total != 0 {
		t.Errorf("expected total=0, got %d", result.Total)
	}
	if result.Learnings != 0 {
		t.Errorf("expected learnings=0, got %d", result.Learnings)
	}
	if result.Decisions != 0 {
		t.Errorf("expected decisions=0, got %d", result.Decisions)
	}
	if result.Failures != 0 {
		t.Errorf("expected failures=0, got %d", result.Failures)
	}
	if result.Patterns != 0 {
		t.Errorf("expected patterns=0, got %d", result.Patterns)
	}
	if result.LastCatalogAt != "" {
		t.Errorf("expected empty last_catalog_at, got %q", result.LastCatalogAt)
	}
	if result.LastVerifyAt != "" {
		t.Errorf("expected empty last_verify_at, got %q", result.LastVerifyAt)
	}
	if result.PendingVerify != 0 {
		t.Errorf("expected pending_verify=0, got %d", result.PendingVerify)
	}
}

func TestCurateVerify_GoalsMDFallback(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	goalsMD := `# Goals

## Mission
Curate verify fallback test

## Gates
| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| always-pass | ` + "`true`" + ` | 1 | always passes |
`
	if err := os.WriteFile(filepath.Join(tmpDir, "GOALS.md"), []byte(goalsMD), 0o644); err != nil {
		t.Fatalf("write GOALS.md: %v", err)
	}

	origOutput := output
	output = "json"
	t.Cleanup(func() { output = origOutput })

	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err = runCurateVerify(nil, nil)
	w.Close()
	os.Stdout = origStdout
	if err != nil {
		t.Fatalf("runCurateVerify: %v", err)
	}

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	r.Close()

	var result curateVerifyResult
	if err := json.Unmarshal(buf[:n], &result); err != nil {
		t.Fatalf("unmarshal result: %v (raw: %s)", err, string(buf[:n]))
	}
	if !result.Verified {
		t.Fatalf("expected verified=true, got false (%+v)", result)
	}
	if result.GatesPassed < 1 {
		t.Fatalf("expected at least one gate passed, got %d", result.GatesPassed)
	}
}

func TestCurateParseFrontmatter_YAMLMultiline(t *testing.T) {
	input := `---
type: learning
date: 2026-02-24
content: |
  line one
  line two
tags:
  - constraint
  - parser
---
Body fallback content
`

	fm, body := curateParseFrontmatter(input)

	if got := curateFrontmatterString(fm, "type"); got != "learning" {
		t.Fatalf("type = %q, want learning", got)
	}
	if got := curateFrontmatterString(fm, "date"); got != "2026-02-24" {
		t.Fatalf("date = %q, want 2026-02-24", got)
	}
	if got := curateFrontmatterString(fm, "content"); !strings.Contains(got, "line one") || !strings.Contains(got, "line two") {
		t.Fatalf("expected multiline content in frontmatter, got %q", got)
	}
	if body != "Body fallback content" {
		t.Fatalf("body = %q, want %q", body, "Body fallback content")
	}
}

func TestCuratePipeline_CatalogAssembleVerify(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	goalsMD := `# Goals

## Mission
Pipeline integration

## Gates
| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| always-pass | ` + "`true`" + ` | 1 | always passes |
`
	if err := os.WriteFile(filepath.Join(tmpDir, "GOALS.md"), []byte(goalsMD), 0o644); err != nil {
		t.Fatalf("write GOALS.md: %v", err)
	}

	artifact := `---
type: learning
date: 2026-02-24
---
Context assembly should read curated JSON artifacts.
`
	artifactPath := filepath.Join(tmpDir, "pipeline-learning.md")
	if err := os.WriteFile(artifactPath, []byte(artifact), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	if err := runCurateCatalog(nil, []string{artifactPath}); err != nil {
		t.Fatalf("runCurateCatalog: %v", err)
	}

	sections := assembleSections(tmpDir, "Read curated artifacts", defaultAssembleMaxChars)
	if len(sections) != 5 {
		t.Fatalf("expected 5 sections, got %d", len(sections))
	}
	if !strings.Contains(sections[2].Content, "Context assembly should read curated JSON artifacts.") {
		t.Fatalf("INTEL section did not include curated learning content: %s", sections[2].Content)
	}

	origOutput := output
	output = "json"
	t.Cleanup(func() { output = origOutput })

	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err = runCurateVerify(nil, nil)
	w.Close()
	os.Stdout = origStdout
	if err != nil {
		t.Fatalf("runCurateVerify: %v", err)
	}

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	r.Close()

	var result curateVerifyResult
	if err := json.Unmarshal(buf[:n], &result); err != nil {
		t.Fatalf("unmarshal result: %v (raw: %s)", err, string(buf[:n]))
	}
	if !result.Verified {
		t.Fatalf("expected verified=true in pipeline path, got false (%+v)", result)
	}
}
