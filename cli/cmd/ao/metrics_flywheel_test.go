package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/types"
	"github.com/spf13/cobra"
)

func TestPrintFlywheelStatus_Compounding(t *testing.T) {
	var buf bytes.Buffer
	m := &types.FlywheelMetrics{
		Timestamp:           time.Now(),
		PeriodStart:         time.Now().AddDate(0, 0, -7),
		PeriodEnd:           time.Now(),
		Delta:               0.17,
		Sigma:               0.8,
		Rho:                 1.5,
		SigmaRho:            1.2,
		Velocity:            1.03,
		AboveEscapeVelocity: true,
		TierCounts:          map[string]int{},
	}

	oldDays := metricsDays
	metricsDays = 7
	defer func() { metricsDays = oldDays }()

	printFlywheelStatus(&buf, m)

	got := buf.String()
	if !strings.Contains(got, "[COMPOUNDING]") {
		t.Errorf("expected [COMPOUNDING] in output, got: %q", got)
	}
	if !strings.Contains(got, "Knowledge is COMPOUNDING") {
		t.Errorf("expected 'Knowledge is COMPOUNDING' in output, got: %q", got)
	}
}

func TestPrintFlywheelStatus_NearEscape(t *testing.T) {
	var buf bytes.Buffer
	m := &types.FlywheelMetrics{
		Timestamp:           time.Now(),
		PeriodStart:         time.Now().AddDate(0, 0, -7),
		PeriodEnd:           time.Now(),
		Delta:               0.17,
		Sigma:               0.3,
		Rho:                 0.5,
		SigmaRho:            0.15,
		Velocity:            -0.02,
		AboveEscapeVelocity: false,
		TierCounts:          map[string]int{},
	}

	oldDays := metricsDays
	metricsDays = 7
	defer func() { metricsDays = oldDays }()

	printFlywheelStatus(&buf, m)

	got := buf.String()
	if !strings.Contains(got, "[NEAR_ESCAPE]") {
		t.Errorf("expected [NEAR_ESCAPE] in output, got: %q", got)
	}
	if !strings.Contains(got, "NEAR escape velocity") {
		t.Errorf("expected 'NEAR escape velocity' in output, got: %q", got)
	}
}

func TestPrintFlywheelStatus_Decaying(t *testing.T) {
	var buf bytes.Buffer
	m := &types.FlywheelMetrics{
		Timestamp:           time.Now(),
		PeriodStart:         time.Now().AddDate(0, 0, -7),
		PeriodEnd:           time.Now(),
		Delta:               0.17,
		Sigma:               0.1,
		Rho:                 0.2,
		SigmaRho:            0.02,
		Velocity:            -0.15,
		AboveEscapeVelocity: false,
		StaleArtifacts:      10,
		TierCounts:          map[string]int{},
	}

	oldDays := metricsDays
	metricsDays = 7
	defer func() { metricsDays = oldDays }()

	printFlywheelStatus(&buf, m)

	got := buf.String()
	if !strings.Contains(got, "[DECAYING]") {
		t.Errorf("expected [DECAYING] in output, got: %q", got)
	}
	if !strings.Contains(got, "Knowledge is DECAYING") {
		t.Errorf("expected 'Knowledge is DECAYING' in output, got: %q", got)
	}
	if !strings.Contains(got, "RECOMMENDATIONS:") {
		t.Errorf("expected RECOMMENDATIONS section, got: %q", got)
	}
}

func TestPrintFlywheelStatus_Recommendations(t *testing.T) {
	// Recommendations only appear in the DECAYING case (velocity <= -0.05)
	tests := []struct {
		name    string
		sigma   float64
		rho     float64
		stale   int
		wantRec string
	}{
		{"low sigma", 0.1, 0.1, 0, "Improve retrieval"},
		{"low rho", 0.5, 0.05, 0, "Cite more learnings"},
		{"many stale", 0.1, 0.1, 10, "Review 10 stale artifacts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			sigmaRho := tt.sigma * tt.rho
			m := &types.FlywheelMetrics{
				Timestamp:           time.Now(),
				PeriodStart:         time.Now().AddDate(0, 0, -7),
				PeriodEnd:           time.Now(),
				Delta:               0.17,
				Sigma:               tt.sigma,
				Rho:                 tt.rho,
				SigmaRho:            sigmaRho,
				Velocity:            sigmaRho - 0.17, // will be <= -0.05 for these cases
				AboveEscapeVelocity: false,
				StaleArtifacts:      tt.stale,
				TierCounts:          map[string]int{},
			}

			oldDays := metricsDays
			metricsDays = 7
			defer func() { metricsDays = oldDays }()

			printFlywheelStatus(&buf, m)

			if !strings.Contains(buf.String(), tt.wantRec) {
				t.Errorf("expected recommendation %q in output, got: %q", tt.wantRec, buf.String())
			}
		})
	}
}

func TestFlywheelStatus_GoldenSignalsAlwaysShown(t *testing.T) {
	dir := t.TempDir()
	// Create minimal structure for golden signals
	for _, rel := range []string{
		filepath.Join(".agents", "ao"),
		filepath.Join(".agents", "learnings"),
		filepath.Join(".agents", "research"),
	} {
		if err := os.MkdirAll(filepath.Join(dir, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Write a few learnings
	os.WriteFile(filepath.Join(dir, ".agents", "learnings", "l1.md"), []byte("# L"), 0644)
	os.WriteFile(filepath.Join(dir, ".agents", "research", "r1.md"), []byte("# R"), 0644)
	// Write empty citation/feedback files
	os.WriteFile(filepath.Join(dir, ".agents", "ao", "citations.jsonl"), []byte{}, 0644)
	os.WriteFile(filepath.Join(dir, ".agents", "ao", "feedback.jsonl"), []byte{}, 0644)

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	oldOutput := output
	defer func() { output = oldOutput }()

	oldDays := metricsDays
	metricsDays = 7
	defer func() { metricsDays = oldDays }()

	// Golden signals should appear in JSON output WITHOUT --golden flag
	output = "json"
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := runFlywheelStatus(cmd, nil); err != nil {
		t.Fatalf("runFlywheelStatus failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("expected valid JSON, got: %q (%v)", buf.String(), err)
	}

	// Verify golden_signals field exists and has expected structure
	gs, ok := parsed["golden_signals"]
	if !ok || gs == nil {
		t.Fatal("expected golden_signals in JSON output by default (without --golden flag)")
	}
	gsMap, ok := gs.(map[string]any)
	if !ok {
		t.Fatal("expected golden_signals to be an object")
	}
	for _, field := range []string{"overall_verdict", "trend_verdict", "pipeline_verdict", "closure_verdict", "concentration_verdict"} {
		if _, ok := gsMap[field]; !ok {
			t.Errorf("expected field %q in golden_signals", field)
		}
	}

	// Golden signals should appear in table output WITHOUT --golden flag
	buf.Reset()
	output = "table"

	cmd2 := &cobra.Command{}
	cmd2.SetOut(&buf)

	if err := runFlywheelStatus(cmd2, nil); err != nil {
		t.Fatalf("runFlywheelStatus table failed: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "GOLDEN SIGNALS") {
		t.Errorf("expected GOLDEN SIGNALS in default table output, got: %q", got)
	}
	if !strings.Contains(got, "OVERALL:") {
		t.Errorf("expected OVERALL verdict in default table output, got: %q", got)
	}
}

func TestRunFlywheelStatus_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	for _, rel := range []string{
		filepath.Join(".agents", "findings"),
		filepath.Join(".agents", "planning-rules"),
		filepath.Join(".agents", "pre-mortem-checks"),
		filepath.Join(".agents", "rpi"),
	} {
		if err := os.MkdirAll(filepath.Join(dir, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, ".agents", "findings", "f-1.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agents", "planning-rules", "f-1.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agents", "pre-mortem-checks", "f-1.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	queue := `{"source_epic":"ag-h83","timestamp":"2026-03-11T17:00:00Z","items":[{"title":"High one","type":"task","severity":"high","source":"council-finding","description":"d1","target_repo":"agentops","consumed":false}],"consumed":false,"claim_status":"available","claimed_by":null,"claimed_at":null,"consumed_by":null,"consumed_at":null}
`
	if err := os.WriteFile(filepath.Join(dir, ".agents", "rpi", "next-work.jsonl"), []byte(queue), 0o644); err != nil {
		t.Fatal(err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	oldOutput := output
	output = "json"
	defer func() { output = oldOutput }()

	oldDays := metricsDays
	metricsDays = 7
	defer func() { metricsDays = oldDays }()

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := runFlywheelStatus(cmd, nil); err != nil {
		t.Fatalf("runFlywheelStatus failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("expected valid JSON, got: %q (%v)", buf.String(), err)
	}

	// Verify expected fields
	for _, field := range []string{"status", "delta", "sigma", "rho", "sigma_rho", "velocity", "compounding", "scorecard", "metrics"} {
		if _, ok := parsed[field]; !ok {
			t.Errorf("expected field %q in JSON output", field)
		}
	}
}

func TestPrintFlywheelStatus_IncludesScorecard(t *testing.T) {
	var buf bytes.Buffer
	m := &types.FlywheelMetrics{
		Timestamp:   time.Now(),
		PeriodStart: time.Now().AddDate(0, 0, -7),
		PeriodEnd:   time.Now(),
		TierCounts:  map[string]int{},
		StigmergicScorecard: &types.StigmergicScorecard{
			PromotedFindings:       3,
			PlanningRules:          3,
			PreMortemChecks:        2,
			UnconsumedItems:        9,
			HighSeverityUnconsumed: 4,
			UnconsumedBatches:      2,
		},
	}

	oldDays := metricsDays
	metricsDays = 7
	defer func() { metricsDays = oldDays }()

	printFlywheelStatus(&buf, m)

	got := buf.String()
	if !strings.Contains(got, "STIGMERGIC SCORECARD:") {
		t.Fatalf("expected scorecard section, got: %q", got)
	}
	if !strings.Contains(got, "Backlog: 9 items, 4 high severity, 2 batches") {
		t.Fatalf("expected backlog line, got: %q", got)
	}
}

func TestRunFlywheelStatus_YAMLOutput(t *testing.T) {
	dir := t.TempDir()

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	oldOutput := output
	output = "yaml"
	defer func() { output = oldOutput }()

	oldDays := metricsDays
	metricsDays = 7
	defer func() { metricsDays = oldDays }()

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := runFlywheelStatus(cmd, nil); err != nil {
		t.Fatalf("runFlywheelStatus failed: %v", err)
	}

	got := buf.String()
	// YAML output should contain key fields
	if !strings.Contains(got, "status:") {
		t.Errorf("expected 'status:' in YAML output, got: %q", got)
	}
	if !strings.Contains(got, "delta:") {
		t.Errorf("expected 'delta:' in YAML output, got: %q", got)
	}
}

func TestRunFlywheelStatus_TableOutput(t *testing.T) {
	dir := t.TempDir()

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	oldOutput := output
	output = "table"
	defer func() { output = oldOutput }()

	oldDays := metricsDays
	metricsDays = 7
	defer func() { metricsDays = oldDays }()

	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := runFlywheelStatus(cmd, nil); err != nil {
		t.Fatalf("runFlywheelStatus failed: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "Flywheel Status:") {
		t.Errorf("expected 'Flywheel Status:' in output, got: %q", got)
	}
	if !strings.Contains(got, "EQUATION:") {
		t.Errorf("expected 'EQUATION:' in output, got: %q", got)
	}
	if !strings.Contains(got, "ESCAPE VELOCITY CHECK:") {
		t.Errorf("expected 'ESCAPE VELOCITY CHECK:' in output, got: %q", got)
	}
}

func TestPrintFlywheelStatus_ContainsEquation(t *testing.T) {
	var buf bytes.Buffer
	m := &types.FlywheelMetrics{
		Timestamp:   time.Now(),
		PeriodStart: time.Now().AddDate(0, 0, -7),
		PeriodEnd:   time.Now(),
		TierCounts:  map[string]int{},
	}

	oldDays := metricsDays
	metricsDays = 7
	defer func() { metricsDays = oldDays }()

	printFlywheelStatus(&buf, m)

	got := buf.String()
	if !strings.Contains(got, "dK/dt") {
		t.Errorf("expected flywheel equation in output, got: %q", got)
	}
}

func TestPrintFlywheelStatus_ShowsPeriod(t *testing.T) {
	var buf bytes.Buffer
	now := time.Now()
	start := now.AddDate(0, 0, -14)
	m := &types.FlywheelMetrics{
		Timestamp:   now,
		PeriodStart: start,
		PeriodEnd:   now,
		TierCounts:  map[string]int{},
	}

	oldDays := metricsDays
	metricsDays = 14
	defer func() { metricsDays = oldDays }()

	printFlywheelStatus(&buf, m)

	got := buf.String()
	if !strings.Contains(got, start.Format("2006-01-02")) {
		t.Errorf("expected period start in output, got: %q", got)
	}
	if !strings.Contains(got, now.Format("2006-01-02")) {
		t.Errorf("expected period end in output, got: %q", got)
	}
	if !strings.Contains(got, "14 days") {
		t.Errorf("expected '14 days' in output, got: %q", got)
	}
}
