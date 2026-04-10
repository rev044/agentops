package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func writeExecutable(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", name, err)
	}
	return path
}

func TestRunOvernightStartDryRunJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testProjectDir = tmpDir
	defer func() { testProjectDir = "" }()

	oldDryRun := dryRun
	oldOutput := output
	oldGoal := overnightGoal
	oldOutputDir := overnightOutputDir
	oldRunTimeout := overnightRunTimeout
	oldKeepAwake := overnightKeepAwake
	oldNoKeepAwake := overnightNoKeepAwake
	oldRunners := append([]string{}, overnightRunners...)
	oldModels := overnightModels
	oldCreative := overnightCreative
	defer func() {
		dryRun = oldDryRun
		output = oldOutput
		overnightGoal = oldGoal
		overnightOutputDir = oldOutputDir
		overnightRunTimeout = oldRunTimeout
		overnightKeepAwake = oldKeepAwake
		overnightNoKeepAwake = oldNoKeepAwake
		overnightRunners = append([]string{}, oldRunners...)
		overnightModels = oldModels
		overnightCreative = oldCreative
	}()

	dryRun = true
	output = "json"
	overnightGoal = "stabilize dream slice"
	overnightRunners = []string{"codex", "claude"}
	overnightCreative = true

	cmd := &cobra.Command{}
	cmd.Flags().String("output-dir", "", "")
	cmd.Flags().String("run-timeout", "", "")
	cmd.Flags().Bool("keep-awake", false, "")
	cmd.Flags().Bool("no-keep-awake", false, "")
	cmd.Flags().StringSlice("runner", nil, "")
	cmd.Flags().String("models", "", "")
	cmd.Flags().Bool("creative-lane", false, "")

	stdout, err := captureStdout(t, func() error {
		return runOvernightStart(cmd, nil)
	})
	if err != nil {
		t.Fatalf("runOvernightStart dry-run: %v", err)
	}

	var summary overnightSummary
	if err := json.Unmarshal([]byte(stdout), &summary); err != nil {
		t.Fatalf("parse dry-run summary: %v\noutput=%s", err, stdout)
	}
	if summary.Status != "dry-run" {
		t.Fatalf("status = %q, want dry-run", summary.Status)
	}
	if summary.Runtime.KeepAwakeMode != "disabled" {
		t.Fatalf("keep_awake_mode = %q, want disabled in dry-run", summary.Runtime.KeepAwakeMode)
	}
	if summary.Council == nil {
		t.Fatal("expected Dream Council plan in dry-run summary")
	}
	if got := strings.Join(summary.Council.RequestedRunners, ","); got != "claude,codex" {
		t.Fatalf("requested_runners = %q, want claude,codex", got)
	}
	if summary.Dreamscape == nil {
		t.Fatal("expected DreamScape in dry-run summary")
	}
	if _, ok := summary.Artifacts["council_packet"]; !ok {
		t.Fatalf("expected council_packet artifact in %#v", summary.Artifacts)
	}
	if !strings.Contains(summary.NextAction, "ao rpi phased") {
		t.Fatalf("next_action = %q, want RPI recommendation", summary.NextAction)
	}
}

func TestRunOvernightSetupDryRunJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testProjectDir = tmpDir
	defer func() { testProjectDir = "" }()

	for _, name := range []string{"codex", "claude", "caffeinate"} {
		writeExecutable(t, tmpDir, name, "#!/bin/sh\nexit 0\n")
	}
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("AGENTOPS_CONFIG", filepath.Join(tmpDir, ".agentops", "config.yaml"))

	oldOutput := output
	oldSetupApply := overnightSetupApply
	oldSetupScheduler := overnightSetupScheduler
	oldSetupAt := overnightSetupAt
	oldSetupRunners := append([]string{}, overnightSetupRunners...)
	oldSetupKeepAwake := overnightSetupKeepAwake
	oldSetupNoKeepAwake := overnightSetupNoKeepAwake
	oldDreamOS := dreamOS
	oldDreamBatteryStatus := dreamBatteryStatus
	defer func() {
		output = oldOutput
		overnightSetupApply = oldSetupApply
		overnightSetupScheduler = oldSetupScheduler
		overnightSetupAt = oldSetupAt
		overnightSetupRunners = append([]string{}, oldSetupRunners...)
		overnightSetupKeepAwake = oldSetupKeepAwake
		overnightSetupNoKeepAwake = oldSetupNoKeepAwake
		dreamOS = oldDreamOS
		dreamBatteryStatus = oldDreamBatteryStatus
	}()

	output = "json"
	overnightSetupScheduler = "auto"
	overnightSetupRunners = nil
	dreamOS = "darwin"
	dreamBatteryStatus = func() bool { return true }

	cmd := &cobra.Command{}
	cmd.Flags().Bool("apply", false, "")
	cmd.Flags().String("scheduler", "auto", "")
	cmd.Flags().String("at", "", "")
	cmd.Flags().StringSlice("runner", nil, "")
	cmd.Flags().Bool("keep-awake", false, "")
	cmd.Flags().Bool("no-keep-awake", false, "")

	stdout, err := captureStdout(t, func() error {
		return runOvernightSetup(cmd, nil)
	})
	if err != nil {
		t.Fatalf("runOvernightSetup dry-run: %v", err)
	}

	var summary dreamSetupSummary
	if err := json.Unmarshal([]byte(stdout), &summary); err != nil {
		t.Fatalf("parse setup summary: %v\noutput=%s", err, stdout)
	}
	if summary.Status != "dry-run" {
		t.Fatalf("status = %q, want dry-run", summary.Status)
	}
	if summary.Host.RecommendedMode != "manual" {
		t.Fatalf("recommended_mode = %q, want manual", summary.Host.RecommendedMode)
	}
	if summary.DreamConfig.SchedulerMode != "manual" {
		t.Fatalf("scheduler_mode = %q, want manual", summary.DreamConfig.SchedulerMode)
	}
	if summary.DreamConfig.KeepAwake == nil || !*summary.DreamConfig.KeepAwake {
		t.Fatalf("keep_awake = %#v, want true", summary.DreamConfig.KeepAwake)
	}
	if got := strings.Join(summary.DreamConfig.Runners, ","); got != "claude,codex" {
		t.Fatalf("dream runners = %q, want claude,codex", got)
	}
	if !strings.Contains(summary.NextAction, "ao overnight start") {
		t.Fatalf("next_action = %q, want overnight start guidance", summary.NextAction)
	}
}

func TestRunOvernightSetupApplyWritesSchedulerArtifact(t *testing.T) {
	tmpDir := t.TempDir()
	testProjectDir = tmpDir
	defer func() { testProjectDir = "" }()

	for _, name := range []string{"codex", "claude", "caffeinate"} {
		writeExecutable(t, tmpDir, name, "#!/bin/sh\nexit 0\n")
	}
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	configPath := filepath.Join(tmpDir, ".agentops", "config.yaml")
	t.Setenv("AGENTOPS_CONFIG", configPath)

	oldOutput := output
	oldSetupApply := overnightSetupApply
	oldSetupScheduler := overnightSetupScheduler
	oldSetupAt := overnightSetupAt
	oldSetupRunners := append([]string{}, overnightSetupRunners...)
	oldDreamOS := dreamOS
	oldDreamBatteryStatus := dreamBatteryStatus
	oldDryRun := dryRun
	defer func() {
		output = oldOutput
		overnightSetupApply = oldSetupApply
		overnightSetupScheduler = oldSetupScheduler
		overnightSetupAt = oldSetupAt
		overnightSetupRunners = append([]string{}, oldSetupRunners...)
		dreamOS = oldDreamOS
		dreamBatteryStatus = oldDreamBatteryStatus
		dryRun = oldDryRun
	}()

	output = "json"
	dryRun = false
	overnightSetupApply = true
	overnightSetupScheduler = "launchd"
	overnightSetupAt = "01:30"
	dreamOS = "darwin"
	dreamBatteryStatus = func() bool { return false }

	cmd := &cobra.Command{}
	cmd.Flags().Bool("apply", false, "")
	cmd.Flags().String("scheduler", "auto", "")
	cmd.Flags().String("at", "", "")
	cmd.Flags().StringSlice("runner", nil, "")
	cmd.Flags().Bool("keep-awake", false, "")
	cmd.Flags().Bool("no-keep-awake", false, "")
	if err := cmd.Flags().Set("at", "01:30"); err != nil {
		t.Fatalf("set --at: %v", err)
	}

	stdout, err := captureStdout(t, func() error {
		return runOvernightSetup(cmd, nil)
	})
	if err != nil {
		t.Fatalf("runOvernightSetup apply: %v", err)
	}

	var summary dreamSetupSummary
	if err := json.Unmarshal([]byte(stdout), &summary); err != nil {
		t.Fatalf("parse setup summary: %v\noutput=%s", err, stdout)
	}
	if summary.Status != "configured" {
		t.Fatalf("status = %q, want configured", summary.Status)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config at %s: %v", configPath, err)
	}
	launchdPath := filepath.Join(tmpDir, ".agentops", "generated", "dream", "com.agentops.dream.plist")
	if _, err := os.Stat(launchdPath); err != nil {
		t.Fatalf("expected launchd artifact at %s: %v", launchdPath, err)
	}
}

func TestRunDreamCouncilWithMockRunners(t *testing.T) {
	tmpDir := t.TempDir()
	writeExecutable(t, tmpDir, "codex", `#!/bin/sh
out=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    shift
    out="$1"
  fi
  shift
done
cat > "$out" <<'JSON'
{"runner":"codex","headline":"Codex says validate","recommended_kind":"validate","recommended_first_action":"Review the overnight council synthesis before shipping.","risks":["retrieval drift"],"opportunities":["promote learnings"],"confidence":"high","wildcard_idea":"Explore a speculative promotion lane."}
JSON
`)
	writeExecutable(t, tmpDir, "claude", `#!/bin/sh
cat <<'JSON'
{"runner":"claude","headline":"Claude agrees","recommended_kind":"validate","recommended_first_action":"Review the overnight council synthesis before shipping.","risks":["low retrieval coverage"],"opportunities":["tighten report copy"],"confidence":"high"}
JSON
`)
	t.Setenv("PATH", tmpDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	summary := overnightSummary{
		RunID:         "test-run",
		Goal:          "stabilize Dream",
		RepoRoot:      tmpDir,
		OutputDir:     filepath.Join(tmpDir, "overnight"),
		Artifacts:     map[string]string{},
		RetrievalLive: map[string]any{"coverage": 0.72},
		MetricsHealth: map[string]any{"escape_velocity": true},
	}
	settings := overnightSettings{
		OutputDir:     summary.OutputDir,
		RunTimeoutRaw: "8h",
		RunTimeout:    8 * time.Hour,
		KeepAwake:     false,
		Runners:       []string{"codex", "claude"},
		RunnerModels:  map[string]string{},
		Consensus:     "majority",
		CreativeLane:  true,
	}
	appendDreamCouncilPlan(&summary, settings)
	var log bytes.Buffer

	if err := runDreamCouncil(context.Background(), tmpDir, &log, &summary, settings); err != nil {
		t.Fatalf("runDreamCouncil: %v", err)
	}
	ensureOvernightDerivedViews(&summary)

	if summary.Council == nil {
		t.Fatal("expected council summary")
	}
	if got := strings.Join(summary.Council.CompletedRunners, ","); got != "claude,codex" {
		t.Fatalf("completed_runners = %q, want claude,codex\ndegraded=%v\nfailed=%v\nlog=%s", got, summary.Degraded, summary.Council.FailedRunners, log.String())
	}
	if summary.Council.RecommendedFirstAction != "Review the overnight council synthesis before shipping." {
		t.Fatalf("recommended_first_action = %q", summary.Council.RecommendedFirstAction)
	}
	if len(summary.Council.WildcardIdeas) != 1 {
		t.Fatalf("wildcard_ideas = %#v, want 1 item", summary.Council.WildcardIdeas)
	}
	if summary.Dreamscape == nil || summary.Dreamscape.FirstMove != summary.Council.RecommendedFirstAction {
		t.Fatalf("dreamscape = %#v, want first move from council", summary.Dreamscape)
	}
}

func TestRunOvernightReportReadsSummaryJSON(t *testing.T) {
	tmpDir := t.TempDir()
	summary := overnightSummary{
		SchemaVersion: 1,
		Mode:          "dream.local-bedtime",
		RunID:         "test-run",
		OutputDir:     tmpDir,
		Status:        "done",
		Runtime: overnightRuntimeSummary{
			KeepAwake:        true,
			KeepAwakeMode:    "caffeinate",
			RequestedTimeout: "8h",
			EffectiveTimeout: "8h0m0s",
		},
		Council: &overnightCouncilSummary{
			RequestedRunners:       []string{"claude", "codex"},
			CompletedRunners:       []string{"claude", "codex"},
			ConsensusPolicy:        "majority",
			ConsensusKind:          "validate",
			RecommendedFirstAction: "Review the overnight council synthesis before shipping.",
		},
	}
	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "summary.json"), data, 0o644); err != nil {
		t.Fatalf("write summary: %v", err)
	}

	oldOutput := output
	oldFrom := overnightReportFrom
	defer func() {
		output = oldOutput
		overnightReportFrom = oldFrom
	}()

	output = "table"
	overnightReportFrom = tmpDir

	stdout, err := captureStdout(t, func() error {
		return runOvernightReport(&cobra.Command{}, nil)
	})
	if err != nil {
		t.Fatalf("runOvernightReport: %v", err)
	}
	if !strings.Contains(stdout, "Dream Morning Report") {
		t.Fatalf("report output missing header: %s", stdout)
	}
	if !strings.Contains(stdout, "test-run") {
		t.Fatalf("report output missing run id: %s", stdout)
	}
	if !strings.Contains(stdout, "DreamScape") {
		t.Fatalf("report output missing DreamScape section: %s", stdout)
	}
	if !strings.Contains(stdout, "Dream Council") {
		t.Fatalf("report output missing Dream Council section: %s", stdout)
	}
}
