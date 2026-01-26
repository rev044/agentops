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
	for _, h := range hooks.SessionStart {
		if len(h.Command) > 1 {
			for _, c := range h.Command {
				if c == "ao inject --apply-decay --max-tokens 1500 2>/dev/null || true" {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("expected ao inject command in SessionStart hooks")
	}

	// Verify Stop contains ao forge
	found = false
	for _, h := range hooks.Stop {
		if len(h.Command) > 1 {
			for _, c := range h.Command {
				if c == "ao forge transcript --last-session --quiet --queue 2>/dev/null; ao task-sync --promote 2>/dev/null || true" {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("expected ao forge command in Stop hooks")
	}
}

func TestHookConfigStructure(t *testing.T) {
	hooks := generateHooksConfig()

	for i, h := range hooks.SessionStart {
		if h.Command == nil {
			t.Errorf("SessionStart hook %d has nil command", i)
		}
		if len(h.Command) < 2 {
			t.Errorf("SessionStart hook %d command too short: %v", i, h.Command)
		}
	}

	for i, h := range hooks.Stop {
		if h.Command == nil {
			t.Errorf("Stop hook %d has nil command", i)
		}
		if len(h.Command) < 2 {
			t.Errorf("Stop hook %d command too short: %v", i, h.Command)
		}
	}
}
