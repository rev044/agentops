package context

import (
	"testing"
	"time"
)

func TestNewBudgetTracker(t *testing.T) {
	bt := NewBudgetTracker("test-session")

	if bt.SessionID != "test-session" {
		t.Errorf("expected SessionID test-session, got %s", bt.SessionID)
	}
	if bt.MaxTokens != DefaultMaxTokens {
		t.Errorf("expected MaxTokens %d, got %d", DefaultMaxTokens, bt.MaxTokens)
	}
	if bt.EstimatedUsage != 0 {
		t.Errorf("expected EstimatedUsage 0, got %d", bt.EstimatedUsage)
	}
}

func TestBudgetTrackerUsage(t *testing.T) {
	bt := NewBudgetTracker("test")
	bt.MaxTokens = 100000

	bt.UpdateUsage(40000)
	if bt.GetUsagePercent() != 0.4 {
		t.Errorf("expected 40%% usage, got %.2f", bt.GetUsagePercent())
	}

	bt.AddTokens(20000)
	if bt.EstimatedUsage != 60000 {
		t.Errorf("expected 60000 tokens, got %d", bt.EstimatedUsage)
	}
}

func TestBudgetTrackerStatus(t *testing.T) {
	bt := NewBudgetTracker("test")
	bt.MaxTokens = 100000

	// Optimal
	bt.UpdateUsage(30000)
	if bt.GetStatus() != StatusOptimal {
		t.Errorf("expected OPTIMAL at 30%%, got %s", bt.GetStatus())
	}

	// Warning
	bt.UpdateUsage(65000)
	if bt.GetStatus() != StatusWarning {
		t.Errorf("expected WARNING at 65%%, got %s", bt.GetStatus())
	}

	// Critical
	bt.UpdateUsage(85000)
	if bt.GetStatus() != StatusCritical {
		t.Errorf("expected CRITICAL at 85%%, got %s", bt.GetStatus())
	}
}

func TestBudgetTrackerNeedsSummarization(t *testing.T) {
	bt := NewBudgetTracker("test")
	bt.MaxTokens = 100000

	bt.UpdateUsage(70000)
	if bt.NeedsSummarization() {
		t.Error("should not need summarization at 70%")
	}

	bt.UpdateUsage(85000)
	if !bt.NeedsSummarization() {
		t.Error("should need summarization at 85%")
	}
}

func TestBudgetTrackerCheckpoints(t *testing.T) {
	bt := NewBudgetTracker("test")
	bt.MaxTokens = 100000
	bt.UpdateUsage(50000)

	cp := bt.CreateCheckpoint("cp1", "Completed feature X", []string{"file1.go"}, "passing")

	if cp.ID != "cp1" {
		t.Errorf("expected checkpoint ID cp1, got %s", cp.ID)
	}
	if cp.TokenUsage != 50000 {
		t.Errorf("expected TokenUsage 50000, got %d", cp.TokenUsage)
	}
	if len(bt.Checkpoints) != 1 {
		t.Errorf("expected 1 checkpoint, got %d", len(bt.Checkpoints))
	}

	last := bt.GetLastCheckpoint()
	if last == nil || last.ID != "cp1" {
		t.Error("GetLastCheckpoint failed")
	}
}

func TestBudgetTrackerRecordSummarization(t *testing.T) {
	bt := NewBudgetTracker("test")
	bt.MaxTokens = 100000
	bt.UpdateUsage(90000)

	bt.RecordSummarization(90000, 50000, []string{"file_changes", "failing_tests"})

	if bt.EstimatedUsage != 50000 {
		t.Errorf("expected usage updated to 50000, got %d", bt.EstimatedUsage)
	}
	if len(bt.SummarizationEvents) != 1 {
		t.Errorf("expected 1 summarization event, got %d", len(bt.SummarizationEvents))
	}

	event := bt.SummarizationEvents[0]
	if event.TokensSaved != 40000 {
		t.Errorf("expected 40000 tokens saved, got %d", event.TokensSaved)
	}
}

func TestBudgetTrackerReport(t *testing.T) {
	bt := NewBudgetTracker("test")
	bt.MaxTokens = 100000
	bt.UpdateUsage(60000)
	bt.CreateCheckpoint("cp1", "test", nil, "passing")

	report := bt.GetReport()

	if report.SessionID != "test" {
		t.Errorf("expected SessionID test, got %s", report.SessionID)
	}
	if report.UsagePercent != 60 {
		t.Errorf("expected UsagePercent 60, got %.2f", report.UsagePercent)
	}
	if report.TokensRemaining != 40000 {
		t.Errorf("expected TokensRemaining 40000, got %d", report.TokensRemaining)
	}
	if report.CheckpointCount != 1 {
		t.Errorf("expected CheckpointCount 1, got %d", report.CheckpointCount)
	}
}

func TestEstimateTokens(t *testing.T) {
	text := "This is a test string with some words"
	tokens := EstimateTokens(text)

	// Rough 4 chars per token
	expected := len(text) / 4
	if tokens != expected {
		t.Errorf("expected %d tokens, got %d", expected, tokens)
	}
}

func TestSummarizerClassifyItem(t *testing.T) {
	bt := NewBudgetTracker("test")
	s := NewSummarizer(bt)

	tests := []struct {
		itemType string
		expected SummaryPriority
	}{
		{"failing_test", PriorityCritical},
		{"file_change", PriorityCritical},
		{"critical_finding", PriorityCritical},
		{"high_finding", PriorityHigh},
		{"medium_finding", PriorityMedium},
		{"low_finding", PriorityLow},
		{"context", PriorityLow},
	}

	for _, tt := range tests {
		result := s.ClassifyItem(tt.itemType, "")
		if result != tt.expected {
			t.Errorf("ClassifyItem(%s) = %d, expected %d", tt.itemType, result, tt.expected)
		}
	}
}

func TestSummarizerCreateContextItem(t *testing.T) {
	bt := NewBudgetTracker("test")
	s := NewSummarizer(bt)

	item := s.CreateContextItem("failing_test", "Test xyz failed", map[string]string{"file": "test.go"})

	if item.Type != "failing_test" {
		t.Errorf("expected type failing_test, got %s", item.Type)
	}
	if item.Priority != PriorityCritical {
		t.Errorf("expected CRITICAL priority for failing test, got %d", item.Priority)
	}
	if item.TokenEstimate == 0 {
		t.Error("expected non-zero token estimate")
	}
}

func TestSummarizerSummarizeContext(t *testing.T) {
	bt := NewBudgetTracker("test")
	bt.MaxTokens = 1000
	bt.UpdateUsage(900) // 90% usage

	s := NewSummarizer(bt)
	s.Config.TargetUsage = 0.5

	items := []ContextItem{
		{Type: "critical_finding", Priority: PriorityCritical, Content: "Critical issue", TokenEstimate: 100},
		{Type: "high_finding", Priority: PriorityHigh, Content: "High priority issue", TokenEstimate: 100},
		{Type: "low_finding", Priority: PriorityLow, Content: "Low priority issue with lots of detail that could be dropped", TokenEstimate: 300},
	}

	result, event := s.SummarizeContext(items)

	if event.TokensBefore != 900 {
		t.Errorf("expected TokensBefore 900, got %d", event.TokensBefore)
	}
	if event.TokensSaved <= 0 {
		t.Error("expected tokens to be saved")
	}

	// Critical items should always be preserved
	hasCritical := false
	for _, item := range result {
		if item.Priority == PriorityCritical {
			hasCritical = true
			break
		}
	}
	if !hasCritical {
		t.Error("critical items should be preserved")
	}
}

func TestSummarizeState(t *testing.T) {
	state := SummarizeState{
		SessionID:    "test-session",
		Timestamp:    time.Now(),
		FilesChanged: []string{"file1.go", "file2.go"},
		TestStatus:   "failing",
		FailingTests: []string{"TestFoo", "TestBar"},
		CurrentTask:  "Implement feature X",
	}

	bt := NewBudgetTracker("test")
	s := NewSummarizer(bt)

	context := s.GenerateResumptionContext(state)

	if context == "" {
		t.Error("expected non-empty resumption context")
	}
	if !contains(context, "file1.go") {
		t.Error("expected files in context")
	}
	if !contains(context, "TestFoo") {
		t.Error("expected failing tests in context")
	}
}

func TestSummaryConfig(t *testing.T) {
	config := DefaultSummaryConfig()

	if config.TargetUsage != 0.5 {
		t.Errorf("expected TargetUsage 0.5, got %f", config.TargetUsage)
	}
	if !config.PreserveFailingTests {
		t.Error("expected PreserveFailingTests true")
	}
	if !config.PreserveFileChanges {
		t.Error("expected PreserveFileChanges true")
	}
}

func TestBudgetRecommendation(t *testing.T) {
	bt := NewBudgetTracker("test")
	bt.MaxTokens = 100000

	bt.UpdateUsage(30000)
	rec := bt.GetRecommendation()
	if !contains(rec, "OPTIMAL") {
		t.Error("expected OPTIMAL recommendation at 30%")
	}

	bt.UpdateUsage(65000)
	rec = bt.GetRecommendation()
	if !contains(rec, "MEDIUM") {
		t.Error("expected MEDIUM recommendation at 65%")
	}

	bt.UpdateUsage(85000)
	rec = bt.GetRecommendation()
	if !contains(rec, "HIGH") {
		t.Error("expected HIGH recommendation at 85%")
	}

	bt.UpdateUsage(95000)
	rec = bt.GetRecommendation()
	if !contains(rec, "CRITICAL") {
		t.Error("expected CRITICAL recommendation at 95%")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
