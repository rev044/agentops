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

	ovn "github.com/boshu2/agentops/cli/internal/overnight"
	v1reader "github.com/boshu2/agentops/cli/cmd/ao/testdata/v1_reader"
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

// TestRunOvernight_HardFail_WritesFailedSummary locks the hard-fail
// contract (na-cdn): when the overnight Dream pipeline aborts on a
// hard-fail step, finalizeOvernightSummary must still write both
// summary.json and summary.md with status="failed", preserving the
// last completed step, the degraded state, and the artifacts map.
//
// This mirrors the runOvernightStart hard-fail path where, on
// ovn.RunLoop error, the command sets summary.Status = "failed" and
// calls finalizeOvernightSummary(&summary, startedAt) before returning
// the error. We exercise finalizeOvernightSummary directly (the same
// pattern TestRunDreamCouncilWithMockRunners uses for its inner
// helper) so the test is hermetic and does not require a real
// .agents/ corpus or lock acquisition.
func TestRunOvernight_HardFail_WritesFailedSummary(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "overnight", "latest")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output: %v", err)
	}

	startedAt := time.Now().UTC().Add(-3 * time.Minute)
	summary := overnightSummary{
		SchemaVersion: 2,
		Mode:          "dream.local-bedtime",
		RunID:         "hardfail-run",
		Goal:          "lock hard-fail contract",
		RepoRoot:      tmpDir,
		OutputDir:     outputDir,
		// Simulate the state runOvernightStart leaves the summary in
		// when ovn.RunLoop returns an error: status flipped to
		// "failed", a last-completed step recorded, degraded entries
		// accumulated, and the baseline artifacts map populated with
		// summary_json / summary_markdown paths.
		Status:    "failed",
		StartedAt: startedAt.Format(time.RFC3339),
		Runtime: overnightRuntimeSummary{
			KeepAwake:          false,
			KeepAwakeMode:      "disabled",
			RequestedTimeout:   "8h",
			EffectiveTimeout:   "8h0m0s",
			LogPath:            filepath.Join(outputDir, "overnight.log"),
			ProcessContractDoc: "docs/contracts/dream-run-contract.md",
			ReportContractDoc:  "docs/contracts/dream-report.md",
		},
		Steps: []overnightStepSummary{
			{
				Name:     "close-loop",
				Status:   "done",
				Command:  "ao flywheel close-loop --threshold 0h --json",
				Artifact: filepath.Join(outputDir, "close-loop.json"),
			},
			{
				Name:    "metrics-health",
				Status:  "failed",
				Command: "ao metrics health --json",
				Note:    "hard-fail: metrics-health exited non-zero",
			},
		},
		Artifacts: map[string]string{
			"close_loop":       filepath.Join(outputDir, "close-loop.json"),
			"metrics_health":   filepath.Join(outputDir, "metrics-health.json"),
			"summary_json":     filepath.Join(outputDir, "summary.json"),
			"summary_markdown": filepath.Join(outputDir, "summary.md"),
		},
		Degraded: []string{"metrics-health: hard-fail simulated for regression test"},
	}

	if err := finalizeOvernightSummary(&summary, startedAt); err != nil {
		t.Fatalf("finalizeOvernightSummary on hard-fail path: %v", err)
	}

	// Assert 1: summary.json exists and parses with status == "failed".
	summaryJSONPath := filepath.Join(outputDir, "summary.json")
	if _, err := os.Stat(summaryJSONPath); err != nil {
		t.Fatalf("expected summary.json at %s: %v", summaryJSONPath, err)
	}
	jsonBytes, err := os.ReadFile(summaryJSONPath)
	if err != nil {
		t.Fatalf("read summary.json: %v", err)
	}
	var persisted overnightSummary
	if err := json.Unmarshal(jsonBytes, &persisted); err != nil {
		t.Fatalf("parse summary.json: %v\npayload=%s", err, string(jsonBytes))
	}
	if persisted.Status != "failed" {
		t.Errorf("summary.json status = %q, want %q", persisted.Status, "failed")
	}

	// Assert 2: summary.md exists and records the failed status.
	summaryMDPath := filepath.Join(outputDir, "summary.md")
	mdBytes, err := os.ReadFile(summaryMDPath)
	if err != nil {
		t.Fatalf("read summary.md: %v", err)
	}
	md := string(mdBytes)
	if !strings.Contains(md, "- Status: `failed`") {
		t.Errorf("summary.md missing failed status line\npayload=%s", md)
	}

	// Assert 3: the LAST completed step (close-loop) is preserved
	// verbatim alongside the failed hard-fail step.
	if len(persisted.Steps) < 2 {
		t.Fatalf("persisted Steps length = %d, want >= 2", len(persisted.Steps))
	}
	var lastDone, hardFail *overnightStepSummary
	for i := range persisted.Steps {
		step := &persisted.Steps[i]
		switch step.Name {
		case "close-loop":
			lastDone = step
		case "metrics-health":
			hardFail = step
		}
	}
	if lastDone == nil {
		t.Fatal("expected last-completed step close-loop in persisted summary")
	}
	if lastDone.Status != "done" {
		t.Errorf("close-loop status = %q, want %q", lastDone.Status, "done")
	}
	if lastDone.Artifact == "" {
		t.Error("close-loop artifact was not preserved on hard-fail summary")
	}
	if hardFail == nil {
		t.Fatal("expected hard-fail step metrics-health in persisted summary")
	}
	if hardFail.Status != "failed" {
		t.Errorf("metrics-health status = %q, want %q", hardFail.Status, "failed")
	}

	// Assert 4: degraded notes and the artifacts map are preserved.
	if len(persisted.Degraded) == 0 {
		t.Error("persisted Degraded unexpectedly empty on hard-fail summary")
	}
	foundDegraded := false
	for _, d := range persisted.Degraded {
		if strings.Contains(d, "metrics-health") {
			foundDegraded = true
			break
		}
	}
	if !foundDegraded {
		t.Errorf("persisted Degraded missing metrics-health entry: %#v", persisted.Degraded)
	}
	if got := persisted.Artifacts["summary_json"]; got != summaryJSONPath {
		t.Errorf("artifacts[summary_json] = %q, want %q", got, summaryJSONPath)
	}
	if got := persisted.Artifacts["summary_markdown"]; got != summaryMDPath {
		t.Errorf("artifacts[summary_markdown] = %q, want %q", got, summaryMDPath)
	}
	if got := persisted.Artifacts["close_loop"]; got == "" {
		t.Error("artifacts[close_loop] unexpectedly empty on hard-fail summary")
	}

	// Assert 5: finalize populates FinishedAt/Duration so the report
	// is well-formed even on the failure path.
	if persisted.FinishedAt == "" {
		t.Error("persisted FinishedAt unexpectedly empty on hard-fail summary")
	}
	if persisted.Duration == "" {
		t.Error("persisted Duration unexpectedly empty on hard-fail summary")
	}
}

// TestRunOvernight_SchemaV2IsV1BackwardCompatible verifies that a v2
// Dream report JSON still parses cleanly into the frozen v1 reader
// struct. This is the pm-008 fix: the compat claim now has a real
// regression test instead of unmarshaling v2 into itself.
//
// The v1 reader lives in cli/cmd/ao/testdata/v1_reader and is a
// frozen snapshot of overnightSummary as it existed on 2026-04-09.
// DO NOT edit the v1 reader to pick up v2-only fields — the whole
// point is that a strict v1 consumer never sees them yet still reads
// every field it does know about.
func TestRunOvernight_SchemaV2IsV1BackwardCompatible(t *testing.T) {
	v2 := overnightSummary{
		SchemaVersion: 2,
		Mode:          "dream.local-bedtime",
		RunID:         "compat-run-2026-04-09",
		Goal:          "verify v1 reader still parses v2 output",
		RepoRoot:      "/tmp/repo",
		OutputDir:     "/tmp/repo/.agents/overnight/compat",
		Status:        "done",
		DryRun:        false,
		StartedAt:     "2026-04-09T00:00:00Z",
		FinishedAt:    "2026-04-09T00:05:00Z",
		Duration:      "5m0s",
		Runtime: overnightRuntimeSummary{
			KeepAwake:          true,
			KeepAwakeMode:      "caffeinate",
			KeepAwakeNote:      "test",
			RequestedTimeout:   "2h",
			EffectiveTimeout:   "2h0m0s",
			LockPath:           "/tmp/repo/.agents/overnight/run.lock",
			LogPath:            "/tmp/repo/.agents/overnight/compat/overnight.log",
			ProcessContractDoc: "docs/contracts/dream-process.md",
			ReportContractDoc:  "docs/contracts/dream-report.md",
		},
		Steps: []overnightStepSummary{
			{
				Name:     "close-loop",
				Status:   "done",
				Command:  "ao learnings close-loop",
				Artifact: "close-loop.json",
			},
			{
				Name:   "defrag-preview",
				Status: "done",
			},
		},
		Artifacts: map[string]string{
			"summary": "summary.json",
		},
		Degraded:    []string{"inject-refresh: soft-skipped in test"},
		Recommended: []string{"promote the top council finding"},
		NextAction:  "ao overnight report",

		// v2 additive fields — the whole point of this test is
		// that these DO NOT break the v1 reader.
		Iterations: []ovn.IterationSummary{
			{
				ID:         "compat-run-2026-04-09-iter-1",
				Index:      1,
				StartedAt:  time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC),
				FinishedAt: time.Date(2026, 4, 9, 0, 2, 30, 0, time.UTC),
				Duration:   "2m30s",
				Status:     "done",
			},
		},
		FitnessDelta: map[string]any{
			"maturity_provisional_or_higher": 0.02,
		},
		PlateauReason:    "",
		RegressionReason: "",
	}

	data, err := json.Marshal(v2)
	if err != nil {
		t.Fatalf("marshal v2 summary: %v", err)
	}

	var v1 v1reader.OvernightSummaryV1
	if err := json.Unmarshal(data, &v1); err != nil {
		t.Fatalf("unmarshal v2 JSON into v1 reader: %v", err)
	}

	if v1.SchemaVersion != 2 {
		t.Errorf("SchemaVersion: got %d, want 2", v1.SchemaVersion)
	}
	if v1.RunID != "compat-run-2026-04-09" {
		t.Errorf("RunID: got %q, want %q", v1.RunID, "compat-run-2026-04-09")
	}
	if v1.Mode != "dream.local-bedtime" {
		t.Errorf("Mode: got %q", v1.Mode)
	}
	if v1.Goal == "" {
		t.Error("Goal: unexpectedly empty")
	}
	if v1.RepoRoot == "" {
		t.Error("RepoRoot: unexpectedly empty")
	}
	if v1.OutputDir == "" {
		t.Error("OutputDir: unexpectedly empty")
	}
	if v1.Status != "done" {
		t.Errorf("Status: got %q, want done", v1.Status)
	}
	if v1.StartedAt == "" {
		t.Error("StartedAt: unexpectedly empty")
	}
	if v1.FinishedAt == "" {
		t.Error("FinishedAt: unexpectedly empty")
	}
	if v1.Duration == "" {
		t.Error("Duration: unexpectedly empty")
	}
	if v1.Runtime.KeepAwakeMode != "caffeinate" {
		t.Errorf("Runtime.KeepAwakeMode: got %q, want caffeinate", v1.Runtime.KeepAwakeMode)
	}
	if v1.Runtime.RequestedTimeout != "2h" {
		t.Errorf("Runtime.RequestedTimeout: got %q, want 2h", v1.Runtime.RequestedTimeout)
	}
	if v1.Runtime.ProcessContractDoc == "" {
		t.Error("Runtime.ProcessContractDoc: unexpectedly empty")
	}
	if len(v1.Steps) != 2 {
		t.Fatalf("Steps: got %d, want 2", len(v1.Steps))
	}
	if v1.Steps[0].Name != "close-loop" {
		t.Errorf("Steps[0].Name: got %q", v1.Steps[0].Name)
	}
	if v1.Steps[0].Status != "done" {
		t.Errorf("Steps[0].Status: got %q", v1.Steps[0].Status)
	}
	if len(v1.Artifacts) != 1 {
		t.Errorf("Artifacts: got %d entries, want 1", len(v1.Artifacts))
	}
	if len(v1.Degraded) == 0 {
		t.Error("Degraded: unexpectedly empty")
	}
	if len(v1.Recommended) == 0 {
		t.Error("Recommended: unexpectedly empty")
	}
	if v1.NextAction == "" {
		t.Error("NextAction: unexpectedly empty")
	}
}
