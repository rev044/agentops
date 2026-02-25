package main

import (
	"testing"
	"time"

	"github.com/boshu2/agentops/cli/internal/vibecheck"
)

// ===========================================================================
// vibe_check.go — outputVibeCheckJSON (zero coverage)
// ===========================================================================

func TestCov3_vibeCheck_outputVibeCheckJSON_emptyResult(t *testing.T) {
	result := &vibecheck.VibeCheckResult{
		Score:    85.0,
		Grade:    "B",
		Metrics:  map[string]float64{},
		Findings: []vibecheck.Finding{},
		Events:   []vibecheck.TimelineEvent{},
	}
	err := outputVibeCheckJSON(result)
	if err != nil {
		t.Fatalf("outputVibeCheckJSON: %v", err)
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
	err := outputVibeCheckJSON(result)
	if err != nil {
		t.Fatalf("outputVibeCheckJSON: %v", err)
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
	err := outputVibeCheckMarkdown(result)
	if err != nil {
		t.Fatalf("outputVibeCheckMarkdown: %v", err)
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
	err := outputVibeCheckMarkdown(result)
	if err != nil {
		t.Fatalf("outputVibeCheckMarkdown: %v", err)
	}
}

// ===========================================================================
// vibe_check.go — printMarkdownMetrics (zero coverage)
// ===========================================================================

func TestCov3_vibeCheck_printMarkdownMetrics_empty(t *testing.T) {
	// Should not panic
	printMarkdownMetrics(map[string]float64{})
}

func TestCov3_vibeCheck_printMarkdownMetrics_withValues(t *testing.T) {
	printMarkdownMetrics(map[string]float64{
		"velocity":   1.5,
		"complexity": 0.3,
		"rework":     0.1,
	})
}

func TestCov3_vibeCheck_printMarkdownMetrics_nil(t *testing.T) {
	printMarkdownMetrics(nil)
}

// ===========================================================================
// vibe_check.go — printMarkdownFindings (zero coverage)
// ===========================================================================

func TestCov3_vibeCheck_printMarkdownFindings_empty(t *testing.T) {
	printMarkdownFindings([]vibecheck.Finding{})
}

func TestCov3_vibeCheck_printMarkdownFindings_withFindings(t *testing.T) {
	findings := []vibecheck.Finding{
		{Severity: "error", Category: "amnesia", Message: "critical issue"},
		{Severity: "warning", Category: "drift", Message: "minor issue", File: "cmd/main.go"},
		{Severity: "info", Category: "test", Message: "informational", File: "pkg/util.go", Line: 10},
	}
	printMarkdownFindings(findings)
}

func TestCov3_vibeCheck_printMarkdownFindings_nil(t *testing.T) {
	printMarkdownFindings(nil)
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
	// Should not panic
	printMarkdownFinding(finding)
}

func TestCov3_vibeCheck_printMarkdownFinding_withFileNoLine(t *testing.T) {
	finding := vibecheck.Finding{
		Severity: "warning",
		Category: "drift",
		Message:  "code drifting",
		File:     "main.go",
	}
	printMarkdownFinding(finding)
}

func TestCov3_vibeCheck_printMarkdownFinding_noFile(t *testing.T) {
	finding := vibecheck.Finding{
		Severity: "info",
		Category: "logging",
		Message:  "missing structured logs",
	}
	printMarkdownFinding(finding)
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
	// Should not print anything (returns early)
	printMarkdownEvents(events)
}

func TestCov3_vibeCheck_printMarkdownEvents_fullWithEvents(t *testing.T) {
	origFull := vibeCheckFull
	defer func() { vibeCheckFull = origFull }()
	vibeCheckFull = true

	events := []vibecheck.TimelineEvent{
		{Timestamp: time.Now(), Author: "dev1", Message: "feat: add feature"},
		{Timestamp: time.Now(), Author: "dev2", Message: "fix: fix bug with a very long message that exceeds fifty characters and should be truncated"},
	}
	printMarkdownEvents(events)
}

func TestCov3_vibeCheck_printMarkdownEvents_fullEmpty(t *testing.T) {
	origFull := vibeCheckFull
	defer func() { vibeCheckFull = origFull }()
	vibeCheckFull = true

	// Empty events should return early
	printMarkdownEvents([]vibecheck.TimelineEvent{})
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

	err := outputVibeCheckTable(result)
	if err != nil {
		t.Fatalf("outputVibeCheckTable: %v", err)
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

	err := outputVibeCheckTable(result)
	if err != nil {
		t.Fatalf("outputVibeCheckTable: %v", err)
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

	err := outputVibeCheckTable(result)
	if err != nil {
		t.Fatalf("outputVibeCheckTable: %v", err)
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

	err := outputVibeCheckTable(result)
	if err != nil {
		t.Fatalf("outputVibeCheckTable: %v", err)
	}
}
