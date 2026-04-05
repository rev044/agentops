package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// gcExecCommand is the function used to create exec.Cmd instances.
// Tests can replace this to intercept shell-outs.
var gcExecCommand = exec.Command

// gcLookPath is the function used to find binaries on PATH.
// Tests can replace this to simulate gc presence/absence.
var gcLookPath = exec.LookPath

// gcMinVersion is the minimum gc version required for bridge compatibility.
const gcMinVersion = "0.13.0"

// GCStatus represents the parsed output of `gc status --json`.
type GCStatus struct {
	City       string          `json:"city"`
	Controller GCController    `json:"controller"`
	Agents     []GCAgentInfo   `json:"agents"`
	Summary    GCStatusSummary `json:"summary"`
}

// GCController represents the controller state within GCStatus.
type GCController struct {
	Running bool   `json:"running"`
	PID     int    `json:"pid"`
	Mode    string `json:"mode"`
}

// GCAgentInfo represents a single agent entry within GCStatus.
type GCAgentInfo struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	Template string `json:"template"`
}

// GCStatusSummary holds aggregate agent counts.
type GCStatusSummary struct {
	Running int `json:"running"`
	Stopped int `json:"stopped"`
	Total   int `json:"total"`
}

// GCSession represents a session from `gc session list --json`.
type GCSession struct {
	ID       string `json:"id"`
	Alias    string `json:"alias"`
	State    string `json:"state"`
	Template string `json:"template"`
}

// gcBridgeAvailable returns true if the gc binary is on PATH.
func gcBridgeAvailable() bool {
	_, err := gcLookPath("gc")
	return err == nil
}

// gcBridgeVersion returns the gc version string.
func gcBridgeVersion() (string, error) {
	out, err := gcExecCommand("gc", "version").Output()
	if err != nil {
		return "", fmt.Errorf("gc version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// gcBridgeCompatible checks if the given version meets the minimum requirement.
func gcBridgeCompatible(version string) bool {
	return compareSemver(version, gcMinVersion) >= 0
}

// compareSemver returns -1, 0, or 1 comparing two semver strings.
func compareSemver(a, b string) int {
	aParts := parseSemverParts(a)
	bParts := parseSemverParts(b)
	for i := 0; i < 3; i++ {
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}
	return 0
}

// parseSemverParts extracts major, minor, patch integers from a version string.
func parseSemverParts(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	var result [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		// Strip any pre-release suffix (e.g. "5-rc1")
		num := strings.SplitN(parts[i], "-", 2)[0]
		result[i], _ = strconv.Atoi(num)
	}
	return result
}

// gcBridgeReady checks both binary availability AND controller running.
// Returns (ready, reason).
func gcBridgeReady(cityPath string) (bool, string) {
	if !gcBridgeAvailable() {
		return false, "gc binary not found on PATH"
	}
	v, err := gcBridgeVersion()
	if err != nil {
		return false, fmt.Sprintf("gc version check failed: %v", err)
	}
	if !gcBridgeCompatible(v) {
		return false, fmt.Sprintf("gc version %s below minimum %s", v, gcMinVersion)
	}
	// Check if controller is running by attempting gc status
	args := []string{"status", "--json"}
	if cityPath != "" {
		args = append([]string{"--city", cityPath}, args...)
	}
	out, err := gcExecCommand("gc", args...).Output()
	if err != nil {
		return false, fmt.Sprintf("gc controller not running: %v", err)
	}
	status, err := parseGCStatus(out)
	if err != nil {
		return false, fmt.Sprintf("gc status parse error: %v", err)
	}
	if !status.Controller.Running {
		return false, "gc controller not running"
	}
	return true, "gc bridge ready"
}

// parseGCStatus parses the JSON output of `gc status --json`.
func parseGCStatus(data []byte) (GCStatus, error) {
	var status GCStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return GCStatus{}, fmt.Errorf("parse gc status: %w", err)
	}
	return status, nil
}

// parseGCSessions parses the JSON output of `gc session list --json`.
func parseGCSessions(data []byte) ([]GCSession, error) {
	var sessions []GCSession
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, fmt.Errorf("parse gc sessions: %w", err)
	}
	return sessions, nil
}

// gcNudgeArgs returns the command arguments for `gc session nudge`.
func gcNudgeArgs(agent, message string) []string {
	return []string{"session", "nudge", agent, message}
}

// gcPeekArgs returns the command arguments for `gc session peek`.
func gcPeekArgs(agent string, lines int) []string {
	return []string{"session", "peek", agent, "--lines", strconv.Itoa(lines)}
}

// gcEventEmitArgs returns the command arguments for `gc event emit`.
func gcEventEmitArgs(eventType, dataJSON string) []string {
	return []string{"event", "emit", eventType, "--payload", dataJSON}
}

// gcBridgeCityPath walks up from cwd looking for city.toml.
// Returns the directory containing city.toml, or empty string if not found.
func gcBridgeCityPath(cwd string) string {
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "city.toml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
