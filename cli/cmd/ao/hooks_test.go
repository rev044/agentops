package main

import (
	"testing"
)

func TestGenerateHooksConfig(t *testing.T) {
	hooks := generateHooksConfig()

	// Verify SessionStart hooks exist
	if len(hooks.SessionStart) == 0 {
		t.Error("expected SessionStart hooks, got none")
	}

	// Verify Stop hooks exist
	if len(hooks.Stop) == 0 {
		t.Error("expected Stop hooks, got none")
	}

	// Verify SessionStart contains ao inject
	found := false
	for _, g := range hooks.SessionStart {
		for _, h := range g.Hooks {
			if h.Type == "command" && h.Command == "ao inject --apply-decay --max-tokens 1500 2>/dev/null || true" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected ao inject command in SessionStart hooks")
	}

	// Verify Stop contains ao forge
	found = false
	for _, g := range hooks.Stop {
		for _, h := range g.Hooks {
			if h.Type == "command" && h.Command == "ao forge transcript --last-session --quiet --queue 2>/dev/null; ao task-sync --promote 2>/dev/null || true" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected ao forge command in Stop hooks")
	}
}

func TestHookConfigStructure(t *testing.T) {
	hooks := generateHooksConfig()

	for i, g := range hooks.SessionStart {
		if len(g.Hooks) == 0 {
			t.Errorf("SessionStart group %d has no hooks", i)
		}
		for j, h := range g.Hooks {
			if h.Type != "command" {
				t.Errorf("SessionStart group %d hook %d: expected type 'command', got '%s'", i, j, h.Type)
			}
			if h.Command == "" {
				t.Errorf("SessionStart group %d hook %d has empty command", i, j)
			}
		}
	}

	for i, g := range hooks.Stop {
		if len(g.Hooks) == 0 {
			t.Errorf("Stop group %d has no hooks", i)
		}
		for j, h := range g.Hooks {
			if h.Type != "command" {
				t.Errorf("Stop group %d hook %d: expected type 'command', got '%s'", i, j, h.Type)
			}
			if h.Command == "" {
				t.Errorf("Stop group %d hook %d has empty command", i, j)
			}
		}
	}
}

func TestHookGroupToMap(t *testing.T) {
	g := HookGroup{
		Hooks: []HookEntry{
			{Type: "command", Command: "echo hello"},
		},
	}

	m := hookGroupToMap(g)

	hooks, ok := m["hooks"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected hooks array in map")
	}
	if len(hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hooks))
	}
	if hooks[0]["type"] != "command" {
		t.Errorf("expected type 'command', got '%v'", hooks[0]["type"])
	}
	if hooks[0]["command"] != "echo hello" {
		t.Errorf("expected command 'echo hello', got '%v'", hooks[0]["command"])
	}

	// No matcher should be present when nil
	if _, exists := m["matcher"]; exists {
		t.Error("expected no matcher key when Matcher is nil")
	}
}
