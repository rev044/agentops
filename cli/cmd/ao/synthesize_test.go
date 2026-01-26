package main

import (
	"testing"

	"github.com/boshu2/agentops/cli/internal/orchestrator"
)

func TestCalculateVerdict(t *testing.T) {
	tests := []struct {
		name     string
		findings []orchestrator.Finding
		want     orchestrator.Severity
	}{
		{
			name:     "empty findings returns PASS",
			findings: nil,
			want:     orchestrator.SeverityPass,
		},
		{
			name: "single LOW returns LOW",
			findings: []orchestrator.Finding{
				{Severity: orchestrator.SeverityLow},
			},
			want: orchestrator.SeverityLow,
		},
		{
			name: "single CRITICAL returns CRITICAL (veto)",
			findings: []orchestrator.Finding{
				{Severity: orchestrator.SeverityCritical},
			},
			want: orchestrator.SeverityCritical,
		},
		{
			name: "CRITICAL among others triggers veto",
			findings: []orchestrator.Finding{
				{Severity: orchestrator.SeverityLow},
				{Severity: orchestrator.SeverityMedium},
				{Severity: orchestrator.SeverityCritical},
				{Severity: orchestrator.SeverityHigh},
			},
			want: orchestrator.SeverityCritical,
		},
		{
			name: "HIGH is highest without CRITICAL",
			findings: []orchestrator.Finding{
				{Severity: orchestrator.SeverityLow},
				{Severity: orchestrator.SeverityMedium},
				{Severity: orchestrator.SeverityHigh},
			},
			want: orchestrator.SeverityHigh,
		},
		{
			name: "MEDIUM is highest without HIGH/CRITICAL",
			findings: []orchestrator.Finding{
				{Severity: orchestrator.SeverityLow},
				{Severity: orchestrator.SeverityMedium},
			},
			want: orchestrator.SeverityMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateVerdict(tt.findings)
			if got != tt.want {
				t.Errorf("calculateVerdict() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestVerdictToGrade(t *testing.T) {
	tests := []struct {
		verdict       orchestrator.Severity
		criticalCount int
		highCount     int
		want          string
	}{
		{orchestrator.SeverityPass, 0, 0, "A"},
		{orchestrator.SeverityLow, 0, 0, "A-"},
		{orchestrator.SeverityMedium, 0, 0, "B"},
		{orchestrator.SeverityHigh, 0, 1, "C"},
		{orchestrator.SeverityHigh, 0, 4, "D"}, // more than 3 HIGH
		{orchestrator.SeverityCritical, 1, 0, "D"},
		{orchestrator.SeverityCritical, 2, 0, "F"}, // more than 1 CRITICAL
	}

	for _, tt := range tests {
		got := verdictToGrade(tt.verdict, tt.criticalCount, tt.highCount)
		if got != tt.want {
			t.Errorf("verdictToGrade(%s, %d, %d) = %s, want %s",
				tt.verdict, tt.criticalCount, tt.highCount, got, tt.want)
		}
	}
}

func TestDeduplicateFindings(t *testing.T) {
	findings := []orchestrator.Finding{
		{ID: "1", Category: "security", Title: "SQL Injection", Severity: orchestrator.SeverityHigh},
		{ID: "2", Category: "security", Title: "sql injection", Severity: orchestrator.SeverityCritical}, // Duplicate, higher severity
		{ID: "3", Category: "quality", Title: "Long function", Severity: orchestrator.SeverityMedium},
	}

	result := deduplicateFindings(findings)

	if len(result) != 2 {
		t.Errorf("expected 2 deduplicated findings, got %d", len(result))
	}

	// The SQL injection finding should have CRITICAL severity (highest)
	for _, f := range result {
		if f.Category == "security" && f.Severity != orchestrator.SeverityCritical {
			t.Errorf("merged security finding should be CRITICAL, got %s", f.Severity)
		}
	}
}

func TestDeduplicateFindingsEmpty(t *testing.T) {
	result := deduplicateFindings(nil)
	if result != nil {
		t.Errorf("deduplicating nil should return nil, got %v", result)
	}

	result = deduplicateFindings([]orchestrator.Finding{})
	if len(result) != 0 {
		t.Errorf("deduplicating empty slice should return empty, got %v", result)
	}
}

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"SQL Injection", "sql injection"},
		{"  SQL Injection  ", "sql injection"},
		{"UPPERCASE", "uppercase"},
		{"", ""},
	}

	for _, tt := range tests {
		got := normalizeTitle(tt.input)
		if got != tt.want {
			t.Errorf("normalizeTitle(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMergeStringSlices(t *testing.T) {
	a := []string{"file1.go", "file2.go"}
	b := []string{"file2.go", "file3.go"}

	result := mergeStringSlices(a, b)

	if len(result) != 3 {
		t.Errorf("expected 3 unique files, got %d", len(result))
	}

	// Check all expected files are present
	expected := map[string]bool{"file1.go": false, "file2.go": false, "file3.go": false}
	for _, f := range result {
		if _, ok := expected[f]; ok {
			expected[f] = true
		}
	}

	for f, found := range expected {
		if !found {
			t.Errorf("expected file %s not found in result", f)
		}
	}
}

func TestSynthesizeFindings(t *testing.T) {
	reports := []AgentFindings{
		{
			Category: "security",
			Findings: []orchestrator.Finding{
				{ID: "1", Severity: orchestrator.SeverityHigh, Title: "XSS", Category: "security"},
			},
			Summary: "Found XSS vulnerability",
		},
		{
			Category: "quality",
			Findings: []orchestrator.Finding{
				{ID: "2", Severity: orchestrator.SeverityMedium, Title: "Long function", Category: "quality"},
			},
			Summary: "Code quality issues",
		},
	}

	result := synthesizeFindings("test-plan", reports)

	if result.PlanID != "test-plan" {
		t.Errorf("expected plan ID test-plan, got %s", result.PlanID)
	}

	if result.Verdict != orchestrator.SeverityHigh {
		t.Errorf("expected verdict HIGH, got %s", result.Verdict)
	}

	if len(result.Findings) != 2 {
		t.Errorf("expected 2 findings, got %d", len(result.Findings))
	}

	if result.HighCount != 1 {
		t.Errorf("expected 1 HIGH count, got %d", result.HighCount)
	}

	if result.MediumCount != 1 {
		t.Errorf("expected 1 MEDIUM count, got %d", result.MediumCount)
	}
}

func TestSynthesizeFindingsEmpty(t *testing.T) {
	result := synthesizeFindings("empty-plan", []AgentFindings{})

	if result.Verdict != orchestrator.SeverityPass {
		t.Errorf("expected PASS verdict for empty findings, got %s", result.Verdict)
	}

	if result.Grade != "A" {
		t.Errorf("expected grade A for empty findings, got %s", result.Grade)
	}
}

func TestSynthesizeSummary(t *testing.T) {
	tests := []struct {
		name     string
		findings []orchestrator.Finding
		verdict  orchestrator.Severity
		critical int
		high     int
		medium   int
		low      int
		wantPart string
	}{
		{
			name:     "no findings",
			findings: nil,
			verdict:  orchestrator.SeverityPass,
			wantPart: "No issues found",
		},
		{
			name: "with critical",
			findings: []orchestrator.Finding{
				{Severity: orchestrator.SeverityCritical},
			},
			verdict:  orchestrator.SeverityCritical,
			critical: 1,
			wantPart: "1 CRITICAL",
		},
		{
			name: "mixed severities",
			findings: []orchestrator.Finding{
				{Severity: orchestrator.SeverityHigh},
				{Severity: orchestrator.SeverityMedium},
			},
			verdict:  orchestrator.SeverityHigh,
			high:     1,
			medium:   1,
			wantPart: "1 HIGH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := synthesizeSummary(tt.findings, tt.verdict, tt.critical, tt.high, tt.medium, tt.low)
			if len(tt.wantPart) > 0 && !contains(got, tt.wantPart) {
				t.Errorf("summary should contain %q, got %q", tt.wantPart, got)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
