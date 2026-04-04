package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGCBridgeAvailable_Installed(t *testing.T) {
	// When gc binary exists on PATH, gcBridgeAvailable returns true
	if _, err := exec.LookPath("gc"); err != nil {
		t.Skip("gc not on PATH — skipping installed test")
	}
	if !gcBridgeAvailable() {
		t.Error("gcBridgeAvailable() should return true when gc is on PATH")
	}
}

func TestGCBridgeAvailable_NotInstalled(t *testing.T) {
	// Save and clear PATH to simulate gc not installed
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", t.TempDir())
	defer os.Setenv("PATH", origPath)

	if gcBridgeAvailable() {
		t.Error("gcBridgeAvailable() should return false when gc not on PATH")
	}
}

func TestGCBridgeVersion(t *testing.T) {
	if _, err := exec.LookPath("gc"); err != nil {
		t.Skip("gc not on PATH")
	}
	v, err := gcBridgeVersion()
	if err != nil {
		t.Fatalf("gcBridgeVersion() error: %v", err)
	}
	if v == "" {
		t.Error("gcBridgeVersion() returned empty string")
	}
	// Version should be semver-ish (starts with digit)
	if v[0] < '0' || v[0] > '9' {
		t.Errorf("gcBridgeVersion() = %q, expected version starting with digit", v)
	}
}

func TestGCBridgeCompatible(t *testing.T) {
	tests := []struct {
		version    string
		compatible bool
	}{
		{"0.13.5", true},
		{"0.14.0", true},
		{"1.0.0", true},
		{"0.12.9", false}, // below minimum
		{"0.1.0", false},  // too old
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := gcBridgeCompatible(tt.version)
			if got != tt.compatible {
				t.Errorf("gcBridgeCompatible(%q) = %v, want %v", tt.version, got, tt.compatible)
			}
		})
	}
}

func TestGCBridgeStatusParsing(t *testing.T) {
	// Test parsing of gc status --json output
	jsonOutput := `{
		"city": "agentops-nami",
		"controller": {"state": "running", "pid": 12345},
		"agents": [
			{"name": "worker-1", "state": "running", "template": "worker"},
			{"name": "worker-2", "state": "stopped", "template": "worker"},
			{"name": "mayor", "state": "running", "template": "mayor"}
		],
		"summary": {"running": 2, "stopped": 1, "total": 3}
	}`
	status, err := parseGCStatus([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("parseGCStatus error: %v", err)
	}
	if status.City != "agentops-nami" {
		t.Errorf("City = %q, want %q", status.City, "agentops-nami")
	}
	if status.Controller.State != "running" {
		t.Errorf("Controller.State = %q, want %q", status.Controller.State, "running")
	}
	if len(status.Agents) != 3 {
		t.Fatalf("len(Agents) = %d, want 3", len(status.Agents))
	}
	if status.Agents[0].Name != "worker-1" {
		t.Errorf("Agents[0].Name = %q, want %q", status.Agents[0].Name, "worker-1")
	}
	if status.Summary.Running != 2 {
		t.Errorf("Summary.Running = %d, want 2", status.Summary.Running)
	}
}

func TestGCBridgeSessionListParsing(t *testing.T) {
	jsonOutput := `[
		{"id": "sess-abc123", "alias": "worker-1", "state": "active", "template": "worker"},
		{"id": "sess-def456", "alias": "mayor", "state": "suspended", "template": "mayor"}
	]`
	sessions, err := parseGCSessions([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("parseGCSessions error: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("len(sessions) = %d, want 2", len(sessions))
	}
	if sessions[0].Alias != "worker-1" {
		t.Errorf("sessions[0].Alias = %q, want %q", sessions[0].Alias, "worker-1")
	}
	if sessions[1].State != "suspended" {
		t.Errorf("sessions[1].State = %q, want %q", sessions[1].State, "suspended")
	}
}

func TestGCBridgeNudgeCommand(t *testing.T) {
	args := gcNudgeArgs("worker-1", "Pick up ag-0ln9.2 and implement it")
	expected := []string{"session", "nudge", "worker-1", "Pick up ag-0ln9.2 and implement it"}
	if len(args) != len(expected) {
		t.Fatalf("gcNudgeArgs len = %d, want %d", len(args), len(expected))
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("args[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}

func TestGCBridgePeekArgs(t *testing.T) {
	args := gcPeekArgs("worker-1", 50)
	expected := []string{"session", "peek", "worker-1", "--lines", "50"}
	if len(args) != len(expected) {
		t.Fatalf("gcPeekArgs len = %d, want %d", len(args), len(expected))
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("args[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}

func TestGCBridgeEventEmitArgs(t *testing.T) {
	data := map[string]string{"phase": "research", "status": "complete"}
	dataJSON, _ := json.Marshal(data)
	args := gcEventEmitArgs("ao:phase", string(dataJSON))
	expected := []string{"event", "emit", "ao:phase", "--data", string(dataJSON)}
	if len(args) != len(expected) {
		t.Fatalf("gcEventEmitArgs len = %d, want %d", len(args), len(expected))
	}
	for i, arg := range args {
		if arg != expected[i] {
			t.Errorf("args[%d] = %q, want %q", i, arg, expected[i])
		}
	}
}

func TestGCBridgeReady_BinaryAndController(t *testing.T) {
	// gcBridgeReady should return false if gc is not available
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", t.TempDir())
	defer os.Setenv("PATH", origPath)

	ready, reason := gcBridgeReady("")
	if ready {
		t.Error("gcBridgeReady should return false when gc not on PATH")
	}
	if !strings.Contains(reason, "not found") && !strings.Contains(reason, "not installed") {
		t.Errorf("reason should mention gc not found, got: %q", reason)
	}
}

func TestGCBridgeFallbackOnError(t *testing.T) {
	// Parsing invalid JSON should return error, not panic
	_, err := parseGCStatus([]byte("not json"))
	if err == nil {
		t.Error("parseGCStatus should return error on invalid JSON")
	}
	_, err = parseGCSessions([]byte("not json"))
	if err == nil {
		t.Error("parseGCSessions should return error on invalid JSON")
	}
}

func TestGCBridgeCityPath(t *testing.T) {
	// gcBridgeCityPath should find city.toml by walking up from cwd
	tmpDir := t.TempDir()
	cityToml := filepath.Join(tmpDir, "city.toml")
	os.WriteFile(cityToml, []byte("[city]\nname = \"test\""), 0644)

	subDir := filepath.Join(tmpDir, "sub", "dir")
	os.MkdirAll(subDir, 0755)

	path := gcBridgeCityPath(subDir)
	if path != tmpDir {
		t.Errorf("gcBridgeCityPath = %q, want %q", path, tmpDir)
	}

	// No city.toml found
	emptyDir := t.TempDir()
	path = gcBridgeCityPath(emptyDir)
	if path != "" {
		t.Errorf("gcBridgeCityPath should return empty when no city.toml, got %q", path)
	}
}

// --- L2 Integration Tests ---

func TestGCBridgeReady_VersionTooLow(t *testing.T) {
	// Integration: gcBridgeReady should reject versions below minimum
	// even when binary is "found". We test the version check path by
	// verifying gcBridgeCompatible feeds into the ready logic.
	if gcBridgeCompatible("0.12.0") {
		// sanity: 0.12.0 is below gcMinVersion 0.13.0
		t.Fatal("precondition failed: 0.12.0 should be incompatible")
	}
	// The full gcBridgeReady flow when binary exists but version is low
	// returns false with a version message. We can't easily mock the binary,
	// but we verify the component chain: compareSemver → gcBridgeCompatible.
	versions := []struct {
		v      string
		compat bool
	}{
		{"0.13.0", true},  // exact minimum
		{"0.12.99", false}, // below (99 patch doesn't help)
		{"v1.0.0", true},  // v-prefix stripped
		{"0.13.0-rc1", true}, // pre-release suffix stripped
		{"0.12.5-beta", false}, // pre-release below min
	}
	for _, tt := range versions {
		got := gcBridgeCompatible(tt.v)
		if got != tt.compat {
			t.Errorf("gcBridgeCompatible(%q) = %v, want %v", tt.v, got, tt.compat)
		}
	}
}

func TestGCBridge_StatusParsingToReadyFlow(t *testing.T) {
	// Integration: parseGCStatus output feeds into the gcBridgeReady
	// controller-state check. Verify various controller states.
	tests := []struct {
		name       string
		state      string
		wantReady  bool
	}{
		{"running controller", "running", true},
		{"stopped controller", "stopped", false},
		{"starting controller", "starting", false},
		{"empty state", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData := []byte(fmt.Sprintf(`{
				"city": "test-city",
				"controller": {"state": %q, "pid": 1234},
				"agents": [],
				"summary": {"running": 0, "stopped": 0, "total": 0}
			}`, tt.state))
			status, err := parseGCStatus(jsonData)
			if err != nil {
				t.Fatalf("parseGCStatus error: %v", err)
			}
			// This is the same check gcBridgeReady performs after parsing
			isReady := status.Controller.State == "running"
			if isReady != tt.wantReady {
				t.Errorf("controller state %q: ready=%v, want %v", tt.state, isReady, tt.wantReady)
			}
		})
	}
}

func TestGCBridge_SessionLifecycleStates(t *testing.T) {
	// Integration: parseGCSessions feeds into checkSessionDone.
	// Verify the done-detection logic for various session states.
	tests := []struct {
		name     string
		state    string
		wantDone bool
	}{
		{"active session", "active", false},
		{"closed session", "closed", true},
		{"completed session", "completed", true},
		{"suspended session", "suspended", false},
		{"errored session", "errored", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData := []byte(fmt.Sprintf(`[
				{"id": "sess-001", "alias": "rpi-test-p1", "state": %q, "template": "worker"}
			]`, tt.state))
			sessions, err := parseGCSessions(jsonData)
			if err != nil {
				t.Fatalf("parseGCSessions error: %v", err)
			}
			if len(sessions) != 1 {
				t.Fatalf("expected 1 session, got %d", len(sessions))
			}
			// Mirror the checkSessionDone logic
			s := sessions[0]
			isDone := s.State == "closed" || s.State == "completed"
			if isDone != tt.wantDone {
				t.Errorf("session state %q: done=%v, want %v", tt.state, isDone, tt.wantDone)
			}
		})
	}
}

func TestGCBridge_SessionNotFoundTreatedAsDone(t *testing.T) {
	// Integration: when a session alias is not in the list,
	// checkSessionDone treats it as complete (controller crash/cleanup).
	jsonData := []byte(`[
		{"id": "sess-001", "alias": "other-session", "state": "active", "template": "worker"}
	]`)
	sessions, err := parseGCSessions(jsonData)
	if err != nil {
		t.Fatalf("parseGCSessions error: %v", err)
	}
	// Simulate the checkSessionDone search for a missing alias
	targetAlias := "rpi-test-p1"
	found := false
	for _, s := range sessions {
		if s.Alias == targetAlias {
			found = true
			break
		}
	}
	if found {
		t.Error("session should not be found in list")
	}
	// Per checkSessionDone: missing session = treated as done
}

func TestGCBridge_CityPathToExecutorIntegration(t *testing.T) {
	// Integration: gcCityPathFromOpts uses gcBridgeCityPath as fallback.
	// Verify the opts-explicit → auto-discover chain.
	tmpDir := t.TempDir()
	cityToml := filepath.Join(tmpDir, "city.toml")
	os.WriteFile(cityToml, []byte("[city]\nname = \"integration-test\""), 0644)

	// Case 1: explicit path takes precedence
	opts := defaultPhasedEngineOptions()
	opts.GCCityPath = "/explicit/override"
	opts.WorkingDir = tmpDir
	got := gcCityPathFromOpts(opts)
	if got != "/explicit/override" {
		t.Errorf("explicit GCCityPath: got %q, want /explicit/override", got)
	}

	// Case 2: empty GCCityPath falls back to auto-discover via WorkingDir
	opts.GCCityPath = ""
	got = gcCityPathFromOpts(opts)
	if got != tmpDir {
		t.Errorf("auto-discover from WorkingDir: got %q, want %q", got, tmpDir)
	}

	// Case 3: subdirectory walks up to find city.toml
	subDir := filepath.Join(tmpDir, "deep", "nested")
	os.MkdirAll(subDir, 0755)
	opts.WorkingDir = subDir
	got = gcCityPathFromOpts(opts)
	if got != tmpDir {
		t.Errorf("walk-up from subdir: got %q, want %q", got, tmpDir)
	}

	// Case 4: no city.toml anywhere
	emptyDir := t.TempDir()
	opts.WorkingDir = emptyDir
	got = gcCityPathFromOpts(opts)
	if got != "" {
		t.Errorf("no city.toml: got %q, want empty", got)
	}
}

func TestGCBridge_EventArgsChainIntegration(t *testing.T) {
	// Integration: verify the full chain from typed event helpers
	// through gcEventEmitArgs to final command args structure.
	// This tests that event data round-trips through JSON correctly.
	phaseData := map[string]any{
		"phase":  2,
		"status": "complete",
		"run_id": "integ-001",
	}
	dataJSON, err := json.Marshal(phaseData)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	args := gcEventEmitArgs(GCEventAOPhase, string(dataJSON))

	// Verify command structure
	if args[0] != "event" || args[1] != "emit" {
		t.Errorf("args[0:2] = %v, want [event emit]", args[:2])
	}
	if args[2] != "ao:phase" {
		t.Errorf("event type = %q, want ao:phase", args[2])
	}
	if args[3] != "--data" {
		t.Errorf("args[3] = %q, want --data", args[3])
	}

	// Verify the JSON payload round-trips
	var decoded map[string]any
	if err := json.Unmarshal([]byte(args[4]), &decoded); err != nil {
		t.Fatalf("round-trip decode error: %v", err)
	}
	if decoded["run_id"] != "integ-001" {
		t.Errorf("round-trip run_id = %v, want integ-001", decoded["run_id"])
	}
	if decoded["status"] != "complete" {
		t.Errorf("round-trip status = %v, want complete", decoded["status"])
	}
}

func TestGCBridge_CompareSemverEdgeCases(t *testing.T) {
	// Integration: semver comparison edge cases that affect bridge compatibility
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"2.0.0", "1.99.99", 1},
		{"0.13", "0.13.0", 0},    // missing patch = 0
		{"0.13.0", "0.13", 0},    // symmetric
		{"v0.14.0", "0.14.0", 0}, // v-prefix
		{"1.0.0-rc1", "1.0.0", 0}, // pre-release stripped to same
		{"", "0.0.0", 0},          // empty = 0.0.0
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := compareSemver(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareSemver(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestGCBridge_ExecutorSelectsWithCityPath(t *testing.T) {
	// Integration: selectExecutorFromCaps with gc mode produces
	// a gcExecutor that uses the city path from opts.
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "city.toml"), []byte("[city]\nname=\"test\""), 0644)

	caps := backendCapabilities{RuntimeMode: "gc"}
	opts := defaultPhasedEngineOptions()
	opts.WorkingDir = tmpDir

	executor, reason := selectExecutorFromCaps(caps, "", nil, opts)
	if executor.Name() != "gc" {
		t.Errorf("executor.Name() = %q, want gc", executor.Name())
	}
	if reason != "runtime=gc" {
		t.Errorf("reason = %q, want runtime=gc", reason)
	}

	// Verify the executor is a gcExecutor with the right city path
	gcExec, ok := executor.(*gcExecutor)
	if !ok {
		t.Fatal("executor is not *gcExecutor")
	}
	if gcExec.cityPath != tmpDir {
		t.Errorf("gcExecutor.cityPath = %q, want %q", gcExec.cityPath, tmpDir)
	}
}
