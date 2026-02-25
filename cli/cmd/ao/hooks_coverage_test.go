package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// HooksConfig.eventGroupPtrs / eventGroupPtr
// ---------------------------------------------------------------------------

func TestHooksCoverage_eventGroupPtrs_ReturnsAllEvents(t *testing.T) {
	config := &HooksConfig{}
	ptrs := config.eventGroupPtrs()
	if len(ptrs) != 12 {
		t.Errorf("expected 12 event pointers, got %d", len(ptrs))
	}
	for _, event := range AllEventNames() {
		if _, ok := ptrs[event]; !ok {
			t.Errorf("missing event pointer for %s", event)
		}
	}
}

func TestHooksCoverage_eventGroupPtr_Unknown(t *testing.T) {
	config := &HooksConfig{}
	ptr := config.eventGroupPtr("NonexistentEvent")
	if ptr != nil {
		t.Error("expected nil for unknown event")
	}
}

func TestHooksCoverage_eventGroupPtr_Known(t *testing.T) {
	config := &HooksConfig{}
	for _, event := range AllEventNames() {
		ptr := config.eventGroupPtr(event)
		if ptr == nil {
			t.Errorf("expected non-nil pointer for event %s", event)
		}
	}
}

// ---------------------------------------------------------------------------
// HooksConfig.GetEventGroups / SetEventGroups
// ---------------------------------------------------------------------------

func TestHooksCoverage_SetEventGroups_UnknownEvent(t *testing.T) {
	config := &HooksConfig{}
	// Should be a no-op, no panic
	config.SetEventGroups("BogusEvent", []HookGroup{
		{Hooks: []HookEntry{{Type: "command", Command: "test"}}},
	})
	// Verify nothing was set
	if got := config.GetEventGroups("BogusEvent"); got != nil {
		t.Errorf("expected nil for unknown event, got %v", got)
	}
}

func TestHooksCoverage_GetEventGroups_Empty(t *testing.T) {
	config := &HooksConfig{}
	for _, event := range AllEventNames() {
		got := config.GetEventGroups(event)
		if len(got) != 0 {
			t.Errorf("expected empty groups for %s on fresh config, got %d", event, len(got))
		}
	}
}

// ---------------------------------------------------------------------------
// generateMinimalHooksConfig
// ---------------------------------------------------------------------------

func TestHooksCoverage_generateMinimalHooksConfig_PluginRootPlaceholder(t *testing.T) {
	config := generateMinimalHooksConfig()
	// All commands should contain ${CLAUDE_PLUGIN_ROOT} before replacement
	for _, g := range config.SessionStart {
		for _, h := range g.Hooks {
			if !strings.Contains(h.Command, "${CLAUDE_PLUGIN_ROOT}") {
				t.Errorf("SessionStart command missing placeholder: %s", h.Command)
			}
		}
	}
	for _, g := range config.SessionEnd {
		for _, h := range g.Hooks {
			if !strings.Contains(h.Command, "${CLAUDE_PLUGIN_ROOT}") {
				t.Errorf("SessionEnd command missing placeholder: %s", h.Command)
			}
		}
	}
	for _, g := range config.Stop {
		for _, h := range g.Hooks {
			if !strings.Contains(h.Command, "${CLAUDE_PLUGIN_ROOT}") {
				t.Errorf("Stop command missing placeholder: %s", h.Command)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// generateHooksConfig (falls back to minimal when full fails)
// ---------------------------------------------------------------------------

func TestHooksCoverage_generateHooksConfig_ReturnsNonNil(t *testing.T) {
	// generateHooksConfig always returns a non-nil config (falls back to minimal)
	config := generateHooksConfig()
	if config == nil {
		t.Fatal("expected non-nil config from generateHooksConfig")
	}
	// Should have at least the core events
	if len(config.SessionStart) == 0 {
		t.Error("expected SessionStart hooks")
	}
	if len(config.Stop) == 0 {
		t.Error("expected Stop hooks")
	}
}

// ---------------------------------------------------------------------------
// replacePluginRoot
// ---------------------------------------------------------------------------

func TestHooksCoverage_replacePluginRoot_EmptyBasePath(t *testing.T) {
	config := generateMinimalHooksConfig()
	replacePluginRoot(config, "")
	// Placeholder should be removed (replaced with empty string)
	for _, g := range config.SessionStart {
		for _, h := range g.Hooks {
			if strings.Contains(h.Command, "${CLAUDE_PLUGIN_ROOT}") {
				t.Errorf("placeholder should be removed, got: %s", h.Command)
			}
		}
	}
}

func TestHooksCoverage_replacePluginRoot_AllEvents(t *testing.T) {
	config := &HooksConfig{}
	for _, event := range AllEventNames() {
		config.SetEventGroups(event, []HookGroup{
			{Hooks: []HookEntry{{Type: "command", Command: "${CLAUDE_PLUGIN_ROOT}/hooks/test.sh"}}},
		})
	}
	replacePluginRoot(config, "/home/user/.agentops")
	for _, event := range AllEventNames() {
		groups := config.GetEventGroups(event)
		for _, g := range groups {
			for _, h := range g.Hooks {
				if strings.Contains(h.Command, "${CLAUDE_PLUGIN_ROOT}") {
					t.Errorf("event %s: placeholder not replaced in %s", event, h.Command)
				}
				if !strings.Contains(h.Command, "/home/user/.agentops/hooks/test.sh") {
					t.Errorf("event %s: expected replaced path, got %s", event, h.Command)
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// ReadHooksManifest
// ---------------------------------------------------------------------------

func TestHooksCoverage_ReadHooksManifest_EmptyHooksObject(t *testing.T) {
	data := []byte(`{"hooks": {}}`)
	config, err := ReadHooksManifest(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return config with no events populated
	for _, event := range AllEventNames() {
		if groups := config.GetEventGroups(event); len(groups) > 0 {
			t.Errorf("expected empty groups for %s, got %d", event, len(groups))
		}
	}
}

// ---------------------------------------------------------------------------
// hookGroupToMap
// ---------------------------------------------------------------------------

func TestHooksCoverage_hookGroupToMap_EmptyHooks(t *testing.T) {
	g := HookGroup{Hooks: []HookEntry{}}
	m := hookGroupToMap(g)
	hooks, ok := m["hooks"].([]map[string]any)
	if !ok {
		t.Fatal("expected hooks key in map")
	}
	if len(hooks) != 0 {
		t.Errorf("expected 0 hooks, got %d", len(hooks))
	}
	if _, exists := m["matcher"]; exists {
		t.Error("expected no matcher key for empty Matcher")
	}
}

func TestHooksCoverage_hookGroupToMap_ZeroTimeout(t *testing.T) {
	g := HookGroup{
		Hooks: []HookEntry{
			{Type: "command", Command: "test", Timeout: 0},
		},
	}
	m := hookGroupToMap(g)
	hooks := m["hooks"].([]map[string]any)
	if _, exists := hooks[0]["timeout"]; exists {
		t.Error("expected no timeout key when Timeout is 0")
	}
}

// ---------------------------------------------------------------------------
// loadHooksSettings
// ---------------------------------------------------------------------------

func TestHooksCoverage_loadHooksSettings_FileNotExist(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nonexistent.json")
	result, err := loadHooksSettings(path)
	if err != nil {
		t.Fatalf("expected nil error for nonexistent file, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestHooksCoverage_loadHooksSettings_ValidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	content := `{"hooks": {"SessionStart": []}, "other": "value"}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := loadHooksSettings(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result["hooks"]; !ok {
		t.Error("expected hooks key in result")
	}
	if result["other"] != "value" {
		t.Error("expected other key preserved")
	}
}

func TestHooksCoverage_loadHooksSettings_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	if err := os.WriteFile(path, []byte("not json{"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := loadHooksSettings(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ---------------------------------------------------------------------------
// writeHooksSettings
// ---------------------------------------------------------------------------

func TestHooksCoverage_writeHooksSettings_CreatesDir(t *testing.T) {
	tmp := t.TempDir()
	settingsPath := filepath.Join(tmp, "subdir", "settings.json")
	rawSettings := map[string]any{"hooks": map[string]any{}}
	if err := writeHooksSettings(settingsPath, rawSettings); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify file exists and is valid JSON
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("written file is not valid JSON: %v", err)
	}
}

// ---------------------------------------------------------------------------
// backupHooksSettings
// ---------------------------------------------------------------------------

func TestHooksCoverage_backupHooksSettings_NoFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nonexistent.json")
	// Should be a no-op, no error
	if err := backupHooksSettings(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHooksCoverage_backupHooksSettings_CreatesBackup(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	content := `{"hooks":{}}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := backupHooksSettings(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify backup file exists
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "settings.json.backup.") {
			found = true
			// Verify backup content matches original
			backupData, err := os.ReadFile(filepath.Join(tmp, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if string(backupData) != content {
				t.Error("backup content doesn't match original")
			}
		}
	}
	if !found {
		t.Error("expected backup file to be created")
	}
}

// ---------------------------------------------------------------------------
// cloneHooksMap
// ---------------------------------------------------------------------------

func TestHooksCoverage_cloneHooksMap_EmptySettings(t *testing.T) {
	rawSettings := map[string]any{}
	result := cloneHooksMap(rawSettings)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestHooksCoverage_cloneHooksMap_WithExistingHooks(t *testing.T) {
	rawSettings := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{},
			"custom":       "value",
		},
	}
	result := cloneHooksMap(rawSettings)
	if len(result) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result))
	}
	if result["custom"] != "value" {
		t.Error("expected custom key preserved")
	}
}

func TestHooksCoverage_cloneHooksMap_NonMapHooks(t *testing.T) {
	rawSettings := map[string]any{
		"hooks": "not-a-map",
	}
	result := cloneHooksMap(rawSettings)
	if len(result) != 0 {
		t.Errorf("expected empty map for non-map hooks, got %v", result)
	}
}

// ---------------------------------------------------------------------------
// existingAoHooksBlock
// ---------------------------------------------------------------------------

func TestHooksCoverage_existingAoHooksBlock_NoHooks(t *testing.T) {
	oldForce := hooksForce
	defer func() { hooksForce = oldForce }()
	hooksForce = false

	rawSettings := map[string]any{}
	if existingAoHooksBlock(rawSettings) {
		t.Error("expected false when no hooks present")
	}
}

func TestHooksCoverage_existingAoHooksBlock_WithForce(t *testing.T) {
	oldForce := hooksForce
	defer func() { hooksForce = oldForce }()
	hooksForce = true

	rawSettings := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"command": "ao inject"},
					},
				},
			},
		},
	}
	if existingAoHooksBlock(rawSettings) {
		t.Error("expected false when --force is set")
	}
}

func TestHooksCoverage_existingAoHooksBlock_WithAoHooks(t *testing.T) {
	oldForce := hooksForce
	defer func() { hooksForce = oldForce }()
	hooksForce = false

	rawSettings := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"command": "ao inject"},
					},
				},
			},
		},
	}
	if !existingAoHooksBlock(rawSettings) {
		t.Error("expected true when ao hooks are present and force is false")
	}
}

func TestHooksCoverage_existingAoHooksBlock_NonAoHooks(t *testing.T) {
	oldForce := hooksForce
	defer func() { hooksForce = oldForce }()
	hooksForce = false

	rawSettings := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"command": "echo hello"},
					},
				},
			},
		},
	}
	if existingAoHooksBlock(rawSettings) {
		t.Error("expected false when no ao hooks are present")
	}
}

// ---------------------------------------------------------------------------
// dryRunPrintSettings
// ---------------------------------------------------------------------------

func TestHooksCoverage_dryRunPrintSettings_NotDryRun(t *testing.T) {
	oldDryRun := hooksDryRun
	defer func() { hooksDryRun = oldDryRun }()
	hooksDryRun = false

	done, err := dryRunPrintSettings("/fake/path", map[string]any{})
	if done {
		t.Error("expected done=false when not in dry-run mode")
	}
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestHooksCoverage_dryRunPrintSettings_DryRun(t *testing.T) {
	oldDryRun := hooksDryRun
	defer func() { hooksDryRun = oldDryRun }()
	hooksDryRun = true

	rawSettings := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{},
		},
	}
	done, err := dryRunPrintSettings("/fake/path", rawSettings)
	if !done {
		t.Error("expected done=true when in dry-run mode")
	}
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// generateHooksForInstall
// ---------------------------------------------------------------------------

func TestHooksCoverage_generateHooksForInstall_Minimal(t *testing.T) {
	oldFull := hooksFull
	defer func() { hooksFull = oldFull }()
	hooksFull = false

	config, events, err := generateHooksForInstall("/home/user/.agentops")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events for minimal install, got %d", len(events))
	}
	// Verify placeholder was replaced
	for _, g := range config.SessionStart {
		for _, h := range g.Hooks {
			if strings.Contains(h.Command, "${CLAUDE_PLUGIN_ROOT}") {
				t.Error("placeholder not replaced in minimal config")
			}
			if !strings.Contains(h.Command, "/home/user/.agentops") {
				t.Errorf("expected install base in command, got %s", h.Command)
			}
		}
	}
}

func TestHooksCoverage_generateHooksForInstall_Full(t *testing.T) {
	oldFull := hooksFull
	defer func() { hooksFull = oldFull }()
	hooksFull = true

	// This should succeed because embedded hooks are always available
	config, events, err := generateHooksForInstall("/home/user/.agentops")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if len(events) != 12 {
		t.Errorf("expected 12 events for full install, got %d", len(events))
	}
}

// ---------------------------------------------------------------------------
// commitHooksSettings
// ---------------------------------------------------------------------------

func TestHooksCoverage_commitHooksSettings(t *testing.T) {
	tmp := t.TempDir()
	settingsPath := filepath.Join(tmp, "settings.json")

	rawSettings := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{},
		},
	}
	newHooks := generateMinimalHooksConfig()
	replacePluginRoot(newHooks, "/test")
	eventsToInstall := []string{"SessionStart", "SessionEnd", "Stop"}

	err := commitHooksSettings(settingsPath, rawSettings, newHooks, eventsToInstall, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("settings file not created: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("written file is not valid JSON: %v", err)
	}
}

// ---------------------------------------------------------------------------
// loadHooksMap
// ---------------------------------------------------------------------------

func TestHooksCoverage_loadHooksMap_FileNotExist(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nonexistent.json")
	result, err := loadHooksMap(path)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result for nonexistent file")
	}
}

func TestHooksCoverage_loadHooksMap_NoHooksKey(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	if err := os.WriteFile(path, []byte(`{"other": "value"}`), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := loadHooksMap(path)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result when no hooks key")
	}
}

func TestHooksCoverage_loadHooksMap_ValidHooks(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	content := `{"hooks": {"SessionStart": []}}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := loadHooksMap(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil hooks map")
	}
}

func TestHooksCoverage_loadHooksMap_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	if err := os.WriteFile(path, []byte("not-json{"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := loadHooksMap(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHooksCoverage_loadHooksMap_HooksNotMap(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	if err := os.WriteFile(path, []byte(`{"hooks": "not-a-map"}`), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := loadHooksMap(path)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result when hooks is not a map")
	}
}

// ---------------------------------------------------------------------------
// countRawGroupHooks
// ---------------------------------------------------------------------------

func TestHooksCoverage_countRawGroupHooks(t *testing.T) {
	tests := []struct {
		name   string
		groups []any
		want   int
	}{
		{
			name:   "empty groups",
			groups: []any{},
			want:   0,
		},
		{
			name: "single group single hook",
			groups: []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"command": "test"},
					},
				},
			},
			want: 1,
		},
		{
			name: "multiple groups multiple hooks",
			groups: []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"command": "test1"},
						map[string]any{"command": "test2"},
					},
				},
				map[string]any{
					"hooks": []any{
						map[string]any{"command": "test3"},
					},
				},
			},
			want: 3,
		},
		{
			name: "non-map group skipped",
			groups: []any{
				"not-a-map",
				map[string]any{
					"hooks": []any{
						map[string]any{"command": "test"},
					},
				},
			},
			want: 1,
		},
		{
			name: "group without hooks key",
			groups: []any{
				map[string]any{
					"matcher": "test",
				},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countRawGroupHooks(tt.groups)
			if got != tt.want {
				t.Errorf("countRawGroupHooks() = %d, want %d", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// printEventCoverage
// ---------------------------------------------------------------------------

func TestHooksCoverage_printEventCoverage_AllInstalled(t *testing.T) {
	hooksMap := make(map[string]any)
	for _, event := range AllEventNames() {
		hooksMap[event] = []any{
			map[string]any{
				"hooks": []any{
					map[string]any{"command": "test"},
				},
			},
		}
	}
	count := printEventCoverage(hooksMap)
	if count != 12 {
		t.Errorf("expected 12 installed events, got %d", count)
	}
}

func TestHooksCoverage_printEventCoverage_NoneInstalled(t *testing.T) {
	hooksMap := make(map[string]any)
	count := printEventCoverage(hooksMap)
	if count != 0 {
		t.Errorf("expected 0 installed events, got %d", count)
	}
}

func TestHooksCoverage_printEventCoverage_Partial(t *testing.T) {
	hooksMap := map[string]any{
		"SessionStart": []any{
			map[string]any{
				"hooks": []any{map[string]any{"command": "test"}},
			},
		},
		"Stop": []any{
			map[string]any{
				"hooks": []any{map[string]any{"command": "test"}},
			},
		},
	}
	count := printEventCoverage(hooksMap)
	if count != 2 {
		t.Errorf("expected 2 installed events, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// countInstalledHookEvents
// ---------------------------------------------------------------------------

func TestHooksCoverage_countInstalledHookEvents(t *testing.T) {
	tests := []struct {
		name     string
		hooksMap map[string]any
		want     int
	}{
		{
			name:     "empty map",
			hooksMap: map[string]any{},
			want:     0,
		},
		{
			name: "one event with groups",
			hooksMap: map[string]any{
				"SessionStart": []any{map[string]any{}},
			},
			want: 1,
		},
		{
			name: "event with empty array",
			hooksMap: map[string]any{
				"SessionStart": []any{},
			},
			want: 0,
		},
		{
			name: "non-standard event ignored",
			hooksMap: map[string]any{
				"CustomEvent": []any{map[string]any{}},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countInstalledHookEvents(tt.hooksMap)
			if got != tt.want {
				t.Errorf("countInstalledHookEvents() = %d, want %d", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// readSettingsHooksMap
// ---------------------------------------------------------------------------

func TestHooksCoverage_readSettingsHooksMap_FileNotExist(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nonexistent.json")
	result, err := readSettingsHooksMap(path)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result != nil {
		t.Error("expected nil result for nonexistent file")
	}
}

func TestHooksCoverage_readSettingsHooksMap_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	if err := os.WriteFile(path, []byte("invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := readSettingsHooksMap(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHooksCoverage_readSettingsHooksMap_NoHooksKey(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	if err := os.WriteFile(path, []byte(`{"other": "value"}`), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := readSettingsHooksMap(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result when no hooks key")
	}
}

func TestHooksCoverage_readSettingsHooksMap_HooksNotMap(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	if err := os.WriteFile(path, []byte(`{"hooks": "string-value"}`), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := readSettingsHooksMap(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result when hooks is not a map")
	}
}

func TestHooksCoverage_readSettingsHooksMap_Valid(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")
	content := `{"hooks": {"SessionStart": [{"hooks": [{"command": "test"}]}]}}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := readSettingsHooksMap(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil hooks map")
	}
	if _, ok := result["SessionStart"]; !ok {
		t.Error("expected SessionStart key in hooks map")
	}
}

// ---------------------------------------------------------------------------
// rawGroupIsAoManaged / rawGroupHooksContainAo / rawGroupLegacyContainsAo
// ---------------------------------------------------------------------------

func TestHooksCoverage_rawGroupIsAoManaged(t *testing.T) {
	tests := []struct {
		name  string
		group map[string]any
		want  bool
	}{
		{
			name: "new format ao command",
			group: map[string]any{
				"hooks": []any{
					map[string]any{"command": "ao inject"},
				},
			},
			want: true,
		},
		{
			name: "new format agentops script",
			group: map[string]any{
				"hooks": []any{
					map[string]any{"command": "/home/user/.agentops/hooks/test.sh"},
				},
			},
			want: true,
		},
		{
			name: "legacy format ao command",
			group: map[string]any{
				"command": []any{"bash", "ao flywheel"},
			},
			want: true,
		},
		{
			name: "non-ao group",
			group: map[string]any{
				"hooks": []any{
					map[string]any{"command": "echo hello"},
				},
			},
			want: false,
		},
		{
			name:  "empty group",
			group: map[string]any{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rawGroupIsAoManaged(tt.group)
			if got != tt.want {
				t.Errorf("rawGroupIsAoManaged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHooksCoverage_rawGroupHooksContainAo(t *testing.T) {
	tests := []struct {
		name  string
		group map[string]any
		want  bool
	}{
		{
			name:  "no hooks key",
			group: map[string]any{},
			want:  false,
		},
		{
			name: "hooks not array",
			group: map[string]any{
				"hooks": "not-array",
			},
			want: false,
		},
		{
			name: "hook entries not maps",
			group: map[string]any{
				"hooks": []any{"not-a-map"},
			},
			want: false,
		},
		{
			name: "command not string",
			group: map[string]any{
				"hooks": []any{
					map[string]any{"command": 123},
				},
			},
			want: false,
		},
		{
			name: "ao command present",
			group: map[string]any{
				"hooks": []any{
					map[string]any{"command": "ao inject"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rawGroupHooksContainAo(tt.group)
			if got != tt.want {
				t.Errorf("rawGroupHooksContainAo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHooksCoverage_rawGroupLegacyContainsAo(t *testing.T) {
	tests := []struct {
		name  string
		group map[string]any
		want  bool
	}{
		{
			name:  "no command key",
			group: map[string]any{},
			want:  false,
		},
		{
			name: "command not array",
			group: map[string]any{
				"command": "string",
			},
			want: false,
		},
		{
			name: "command array too short",
			group: map[string]any{
				"command": []any{"bash"},
			},
			want: false,
		},
		{
			name: "second element not string",
			group: map[string]any{
				"command": []any{"bash", 123},
			},
			want: false,
		},
		{
			name: "legacy ao command",
			group: map[string]any{
				"command": []any{"bash", "ao flywheel close-loop"},
			},
			want: true,
		},
		{
			name: "legacy non-ao command",
			group: map[string]any{
				"command": []any{"bash", "echo hello"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rawGroupLegacyContainsAo(tt.group)
			if got != tt.want {
				t.Errorf("rawGroupLegacyContainsAo() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// filterNonAoHookGroups
// ---------------------------------------------------------------------------

func TestHooksCoverage_filterNonAoHookGroups_EventNotPresent(t *testing.T) {
	hooksMap := map[string]any{}
	result := filterNonAoHookGroups(hooksMap, "SessionStart")
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d", len(result))
	}
}

func TestHooksCoverage_filterNonAoHookGroups_NonMapEntries(t *testing.T) {
	hooksMap := map[string]any{
		"SessionStart": []any{
			"not-a-map",
			map[string]any{
				"hooks": []any{
					map[string]any{"command": "echo hello"},
				},
			},
		},
	}
	result := filterNonAoHookGroups(hooksMap, "SessionStart")
	if len(result) != 1 {
		t.Errorf("expected 1 group (non-map skipped), got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// hooksCopyFile
// ---------------------------------------------------------------------------

func TestHooksCoverage_hooksCopyFile(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "source.sh")
	dst := filepath.Join(tmp, "subdir", "dest.sh")

	content := "#!/bin/bash\necho hello"
	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := hooksCopyFile(src, dst); err != nil {
		t.Fatalf("hooksCopyFile failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("destination not created: %v", err)
	}
	if string(data) != content {
		t.Error("content mismatch")
	}
}

func TestHooksCoverage_hooksCopyFile_SourceNotExist(t *testing.T) {
	tmp := t.TempDir()
	err := hooksCopyFile(filepath.Join(tmp, "nonexistent"), filepath.Join(tmp, "dest"))
	if err == nil {
		t.Error("expected error for nonexistent source")
	}
}

// ---------------------------------------------------------------------------
// copyDir
// ---------------------------------------------------------------------------

func TestHooksCoverage_copyDir(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src")
	dstDir := filepath.Join(tmp, "dst")
	if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "sub", "b.txt"), []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}

	count, err := copyDir(srcDir, dstDir)
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 files copied, got %d", count)
	}

	// Verify files exist
	if _, err := os.Stat(filepath.Join(dstDir, "a.txt")); err != nil {
		t.Error("a.txt not copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "sub", "b.txt")); err != nil {
		t.Error("sub/b.txt not copied")
	}
}

// ---------------------------------------------------------------------------
// copyShellScripts
// ---------------------------------------------------------------------------

func TestHooksCoverage_copyShellScripts(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src")
	dstDir := filepath.Join(tmp, "dst")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create mixed files
	for _, name := range []string{"hook1.sh", "hook2.sh", "readme.md"} {
		if err := os.WriteFile(filepath.Join(srcDir, name), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Create a subdirectory (should be skipped)
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	count, err := copyShellScripts(srcDir, dstDir)
	if err != nil {
		t.Fatalf("copyShellScripts failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 scripts copied, got %d", count)
	}

	// Verify scripts are executable
	for _, name := range []string{"hook1.sh", "hook2.sh"} {
		info, err := os.Stat(filepath.Join(dstDir, name))
		if err != nil {
			t.Fatalf("script %s not copied: %v", name, err)
		}
		if info.Mode()&0111 == 0 {
			t.Errorf("script %s should be executable", name)
		}
	}

	// readme.md should not be copied
	if _, err := os.Stat(filepath.Join(dstDir, "readme.md")); !os.IsNotExist(err) {
		t.Error("readme.md should not be copied")
	}
}

func TestHooksCoverage_copyShellScripts_SrcNotExist(t *testing.T) {
	_, err := copyShellScripts("/nonexistent/path", "/also/nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent source directory")
	}
}

// ---------------------------------------------------------------------------
// copyOptionalFile
// ---------------------------------------------------------------------------

func TestHooksCoverage_copyOptionalFile_Missing(t *testing.T) {
	tmp := t.TempDir()
	count, err := copyOptionalFile(filepath.Join(tmp, "nofile"), filepath.Join(tmp, "dst"), "test")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 copied, got %d", count)
	}
}

func TestHooksCoverage_copyOptionalFile_Present(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "source.txt")
	dst := filepath.Join(tmp, "subdir", "dest.txt")
	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	count, err := copyOptionalFile(src, dst, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 copied, got %d", count)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal("destination not created")
	}
	if string(data) != "content" {
		t.Error("content mismatch")
	}
}

// ---------------------------------------------------------------------------
// copyOptionalDir
// ---------------------------------------------------------------------------

func TestHooksCoverage_copyOptionalDir_Missing(t *testing.T) {
	tmp := t.TempDir()
	count, err := copyOptionalDir(filepath.Join(tmp, "nodir"), filepath.Join(tmp, "dst"), "test")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 copied, got %d", count)
	}
}

func TestHooksCoverage_copyOptionalDir_Present(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src")
	dstDir := filepath.Join(tmp, "dst")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	count, err := copyOptionalDir(srcDir, dstDir, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 file copied, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// installFullHooks
// ---------------------------------------------------------------------------

func TestHooksCoverage_installFullHooks_NoGitDir(t *testing.T) {
	tmp := t.TempDir()
	// Source dir without .git - should fail
	srcDir := filepath.Join(tmp, "src")
	if err := os.MkdirAll(filepath.Join(srcDir, "hooks"), 0755); err != nil {
		t.Fatal(err)
	}
	installBase := filepath.Join(tmp, "install")

	_, err := installFullHooks(srcDir, installBase)
	if err == nil {
		t.Error("expected error when .git directory is missing")
	}
	if !strings.Contains(err.Error(), "not a git root") {
		t.Errorf("expected 'not a git root' error, got: %v", err)
	}
}

func TestHooksCoverage_installFullHooks_WithGitDir(t *testing.T) {
	tmp := t.TempDir()
	srcDir := filepath.Join(tmp, "src")
	installBase := filepath.Join(tmp, "install")

	// Create source dir with .git and hooks
	if err := os.MkdirAll(filepath.Join(srcDir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	hooksDir := filepath.Join(srcDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "test.sh"), []byte("#!/bin/bash"), 0644); err != nil {
		t.Fatal(err)
	}

	count, err := installFullHooks(srcDir, installBase)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count < 1 {
		t.Errorf("expected at least 1 file copied, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// installFullHooksFromEmbed
// ---------------------------------------------------------------------------

func TestHooksCoverage_installFullHooksFromEmbed(t *testing.T) {
	tmp := t.TempDir()
	count, err := installFullHooksFromEmbed(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count == 0 {
		t.Error("expected at least 1 file extracted")
	}

	// Verify hooks.json was extracted
	hooksJSON := filepath.Join(tmp, "hooks", "hooks.json")
	if _, err := os.Stat(hooksJSON); err != nil {
		t.Errorf("hooks.json not extracted: %v", err)
	}
}

// ---------------------------------------------------------------------------
// printHooksInstallSummary (just verify no panic)
// ---------------------------------------------------------------------------

func TestHooksCoverage_printHooksInstallSummary_Minimal(t *testing.T) {
	oldFull := hooksFull
	defer func() { hooksFull = oldFull }()
	hooksFull = false

	newHooks := generateMinimalHooksConfig()
	// Should not panic
	printHooksInstallSummary("/fake/settings.json", newHooks, []string{"SessionStart", "SessionEnd", "Stop"}, 3)
}

func TestHooksCoverage_printHooksInstallSummary_Full(t *testing.T) {
	oldFull := hooksFull
	defer func() { hooksFull = oldFull }()
	hooksFull = true

	newHooks := generateMinimalHooksConfig()
	// Should not panic
	printHooksInstallSummary("/fake/settings.json", newHooks, AllEventNames(), 3)
}

// ---------------------------------------------------------------------------
// runSettingsCoverageTest
// ---------------------------------------------------------------------------

func TestHooksCoverage_runSettingsCoverageTest_NoSettings(t *testing.T) {
	tmp := t.TempDir()
	allPassed := true
	// Should print warning, not crash
	runSettingsCoverageTest(1, tmp, &allPassed)
}

func TestHooksCoverage_runSettingsCoverageTest_WithSettings(t *testing.T) {
	tmp := t.TempDir()
	claudeDir := filepath.Join(tmp, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `{"hooks": {"SessionStart": [{"hooks": [{"command": "test"}]}]}}`
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	allPassed := true
	runSettingsCoverageTest(1, tmp, &allPassed)
	// Should still pass
	if !allPassed {
		t.Error("expected allPassed to remain true")
	}
}

func TestHooksCoverage_runSettingsCoverageTest_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	claudeDir := filepath.Join(tmp, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte("invalid"), 0644); err != nil {
		t.Fatal(err)
	}
	allPassed := true
	runSettingsCoverageTest(1, tmp, &allPassed)
	if allPassed {
		t.Error("expected allPassed to be false for invalid JSON")
	}
}

// ---------------------------------------------------------------------------
// runHookScriptsAccessTest
// ---------------------------------------------------------------------------

func TestHooksCoverage_runHookScriptsAccessTest_NoAgentopsDir(t *testing.T) {
	tmp := t.TempDir()
	// Should not panic, prints "not installed"
	runHookScriptsAccessTest(1, tmp)
}

func TestHooksCoverage_runHookScriptsAccessTest_WithScripts(t *testing.T) {
	tmp := t.TempDir()
	hooksDir := filepath.Join(tmp, ".agentops", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create executable scripts
	for _, name := range []string{"test1.sh", "test2.sh"} {
		if err := os.WriteFile(filepath.Join(hooksDir, name), []byte("#!/bin/bash"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	// Should not panic
	runHookScriptsAccessTest(1, tmp)
}

func TestHooksCoverage_runHookScriptsAccessTest_NonExecutableScripts(t *testing.T) {
	tmp := t.TempDir()
	hooksDir := filepath.Join(tmp, ".agentops", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create non-executable script
	if err := os.WriteFile(filepath.Join(hooksDir, "test.sh"), []byte("#!/bin/bash"), 0644); err != nil {
		t.Fatal(err)
	}
	// Should not panic, prints warning
	runHookScriptsAccessTest(1, tmp)
}

func TestHooksCoverage_runHookScriptsAccessTest_WithSettingsAndScripts(t *testing.T) {
	tmp := t.TempDir()

	// Create hooks scripts
	hooksDir := filepath.Join(tmp, ".agentops", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"session-start.sh", "orphan.sh"} {
		if err := os.WriteFile(filepath.Join(hooksDir, name), []byte("#!/bin/bash"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create settings with wired hooks
	claudeDir := filepath.Join(tmp, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	settings := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{
							"type":    "command",
							"command": fmt.Sprintf("%s/session-start.sh", hooksDir),
						},
					},
				},
			},
		},
	}
	data, _ := json.MarshalIndent(settings, "", "  ")
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	// Should not panic, should report scripts and unwired count
	runHookScriptsAccessTest(1, tmp)
}

// ---------------------------------------------------------------------------
// runForgeTranscriptAccessTest
// ---------------------------------------------------------------------------

func TestHooksCoverage_runForgeTranscriptAccessTest_DryRun(t *testing.T) {
	oldDryRun := hooksDryRun
	defer func() { hooksDryRun = oldDryRun }()
	hooksDryRun = true

	// Should skip, not panic
	runForgeTranscriptAccessTest(1, t.TempDir())
}

func TestHooksCoverage_runForgeTranscriptAccessTest_NoProjectsDir(t *testing.T) {
	oldDryRun := hooksDryRun
	defer func() { hooksDryRun = oldDryRun }()
	hooksDryRun = false

	tmp := t.TempDir()
	// No .claude/projects directory
	runForgeTranscriptAccessTest(1, tmp)
}

func TestHooksCoverage_runForgeTranscriptAccessTest_WithProjectsDir(t *testing.T) {
	oldDryRun := hooksDryRun
	defer func() { hooksDryRun = oldDryRun }()
	hooksDryRun = false

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".claude", "projects"), 0755); err != nil {
		t.Fatal(err)
	}
	runForgeTranscriptAccessTest(1, tmp)
}

// ---------------------------------------------------------------------------
// runInjectCommandTest (dry-run only)
// ---------------------------------------------------------------------------

func TestHooksCoverage_runInjectCommandTest_DryRun(t *testing.T) {
	oldDryRun := hooksDryRun
	defer func() { hooksDryRun = oldDryRun }()
	hooksDryRun = true

	allPassed := true
	// Should skip
	runInjectCommandTest(1, &allPassed)
	if !allPassed {
		t.Error("expected allPassed to remain true in dry-run mode")
	}
}

// ---------------------------------------------------------------------------
// resolveSourceDir
// ---------------------------------------------------------------------------

func TestHooksCoverage_resolveSourceDir_ExplicitValid(t *testing.T) {
	tmp := t.TempDir()
	hooksDir := filepath.Join(tmp, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "hooks.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	oldSourceDir := hooksSourceDir
	defer func() { hooksSourceDir = oldSourceDir }()
	hooksSourceDir = tmp

	result, err := resolveSourceDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != tmp {
		t.Errorf("expected %s, got %s", tmp, result)
	}
}

func TestHooksCoverage_resolveSourceDir_ExplicitInvalid(t *testing.T) {
	tmp := t.TempDir()

	oldSourceDir := hooksSourceDir
	defer func() { hooksSourceDir = oldSourceDir }()
	hooksSourceDir = tmp

	_, err := resolveSourceDir()
	if err == nil {
		t.Error("expected error for invalid source dir")
	}
}

// ---------------------------------------------------------------------------
// collectScriptNames
// ---------------------------------------------------------------------------

func TestHooksCoverage_collectScriptNames_WithMixedFiles(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{"a.sh", "b.sh", "c.txt", "d.py"} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}
	names := collectScriptNames(tmp)
	if len(names) != 2 {
		t.Errorf("expected 2, got %d", len(names))
	}
	if !names["a.sh"] || !names["b.sh"] {
		t.Error("expected a.sh and b.sh")
	}
}

// ---------------------------------------------------------------------------
// collectWiredScripts
// ---------------------------------------------------------------------------

func TestHooksCoverage_collectWiredScripts_MalformedEntries(t *testing.T) {
	hooksMap := map[string]any{
		"SessionStart": []any{
			"not-a-map", // should be skipped
			map[string]any{
				"hooks": "not-an-array", // should be skipped
			},
			map[string]any{
				"hooks": []any{
					"not-a-map-hook", // should be skipped
					map[string]any{
						"command": 123, // not a string, should be skipped
					},
				},
			},
		},
	}
	eventScriptCount, wiredScripts := collectWiredScripts(hooksMap)
	if len(eventScriptCount) != 0 {
		t.Errorf("expected 0 events, got %d", len(eventScriptCount))
	}
	if len(wiredScripts) != 0 {
		t.Errorf("expected 0 wired scripts, got %d", len(wiredScripts))
	}
}

// ---------------------------------------------------------------------------
// isAoManagedHookCommand
// ---------------------------------------------------------------------------

func TestHooksCoverage_isAoManagedHookCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"ao inject --max-tokens 100", true},
		{"/home/user/.agentops/hooks/session-start.sh", true},
		{"echo hello", false},
		{"", false},
		{"/opt/tools/ao-runner.sh", false}, // "ao" not followed by space
		{"ao ", true},                       // "ao " is detected
	}
	for _, tt := range tests {
		got := isAoManagedHookCommand(tt.cmd)
		if got != tt.want {
			t.Errorf("isAoManagedHookCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// HookEntry / HookGroup JSON round-trip
// ---------------------------------------------------------------------------

func TestHooksCoverage_HookEntry_JSONRoundTrip(t *testing.T) {
	entry := HookEntry{
		Type:    "command",
		Command: "ao inject",
		Timeout: 30,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	var decoded HookEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Type != entry.Type || decoded.Command != entry.Command || decoded.Timeout != entry.Timeout {
		t.Error("round-trip mismatch")
	}
}

func TestHooksCoverage_HookEntry_ZeroTimeout_OmittedInJSON(t *testing.T) {
	entry := HookEntry{Type: "command", Command: "test", Timeout: 0}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("timeout")) {
		t.Error("expected timeout to be omitted when 0")
	}
}

func TestHooksCoverage_HookGroup_JSONRoundTrip(t *testing.T) {
	group := HookGroup{
		Matcher: "Write|Edit",
		Hooks: []HookEntry{
			{Type: "command", Command: "test1", Timeout: 5},
			{Type: "command", Command: "test2"},
		},
	}
	data, err := json.Marshal(group)
	if err != nil {
		t.Fatal(err)
	}
	var decoded HookGroup
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Matcher != group.Matcher {
		t.Errorf("matcher mismatch: got %q, want %q", decoded.Matcher, group.Matcher)
	}
	if len(decoded.Hooks) != 2 {
		t.Errorf("expected 2 hooks, got %d", len(decoded.Hooks))
	}
}

// ---------------------------------------------------------------------------
// HooksConfig JSON round-trip
// ---------------------------------------------------------------------------

func TestHooksCoverage_HooksConfig_JSONRoundTrip(t *testing.T) {
	config := &HooksConfig{}
	for _, event := range AllEventNames() {
		config.SetEventGroups(event, []HookGroup{
			{
				Matcher: "test-matcher",
				Hooks: []HookEntry{
					{Type: "command", Command: fmt.Sprintf("test-%s", event), Timeout: 10},
				},
			},
		})
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}

	var decoded HooksConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	for _, event := range AllEventNames() {
		groups := decoded.GetEventGroups(event)
		if len(groups) != 1 {
			t.Errorf("event %s: expected 1 group after round-trip, got %d", event, len(groups))
		}
	}
}

// ---------------------------------------------------------------------------
// hooksManifest type
// ---------------------------------------------------------------------------

func TestHooksCoverage_hooksManifest_Unmarshal(t *testing.T) {
	data := []byte(`{"hooks": {"SessionStart": [{"hooks": [{"type": "command", "command": "test"}]}]}}`)
	var m hooksManifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m.Hooks == nil {
		t.Fatal("expected non-nil Hooks")
	}
	if len(m.Hooks.SessionStart) != 1 {
		t.Errorf("expected 1 SessionStart group, got %d", len(m.Hooks.SessionStart))
	}
}

// ---------------------------------------------------------------------------
// mergeHookEvents edge cases
// ---------------------------------------------------------------------------

func TestHooksCoverage_mergeHookEvents_MultipleEvents(t *testing.T) {
	hooksMap := map[string]any{}
	config := &HooksConfig{}
	config.SetEventGroups("SessionStart", []HookGroup{
		{Hooks: []HookEntry{{Command: "ao start"}}},
	})
	config.SetEventGroups("Stop", []HookGroup{
		{Hooks: []HookEntry{{Command: "ao stop"}}},
	})
	config.SetEventGroups("PreToolUse", []HookGroup{
		{Matcher: "Write", Hooks: []HookEntry{{Command: "ao pre"}}},
	})

	count := mergeHookEvents(hooksMap, config, []string{"SessionStart", "Stop", "PreToolUse", "PostToolUse"})
	if count != 3 {
		t.Errorf("expected 3 installed events (PostToolUse empty), got %d", count)
	}
}

// ---------------------------------------------------------------------------
// runHooksInit output format validation
// ---------------------------------------------------------------------------

func TestHooksCoverage_runHooksInit_UnknownFormat(t *testing.T) {
	oldFormat := hooksOutputFormat
	defer func() { hooksOutputFormat = oldFormat }()
	hooksOutputFormat = "xml"

	err := runHooksInit(nil, nil)
	if err == nil {
		t.Error("expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("expected 'unknown format' error, got: %v", err)
	}
}
