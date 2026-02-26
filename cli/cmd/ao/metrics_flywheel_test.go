package main

import (
	"bytes"
	"encoding/json"
	"os"
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

func TestRunFlywheelStatus_JSONOutput(t *testing.T) {
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
	for _, field := range []string{"status", "delta", "sigma", "rho", "sigma_rho", "velocity", "compounding", "metrics"} {
		if _, ok := parsed[field]; !ok {
			t.Errorf("expected field %q in JSON output", field)
		}
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
