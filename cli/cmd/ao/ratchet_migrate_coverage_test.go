package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// --- runRatchetMigrate tests ---

func TestCov3_ratchetMigrate_runRatchetMigrate_dryRun(t *testing.T) {
	tmp := t.TempDir()
	chdirTo(t, tmp)

	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	cmd := &cobra.Command{}
	got := captureJSONStdout(t, func() {
		err := runRatchetMigrate(cmd, nil)
		if err != nil {
			t.Fatalf("runRatchetMigrate dry-run: %v", err)
		}
	})

	if !strings.Contains(got, "Would migrate chain from") {
		t.Fatalf("expected dry-run output, got: %s", got)
	}
}

func TestCov3_ratchetMigrate_runRatchetMigrate_noLegacy(t *testing.T) {
	tmp := t.TempDir()
	chdirTo(t, tmp)

	// Create .agents/ dir but no legacy chain
	if err := os.MkdirAll(filepath.Join(tmp, ".agents"), 0755); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	cmd := &cobra.Command{}
	err := runRatchetMigrate(cmd, nil)
	if err == nil {
		t.Fatal("expected error when no legacy chain")
	}
	if !strings.Contains(err.Error(), "migrate chain") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCov3_ratchetMigrate_runRatchetMigrate_withLegacy(t *testing.T) {
	tmp := t.TempDir()
	chdirTo(t, tmp)

	// Create .agents/provenance/chain.yaml with valid YAML
	legacyDir := filepath.Join(tmp, ".agents", "provenance")
	aoDir := filepath.Join(tmp, ".agents", "ao")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(aoDir, 0755); err != nil {
		t.Fatal(err)
	}

	legacyContent := `entries:
  - step: research
    gate: research
    status: passed
    timestamp: "2026-01-01T00:00:00Z"
    artifact_path: ".agents/research/test.md"
`
	if err := os.WriteFile(filepath.Join(legacyDir, "chain.yaml"), []byte(legacyContent), 0644); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	cmd := &cobra.Command{}
	got := captureJSONStdout(t, func() {
		err := runRatchetMigrate(cmd, nil)
		if err != nil {
			t.Fatalf("runRatchetMigrate with legacy: %v", err)
		}
	})

	if !strings.Contains(got, "Migrated") || !strings.Contains(got, "Migration complete") {
		t.Fatalf("expected migration output, got: %s", got)
	}
}

// --- runMigrateArtifacts tests ---

func TestCov3_ratchetMigrate_runMigrateArtifacts_noArgs(t *testing.T) {
	tmp := t.TempDir()
	chdirTo(t, tmp)

	// Create .agents/ directory with a markdown file
	agentsDir := filepath.Join(tmp, ".agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	mdContent := "# Test Artifact\n**Date:** 2026-01-01\nSome content\n"
	if err := os.WriteFile(filepath.Join(agentsDir, "test.md"), []byte(mdContent), 0644); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	cmd := &cobra.Command{}
	got := captureJSONStdout(t, func() {
		err := runMigrateArtifacts(cmd, nil)
		if err != nil {
			t.Fatalf("runMigrateArtifacts no args: %v", err)
		}
	})

	if !strings.Contains(got, "Summary:") {
		t.Fatalf("expected summary output, got: %s", got)
	}
}

func TestCov3_ratchetMigrate_runMigrateArtifacts_withPath(t *testing.T) {
	tmp := t.TempDir()

	// Create a subdir with a markdown file
	subDir := filepath.Join(tmp, "learnings")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	mdContent := "# Learning\n**Date:** 2026-01-01\nContent\n"
	if err := os.WriteFile(filepath.Join(subDir, "learn.md"), []byte(mdContent), 0644); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	cmd := &cobra.Command{}
	got := captureJSONStdout(t, func() {
		err := runMigrateArtifacts(cmd, []string{subDir})
		if err != nil {
			t.Fatalf("runMigrateArtifacts with path: %v", err)
		}
	})

	if !strings.Contains(got, "Summary:") {
		t.Fatalf("expected summary output, got: %s", got)
	}
}

func TestCov3_ratchetMigrate_runMigrateArtifacts_dryRun(t *testing.T) {
	tmp := t.TempDir()

	mdContent := "# Artifact\n**Epic:** test-epic\nSome content\n"
	if err := os.WriteFile(filepath.Join(tmp, "artifact.md"), []byte(mdContent), 0644); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = true
	defer func() { dryRun = oldDryRun }()

	cmd := &cobra.Command{}
	got := captureJSONStdout(t, func() {
		err := runMigrateArtifacts(cmd, []string{tmp})
		if err != nil {
			t.Fatalf("runMigrateArtifacts dry-run: %v", err)
		}
	})

	if !strings.Contains(got, "Would add schema_version") {
		t.Fatalf("expected dry-run output, got: %s", got)
	}
}

// --- shouldMigrateFile tests ---

func TestCov3_ratchetMigrate_shouldMigrateFile_markdown(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "test.md")
	if err := os.WriteFile(p, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if !shouldMigrateFile(p, info) {
		t.Fatal("expected markdown file to be eligible")
	}
}

func TestCov3_ratchetMigrate_shouldMigrateFile_nonMarkdown(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "test.json")
	if err := os.WriteFile(p, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if shouldMigrateFile(p, info) {
		t.Fatal("expected non-markdown file to be ineligible")
	}
}

func TestCov3_ratchetMigrate_shouldMigrateFile_directory(t *testing.T) {
	tmp := t.TempDir()
	info, err := os.Stat(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if shouldMigrateFile(tmp, info) {
		t.Fatal("expected directory to be ineligible")
	}
}

// --- findSchemaInsertPoint tests ---

func TestCov3_ratchetMigrate_findSchemaInsertPoint_afterDate(t *testing.T) {
	lines := []string{"# Title", "**Date:** 2026-01-01", "Content"}
	idx := findSchemaInsertPoint(lines)
	if idx != 2 {
		t.Fatalf("expected insert at 2, got %d", idx)
	}
}

func TestCov3_ratchetMigrate_findSchemaInsertPoint_afterEpic(t *testing.T) {
	lines := []string{"# Title", "**Epic:** test", "Content"}
	idx := findSchemaInsertPoint(lines)
	if idx != 2 {
		t.Fatalf("expected insert at 2, got %d", idx)
	}
}

func TestCov3_ratchetMigrate_findSchemaInsertPoint_afterHeading(t *testing.T) {
	lines := []string{"# Title", "Some content", "More"}
	idx := findSchemaInsertPoint(lines)
	if idx != 1 {
		t.Fatalf("expected insert at 1 (after heading), got %d", idx)
	}
}

func TestCov3_ratchetMigrate_findSchemaInsertPoint_noMatch(t *testing.T) {
	lines := []string{"no heading", "no date", "no epic"}
	idx := findSchemaInsertPoint(lines)
	if idx != -1 {
		t.Fatalf("expected -1 for no match, got %d", idx)
	}
}

// --- migrateFile tests ---

func TestCov3_ratchetMigrate_migrateFile_alreadyHasSchema(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "already.md")
	content := "# Title\n**Schema Version:** 1\nContent\n"
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	result := migrateFile(p, info)
	if result != migrateResultSkipped {
		t.Fatalf("expected skipped for file with schema, got %d", result)
	}
}

func TestCov3_ratchetMigrate_migrateFile_success(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "migrate-me.md")
	content := "# Title\n**Date:** 2026-01-01\nContent\n"
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	oldDryRun := dryRun
	dryRun = false
	defer func() { dryRun = oldDryRun }()

	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}

	captureJSONStdout(t, func() {
		result := migrateFile(p, info)
		if result != migrateResultSuccess {
			t.Fatalf("expected success, got %d", result)
		}
	})

	// Verify schema_version was added
	updated, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updated), "**Schema Version:** 1") {
		t.Fatalf("expected schema version in updated file, got: %s", string(updated))
	}
}

func TestCov3_ratchetMigrate_migrateFile_noInsertPoint(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "no-heading.md")
	content := "plain text\nno heading\nno date\n"
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	result := migrateFile(p, info)
	if result != migrateResultSkipped {
		t.Fatalf("expected skipped for no insert point, got %d", result)
	}
}
