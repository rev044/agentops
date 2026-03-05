package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseBeadIDs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantFirst string
		wantErr   bool
	}{
		{
			name:      "array of beads",
			input:     `[{"id":"ol-001"},{"id":"ol-002"},{"id":"ol-003"}]`,
			wantCount: 3,
			wantFirst: "ol-001",
		},
		{
			name:      "single bead object",
			input:     `{"id":"ol-001"}`,
			wantCount: 1,
			wantFirst: "ol-001",
		},
		{
			name:      "empty array",
			input:     `[]`,
			wantCount: 0,
		},
		{
			name:      "empty input",
			input:     "",
			wantCount: 0,
		},
		{
			name:    "invalid JSON",
			input:   "not json",
			wantErr: true,
		},
		{
			name:      "single object with empty ID",
			input:     `{"id":""}`,
			wantCount: 0,
		},
		{
			name:      "array with extra fields",
			input:     `[{"id":"ag-m0r","title":"test","status":"open"}]`,
			wantCount: 1,
			wantFirst: "ag-m0r",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBeadIDs([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("got %d IDs, want %d; got %v", len(got), tt.wantCount, got)
			}
			if tt.wantFirst != "" && len(got) > 0 && got[0] != tt.wantFirst {
				t.Errorf("first ID = %q, want %q", got[0], tt.wantFirst)
			}
		})
	}
}

func TestDefaultFireConfig(t *testing.T) {
	cfg := DefaultFireConfig()
	if cfg.MaxPolecats != 4 {
		t.Errorf("MaxPolecats = %d, want 4", cfg.MaxPolecats)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", cfg.MaxRetries)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("PollInterval = %v, want 30s", cfg.PollInterval)
	}
	if cfg.BackoffBase != 30*time.Second {
		t.Errorf("BackoffBase = %v, want 30s", cfg.BackoffBase)
	}
}

func TestIsComplete(t *testing.T) {
	t.Run("empty ready and burning is complete", func(t *testing.T) {
		state := &FireState{
			Ready:   []string{},
			Burning: []string{},
			Reaped:  []string{"ol-001"},
		}
		if !isComplete(state) {
			t.Error("expected complete when ready and burning are empty")
		}
	})

	t.Run("has ready issues means not complete", func(t *testing.T) {
		state := &FireState{
			Ready:   []string{"ol-001"},
			Burning: []string{},
		}
		if isComplete(state) {
			t.Error("expected not complete when ready has issues")
		}
	})

	t.Run("has burning issues means not complete", func(t *testing.T) {
		state := &FireState{
			Ready:   []string{},
			Burning: []string{"ol-002"},
		}
		if isComplete(state) {
			t.Error("expected not complete when burning has issues")
		}
	})

	t.Run("nil slices treated as empty (complete)", func(t *testing.T) {
		state := &FireState{}
		if !isComplete(state) {
			t.Error("expected complete for empty state")
		}
	})
}

// ---------------------------------------------------------------------------
// FireState
// ---------------------------------------------------------------------------

func TestFireCoverage_FireStateJSON(t *testing.T) {
	state := FireState{
		EpicID:   "epic-001",
		Rig:      "agentops",
		Ready:    []string{"issue-1", "issue-2"},
		Burning:  []string{"issue-3"},
		Reaped:   []string{"issue-4"},
		Blocked:  []string{"issue-5"},
		ConvoyID: "convoy-123",
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded FireState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.EpicID != state.EpicID {
		t.Errorf("EpicID = %q, want %q", decoded.EpicID, state.EpicID)
	}
	if len(decoded.Ready) != 2 {
		t.Errorf("Ready len = %d, want 2", len(decoded.Ready))
	}
	if len(decoded.Burning) != 1 {
		t.Errorf("Burning len = %d, want 1", len(decoded.Burning))
	}
}

// ---------------------------------------------------------------------------
// RetryInfo
// ---------------------------------------------------------------------------

func TestFireCoverage_RetryInfoJSON(t *testing.T) {
	now := time.Now()
	info := RetryInfo{
		IssueID:      "issue-1",
		Attempt:      2,
		LastAttempt:  now,
		NextAttempt:  now.Add(60 * time.Second),
		FailureNotes: []string{"first fail", "second fail"},
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded RetryInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.IssueID != "issue-1" {
		t.Errorf("IssueID = %q, want %q", decoded.IssueID, "issue-1")
	}
	if decoded.Attempt != 2 {
		t.Errorf("Attempt = %d, want 2", decoded.Attempt)
	}
	if len(decoded.FailureNotes) != 2 {
		t.Errorf("FailureNotes len = %d, want 2", len(decoded.FailureNotes))
	}
}

// ---------------------------------------------------------------------------
// FireConfig
// ---------------------------------------------------------------------------

func TestFireCoverage_FireConfig(t *testing.T) {
	cfg := FireConfig{
		EpicID:       "epic-001",
		Rig:          "agentops",
		MaxPolecats:  8,
		PollInterval: 60 * time.Second,
		MaxRetries:   5,
		BackoffBase:  15 * time.Second,
	}

	if cfg.EpicID != "epic-001" {
		t.Errorf("EpicID = %q, want %q", cfg.EpicID, "epic-001")
	}
	if cfg.MaxPolecats != 8 {
		t.Errorf("MaxPolecats = %d, want 8", cfg.MaxPolecats)
	}
	if cfg.PollInterval != 60*time.Second {
		t.Errorf("PollInterval = %v, want 60s", cfg.PollInterval)
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", cfg.MaxRetries)
	}
	if cfg.BackoffBase != 15*time.Second {
		t.Errorf("BackoffBase = %v, want 15s", cfg.BackoffBase)
	}
}

// ---------------------------------------------------------------------------
// containsString
// ---------------------------------------------------------------------------

func TestFireCoverage_ContainsString(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		s     string
		want  bool
	}{
		{name: "found", slice: []string{"a", "b", "c"}, s: "b", want: true},
		{name: "not found", slice: []string{"a", "b", "c"}, s: "d", want: false},
		{name: "empty slice", slice: nil, s: "a", want: false},
		{name: "empty string found", slice: []string{""}, s: "", want: true},
		{name: "first element", slice: []string{"x", "y"}, s: "x", want: true},
		{name: "last element", slice: []string{"x", "y"}, s: "y", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsString(tt.slice, tt.s)
			if got != tt.want {
				t.Errorf("containsString(%v, %q) = %v, want %v", tt.slice, tt.s, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// collectDueRetries
// ---------------------------------------------------------------------------

func TestFireCoverage_CollectDueRetries(t *testing.T) {
	t.Run("collects due items", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)
		future := time.Now().Add(1 * time.Hour)

		queue := map[string]*RetryInfo{
			"issue-1": {IssueID: "issue-1", NextAttempt: past},
			"issue-2": {IssueID: "issue-2", NextAttempt: future},
			"issue-3": {IssueID: "issue-3", NextAttempt: past},
		}

		due := collectDueRetries(queue, 10)
		if len(due) != 2 {
			t.Errorf("due = %d, want 2", len(due))
		}
		// Collected items should be removed from queue
		if len(queue) != 1 {
			t.Errorf("remaining queue = %d, want 1", len(queue))
		}
	})

	t.Run("respects capacity", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)
		queue := map[string]*RetryInfo{
			"issue-1": {IssueID: "issue-1", NextAttempt: past},
			"issue-2": {IssueID: "issue-2", NextAttempt: past},
			"issue-3": {IssueID: "issue-3", NextAttempt: past},
		}

		due := collectDueRetries(queue, 1)
		if len(due) != 1 {
			t.Errorf("due = %d, want 1 (capped)", len(due))
		}
	})

	t.Run("empty queue", func(t *testing.T) {
		queue := map[string]*RetryInfo{}
		due := collectDueRetries(queue, 10)
		if len(due) != 0 {
			t.Errorf("due = %d, want 0", len(due))
		}
	})
}

// ---------------------------------------------------------------------------
// collectReadyIssues
// ---------------------------------------------------------------------------

func TestFireCoverage_CollectReadyIssues(t *testing.T) {
	t.Run("adds new ready issues", func(t *testing.T) {
		result := collectReadyIssues([]string{"a", "b", "c"}, nil, 5)
		if len(result) != 3 {
			t.Errorf("len = %d, want 3", len(result))
		}
	})

	t.Run("skips duplicates from already", func(t *testing.T) {
		result := collectReadyIssues([]string{"a", "b"}, []string{"a"}, 5)
		if len(result) != 2 {
			t.Errorf("len = %d, want 2", len(result))
		}
		// Should have "a" (from already) and "b" (new)
		found := make(map[string]bool)
		for _, id := range result {
			found[id] = true
		}
		if !found["a"] || !found["b"] {
			t.Errorf("expected both a and b, got %v", result)
		}
	})

	t.Run("respects capacity", func(t *testing.T) {
		result := collectReadyIssues([]string{"a", "b", "c"}, nil, 2)
		if len(result) != 2 {
			t.Errorf("len = %d, want 2 (capped)", len(result))
		}
	})

	t.Run("already exceeds capacity still includes new until break", func(t *testing.T) {
		result := collectReadyIssues([]string{"c"}, []string{"a", "b"}, 2)
		// already=[a,b] is len 2, then "c" is appended (len 3), capacity check breaks
		if len(result) != 3 {
			t.Errorf("len = %d, want 3 (already items always included)", len(result))
		}
	})

	t.Run("empty ready", func(t *testing.T) {
		result := collectReadyIssues(nil, []string{"a"}, 5)
		if len(result) != 1 {
			t.Errorf("len = %d, want 1", len(result))
		}
	})
}

// ---------------------------------------------------------------------------
// printState (smoke test)
// ---------------------------------------------------------------------------

func TestFireCoverage_PrintState(t *testing.T) {
	// Just ensure no panic
	state := &FireState{
		Ready:   []string{"a", "b"},
		Burning: []string{"c"},
		Reaped:  []string{"d", "e", "f"},
		Blocked: []string{},
	}
	printState(state)
}

// ---------------------------------------------------------------------------
// sendMail
// ---------------------------------------------------------------------------

func TestFireCoverage_SendMail(t *testing.T) {
	tmp := t.TempDir()
	prev, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prev) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := sendMail("mayor", "Test message body", "blocker"); err != nil {
		t.Fatalf("sendMail: %v", err)
	}

	// Verify message file was created
	messagesPath := filepath.Join(tmp, ".agents", "mail", "messages.jsonl")
	data, err := os.ReadFile(messagesPath)
	if err != nil {
		t.Fatalf("read messages: %v", err)
	}

	var msg Message
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		// JSONL has newline, try trimming
		lines := trimLines(string(data))
		if len(lines) > 0 {
			if err := json.Unmarshal([]byte(lines[0]), &msg); err != nil {
				t.Fatalf("unmarshal message: %v", err)
			}
		}
	}

	if msg.From != "fire-loop" {
		t.Errorf("From = %q, want %q", msg.From, "fire-loop")
	}
	if msg.To != "mayor" {
		t.Errorf("To = %q, want %q", msg.To, "mayor")
	}
	if msg.Body != "Test message body" {
		t.Errorf("Body = %q, want %q", msg.Body, "Test message body")
	}
	if msg.Type != "blocker" {
		t.Errorf("Type = %q, want %q", msg.Type, "blocker")
	}
}

// trimLines is a test helper that splits and trims lines.
func trimLines(s string) []string {
	var lines []string
	for _, line := range fireCovSplitNonEmpty(s) {
		trimmed := trimLine(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines
}

func fireCovSplitNonEmpty(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 {
				result = append(result, line)
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

func trimLine(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// ---------------------------------------------------------------------------
// escalatePhase
// ---------------------------------------------------------------------------

func TestFireCoverage_EscalatePhase(t *testing.T) {
	tmp := t.TempDir()
	prev, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prev) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	cfg := FireConfig{
		MaxRetries:  2,
		BackoffBase: 100 * time.Millisecond,
	}

	t.Run("first failure schedules retry", func(t *testing.T) {
		retryQueue := make(map[string]*RetryInfo)
		escalated, err := escalatePhase([]string{"issue-1"}, retryQueue, cfg)
		if err != nil {
			t.Fatalf("escalate: %v", err)
		}
		if len(escalated) != 0 {
			t.Errorf("expected 0 escalated on first failure, got %d", len(escalated))
		}
		if _, ok := retryQueue["issue-1"]; !ok {
			t.Error("expected issue-1 in retry queue")
		}
		if retryQueue["issue-1"].Attempt != 1 {
			t.Errorf("Attempt = %d, want 1", retryQueue["issue-1"].Attempt)
		}
	})

	t.Run("max retries reached escalates", func(t *testing.T) {
		retryQueue := map[string]*RetryInfo{
			"issue-2": {IssueID: "issue-2", Attempt: 1},
		}
		escalated, err := escalatePhase([]string{"issue-2"}, retryQueue, cfg)
		if err != nil {
			t.Fatalf("escalate: %v", err)
		}
		if len(escalated) != 1 {
			t.Errorf("expected 1 escalated, got %d", len(escalated))
		}
		if _, ok := retryQueue["issue-2"]; ok {
			t.Error("expected issue-2 removed from retry queue after escalation")
		}
	})

	t.Run("no failures no escalation", func(t *testing.T) {
		retryQueue := make(map[string]*RetryInfo)
		escalated, err := escalatePhase(nil, retryQueue, cfg)
		if err != nil {
			t.Fatalf("escalate: %v", err)
		}
		if len(escalated) != 0 {
			t.Errorf("expected 0 escalated, got %d", len(escalated))
		}
	})
}

// ---------------------------------------------------------------------------
// parseBeadIDs edge cases
// ---------------------------------------------------------------------------

func TestFireCoverage_ParseBeadIDsEdgeCases(t *testing.T) {
	t.Run("array with multiple fields", func(t *testing.T) {
		input := `[{"id":"a","status":"open","title":"test"},{"id":"b","status":"closed"}]`
		ids, err := parseBeadIDs([]byte(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ids) != 2 {
			t.Errorf("expected 2 IDs, got %d", len(ids))
		}
	})

	t.Run("single object with extra fields", func(t *testing.T) {
		input := `{"id":"solo","title":"my issue"}`
		ids, err := parseBeadIDs([]byte(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ids) != 1 || ids[0] != "solo" {
			t.Errorf("expected [solo], got %v", ids)
		}
	})

	t.Run("nested invalid JSON", func(t *testing.T) {
		input := `{invalid}`
		_, err := parseBeadIDs([]byte(input))
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

// ---------------------------------------------------------------------------
// isComplete additional cases
// ---------------------------------------------------------------------------

func TestFireCoverage_IsCompleteEdgeCases(t *testing.T) {
	t.Run("both ready and burning non-empty", func(t *testing.T) {
		state := &FireState{
			Ready:   []string{"a"},
			Burning: []string{"b"},
		}
		if isComplete(state) {
			t.Error("expected not complete")
		}
	})

	t.Run("with blocked and reaped but no ready/burning", func(t *testing.T) {
		state := &FireState{
			Blocked: []string{"x"},
			Reaped:  []string{"y", "z"},
		}
		if !isComplete(state) {
			t.Error("expected complete (blocked/reaped don't affect completion)")
		}
	})
}

// ---------------------------------------------------------------------------
// DefaultFireConfig field validation
// ---------------------------------------------------------------------------

func TestFireCoverage_DefaultFireConfigFields(t *testing.T) {
	cfg := DefaultFireConfig()
	if cfg.EpicID != "" {
		t.Errorf("EpicID = %q, want empty", cfg.EpicID)
	}
	if cfg.Rig != "" {
		t.Errorf("Rig = %q, want empty", cfg.Rig)
	}
	if cfg.MaxPolecats <= 0 {
		t.Errorf("MaxPolecats = %d, want > 0", cfg.MaxPolecats)
	}
	if cfg.PollInterval <= 0 {
		t.Errorf("PollInterval = %v, want > 0", cfg.PollInterval)
	}
	if cfg.MaxRetries <= 0 {
		t.Errorf("MaxRetries = %d, want > 0", cfg.MaxRetries)
	}
	if cfg.BackoffBase <= 0 {
		t.Errorf("BackoffBase = %v, want > 0", cfg.BackoffBase)
	}
}

// ===========================================================================
// fire.go — runFireIteration (from fire_deep_test.go)
// ===========================================================================

func TestFire_runFireIteration_emptyEpic(t *testing.T) {
	// runFireIteration calls findPhase which calls bdReady, bdListByStatus, etc.
	// With no bd on PATH, findPhase returns an error, and runFireIteration wraps it.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("PATH", tmp) // no bd on PATH

	cfg := FireConfig{
		EpicID:       "test-epic",
		Rig:          "test-rig",
		MaxPolecats:  2,
		PollInterval: time.Second,
		MaxRetries:   3,
		BackoffBase:  time.Second,
	}
	retryQueue := make(map[string]*RetryInfo)

	done, err := runFireIteration(cfg, retryQueue)
	if done {
		t.Error("expected done=false when findPhase fails")
	}
	if err == nil {
		t.Fatal("expected error when bd is not available")
	}
}

// ===========================================================================
// fire.go — ignitePhase (from fire_deep_test.go)
// ===========================================================================

func TestFire_ignitePhase_atCapacity(t *testing.T) {
	state := &FireState{
		Burning: []string{"a", "b", "c", "d"},
	}
	cfg := FireConfig{MaxPolecats: 4, Rig: "test"}
	retryQueue := make(map[string]*RetryInfo)

	ignited, err := ignitePhase(state, cfg, retryQueue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ignited) != 0 {
		t.Errorf("expected no ignitions at capacity, got %v", ignited)
	}
}

func TestFire_ignitePhase_noReadyNoDue(t *testing.T) {
	state := &FireState{
		Burning: []string{"a"},
	}
	cfg := FireConfig{MaxPolecats: 4, Rig: "test"}
	retryQueue := make(map[string]*RetryInfo)

	ignited, err := ignitePhase(state, cfg, retryQueue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ignited) != 0 {
		t.Errorf("expected no ignitions with no ready/due, got %v", ignited)
	}
}

func TestFire_ignitePhase_withReadyButNoGt(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("PATH", tmp) // no gt on PATH

	state := &FireState{
		Ready:   []string{"issue-1"},
		Burning: []string{},
	}
	cfg := FireConfig{MaxPolecats: 4, Rig: "test"}
	retryQueue := make(map[string]*RetryInfo)

	// ignitePhase calls slingIssues which calls gtSling; without gt, sling fails
	// but ignitePhase returns nil error (failures are logged, not returned)
	ignited, err := ignitePhase(state, cfg, retryQueue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No items should be successfully ignited since gt is missing
	if len(ignited) != 0 {
		t.Errorf("expected 0 ignited without gt, got %d", len(ignited))
	}
}

func TestFire_ignitePhase_withDueRetries(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("PATH", tmp) // no gt

	state := &FireState{
		Burning: []string{},
	}
	cfg := FireConfig{MaxPolecats: 4, Rig: "test"}
	retryQueue := map[string]*RetryInfo{
		"retry-issue": {
			IssueID:     "retry-issue",
			Attempt:     1,
			NextAttempt: time.Now().Add(-time.Minute), // already due
		},
	}

	ignited, err := ignitePhase(state, cfg, retryQueue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// retryQueue entry should have been consumed even though sling failed
	if _, exists := retryQueue["retry-issue"]; exists {
		t.Error("expected retry-issue to be removed from retry queue")
	}
	_ = ignited
}

// ===========================================================================
// fire.go — reapPhase (from fire_deep_test.go)
// ===========================================================================

func TestFire_reapPhase_noBurning(t *testing.T) {
	state := &FireState{Burning: []string{}}
	reaped, failures, err := reapPhase(state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reaped) != 0 {
		t.Errorf("expected no reaped, got %v", reaped)
	}
	if len(failures) != 0 {
		t.Errorf("expected no failures, got %v", failures)
	}
}

func TestFire_reapPhase_withBurningNoBd(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("PATH", tmp) // no bd

	state := &FireState{Burning: []string{"issue-1", "issue-2"}}
	reaped, failures, err := reapPhase(state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// bdShowStatus fails for all issues -> all skipped (continue in switch)
	if len(reaped) != 0 {
		t.Errorf("expected 0 reaped without bd, got %d", len(reaped))
	}
	if len(failures) != 0 {
		t.Errorf("expected 0 failures without bd, got %d", len(failures))
	}
}

// ===========================================================================
// fire.go — sendMail (from fire_deep_test.go)
// ===========================================================================

func TestFire_sendMail(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	err := sendMail("mayor", "test message body", "blocker")
	if err != nil {
		t.Fatalf("sendMail failed: %v", err)
	}

	// Verify the file was created
	messagesPath := filepath.Join(tmp, ".agents", "mail", "messages.jsonl")
	if _, err := os.Stat(messagesPath); os.IsNotExist(err) {
		t.Fatal("messages.jsonl was not created")
	}

	content, err := os.ReadFile(messagesPath)
	if err != nil {
		t.Fatalf("read messages.jsonl: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("messages.jsonl is empty")
	}
}

// ===========================================================================
// fire.go — printState (from fire_deep_test.go)
// ===========================================================================

func TestFire_printState_doesNotPanic(t *testing.T) {
	// printState prints to stdout; verify it doesn't panic
	state := &FireState{
		Ready:   []string{"a", "b"},
		Burning: []string{"c"},
		Reaped:  []string{"d", "e", "f"},
		Blocked: []string{"g"},
	}
	// Should not panic
	printState(state)
}

func TestFire_printState_emptyState(t *testing.T) {
	state := &FireState{}
	printState(state)
}
