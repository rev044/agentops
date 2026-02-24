package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGoalSlug(t *testing.T) {
	tests := []struct {
		goal string
		want string
	}{
		{"add evolve watchdog", "evolve-watchdog"},
		{"Add CLI dashboard and reporting", "cli-dashboard-reporting"},
		{"consolidate skills to 40", "consolidate-skills-40"},
		{"Create a Codex-native plugin", "create-codexnative-plugin"},
		{"fix the bug", "fix-bug"},
		{"", ""},
		{"the a an to for", ""},
	}
	for _, tt := range tests {
		got := goalSlug(tt.goal)
		if got != tt.want {
			t.Errorf("goalSlug(%q) = %q, want %q", tt.goal, got, tt.want)
		}
	}
}

func TestResolveParallelEpics_Args(t *testing.T) {
	// Reset global.
	parallelManifest = ""

	args := []string{"add evolve watchdog", "add CLI dashboard"}
	epics, err := resolveParallelEpics(args)
	if err != nil {
		t.Fatalf("resolveParallelEpics: %v", err)
	}
	if len(epics) != 2 {
		t.Fatalf("expected 2 epics, got %d", len(epics))
	}
	if epics[0].Goal != "add evolve watchdog" {
		t.Errorf("epic 0 goal = %q, want %q", epics[0].Goal, "add evolve watchdog")
	}
	if epics[0].MergeOrder != 1 {
		t.Errorf("epic 0 merge_order = %d, want 1", epics[0].MergeOrder)
	}
	if epics[1].MergeOrder != 2 {
		t.Errorf("epic 1 merge_order = %d, want 2", epics[1].MergeOrder)
	}
}

func TestResolveParallelEpics_Manifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifest := parallelManifestFile{
		Epics: []parallelEpic{
			{Name: "evolve", Goal: "Add watchdog", MergeOrder: 2},
			{Name: "cli-obs", Goal: "Add dashboard", MergeOrder: 1},
		},
	}
	data, _ := json.Marshal(manifest)
	manifestPath := filepath.Join(tmpDir, "epics.json")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	parallelManifest = manifestPath
	defer func() { parallelManifest = "" }()

	epics, err := resolveParallelEpics(nil)
	if err != nil {
		t.Fatalf("resolveParallelEpics manifest: %v", err)
	}
	if len(epics) != 2 {
		t.Fatalf("expected 2 epics, got %d", len(epics))
	}
	if epics[0].Name != "evolve" {
		t.Errorf("epic 0 name = %q, want %q", epics[0].Name, "evolve")
	}
	if epics[1].MergeOrder != 1 {
		t.Errorf("epic 1 merge_order = %d, want 1", epics[1].MergeOrder)
	}
}

func TestResolveParallelEpics_ManifestDuplicateName(t *testing.T) {
	tmpDir := t.TempDir()
	manifest := parallelManifestFile{
		Epics: []parallelEpic{
			{Name: "evolve", Goal: "Goal 1"},
			{Name: "evolve", Goal: "Goal 2"},
		},
	}
	data, _ := json.Marshal(manifest)
	manifestPath := filepath.Join(tmpDir, "epics.json")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	parallelManifest = manifestPath
	defer func() { parallelManifest = "" }()

	_, err := resolveParallelEpics(nil)
	if err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	}
}

func TestResolveParallelEpics_Empty(t *testing.T) {
	parallelManifest = ""
	epics, err := resolveParallelEpics(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(epics) != 0 {
		t.Fatalf("expected 0 epics, got %d", len(epics))
	}
}

func TestResolveMergeOrder_Default(t *testing.T) {
	epics := []parallelEpic{
		{Name: "c", MergeOrder: 3},
		{Name: "a", MergeOrder: 1},
		{Name: "b", MergeOrder: 2},
	}
	results := []parallelResult{
		{Success: true},
		{Success: true},
		{Success: true},
	}
	parallelMergeOrder = ""
	order := resolveMergeOrder(epics, results)
	// Should be sorted by MergeOrder: a(1), b(2), c(3)
	if order[0] != 1 || order[1] != 2 || order[2] != 0 {
		t.Errorf("merge order = %v, want [1 2 0]", order)
	}
}

func TestResolveMergeOrder_Explicit(t *testing.T) {
	epics := []parallelEpic{
		{Name: "alpha"},
		{Name: "beta"},
		{Name: "gamma"},
	}
	results := []parallelResult{{}, {}, {}}
	parallelMergeOrder = "gamma,alpha,beta"
	defer func() { parallelMergeOrder = "" }()

	order := resolveMergeOrder(epics, results)
	if order[0] != 2 || order[1] != 0 || order[2] != 1 {
		t.Errorf("merge order = %v, want [2 0 1]", order)
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "''"},
		{"hello", "'hello'"},
		{"hello world", "'hello world'"},
		{"it's", "'it'\\''s'"},
		{"back`tick", "'back`tick'"},
		{`double"quote`, `'double"quote'`},
		{"multi'ple'quotes", "'multi'\\''ple'\\''quotes'"},
		{"$VAR", "'$VAR'"},
	}
	for _, tt := range tests {
		got := shellQuote(tt.input)
		if got != tt.want {
			t.Errorf("shellQuote(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTmuxPaneIsDead_Parsing(t *testing.T) {
	// tmuxPaneIsDead calls exec.Command internally, so we test the parsing
	// logic by verifying the function handles missing tmux gracefully.
	// With a non-existent tmux binary, the pane is considered dead.
	dead, code := tmuxPaneIsDead("__nonexistent_tmux_binary__", "nosession:nowindow")
	if !dead {
		t.Error("expected dead=true for missing tmux binary")
	}
	if code != 1 {
		t.Errorf("expected exitCode=1, got %d", code)
	}
}
