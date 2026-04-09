package ratchet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSkillDrafts_RequiresThreeSessionRefs(t *testing.T) {
	baseDir := t.TempDir()
	patternsDir := filepath.Join(baseDir, ".agents", "patterns")
	sessionsDir := filepath.Join(baseDir, ".agents", "ao", "sessions")
	if err := os.MkdirAll(patternsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	patternPath := filepath.Join(patternsDir, "2026-04-09-test-pattern.md")
	if err := os.WriteFile(patternPath, []byte("---\nconfidence: 0.9\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	writeSessionRef := func(name string) {
		t.Helper()
		path := filepath.Join(sessionsDir, name)
		if err := os.WriteFile(path, []byte("used 2026-04-09-test-pattern.md\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeSessionRef("sess1.jsonl")
	writeSessionRef("sess2.md")

	result, err := GenerateSkillDrafts(baseDir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Generated != 0 {
		t.Fatalf("expected no drafts with two session refs, got %+v", result)
	}

	writeSessionRef("sess3.jsonl")

	result, err = GenerateSkillDrafts(baseDir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Generated != 1 {
		t.Fatalf("expected one draft with three session refs, got %+v", result)
	}
	if len(result.Paths) != 1 {
		t.Fatalf("expected one draft path, got %+v", result.Paths)
	}

	draftPath := result.Paths[0]
	data, err := os.ReadFile(draftPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "name: test") {
		t.Fatalf("expected draft frontmatter name to derive from pattern slug, got:\n%s", text)
	}
	if !strings.Contains(text, "Draft generated from recurring pattern evidence") {
		t.Fatalf("expected generated draft description, got:\n%s", text)
	}

	evidencePath := filepath.Join(filepath.Dir(draftPath), "evidence.json")
	evidenceBytes, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatal(err)
	}

	var evidence skillDraftEvidence
	if err := json.Unmarshal(evidenceBytes, &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.SessionRefs != 3 {
		t.Fatalf("expected session_refs=3, got %+v", evidence)
	}
	if evidence.PatternPath != patternPath {
		t.Fatalf("expected evidence pattern path %q, got %+v", patternPath, evidence)
	}
}
