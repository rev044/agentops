package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/boshu2/agentops/cli/internal/goals"
)

func TestGoalsHistory_NoHistoryFile(t *testing.T) {
	dir := t.TempDir()

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldJSON := goalsJSON
	oldSince := goalsHistorySince
	oldGoalID := goalsHistoryGoalID
	defer func() {
		goalsJSON = oldJSON
		goalsHistorySince = oldSince
		goalsHistoryGoalID = oldGoalID
	}()
	goalsJSON = false
	goalsHistorySince = ""
	goalsHistoryGoalID = ""

	// Should print "no history" message, not error
	err := goalsHistoryCmd.RunE(goalsHistoryCmd, nil)
	if err != nil {
		t.Fatalf("history returned error: %v", err)
	}
}

func TestGoalsHistory_WithEntries(t *testing.T) {
	dir := t.TempDir()

	historyDir := filepath.Join(dir, ".agents/ao/goals")
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	entries := []goals.HistoryEntry{
		{Timestamp: "2025-06-01T12:00:00Z", GoalsPassing: 3, GoalsTotal: 5, Score: 60.0, GitSHA: "abc1234"},
		{Timestamp: "2025-06-02T12:00:00Z", GoalsPassing: 4, GoalsTotal: 5, Score: 80.0, GitSHA: "def5678"},
	}

	historyPath := filepath.Join(historyDir, "history.jsonl")
	f, err := os.Create(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		data, _ := json.Marshal(e)
		_, _ = f.Write(data)
		_, _ = f.Write([]byte("\n"))
	}
	_ = f.Close()

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldJSON := goalsJSON
	oldSince := goalsHistorySince
	oldGoalID := goalsHistoryGoalID
	defer func() {
		goalsJSON = oldJSON
		goalsHistorySince = oldSince
		goalsHistoryGoalID = oldGoalID
	}()
	goalsJSON = false
	goalsHistorySince = ""
	goalsHistoryGoalID = ""

	err = goalsHistoryCmd.RunE(goalsHistoryCmd, nil)
	if err != nil {
		t.Fatalf("history returned error: %v", err)
	}
}

func TestGoalsHistory_JSONOutput(t *testing.T) {
	dir := t.TempDir()

	historyDir := filepath.Join(dir, ".agents/ao/goals")
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	entries := []goals.HistoryEntry{
		{Timestamp: "2025-06-01T12:00:00Z", GoalsPassing: 5, GoalsTotal: 5, Score: 100.0, GitSHA: "abc1234"},
	}

	historyPath := filepath.Join(historyDir, "history.jsonl")
	f, err := os.Create(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		data, _ := json.Marshal(e)
		_, _ = f.Write(data)
		_, _ = f.Write([]byte("\n"))
	}
	_ = f.Close()

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldJSON := goalsJSON
	oldSince := goalsHistorySince
	oldGoalID := goalsHistoryGoalID
	defer func() {
		goalsJSON = oldJSON
		goalsHistorySince = oldSince
		goalsHistoryGoalID = oldGoalID
	}()
	goalsJSON = true
	goalsHistorySince = ""
	goalsHistoryGoalID = ""

	// Capture stdout
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = goalsHistoryCmd.RunE(goalsHistoryCmd, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("history returned error: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)

	var decoded []goals.HistoryEntry
	if err := json.Unmarshal(buf[:n], &decoded); err != nil {
		t.Fatalf("failed to decode JSON output: %v (raw: %s)", err, string(buf[:n]))
	}
	if len(decoded) != 1 {
		t.Errorf("expected 1 entry, got %d", len(decoded))
	}
	if decoded[0].Score != 100.0 {
		t.Errorf("Score = %f, want 100.0", decoded[0].Score)
	}
}

func TestGoalsHistory_SinceFilter(t *testing.T) {
	dir := t.TempDir()

	historyDir := filepath.Join(dir, ".agents/ao/goals")
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	entries := []goals.HistoryEntry{
		{Timestamp: "2025-01-01T12:00:00Z", GoalsPassing: 1, GoalsTotal: 5, Score: 20.0, GitSHA: "old1"},
		{Timestamp: "2025-06-15T12:00:00Z", GoalsPassing: 4, GoalsTotal: 5, Score: 80.0, GitSHA: "new1"},
		{Timestamp: "2025-07-01T12:00:00Z", GoalsPassing: 5, GoalsTotal: 5, Score: 100.0, GitSHA: "new2"},
	}

	historyPath := filepath.Join(historyDir, "history.jsonl")
	f, err := os.Create(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		data, _ := json.Marshal(e)
		_, _ = f.Write(data)
		_, _ = f.Write([]byte("\n"))
	}
	_ = f.Close()

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldJSON := goalsJSON
	oldSince := goalsHistorySince
	oldGoalID := goalsHistoryGoalID
	defer func() {
		goalsJSON = oldJSON
		goalsHistorySince = oldSince
		goalsHistoryGoalID = oldGoalID
	}()
	goalsJSON = true
	goalsHistorySince = "2025-06-01"
	goalsHistoryGoalID = ""

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err = goalsHistoryCmd.RunE(goalsHistoryCmd, nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("history returned error: %v", err)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)

	var decoded []goals.HistoryEntry
	if err := json.Unmarshal(buf[:n], &decoded); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if len(decoded) != 2 {
		t.Errorf("expected 2 entries after --since 2025-06-01, got %d", len(decoded))
	}
}

func TestGoalsHistory_InvalidSinceDate(t *testing.T) {
	dir := t.TempDir()

	historyDir := filepath.Join(dir, ".agents/ao/goals")
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write at least one entry so we don't hit the "no entries" path
	historyPath := filepath.Join(historyDir, "history.jsonl")
	entry := goals.HistoryEntry{Timestamp: "2025-06-01T12:00:00Z", GoalsPassing: 1, GoalsTotal: 1, Score: 100.0}
	data, _ := json.Marshal(entry)
	if err := os.WriteFile(historyPath, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	oldSince := goalsHistorySince
	oldGoalID := goalsHistoryGoalID
	defer func() {
		goalsHistorySince = oldSince
		goalsHistoryGoalID = oldGoalID
	}()
	goalsHistorySince = "not-a-date"
	goalsHistoryGoalID = ""

	err := goalsHistoryCmd.RunE(goalsHistoryCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid --since date")
	}
	if !strings.Contains(err.Error(), "invalid --since date") {
		t.Errorf("error = %q, want 'invalid --since date'", err.Error())
	}
}

func TestGoalsHistory_CmdAttributes(t *testing.T) {
	if goalsHistoryCmd.Use != "history" {
		t.Errorf("Use = %q, want history", goalsHistoryCmd.Use)
	}
	if goalsHistoryCmd.GroupID != "analysis" {
		t.Errorf("GroupID = %q, want analysis", goalsHistoryCmd.GroupID)
	}
	found := false
	for _, a := range goalsHistoryCmd.Aliases {
		if a == "h" {
			found = true
		}
	}
	if !found {
		t.Error("expected alias 'h' for history command")
	}
}
