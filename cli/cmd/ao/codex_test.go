package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDetectCodexLifecycleProfile_Hookless(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_THREAD_ID", "019d1bf7-58ea-79e1-9f5d-02109d930081")
	t.Setenv("CODEX_INTERNAL_ORIGINATOR_OVERRIDE", "Codex Desktop")

	indexPath := filepath.Join(home, ".codex", "session_index.jsonl")
	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		t.Fatal(err)
	}
	line := `{"id":"019d1bf7-58ea-79e1-9f5d-02109d930081","thread_name":"Codex lifecycle fallback","updated_at":"2026-03-23T12:00:00Z"}`
	if err := os.WriteFile(indexPath, []byte(line+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	profile := detectCodexLifecycleProfile()
	if profile.Runtime != runtimeKindCodex {
		t.Fatalf("runtime = %q, want %q", profile.Runtime, runtimeKindCodex)
	}
	if profile.Mode != lifecycleModeCodexHookless {
		t.Fatalf("mode = %q, want %q", profile.Mode, lifecycleModeCodexHookless)
	}
	if profile.HookCapable {
		t.Fatal("hook_capable = true, want false for Codex")
	}
	if profile.ThreadName != "Codex lifecycle fallback" {
		t.Fatalf("thread_name = %q, want %q", profile.ThreadName, "Codex lifecycle fallback")
	}
}

func TestFindLastSession_PrefersNewestCodexArchivedTranscript(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	claudePath := filepath.Join(home, ".claude", "projects", "repo", "conversations", "claude.jsonl")
	if err := os.MkdirAll(filepath.Dir(claudePath), 0o755); err != nil {
		t.Fatal(err)
	}
	claudeContent := strings.Repeat(`{"type":"user","message":{"content":"hello"}}`+"\n", 4)
	if err := os.WriteFile(claudePath, []byte(claudeContent), 0o644); err != nil {
		t.Fatal(err)
	}
	older := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(claudePath, older, older); err != nil {
		t.Fatal(err)
	}

	codexPath := filepath.Join(home, ".codex", "archived_sessions", "rollout-2026-03-23T12-00-00-019d1bf7-58ea-79e1-9f5d-02109d930081.jsonl")
	if err := os.MkdirAll(filepath.Dir(codexPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(codexPath, []byte(`{"type":"session_meta","payload":{"id":"019d1bf7-58ea-79e1-9f5d-02109d930081"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	newer := older.Add(24 * time.Hour)
	if err := os.Chtimes(codexPath, newer, newer); err != nil {
		t.Fatal(err)
	}

	got, err := findLastSession()
	if err != nil {
		t.Fatalf("findLastSession: %v", err)
	}
	if got != codexPath {
		t.Fatalf("findLastSession = %q, want %q", got, codexPath)
	}
}

func TestFindTranscriptBySessionID_FindsCodexArchivedTranscript(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	sessionID := "019d1bf7-58ea-79e1-9f5d-02109d930081"
	path := filepath.Join(home, ".codex", "archived_sessions", "rollout-2026-03-23T12-00-00-"+sessionID+".jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"type":"session_meta","payload":{"id":"`+sessionID+`"}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := findTranscriptBySessionID(sessionID)
	if err != nil {
		t.Fatalf("findTranscriptBySessionID: %v", err)
	}
	if got != path {
		t.Fatalf("findTranscriptBySessionID = %q, want %q", got, path)
	}
}

func TestSynthesizeCodexHistoryTranscript(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	sessionID := "019d1bf7-58ea-79e1-9f5d-02109d930081"
	historyPath := filepath.Join(home, ".codex", "history.jsonl")
	indexPath := filepath.Join(home, ".codex", "session_index.jsonl")
	if err := os.MkdirAll(filepath.Dir(historyPath), 0o755); err != nil {
		t.Fatal(err)
	}

	history := []string{
		`{"session_id":"` + sessionID + `","ts":1766945654,"text":"exit"}`,
		`{"session_id":"` + sessionID + `","ts":1766945655,"text":"Investigate lifecycle fallback"}`,
		`{"session_id":"` + sessionID + `","ts":1766945656,"text":"Investigate lifecycle fallback"}`,
		`{"session_id":"` + sessionID + `","ts":1766945657,"text":"/model default"}`,
		`{"session_id":"` + sessionID + `","ts":1766945658,"text":"Record explicit Codex closeout"}`,
	}
	if err := os.WriteFile(historyPath, []byte(strings.Join(history, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexPath, []byte(`{"id":"`+sessionID+`","thread_name":"Lifecycle fallback","updated_at":"2026-03-23T12:00:00Z"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := t.TempDir()
	path, err := synthesizeCodexHistoryTranscript(repo, sessionID)
	if err != nil {
		t.Fatalf("synthesizeCodexHistoryTranscript: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, `"type":"session_meta"`) {
		t.Fatalf("expected session_meta line in %s", content)
	}
	if !strings.Contains(content, `"thread_name":"Lifecycle fallback"`) {
		t.Fatalf("expected thread_name in synthesized transcript: %s", content)
	}
	if strings.Contains(content, `"message":"exit"`) {
		t.Fatalf("control message should be filtered out: %s", content)
	}
	if strings.Count(content, `"type":"event_msg"`) != 2 {
		t.Fatalf("expected 2 event messages after compaction, got %d in %s", strings.Count(content, `"type":"event_msg"`), content)
	}
}

func TestCodexStartJSONWritesStateAndCitations(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_THREAD_ID", "019d1bf7-58ea-79e1-9f5d-02109d930081")
	t.Setenv("CODEX_INTERNAL_ORIGINATOR_OVERRIDE", "Codex Desktop")

	indexPath := filepath.Join(home, ".codex", "session_index.jsonl")
	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexPath, []byte(`{"id":"019d1bf7-58ea-79e1-9f5d-02109d930081","thread_name":"explicit lifecycle","updated_at":"2026-03-23T12:00:00Z"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".agents", "learnings"), 0o755); err != nil {
		t.Fatal(err)
	}
	learningPath := filepath.Join(repo, ".agents", "learnings", "codex-lifecycle.md")
	learning := `---
id: codex-lifecycle
type: learning
date: 2026-03-23
source: codex-test
maturity: provisional
utility: 0.9
---

# Explicit Codex lifecycle

Use ao codex start and ao codex stop when runtime hooks are unavailable.
`
	if err := os.WriteFile(learningPath, []byte(learning), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	out, err := executeCommand("codex", "start", "--json", "--query", "explicit codex lifecycle")
	if err != nil {
		t.Fatalf("codex start --json: %v\noutput: %s", err, out)
	}

	var result codexStartResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("parse codex start json: %v\noutput: %s", err, out)
	}
	if result.Runtime.Mode != lifecycleModeCodexHookless {
		t.Fatalf("runtime mode = %q, want %q", result.Runtime.Mode, lifecycleModeCodexHookless)
	}
	if result.StartupContextPath == "" || !fileExists(result.StartupContextPath) {
		t.Fatalf("startup context path missing or unreadable: %q", result.StartupContextPath)
	}
	if result.StatePath == "" || !fileExists(result.StatePath) {
		t.Fatalf("state path missing or unreadable: %q", result.StatePath)
	}
	citationsPath := filepath.Join(repo, ".agents", "ao", "citations.jsonl")
	data, err := os.ReadFile(citationsPath)
	if err != nil {
		t.Fatalf("read citations: %v", err)
	}
	if !strings.Contains(string(data), `"citation_type":"retrieved"`) {
		t.Fatalf("expected retrieved citation in %s", string(data))
	}
}

func TestCodexStopJSONUsesHistoryFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_THREAD_ID", "019d1bf7-58ea-79e1-9f5d-02109d930081")
	t.Setenv("CODEX_INTERNAL_ORIGINATOR_OVERRIDE", "Codex Desktop")

	sessionID := "019d1bf7-58ea-79e1-9f5d-02109d930081"
	historyPath := filepath.Join(home, ".codex", "history.jsonl")
	indexPath := filepath.Join(home, ".codex", "session_index.jsonl")
	if err := os.MkdirAll(filepath.Dir(historyPath), 0o755); err != nil {
		t.Fatal(err)
	}
	history := []string{
		`{"session_id":"` + sessionID + `","ts":1766945655,"text":"Design Codex fallback lifecycle"}`,
		`{"session_id":"` + sessionID + `","ts":1766945658,"text":"Implement explicit ao codex stop closeout"}`,
	}
	if err := os.WriteFile(historyPath, []byte(strings.Join(history, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexPath, []byte(`{"id":"`+sessionID+`","thread_name":"Lifecycle fallback stop","updated_at":"2026-03-23T12:00:00Z"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	out, err := executeCommand("codex", "stop", "--json")
	if err != nil {
		t.Fatalf("codex stop --json: %v\noutput: %s", err, out)
	}

	var result codexStopResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("parse codex stop json: %v\noutput: %s", err, out)
	}
	if result.TranscriptSource != "history-fallback" {
		t.Fatalf("transcript_source = %q, want history-fallback", result.TranscriptSource)
	}
	if !result.SyntheticTranscript {
		t.Fatal("synthetic_transcript = false, want true")
	}
	if result.Session.SessionID == "" {
		t.Fatal("expected session close result to include a session ID")
	}
	if !fileExists(result.TranscriptPath) {
		t.Fatalf("synthetic transcript path missing: %s", result.TranscriptPath)
	}
}

func TestCodexStatusJSONReflectsHooklessHealthAndSearchCitations(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CODEX_THREAD_ID", "019d1bf7-58ea-79e1-9f5d-02109d930081")
	t.Setenv("CODEX_INTERNAL_ORIGINATOR_OVERRIDE", "Codex Desktop")

	indexPath := filepath.Join(home, ".codex", "session_index.jsonl")
	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexPath, []byte(`{"id":"019d1bf7-58ea-79e1-9f5d-02109d930081","thread_name":"Codex lifecycle status","updated_at":"2026-03-23T12:00:00Z"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".agents", "learnings"), 0o755); err != nil {
		t.Fatal(err)
	}
	learningPath := filepath.Join(repo, ".agents", "learnings", "codex-status.md")
	learning := `---
id: codex-status
type: learning
date: 2026-03-23
source: codex-test
maturity: provisional
utility: 0.9
---

# Codex lifecycle status

Use ao codex start, ao search --cite applied, and ao codex stop when runtime hooks are unavailable.
`
	if err := os.WriteFile(learningPath, []byte(learning), 0o644); err != nil {
		t.Fatal(err)
	}
	searchableLearningPath := filepath.Join(repo, ".agents", "learnings", "codex-status.jsonl")
	searchableLearning := `{"summary":"codex lifecycle status from explicit search citation","utility":0.8,"maturity":"provisional"}`
	if err := os.WriteFile(searchableLearningPath, []byte(searchableLearning+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".agents", "knowledge", "pending"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".agents", "knowledge", "pending", "queued-learning.md"), []byte("# queued\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".agents", "knowledge", "pending", ".quarantine"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, ".agents", "knowledge", "pending", ".quarantine", "truncated-learning.md"), []byte("# quarantined\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if out, err := executeCommand("codex", "start", "--json", "--no-maintenance", "--query", "codex lifecycle status"); err != nil {
		t.Fatalf("codex start --json: %v\noutput: %s", err, out)
	}

	recordSearchCitations(cwd, []searchResult{{Path: filepath.Join(cwd, ".agents", "learnings", "codex-status.jsonl")}}, resolveSessionID(""), "codex lifecycle status", "applied")

	citationsPath := filepath.Join(repo, ".agents", "ao", "citations.jsonl")
	citations, err := os.ReadFile(citationsPath)
	if err != nil {
		t.Fatalf("read citations: %v", err)
	}
	citationsText := string(citations)
	if !strings.Contains(citationsText, `"session_id":"019d1bf7-58ea-79e1-9f5d-02109d930081"`) {
		t.Fatalf("expected Codex session ID in citations: %s", citationsText)
	}
	if !strings.Contains(citationsText, `"citation_type":"applied"`) {
		t.Fatalf("expected applied citation in %s", citationsText)
	}

	out, err := executeCommand("codex", "status", "--json", "--days", "30")
	if err != nil {
		t.Fatalf("codex status --json: %v\noutput: %s", err, out)
	}

	var result codexStatusResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Fatalf("parse codex status json: %v\noutput: %s", err, out)
	}
	if result.Runtime.Mode != lifecycleModeCodexHookless {
		t.Fatalf("runtime mode = %q, want %q", result.Runtime.Mode, lifecycleModeCodexHookless)
	}
	if result.Capture.PendingKnowledge != 1 {
		t.Fatalf("pending knowledge = %d, want 1", result.Capture.PendingKnowledge)
	}
	if result.Capture.PendingQuarantine != 1 {
		t.Fatalf("pending quarantine = %d, want 1", result.Capture.PendingQuarantine)
	}
	if result.Citations.Total != 3 {
		t.Fatalf("citation total = %d, want 3", result.Citations.Total)
	}
	if result.Citations.UniqueArtifacts != 2 {
		t.Fatalf("unique artifacts = %d, want 2", result.Citations.UniqueArtifacts)
	}
	if result.Citations.Retrieved != 2 {
		t.Fatalf("retrieved citations = %d, want 2", result.Citations.Retrieved)
	}
	if result.Citations.Applied != 1 {
		t.Fatalf("applied citations = %d, want 1", result.Citations.Applied)
	}
	if result.State == nil || result.State.LastStart == nil {
		t.Fatalf("expected persisted codex lifecycle state, got %+v", result.State)
	}
}
