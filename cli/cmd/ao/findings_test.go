package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestFindingsListJSON_ExcludesRetiredByDefault(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeFindingFixture(t, repo, "f-active", "Active finding", "active", 2)
	writeFindingFixture(t, repo, "f-retired", "Retired finding", "retired", 9)

	out, err := executeCommand("findings", "list", "--json")
	if err != nil {
		t.Fatalf("findings list failed: %v\n%s", err, out)
	}

	var findings []knowledgeFinding
	if err := json.Unmarshal([]byte(out), &findings); err != nil {
		t.Fatalf("parse findings list json: %v\n%s", err, out)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 active finding, got %d", len(findings))
	}
	if findings[0].ID != "f-active" {
		t.Fatalf("expected active finding, got %q", findings[0].ID)
	}
}

func TestFindingsRetire_UpdatesFrontMatter(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	path := writeFindingFixture(t, repo, "f-retire", "Retire me", "active", 1)

	out, err := executeCommand("findings", "retire", "f-retire", "--by", "tester", "--json")
	if err != nil {
		t.Fatalf("findings retire failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read retired finding: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "status: retired") {
		t.Fatalf("retired file missing retired status:\n%s", content)
	}
	if !strings.Contains(content, "retired_by: tester") {
		t.Fatalf("retired file missing retired_by marker:\n%s", content)
	}
}

func TestFindingsPullAndExport_JSON(t *testing.T) {
	sourceRepo := t.TempDir()
	destRepo := t.TempDir()
	exportRepo := t.TempDir()

	writeFindingFixture(t, sourceRepo, "f-pull", "Pulled finding", "active", 4)

	t.Chdir(destRepo)
	out, err := executeCommand("findings", "pull", "f-pull", "--from", sourceRepo, "--json")
	if err != nil {
		t.Fatalf("findings pull failed: %v\n%s", err, out)
	}

	localPath := filepath.Join(destRepo, ".agents", SectionFindings, "f-pull.md")
	if _, err := os.Stat(localPath); err != nil {
		t.Fatalf("expected pulled finding at %s: %v", localPath, err)
	}

	out, err = executeCommand("findings", "export", "f-pull", "--to", exportRepo, "--json")
	if err != nil {
		t.Fatalf("findings export failed: %v\n%s", err, out)
	}

	exportPath := filepath.Join(exportRepo, ".agents", SectionFindings, "f-pull.md")
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("expected exported finding at %s: %v", exportPath, err)
	}
}

func TestFindingsStats_JSON(t *testing.T) {
	repo := t.TempDir()
	t.Chdir(repo)
	writeFindingFixture(t, repo, "f-one", "One", "active", 3)
	writeFindingFixture(t, repo, "f-two", "Two", "retired", 5)

	out, err := executeCommand("findings", "stats", "--json")
	if err != nil {
		t.Fatalf("findings stats failed: %v\n%s", err, out)
	}

	var stats findingStats
	if err := json.Unmarshal([]byte(out), &stats); err != nil {
		t.Fatalf("parse findings stats json: %v\n%s", err, out)
	}
	if stats.Total != 2 {
		t.Fatalf("total = %d, want 2", stats.Total)
	}
	if stats.ByStatus["active"] != 1 || stats.ByStatus["retired"] != 1 {
		t.Fatalf("unexpected status counts: %+v", stats.ByStatus)
	}
	if stats.TotalHits != 8 {
		t.Fatalf("total_hits = %d, want 8", stats.TotalHits)
	}
	if len(stats.MostCited) == 0 || stats.MostCited[0].ID != "f-two" {
		t.Fatalf("most cited ordering incorrect: %+v", stats.MostCited)
	}
}

func writeFindingFixture(t *testing.T, repoRoot, id, title, status string, hits int) string {
	t.Helper()
	dir := filepath.Join(repoRoot, ".agents", SectionFindings)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir findings dir: %v", err)
	}
	path := filepath.Join(dir, id+".md")
	content := `---
id: ` + id + `
title: ` + title + `
source_skill: vibe
severity: high
detectability: mechanical
status: ` + status + `
compiler_targets: [inject, lookup]
scope_tags: [cli, validation]
applicable_when: [plan-shape, validation-gap]
applicable_languages: [go]
hit_count: ` + strconv.Itoa(hits) + `
last_cited: 2026-03-10T01:00:00Z
---

# ` + title + `

Summary text for ` + id + `.
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write finding fixture: %v", err)
	}
	return path
}
