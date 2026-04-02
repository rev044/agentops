package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFactoryStartJSONBuildsBriefingThenRunsCodexStartup(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_THREAD_ID", "019d1bf7-58ea-79e1-9f5d-02109d930081")
	t.Setenv("CODEX_INTERNAL_ORIGINATOR_OVERRIDE", "Codex Desktop")

	indexPath := filepath.Join(home, ".codex", "session_index.jsonl")
	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexPath, []byte(`{"id":"019d1bf7-58ea-79e1-9f5d-02109d930081","thread_name":"factory startup","updated_at":"2026-03-23T12:00:00Z"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := t.TempDir()
	writeKnowledgeCorpusFixtures(t, repo)

	origProjectDir := testProjectDir
	testProjectDir = repo
	defer func() { testProjectDir = origProjectDir }()

	out, err := executeCommand("factory", "start", "--json", "--no-maintenance", "--goal", "Healthy topic rollout")
	if err != nil {
		t.Fatalf("factory start --json: %v\noutput: %s", err, out)
	}

	var result factoryStartResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("parse factory start json: %v\noutput: %s", err, out)
	}

	if result.Goal != "Healthy topic rollout" {
		t.Fatalf("goal = %q, want %q", result.Goal, "Healthy topic rollout")
	}
	if result.Briefing == "" || !knowledgePathExists(result.Briefing) {
		t.Fatalf("briefing missing: %q", result.Briefing)
	}
	if result.Codex.StartupContextPath == "" || !fileExists(result.Codex.StartupContextPath) {
		t.Fatalf("startup context missing: %q", result.Codex.StartupContextPath)
	}
	if len(result.Codex.Briefings) == 0 {
		t.Fatal("expected codex startup to surface the generated briefing")
	}
	if len(result.Recommended) == 0 {
		t.Fatal("expected recommended next steps")
	}

	data, err := os.ReadFile(result.Codex.StartupContextPath)
	if err != nil {
		t.Fatalf("read startup context: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "# Briefings") {
		t.Fatalf("startup context missing briefings section:\n%s", content)
	}
	if !strings.Contains(content, "Healthy topic rollout") {
		t.Fatalf("startup context missing goal-specific briefing:\n%s", content)
	}
}

func TestFactoryStartJSONWarnsWhenBriefingCannotBeBuiltYet(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_THREAD_ID", "019d1bf7-58ea-79e1-9f5d-02109d930082")
	t.Setenv("CODEX_INTERNAL_ORIGINATOR_OVERRIDE", "Codex Desktop")

	indexPath := filepath.Join(home, ".codex", "session_index.jsonl")
	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexPath, []byte(`{"id":"019d1bf7-58ea-79e1-9f5d-02109d930082","thread_name":"factory startup no corpus","updated_at":"2026-03-23T12:00:00Z"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}

	origProjectDir := testProjectDir
	testProjectDir = repo
	defer func() { testProjectDir = origProjectDir }()

	out, err := executeCommand("factory", "start", "--json", "--no-maintenance", "--goal", "No topic packets yet")
	if err != nil {
		t.Fatalf("factory start without topic packets should still succeed: %v\noutput: %s", err, out)
	}

	var result factoryStartResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("parse factory start json: %v\noutput: %s", err, out)
	}
	if result.Briefing != "" {
		t.Fatalf("briefing = %q, want empty when no topic packets exist", result.Briefing)
	}
	if result.BriefingWarning == "" {
		t.Fatal("expected briefing warning when no topic packets exist")
	}
	if result.Codex.StartupContextPath == "" || !fileExists(result.Codex.StartupContextPath) {
		t.Fatalf("startup context missing: %q", result.Codex.StartupContextPath)
	}
}
