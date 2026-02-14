package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractCouncilVerdict(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
		wantErr  bool
	}{
		{
			name:     "PASS verdict",
			content:  "# Pre-Mortem\n\n## Council Verdict: PASS\n\nDetails here.",
			expected: "PASS",
		},
		{
			name:     "WARN verdict",
			content:  "## Council Verdict: WARN\n\nSome concerns.",
			expected: "WARN",
		},
		{
			name:     "FAIL verdict",
			content:  "## Council Verdict: FAIL\n\nCritical issues.",
			expected: "FAIL",
		},
		{
			name:    "no verdict",
			content: "# Report\n\nNo verdict line here.",
			wantErr: true,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
		},
		{
			name:     "verdict with extra whitespace",
			content:  "## Council Verdict:  PASS \n",
			expected: "PASS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "report.md")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			verdict, err := extractCouncilVerdict(path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got verdict %q", verdict)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if verdict != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, verdict)
			}
		})
	}
}

func TestExtractCouncilFindings(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		max      int
		expected int
	}{
		{
			name:     "structured findings",
			content:  "FINDING: Missing auth | FIX: Add middleware | REF: auth.go:10\nFINDING: No tests | FIX: Add tests | REF: auth_test.go",
			max:      5,
			expected: 2,
		},
		{
			name:     "max limit applied",
			content:  "FINDING: A | FIX: B | REF: C\nFINDING: D | FIX: E | REF: F\nFINDING: G | FIX: H | REF: I",
			max:      2,
			expected: 2,
		},
		{
			name:     "fallback to markdown findings",
			content:  "## Shared Findings\n\n1. **Missing auth** — No middleware\n2. **No tests** — Zero coverage",
			max:      5,
			expected: 2,
		},
		{
			name:     "no findings",
			content:  "# Empty report",
			max:      5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "report.md")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			findings, err := extractCouncilFindings(path, tt.max)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(findings) != tt.expected {
				t.Errorf("expected %d findings, got %d", tt.expected, len(findings))
			}
		})
	}
}

func TestBuildPromptForPhase(t *testing.T) {
	tests := []struct {
		name     string
		phase    int
		state    *phasedState
		contains string
	}{
		{
			name:     "research phase",
			phase:    1,
			state:    &phasedState{Goal: "add auth"},
			contains: `/research "add auth" --auto`,
		},
		{
			name:     "plan phase",
			phase:    2,
			state:    &phasedState{Goal: "add auth"},
			contains: `/plan "add auth" --auto`,
		},
		{
			name:     "pre-mortem normal",
			phase:    3,
			state:    &phasedState{},
			contains: "/pre-mortem",
		},
		{
			name:     "pre-mortem fast path",
			phase:    3,
			state:    &phasedState{FastPath: true},
			contains: "--quick",
		},
		{
			name:     "crank with epic",
			phase:    4,
			state:    &phasedState{EpicID: "ag-5k2"},
			contains: "/crank ag-5k2",
		},
		{
			name:     "crank with test-first",
			phase:    4,
			state:    &phasedState{EpicID: "ag-5k2", TestFirst: true},
			contains: "--test-first",
		},
		{
			name:     "vibe normal",
			phase:    5,
			state:    &phasedState{},
			contains: "/vibe recent",
		},
		{
			name:     "vibe fast path",
			phase:    5,
			state:    &phasedState{FastPath: true},
			contains: "/vibe --quick recent",
		},
		{
			name:     "post-mortem with epic",
			phase:    6,
			state:    &phasedState{EpicID: "ag-5k2"},
			contains: "/post-mortem ag-5k2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := buildPromptForPhase(tt.phase, tt.state, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !containsStr(prompt, tt.contains) {
				t.Errorf("prompt %q does not contain %q", prompt, tt.contains)
			}
		})
	}
}

func TestBuildPromptForPhase_Retry(t *testing.T) {
	state := &phasedState{Goal: "add auth", EpicID: "ag-5k2"}
	retryCtx := &retryContext{
		Attempt: 2,
		Findings: []finding{
			{Description: "Missing error handling", Fix: "Add try-catch", Ref: "auth.go:42"},
		},
		Verdict: "FAIL",
	}

	// Pre-mortem retry → re-plan
	prompt, err := buildRetryPrompt(3, state, retryCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsStr(prompt, "/plan") {
		t.Errorf("pre-mortem retry should invoke /plan, got: %q", prompt)
	}
	if !containsStr(prompt, "Missing error handling") {
		t.Errorf("retry prompt should contain finding description, got: %q", prompt)
	}

	// Vibe retry → re-crank
	prompt, err = buildRetryPrompt(5, state, retryCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsStr(prompt, "/crank") {
		t.Errorf("vibe retry should invoke /crank, got: %q", prompt)
	}
}

func TestPhasedState_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	stateDir := filepath.Join(tmpDir, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}

	original := &phasedState{
		Goal:      "test goal",
		EpicID:    "ag-test",
		Phase:     3,
		Cycle:     1,
		FastPath:  true,
		TestFirst: false,
		Verdicts:  map[string]string{"pre_mortem": "PASS"},
		Attempts:  map[string]int{"phase_3": 1},
		StartedAt: "2026-02-14T12:00:00Z",
	}

	if err := savePhasedState(tmpDir, original); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := loadPhasedState(tmpDir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if loaded.Goal != original.Goal {
		t.Errorf("goal: got %q, want %q", loaded.Goal, original.Goal)
	}
	if loaded.EpicID != original.EpicID {
		t.Errorf("epic_id: got %q, want %q", loaded.EpicID, original.EpicID)
	}
	if loaded.Phase != original.Phase {
		t.Errorf("phase: got %d, want %d", loaded.Phase, original.Phase)
	}
	if loaded.FastPath != original.FastPath {
		t.Errorf("fast_path: got %v, want %v", loaded.FastPath, original.FastPath)
	}
	if loaded.Verdicts["pre_mortem"] != "PASS" {
		t.Errorf("verdicts: got %v, want pre_mortem=PASS", loaded.Verdicts)
	}

	// Verify JSON round-trip
	data, _ := json.Marshal(original)
	var roundTrip phasedState
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if roundTrip.Goal != original.Goal {
		t.Errorf("round-trip goal mismatch")
	}
}

func TestPhaseNameToNum(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{"research", 1},
		{"plan", 2},
		{"pre-mortem", 3},
		{"premortem", 3},
		{"pre_mortem", 3},
		{"crank", 4},
		{"implement", 4},
		{"vibe", 5},
		{"validate", 5},
		{"post-mortem", 6},
		{"postmortem", 6},
		{"post_mortem", 6},
		{"unknown", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := phaseNameToNum(tt.name)
			if got != tt.expected {
				t.Errorf("phaseNameToNum(%q) = %d, want %d", tt.name, got, tt.expected)
			}
		})
	}
}

func TestFindLatestCouncilReport(t *testing.T) {
	tmpDir := t.TempDir()
	councilDir := filepath.Join(tmpDir, ".agents", "council")
	if err := os.MkdirAll(councilDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create two reports with different timestamps
	if err := os.WriteFile(filepath.Join(councilDir, "2026-02-13-pre-mortem-auth.md"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(councilDir, "2026-02-14-pre-mortem-auth.md"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	// Unrelated report
	if err := os.WriteFile(filepath.Join(councilDir, "2026-02-14-vibe-recent.md"), []byte("vibe"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should find the latest pre-mortem report
	report, err := findLatestCouncilReport(tmpDir, "pre-mortem")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsStr(report, "2026-02-14-pre-mortem") {
		t.Errorf("expected latest report, got: %s", report)
	}

	// Should find vibe report
	report, err = findLatestCouncilReport(tmpDir, "vibe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !containsStr(report, "vibe-recent") {
		t.Errorf("expected vibe report, got: %s", report)
	}

	// Should error on missing pattern
	_, err = findLatestCouncilReport(tmpDir, "nonexistent")
	if err == nil {
		t.Error("expected error for missing pattern")
	}
}

// containsStr is a helper to check substring presence.
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
