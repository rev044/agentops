package main

import (
	"encoding/json"
	"testing"
)

func TestGenerateMinimalHooksConfig(t *testing.T) {
	hooks := generateMinimalHooksConfig()

	if len(hooks.SessionStart) == 0 {
		t.Error("expected SessionStart hooks, got none")
	}
	if len(hooks.Stop) == 0 {
		t.Error("expected Stop hooks, got none")
	}

	// Verify SessionStart contains ao inject
	found := false
	for _, g := range hooks.SessionStart {
		for _, h := range g.Hooks {
			if h.Type == "command" && h.Command == "ao inject --apply-decay --max-tokens 1500 2>/dev/null || true" {
				found = true
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
			}
		}
	}
	if !found {
		t.Error("expected ao forge command in Stop hooks")
	}
}

func TestAllEventNames(t *testing.T) {
	events := AllEventNames()
	if len(events) != 8 {
		t.Fatalf("expected 8 events, got %d", len(events))
	}
	expected := []string{
		"SessionStart", "SessionEnd",
		"PreToolUse", "PostToolUse",
		"UserPromptSubmit", "TaskCompleted",
		"Stop", "PreCompact",
	}
	for i, e := range expected {
		if events[i] != e {
			t.Errorf("event %d: expected %s, got %s", i, e, events[i])
		}
	}
}

func TestHooksConfigGetSetEventGroups(t *testing.T) {
	config := &HooksConfig{}
	groups := []HookGroup{
		{Hooks: []HookEntry{{Type: "command", Command: "test"}}},
	}

	for _, event := range AllEventNames() {
		config.SetEventGroups(event, groups)
		got := config.GetEventGroups(event)
		if len(got) != 1 {
			t.Errorf("event %s: expected 1 group after set, got %d", event, len(got))
		}
	}

	// Unknown event returns nil
	if got := config.GetEventGroups("Unknown"); got != nil {
		t.Error("expected nil for unknown event")
	}
}

func TestHookGroupToMapStringMatcher(t *testing.T) {
	g := HookGroup{
		Matcher: "Write|Edit",
		Hooks: []HookEntry{
			{Type: "command", Command: "echo hello"},
		},
	}

	m := hookGroupToMap(g)

	// Matcher should be a string
	matcher, ok := m["matcher"].(string)
	if !ok {
		t.Fatal("expected matcher to be a string")
	}
	if matcher != "Write|Edit" {
		t.Errorf("expected matcher 'Write|Edit', got '%s'", matcher)
	}

	hooks, ok := m["hooks"].([]map[string]interface{})
	if !ok {
		t.Fatal("expected hooks array in map")
	}
	if len(hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hooks))
	}
}

func TestHookGroupToMapEmptyMatcher(t *testing.T) {
	g := HookGroup{
		Matcher: "",
		Hooks: []HookEntry{
			{Type: "command", Command: "echo hello"},
		},
	}

	m := hookGroupToMap(g)
	if _, exists := m["matcher"]; exists {
		t.Error("expected no matcher key when Matcher is empty string")
	}
}

func TestHookGroupToMapTimeout(t *testing.T) {
	g := HookGroup{
		Hooks: []HookEntry{
			{Type: "command", Command: "test", Timeout: 120},
		},
	}

	m := hookGroupToMap(g)
	hooks := m["hooks"].([]map[string]interface{})
	if hooks[0]["timeout"] != 120 {
		t.Errorf("expected timeout 120, got %v", hooks[0]["timeout"])
	}

	// Zero timeout should be omitted
	g2 := HookGroup{
		Hooks: []HookEntry{
			{Type: "command", Command: "test", Timeout: 0},
		},
	}
	m2 := hookGroupToMap(g2)
	hooks2 := m2["hooks"].([]map[string]interface{})
	if _, exists := hooks2[0]["timeout"]; exists {
		t.Error("expected no timeout key when Timeout is 0")
	}
}

func TestReadHooksManifest(t *testing.T) {
	manifest := `{
		"$schema": "test",
		"hooks": {
			"SessionStart": [{"hooks": [{"type": "command", "command": "test-start"}]}],
			"SessionEnd": [{"hooks": [{"type": "command", "command": "test-end"}]}],
			"PreToolUse": [{"matcher": "Write|Edit", "hooks": [{"type": "command", "command": "test-pre", "timeout": 2}]}],
			"PostToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "test-post"}]}],
			"UserPromptSubmit": [{"hooks": [{"type": "command", "command": "test-prompt"}]}],
			"TaskCompleted": [{"hooks": [{"type": "command", "command": "test-task", "timeout": 120}]}],
			"Stop": [{"hooks": [{"type": "command", "command": "test-stop"}]}],
			"PreCompact": [{"hooks": [{"type": "command", "command": "test-compact"}]}]
		}
	}`

	config, err := ReadHooksManifest([]byte(manifest))
	if err != nil {
		t.Fatalf("ReadHooksManifest failed: %v", err)
	}

	// Verify all 8 events parsed
	for _, event := range AllEventNames() {
		groups := config.GetEventGroups(event)
		if len(groups) == 0 {
			t.Errorf("event %s: expected at least 1 group, got 0", event)
		}
	}

	// Verify PreToolUse has string matcher
	if len(config.PreToolUse) > 0 && config.PreToolUse[0].Matcher != "Write|Edit" {
		t.Errorf("PreToolUse matcher: expected 'Write|Edit', got '%s'", config.PreToolUse[0].Matcher)
	}

	// Verify timeout preserved
	if len(config.TaskCompleted) > 0 && len(config.TaskCompleted[0].Hooks) > 0 {
		if config.TaskCompleted[0].Hooks[0].Timeout != 120 {
			t.Errorf("TaskCompleted timeout: expected 120, got %d", config.TaskCompleted[0].Hooks[0].Timeout)
		}
	}

	// Verify PreToolUse hook timeout
	if len(config.PreToolUse) > 0 && len(config.PreToolUse[0].Hooks) > 0 {
		if config.PreToolUse[0].Hooks[0].Timeout != 2 {
			t.Errorf("PreToolUse timeout: expected 2, got %d", config.PreToolUse[0].Hooks[0].Timeout)
		}
	}
}

func TestReadHooksManifestInvalid(t *testing.T) {
	// Missing hooks key
	_, err := ReadHooksManifest([]byte(`{"other": "data"}`))
	if err == nil {
		t.Error("expected error for missing hooks key")
	}

	// Invalid JSON
	_, err = ReadHooksManifest([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestReplacePluginRoot(t *testing.T) {
	config := &HooksConfig{
		PreToolUse: []HookGroup{
			{
				Matcher: "Write|Edit",
				Hooks: []HookEntry{
					{Type: "command", Command: "${CLAUDE_PLUGIN_ROOT}/hooks/standards-injector.sh"},
				},
			},
		},
		Stop: []HookGroup{
			{
				Hooks: []HookEntry{
					{Type: "command", Command: "${CLAUDE_PLUGIN_ROOT}/hooks/stop-team-guard.sh"},
				},
			},
		},
	}

	replacePluginRoot(config, "/home/user/.agentops")

	if config.PreToolUse[0].Hooks[0].Command != "/home/user/.agentops/hooks/standards-injector.sh" {
		t.Errorf("PreToolUse command not rewritten: %s", config.PreToolUse[0].Hooks[0].Command)
	}
	if config.Stop[0].Hooks[0].Command != "/home/user/.agentops/hooks/stop-team-guard.sh" {
		t.Errorf("Stop command not rewritten: %s", config.Stop[0].Hooks[0].Command)
	}
}

func TestFilterNonAoHookGroupsAllEvents(t *testing.T) {
	// Build a hooksMap with ao and non-ao groups for every event
	hooksMap := make(map[string]interface{})
	for _, event := range AllEventNames() {
		hooksMap[event] = []interface{}{
			map[string]interface{}{
				"hooks": []interface{}{
					map[string]interface{}{"type": "command", "command": "ao inject 2>/dev/null"},
				},
			},
			map[string]interface{}{
				"hooks": []interface{}{
					map[string]interface{}{"type": "command", "command": "my-custom-hook"},
				},
			},
		}
	}

	for _, event := range AllEventNames() {
		filtered := filterNonAoHookGroups(hooksMap, event)
		if len(filtered) != 1 {
			t.Errorf("event %s: expected 1 non-ao group, got %d", event, len(filtered))
		}
		if hooks, ok := filtered[0]["hooks"].([]interface{}); ok {
			if hook, ok := hooks[0].(map[string]interface{}); ok {
				if hook["command"] != "my-custom-hook" {
					t.Errorf("event %s: expected non-ao hook preserved, got %v", event, hook["command"])
				}
			}
		}
	}
}

func TestHookGroupContainsAoAllEvents(t *testing.T) {
	hooksMap := make(map[string]interface{})
	for _, event := range AllEventNames() {
		hooksMap[event] = []interface{}{
			map[string]interface{}{
				"hooks": []interface{}{
					map[string]interface{}{"type": "command", "command": "ao inject stuff"},
				},
			},
		}
	}

	for _, event := range AllEventNames() {
		if !hookGroupContainsAo(hooksMap, event) {
			t.Errorf("event %s: expected ao hook detected", event)
		}
	}
}

func TestBackwardsCompatDefaultInstall(t *testing.T) {
	// generateMinimalHooksConfig should ALWAYS return SessionStart + Stop
	hooks := generateMinimalHooksConfig()
	if len(hooks.SessionStart) == 0 {
		t.Error("minimal config missing SessionStart")
	}
	if len(hooks.Stop) == 0 {
		t.Error("minimal config missing Stop")
	}
	// Should NOT have other events
	if len(hooks.PreToolUse) > 0 {
		t.Error("minimal config should not have PreToolUse")
	}
	if len(hooks.TaskCompleted) > 0 {
		t.Error("minimal config should not have TaskCompleted")
	}
}

func TestMatcherJSONRoundTrip(t *testing.T) {
	original := HookGroup{
		Matcher: "Write|Edit",
		Hooks: []HookEntry{
			{Type: "command", Command: "test", Timeout: 5},
		},
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Unmarshal
	var roundTripped HookGroup
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if roundTripped.Matcher != "Write|Edit" {
		t.Errorf("matcher lost in round-trip: got '%s'", roundTripped.Matcher)
	}
	if roundTripped.Hooks[0].Timeout != 5 {
		t.Errorf("timeout lost in round-trip: got %d", roundTripped.Hooks[0].Timeout)
	}
}
