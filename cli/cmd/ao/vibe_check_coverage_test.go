package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/vibecheck"
)

// ===========================================================================
// vibe_check.go — outputVibeCheckJSON (zero coverage)
// ===========================================================================

// captureVibeStdout captures stdout output from a function for assertion.
func captureVibeStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = io.ReadAll(io.TeeReader(r, &buf))
	return buf.String()
}

func TestCov3_vibeCheck_outputVibeCheckJSON_emptyResult(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score:    85.0,
		Grade:    "B",
		Metrics:  map[string]float64{},
		Findings: []vibecheck.Finding{},
		Events:   []vibecheck.TimelineEvent{},
	}
	out, err := captureStdout(t, func() error {
		return outputVibeCheckJSON(result)
	})
	if err != nil {
		t.Fatalf("outputVibeCheckJSON: %v", err)
	}
	if !strings.Contains(out, "85") {
		t.Errorf("expected JSON output to contain score 85, got: %s", out)
	}
}

func TestCov3_vibeCheck_outputVibeCheckJSON_withData(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score: 92.5,
		Grade: "A",
		Metrics: map[string]float64{
			"velocity":   1.5,
			"complexity": 0.3,
		},
		Findings: []vibecheck.Finding{
			{Severity: "warning", Category: "drift", Message: "test warning"},
		},
		Events: []vibecheck.TimelineEvent{
			{
				Timestamp: time.Now(),
				Author:    "test",
				Message:   "initial commit",
			},
		},
	}
	out, err := captureStdout(t, func() error {
		return outputVibeCheckJSON(result)
	})
	if err != nil {
		t.Fatalf("outputVibeCheckJSON: %v", err)
	}
	if !strings.Contains(out, "92.5") {
		t.Errorf("expected JSON to contain score 92.5, got: %s", out)
	}
	if !strings.Contains(out, "drift") {
		t.Errorf("expected JSON to contain finding category 'drift', got: %s", out)
	}
}

// ===========================================================================
// vibe_check.go — outputVibeCheckMarkdown (zero coverage)
// ===========================================================================

func TestCov3_vibeCheck_outputVibeCheckMarkdown_emptyResult(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score:    50.0,
		Grade:    "C",
		Metrics:  map[string]float64{},
		Findings: []vibecheck.Finding{},
		Events:   []vibecheck.TimelineEvent{},
	}
	out, err := captureStdout(t, func() error {
		return outputVibeCheckMarkdown(result)
	})
	if err != nil {
		t.Fatalf("outputVibeCheckMarkdown: %v", err)
	}
	if !strings.Contains(out, "50") {
		t.Errorf("expected markdown to contain score 50, got: %s", out)
	}
}

func TestCov3_vibeCheck_outputVibeCheckMarkdown_withData(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score: 72.0,
		Grade: "B-",
		Metrics: map[string]float64{
			"velocity":   1.2,
			"complexity": 0.5,
		},
		Findings: []vibecheck.Finding{
			{Severity: "error", Category: "amnesia", Message: "repeating mistakes"},
			{Severity: "info", Category: "drift", Message: "minor drift detected", File: "main.go", Line: 42},
		},
		Events: []vibecheck.TimelineEvent{
			{
				Timestamp: time.Now(),
				Author:    "dev",
				Message:   "fix: something",
			},
		},
	}
	out, err := captureStdout(t, func() error {
		return outputVibeCheckMarkdown(result)
	})
	if err != nil {
		t.Fatalf("outputVibeCheckMarkdown: %v", err)
	}
	if !strings.Contains(out, "72") {
		t.Errorf("expected markdown to contain score 72, got: %s", out)
	}
	if !strings.Contains(out, "amnesia") {
		t.Errorf("expected markdown to contain finding category 'amnesia', got: %s", out)
	}
}

// ===========================================================================
// vibe_check.go — printMarkdownMetrics (zero coverage)
// ===========================================================================

func TestCov3_vibeCheck_printMarkdownMetrics_empty(t *testing.T) {
	out := captureVibeStdout(t, func() {
		printMarkdownMetrics(map[string]float64{})
	})
	// Empty metrics should produce minimal or no output, but must not panic
	if strings.Contains(out, "panic") {
		t.Errorf("printMarkdownMetrics panicked on empty map")
	}
}

func TestCov3_vibeCheck_printMarkdownMetrics_withValues(t *testing.T) {
	out := captureVibeStdout(t, func() {
		printMarkdownMetrics(map[string]float64{
			"velocity":   1.5,
			"complexity": 0.3,
			"rework":     0.1,
		})
	})
	if !strings.Contains(out, "velocity") {
		t.Errorf("expected output to contain 'velocity', got: %s", out)
	}
	if !strings.Contains(out, "complexity") {
		t.Errorf("expected output to contain 'complexity', got: %s", out)
	}
}

func TestCov3_vibeCheck_printMarkdownMetrics_nil(t *testing.T) {
	out := captureVibeStdout(t, func() {
		printMarkdownMetrics(nil)
	})
	// nil metrics should produce minimal or no output, but must not panic
	if strings.Contains(out, "panic") {
		t.Errorf("printMarkdownMetrics panicked on nil map")
	}
}

// ===========================================================================
// vibe_check.go — printMarkdownFindings (zero coverage)
// ===========================================================================

func TestCov3_vibeCheck_printMarkdownFindings_empty(t *testing.T) {
	out := captureVibeStdout(t, func() {
		printMarkdownFindings([]vibecheck.Finding{})
	})
	// Empty findings should not produce finding-specific output
	if strings.Contains(out, "error") || strings.Contains(out, "amnesia") {
		t.Errorf("expected no finding output for empty list, got: %s", out)
	}
}

func TestCov3_vibeCheck_printMarkdownFindings_withFindings(t *testing.T) {
	findings := []vibecheck.Finding{
		{Severity: "error", Category: "amnesia", Message: "critical issue"},
		{Severity: "warning", Category: "drift", Message: "minor issue", File: "cmd/main.go"},
		{Severity: "info", Category: "test", Message: "informational", File: "pkg/util.go", Line: 10},
	}
	out := captureVibeStdout(t, func() {
		printMarkdownFindings(findings)
	})
	if !strings.Contains(out, "critical issue") {
		t.Errorf("expected output to contain 'critical issue', got: %s", out)
	}
	if !strings.Contains(out, "minor issue") {
		t.Errorf("expected output to contain 'minor issue', got: %s", out)
	}
}

func TestCov3_vibeCheck_printMarkdownFindings_nil(t *testing.T) {
	out := captureVibeStdout(t, func() {
		printMarkdownFindings(nil)
	})
	// nil findings should produce minimal or no output, but must not panic
	if strings.Contains(out, "panic") {
		t.Errorf("printMarkdownFindings panicked on nil")
	}
}

// ===========================================================================
// vibe_check.go — severityEmoji (zero coverage)
// ===========================================================================

func TestCov3_vibeCheck_severityEmoji(t *testing.T) {
	tests := []struct {
		severity string
		want     string
	}{
		{"error", "❌"},
		{"info", "ℹ️"},
		{"warning", "⚠️"},
		{"unknown", "⚠️"},
		{"", "⚠️"},
	}
	for _, tc := range tests {
		t.Run(tc.severity, func(t *testing.T) {
			got := severityEmoji(tc.severity)
			if got != tc.want {
				t.Errorf("severityEmoji(%q) = %q, want %q", tc.severity, got, tc.want)
			}
		})
	}
}

// ===========================================================================
// vibe_check.go — printMarkdownFinding (zero coverage)
// ===========================================================================

func TestCov3_vibeCheck_printMarkdownFinding_withFileLine(t *testing.T) {
	finding := vibecheck.Finding{
		Severity: "error",
		Category: "test-lies",
		Message:  "test assertions are empty",
		File:     "pkg/handler.go",
		Line:     42,
	}
	out := captureVibeStdout(t, func() {
		printMarkdownFinding(finding)
	})
	if !strings.Contains(out, "test assertions are empty") {
		t.Errorf("expected output to contain finding message, got: %s", out)
	}
	if !strings.Contains(out, "pkg/handler.go") {
		t.Errorf("expected output to contain file path, got: %s", out)
	}
}

func TestCov3_vibeCheck_printMarkdownFinding_withFileNoLine(t *testing.T) {
	finding := vibecheck.Finding{
		Severity: "warning",
		Category: "drift",
		Message:  "code drifting",
		File:     "main.go",
	}
	out := captureVibeStdout(t, func() {
		printMarkdownFinding(finding)
	})
	if !strings.Contains(out, "code drifting") {
		t.Errorf("expected output to contain finding message, got: %s", out)
	}
	if !strings.Contains(out, "main.go") {
		t.Errorf("expected output to contain file path, got: %s", out)
	}
}

func TestCov3_vibeCheck_printMarkdownFinding_noFile(t *testing.T) {
	finding := vibecheck.Finding{
		Severity: "info",
		Category: "logging",
		Message:  "missing structured logs",
	}
	out := captureVibeStdout(t, func() {
		printMarkdownFinding(finding)
	})
	if !strings.Contains(out, "missing structured logs") {
		t.Errorf("expected output to contain finding message, got: %s", out)
	}
}

// ===========================================================================
// vibe_check.go — printMarkdownEvents (zero coverage)
// ===========================================================================

func TestCov3_vibeCheck_printMarkdownEvents_notFull(t *testing.T) {
	// vibeCheckFull is false by default, so events should not be printed
	origFull := vibeCheckFull
	defer func() { vibeCheckFull = origFull }()
	vibeCheckFull = false

	events := []vibecheck.TimelineEvent{
		{Timestamp: time.Now(), Author: "dev", Message: "commit 1"},
	}
	out := captureVibeStdout(t, func() {
		printMarkdownEvents(events)
	})
	// Not full mode — should produce no output
	if strings.Contains(out, "commit 1") {
		t.Errorf("expected no event output in non-full mode, got: %s", out)
	}
}

func TestCov3_vibeCheck_printMarkdownEvents_fullWithEvents(t *testing.T) {
	origFull := vibeCheckFull
	defer func() { vibeCheckFull = origFull }()
	vibeCheckFull = true

	events := []vibecheck.TimelineEvent{
		{Timestamp: time.Now(), Author: "dev1", Message: "feat: add feature"},
		{Timestamp: time.Now(), Author: "dev2", Message: "fix: fix bug with a very long message that exceeds fifty characters and should be truncated"},
	}
	out := captureVibeStdout(t, func() {
		printMarkdownEvents(events)
	})
	if !strings.Contains(out, "dev1") {
		t.Errorf("expected output to contain author 'dev1', got: %s", out)
	}
	if !strings.Contains(out, "feat: add feature") {
		t.Errorf("expected output to contain event message, got: %s", out)
	}
}

func TestCov3_vibeCheck_printMarkdownEvents_fullEmpty(t *testing.T) {
	origFull := vibeCheckFull
	defer func() { vibeCheckFull = origFull }()
	vibeCheckFull = true

	// Empty events should return early
	out := captureVibeStdout(t, func() {
		printMarkdownEvents([]vibecheck.TimelineEvent{})
	})
	// Empty events in full mode should produce no event-specific output
	if strings.Contains(out, "Author") && strings.Contains(out, "Message") {
		t.Errorf("expected no event rows for empty list, got: %s", out)
	}
}

// ===========================================================================
// vibe_check.go — outputVibeCheckTable (zero coverage)
// ===========================================================================

func TestCov3_vibeCheck_outputVibeCheckTable_empty(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score:    0.0,
		Grade:    "F",
		Metrics:  map[string]float64{},
		Findings: []vibecheck.Finding{},
		Events:   []vibecheck.TimelineEvent{},
	}

	origFull := vibeCheckFull
	defer func() { vibeCheckFull = origFull }()
	vibeCheckFull = false

	out, err := captureStdout(t, func() error {
		return outputVibeCheckTable(result)
	})
	if err != nil {
		t.Fatalf("outputVibeCheckTable: %v", err)
	}
	if !strings.Contains(out, "F") {
		t.Errorf("expected table to contain grade 'F', got: %s", out)
	}
}

func TestCov3_vibeCheck_outputVibeCheckTable_withMetrics(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score: 75.0,
		Grade: "B",
		Metrics: map[string]float64{
			"velocity":   1.5,
			"complexity": 0.3,
		},
		Findings: []vibecheck.Finding{
			{Severity: "warning", Category: "drift", Message: "minor drift"},
			{Severity: "error", Category: "amnesia", Message: "repeating", File: "main.go", Line: 5},
		},
		Events: []vibecheck.TimelineEvent{
			{Timestamp: time.Now(), Author: "test", Message: "commit"},
		},
	}

	origFull := vibeCheckFull
	defer func() { vibeCheckFull = origFull }()
	vibeCheckFull = false

	out, err := captureStdout(t, func() error {
		return outputVibeCheckTable(result)
	})
	if err != nil {
		t.Fatalf("outputVibeCheckTable: %v", err)
	}
	if !strings.Contains(out, "75") {
		t.Errorf("expected table to contain score 75, got: %s", out)
	}
}

func TestCov3_vibeCheck_outputVibeCheckTable_fullMode(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score:    90.0,
		Grade:    "A",
		Metrics:  map[string]float64{"velocity": 2.0},
		Findings: []vibecheck.Finding{},
		Events: []vibecheck.TimelineEvent{
			{Timestamp: time.Now(), Author: "dev", Message: "initial"},
		},
	}

	origFull := vibeCheckFull
	defer func() { vibeCheckFull = origFull }()
	vibeCheckFull = true

	out, err := captureStdout(t, func() error {
		return outputVibeCheckTable(result)
	})
	if err != nil {
		t.Fatalf("outputVibeCheckTable: %v", err)
	}
	if !strings.Contains(out, "90") {
		t.Errorf("expected table to contain score 90, got: %s", out)
	}
}

func TestCov3_vibeCheck_outputVibeCheckTable_findingsNoFile(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score:   60.0,
		Grade:   "D",
		Metrics: nil,
		Findings: []vibecheck.Finding{
			{Severity: "warning", Category: "test", Message: "no file location"},
		},
		Events: nil,
	}

	out, err := captureStdout(t, func() error {
		return outputVibeCheckTable(result)
	})
	if err != nil {
		t.Fatalf("outputVibeCheckTable: %v", err)
	}
	if !strings.Contains(out, "no file location") {
		t.Errorf("expected table to contain finding message, got: %s", out)
	}
}
