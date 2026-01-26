package orchestrator

import (
	"testing"
	"time"
)

func TestNewWaveDispatcher(t *testing.T) {
	config := DefaultDispatchConfig()
	d := NewWaveDispatcher(config)

	if d.Config.PodSize != 6 {
		t.Errorf("expected PodSize 6, got %d", d.Config.PodSize)
	}
	if d.Config.QuorumThreshold != 0.7 {
		t.Errorf("expected QuorumThreshold 0.7, got %f", d.Config.QuorumThreshold)
	}
}

func TestHigherSeverity(t *testing.T) {
	tests := []struct {
		a, b     Severity
		expected Severity
	}{
		{SeverityCritical, SeverityHigh, SeverityCritical},
		{SeverityHigh, SeverityCritical, SeverityCritical},
		{SeverityMedium, SeverityLow, SeverityMedium},
		{SeverityPass, SeverityLow, SeverityLow},
		{SeverityCritical, SeverityCritical, SeverityCritical},
	}

	for _, tt := range tests {
		result := HigherSeverity(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("HigherSeverity(%s, %s) = %s, expected %s",
				tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestConsensusSingleVeto(t *testing.T) {
	c := NewConsensus()

	findings := []Finding{
		{ID: "1", Severity: SeverityMedium},
		{ID: "2", Severity: SeverityHigh},
		{ID: "3", Severity: SeverityLow},
	}

	result := c.ApplySingleVeto(findings)
	if result != SeverityHigh {
		t.Errorf("expected HIGH, got %s", result)
	}

	// Add CRITICAL
	findings = append(findings, Finding{ID: "4", Severity: SeverityCritical})
	result = c.ApplySingleVeto(findings)
	if result != SeverityCritical {
		t.Errorf("expected CRITICAL with veto, got %s", result)
	}
}

func TestConsensusDeduplicate(t *testing.T) {
	c := NewConsensus()

	findings := []Finding{
		{ID: "1", Severity: SeverityHigh, Category: "security", Title: "SQL injection"},
		{ID: "2", Severity: SeverityMedium, Category: "security", Title: "SQL injection"},
		{ID: "3", Severity: SeverityLow, Category: "quality", Title: "Long function"},
	}

	result := c.Deduplicate(findings)

	if len(result) != 2 {
		t.Errorf("expected 2 findings after dedup, got %d", len(result))
	}

	// The security finding should have HIGH severity (highest of duplicates)
	for _, f := range result {
		if f.Category == "security" && f.Severity != SeverityHigh {
			t.Errorf("expected merged security finding to be HIGH, got %s", f.Severity)
		}
	}
}

func TestConsensusFilterByContextBudget(t *testing.T) {
	c := NewConsensus()

	findings := []Finding{
		{ID: "1", Severity: SeverityCritical},
		{ID: "2", Severity: SeverityHigh},
		{ID: "3", Severity: SeverityMedium},
		{ID: "4", Severity: SeverityLow},
	}

	// At 60% usage, drop LOW
	result := c.FilterByContextBudget(findings, 0.6)
	for _, f := range result {
		if f.Severity == SeverityLow {
			t.Error("LOW findings should be filtered at 60% usage")
		}
	}

	// At 80% usage, keep only CRITICAL and HIGH
	result = c.FilterByContextBudget(findings, 0.8)
	for _, f := range result {
		if f.Severity != SeverityCritical && f.Severity != SeverityHigh {
			t.Errorf("only CRITICAL and HIGH should remain at 80%%, got %s", f.Severity)
		}
	}
}

func TestConflictResolverWithinPod(t *testing.T) {
	r := NewConflictResolver()

	agents := []AgentResult{
		{
			AgentID: "agent-1",
			Findings: []Finding{
				{ID: "1", Severity: SeverityHigh, Category: "security", Title: "XSS"},
			},
		},
		{
			AgentID: "agent-2",
			Findings: []Finding{
				{ID: "2", Severity: SeverityCritical, Category: "security", Title: "SQL injection"},
			},
		},
	}

	result := r.ResolveWithinPod(agents)

	if len(result.Findings) != 2 {
		t.Errorf("expected 2 findings, got %d", len(result.Findings))
	}
}

func TestConflictResolverAcrossPods(t *testing.T) {
	r := NewConflictResolver()

	pods := []PodResult{
		{
			Config: PodConfig{Name: "security"},
			Findings: []Finding{
				{ID: "1", Severity: SeverityHigh, Category: "security", Title: "XSS vulnerability"},
			},
		},
		{
			Config: PodConfig{Name: "quality"},
			Findings: []Finding{
				{ID: "2", Severity: SeverityHigh, Category: "security", Title: "XSS vulnerability"},
			},
		},
	}

	result := r.ResolveAcrossPods(pods)

	// Should merge the duplicate XSS findings
	if len(result.MergedFindings) != 1 {
		t.Errorf("expected 1 merged finding, got %d", len(result.MergedFindings))
	}
}

func TestDispatcherCreatePods(t *testing.T) {
	config := DefaultDispatchConfig()
	d := NewWaveDispatcher(config)

	files := []string{"file1.go", "file2.go", "file3.go"}
	pods := d.createPods(files)

	if len(pods) == 0 {
		t.Error("expected at least one pod")
	}

	for _, pod := range pods {
		if len(pod.Files) != len(files) {
			t.Errorf("expected all files in pod, got %d", len(pod.Files))
		}
	}
}

func TestDispatcherCalculateVerdict(t *testing.T) {
	d := NewWaveDispatcher(DefaultDispatchConfig())

	tests := []struct {
		findings []Finding
		expected Severity
	}{
		{[]Finding{}, SeverityPass},
		{[]Finding{{Severity: SeverityLow}}, SeverityLow},
		{[]Finding{{Severity: SeverityMedium}}, SeverityMedium},
		{[]Finding{{Severity: SeverityHigh}}, SeverityHigh},
		{[]Finding{{Severity: SeverityCritical}}, SeverityCritical},
		{[]Finding{
			{Severity: SeverityLow},
			{Severity: SeverityCritical},
			{Severity: SeverityMedium},
		}, SeverityCritical},
	}

	for _, tt := range tests {
		result := d.calculateVerdict(tt.findings)
		if result != tt.expected {
			t.Errorf("calculateVerdict() = %s, expected %s", result, tt.expected)
		}
	}
}

func TestFinding(t *testing.T) {
	f := Finding{
		ID:             "test-finding",
		Severity:       SeverityHigh,
		Category:       "security",
		Title:          "Test Finding",
		Description:    "A test finding for unit tests",
		Files:          []string{"file1.go", "file2.go"},
		Lines:          []int{10, 20, 30},
		Recommendation: "Fix the issue",
		Confidence:     0.85,
		FoundAt:        time.Now(),
	}

	if f.ID != "test-finding" {
		t.Errorf("expected ID test-finding, got %s", f.ID)
	}
	if len(f.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(f.Files))
	}
}

func TestPodConfig(t *testing.T) {
	// Test standard pod categories
	if len(PodCategories) == 0 {
		t.Error("expected predefined pod categories")
	}

	for _, pod := range PodCategories {
		if pod.Name == "" {
			t.Error("pod name should not be empty")
		}
		if pod.AgentCount == 0 {
			t.Errorf("pod %s should have agents", pod.Name)
		}
	}
}

func TestSeverityOrder(t *testing.T) {
	if SeverityOrder[SeverityCritical] <= SeverityOrder[SeverityHigh] {
		t.Error("CRITICAL should be higher than HIGH")
	}
	if SeverityOrder[SeverityHigh] <= SeverityOrder[SeverityMedium] {
		t.Error("HIGH should be higher than MEDIUM")
	}
	if SeverityOrder[SeverityMedium] <= SeverityOrder[SeverityLow] {
		t.Error("MEDIUM should be higher than LOW")
	}
	if SeverityOrder[SeverityLow] <= SeverityOrder[SeverityPass] {
		t.Error("LOW should be higher than PASS")
	}
}
