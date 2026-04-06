// Package bridge provides pure helper functions extracted from cmd/ao bridge-related
// files (hooks, codex, factory). These are parsing, formatting, and validation helpers
// with no I/O, exec, or cobra dependencies.
package bridge

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// HookEntry represents a single hook command (e.g., {"type": "command", "command": "..."}).
type HookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

// HookGroup represents a hook group with optional matcher and a hooks array.
type HookGroup struct {
	Matcher string      `json:"matcher,omitempty"`
	Hooks   []HookEntry `json:"hooks"`
}

// HooksConfig represents the hooks section of Claude settings.
// Supports all 12 hook events (Claude Code runtime).
type HooksConfig struct {
	SessionStart     []HookGroup `json:"SessionStart,omitempty"`
	SessionEnd       []HookGroup `json:"SessionEnd,omitempty"`
	PreToolUse       []HookGroup `json:"PreToolUse,omitempty"`
	PostToolUse      []HookGroup `json:"PostToolUse,omitempty"`
	UserPromptSubmit []HookGroup `json:"UserPromptSubmit,omitempty"`
	TaskCompleted    []HookGroup `json:"TaskCompleted,omitempty"`
	Stop             []HookGroup `json:"Stop,omitempty"`
	PreCompact       []HookGroup `json:"PreCompact,omitempty"`
	SubagentStop     []HookGroup `json:"SubagentStop,omitempty"`
	WorktreeCreate   []HookGroup `json:"WorktreeCreate,omitempty"`
	WorktreeRemove   []HookGroup `json:"WorktreeRemove,omitempty"`
	ConfigChange     []HookGroup `json:"ConfigChange,omitempty"`
}

// eventGroupPtrs returns a map from event name to a pointer to the corresponding
// []HookGroup field. Used by GetEventGroups and SetEventGroups.
func (c *HooksConfig) eventGroupPtrs() map[string]*[]HookGroup {
	return map[string]*[]HookGroup{
		"SessionStart":     &c.SessionStart,
		"SessionEnd":       &c.SessionEnd,
		"PreToolUse":       &c.PreToolUse,
		"PostToolUse":      &c.PostToolUse,
		"UserPromptSubmit": &c.UserPromptSubmit,
		"TaskCompleted":    &c.TaskCompleted,
		"Stop":             &c.Stop,
		"PreCompact":       &c.PreCompact,
		"SubagentStop":     &c.SubagentStop,
		"WorktreeCreate":   &c.WorktreeCreate,
		"WorktreeRemove":   &c.WorktreeRemove,
		"ConfigChange":     &c.ConfigChange,
	}
}

// eventGroupPtr returns a pointer to the []HookGroup field for the given event name,
// or nil if the event is unknown.
func (c *HooksConfig) eventGroupPtr(event string) *[]HookGroup {
	return c.eventGroupPtrs()[event]
}

// GetEventGroups returns the hook groups for a given event name.
func (c *HooksConfig) GetEventGroups(event string) []HookGroup {
	ptr := c.eventGroupPtr(event)
	if ptr == nil {
		return nil
	}
	return *ptr
}

// SetEventGroups sets the hook groups for a given event name.
func (c *HooksConfig) SetEventGroups(event string, groups []HookGroup) {
	ptr := c.eventGroupPtr(event)
	if ptr == nil {
		return
	}
	*ptr = groups
}

// AllEventNames returns all 12 hook event names in canonical order.
func AllEventNames() []string {
	return []string{
		"SessionStart", "SessionEnd",
		"PreToolUse", "PostToolUse",
		"UserPromptSubmit", "TaskCompleted",
		"Stop", "PreCompact",
		"SubagentStop", "WorktreeCreate",
		"WorktreeRemove", "ConfigChange",
	}
}

// HookCoverageContract describes the active coverage contract for runtime checks.
type HookCoverageContract struct {
	ActiveEvents   []string
	FallbackReason string
}

// FallbackHookCoverageContract returns a contract covering all events with a fallback reason.
func FallbackHookCoverageContract(reason string) HookCoverageContract {
	events := append([]string(nil), AllEventNames()...)
	return HookCoverageContract{
		ActiveEvents:   events,
		FallbackReason: reason,
	}
}

// ActiveEventNamesFromConfig extracts the list of events that have at least one hook group.
func ActiveEventNamesFromConfig(config *HooksConfig) []string {
	active := make([]string, 0)
	for _, event := range AllEventNames() {
		if len(config.GetEventGroups(event)) > 0 {
			active = append(active, event)
		}
	}
	return active
}

// CountInstalledEventsForList counts how many events have non-empty hook groups in a raw hooks map.
func CountInstalledEventsForList(hooksMap map[string]any, events []string) int {
	installed := 0
	for _, event := range events {
		if groups, ok := hooksMap[event].([]any); ok && len(groups) > 0 {
			installed++
		}
	}
	return installed
}

// IsAoManagedHookCommand checks whether a command string references an ao-managed hook.
func IsAoManagedHookCommand(cmd string) bool {
	if strings.Contains(cmd, "ao ") {
		return true
	}
	normalized := filepath.ToSlash(cmd)
	return strings.Contains(normalized, "/.agentops/hooks/")
}

// RawGroupHooksContainAo checks the new-format hooks array for ao commands.
func RawGroupHooksContainAo(group map[string]any) bool {
	hooks, ok := group["hooks"].([]any)
	if !ok {
		return false
	}
	for _, h := range hooks {
		hook, ok := h.(map[string]any)
		if !ok {
			continue
		}
		if cmd, ok := hook["command"].(string); ok && IsAoManagedHookCommand(cmd) {
			return true
		}
	}
	return false
}

// RawGroupLegacyContainsAo checks the legacy format for ao commands.
func RawGroupLegacyContainsAo(group map[string]any) bool {
	cmd, ok := group["command"].([]any)
	if !ok || len(cmd) <= 1 {
		return false
	}
	cmdStr, ok := cmd[1].(string)
	return ok && IsAoManagedHookCommand(cmdStr)
}

// RawGroupIsAoManaged returns true if a raw hook group (from settings.json) is ao-managed.
func RawGroupIsAoManaged(group map[string]any) bool {
	return RawGroupHooksContainAo(group) || RawGroupLegacyContainsAo(group)
}

// HookGroupContainsAo checks if any hook group in the given event contains an ao command.
func HookGroupContainsAo(hooksMap map[string]any, event string) bool {
	groups, ok := hooksMap[event].([]any)
	if !ok {
		return false
	}
	for _, g := range groups {
		group, ok := g.(map[string]any)
		if !ok {
			continue
		}
		if RawGroupIsAoManaged(group) {
			return true
		}
	}
	return false
}

// CollectLegacyAoManagedEvents returns events that have ao-managed hooks but are not in the active set.
func CollectLegacyAoManagedEvents(hooksMap map[string]any, activeEvents []string) []string {
	activeSet := make(map[string]struct{}, len(activeEvents))
	for _, event := range activeEvents {
		activeSet[event] = struct{}{}
	}

	legacyEvents := make([]string, 0)
	for _, event := range AllEventNames() {
		if _, ok := activeSet[event]; ok {
			continue
		}
		if HookGroupContainsAo(hooksMap, event) {
			legacyEvents = append(legacyEvents, event)
		}
	}
	return legacyEvents
}

// FormatLegacyPreservationReport formats a human-readable report of legacy events.
func FormatLegacyPreservationReport(legacyEvents []string) string {
	if len(legacyEvents) == 0 {
		return ""
	}
	return fmt.Sprintf(
		"Preserved legacy ao-managed hooks outside active contract (%d): %s",
		len(legacyEvents),
		strings.Join(legacyEvents, ", "),
	)
}

// hooksManifest wraps the hooks.json file format which has a top-level "hooks" key.
type hooksManifest struct {
	Hooks *HooksConfig `json:"hooks"`
}

// ReadHooksManifest parses a hooks.json manifest from raw bytes.
// The manifest wraps events in a top-level "hooks" key and may contain a "$schema" key.
func ReadHooksManifest(data []byte) (*HooksConfig, error) {
	var manifest hooksManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse hooks manifest: %w", err)
	}
	if manifest.Hooks == nil {
		return nil, fmt.Errorf("hooks manifest missing 'hooks' key")
	}
	return manifest.Hooks, nil
}

// ReplacePluginRoot replaces ${CLAUDE_PLUGIN_ROOT} in command strings with the given base path.
func ReplacePluginRoot(config *HooksConfig, basePath string) {
	for _, event := range AllEventNames() {
		groups := config.GetEventGroups(event)
		for i := range groups {
			for j := range groups[i].Hooks {
				groups[i].Hooks[j].Command = strings.ReplaceAll(
					groups[i].Hooks[j].Command,
					"${CLAUDE_PLUGIN_ROOT}",
					basePath,
				)
			}
		}
	}
}

// GenerateMinimalHooksConfig returns the bare-minimum flywheel config (SessionStart + SessionEnd + Stop).
func GenerateMinimalHooksConfig() *HooksConfig {
	return &HooksConfig{
		SessionStart: []HookGroup{
			{
				Hooks: []HookEntry{
					{Type: "command", Command: "${CLAUDE_PLUGIN_ROOT}/hooks/session-start.sh"},
				},
			},
		},
		SessionEnd: []HookGroup{
			{
				Hooks: []HookEntry{
					{Type: "command", Command: "${CLAUDE_PLUGIN_ROOT}/hooks/session-end-maintenance.sh", Timeout: 35},
				},
			},
		},
		Stop: []HookGroup{
			{
				Hooks: []HookEntry{
					{Type: "command", Command: "${CLAUDE_PLUGIN_ROOT}/hooks/ao-flywheel-close.sh", Timeout: 15},
				},
			},
		},
	}
}

// HookGroupToMap converts a HookGroup to a map for JSON serialization.
func HookGroupToMap(g HookGroup) map[string]any {
	hooks := make([]map[string]any, len(g.Hooks))
	for i, h := range g.Hooks {
		entry := map[string]any{
			"type":    h.Type,
			"command": h.Command,
		}
		if h.Timeout > 0 {
			entry["timeout"] = h.Timeout
		}
		hooks[i] = entry
	}
	result := map[string]any{
		"hooks": hooks,
	}
	if g.Matcher != "" {
		result["matcher"] = g.Matcher
	}
	return result
}
